// Package calcutil wraps calculate functions
package calcutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v -timeout 30s -run ^TestAtomicMinusUint64$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/calcutil
func TestAtomicMinusUint64(t *testing.T) {
	cases := []struct {
		name  string
		val   uint64
		minus int64
		out   uint64
	}{
		{
			name:  "positive",
			val:   20,
			minus: 9,
			out:   11,
		},
		{
			name:  "negative",
			val:   20,
			minus: -9,
			out:   11,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := AtomicMinusUint64(&tc.val, tc.minus)
			assert.Equal(t, tc.out, out)
		})
	}
}
