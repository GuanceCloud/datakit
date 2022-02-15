package cmds

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/spf13/pflag"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	FlagUpdateLogFile string
	FlagVersion,
	FlagShowTestingVersions,
	FlagAcceptRCVersion, // deprecated
	FlagCheckUpdate bool

	FlagGrokq bool
	FlagPipeline,
	FlagText string

	FlagProm      string
	FlagTestInput string

	FlagDefConf bool
	FlagWorkDir string

	FlagDisableTFMono, FlagMan bool
	FlagIgnore,
	FlagExportManuals, // Deprecated
	FlagExportIntegration,
	FlagManVersion, // Deprecated
	FlagTODO string
	FlagExportMetaInfo string

	FlagInstallExternal string

	FlagStart,
	FlagStop,
	FlagRestart,
	FlagAPIRestart,
	FlagStatus,
	FlagUninstall,
	FlagReinstall bool

	FlagDQL      bool
	FlagJSON     bool
	FlagAutoJSON bool
	FlagForce    bool

	FlagRunDQL,
	FlagCSV string

	FlagToken         string
	FlagWorkspaceInfo bool

	FlagUpdateIPDB bool
	FlagAddr       string
	FlagInterval   time.Duration

	FlagShowCloudInfo    string
	FlagIPInfo           string
	FlagConfigDir        string
	FlagMonitor          bool
	FlagCheckConfig      bool
	FlagCheckSample      bool
	FlagDocker           bool
	FlagDisableSelfInput bool
	FlagVVV              bool
	FlagCmdLogPath       string
	FlagDumpSamples      string

	FlagUploadLog bool
)

