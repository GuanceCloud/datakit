// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/installer/installer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/upgrader/upgrader"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dl "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/downloader"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
)

var (
	DataKitVersion = "not-set"
	DataKitBaseURL = "not-set"

	oldInstallDir      = "/usr/local/cloudcare/dataflux/datakit"
	oldInstallDirWin   = `C:\Program Files\dataflux\datakit`
	oldInstallDirWin32 = `C:\Program Files (x86)\dataflux\datakit`

	l = logger.DefaultSLogger("installer")

	args = installer.DefaultInstallArgs()
)

//nolint:gochecknoinits,lll
func init() {
	flag.BoolVar(&args.FlagDKUpgrade, "upgrade", false, "")

	flag.IntVar(&args.FlagUpgraderEnabled, "upgrade-manager", 0, "whether we should upgrade the Datakit upgrade service")
	flag.StringVar(&args.FlagUpgraderIPWhiteList, "upgrade-ip-whitelist", "", "set datakit upgrade http service allowed request client ip, split by ','")
	flag.StringVar(&args.FlagUpgraderListen, "upgrade-listen", "0.0.0.0:9542", "set datakit upgrade HTTP server bind ip:port")

	flag.StringVar(&args.FlagInstallLog, "install-log", "dk-install-upgrade.log", "log file during install or upgrade")
	flag.StringVar(&args.FlagSrc, "srcs", fmt.Sprintf("./datakit-%s-%s-%s.tar.gz,./data.tar.gz", runtime.GOOS, runtime.GOARCH, DataKitVersion), `local path of install files`)
	flag.IntVar(&args.FlagInstallOnly, "install-only", 0, "install only, not start")
	flag.BoolVar(&args.FlagInfo, "info", false, "show installer info")
	flag.BoolVar(&args.FlagOffline, "offline", false, "-offline option removed")
	flag.BoolVar(&args.FlagDownloadOnly, "download-only", false, "only download install packages")
	flag.StringVar(&args.DistBaseURL, "installer_base_url", "", "install datakit and data BaseUrl")
	flag.StringVar(&args.FlagUserName, "user-name",
		func() string {
			if runtime.GOOS == datakit.OSLinux {
				return "root"
			} else {
				return ""
			}
		}(), "install log") // user & group.
	flag.StringVar(&args.FlagLite, "lite", "", "install datakit lite")
	flag.StringVar(&args.FlagELinker, "elinker", "", "install datakit eBPF span linker")
	flag.StringVar(&args.InstrumentationEnabled, "apm-instrumentation-enabled", "", "enable APM instrumentation")
	flag.StringVar(&args.DatawayURLs, "dataway", "", "DataWay host(https://guance.openway.com?token=xxx)")
	flag.StringVar(&args.Proxy, "proxy", "", "http proxy http://ip:port for datakit")
	flag.StringVar(&args.DatakitName, "name", "", "specify DataKit name, example: prod-env-datakit")
	flag.StringVar(&args.EnableInputs, "enable-inputs", "", "default enable inputs(comma splited, example:cpu,mem,disk)")
	flag.StringVar(&args.HTTPPublicAPIs, "http-public-apis", "", "set which apis can be access by remote, split by comma.")
	flag.StringVar(&args.HTTPDisabledAPIs, "http-disabled-apis", "", "(Deprecated) set which apis are disallowed access by remote, split by comma.")
	flag.IntVar(&args.InstallRUMSymbolTools, "install-rum-symbol-tools", 0, "whether to install RUM source map tools")
	flag.BoolVar(&args.OTA, "ota", false, "auto update")
	flag.StringVar(&args.InstallExternals, "install-externals", "", "install some external inputs")

	// DCA flags
	flag.StringVar(&args.DCAEnable, "dca-enable", "", "enable DCA")
	flag.StringVar(&args.DCAWebsocketServer, "dca-websocket-server", "", "DCA websocket server")

	// global-host-tags flags
	flag.StringVar(&args.GlobalHostTags, "global-tags", "", "Deprecated, use global-host-tag")
	flag.StringVar(&args.GlobalHostTags, "global-host-tags", "", "enable global host tags, example: host= __datakit_hostname,ip= __datakit_ip")

	// election flags
	flag.StringVar(&args.GlobalElectionTags, "global-election-tags", "", "enable global environment tags, example: project=my-project,cluster=my-cluster")
	flag.StringVar(&args.GlobalElectionTags, "global-env-tags", "", "Deprecated, use --global-election-tags")
	flag.StringVar(&args.EnableElection, "enable-election", "", "datakit election")
	flag.StringVar(&args.ElectionNamespace, "namespace", "", "datakit namespace")

	// datakit HTTP flags
	flag.IntVar(&args.HTTPPort, "port", 0, "datakit HTTP port")
	flag.StringVar(&args.HTTPListen, "listen", "", "datakit HTTP listen")

	flag.StringVar(&args.HostName, "env_hostname", "", "host name")
	flag.StringVar(&args.IPDBType, "ipdb-type", "", "ipdb type")
	flag.StringVar(&args.CloudProvider, "cloud-provider", "", "specify cloud provider(accept aliyun/tencent/aws)")

	flag.Float64Var(&args.WALCapacity, "wal-capacity", 0, "Set WAL cache capacity in GB(default 2.0)")
	flag.IntVar(&args.WALWorkers, "wal-workers", 0, "Set WAL flush workers(default limited-CPU-cores * 8)")

	// confd flags
	flag.StringVar(&args.ConfdBackend, "confd-backend", "", "backend kind")
	flag.StringVar(&args.ConfdBasicAuth, "confd-basic-auth", "", "if backend need auth")
	flag.StringVar(&args.ConfdClientCaKeys, "confd-client-ca-keys", "", "backend ca key")
	flag.StringVar(&args.ConfdClientCert, "confd-client-cert", "", "backend cert key")
	flag.StringVar(&args.ConfdClientKey, "confd-client-key", "", "backend cert key id")
	flag.StringVar(&args.ConfdBackendNodes, "confd-backend-nodes", "", "backend nodes ip")
	flag.StringVar(&args.ConfdPassword, "confd-password", "", "backend login password")
	flag.StringVar(&args.ConfdScheme, "confd-scheme", "", "backend scheme")
	flag.StringVar(&args.ConfdSeparator, "confd-separator", "", "backend separator")
	flag.StringVar(&args.ConfdUsername, "confd-username", "", "backend login username")
	flag.StringVar(&args.ConfdAccessKey, "confd-access-key", "", "backend access key id")
	flag.StringVar(&args.ConfdSecretKey, "confd-secret-key", "", "backend secret key")
	flag.StringVar(&args.ConfdConfdNamespace, "confd-confd-namespace", "", "confd config namespace id")
	flag.StringVar(&args.ConfdPipelineNamespace, "confd-pipeline-namespace", "", "pipeline config namespace id")
	flag.StringVar(&args.ConfdRegion, "confd-region", "", "aws region")
	flag.IntVar(&args.ConfdCircleInterval, "confd-circle-interval", 60, "backend loop search interval second")

	// gitrepo flags
	flag.StringVar(&args.GitURL, "git-url", "", "git repository url")
	flag.StringVar(&args.GitKeyPath, "git-key-path", "", "git repository access private key path")
	flag.StringVar(&args.GitKeyPW, "git-key-pw", "", "git repository access private use password")
	flag.StringVar(&args.GitBranch, "git-branch", "", "git repository branch name")
	flag.StringVar(&args.GitPullInterval, "git-pull-interval", "", "git repository pull interval")

	// rum flags
	flag.StringVar(&args.RumOriginIPHeader, "rum-origin-ip-header", "", "rum only")
	flag.StringVar(&args.RumDisable404Page, "disable-404page", "", "datakit rum 404 page")

	// log flags
	flag.StringVar(&args.LogLevel, "log-level", "", "log level setting")
	flag.StringVar(&args.Log, "log", "", "log setting")
	flag.StringVar(&args.GinLog, "gin-log", "", "gin log setting")

	// pprof flags
	flag.StringVar(&args.PProfListen, "pprof-listen", "", "pprof listen")

	// sinker flags
	flag.StringVar(&args.EnableSinker, "enable-dataway-sinker", "", "enable dataway sinker")
	flag.StringVar(&args.SinkerGlobalCustomerKeys, "sinker-global-customer-keys", "", "sinker configures")

	// resource limit flags
	flag.IntVar(&args.LimitDisabled, "limit-disabled", 0, "enable disable resource limits for CPU and memory in linux and windows")

	flag.Float64Var(&args.LimitCPUMax, "limit-cpumax", 0.0, "CPU max usage(Deprecated: use --limit-cpucores)")
	flag.Float64Var(&args.LimitCPUCores, "limit-cpucores", 0.0, "limited CPU cores")

	flag.Float64Var(&args.LimitCPUMin, "limit-cpumin", 0.0, "CPU min usage, Deprecated")
	flag.Int64Var(&args.LimitMemMax, "limit-memmax", 0, "memory limit")

	flag.StringVar(&args.CryptoAESKey, "crypto-aes_key", "", "ENC crypto for AES key")
	flag.StringVar(&args.CryptoAESKeyFile, "crypto-aes_key_file", "", "ENC crypto for AES key filepath")
}

