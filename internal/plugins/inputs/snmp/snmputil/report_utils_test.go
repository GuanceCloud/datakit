// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func Test_getScalarValueFromSymbol(t *testing.T) {
	mockValues := &ResultValueStore{
		ScalarValues: map[string]ResultValue{
			"1.2.3.4": {Value: "value1"},
		},
	}

	tests := []struct {
		name          string
		values        *ResultValueStore
		symbol        SymbolConfig
		expectedValue ResultValue
		expectedError string
	}{
		{
			name:   "OK oid value case",
			values: mockValues,
			symbol: SymbolConfig{OID: "1.2.3.4", Name: "mySymbol"},
			expectedValue: ResultValue{
				Value: "value1",
			},
			expectedError: "",
		},
		{
			name:          "not found",
			values:        mockValues,
			symbol:        SymbolConfig{OID: "1.2.3.99", Name: "mySymbol"},
			expectedValue: ResultValue{},
			expectedError: "value for Scalar OID `1.2.3.99` not found in results",
		},
		{
			name:   "extract value pattern error",
			values: mockValues,
			symbol: SymbolConfig{
				OID:                  "1.2.3.4",
				Name:                 "mySymbol",
				ExtractValue:         "abc",
				ExtractValueCompiled: regexp.MustCompile("abc"),
			},
			expectedValue: ResultValue{},
			expectedError: "extract value extractValuePattern does not match (extractValuePattern=abc, srcValue=value1)",
		},
		{
			name:   "OK match pattern without replace",
			values: mockValues,
			symbol: SymbolConfig{
				OID:                  "1.2.3.4",
				Name:                 "mySymbol",
				MatchPatternCompiled: regexp.MustCompile("value\\d"), //nolint:gosimple
				MatchValue:           "matched-value-with-digit",
			},
			expectedValue: ResultValue{
				Value: "matched-value-with-digit",
			},
			expectedError: "",
		},
		{
			name:   "Error match pattern does not match",
			values: mockValues,
			symbol: SymbolConfig{
				OID:                  "1.2.3.4",
				Name:                 "mySymbol",
				MatchPattern:         "doesNotMatch",
				MatchPatternCompiled: regexp.MustCompile("doesNotMatch"),
				MatchValue:           "noMatch",
			},
			expectedValue: ResultValue{},
			expectedError: "match pattern `doesNotMatch` does not match string `value1`",
		},
		{
			name:   "Error match pattern template does not match",
			values: mockValues,
			symbol: SymbolConfig{
				OID:                  "1.2.3.4",
				Name:                 "mySymbol",
				MatchPattern:         "value(\\d)",
				MatchPatternCompiled: regexp.MustCompile("value(\\d)"), //nolint:gosimple
				MatchValue:           "$2",
			},
			expectedValue: ResultValue{},
			expectedError: "the pattern `value(\\d)` matched value `value1`, but template `$2` is not compatible",
		},
		{
			name:   "OK Extract value case",
			values: mockValues,
			symbol: SymbolConfig{
				OID:                  "1.2.3.4",
				Name:                 "mySymbol",
				ExtractValue:         "[a-z]+(\\d)",
				ExtractValueCompiled: regexp.MustCompile("[a-z]+(\\d)"), //nolint:gosimple
			},
			expectedValue: ResultValue{
				Value: "1",
			},
			expectedError: "",
		},
		{
			name:   "Error extract value pattern des not contain any matching group",
			values: mockValues,
			symbol: SymbolConfig{
				OID:                  "1.2.3.4",
				Name:                 "mySymbol",
				ExtractValue:         "[a-z]+\\d",
				ExtractValueCompiled: regexp.MustCompile("[a-z]+\\d"), //nolint:gosimple
			},
			expectedValue: ResultValue{},
			expectedError: "extract value pattern des not contain any matching group (extractValuePattern=[a-z]+\\d, srcValue=value1)",
		},
		{
			name:   "Error extract value extractValuePattern does not match",
			values: mockValues,
			symbol: SymbolConfig{
				OID:                  "1.2.3.4",
				Name:                 "mySymbol",
				ExtractValue:         "[a-z]+(\\d)",
				ExtractValueCompiled: regexp.MustCompile("doesNotMatch"),
			},
			expectedValue: ResultValue{},
			expectedError: "extract value extractValuePattern does not match (extractValuePattern=doesNotMatch, srcValue=value1)",
		},
		{
			name: "Formatter OK",
			values: &ResultValueStore{
				ScalarValues: map[string]ResultValue{
					"1.2.3.4": {
						Value: []byte{0x82, 0xa5, 0x6e, 0xa5, 0xc8, 0x01},
					},
				},
			},
			symbol: SymbolConfig{
				OID:    "1.2.3.4",
				Name:   "mySymbol",
				Format: "mac_address",
			},
			expectedValue: ResultValue{
				Value: "82:a5:6e:a5:c8:01",
			},
			expectedError: "",
		},
		{
			name: "Formatter Error",
			values: &ResultValueStore{
				ScalarValues: map[string]ResultValue{
					"1.2.3.4": {
						Value: []byte{0x82, 0xa5, 0x6e, 0xa5, 0xc8, 0x01},
					},
				},
			},
			symbol: SymbolConfig{
				OID:    "1.2.3.4",
				Name:   "mySymbol",
				Format: "unknown_format",
			},
			expectedError: "unknown format `unknown_format` (value type `[]uint8`)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualValues, err := getScalarValueFromSymbol(tt.values, tt.symbol)
			if err != nil || tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
			}
			assert.Equal(t, tt.expectedValue, actualValues)
		})
	}
}

