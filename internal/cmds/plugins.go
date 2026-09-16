// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"fmt"
	"runtime"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var ExternalInstallDir = map[string]string{
	datakit.OSArchWinAmd64:    `C:\Program Files\`,
	datakit.OSArchWin386:      `C:\Program Files (x86)\`,
	datakit.OSArchLinuxArm:    `/usr/local/`,
	datakit.OSArchLinuxArm64:  `/usr/local/`,
	datakit.OSArchLinuxAmd64:  `/usr/local/`,
	datakit.OSArchLinux386:    `/usr/local/`,
	datakit.OSArchDarwinAmd64: `/usr/local/`,
	datakit.OSArchDarwinArm64: `/usr/local/`,
}

type InstallOptions struct {
	LogPath     string
	Telegraf    bool
	Scheck      bool
	IPDB        string
	SymbolTools bool
}

func RunInstall(opts InstallOptions) error {
	ConfigureCommandLog(opts.LogPath)

	dir := runtime.GOOS + "/" + runtime.GOARCH

	if _, ok := ExternalInstallDir[dir]; !ok {
		return fmt.Errorf("%v/%v not suppotrted", runtime.GOOS, runtime.GOARCH)
	}

	switch {
	case opts.Telegraf:
		return installTelegraf(ExternalInstallDir[dir])
	case opts.Scheck:
		return installScheck()
	case opts.IPDB != "":
		switch opts.IPDB {
		case "iploc":
			return installIPDB("iploc")
		case "geolite2":
			return installIPDB("geolite2")
		default:
			return fmt.Errorf("unknown ipdb `%s'", opts.IPDB)
		}
	case opts.SymbolTools:
		return InstallSymbolTools()
	default:
		return fmt.Errorf("unknown package or plugin")
	}
}
