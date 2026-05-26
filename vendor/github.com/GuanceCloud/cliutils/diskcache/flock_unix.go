// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows

package diskcache

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

func (wl *walLock) tryLock() (bool, error) {
	f, err := os.OpenFile(wl.file, os.O_CREATE|os.O_RDWR, 0o666) // nolint: gosec
	if err != nil {
		return false, err
	}
	wl.f = f

	// LOCK_EX = Exclusive, LOCK_NB = Non-blocking
	err = syscall.Flock(int(wl.f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		// If the error is EWOULDBLOCK, it means someone else has the lock
		if errors.Is(err, syscall.EWOULDBLOCK) {
			if err := wl.f.Close(); err != nil {
				l.Errorf("Close: %s", err.Error())
			}
			return false, fmt.Errorf("locked")
		}

		if err := wl.f.Close(); err != nil {
			l.Errorf("Close: %s", err.Error())
		}
		return false, err
	}
	return true, nil
}

func (wl *walLock) unlock() {
	if wl.f != nil {
		if err := syscall.Flock(int(wl.f.Fd()), syscall.LOCK_UN); err != nil {
			l.Errorf("Flock: %s", err.Error())
		}

		if err := wl.f.Close(); err != nil {
			l.Errorf("CLose: %s", err.Error())
		}

		if err := os.Remove(wl.file); err != nil { // Optional on Unix
			l.Errorf("Remove: %s", err.Error())
		}
	}
}
