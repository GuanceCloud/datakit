package service

import (
	"fmt"
	"runtime"

	"github.com/kardianos/service"
)

var (
	ServiceName        = "datakit"
	ServiceDisplayName = "datakit"
	ServiceDescription = `Collects data and upload it to DataFlux.`
	ServiceExecutable  string
	ServiceArguments   []string

	Entry func()

	StopCh     = make(chan interface{})
	waitstopCh = make(chan interface{})
	slogger    service.Logger
)

type program struct{}

func NewService() (service.Service, error) {

	prog := &program{}

	scfg := &service.Config{
		Name:        ServiceName,
		DisplayName: ServiceName,
		Description: ServiceDescription,
		Executable:  ServiceExecutable,
		Arguments:   ServiceArguments,
	}

	if runtime.GOOS == "darwin" {
		scfg.Name = "cn.dataflux.datakit"
	}

	svc, err := service.New(prog, scfg)

	if err != nil {
		return nil, err
	}

	return svc, nil
}

func StartService() error {

	svc, err := NewService()
	if err != nil {
		return err
	}

	errch := make(chan error, 32)
	slogger, err = svc.Logger(errch)
	if err != nil {
		return err
	}

	if err := slogger.Info("datakit set service logger ok, starting..."); err != nil {
		return err
	}

	if err := svc.Run(); err != nil {
		if serr := slogger.Errorf("start service failed: %s", err.Error()); serr != nil {
			return serr
		}
		return err
	}

	if err := slogger.Info("datakit service exited"); err != nil {
		return err
	}

	return nil
}

func (p *program) Start(s service.Service) error {
	if Entry == nil {
		return fmt.Errorf("entry not set")
	}

	go Entry()
	return nil
}

func (p *program) Stop(s service.Service) error {
	close(StopCh)

	// We must wait here:
	// On windows, we stop datakit in services.msc, if datakit process do not
	// echo to here, services.msc will complain the datakit process has been
	// exit unexpected
	<-waitstopCh
	return nil
}

func Stop() {
	close(waitstopCh)
}
