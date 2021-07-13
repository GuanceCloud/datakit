package kafka

import (
	"os"
	"path/filepath"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	defaultInterval = "60s"
	inputName       = "kafka"
)

var (
	l = logger.DefaultSLogger(inputName)
)

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
	Match             string   `toml:"match"`
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)

	if i.Interval == "" {
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
		Match:             i.Log.Match,
	}

	pl := filepath.Join(datakit.PipelineDir, i.Log.Pipeline)
	if _, err := os.Stat(pl); err != nil {
		l.Warn("%s missing: %s", pl, err.Error())
	} else {
		opt.Pipeline = pl
	}

	var err error
	i.tail, err = tailer.NewTailer(i.Log.Files, opt)
	if err != nil {
		l.Error(err)
		io.FeedLastError(inputName, err.Error())
		return
	}

	go i.tail.Start()
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
