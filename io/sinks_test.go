package io

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v -timeout 30s -run ^TestCheckSinksConfig$ gitlab.jiagouyun.com/cloudcare-tools/datakit/io
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
			expectError: fmt.Errorf("invalid id: empty"),
		},
		{
			name: "id_empty_2",
			in: []map[string]interface{}{
				{"id": ""},
			},
			expectError: fmt.Errorf("invalid id: empty"),
		},
		{
			name: "id_empty_3",
			in: []map[string]interface{}{
				{"id": "  "},
			},
			expectError: fmt.Errorf("invalid id: empty"),
		},
		{
			name: "id_repeat",
			in: []map[string]interface{}{
				{"id": "abc"},
				{"id": "bcd"},
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
			err := checkSinksConfig(tc.in)
			assert.Equal(t, tc.expectError, err)
		})
	}
}
