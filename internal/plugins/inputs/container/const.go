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

	minInterval      = time.Second * 10
	maxInterval      = time.Minute * 5
	maxMessageLength = 256 * 1024
)

const sampleCfg = `
# Container input plugin configuration
[inputs.container]

  # Container runtime endpoints to connect to
  endpoints = [
    "unix:///var/run/docker.sock",
    "unix:///var/run/containerd/containerd.sock",
    "unix:///var/run/crio/crio.sock",
  ]

  # Collection intervals (uncomment to customize)
  # metric_collect_interval = "60s"    # Default: 60s
  # object_collect_interval = "5m"     # Default: 5m
  # logging_search_interval = "60s"    # Default: 60s

  # Feature toggles
  enable_container_metric = true  # Collect container metrics
  enable_k8s_metric       = true  # Collect Kubernetes metrics
  enable_pod_metric       = false # Collect pod-specific metrics
  enable_k8s_event        = true  # Collect Kubernetes events
  enable_k8s_node_local   = true  # Enable node-local collection
  enable_collect_k8s_job  = true  # Collect Kubernetes jobs

  # Kubernetes label extraction as tags
  # Specify label keys to extract as tags (containers use pod labels)
  # Example: ["app", "name", "version"]
  # extract_k8s_label_as_tags_v2            = []
  # extract_k8s_label_as_tags_v2_for_metric = []

  # Container log filtering (supports glob patterns)
  container_include_log = []                                    # Include specific containers (empty = all)
  container_exclude_log = ["image:*logfwd*", "image:*datakit*"] # Exclude containers by image pattern

  # Pod metric filtering (supports glob patterns)
  pod_include_metric = [] # Include specific pods (empty = all)
  pod_exclude_metric = [] # Exclude specific pods

  # Logging configuration
  logging_enable_multiline              = true  # Enable multiline log detection
  logging_auto_multiline_detection      = true  # Auto-detect multiline patterns
  logging_auto_multiline_extra_patterns = []    # Additional multiline patterns

  # Log field filtering
  logging_field_white_list = [] # Only retain specified fields (empty = all fields)

  # Log processing options
  logging_remove_ansi_escape_codes = false # Remove ANSI escape codes from logs
  logging_file_from_beginning      = false # Start reading from beginning of log files

  # Performance tuning
  # logging_max_open_files = 500 # Maximum open files (-1 = unlimited)

  # Source mapping for log collection
  # Maps regex patterns to new source names
  [inputs.container.logging_extra_source_map]
    # ".*nginx.*" = "nginx_logs"
    # ".*apache.*" = "apache_logs"

  # Multiline configuration per source
  # Define multiline patterns for specific log sources
  [inputs.container.logging_source_multiline_map]
    # "java_logs" = '''^\d{4}-\d{2}-\d{2}'''
    # "python_logs" = '''^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}'''

  # Global tags for all collected data
  [inputs.container.tags]
    # environment = "production"
    # cluster = "main"
    # region = "us-west-1"
`

type DeprecatedConf struct {
	LoggingMinFlushInterval                           time.Duration `toml:"logging_min_flush_interval"`
	LoggingMaxMultilineLifeDuration                   time.Duration `toml:"logging_max_multiline_life_duration"`
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
