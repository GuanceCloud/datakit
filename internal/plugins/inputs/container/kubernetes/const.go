// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var (
	name                   = "kubernetes"
	maxMessageLength       = 256 * 1024 // 256KB
	queryLimit       int64 = 100

	measurements []inputs.Measurement
)

type pointKVs []*typed.PointKV

func Measurements() []inputs.Measurement {
	return measurements
}

func registerMeasurements(meas ...inputs.Measurement) {
	measurements = append(measurements, meas...)
}

func Name() string { return name }
