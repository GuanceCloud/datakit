// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package jvm collects JVM metrics.
package jvm

import (
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/jolokia"
)

const (
	defaultInterval   = "60s"
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

var jvmTypeMap = map[string]string{
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

type Input struct {
	jolokia.JolokiaAgent
	Tags map[string]string `toml:"tags"`
}

var log = logger.DefaultSLogger(inputName)

func (ipt *Input) Run() {
	log = logger.SLogger(inputName)
	if ipt.JolokiaAgent.Interval == "" {
		ipt.JolokiaAgent.Interval = defaultInterval
	}

	ipt.JolokiaAgent.PluginName = inputName

	ipt.JolokiaAgent.Tags = ipt.Tags
	ipt.JolokiaAgent.Types = jvmTypeMap
	ipt.JolokiaAgent.L = log
	ipt.JolokiaAgent.Collect()
}

func (ipt *Input) Catalog() string      { return inputName }
func (ipt *Input) SampleConfig() string { return confSample }
func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&JavaRuntimeMemt{},
		&JavaMemoryMemt{},
		&JavaGcMemt{},
		&JavaThreadMemt{},
		&JavaClassLoadMemt{},
		&JavaMemoryPoolMemt{},
	}
}

func (ipt *Input) AvailableArchs() []string {
	return datakit.AllOS
}

func defaultInput() *Input {
	return &Input{
		JolokiaAgent: *jolokia.DefaultInput(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
