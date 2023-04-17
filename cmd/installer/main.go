// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/kardianos/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/installer/installer"
	upgrader2 "gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/installer/upgrader"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/upgrader/upgrader"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	dl "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/downloader"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
)

var (
	oldInstallDir      = "/usr/local/cloudcare/dataflux/datakit"
	oldInstallDirWin   = `C:\Program Files\dataflux\datakit`
	oldInstallDirWin32 = `C:\Program Files (x86)\dataflux\datakit`

	DataKitBaseURL = ""
	DataKitVersion = ""
	dataURL        = "https://" + path.Join(DataKitBaseURL, "data.tar.gz")
	dkUpgraderURL  = "https://" + path.Join(DataKitBaseURL,
		fmt.Sprintf("%s-%s-%s.tar.gz", upgrader.BuildBinName, runtime.GOOS, runtime.GOARCH))
	datakitURL = "https://" + path.Join(DataKitBaseURL,
		fmt.Sprintf("datakit-%s-%s-%s.tar.gz",
			runtime.GOOS,
			runtime.GOARCH,
			DataKitVersion))
	InstallerBaseURL = ""
	l                = logger.DefaultSLogger("installer")
)

// Installer flags.
var (
	flagDKUpgrade,
	flagOffline,
	flagDownloadOnly,
	flagInfo bool

	flagUpgradeServIPWhiteList,
	flagUserName,
	flagInstallLog,
	flagSrc string

	flagUpgradeManagerService,
	flagInstallOnly int
)

const (
	datakitBin = "datakit"
)

