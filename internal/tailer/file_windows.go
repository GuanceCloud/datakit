// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package tailer

import (
	"fmt"
	"os"
	"syscall"
)

func openFile(path string) (*os.File, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("File '%s' not found. Error: %v", path, syscall.ERROR_FILE_NOT_FOUND)
	}

	pathp, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, fmt.Errorf("Error converting to UTF16: %v", err)
	}

	var access uint32
	access = syscall.GENERIC_READ

	sharemode := uint32(syscall.FILE_SHARE_READ | syscall.FILE_SHARE_WRITE | syscall.FILE_SHARE_DELETE)

	var sa *syscall.SecurityAttributes

	var createmode uint32

	createmode = syscall.OPEN_EXISTING

	handle, err := syscall.CreateFile(pathp, access, sharemode, sa, createmode, syscall.FILE_ATTRIBUTE_NORMAL, 0)

	if err != nil {
		return nil, fmt.Errorf("Error creating file '%s': %v", path, err)
	}

	return os.NewFile(uintptr(handle), path), nil
}
