package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"syscall"
	"time"
)

func RunWithTimeout(timeout time.Duration, sudo bool, command string, args ...string) ([]byte, error) {
	if sudo {
		args = append([]string{"-n", command}, args...)
		command = "sudo"
	}
	cmd := exec.Command(command, args...)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	after := time.AfterFunc(timeout, func() {
		err = cmd.Process.Kill()
	})

	cmderr := cmd.Wait()

	if !after.Stop() {
		if err == nil {
			return nil, fmt.Errorf("command %q process overtime", command)
		}

		return nil, err
	}

	output, err := ioutil.ReadAll(&buf)
	if err != nil {
		return nil, err
	}

	return output, cmderr
}

// Command line parse errors are denoted by the exit code having the 0 bit set.
// All other errors are drive/communication errors and should be ignored.
func ExitStatus(err error) (int, error) {
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus(), nil
		}
	}

	return 0, err
}
