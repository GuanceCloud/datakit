// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package socket

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	successString  string = "succeeded!"
	ncUnknownError string = "unknown fail reason or nc run time out"
)

func findNc() (string, error) {
	ncPath, err := exec.LookPath("nc")
	if err != nil {
		return " ", err
	} else {
		return ncPath, nil
	}
}

func (i *Input) CollectUDP(destHost string, destPort string) error {
	ncPath, err := findNc()
	if err != nil {
		l.Warnf("input socket: %s", err)
		return err
	}

	// -vuz 1.1.1.1 5555
	args := []string{"-vuz", destHost, destPort}
	tags := map[string]string{
		"dest_host": destHost,
		"dest_port": destPort,
		"proto":     "udp",
	}

	fields := map[string]interface{}{
		"success": int64(-1),
	}
	// nolint
	cmd := exec.Command(ncPath, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = RunTimeout(cmd, i.UDPTimeOut.Duration)
	if err != nil {
		l.Warnf("error running nc or services port host or port error: nc %s", err)
	}
	res := out.String()
	if len(res) == 0 {
		res = ncUnknownError
	}
	if i.platform == datakit.OSWindows {
		if !strings.Contains(res, destPort+"(?)") {
			fields["success"] = 1
		}
	} else {
		if strings.Contains(res, successString) {
			fields["success"] = 1
		}
	}
	ts := time.Now()
	tmp := &UDPMeasurement{name: "udp", tags: tags, fields: fields, ts: ts}
	i.collectCache = append(i.collectCache, tmp)

	return nil
}

func RunTimeout(c *exec.Cmd, timeout time.Duration) error {
	if err := c.Start(); err != nil {
		return err
	}
	return WaitTimeout(c, timeout)
}

func WaitTimeout(c *exec.Cmd, timeout time.Duration) error {
	var kill *time.Timer
	term := time.AfterFunc(timeout, func() {
		err := c.Process.Signal(syscall.SIGTERM)
		if err != nil {
			l.Infof("input socket nc terminating process: %s", err)
			return
		}

		kill = time.AfterFunc(KillGrace, func() {
			err := c.Process.Kill()
			if err != nil {
				l.Infof("input socket nc Error killing process: %s", err)
				return
			}
		})
	})

	err := c.Wait()

	// Shutdown all timers
	if kill != nil {
		kill.Stop()
	}

	// If the process exited without error treat it as success.  This allows a
	// process to do a clean shutdown on signal.
	if err == nil {
		return nil
	}

	// If SIGTERM was sent then treat any process error as a timeout.
	if !term.Stop() {
		return fmt.Errorf("\"nc command timed out\"")
	}

	// Otherwise there was an error unrelated to termination.
	return err
}
