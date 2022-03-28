// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestCheckUpgradeVersion(t *testing.T) {
	cases := []struct {
		id, s string
		fail  bool
	}{
		{
			id: "normal",
			s:  "1.2.3",
		},
		{
			id: "zero-minor-version",
			s:  "1.0.3",
		},

		{
			id: "large minor version",
			s:  "1.1024.3",
		},
		{
			id:   `too-large-minor-version`,
			s:    "1.1026.3",
			fail: true,
		},
		{
			id:   `unstable-version`,
			s:    "1.3.3",
			fail: true,
		},

		{
			id:   `1.1.x-stable-rc-version`,
			s:    "1.1.9-rc1", // treat 1.1.x as stable
			fail: false,
		},

		{
			id:   `1.1.x-stable-rc-testing-version`,
			s:    "1.1.7-rc1-125-g40c4860c", // also as stable
			fail: false,
		},

		{
			id:   `1.1.x-stable-rc-hotfix-version`,
			s:    "1.1.7-rc7.1", // stable
			fail: false,
		},

		{
			id:   `invalid-version-string`,
			s:    "2.1.7.0-rc1-126-g40c4860c",
			fail: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			err := checkUpgradeVersion(tc.s)
			if tc.fail {
				tu.NotOk(t, err, "")
				t.Logf("expect error: %s -> %s", tc.s, err)
			} else {
				tu.Ok(t, err)
			}
		})
	}
}

// go test -v -timeout 30s -run ^TestParseSinkSingle$ gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/installer
func TestParseSinkSingle(t *testing.T) {
	cases := []struct {
		name        string
		in          string
		out         map[string]string
		expectError error
	}{
		{
			name: "normal",
			in:   "influxdb://1.1.1.1:8086?protocol=http&database=db0&timeout=15s",
			out: map[string]string{
				"target":   "influxdb",
				"protocol": "http",
				"host":     "1.1.1.1:8086",
				"database": "db0",
				"timeout":  "15s",
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

// go test -v -timeout 30s -run ^TestPolymerizeSinkCategory$ gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/installer
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
			categoryShort: datakit.SinkCategoryMetric,
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
			categoryShort: datakit.SinkCategoryLogging,
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
			categoryShort: datakit.SinkCategoryLogging,
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
			categoryShort: datakit.SinkCategoryMetric,
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
