// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	timex "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/time"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "container"
	catelog   = "container"

	dockerEndpoint    = "unix:///var/run/docker.sock"
	containerdAddress = "/var/run/containerd/containerd.sock"
)

var measurements = []inputs.Measurement{}

func registerMeasurement(mea inputs.Measurement) {
	measurements = append(measurements, mea)
}

const sampleCfg = `
[inputs.container]
  docker_endpoint = "unix:///var/run/docker.sock"
  containerd_address = "/var/run/containerd/containerd.sock"

  enable_container_metric = true
  enable_k8s_metric = true
  enable_pod_metric = true
  extract_k8s_label_as_tags = false

  ## Auto-Discovery of PrometheusMonitoring Annotations/CRDs
  enable_autdo_discovery_of_prometheus_service_annotations = false
  enable_autdo_discovery_of_prometheus_pod_monitors = false
  enable_autdo_discovery_of_prometheus_service_monitors = false


  ## Containers logs to include and exclude, default collect all containers. Globs accepted.
  container_include_log = []
  container_exclude_log = ["image:pubrepo.jiagouyun.com/datakit/logfwd*", "image:pubrepo.jiagouyun.com/datakit/datakit*"]

  exclude_pause_container = true

  ## Removes ANSI escape codes from text strings
  logging_remove_ansi_escape_codes = false
  ## Search logging interval, default "60s"
  #logging_search_interval = ""

  ## If the data sent failure, will retry forevery
  logging_blocking_mode = true

  kubernetes_url = "https://kubernetes.default:443"

  ## Authorization level:
  ##   bearer_token -> bearer_token_string -> TLS
  ## Use bearer token for authorization. ('bearer_token' takes priority)
  ## linux at:   /run/secrets/kubernetes.io/serviceaccount/token
  ## windows at: C:\var\run\secrets\kubernetes.io\serviceaccount\token
  bearer_token = "/run/secrets/kubernetes.io/serviceaccount/token"
  # bearer_token_string = "<your-token-string>"

  logging_auto_multiline_detection = true
  logging_auto_multiline_extra_patterns = []

  ## Set true to enable election for k8s metric collection
  election = true

  [inputs.container.logging_extra_source_map]
    # source_regexp = "new_source"

  [inputs.container.logging_source_multiline_map]
    # source = '''^\d{4}'''

  [inputs.container.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`

type DepercatedConf struct {
	EnableMetric           bool           `toml:"enable_metric,omitempty"`
	EnableObject           bool           `toml:"enable_object,omitempty"`
	EnableLogging          bool           `toml:"enable_logging,omitempty"`
	MetricInterval         timex.Duration `toml:"metric_interval,omitempty"`
	MaxLoggingLength       int            `toml:"max_logging_length"`
	IgnoreImageName        []string       `toml:"ignore_image_name,omitempty"`
	IgnoreContainerName    []string       `toml:"ignore_container_name,omitempty"`
	DropTags               []string       `toml:"drop_tags,omitempty"`
	ContainerIncludeMetric []string       `toml:"container_include_metric"`
	ContainerExcludeMetric []string       `toml:"container_exclude_metric"`
	Kubernetes             struct {
		URL                string   `toml:"kubelet_url,omitempty"`
		IgnorePodName      []string `toml:"ignore_pod_name,omitempty"`
		BearerToken        string   `toml:"bearer_token,omitempty"`
		BearerTokenString  string   `toml:"bearer_token_string,omitempty"`
		TLSCA              string   `toml:"tls_ca,omitempty"`
		TLSCert            string   `toml:"tls_cert,omitempty"`
		TLSKey             string   `toml:"tls_key,omitempty"`
		InsecureSkipVerify bool     `toml:"insecure_skip_verify,omitempty"`
	} `toml:"kubelet,omitempty"`
	Logs []struct {
		MatchBy           string   `toml:"match_by,omitempty"`
		Match             []string `toml:"match,omitempty"`
		Source            string   `toml:"source,omitempty"`
		Service           string   `toml:"service,omitempty"`
		Pipeline          string   `toml:"pipeline,omitempty"`
		IgnoreStatus      []string `toml:"ignore_status,omitempty"`
		CharacterEncoding string   `toml:"character_encoding,omitempty"`
		MultilineMatch    string   `toml:"multiline_match,omitempty"`
	} `toml:"log,omitempty"`
	LogDepercated struct {
		FilterMessage []string `toml:"filter_message,omitempty"`
		Source        string   `toml:"source,omitempty"`
		Service       string   `toml:"service,omitempty"`
		Pipeline      string   `toml:"pipeline,omitempty"`
	} `toml:"logfilter,omitempty"`
	PodNameRewriteDeprecated []string `toml:"pod_name_write,omitempty"`
	PodnameRewriteDeprecated []string `toml:"pod_name_rewrite,omitempty"`
}
