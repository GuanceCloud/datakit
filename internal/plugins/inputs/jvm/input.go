// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package jvm collects JVM metrics.
package jvm

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/influxdata/telegraf/plugins/common/tls"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/jolokia"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	defaultInterval   = time.Second * 60
	MaxGatherInterval = 30 * time.Minute
	MinGatherInterval = 1 * time.Second
	inputName         = "jvm"
)

const (
	confSample = `[[inputs.jvm]]
  # default_tag_prefix      = ""
  # default_field_prefix    = ""
  # default_field_separator = "."

  # username = ""
  # password = ""
  # response_timeout = "5s"

  ## Optional TLS config
  # tls_ca   = "/var/private/ca.pem"
  # tls_cert = "/var/private/client.pem"
  # tls_key  = "/var/private/client-key.pem"
  # insecure_skip_verify = false

  ## Monitor Intreval
  # interval   = "60s"

  # Add agents URLs to query
  urls = ["http://localhost:8080/jolokia"]

  ## v2+ override all measurement names to "jvm", default: v2
  ## If you want to use the old metric set, you can change it to "v1"
  measurement_version = "v2"

  ## Add metrics to read
  [[inputs.jvm.metric]]
    name  = "java_runtime"
    mbean = "java.lang:type=Runtime"
    paths = ["Uptime"]

  [[inputs.jvm.metric]]
    name  = "java_memory"
    mbean = "java.lang:type=Memory"
    paths = ["HeapMemoryUsage", "NonHeapMemoryUsage", "ObjectPendingFinalizationCount"]

  [[inputs.jvm.metric]]
    name     = "java_garbage_collector"
    mbean    = "java.lang:name=*,type=GarbageCollector"
    paths    = ["CollectionTime", "CollectionCount"]
    tag_keys = ["name"]

  [[inputs.jvm.metric]]
    name  = "java_threading"
    mbean = "java.lang:type=Threading"
    paths = ["TotalStartedThreadCount", "ThreadCount", "DaemonThreadCount", "PeakThreadCount"]

  [[inputs.jvm.metric]]
    name  = "java_class_loading"
    mbean = "java.lang:type=ClassLoading"
    paths = ["LoadedClassCount", "UnloadedClassCount", "TotalLoadedClassCount"]

  [[inputs.jvm.metric]]
    name     = "java_memory_pool"
    mbean    = "java.lang:name=*,type=MemoryPool"
    paths    = ["Usage", "PeakUsage", "CollectionUsage"]
    tag_keys = ["name"]

  [inputs.jvm.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...`
)

type Input struct {
	URLs            []string      `toml:"urls"`
	Username        string        `toml:"username"`
	Password        string        `toml:"password"`
	ResponseTimeout time.Duration `toml:"response_timeout"`
	Interval        string        `toml:"interval"`
	Election        bool          `toml:"election"`

	tls.ClientConfig

	Metrics            []MetricConfig `toml:"metric"`
	MeasurementVersion string         `toml:"measurement_version"`

	DefaultTagPrefix      string `toml:"default_tag_prefix"`
	DefaultFieldPrefix    string `toml:"default_field_prefix"`
	DefaultFieldSeparator string `toml:"default_field_separator"`

	Tags map[string]string `toml:"tags"`

	duration time.Duration
	clients  []*jolokia.Client
	Types    map[string]string
	SemStop  *cliutils.Sem
	Feeder   dkio.Feeder
	Tagger   datakit.GlobalTagger
	g        *goroutine.Group

	paused atomic.Bool
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

var (
	jvmTypeMap = map[string]string{
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
	l                      = logger.DefaultSLogger(inputName)
	_ inputs.ElectionInput = (*Input)(nil)
)

func (ipt *Input) Resume() error {
	ipt.paused.Store(false)
	return nil
}

func (ipt *Input) Pause() error {
	ipt.paused.Store(true)
	return nil
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
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
	l.Debugf("jvm urls:%v", ipt.URLs)

	for {
		if ipt.paused.Load() {
			l.Debugf("JVM plugin %s paused", inputName)
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
		ipt.g = goroutine.NewGroup(goroutine.Option{Name: "jvm_collectors"})
	}

	// Initialize clients
	if err := ipt.initClients(); err != nil {
		return fmt.Errorf("initClients failed: %w", err)
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
	ipt.Types = jvmTypeMap

	return nil
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

func (ipt *Input) Terminate() {
	if ipt.SemStop != nil {
		ipt.SemStop.Close()
	}
}

func (ipt *Input) exit() {
	// JVM doesn't have tailer or other resources to clean up
}

func (ipt *Input) Catalog() string      { return inputName }
func (ipt *Input) SampleConfig() string { return confSample }
func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&jvmMeasurement{}, // Unified v2 measurement
		&inputs.UpMeasurement{},
	}
}

func (ipt *Input) AvailableArchs() []string {
	return datakit.AllOS
}

func defaultInput() *Input {
	return &Input{
		SemStop:            cliutils.NewSem(),
		paused:             atomic.Bool{},
		Election:           true,
		Tagger:             datakit.DefaultGlobalTagger(),
		Feeder:             dkio.DefaultFeeder(),
		Types:              jvmTypeMap,
		ResponseTimeout:    5 * time.Second,
		MeasurementVersion: "v2",
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
