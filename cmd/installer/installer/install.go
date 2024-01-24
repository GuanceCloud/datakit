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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
)

var InstallExternals = ""

func Install(svc service.Service, userName string) {
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
	mc.DatakitUser = userName

	// prepare dataway info and check token format
	if len(Dataway) != 0 {
		var err error
		mc.Dataway, err = getDataway()
		if err != nil {
			l.Errorf("getDataway failed: %s", err.Error())
			l.Fatal(err)
		}

		l.Infof("Set dataway to %s", Dataway)

		mc.Dataway.GlobalCustomerKeys = dataway.ParseGlobalCustomerKeys(SinkerGlobalCustomerKeys)
		mc.Dataway.EnableSinker = (EnableSinker != "")

		l.Infof("Set dataway global sinker customer keys: %+#v", mc.Dataway.GlobalCustomerKeys)
	}

	if OTA {
		mc.AutoUpdate = OTA
		l.Info("set auto update(OTA enabled)")
	}

	if HTTPPublicAPIs != "" {
		apis := strings.Split(HTTPPublicAPIs, ",")
		idx := 0
		for _, api := range apis {
			api = strings.TrimSpace(api)
			if api != "" {
				if !strings.HasPrefix(api, "/") {
					api = "/" + api
				}
				apis[idx] = api
				idx++
			}
		}
		mc.HTTPAPI.PublicAPIs = apis[:idx]
		l.Infof("set PublicAPIs to %s", strings.Join(apis[:idx], ","))
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

	// Only supports linux and windows
	if LimitDisabled != 1 && (runtime.GOOS == datakit.OSLinux || runtime.GOOS == datakit.OSWindows) {
		mc.ResourceLimitOptions.Enable = true

		if LimitCPUMax > 0 {
			mc.ResourceLimitOptions.CPUMax = LimitCPUMax
		}

		if LimitMemMax > 0 {
			l.Infof("resource limit set max memory to %dMB", LimitMemMax)
			mc.ResourceLimitOptions.MemMax = LimitMemMax
		} else {
			l.Infof("resource limit max memory not set")
		}

		l.Infof("resource limit enabled under %s, cpu: %f, mem: %dMB",
			runtime.GOOS, mc.ResourceLimitOptions.CPUMax, mc.ResourceLimitOptions.MemMax)
	} else {
		mc.ResourceLimitOptions.Enable = false
		l.Infof("resource limit disabled, OS: %s", runtime.GOOS)
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

	writeDefInputToMainCfg(mc, false)

	// build datakit main config
	if err := mc.InitCfg(datakit.MainConfPath); err != nil {
		l.Fatalf("failed to init datakit main config: %s", err.Error())
	}

	l.Infof("Installing service %q...", dkservice.Name())
	if err := service.Control(svc, "install"); err != nil {
		l.Warnf("uninstall service failed %s", err.Error()) //nolint:lll
	}

	if InstallRUMSymbolTools != 0 {
		if err := cmds.InstallSymbolTools(); err != nil {
			l.Fatalf("unable to install RUM symbol tools: %s", err)
		}
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
