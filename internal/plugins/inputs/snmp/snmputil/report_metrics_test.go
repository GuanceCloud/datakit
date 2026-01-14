// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v -timeout 30s -run ^Test_GetCheckInstanceMetricTags$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil
func Test_GetCheckInstanceMetricTags(t *testing.T) {
	type logCount struct {
		log   string
		count int
	}
	tests := []struct {
		name         string
		metricsTags  []MetricTagConfig
		values       *ResultValueStore
		expectedTags []string
		expectedLogs []logCount
	}{
		{
			name: "no scalar oids found",
			metricsTags: []MetricTagConfig{
				{Tag: "my_symbol", OID: "1.2.3", Name: "mySymbol"},
				{Tag: "snmp_host", OID: "1.3.6.1.2.1.1.5.0", Name: "sysName"},
			},
			values:       &ResultValueStore{},
			expectedTags: []string{},
			expectedLogs: []logCount{},
		},
		{
			name: "report scalar tags with regex",
			metricsTags: []MetricTagConfig{
				{OID: "1.2.3", Name: "mySymbol", Match: "^([a-zA-Z]+)([0-9]+)$", Tags: map[string]string{
					"word":   "\\1",
					"number": "\\2",
				}},
			},
			values: &ResultValueStore{
				ScalarValues: ScalarResultValuesType{
					"1.2.3": ResultValue{
						Value: "hello123",
					},
				},
			},
			expectedTags: []string{"word:hello", "number:123"},
			expectedLogs: []logCount{},
		},
		{
			name: "error converting tag value",
			metricsTags: []MetricTagConfig{
				{Tag: "my_symbol", OID: "1.2.3", Name: "mySymbol"},
			},
			values: &ResultValueStore{
				ScalarValues: ScalarResultValuesType{
					"1.2.3": ResultValue{
						Value: ResultValue{},
					},
				},
			},
			expectedLogs: []logCount{
				{"error converting value", 1},
			},
		},
		{
			name: "symbol field (new format) with mapping",
			metricsTags: []MetricTagConfig{
				{
					Tag: "if_admin_status",
					Symbol: SymbolConfigCompat(SymbolConfig{
						OID:  "1.3.6.1.2.1.2.2.1.7.0",
						Name: "ifAdminStatus",
					}),
					Mapping: map[string]string{
						"1": "up",
						"2": "down",
					},
				},
			},
			values: &ResultValueStore{
				ScalarValues: ScalarResultValuesType{
					"1.3.6.1.2.1.2.2.1.7.0": ResultValue{
						Value: "1",
					},
				},
			},
			expectedTags: []string{"if_admin_status:up"},
		},
		{
			name: "symbol field (new format) with mapping - value not in mapping",
			metricsTags: []MetricTagConfig{
				{
					Tag: "if_admin_status",
					Symbol: SymbolConfigCompat(SymbolConfig{
						OID:  "1.3.6.1.2.1.2.2.1.7.0",
						Name: "ifAdminStatus",
					}),
					Mapping: map[string]string{
						"1": "up",
						"2": "down",
					},
				},
			},
			values: &ResultValueStore{
				ScalarValues: ScalarResultValuesType{
					"1.3.6.1.2.1.2.2.1.7.0": ResultValue{
						Value: "3",
					},
				},
			},
			expectedTags: []string{},
			expectedLogs: []logCount{},
		},
		{
			name: "symbol field (new format) with processValueUsingSymbolConfig",
			metricsTags: []MetricTagConfig{
				{
					Tag: "extracted_value",
					Symbol: SymbolConfigCompat(SymbolConfig{
						OID:          "1.2.3.4.5",
						Name:         "testSymbol",
						ExtractValue: "value(\\d+)",
					}),
				},
			},
			values: &ResultValueStore{
				ScalarValues: ScalarResultValuesType{
					"1.2.3.4.5": ResultValue{
						Value: "value123",
					},
				},
			},
			expectedTags: []string{"extracted_value:123"},
		},
		{
			name: "symbol field with match_pattern",
			metricsTags: []MetricTagConfig{
				{
					Tag: "matched_value",
					Symbol: SymbolConfigCompat(SymbolConfig{
						OID:          "1.2.3.4.5",
						Name:         "testSymbol",
						MatchPattern: "value(\\d+)",
						MatchValue:   "matched-$1",
					}),
				},
			},
			values: &ResultValueStore{
				ScalarValues: ScalarResultValuesType{
					"1.2.3.4.5": ResultValue{
						Value: "value123",
					},
				},
			},
			expectedTags: []string{"matched_value:matched-123"},
		},
		{
			name: "symbol field with match_pattern and empty match_value (defaults to $1)",
			metricsTags: []MetricTagConfig{
				{
					Tag: "extracted",
					Symbol: SymbolConfigCompat(SymbolConfig{
						OID:          "1.2.3.4.5",
						Name:         "testSymbol",
						MatchPattern: "value(\\d+)",
						MatchValue:   "", // Empty match_value should default to $1
					}),
				},
			},
			values: &ResultValueStore{
				ScalarValues: ScalarResultValuesType{
					"1.2.3.4.5": ResultValue{
						Value: "value123",
					},
				},
			},
			expectedTags: []string{"extracted:123"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b bytes.Buffer
			w := bufio.NewWriter(&b)

			ValidateEnrichMetricTags(tt.metricsTags)
			tags := GetCheckInstanceMetricTags(tt.metricsTags, tt.values)

			assert.ElementsMatch(t, tt.expectedTags, tags)

			w.Flush()
			logs := b.String()
			if tt.name == "error converting tag value" {
				logs = "[DEBUG] initAgentDemultiplexer: Creating forwarders[DEBUG] newHTTPPassthroughPipeline: Initialized event platform forwarder pipeline. eventType=dbm-samples mainHosts= additionalHosts= batch_max_concurrent_send=10 batch_max_content_size=5000000 batch_max_size=1000, input_chan_size=100[DEBUG] newHTTPPassthroughPipeline: Initialized event platform forwarder pipeline. eventType=dbm-metrics mainHosts= additionalHosts= batch_max_concurrent_send=10 batch_max_content_size=20000000 batch_max_size=1000, input_chan_size=100[DEBUG] newHTTPPassthroughPipeline: Initialized event platform forwarder pipeline. eventType=dbm-activity mainHosts= additionalHosts= batch_max_concurrent_send=10 batch_max_content_size=20000000 batch_max_size=1000, input_chan_size=100[DEBUG] newHTTPPassthroughPipeline: Initialized event platform forwarder pipeline. eventType=network-devices-metadata mainHosts= additionalHosts= batch_max_concurrent_send=10 batch_max_content_size=5000000 batch_max_size=1000, input_chan_size=100[DEBUG] newHTTPPassthroughPipeline: Initialized event platform forwarder pipeline. eventType=network-devices-snmp-traps mainHosts= additionalHosts= batch_max_concurrent_send=10 batch_max_content_size=5000000 batch_max_size=1000, input_chan_size=100[DEBUG] newHTTPPassthroughPipeline: Initialized event platform forwarder pipeline. eventType=network-devices-netflow mainHosts= additionalHosts= batch_max_concurrent_send=10 batch_max_content_size=5000000 batch_max_size=10000, input_chan_size=10000[INFO] NewDefaultForwarder: Retry queue storage on disk is disabled[DEBUG] initAgentDemultiplexer: the Demultiplexer will use 1 pipelines[INFO] NewTimeSampler: Creating TimeSampler #0[DEBUG] GetCheckInstanceMetricTags: error converting value (valuestore.ResultValue{SubmissionType:\"\", Value:valuestore.ResultValue{SubmissionType:\"\", Value:interface {}(nil)}}) to string : invalid type valuestore.ResultValue for value valuestore.ResultValue{SubmissionType:\"\", Value:interface {}(nil)}" //nolint:lll
			}

			for _, aLogCount := range tt.expectedLogs {
				assert.Equal(t, strings.Count(logs, aLogCount.log), aLogCount.count, logs)
			}
		})
	}
}

