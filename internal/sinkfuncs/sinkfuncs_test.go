// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sinkfuncs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v -timeout 30s -run ^TestParseSinkSingle$ gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/installer
func TestParseSinkSingle(t *testing.T) {
	cases := []struct {
		name        string
		in          string
		out         map[string]interface{}
		expectError error
	}{
		{
			name: "normal",
			in:   "influxdb://1.1.1.1:8086?protocol=http&database=db0&timeout=15s",
			out: map[string]interface{}{
				"target":   "influxdb",
				"protocol": "http",
				"host":     "1.1.1.1:8086",
				"database": "db0",
				"timeout":  "15s",
			},
		},
		{
			name: "dataway",
			in:   "dataway://?url=https://openway.guance.com&token=tkn_xxxxx&proxy=127.0.0.1:1080",
			out: map[string]interface{}{
				"target": "dataway",
				"url":    "https://openway.guance.com",
				"token":  "tkn_xxxxx",
				"proxy":  "127.0.0.1:1080",
			},
		},
		{
			name: "dataway filters",
			in:   "dataway://?url=https://openway.guance.com&token=tkn_xxxxx&filters={interface='en0'}&filters={abc='def'}",
			out: map[string]interface{}{
				"target":  "dataway",
				"url":     "https://openway.guance.com",
				"token":   "tkn_xxxxx",
				"filters": []string{"{interface='en0'}", "{abc='def'}"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mVal, err := parseSinkSingle(tc.in)
			assert.Equal(t, tc.expectError, err)
			assert.Equal(t, tc.out, mVal)
		})
	}
}

// go test -v -timeout 30s -run ^TestPolymerizeSinkCategory$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/sinkfuncs
func TestPolymerizeSinkCategory(t *testing.T) {
	cases := []struct {
		name          string
		categoryShort string
		arg           string
		sinks         []map[string]interface{}
		expectError   error
		expectSinks   []map[string]interface{}
	}{
		{
			name:          "normal",
			categoryShort: "M",
			arg:           "influxdb://1.1.1.1:8086?protocol=http&database=db0&timeout=15s",
			expectSinks: []map[string]interface{}{
				{
					"database":   "db0",
					"host":       "1.1.1.1:8086",
					"protocol":   "http",
					"target":     "influxdb",
					"timeout":    "15s",
					"categories": []string{"M"},
				},
			},
		},
		{
			name:          "append_new_all",
			categoryShort: "L",
			arg:           "influxdb://1.1.1.1:8087?protocol=http&database=db1&timeout=15s",
			sinks: []map[string]interface{}{
				{
					"database":   "db0",
					"host":       "1.1.1.1:8086",
					"protocol":   "http",
					"target":     "influxdb",
					"timeout":    "15s",
					"categories": []string{"M"},
				},
			},
			expectSinks: []map[string]interface{}{
				{
					"database":   "db0",
					"host":       "1.1.1.1:8086",
					"protocol":   "http",
					"target":     "influxdb",
					"timeout":    "15s",
					"categories": []string{"M"},
				},
				{
					"database":   "db1",
					"host":       "1.1.1.1:8087",
					"protocol":   "http",
					"target":     "influxdb",
					"timeout":    "15s",
					"categories": []string{"L"},
				},
			},
		},
		{
			name:          "append_new_category",
			categoryShort: "L",
			arg:           "influxdb://1.1.1.1:8086?protocol=http&database=db0&timeout=15s",
			sinks: []map[string]interface{}{
				{
					"database":   "db0",
					"host":       "1.1.1.1:8086",
					"protocol":   "http",
					"target":     "influxdb",
					"timeout":    "15s",
					"categories": []string{"M"},
				},
			},
			expectSinks: []map[string]interface{}{
				{
					"database":   "db0",
					"host":       "1.1.1.1:8086",
					"protocol":   "http",
					"target":     "influxdb",
					"timeout":    "15s",
					"categories": []string{"M", "L"},
				},
			},
		},
		{
			name:          "repeat",
			categoryShort: "M",
			arg:           "influxdb://1.1.1.1:8086?protocol=http&database=db0&timeout=15s",
			sinks: []map[string]interface{}{
				{
					"database":   "db0",
					"host":       "1.1.1.1:8086",
					"protocol":   "http",
					"target":     "influxdb",
					"timeout":    "15s",
					"categories": []string{"M"},
				},
			},
			expectSinks: []map[string]interface{}{
				{
					"database":   "db0",
					"host":       "1.1.1.1:8086",
					"protocol":   "http",
					"target":     "influxdb",
					"timeout":    "15s",
					"categories": []string{"M"},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := polymerizeSinkCategory(tc.categoryShort, tc.arg, &(tc.sinks))
			assert.Equal(t, tc.expectError, err)
			assert.Equal(t, tc.expectSinks, tc.sinks)
		})
	}
}

// go test -v -timeout 30s -run ^TestGetSinkFromEnvs$ gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/installer
func TestGetSinkFromEnvs(t *testing.T) {
	categoryShorts := []string{
		"M",
		"N",
		"K",
		"O",
		"CO",
		"L",
		"T",
		"R",
		"S",
	}

	cases := []struct {
		name        string
		in          []string
		out         []map[string]interface{}
		expectError error
	}{
		{
			name: "normal",
			in: []string{
				"influxdb://1.1.1.1:8081?protocol=http&database=db0&timeout=15s",
				"influxdb://1.1.1.1:8082?protocol=http&database=db0&timeout=15s",
				"influxdb://1.1.1.1:8083?protocol=http&database=db0&timeout=15s",
				"influxdb://1.1.1.1:8084?protocol=http&database=db0&timeout=15s",
				"influxdb://1.1.1.1:8085?protocol=http&database=db0&timeout=15s",
				"influxdb://1.1.1.1:8086?protocol=http&database=db0&timeout=15s",
				"influxdb://1.1.1.1:8087?protocol=http&database=db0&timeout=15s",
				"influxdb://1.1.1.1:8088?protocol=http&database=db0&timeout=15s",
				"influxdb://1.1.1.1:8089?protocol=http&database=db0&timeout=15s",
			},
			out: []map[string]interface{}{
				{
					"target":     "influxdb",
					"protocol":   "http",
					"host":       "1.1.1.1:8081",
					"database":   "db0",
					"timeout":    "15s",
					"categories": []string{"M"},
				},
				{
					"target":     "influxdb",
					"protocol":   "http",
					"host":       "1.1.1.1:8082",
					"database":   "db0",
					"timeout":    "15s",
					"categories": []string{"N"},
				},
				{
					"target":     "influxdb",
					"protocol":   "http",
					"host":       "1.1.1.1:8083",
					"database":   "db0",
					"timeout":    "15s",
					"categories": []string{"K"},
				},
				{
					"target":     "influxdb",
					"protocol":   "http",
					"host":       "1.1.1.1:8084",
					"database":   "db0",
					"timeout":    "15s",
					"categories": []string{"O"},
				},
				{
					"target":     "influxdb",
					"protocol":   "http",
					"host":       "1.1.1.1:8085",
					"database":   "db0",
					"timeout":    "15s",
					"categories": []string{"CO"},
				},
				{
					"target":     "influxdb",
					"protocol":   "http",
					"host":       "1.1.1.1:8086",
					"database":   "db0",
					"timeout":    "15s",
					"categories": []string{"L"},
				},
				{
					"target":     "influxdb",
					"protocol":   "http",
					"host":       "1.1.1.1:8087",
					"database":   "db0",
					"timeout":    "15s",
					"categories": []string{"T"},
				},
				{
					"target":     "influxdb",
					"protocol":   "http",
					"host":       "1.1.1.1:8088",
					"database":   "db0",
					"timeout":    "15s",
					"categories": []string{"R"},
				},
				{
					"target":     "influxdb",
					"protocol":   "http",
					"host":       "1.1.1.1:8089",
					"database":   "db0",
					"timeout":    "15s",
					"categories": []string{"S"},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mVal, err := GetSinkFromEnvs(categoryShorts, tc.in)
			assert.Equal(t, tc.expectError, err)
			assert.Equal(t, tc.out, mVal)
		})
	}
}
