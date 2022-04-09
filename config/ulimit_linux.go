//go:build linux
// +build linux

package config

import (
	"syscall"
)

func setUlimit(n uint64) error {
	var rLimit syscall.Rlimit //nolint:typecheck
	// Set both soft limit and hard limit to n.
	rLimit.Cur = n
	rLimit.Max = n
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil { //nolint:typecheck
		return err
	}
	return nil
}

func getUlimit() (soft, hard uint64, err error) {
	var rLimit syscall.Rlimit                                                 //nolint:typecheck
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil { //nolint:typecheck
		return 0, 0, err
	}
	return rLimit.Cur, rLimit.Max, nil
}
