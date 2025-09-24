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

func (args *InstallerArgs) uninstallDKService(svc service.Service) {
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
}

// Install will stop/uninstall/install svc.
func (args *InstallerArgs) Install(mc *config.Config, svc service.Service) (err error) {
	args.uninstallDKService(svc)

	if err := args.injectDefInputs(mc); err != nil {
		l.Warnf("WriteDefInputs: %s, ignored", err)
	}

	// build datakit main config
	if err := mc.InitCfg(datakit.MainConfPath); err != nil {
		l.Fatalf("failed to init datakit main config: %s", err.Error())
	}

	l.Infof("Installing service %q...", dkservice.Name())
	if err := service.Control(svc, "install"); err != nil {
		l.Warnf("uninstall service failed %s", err.Error()) //nolint:lll
	}

	if args.InstallRUMSymbolTools != 0 {
		if err := cmds.InstallSymbolTools(); err != nil {
			l.Errorf("unable to install RUM symbol tools: %s", err)
			return err
		}
	}

	return nil
}

func (args *InstallerArgs) addConfdConfig(mcPrt *config.Config) {
	if args.ConfdBackend != "" {
		// 个别数据类型需要转换

		// 解析后台源，应该填写成 "[地址A:端口号A,地址B:端口号B]"字样
		if i := strings.Index(args.ConfdBackendNodes, "["); i > -1 {
			args.ConfdBackendNodes = args.ConfdBackendNodes[i+1:]
		}
		if i := strings.Index(args.ConfdBackendNodes, "]"); i > -1 {
			args.ConfdBackendNodes = args.ConfdBackendNodes[:i]
		}
		backendNodes := strings.Split(args.ConfdBackendNodes, ",")
		for i := 0; i < len(backendNodes); i++ {
			backendNodes[i] = strings.TrimSpace(backendNodes[i])
		}

		basicAuth := false
		if args.ConfdBasicAuth == "true" {
			basicAuth = true
		}

		mcPrt.Confds = []*config.ConfdCfg{{
			Enable:            true,
			Backend:           args.ConfdBackend,
			BasicAuth:         basicAuth,
			ClientCaKeys:      args.ConfdClientCaKeys,
			ClientCert:        args.ConfdClientCert,
			ClientKey:         args.ConfdClientKey,
			BackendNodes:      append(backendNodes[0:0], backendNodes...),
			Password:          args.ConfdPassword,
			Scheme:            args.ConfdScheme,
			Separator:         args.ConfdSeparator,
			Username:          args.ConfdUsername,
			AccessKey:         args.ConfdAccessKey,
			SecretKey:         args.ConfdSecretKey,
			ConfdNamespace:    args.ConfdConfdNamespace,
			PipelineNamespace: args.ConfdPipelineNamespace,
			Region:            args.ConfdRegion,
			CircleInterval:    args.ConfdCircleInterval,
		}}
	}
}
