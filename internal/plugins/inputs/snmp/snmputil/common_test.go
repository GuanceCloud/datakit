// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_makeStringBatches(t *testing.T) {
	tests := []struct {
		name            string
		elements        []string
		size            int
		expectedBatches [][]string
		expectedError   error
	}{
		{
			"three batches, last with diff length",
			[]string{"aa", "bb", "cc", "dd", "ee"},
			2,
			[][]string{
				{"aa", "bb"},
				{"cc", "dd"},
				{"ee"},
			},
			nil,
		},
		{
			"two batches same length",
			[]string{"aa", "bb", "cc", "dd", "ee", "ff"},
			3,
			[][]string{
				{"aa", "bb", "cc"},
				{"dd", "ee", "ff"},
			},
			nil,
		},
		{
			"one full batch",
			[]string{"aa", "bb", "cc"},
			3,
			[][]string{
				{"aa", "bb", "cc"},
			},
			nil,
		},
		{
			"one partial batch",
			[]string{"aa"},
			3,
			[][]string{
				{"aa"},
			},
			nil,
		},
		{
			"large batch size",
			[]string{"aa", "bb", "cc", "dd", "ee", "ff"},
			100,
			[][]string{
				{"aa", "bb", "cc", "dd", "ee", "ff"},
			},
			nil,
		},
		{
			"zero element",
			[]string{},
			2,
			[][]string(nil),
			nil,
		},
		{
			"zero batch size",
			[]string{"aa", "bb", "cc", "dd", "ee"},
			0,
			nil,
			fmt.Errorf("batch size must be positive. invalid size: 0"),
		},
		{
			"negative batch size",
			[]string{"aa", "bb", "cc", "dd", "ee"},
			-1,
			nil,
			fmt.Errorf("batch size must be positive. invalid size: -1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batches, err := CreateStringBatches(tt.elements, tt.size)
			assert.Equal(t, tt.expectedBatches, batches)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func Test_CopyStrings(t *testing.T) {
	tags := []string{"aa", "bb"}
	newTags := CopyStrings(tags)
	assert.Equal(t, tags, newTags)
	assert.NotEqual(t, fmt.Sprintf("%p", tags), fmt.Sprintf("%p", newTags))
	assert.NotEqual(t, fmt.Sprintf("%p", &tags[0]), fmt.Sprintf("%p", &newTags[0]))
}

func TestValidateNamespace(t *testing.T) {
	assert := assert.New(t)
	long := strings.Repeat("a", 105)
	_, err := NormalizeNamespace(long)
	assert.NotNil(err, "namespace should not be too long")

	namespace, err := NormalizeNamespace("a<b")
	assert.Nil(err, "namespace with symbols should be normalized")
	assert.Equal("a-b", namespace, "namespace should not contain symbols")

	namespace, err = NormalizeNamespace("a\nb")
	assert.Nil(err, "namespace with symbols should be normalized")
	assert.Equal("ab", namespace, "namespace should not contain symbols")

	// Invalid namespace as bytes that would look like this: 9cbef2d1-8c20-4bf2-97a5-7d70��
	b := []byte{
		57, 99, 98, 101, 102, 50, 100, 49, 45, 56, 99, 50, 48, 45,
		52, 98, 102, 50, 45, 57, 55, 97, 53, 45, 55, 100, 55, 48,
		0, 0, 0, 0, 239, 191, 189, 239, 191, 189, 1, // these are bad bytes
	}
	_, err = NormalizeNamespace(string(b))
	assert.NotNil(err, "namespace should not contain bad bytes")
}

// go test -v -timeout 30s -run ^Test_getIPInterfaceByTags$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil
func Test_getIPInterfaceByTags(t *testing.T) {
	cases := []struct {
		name                string
		tags                []string
		expectIP, expectInf string
	}{
		{
			name: "normal",
			tags: []string{
				"snmp_profile:cisco-catalyst",
				"device_vendor:cisco",
				"snmp_host:jk",
				"ip:1.1.1.1",
				"agent_host:",
				"agent_version:1.6.1-459-g0d3783817f",
				"interface:Gi1/0/3",
				"interface_alias:conn-to-4XF4BX2-2port-2.2.2.2",
			},
			expectIP:  "1.1.1.1",
			expectInf: "Gi1/0/3",
		},

		{
			name: "unexpect",
			tags: []string{
				"some",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ip, inf := getIPInterfaceByTags(tc.tags)
			assert.Equal(t, tc.expectIP, ip)
			assert.Equal(t, tc.expectInf, inf)
		})
	}
}

// go test -v -timeout 30s -run ^Test_getPreviousBandwidthUsageRateKeyName$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil
func Test_getPreviousBandwidthUsageRateKeyName(t *testing.T) {
	cases := []struct {
		name                string
		ip, inf, metricName string
		expect              string
	}{
		{
			name:       "normal",
			ip:         "1.1.1.1",
			inf:        "Gi1/0/3",
			metricName: "ifBandwidthInUsage.rate",
			expect:     "1.1.1.1_Gi1/0/3_ifBandwidthInUsage.rate",
		},

		{
			name:       "emptyIP",
			inf:        "Gi1/0/3",
			metricName: "ifBandwidthInUsage.rate",
			expect:     "",
		},

		{
			name:       "emptyInterface",
			ip:         "1.1.1.1",
			metricName: "ifBandwidthInUsage.rate",
			expect:     "",
		},

		{
			name:   "emptyMetricName",
			ip:     "1.1.1.1",
			inf:    "Gi1/0/3",
			expect: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := getPreviousBandwidthUsageRateKeyName(tc.ip, tc.inf, tc.metricName)
			assert.Equal(t, tc.expect, out)
		})
	}
}

// go test -v -timeout 30s -run ^Test_calculateBandwidthUtilization$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil
func Test_calculateBandwidthUtilization(t *testing.T) {
	preEmpty := func() {
		previousBandwidthUsageRate = sync.Map{}
	}

	preInvalidValueItem := func() {
		previousBandwidthUsageRate = sync.Map{}

		previousBandwidthUsageRate.Store("1.1.1.1_Gi1/0/3_ifBandwidthInUsage.rate", 1)
	}

	preNegative := func() {
		previousBandwidthUsageRate = sync.Map{}

		previousBandwidthUsageRate.Store(
			"1.1.1.1_Gi1/0/3_ifBandwidthInUsage.rate",
			newValueItem(10000, 1),
		)
	}

	prePreZero := func() {
		previousBandwidthUsageRate = sync.Map{}

		previousBandwidthUsageRate.Store(
			"1.1.1.1_Gi1/0/3_ifBandwidthInUsage.rate",
			newValueItem(0, 0),
		)
	}

	preNewMetric := func() {
		previousBandwidthUsageRate = sync.Map{}

		previousBandwidthUsageRate.Store(
			"1.1.1.1_Gi1/0/3_ifBandwidthInUsage.rate",
			newValueItem(1000, 1),
		)
	}

	_ = preEmpty
	_ = preInvalidValueItem
	_ = preNegative
	_ = preNewMetric

	cases := []struct {
		name                   string
		previous               func()
		ip, inf, metricName    string
		metricValue, timestamp float64
		expectErr              error
		expect                 float64
	}{
		{
			name:        "Zero",
			previous:    preEmpty,
			ip:          "1.1.1.1",
			inf:         "Gi1/0/3",
			metricName:  "ifBandwidthInUsage.rate",
			metricValue: 0,
			timestamp:   5,
			expect:      0,
		},

		{
			name:        "InvalidInterface",
			previous:    preEmpty,
			ip:          "1.1.1.1",
			metricName:  "ifBandwidthInUsage.rate",
			metricValue: 10000,
			timestamp:   5,
			expectErr:   fmt.Errorf("unexpected ip and interface"),
			expect:      0,
		},

		{
			name:        "EmptyKey",
			previous:    preEmpty,
			ip:          "1.1.1.1",
			inf:         "Gi1/0/3",
			metricValue: 10000,
			timestamp:   5,
			expectErr:   fmt.Errorf("unexpected key name"),
			expect:      0,
		},

		{
			name:        "InvalidValueItem",
			previous:    preInvalidValueItem,
			ip:          "1.1.1.1",
			inf:         "Gi1/0/3",
			metricName:  "ifBandwidthInUsage.rate",
			metricValue: 10000,
			timestamp:   5,
			expectErr:   fmt.Errorf("invalid *valueItem"),
			expect:      0,
		},

		{
			name:        "Negative",
			previous:    preNegative,
			ip:          "1.1.1.1",
			inf:         "Gi1/0/3",
			metricName:  "ifBandwidthInUsage.rate",
			metricValue: 1000,
			timestamp:   5,
			expect:      0,
		},

		{
			name:        "PreZero",
			previous:    prePreZero,
			ip:          "1.1.1.1",
			inf:         "Gi1/0/3",
			metricName:  "ifBandwidthInUsage.rate",
			metricValue: 1000,
			timestamp:   5,
			expect:      0,
		},

		{
			name:        "NewDevice",
			previous:    preEmpty,
			ip:          "1.1.1.1",
			inf:         "Gi1/0/3",
			metricName:  "ifBandwidthInUsage.rate",
			metricValue: 10000,
			timestamp:   5,
			expect:      0,
		},

		{
			name:        "NewMetric",
			previous:    preNewMetric,
			ip:          "1.1.1.1",
			inf:         "Gi1/0/3",
			metricName:  "ifBandwidthInUsage.rate",
			metricValue: 10000,
			timestamp:   5,
			expect:      2250,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.previous()

			out, err := calculateBandwidthUtilization(tc.ip, tc.inf, tc.metricName, tc.metricValue, tc.timestamp)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expect, out)
		})
	}
}
