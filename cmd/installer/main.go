// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/service"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	dl "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/downloader"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/sinkfuncs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/version"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
)

var (
	oldInstallDir      = "/usr/local/cloudcare/dataflux/datakit"
	oldInstallDirWin   = `C:\Program Files\dataflux\datakit`
	oldInstallDirWin32 = `C:\Program Files (x86)\dataflux\datakit`

	DataKitBaseURL = ""
	DataKitVersion = ""
	dataURL        = "https://" + path.Join(DataKitBaseURL, "data.tar.gz")
	datakitURL     = "https://" + path.Join(DataKitBaseURL,
		fmt.Sprintf("datakit-%s-%s-%s.tar.gz",
			runtime.GOOS,
			runtime.GOARCH,
			DataKitVersion))
	l = logger.DefaultSLogger("installer")
)

// Installer flags.
var (
	flagHostName string
	flagDKUpgrade,
	flagOffline,
	flagDownloadOnly,
	flagInfo,
	flagOTA bool

	flagDataway,
	flagDCAEnable,
	flagEnableInputs,
	flagDatakitName,
	flagGlobalTags,
	flagProxy,
	flagDatakitHTTPListen,
	flagNamespace,
	flagInstallLog,
	flagInstallExternals,
	flagDCAListen,
	flagDCAWhiteList,
	flagGitURL,
	flagGitKeyPath,
	flagGitKeyPW,
	flagGitBranch,
	flagGitPullInterval,
	flagSrc,
	flagCloudProvider,
	flagRumOriginIPHeader,
	flagLogLevel,
	flagLog,
	flagIpdb,
	flagGinLog,
	flagEnableElection,
	flagEnablePProf,
	flagPProfListen,
	flagDisable404Page,
	flagSinkMetric,
	flagSinkNetwork,
	flagSinkKeyEvent,
	flagSinkObject,
	flagSinkCustomObject,
	flagSinkLogging,
	flagSinkTracing,
	flagSinkRUM,
	flagSinkSecurity string

	flagInstallOnly,
	flagCgroupDisabled,
	envEnableExperimental,
	flagDatakitHTTPPort int

	flagLimitCPUMax float64
	flagLimitMemMax int64
	flagLimitCPUMin float64
)

const (
	datakitBin = "datakit"
)

