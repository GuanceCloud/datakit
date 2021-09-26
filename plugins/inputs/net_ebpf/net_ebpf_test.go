// +build linux

package net_ebpf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckLinuxKernelVersion(t *testing.T) {
	cases := map[string]bool{
		"4.9.0":   true,
		"5.4.0":   true,
		"5.3.10":  true,
		"5.11.10": true,
		"4.8.11":  false,
		"3.10.0":  false,
		"3.10.11": false,
	}
	for k := range cases {
		r, err := checkLinuxKernelVesion(k)
		assert.Equal(t, cases[k], r, k)
		if cases[k] == false {
			if err == nil {
				t.Error("empty error")
			}
		}
	}
}
