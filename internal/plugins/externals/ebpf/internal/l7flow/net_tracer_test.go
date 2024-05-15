//go:build linux
// +build linux

package l7flow

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
)

func TestSort(t *testing.T) {
	fn := func(cases []uint32, expected []uint32, start ...uint32) {
		rst := []uint32{}
		netdata := netdata{}
		if len(start) > 0 {
			netdata.prvDataPos = start[0]
		}
		for _, v := range cases {
			r := netdata.Queue(&comm.NetwrkData{Index: v})
			for _, d := range r {
				rst = append(rst, d.Index)
			}
		}
		assert.Equal(t, expected, rst)
	}

	t.Run("c1", func(t *testing.T) {
		li := []uint32{1, 4, 3, 5, 2}
		fn(li, []uint32{1, 2, 3, 4, 5})
	})

	t.Run("c2", func(t *testing.T) {
		li := []uint32{5, 4, 3, 2, 1}
		fn(li, []uint32{1, 2, 3, 4, 5})
	})

	t.Run("c3", func(t *testing.T) {
		li := []uint32{1, 2, 3, 4, 5}
		fn(li, []uint32{1, 2, 3, 4, 5})
	})

	t.Run("c4", func(t *testing.T) {
		li := []uint32{1, 2}
		fn(li, []uint32{1, 2})
	})

	t.Run("c5", func(t *testing.T) {
		li := []uint32{2, 1}
		fn(li, []uint32{1, 2})
	})
	t.Run("c5", func(t *testing.T) {
		li := []uint32{2, 4, 3, 1}
		fn(li, []uint32{1, 2, 3, 4})
	})

	t.Run("c6", func(t *testing.T) {
		li := []uint32{1, 6, 2, 3, 4, 5, 7, 9, 10, 8, 11, 14, 13, 12}
		fn(li, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14})
	})

	t.Run("c7", func(t *testing.T) {
		var li []uint32
		for i := 2; i < 2+1024; i++ {
			li = append(li, uint32(i))
		}
		fn(li, li[:128])
	})

	t.Run("c8", func(t *testing.T) {
		var li, rst []uint32
		startPos := uint32(math.MaxUint32 - 100)
		for i := startPos + 1; i < math.MaxUint32; i++ {
			li = append(li, i)
			rst = append(rst, i)
		}

		li = append(li, 1, math.MaxUint32, 2, 0, 3)
		rst = append(rst, math.MaxUint32, 0, 1, 2, 3)
		fn(li, rst, startPos)
	})
}
