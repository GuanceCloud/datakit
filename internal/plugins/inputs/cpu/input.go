// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cpu collect CPU metrics.
package cpu

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"go.uber.org/atomic"
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
	inputName   = "cpu"
	metricName  = inputName
)

var (
	_ inputs.ReadEnv   = (*Input)(nil)
	_ inputs.Singleton = (*Input)(nil)
	l                  = logger.DefaultSLogger(inputName)
	g                  = datakit.G("inputs_" + inputName)
)

type Input struct {
	TotalCPU                  bool `toml:"totalcpu"`                              // deprecated
	CollectCPUTime            bool `toml:"collect_cpu_time"`                      // deprecated
	ReportActive              bool `toml:"report_active"`                         // deprecated
	DisableTemperatureCollect bool `toml:"disable_temperature_collect,omitempty"` // deprecated

	PerCPU            bool `toml:"percpu"`
	EnableTemperature bool `toml:"enable_temperature"`
	EnableLoad5s      bool `toml:"enable_load5s"`

	Interval time.Duration
	Tags     map[string]string

	semStop      *cliutils.Sem
	collectCache []*point.Point
	// collectCacheLast1Ptr *cpuMeasurement
	feeder     dkio.Feeder
	mergedTags map[string]string
	tagger     datakit.GlobalTagger

	lastStats map[string]cpu.TimesStat
	load5s    atomic.Int32
	lastLoad1 float64
	ps        CPUStatInfo
}

func (ipt *Input) Run() {
	ipt.setup()

	if ipt.EnableLoad5s {
		g.Go(func(ctx context.Context) error {
			ipt.calLoad5s()
			return nil
		})
	}

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	start := ntp.Now()
	for {
		collectStart := time.Now()
		if err := ipt.collect(start.UnixNano()); err != nil {
			l.Errorf("collect: %s", err)
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorCategory(point.Metric),
			)
		}

		if len(ipt.collectCache) > 0 {
			if err := ipt.feeder.Feed(point.Metric, ipt.collectCache,
				dkio.WithCollectCost(time.Since(collectStart)),
				dkio.WithElection(false),
				dkio.WithSource(metricName)); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(inputName),
					metrics.WithLastErrorCategory(point.Metric),
				)
				l.Errorf("feed measurement: %s", err)
			}
		}

		select {
		case tt := <-tick.C:
			start = inputs.AlignTime(tt, start, ipt.Interval)

		case <-datakit.Exit.Wait():
			l.Infof("%s input exit", inputName)
			return
		case <-ipt.semStop.Wait():
			l.Infof("%s input return", inputName)
			return
		}
	}
}

func (ipt *Input) setup() {
	l = logger.SLogger(inputName)

	l.Infof("%s input started", inputName)
	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)
	ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")
	l.Debugf("merged tags: %+#v", ipt.mergedTags)
}

func (ipt *Input) collect(ptTS int64) error {
	ipt.collectCache = make([]*point.Point, 0)
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(ptTS))

	// totalCPU only
	cpuTimes, err := ipt.ps.CPUTimes(ipt.PerCPU, true)
	if err != nil {
		return fmt.Errorf("error gather cpu stats: %w", err)
	}

	var coreTemp map[string]float64
	if ipt.EnableTemperature {
		var errTemp error
		coreTemp, errTemp = CoreTemp()
		if errTemp != nil {
			l.Warn("failed to collect core temperature data: ", errTemp)
			ipt.EnableTemperature = false
			l.Warn("skip core temperature data collection")
		}
	}

	for _, cts := range cpuTimes {
		var kvs point.KVs

		kvs = kvs.Add("cpu", cts.CPU, true, true)

		_, total := CPUActiveTotalTime(cts)

		lastCts, ok := ipt.lastStats[cts.CPU]
		if !ok {
			continue
		}
		_, lastTotal := CPUActiveTotalTime(lastCts)
		totalDelta := total - lastTotal
		if totalDelta < 0 {
			l.Warn("error: current total cpu time less than previous total cpu time")
			break
		}
		if totalDelta == 0 {
			continue
		}
		cpuUsage, _ := CalculateUsage(cts, lastCts, totalDelta)

		kvs = kvs.Add("usage_user", cpuUsage.User, false, true)
		kvs = kvs.Add("usage_system", cpuUsage.System, false, true)
		kvs = kvs.Add("usage_idle", cpuUsage.Idle, false, true)
		kvs = kvs.Add("usage_nice", cpuUsage.Nice, false, true)
		kvs = kvs.Add("usage_iowait", cpuUsage.Iowait, false, true)
		kvs = kvs.Add("usage_irq", cpuUsage.Irq, false, true)
		kvs = kvs.Add("usage_softirq", cpuUsage.Softirq, false, true)
		kvs = kvs.Add("usage_steal", cpuUsage.Steal, false, true)
		kvs = kvs.Add("usage_guest", cpuUsage.Guest, false, true)
		kvs = kvs.Add("usage_guest_nice", cpuUsage.GuestNice, false, true)
		kvs = kvs.Add("usage_total", cpuUsage.Total, false, true)

		if ipt.EnableLoad5s {
			kvs = kvs.Add("load5s", ipt.getLoad5s(), false, true)
		}

		if len(coreTemp) > 0 && cts.CPU == "cpu-total" {
			if v, ok := coreTemp[cts.CPU]; ok {
				kvs = kvs.Add("core_temperature", v, false, true)
			}
		}

		for k, v := range ipt.mergedTags {
			kvs = kvs.AddTag(k, v)
		}

		ipt.collectCache = append(ipt.collectCache, point.NewPointV2(inputName, kvs, opts...))
	}
	// last cputimes stats
	ipt.lastStats = make(map[string]cpu.TimesStat)
	for _, cts := range cpuTimes {
		ipt.lastStats[cts.CPU] = cts
	}
	return nil
}

