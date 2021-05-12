package jvm

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = "jvm"
)

const (
	defaultInterval   = "60s"
	MaxGatherInterval = 30 * time.Minute
	MinGatherInterval = 1 * time.Second
)

var JvmTypeMap = map[string]string{
	"Uptime":                         "int",
	"HeapMemoryUsageinit":            "int",
	"HeapMemoryUsageused":            "int",
	"HeapMemoryUsagemax":             "int",
	"HeapMemoryUsagecommitted":       "int",
	"NonHeapMemoryUsageinit":         "int",
	"NonHeapMemoryUsageused":         "int",
	"NonHeapMemoryUsagemax":          "int",
	"NonHeapMemoryUsagecommitted":    "int",
	"ObjectPendingFinalizationCount": "int",
	"CollectionTime":                 "int",
	"CollectionCount":                "int",
	"DaemonThreadCount":              "int",
	"PeakThreadCount":                "int",
	"ThreadCount":                    "int",
	"TotalStartedThreadCount":        "int",
	"LoadedClassCount":               "int",
	"TotalLoadedClassCount":          "int",
	"UnloadedClassCount":             "int",
	"Usageinit":                      "int",
	"Usagemax":                       "int",
	"Usagecommitted":                 "int",
	"Usageused":                      "int",
	"PeakUsageinit":                  "int",
	"PeakUsagemax":                   "int",
	"PeakUsagecommitted":             "int",
	"PeakUsageused":                  "int",
}

type Input struct {
	JolokiaAgent
	Tags map[string]string
}

func (i *Input) Run() {
	if i.Interval == "" {
		i.Interval = defaultInterval
	}

	i.PluginName = inputName

	i.JolokiaAgent.Tags = i.Tags
	i.JolokiaAgent.Types = JvmTypeMap

	i.JolokiaAgent.Collect()
}

func (i *Input) Catalog() string      { return inputName }
func (i *Input) SampleConfig() string { return JvmConfigSample }
func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&JavaRuntimeMemt{},
		&JavaMemoryMemt{},
		&JavaGcMemt{},
		//&JavaLastGcMemt{},
		&JavaThreadMemt{},
		&JavaClassLoadMemt{},
		&JavaMemoryPoolMemt{},
	}
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}

func (j *JolokiaAgent) Collect() {
	j.l = logger.DefaultSLogger(j.PluginName)
	j.l.Infof("%s input started...", j.PluginName)

	duration, err := time.ParseDuration(j.Interval)
	if err != nil {
		j.l.Error(err)
		return
	}
	duration = datakit.ProtectedInterval(MinGatherInterval, MaxGatherInterval, duration)
	tick := time.NewTicker(duration)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			start := time.Now()
			if err := j.Gather(); err != nil {
				io.FeedLastError(j.PluginName, err.Error())
				j.l.Error(err)
			} else {
				inputs.FeedMeasurement(j.PluginName, datakit.Metric, j.collectCache,
					&io.Option{CollectCost: time.Since(start), HighFreq: false})

				j.collectCache = j.collectCache[:] // NOTE: do not forget to clean cache
			}

		case <-datakit.Exit.Wait():
			j.l.Infof("input %s exit", j.PluginName)
			return
		}
	}
}
