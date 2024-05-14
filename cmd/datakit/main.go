// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/checkutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/confd"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dnswatcher"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/election"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/gitrepo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	plRemote "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/remote"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/all"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/resourcelimit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/usagetrace"
)

var (
	l = logger.DefaultSLogger("main")

	// injected during building: -X.
	InputsReleaseType = ""
	ReleaseVersion    = ""
	Lite              = "false"

	runtimeID string
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano()) // rand seed global

	datakit.Version = ReleaseVersion
	if ReleaseVersion != "" {
		datakit.Version = ReleaseVersion
	}

	if v, err := strconv.ParseBool(Lite); err == nil {
		datakit.Lite = v
	}

	cmds.ReleaseVersion = ReleaseVersion
	cmds.InputsReleaseType = InputsReleaseType
	cmds.Lite = datakit.Lite

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

	datakit.SetLog()

	if datakit.Docker {
		// This may throw `Unix syslog delivery error` within docker, so we just
		// start the entry under docker.
		run()
	} else {
		// Auto enable resource limit under host running(debug mode and service mode)
		if config.Cfg.ResourceLimitOptions != nil {
			resourcelimit.Run(config.Cfg.ResourceLimitOptions, config.Cfg.DatakitUser)
		}

		if workdir != "" {
			run()
		} else { // running as system service
			if err := service.StartService(serviceEntry); err != nil {
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

	maxRunTick := time.NewTicker(time.Duration(int64(math.MaxInt64)))
	if v := datakit.GetEnv("DK_DEBUG_MAX_RUN_DURATION"); v != "" {
		du, err := time.ParseDuration(v)
		if err == nil {
			l.Infof("set max-run-duration to %s", du)
			maxRunTick.Reset(du)
		}
	}
	defer maxRunTick.Stop()

	l.Info("datakit start ok. Wait signal or service stop...")

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
			l.Infof("get signal %v, wait & exit", sig)
			quit()
			l.Info("datakit exit.")
			goto exit

		case <-service.Wait():
			l.Infof("service stopping")
			quit()
			l.Info("datakit exit.")
			goto exit
		case <-maxRunTick.C:
			l.Infof("reach max run duration")
			quit()
			goto exit
		}
	}

exit:
	time.Sleep(time.Second)
}

func quit() {
	datakit.GlobalExitTime = time.Now()
	if err := os.Remove(datakit.PidFile); err != nil {
		l.Warnf("remove PID file(%s) failed: %s, ignored", datakit.PidFile, err)
	}

	datakit.Exit.Close()
	datakit.WG.Wait()
	datakit.GWait()
	service.Stop()
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

	runtimeID = cliutils.XID("dkrun_")

	l.Infof("datakit run ID: %s, version: %s", runtimeID, datakit.Version)
}

