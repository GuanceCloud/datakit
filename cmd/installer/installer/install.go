// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package installer implements datakit's install and upgrade
package installer

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kardianos/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/sinkfuncs"
)

var InstallExternals = ""

func Install(svc service.Service) {
	svcStatus, err := svc.Status()
	if err != nil {
		if errors.Is(err, service.ErrNotInstalled) {
			l.Infof("datakit service not installed before")
			// pass
		} else {
			l.Warnf("svc.Status: %s", err.Error())
		}
	} else {
		switch svcStatus {
		case service.StatusUnknown: // pass
		case service.StatusStopped:
			if err := service.Control(svc, "uninstall"); err != nil {
				l.Warnf("uninstall service failed: %s", err.Error())
			}
		case service.StatusRunning: // should not been here
			l.Warnf("unexpected: datakit service should have stopped")
			if err := service.Control(svc, "uninstall"); err != nil {
				l.Warnf("uninstall service failed: %s", err.Error())
			}
		}
	}

	mc := config.Cfg

	// prepare dataway info and check token format
	if len(Dataway) != 0 {
		var err error
		mc.DataWay, err = getDataWay()
		if err != nil {
			l.Fatal(err)
		}
		cp.Infof("Set dataway to %s\n", Dataway)
	}

	if OTA {
		mc.AutoUpdate = OTA
		l.Info("set auto update(OTA enabled)")
	}

	if DCAEnable != "" {
		config.Cfg.DCAConfig.Enable = true
		if DCAWhiteList != "" {
			config.Cfg.DCAConfig.WhiteList = strings.Split(DCAWhiteList, ",")
		}

		// check white list whether is empty or invalid
		if len(config.Cfg.DCAConfig.WhiteList) == 0 {
			l.Fatalf("DCA service is enabled, but white list is empty! ")
		}

		for _, cidr := range config.Cfg.DCAConfig.WhiteList {
			if _, _, err := net.ParseCIDR(cidr); err != nil {
				if net.ParseIP(cidr) == nil {
					l.Fatalf("DCA white list set error, invalid IP: %s", cidr)
				}
			}
		}

		if DCAListen != "" {
			config.Cfg.DCAConfig.Listen = DCAListen
		}

		l.Infof("DCA enabled, listen on %s, whiteliste: %s", DCAListen, DCAWhiteList)
	}

	if EnablePProf != "" {
		config.Cfg.EnablePProf = true
		if PProfListen != "" {
			config.Cfg.PProfListen = PProfListen
		}
		l.Infof("pprof enabled? %v, listen on %s", config.Cfg.EnablePProf, config.Cfg.PProfListen)
	}

	// Only linux support cgroup.
	if CgroupDisabled != 1 && runtime.GOOS == datakit.OSLinux {
		mc.Cgroup.Enable = true

		if LimitCPUMin > 0 {
			mc.Cgroup.CPUMin = LimitCPUMin
		}

		if LimitCPUMax > 0 {
			mc.Cgroup.CPUMax = LimitCPUMax
		}

		if mc.Cgroup.CPUMax < mc.Cgroup.CPUMin {
			l.Fatalf("invalid CGroup CPU limit, max should larger than min")
		}

		if LimitMemMax > 0 {
			l.Infof("cgroup set max memory to %dMB", LimitMemMax)
			mc.Cgroup.MemMax = LimitMemMax
		} else {
			l.Infof("cgroup max memory not set")
		}

		l.Infof("croups enabled under %s, cpu: [%f, %f], mem: %dMB",
			runtime.GOOS, mc.Cgroup.CPUMin, mc.Cgroup.CPUMin, mc.Cgroup.MemMax)
	} else {
		mc.Cgroup.Enable = false
		l.Infof("cgroup disabled, OS: %s", runtime.GOOS)
	}

	if HostName != "" {
		mc.Environments["ENV_HOSTNAME"] = HostName
		l.Infof("set ENV_HOSTNAME to %s", HostName)
	}

	if GlobalHostTags != "" {
		mc.GlobalHostTags = config.ParseGlobalTags(GlobalHostTags)

		l.Infof("set global host tags to %+#v", mc.GlobalHostTags)
	}

	if GlobalElectionTags != "" {
		mc.Election.Tags = config.ParseGlobalTags(GlobalElectionTags)
		l.Infof("set global election tags to %+#v", mc.Election.Tags)
	}

	if EnableElection != "" {
		mc.Election.Enable = true
		l.Infof("election enabled? %v", true)
	}
	mc.Election.Namespace = ElectionNamespace
	l.Infof("set election namespace to %s", mc.Election.Namespace)

	mc.HTTPAPI.Listen = fmt.Sprintf("%s:%d", HTTPListen, HTTPPort)
	l.Infof("set HTTP listen to %s", mc.HTTPAPI.Listen)

	mc.InstallVer = DataKitVersion
	l.Infof("install version %s", mc.InstallVer)

	if DatakitName != "" {
		mc.Name = DatakitName
		l.Infof("set datakit name to %s", mc.Name)
	}

	if GitURL != "" {
		mc.GitRepos = &config.GitRepost{
			PullInterval: GitPullInterval,
			Repos: []*config.GitRepository{
				{
					Enable:                true,
					URL:                   GitURL,
					SSHPrivateKeyPath:     GitKeyPath,
					SSHPrivateKeyPassword: GitKeyPW,
					Branch:                GitBranch,
				}, // GitRepository
			}, // Repos
		} // GitRepost
	}

	if RumDisable404Page != "" {
		l.Infof("set disable 404 page: %v", RumDisable404Page)
		mc.HTTPAPI.Disable404Page = true
	}

	if RumOriginIPHeader != "" {
		l.Infof("set rum origin IP header: %s", RumOriginIPHeader)
		mc.HTTPAPI.RUMOriginIPHeader = RumOriginIPHeader
	}

	if LogLevel != "" {
		mc.Logging.Level = LogLevel
		l.Infof("set log level to %s", LogLevel)
	}

	if Log != "" {
		mc.Logging.Log = Log
		l.Infof("set log to %s", Log)
	}

	if GinLog != "" {
		l.Infof("set gin log to %s", GinLog)
		mc.Logging.GinLog = GinLog
	}

	// parse sink
	if err := parseSinkArgs(mc); err != nil {
		mc.Sinks.Sink = []map[string]interface{}{{}} // clear
		l.Fatalf("parseSinkArgs failed: %s", err.Error())
	}
	if LogSinkDetail != "" {
		l.Info("set enable log sink detail.")
		mc.LogSinkDetail = true
	}

	writeDefInputToMainCfg(mc)

	// build datakit main config
	if err := mc.InitCfg(datakit.MainConfPath); err != nil {
		l.Fatalf("failed to init datakit main config: %s", err.Error())
	}

	installExts := map[string]struct{}{}
	for _, v := range strings.Split(InstallExternals, ",") {
		installExts[v] = struct{}{}
	}
	updateEBPF := false
	if runtime.GOOS == datakit.OSLinux && (runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64") {
		if _, err := os.Stat(filepath.Join(datakit.InstallDir, "externals", "datakit-ebpf")); err == nil {
			updateEBPF = true
		}
		if _, ok := installExts["ebpf"]; ok {
			updateEBPF = true
		}
	}
	if updateEBPF {
		cp.Infof("Install DataKit eBPF plugin...")
		// nolint:gosec
		cmd := exec.Command(filepath.Join(datakit.InstallDir, "datakit"), "install", "--ebpf")
		if msg, err := cmd.CombinedOutput(); err != nil {
			l.Errorf("upgradde external input plugin %s failed: %s msg: %s", "ebpf", err.Error(), msg)
		}
	}

	cp.Infof("Installing service %s...\n", dkservice.Name)
	if err := service.Control(svc, "install"); err != nil {
		l.Warnf("uninstall service failed %s", err.Error()) //nolint:lll
	}
}

func parseSinkArgs(mc *config.Config) error {
	if mc == nil {
		return fmt.Errorf("invalid main config")
	}

	if mc.Sinks == nil {
		return fmt.Errorf("invalid main config sinks")
	}

	categoryShorts := []string{
		datakit.SinkCategoryMetric,
		datakit.SinkCategoryNetwork,
		datakit.SinkCategoryKeyEvent,
		datakit.SinkCategoryObject,
		datakit.SinkCategoryCustomObject,
		datakit.SinkCategoryLogging,
		datakit.SinkCategoryTracing,
		datakit.SinkCategoryRUM,
		datakit.SinkCategorySecurity,
		datakit.SinkCategoryProfiling,
	}

	args := []string{
		SinkMetric,
		SinkNetwork,
		SinkKeyEvent,
		SinkObject,
		SinkCustomObject,
		SinkLogging,
		SinkTracing,
		SinkRUM,
		SinkSecurity,
		SinkProfiling,
	}

	sinks, err := sinkfuncs.GetSinkFromEnvs(categoryShorts, args)
	if err != nil {
		return err
	}

	mc.Sinks.Sink = sinks
	return nil
}
