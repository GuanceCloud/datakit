package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	flag "github.com/spf13/pflag"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/gitrepo"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cgroup"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tracer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/election"
	plworker "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/worker"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/all"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/pythond"
)

func init() { //nolint:gochecknoinits
	flag.BoolVarP(&cmds.FlagVersion, "version", "V", false, `show version info`)
	flag.BoolVar(&cmds.FlagCheckUpdate, "check-update", false, "check if new version available")
	flag.BoolVar(&cmds.FlagAcceptRCVersion, "accept-rc-version", false, "during update, accept RC version if available")
	flag.BoolVar(&cmds.FlagShowTestingVersions, "show-testing-version", false, "show testing versions on -version flag")
	flag.StringVar(&cmds.FlagUpdateLogFile, "update-log", "", "update history log file")

	flag.StringVar(&cmds.FlagWorkDir, "workdir", "", "set datakit work dir")
	flag.BoolVar(&cmds.FlagDefConf, "default-main-conf", false, "get datakit default main configure")

	// debug grok
	flag.StringVar(&cmds.FlagPipeline, "pl", "", "pipeline script to test(name only, do not use file path)")
	flag.BoolVar(&cmds.FlagGrokq, "grokq", false, "query groks interactively")
	flag.StringVar(&cmds.FlagText, "txt", "", "text string for the pipeline or grok(json or raw text)")

	flag.StringVar(&cmds.FlagProm, "prom-conf", "", "prom config file to test")
	flag.StringVar(&cmds.FlagTestInput, "test-input", "", "specify config file to test")

	// manuals related
	flag.BoolVar(&cmds.FlagMan, "man", false, "read manuals of inputs")
	flag.StringVar(&cmds.FlagExportMan, "export-manuals", "", "export all inputs and related manuals to specified path")
	flag.StringVar(&cmds.FlagExportMetaInfo, "export-metainfo", "", "output metainfo to specified file")
	flag.BoolVar(&cmds.FlagDisableTFMono, "disable-tf-mono", false, "use normal font on tag/field")
	flag.StringVar(&cmds.FlagIgnore, "ignore", "", "disable list, i.e., --ignore nginx,redis,mem")
	flag.StringVar(&cmds.FlagExportIntegration, "export-integration", "", "export all integrations")
	flag.StringVar(&cmds.FlagManVersion, "man-version", datakit.Version, "specify manuals version")
	flag.StringVar(&cmds.FlagTODO, "TODO", "TODO", "set TODO")

	// install 3rd-party kit
	flag.StringVar(&cmds.FlagInstallExternal, "install", "", "install external tool/software")

	// managing service
	flag.BoolVar(&cmds.FlagStart, "start", false, "start datakit")
	flag.BoolVar(&cmds.FlagStop, "stop", false, "stop datakit")
	flag.BoolVar(&cmds.FlagRestart, "restart", false, "restart datakit")
	flag.BoolVar(&cmds.FlagAPIRestart, "api-restart", false, "restart datakit for api only")
	flag.BoolVar(&cmds.FlagStatus, "status", false, "show datakit service status")
	flag.BoolVar(&cmds.FlagUninstall, "uninstall", false, "uninstall datakit service(not delete DataKit files)")
	flag.BoolVar(&cmds.FlagReinstall, "reinstall", false, "re-install datakit service")

	// DQL
	flag.BoolVarP(&cmds.FlagDQL, "dql", "Q", false, "under DQL, query interactively")
	flag.BoolVar(&cmds.FlagJSON, "json", false, "under DQL, output in json format")
	flag.BoolVar(&cmds.FlagForce, "force", false, "Mandatory modification")
	flag.BoolVar(&cmds.FlagAutoJSON, "auto-json", false, "under DQL, pretty output string if it's JSON")
	flag.StringVar(&cmds.FlagRunDQL, "run-dql", "", "run single DQL")
	flag.StringVar(&cmds.FlagToken, "token", "", "query under specific token")
	flag.StringVar(&cmds.FlagCSV, "csv", "", "Specify the directory")

	// update online data
	flag.BoolVar(&cmds.FlagUpdateIPDB, "update-ip-db", false, "update ip db")
	flag.StringVarP(&cmds.FlagAddr, "addr", "A", "", "url path")
	flag.DurationVar(&cmds.FlagInterval, "interval", time.Second*5, "auxiliary option, time interval")

	// utils
	flag.StringVar(&cmds.FlagShowCloudInfo, "show-cloud-info", "", "show current host's cloud info(aliyun/tencent/aws)")
	flag.StringVar(&cmds.FlagIPInfo, "ipinfo", "", "show IP geo info")
	flag.BoolVar(&cmds.FlagWorkspaceInfo, "workspace-info", false, "show workspace info")

	if runtime.GOOS != datakit.OSWindows { // unsupported options under windows
		flag.BoolVarP(&cmds.FlagMonitor, "monitor", "M", false, "show monitor info of current datakit")
		flag.BoolVar(&cmds.FlagDocker, "docker", false, "run within docker")
	}

	flag.BoolVar(&cmds.FlagCheckConfig, "check-config", false, "check inputs configure and main configure")
	flag.StringVar(&cmds.FlagConfigDir, "config-dir", "", "check configures under specified path")
	flag.BoolVar(&cmds.FlagCheckSample, "check-sample", false, "check inputs configure samples")
	flag.BoolVar(&cmds.FlagVVV, "vvv", false, "more verbose info")
	flag.StringVar(&cmds.FlagCmdLogPath, "cmd-log", "/dev/null", "command line log path")
	flag.StringVar(&cmds.FlagDumpSamples, "dump-samples", "", "dump all inputs samples")

	flag.BoolVar(&config.DisableSelfInput, "disable-self-input", false, "disable self input")
	flag.BoolVar(&io.DisableDatawayList, "disable-dataway-list", false, "disable list available dataway")
	flag.BoolVar(&io.DisableLogFilter, "disable-logfilter", false, "disable logfilter")
	flag.BoolVar(&io.DisableHeartbeat, "disable-heartbeat", false, "disable heartbeat")

	flag.BoolVar(&cmds.FlagUploadLog, "upload-log", false, "upload log")
}

