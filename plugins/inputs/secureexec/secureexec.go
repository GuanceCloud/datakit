package secureexec

import (
	"fmt"
	"runtime"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type SecureExec struct {
	ShellPath string
	Username  string
}

const (
	inputName            = "secureexec"
	secureExecConfSample = `#[inputs.secureexec]
#	shellPath   = "/bin/bash"      # windows: "cmd"" or "powershell"
#	username    = "/your/username" # only for linux
`
	LinuxShellPath   = "/bin/bash"
	WindowsShellPath = "cmd"
	DarwinShellPath  = "/bin/bash"
)

var (
	Log       *logger.Logger
	ShellPath       = ""
	GlobalErr error = nil
)

func (s *SecureExec) Catalog() string {
	return inputName
}

func (s *SecureExec) SampleConfig() string {
	return secureExecConfSample
}

func (s *SecureExec) Run() {
	Log = logger.SLogger(inputName)

	os := runtime.GOOS
	switch os {
	case "linux":
		ShellPath = LinuxShellPath
	case "windows":
		ShellPath = WindowsShellPath
	case "darwin":
		ShellPath = DarwinShellPath
	default:
		GlobalErr = fmt.Errorf("Os %v unsupported", os)
		Log.Errorf("%s", GlobalErr)
		return
	}
	if s.ShellPath != "" {
		ShellPath = s.ShellPath
	}

	Log.Infof("%v input started...", inputName)
	Log.Debugf("shell path = %v, username = %v", ShellPath, s.Username)
	s.ExecInit()
	<-datakit.Exit.Wait()
	Log.Infof("input %v exit", inputName)
}

func Exec(cmds string) (string, error) {
	os := runtime.GOOS
	Log.Debugf("exec cmds: %v on Os: %v", cmds, os)

	if GlobalErr != nil {
		Log.Errorf(GlobalErr.Error())
		return "", GlobalErr
	}

	out, err := ExecCmd(cmds)
	if err != nil {
		Log.Errorf("exec on %v failed: %v", os, err)
	}
	return out, err
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		p := &SecureExec{}
		return p
	})
}
