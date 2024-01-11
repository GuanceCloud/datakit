// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package container collect container metrics/loggings/objects.
package container

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

var (
	_ inputs.ReadEnv   = (*Input)(nil)
	_ inputs.Singleton = (*Input)(nil)
)

type Input struct {
	Endpoints                   []string `toml:"endpoints"`
	DeprecatedDockerEndpoint    string   `toml:"docker_endpoint"`
	DeprecatedContainerdAddress string   `toml:"containerd_address"`

	EnableContainerMetric                 bool     `toml:"enable_container_metric"`
	EnableK8sMetric                       bool     `toml:"enable_k8s_metric"`
	EnablePodMetric                       bool     `toml:"enable_pod_metric"`
	EnableK8sEvent                        bool     `toml:"enable_k8s_event"`
	EnableK8sNodeLocal                    bool     `toml:"enable_k8s_node_local"`
	DeprecatedEnableExtractK8sLabelAsTags bool     `toml:"extract_k8s_label_as_tags"`
	ExtractK8sLabelAsTagsV2               []string `toml:"extract_k8s_label_as_tags_v2"`
	ExtractK8sLabelAsTagsV2ForMetric      []string `toml:"extract_k8s_label_as_tags_v2_for_metric"`
	Election                              bool     `toml:"election"`

	K8sURL                                            string `toml:"kubernetes_url"`
	K8sBearerToken                                    string `toml:"bearer_token"`
	K8sBearerTokenString                              string `toml:"bearer_token_string"`
	EnableAutoDiscoveryOfPrometheusPodAnnotations     bool   `toml:"enable_auto_discovery_of_prometheus_pod_annotations"`
	EnableAutoDiscoveryOfPrometheusServiceAnnotations bool   `toml:"enable_auto_discovery_of_prometheus_service_annotations"`
	EnableAutoDiscoveryOfPrometheusPodMonitors        bool   `toml:"enable_auto_discovery_of_prometheus_pod_monitors"`
	EnableAutoDiscoveryOfPrometheusServiceMonitors    bool   `toml:"enable_auto_discovery_of_prometheus_service_monitors"`
	autoDiscoveryOfPromStreamSize                     int

	ContainerIncludeLog                   []string          `toml:"container_include_log"`
	ContainerExcludeLog                   []string          `toml:"container_exclude_log"`
	LoggingExtraSourceMap                 map[string]string `toml:"logging_extra_source_map"`
	LoggingSourceMultilineMap             map[string]string `toml:"logging_source_multiline_map"`
	LoggingAutoMultilineDetection         bool              `toml:"logging_auto_multiline_detection"`
	LoggingAutoMultilineExtraPatterns     []string          `toml:"logging_auto_multiline_extra_patterns"`
	LoggingSearchInterval                 time.Duration     `toml:"logging_search_interval"`
	DeprecatedLoggingMinFlushInterval     time.Duration     `toml:"logging_min_flush_interval"`
	LoggingForceFlushLimit                int               `toml:"logging_force_flush_limit"`
	LoggingMaxMultilineLifeDuration       time.Duration     `toml:"logging_max_multiline_life_duration"`
	LoggingFileFromBeginningThresholdSize int               `toml:"logging_file_from_beginning_threshold_size"`
	LoggingRemoveAnsiEscapeCodes          bool              `toml:"logging_remove_ansi_escape_codes"`

	CollectMetricInterval time.Duration `toml:"-"`

	Tags map[string]string `toml:"tags"`
	DeprecatedConf

	Feeder dkio.Feeder
	Tagger datakit.GlobalTagger

	semStop *cliutils.Sem // start stop signal
	pause   *atomic.Bool
	chPause chan bool
}

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) Catalog() string { return "container" }

func (*Input) PipelineConfig() map[string]string { return nil }

func (*Input) GetPipeline() []*tailer.Option { return nil }

func (*Input) RunPipeline() { /*nil*/ }

func (*Input) Singleton() { /*nil*/ }

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLabelLinux, datakit.LabelK8s, datakit.LabelDocker}
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return getCollectorMeasurement()
}

func (ipt *Input) ElectionEnabled() bool { return ipt.Election }

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	select {
	case ipt.chPause <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	select {
	case ipt.chPause <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func newInput() *Input {
	pause := &atomic.Bool{}
	pause.Store(true)
	return &Input{
		EnableContainerMetric:     true,
		EnableK8sMetric:           true,
		EnableK8sEvent:            true,
		EnableK8sNodeLocal:        true,
		Tags:                      make(map[string]string),
		LoggingExtraSourceMap:     make(map[string]string),
		LoggingSourceMultilineMap: make(map[string]string),
		Election:                  true,
		Feeder:                    dkio.DefaultFeeder(),
		Tagger:                    datakit.DefaultGlobalTagger(),
		pause:                     pause,
		chPause:                   make(chan bool, inputs.ElectionPauseChannelLength),
		semStop:                   cliutils.NewSem(),
	}
}

//nolint:gochecknoinits
func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newInput()
	})
	setupMetrics()
}
