// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cpu collect CPU metrics.
package cpu

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"go.uber.org/atomic"
)

var (
	_ inputs.ReadEnv = (*Input)(nil)
	l                = logger.DefaultSLogger(inputName)
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
)

const (
	inputName  = "cpu"
	metricName = inputName
)

type Input struct {
	TotalCPU                  bool `toml:"totalcpu"`                              // deprecated
	CollectCPUTime            bool `toml:"collect_cpu_time"`                      // deprecated
	ReportActive              bool `toml:"report_active"`                         // deprecated
	DisableTemperatureCollect bool `toml:"disable_temperature_collect,omitempty"` // deprecated

	PerCPU            bool `toml:"percpu"`
	EnableTemperature bool `toml:"enable_temperature"`
	EnableLoad5s      bool `toml:"enable_load5s"`

	Interval datakit.Duration
	Tags     map[string]string

	collectCache         []inputs.Measurement
	collectCacheLast1Ptr *cpuMeasurement

	lastStats map[string]cpu.TimesStat
	load5s    atomic.Int32
	lastLoad1 float64
	ps        CPUStatInfo

	semStop *cliutils.Sem // start stop signal
}

func (ipt *Input) appendMeasurement(name string, tags map[string]string, fields map[string]interface{}, ts time.Time) {
	tmp := &cpuMeasurement{name: name, tags: tags, fields: fields, ts: ts}
	ipt.collectCache = append(ipt.collectCache, tmp)
	ipt.collectCacheLast1Ptr = tmp
}

func (*Input) Catalog() string {
	return "host"
}

func (*Input) SampleConfig() string {
	return sampleCfg
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&cpuMeasurement{},
	}
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (ipt *Input) Collect() error {
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

	now := time.Now()
	for _, cts := range cpuTimes {
		tags := map[string]string{
			"cpu": cts.CPU,
		}
		for k, v := range ipt.Tags {
			tags[k] = v
		}

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

		fields := map[string]interface{}{
			"usage_user":       cpuUsage.User,
			"usage_system":     cpuUsage.System,
			"usage_idle":       cpuUsage.Idle,
			"usage_nice":       cpuUsage.Nice,
			"usage_iowait":     cpuUsage.Iowait,
			"usage_irq":        cpuUsage.Irq,
			"usage_softirq":    cpuUsage.Softirq,
			"usage_steal":      cpuUsage.Steal,
			"usage_guest":      cpuUsage.Guest,
			"usage_guest_nice": cpuUsage.GuestNice,
			"usage_total":      cpuUsage.Total,
		}

		if ipt.EnableLoad5s {
			fields["load5s"] = ipt.getLoad5s()
		}

		if len(coreTemp) > 0 && cts.CPU == "cpu-total" {
			if v, ok := coreTemp[cts.CPU]; ok {
				fields["core_temperature"] = v
			}
		}
		ipt.appendMeasurement(inputName, tags, fields, now)
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

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("cpu input started")

	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)

	if ipt.EnableLoad5s {
		go ipt.calLoad5s()
	}

	if err := ipt.Collect(); err != nil { // gather lastSats
		l.Errorf("Collect: %s", err.Error())
	}

	time.Sleep(ipt.Interval.Duration)

	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	for {
		start := time.Now()
		if err := ipt.Collect(); err != nil {
			l.Errorf("Collect: %s", err)
			io.FeedLastError(inputName, err.Error())
		}

		if len(ipt.collectCache) > 0 {
			if err := inputs.FeedMeasurement(metricName, datakit.Metric, ipt.collectCache,
				&io.Option{CollectCost: time.Since((start))}); err != nil {
				l.Errorf("FeedMeasurement: %s", err)
			}
		}
		ipt.collectCache = make([]inputs.Measurement, 0)

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Infof("cpu input exit")
			return

		case <-ipt.semStop.Wait():
			l.Info("cpu input return")
			return
		}
	}
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

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

// ReadEnv support envs：
//   ENV_INPUT_CPU_PERCPU : booler
//   ENV_INPUT_CPU_ENABLE_TEMPERATURE : booler
//   ENV_INPUT_CPU_INTERVAL : datakit.Duration
//   ENV_INPUT_CPU_DISABLE_TEMPERATURE_COLLECT : bool
//   ENV_INPUT_CPU_ENABLE_LOAD5S : bool
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

	//   ENV_INPUT_CPU_INTERVAL : datakit.Duration
	//   ENV_INPUT_CPU_DISABLE_TEMPERATURE_COLLECT : bool
	//   ENV_INPUT_CPU_ENABLE_LOAD5S : bool
	if str, ok := envs["ENV_INPUT_CPU_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CPU_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval.Duration = config.ProtectedInterval(minInterval,
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
			Interval:          datakit.Duration{Duration: time.Second * 10},
			EnableTemperature: true,

			semStop: cliutils.NewSem(),
			Tags:    make(map[string]string),
		}
	})
}
