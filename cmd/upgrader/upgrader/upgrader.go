// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package upgrader

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/kardianos/service"
	hostutil "github.com/shirou/gopsutil/host"
	"go.uber.org/atomic"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
)

type pingInfo struct {
	Content httpapi.Ping `json:"content"`
}

var Docker = false

const (
	statusNoUpgrade = 0
	statusUpgrading = 1
)

type upgraderImpl struct {
	c             *MainConfig
	upgradeStatus *atomic.Int32
}

type upgradeOptions struct {
	version string
	force   bool
}

type upgradeOpt func(*upgradeOptions)

var checkOSCompatibility = isHostV2Compatible

func withVersion(v string) upgradeOpt {
	return func(uo *upgradeOptions) {
		uo.version = v
	}
}

func withForce(on bool) upgradeOpt {
	return func(uo *upgradeOptions) {
		uo.force = on
	}
}

func (u *upgraderImpl) upgrade(opts ...upgradeOpt) error {
	uo := &upgradeOptions{}

	for _, opt := range opts {
		opt(uo)
	}

	l.Debugf("upgrade options: %+#v", uo)

	if !u.upgradeStatus.CompareAndSwap(statusNoUpgrade, statusUpgrading) {
		return httpapi.ErrIsUpgrading
	}

	if u.c == nil {
		u.upgradeStatus.Store(statusNoUpgrade)
		return fmt.Errorf("upgrader config not initialized")
	}

	defer u.upgradeStatus.Store(statusNoUpgrade)

	var (
		dkv         *pingInfo
		err         error
		baseURL     = "https://" + cmds.StaticCDN
		upToDate    = false
		downloadURL = ""
		scriptBase  = ""
	)

	if u.c.InstallerBaseURL != "" {
		baseURL = u.c.InstallerBaseURL
	}

	if uo.force { // force upgrade/downgrade current DK version
		dkv = &pingInfo{
			Content: httpapi.Ping{}, // fake ping info
		}
	} else {
		dkv, err = u.fetchCurrentDKVersion()
		if err != nil {
			return uhttp.Errorf(httpapi.ErrUpgradeFailed, "get datakit version failed: %s", err.Error())
		}

		l.Infof("current DK version: %s, commit: %s", dkv.Content.Version, dkv.Content.Commit)
	}

	if uo.version == "" { // auto-upgrade to latest version from current OSS path.
		if v2OK, msg := checkOSCompatibility(); !v2OK {
			l.Warnf("OS may not meet v2 requirements: %s. Upgrade will proceed, but the installer may reject this host.", msg)
		}

		onlineVer, err := cmds.GetOnlineVersions(baseURL, u.c.Proxy, 30*time.Second)
		if err != nil {
			return uhttp.Errorf(httpapi.ErrUpgradeFailed, "unable to get online version: %s", err)
		}

		scriptBase = baseURL

		if uo.force {
			downloadURL = onlineVer.DownloadURL
		} else {
			if onlineVer.Commit != dkv.Content.Commit ||
				onlineVer.VersionString != dkv.Content.Version {
				l.Infof("current version is %q, target online version is %q", dkv.Content.Version, onlineVer.VersionString)
				downloadURL = onlineVer.DownloadURL
			} else {
				upToDate = true
			}
		}
	} else {
		l.Infof("upgrade strategy: specified version %q provided", uo.version)

		if uo.force {
			downloadURL = cmds.CanonicalInstallBaseURL(baseURL)
			scriptBase = downloadURL
		} else {
			// only check if version-string equal, user will not supply Git commit ID in API.
			if dkv.Content.Version != uo.version {
				l.Infof("current version is %q, specified version is %q", dkv.Content.Version, uo.version)
				downloadURL = cmds.CanonicalInstallBaseURL(baseURL)
				scriptBase = downloadURL
			} else {
				upToDate = true
			}
		}
	}

	if downloadURL == "" {
		if upToDate {
			l.Debugf("current version up-to-date according to base URL %s", baseURL)
			return httpapi.ErrDKVersionUptoDate
		} else {
			l.Errorf("new version not found")
			return uhttp.Errorf(httpapi.ErrUpgradeFailed, "can't get newer datakit version")
		}
	}

	scriptFile, err := u.saveUpgradeScript(downloadURL, uo.version)
	if err != nil {
		l.Errorf("unable to download upgrade script: %s", err.Error())
		return err
	}

	if scriptBase == "" {
		scriptBase = baseURL
	}

	if err := u.doUpgrade(scriptFile, scriptBase); err != nil {
		l.Errorf("doUpgrade: %s", err.Error())
		return uhttp.Errorf(httpapi.ErrUpgradeFailed, "doUpgrade: %s", err)
	}

	// If the backened upgrading procedure failed to start the service,
	// we tried here to start it.
	u.tryStartService()

	return nil
}

