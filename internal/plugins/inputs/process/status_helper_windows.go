// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package process

import (
	pr "github.com/shirou/gopsutil/v3/process"
	"golang.org/x/sys/windows"
)

// getProcessStatusImpl returns process status for Windows processes.
// Windows doesn't have the same process state model as Linux, so we infer status.
// Possible return values: "R" (Running), "I" (Idle/Waiting), "Z" (Zombie/Dead).
func getProcessStatusImpl(proc *pr.Process) ([]string, error) {
	// Try to open the process handle.
	handle, err := windows.OpenProcess(
		windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_QUERY_LIMITED_INFORMATION,
		false,
		uint32(proc.Pid),
	)
	if err != nil {
		// Process no longer exists or is a zombie.
		return []string{pr.Zombie}, nil
	}
	defer windows.CloseHandle(handle)

	// Check if the process has exited by trying to get its exit code.
	var exitCode uint32
	err = windows.GetExitCodeProcess(handle, &exitCode)
	if err != nil {
		// If we can't get exit code, assume it's running
		return []string{pr.Running}, nil
	}

	// If exit code is not STILL_ACTIVE, the process is dead/zombie.
	const stillActive = 259 // Windows STILL_ACTIVE constant.
	if exitCode != stillActive {
		return []string{pr.Zombie}, nil
	}

	// Process is still active. Check thread states for more detailed status.
	// For now, we return Running as the default for active processes.
	// A more sophisticated check could use NtQuerySystemInformation to check thread wait reasons.
	return []string{pr.Running}, nil
}
