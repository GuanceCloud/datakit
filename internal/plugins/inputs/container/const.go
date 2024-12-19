// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"time"
)

const (
	inputName = "container"

	objectInterval = time.Minute * 5
	metricInterval = time.Second * 60

	maxMessageLength = 256 * 1024 // 256KB
)

const sampleCfg = `
[inputs.container]
  endpoints = [
    "unix:///var/run/docker.sock",
    "unix:///var/run/containerd/containerd.sock",
    "unix:///var/run/crio/crio.sock",
  ]

  enable_container_metric = true
  enable_k8s_metric       = true
  enable_pod_metric       = false
  enable_k8s_event        = true
  enable_k8s_node_local   = true

  ## Add resource Label as Tags (container use Pod Label), need to specify Label keys.
  ## e.g. ["app", "name"]
  # extract_k8s_label_as_tags_v2            = []
  # extract_k8s_label_as_tags_v2_for_metric = []

  ## Containers logs to include and exclude, default collect all containers. Globs accepted.
  container_include_log = []
  container_exclude_log = ["image:*logfwd*", "image:*datakit*"]

  kubernetes_url = "https://kubernetes.default:443"

  ## Authorization level:
  ##   bearer_token -> bearer_token_string -> TLS
  ## Use bearer token for authorization. ('bearer_token' takes priority)
  ## linux at:   /run/secrets/kubernetes.io/serviceaccount/token
  bearer_token = "/run/secrets/kubernetes.io/serviceaccount/token"
  # bearer_token_string = "<your-token-string>"

  ## Set true to enable election for k8s metric collection
  election = true

  logging_enable_multiline             = true
  logging_auto_multiline_detection     = true
  logging_auto_multiline_extra_patterns = []

  ## Only retain the fields specified in the whitelist.
  logging_field_white_list = []

  ## Removes ANSI escape codes from text strings.
  logging_remove_ansi_escape_codes = false

  ## Whether to collect logs from the begin of the file.
  logging_file_from_beginning = false

  ## The maximum allowed number of open files, default is 500. If it is -1, it means no limit.
  # logging_max_open_files = 500

  ## Search logging interval, default "60s"
  #logging_search_interval = ""

  [inputs.container.logging_extra_source_map]
    # source_regexp = "new_source"

  [inputs.container.logging_source_multiline_map]
    # source = '''^\d{4}'''

  [inputs.container.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`

type DeprecatedConf struct {
	LoggingMinFlushInterval                           time.Duration `toml:"logging_min_flush_interval"`
	LoggingBlockingMode                               bool          `toml:"logging_blocking_mode"`
	LoggingForceFlushLimit                            int           `toml:"logging_force_flush_limit"`
	ExcludePauseContainer                             bool          `toml:"exclude_pause_container"`
	DisableK8sEvents                                  bool          `toml:"disable_k8s_events"`
	EnableAutoDiscoveryOfPrometheusPodAnnotations     bool          `toml:"enable_auto_discovery_of_prometheus_pod_annotations"`
	EnableAutoDiscoveryOfPrometheusServiceAnnotations bool          `toml:"enable_auto_discovery_of_prometheus_service_annotations"`
	EnableAutoDiscoveryOfPrometheusPodMonitors        bool          `toml:"enable_auto_discovery_of_prometheus_pod_monitors"`
	EnableAutoDiscoveryOfPrometheusServiceMonitors    bool          `toml:"enable_auto_discovery_of_prometheus_service_monitors"`
	KeepExistPrometheusMetricName                     bool          `toml:"keep_exist_prometheus_metric_name"`
	EnableK8sSelfMetricByProm                         bool          `toml:"enable_k8s_self_metric_by_prom"`
}