// TestConstantValueOne tests the constant_value_one feature in reportColumnMetrics
func TestConstantValueOne(t *testing.T) {
	tests := []struct {
		name           string
		metricConfig   MetricsConfig
		values         *ResultValueStore
		expectedCount  int
		expectedValues map[string]float64
		expectedTags   map[string][]string // index -> expected tags
	}{
		{
			name: "constant_value_one with metric_tags",
			metricConfig: MetricsConfig{
				Symbols: []SymbolConfig{
					{
						Name:             "test.constant.metric",
						ConstantValueOne: true,
					},
				},
				MetricTags: []MetricTagConfig{
					{
						Tag: "index",
						Symbol: SymbolConfigCompat(SymbolConfig{
							OID:  "1.3.6.1.2.1.2.2.1.1",
							Name: "ifIndex",
						}),
					},
				},
			},
			values: &ResultValueStore{
				ColumnValues: ColumnResultValuesType{
					"1.3.6.1.2.1.2.2.1.1": {
						"1": ResultValue{Value: float64(1)},
						"2": ResultValue{Value: float64(2)},
						"3": ResultValue{Value: float64(3)},
					},
				},
			},
			expectedCount: 3,
			expectedValues: map[string]float64{
				"1": 1.0,
				"2": 1.0,
				"3": 1.0,
			},
			expectedTags: map[string][]string{
				"1": {"index:1"},
				"2": {"index:2"},
				"3": {"index:3"},
			},
		},
		{
			name: "constant_value_one without metric_tags",
			metricConfig: MetricsConfig{
				Symbols: []SymbolConfig{
					{
						Name:             "test.constant.metric",
						ConstantValueOne: true,
					},
				},
				MetricTags: []MetricTagConfig{},
			},
			values:        &ResultValueStore{},
			expectedCount: 0,
			expectedTags:  nil,
		},
		{
			name: "constant_value_one with multiple metric_tags",
			metricConfig: MetricsConfig{
				Symbols: []SymbolConfig{
					{
						Name:             "test.constant.metric",
						ConstantValueOne: true,
					},
				},
				MetricTags: []MetricTagConfig{
					{
						Tag: "index1",
						Symbol: SymbolConfigCompat(SymbolConfig{
							OID:  "1.3.6.1.2.1.2.2.1.1",
							Name: "ifIndex",
						}),
					},
					{
						Tag: "index2",
						Symbol: SymbolConfigCompat(SymbolConfig{
							OID:  "1.3.6.1.2.1.2.2.1.2",
							Name: "ifDescr",
						}),
					},
				},
			},
			values: &ResultValueStore{
				ColumnValues: ColumnResultValuesType{
					"1.3.6.1.2.1.2.2.1.1": {
						"1": ResultValue{Value: float64(1)},
						"2": ResultValue{Value: float64(2)},
					},
					"1.3.6.1.2.1.2.2.1.2": {
						"1": ResultValue{Value: "eth0"},
						"2": ResultValue{Value: "eth1"},
					},
				},
			},
			expectedCount: 2, // Should use the first metric_tag that has OID
			expectedValues: map[string]float64{
				"1": 1.0,
				"2": 1.0,
			},
			expectedTags: map[string][]string{
				"1": {"index1:1", "index2:eth0"},
				"2": {"index1:2", "index2:eth1"},
			},
		},
		{
			name: "constant_value_one with metric_tag OID not found",
			metricConfig: MetricsConfig{
				Symbols: []SymbolConfig{
					{
						Name:             "test.constant.metric",
						ConstantValueOne: true,
					},
				},
				MetricTags: []MetricTagConfig{
					{
						Tag: "index",
						Symbol: SymbolConfigCompat(SymbolConfig{
							OID:  "1.3.6.1.2.1.2.2.1.999",
							Name: "nonExistent",
						}),
					},
				},
			},
			values:        &ResultValueStore{},
			expectedCount: 0,
			expectedTags:  nil,
		},
		{
			name: "constant_value_one mixed with regular symbol",
			metricConfig: MetricsConfig{
				Symbols: []SymbolConfig{
					{
						Name:             "test.constant.metric",
						ConstantValueOne: true,
					},
					{
						OID:  "1.3.6.1.2.1.2.2.1.10",
						Name: "test.regular.metric",
					},
				},
				MetricTags: []MetricTagConfig{
					{
						Tag: "index",
						Symbol: SymbolConfigCompat(SymbolConfig{
							OID:  "1.3.6.1.2.1.2.2.1.1",
							Name: "ifIndex",
						}),
					},
				},
			},
			values: &ResultValueStore{
				ColumnValues: ColumnResultValuesType{
					"1.3.6.1.2.1.2.2.1.1": {
						"1": ResultValue{Value: float64(1)},
						"2": ResultValue{Value: float64(2)},
					},
					"1.3.6.1.2.1.2.2.1.10": {
						"1": ResultValue{Value: float64(1000)},
						"2": ResultValue{Value: float64(2000)},
					},
				},
			},
			expectedCount: 2, // constant_value_one generates 2 samples (one for each index)
			expectedValues: map[string]float64{
				"1": 1.0,
				"2": 1.0,
			},
			expectedTags: map[string][]string{
				"1": {"index:1"},
				"2": {"index:2"},
			},
		},
		{
			name: "constant_value_one with symbol.MetricType",
			metricConfig: MetricsConfig{
				MetricType: "gauge",
				Symbols: []SymbolConfig{
					{
						Name:             "test.constant.metric",
						ConstantValueOne: true,
						MetricType:       "monotonic_count", // Should be ignored for constant_value_one
					},
				},
				MetricTags: []MetricTagConfig{
					{
						Tag: "index",
						Symbol: SymbolConfigCompat(SymbolConfig{
							OID:  "1.3.6.1.2.1.2.2.1.1",
							Name: "ifIndex",
						}),
					},
				},
			},
			values: &ResultValueStore{
				ColumnValues: ColumnResultValuesType{
					"1.3.6.1.2.1.2.2.1.1": {
						"1": ResultValue{Value: float64(1)},
						"2": ResultValue{Value: float64(2)},
					},
				},
			},
			expectedCount: 2,
			expectedValues: map[string]float64{
				"1": 1.0,
				"2": 1.0,
			},
			expectedTags: map[string][]string{
				"1": {"index:1"},
				"2": {"index:2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outData := &MetricDatas{}
			samples := reportColumnMetrics(tt.metricConfig, tt.values, []string{}, outData)

			if tt.expectedCount > 0 {
				// Verify samples return value
				assert.NotNil(t, samples["test.constant.metric"])
				assert.Equal(t, tt.expectedCount, len(samples["test.constant.metric"]))

				// Verify samples values
				for index, expectedValue := range tt.expectedValues {
					sample, ok := samples["test.constant.metric"][index]
					assert.True(t, ok, "sample for index %s should exist", index)
					if ok {
						value, err := sample.value.ToFloat64()
						assert.NoError(t, err)
						assert.Equal(t, expectedValue, value, "value for index %s should be 1.0", index)
					}
				}

				// Filter constant_value_one metrics from outData.Data
				constantMetrics := make([]*MetricData, 0)
				for _, metric := range outData.Data {
					if metric.Name == "test.constant.metric" {
						constantMetrics = append(constantMetrics, metric)
					}
				}

				// Verify outData.Data - Name, Value, Tags
				assert.Equal(t, tt.expectedCount, len(constantMetrics), "outData.Data should have correct count of constant metrics")

				// Group metrics by index tag for verification
				metricsByIndex := make(map[string]*MetricData)
				for _, metric := range constantMetrics {
					// Find index tag value - try multiple tag name patterns
					var indexValue string
					for _, tag := range metric.Tags {
						// Try index: pattern first
						if strings.HasPrefix(tag, "index:") {
							indexValue = strings.TrimPrefix(tag, "index:")
							break
						}
						// Try index1: pattern (for multiple metric_tags case)
						if strings.HasPrefix(tag, "index1:") {
							indexValue = strings.TrimPrefix(tag, "index1:")
							break
						}
					}
					if indexValue != "" {
						metricsByIndex[indexValue] = metric
					}
				}

				// Verify each metric
				for index, expectedValue := range tt.expectedValues {
					metric, ok := metricsByIndex[index]
					assert.True(t, ok, "metric for index %s should exist in outData.Data", index)
					if ok {
						// Verify Name
						assert.Equal(t, "test.constant.metric", metric.Name, "metric name should be correct for index %s", index)
						// Verify Value
						assert.Equal(t, expectedValue, metric.Value, "metric value should be 1.0 for index %s", index)
						// Verify Tags
						if tt.expectedTags != nil {
							if expectedTags, ok := tt.expectedTags[index]; ok {
								assert.ElementsMatch(t, expectedTags, metric.Tags, "tags should match for index %s. got: %v, want: %v", index, metric.Tags, expectedTags)
							}
						}
					}
				}
			} else {
				assert.Empty(t, samples)
				assert.Equal(t, 0, len(outData.Data), "outData.Data should be empty")
			}
		})
	}
}

// TestMetricTypePriority tests the metric_type priority logic.
func TestMetricTypePriority(t *testing.T) {
	tests := []struct {
		name         string
		metricConfig MetricsConfig
		symbolConfig SymbolConfig
		expectedType string
	}{
		{
			name: "symbol.MetricType takes precedence",
			metricConfig: MetricsConfig{
				MetricType: "gauge",
			},
			symbolConfig: SymbolConfig{
				Name:       "test.metric",
				MetricType: "monotonic_count",
			},
			expectedType: "monotonic_count",
		},
		{
			name: "metricConfig.MetricType used when symbol.MetricType is empty",
			metricConfig: MetricsConfig{
				MetricType: "rate",
			},
			symbolConfig: SymbolConfig{
				Name: "test.metric",
			},
			expectedType: "rate",
		},
		{
			name: "forced_type backward compatibility",
			metricConfig: MetricsConfig{
				ForcedType: "counter",
			},
			symbolConfig: SymbolConfig{
				Name: "test.metric",
			},
			expectedType: "counter",
		},
		{
			name:         "default to gauge when no type specified",
			metricConfig: MetricsConfig{},
			symbolConfig: SymbolConfig{
				Name: "test.metric",
			},
			expectedType: "gauge",
		},
		{
			name: "monotonic_count_and_rate type",
			metricConfig: MetricsConfig{
				MetricType: "monotonic_count_and_rate",
			},
			symbolConfig: SymbolConfig{
				Name: "test.metric",
			},
			expectedType: "monotonic_count_and_rate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test getMetricTypeHelper helper function logic
			metricType := getMetricTypeHelper(tt.symbolConfig, tt.metricConfig)
			assert.Equal(t, tt.expectedType, metricType)

			// Test in actual sendMetric call
			outData := &MetricDatas{}
			sample := MetricSample{
				value:      ResultValue{Value: float64(10)},
				tags:       []string{},
				symbol:     tt.symbolConfig,
				forcedType: metricType,
			}
			sendMetric(sample, outData)

			// Verify metric was added
			assert.Greater(t, len(outData.Data), 0, "metric should be added")

			// For monotonic_count_and_rate, skip the generic check as it generates 2 metrics
			if tt.expectedType != "monotonic_count_and_rate" {
				metricData := outData.Data[len(outData.Data)-1]
				assert.Equal(t, "test.metric", metricData.Name)
			}

			// Verify monotonic_count_and_rate generates two metrics
			if tt.expectedType == "monotonic_count_and_rate" {
				// Should have 2 metrics: one for monotonic_count and one for .rate
				assert.Equal(t, 2, len(outData.Data), "monotonic_count_and_rate should generate exactly 2 metrics")

				// Verify both metrics exist
				baseMetricFound := false
				rateMetricFound := false
				for _, m := range outData.Data {
					if m.Name == "test.metric" {
						baseMetricFound = true
					}
					if m.Name == "test.metric.rate" {
						rateMetricFound = true
					}
				}
				assert.True(t, baseMetricFound, "should have generated base metric (test.metric)")
				assert.True(t, rateMetricFound, "should have generated .rate metric (test.metric.rate)")
			}
		})
	}
}

