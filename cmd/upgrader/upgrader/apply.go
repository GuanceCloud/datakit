// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package upgrader

import (
	"errors"
	"fmt"
	"os"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/kardianos/service"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func newUpgradeService(username string) (service.Service, error) {
	envArgs := []string{"env"}

	if envPath := os.Getenv("PATH"); envPath != "" {
		// Set service bootstrap PATH env variable
		envArgs = append(envArgs, "--PATH="+envPath)
	}

	if dkHome := os.Getenv(ENVDatakitHome); dkHome != "" {
		envArgs = append(envArgs, fmt.Sprintf("--%s=%s", ENVDatakitHome, dkHome))
	}

	if dkUpgradeHome := os.Getenv(ENVDKUpgraderHome); dkUpgradeHome != "" {
		envArgs = append(envArgs, fmt.Sprintf("--%s=%s", ENVDKUpgraderHome, dkUpgradeHome))
	}

	var args []string
	if len(envArgs) > 1 {
		args = append(args, envArgs...)
	}

	return NewDefaultService(username, args)
}

// StopUpgradeService stop upgrader service.
func StopUpgradeService(username string) {
	l = logger.SLogger(ServiceName)

	serv, err := newUpgradeService(username)
	if err != nil {
		l.Fatalf("unable to create upgrade service: %s", err)
	}

	svcStatus, err := serv.Status()
	if err != nil {
		if errors.Is(err, service.ErrNotInstalled) {
			cp.Infof("%s service not installed before\n", ServiceName)
		} else {
			cp.Warnf("svc.Status: %s, ignored\n", err.Error())
		}
	} else {
		switch svcStatus {
		case service.StatusUnknown: // not installed
			cp.Infof("%s service maybe not installed\n", ServiceName)
		case service.StatusStopped: // pass
			cp.Infof("%s service stopped\n", ServiceName)
		case service.StatusRunning:
			cp.Infof("Stopping running %s...\n", ServiceName)
			if err = serv.Stop(); err != nil {
				l.Warnf("stop service failed %s, ignored", err.Error())
			}
		}
	}
}

// InstallUpgradeService install upgrader service used to upgrade datakit version remotely.
func InstallUpgradeService(opts ...UpgraderOpt) error {
	l = logger.SLogger(ServiceName)

	c := Cfg

	// load exist(old) main.conf
	if datakit.FileExist(MainConfigFile) {
		if err := c.LoadMainTOML(MainConfigFile); err != nil {
			l.Warnf("unable to load current main config: %s", err)
		}
	}

	// apply new options based on exist main.conf.
	for _, opt := range opts {
		opt(c)
	}

	// Under DK upgrade, upgrader-service itself should not upgrade by default
	// except that user need to upgrade the ugrader-service.
	// Under DK install, the upgrader service installed and started by default.
	if !c.dkUpgrade || c.upgradeUpgraderService {
		serv, err := newUpgradeService(c.Username)
		if err != nil {
			l.Warnf("unable to create upgrade service: %s", err)
			return err
		}

		if err := createDirs(); err != nil {
			l.Warnf("unable to create directory: %s", err)
			return err
		}

		status, err := serv.Status()
		if err != nil {
			l.Warnf("get dk-upgrader service status failed: %s, ignored", err)
		}

		switch status {
		case service.StatusRunning:
			if err := serv.Stop(); err != nil {
				l.Warnf("stop dk-upgrader service failed: %s, ignored", err)
			}
			if err := serv.Uninstall(); err != nil {
				l.Warnf("uninstall dk-upgrader service failed: %s, ignored", err)
			}
		case service.StatusStopped:
			l.Warnf("dk-upgrader service stopped")
			if err := serv.Uninstall(); err != nil {
				l.Warnf("uninstall dk-upgrader service failed: %s, ignored", err)
			}

		case service.StatusUnknown: // pass
			l.Infof("dk-upgrader service status unknown, maybe 1st time install")
		}

		if err := serv.Install(); err != nil {
			l.Warnf("unable to install %s service: %s", ServiceName, err)
			return err
		}

		// create/re-write main.conf
		if err := c.createCfgFile(); err != nil {
			l.Warnf("unable to create main config file: %s", err)
			return err
		}

		if c.installOnly {
			l.Warnf("Only install service %s, NOT started", ServiceName)
		} else {
			l.Infof("Starting service %s...", ServiceName)
			if err = serv.Start(); err != nil {
				l.Warnf("Start service %s failed: %s", ServiceName, err.Error())
				return err
			}
		}
	}

	return nil
}
