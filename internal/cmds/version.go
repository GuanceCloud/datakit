// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"encoding/json"
	"fmt"
	"io"
	nhttp "net/http"
	"runtime"
	"strings"
	"time"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/version"
)

func runVersionFlags(disableUpgradeInfo bool) error {
	showVersion(ReleaseVersion)

	if !disableUpgradeInfo {
		vi, err := CheckNewVersion(ReleaseVersion)
		if err != nil {
			return err
		}

		if vi != nil {
			cp.Infof("\n\n%s version available: %s, commit %s (release at %s)\n\nUpgrade:\n",
				vi.versionType,
				vi.NewVersion.VersionString,
				vi.NewVersion.Commit,
				vi.NewVersion.ReleaseDate)

			cp.Infof("%s\n", getUpgradeCommand(runtime.GOOS,
				vi.NewVersion.DownloadURL,
				config.Cfg.Dataway.HTTPProxy))
		}
	}

	return nil
}

func showVersion(curverStr string) {
	buildTag := "full"
	if Lite {
		buildTag = "lite"
	} else if ELinker {
		buildTag = "elinker"
	}

	cp.Printf(`
       Version: %s
        Commit: %s
        Branch: %s
 Build At(UTC): %s
Golang Version: %s
     Build Tag: %s
`, curverStr, git.Commit, git.Branch, git.BuildAt, git.Golang, buildTag)
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

func CheckNewVersion(curverStr string) (*newVersionInfo, error) {
	proxy := ""
	if config.Cfg.Dataway.HTTPProxy != "" {
		proxy = config.Cfg.Dataway.HTTPProxy
	}

	ver, err := GetOnlineVersions(OnlineBaseURL, proxy, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("GetOnlineVersions: %w", err)
	}

	curver, err := getLocalVersion(curverStr)
	if err != nil {
		return nil, fmt.Errorf("getLocalVersion: %w", err)
	}

	var vis *newVersionInfo

	l.Debugf("compare %s <=> %s", ver, curver)

	if version.IsNewVersion(ver, curver, true) {
		vis = &newVersionInfo{
			versionType: versionTypeOnline,
			upgrade:     true,
			NewVersion:  ver,
		}
	}

	return vis, nil
}

const (
	versionTypeOnline = "Online"
)

func getUpgradeCommand(os, dlurl, proxy string) string {
	p := &export.Params{}

	cmd := export.InstallCommand(
		p.WithUpgrade(true),
		p.WithIndent(4),
		p.WithSourceURL(dlurl))

	if proxy != "" {
		p.WithEnvs("HTTPS_PROXY", proxy)(cmd)
		p.WithProxy(true)(cmd)
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

func getVersionInfo(addr, proxy string, timeout time.Duration) (*version.VerInfo, error) {
	cli := GetHTTPClient(proxy)
	if timeout > 0 {
		cli.Timeout = timeout
	}

	urladdr := addr
	if strings.HasSuffix(addr, "/") {
		urladdr += "version"
	} else {
		urladdr += "/version"
	}

	req, err := nhttp.NewRequest("GET", urladdr, nil)
	if err != nil {
		return nil, fmt.Errorf("http new request err=%w", err)
	}

	resp, err := cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http Do request err=%w", err)
	}

	defer resp.Body.Close() //nolint:errcheck
	infobody, err := io.ReadAll(resp.Body)
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

// CanonicalInstallBaseURL add support for install_base_url with the suffix "/datakit" or not.
// The canonical install_base_url ends with "/datakit/".
func CanonicalInstallBaseURL(installBaseURL string) string {
	suffix := "/datakit/"
	sb := &strings.Builder{}
	sb.Grow(len(installBaseURL) + len(suffix))
	sb.WriteString(installBaseURL)

	if !strings.HasSuffix(installBaseURL, "/") {
		sb.WriteByte('/')
	}

	if !strings.HasSuffix(sb.String(), suffix) {
		sb.WriteString("datakit/")
	}

	return sb.String()
}

func GetOnlineVersions(baseURL, proxy string, timeout time.Duration) (*version.VerInfo, error) {
	vi, err := getVersionInfo(CanonicalInstallBaseURL(baseURL), proxy, timeout)
	if err != nil {
		return nil, fmt.Errorf("get version from %s failed: %w", baseURL, err)
	}
	l.Debugf("get %s version: %s", versionTypeOnline, vi)

	return vi, nil
}