//nolint:gochecknoinits,lll
func init() {
	flag.BoolVar(&flagDKUpgrade, "upgrade", false, "")
	flag.IntVar(&flagUpgradeManagerService, "upgrade-manager", 0, "whether we should upgrade the Datakit upgrade service")
	flag.StringVar(&flagUpgradeServIPWhiteList, "upgrade-ip-whitelist", "", "set datakit upgrade http service allowed request client ip, split by ','")
	flag.StringVar(&flagInstallLog, "install-log", "install.log", "install log")
	flag.StringVar(&flagSrc, "srcs", fmt.Sprintf("./datakit-%s-%s-%s.tar.gz,./data.tar.gz", runtime.GOOS, runtime.GOARCH, DataKitVersion), `local path of install files`)
	flag.IntVar(&flagInstallOnly, "install-only", 0, "install only, not start")
	flag.BoolVar(&flagInfo, "info", false, "show installer info")
	flag.BoolVar(&flagOffline, "offline", false, "-offline option removed")
	flag.BoolVar(&flagDownloadOnly, "download-only", false, "only download install packages")
	flag.StringVar(&InstallerBaseURL, "installer_base_url", "", "install datakit and data BaseUrl")
	flag.StringVar(&flagUserName, "user-name", "root", "install log") // user & group.

	flag.StringVar(&installer.Dataway, "dataway", "", "DataWay host(https://guance.openway.com?token=xxx)")
	flag.StringVar(&installer.Proxy, "proxy", "", "http proxy http://ip:port for datakit")
	flag.StringVar(&installer.DatakitName, "name", "", "specify DataKit name, example: prod-env-datakit")
	flag.StringVar(&installer.EnableInputs, "enable-inputs", "", "default enable inputs(comma splited, example:cpu,mem,disk)")
	flag.BoolVar(&installer.OTA, "ota", false, "auto update")
	flag.StringVar(&installer.InstallExternals, "install-externals", "", "install some external inputs")

	// DCA flags
	flag.StringVar(&installer.DCAEnable, "dca-enable", "", "enable DCA")
	flag.StringVar(&installer.DCAListen, "dca-listen", "0.0.0.0:9531", "DCA listen address and port")
	flag.StringVar(&installer.DCAWhiteList, "dca-white-list", "", "DCA white list")

	// global-host-tags flags
	flag.StringVar(&installer.GlobalHostTags, "global-tags", "", "Deprecated, use global-host-tag")
	flag.StringVar(&installer.GlobalHostTags, "global-host-tags", "", "enable global host tags, example: host= __datakit_hostname,ip= __datakit_ip")

	// election flags
	flag.StringVar(&installer.GlobalElectionTags, "global-election-tags", "", "enable global environment tags, example: project=my-project,cluster=my-cluster")
	flag.StringVar(&installer.GlobalElectionTags, "global-env-tags", "", "Deprecated, use --global-election-tags")
	flag.StringVar(&installer.EnableElection, "enable-election", "", "datakit election")
	flag.StringVar(&installer.ElectionNamespace, "namespace", "", "datakit namespace")

	// datakit HTTP flags
	flag.IntVar(&installer.HTTPPort, "port", 9529, "datakit HTTP port")
	flag.StringVar(&installer.HTTPListen, "listen", "localhost", "datakit HTTP listen")
	flag.StringVar(&installer.HostName, "env_hostname", "", "host name")
	flag.StringVar(&installer.IPDBType, "ipdb-type", "", "ipdb type")
	flag.StringVar(&installer.CloudProvider, "cloud-provider", "", "specify cloud provider(accept aliyun/tencent/aws)")

	// confd flags
	flag.StringVar(&installer.ConfdBackend, "confd-backend", "", "backend kind")
	flag.StringVar(&installer.ConfdBasicAuth, "confd-basic-auth", "", "if backend need auth")
	flag.StringVar(&installer.ConfdClientCaKeys, "confd-client-ca-keys", "", "backend ca key")
	flag.StringVar(&installer.ConfdClientCert, "confd-client-cert", "", "backend cert key")
	flag.StringVar(&installer.ConfdClientKey, "confd-client-key", "", "backend cert key id")
	flag.StringVar(&installer.ConfdBackendNodes, "confd-backend-nodes", "", "backend nodes ip")
	flag.StringVar(&installer.ConfdPassword, "confd-password", "", "backend login password")
	flag.StringVar(&installer.ConfdScheme, "confd-scheme", "", "backend scheme")
	flag.StringVar(&installer.ConfdSeparator, "confd-separator", "", "backend separator")
	flag.StringVar(&installer.ConfdUsername, "confd-username", "", "backend login username")
	flag.StringVar(&installer.ConfdAccessKey, "confd-access-key", "", "backend access key id")
	flag.StringVar(&installer.ConfdSecretKey, "confd-secret-key", "", "backend secret key")
	flag.StringVar(&installer.ConfdConfdNamespace, "confd-confd-namespace", "", "confd config namespace id")
	flag.StringVar(&installer.ConfdPipelineNamespace, "confd-pipeline-namespace", "", "pipeline config namespace id")
	flag.StringVar(&installer.ConfdRegion, "confd-region", "", "aws region")
	flag.IntVar(&installer.ConfdCircleInterval, "confd-circle-interval", 60, "backend loop search interval second")

	// gitrepo flags
	flag.StringVar(&installer.GitURL, "git-url", "", "git repository url")
	flag.StringVar(&installer.GitKeyPath, "git-key-path", "", "git repository access private key path")
	flag.StringVar(&installer.GitKeyPW, "git-key-pw", "", "git repository access private use password")
	flag.StringVar(&installer.GitBranch, "git-branch", "", "git repository branch name")
	flag.StringVar(&installer.GitPullInterval, "git-pull-interval", "", "git repository pull interval")

	// rum flags
	flag.StringVar(&installer.RumOriginIPHeader, "rum-origin-ip-header", "", "rum only")
	flag.StringVar(&installer.RumDisable404Page, "disable-404page", "", "datakit rum 404 page")

	// log flags
	flag.StringVar(&installer.LogLevel, "log-level", "", "log level setting")
	flag.StringVar(&installer.Log, "log", "", "log setting")
	flag.StringVar(&installer.GinLog, "gin-log", "", "gin log setting")

	// pprof flags
	flag.StringVar(&installer.EnablePProf, "enable-pprof", "", "enable pprof")
	flag.StringVar(&installer.PProfListen, "pprof-listen", "", "pprof listen")

	// sinker flags
	flag.StringVar(&installer.Sinker, "sinker", "", "sinker configures")

	// cgroup flags
	flag.IntVar(&installer.CgroupDisabled, "cgroup-disabled", 0, "enable disable cgroup(Linux) limits for CPU and memory")
	flag.Float64Var(&installer.LimitCPUMax, "limit-cpumax", 30.0, "Cgroup CPU max usage")
	flag.Float64Var(&installer.LimitCPUMin, "limit-cpumin", 5.0, "Cgroup CPU min usage")
	flag.Int64Var(&installer.LimitMemMax, "limit-mem", 4096, "Cgroup memory limit")
}