func init() { //nolint:gochecknoinits
	flag.BoolVar(&flagDKUpgrade, "upgrade", false, "")
	flag.BoolVar(&flagOTA, "ota", false, "auto update")
	flag.StringVar(&flagDCAEnable, "dca-enable", "", "enable DCA")
	flag.StringVar(&flagDCAListen, "dca-listen", "0.0.0.0:9531", "DCA listen address and port")
	flag.StringVar(&flagDCAWhiteList, "dca-white-list", "", "DCA white list")
	flag.StringVar(&flagDataway, "dataway", "", "DataWay host(https://guance.openway.com?token=xxx)")
	flag.StringVar(&flagEnableInputs,
		"enable-inputs", "", "default enable inputs(comma splited, example:cpu,mem,disk)")
	flag.StringVar(&flagDatakitName, "name", "", "specify DataKit name, example: prod-env-datakit")
	flag.StringVar(&flagGlobalTags, "global-tags", "",
		"enable global tags, example: host= __datakit_hostname,ip= __datakit_ip")
	flag.StringVar(&flagProxy, "proxy", "", "http proxy http://ip:port for datakit")
	flag.StringVar(&flagDatakitHTTPListen, "listen", "localhost", "datakit HTTP listen")
	flag.StringVar(&flagNamespace, "namespace", "", "datakit namespace")
	flag.StringVar(&flagInstallLog, "install-log", "", "install log")
	flag.StringVar(&flagHostName, "env_hostname", "", "host name")
	flag.StringVar(&flagIpdb, "ipdb-type", "", "ipdb type")
	flag.StringVar(&flagInstallExternals, "install-externals", "", "install some external inputs")
	flag.StringVar(&flagCloudProvider,
		"cloud-provider", "", "specify cloud provider(accept aliyun/tencent/aws)")
	flag.StringVar(&flagGitURL, "git-url", "", "git repo url")
	flag.StringVar(&flagGitKeyPath, "git-key-path", "", "git repo access private key path")
	flag.StringVar(&flagGitKeyPW, "git-key-pw", "", "git repo access private use password")
	flag.StringVar(&flagGitBranch, "git-branch", "", "git repo branch name")
	flag.StringVar(&flagGitPullInterval, "git-pull-interval", "", "git repo pull interval")
	flag.StringVar(&flagEnableElection, "enable-election", "", "datakit election")
	flag.StringVar(&flagRumOriginIPHeader, "rum-origin-ip-header", "", "rum only")
	flag.StringVar(&flagDisable404Page, "disable-404page", "", "datakit rum 404 page")
	flag.StringVar(&flagLogLevel, "log-level", "", "log level setting")
	flag.StringVar(&flagLog, "log", "", "log setting")
	flag.StringVar(&flagGinLog, "gin-log", "", "gin log setting")
	flag.StringVar(&flagEnablePProf, "enable-pprof", "", "enable pprof")
	flag.StringVar(&flagPProfListen, "pprof-listen", "", "pprof listen")
	flag.StringVar(&flagSrc, "srcs",
		fmt.Sprintf("./datakit-%s-%s-%s.tar.gz,./data.tar.gz",
			runtime.GOOS, runtime.GOARCH, DataKitVersion),
		`local path of install files`)

	// sink parameters
	flag.StringVar(&flagSinkMetric, "sink-metric", "", "sink for Metric")
	flag.StringVar(&flagSinkNetwork, "sink-network", "", "sink for Network")
	flag.StringVar(&flagSinkKeyEvent, "sink-keyevent", "", "sink for Key Event")
	flag.StringVar(&flagSinkObject, "sink-object", "", "sink for Object")
	flag.StringVar(&flagSinkCustomObject, "sink-custom-object", "", "sink for CustomObject")
	flag.StringVar(&flagSinkLogging, "sink-logging", "", "sink for Logging")
	flag.StringVar(&flagSinkTracing, "sink-tracing", "", "sink for Tracing")
	flag.StringVar(&flagSinkRUM, "sink-rum", "", "sink for RUM")
	flag.StringVar(&flagSinkSecurity, "sink-security", "", "sink for Security")

	flag.Float64Var(&flagLimitCPUMax, "limit-cpumax", 30.0, "Cgroup CPU max usage")
	flag.Float64Var(&flagLimitCPUMin, "limit-cpumin", 5.0, "Cgroup CPU min usage")
	flag.Int64Var(&flagLimitMemMax, "limit-mem", 4096, "Cgroup memory limit")

	flag.IntVar(&flagCgroupDisabled, "cgroup-disabled", 0, "enable disable cgroup(Linux) limits for CPU and memory")
	flag.IntVar(&flagDatakitHTTPPort, "port", 9529, "datakit HTTP port")
	flag.IntVar(&flagInstallOnly, "install-only", 0, "install only, not start")

	flag.BoolVar(&flagInfo, "info", false, "show installer info")
	flag.BoolVar(&flagOffline, "offline", false, "-offline option removed")
	flag.BoolVar(&flagDownloadOnly, "download-only", false, "only download install packages")
	flag.IntVar(&envEnableExperimental, "env_enable_experimental", 0, "")
}

