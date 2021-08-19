package cmds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	nhttp "net/http"
	"os"
	"path"
	"runtime"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/version"
)

const (
	winUpgradeCmd      = `$env:DK_UPGRADE="1"; Import-Module bitstransfer; start-bitstransfer -source %s -destination .install.ps1; powershell .install.ps1;`
	winUpgradeCmdProxy = `$env:HTTPS_PROXY="%s"; $env:DK_UPGRADE="1"; Import-Module bitstransfer; start-bitstransfer -ProxyUsage Override -ProxyList $env:HTTP_PROXY -source %s -destination .install.ps1; powershell .install.ps1;`

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
		return 42
	} else {
		if acceptRC {
			l.Infof("Up to date(%s)", curver.VersionString)
		} else {
			l.Infof("Up to date(%s), RC version skipped", curver.VersionString)
		}
	}
	return 0
}

func showVersion(curverStr, releaseType string, showTestingVer bool) {
	fmt.Printf(`
       Version: %s
        Commit: %s
        Branch: %s
 Build At(UTC): %s
Golang Version: %s
      Uploader: %s
ReleasedInputs: %s
`, curverStr, git.Commit, git.Branch, git.BuildAt, git.Golang, git.Uploader, releaseType)
	vers, err := getOnlineVersions(showTestingVer)
	if err != nil {
		fmt.Printf("Get online version failed: \n%s\n", err.Error())
		os.Exit(-1)
	}
	curver, err := getLocalVersion(curverStr)
	if err != nil {
		fmt.Printf("Get local version failed: \n%s\n", err.Error())
		os.Exit(-1)
	}

	for k, v := range vers {

		// always show testing verison if showTestingVer is true
		l.Debugf("compare %s <=> %s", v, curver)
		if k == "Testing" || version.IsNewVersion(v, curver, true) { // show version info, also show RC verison info
			fmt.Println("---------------------------------------------------")
			fmt.Printf("\n\n%s version available: %s, commit %s (release at %s)\n\nUpgrade:\n\t",
				k, v.VersionString, v.Commit, v.ReleaseDate)

			fmt.Println(getUpgradeCommand(v.DownloadURL))
		}
	}
}

func getUpgradeCommand(dlurl string) string {
	upgradeFmt := ""
	proxy := config.Cfg.DataWay.HttpProxy
	switch runtime.GOOS {
	case "windows":
		if proxy != "" {
			upgradeFmt = fmt.Sprintf(winUpgradeCmdProxy, proxy, dlurl)
		} else {
			upgradeFmt = fmt.Sprintf(winUpgradeCmd, dlurl)
		}

	default: // Linux/Mac

		if proxy != "" {
			upgradeFmt = fmt.Sprintf(unixUpgradeCmdProxy, proxy, proxy, dlurl)
		} else {
			upgradeFmt = fmt.Sprintf(unixUpgradeCmd, dlurl)
		}
	}

	return upgradeFmt
}

func getLocalVersion(ver string) (*version.VerInfo, error) {
	v := &version.VerInfo{
		VersionString: strings.TrimPrefix(ver, "v"),
		Commit:        git.Commit,
		ReleaseDate:   git.BuildAt}
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

	defer resp.Body.Close()
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

	if runtime.GOOS == "windows" {
		ver.DownloadURL = fmt.Sprintf("https://%s/install.ps1", addr)
	}
	return &ver, nil
}

func getOnlineVersions(showTestingVer bool) (res map[string]*version.VerInfo, err error) {

	res = map[string]*version.VerInfo{}

	onlineVer, err := getVersion("static.dataflux.cn/datakit")
	if err != nil {
		return nil, err
	}
	res["Online"] = onlineVer
	l.Debugf("online version: %s", onlineVer)

	if showTestingVer {
		testVer, err := getVersion("zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/datakit")
		if err != nil {
			return nil, err
		}
		res["Testing"] = testVer
		l.Debugf("testing version: %s", testVer)
	}

	return
}