var (
	l = logger.DefaultSLogger("main")

	// injected during building: -X.
	InputsReleaseType = ""
	ReleaseVersion    = ""
)

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
		if err := flag.CommandLine.MarkHidden(f); err != nil {
			l.Warnf("CommandLine.MarkHidden: %s, ignored", err)
		}
	}

	flag.CommandLine.SortFlags = false
	flag.ErrHelp = errors.New("") // disable `pflag: help requested`

	if runtime.GOOS == datakit.OSWindows {
		cmds.FlagCmdLogPath = "nul" // under windows, nul is /dev/null
	}
}

func main() {
	datakit.Version = ReleaseVersion
	if ReleaseVersion != "" {
		datakit.Version = ReleaseVersion
	}

	setupFlags()
	flag.Parse()
	applyFlags()

	if err := datakit.SavePid(); err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	tryLoadConfig()

	tracer.Start()
	defer tracer.Stop()

	datakit.SetLog()

	if cmds.FlagDocker {
		// This may throw `Unix syslog delivery error` within docker, so we just
		// start the entry under docker.
		run()
	} else {
		go cgroup.Run()
		service.Entry = run
		if cmds.FlagWorkDir != "" { // debugging running, not start as service
			run()
		} else if err := service.StartService(); err != nil {
			l.Errorf("start service failed: %s", err.Error())

			return
		}
	}

	l.Info("datakit exited")
}

func applyFlags() {
	inputs.TODO = cmds.FlagTODO

	if cmds.FlagWorkDir != "" {
		datakit.SetWorkDir(cmds.FlagWorkDir)
	}

	datakit.EnableUncheckInputs = (InputsReleaseType == "all")

	if cmds.FlagDocker {
		datakit.Docker = true
	}

	cmds.ReleaseVersion = ReleaseVersion
	cmds.InputsReleaseType = InputsReleaseType

	cmds.RunCmds()
}

func run() {
	l.Info("datakit start...")
	if err := doRun(); err != nil {
		return
	}

	io.FeedEventLog(&io.Reporter{Message: "datakit start ok, ready for collecting metrics."})

	l.Info("datakit start ok. Wait signal or service stop...")

	// NOTE:
	// Actually, the datakit process been managed by system service, no matter on
	// windows/UNIX, datakit should exit via `service-stop' operation, so the signal
	// branch should not reached, but for daily debugging(ctrl-c), we kept the signal
	// exit option.
	signals := make(chan os.Signal, datakit.CommonChanCap)
	for {
		signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
		select {
		case sig := <-signals:
			l.Infof("get signal %v, wait & exit", sig)
			datakit.Quit()
			l.Info("datakit exit.")
			goto exit

		case <-service.StopCh:
			l.Infof("service stopping")
			datakit.Quit()
			l.Info("datakit exit.")
			goto exit
		}
	}
exit:
	time.Sleep(time.Second)
}

func tryLoadConfig() {
	config.MoveDeprecatedCfg()

	for {
		if err := config.LoadCfg(config.Cfg, datakit.MainConfPath); err != nil {
			l.Errorf("load config failed: %s", err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	l = logger.SLogger("main")

	l.Infof("datakit run ID: %s", cliutils.XID("dkrun_"))
}

func initPythonCore() error {
	// remove core dir
	if err := os.RemoveAll(datakit.PythonCoreDir); err != nil {
		return err
	}

	// generate new core dir
	if err := os.MkdirAll(datakit.PythonCoreDir, datakit.ConfPerm); err != nil {
		return err
	}

	for k, v := range pythond.PythonDCoreFiles {
		bFile := filepath.Join(datakit.PythonCoreDir, k)
		if err := ioutil.WriteFile(bFile, []byte(v), datakit.ConfPerm); err != nil {
			return err
		}
	}

	return nil
}

func doRun() error {
	if err := io.Start(); err != nil {
		return err
	}

	plworker.InitManager(-1)

	if config.Cfg.EnableElection {
		election.Start(config.Cfg.Namespace, config.Cfg.Hostname, config.Cfg.DataWay)
	}

	if err := initPythonCore(); err != nil {
		l.Errorf("initPythonCore failed: %v", err)
		return err
	}

	if config.GitHasEnabled() {
		if err := gitrepo.StartPull(); err != nil {
			l.Errorf("gitrepo.StartPull failed: %v", err)
			return err
		}
	} else {
		if err := inputs.RunInputs(false); err != nil {
			l.Error("error running inputs: %v", err)
			return err
		}
	}

	// NOTE: Should we wait all inputs ok, then start http server?
	dkhttp.Start(&dkhttp.Option{
		APIConfig:      config.Cfg.HTTPAPI,
		DCAConfig:      config.Cfg.DCAConfig,
		GinLog:         config.Cfg.Logging.GinLog,
		GinRotate:      config.Cfg.Logging.Rotate,
		GinReleaseMode: strings.ToLower(config.Cfg.Logging.Level) != "debug",

		DataWay: config.Cfg.DataWay,
		PProf:   config.Cfg.EnablePProf,
	})

	time.Sleep(time.Second) // wait http server ok

	return nil
}
