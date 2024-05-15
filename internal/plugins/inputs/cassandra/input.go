// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cassandra collects cassandra metrics.
package cassandra

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	inputName   = "cassandra"
	catalogName = "db"
)

type Input struct{}

func (*Input) Run() {}

var _ inputs.InputV2 = (*Input)(nil)

func (*Input) Terminate()               {}
func (*Input) Catalog() string          { return catalogName }
func (*Input) SampleConfig() string     { return sampleCfg }
func (*Input) AvailableArchs() []string { return datakit.AllOS }
func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&CassandraMeasurement{},
		&CassandraJVMMeasurement{},
		&CassandraJMXMeasurement{},
		&CassandraDDtraceMeasurement{},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
