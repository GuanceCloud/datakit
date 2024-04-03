//go:build linux
// +build linux

package run

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/shirou/gopsutil/process"
)

func quit(fp string) {
	_ = os.Remove(fp)
}

func savePid(fp string) error {
	if isRuning(fp) {
		return fmt.Errorf("ebpf still running, PID: %s", fp)
	}

	pid := os.Getpid()
	return os.WriteFile(fp, []byte(fmt.Sprintf("%d", pid)), 0o644) //nolint:gosec
}

func isRuning(fp string) bool {
	var oidPid int64
	var name string
	var p *process.Process

	cont, err := os.ReadFile(filepath.Clean(fp))
	// pid file does not exist
	if err != nil {
		return false
	}

	oidPid, err = strconv.ParseInt(string(cont), 10, 32)
	if err != nil {
		return false
	}

	p, _ = process.NewProcess(int32(oidPid))
	name, _ = p.Name()

	return name == "datakit-ebpf"
}
