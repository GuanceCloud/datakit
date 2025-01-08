// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

// Package utils contains utils
package utils

import (
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/unix"
)

func RunCmd(name string, args []string, stdout, stderr *os.File) (int, error) {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: unix.SIGKILL,
		Setpgid:   true,
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Start()
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		return -1, err
	}

	var exitCode int
	err = cmd.Wait()
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	return exitCode, err
}
