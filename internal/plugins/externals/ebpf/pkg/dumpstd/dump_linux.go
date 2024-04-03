//go:build linux
// +build linux

// Package dumpstd dump stderr to file
package dumpstd

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const (
	FileModeRW = 0o644
	DirModeRW  = 0o755
)

func DumpStderr2File(dir string) error {
	dirpath := filepath.Join(dir, "externals")
	filepath := filepath.Join(dirpath, "datakit-ebpf.stderr")
	if err := os.MkdirAll(dirpath, DirModeRW); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_RDWR, FileModeRW) //nolint:gosec
	if err != nil {
		return err
	}
	_, err = f.WriteString(fmt.Sprintf("\n========= %s =========\n", time.Now().UTC()))
	if err != nil {
		return err
	}
	_ = syscall.Dup3(int(f.Fd()), int(os.Stderr.Fd()), 0) // for arm64 arch
	if err = f.Close(); err != nil {
		return err
	}
	return nil
}
