// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package service is datakit's service manager
package service

import (
	"fmt"
	"runtime"

	"github.com/kardianos/service"
)

var (
	Name        = "datakit"
	description = `Collects data and upload it to DataFlux.`
	Executable  string
	Arguments   []string

	Entry func()

	option = map[string]interface{}{
		"RestartSec":         10, // 重启间隔.
		"StartLimitInterval": 60, // 60秒内5次重启之后便不再启动.
		"StartLimitBurst":    5,
	}

	StopCh     = make(chan interface{})
	waitStopCh = make(chan interface{})
	sLogger    service.Logger
)

type program struct{}

func NewService(userName string) (service.Service, error) {
	prog := &program{}

	scfg := &service.Config{
		Name:        Name,
		DisplayName: Name,
		Description: description,
		Executable:  Executable,
		Arguments:   Arguments,
		Option:      option,
		UserName:    userName,
	}

	if runtime.GOOS == "darwin" {
		scfg.Name = "com.guance.datakit"
	}

	svc, err := service.New(prog, scfg)
	if err != nil {
		return nil, err
	}

	return svc, nil
}

func StartService() error {
	svc, err := NewService("")
	if err != nil {
		return err
	}

	errch := make(chan error, 32) //nolint:gomnd
	sLogger, err = svc.Logger(errch)
	if err != nil {
		return err
	}

	if err := sLogger.Info("datakit set service logger ok, starting..."); err != nil {
		return err
	}

	if err := svc.Run(); err != nil {
		if serr := sLogger.Errorf("start service failed: %s", err.Error()); serr != nil {
			return serr
		}
		return err
	}

	if err := sLogger.Info("datakit service exited"); err != nil {
		return err
	}

	return nil
}

func (p *program) Start(s service.Service) error {
	if Entry == nil {
		return fmt.Errorf("entry not set")
	}

	Entry()
	return nil
}

func (p *program) Stop(s service.Service) error {
	close(StopCh)

	// We must wait here:
	// On windows, we stop datakit in services.msc, if datakit process do not
	// echo to here, services.msc will complain the datakit process has been
	// exit unexpected
	<-waitStopCh
	return nil
}

func Stop() {
	close(waitStopCh)
}
