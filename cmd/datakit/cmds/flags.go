package cmds

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	FlagUpdateLogFile string
	FlagVersion,
	FlagShowTestingVersions,
	FlagCheckUpdate,
	FlagAcceptRCVersion bool

	FlagGrokq bool
	FlagPipeline,
	FlagText string

	FlagProm string

	FlagDefConf bool
	FlagWorkDir string

	FlagMan bool
	FlagIgnore,
	FlagExportMan,
	FlagExportIntegration,
	FlagManVersion,
	FlagTODO string

	FlagInstallExternal string

	FlagStart,
	FlagStop,
	FlagRestart,
	FlagApiRestart,
	FlagStatus,
	FlagUninstall,
	FlagReinstall bool

	FlagDQL      bool
	FlagJSON     bool
	FlagAutoJSON bool
	FlagForce    bool
	FlagCSV      string
	FlagRunDQL,  // TODO: dump dql query result to specified CSV file

	FlagToken string
	FlagWorkspaceInfo bool

	FlagUpdateIPDB bool
	FlagAddr       string
	FlagInterval   time.Duration

	FlagShowCloudInfo    string
	FlagIPInfo           string
	FlagMonitor          bool
	FlagCheckConfig      bool
	FlagCheckSample      bool
	FlagDocker           bool
	FlagDisableSelfInput bool
	FlagVVV              bool
	FlagCmdLogPath       string
	FlagDumpSamples      string
)

var (
	ReleaseVersion string
	ReleaseType    string
)

func tryLoadMainCfg() {
	if err := config.Cfg.LoadMainTOML(datakit.MainConfPath); err != nil {
		l.Warnf("load config %s failed: %s, ignore", datakit.MainConfPath, err)
	}
}

//nolint:funlen,gocyclo
func RunCmds() {
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
		ret := checkUpdate(ReleaseVersion, FlagAcceptRCVersion)
		os.Exit(ret)
	}

	if FlagVersion {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)
		showVersion(ReleaseVersion, ReleaseType, FlagShowTestingVersions)
		os.Exit(0)
	}

	if FlagCheckConfig {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)
		checkConfig()
		os.Exit(0)
	}

	if FlagCheckSample {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)
		checkSample()
		os.Exit(0)
	}

	if FlagDQL {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)
		dql(config.Cfg.HTTPAPI.Listen)
		os.Exit(0)
	}

	if FlagRunDQL != "" {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)
		datakitHost = config.Cfg.HTTPAPI.Listen
		runSingleDQL(FlagRunDQL)
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
			fmt.Printf("\t% 24s: %v\n", k, info[k])
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
				fmt.Printf("\t% 8s: %s\n", k, v)
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
		promDebugger(FlagProm) //nolint:errcheck
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

	if FlagExportMan != "" {
		setCmdRootLog(FlagCmdLogPath)
		if err := exportMan(FlagExportMan, FlagIgnore, FlagManVersion); err != nil {
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
			errorf("[E] start DataKit failed: %s\n", err.Error())
			os.Exit(-1)
		}

		fmt.Println("Start DataKit OK") // TODO: 需说明 PID 是多少
		os.Exit(0)
	}

	if FlagStop {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)

		if err := stopDatakit(); err != nil {
			errorf("[E] stop DataKit failed: %s\n", err.Error())
			os.Exit(-1)
		}

		fmt.Println("Stop DataKit OK")
		os.Exit(0)
	}

	if FlagRestart {
		tryLoadMainCfg()

		setCmdRootLog(FlagCmdLogPath)

		if err := restartDatakit(); err != nil {
			errorf("[E] restart DataKit failed: %s\n", err.Error())
			os.Exit(-1)
		}

		fmt.Println("Restart DataKit OK")
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
		fmt.Println(x)
		os.Exit(0)
	}

	if FlagUninstall {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)
		if err := uninstallDatakit(); err != nil {
			errorf("[E] uninstall DataKit failed: %s\n", err.Error())
			os.Exit(-1)
		}

		fmt.Println("Uninstall DataKit OK")
		os.Exit(0)
	}

	if FlagReinstall {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)
		if err := reinstallDatakit(); err != nil {
			errorf("[E] reinstall DataKit failed: %s\n", err.Error())
			os.Exit(-1)
		}

		fmt.Println("Reinstall DataKit OK")
		os.Exit(0)
	}

	if FlagUpdateIPDB {
		tryLoadMainCfg()
		setCmdRootLog(FlagCmdLogPath)

		if err := updateIPDB(FlagAddr); err != nil {
			errorf("[E] update IPDB failed: %s\n", err.Error())
			os.Exit(-1)
		}

		fmt.Println("Update IPdb OK, please restart datakit to load new IPDB")
		os.Exit(0)
	}

	if FlagApiRestart {
		tryLoadMainCfg()
		logPath := config.Cfg.Logging.Log
		setCmdRootLog(logPath)
		apiRestart()
		os.Exit(0)
	}
}

func getcli() *http.Client {
	proxy := config.Cfg.DataWay.HttpProxy

	cliopt := &ihttp.Options{
		InsecureSkipVerify: true,
	}

	if proxy != "" {
		if u, err := url.Parse(proxy); err == nil {
			cliopt.ProxyURL = u
		}
	}

	return ihttp.Cli(cliopt)
}
