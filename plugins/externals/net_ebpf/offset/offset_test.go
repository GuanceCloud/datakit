//go:build (linux && ignore) || ebpf
// +build linux,ignore ebpf

package offset

import (
	"syscall"
	"testing"
)

func ownNetNS() (uint64, error) {
	var s syscall.Stat_t
	if err := syscall.Stat("/proc/self/ns/net", &s); err != nil {
		return 0, err
	}
	return s.Ino, nil
}

func TestF(t *testing.T) {
	a, err := ownNetNS()
	if err != nil {
		t.Error(err)
	} else {
		t.Error(a)
	}
	v, err := getLinuxKernelVesion()
	if err == nil {
		t.Errorf("%x", v)
	}
}
