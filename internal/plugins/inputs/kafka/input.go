// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kafka collect kafka metrics
package kafka

import (
	"context"

	"github.com/GuanceCloud/cliutils/logger"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/jolokia"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

const (
	defaultInterval = "60s"
	inputName       = "kafka"
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	jolokia.JolokiaAgent
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

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	if ipt.Interval == "" { //nolint:typecheck
		ipt.JolokiaAgent.Interval = defaultInterval
	}

	ipt.JolokiaAgent.L = l
	ipt.JolokiaAgent.PluginName = inputName
	ipt.JolokiaAgent.Tags = ipt.Tags
	ipt.JolokiaAgent.Types = KafkaTypeMap
	l.Debugf("kafka url:%s", ipt.JolokiaAgent.URLs)
	ipt.JolokiaAgent.Collect()
}

func (ipt *Input) Terminate() {
	if ipt.tail != nil {
		ipt.tail.Close()
	}
	ipt.JolokiaAgent.Terminate()
}

func (ipt *Input) RunPipeline() {
	if ipt.Log == nil || len(ipt.Log.Files) == 0 {
		return
	}

	opts := []tailer.Option{
		tailer.WithSource(inputName),
		tailer.WithService(inputName),
		tailer.WithPipeline(ipt.Log.Pipeline),
		tailer.WithIgnoredStatuses(ipt.Log.IgnoreStatus),
		tailer.WithCharacterEncoding(ipt.Log.CharacterEncoding),
		tailer.EnableMultiline(true),
		tailer.WithMaxMultilineLength(int64(float64(config.Cfg.Dataway.MaxRawBodySize) * 0.8)),
		tailer.WithMultilinePatterns([]string{ipt.Log.MultilineMatch}),
		tailer.WithExtraTags(inputs.MergeTags(ipt.JolokiaAgent.Tagger.HostTags(), ipt.Tags, "")),
		tailer.EnableDebugFields(config.Cfg.EnableDebugFields),
	}

	var err error
	ipt.tail, err = tailer.NewTailer(ipt.Log.Files, opts...)
	if err != nil {
		l.Errorf("NewTailer: %s", err)
		metrics.FeedLastError(inputName, err.Error())
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_kafka"})
	g.Go(func(ctx context.Context) error {
		ipt.tail.Start()
		return nil
	})
}

func (*Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName: pipelineCfg,
	}
	return pipelineMap
}

//nolint:lll
func (ipt *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		inputName: {
			"Kafka log": `[2020-07-07 15:04:29,333] DEBUG Progress event: HTTP_REQUEST_COMPLETED_EVENT, bytes: 0 (io.confluent.connect.s3.storage.S3OutputStream:286)`,
		},
	}
}

func (ipt *Input) GetPipeline() []tailer.Option {
	opts := []tailer.Option{
		tailer.WithSource(inputName),
		tailer.WithService(inputName),
	}
	if ipt.Log != nil {
		opts = append(opts, tailer.WithPipeline(ipt.Log.Pipeline))
	}
	return opts
}

func (*Input) Catalog() string      { return "db" }
func (*Input) SampleConfig() string { return kafkaConfSample }

func (*Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
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
		&KafkaZooKeeperMment{},
		&KafkaNetworkMment{},
		&KafkaLogMment{},
		&KafkaConsumerMment{},
		&KafkaProducerMment{},
		&KafkaConnectMment{},
	}
}

func defaultInput() *Input {
	j := jolokia.DefaultInput()
	j.PluginName = inputName

	return &Input{
		JolokiaAgent: *j,
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
