// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package prom

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
)

const (
	mockHeader = `
# HELP datakit_http_worker_number The number of the worker
# TYPE datakit_http_worker_number gauge
`
	mockBody = `
datakit_http_worker_number{category="metric",domain="dataway.testing.com",status="%d"} 11.0 1755681983000
datakit_http_worker_number{category="metric",domain="dataway.testing.com",status="%d"} 12.2 1755681983000
datakit_http_worker_number{category="metric",domain="dataway.testing.com",status="%d"} 13.0 1755681983000
datakit_http_worker_number{category="metric",domain="dataway.testing.com",status="%d"} 14.2 1755681983000
datakit_http_worker_number{category="metric",domain="dataway.testing.com",status="%d"} 15.0 1755681983000
`
)

func TestParseMetrics(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString(mockHeader)
	for i := 0; i < 10; i++ {
		buf.WriteString(fmt.Sprintf(mockBody, i, i, i, i, i))
	}

	count := 0

	opts := []PromOption{
		WithMeasurementName("testing-meas"),
		HonorTimestamps(true),
		WithTags(map[string]string{"key-01": "value-01"}),
		WithMaxBatchCallback(1, func(pts []*point.Point) error {
			// for _, pt := range pts {
			// 	t.Log(pt.Pretty())
			// }
			count += len(pts)
			return nil
		}),
	}
	prom, err := NewProm(opts...)
	assert.NoError(t, err)

	_, err = prom.text2MetricsBatch(&buf, "")
	assert.NoError(t, err)

	t.Logf("count: %d\n", count)
}

func BenchmarkParseMetrics(b *testing.B) {
	var buf bytes.Buffer
	buf.WriteString(mockHeader)
	for i := 0; i < 10000; i++ {
		buf.WriteString(fmt.Sprintf(mockBody, i, i, i, i, i))
	}

	opts := []PromOption{
		WithMeasurementName("testing-meas"),
		WithTags(map[string]string{"key-01": "value-01"}),
		WithMaxBatchCallback(1, func(pts []*point.Point) error {
			return nil
		}),
	}
	prom, err := NewProm(opts...)
	assert.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = prom.text2MetricsBatch(&buf, "")
		assert.NoError(b, err)
	}
}

func TestGetNamesByDefault(t *testing.T) {
	cases := []struct {
		inName                string
		inKeepExistMetricName bool

		outMeasurementName string
		outFieldName       string
	}{
		{
			inName:                "etcd_write_bytes_total",
			inKeepExistMetricName: false,
			outMeasurementName:    "etcd",
			outFieldName:          "write_bytes_total",
		},
		{
			inName:                "etcd_write_bytes_total",
			inKeepExistMetricName: true,
			outMeasurementName:    "etcd",
			outFieldName:          "etcd_write_bytes_total",
		},
		{
			inName:                "_",
			inKeepExistMetricName: false,
			outMeasurementName:    "unknown",
			outFieldName:          "unknown",
		},
		{
			inName:                "__",
			inKeepExistMetricName: false,
			outMeasurementName:    "unknown",
			outFieldName:          "unknown",
		},
		{
			inName:                "etcd_",
			inKeepExistMetricName: false,
			outMeasurementName:    "etcd",
			outFieldName:          "etcd",
		},
		{
			inName:                "etcd_",
			inKeepExistMetricName: true,
			outMeasurementName:    "etcd",
			outFieldName:          "etcd",
		},
		{
			inName:                "_etcd",
			inKeepExistMetricName: false,
			outMeasurementName:    "etcd",
			outFieldName:          "etcd",
		},
		{
			inName:                "_etcd_write_bytes_total",
			inKeepExistMetricName: false,
			outMeasurementName:    "etcd",
			outFieldName:          "write_bytes_total",
		},
		{
			inName:                "_etcd_write_bytes_total",
			inKeepExistMetricName: true,
			outMeasurementName:    "etcd",
			outFieldName:          "etcd_write_bytes_total",
		},
		{
			inName:                "___etcd_write_bytes_total",
			inKeepExistMetricName: false,
			outMeasurementName:    "etcd",
			outFieldName:          "write_bytes_total",
		},
		{
			inName:                "___etcd_write_bytes_total",
			inKeepExistMetricName: true,
			outMeasurementName:    "etcd",
			outFieldName:          "etcd_write_bytes_total",
		},
	}

	for _, tc := range cases {
		p := Prom{opt: &option{}}
		p.opt.keepExistMetricName = tc.inKeepExistMetricName

		measurementName, fieldName := p.getNamesByDefault(tc.inName)

		assert.Equal(t, tc.outMeasurementName, measurementName)
		assert.Equal(t, tc.outFieldName, fieldName)
	}
}
