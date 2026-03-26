// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package core

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/checkutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/resourcelimit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
)

var log = logger.DefaultSLogger("core")

// Core manages the main DataKit application.
type Core struct {
	runtimeID string
	cfg       *config.Config
}

// New creates a new Core instance.
func New() *Core {
	return &Core{}
}

// Initialize sets up the core application.
func (c *Core) Initialize(releaseVersion, inputsReleaseType, lite, eLinker string) error {
	rand.Seed(time.Now().UTC().UnixNano()) // rand seed global

	if releaseVersion != "" {
		datakit.Version = releaseVersion
	}

	if v, err := strconv.ParseBool(lite); err == nil {
		datakit.Lite = v
	}
	if v, err := strconv.ParseBool(eLinker); err == nil {
		datakit.ELinker = v
	}

	cmds.ReleaseVersion = releaseVersion
	cmds.InputsReleaseType = inputsReleaseType
	cmds.Lite = datakit.Lite
	cmds.ELinker = datakit.ELinker

	var workdir string
	// Debugging running, not start as service
	if v := datakit.GetEnv("DK_DEBUG_WORKDIR"); v != "" {
		datakit.SetWorkDir(v)
		workdir = v
	}

	cmds.ParseFlags()
	c.applyFlags()

	if err := c.tryLoadConfig(); err != nil {
		return err
	}

	datakit.SetLog()
	service.SetLog()

	if datakit.Docker {
		// This may throw `Unix syslog delivery error` within docker, so we just
		// start the entry under docker.
		if err := c.run(); err != nil {
			return err
		}

		// NOTE: do not set PID on docker running
		// Under docker, people can set restart on crash, so do not set and check pid:
		//   docker run -d --restart=on-failure ...
	} else {
		if err := datakit.SavePid(); err != nil { // save PID on host-running
			return fmt.Errorf("save pid failed: %w", err)
		}

		// Auto enable resource limit under host running(debug mode and service mode)
		if c.cfg.ResourceLimitOptions != nil {
			resourcelimit.Run(c.cfg.ResourceLimitOptions, c.cfg.DatakitUser)
		}

		if workdir != "" {
			if err := c.run(); err != nil {
				return err
			}
		} else { // running as system service
			if err := service.StartService(c.serviceEntry); err != nil {
				return fmt.Errorf("service.StartService: %w", err)
			}
		}
	}

	return nil
}

// applyFlags applies command line flags.
func (c *Core) applyFlags() {
	if *cmds.FlagRunInContainer {
		datakit.Docker = true
	}
}

// serviceEntry is the entry point for service mode.
func (c *Core) serviceEntry() {
	go func() {
		if err := c.run(); err != nil {
			log.Errorf("run failed: %v", err)
		}
	}()
}

// run executes the main runtime logic.
func (c *Core) run() error {
	log.Info("datakit start...")

	switch c.cfg.RunMode {
	case datakit.ModeNormal:
		if err := c.doRun(); err != nil {
			return err
		}

	case datakit.ModeDev:
		c.startHTTP()

	default:
		return nil
	}

	maxRunTick := time.NewTicker(time.Duration(int64(math.MaxInt64)))
	if v := datakit.GetEnv("DK_DEBUG_MAX_RUN_DURATION"); v != "" {
		du, err := time.ParseDuration(v)
		if err == nil {
			log.Infof("set max-run-duration to %s", du)
			maxRunTick.Reset(du)
		}
	}
	defer maxRunTick.Stop()

	log.Info("datakit start ok. Wait signal or service stop...")

	// NOTE:
	// Actually, the datakit process been managed by system service, no matter on
	// windows/UNIX, datakit should exit via `service-stop' operation, so the signal
	// branch should not reached, but for daily debugging(ctrl-c), we kept the signal
	// exit option.
	signals := make(chan os.Signal, datakit.CommonChanCap)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case sig := <-signals:
			log.Infof("get signal %v, wait & exit", sig)
			c.quit()
			log.Info("datakit exit.")
			goto exit

		case <-service.Wait():
			log.Infof("service stopping")
			c.quit()
			log.Info("datakit exit.")
			goto exit
		case <-maxRunTick.C:
			log.Infof("reach max run duration")
			c.quit()
			log.Info("datakit exit.")
			goto exit
		}
	}

exit:
	time.Sleep(time.Second)
	return nil
}

// quit performs graceful shutdown.
func (c *Core) quit() {
	datakit.GlobalExitTime = time.Now()
	if err := os.Remove(datakit.PidFile); err != nil {
		log.Warnf("remove PID file(%s) failed: %s, ignored", datakit.PidFile, err)
	}

	datakit.Exit.Close()
	datakit.WG.Wait()
	goroutine.GWait()
	service.Stop()
}

// tryLoadConfig loads and validates configuration.
func (c *Core) tryLoadConfig() error {
	config.MoveDeprecatedCfg()

	log.Infof("load config from %s...", datakit.MainConfPath)
	checkutil.CheckConditionExit(func() bool {
		if err := config.LoadCfg(config.Cfg, datakit.MainConfPath, datakit.Docker); err != nil {
			log.Errorf("load config failed: %s", err)
			return false
		}

		return true
	})

	log = logger.SLogger("main")

	c.runtimeID = cliutils.XID("dkrun_")
	datakit.RuntimeID = c.runtimeID
	c.cfg = config.Cfg

	log.Infof("datakit run ID: %s, version: %s", c.runtimeID, datakit.Version)
	return nil
}

// startHTTP starts the HTTP API server.
func (c *Core) startHTTP() {
	httpapi.Start(
		httpapi.WithAPIConfig(c.cfg.HTTPAPI),
		httpapi.WithDCAConfig(c.cfg.DCAConfig),
		httpapi.WithGinLog(c.cfg.Logging.GinLog),
		httpapi.WithGinRotateMB(c.cfg.Logging.Rotate),
		httpapi.WithGinReleaseMode(strings.ToLower(c.cfg.Logging.Level) != "debug"),
		httpapi.WithDataway(c.cfg.Dataway),
		httpapi.WithPProf(c.cfg.EnablePProf),
		httpapi.WithPProfListen(c.cfg.PProfListen),
	)

	time.Sleep(time.Second) // wait http server ok
}
