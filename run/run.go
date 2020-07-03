package run

import (
	"time"

	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	l *zap.SugaredLogger
)

type Agent struct {
}

func NewAgent() (*Agent, error) {
	a := &Agent{}
	return a, nil
}

func (a *Agent) Run() error {

	l = logger.SLogger("run")

	io.Init()

	datakit.WG.Add(1)
	go func() {
		defer datakit.WG.Done()
		io.Start()
		l.Info("io goroutine exit")
	}()

	datakit.WG.Add(1)
	go func() {
		defer datakit.WG.Done()
		io.GRPCServer()
		l.Info("gRPC goroutine exit")
	}()

	if err := a.runInputs(); err != nil {
		l.Error("error running inputs: %v", err)
	}

	time.Sleep(time.Second * 2) // wait all plugin start
	datakit.WG.Add(1)
	go func() {
		defer datakit.WG.Done()
		io.HTTPServer()
		l.Info("HTTP server goroutine exit")
	}()

	return nil
}

func (a *Agent) runInputs() error {

	for _, input := range config.Cfg.Inputs {

		switch input.Input.(type) {

		//case telegraf.ServiceInput:
		//	l.Info("ignore service input ...")
		//	l.Info("starting service input ...")
		//	if err := a.runServiceInput(input, dst.ch); err != nil {
		//		return err
		//	}

		case inputs.Input:
			l.Infof("starting input %s ...", input.Config.Name)
			datakit.WG.Add(1)
			go func(i inputs.Input, name string) {
				defer datakit.WG.Done()
				i.Run()
				l.Infof("input %s exited", name)
			}(input.Input.(inputs.Input), input.Config.Name)

		default:
			l.Info("ignore interval input %s", input.Config.Name)
			//l.Info("starting interval input ...")
			//if err := a.runIntervalInput(ctx, input, startTime, dst.ch, &wg); err != nil {
			//	return err
			//}
		}
	}

	return nil
}
