// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

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
		"telegraf", "scheck", "ebpf",
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
			return installIpdb("iploc")
		default:
			return fmt.Errorf("unknown ipdb `%s'", *flagInstallIPDB)
		}
	case *flagInstallEbpf:
		return InstallEbpf()
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
