package cmds

import (
	"net/http"
	"net/url"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
)

var (
	ReleaseVersion    string
	InputsReleaseType string
	windowsCmdErrMsg  = "Stop-Service -Name datakit"
	darwinCmdErrMsg   = "sudo launchctl unload -w /Library/LaunchDaemons/cn.dataflux.datakit.plist"
	linuxCmdErrMsg    = "systemctl stop datakit"

	errMsg = map[string]string{
		datakit.OSWindows: windowsCmdErrMsg,
		datakit.OSLinux:   linuxCmdErrMsg,
		datakit.OSDarwin:  darwinCmdErrMsg,
	}
)

func tryLoadMainCfg() {
	if err := config.Cfg.LoadMainTOML(datakit.MainConfPath); err != nil {
		warnf("[W] load config %s failed: %s, ignored\n", datakit.MainConfPath, err)
	}
}

func getcli() *http.Client {
	proxy := config.Cfg.DataWay.HTTPProxy

	cliopt := &ihttp.Options{
		InsecureSkipVerify: true,
	}

	if proxy != "" {
		if u, err := url.Parse(proxy); err == nil {
			cliopt.ProxyURL = u
		}
	}

	return ihttp.Cli(cliopt)
}