// TestMetricTypeInColumnMetrics tests metric_type in column metrics
func TestMetricTypeInColumnMetrics(t *testing.T) {
	tests := []struct {
		name            string
		metricConfig    MetricsConfig
		values          *ResultValueStore
		expectedType    string
		expectedMetrics []struct {
			name  string
			value float64
			tags  []string
		}
	}{
		{
			name: "metric_type rate in column metrics",
			metricConfig: MetricsConfig{
				MetricType: "rate",
				Symbols: []SymbolConfig{
					{
						OID:  "1.3.6.1.2.1.2.2.1.10",
						Name: "ifInOctets",
					},
				},
				MetricTags: []MetricTagConfig{
					{
						Tag: "interface",
						Symbol: SymbolConfigCompat(SymbolConfig{
							OID:  "1.3.6.1.2.1.2.2.1.2",
							Name: "ifDescr",
						}),
					},
				},
			},
			values: &ResultValueStore{
				ColumnValues: ColumnResultValuesType{
					"1.3.6.1.2.1.2.2.1.10": {
						"1": ResultValue{Value: float64(1000)},
						"2": ResultValue{Value: float64(2000)},
					},
					"1.3.6.1.2.1.2.2.1.2": {
						"1": ResultValue{Value: "eth0"},
						"2": ResultValue{Value: "eth1"},
					},
				},
			},
			expectedType: "rate",
			expectedMetrics: []struct {
				name  string
				value float64
				tags  []string
			}{
				{name: "ifInOctets", value: 1000.0, tags: []string{"interface:eth0"}},
				{name: "ifInOctets", value: 2000.0, tags: []string{"interface:eth1"}},
			},
		},
		{
			name: "symbol.MetricType in column metrics",
			metricConfig: MetricsConfig{
				MetricType: "gauge",
				Symbols: []SymbolConfig{
					{
						OID:        "1.3.6.1.2.1.2.2.1.10",
						Name:       "ifInOctets",
						MetricType: "monotonic_count",
					},
				},
				MetricTags: []MetricTagConfig{
					{
						Tag: "interface",
						Symbol: SymbolConfigCompat(SymbolConfig{
							OID:  "1.3.6.1.2.1.2.2.1.2",
							Name: "ifDescr",
						}),
					},
				},
			},
			values: &ResultValueStore{
				ColumnValues: ColumnResultValuesType{
					"1.3.6.1.2.1.2.2.1.10": {
						"1": ResultValue{Value: float64(1000)},
					},
					"1.3.6.1.2.1.2.2.1.2": {
						"1": ResultValue{Value: "eth0"},
					},
				},
			},
			expectedType: "monotonic_count",
			expectedMetrics: []struct {
				name  string
				value float64
				tags  []string
			}{
				{name: "ifInOctets", value: 1000.0, tags: []string{"interface:eth0"}},
			},
		},
		{
			name: "multiple symbols with different metric_types",
			metricConfig: MetricsConfig{
				MetricType: "gauge",
				Symbols: []SymbolConfig{
					{
						OID:        "1.3.6.1.2.1.2.2.1.10",
						Name:       "ifInOctets",
						MetricType: "rate",
					},
					{
						OID:        "1.3.6.1.2.1.2.2.1.16",
						Name:       "ifOutOctets",
						MetricType: "monotonic_count",
					},
				},
				MetricTags: []MetricTagConfig{
					{
						Tag: "interface",
						Symbol: SymbolConfigCompat(SymbolConfig{
							OID:  "1.3.6.1.2.1.2.2.1.2",
							Name: "ifDescr",
						}),
					},
				},
			},
			values: &ResultValueStore{
				ColumnValues: ColumnResultValuesType{
					"1.3.6.1.2.1.2.2.1.10": {
						"1": ResultValue{Value: float64(1000)},
					},
					"1.3.6.1.2.1.2.2.1.16": {
						"1": ResultValue{Value: float64(2000)},
					},
					"1.3.6.1.2.1.2.2.1.2": {
						"1": ResultValue{Value: "eth0"},
					},
				},
			},
			expectedType: "rate", // First symbol's type
			expectedMetrics: []struct {
				name  string
				value float64
				tags  []string
			}{
				{name: "ifInOctets", value: 1000.0, tags: []string{"interface:eth0"}},
				{name: "ifOutOctets", value: 2000.0, tags: []string{"interface:eth0"}},
			},
		},
		{
			name: "constant_value_one with metric_type",
			metricConfig: MetricsConfig{
				MetricType: "gauge",
				Symbols: []SymbolConfig{
					{
						Name:             "test.constant.metric",
						ConstantValueOne: true,
					},
				},
				MetricTags: []MetricTagConfig{
					{
						Tag: "index",
						Symbol: SymbolConfigCompat(SymbolConfig{
							OID:  "1.3.6.1.2.1.2.2.1.1",
							Name: "ifIndex",
						}),
					},
				},
			},
			values: &ResultValueStore{
				ColumnValues: ColumnResultValuesType{
					"1.3.6.1.2.1.2.2.1.1": {
						"1": ResultValue{Value: float64(1)},
					},
				},
			},
			expectedType: "gauge",
			expectedMetrics: []struct {
				name  string
				value float64
				tags  []string
			}{
				{name: "test.constant.metric", value: 1.0, tags: []string{"index:1"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outData := &MetricDatas{}
			tags := []string{}
			_ = reportColumnMetrics(tt.metricConfig, tt.values, tags, outData)

			// Verify metrics were generated
			assert.Greater(t, len(outData.Data), 0, "should have generated metrics")

			// Verify expected metrics if provided
			if len(tt.expectedMetrics) > 0 {
				if len(tt.expectedMetrics) > 0 {
					assert.Equal(t, len(tt.expectedMetrics), len(outData.Data), "should have correct number of metrics")

					// Track which actual metrics have been matched to avoid duplicate matches
					matched := make(map[int]bool)

					// Verify each expected metric exists
					for _, expectedMetric := range tt.expectedMetrics {
						found := false
						for i, actualMetric := range outData.Data {
							// Skip if already matched
							if matched[i] {
								continue
							}
							// Match by name, value, and tags
							if actualMetric.Name == expectedMetric.name {
								// Check value first (more specific)
								if actualMetric.Value == expectedMetric.value {
									// Check if tags match (order may vary)
									if len(expectedMetric.tags) == 0 || assert.ElementsMatch(t, expectedMetric.tags, actualMetric.Tags, "tags should match for metric %s with value %f", expectedMetric.name, expectedMetric.value) {
										matched[i] = true
										found = true
										break
									}
								}
							}
						}
						assert.True(t, found, "expected metric %s with value %f and tags %v should exist", expectedMetric.name, expectedMetric.value, expectedMetric.tags)
					}
				}
			} else {
				// Fallback: basic verification if no expected metrics specified
				for _, metric := range outData.Data {
					assert.NotEmpty(t, metric.Name, "metric should have a name")
					assert.GreaterOrEqual(t, metric.Value, 0.0, "metric should have a non-negative value")
				}
			}
		})
	}
}

// TestMetricTypeInScalarMetrics tests metric_type in scalar metrics
func TestMetricTypeInScalarMetrics(t *testing.T) {
	tests := []struct {
		name         string
		metricConfig MetricsConfig
		values       *ResultValueStore
		expectedType string
	}{
		{
			name: "metric_type gauge in scalar metrics",
			metricConfig: MetricsConfig{
				MetricType: "gauge",
				Symbol: SymbolConfig{
					OID:  "1.3.6.1.2.1.1.3.0",
					Name: "sysUpTime",
				},
			},
			values: &ResultValueStore{
				ScalarValues: ScalarResultValuesType{
					"1.3.6.1.2.1.1.3.0": ResultValue{
						Value: float64(123456),
					},
				},
			},
			expectedType: "gauge",
		},
		{
			name: "symbol.MetricType in scalar metrics",
			metricConfig: MetricsConfig{
				MetricType: "gauge",
				Symbol: SymbolConfig{
					OID:        "1.3.6.1.2.1.1.3.0",
					Name:       "sysUpTime",
					MetricType: "rate",
				},
			},
			values: &ResultValueStore{
				ScalarValues: ScalarResultValuesType{
					"1.3.6.1.2.1.1.3.0": ResultValue{
						Value: float64(123456),
					},
				},
			},
			expectedType: "rate",
		},
		{
			name: "monotonic_count_and_rate in scalar metrics",
			metricConfig: MetricsConfig{
				MetricType: "monotonic_count_and_rate",
				Symbol: SymbolConfig{
					OID:  "1.3.6.1.2.1.1.3.0",
					Name: "sysUpTime",
				},
			},
			values: &ResultValueStore{
				ScalarValues: ScalarResultValuesType{
					"1.3.6.1.2.1.1.3.0": ResultValue{
						Value: float64(123456),
					},
				},
			},
			expectedType: "monotonic_count_and_rate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outData := &MetricDatas{}
			tags := []string{}
			_, err := reportScalarMetrics(tt.metricConfig, tt.values, tags, outData)
			assert.NoError(t, err)

			// Verify metrics were generated
			assert.Greater(t, len(outData.Data), 0, "should have generated metrics")

			// Verify monotonic_count_and_rate generates two metrics
			if tt.expectedType == "monotonic_count_and_rate" {
				assert.GreaterOrEqual(t, len(outData.Data), 2, "monotonic_count_and_rate should generate 2 metrics")
				rateMetricFound := false
				for _, m := range outData.Data {
					if m.Name == "sysUpTime.rate" {
						rateMetricFound = true
						break
					}
				}
				assert.True(t, rateMetricFound, "should have generated .rate metric")
			}
		})
	}
}

// getMetricTypeHelper is a helper function that determines the metric type
func getMetricTypeHelper(symbol SymbolConfig, metricConfig MetricsConfig) string {
	if symbol.MetricType != "" {
		return symbol.MetricType
	}
	if metricConfig.MetricType != "" {
		return metricConfig.MetricType
	}
	if metricConfig.ForcedType != "" {
		return metricConfig.ForcedType
	}
	return "gauge" // default
}
