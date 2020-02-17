//+build !windows

package telegrafwrap

import "syscall"

func CheckPid(pid int) error {
	return syscall.Kill(pid, 0)
}
