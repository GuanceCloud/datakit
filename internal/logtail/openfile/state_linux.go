// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package openfile

import (
	"strconv"
	"syscall"
)

func FileKey(file string) string {
	return file + "::" + FileInode(file)
}

func FileInode(file string) string {
	inodeStr := "inode"
	var stat syscall.Stat_t
	if err := syscall.Stat(file, &stat); err == nil {
		inodeStr = strconv.Itoa(int(stat.Ino))
	}
	return inodeStr
}