func setupLogging() error {
	if args.FlagInstallLog == "stdout" {
		cp.Infof("Set log file to stdout\n")

		if err := logger.InitRoot(&logger.Option{
			Level: logger.DEBUG,
			Flags: logger.OPT_DEFAULT | logger.OPT_STDOUT,
		}); err != nil {
			return fmt.Errorf("set root log failed: %w", err)
		}
	} else {
		l.Infof("Set log file to %s\n", args.FlagInstallLog)
		if err := logger.InitRoot(&logger.Option{
			Path:  args.FlagInstallLog,
			Level: logger.DEBUG,
			Flags: logger.OPT_DEFAULT,
		}); err != nil {
			return fmt.Errorf("logger.InitRoot: %w", err)
		}
	}

	// setup module logger
	l = logger.SLogger("installer")

	l.Infof("install/upgrade id: %s", cliutils.XID("iu_"))

	// setup sub-module's logger.
	config.SetLog()
	installer.SetLog()

	return nil
}

// offlineExtract extrac all downloaded files to installer dirs.
func offlineExtract() error {
	for _, f := range strings.Split(args.FlagSrc, ",") {
		fd, err := os.Open(filepath.Clean(f))
		if err != nil {
			return fmt.Errorf("Open(%q): %w", f, err)
		}

		// default extract to /usr/local/datakit
		destDir := datakit.InstallDir

		switch {
		// dk_upgrader should extract to dir /usr/local/dk_upgrader
		case strings.HasPrefix(f, "dk_upgrader"): // e.g., dk_upgrader-linux-amd64.tar.gz
			destDir = upgrader.InstallDir
		default: // pass: others are datakit.tar.gz and data.tar.gz
		}

		if err := dl.Extract(fd, destDir); err != nil {
			return fmt.Errorf("download Extract(): %w", err)
		} else if err := fd.Close(); err != nil {
			l.Warnf("Close: %s, ignored", err)
		}
	}

	return nil
}