func downloadFiles(to string) error {
	dl.CurDownloading = "datakit"

	cliopt := &ihttp.Options{
		InsecureSkipVerify: true, // ignore SSL error
	}

	if flagProxy != "" {
		u, err := url.Parse(flagProxy)
		if err != nil {
			return err
		}
		cliopt.ProxyURL = u
		l.Infof("set proxy to %s ok", flagProxy)
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

	if flagIpdb != "" {
		fmt.Printf("\n")
		baseURL := "https://" + DataKitBaseURL

		l.Debugf("get ipdb from %s", baseURL)
		if _, err := cmds.InstallIpdb(baseURL, flagIpdb); err != nil {
			l.Warnf("ipdb install failed error: %s, please try later.", err.Error())
			time.Sleep(1 * time.Second)
		} else {
			config.Cfg.Pipeline.IPdbType = flagIpdb
		}
	}

	fmt.Printf("\n")
	return nil
}

func applyFlags() {
	var err error

	// set logging
	if flagInstallLog == "" {
		if err = logger.InitRoot(
			&logger.Option{
				Level: logger.DEBUG,
				Flags: logger.OPT_DEFAULT | logger.OPT_STDOUT,
			}); err != nil {
			l.Errorf("set root log faile: %s", err.Error())
		}
	} else {
		l.Infof("set log file to %s", flagInstallLog)

		if err = logger.InitRoot(&logger.Option{
			Path:  flagInstallLog,
			Level: logger.DEBUG,
			Flags: logger.OPT_DEFAULT,
		}); err != nil {
			l.Errorf("set root log faile: %s", err.Error())
		}
	}

	l = logger.SLogger("installer")

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

	if flagProxy != "" {
		if !strings.HasPrefix(flagProxy, "http") {
			flagProxy = "http://" + flagProxy
		}

		if _, err = url.Parse(flagProxy); err != nil {
			l.Warnf("bad proxy config expect http://ip:port given %s", flagProxy)
		} else {
			l.Infof("set proxy to %s", flagProxy)
		}
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

	dkservice.ServiceExecutable = filepath.Join(datakit.InstallDir, datakitBin)
	if runtime.GOOS == datakit.OSWindows {
		dkservice.ServiceExecutable += ".exe"
	}

	svc, err := dkservice.NewService()
	if err != nil {
		l.Errorf("new %s service failed: %s", runtime.GOOS, err.Error())
		return
	}

	svcStatus, err := svc.Status()
	if err != nil {
		if errors.Is(err, service.ErrNotInstalled) {
			l.Infof("datakit service not installed before")
		} else {
			l.Warnf("svc.Status: %s, ignored", err.Error())
		}
	} else {
		switch svcStatus {
		case service.StatusUnknown: // not installed
			l.Info("datakit service maybe not installed")
		case service.StatusStopped: // pass
			l.Info("datakit service stopped")
		case service.StatusRunning:
			l.Info("stoping datakit...")
			if err = service.Control(svc, "stop"); err != nil {
				l.Warnf("stop service failed %s, ignored", err.Error()) //nolint:lll
			}
		}
	}

	applyFlags()

	// 迁移老版本 datakit 数据目录
	mvOldDatakit(svc)

	if !flagOffline {
		for i := 0; i < 5; i++ {
			if err = downloadFiles(datakit.InstallDir); err != nil { // download 过程直接覆盖已有安装
				l.Errorf("[%d] download failed: %s, retry...", i, err.Error())
				continue
			}
			l.Infof("[%d] download installer ok", i)
			break
		}
	}

	datakit.InitDirs()

	if flagDKUpgrade { // upgrade new version
		if err := checkUpgradeVersion(git.Version); err != nil {
			l.Fatalf("upgrade datakit: %s", err.Error())
		}

		l.Infof("Upgrading to version %s...", DataKitVersion)
		if err = upgradeDatakit(svc); err != nil {
			l.Warnf("upgrade datakit failed: %s", err.Error())
		}
	} else { // install new datakit
		l.Infof("Installing version %s...", DataKitVersion)
		installNewDatakit(svc)
	}

	if flagInstallOnly != 0 {
		l.Infof("only install service %s, NOT started", dkservice.ServiceName)
	} else {
		l.Infof("starting service %s...", dkservice.ServiceName)
		if err = service.Control(svc, "start"); err != nil {
			l.Warnf("start service failed: %s", err.Error()) //nolint:lll
		}
	}

	if err := config.CreateSymlinks(); err != nil {
		l.Errorf("CreateSymlinks: %s", err.Error())
	}

	if err := checkIsNewVersion("http://"+config.Cfg.HTTPAPI.Listen, DataKitVersion); err != nil {
		l.Errorf("checkIsNewVersion: %s", err.Error()) //nolint:lll
	} else {
		l.Infof("current running datakit is version: %s", DataKitVersion) //nolint:lll

		if flagDKUpgrade {
			l.Info(":) Upgrade Success!")
		} else {
			l.Info(":) Install Success!")
		}
		promptReferences()
	}
}

func checkIsNewVersion(host, version string) error {
	x := struct {
		Content map[string]string `json:"content"`
	}{}

	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)

		resp, err := http.Get(host + "/v1/ping")
		if err != nil {
			l.Errorf("http.Get: %s", err)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			l.Errorf("ioutil.ReadAll: %s", err)
			continue
		}

		defer resp.Body.Close()

		if err := json.Unmarshal(body, &x); err != nil {
			l.Errorf("json.Unmarshal: %s", err)
			return err
		}

		if x.Content["version"] != version {
			return fmt.Errorf("current version: %s, expect %s", x.Content["version"], version)
		} else {
			return nil
		}
	}

	return fmt.Errorf("check current version failed")
}

