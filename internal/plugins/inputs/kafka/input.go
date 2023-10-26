// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kafka collect kafka metrics
package kafka

import (
	"context"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
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

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	if ipt.Interval == "" { //nolint:typecheck
		ipt.JolokiaAgent.Interval = defaultInterval
	}

	ipt.JolokiaAgent.L = l
	ipt.JolokiaAgent.PluginName = inputName
	ipt.JolokiaAgent.Tags = ipt.Tags
	ipt.JolokiaAgent.Types = KafkaTypeMap

	ipt.JolokiaAgent.Collect()
}

func (ipt *Input) RunPipeline() {
	if ipt.Log == nil || len(ipt.Log.Files) == 0 {
		return
	}

	opt := &tailer.Option{
		Source:            inputName,
		Service:           inputName,
		Pipeline:          ipt.Log.Pipeline,
		GlobalTags:        ipt.Tags,
		IgnoreStatus:      ipt.Log.IgnoreStatus,
		CharacterEncoding: ipt.Log.CharacterEncoding,
		MultilinePatterns: []string{ipt.Log.MultilineMatch},
		Done:              ipt.SemStop.Wait(), // nolint:typecheck
	}

	var err error
	ipt.tail, err = tailer.NewTailer(ipt.Log.Files, opt)
	if err != nil {
		l.Errorf("NewTailer: %s", err)
		io.FeedLastError(inputName, err.Error())
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

func (ipt *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if ipt.Log != nil {
					return ipt.Log.Pipeline
				}
				return ""
			}(),
		},
	}
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

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			JolokiaAgent: inputs.JolokiaAgent{
				SemStop: cliutils.NewSem(),
			},
		}
	})
}
