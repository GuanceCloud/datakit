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
				inputs.FeedMeasurement(j.PluginName, io.Metric, j.collectCache,
					&io.Option{CollectCost: time.Since(start), HighFreq: false})

				j.collectCache = j.collectCache[:] // NOTE: do not forget to clean cache
			}

		case <-datakit.Exit.Wait():
			j.l.Infof("input %s exit", j.PluginName)
			return
		}
	}
}
