// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package main is for Datakit upgrade
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/kardianos/service"
	"github.com/shirou/gopsutil/v3/process"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/upgrader/upgrader"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var l = logger.DefaultSLogger("main")

func isServiceRunning() (int, bool) {
	cont, err := os.ReadFile(upgrader.PidFile)
	if err != nil {
		l.Infof("unable to open manager pid file[%s]: %s", upgrader.PidFile, err)
		return 0, false
	}

	pid, err := strconv.ParseInt(strings.TrimSpace(string(cont)), 10, 32)
	if err != nil {
		l.Errorf("unable to resolve pid from [%s]", string(cont))
		return 0, false
	}

	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return 0, false
	}
	args, _ := p.CmdlineSlice()

	if len(args) > 1 {
		if strings.Contains(strings.ToLower(filepath.Base(args[0])), upgrader.ServiceName) {
			return int(pid), true
		}
	}
	return int(pid), false
}

func main() {
	upgrader.ParseAndRunSubCommand()

	// only set debug mode by env
	if v := datakit.GetEnv("DK_UPGRADER_DEBUG"); v != "" {
		upgrader.DebugRun()
		return
	}

	pid, ok := isServiceRunning()
	if ok {
		l.Errorf("service %s is already running, pid[%d]", upgrader.ServiceName, pid)
		os.Exit(upgrader.ExitStatusAlreadyRunning)
	}

	if err := doRunService(); err != nil {
		l.Errorf("unable to run datakit manager: %s", err)
		os.Exit(upgrader.ExitStatusUnableToRun)
	}
}

func doRunService() error {
	if err := os.WriteFile(upgrader.PidFile, []byte(strconv.Itoa(os.Getpid())), 0o600); err != nil {
		return fmt.Errorf("service %s unable to save pid to file: %w", upgrader.ServiceName, err)
	}

	if err := upgrader.Cfg.LoadMainTOML(upgrader.MainConfigFile); err != nil {
		l.Warnf("unable to load main config file: %s", err)
	}
	upgrader.Cfg.SetLogging()

	l = logger.SLogger("main")

	serv, err := upgrader.NewDefaultService("", nil)
	if err != nil {
		return fmt.Errorf("unable to create service: %w", err)
	}

	if err := runService(serv); err != nil {
		return fmt.Errorf("unable to start service: %w", err)
	}
	return nil
}

func runService(serv service.Service) error {
	errch := make(chan error, 32) //nolint:gomnd
	sLogger, err := serv.Logger(errch)
	if err != nil {
		return fmt.Errorf("unable to get service logger: %w", err)
	}

	if err := sLogger.Infof("%s set service logger ok, starting...", upgrader.ServiceName); err != nil {
		return err
	}

	if err := serv.Run(); err != nil {
		if serr := sLogger.Errorf("start service failed: %s", err.Error()); serr != nil {
			return serr
		}
		return err
	}

	if err := sLogger.Infof("%s service exited", upgrader.ServiceName); err != nil {
		return err
	}

	return nil
}
