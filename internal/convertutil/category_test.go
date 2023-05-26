// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package convertutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

//------------------------------------------------------------------------------

// go test -v -timeout 30s -run ^TestGetMapCategoryShortToFull$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/convertutil
func TestGetMapCategoryShortToFull(t *testing.T) {
	cases := []struct {
		name        string
		in          string
		expectError error
		expect      string
	}{
		{
			name:   "metric",
			in:     datakit.CategoryMetric,
			expect: datakit.Metric,
		},
		{
			name:   "network",
			in:     datakit.CategoryNetwork,
			expect: datakit.Network,
		},
		{
			name:   "keyEvent",
			in:     datakit.CategoryKeyEvent,
			expect: datakit.KeyEvent,
		},
		{
			name:   "object",
			in:     datakit.CategoryObject,
			expect: datakit.Object,
		},
		{
			name:   "custom_object",
			in:     datakit.CategoryCustomObject,
			expect: datakit.CustomObject,
		},
		{
			name:   "logging",
			in:     datakit.CategoryLogging,
			expect: datakit.Logging,
		},
		{
			name:   "tracing",
			in:     datakit.CategoryTracing,
			expect: datakit.Tracing,
		},
		{
			name:   "profiling",
			in:     datakit.CategoryProfiling,
			expect: datakit.Profiling,
		},
		{
			name:   "rum",
			in:     datakit.CategoryRUM,
			expect: datakit.RUM,
		},
		{
			name:   "security",
			in:     datakit.CategorySecurity,
			expect: datakit.Security,
		},
		{
			name:        "unrecognized category",
			in:          "",
			expectError: fmt.Errorf("unrecognized category"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := GetMapCategoryShortToFull(tc.in)
			assert.Equal(t, tc.expectError, err)
			assert.Equal(t, tc.expect, out)
		})
	}
}

// go test -v -timeout 30s -run ^TestGetMapCategoryFullToShort$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/convertutil
func TestGetMapCategoryFullToShort(t *testing.T) {
	cases := []struct {
		name        string
		in          string
		expectError error
		expect      string
	}{
		{
			name:   "metric",
			in:     datakit.Metric,
			expect: datakit.CategoryMetric,
		},
		{
			name:   "network",
			in:     datakit.Network,
			expect: datakit.CategoryNetwork,
		},
		{
			name:   "keyEvent",
			in:     datakit.KeyEvent,
			expect: datakit.CategoryKeyEvent,
		},
		{
			name:   "object",
			in:     datakit.Object,
			expect: datakit.CategoryObject,
		},
		{
			name:   "custom_object",
			in:     datakit.CustomObject,
			expect: datakit.CategoryCustomObject,
		},
		{
			name:   "logging",
			in:     datakit.Logging,
			expect: datakit.CategoryLogging,
		},
		{
			name:   "tracing",
			in:     datakit.Tracing,
			expect: datakit.CategoryTracing,
		},
		{
			name:   "profiling",
			in:     datakit.Profiling,
			expect: datakit.CategoryProfiling,
		},
		{
			name:   "rum",
			in:     datakit.RUM,
			expect: datakit.CategoryRUM,
		},
		{
			name:   "security",
			in:     datakit.Security,
			expect: datakit.CategorySecurity,
		},
		{
			name:        "unrecognized category",
			expect:      "",
			expectError: fmt.Errorf("unrecognized category"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := GetMapCategoryFullToShort(tc.in)
			assert.Equal(t, tc.expectError, err)
			assert.Equal(t, tc.expect, out)
		})
	}
}
