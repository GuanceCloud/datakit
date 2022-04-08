package cgroup

import (
	"testing"
)

func TestSetup(t *testing.T) {
	// cases := []*CgroupOptions{
	cases := []struct {
		name string
		opt  *CgroupOptions
	}{
		{
			name: "mem-100m",
			opt: &CgroupOptions{
				Enable: true,
				Path:   "/test-setup",
				CPUMax: 10.0,
				CPUMin: 20.0,
				MemMax: 1024 * 1024 * 10,
			},
		},
	}

	cg := Cgroup{}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cg.opt = tc.opt
			if err := cg.setup(); err != nil {
				t.Logf("cg.setup: %s", err)
			} else {
				m, err := cg.control.Stat()
				if err != nil {
					t.Logf("cg.control.Stat: %s", err)
				} else {
					t.Logf("Mem.Swap: %v, Mem.Usage: %v", m.Memory.Swap, m.Memory.Usage)
				}

				cg.stop()
			}
		})
	}
}