func downloadFiles(to string) error {
	dl.CurDownloading = "datakit"

	cliopt := &ihttp.Options{
		InsecureSkipVerify: true, // ignore SSL error
	}

	if installer.Proxy != "" {
		u, err := url.Parse(installer.Proxy)
		if err != nil {
			return err
		}
		cliopt.ProxyURL = u
		l.Infof("set proxy to %s ok", installer.Proxy)
	}

	cli := ihttp.Cli(cliopt)

	if err := dl.Download(cli, datakitURL, to, true, flagDownloadOnly); err != nil {
		return err
	}

	fmt.Printf("\n")

	dl.CurDownloading = "data"
	if err := dl.Download(cli, dataURL, to, true, flagDownloadOnly); err != nil {
		return err
	}

	// We will not upgrade dk-upgrader default when upgrading Datakit except for setting flagUpgradeManagerService flag
	if !flagDKUpgrade || (flagDKUpgrade && flagUpgradeManagerService == 1) || flagDownloadOnly {
		if !flagDownloadOnly {
			to = upgrader.InstallDir
		}
		dl.CurDownloading = upgrader.BuildBinName
		if err := dl.Download(cli, dkUpgraderURL, to, true, flagDownloadOnly); err != nil {
			return fmt.Errorf("unable to download %s from [%s]: %w", upgrader.BuildBinName, dkUpgraderURL, err)
		}
	}

	if installer.IPDBType != "" {
		fmt.Printf("\n")
		baseURL := "https://" + DataKitBaseURL

		l.Debugf("get ipdb from %s", baseURL)
		if _, err := cmds.InstallIPDB(baseURL, installer.IPDBType); err != nil {
			l.Warnf("install IPDB %s failed error: %s, please try later.", installer.IPDBType, err.Error())
			time.Sleep(1 * time.Second)
		} else {
			config.Cfg.Pipeline.IPdbType = installer.IPDBType
		}
	}

	fmt.Printf("\n")
	return nil
}

