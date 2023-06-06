// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_metricSender_sendBandwidthUsageMetric(t *testing.T) {
	type Metric struct {
		name  string
		value float64
	}
	tests := []struct {
		name           string
		symbol         SymbolConfig
		fullIndex      string
		values         *ResultValueStore
		expectedMetric []Metric
		expectedError  error
	}{
		{
			"ifBandwidthInUsage.Rate submitted",
			SymbolConfig{OID: "1.3.6.1.2.1.31.1.1.1.6", Name: "ifHCInOctets"},
			"9",
			&ResultValueStore{
				ColumnValues: ColumnResultValuesType{
					// ifHCInOctets
					"1.3.6.1.2.1.31.1.1.1.6": map[string]ResultValue{
						"9": {
							Value: 5000000.0,
						},
					},
					// ifHCOutOctets
					"1.3.6.1.2.1.31.1.1.1.10": map[string]ResultValue{
						"9": {
							Value: 1000000.0,
						},
					},
					// ifHighSpeed
					"1.3.6.1.2.1.31.1.1.1.15": map[string]ResultValue{
						"9": {
							Value: 80.0,
						},
					},
				},
			},
			[]Metric{
				// ((5000000 * 8) / (80 * 1000000)) * 100 = 50.0
				{"ifBandwidthInUsage.rate", 50.0},
			},
			nil,
		},
		{
			"ifBandwidthOutUsage.Rate submitted",
			SymbolConfig{OID: "1.3.6.1.2.1.31.1.1.1.10", Name: "ifHCOutOctets"},
			"9",
			&ResultValueStore{
				ColumnValues: ColumnResultValuesType{
					// ifHCInOctets
					"1.3.6.1.2.1.31.1.1.1.6": map[string]ResultValue{
						"9": {
							Value: 5000000.0,
						},
					},
					// ifHCOutOctets
					"1.3.6.1.2.1.31.1.1.1.10": map[string]ResultValue{
						"9": {
							Value: 1000000.0,
						},
					},
					// ifHighSpeed
					"1.3.6.1.2.1.31.1.1.1.15": map[string]ResultValue{
						"9": {
							Value: 80.0,
						},
					},
				},
			},
			[]Metric{
				// ((1000000 * 8) / (80 * 1000000)) * 100 = 10.0
				{"ifBandwidthOutUsage.rate", 10.0},
			},
			nil,
		},
		{
			"not a bandwidth metric",
			SymbolConfig{OID: "1.3.6.1.2.1.31.1.1.1.99", Name: "notABandwidthMetric"},
			"9",
			&ResultValueStore{
				ColumnValues: ColumnResultValuesType{},
			},
			[]Metric{},
			nil,
		},
		{
			"missing ifHighSpeed",
			SymbolConfig{OID: "1.3.6.1.2.1.31.1.1.1.6", Name: "ifHCInOctets"},
			"9",
			&ResultValueStore{
				ColumnValues: ColumnResultValuesType{
					// ifHCInOctets
					"1.3.6.1.2.1.31.1.1.1.6": map[string]ResultValue{
						"9": {
							Value: 5000000.0,
						},
					},
					// ifHCOutOctets
					"1.3.6.1.2.1.31.1.1.1.10": map[string]ResultValue{
						"9": {
							Value: 1000000.0,
						},
					},
				},
			},
			[]Metric{},
			fmt.Errorf("bandwidth usage: missing `ifHighSpeed` metric, skipping metric. fullIndex=9"),
		},
		{
			"missing ifHCInOctets",
			SymbolConfig{OID: "1.3.6.1.2.1.31.1.1.1.6", Name: "ifHCInOctets"},
			"9",
			&ResultValueStore{
				ColumnValues: ColumnResultValuesType{
					// ifHCOutOctets
					"1.3.6.1.2.1.31.1.1.1.10": map[string]ResultValue{
						"9": {
							Value: 1000000.0,
						},
					},
					// ifHighSpeed
					"1.3.6.1.2.1.31.1.1.1.15": map[string]ResultValue{
						"9": {
							Value: 80.0,
						},
					},
				},
			},
			[]Metric{},
			fmt.Errorf("bandwidth usage: missing `ifHCInOctets` metric, skipping this row. fullIndex=9"),
		},
		{
			"missing ifHCOutOctets",
			SymbolConfig{OID: "1.3.6.1.2.1.31.1.1.1.6", Name: "ifHCOutOctets"},
			"9",
			&ResultValueStore{
				ColumnValues: ColumnResultValuesType{
					// ifHCOutOctets
					"1.3.6.1.2.1.31.1.1.1.10": map[string]ResultValue{
						"9": {
							Value: 1000000.0,
						},
					},
					// ifHighSpeed
					"1.3.6.1.2.1.31.1.1.1.15": map[string]ResultValue{
						"9": {
							Value: 80.0,
						},
					},
				},
			},
			[]Metric{},
			fmt.Errorf("bandwidth usage: missing `ifHCOutOctets` metric, skipping this row. fullIndex=9"),
		},
		{
			"missing ifHCInOctets value",
			SymbolConfig{OID: "1.3.6.1.2.1.31.1.1.1.6", Name: "ifHCInOctets"},
			"9",
			&ResultValueStore{
				ColumnValues: ColumnResultValuesType{
					// ifHCInOctets
					"1.3.6.1.2.1.31.1.1.1.6": map[string]ResultValue{
						"9999": {
							Value: 5000000.0,
						},
					},
					// ifHCOutOctets
					"1.3.6.1.2.1.31.1.1.1.10": map[string]ResultValue{
						"9": {
							Value: 1000000.0,
						},
					},
					// ifHighSpeed
					"1.3.6.1.2.1.31.1.1.1.15": map[string]ResultValue{
						"9": {
							Value: 80.0,
						},
					},
				},
			},
			[]Metric{},
			fmt.Errorf("bandwidth usage: missing value for `ifHCInOctets` metric, skipping this row. fullIndex=9"),
		},
		{
			"missing ifHighSpeed value",
			SymbolConfig{OID: "1.3.6.1.2.1.31.1.1.1.6", Name: "ifHCInOctets"},
			"9",
			&ResultValueStore{
				ColumnValues: ColumnResultValuesType{
					// ifHCInOctets
					"1.3.6.1.2.1.31.1.1.1.6": map[string]ResultValue{
						"9": {
							Value: 5000000.0,
						},
					},
					// ifHCOutOctets
					"1.3.6.1.2.1.31.1.1.1.10": map[string]ResultValue{
						"9": {
							Value: 1000000.0,
						},
					},
					// ifHighSpeed
					"1.3.6.1.2.1.31.1.1.1.15": map[string]ResultValue{
						"999": {
							Value: 80.0,
						},
					},
				},
			},
			[]Metric{},
			fmt.Errorf("bandwidth usage: missing value for `ifHighSpeed`, skipping this row. fullIndex=9"),
		},
		{
			"cannot convert ifHighSpeed to float",
			SymbolConfig{OID: "1.3.6.1.2.1.31.1.1.1.6", Name: "ifHCInOctets"},
			"9",
			&ResultValueStore{
				ColumnValues: ColumnResultValuesType{
					// ifHCInOctets
					"1.3.6.1.2.1.31.1.1.1.6": map[string]ResultValue{
						"9": {
							Value: 5000000.0,
						},
					},
					// ifHCOutOctets
					"1.3.6.1.2.1.31.1.1.1.10": map[string]ResultValue{
						"9": {
							Value: 1000000.0,
						},
					},
					// ifHighSpeed
					"1.3.6.1.2.1.31.1.1.1.15": map[string]ResultValue{
						"9": {
							Value: "abc",
						},
					},
				},
			},
			[]Metric{},
			fmt.Errorf("failed to convert ifHighSpeedValue to float64: failed to parse `abc`: strconv.ParseFloat: parsing \"abc\": invalid syntax"),
		},
		{
			"cannot convert ifHCInOctets to float",
			SymbolConfig{OID: "1.3.6.1.2.1.31.1.1.1.6", Name: "ifHCInOctets"},
			"9",
			&ResultValueStore{
				ColumnValues: ColumnResultValuesType{
					// ifHCInOctets
					"1.3.6.1.2.1.31.1.1.1.6": map[string]ResultValue{
						"9": {
							Value: "abc",
						},
					},
					// ifHCOutOctets
					"1.3.6.1.2.1.31.1.1.1.10": map[string]ResultValue{
						"9": {
							Value: 1000000.0,
						},
					},
					// ifHighSpeed
					"1.3.6.1.2.1.31.1.1.1.15": map[string]ResultValue{
						"9": {
							Value: 80.0,
						},
					},
				},
			},
			[]Metric{},
			fmt.Errorf("failed to convert octetsValue to float64: failed to parse `abc`: strconv.ParseFloat: parsing \"abc\": invalid syntax"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags := []string{"foo:bar"}
			outData := &MetricDatas{}
			err := sendBandwidthUsageMetric(tt.symbol, tt.fullIndex, tt.values, tags, outData)
			assert.Equal(t, tt.expectedError, err)

			for k, metric := range tt.expectedMetric {
				assert.Equal(t, metric.name, outData.Data[k].Name)
				assert.Equal(t, metric.value, outData.Data[k].Value)
				assert.Equal(t, tags, outData.Data[k].Tags)
			}
		})
	}
}