func (u *upgraderImpl) saveUpgradeScript(downloadURL, version string) (string, error) {
	downloadURL = strings.TrimRight(downloadURL, "/ ")
	scriptExt := ".sh"

	verSuffix := ""
	if version != "" {
		verSuffix = "-" + version
	}

	if runtime.GOOS == datakit.OSWindows {
		downloadURL = fmt.Sprintf("%s/install%s.ps1", downloadURL, verSuffix)
		scriptExt = ".ps1"
	} else {
		downloadURL = fmt.Sprintf("%s/install%s.sh", downloadURL, verSuffix)
	}

	f, err := os.CreateTemp(u.c.InstallDir, fmt.Sprintf("tmp-dk-upgrader-*%s", scriptExt))
	if err != nil {
		return "", uhttp.Errorf(httpapi.ErrUpgradeFailed,
			"unable to create Datakit temporary setup script file: %s",
			err)
	}
	defer f.Close() //nolint

	fileABSPath, err := filepath.Abs(f.Name())
	if err != nil {
		return "", uhttp.Errorf(httpapi.ErrUpgradeFailed,
			"unable to get setup file absolute path: %s", err)
	}

	if err := f.Truncate(0); err != nil {
		return "", uhttp.Errorf(httpapi.ErrUpgradeFailed,
			"unable to truncate DK setup temp file %s: %s",
			fileABSPath,
			err.Error())
	}

	l.Infof("download script from %s", downloadURL)
	resp, err := http.Get(downloadURL) // nolint:gosec
	if err != nil {
		return "", uhttp.Errorf(httpapi.ErrUpgradeFailed,
			"unable to download script file %s: %s",
			downloadURL,
			err)
	}

	if resp.StatusCode/100 != 2 {
		return "", uhttp.Errorf(httpapi.ErrUpgradeFailed,
			"unable to download script file %s: resonse status: %s",
			downloadURL,
			resp.Status)
	}
	defer resp.Body.Close() // nolint:errcheck

	if n, err := io.Copy(f, resp.Body); err != nil {
		return "", uhttp.Errorf(httpapi.ErrUpgradeFailed, "io.Copy(%s): %s", fileABSPath, err)
	} else {
		l.Infof("save %s(bytes: %d)", fileABSPath, n)
	}

	return fileABSPath, nil
}

func (u *upgraderImpl) tryStartService() {
	svc, err := dkservice.NewService()
	if err != nil {
		l.Warnf("new %s service failed: %s, ignored", runtime.GOOS, err.Error())
	}

	l.Infof("try starting Datakit...")
	if err = service.Control(svc, "start"); err != nil {
		l.Warnf("stop service failed %s, ignored", err.Error())
	} else {
		l.Infof("Start datakit ok")
	}
}

