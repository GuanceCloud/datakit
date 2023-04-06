// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import (
	"fmt"
	"io/ioutil"
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
		x, err := ioutil.ReadFile(l.file)
		if err != nil {
			return err
		}

		pidInFile, err := strconv.Atoi(string(x))
		if err != nil {
			return err
		} else {
			switch pidInFile {
			case -1: // unlocked
				goto write
			case curPid:
				return fmt.Errorf("lock failed(locked by pid %d)", curPid)
			default: // other pid, may terminated
				if pidAlive(pidInFile) {
					return fmt.Errorf("lock failed(locked by alive %d)", pidInFile)
				}
			}
		}
	}

write:
	return ioutil.WriteFile(l.file, []byte(fmt.Sprintf("%d", curPid)), 0o600)
}

func (l *flock) unlock() error {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	return ioutil.WriteFile(l.file, []byte(fmt.Sprintf("%d", -1)), 0o600)
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
