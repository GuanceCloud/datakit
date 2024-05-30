// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package prom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
