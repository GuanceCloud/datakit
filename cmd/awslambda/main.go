// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/checkutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	plRemote "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/remote"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/all"
)

var (
	l = logger.DefaultSLogger("main")

	// injected during building: -X.

	ReleaseVersion = ""

	runtimeID string
)

var signals = make(chan os.Signal, 1)

func init() { //nolint:gochecknoinits
	signal.Notify(signals, syscall.SIGTERM)
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano()) // rand seed global

	datakit.Version = ReleaseVersion
	if ReleaseVersion != "" {
		datakit.Version = ReleaseVersion
	}

	if v := datakit.GetEnv("DK_DEBUG_WORKDIR"); v != "" {
		datakit.SetWorkDir(v)
	} else {
		// only /tmp dir are write-able under AWS lambda.
		v = "/tmp/datakit"
		datakit.SetWorkDir(v)
	}

	loadLambdaDefaultConf()
	tryLoadConfig()
	datakit.SetLog()

	run()

	l.Info("datakit exited")
}

func loadLambdaDefaultConf() {
	datakit.Docker = true // set docker mode under lambda

	// All logs default write to stdout. Other log related settings
	// should set via ENV_LOG_XXX envs.
	config.Cfg.Logging.GinLog = "stdout"
	config.Cfg.Logging.Log = "stdout"

	// Default listen to non-localhost: we may accept trace related API request from user lambda apps.
	config.Cfg.HTTPAPI.Listen = "0.0.0.0:9529"
	config.Cfg.DefaultEnabledInputs = []string{"awslambda", "ddtrace", "opentelemetry", "statsd"}
	config.Cfg.Dataway.MaxRawBodySize = dataway.DefaultMaxRawBodySize
	config.Cfg.IO.CompactWorkers = 1 // limit workers to save CPU cost
}

func run() {
	l.Info("datakit start...")

	switch config.Cfg.RunMode {
	case datakit.ModeNormal:
		if err := doRun(); err != nil {
			return
		}
	default:
		return
	}

	l.Info("datakit start ok. Wait signal or service stop...")
	sig := <-signals
	l.Infof("get signal %v, wait & exit", sig)
	quit()

	l.Info("datakit exit.")
	time.Sleep(time.Second)
}

func quit() {
	datakit.GlobalExitTime = time.Now()
	datakit.Exit.Close()
	datakit.WG.Wait()
	datakit.GWait()
}

func tryLoadConfig() {
	l.Infof("load config from %s...", datakit.MainConfPath)
	checkutil.CheckConditionExit(func() bool {
		if err := config.LoadCfg(config.Cfg, datakit.MainConfPath); err != nil {
			l.Errorf("load config failed: %s", err)
			return false
		}

		return true
	})

	l = logger.SLogger("main")

	runtimeID = cliutils.XID("dkrun_")

	l.Infof("datakit run ID: %s, version: %s", runtimeID, datakit.Version)
}

func startIO() {
	c := config.Cfg.IO
	opts := []dkio.IOOption{
		dkio.WithFeederOutputer(dkio.NewAwsLambdaOutput()),
		dkio.WithDataway(config.Cfg.Dataway),
		dkio.WithCompactAt(c.MaxCacheCount),
		dkio.WithFilters(c.Filters),
		dkio.WithCompactWorkers(c.CompactWorkers),
		dkio.WithRecorder(config.Cfg.Recorder),
		dkio.WithCompactInterval(c.CompactInterval),
		dkio.WithCompactor(false),
	}

	dkio.Start(opts...)
}

func doRun() error {
	if config.Cfg.PointPool.Enable {
		l.Info("point pool enabled with reserved capacity %d", config.Cfg.PointPool.ReservedCapacity)
		datakit.SetupPointPool(config.Cfg.PointPool.ReservedCapacity)
	}

	startIO()

	if config.Cfg.Dataway != nil {
		if len(config.Cfg.Dataway.URLs) == 1 {
			// https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/524
			plRemote.StartPipelineRemote(config.Cfg.Dataway.URLs, plRemote.DefaultPipelineRemote())
		} else {
			l.Warn("dataway empty or multi, not run pipeline remote")
		}
	} else {
		l.Warn("Ignore election or pipeline remote because dataway is not set")
	}
	if err := inputs.RunInputs(); err != nil {
		l.Error("error running inputs: %v", err)
		return err
	}

	// NOTE: Should we wait all inputs ok, then start http server?
	startHTTP()

	return nil
}

func startHTTP() {
	httpapi.Start(
		httpapi.WithAPIConfig(config.Cfg.HTTPAPI),
		httpapi.WithDCAConfig(config.Cfg.DCAConfig),
		httpapi.WithGinLog(config.Cfg.Logging.GinLog),
		httpapi.WithGinRotateMB(config.Cfg.Logging.Rotate),
		httpapi.WithGinReleaseMode(strings.ToLower(config.Cfg.Logging.Level) != "debug"),
		httpapi.WithDataway(config.Cfg.Dataway),
		httpapi.WithPProf(config.Cfg.EnablePProf),
		httpapi.WithPProfListen(config.Cfg.PProfListen),
	)
}