func (u *upgraderImpl) forceStopService() error {
	svc, err := dkservice.NewService()
	if err != nil {
		l.Warnf("new %s service failed: %s, ignored", runtime.GOOS, err.Error())
	}

	if status, err := svc.Status(); err == nil {
		if status == service.StatusRunning {
			l.Infof("stopping running datakit...")
			if err = service.Control(svc, "stop"); err != nil {
				l.Warnf("stop service failed: %s, ignored", err.Error())
				return err
			}
		}
	} else {
		l.Warnf("get status of datakit service failed: %s, ignored", err.Error())
	}

	return nil
}

func (u *upgraderImpl) restartService() error {
	if err := cmds.RestartDatakit(); err != nil {
		return err
	}

	return nil
}

func (u *upgraderImpl) doUpgrade(scriptFile, installBaseURL string) error {
	// Force stop current running datakit service.
	// dk-install may failed to stop(why?) the datakit service, and during
	// download new version datakit binary, we'll get `text file busy' error.
	//
	// I don't know why backed-started upgrade procedure failed to operate on datakit service,
	// such as get current service status/stop service/start service. So it's better to start/stop
	// datakit service within dk_upgrader, not within the installer.
	if err := u.forceStopService(); err != nil {
		return err
	}

	shell := "bash"
	args := []string{scriptFile}
	if runtime.GOOS == datakit.OSWindows {
		shell = "powershell"
		// Powershell can not invoke a script at a path with blanks
		// see: https://stackoverflow.com/questions/18537098/spaces-cause-split-in-path-with-powershell
		args = []string{
			"-c",
			fmt.Sprintf(`Set-ExecutionPolicy Bypass -scope Process -Force;& "%s"`, scriptFile),
		}
	}

	shellBin, err := exec.LookPath(shell)
	if err != nil {
		return fmt.Errorf("%s command not found: %w", shell, err)
	}

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	cmd := exec.Command(shellBin, args...) // nolint:gosec
	cmd.Dir = filepath.Dir(scriptFile)
	cmd.Stderr = stderr
	cmd.Stdout = stdout

	envs := os.Environ()
	envs = append(envs, "DK_UPGRADE=1")
	envs = append(envs, fmt.Sprintf("TMPDIR=%s", cmd.Dir))

	if u.c.Proxy != "" {
		envs = append(envs, "HTTPS_PROXY="+u.c.Proxy)
	}

	if installBaseURL != "" {
		envs = append(envs, fmt.Sprintf("DK_INSTALLER_BASE_URL=%s",
			strings.TrimRight(cmds.CanonicalInstallBaseURL(installBaseURL), "/")))
	}

	cmd.Env = envs

	l.Infof("run upgrade script envs: %s", formatUpgradeEnvsForLog(cmd.Env))

	l.Infof("datakit manager will start execute upgrade cmd: %s", cmd.String())
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to execute upgrade cmd[%s]: %w", cmd.String(), err)
	}

	err = cmd.Wait()
	if x := stdout.String(); x != "" {
		l.Infof("upgrade process stdout:\n%s\n", x)
	}

	if x := stderr.String(); x != "" {
		l.Errorf("upgrade process stderr:\n%s\n", x)
	}

	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return fmt.Errorf("upgrade process exit abnormal: %s, err: %w, stdout:%s, stderr: %s",
				ee.ProcessState.String(), ee, stdout.String(), stderr.String())
		}
		return fmt.Errorf("upgrade process execute fail: %w, stdout:%s, stderr: %s", err, stdout.String(), stderr.String())
	}

	return nil
}

func formatUpgradeEnvsForLog(envs []string) string {
	formatted := make([]string, 0, len(envs))
	for _, env := range envs {
		key, _, ok := strings.Cut(env, "=")
		if !ok {
			formatted = append(formatted, env)
			continue
		}

		if isSensitiveEnvKey(key) {
			formatted = append(formatted, key+"=******")
			continue
		}

		formatted = append(formatted, env)
	}

	return strings.Join(formatted, "\n\t")
}

func isSensitiveEnvKey(key string) bool {
	key = strings.ToUpper(key)
	for _, pattern := range []string{"TOKEN", "KEY", "SECRET", "PASSWORD", "PASSWD", "PROXY", "AUTH", "CREDENTIAL"} {
		if strings.Contains(key, pattern) {
			return true
		}
	}

	return false
}

