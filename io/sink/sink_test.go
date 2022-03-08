package sink

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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
					"id":             "influxdb_1",
					"target":         "influxdb",
					"addr":           "http://10.200.7.21:8086",
					"precision":      "ns",
					"database":       "db0",
					"user_agent":     "go_test_client",
					"timeout":        "6s",
					"write_encoding": "",
				},
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
			name: "categories_empty",
			in: []map[string]interface{}{
				{
					"categories": []string{},
				},
			},
			expectError: fmt.Errorf("invalid categories: empty"),
		},
		{
			name: "categories_not_[]string",
			in: []map[string]interface{}{
				{
					"categories": "",
				},
			},
			expectError: fmt.Errorf("invalid categories: not []string"),
		},
		{
			name: "invalid_id",
			in: []map[string]interface{}{
				{
					"id":         123,
					"categories": []string{"M"},
				},
			},
			expectError: fmt.Errorf("invalid id: not string"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := aggregationCategorys(tc.in)
			assert.Equal(t, tc.expectError, err)
		})
	}
}