func promptReferences() {
	fmt.Println("\n\tVisit https://www.yuque.com/dataflux/datakit/changelog to see DataKit change logs.")
	if config.Cfg.HTTPAPI.Listen != "localhost:9529" {
		fmt.Printf("\tUse `datakit monitor --to %s` to see DataKit running status.\n", config.Cfg.HTTPAPI.Listen)
	} else {
		fmt.Println("\tUse `datakit monitor` to see DataKit running status.")
	}
}

func upgradeDatakit(svc service.Service) error {
	mc := config.Cfg

	if err := mc.LoadMainTOML(datakit.MainConfPath); err == nil {
		mc = upgradeMainConfig(mc)

		if flagOTA {
			l.Debugf("set auto update flag")
			mc.AutoUpdate = flagOTA
		}

		writeDefInputToMainCfg(mc)
	} else {
		l.Warnf("load main config: %s, ignored", err.Error())
		return err
	}

	// build datakit main config
	if err := mc.InitCfg(datakit.MainConfPath); err != nil {
		l.Fatalf("failed to init datakit main config: %s", err.Error())
	}

	for _, dir := range []string{datakit.DataDir, datakit.ConfdDir} {
		if err := os.MkdirAll(dir, datakit.ConfPerm); err != nil {
			return err
		}
	}

	installExternals := map[string]struct{}{}
	for _, v := range strings.Split(flagInstallExternals, ",") {
		installExternals[v] = struct{}{}
	}
	updateEBPF := false
	if runtime.GOOS == datakit.OSLinux && runtime.GOARCH == "amd64" {
		if _, err := os.Stat(filepath.Join(datakit.InstallDir, "externals", "datakit-ebpf")); err == nil {
			updateEBPF = true
		}
		if _, ok := installExternals["datakit-ebpf"]; ok {
			updateEBPF = true
		}
	}
	if updateEBPF {
		l.Info("upgrade datakit-ebpf...")
		// nolint:gosec
		cmd := exec.Command(filepath.Join(datakit.InstallDir, "datakit"), "install", "--datakit-ebpf")
		if msg, err := cmd.CombinedOutput(); err != nil {
			l.Warnf("upgrade datakit-ebpf failed: %s(%s), ignored", err.Error(), msg)
		}
	}

	if err := service.Control(svc, "install"); err != nil {
		return err
	}

	return nil
}

