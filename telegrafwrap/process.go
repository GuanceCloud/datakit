//+build !windows

package telegrafwrap

import "syscall"

func KillProcess(pid int) error {
	return syscall.Kill(pid, 0)
}
