// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"crypto/tls"
	"net/http"
	"net/url"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpcli"
)

var (
	ReleaseVersion    string
	InputsReleaseType string
	Lite              bool
	ELinker           bool
	windowsCmdErrMsg  = "Stop-Service -Name datakit"
	darwinCmdErrMsg   = "sudo launchctl unload -w /Library/LaunchDaemons/com.datakit.plist"
	linuxCmdErrMsg    = "systemctl stop datakit"

	errMsg = map[string]string{
		datakit.OSWindows: windowsCmdErrMsg,
		datakit.OSLinux:   linuxCmdErrMsg,
		datakit.OSDarwin:  darwinCmdErrMsg,
	}
)

func tryLoadMainCfg() {
	if err := config.Cfg.LoadMainTOML(datakit.MainConfPath); err != nil {
		cp.Warnf("[W] load config %s failed: %s, ignored\n", datakit.MainConfPath, err)
	}

	config.Cfg.SetCommandLineMode(true)
}

func GetHTTPClient(proxy string) *http.Client {
	cliopt := &httpcli.Options{
		// ignore SSL error
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint
	}

	if proxy != "" {
		if u, err := url.Parse(proxy); err == nil {
			cliopt.ProxyURL = u
		}
	}

	return httpcli.Cli(cliopt)
}
