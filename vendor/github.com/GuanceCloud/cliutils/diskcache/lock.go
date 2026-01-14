// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"syscall"
)

type flock struct {
	file string
	mtx  *sync.Mutex
}

func newFlock(path string) *flock {
	return &flock{
		file: filepath.Clean(filepath.Join(path, ".lock")),
		mtx:  &sync.Mutex{},
	}
}

func (l *flock) lock() error {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	curPid := os.Getpid()

	if _, err := os.Stat(l.file); err != nil {
		goto write // file not exist
	} else {
		x, err := os.ReadFile(l.file)
		if err != nil {
			return WrapFileOperationError(OpRead, err, "", l.file).
				WithDetails("failed_to_read_lock_file")
		}

		if len(x) == 0 {
			goto write
		}

		pidInFile, err := strconv.Atoi(string(x))
		if err != nil {
			return NewCacheError(OpLock, err,
				fmt.Sprintf("failed_to_parse_pid_from_lock_file: content=%q", string(x))).
				WithFile(l.file)
		} else {
			switch pidInFile {
			case -1: // unlocked
				goto write
			case curPid:
				return NewCacheError(OpLock, fmt.Errorf("already_locked_by_current_process"), "").
					WithFile(l.file).WithDetails(fmt.Sprintf("current_pid=%d", curPid))
			default: // other pid, may terminated
				if pidAlive(pidInFile) {
					return WrapLockError(fmt.Errorf("process_already_has_lock"), "", pidInFile).
						WithFile(l.file)
				}
			}
		}
	}

write:
	if err := os.WriteFile(l.file, []byte(strconv.Itoa(curPid)), 0o600); err != nil {
		return WrapFileOperationError(OpWrite, err, "", l.file).
			WithDetails(fmt.Sprintf("failed_to_write_pid_to_lock_file: pid=%d", curPid))
	}
	return nil
}

func (l *flock) unlock() error {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	if err := os.WriteFile(l.file, []byte(strconv.Itoa(-1)), 0o600); err != nil {
		return WrapFileOperationError(OpWrite, err, "", l.file).
			WithDetails("failed_to_write_unlock_marker")
	}
	return nil
}

func pidAlive(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Signal not available on windows.
	if runtime.GOOS == "windows" {
		return true
	}

	if err := p.Signal(syscall.Signal(0)); err != nil {
		switch err.Error() {
		case "operation not permitted":
			return true
		default:
			return false
		}
	} else {
		return true
	}
}
