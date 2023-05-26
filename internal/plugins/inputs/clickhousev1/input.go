// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package clickhousev1 collect clickhouse metrics by using input prom.
package clickhousev1

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	inputName    = "clickhousev1"
	catalogName  = "db"
	configSample = `
[[inputs.prom]]
  ## Exporter HTTP URL.
  url = "http://127.0.0.1:9363/metrics"

  ## Collector alias.
  source = "clickhouse"

  ## Collect data output.
  # Fill this when want to collect the data to local file nor center.
  # After filling, could use 'datakit --prom-conf /path/to/this/conf' to debug local storage measurement set.
  # Using '--prom-conf' when priority debugging data in 'output' path.
  # output = "/abs/path/to/file"

  ## Collect data upper limit as bytes.
  # Only available when set output to local file.
  # If collect data exceeded the limit, the data would be dropped.
  # Default is 32MB.
  # max_file_size = 0

  ## Metrics type whitelist. Optional: counter, gauge, histogram, summary
  # Default only collect 'counter' and 'gauge'.
  # Collect all if empty.
  metric_types = ["counter", "gauge"]

  ## Metrics name whitelist.
  # Regex supported. Multi supported, conditions met when one matched.
  # Collect all if empty.
  # metric_name_filter = ["cpu"]

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
  prefix = "ClickHouseProfileEvents_"
  name = "ClickHouseProfileEvents"

  [[inputs.prom.measurements]]
  prefix = "ClickHouseMetrics_"
  name = "ClickHouseMetrics"

  [[inputs.prom.measurements]]
  prefix = "ClickHouseAsyncMetrics_"
  name = "ClickHouseAsyncMetrics"

  [[inputs.prom.measurements]]
  prefix = "ClickHouseStatusInfo_"
  name = "ClickHouseStatusInfo"

  ## Customize tags.
  [inputs.prom.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
)

type Input struct{}

var _ inputs.InputV2 = (*Input)(nil)

func (i *Input) Terminate() {
	// do nothing
}

func (i *Input) Catalog() string {
	return catalogName
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
		&AsyncMetricsMeasurement{},
		&MetricsMeasurement{},
		&ProfileEventsMeasurement{},
		&StatusInfoMeasurement{},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
