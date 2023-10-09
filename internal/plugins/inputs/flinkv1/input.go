// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package flinkv1 collect flink metrics by using input prom.
package flinkv1

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	inputName    = "flinkv1"
	configSample = `
[[inputs.prom]]
  ## Push gateway URL.
  url = "http://<pushgateway-host>:9091/metrics"

  ## Stream Size. 
  ## The source stream segmentation size.
  ## Default 1, source stream undivided. 
  # stream_size = 1

  ## Collector alias.
  source = "flink"

  ## Metrics type whitelist. Optional: counter, gauge, histogram, summary
  # Default only collect 'counter' and 'gauge'.
  # Collect all if empty.
  metric_types = ["counter", "gauge"]

  ## Metrics name whitelist.
  # Regex supported. Multi supported, conditions met when one matched.
  # Collect all if empty.
  # metric_name_filter = [""]

  ## Measurement prefix.
  # Add prefix to measurement set name.
  measurement_prefix = ""

  ## Measurement name.
  # If measurement_name is empty, split metric name by '_', the first field after split as measurement set name, the rest as current metric name.
  # If measurement_name is not empty, using this as measurement set name.
  # Always add 'measurement_prefix' prefix at last.
  # measurement_name = "prom"

  ## Collect interval, support "ns", "us" (or "µs"), "ms", "s", "m", "h".
  interval = "10s"

  # Ignore tags. Multi supported.
  # The matched tags would be dropped, but the item would still be sent.
  # tags_ignore = ["xxxx"]

  ## TLS configuration.
  tls_open = false
  # tls_ca = "/tmp/ca.crt"
  # tls_cert = "/tmp/peer.crt"
  # tls_key = "/tmp/peer.key"

  ## Customize measurement set name.
  # Treat those metrics with prefix as one set.
  # Prioritier over 'measurement_name' configuration.
  [[inputs.prom.measurements]]
  prefix = "flink_jobmanager_"
  name = "flink_jobmanager"

  [[inputs.prom.measurements]]
  prefix = "flink_taskmanager_"
  name = "flink_taskmanager"

  ## Customize authentification. For now support Bearer Token only.
  # Filling in 'token' or 'token_file' is acceptable.
  # [inputs.prom.auth]
  # type = "bearer_token"
  # token = "xxxxxxxx"
  # token_file = "/tmp/token"

  ## Customize tags.
  # some_tag = "some_value"
`
)

type Input struct{}

var _ inputs.InputV2 = (*Input)(nil)

func (i *Input) Terminate() {
	// do nothing
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) SampleConfig() string {
	return configSample
}

func (*Input) Run() {
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&JobmanagerMeasurement{},
		&TaskmanagerMeasurement{},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
