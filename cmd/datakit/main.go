package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

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

var (
	l = logger.DefaultSLogger("main")

	// injected during building: -X.
	InputsReleaseType = ""
	ReleaseVersion    = ""
)

func main() {
	datakit.Version = ReleaseVersion
	if ReleaseVersion != "" {
		datakit.Version = ReleaseVersion
	}

	cmds.ParseFlags()
	applyFlags()

	if err := datakit.SavePid(); err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	tryLoadConfig()

	tracer.Start()
	defer tracer.Stop()

	datakit.SetLog()

	if datakit.Docker {
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

	if cmds.FlagDocker /* Deprecated */ || *cmds.FlagRunInContainer {
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
