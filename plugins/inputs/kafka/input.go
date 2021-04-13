package kafka

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/jvm"
)

var (
	inputName = "kafka"
)

const (
	defaultInterval = "60s"
)

type Input struct {
	jvm.JolokiaAgent
}

func (i *Input) Run() {
	if i.Interval == "" {
		i.Interval = defaultInterval
	}

	i.PluginName = inputName

	i.JolokiaAgent.Collect()
}

func (i *Input) Catalog() string      { return inputName }
func (i *Input) SampleConfig() string { return kafkaConfSample }

//func (i *Input) SampleMeasurement() []inputs.Measurement {
//	return []inputs.Measurement{
//		&JvmMeasurement{},
//	}
//}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
