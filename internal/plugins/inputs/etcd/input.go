// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package etcd collect etcd metrics by using input prom.
package etcd

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	inputName    = "etcd"
	configSample = `
[[inputs.prom]]
  # Exporter URL or file path.
  ## Windows example: C:\\Users
  ## UNIX-like example: /usr/local/
  url = "http://127.0.0.1:2379/metrics"

  ## Collector alias.
  source = "etcd"

  ## Metrics type whitelist. Optional: counter, gauge, histogram, summary
  # Default only collect 'counter' and 'gauge'.
  # Collect all if empty.
  metric_types = ["counter", "gauge"]

  ## Metrics name whitelist.
  # Regex supported. Multi supported, conditions met when one matched.
  # Collect all if empty.
  metric_name_filter = ["etcd_server_proposals","etcd_server_leader","etcd_server_has","etcd_network_client"]

  ## Measurement prefix.
  # Add prefix to measurement set name.
  measurement_prefix = ""

  ## Measurement name.
  # If measurement_name is empty, split metric name by '_', the first field after split as measurement set name, the rest as current metric name.
  # If measurement_name is not empty, using this as measurement set name.
  # Always add 'measurement_prefix' prefix at last.
  measurement_name = "etcd"

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
  #[[inputs.prom.measurements]]
  #  prefix = "etcd_"
  #  name = "etcd"

  ## Customize authentification. For now support Bearer Token only.
  # Filling in 'token' or 'token_file' is acceptable.
  # [inputs.prom.auth]
  # type = "bearer_token"
  # token = "xxxxxxxx"
  # token_file = "/tmp/token"

  ## Customize tags.
  [inputs.prom.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
)

var _ inputs.InputV2 = (*Input)(nil)

type Input struct{}

func (i *Input) Terminate() { /* do nothing */ }

func (i *Input) Catalog() string {
	return "etcd"
}

func (i *Input) SampleConfig() string {
	return configSample
}

func (i *Input) Run() {
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&etcdMeasurement{},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
