// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package process

import (
	pr "github.com/shirou/gopsutil/v3/process"
)

// getProcessStatus returns process status compatible with cross-platform format.
// On Linux, it uses the gopsutil Status() method which reads from /proc/[pid]/status.
// On Windows, it infers status from process exit code and handle state.
// On other platforms (Darwin/BSD), it uses gopsutil implementation.
// Returns a string slice with status character(s) like: R (Running), S (Sleep), I (Idle), Z (Zombie), etc.
func getProcessStatus(proc *pr.Process) ([]string, error) {
	return getProcessStatusImpl(proc)
}
