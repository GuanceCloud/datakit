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

func Test_metricSender_sendMemoryUsageMetric(t *testing.T) {
	type MetricSamplesStore struct {
		scalarSamples map[string]MetricSample
		columnSamples map[string]map[string]MetricSample
	}
	type Metric struct {
		name  string
		value float64
		tags  []string
	}
	tests := []struct {
		name            string
		samplesStore    MetricSamplesStore
		expectedMetrics []Metric
		expectedError   error
	}{
		{
			"should not emit evaluated memory.usage when scalar memory.usage is collected",
			MetricSamplesStore{scalarSamples: map[string]MetricSample{
				"memory.usage": {
					value:      ResultValue{Value: 100.0},
					tags:       []string{"device_namespace:default", "ip_address:192.168.10.24"},
					symbol:     SymbolConfig{Name: "memory.usage"},
					options:    MetricsConfigOption{},
					forcedType: "",
				},
			}},
			[]Metric{},
			nil,
		},
		{
			"should not emit evaluated memory.usage when column memory.usage is collected",
			MetricSamplesStore{columnSamples: map[string]map[string]MetricSample{
				"memory.usage": {
					"123": {
						value:      ResultValue{Value: 100.0},
						tags:       []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:123"},
						symbol:     SymbolConfig{Name: "memory.usage"},
						options:    MetricsConfigOption{},
						forcedType: "",
					},
					"567": {
						value:      ResultValue{Value: 100.0},
						tags:       []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:567"},
						symbol:     SymbolConfig{Name: "memory.usage"},
						options:    MetricsConfigOption{},
						forcedType: "",
					},
				},
			}},
			[]Metric{},
			nil,
		},
		{
			"should not emit evaluated memory.usage when only scalar memory.used is collected",
			MetricSamplesStore{scalarSamples: map[string]MetricSample{
				"memory.used": {
					value:      ResultValue{Value: 100.0},
					tags:       []string{"device_namespace:default", "ip_address:192.168.10.24"},
					symbol:     SymbolConfig{Name: "memory.used"},
					options:    MetricsConfigOption{},
					forcedType: "",
				},
			}},
			[]Metric{},
			fmt.Errorf("missing free, total memory metrics, skipping scalar memory usage"),
		},
		{
			"should not emit evaluated memory.usage when only scalar memory.free is collected",
			MetricSamplesStore{scalarSamples: map[string]MetricSample{
				"memory.free": {
					value:      ResultValue{Value: 100.0},
					tags:       []string{"device_namespace:default", "ip_address:192.168.10.24"},
					symbol:     SymbolConfig{Name: "memory.free"},
					options:    MetricsConfigOption{},
					forcedType: "",
				},
			}},
			[]Metric{},
			fmt.Errorf("missing used, total memory metrics, skipping scalar memory usage"),
		},
		{
			"should not emit evaluated memory.usage when only scalar memory.total is collected",
			MetricSamplesStore{scalarSamples: map[string]MetricSample{
				"memory.total": {
					value:      ResultValue{Value: 100.0},
					tags:       []string{"device_namespace:default", "ip_address:192.168.10.24"},
					symbol:     SymbolConfig{Name: "memory.total"},
					options:    MetricsConfigOption{},
					forcedType: "",
				},
			}},
			[]Metric{},
			fmt.Errorf("missing used, free memory metrics, skipping scalar memory usage"),
		},
		{
			"should not emit evaluated memory.usage when only column memory.used is collected",
			MetricSamplesStore{columnSamples: map[string]map[string]MetricSample{
				"memory.used": {
					"123": {
						value:      ResultValue{Value: 100.0},
						tags:       []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:123"},
						symbol:     SymbolConfig{Name: "memory.used"},
						options:    MetricsConfigOption{},
						forcedType: "",
					},
					"567": {
						value:      ResultValue{Value: 100.0},
						tags:       []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:567"},
						symbol:     SymbolConfig{Name: "memory.used"},
						options:    MetricsConfigOption{},
						forcedType: "",
					},
				},
			}},
			[]Metric{},
			fmt.Errorf("missing free, total memory metrics, skipping column memory usage"),
		},
		{
			"should not emit evaluated memory.usage when only column memory.free is collected",
			MetricSamplesStore{columnSamples: map[string]map[string]MetricSample{
				"memory.free": {
					"123": {
						value:      ResultValue{Value: 100.0},
						tags:       []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:123"},
						symbol:     SymbolConfig{Name: "memory.free"},
						options:    MetricsConfigOption{},
						forcedType: "",
					},
					"567": {
						value:      ResultValue{Value: 100.0},
						tags:       []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:567"},
						symbol:     SymbolConfig{Name: "memory.free"},
						options:    MetricsConfigOption{},
						forcedType: "",
					},
				},
			}},
			[]Metric{},
			fmt.Errorf("missing used, total memory metrics, skipping column memory usage"),
		},
		{
			"should not emit evaluated memory.usage when only column memory.total is collected",
			MetricSamplesStore{columnSamples: map[string]map[string]MetricSample{
				"memory.total": {
					"123": {
						value:      ResultValue{Value: 100.0},
						tags:       []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:123"},
						symbol:     SymbolConfig{Name: "memory.total"},
						options:    MetricsConfigOption{},
						forcedType: "",
					},
					"567": {
						value:      ResultValue{Value: 100.0},
						tags:       []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:567"},
						symbol:     SymbolConfig{Name: "memory.total"},
						options:    MetricsConfigOption{},
						forcedType: "",
					},
				},
			}},
			[]Metric{},
			fmt.Errorf("missing used, free memory metrics, skipping column memory usage"),
		},
		{
			"should not emit evaluated memory.usage when no memory metric is collected",
			MetricSamplesStore{},
			[]Metric{},
			fmt.Errorf("missing used, free, total memory metrics, skipping column memory usage"),
		},
		{
			"should emit evaluated memory.usage when scalar memory.used and memory.total are collected",
			MetricSamplesStore{scalarSamples: map[string]MetricSample{
				"memory.used": {
					value:      ResultValue{Value: 50.0},
					tags:       []string{"device_namespace:default", "ip_address:192.168.10.24"},
					symbol:     SymbolConfig{Name: "memory.used"},
					options:    MetricsConfigOption{},
					forcedType: "",
				},
				"memory.total": {
					value:      ResultValue{Value: 200.0},
					tags:       []string{"device_namespace:default", "ip_address:192.168.10.24"},
					symbol:     SymbolConfig{Name: "memory.total"},
					options:    MetricsConfigOption{},
					forcedType: "",
				},
			}},
			[]Metric{
				{"memory.usage", 25.0, []string{"device_namespace:default", "ip_address:192.168.10.24"}},
			},
			nil,
		},
		{
			"should emit evaluated memory.usage when scalar memory.used and memory.free are collected",
			MetricSamplesStore{scalarSamples: map[string]MetricSample{
				"memory.used": {
					value:      ResultValue{Value: 50.0},
					tags:       []string{"device_namespace:default", "ip_address:192.168.10.24"},
					symbol:     SymbolConfig{Name: "memory.used"},
					options:    MetricsConfigOption{},
					forcedType: "",
				},
				"memory.free": {
					value:      ResultValue{Value: 150.0},
					tags:       []string{"device_namespace:default", "ip_address:192.168.10.24"},
					symbol:     SymbolConfig{Name: "memory.free"},
					options:    MetricsConfigOption{},
					forcedType: "",
				},
			}},
			[]Metric{
				{"memory.usage", 25.0, []string{"device_namespace:default", "ip_address:192.168.10.24"}},
			},
			nil,
		},
		{
			"should emit evaluated memory.usage when scalar memory.free and memory.total are collected",
			MetricSamplesStore{scalarSamples: map[string]MetricSample{
				"memory.free": {
					value:      ResultValue{Value: 150.0},
					tags:       []string{"device_namespace:default", "ip_address:192.168.10.24"},
					symbol:     SymbolConfig{Name: "memory.free"},
					options:    MetricsConfigOption{},
					forcedType: "",
				},
				"memory.total": {
					value:      ResultValue{Value: 200.0},
					tags:       []string{"device_namespace:default", "ip_address:192.168.10.24"},
					symbol:     SymbolConfig{Name: "memory.total"},
					options:    MetricsConfigOption{},
					forcedType: "",
				},
			}},
			[]Metric{
				{"memory.usage", 25.0, []string{"device_namespace:default", "ip_address:192.168.10.24"}},
			},
			nil,
		},
		{
			"should emit evaluated memory.usage when column memory.used and memory.total are collected",
			MetricSamplesStore{columnSamples: map[string]map[string]MetricSample{
				"memory.used": {
					"123": {
						value:      ResultValue{Value: 100.0},
						tags:       []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:123"},
						symbol:     SymbolConfig{Name: "memory.used"},
						options:    MetricsConfigOption{},
						forcedType: "",
					},
					"567": {
						value:      ResultValue{Value: 20.0},
						tags:       []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:567"},
						symbol:     SymbolConfig{Name: "memory.used"},
						options:    MetricsConfigOption{},
						forcedType: "",
					},
				},
				"memory.total": {
					"123": {
						value:      ResultValue{Value: 200.0},
						tags:       []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:123"},
						symbol:     SymbolConfig{Name: "memory.total"},
						options:    MetricsConfigOption{},
						forcedType: "",
					},
					"567": {
						value:      ResultValue{Value: 200.0},
						tags:       []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:567"},
						symbol:     SymbolConfig{Name: "memory.total"},
						options:    MetricsConfigOption{},
						forcedType: "",
					},
				},
			}},
			[]Metric{
				{"memory.usage", 50.0, []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:123"}},
				{"memory.usage", 10.0, []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:567"}},
			},
			nil,
		},
		{
			"should emit evaluated memory.usage when column memory.used and memory.free are collected",
			MetricSamplesStore{columnSamples: map[string]map[string]MetricSample{
				"memory.used": {
					"123": {
						value:      ResultValue{Value: 100.0},
						tags:       []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:123"},
						symbol:     SymbolConfig{Name: "memory.usage"},
						options:    MetricsConfigOption{},
						forcedType: "",
					},
					"567": {
						value:      ResultValue{Value: 20.0},
						tags:       []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:567"},
						symbol:     SymbolConfig{Name: "memory.usage"},
						options:    MetricsConfigOption{},
						forcedType: "",
					},
				},
				"memory.free": {
					"123": {
						value:      ResultValue{Value: 100.0},
						tags:       []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:123"},
						symbol:     SymbolConfig{Name: "memory.usage"},
						options:    MetricsConfigOption{},
						forcedType: "",
					},
					"567": {
						value:      ResultValue{Value: 180.0},
						tags:       []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:567"},
						symbol:     SymbolConfig{Name: "memory.usage"},
						options:    MetricsConfigOption{},
						forcedType: "",
					},
				},
			}},
			[]Metric{
				{"memory.usage", 50.0, []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:123"}},
				{"memory.usage", 10.0, []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:567"}},
			},
			nil,
		},
		{
			"should emit evaluated memory.usage when column memory.free and memory.total are collected",
			MetricSamplesStore{columnSamples: map[string]map[string]MetricSample{
				"memory.free": {
					"123": {
						value:      ResultValue{Value: 100.0},
						tags:       []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:123"},
						symbol:     SymbolConfig{Name: "memory.usage"},
						options:    MetricsConfigOption{},
						forcedType: "",
					},
					"567": {
						value:      ResultValue{Value: 180.0},
						tags:       []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:567"},
						symbol:     SymbolConfig{Name: "memory.usage"},
						options:    MetricsConfigOption{},
						forcedType: "",
					},
				},
				"memory.total": {
					"123": {
						value:      ResultValue{Value: 200.0},
						tags:       []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:123"},
						symbol:     SymbolConfig{Name: "memory.usage"},
						options:    MetricsConfigOption{},
						forcedType: "",
					},
					"567": {
						value:      ResultValue{Value: 200.0},
						tags:       []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:567"},
						symbol:     SymbolConfig{Name: "memory.usage"},
						options:    MetricsConfigOption{},
						forcedType: "",
					},
				},
			}},
			[]Metric{
				{"memory.usage", 50.0, []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:123"}},
				{"memory.usage", 10.0, []string{"device_namespace:default", "ip_address:192.168.10.24", "mem:567"}},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outData := &MetricDatas{}
			err := tryReportMemoryUsage(tt.samplesStore.scalarSamples, tt.samplesStore.columnSamples, outData)
			assert.Equal(t, tt.expectedError, err)

			assert.Equal(t, len(tt.expectedMetrics), len(outData.Data))
			for _, metric := range tt.expectedMetrics {
				var bEqualName, bEqualValue bool
				for _, v := range outData.Data {
					if bEqualName && bEqualValue {
						break
					}

					if metric.name == v.Name {
						bEqualName = true
					} else {
						continue // not this one.
					}
					if metric.value == v.Value {
						bEqualValue = true
					}

					if bEqualName && bEqualValue {
						assert.Equal(t, metric.tags, v.Tags)
					}
				}
				assert.Equal(t, true, bEqualName && bEqualValue)
			}
		})
	}
}