func applyFlags(mc *config.Config) (err error) {
	args.DataKitVersion = DataKitVersion
	args.DataKitBaseURL = DataKitBaseURL

	// NOTE: we should update base-urls before other settings
	if err = args.UpdateDownloadURLs(); err != nil {
		return fmt.Errorf("options.UpdateDownloadURLs(): %w", err)
	}

	// show installer info
	if args.FlagInfo {
		cp.Printf(`
Version        : %s
Build At       : %s
Golang Version : %s
BaseUrl        : %s
Data           : %s
`, DataKitVersion, git.BuildAt, git.Golang,
			DataKitBaseURL, args.DistDataURL)
		os.Exit(0)
	}

	if err = setupLogging(); err != nil {
		return err
	}

	args.SetDatakitLiteAndELinker()

	if !args.FlagDKUpgrade || args.FlagUpgraderEnabled == 1 {
		upgrader.StopUpgradeService(args.FlagUserName)
	}

	if args.FlagDownloadOnly {
		if err = args.DownloadFiles(""); err != nil { // download 过程直接覆盖已有安装
			cp.Errorf("download failed: %s", err.Error())
		}
		os.Exit(0)
	}

	if args.FlagSrc != "" && args.FlagOffline {
		if err = offlineExtract(); err != nil {
			l.Warnf("offlineExtract: %s, ignored", err)
		}
	}

	// try add 'http://' prefix to proxy.
	if args.Proxy != "" {
		if !strings.HasPrefix(args.Proxy, "http") {
			args.Proxy = "http://" + args.Proxy
		}

		if _, err := url.Parse(args.Proxy); err != nil {
			l.Warnf("bad proxy config expect http://ip:port given %s", args.Proxy)
		} else {
			l.Infof("set proxy to %s", args.Proxy)
		}
	}

	// setup working dirs
	datakit.InitDirs()

	// load args and apply them into datakit.conf(mc)
	mc, err = args.LoadInstallerArgs(mc)
	if err != nil {
		return err
	}

	if err = args.SetupUserGroup(mc); err != nil {
		l.Errorf("SetupUserGroup: %s, ignored", err)
	}

	return nil
}

