// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package upgrader is used for Datakit remote upgrade
package upgrader

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/kardianos/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/upgrader/upgrader"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
)

var l = logger.DefaultSLogger("upgrade")

func newUpgradeService(username string) (service.Service, error) {
	envArgs := []string{"env"}

	if envPath := os.Getenv("PATH"); envPath != "" {
		// Set service bootstrap PATH env variable
		envArgs = append(envArgs, "--PATH="+envPath)
	}

	if dkHome := os.Getenv(upgrader.ENVDatakitHome); dkHome != "" {
		envArgs = append(envArgs, fmt.Sprintf("--%s=%s", upgrader.ENVDatakitHome, dkHome))
	}

	if dkUpgradeHome := os.Getenv(upgrader.ENVDKUpgraderHome); dkUpgradeHome != "" {
		envArgs = append(envArgs, fmt.Sprintf("--%s=%s", upgrader.ENVDKUpgraderHome, dkUpgradeHome))
	}

	var args []string
	if len(envArgs) > 1 {
		args = append(args, envArgs...)
	}

	return upgrader.NewDefaultService(username, args)
}

func StopUpgradeService(username string) {
	l = logger.SLogger(upgrader.ServiceName)

	serv, err := newUpgradeService(username)
	if err != nil {
		l.Fatalf("unable to create upgrade service: %s", err)
	}

	svcStatus, err := serv.Status()
	if err != nil {
		if errors.Is(err, service.ErrNotInstalled) {
			cp.Infof("%s service not installed before\n", upgrader.ServiceName)
		} else {
			l.Warnf("svc.Status: %s, ignored", err.Error())
		}
	} else {
		switch svcStatus {
		case service.StatusUnknown: // not installed
			cp.Infof("%s service maybe not installed\n", upgrader.ServiceName)
		case service.StatusStopped: // pass
			cp.Infof("%s service stopped\n", upgrader.ServiceName)
		case service.StatusRunning:
			cp.Infof("Stopping running %s...\n", upgrader.ServiceName)
			if err = serv.Stop(); err != nil {
				l.Warnf("stop service failed %s, ignored", err.Error())
			}
		}
	}
}

func InstallUpgradeService(username string, flagDKUpgrade bool, flagInstallOnly int,
	flagUpgradeManagerService int, flagUpgradeServIPWhiteList string,
) {
	l = logger.SLogger(upgrader.ServiceName)

	serv, err := newUpgradeService(username)
	if err != nil {
		l.Fatalf("unable to create upgrade service: %s", err)
	}

	if err := upgrader.CreateDirs(); err != nil {
		l.Fatalf("unable to create directory: %s", err)
	}

	if !flagDKUpgrade || flagUpgradeManagerService == 1 {
		if err := serv.Install(); err != nil {
			l.Warnf("unable to install %s service: %s", upgrader.ServiceName, err)
		}

		loadMainConfOK := true
		if flagUpgradeManagerService == 1 && datakit.FileExist(upgrader.MainConfigFile) {
			if err := upgrader.Cfg.LoadMainTOML(upgrader.MainConfigFile); err != nil {
				l.Warnf("unable to load current main config: %s", err)
				loadMainConfOK = false
			}
		}
		if flagUpgradeServIPWhiteList != "" {
			upgrader.Cfg.IPWhiteList = strings.Split(strings.TrimSpace(flagUpgradeServIPWhiteList), ",")
		}
		if loadMainConfOK {
			if err := upgrader.Cfg.CreateCfgFile(); err != nil {
				l.Warnf("unable to create main config file: %s", err)
			}
		}

		if flagInstallOnly != 0 {
			cp.Warnf("Only install service %s, NOT started\n", upgrader.ServiceName)
		} else {
			cp.Infof("Starting service %s...\n", upgrader.ServiceName)
			if err = serv.Start(); err != nil {
				cp.Warnf("Start service %s failed: %s\n", upgrader.ServiceName, err.Error())
			}
		}
	}
}
