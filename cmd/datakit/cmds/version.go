package cmds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	nhttp "net/http"
	"path"
	"runtime"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/version"
)

//nolint:lll
const (
	winUpgradeCmd      = `$env:DK_UPGRADE="1"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source %s -destination .install.ps1; powershell .install.ps1;`
	winUpgradeCmdProxy = `$env:HTTPS_PROXY="%s"; $env:DK_UPGRADE="1"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -ProxyUsage Override -ProxyList $env:HTTP_PROXY -source %s -destination .install.ps1; powershell .install.ps1;`

	unixUpgradeCmd      = `DK_UPGRADE=1 bash -c "$(curl -L %s)"`
	unixUpgradeCmdProxy = `HTTPS_PROXY="%s" DK_UPGRADE=1 bash -c "$(curl -x "%s" -L %s)"`
)

func checkUpdate(curverStr string, acceptRC bool) int {
	l = logger.SLogger("ota-update")

	l.Debugf("get online version...")
	vers, err := getOnlineVersions(false)
	if err != nil {
		l.Errorf("Get online version failed: \n%s\n", err.Error())
		return 0
	}

	ver := vers["Online"]

	curver, err := getLocalVersion(curverStr)
	if err != nil {
		l.Errorf("Get online version failed: \n%s\n", err.Error())
		return -1
	}

	l.Debugf("online version: %v, local version: %v", ver, curver)

	if ver != nil && version.IsNewVersion(ver, curver, acceptRC) {
		l.Infof("New online version available: %s, commit %s (release at %s)",
			ver.VersionString, ver.Commit, ver.ReleaseDate)
		return 42 // nolint
	} else {
		if acceptRC {
			l.Infof("Up to date(%s)", curver.VersionString)
		} else {
			l.Infof("Up to date(%s), RC version skipped", curver.VersionString)
		}
	}
	return 0
}

func showVersion(curverStr, releaseType string) {
	fmt.Printf(`
       Version: %s
        Commit: %s
        Branch: %s
 Build At(UTC): %s
Golang Version: %s
      Uploader: %s
ReleasedInputs: %s
     InstallAt: %s
     UpgradeAt: %s
`, curverStr, git.Commit, git.Branch, git.BuildAt, git.Golang, git.Uploader,
		releaseType, config.Cfg.InstallDate, func() string {
			if config.Cfg.UpgradeDate.Unix() < 0 {
				return "not upgraded"
			}
			return fmt.Sprintf("%v", config.Cfg.UpgradeDate)
		}())
}

type newVersionInfo struct {
	versionType string
	upgrade     bool
	install     bool
	newVersion  *version.VerInfo
}

func (vi *newVersionInfo) String() string {
	return fmt.Sprintf("%s/%v/%v\n", vi.versionType, vi.upgrade, vi.install) + func() string {
		if vi.upgrade {
			return getUpgradeCommand(vi.newVersion.DownloadURL)
		} else {
			return getInstallCommand()
		}
	}()
}

func checkNewVersion(curverStr string, showTestingVer bool) (map[string]*newVersionInfo, error) {
	vers, err := getOnlineVersions(showTestingVer)
	if err != nil {
		return nil, fmt.Errorf("getOnlineVersions: %w", err)
	}

	curver, err := getLocalVersion(curverStr)
	if err != nil {
		return nil, fmt.Errorf("getLocalVersion: %w", err)
	}

	vis := map[string]*newVersionInfo{}

	for k, v := range vers {
		// always show testing version if showTestingVer is true
		l.Debugf("compare %s <=> %s", v, curver)

		if version.IsNewVersion(v, curver, true) {
			vis[k] = &newVersionInfo{
				versionType: k,
				upgrade:     true,
				newVersion:  v,
			}
		}
	}
	return vis, nil
}

const (
	versionTypeOnline  = "Online"
	versionTypeTesting = "Testing"
)

var versionInfos = map[string]string{
	versionTypeOnline:  "static.guance.com/datakit",
	versionTypeTesting: "zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/datakit",
}

func getInstallCommand() string {
	return "Current is unstable version, please reinstall DataKit."
}

func getUpgradeCommand(dlurl string) string {
	proxy := config.Cfg.DataWay.HTTPProxy
	var upgradeCmd string

	switch runtime.GOOS {
	case datakit.OSWindows:
		if proxy != "" {
			upgradeCmd = fmt.Sprintf(winUpgradeCmdProxy, proxy, dlurl)
		} else {
			upgradeCmd = fmt.Sprintf(winUpgradeCmd, dlurl)
		}

	default: // Linux/Mac

		if proxy != "" {
			upgradeCmd = fmt.Sprintf(unixUpgradeCmdProxy, proxy, proxy, dlurl)
		} else {
			upgradeCmd = fmt.Sprintf(unixUpgradeCmd, dlurl)
		}
	}

	return upgradeCmd
}

func getLocalVersion(ver string) (*version.VerInfo, error) {
	v := &version.VerInfo{
		VersionString: strings.TrimPrefix(ver, "v"),
		Commit:        git.Commit,
		ReleaseDate:   git.BuildAt,
	}
	if err := v.Parse(); err != nil {
		return nil, err
	}
	return v, nil
}

func getVersion(addr string) (*version.VerInfo, error) {
	cli := getcli()

	req, err := nhttp.NewRequest("GET", "http://"+path.Join(addr, "version"), nil)
	if err != nil {
		return nil, err
	}

	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close() //nolint:errcheck
	infobody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ver version.VerInfo
	if err = json.Unmarshal(infobody, &ver); err != nil {
		return nil, err
	}

	if err := ver.Parse(); err != nil {
		return nil, err
	}

	ver.DownloadURL = fmt.Sprintf("https://%s/install.sh", addr)

	if runtime.GOOS == datakit.OSWindows {
		ver.DownloadURL = fmt.Sprintf("https://%s/install.ps1", addr)
	}
	return &ver, nil
}

func getOnlineVersions(showTestingVer bool) (map[string]*version.VerInfo, error) {
	res := map[string]*version.VerInfo{}
	for k, v := range versionInfos {
		if k == versionTypeTesting && !showTestingVer {
			continue
		}

		vi, err := getVersion(v)
		if err != nil {
			return nil, err
		}

		res[k] = vi
		l.Debugf("get %s version: %s", k, vi)
	}

	return res, nil
}