func applyFlags() {
	var err error

	// setup logging
	if flagInstallLog == "stdout" {
		cp.Infof("Set log file to stdout\n")

		if err = logger.InitRoot(
			&logger.Option{
				Level: logger.DEBUG,
				Flags: logger.OPT_DEFAULT | logger.OPT_STDOUT,
			}); err != nil {
			cp.Errorf("Set root log faile: %s\n", err.Error())
		}
	} else {
		cp.Infof("Set log file to %s\n", flagInstallLog)
		if err = logger.InitRoot(&logger.Option{
			Path:  flagInstallLog,
			Level: logger.DEBUG,
			Flags: logger.OPT_DEFAULT,
		}); err != nil {
			cp.Errorf("Set root log faile: %s", err.Error())
		}
	}

	config.SetLog()
	installer.SetLog()
	l = logger.SLogger("installer")

	installer.DataKitVersion = DataKitVersion

	if flagDownloadOnly {
		if err = downloadFiles(""); err != nil { // download 过程直接覆盖已有安装
			l.Fatalf("download failed: %s", err.Error())
		}
		os.Exit(0)
	}

	if flagSrc != "" && flagOffline {
		for _, f := range strings.Split(flagSrc, ",") {
			fd, err := os.Open(filepath.Clean(f))
			if err != nil {
				l.Fatalf("Open: %s", err)
			}

			if err := dl.Extract(fd, datakit.InstallDir); err != nil {
				l.Fatalf("Extract: %s", err)
			} else if err := fd.Close(); err != nil {
				l.Warnf("Close: %s, ignored", err)
			}
		}

		// NOTE: continue to install/upgrade
	}

	if installer.Proxy != "" {
		if !strings.HasPrefix(installer.Proxy, "http") {
			installer.Proxy = "http://" + installer.Proxy
		}

		if _, err = url.Parse(installer.Proxy); err != nil {
			l.Warnf("bad proxy config expect http://ip:port given %s", installer.Proxy)
		} else {
			l.Infof("set proxy to %s", installer.Proxy)
		}
	}

	if InstallerBaseURL != "" {
		_, err := url.Parse(InstallerBaseURL)
		if err != nil {
			l.Errorf("ENV:$DK_INSTALLER_BASE_URL can not parse to URL, err=%v", err)
			os.Exit(0)
		}
		if !strings.HasSuffix(InstallerBaseURL, "/") {
			InstallerBaseURL += "/"
		}

		cp.Infof("Set installer base URL to %s\n", InstallerBaseURL)
		dataURL = InstallerBaseURL + "data.tar.gz"

		datakitURL = InstallerBaseURL + fmt.Sprintf("datakit-%s-%s-%s.tar.gz",
			runtime.GOOS,
			runtime.GOARCH,
			DataKitVersion)
	}
}

func main() {
	flag.Parse()

	if flagInfo {
		fmt.Printf(`
Version        : %s
Build At       : %s
Golang Version : %s
BaseUrl        : %s
Data           : %s
`, DataKitVersion, git.BuildAt, git.Golang, datakitURL, dataURL)
		os.Exit(0)
	}

	var err error

	dkservice.Executable = filepath.Join(datakit.InstallDir, datakitBin)
	if runtime.GOOS == datakit.OSWindows {
		dkservice.Executable += ".exe"
	}

	// fix user name.
	var userName, groupAdd, userAdd string
	if runtime.GOOS == datakit.OSLinux && len(flagUserName) > 0 && flagUserName != "root" {
		// check add group and user command.
		groupAdd, userAdd, err = checkUserGroupCmdOK()
		if err != nil {
			cp.Errorf("check command failed: %v\n", err)
			return
		}
		userName = builtInUserName // set as 'datakit'(default).

		cp.Infof("datakit service run as user: '%s'\n", userName)
	}

	svc, err := dkservice.NewService(userName)
	if err != nil {
		l.Errorf("new %s service failed: %s", runtime.GOOS, err.Error())
		return
	}

	svcStatus, err := svc.Status()
	if err != nil {
		if errors.Is(err, service.ErrNotInstalled) {
			cp.Infof("datakit service not installed before\n")
		} else {
			l.Warnf("svc.Status: %s, ignored", err.Error())
		}
	} else {
		switch svcStatus {
		case service.StatusUnknown: // not installed
			cp.Infof("DataKit service maybe not installed\n")
		case service.StatusStopped: // pass
			cp.Infof("DataKit service stopped\n")
		case service.StatusRunning:
			cp.Infof("Stopping running DataKit...\n")
			if err = service.Control(svc, "stop"); err != nil {
				l.Warnf("stop service failed %s, ignored", err.Error())
			}
		}
	}

	if !flagDKUpgrade || flagUpgradeManagerService == 1 {
		upgrader2.StopUpgradeService(userName)
	}

	applyFlags()

	// 迁移老版本 datakit 数据目录
	mvOldDatakit(svc)

	if !flagOffline {
		dlRetry := 5

		cp.Infof("Download installer...")

		for i := 0; i < dlRetry; i++ {
			if err = downloadFiles(datakit.InstallDir); err != nil { // download 过程直接覆盖已有安装
				cp.Warnf("[%d] download failed: %s, retry...", i, err.Error())
				continue
			}

			goto __downloadOK
		}

		cp.Errorf("Download failed, please check your network settings.\n")
		return
	}

__downloadOK:
	datakit.InitDirs()

	upgrader2.InstallUpgradeService(userName, flagDKUpgrade, flagInstallOnly, flagUpgradeManagerService, flagUpgradeServIPWhiteList)

	if flagDKUpgrade { // upgrade new version
		cp.Infof("Upgrading to version %s...\n", DataKitVersion)
		if err = installer.Upgrade(); err != nil {
			cp.Warnf("upgrade datakit failed: %s, ignored\n", err.Error())
		}
	} else { // install new datakit
		cp.Infof("Installing version %s...\n", DataKitVersion)
		installer.Install(svc)
	}

	setupUserGroup(userName, userName, groupAdd, userAdd)

	if flagInstallOnly != 0 {
		cp.Warnf("Only install service %s, NOT started\n", dkservice.Name)
	} else {
		cp.Infof("Starting service %s...\n", dkservice.Name)
		if err = service.Control(svc, "start"); err != nil {
			cp.Warnf("Start service failed: %s\n", err.Error())
		}
	}

	cp.Infof("Create symlinks...\n")
	if err := config.CreateSymlinks(); err != nil {
		l.Errorf("CreateSymlinks: %s", err.Error())
	}

	if err := checkIsNewVersion("http://"+config.Cfg.HTTPAPI.Listen, DataKitVersion); err != nil {
		promptFixVersionChecking()
		return
	}

	cp.Infof("Current running datakit version: %s\n", DataKitVersion)

	if flagDKUpgrade {
		cp.Infof("Upgrade OK.\n")
	} else {
		cp.Infof("Install OK.\n")
	}
	promptReferences()
}

