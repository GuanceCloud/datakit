// +build linux

package netebpf

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
		"4.8.11":  true,
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

func TestCheckThrLSBInfo(t *testing.T) {
	cases := map[[3]string]bool{
		{"ubuntu", "debian", "20.04"}:  true,
		{"ubuntu", "debian", "12.04"}:  false,
		{"centos", "rhel", "7.6.1810"}: true,
	}
	for k, v := range cases {
		if v != checkIsCentos76Ubuntu1604(k[0], k[2]) {
			t.Error(k)
		}
	}
}
