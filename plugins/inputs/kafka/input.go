package kafka

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = "kafka"
)

const (
	defaultInterval = "60s"
)

var (
	l *logger.Logger
)

type Input struct {
	inputs.JolokiaAgent
	Log  *inputs.TailerOption `toml:"log"`
	Tags map[string]string    `toml:"tags"`
}

func (i *Input) Run() {
	if i.Interval == "" {
		i.Interval = defaultInterval
	}

	l = logger.SLogger(inputName)

	i.PluginName = inputName
	i.JolokiaAgent.Tags = i.Tags
	i.JolokiaAgent.Types = KafkaTypeMap

	if i.Log != nil {
		go i.runLog()
	}
	i.JolokiaAgent.Collect()
}

func (i *Input) runLog() {
	inputs.JoinPipelinePath(i.Log, "kafka.p")
	i.Log.Source = "kafka"
	i.Log.Tags = make(map[string]string)
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

func (_ *Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName: pipelineCfg,
	}
	return pipelineMap
}

func (i *Input) Catalog() string      { return inputName }
func (i *Input) SampleConfig() string { return kafkaConfSample }

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&KafkaControllerMment{},
		&KafkaReplicaMment{},
		&KafkaPurgatoryMment{},
		&KafkaClientMment{},
		&KafkaRequestMment{},
		&KafkaTopicsMment{},
		&KafkaTopicMment{},
		&KafkaPartitionMment{},
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