func initOldStyleFlags() { //nolint:gochecknoinits
	pflag.BoolVarP(&FlagVersion, "version", "V", false, `show version info`)
	pflag.BoolVar(&FlagCheckUpdate, "check-update", false, "check if new version available")
	pflag.BoolVar(&FlagAcceptRCVersion, "accept-rc-version", false, "during update, accept RC version if available")
	pflag.BoolVar(&FlagShowTestingVersions, "show-testing-version", false, "show testing versions on -version flag")
	pflag.StringVar(&FlagUpdateLogFile, "update-log", "", "update history log file")

	pflag.StringVar(&FlagWorkDir, "workdir", "", "set datakit work dir")
	pflag.BoolVar(&FlagDefConf, "default-main-conf", false, "get datakit default main configure")

	// debug grok
	pflag.StringVar(&FlagPipeline, "pl", "", "pipeline script to test(name only, do not use file path)")
	pflag.BoolVar(&FlagGrokq, "grokq", false, "query groks interactively")
	pflag.StringVar(&FlagText, "txt", "", "text string for the pipeline or grok(json or raw text)")

	pflag.StringVar(&FlagProm, "prom-conf", "", "prom config file to test")
	pflag.StringVar(&FlagTestInput, "test-input", "", "specify config file to test")

	// manuals related
	pflag.BoolVar(&FlagMan, "man", false, "read manuals of inputs")
	pflag.StringVar(&FlagExportManuals, "export-manuals", "", "export all inputs and related manuals to specified path")
	pflag.StringVar(&FlagExportMetaInfo, "export-metainfo", "", "output metainfo to specified file")
	pflag.BoolVar(&FlagDisableTFMono, "disable-tf-mono", false, "use normal font on tag/field")
	pflag.StringVar(&FlagIgnore, "ignore", "", "disable list, i.e., --ignore nginx,redis,mem")
	pflag.StringVar(&FlagExportIntegration, "export-integration", "", "export all integrations")
	pflag.StringVar(&FlagManVersion, "man-version", datakit.Version, "specify manuals version")
	pflag.StringVar(&FlagTODO, "TODO", "TODO", "set TODO")

	// install 3rd-party kit
	pflag.StringVar(&FlagInstallExternal, "install", "", "install external tool/software")

	// managing service
	pflag.BoolVar(&FlagStart, "start", false, "start datakit")
	pflag.BoolVar(&FlagStop, "stop", false, "stop datakit")
	pflag.BoolVar(&FlagRestart, "restart", false, "restart datakit")
	pflag.BoolVar(&FlagAPIRestart, "api-restart", false, "restart datakit for api only")
	pflag.BoolVar(&FlagStatus, "status", false, "show datakit service status")
	pflag.BoolVar(&FlagUninstall, "uninstall", false, "uninstall datakit service(not delete DataKit files)")
	pflag.BoolVar(&FlagReinstall, "reinstall", false, "re-install datakit service")

	// DQL
	pflag.BoolVarP(&FlagDQL, "dql", "Q", false, "under DQL, query interactively")
	pflag.BoolVar(&FlagJSON, "json", false, "under DQL, output in json format")
	pflag.BoolVar(&FlagForce, "force", false, "Mandatory modification")
	pflag.BoolVar(&FlagAutoJSON, "auto-json", false, "under DQL, pretty output string if it's JSON")
	pflag.StringVar(&FlagRunDQL, "run-dql", "", "run single DQL")
	pflag.StringVar(&FlagToken, "token", "", "query under specific token")
	pflag.StringVar(&FlagCSV, "csv", "", "Specify the directory")

	// update online data
	pflag.BoolVar(&FlagUpdateIPDB, "update-ip-db", false, "update ip db")
	pflag.StringVarP(&FlagAddr, "addr", "A", "", "url path")
	pflag.DurationVar(&FlagInterval, "interval", time.Second*5, "auxiliary option, time interval")

	// utils
	pflag.StringVar(&FlagShowCloudInfo, "show-cloud-info", "", "show current host's cloud info(aliyun/tencent/aws)")
	pflag.StringVar(&FlagIPInfo, "ipinfo", "", "show IP geo info")
	pflag.BoolVar(&FlagWorkspaceInfo, "workspace-info", false, "show workspace info")

	if runtime.GOOS != datakit.OSWindows { // unsupported options under windows
		pflag.BoolVarP(&FlagMonitor, "monitor", "M", false, "show monitor info of current datakit")
		pflag.BoolVar(&FlagDocker, "docker", false, "run within docker")
	}

	pflag.BoolVar(&FlagCheckConfig, "check-config", false, "check inputs configure and main configure")
	pflag.StringVar(&FlagConfigDir, "config-dir", "", "check configures under specified path")
	pflag.BoolVar(&FlagCheckSample, "check-sample", false, "check inputs configure samples")
	pflag.BoolVar(&FlagVVV, "vvv", false, "more verbose info")
	pflag.StringVar(&FlagCmdLogPath, "cmd-log", "/dev/null", "command line log path")
	pflag.StringVar(&FlagDumpSamples, "dump-samples", "", "dump all inputs samples")

	pflag.BoolVar(&config.DisableSelfInput, "disable-self-input", false, "disable self input")
	pflag.BoolVar(&io.DisableDatawayList, "disable-dataway-list", false, "disable list available dataway")
	pflag.BoolVar(&io.DisableLogFilter, "disable-logfilter", false, "disable logfilter")
	pflag.BoolVar(&io.DisableHeartbeat, "disable-heartbeat", false, "disable heartbeat")

	pflag.BoolVar(&FlagUploadLog, "upload-log", false, "upload log")
}

// setupFlags deprecated
func setupFlags() {
	// hidden flags
	for _, f := range []string{
		"TODO",
		"man-version",
		"export-integration",
		"addr",
		"show-testing-version",
		"update-log",
		"dump-samples",
		"workdir",
		"default-main-conf",
		"disable-self-input",
		"disable-dataway-list",
		"disable-logfilter",
		"disable-heartbeat",
		"api-restart",
		"export-metainfo",
	} {
		if err := pflag.CommandLine.MarkHidden(f); err != nil {
			l.Warnf("CommandLine.MarkHidden: %s, ignored", err)
		}
	}

	pflag.CommandLine.SortFlags = false
	pflag.ErrHelp = errors.New("") // disable `pflag: help requested`

	if runtime.GOOS == datakit.OSWindows {
		FlagCmdLogPath = "nul" // under windows, nul is /dev/null
	}
}

// parseOldStyleFlags deprecated
func parseOldStyleFlags() {
	setupFlags()
	pflag.Parse()
}