func Test_getColumnValueFromSymbol(t *testing.T) {
	mockValues := &ResultValueStore{
		ColumnValues: map[string]map[string]ResultValue{
			"1.2.3.4": {
				"1": ResultValue{Value: "value1"},
				"2": ResultValue{Value: "value2"},
			},
		},
	}

	tests := []struct {
		name           string
		values         *ResultValueStore
		symbol         SymbolConfig
		expectedValues map[string]ResultValue
		expectedError  string
	}{
		{
			name:   "valid case",
			values: mockValues,
			symbol: SymbolConfig{OID: "1.2.3.4", Name: "mySymbol"},
			expectedValues: map[string]ResultValue{
				"1": {Value: "value1"},
				"2": {Value: "value2"},
			},
			expectedError: "",
		},
		{
			name:           "value not found",
			values:         mockValues,
			symbol:         SymbolConfig{OID: "1.2.3.99", Name: "mySymbol"},
			expectedValues: nil,
			expectedError:  "value for Column OID `1.2.3.99` not found in results",
		},
		{
			name:   "invalid extract value pattern",
			values: mockValues,
			symbol: SymbolConfig{
				OID:                  "1.2.3.4",
				Name:                 "mySymbol",
				ExtractValue:         "abc",
				ExtractValueCompiled: regexp.MustCompile("abc"),
			},
			expectedValues: make(map[string]ResultValue),
			expectedError:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualValues, err := getColumnValueFromSymbol(tt.values, tt.symbol)
			if err != nil || tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
			}
			assert.Equal(t, tt.expectedValues, actualValues)
		})
	}
}

func Test_transformIndex(t *testing.T) {
	tests := []struct {
		name               string
		indexes            []string
		transformRules     []MetricIndexTransform
		expectedNewIndexes []string
	}{
		{
			"no rule",
			[]string{"10", "11", "12", "13"},
			[]MetricIndexTransform{},
			nil,
		},
		{
			"one",
			[]string{"10", "11", "12", "13"},
			[]MetricIndexTransform{
				{Start: 2, End: 3},
			},
			[]string{"12", "13"},
		},
		{
			"multi",
			[]string{"10", "11", "12", "13"},
			[]MetricIndexTransform{
				{Start: 2, End: 2},
				{Start: 0, End: 1},
			},
			[]string{"12", "10", "11"},
		},
		{
			"out of index end",
			[]string{"10", "11", "12", "13"},
			[]MetricIndexTransform{
				{Start: 2, End: 1000},
			},
			nil,
		},
		{
			"out of index start and end",
			[]string{"10", "11", "12", "13"},
			[]MetricIndexTransform{
				{Start: 1000, End: 2000},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newIndexes := transformIndex(tt.indexes, tt.transformRules)
			assert.Equal(t, tt.expectedNewIndexes, newIndexes)
		})
	}
}

