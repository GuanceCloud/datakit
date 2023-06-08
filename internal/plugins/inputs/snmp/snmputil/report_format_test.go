// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_formatValue(t *testing.T) {
	tests := []struct {
		name          string
		value         ResultValue
		format        string
		expectedValue ResultValue
		expectedError string
	}{
		{
			name: "format mac address",
			value: ResultValue{
				Value: []byte{0x82, 0xa5, 0x6e, 0xa5, 0xc8, 0x01},
			},
			format: "mac_address",
			expectedValue: ResultValue{
				Value: "82:a5:6e:a5:c8:01",
			},
		},
		{
			name: "error unknown value type",
			value: ResultValue{
				Value: ResultValue{},
			},
			format:        "mac_address",
			expectedError: "value type `snmputil.ResultValue` not supported (format `mac_address`)",
		},
		{
			name: "error unknown format type",
			value: ResultValue{
				Value: []byte{0x82, 0xa5, 0x6e, 0xa5, 0xc8, 0x01},
			},
			format:        "unknown_format",
			expectedError: "unknown format `unknown_format` (value type `[]uint8`)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := formatValue(tt.value, tt.format)
			assert.Equal(t, tt.expectedValue, value)
			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
			}
		})
	}
}
