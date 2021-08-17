package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	flag "github.com/spf13/pflag"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cgroup"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/election"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/all"
)

func init() {
	flag.BoolVarP(&cmds.FlagVersion, "version", "V", false, `show version info`)
	flag.BoolVar(&cmds.FlagCheckUpdate, "check-update", false, "check if new verison available")
	flag.BoolVar(&cmds.FlagAcceptRCVersion, "accept-rc-version", false, "during update, accept RC version if available")
	flag.BoolVar(&cmds.FlagShowTestingVersions, "show-testing-version", false, "show testing versions on -version flag")
	flag.StringVar(&cmds.FlagUpdateLogFile, "update-log", "", "update history log file")

	// debug grok
	flag.StringVar(&cmds.FlagPipeline, "pl", "", "pipeline script to test(name only, do not use file path)")
	flag.BoolVar(&cmds.FlagGrokq, "grokq", false, "query groks interactively")
	flag.StringVar(&cmds.FlagText, "txt", "", "text string for the pipeline or grok(json or raw text)")

	flag.StringVar(&cmds.FlagProm, "prom-conf", "", "prom config file to test")

	// manuals related
	flag.BoolVar(&cmds.FlagMan, "man", false, "read manuals of inputs")
	flag.StringVar(&cmds.FlagExportMan, "export-manuals", "", "export all inputs and related manuals to specified path")
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
	flag.BoolVar(&cmds.FlagReload, "reload", false, "reload datakit")
	flag.BoolVar(&cmds.FlagStatus, "status", false, "show datakit service status")
	flag.BoolVar(&cmds.FlagUninstall, "uninstall", false, "uninstall datakit service(not delete DataKit files)")
	flag.BoolVar(&cmds.FlagReinstall, "reinstall", false, "re-install datakit service")

	flag.StringVarP(&cmds.FlagDatakitHost, "datakit-host", "H", "localhost:9529", "datakit HTTP host")

	// DQL
	flag.BoolVarP(&cmds.FlagDQL, "dql", "Q", false, "query DQL interactively")
	flag.StringVar(&cmds.FlagRunDQL, "run-dql", "", "run single DQL")

	// update online data
	flag.BoolVar(&cmds.FlagUpdateIPDB, "update-ip-db", false, "update ip db")
	flag.StringVarP(&cmds.FlagAddr, "addr", "A", "", "url path")
	flag.DurationVar(&cmds.FlagInterval, "interval", time.Second*3, "auxiliary option, time interval")

	// utils
	flag.StringVar(&cmds.FlagShowCloudInfo, "show-cloud-info", "", "show current host's cloud info              ( aliyun/tencent/aws)")
	flag.StringVar(&cmds.FlagIPInfo, "ipinfo", "", "show IP geo info")
	flag.BoolVarP(&cmds.FlagMonitor, "monitor", "M", false, "show monitor info of current datakit")
	flag.BoolVar(&cmds.FlagCheckConfig, "check-config", false, "check inputs configure and main configure")
	flag.BoolVar(&cmds.FlagVVV, "vvv", false, "more verbose info")
	flag.StringVar(&cmds.FlagCmdLogPath, "cmd-log", "/dev/null", "command line log path")
	flag.StringVar(&cmds.FlagDumpSamples, "dump-samples", "", "dump all inputs samples")
	flag.BoolVar(&cmds.FlagDocker, "docker", false, "run within docker")
}

var (

	// deprecated
	flagCmdDeprecated = flag.Bool("cmd", false, "run datakit under command line mode")

	////////////////////////////////////////////////////////////
	// Commands
	////////////////////////////////////////////////////////////
)

var (
	l = logger.DefaultSLogger("main")

	// injected during building: -X
	ReleaseType    = ""
	ReleaseVersion = ""
)

func setupFlags() {
	// deprecated
	flag.CommandLine.MarkDeprecated("cmd", "--cmd deprecated and not required")

	// hidden flags
	for _, f := range []string{
		"TODO",
		"check-update",
		"man-version",
		"export-integration",
		"addr",
		"show-testing-version",
		"update-log",
		"k8s-deploy",
		"interactive",
		"dump-samples",
	} {
		flag.CommandLine.MarkHidden(f)
	}

	if runtime.GOOS == "windows" {
		flag.CommandLine.MarkHidden("reload")
	}

	flag.CommandLine.SortFlags = false
	flag.ErrHelp = errors.New("") // disable `pflag: help requested`

	if runtime.GOOS == "windows" {
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

	datakit.SetLog()

	// This may throw `Unix syslog delivery error` within docker, so we just
	// start the entry under docker.
	if cmds.FlagDocker {
		run()
	} else {
		go cgroup.Run()
		service.Entry = run
		if err := service.StartService(); err != nil {
			l.Errorf("start service failed: %s", err.Error())
			return
		}
	}

	l.Info("datakit exited")
}

func applyFlags() {

	inputs.TODO = cmds.FlagTODO

	datakit.EnableUncheckInputs = (ReleaseType == "all")

	if cmds.FlagDocker {
		datakit.Docker = true
	}

	cmds.ReleaseVersion = ReleaseVersion
	cmds.ReleaseType = ReleaseType

	cmds.RunCmds()
}

func run() {

	l.Info("datakit start...")
	if err := doRun(); err != nil {
		return
	}

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
			if sig == syscall.SIGHUP {
				l.Info("under signal SIGHUP, reloading...")
				cmds.SetLog()
				cmds.Reload()
			} else {
				l.Infof("get signal %v, wait & exit", sig)
				datakit.Quit()
				l.Info("datakit exit.")
				goto exit
			}

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
}

func doRun() error {

	for _, x := range os.Environ() {
		l.Infof("get env %s", x)
	}

	io.Start()

	if config.Cfg.EnableElection {
		election.Start(config.Cfg.Namespace, config.Cfg.Hostname, config.Cfg.DataWay)
	}

	if err := inputs.RunInputs(); err != nil {
		l.Error("error running inputs: %v", err)
		return err
	}

	// FIXME: wait all inputs ok, then start http server

	dkhttp.Start(&dkhttp.Option{
		APIConfig:      config.Cfg.HTTPAPI,
		GinLog:         config.Cfg.Logging.GinLog,
		GinRotate:      config.Cfg.Logging.Rotate,
		GinReleaseMode: strings.ToLower(config.Cfg.Logging.Level) != "debug",

		DataWay: config.Cfg.DataWay,
		PProf:   config.Cfg.EnablePProf,
	})

	time.Sleep(time.Second) // wait http server ok

	return nil
}
