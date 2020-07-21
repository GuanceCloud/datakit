package run

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/run/httpserver"
)

var (
	l *logger.Logger
)

type Agent struct {
}

func NewAgent() (*Agent, error) {
	a := &Agent{}
	return a, nil
}

func (a *Agent) Run() error {

	l = logger.SLogger("run")

	io.Start()

	for name := range config.Cfg.Inputs {
		if _, ok := httpserver.OwnerList[name]; ok {
			datakit.WG.Add(1)
			go func() {
				defer datakit.WG.Done()
				httpserver.Start(config.Cfg.MainCfg.HTTPServerAddr)
				l.Info("HTTPServer goroutine exit")
			}()
			break
		}
	}

	if err := a.runInputs(); err != nil {
		l.Error("error running inputs: %v", err)
	}

	return nil
}

func (a *Agent) runInputs() error {

	for name, ips := range config.Cfg.Inputs {

		for _, input := range ips {

			switch input.(type) {

			case inputs.Input:
				l.Infof("starting input %s ...", name)
				datakit.WG.Add(1)
				go func(i inputs.Input, name string) {
					defer datakit.WG.Done()
					i.Run()
					l.Infof("input %s exited", name)
				}(input, name)

			default:
				l.Warn("ignore input %s", name)
			}
		}

	}

	return nil
}
