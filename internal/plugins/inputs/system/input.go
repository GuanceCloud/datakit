// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package system collect system level metrics
package system

import (
	"errors"
	"runtime"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/process"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	conntrackutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hostutil/conntrack"
	filefdutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hostutil/filefd"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/mem"
)

var (
	_ inputs.ReadEnv   = (*Input)(nil)
	_ inputs.Singleton = (*Input)(nil)
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
)

const (
	inputName           = "system"
	metricNameSystem    = "system"
	metricNameConntrack = "conntrack"
	metricNameFilefd    = "filefd"
	sampleCfg           = `
[[inputs.system]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

  [inputs.system.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	Interval  datakit.Duration
	Fielddrop []string // Deprecated
	Tags      map[string]string

	collectCache []*point.Point

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger
	ptsTime time.Time
}

func (ipt *Input) Singleton() {
}

func (*Input) Catalog() string {
	return "host"
}

func (*Input) SampleConfig() string {
	return sampleCfg
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&systemMeasurement{},
		&conntrackMeasurement{},
		&filefdMeasurement{},
	}
}

func (ipt *Input) Collect() error {
	opts := append(point.DefaultMetricOptions(), point.WithTime(ipt.ptsTime))

	loadAvg, err := load.Avg()
	if err != nil && !strings.Contains(err.Error(), "not implemented") {
		return err
	}

	tags := map[string]string{}
	for k, v := range ipt.Tags {
		tags[k] = v
	}

	tags = inputs.MergeTags(ipt.Tagger.HostTags(), tags, "")

	if runtime.GOOS == "linux" {
		conntrackStat := conntrackutil.GetConntrackInfo()

		var kvs point.KVs
		kvs = kvs.Set("entries", conntrackStat.Current).
			Set("entries_limit", conntrackStat.Limit).
			Set("stat_found", conntrackStat.Found).
			Set("stat_invalid", conntrackStat.Invalid).
			Set("stat_ignore", conntrackStat.Ignore).
			Set("stat_insert", conntrackStat.Insert).
			Set("stat_insert_failed", conntrackStat.InsertFailed).
			Set("stat_drop", conntrackStat.Drop).
			Set("stat_early_drop", conntrackStat.EarlyDrop).
			Set("stat_search_restart", conntrackStat.SearchRestart)

		for k, v := range tags {
			kvs = kvs.AddTag(k, v)
		}

		ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameConntrack, kvs, opts...))

		filefdStat, err := filefdutil.GetFileFdInfo()
		if err != nil {
			l.Warnf("filefdutil.GetFileFdInfo(): %s, ignored", err.Error())
		} else {
			var kvs point.KVs
			kvs = kvs.Set("allocated", filefdStat.Allocated).
				Set("maximum_mega", filefdStat.MaximumMega)
			for k, v := range tags {
				kvs = kvs.AddTag(k, v)
			}

			ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameFilefd, kvs, opts...))
		}
	}

	var kvs point.KVs
	cpuTotal, err := cpu.Percent(0, false)
	if err != nil {
		l.Warnf("CPU stat error: %s, ignored", err.Error())
	} else if len(cpuTotal) > 0 {
		kvs = kvs.Set("cpu_total_usage", cpuTotal[0])
	}

	if vm, err := mem.VirtualMemoryStat(); err != nil {
		l.Warnf("error getting virtual memory info: %w", err)
	} else {
		if vm == nil {
			return errors.New("get virtual memory stat fail")
		}
		kvs = kvs.Set("memory_usage", vm.UsedPercent)
	}

	if pids, err := process.Pids(); err != nil {
		l.Warnf("error getting Pids: %w", err)
	} else {
		kvs = kvs.Set("process_count", len(pids))
	}

	numCPUs, err := cpu.Counts(true)
	if err != nil {
		return err
	}

	kvs = kvs.Set("load1_per_core", loadAvg.Load1/float64(numCPUs)).
		Set("load1", loadAvg.Load1).
		Set("load15_per_core", loadAvg.Load15/float64(numCPUs)).
		Set("load15", loadAvg.Load15).
		Set("load5_per_core", loadAvg.Load5/float64(numCPUs)).
		Set("load5", loadAvg.Load5).
		Set("n_cpus", numCPUs)

	if users, err := host.Users(); err != nil {
		l.Warnf("Users: %s, ignored", err.Error())
	} else {
		kvs = kvs.Set("n_users", len(users))
	}

	uptime, err := host.Uptime()
	if err != nil {
		l.Warnf("Uptime: %s, ignored", err.Error())
	} else {
		kvs = kvs.Set("uptime", uptime)
	}
	ipt.collectCache = append(ipt.collectCache, point.NewPoint(metricNameSystem, kvs, opts...))

	return nil
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("system input started")
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	ipt.ptsTime = ntp.Now()
	for {
		collectStart := time.Now()
		if err := ipt.Collect(); err != nil {
			l.Errorf("Collect: %s", err)
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
			)
		}

		if len(ipt.collectCache) > 0 {
			if err := ipt.feeder.Feed(point.Metric, ipt.collectCache,
				dkio.WithCollectCost(time.Since(collectStart)),
				dkio.WithSource(inputName),
			); err != nil {
				l.Errorf("Feed failed: %v", err)
			}

			ipt.collectCache = ipt.collectCache[:0]
		}

		select {
		case tt := <-tick.C:
			ipt.ptsTime = inputs.AlignTime(tt, ipt.ptsTime, ipt.Interval.Duration)
		case <-datakit.Exit.Wait():
			l.Info("system input exit")
			return

		case <-ipt.semStop.Wait():
			l.Info("system input return")
			return
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "Interval"},
		{FieldName: "Tags"},
	}

	return doc.SetENVDoc("ENV_INPUT_SYSTEM_", infos)
}

// ReadEnv support envs：
//
//	ENV_INPUT_SYSTEM_TAGS : "a=b,c=d"
//	ENV_INPUT_SYSTEM_INTERVAL : datakit.Duration
func (ipt *Input) ReadEnv(envs map[string]string) {
	if tagsStr, ok := envs["ENV_INPUT_SYSTEM_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	//   ENV_INPUT_SYSTEM_INTERVAL : datakit.Duration
	if str, ok := envs["ENV_INPUT_SYSTEM_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_SYSTEM_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval.Duration = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}
}

func defaultInput() *Input {
	return &Input{
		Interval: datakit.Duration{Duration: time.Second * 10},

		semStop: cliutils.NewSem(),
		Tags:    make(map[string]string),
		feeder:  dkio.DefaultFeeder(),
		Tagger:  datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
