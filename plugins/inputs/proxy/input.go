package proxy

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/elazarl/goproxy"

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
	l = logger.DefaultSLogger(inputName)
)

type Input struct {
	Bind    string `toml:"bind"`
	Port    int    `toml:"port"`
	Verbose bool   `toml:"verbose"`
}

type proxyLogger struct{}

func (pl *proxyLogger) Printf(format string, v ...interface{}) {
	l.Infof(format, v...)
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
	return []inputs.Measurement{
		//&measurement{}
	}
}

func (h *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("http proxy input started...")

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = h.Verbose
	proxy.Logger = &proxyLogger{}
	proxysrv := &http.Server{
		Addr:    fmt.Sprintf("%s:%v", h.Bind, h.Port),
		Handler: proxy,
	}

	go func(proxysvr *http.Server) {
		l.Infof("http proxy server listening on %s", proxysrv.Addr)
		if err := proxysrv.ListenAndServe(); err != http.ErrServerClosed {
			l.Errorf("proxy server not gracefully shutdown, err :%v\n", err)
		} else {
			l.Error(err)
		}
	}(proxysrv)

	<-datakit.Exit.Wait()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := proxysrv.Shutdown(ctx); nil != err {
		l.Errorf("server shutdown failed, err: %v\n", err)
	} else {
		l.Info("proxy server gracefully shutdown")
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
