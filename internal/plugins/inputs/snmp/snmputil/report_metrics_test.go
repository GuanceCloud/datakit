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
