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
	"runtime"
	"strings"

	"github.com/kardianos/service"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
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
		mc.Dataway, err = getDataway()
		if err != nil {
			cp.Errorf("%s\n", err.Error())
			l.Fatal(err)
		}
		cp.Infof("Set dataway to %s\n", Dataway)
	}

	if Sinker != "" {
		mc.LoadSink(Sinker)
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

	if PProfListen != "" {
		config.Cfg.PProfListen = PProfListen
	}
	l.Infof("pprof enabled? %v, listen on %s", config.Cfg.EnablePProf, config.Cfg.PProfListen)

	// Only linux support cgroup.
	if CgroupDisabled != 1 && runtime.GOOS == datakit.OSLinux {
		mc.Cgroup.Enable = true

		if LimitCPUMax > 0 {
			mc.Cgroup.CPUMax = LimitCPUMax
		}

		if LimitMemMax > 0 {
			l.Infof("cgroup set max memory to %dMB", LimitMemMax)
			mc.Cgroup.MemMax = LimitMemMax
		} else {
			l.Infof("cgroup max memory not set")
		}

		l.Infof("croups enabled under %s, cpu: %f, mem: %dMB",
			runtime.GOOS, mc.Cgroup.CPUMax, mc.Cgroup.MemMax)
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

	// 一个 func 最大 230 行
	addConfdConfig(mc)

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

	writeDefInputToMainCfg(mc)

	// build datakit main config
	if err := mc.InitCfg(datakit.MainConfPath); err != nil {
		l.Fatalf("failed to init datakit main config: %s", err.Error())
	}

	cp.Infof("Installing service %s...\n", dkservice.Name)
	if err := service.Control(svc, "install"); err != nil {
		l.Warnf("uninstall service failed %s", err.Error()) //nolint:lll
	}
}

func addConfdConfig(mcPrt *config.Config) {
	if ConfdBackend != "" {
		// 个别数据类型需要转换

		// 解析后台源，应该填写成 "[地址A:端口号A,地址B:端口号B]"字样
		if i := strings.Index(ConfdBackendNodes, "["); i > -1 {
			ConfdBackendNodes = ConfdBackendNodes[i+1:]
		}
		if i := strings.Index(ConfdBackendNodes, "]"); i > -1 {
			ConfdBackendNodes = ConfdBackendNodes[:i]
		}
		backendNodes := strings.Split(ConfdBackendNodes, ",")
		for i := 0; i < len(backendNodes); i++ {
			backendNodes[i] = strings.TrimSpace(backendNodes[i])
		}

		basicAuth := false
		if ConfdBasicAuth == "true" {
			basicAuth = true
		}

		mcPrt.Confds = []*config.ConfdCfg{{
			Enable:            true,
			Backend:           ConfdBackend,
			BasicAuth:         basicAuth,
			ClientCaKeys:      ConfdClientCaKeys,
			ClientCert:        ConfdClientCert,
			ClientKey:         ConfdClientKey,
			BackendNodes:      append(backendNodes[0:0], backendNodes...),
			Password:          ConfdPassword,
			Scheme:            ConfdScheme,
			Separator:         ConfdSeparator,
			Username:          ConfdUsername,
			AccessKey:         ConfdAccessKey,
			SecretKey:         ConfdSecretKey,
			ConfdNamespace:    ConfdConfdNamespace,
			PipelineNamespace: ConfdPipelineNamespace,
			Region:            ConfdRegion,
			CircleInterval:    ConfdCircleInterval,
		}}
	}
}
