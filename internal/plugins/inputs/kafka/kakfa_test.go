// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kafka

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseMbeanName(t *testing.T) {
	cases := []struct {
		name     string
		mbean    string
		domain   string
		propsLen int
		props    map[string]string
	}{
		{
			name:     "simple domain",
			mbean:    "kafka.server",
			domain:   "kafka.server",
			propsLen: 0,
		},
		{
			name:     "with properties",
			mbean:    "kafka.server:type=BrokerTopicMetrics,name=MessagesInPerSec",
			domain:   "kafka.server",
			propsLen: 2,
			props: map[string]string{
				"type": "BrokerTopicMetrics",
				"name": "MessagesInPerSec",
			},
		},
		{
			name:     "with special characters",
			mbean:    `kafka.server:type=Test,name="test-value"`,
			domain:   "kafka.server",
			propsLen: 2,
			props: map[string]string{
				"type": "Test",
				"name": `"test_value"`, // - replaced with _, quotes kept
			},
		},
		{
			name:     "with multiple properties",
			mbean:    "kafka.log:name=Size,topic=test-topic,partition=0,type=Log",
			domain:   "kafka.log",
			propsLen: 4,
			props: map[string]string{
				"name":      "Size",
				"topic":     "test_topic", // - replaced with _
				"partition": "0",
				"type":      "Log",
			},
		},
		{
			name:     "empty properties",
			mbean:    "kafka.server:",
			domain:   "kafka.server",
			propsLen: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			domain, props := parseMbeanName(tc.mbean)
			assert.Equal(t, tc.domain, domain)
			assert.Equal(t, tc.propsLen, len(props))
			for k, v := range tc.props {
				assert.Equal(t, v, props[k], "property %s", k)
			}
		})
	}
}

