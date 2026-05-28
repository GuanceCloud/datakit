// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promremote

import (
	"testing"

	"github.com/stretchr/testify/assert"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
)

func TestGetNamesByDefaultRule(t *testing.T) {
	cases := []struct {
		name                  string
		inMetric              string
		inKeepExistMetricName bool
		inMeasurements        []iprom.Rule
		outMeasurementName    string
		outMetricName         string
	}{
		// ---- default behavior (keep_exist_metric_name = false) ----
		{
			name:                  "default-split",
			inMetric:              "etcd_write_bytes_total",
			inKeepExistMetricName: false,
			outMeasurementName:    "etcd",
			outMetricName:         "write_bytes_total",
		},
		{
			name:                  "no-underscore",
			inMetric:              "goroutines",
			inKeepExistMetricName: false,
			outMeasurementName:    "goroutines",
			outMetricName:         "goroutines",
		},

		// ---- keep_exist_metric_name = true ----
		{
			name:                  "keep-exist",
			inMetric:              "etcd_write_bytes_total",
			inKeepExistMetricName: true,
			outMeasurementName:    "etcd",
			outMetricName:         "etcd_write_bytes_total",
		},
		{
			name:                  "keep-exist-no-underscore",
			inMetric:              "goroutines",
			inKeepExistMetricName: true,
			outMeasurementName:    "goroutines",
			outMetricName:         "goroutines",
		},

		// ---- custom measurements rule (unaffected by keep_exist_metric_name) ----
		{
			name:                  "custom-rule-keep-exist-false",
			inMetric:              "etcd_server_has_leader",
			inKeepExistMetricName: false,
			inMeasurements: []iprom.Rule{
				{Prefix: "etcd_server_", Name: "etcd_server"},
			},
			outMeasurementName: "etcd_server",
			outMetricName:      "has_leader",
		},
		{
			name:                  "custom-rule-keep-exist-true",
			inMetric:              "etcd_server_has_leader",
			inKeepExistMetricName: true,
			inMeasurements: []iprom.Rule{
				{Prefix: "etcd_server_", Name: "etcd_server"},
			},
			outMeasurementName: "etcd_server",
			outMetricName:      "has_leader",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &Parser{
				KeepExistMetricName: tc.inKeepExistMetricName,
				Measurements:        tc.inMeasurements,
			}
			measurementName, metricName := p.getNamesByDefaultRule(tc.inMetric)
			assert.Equal(t, tc.outMeasurementName, measurementName)
			assert.Equal(t, tc.outMetricName, metricName)
		})
	}
}