func installNewDatakit(svc service.Service) {
	svcStatus, err := svc.Status()
	if err != nil {
		if errors.Is(err, service.ErrNotInstalled) {
			l.Infof("datakit service not installed before")
			// pass
		} else {
			l.Warnf("svc.Status: %s", err.Error())
		}
	} else {
		switch svcStatus {
		case service.StatusUnknown: // pass
		case service.StatusStopped:
			if err := service.Control(svc, "uninstall"); err != nil {
				l.Warnf("uninstall service failed: %s", err.Error())
			}
		case service.StatusRunning: // should not been here
			l.Warnf("unexpected: datakit service should have stopped")
		}
	}

	mc := config.Cfg

	// prepare dataway info and check token format
	if len(flagDataway) != 0 {
		var err error
		mc.DataWay, err = getDataWay()
		if err != nil {
			l.Fatal(err)
		}
	}

	if flagOTA {
		l.Debugf("set auto update flag")
		mc.AutoUpdate = flagOTA
	}
	if flagDCAListen != "" {
		config.Cfg.DCAConfig.Listen = flagDCAListen
	}

	if flagDCAWhiteList != "" {
		config.Cfg.DCAConfig.WhiteList = strings.Split(flagDCAWhiteList, ",")
	}

	if flagDCAEnable != "" {
		config.Cfg.DCAConfig.Enable = true

		// check white list whether is empty or invalid
		if len(config.Cfg.DCAConfig.WhiteList) == 0 {
			l.Fatalf("DCA service is enabled, but white list is empty! ")
		}
		for _, cidr := range config.Cfg.DCAConfig.WhiteList {
			_, _, err := net.ParseCIDR(cidr)
			if err != nil {
				if net.ParseIP(cidr) == nil {
					l.Fatalf("DCA white list set error: invalid ip, %s", cidr)
				}
			}
		}
	}

	if flagEnablePProf != "" {
		config.Cfg.EnablePProf = true
	}

	if flagPProfListen != "" {
		config.Cfg.PProfListen = flagPProfListen
	}

	// Only linux support cgroup.
	if flagCgroupDisabled != 1 && runtime.GOOS == datakit.OSLinux {
		l.Infof("Croups enabled under Linux")
		mc.Cgroup.Enable = true

		if flagLimitCPUMin > 0 {
			mc.Cgroup.CPUMin = flagLimitCPUMin
		}

		if flagLimitCPUMax > 0 {
			mc.Cgroup.CPUMax = flagLimitCPUMax
		}

		if mc.Cgroup.CPUMax < mc.Cgroup.CPUMin {
			l.Fatalf("invalid CGroup CPU limit, max should larger than min")
		}

		if flagLimitMemMax > 0 {
			l.Infof("cgroup set max memory to %dMB", flagLimitMemMax)
			mc.Cgroup.MemMax = flagLimitMemMax
		} else {
			l.Infof("cgroup max memory not set")
		}
	}

	if flagLimitCPUMax != 0 {
		if flagLimitCPUMax < 0 || flagLimitCPUMax > 100 {
			l.Errorf("Limit CPU max can not less than zero or bigger than one hundred")
			flagLimitCPUMax = 20.0
		}
		mc.Cgroup.CPUMax = flagLimitCPUMax
	}

	if flagHostName != "" {
		l.Debugf("set ENV_HOSTNAME to %s", flagHostName)
		mc.Environments["ENV_HOSTNAME"] = flagHostName
	}

	// accept any install options
	if flagGlobalTags != "" {
		l.Infof("set global tags...")
		mc.GlobalTags = config.ParseGlobalTags(flagGlobalTags)

		l.Infof("set global tags %+#v", mc.GlobalTags)
	}

	mc.Namespace = flagNamespace
	mc.HTTPAPI.Listen = fmt.Sprintf("%s:%d", flagDatakitHTTPListen, flagDatakitHTTPPort)
	mc.InstallDate = time.Now()
	mc.InstallVer = DataKitVersion

	if flagDatakitName != "" {
		mc.Name = flagDatakitName
	}

	if flagGitURL != "" {
		mc.GitRepos = &config.GitRepost{
			PullInterval: flagGitPullInterval,
			Repos: []*config.GitRepository{
				{
					Enable:                true,
					URL:                   flagGitURL,
					SSHPrivateKeyPath:     flagGitKeyPath,
					SSHPrivateKeyPassword: flagGitKeyPW,
					Branch:                flagGitBranch,
				}, // GitRepository
			}, // Repos
		} // GitRepost
	}

	if flagEnableElection != "" {
		l.Infof("set enable election: %v", flagEnableElection)
		mc.EnableElection = true
	}
	if flagDisable404Page != "" {
		l.Infof("set disable 404 page: %v", flagDisable404Page)
		mc.Disable404PageDeprecated = true
	}
	if flagRumOriginIPHeader != "" {
		l.Infof("set rum origin IP header: %s", flagRumOriginIPHeader)
		mc.HTTPAPI.RUMOriginIPHeader = flagRumOriginIPHeader
	}
	if flagLogLevel != "" {
		l.Infof("set log level: %s", flagLogLevel)
		mc.Logging.Level = flagLogLevel
	}
	if flagLog != "" {
		l.Infof("set log: %s", flagLog)
		mc.Logging.Log = flagLog
	}
	if flagGinLog != "" {
		l.Infof("set gin log: %s", flagGinLog)
		mc.Logging.GinLog = flagGinLog
	}

	// parse sink
	if err := parseSinkArgs(mc); err != nil {
		mc.Sinks.Sink = []map[string]interface{}{{}} // clear
		l.Fatalf("parseSinkArgs failed: %s", err.Error())
	}

	writeDefInputToMainCfg(mc)

	// build datakit main config
	if err := mc.InitCfg(datakit.MainConfPath); err != nil {
		l.Fatalf("failed to init datakit main config: %s", err.Error())
	}

	installExternals := map[string]struct{}{}
	for _, v := range strings.Split(flagInstallExternals, ",") {
		installExternals[v] = struct{}{}
	}
	updateEBPF := false
	if runtime.GOOS == datakit.OSLinux && runtime.GOARCH == "amd64" {
		if _, err := os.Stat(filepath.Join(datakit.InstallDir, "externals", "datakit-ebpf")); err == nil {
			updateEBPF = true
		}
		if _, ok := installExternals["datakit-ebpf"]; ok {
			updateEBPF = true
		}
	}
	if updateEBPF {
		l.Info("install datakit-ebpf...")
		// nolint:gosec
		cmd := exec.Command(filepath.Join(datakit.InstallDir, "datakit"), "install", "--datakit-ebpf")
		if msg, err := cmd.CombinedOutput(); err != nil {
			l.Errorf("upgradde external input %s failed: %s msg: %s", "datakit-ebpf", err.Error(), msg)
		}
	}

	l.Infof("installing service %s...", dkservice.ServiceName)
	if err := service.Control(svc, "install"); err != nil {
		l.Warnf("uninstall service failed %s", err.Error()) //nolint:lll
	}
}

