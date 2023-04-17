// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	nhttp "net/http"
	"runtime"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/version"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/man"
)

func runVersionFlags(disableUpgradeInfo bool) error {
	showVersion(ReleaseVersion)

	if !disableUpgradeInfo {
		vis, err := CheckNewVersion(ReleaseVersion)
		if err != nil {
			return err
		}

		for _, vi := range vis {
			cp.Infof("\n\n%s version available: %s, commit %s (release at %s)\n\nUpgrade:\n",
				vi.versionType, vi.NewVersion.VersionString, vi.NewVersion.Commit, vi.NewVersion.ReleaseDate)
			cp.Infof("%s\n", getUpgradeCommand(runtime.GOOS, vi.NewVersion.DownloadURL, config.Cfg.Dataway.HTTPProxy))
		}
	}

	return nil
}

func showVersion(curverStr string) {
	fmt.Printf(`
       Version: %s
        Commit: %s
        Branch: %s
 Build At(UTC): %s
Golang Version: %s
`, curverStr, git.Commit, git.Branch, git.BuildAt, git.Golang)
}

type newVersionInfo struct {
	versionType string
	upgrade     bool
	install     bool
	NewVersion  *version.VerInfo
}

func (vi *newVersionInfo) String() string {
	if vi.NewVersion == nil {
		return ""
	}

	return fmt.Sprintf("%s/%v/%v\n%s",
		vi.versionType,
		vi.upgrade,
		vi.install,
		getUpgradeCommand(runtime.GOOS, vi.NewVersion.DownloadURL, config.Cfg.Dataway.HTTPProxy))
}

func CheckNewVersion(curverStr string) (map[string]*newVersionInfo, error) {
	vers, err := GetOnlineVersions()
	if err != nil {
		return nil, fmt.Errorf("GetOnlineVersions: %w", err)
	}

	curver, err := getLocalVersion(curverStr)
	if err != nil {
		return nil, fmt.Errorf("getLocalVersion: %w", err)
	}

	vis := map[string]*newVersionInfo{}

	for k, v := range vers {
		l.Debugf("compare %s <=> %s", v, curver)

		if version.IsNewVersion(v, curver, true) {
			vis[k] = &newVersionInfo{
				versionType: k,
				upgrade:     true,
				NewVersion:  v,
			}
		}
	}
	return vis, nil
}

const (
	versionTypeOnline = "Online"
)

func getUpgradeCommand(os, dlurl, proxy string) string {
	p := &man.Params{}

	cmd := man.InstallCommand(
		p.WithUpgrade(true),
		p.WithIndent(4),
		p.WithSourceURL(dlurl))

	if proxy != "" {
		p.WithEnvs("HTTPS_PROXY", proxy)(cmd)
	}

	switch os {
	case "windows":
		p.WithPlatform("windows")(cmd)

		// nolint:lll
		// Add proxy settings for bitstransfer. See:
		//	 https://learn.microsoft.com/en-us/powershell/module/bitstransfer/start-bitstransfer?view=windowsserver2022-ps&viewFallbackFrom=winserver2012-ps
		if proxy != "" {
			p.WithBitstransferOpts("-ProxyUsage Override -ProxyList $env:HTTPS_PROXY")(cmd)
		}
	default:
		p.WithPlatform("unix")(cmd)
	}

	return cmd.String()
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
	cli.Timeout = time.Second * 5
	urladdr := addr + "/version"

	req, err := nhttp.NewRequest("GET", urladdr, nil)
	if err != nil {
		return nil, fmt.Errorf("http new request err=%w", err)
	}

	resp, err := cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http Do request err=%w", err)
	}

	defer resp.Body.Close() //nolint:errcheck
	infobody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("http read body err=%w", err)
	}

	var ver version.VerInfo
	if err = json.Unmarshal(infobody, &ver); err != nil {
		return nil, fmt.Errorf("json unmarshal err=%w", err)
	}

	if err := ver.Parse(); err != nil {
		return nil, err
	}

	ver.DownloadURL = addr

	return &ver, nil
}

func GetOnlineVersions() (map[string]*version.VerInfo, error) {
	res := map[string]*version.VerInfo{}

	if v := datakit.GetEnv("DK_INSTALLER_BASE_URL"); v != "" {
		cp.Warnf("setup base URL to %s\n", v)
		OnlineBaseURL = v
	}

	versionInfos := map[string]string{
		versionTypeOnline: (OnlineBaseURL + "/datakit"),
	}

	for k, v := range versionInfos {
		vi, err := getVersion(v)
		if err != nil {
			return nil, fmt.Errorf("get version from %s failed: %w", v, err)
		}
		res[k] = vi
		l.Debugf("get %s version: %s", k, vi)
	}

	return res, nil
}