// test if installed/upgraded to expected version.
func checkIsNewVersion(host, version string) error {
	x := struct {
		Content map[string]string `json:"content"`
	}{}

	for i := 0; i < 10; i++ {
		time.Sleep(time.Second * time.Duration(i+1))

		resp, err := http.Get(host + "/v1/ping")
		if err != nil {
			l.Errorf("http.Get: %s", err)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			l.Errorf("ioutil.ReadAll: %s", err)
		}

		resp.Body.Close() //nolint:errcheck,gosec

		if err := json.Unmarshal(body, &x); err != nil {
			l.Errorf("json.Unmarshal: %s", err)
		}

		if x.Content["version"] != version {
			l.Warnf("current version: %s, expect %s", x.Content["version"], version)
		} else {
			return nil
		}
	}

	return fmt.Errorf("check version failed")
}

func promptFixVersionChecking() {
	cp.Warnf("\n\tVisit https://docs.guance.com/datakit/datakit-update/#version-check-failed to fix the issue.\n")
}

func promptReferences() {
	cp.Infof("\nVisit https://docs.guance.com/datakit/changelog/ to see DataKit change logs.\n")
	if config.Cfg.HTTPAPI.Listen != "localhost:9529" {
		cp.Infof("Use `datakit monitor --to %s` to see DataKit running status.\n", config.Cfg.HTTPAPI.Listen)
	} else {
		cp.Infof("Use `datakit monitor` to see DataKit running status.\n")
	}
}

func mvOldDatakit(svc service.Service) {
	olddir := oldInstallDir
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case datakit.OSArchWinAmd64:
		olddir = oldInstallDirWin
	case datakit.OSArchWin386:
		olddir = oldInstallDirWin32
	}

	if _, err := os.Stat(olddir); err != nil {
		l.Infof("deprecated install path %s not found\n", olddir)
		return
	}

	if err := service.Control(svc, "uninstall"); err != nil {
		l.Warnf("uninstall service failed: %s", err.Error())
	}

	if err := os.Rename(olddir, datakit.InstallDir); err != nil {
		l.Fatalf("move %s -> %s failed: %s", olddir, datakit.InstallDir, err.Error())
	}
}

