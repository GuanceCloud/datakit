// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jolokia

import (
	"github.com/GuanceCloud/cliutils/point"
)

type JolokiaMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     int64
}

// Point implement MeasurementV2.
func (m *JolokiaMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}