func retryingDownloadFiles(dlRetry int) error {
	l.Infof("Download installer and data files(with %d retry)...", dlRetry)

	for i := 0; i < dlRetry; i++ {
		if err := args.DownloadFiles(datakit.InstallDir); err != nil { // download 过程直接覆盖已有安装
			cp.Warnf("download failed: %s, %dth retry...\n", err.Error(), i)
			continue
		}
		return nil
	}

	return fmt.Errorf("download failed")
}

func main() {
	flag.Parse()
	if err := applyFlags(config.Cfg); err != nil {
		cp.Errorf("applyFlags: %s", err.Error())
		os.Exit(-1)
	}

	// create datakit system service
	svc, err := args.SetupService()
	if err != nil {
		cp.Errorf("SetupService: %s", err.Error())
		os.Exit(-1)
	}

	// 迁移老版本 datakit 数据目录
	moveOldDatakit(svc)

	if !args.FlagOffline {
		if err := retryingDownloadFiles(5); err != nil {
			cp.Errorf("Download installer and data files failed, please check your network settings and check installer log at %s.\n", args.FlagInstallLog)
			return
		}
	}

	if args.FlagDKUpgrade { // upgrade new version
		l.Infof("Upgrading to version %s...", DataKitVersion)
		if err = args.Upgrade(config.Cfg); err != nil {
			l.Warnf("upgrade datakit failed: %s, ignored", err.Error())
		}
	} else { // install new datakit
		l.Infof("Installing version %s...", DataKitVersion)
		if err := args.Install(config.Cfg, svc); err != nil {
			l.Fatalf("Install: %s", err)
		}
	}

	if args.FlagInstallOnly != 0 {
		l.Warnf("Only install service %q, NOT started", dkservice.Name())
	} else {
		if err = service.Control(svc, "start"); err != nil {
			l.Warnf("Start service %q failed: %s", dkservice.Name(), err.Error())
		} else {
			l.Infof("Starting service %q ok", dkservice.Name())
		}
	}

	if err := config.CreateSymlinks(); err != nil {
		l.Errorf("CreateSymlinks failed: %s", err.Error())
		l.Infof("Your may need to run datakit command under install path %q", datakit.InstallDir)
	} else {
		l.Infof("Create symlinks ok")
	}

	// Checking running datakit version to make sure the new version is installed and running ok.
	if args.FlagInstallOnly == 0 {
		if err := checkIsNewVersion("http://"+config.Cfg.HTTPAPI.Listen, DataKitVersion); err != nil {
			l.Warnf("Check current datakit version failed(expect version %q), we can ignore the message and go on", DataKitVersion)
			l.Infof("Please visit https://docs.guance.com/datakit/datakit-update/#version-check-failed to see more info about the version checking.")
		} else {
			l.Infof("Current datakit version is %q", DataKitVersion)
		}
	} else {
		l.Infof("Current datakit version is %q(NOT running)", DataKitVersion)
	}

	// After datakit setup/upgrade ok, it's time for dk-upgrader service.
	args.SetupUpgraderService()

	if args.FlagDKUpgrade {
		l.Infof("Upgrade OK.")
	} else {
		l.Infof("Install OK.")
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
			l.Warnf("get datakit current version failed: %s, %dth retrying...", err, i)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			l.Errorf("io.ReadAll: %s", err)
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

func promptReferences() {
	cp.Infof("\nVisit https://docs.guance.com/datakit/changelog-%d/ to see DataKit change logs.\n", time.Now().Year())
	cp.Infof("Use `datakit monitor` to see DataKit running status.\n")
}

func moveOldDatakit(svc service.Service) {
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
