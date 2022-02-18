package cmds

import (
	"fmt"
	"runtime"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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

	availablePlugins = []string{
		"telegraf", "scheck",
	}
)

func installPlugins() error {
	dir := runtime.GOOS + "/" + runtime.GOARCH

	if _, ok := ExternalInstallDir[dir]; !ok {
		return fmt.Errorf("%v/%v not suppotrted", runtime.GOOS, runtime.GOARCH)
	}

	switch {
	case *flagInstallTelegraf:
		return installTelegraf(ExternalInstallDir[dir])
	case *flagInstallScheck:
		return installScheck()
	case *flagInstallIPDB != "":
		switch *flagInstallIPDB {
		case "iploc":
			// TODO: add another ipdb.go to install ipdb files
			return nil
		default:
			return fmt.Errorf("unknown ipdb `%s'", *flagInstallIPDB)
		}
	default:
		return fmt.Errorf("unknown package or plugin")
	}
}

// Deprecated: old flag handler.
func installExternal(service string) error {
	name := strings.ToLower(service)
	dir := runtime.GOOS + "/" + runtime.GOARCH

	if _, ok := ExternalInstallDir[dir]; !ok {
		return fmt.Errorf("%v/%v not suppotrted", runtime.GOOS, runtime.GOARCH)
	}

	switch name {
	case "telegraf":
		return installTelegraf(ExternalInstallDir[dir])
	case "sec-checker", // deprecated
		"scheck":
		return installScheck()
	default:
		return fmt.Errorf("unsupport install %s(available plugins: %s)",
			service, strings.Join(availablePlugins, "/"))
	}
}
