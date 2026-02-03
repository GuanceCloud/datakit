// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package process

import (
	pr "github.com/shirou/gopsutil/v3/process"
)

// getProcessStatusImpl returns process status for Linux processes.
// Uses the native gopsutil implementation which reads from /proc/[pid]/status.
func getProcessStatusImpl(proc *pr.Process) ([]string, error) {
	return proc.Status()
}
