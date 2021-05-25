package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
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
		} else {
			return nil, err
		}
	}

	output, err := ioutil.ReadAll(&buf)
	if err != nil {
		return nil, err
	}

	return output, cmderr
}
