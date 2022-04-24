// Package proxy used to proxy HTTP request for no-public-network datakits.
package proxy

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/elazarl/goproxy"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName    = "proxy"
	sampleConfig = `
[[inputs.proxy]]
  ## default bind ip address
  bind = "0.0.0.0"
  ## default bind port
  port = 9530
`
	log = logger.DefaultSLogger(inputName)
)

type proxyLogger struct{}

func (pl *proxyLogger) Printf(format string, v ...interface{}) {
	log.Infof(format, v...)
}

type Input struct {
	Bind string `toml:"bind"`
	Port int    `toml:"port"`

	semStop *cliutils.Sem // start stop signal

	PathDeprecated   string `toml:"path,omitempty"`
	WSBindDeprecated string `toml:"ws_bind,omitempty"`
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return nil
}

func (ipt *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("http proxy input started...")

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false
	proxy.Logger = &proxyLogger{}
	proxysrv := &http.Server{
		Addr:    fmt.Sprintf("%s:%v", ipt.Bind, ipt.Port),
		Handler: proxy,
	}

	go func(proxysrv *http.Server) {
		log.Infof("http proxy server listening on %s", proxysrv.Addr)
		if err := proxysrv.ListenAndServe(); err != nil {
			log.Error(err)
		}
	}(proxysrv)

	stopFunc := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := proxysrv.Shutdown(ctx); nil != err {
			log.Errorf("server shutdown failed, err: %sn", err.Error())
		} else {
			log.Info("proxy server gracefully shutdown")
		}
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			stopFunc()
			return

		case <-ipt.semStop.Wait():
			stopFunc()
			return
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			semStop: cliutils.NewSem(),
		}
	})
}
