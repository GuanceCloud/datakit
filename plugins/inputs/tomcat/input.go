package tomcat

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName   = "tomcat"
	minInterval = time.Second * 10
	maxInterval = time.Minute * 20
)

type Input struct {
	inputs.JolokiaAgent
	Log  *inputs.TailerOption `toml:"log"`
	Tags map[string]string    `toml:"tags"`
}

func (i *Input) Catalog() string {
	return inputName
}

func (i *Input) SampleConfig() string {
	return tomcatSampleCfg
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&TomcatGlobalRequestProcessorM{},
		&TomcatJspMonitorM{},
		&TomcatThreadPoolM{},
		&TomcatServletM{},
		&TomcatCacheM{},
	}
}
func (i *Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName: pipelineCfg,
	}
	return pipelineMap
}

func (i *Input) Run() {
	if d, err := time.ParseDuration(i.Interval); err != nil {
		i.Interval = (time.Second * 10).String()
	} else {
		i.Interval = config.ProtectedInterval(minInterval, maxInterval, d).String()
	}
	i.PluginName = inputName
	i.JolokiaAgent.Tags = i.Tags
	i.JolokiaAgent.Types = TomcatMetricType
	if i.Log != nil && len(i.Log.Files) > 0 {
		go i.gatherLog()
	}
	i.JolokiaAgent.Collect()
}

func (i *Input) gatherLog() {
	inputs.JoinPipelinePath(i.Log, inputName+".p")
	i.Log.Source = inputName
	i.Log.Tags = map[string]string{}
	for k, v := range i.Tags {
		i.Log.Tags[k] = v
	}
	tail, err := inputs.NewTailer(i.Log)
	if err != nil {
		return
	}
	defer tail.Close()
	tail.Run()
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		i := &Input{}
		return i
	})
}