func parseSinkArgs(mc *config.Config) error {
	if mc == nil {
		return fmt.Errorf("invalid main config")
	}

	if mc.Sinks == nil {
		return fmt.Errorf("invalid main config sinks")
	}

	categoryShorts := []string{
		datakit.SinkCategoryMetric,
		datakit.SinkCategoryNetwork,
		datakit.SinkCategoryKeyEvent,
		datakit.SinkCategoryObject,
		datakit.SinkCategoryCustomObject,
		datakit.SinkCategoryLogging,
		datakit.SinkCategoryTracing,
		datakit.SinkCategoryRUM,
		datakit.SinkCategorySecurity,
	}

	args := []string{
		flagSinkMetric,
		flagSinkNetwork,
		flagSinkKeyEvent,
		flagSinkObject,
		flagSinkCustomObject,
		flagSinkLogging,
		flagSinkTracing,
		flagSinkRUM,
		flagSinkSecurity,
	}

	sinks, err := sinkfuncs.GetSinkFromEnvs(categoryShorts, args)
	if err != nil {
		return err
	}

	mc.Sinks.Sink = sinks
	return nil
}

var (
	defaultHostInputs = []string{
		"cpu",
		"disk",
		"diskio",
		"mem",
		"swap",
		"system",
		"hostobject",
		"net",
		"host_processes",
	}
	defaultHostInputsForLinux = []string{
		"cpu",
		"disk",
		"diskio",
		"mem",
		"swap",
		"system",
		"hostobject",
		"net",
		"host_processes",
		"container",
	}
)