const (
	installDir      = "/usr/local/datakit"
	defaultLogDir   = "/var/log/datakit"
	builtInUserName = "datakit"
)

func setupUserGroup(userName, groupName, groupAdd, userAdd string) {
	l.Info("setupUserGroup entry")

	if len(userName) == 0 || userName == "root" || runtime.GOOS != datakit.OSLinux {
		return
	}

	if len(groupAdd) == 0 || len(userAdd) == 0 {
		l.Errorf("groupAdd or userAdd command not set.")
		return
	}

	l.Info("setupUserGroup start")

	// set as 'datakit'.
	userName = builtInUserName
	_ = groupName // for lint.
	groupName = builtInUserName

	// add group.
	if _, err := user.LookupGroup(groupName); err != nil {
		if err.Error() == user.UnknownGroupError(groupName).Error() {
			if err = executeCmd(groupAdd, "--system", groupName); err != nil {
				l.Errorf("%s failed: %v", groupAdd, err)
				return
			}
		} else {
			l.Errorf("LookupGroup failed: %v", err)
		}
	}

	// add user.
	if _, err := user.Lookup(userName); err != nil {
		if err.Error() == user.UnknownUserError(userName).Error() {
			if err = executeCmd(userAdd, "--system", "--no-create-home", "--disabled-password", "--ingroup", groupName, userName); err != nil {
				l.Errorf("%s failed: %v", userAdd, err)
				return
			}
		} else {
			l.Errorf("Lookup failed: %v", err)
		}
	}

	// make dirs.
	if err := os.MkdirAll(installDir, os.ModePerm); err != nil {
		l.Errorf("make installDir failed: %v", err)
	}
	if err := os.MkdirAll(defaultLogDir, os.ModePerm); err != nil {
		l.Errorf("make defaultLogDir failed: %v", err)
	}

	// chown.
	if err := executeCmd("chown", "-R", fmt.Sprintf("%s:%s", userName, groupName), installDir, defaultLogDir); err != nil {
		l.Errorf("chown failed: %v", err)
		return
	}
	// chmod.
	if err := executeCmd("chmod", "-R", "755", installDir, defaultLogDir); err != nil {
		l.Errorf("chmod failed: %v", err)
		return
	}

	// chown.
	if err := executeCmd("chown", "-R", fmt.Sprintf("%s:%s", userName, groupName), upgrader.InstallDir, upgrader.DefaultLogDir); err != nil {
		l.Errorf("chown failed: %v", err)
		return
	}
	// chmod.
	if err := executeCmd("chmod", "-R", "755", upgrader.InstallDir, upgrader.DefaultLogDir); err != nil {
		l.Errorf("chmod failed: %v", err)
		return
	}
}

func executeCmd(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	return cmd.Run()
}

func checkUserGroupCmdOK() (groupAdd string, userAdd string, err error) {
	// check group
	groupAdd, err = checkCmd("addgroup", "groupadd")
	if err != nil {
		return "", "", fmt.Errorf("neither 'addgroup' or 'groupadd' executable file found in $PATH")
	}

	userAdd, err = checkCmd("adduser", "useradd")
	if err != nil {
		return "", "", fmt.Errorf("neither 'adduser' or 'useradd' executable file found in $PATH")
	}

	return groupAdd, userAdd, nil
}

func checkCmd(candidates ...string) (cmdString string, err error) {
	for _, v := range candidates {
		if err = commandExists(v); err != nil {
			cp.Infof("try '%s' failed: %v\n", v, err)
		} else {
			l.Infof("try '%s' ok.", v)
			return v, nil
		}
	}
	return "", err
}

func commandExists(cmd string) error {
	_, err := exec.LookPath(cmd)
	return err
}
