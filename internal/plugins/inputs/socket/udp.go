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

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

const (
	successString  string = "succe"
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

func (i *input) collectUDP(destHost string, destPort string) *point.Point {
	ncPath, err := findNc()
	if err != nil {
		l.Warnf("input socket: %s", err)

		i.feeder.FeedLastError(err.Error(),
			io.WithLastErrorInput(inputName),
			io.WithLastErrorCategory(point.Metric))
		return nil
	}

	// -vuz 1.1.1.1 5555
	args := []string{"-vuz", destHost, destPort}

	var kvs point.KVs
	kvs = kvs.MustAddTag("dest_host", destHost)
	kvs = kvs.MustAddTag("dest_port", destPort)
	kvs = kvs.MustAddTag("proto", "udp")

	// default set failed
	kvs = kvs.Add("success", -1, false, true)

	// nolint
	cmd := exec.Command(ncPath, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = runTimeout(cmd, i.UDPTimeOut.Duration)
	if err != nil {
		i.feeder.FeedLastError(err.Error(),
			io.WithLastErrorInput(inputName),
			io.WithLastErrorCategory(point.Metric))
	}

	res := out.String()
	if len(res) == 0 {
		res = ncUnknownError
	}

	if i.platform == datakit.OSWindows {
		if !strings.Contains(res, destPort+"(?)") {
			kvs = kvs.Add("success", 1, false, true)
		}
	} else {
		if strings.Contains(res, successString) {
			kvs = kvs.Add("success", 1, false, true)
		}
	}

	return point.NewPointV2("udp", kvs, point.DefaultMetricOptions()...)
}

func runTimeout(c *exec.Cmd, timeout time.Duration) error {
	if err := c.Start(); err != nil {
		return err
	}
	return waitTimeout(c, timeout)
}

func waitTimeout(c *exec.Cmd, timeout time.Duration) error {
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
