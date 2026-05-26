// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package diskcache

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	modkernel32    = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx = modkernel32.NewProc("LockFileEx")
)

func (wl *walLock) tryLock() (bool, error) {
	f, err := os.OpenFile(wl.file, os.O_CREATE|os.O_RDWR, 0o666)
	if err != nil {
		return false, err
	}
	wl.f = f

	// LOCKFILE_EXCLUSIVE_LOCK = 2, LOCKFILE_FAIL_IMMEDIATELY = 1
	flags := uint32(2 | 1)
	var overlapped syscall.Overlapped

	// Call Win32 LockFileEx
	ret, _, err := procLockFileEx.Call(
		uintptr(wl.f.Fd()),
		uintptr(flags),
		0, // reserved
		0, // length low
		1, // length high (lock 1 byte)
		uintptr(unsafe.Pointer(&overlapped)),
	)

	if ret == 0 {
		// ERROR_LOCK_VIOLATION = 33
		if errno, ok := err.(syscall.Errno); ok && errno == 33 {
			wl.f.Close()
			return false, nil
		}
		wl.f.Close()
		return false, err
	}

	return true, nil
}

func (wl *walLock) unlock() {
	if wl.f != nil {
		wl.f.Close() // Closing the file handle automatically releases the lock in Windows
		os.Remove(wl.file)
	}
}
