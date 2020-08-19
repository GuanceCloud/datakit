package datakit

import (
	"fmt"

	"github.com/kardianos/service"

	L "gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

var (
	ServiceName        = "datakit"
	ServiceDisplayName = "datakit"
	ServiceDescription = `Collects data and upload it to DataFlux.`
	ServiceExecutable  string
	ServiceArguments   []string

	Entry func() error

	StopCh     = make(chan interface{})
	waitstopCh = make(chan interface{})
	logger     service.Logger

	l = L.DefaultSLogger("datakit")
)

type program struct{}

func NewService() (service.Service, error) {

	prog := &program{}

	svc, err := service.New(prog, &service.Config{
		Name:        ServiceName,
		DisplayName: ServiceName,
		Description: ServiceDescription,
		Executable:  ServiceExecutable,
		Arguments:   ServiceArguments,
	})

	if err != nil {
		return nil, err
	}

	return svc, nil
}

func StartService() error {

	l = L.SLogger("datakit")

	svc, err := NewService()
	if err != nil {
		return err
	}

	errch := make(chan error, 5)
	logger, err = svc.Logger(errch)
	if err != nil {
		return err
	}

	logger.Info("datakit set service logger ok, starting...")

	if err := svc.Run(); err != nil {
		logger.Errorf("start service failed: %s", err.Error())
	}

	logger.Error("datakit service exited")
	return nil
}

func (p *program) Start(s service.Service) error {
	if Entry == nil {
		return fmt.Errorf("entry not set")
	}

	return Entry()
}

func (p *program) Stop(s service.Service) error {
	//close(StopCh)

	// We must wait here:
	// On windows, we stop datakit in services.msc, if datakit process do not
	// echo to here, services.msc will complain the datakit process has been
	// exit unexpected
	//l.Info("wait waitstopCh...")
	//<-waitstopCh
	//l.Info("wait waitstopCh done")
	return nil
}

func Quit() {
	Exit.Close()

	l.Info("wait...")
	WG.Wait()

	l.Info("wg wait done")
	//close(waitstopCh)
}