func Test_normalizeValue(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normal value",
			input: "MessagesInPerSec",
			want:  "MessagesInPerSec",
		},
		{
			name:  "with dash",
			input: "test-value",
			want:  "test_value",
		},
		{
			name:  "with quotes",
			input: `"test-value"`,
			want:  `"test_value"`, // only - replaced with _, quotes kept
		},
		{
			name:  "with special chars",
			input: "test/value:name*[test]",
			want:  "test/value:name*[test]", // only - replaced with _, other special chars kept
		},
		{
			name:  "with space",
			input: "test value",
			want:  "test value", // only - replaced with _, space kept
		},
		{
			name:  "with pipe",
			input: "test|value",
			want:  "test|value", // only - replaced with _, pipe kept
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeValue(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_buildFieldKey(t *testing.T) {
	cases := []struct {
		name   string
		domain string
		props  map[string]string
		attr   string
		want   string
	}{
		{
			name:   "full format",
			domain: "kafka.server",
			props: map[string]string{
				"type": "BrokerTopicMetrics",
				"name": "MessagesInPerSec",
			},
			attr: "Count",
			want: "kafka.server.BrokerTopicMetrics.MessagesInPerSec.Count",
		},
		{
			name:   "without type",
			domain: "kafka.server",
			props: map[string]string{
				"name": "MessagesInPerSec",
			},
			attr: "Count",
			want: "kafka.server.MessagesInPerSec.Count",
		},
		{
			name:   "without name",
			domain: "kafka.server",
			props: map[string]string{
				"type": "BrokerTopicMetrics",
			},
			attr: "Count",
			want: "kafka.server.BrokerTopicMetrics.Count",
		},
		{
			name:   "only domain and attr",
			domain: "kafka.server",
			props:  map[string]string{},
			attr:   "Count",
			want:   "kafka.server.Count",
		},
		{
			name:   "with special chars in attr",
			domain: "kafka.server",
			props: map[string]string{
				"type": "Test",
				"name": "test",
			},
			attr: "test-value",
			want: "kafka.server.Test.test.test_value",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildFieldKey(tc.domain, tc.props, tc.attr)
			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_convertValue(t *testing.T) {
	cases := []struct {
		name  string
		input interface{}
		want  interface{}
	}{
		{
			name:  "json.Number int",
			input: json.Number("123"),
			want:  int64(123),
		},
		{
			name:  "json.Number float",
			input: json.Number("123.45"),
			want:  123.45,
		},
		{
			name:  "string",
			input: "test",
			want:  nil, // should skip
		},
		{
			name:  "bool",
			input: true,
			want:  nil, // should skip
		},
		{
			name:  "map",
			input: map[string]interface{}{"key": "value"},
			want:  nil, // should skip
		},
		{
			name:  "array",
			input: []interface{}{1, 2, 3},
			want:  nil, // should skip
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := convertValue(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_extractFieldsFromValue(t *testing.T) {
	cases := []struct {
		name   string
		domain string
		props  map[string]string
		value  interface{}
		want   map[string]interface{}
	}{
		{
			name:   "map with numeric values",
			domain: "kafka.server",
			props: map[string]string{
				"type": "BrokerTopicMetrics",
				"name": "MessagesInPerSec",
			},
			value: map[string]interface{}{
				"Count":    json.Number("100"),
				"MeanRate": json.Number("10.5"),
			},
			want: map[string]interface{}{
				"kafka.server.BrokerTopicMetrics.MessagesInPerSec.Count":    int64(100),
				"kafka.server.BrokerTopicMetrics.MessagesInPerSec.MeanRate": 10.5,
			},
		},
		{
			name:   "map with non-numeric values",
			domain: "kafka.server",
			props:  map[string]string{"type": "Test"},
			value: map[string]interface{}{
				"Count":  json.Number("100"),
				"String": "test",
				"Bool":   true,
			},
			want: map[string]interface{}{
				"kafka.server.Test.Count": int64(100),
			},
		},
		{
			name:   "map with array",
			domain: "kafka.server",
			props:  map[string]string{"type": "Test"},
			value: map[string]interface{}{
				"Count": json.Number("100"),
				"Array": []interface{}{1, 2, 3},
			},
			want: map[string]interface{}{
				"kafka.server.Test.Count": int64(100),
			},
		},
		{
			name:   "non-map value",
			domain: "kafka.server",
			props:  map[string]string{"type": "Test"},
			value:  json.Number("100"),
			want: map[string]interface{}{
				"kafka.server.Test.value": int64(100),
			},
		},
		{
			name:   "non-map non-numeric",
			domain: "kafka.server",
			props:  map[string]string{"type": "Test"},
			value:  "test",
			want:   map[string]interface{}{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractFieldsFromValue(tc.value, tc.domain, tc.props)
			assert.Equal(t, len(tc.want), len(got))
			for k, v := range tc.want {
				assert.Equal(t, v, got[k], "field %s", k)
			}
		})
	}
}

func Test_isMBeanBlacklisted(t *testing.T) {
	cases := []struct {
		name      string
		blacklist []string
		mbean     string
		want      bool
	}{
		{
			name:      "empty blacklist",
			blacklist: []string{},
			mbean:     "kafka.server:type=Test",
			want:      false,
		},
		{
			name:      "exact match",
			blacklist: []string{"kafka.server:type=Test"},
			mbean:     "kafka.server:type=Test",
			want:      true,
		},
		{
			name:      "no match",
			blacklist: []string{"kafka.server:type=Test"},
			mbean:     "kafka.server:type=Other",
			want:      false,
		},
		{
			name:      "wildcard match all",
			blacklist: []string{"*.*:*"},
			mbean:     "kafka.server:type=Test",
			want:      true,
		},
		{
			name:      "wildcard match domain",
			blacklist: []string{"kafka.log:*"},
			mbean:     "kafka.log:type=Test",
			want:      true,
		},
		{
			name:      "wildcard match domain no match",
			blacklist: []string{"kafka.log:*"},
			mbean:     "kafka.server:type=Test",
			want:      false,
		},
		{
			name:      "wildcard match type",
			blacklist: []string{"kafka.server:type=BrokerTopicMetrics,*"},
			mbean:     "kafka.server:type=BrokerTopicMetrics,name=MessagesInPerSec",
			want:      true,
		},
		{
			name:      "wildcard match type no match",
			blacklist: []string{"kafka.server:type=BrokerTopicMetrics,*"},
			mbean:     "kafka.server:type=Other,name=MessagesInPerSec",
			want:      false,
		},
		{
			name:      "multiple patterns match first",
			blacklist: []string{"kafka.log:*", "kafka.server:type=Test"},
			mbean:     "kafka.log:type=Test",
			want:      true,
		},
		{
			name:      "multiple patterns match second",
			blacklist: []string{"kafka.log:*", "kafka.server:type=Test"},
			mbean:     "kafka.server:type=Test",
			want:      true,
		},
		{
			name:      "multiple patterns no match",
			blacklist: []string{"kafka.log:*", "kafka.server:type=Test"},
			mbean:     "kafka.server:type=Other",
			want:      false,
		},
		{
			name:      "wildcard with question mark",
			blacklist: []string{"kafka.server:type=Test?,*"},
			mbean:     "kafka.server:type=Test1,name=test",
			want:      true,
		},
		{
			name:      "wildcard with question mark no match",
			blacklist: []string{"kafka.server:type=Test?,*"},
			mbean:     "kafka.server:type=Test12,name=test",
			want:      false,
		},
		{
			name:      "invalid pattern silently ignored",
			blacklist: []string{"[invalid", "kafka.server:type=Test"},
			mbean:     "kafka.server:type=Test",
			want:      true, // should match the valid pattern
		},
		{
			name:      "invalid pattern no match",
			blacklist: []string{"[invalid"},
			mbean:     "kafka.server:type=Test",
			want:      false, // invalid pattern is ignored
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ipt := &Input{
				MBeanBlacklist: tc.blacklist,
			}
			got := ipt.isMBeanBlacklisted(tc.mbean)
			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_filterMBeansByBlacklist(t *testing.T) {
	cases := []struct {
		name      string
		blacklist []string
		mbeans    []string
		want      []string
	}{
		{
			name:      "empty blacklist",
			blacklist: []string{},
			mbeans:    []string{"kafka.server:type=Test1", "kafka.server:type=Test2"},
			want:      []string{"kafka.server:type=Test1", "kafka.server:type=Test2"},
		},
		{
			name:      "filter all",
			blacklist: []string{"*.*:*"},
			mbeans:    []string{"kafka.server:type=Test1", "kafka.server:type=Test2"},
			want:      []string{},
		},
		{
			name:      "filter by domain",
			blacklist: []string{"kafka.log:*"},
			mbeans: []string{
				"kafka.server:type=Test1",
				"kafka.log:type=Test2",
				"kafka.server:type=Test3",
			},
			want: []string{
				"kafka.server:type=Test1",
				"kafka.server:type=Test3",
			},
		},
		{
			name:      "filter by type",
			blacklist: []string{"kafka.server:type=BrokerTopicMetrics,*"},
			mbeans: []string{
				"kafka.server:type=BrokerTopicMetrics,name=Test1",
				"kafka.server:type=Other,name=Test2",
				"kafka.server:type=BrokerTopicMetrics,name=Test3",
			},
			want: []string{
				"kafka.server:type=Other,name=Test2",
			},
		},
		{
			name:      "multiple patterns",
			blacklist: []string{"kafka.log:*", "kafka.server:type=Test1,*"},
			mbeans: []string{
				"kafka.server:type=Test1,name=test",
				"kafka.log:type=Test2",
				"kafka.server:type=Test3,name=test",
			},
			want: []string{
				"kafka.server:type=Test3,name=test",
			},
		},
		{
			name:      "no matches",
			blacklist: []string{"kafka.log:*"},
			mbeans: []string{
				"kafka.server:type=Test1",
				"kafka.server:type=Test2",
			},
			want: []string{
				"kafka.server:type=Test1",
				"kafka.server:type=Test2",
			},
		},
		{
			name:      "empty mbeans",
			blacklist: []string{"kafka.log:*"},
			mbeans:    []string{},
			want:      []string{},
		},
		{
			name:      "filter with question mark wildcard",
			blacklist: []string{"kafka.server:type=Test?,*"},
			mbeans: []string{
				"kafka.server:type=Test1,name=test",
				"kafka.server:type=Test2,name=test",
				"kafka.server:type=Test10,name=test",
				"kafka.server:type=Other,name=test",
			},
			want: []string{
				"kafka.server:type=Test10,name=test",
				"kafka.server:type=Other,name=test",
			},
		},
		{
			name:      "invalid pattern in blacklist",
			blacklist: []string{"[invalid", "kafka.log:*"},
			mbeans: []string{
				"kafka.server:type=Test1",
				"kafka.log:type=Test2",
			},
			want: []string{
				"kafka.server:type=Test1",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ipt := &Input{
				MBeanBlacklist: tc.blacklist,
			}
			got := ipt.filterMBeansByBlacklist(tc.mbeans)
			assert.Equal(t, tc.want, got)
		})
	}
}
