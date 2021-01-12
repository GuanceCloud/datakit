// +build darwin

package secureexec

import (
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
)

var (
	Uid uint32
	Gid uint32
)

func (s *SecureExec) ExecInit() error {
	user, err := user.Lookup(s.Username)
	if err != nil {
		return err
	}
	uid, _ := strconv.ParseUint(user.Uid, 10, 32)
	gid, _ := strconv.ParseUint(user.Gid, 10, 32)

	Uid = uint32(uid)
	Gid = uint32(gid)
	return nil
}

func ExecCmd(cmds string) (string, error) {
	cmd := exec.Command(ShellPath, "-c", cmds)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: Uid, Gid: Gid}
	stdout, err := cmd.CombinedOutput()
	return string(stdout), err
}
