// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package installer implements datakit's install and upgrade
package installer

import (
	"errors"
	"strings"

	"github.com/kardianos/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
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
	// load DK_XXX env config
	mc = loadDKEnvCfg(mc)

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
