package cmds

import (
	"fmt"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"runtime"
	"strings"
)

var (
	ExternalInstallDir = map[string]string{
		datakit.OSArchWinAmd64:    `C:\Program Files\`,
		datakit.OSArchWin386:      `C:\Program Files (x86)\`,
		datakit.OSArchLinuxArm:    `/usr/local/`,
		datakit.OSArchLinuxArm64:  `/usr/local/`,
		datakit.OSArchLinuxAmd64:  `/usr/local/`,
		datakit.OSArchLinux386:    `/usr/local/`,
		datakit.OSArchDarwinAmd64: `/usr/local/`,
	}
)

func InstallExternal(service string) error {

	l.Info("Comming soon")
	return nil

	name := strings.ToLower(service)
	dir := runtime.GOOS + "/" + runtime.GOARCH

	if _, ok := ExternalInstallDir[dir]; !ok {
		return fmt.Errorf("%v/%v not suppotrted", runtime.GOOS, runtime.GOARCH)
	}

	switch name {
	case "telegraf":
		return InstallTelegraf(ExternalInstallDir[dir])
	default:
		return fmt.Errorf("Unsupport install %s", service)
	}

	return nil
}