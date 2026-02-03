// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build darwin || freebsd || openbsd || netbsd || solaris || dragonfly
// +build darwin freebsd openbsd netbsd solaris dragonfly

package process

import (
	pr "github.com/shirou/gopsutil/v3/process"
)

// getProcessStatusImpl returns process status for non-Linux/non-Windows platforms.
// Uses the gopsutil implementation which works on Darwin/macOS and BSD variants.
func getProcessStatusImpl(proc *pr.Process) ([]string, error) {
	return proc.Status()
}
