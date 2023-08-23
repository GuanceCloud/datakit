// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package couchdb collect CouchDB metrics by using input prom
package couchdb

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	inputName    = "couchdb"
	configSample = `
[[inputs.prom]]
  ## Collector alias.
  source = "couchdb"

  ## Exporter URLs.
  urls = ["http://127.0.0.1:17986/_node/_local/_prometheus"]

  ## TLS configuration.
  tls_open = false
  # tls_ca = "/tmp/ca.crt"
  # tls_cert = "/tmp/peer.crt"
  # tls_key = "/tmp/peer.key"

  ## Set to 'true' to enable election.
  election = true

  ## Customize tags.
  [inputs.prom.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
  
  ## (Optional) Collect interval: (defaults to "30s").
  # interval = "30s"
`
)

type Input struct{}

var _ inputs.InputV2 = (*Input)(nil)

func (*Input) Terminate() {
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
		&Measurement{},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
