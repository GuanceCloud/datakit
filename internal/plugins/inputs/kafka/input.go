// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kafka collect kafka metrics
package kafka

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/influxdata/telegraf/plugins/common/tls"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/jolokia"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

const (
	defaultInterval     = time.Second * 60
	inputName           = "kafka"
	defaultBatchSize    = 1024
	defaultMBeanPattern = "kafka.*:*"
	kafkaAppInfoMBean   = "kafka.server:type=app-info"
)

var (
	l                      = logger.DefaultSLogger(inputName)
	_ inputs.ElectionInput = (*Input)(nil)
	_ inputs.InputV2       = (*Input)(nil)
)

type Input struct {
	URLs            []string      `toml:"urls"`
	Username        string        `toml:"username"`
	Password        string        `toml:"password"`
	ResponseTimeout time.Duration `toml:"response_timeout"`
	Interval        string        `toml:"interval"`
	Election        bool          `toml:"election"`

	tls.ClientConfig

	Metrics           []MetricConfig `toml:"metric"`
	EnableAutoCollect bool           `toml:"enable_auto_collect"`
	MBeanBlacklist    []string       `toml:"mbean_blacklist"`

	DefaultTagPrefix      string   `toml:"default_tag_prefix"`
	DefaultFieldPrefix    string   `toml:"default_field_prefix"`
	DefaultFieldSeparator string   `toml:"default_field_separator"`
	TagBlacklist          []string `toml:"tag_blacklist"`
	TagBlackMap           map[string]struct{}

	Log  *kafkalog         `toml:"log"`
	Tags map[string]string `toml:"tags"`

	duration time.Duration
	clients  []*jolokia.Client
	Types    map[string]string
	SemStop  *cliutils.Sem
	Feeder   dkio.Feeder
	Tagger   datakit.GlobalTagger
	g        *goroutine.Group
	pause    atomic.Bool
	tail     *tailer.Tailer
}

type MetricConfig struct {
	Name           string   `toml:"name"`
	Mbean          string   `toml:"mbean"`
	Paths          []string `toml:"paths"`
	FieldName      *string  `toml:"field_name"`
	FieldPrefix    *string  `toml:"field_prefix"`
	FieldSeparator *string  `toml:"field_separator"`
	TagPrefix      *string  `toml:"tag_prefix"`
	TagKeys        []string `toml:"tag_keys"`
}

type kafkalog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	MultilineMatch    string   `toml:"multiline_match"`
}

func (ipt *Input) Run() {
	if err := ipt.setup(); err != nil {
		l.Errorf("setup failed: %v", err)
		return
	}
	duration := ipt.duration

	tick := time.NewTicker(duration)
	defer tick.Stop()
	start := ntp.Now()

	l.Infof("%s input started...", inputName)
	l.Debugf("kafka urls:%v", ipt.URLs)

	for {
		if ipt.pause.Load() {
			l.Debugf("Kafka plugin %s paused", inputName)
		} else {
			if err := ipt.collect(start.UnixNano()); err != nil {
				ipt.Feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(inputName),
					metrics.WithLastErrorCategory(point.Metric),
				)
			}
		}

		select {
		case tt := <-tick.C:
			start = inputs.AlignTime(tt, start, duration)

		case <-datakit.Exit.Wait():
			l.Infof("input %s exit", inputName)
			ipt.exit()
			return

		case <-ipt.SemStop.Wait():
			l.Infof("input %s return", inputName)
			ipt.exit()
			return
		}
	}
}

func (ipt *Input) setup() error {
	l = logger.SLogger(inputName)

	// Adapt metrics: replace # with $ in FieldPrefix, FieldSeparator, and FieldName
	ipt.adaptor()

	if ipt.g == nil {
		ipt.g = goroutine.NewGroup(goroutine.Option{Name: "kafka_collectors"})
	}

	// Initialize clients
	if err := ipt.initClients(); err != nil {
		return fmt.Errorf("initClients failed: %w", err)
	}

	if ipt.TagBlackMap == nil {
		ipt.TagBlackMap = make(map[string]struct{})
	}
	for _, tag := range ipt.TagBlacklist {
		ipt.TagBlackMap[tag] = struct{}{}
	}

	var duration time.Duration
	var err error
	if len(ipt.Interval) > 0 {
		duration, err = time.ParseDuration(ipt.Interval)
		if err != nil {
			return fmt.Errorf("time.ParseDuration: %w", err)
		}
	} else {
		duration = defaultInterval
	}

	ipt.duration = duration

	return nil
}

func (ipt *Input) Terminate() {
	if ipt.SemStop != nil {
		ipt.SemStop.Close()
	}
}

func (ipt *Input) exit() {
	if ipt.tail != nil {
		ipt.tail.Close()
	}
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
		tailer.WithExtraTags(inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, "")),
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

func (ipt *Input) Pause() error {
	ipt.pause.Store(true)
	return nil
}

func (ipt *Input) Resume() error {
	ipt.pause.Store(false)
	return nil
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
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
		&kafkaMeasurement{}, // Auto collect mode measurement
		&inputs.UpMeasurement{},
	}
}

func defaultInput() *Input {
	return &Input{
		SemStop:           cliutils.NewSem(),
		pause:             atomic.Bool{},
		Election:          true,
		Tagger:            datakit.DefaultGlobalTagger(),
		Feeder:            dkio.DefaultFeeder(),
		Types:             KafkaTypeMap,
		ResponseTimeout:   5 * time.Second,
		TagBlacklist:      []string{"domain", "type", "name"},
		TagBlackMap:       make(map[string]struct{}),
		EnableAutoCollect: true,
	}
}

func (ipt *Input) adaptor() {
	for i, m := range ipt.Metrics {
		if m.FieldPrefix != nil {
			t := strings.ReplaceAll(*m.FieldPrefix, "#", "$")
			m.FieldPrefix = &t
		}

		if m.FieldSeparator != nil {
			t := strings.ReplaceAll(*m.FieldSeparator, "#", "$")
			m.FieldSeparator = &t
		}

		if m.FieldName != nil {
			t := strings.ReplaceAll(*m.FieldName, "#", "$")
			m.FieldName = &t
		}

		ipt.Metrics[i] = m
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}

func (ipt *Input) initClients() error {
	ipt.clients = make([]*jolokia.Client, 0, len(ipt.URLs))
	for _, url := range ipt.URLs {
		config := jolokia.Config{
			URL:             url,
			Username:        ipt.Username,
			Password:        ipt.Password,
			ResponseTimeout: ipt.ResponseTimeout,
			TLS:             ipt.ClientConfig,
			Input:           inputName,
		}

		client, err := jolokia.NewClient(config)
		if err != nil {
			return fmt.Errorf("create client for %s: %w", url, err)
		}

		ipt.clients = append(ipt.clients, client)
	}
	return nil
}