func writeDefInputToMainCfg(mc *config.Config) {
	hostInputs := defaultHostInputs
	if runtime.GOOS == datakit.OSLinux {
		hostInputs = defaultHostInputsForLinux
	}

	if flagEnableInputs == "" {
		flagEnableInputs = strings.Join(hostInputs, ",")
	} else {
		flagEnableInputs = flagEnableInputs + "," + strings.Join(hostInputs, ",")
	}

	mc.EnableDefaultsInputs(flagEnableInputs)

	if flagCloudProvider != "" {
		if err := injectCloudProvider(flagCloudProvider); err != nil {
			l.Fatalf("failed to inject cloud-provider: %s", err.Error())
		} else {
			l.Infof("set cloud provider to %s ok", flagCloudProvider)
		}
	} else {
		l.Infof("cloud provider not set")
	}

	l.Debugf("main config:\n%s", mc.String())
}

func injectCloudProvider(p string) error {
	switch p {
	case "aliyun", "tencent", "aws", "hwcloud", "azure":

		l.Infof("try set cloud provider to %s...", p)

		conf := preEnableHostobjectInput(p)

		if err := os.MkdirAll(filepath.Join(datakit.ConfdDir, "host"), datakit.ConfPerm); err != nil {
			l.Fatalf("failed to init hostobject conf: %s", err.Error())
		}

		cfgpath := filepath.Join(datakit.ConfdDir, "host", "hostobject.conf")
		if err := ioutil.WriteFile(cfgpath, conf, datakit.ConfPerm); err != nil {
			return err
		}

	case "": // pass

	default:
		l.Warnf("unknown cloud provider %s, ignored", p)
	}

	return nil
}

func preEnableHostobjectInput(cloud string) []byte {
	// I don't want to import hostobject input, cause the installer binary bigger
	sample := []byte(`
[inputs.hostobject]

#pipeline = '' # optional

## Datakit does not collect network virtual interfaces under the linux system.
## Setting enable_net_virtual_interfaces to true will collect network virtual interfaces stats for linux.
# enable_net_virtual_interfaces = true

## Ignore mount points by filesystem type. Default ignored following FS types
# ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "autofs", "squashfs", "aufs"]


[inputs.hostobject.tags] # (optional) custom tags
# cloud_provider = "aliyun" # aliyun/tencent/aws
# some_tag = "some_value"
# more_tag = "some_other_value"
# ...`)

	conf := bytes.ReplaceAll(sample,
		[]byte(`# cloud_provider = "aliyun"`),
		[]byte(fmt.Sprintf(`  cloud_provider = "%s"`, cloud)))

	return conf
}