//nolint:funlen,gocyclo
func runOldStyleCmds() {
	if FlagDefConf {
		defconf := config.DefaultConfig()
		fmt.Println(defconf.String())
		os.Exit(0)
	}

	if FlagCheckUpdate { // 更新日志单独存放，不跟 cmd.log 一块
		tryLoadMainCfg()

		if FlagUpdateLogFile != "" {
			if err := logger.InitRoot(&logger.Option{
				Path:  FlagUpdateLogFile,
				Level: logger.DEBUG,
				Flags: logger.OPT_DEFAULT,
			}); err != nil {
				l.Errorf("set root log faile: %s", err.Error())
			}
		}

		// deprecated: after 1.2.x, RC version can't be upgraded, see issue #484
		if FlagAcceptRCVersion {
			warnf("[W] --accept-rc-version deprecated\n")
		}

		ret := checkUpdate(ReleaseVersion, false)
		os.Exit(ret)
	}

	if FlagVersion {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)
		showVersion(ReleaseVersion, InputsReleaseType)

		vis, err := checkNewVersion(ReleaseVersion, FlagShowTestingVersions)
		if err != nil {
			errorf("get online version info failed: %s\n", err)
			os.Exit(-1)
		}

		for _, vi := range vis {
			infof("\n\n%s version available: %s, commit %s (release at %s)\n\nUpgrade:\n\t",
				vi.versionType, vi.newVersion.VersionString, vi.newVersion.Commit, vi.newVersion.ReleaseDate)
			infof("%s\n", getUpgradeCommand(vi.newVersion.DownloadURL))
		}

		os.Exit(0)
	}

	if FlagCheckConfig {
		confdir := FlagConfigDir
		if confdir == "" {
			tryLoadMainCfg()
			confdir = datakit.ConfdDir
		}

		setCmdRootLog(FlagCmdLogPath)
		if err := checkConfig(confdir, ""); err != nil {
			os.Exit(-1)
		}
		os.Exit(0)
	}

	if FlagCheckSample {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)
		if err := checkSample(); err != nil {
			os.Exit(-1)
		}
		os.Exit(0)
	}

	if FlagDQL || FlagRunDQL != "" {

		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)

		dc := &dqlCmd{
			json:          FlagJSON,
			autoJson:      FlagAutoJSON,
			dqlString:     FlagRunDQL,
			token:         FlagToken,
			csv:           FlagCSV,
			forceWriteCSV: FlagForce,
		}

		if err := dc.prepare(); err != nil {
			errorf("dc.prepare: %s\n", err.Error())
			os.Exit(1)
		}

		infof("dqlcmd: %+#v\n", dc)
		dc.run()
		os.Exit(0)
	}

	if FlagWorkspaceInfo {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)
		requrl := fmt.Sprintf("http://%s%s", config.Cfg.HTTPAPI.Listen, workspace)
		body, err := doWorkspace(requrl)
		if err != nil {
			errorf("get worksapceInfo fail %s\n", err.Error())
		}
		outputWorkspaceInfo(body)
		os.Exit(0)
	}
	if FlagShowCloudInfo != "" {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)
		info, err := showCloudInfo(FlagShowCloudInfo)
		if err != nil {
			errorf("[E] Get cloud info failed: %s\n", err.Error())
			os.Exit(-1)
		}

		keys := []string{}
		for k := range info {
			keys = append(keys, k)
		}

		sort.Strings(keys)
		for _, k := range keys {
			infof("\t% 24s: %v\n", k, info[k])
		}

		os.Exit(0)
	}

	if FlagMonitor {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)

		cmdMonitor(FlagInterval, FlagVVV)
		os.Exit(0)
	}

	if FlagIPInfo != "" {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)
		x, err := ipInfo(FlagIPInfo)
		if err != nil {
			errorf("[E] get IP info failed: %s\n", err.Error())
		} else {
			for k, v := range x {
				infof("\t% 8s: %s\n", k, v)
			}
		}

		os.Exit(0)
	}

	if FlagDumpSamples != "" {
		tryLoadMainCfg()
		fpath := FlagDumpSamples

		if err := os.MkdirAll(fpath, datakit.ConfPerm); err != nil {
			panic(err)
		}

		for k, v := range inputs.Inputs {
			sample := v().SampleConfig()
			if err := ioutil.WriteFile(filepath.Join(fpath, k+".conf"),
				[]byte(sample), datakit.ConfPerm); err != nil {
				panic(err)
			}
		}
		os.Exit(0)
	}

	if FlagPipeline != "" {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)
		if err := pipelineDebugger(FlagPipeline, FlagText); err != nil {
			errorf("[E] %s\n", err)
			os.Exit(-1)
		}

		os.Exit(0)
	}

	if FlagProm != "" {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)
		if err := promDebugger(FlagProm); err != nil {
			errorf("[E] %s\n", err)
			os.Exit(-1)
		}

		os.Exit(0)
	}

	if FlagTestInput != "" {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)
		if err := inputDebugger(FlagTestInput); err != nil {
			l.Errorf("inputDebugger: %s", err)
		}

		os.Exit(0)
	}

	if FlagGrokq {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)
		grokq()
		os.Exit(0)
	}

	if FlagMan {
		setCmdRootLog(FlagCmdLogPath)
		cmdMan()
		os.Exit(0)
	}

	if FlagExportManuals != "" {
		setCmdRootLog(FlagCmdLogPath)
		if err := exportMan(FlagExportManuals,
			FlagIgnore,
			FlagManVersion,
			FlagDisableTFMono); err != nil {
			l.Error(err)
		}
		os.Exit(0)
	}

	if FlagExportIntegration != "" {
		setCmdRootLog(FlagCmdLogPath)
		if err := exportIntegration(FlagExportIntegration, FlagIgnore); err != nil {
			l.Error(err)
		}
		os.Exit(0)
	}
	if FlagExportMetaInfo != "" {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)
		if err := ExportMetaInfo(FlagExportMetaInfo); err != nil {
			l.Error(err)
		}
		os.Exit(0)
	}
	if FlagInstallExternal != "" {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)

		if err := installExternal(FlagInstallExternal); err != nil {
			l.Error(err)
		}
		os.Exit(0)
	}

	if FlagStart {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)

		if err := startDatakit(); err != nil {
			errorf("[E] start DataKit failed: %s\n using command to stop : %s\n", err.Error(), errMsg[runtime.GOOS])
			os.Exit(-1)
		}

		infof("Start DataKit OK\n") // TODO: 需说明 PID 是多少
		os.Exit(0)
	}

	if FlagStop {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)

		if err := stopDatakit(); err != nil {
			errorf("[E] stop DataKit failed: %s\n", err.Error())
			os.Exit(-1)
		}

		infof("Stop DataKit OK\n")
		os.Exit(0)
	}

	if FlagRestart {
		tryLoadMainCfg()

		setCmdRootLog(FlagCmdLogPath)

		if err := restartDatakit(); err != nil {
			errorf("[E] restart DataKit failed:%s\n using command to restart: %s\n", err.Error(), errMsg[runtime.GOOS])
			os.Exit(-1)
		}

		infof("Restart DataKit OK\n")
		os.Exit(0)
	}

	if FlagStatus {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)
		x, err := datakitStatus()
		if err != nil {
			errorf("[E] get DataKit status failed: %s\n", err.Error())
			os.Exit(-1)
		}
		infof("%s\n", x)
		os.Exit(0)
	}

	if FlagUninstall {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)
		if err := uninstallDatakit(); err != nil {
			errorf("[E] uninstall DataKit failed: %s\n", err.Error())
			os.Exit(-1)
		}

		infof("Uninstall DataKit OK\n")
		os.Exit(0)
	}

	if FlagReinstall {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)
		if err := reinstallDatakit(); err != nil {
			errorf("[E] reinstall DataKit failed: %s\n", err.Error())
			os.Exit(-1)
		}

		infof("Reinstall DataKit OK\n")
		os.Exit(0)
	}

	if FlagUpdateIPDB {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)

		if err := updateIPDB(FlagAddr); err != nil {
			errorf("[E] update IPDB failed: %s\n", err.Error())
			os.Exit(-1)
		}

		infof("Update IPdb OK, please restart datakit to load new IPDB\n")
		os.Exit(0)
	}

	if FlagAPIRestart {
		tryLoadMainCfg()
		logPath := config.Cfg.Logging.Log
		setCmdRootLog(logPath)
		apiRestart()
		os.Exit(0)
	}

	if FlagUploadLog {
		tryLoadMainCfg()
		infof("Upload log start...\n")
		if err := uploadLog(config.Cfg.DataWay.URLs); err != nil {
			errorf("[E] upload log failed : %s\n", err.Error())
			os.Exit(-1)
		}
		infof("Upload ok.\n")
		os.Exit(0)
	}
}