func Test_metricSender_trySendBandwidthUsageMetric(t *testing.T) {
	type Metric struct {
		name  string
		value float64
	}
	tests := []struct {
		name           string
		symbol         SymbolConfig
		fullIndex      string
		values         *ResultValueStore
		expectedMetric []Metric
	}{
		{
			"ifBandwidthInUsage.Rate submitted",
			SymbolConfig{OID: "1.3.6.1.2.1.31.1.1.1.6", Name: "ifHCInOctets"},
			"9",
			&ResultValueStore{
				ColumnValues: ColumnResultValuesType{
					// ifHCInOctets
					"1.3.6.1.2.1.31.1.1.1.6": map[string]ResultValue{
						"9": {
							Value: 5000000.0,
						},
					},
					// ifHCOutOctets
					"1.3.6.1.2.1.31.1.1.1.10": map[string]ResultValue{
						"9": {
							Value: 1000000.0,
						},
					},
					// ifHighSpeed
					"1.3.6.1.2.1.31.1.1.1.15": map[string]ResultValue{
						"9": {
							Value: 80.0,
						},
					},
				},
			},
			[]Metric{
				// ((5000000 * 8) / (80 * 1000000)) * 100 = 50.0
				{"ifBandwidthInUsage.rate", 50.0},
			},
		},
		{
			"should complete even on error",
			SymbolConfig{OID: "1.3.6.1.2.1.31.1.1.1.6", Name: "ifHCInOctets"},
			"9",
			&ResultValueStore{
				ColumnValues: ColumnResultValuesType{
					// ifHCInOctets
					"1.3.6.1.2.1.31.1.1.1.6": map[string]ResultValue{
						"9": {
							Value: 5000000.0,
						},
					},
					// ifHCOutOctets
					"1.3.6.1.2.1.31.1.1.1.10": map[string]ResultValue{
						"9": {
							Value: 1000000.0,
						},
					},
					// ifHighSpeed
					"1.3.6.1.2.1.31.1.1.1.15": map[string]ResultValue{
						"999": {
							Value: 80.0,
						},
					},
				},
			},
			[]Metric{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags := []string{"foo:bar"}
			outData := &MetricDatas{}
			trySendBandwidthUsageMetric(tt.symbol, tt.fullIndex, tt.values, tags, outData)

			for k, metric := range tt.expectedMetric {
				assert.Equal(t, metric.name, outData.Data[k].Name)
				assert.Equal(t, metric.value, outData.Data[k].Value)
				assert.Equal(t, tags, outData.Data[k].Tags)
			}
		})
	}
}
