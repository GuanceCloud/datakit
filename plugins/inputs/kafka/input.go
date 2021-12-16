// Package kafka collect kafka metrics
package kafka

import (
	"os"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	defaultInterval = "60s"
	inputName       = "kafka"
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	inputs.JolokiaAgent
	Log  *kafkalog         `toml:"log"`
	Tags map[string]string `toml:"tags"`

	tail *tailer.Tailer
}

type kafkalog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	MultilineMatch    string   `toml:"multiline_match"`
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)

	if i.Interval == "" { //nolint:typecheck
		i.Interval = defaultInterval
	}

	i.JolokiaAgent.L = l
	i.PluginName = inputName
	i.JolokiaAgent.Tags = i.Tags
	i.JolokiaAgent.Types = KafkaTypeMap

	i.JolokiaAgent.Collect()
}

func (i *Input) RunPipeline() {
	if i.Log == nil || len(i.Log.Files) == 0 {
		return
	}

	if i.Log.Pipeline == "" {
		i.Log.Pipeline = inputName + ".p" // use default
	}

	opt := &tailer.Option{
		Source:            inputName,
		Service:           inputName,
		GlobalTags:        i.Tags,
		IgnoreStatus:      i.Log.IgnoreStatus,
		CharacterEncoding: i.Log.CharacterEncoding,
		MultilineMatch:    i.Log.MultilineMatch,
	}

	pl, err := config.GetPipelinePath(i.Log.Pipeline)
	if err != nil {
		l.Error(err)
		io.FeedLastError(inputName, err.Error())
		return
	}
	if _, err := os.Stat(pl); err != nil {
		l.Warn("%s missing: %s", pl, err.Error())
	} else {
		opt.Pipeline = pl
	}

	i.tail, err = tailer.NewTailer(i.Log.Files, opt)
	if err != nil {
		l.Errorf("NewTailer: %s", err)

		io.FeedLastError(inputName, err.Error())
		return
	}

	go i.tail.Start()
}

func (*Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName: pipelineCfg,
	}
	return pipelineMap
}

func (i *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if i.Log != nil {
					return i.Log.Pipeline
				}
				return ""
			}(),
		},
	}
}

func (*Input) Catalog() string      { return "db" }
func (*Input) SampleConfig() string { return kafkaConfSample }

func (*Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&KafkaControllerMment{},
		&KafkaReplicaMment{},
		&KafkaPurgatoryMment{},
		&KafkaRequestMment{},
		&KafkaTopicsMment{},
		&KafkaTopicMment{},
		&KafkaPartitionMment{},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
