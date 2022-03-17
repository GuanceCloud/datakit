// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sink

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
)

// go test -v -timeout 30s -run ^TestCheckSinksConfig$ gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink
func TestCheckSinksConfig(t *testing.T) {
	cases := []struct {
		name        string
		in          []map[string]interface{}
		expectError error
	}{
		{
			name: "id_unique",
			in: []map[string]interface{}{
				{"id": "abc"},
				{"id": "bcd"},
				{"id": "efg"},
			},
		},
		{
			name: "empty",
			in: []map[string]interface{}{
				{},
			},
		},
		{
			name: "id_empty_1",
			in: []map[string]interface{}{
				{"id": " "},
			},
			expectError: fmt.Errorf("%s could not be empty", "id"),
		},
		{
			name: "id_empty_2",
			in: []map[string]interface{}{
				{"id": ""},
			},
			expectError: fmt.Errorf("%s could not be empty", "id"),
		},
		{
			name: "id_empty_3",
			in: []map[string]interface{}{
				{"id": "  "},
			},
			expectError: fmt.Errorf("%s could not be empty", "id"),
		},
		{
			name: "id_repeat",
			in: []map[string]interface{}{
				{"id": "abc"},
				{"id": "abc"},
			},
			expectError: fmt.Errorf("invalid sink config: id not unique"),
		},
		{
			name: "id_digit",
			in: []map[string]interface{}{
				{"id": 123},
			},
			expectError: fmt.Errorf("invalid id: not string"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkSinkConfig(tc.in)
			assert.Equal(t, tc.expectError, err)
		})
	}
}

// go test -v -timeout 30s -run ^TestBuildSinkImpls$ gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink
func TestBuildSinkImpls(t *testing.T) {
	cases := []struct {
		name        string
		in          []map[string]interface{}
		expectError error
	}{
		{
			name: "normal",
			in: []map[string]interface{}{
				{
					"id":         "influxdb_1",
					"target":     "influxdb",
					"addr":       "http://1.1.1.1:8086",
					"precision":  "ns",
					"database":   "db0",
					"user_agent": "go_test_client",
					"timeout":    "6s",
				},
			},
		},
		{
			name: "empty",
			in: []map[string]interface{}{
				{},
			},
		},
		{
			name: "invaid_target",
			in: []map[string]interface{}{
				{
					"target": "influxdb1",
				},
			},
			expectError: fmt.Errorf("%s not implemented yet", "influxdb1"),
		},
		{
			name: "invaid_target_type",
			in: []map[string]interface{}{
				{
					"target": 123,
				},
			},
			expectError: fmt.Errorf("invalid %s: not string", "target"),
		},
		{
			name: "example",
			in: []map[string]interface{}{
				{
					"target": datakit.SinkTargetExample,
				},
			},
		},
		{
			name: "id_empty",
			in: []map[string]interface{}{
				{
					"target": "influxdb",
				},
			},
			expectError: fmt.Errorf("%s could not be empty", "id"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := buildSinkImpls(tc.in)
			assert.Equal(t, tc.expectError, err)
		})
	}
}

// go test -v -timeout 30s -run ^TestAggregationCategorys$ gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink
func TestAggregationCategorys(t *testing.T) {
	cases := []struct {
		name        string
		in          []map[string]interface{}
		expectError error
	}{
		{
			name: "normal",
			in: []map[string]interface{}{
				{
					"id":         "influxdb_1",
					"target":     "influxdb",
					"addr":       "http://1.1.1.1:8086",
					"precision":  "ns",
					"database":   "db0",
					"user_agent": "go_test_client",
					"timeout":    "6s",
					"categories": []interface{}{"M"},
				},
			},
		},
		{
			name: "categories_empty",
			in: []map[string]interface{}{
				{
					"categories": []interface{}{},
				},
			},
			expectError: fmt.Errorf("invalid categories: empty"),
		},
		{
			name: "categories_not_[]interface{}",
			in: []map[string]interface{}{
				{
					"categories": "",
				},
			},
			expectError: fmt.Errorf("invalid categories: not []interface{}: %#v", ""),
		},
		{
			name: "invalid_id",
			in: []map[string]interface{}{
				{
					"id":         123,
					"categories": []interface{}{"M"},
				},
			},
			expectError: fmt.Errorf("invalid id: not string"),
		},
		{
			name: "empty",
			in: []map[string]interface{}{
				{},
			},
		},
		{
			name: "no_categories",
			in: []map[string]interface{}{
				{
					"id": 123,
				},
			},
			expectError: fmt.Errorf("invalid categories: not found"),
		},
		{
			name: "invalid_icategories_not_string",
			in: []map[string]interface{}{
				{
					"categories": []interface{}{123},
				},
			},
			expectError: fmt.Errorf("invalid categories: not string"),
		},
		{
			name: "unrecognized category",
			in: []map[string]interface{}{
				{
					"id":         "influxdb_1",
					"target":     "influxdb",
					"addr":       "http://1.1.1.1:8086",
					"precision":  "ns",
					"database":   "db0",
					"user_agent": "go_test_client",
					"timeout":    "6s",
					"categories": []interface{}{"M1"},
				},
			},
			expectError: fmt.Errorf("unrecognized category"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if len(sinkcommon.SinkImpls) == 0 {
				TestBuildSinkImpls(t)
			}

			err := aggregationCategorys(tc.in)
			assert.Equal(t, tc.expectError, err)
		})
	}
}

// go test -v -timeout 30s -run ^TestGetMapCategory$ gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink
func TestGetMapCategory(t *testing.T) {
	cases := []struct {
		name        string
		in          []string
		out         []string
		expectError error
	}{
		{
			name: "normal",
			in: []string{
				datakit.SinkCategoryMetric,
				datakit.SinkCategoryNetwork,
				datakit.SinkCategoryKeyEvent,
				datakit.SinkCategoryObject,
				datakit.SinkCategoryCustomObject,
				datakit.SinkCategoryLogging,
				datakit.SinkCategoryTracing,
				datakit.SinkCategoryRUM,
				datakit.SinkCategorySecurity,
			},
			out: []string{
				datakit.Metric,
				datakit.Network,
				datakit.KeyEvent,
				datakit.Object,
				datakit.CustomObject,
				datakit.Logging,
				datakit.Tracing,
				datakit.RUM,
				datakit.Security,
			},
		},
		{
			name:        "unrecognized category",
			in:          []string{datakit.MetricDeprecated},
			out:         []string{""},
			expectError: fmt.Errorf("unrecognized category"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for k, str := range tc.in {
				newCategory, err := getMapCategory(str)
				assert.Equal(t, tc.expectError, err)
				assert.Equal(t, tc.out[k], newCategory)
			}
		})
	}
}
