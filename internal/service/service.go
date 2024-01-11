// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package service is datakit's service manager
package service

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/kardianos/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var (
	serviceName  = "datakit"
	displayName  = "datakit"
	serviceEntry func()
	stopCh       = make(chan any)
	waitStopCh   = make(chan any)
	sLogger      service.Logger
)

type serviceOption func(sconf *service.Config)

func Name() string {
	return serviceName
}

func DisplayName() string {
	return displayName
}

func WithUser(userName string) serviceOption {
	return func(sconf *service.Config) {
		if userName != "" {
			sconf.UserName = userName
		}
	}
}

func WithMemLimit(mem string) serviceOption {
	return func(sconf *service.Config) {
		if mem != "" {
			sconf.Option["MemoryLimit"] = mem
		}
	}
}

func WithCPULimit(cpu string) serviceOption {
	return func(sconf *service.Config) {
		if cpu != "" {
			sconf.Option["CPUQuota"] = cpu
		}
	}
}

func WithName(name string) serviceOption {
	return func(sconf *service.Config) {
		sconf.Name = name
		sconf.DisplayName = name
	}
}

func WithDescription(desc string) serviceOption {
	return func(sconf *service.Config) {
		sconf.Description = desc
	}
}

func WithExecutable(exec string, args []string) serviceOption {
	return func(sconf *service.Config) {
		sconf.Executable = exec
		sconf.Arguments = args
	}
}

type program struct{}

func NewService(opts ...serviceOption) (service.Service, error) {
	scfg := &service.Config{
		Option: map[string]any{
			"RestartSec":         10, // 重启间隔.
			"StartLimitInterval": 60, // 60秒内5次重启之后便不再启动.
			"StartLimitBurst":    5,
		},
		Name:        serviceName,
		DisplayName: displayName,
		Description: "Collects data and upload it to Guance Cloud.",
		Executable:  datakit.DatakitBinaryPath(),
		UserName:    "root",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(scfg)
		}
	}

	if scfg.UserName == "root" {
		scfg.Option["MemoryLimit"] = ""
		scfg.Option["CPUQuota"] = ""
	}

	if runtime.GOOS == "darwin" {
		scfg.Name = "com.guance.datakit"
	}

	prog := &program{}
	svc, err := service.New(prog, scfg)
	if err != nil {
		return nil, err
	}

	return svc, nil
}

func StartService(entry func()) error {
	serviceEntry = entry

	executable := filepath.Join(datakit.InstallDir, "datakit")
	if runtime.GOOS == datakit.OSWindows {
		executable += ".exe"
	}

	svc, err := NewService(WithExecutable(executable, nil))
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
	if serviceEntry == nil {
		return fmt.Errorf("service entry not set")
	}

	serviceEntry()
	return nil
}

func (p *program) Stop(s service.Service) error {
	close(stopCh)

	// We must wait here:
	// On windows, we stop datakit in services.msc, if datakit process do not
	// echo to here, services.msc will complain the datakit process has been
	// exit unexpected
	<-waitStopCh
	return nil
}

func Wait() <-chan any {
	return stopCh
}

func Stop() {
	close(waitStopCh)
}
