// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package openfile

import (
	"fmt"
	"strconv"
	"syscall"
)

func FileKey(file string) string {
	return file + "::" + Inode(file)
}

func UniqueID(file string) string {
	return fmt.Sprintf("dev:%s/ino:%s", Device(file), Inode(file))
}

func Inode(file string) string {
	ino := "inode"
	var stat syscall.Stat_t
	if err := syscall.Stat(file, &stat); err == nil {
		ino = strconv.Itoa(int(stat.Ino))
	}
	return ino
}

func Device(file string) string {
	dev := "driveD"
	var stat syscall.Stat_t
	if err := syscall.Stat(file, &stat); err == nil {
		dev = strconv.Itoa(int(stat.Dev))
	}
	return dev
}
