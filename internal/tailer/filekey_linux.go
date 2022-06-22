//go:build linux
// +build linux

package tailer

import (
	"strconv"
	"syscall"
)

//nolint
func getFileKey(file string) string {
	inodeStr := "inode"
	var stat syscall.Stat_t
	if err := syscall.Stat(file, &stat); err == nil {
		inodeStr = strconv.Itoa(int(stat.Ino))
	}
	return file + "::" + inodeStr
}
