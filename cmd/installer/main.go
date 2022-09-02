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
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/service"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/installer/installer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
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
	datakitURL     = "https://" + path.Join(DataKitBaseURL,
		fmt.Sprintf("datakit-%s-%s-%s.tar.gz",
			runtime.GOOS,
			runtime.GOARCH,
			DataKitVersion))
	l = logger.DefaultSLogger("installer")
)

// Installer flags.
var (
	flagDKUpgrade,
	flagOffline,
	flagDownloadOnly,
	flagInfo bool

	flagInstallLog,
	flagSrc string

	flagInstallOnly int
)

const (
	datakitBin = "datakit"
)

//nolint:gochecknoinits,lll
func init() {
	flag.BoolVar(&flagDKUpgrade, "upgrade", false, "")
	flag.StringVar(&flagInstallLog, "install-log", "", "install log")
	flag.StringVar(&flagSrc, "srcs", fmt.Sprintf("./datakit-%s-%s-%s.tar.gz,./data.tar.gz", runtime.GOOS, runtime.GOARCH, DataKitVersion), `local path of install files`)
	flag.IntVar(&flagInstallOnly, "install-only", 0, "install only, not start")
	flag.BoolVar(&flagInfo, "info", false, "show installer info")
	flag.BoolVar(&flagOffline, "offline", false, "-offline option removed")
	flag.BoolVar(&flagDownloadOnly, "download-only", false, "only download install packages")

	flag.StringVar(&installer.Dataway, "dataway", "", "DataWay host(https://guance.openway.com?token=xxx)")
	flag.StringVar(&installer.Proxy, "proxy", "", "http proxy http://ip:port for datakit")
	flag.StringVar(&installer.DatakitName, "name", "", "specify DataKit name, example: prod-env-datakit")
	flag.StringVar(&installer.EnableInputs, "enable-inputs", "", "default enable inputs(comma splited, example:cpu,mem,disk)")
	flag.BoolVar(&installer.OTA, "ota", false, "auto update")
	flag.IntVar(&installer.EnableExperimental, "enable-experimental", 0, "")
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

	// gitrepo flags
	flag.StringVar(&installer.GitURL, "git-url", "", "git repo url")
	flag.StringVar(&installer.GitKeyPath, "git-key-path", "", "git repo access private key path")
	flag.StringVar(&installer.GitKeyPW, "git-key-pw", "", "git repo access private use password")
	flag.StringVar(&installer.GitBranch, "git-branch", "", "git repo branch name")
	flag.StringVar(&installer.GitPullInterval, "git-pull-interval", "", "git repo pull interval")

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

	// sink flags
	flag.StringVar(&installer.SinkMetric, "sink-metric", "", "sink for Metric")
	flag.StringVar(&installer.SinkNetwork, "sink-network", "", "sink for Network")
	flag.StringVar(&installer.SinkKeyEvent, "sink-keyevent", "", "sink for Key Event")
	flag.StringVar(&installer.SinkObject, "sink-object", "", "sink for Object")
	flag.StringVar(&installer.SinkCustomObject, "sink-custom-object", "", "sink for CustomObject")
	flag.StringVar(&installer.SinkLogging, "sink-logging", "", "sink for Logging")
	flag.StringVar(&installer.SinkTracing, "sink-tracing", "", "sink for Tracing")
	flag.StringVar(&installer.SinkRUM, "sink-rum", "", "sink for RUM")
	flag.StringVar(&installer.SinkSecurity, "sink-security", "", "sink for Security")
	flag.StringVar(&installer.SinkProfiling, "sink-profile", "", "sink for Profiling")
	flag.StringVar(&installer.LogSinkDetail, "log-sink-detail", "", "log sink detail")

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
				l.Warnf("stop service failed %s, ignored", err.Error())
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
		if err := installer.CheckVersion(git.Version); err != nil {
			l.Fatalf("upgrade datakit: %s", err.Error())
			return
		}

		l.Infof("Upgrading to version %s...", DataKitVersion)
		if err = installer.Upgrade(svc); err != nil {
			l.Warnf("upgrade datakit failed: %s", err.Error())
		}
	} else { // install new datakit
		l.Infof("Installing version %s...", DataKitVersion)
		installer.Install(svc)
	}

	if flagInstallOnly != 0 {
		l.Infof("only install service %s, NOT started", dkservice.ServiceName)
	} else {
		l.Infof("starting service %s...", dkservice.ServiceName)
		if err = service.Control(svc, "start"); err != nil {
			l.Warnf("start service failed: %s", err.Error())
		}
	}

	if err := config.CreateSymlinks(); err != nil {
		l.Errorf("CreateSymlinks: %s", err.Error())
	}

	if err := checkIsNewVersion("http://"+config.Cfg.HTTPAPI.Listen, DataKitVersion); err != nil {
		l.Errorf("checkIsNewVersion: %s", err.Error())
	} else {
		l.Infof("current running datakit is version: %s", DataKitVersion)

		if flagDKUpgrade {
			l.Info(":) Upgrade Success!")
		} else {
			l.Info(":) Install Success!")
		}
		promptReferences()
	}
}

// test if installed/upgraded to expected version.
func checkIsNewVersion(host, version string) error {
	x := struct {
		Content map[string]string `json:"content"`
	}{}

	for i := 0; i < 10; i++ {
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

		defer resp.Body.Close() //nolint:errcheck

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
	fmt.Println("\n\tVisit https://docs.guance.com/datakit/changelog/ to see DataKit change logs.")
	if config.Cfg.HTTPAPI.Listen != "localhost:9529" {
		fmt.Printf("\tUse `datakit monitor --to %s` to see DataKit running status.\n", config.Cfg.HTTPAPI.Listen)
	} else {
		fmt.Println("\tUse `datakit monitor` to see DataKit running status.")
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
		l.Infof("deprecated install path %s not found", olddir)
		return
	}

	if err := service.Control(svc, "uninstall"); err != nil {
		l.Warnf("uninstall service failed: %s", err.Error())
	}

	if err := os.Rename(olddir, datakit.InstallDir); err != nil {
		l.Fatalf("move %s -> %s failed: %s", olddir, datakit.InstallDir, err.Error())
	}
}
