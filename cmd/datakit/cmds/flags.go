package cmds

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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
	FlagReload,
	FlagStatus,
	FlagUninstall,
	FlagReinstall bool

	FlagDatakitHost string

	FlagDQL    bool
	FlagRunDQL string

	FlagUpdateIPDB bool
	FlagAddr       string
	FlagInterval   time.Duration

	FlagShowCloudInfo string
	FlagIPInfo        string
	FlagMonitor       bool
	FlagCheckConfig   bool
	FlagDocker        bool
	FlagVVV           bool
	FlagCmdLogPath    string
	FlagDumpSamples   string

	ReleaseVersion string
	ReleaseType    string
)

func RunCmds() {

	if FlagCheckUpdate { // 更新日志单独存放，不跟 cmd.log 一块
		if FlagUpdateLogFile != "" {
			if err := logger.SetGlobalRootLogger(FlagUpdateLogFile, logger.DEBUG, logger.OPT_DEFAULT); err != nil {
				l.Errorf("set root log faile: %s", err.Error())
			}
		}
		ret := checkUpdate(ReleaseVersion, FlagAcceptRCVersion)
		os.Exit(ret)
	}

	if FlagVersion {
		setCmdRootLog(FlagCmdLogPath)

		showVersion(ReleaseVersion, ReleaseType, FlagShowTestingVersions)
		os.Exit(0)
	}

	if FlagCheckConfig {
		setCmdRootLog(FlagCmdLogPath)
		checkConfig()
		os.Exit(0)
	}

	if FlagDQL {
		setCmdRootLog(FlagCmdLogPath)
		dql(FlagDatakitHost)
		os.Exit(0)
	}

	if FlagRunDQL != "" {
		setCmdRootLog(FlagCmdLogPath)
		doDQL(FlagRunDQL)
		os.Exit(0)
	}

	if FlagShowCloudInfo != "" {
		setCmdRootLog(FlagCmdLogPath)
		info, err := showCloudInfo(FlagShowCloudInfo)
		if err != nil {
			fmt.Printf("Get cloud info failed: %s\n", err.Error())
			os.Exit(-1)
		}

		keys := []string{}
		for k, _ := range info {
			keys = append(keys, k)
		}

		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("\t% 24s: %v\n", k, info[k])
		}

		os.Exit(0)
	}

	if FlagMonitor {
		setCmdRootLog(FlagCmdLogPath)
		if runtime.GOOS == "windows" {
			fmt.Println("unsupport under Windows")
			os.Exit(-1)
		}

		cmdMonitor(FlagInterval, FlagDatakitHost, FlagVVV)
		os.Exit(0)
	}

	if FlagIPInfo != "" {
		setCmdRootLog(FlagCmdLogPath)
		x, err := ipInfo(FlagIPInfo)
		if err != nil {
			fmt.Printf("\t%v\n", err)
		} else {
			for k, v := range x {
				fmt.Printf("\t% 8s: %s\n", k, v)
			}
		}

		os.Exit(0)
	}

	if FlagDumpSamples != "" {
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
		setCmdRootLog(FlagCmdLogPath)
		pipelineDebugger(FlagPipeline, FlagText)
		os.Exit(0)
	}

	if FlagProm != "" {
		setCmdRootLog(FlagCmdLogPath)
		promDebugger(FlagProm)
		os.Exit(0)
	}

	if FlagGrokq {
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
		setCmdRootLog(FlagCmdLogPath)

		if err := installExternal(FlagInstallExternal); err != nil {
			l.Error(err)
		}
		os.Exit(0)
	}

	if FlagStart {

		setCmdRootLog(FlagCmdLogPath)

		if err := startDatakit(); err != nil {
			fmt.Printf("Start DataKit failed: %s\n", err)
			os.Exit(-1)
		}

		fmt.Println("Start DataKit OK") // TODO: 需说明 PID 是多少
		os.Exit(0)
	}

	if FlagStop {

		setCmdRootLog(FlagCmdLogPath)

		if err := stopDatakit(); err != nil {
			fmt.Printf("Stop DataKit failed: %s\n", err)
			os.Exit(-1)
		}

		fmt.Println("Stop DataKit OK")
		os.Exit(0)
	}

	if FlagRestart {

		setCmdRootLog(FlagCmdLogPath)

		if err := restartDatakit(); err != nil {
			fmt.Printf("Restart DataKit failed: %s\n", err)
			os.Exit(-1)
		}

		fmt.Println("Restart DataKit OK")
		os.Exit(0)
	}

	if FlagReload {
		if runtime.GOOS == "windows" {
			fmt.Println("unsupport under Windows")
			os.Exit(-1)
		}

		setCmdRootLog(FlagCmdLogPath)

		if err := reloadDatakit(FlagDatakitHost); err != nil {
			fmt.Printf("Reload DataKit Failed\n")
			os.Exit(-1)
		}

		fmt.Println("Reload DataKit OK")
		os.Exit(0)
	}

	if FlagStatus {

		setCmdRootLog(FlagCmdLogPath)
		x, err := datakitStatus()
		if err != nil {
			fmt.Printf("Get DataKit status failed: %s\n", err)
			os.Exit(-1)
		}
		fmt.Println(x)
		os.Exit(0)
	}

	if FlagUninstall {
		setCmdRootLog(FlagCmdLogPath)
		if err := uninstallDatakit(); err != nil {
			fmt.Printf("Get DataKit status failed: %s\n", err)
			os.Exit(-1)
		}
		os.Exit(0)
	}

	if FlagReinstall {
		setCmdRootLog(FlagCmdLogPath)
		if err := reinstallDatakit(); err != nil {
			fmt.Printf("Reinstall DataKit failed: %s\n", err)
			os.Exit(-1)
		}
		os.Exit(0)
	}

	if FlagUpdateIPDB {
		setCmdRootLog(FlagCmdLogPath)

		if runtime.GOOS == datakit.OSWindows {
			fmt.Println("[E] not supported")
			os.Exit(-1)
		}

		if err := updateIPDB(FlagDatakitHost, FlagAddr); err != nil {
			fmt.Printf("Reload DataKit failed: %s\n", err)
			os.Exit(-1)
		}

		fmt.Println("Update IPdb ok!")

		os.Exit(0)
	}
}
