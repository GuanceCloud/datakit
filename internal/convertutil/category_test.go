package convertutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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
			name:   "Metric",
			in:     datakit.CategoryMetric,
			expect: datakit.Metric,
		},
		{
			name:   "Network",
			in:     datakit.CategoryNetwork,
			expect: datakit.Network,
		},
		{
			name:   "KeyEvent",
			in:     datakit.CategoryKeyEvent,
			expect: datakit.KeyEvent,
		},
		{
			name:   "Object",
			in:     datakit.CategoryObject,
			expect: datakit.Object,
		},
		{
			name:   "CustomObject",
			in:     datakit.CategoryCustomObject,
			expect: datakit.CustomObject,
		},
		{
			name:   "Logging",
			in:     datakit.CategoryLogging,
			expect: datakit.Logging,
		},
		{
			name:   "Tracing",
			in:     datakit.CategoryTracing,
			expect: datakit.Tracing,
		},
		{
			name:   "RUM",
			in:     datakit.CategoryRUM,
			expect: datakit.RUM,
		},
		{
			name:   "Security",
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