func (ipt *Input) getLoad5s() int {
	return int(ipt.load5s.Load())
}

// calLoad5s gets average load information every five seconds,
// calculates load5s and store it in Input.
func (ipt *Input) calLoad5s() {
	tick := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-tick.C:
			var load5s int32
			avg, err := load.Avg()
			if err != nil {
				l.Warnf("fail to get load average: %v", err)
				ipt.lastLoad1 = 0
				continue
			}
			if ipt.lastLoad1 == 0 || avg.Load1 == 0 {
				load5s = int32(avg.Load1)
			} else {
				load5s = int32(math.Round((2048*(2048*avg.Load1-10) - 1024 -
					1884*2048*ipt.lastLoad1) / 164 / 2048))
			}
			ipt.lastLoad1 = avg.Load1
			ipt.load5s.Store(load5s)
		case <-datakit.Exit.Wait():
			l.Infof("load5s calculator exited")
			return
		case <-ipt.semStop.Wait():
			l.Info("load5s calculator returned")
			return
		}
	}
}

func (*Input) Singleton() {}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}
func (*Input) Catalog() string      { return "host" }
func (*Input) SampleConfig() string { return sampleCfg }
func (*Input) AvailableArchs() []string {
	return []string{
		datakit.OSLabelLinux, datakit.OSLabelWindows,
		datakit.LabelK8s, datakit.LabelDocker,
	}
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&docMeasurement{},
	}
}

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "Interval", Type: doc.TimeDuration, Default: "`10s`", Desc: "Collect interval", DescZh: "采集器重复间隔时长"},
		{FieldName: "PerCPU", ENVName: "PERCPU", Type: doc.Boolean, Default: `false`, Desc: "Collect CPU usage per core", DescZh: "采集每一个 CPU 核"},
		{FieldName: "EnableTemperature", Type: doc.Boolean, Default: "`true`", Desc: "Enable to collect core temperature data", DescZh: "采集 CPU 温度"},
		{FieldName: "EnableLoad5s", ENVName: "ENABLE_LOAD5S", Type: doc.Boolean, Default: "`false`", Desc: "Enable gets average load information every five seconds", DescZh: "每五秒钟获取一次平均负载信息"},
		{FieldName: "Tags", Type: doc.String, Example: "`tag1=value1,tag2=value2`", Desc: "Customize tags. If there is a tag with the same name in the configuration file, it will be overwritten", DescZh: "自定义标签。如果配置文件有同名标签，将会覆盖它"},
	}

	return doc.SetENVDoc("ENV_INPUT_CPU_", infos)
}

// ReadEnv support envs：
//
//	ENV_INPUT_CPU_PERCPU : booler
//	ENV_INPUT_CPU_ENABLE_TEMPERATURE : booler
//	ENV_INPUT_CPU_INTERVAL : time.Duration
//	ENV_INPUT_CPU_DISABLE_TEMPERATURE_COLLECT : bool
//	ENV_INPUT_CPU_ENABLE_LOAD5S : bool
func (ipt *Input) ReadEnv(envs map[string]string) {
	if percpu, ok := envs["ENV_INPUT_CPU_PERCPU"]; ok {
		b, err := strconv.ParseBool(percpu)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CPU_PERCPU to bool: %s, ignore", err)
		} else {
			ipt.PerCPU = b
		}
	}

	if enableTemperature, ok := envs["ENV_INPUT_CPU_ENABLE_TEMPERATURE"]; ok {
		b, err := strconv.ParseBool(enableTemperature)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CPU_ENABLE_TEMPERATURE to bool: %s, ignore", err)
		} else {
			ipt.EnableTemperature = b
		}
	}

	if tagsStr, ok := envs["ENV_INPUT_CPU_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	//   ENV_INPUT_CPU_INTERVAL : time.Duration
	//   ENV_INPUT_CPU_DISABLE_TEMPERATURE_COLLECT : bool
	//   ENV_INPUT_CPU_ENABLE_LOAD5S : bool
	if str, ok := envs["ENV_INPUT_CPU_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CPU_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}

	if str := envs["ENV_INPUT_CPU_DISABLE_TEMPERATURE_COLLECT"]; str != "" {
		ipt.DisableTemperatureCollect = true
	}

	if str := envs["ENV_INPUT_CPU_ENABLE_LOAD5S"]; str != "" {
		ipt.EnableLoad5s = true
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			ps:                &CPUInfo{},
			Interval:          time.Second * 10,
			EnableTemperature: true,

			semStop: cliutils.NewSem(),
			feeder:  dkio.DefaultFeeder(),
			Tags:    make(map[string]string),
			tagger:  datakit.DefaultGlobalTagger(),
		}
	})
}