func parseMajorMinor(v string) (int, int, error) {
	x := strings.TrimSpace(v)
	if x == "" {
		return 0, 0, fmt.Errorf("empty version")
	}

	parts := strings.Split(x, ".")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("invalid version %q", v)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parse major from %q failed: %w", v, err)
	}

	minorPart := strings.Builder{}
	for _, ch := range parts[1] {
		if ch >= '0' && ch <= '9' {
			minorPart.WriteRune(ch)
		} else {
			break
		}
	}

	if minorPart.Len() == 0 {
		return 0, 0, fmt.Errorf("parse minor from %q failed", v)
	}

	minor, err := strconv.Atoi(minorPart.String())
	if err != nil {
		return 0, 0, fmt.Errorf("parse minor from %q failed: %w", v, err)
	}

	return major, minor, nil
}

func isHostV2Compatible() (bool, string) {
	info, err := hostutil.Info()
	if err != nil {
		return false, fmt.Sprintf("host detection failed(%s)", err)
	}

	switch runtime.GOOS {
	case datakit.OSLinux:
		major, minor, err := parseMajorMinor(info.KernelVersion)
		if err != nil {
			return false, fmt.Sprintf("linux kernel %q parse failed(%s)", info.KernelVersion, err)
		}

		if major > 3 || (major == 3 && minor >= 2) {
			return true, fmt.Sprintf("linux kernel=%s >= 3.2, meets v2 requirement", info.KernelVersion)
		}
		return false, fmt.Sprintf("linux kernel=%s < 3.2, below v2 requirement", info.KernelVersion)

	case datakit.OSWindows:
		major, _, err := parseMajorMinor(info.PlatformVersion)
		if err != nil {
			return false, fmt.Sprintf("windows platform_version %q parse failed(%s)", info.PlatformVersion, err)
		}

		if major >= 10 {
			return true, fmt.Sprintf("windows platform_version=%s (Win10/Server2016+), meets v2 requirement", info.PlatformVersion)
		}
		return false, fmt.Sprintf("windows platform_version=%s (<10), below v2 requirement", info.PlatformVersion)

	case datakit.OSDarwin:
		major, _, err := parseMajorMinor(info.PlatformVersion)
		if err != nil {
			return false, fmt.Sprintf("macOS platform_version %q parse failed(%s)", info.PlatformVersion, err)
		}

		if major >= 12 {
			return true, fmt.Sprintf("macOS platform_version=%s >= 12, meets v2 requirement", info.PlatformVersion)
		}
		return false, fmt.Sprintf("macOS platform_version=%s < 12, below v2 requirement", info.PlatformVersion)

	default:
		return false, fmt.Sprintf("unsupported OS %q, unable to check v2 compatibility", runtime.GOOS)
	}
}

func dkPing(dklisten string, https bool) ([]byte, error) {
	var (
		cli    = cmds.GetHTTPClient("")
		getURL = fmt.Sprintf("http://%s/v1/ping", dklisten)
	)

	if https {
		getURL = fmt.Sprintf("https://%s/v1/ping", dklisten)
	}

	req, err := http.NewRequest("GET", getURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to query current Datakit version: %w", err)
	}
	defer resp.Body.Close() //nolint: errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read Datakit ping result: %w", err)
	}

	l.Debugf("ping body from %s: %q", getURL, body)

	return body, err
}

func (u *upgraderImpl) fetchCurrentDKVersion() (*pingInfo, error) {
	body, err := dkPing(u.c.DatakitAPIListen, u.c.DatakitAPIHTTPS)
	if err != nil {
		return nil, err
	}

	var pi pingInfo
	if err := json.Unmarshal(body, &pi); err != nil {
		return nil, fmt.Errorf("unmarshal datakit ping response: %w", err)
	}

	return &pi, nil
}