func upgradeMainConfig(c *config.Config) *config.Config {
	if c.DataWayCfg != nil {
		c.DataWayCfg.DeprecatedURL = ""
	}

	// XXX: 无脑更改日志位置
	switch runtime.GOOS {
	case datakit.OSWindows:
		c.Logging.Log = filepath.Join(datakit.InstallDir, "log")
		c.Logging.GinLog = filepath.Join(datakit.InstallDir, "gin.log")
	default:
		c.Logging.Log = "/var/log/datakit/log"
		c.Logging.GinLog = "/var/log/datakit/gin.log"
	}
	l.Debugf("set log to %s, remove ", c.Logging.Log)
	l.Debugf("set gin log to %s", c.Logging.GinLog)

	if c.LogDeprecated != "" {
		c.Logging.Log = c.LogDeprecated
		c.LogDeprecated = ""
	}

	if c.LogLevelDeprecated != "" {
		c.Logging.Level = c.LogLevelDeprecated
		c.LogLevelDeprecated = ""
	}

	if c.LogRotateDeprecated != 0 {
		c.Logging.Rotate = c.LogRotateDeprecated
		c.LogRotateDeprecated = 0
	}

	if c.GinLogDeprecated != "" {
		c.Logging.GinLog = c.GinLogDeprecated
		c.GinLogDeprecated = ""
	}

	if c.HTTPListenDeprecated != "" {
		c.HTTPAPI.Listen = c.HTTPListenDeprecated
		c.HTTPListenDeprecated = ""
	}

	if c.Disable404PageDeprecated {
		c.HTTPAPI.Disable404Page = true
		c.Disable404PageDeprecated = false
	}

	if c.IOCacheCountDeprecated != 0 {
		c.IOConf.MaxCacheCount = c.IOCacheCountDeprecated
		c.IOCacheCountDeprecated = 0
	}

	if c.OutputFileDeprecated != "" {
		c.IOConf.OutputFile = c.OutputFileDeprecated
		c.OutputFileDeprecated = ""
	}

	if c.IntervalDeprecated != "" {
		c.IntervalDeprecated = ""
	}

	if c.DataWayCfg != nil {
		c.DataWayCfg.HTTPProxy = flagProxy
	}

	c.InstallVer = DataKitVersion
	c.UpgradeDate = time.Now()

	return c
}

func getDataWay() (dataway.DataWay, error) {
	var dwCfg *dataway.DataWayCfg
	if flagDataway != "" {
		dwCfg = &dataway.DataWayCfg{}
		dwCfg.URLs = strings.Split(flagDataway, ",")

		if flagProxy != "" {
			l.Debugf("set proxy to %s", flagProxy)
			dwCfg.HTTPProxy = flagProxy
		}

		dw := &dataway.DataWayDefault{}
		if err := dw.Init(dwCfg); err != nil {
			return nil, err
		} else {
			tokens := dw.GetTokens()
			if len(tokens) == 0 {
				return nil, fmt.Errorf("dataway token should not be empty")
			}

			if err := dw.CheckToken(tokens[0]); err != nil {
				return nil, err
			}
			config.Cfg.DataWayCfg = dwCfg
			return dw, nil
		}
	} else {
		return nil, fmt.Errorf("dataway is not set")
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
		l.Debugf("deprecated install path %s not exists, ignored", olddir)
		return
	}

	if err := service.Control(svc, "uninstall"); err != nil {
		l.Warnf("uninstall service failed: %s", err.Error())
	}

	if err := os.Rename(olddir, datakit.InstallDir); err != nil {
		l.Fatalf("move %s -> %s failed: %s", olddir, datakit.InstallDir, err.Error())
	}
}

func checkUpgradeVersion(s string) error {
	v := version.VerInfo{VersionString: s}
	if err := v.Parse(); err != nil {
		return err
	}

	// 对 1.1.x 版本的 datakit，此处暂且认为是 stable 版本，不然
	// 无法从 1.1.x 升级到 1.2.x
	// 1.2 以后的版本（1.3/1.5/...）等均视为 unstable 版本
	if v.GetMinor() == 1 {
		return nil
	}

	if !v.IsStable() {
		if envEnableExperimental == 1 {
			l.Info("upgrade version is unstable")
		} else {
			return fmt.Errorf("upgrade to %s is not stable version, use env: <$DK_ENABLE_EXPEIMENTAL=1> to upgrade", DataKitVersion)
		}
	}
	return nil
}