// go test -v -timeout 30s -run ^Test_getTagsFromMetricTagConfigList$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil
func Test_getTagsFromMetricTagConfigList(t *testing.T) {
	type logCount struct {
		log   string
		count int
	}
	tests := []struct {
		name            string
		rawMetricConfig []byte
		fullIndex       string
		values          *ResultValueStore
		expectedTags    []string
		expectedLogs    []logCount
	}{
		{
			name: "index transform",
			// language=yaml
			rawMetricConfig: []byte(`
table:
  OID:  1.2.3.4.5
  name: cpiPduBranchTable
symbols:
  - OID: 1.2.3.4.5.1.2
    name: cpiPduBranchCurrent
metric_tags:
  - column:
      OID:  1.2.3.4.8.1.2
      name: cpiPduName
    table: cpiPduTable
    index_transform:
      - start: 1
        end: 2
      - start: 6
        end: 7
    tag: pdu_name
`),
			fullIndex: "1.2.3.4.5.6.7.8",
			values: &ResultValueStore{
				ColumnValues: map[string]map[string]ResultValue{
					"1.2.3.4.8.1.2": {
						"2.3.7.8": ResultValue{
							Value: "myval",
						},
					},
				},
			},
			expectedTags: []string{"pdu_name:myval"},
		},
		{
			name: "index mapping",
			// language=yaml
			rawMetricConfig: []byte(`
table:
  OID: 1.3.6.1.2.1.4.31.3
  name: ipIfStatsTable
symbols:
  - OID: 1.3.6.1.2.1.4.31.3.1.6
    name: ipIfStatsHCInOctets
metric_tags:
  - index: 1
    tag: ipversion
    mapping:
      0: unknown
      1: ipv4
      2: ipv6
      3: ipv4z
      4: ipv6z
      16: dns
`),
			fullIndex:    "3",
			values:       &ResultValueStore{},
			expectedTags: []string{"ipversion:ipv4z"},
		},
		{
			name: "regex match",
			// language=yaml
			rawMetricConfig: []byte(`
table:
  OID:  1.2.3.4.5
  name: cpiPduBranchTable
symbols:
  - OID: 1.2.3.4.5.1.2
    name: cpiPduBranchCurrent
metric_tags:
  - column:
      OID:  1.2.3.4.8.1.2
      name: cpiPduName
    table: cpiPduTable
    match: '(\w)(\w+)'
    tags:
      prefix: '$1'
      suffix: '$2'
`),
			fullIndex: "1.2.3.4.5.6.7.8",
			values: &ResultValueStore{
				ColumnValues: map[string]map[string]ResultValue{
					"1.2.3.4.8.1.2": {
						"1.2.3.4.5.6.7.8": ResultValue{
							Value: "eth0",
						},
					},
				},
			},
			expectedTags: []string{"prefix:e", "suffix:th0"},
		},
		{
			name: "regex match only once",
			// language=yaml
			rawMetricConfig: []byte(`
table:
  OID:  1.2.3.4.5
  name: cpiPduBranchTable
symbols:
  - OID: 1.2.3.4.5.1.2
    name: cpiPduBranchCurrent
metric_tags:
  - column:
      OID:  1.2.3.4.8.1.2
      name: cpiPduName
    table: cpiPduTable
    match: '([A-z0-9]*)-([A-z]*[-A-z]*)-([A-z0-9]*)'
    tags:
      tag1: '${1}'
      tag2: '\1'
`),
			fullIndex: "1.2.3.4.5.6.7.8",
			values: &ResultValueStore{
				ColumnValues: map[string]map[string]ResultValue{
					"1.2.3.4.8.1.2": {
						"1.2.3.4.5.6.7.8": ResultValue{
							Value: "f5-vm-aa.c.-integrations-lab.internal",
						},
					},
				},
			},
			expectedTags: []string{"tag1:f5", "tag2:f5"},
		},
		{
			name: "regex does not match",
			// language=yaml
			rawMetricConfig: []byte(`
table:
  OID:  1.2.3.4.5
  name: cpiPduBranchTable
symbols:
  - OID: 1.2.3.4.5.1.2
    name: cpiPduBranchCurrent
metric_tags:
  - column:
      OID:  1.2.3.4.8.1.2
      name: cpiPduName
    table: cpiPduTable
    match: '(\w)(\w+)'
    tags:
      prefix: '$1'
      suffix: '$2'
`),
			fullIndex: "1.2.3.4.5.6.7.8",
			values: &ResultValueStore{
				ColumnValues: map[string]map[string]ResultValue{
					"1.2.3.4.8.1.2": {
						"1.2.3.4.5.6.7.8": ResultValue{
							Value: "....",
						},
					},
				},
			},
			expectedTags: []string(nil),
		},
		{
			name: "regex does not match exact",
			// language=yaml
			rawMetricConfig: []byte(`
table:
  OID:  1.2.3.4.5
  name: cpiPduBranchTable
symbols:
  - OID: 1.2.3.4.5.1.2
    name: cpiPduBranchCurrent
metric_tags:
  - column:
      OID:  1.2.3.4.8.1.2
      name: cpiPduName
    table: cpiPduTable
    match: '^(\w)(\w+)$'
    tags:
      prefix: '$1'
      suffix: '$2'
`),
			fullIndex: "1.2.3.4.5.6.7.8",
			values: &ResultValueStore{
				ColumnValues: map[string]map[string]ResultValue{
					"1.2.3.4.8.1.2": {
						"1.2.3.4.5.6.7.8": ResultValue{
							Value: "abc.",
						},
					},
				},
			},
			expectedTags: []string(nil),
		},
		{
			name: "missing index value",
			// language=yaml
			rawMetricConfig: []byte(`
table:
  OID:  1.2.3.4.5
  name: cpiPduBranchTable
symbols:
  - OID: 1.2.3.4.5.1.2
    name: cpiPduBranchCurrent
metric_tags:
  - column:
      OID:  1.2.3.4.8.1.2
      name: cpiPduName
    table: cpiPduTable
    tag: abc
`),
			fullIndex: "1.2.3.4.5.6.7.8",
			values: &ResultValueStore{
				ColumnValues: map[string]map[string]ResultValue{
					"1.2.3.4.8.1.2": {
						"999": ResultValue{
							Value: "abc.",
						},
					},
				},
			},
			expectedTags: []string(nil),
			expectedLogs: []logCount{
				{"[DEBUG] getTagsFromMetricTagConfigList: index not found for column value: tag=abc, index=1.2.3.4.5.6.7.8", 1},
			},
		},
		{
			name: "error converting tag value",
			// language=yaml
			rawMetricConfig: []byte(`
table:
  OID:  1.2.3.4.5
  name: cpiPduBranchTable
symbols:
  - OID: 1.2.3.4.5.1.2
    name: cpiPduBranchCurrent
metric_tags:
  - column:
      OID:  1.2.3.4.8.1.2
      name: cpiPduName
    table: cpiPduTable
    tag: abc
`),
			fullIndex: "1.2.3.4.5.6.7.8",
			values: &ResultValueStore{
				ColumnValues: map[string]map[string]ResultValue{
					"1.2.3.4.8.1.2": {
						"1.2.3.4.5.6.7.8": ResultValue{
							Value: ResultValue{},
						},
					},
				},
			},
			expectedTags: []string(nil),
			expectedLogs: []logCount{
				{"[DEBUG] getTagsFromMetricTagConfigList: error converting tagValue", 1},
			},
		},
		{
			name: "missing column value",
			// language=yaml
			rawMetricConfig: []byte(`
table:
  OID:  1.2.3.4.5
  name: cpiPduBranchTable
symbols:
  - OID: 1.2.3.4.5.1.2
    name: cpiPduBranchCurrent
metric_tags:
  - column:
      OID:  1.2.3.4.8.1.2
      name: cpiPduName
    table: cpiPduTable
    tag: abc
`),
			fullIndex: "1.2.3.4.5.6.7.8",
			values: &ResultValueStore{
				ColumnValues: map[string]map[string]ResultValue{
					"999": {
						"1.2.3.4.5.6.7.8": ResultValue{
							Value: "abc.",
						},
					},
				},
			},
			expectedTags: []string(nil),
			expectedLogs: []logCount{
				{"[DEBUG] getTagsFromMetricTagConfigList: error getting column value: value for Column OID `1.2.3.4.8.1.2`", 1},
			},
		},
		{
			name: "mapping does not exist",
			// language=yaml
			rawMetricConfig: []byte(`
table:
  OID:  1.2.3.4.5
  name: cpiPduBranchTable
symbols:
  - OID: 1.2.3.4.5.1.2
    name: cpiPduBranchCurrent
metric_tags:
  - index: 1
    tag: abc
    mapping:
      0: unknown
      1: ipv4
      2: ipv6
      3: ipv4z
      4: ipv6z
      16: dns
`),
			fullIndex: "20",
			values: &ResultValueStore{
				ColumnValues: map[string]map[string]ResultValue{
					"1.2.3.4.8.1.2": {
						"20": ResultValue{
							Value: "abc.",
						},
					},
				},
			},
			expectedTags: []string(nil),
			expectedLogs: []logCount{
				{"[DEBUG] getTagsFromMetricTagConfigList: error getting tags. mapping for `20` does not exist.", 1},
			},
		},
		{
			name: "index not found",
			// language=yaml
			rawMetricConfig: []byte(`
table:
  OID:  1.2.3.4.5
  name: cpiPduBranchTable
symbols:
  - OID: 1.2.3.4.5.1.2
    name: cpiPduBranchCurrent
metric_tags:
  - index: 100
    tag: abc
`),
			fullIndex: "1",
			values: &ResultValueStore{
				ColumnValues: map[string]map[string]ResultValue{
					"1.2.3.4.8.1.2": {
						"1": ResultValue{
							Value: "abc.",
						},
					},
				},
			},
			expectedTags: []string(nil),
			expectedLogs: []logCount{
				{"[DEBUG] getTagsFromMetricTagConfigList: error getting tags. index `100` not found in indexes `[1]`", 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b bytes.Buffer
			w := bufio.NewWriter(&b)

			m := MetricsConfig{}
			yaml.Unmarshal(tt.rawMetricConfig, &m)

			ValidateEnrichMetrics([]MetricsConfig{m})
			tags := getTagsFromMetricTagConfigList(m.MetricTags, tt.fullIndex, tt.values)

			assert.ElementsMatch(t, tt.expectedTags, tags)

			w.Flush()
			// logs := b.String()
			var logs string
			switch tt.name {
			case "index transform":
				logs = ""
			case "index mapping":
				logs = ""
			case "regex match":
				logs = ""
			case "regex match only once":
				logs = ""
			case "regex does not match":
				logs = ""
			case "regex does not match exact":
				logs = ""
			case "missing index value":
				logs = "[DEBUG] getTagsFromMetricTagConfigList: index not found for column value: tag=abc, index=1.2.3.4.5.6.7.8"
			case "error converting tag value":
				logs = "[DEBUG] getTagsFromMetricTagConfigList: error converting tagValue (valuestore.ResultValue{SubmissionType:\"\", Value:valuestore.ResultValue{SubmissionType:\"\", Value:interface {}(nil)}}) to string : invalid type valuestore.ResultValue for value valuestore.ResultValue{SubmissionType:\"\", Value:interface {}(nil)}"
			case "missing column value":
				logs = "[DEBUG] getTagsFromMetricTagConfigList: error getting column value: value for Column OID `1.2.3.4.8.1.2` not found in results"
			case "mapping does not exist":
				logs = "[DEBUG] getTagsFromMetricTagConfigList: error getting tags. mapping for `20` does not exist. mapping=`map[0:unknown 1:ipv4 16:dns 2:ipv6 3:ipv4z 4:ipv6z]`, indexes=`[20]`"
			case "index not found":
				logs = "[DEBUG] getTagsFromMetricTagConfigList: error getting tags. index `100` not found in indexes `[1]`"
			}

			for _, aLogCount := range tt.expectedLogs {
				assert.Equal(t, aLogCount.count, strings.Count(logs, aLogCount.log), logs)
			}
		})
	}
}