func startIO() {
	c := config.Cfg.IO
	opts := []dkio.IOOption{
		dkio.WithFeederOutputer(dkio.NewDatawayOutput(c.FeedChanSize)),
		dkio.WithDataway(config.Cfg.Dataway),
		dkio.WithMaxCacheCount(c.MaxCacheCount),
		dkio.WithDiskCache(c.EnableCache),
		dkio.WithDiskCacheSize(c.CacheSizeGB),
		dkio.WithFilters(c.Filters),
		dkio.WithCacheAll(c.CacheAll),
		dkio.WithFlushWorkers(c.FlushWorkers),
		dkio.WithRecorder(config.Cfg.Recorder),
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

func gc(du time.Duration) {
	tick := time.NewTicker(du)
	defer tick.Stop()

	l.Infof("setup GC on interval %s", du)
	for {
		select {
		case <-datakit.Exit.Wait():
			return
		case <-tick.C:
			runtime.GC()
		}
	}
}

func doRun() error {
	if config.Cfg.PointPool.Enable {
		l.Info("point pool enabled with reserved capacity %d", config.Cfg.PointPool.ReservedCapacity)
		datakit.SetupPointPool(config.Cfg.PointPool.ReservedCapacity)
	}

	if v := os.Getenv("DK_DEBUG_GC_DURATION"); v != "" {
		du, err := time.ParseDuration(v)
		if err != nil {
			l.Warnf("invalid ENV_GC_DURATION: %q, ignored", v)
		}
		if du < time.Second*10 {
			l.Infof("reset GC ticker from %s to 10s", du)
			du = time.Second * 10
		}

		go gc(du)
	}

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
			l.Warn("dataway empty or multi, not run pipeline remote")
		}
	} else {
		l.Warn("Ignore election or pipeline remote because dataway is not set")
	}

	// start CPU-core-based datakit running instance counting.
	usagetrace.Start(usagetrace.WithRefresher(config.Cfg.Dataway),
		usagetrace.WithServerListens(config.Cfg.HTTPAPI.Listen),
		usagetrace.WithCPULimits(getCurrentCPULimits()),
		usagetrace.WithDatakitHostname(config.Cfg.Hostname),
		usagetrace.WithDatakitRuntimeID(runtimeID),
		usagetrace.WithDatakitVersion(ReleaseVersion),
		usagetrace.WithDatakitStartTime(metrics.Uptime.Unix()),
		usagetrace.WithRunInContainer(datakit.Docker),
		usagetrace.WithReservedInputs("rum", "kafkamq", "prom_remote_write", "beats_output"),
		usagetrace.WithExitChan(datakit.Exit.Wait()),
		usagetrace.WithRefreshDuration(time.Minute*5),

		usagetrace.WithUpgraderServer(func() string {
			if config.Cfg.DKUpgrader != nil {
				return fmt.Sprintf("%s:%d", config.Cfg.DKUpgrader.Host, config.Cfg.DKUpgrader.Port)
			} else {
				return ""
			}
		}()),

		usagetrace.WithMainIP(func() string {
			if ip, err := datakit.LocalIP(); err != nil {
				return fmt.Sprintf("get datakit local IP failed: %s", err.Error())
			} else {
				return ip
			}
		}()),

		usagetrace.WithWorkspaceToken(func() string {
			arr := config.Cfg.Dataway.GetTokens()
			if len(arr) > 0 {
				return arr[0] // only use the 1st token configured.
			}
			return "datakit's workspace token not set"
		}()),

		usagetrace.WithDCAAPIServer(func() string {
			if !config.Cfg.DCAConfig.Enable {
				return ""
			} else {
				return config.Cfg.DCAConfig.Listen
			}
		}()),

		usagetrace.WithDatakitPodname(func() string {
			if v := datakit.GetEnv("POD_NAME"); v != "" {
				return v
			} else {
				return ""
			}
		}()),
	)

	if config.ConfdEnabled() {
		if err := confd.Run(config.Cfg.Confds); err != nil {
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

	time.Sleep(time.Second) // wait http server ok
}

func getCurrentCPULimits() float64 {
	if datakit.Docker {
		if limit, err := getCurrentCPUMaxFromCgroupv2(); err != nil {
			l.Warnf("failed to get cpu limit from cgroupv2, %s", err)
		} else {
			l.Infof("used cpu limit %v from cgroupv2", limit)
			return limit
		}

		if limit, err := getCurrentCPUMaxFromCgroupv1(); err != nil {
			l.Warnf("failed to get cpu limit from cgroupv1, %s", err)
		} else {
			l.Infof("used cpu limit %v from cgroupv1", limit)
			return limit
		}

		l.Warn("used default limit 1.0")
		return 1.0
	} else {
		if config.Cfg.ResourceLimitOptions.Enable {
			if config.Cfg.ResourceLimitOptions.CPUMax > 100.0 {
				return float64(runtime.NumCPU())
			} else {
				return (config.Cfg.ResourceLimitOptions.CPUMax / 100.0 * float64(runtime.NumCPU()))
			}
		} else {
			return float64(runtime.NumCPU()) // if no limit, set it to full-CPU cores
		}
	}
}

// Reference: https://docs.kernel.org/admin-guide/cgroup-v2.html#cpu-interface-files
func getCurrentCPUMaxFromCgroupv2() (float64, error) {
	const cpuMax = "/sys/fs/cgroup/cpu.max"

	data, err := os.ReadFile(cpuMax)
	if err != nil {
		return 0, err
	}

	content := strings.TrimSuffix(string(data), "\n")
	array := strings.Split(content, " ")
	if len(array) != 2 {
		return 0, fmt.Errorf("invalid cgroupv2 file")
	}

	return parseCurrentCPUMax(array[0], array[1])
}

// Reference: https://docs.kernel.org/scheduler/sched-bwc.html#management
func getCurrentCPUMaxFromCgroupv1() (float64, error) {
	const (
		cpuQuota  = "/sys/fs/cgroup/cpu/cpu.cfs_quota_us"
		cpuPeriod = "/sys/fs/cgroup/cpu/cpu.cfs_period_us"
	)

	quota, err := os.ReadFile(cpuQuota)
	if err != nil {
		return 0, err
	}
	period, err := os.ReadFile(cpuPeriod)
	if err != nil {
		return 0, err
	}

	quotaStr := strings.TrimSuffix(string(quota), "\n")
	periodStr := strings.TrimSuffix(string(period), "\n")

	return parseCurrentCPUMax(quotaStr, periodStr)
}

var getNumCPU = func() int { return runtime.NumCPU() } //nolint:gocritic

func parseCurrentCPUMax(quotaStr, periodStr string) (float64, error) {
	var err error
	var quota, period int

	if quotaStr == "max" || quotaStr == "-1" {
		quota = getNumCPU() * 100000 /*time quota in microseconds*/
	} else {
		quota, err = strconv.Atoi(quotaStr)
		if err != nil {
			return 0, fmt.Errorf("not parse quota, %w", err)
		}
	}
	if quota <= 0 {
		return 0, fmt.Errorf("unexpected quota %s", quotaStr)
	}

	period, err = strconv.Atoi(periodStr)
	if err != nil {
		return 0, fmt.Errorf("not parse period, %w", err)
	}
	if period <= 0 {
		return 0, fmt.Errorf("unexpected period %s", quotaStr)
	}

	maxCPU := math.Ceil(float64(quota) / float64(period))
	// Not more than NumCPU
	if numCPU := getNumCPU(); numCPU < int(maxCPU) {
		return float64(numCPU), nil
	}
	return maxCPU, nil
}
