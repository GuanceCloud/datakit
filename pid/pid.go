//+build !windows

package pid

import "syscall"

func CheckPid(pid int) error {
	return syscall.Kill(pid, 0)
}
