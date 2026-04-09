// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package core

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/checkutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/confd"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dnswatcher"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/election"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/gitrepo"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	plRemote "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/remote"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/usagetrace"
)

// doRun executes the main runtime logic.
func (c *Core) doRun() error {
	if c.cfg.PointPool.Enable {
		log.Info("point pool enabled with reserved capacity %d", c.cfg.PointPool.ReservedCapacity)
		datakit.SetupPointPool(c.cfg.PointPool.ReservedCapacity)
	}

	// Setup GC if debug environment variable is set
	if v := os.Getenv("DK_DEBUG_GC_DURATION"); v != "" {
		du, err := time.ParseDuration(v)
		if err != nil {
			log.Warnf("invalid ENV_GC_DURATION: %q, ignored", v)
		}
		if du < time.Second*10 {
			log.Infof("reset GC ticker from %s to 10s", du)
			du = time.Second * 10
		}

		go c.gc(du)
	}

	// Get CPU limits and adjust available CPUs
	cpuLimit := c.getCurrentCPULimits()
	log.Infof("get limited cpu cores: %f", cpuLimit)
	if cpuLimit > 1.0 {
		datakit.AvailableCPUs = int(cpuLimit)
	}

	// Start core services
	c.startDatawayWorkers()
	c.startIO()

	// Start NTP syncer if enabled
	if n := c.cfg.Dataway.NTP; n != nil && n.Enable {
		ntp.StartNTP(c.cfg.Dataway,
			n.Interval,
			uint64(n.SyncOnDiff/time.Second))
	}

	// Start DNS watcher
	checkutil.CheckConditionExit(func() bool {
		if err := dnswatcher.StartWatch(); err != nil {
			return false
		}

		return true
	})

	// Start election and pipeline remote if dataway is configured
	if c.cfg.Dataway != nil {
		c.startElection()

		if len(c.cfg.Dataway.URLs) == 1 {
			plRemote.StartPipelineRemote(c.cfg.Dataway.URLs, plRemote.DefaultPipelineRemote())
		} else {
			log.Warn("dataway empty or multi, not run pipeline remote")
		}
	} else {
		log.Warn("Ignore election or pipeline remote because dataway is not set")
	}

	// Start usage tracing
	c.startUsageTrace(cpuLimit)

	// Start configuration sources (confd, git, or direct inputs)
	if err := c.startConfigSources(); err != nil {
		return err
	}

	// Start HTTP server
	c.startHTTP()

	return nil
}

// startIO starts the IO subsystem.
func (c *Core) startIO() {
	ioCfg := c.cfg.IO
	opts := []dkio.IOOption{
		dkio.WithFeederOutputer(dkio.NewDatawayOutput(ioCfg.FeedChanSize)),
		dkio.WithDataway(c.cfg.Dataway),
		dkio.WithCompactAt(ioCfg.MaxCacheCount),
		dkio.WithFilters(ioCfg.Filters),
		dkio.WithCompactWorkers(ioCfg.CompactWorkers),
		dkio.WithRecorder(c.cfg.Recorder),
		dkio.WithRemoteJob(c.cfg.RemoteJob, c.cfg.Dataway),
		dkio.WithAvailableCPUs(datakit.AvailableCPUs),
		dkio.WithTimeCorrect(ioCfg.AutoTimestampCorrection),
	}

	dkio.Start(opts...)
}

// startDatawayWorkers starts dataway workers.
func (c *Core) startDatawayWorkers() {
	dw := c.cfg.Dataway

	// setup extra options on @dw
	if dw.WAL.Workers == 0 {
		n := datakit.AvailableCPUs * 8 // we need more workers on WAL upload
		log.Infof("set %d flush WAL workers", n)
		dataway.WithWALWorkers(n)(dw)
	}

	for {
		if err := dw.StartFlushWorkers(); err != nil {
			log.Errorf("StartFlushWorkers failed: %s, retrying...", err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}
}

// gc runs periodic garbage collection.
func (c *Core) gc(du time.Duration) {
	tick := time.NewTicker(du)
	defer tick.Stop()

	log.Infof("setup GC on interval %s", du)
	for {
		select {
		case <-datakit.Exit.Wait():
			return
		case <-tick.C:
			runtime.GC()
		}
	}
}

// startElection starts the election process.
func (c *Core) startElection() {
	electionsOpts := []election.ElectionOption{
		election.WithElectionEnabled(c.cfg.Election.Enable),
		election.WithElectionWhitelist(c.cfg.Election.NodeWhitelist),
		election.WithID(datakit.DKHost),
		election.WithNamespace(c.cfg.Election.Namespace),
	}

	electionsOpts = append(electionsOpts, election.WithDatawayPuller(c.cfg.Dataway))
	election.Start(electionsOpts...)
}

// startUsageTrace starts usage tracing.
func (c *Core) startUsageTrace(cpuLimit float64) {
	usagetrace.Start(usagetrace.WithRefresher(c.cfg.Dataway),
		usagetrace.WithServerListens(c.cfg.HTTPAPI.Listen),
		usagetrace.WithCPULimits(cpuLimit),
		usagetrace.WithDatakitHostname(datakit.DKHost),
		usagetrace.WithDatakitRuntimeID(c.runtimeID),
		usagetrace.WithDatakitVersion(datakit.Version),
		usagetrace.WithDatakitStartTime(metrics.Uptime.Unix()),
		usagetrace.WithRunInContainer(datakit.Docker),
		usagetrace.WithReservedInputs("rum", "kafkamq", "prom_remote_write", "beats_output"),
		usagetrace.WithExitChan(datakit.Exit.Wait()),
		usagetrace.WithRefreshDuration(time.Minute*5),

		usagetrace.WithMainIP(func() string {
			if ip, err := datakit.LocalIP(); err != nil {
				return fmt.Sprintf("get datakit local IP failed: %s", err.Error())
			} else {
				return ip
			}
		}()),

		usagetrace.WithWorkspaceToken(func() string {
			arr := c.cfg.Dataway.GetTokens()
			if len(arr) > 0 {
				return arr[0] // only use the 1st token configured.
			}
			return "datakit's workspace token not set"
		}()),

		usagetrace.WithDatakitPodname(func() string {
			if v := datakit.GetEnv("POD_NAME"); v != "" {
				return v
			} else {
				return ""
			}
		}()),
	)
}

// startConfigSources starts configuration sources.
func (c *Core) startConfigSources() error {
	if c.confdEnabled() {
		if err := confd.Run(c.cfg.Confds); err != nil {
			return err
		}
	} else {
		if config.GitHasEnabled() {
			if err := gitrepo.StartPull(); err != nil {
				log.Errorf("gitrepo.StartPull failed: %v", err)
				return err
			}
		} else {
			if err := inputs.RunInputs(inputs.AllInputsInfo); err != nil {
				log.Error("error running inputs: %v", err)
				return err
			}
		}
	}
	return nil
}

// confdEnabled reports whether any confd backend is enabled in current config.
func (c *Core) confdEnabled() bool {
	if c.cfg == nil {
		return false
	}

	for _, confd := range c.cfg.Confds {
		if confd.Enable {
			return true
		}
	}

	return false
}
