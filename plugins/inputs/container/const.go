package container

import (
	"reflect"
	"time"

	timex "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/time"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "container"
	catelog   = "container"

	dockerEndpoint = "unix:///var/run/docker.sock"

	apiTimeoutDuration = time.Second * 5
)

var measurements = make(map[reflect.Type]inputs.Measurement)

func registerMeasurement(mea inputs.Measurement) {
	measurements[reflect.TypeOf(mea)] = mea
}

const sampleCfg = `
[inputs.container]
  endpoint = "unix:///var/run/docker.sock"

  ## Containers metrics to include and exclude, default not collect. Globs accepted.
  container_include_metric = []
  container_exclude_metric = ["image:*"]

  ## Containers logs to include and exclude, default collect all containers. Globs accepted.
  container_include_log = ["image:*"]
  container_exclude_log = []

  exclude_pause_container = true

  ## Removes ANSI escape codes from text strings
  logging_remove_ansi_escape_codes = false
  
  kubernetes_url = "https://kubernetes.default:443"

  ## Authorization level:
  ##   bearer_token -> bearer_token_string -> TLS
  ## Use bearer token for authorization. ('bearer_token' takes priority)
  ## linux at:   /run/secrets/kubernetes.io/serviceaccount/token
  ## windows at: C:\var\run\secrets\kubernetes.io\serviceaccount\token
  bearer_token = "/run/secrets/kubernetes.io/serviceaccount/token"
  # bearer_token_string = "<your-token-string>"

  [inputs.container.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`

type DepercatedConf struct {
	EnableMetric        bool           `toml:"enable_metric,omitempty"`
	EnableObject        bool           `toml:"enable_object,omitempty"`
	EnableLogging       bool           `toml:"enable_logging,omitempty"`
	MetricInterval      timex.Duration `toml:"metric_interval,omitempty"`
	IgnoreImageName     []string       `toml:"ignore_image_name,omitempty"`
	IgnoreContainerName []string       `toml:"ignore_container_name,omitempty"`
	DropTags            []string       `toml:"drop_tags,omitempty"`
	Kubernetes          struct {
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
