// Package compareutil contains compare utils
package compareutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v -timeout 30s -run ^TestCompareListDisordered$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/compareutil
func TestCompareListDisordered(t *testing.T) {
	cases := []struct {
		name   string
		left   []string
		right  []string
		expect bool
	}{
		{
			name:   "same",
			left:   []string{"1", "2", "3"},
			right:  []string{"1", "2", "3"},
			expect: true,
		},
		{
			name:   "disorder",
			left:   []string{"1", "2", "3"},
			right:  []string{"3", "2", "1"},
			expect: true,
		},
		{
			name:   "not_equal",
			left:   []string{"1", "2", "3"},
			right:  []string{"3", "2"},
			expect: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := CompareListDisordered(tc.left, tc.right)
			assert.Equal(t, tc.expect, out)
		})
	}
}
