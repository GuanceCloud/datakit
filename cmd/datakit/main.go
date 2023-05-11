// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/gitrepo"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cgroup"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/checkutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dnswatcher"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/election"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	plRemote "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/remote"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/all"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/tracer"
)

var (
	l = logger.DefaultSLogger("main")

	// injected during building: -X.
	InputsReleaseType = ""
	ReleaseVersion    = ""
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano()) // rand seed global

	datakit.Version = ReleaseVersion
	if ReleaseVersion != "" {
		datakit.Version = ReleaseVersion
	}

	cmds.ReleaseVersion = ReleaseVersion
	cmds.InputsReleaseType = InputsReleaseType

	var workdir string
	// Debugging running, not start as service
	if v := datakit.GetEnv("DK_DEBUG_WORKDIR"); v != "" {
		datakit.SetWorkDir(v)
		workdir = v
	}

	cmds.ParseFlags()
	applyFlags()

	if err := datakit.SavePid(); err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	tryLoadConfig()

	// start up global tracer
	tracer.Start()
	defer tracer.Stop()

	datakit.SetLog()

	if datakit.Docker {
		// This may throw `Unix syslog delivery error` within docker, so we just
		// start the entry under docker.
		run()
	} else {
		// Auto enable cgroup limit under host running(debug mode and service mode)
		cgroup.Run(config.Cfg.Cgroup)

		if workdir != "" {
			run()
		} else { // running as System service
			service.Entry = serviceEntry
			if err := service.StartService(); err != nil {
				l.Errorf("start service failed: %s", err.Error())
				return
			}
		}
	}

	l.Info("datakit exited")
}

func applyFlags() {
	if *cmds.FlagRunInContainer {
		datakit.Docker = true
	}
}

func serviceEntry() {
	go run()
}

func run() {
	l.Info("datakit start...")

	switch config.Cfg.RunMode {
	case datakit.ModeNormal:
		if err := doRun(); err != nil {
			return
		}

	case datakit.ModeDev:
		startHTTP()

	default:
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

	l.Infof("load config from %s...", datakit.MainConfPath)
	checkutil.CheckConditionExit(func() bool {
		if err := config.LoadCfg(config.Cfg, datakit.MainConfPath); err != nil {
			l.Errorf("load config failed: %s", err)
			return false
		}

		return true
	})

	l = logger.SLogger("main")

	l.Infof("datakit run ID: %s, version: %s", cliutils.XID("dkrun_"), datakit.Version)
}

func startIO() {
	c := config.Cfg.IO
	opts := []dkio.IOOption{
		dkio.WithDataway(config.Cfg.Dataway),
		dkio.WithFeedCap(c.FeedChanSize),
		dkio.WithMaxCacheCount(c.MaxCacheCount),
		dkio.WithOutputFile(c.OutputFile),
		dkio.WithOutputFileOnInputs(c.OutputFileInputs),
		dkio.WithDiskCache(c.EnableCache),
		dkio.WithDiskCacheSize(c.CacheSizeGB),
		dkio.WithFilters(c.Filters),
		dkio.WithCacheAll(c.CacheAll),
		dkio.WithFlushWorkers(c.FlushWorkers),
	}

	du, err := time.ParseDuration(c.FlushInterval)
	if err != nil {
	} else {
		opts = append(opts, dkio.WithFlushInterval(du))
	}

	du, err = time.ParseDuration(c.CacheCleanInterval)
	if err != nil {
		l.Warnf("parse CacheCleanInterval failed: %s, use default 5s", err)
	} else {
		opts = append(opts, dkio.WithDiskCacheCleanInterval(du))
	}

	dkio.Start(opts...)
}

func doRun() error {
	startIO()

	checkutil.CheckConditionExit(func() bool {
		if err := dnswatcher.StartWatch(); err != nil {
			return false
		}

		return true
	})

	if config.Cfg.Dataway != nil {
		electionsOpts := []election.ElectionOption{
			election.WithElectionEnabled(config.Cfg.Election.Enable),
			election.WithID(config.Cfg.Hostname),
			election.WithNamespace(config.Cfg.Election.Namespace),
		}

		if err := config.Cfg.Operator.Ping(); err == nil {
			l.Infof("datakit-operator connection successed.")
			electionsOpts = append(electionsOpts, election.WithOperatorPuller(config.Cfg.Operator))
		} else {
			l.Infof("datakit-operator connection refused, reason: %s", err)
			electionsOpts = append(electionsOpts, election.WithDatawayPuller(config.Cfg.Dataway))
		}

		election.Start(electionsOpts...)

		if len(config.Cfg.Dataway.URLs) == 1 {
			// https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/524
			plRemote.StartPipelineRemote(config.Cfg.Dataway.URLs)
		} else {
			dkio.FeedLastError(datakit.DatakitInputName, "dataway empty or multi, not run pipeline remote")
		}
	} else {
		l.Warn("Ignore election or pipeline remote because dataway is not set")
	}

	if config.IsUseConfd() {
		// First need RunInputs. lots of start in this func
		// must befor StartConfd()
		if err := inputs.RunInputs(); err != nil {
			l.Error("error running inputs: %v", err)
			return err
		}

		// if use config source from confd, like etcd zookeeper concul tredis ...
		if err := config.StartConfd(); err != nil {
			l.Errorf("config.StartConfd failed: %v", err)
			return err
		}
	} else {
		if config.GitHasEnabled() {
			if err := gitrepo.StartPull(); err != nil {
				l.Errorf("gitrepo.StartPull failed: %v", err)
				return err
			}
		} else {
			if err := inputs.RunInputs(); err != nil {
				l.Error("error running inputs: %v", err)
				return err
			}
		}
	}

	// NOTE: Should we wait all inputs ok, then start http server?
	startHTTP()

	return nil
}

func startHTTP() {
	dkhttp.Start(&dkhttp.Option{
		APIConfig:      config.Cfg.HTTPAPI,
		DCAConfig:      config.Cfg.DCAConfig,
		Log:            config.Cfg.Logging.Log,
		GinLog:         config.Cfg.Logging.GinLog,
		GinRotate:      config.Cfg.Logging.Rotate,
		GinReleaseMode: strings.ToLower(config.Cfg.Logging.Level) != "debug",

		DataWay:     config.Cfg.Dataway,
		PProf:       config.Cfg.EnablePProf,
		PProfListen: config.Cfg.PProfListen,
	})

	time.Sleep(time.Second) // wait http server ok
}
