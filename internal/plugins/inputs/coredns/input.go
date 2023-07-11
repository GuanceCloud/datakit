// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package coredns collect coreDNS metrics by using input prom
package coredns

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	inputName    = "coredns"
	configSample = `
[[inputs.prom]]
url = "http://127.0.0.1:9153/metrics"
source = "coredns"
metric_types = ["counter", "gauge"]

## filter metrics by names
metric_name_filter = ["^coredns_(acl|cache|dnssec|forward|grpc|hosts|template|dns)_([a-z_]+)$"]

# measurement_prefix = ""
# measurement_name = "prom"

interval = "10s"

# tags_ignore = [""]

## TLS config
tls_open = false
# tls_ca = "/tmp/ca.crt"
# tls_cert = "/tmp/peer.crt"
# tls_key = "/tmp/peer.key"

## customize metrics
[[inputs.prom.measurements]]
  prefix = "coredns_acl_"
  name = "coredns_acl"

[[inputs.prom.measurements]]
  prefix = "coredns_cache_"
  name = "coredns_cache"

[[inputs.prom.measurements]]
  prefix = "coredns_dnssec_"
  name = "coredns_dnssec"

[[inputs.prom.measurements]]
  prefix = "coredns_forward_"
  name = "coredns_forward"

[[inputs.prom.measurements]]
  prefix = "coredns_grpc_"
  name = "coredns_grpc"

[[inputs.prom.measurements]]
  prefix = "coredns_hosts_"
  name = "coredns_hosts"

[[inputs.prom.measurements]]
  prefix = "coredns_template_"
  name = "coredns_template"

[[inputs.prom.measurements]]
  prefix = "coredns_dns_"
  name = "coredns"`
)

type Input struct{}

var _ inputs.InputV2 = (*Input)(nil)

func (i *Input) Terminate() {
	// do nothing
}

func (i *Input) Catalog() string {
	return inputName
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
		&ACLMeasurement{},
		&CacheMeasurement{},
		&DNSSecMeasurement{},
		&ForwardMeasurement{},
		&GrpcMeasurement{},
		&HostsMeasurement{},
		&TemplateMeasurement{},
		&PromMeasurement{},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
