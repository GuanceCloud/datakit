// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package upgrader

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

const (
	ExitStatusUnableToRun    = 101
	ExitStatusAlreadyRunning = 120
)

var (
	l = logger.DefaultSLogger(ServiceName)

	PidFile = filepath.Join(InstallDir, ServiceName+".pid")
)

const (
	DatakitCmd         = "datakit"
	BuildEntranceFile  = "cmd/upgrader/main.go"
	BuildBinName       = "dk_upgrader"
	ServiceName        = BuildBinName
	DarwinServiceName  = "com.guance." + ServiceName
	ServiceDescription = "datakit upgrade service"
)

var optionalInstallDir = map[string]string{
	datakit.OSArchWinAmd64: `C:\Program Files\` + ServiceName,
	datakit.OSArchWin386:   `C:\Program Files (x86)\` + ServiceName,

	datakit.OSArchLinuxArm:    `/usr/local/` + ServiceName,
	datakit.OSArchLinuxArm64:  `/usr/local/` + ServiceName,
	datakit.OSArchLinuxAmd64:  `/usr/local/` + ServiceName,
	datakit.OSArchLinux386:    `/usr/local/` + ServiceName,
	datakit.OSArchDarwinAmd64: `/usr/local/` + ServiceName,
	datakit.OSArchDarwinArm64: `/usr/local/` + ServiceName,
}

var (
	InstallDir     = optionalInstallDir[runtime.GOOS+"/"+runtime.GOARCH]
	DefaultLogDir  = filepath.Join("/var/log", ServiceName)
	MainConfigFile = filepath.Join(InstallDir, "main.conf")

	defaultServiceOpts = map[string]interface{}{
		"RestartSec":         10, // 重启间隔.
		"StartLimitInterval": 60, // 60秒内5次重启之后便不再启动.
		"StartLimitBurst":    5,
		"OnFailure":          "restart", // windows
	}

	defaultServiceImpl = newProgram()
)

type serviceImpl struct {
	entry func(*serviceImpl)
	stop  chan struct{}
	done  chan struct{}
}

func newProgram() *serviceImpl {
	return &serviceImpl{
		entry: entryFunc,
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}
}

func createDirs() error {
	for _, dir := range []string{
		InstallDir,
		DefaultLogDir,
	} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("create %s failed: %w", dir, err)
		}
	}
	return nil
}

func (p *serviceImpl) Start(s service.Service) error {
	if p.entry == nil {
		return fmt.Errorf("entry not set")
	}

	p.entry(p)
	return nil
}

func (p *serviceImpl) Stop(s service.Service) error {
	close(p.stop)

	// We must wait here:
	// On Windows, we stop datakit in services.msc, if datakit process do not
	// echo to here, services.msc will complain the datakit process has been
	// exit unexpected
	<-p.done
	return nil
}

func NewDefaultService(username string, args []string) (service.Service, error) {
	l = logger.SLogger(ServiceName)
	return NewService(defaultServiceImpl, username, args)
}

func NewService(program service.Interface, username string, args []string) (service.Service, error) {
	if program == nil {
		program = defaultServiceImpl
	}

	executable := filepath.Join(InstallDir, BuildBinName)
	if runtime.GOOS == datakit.OSWindows {
		executable += ".exe"
	}

	scfg := &service.Config{
		Name:        ServiceName,
		DisplayName: ServiceName,
		Description: ServiceDescription,
		Executable:  executable,
		Arguments:   args,
		Option:      defaultServiceOpts,
		UserName:    username,
	}

	if runtime.GOOS == "darwin" {
		scfg.Name = "com.guance." + ServiceName
	}

	svc, err := service.New(program, scfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create service: %w", err)
	}

	return svc, nil
}

func entryFunc(p *serviceImpl) {
	ui.c = Cfg

	go func() {
		if err := startDCA(p); err != nil {
			l.Errorf("startDCA failed: %s", err.Error())
		}
	}()

	go func() {
		<-p.stop
		close(p.done)
	}()
}
