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

const (
	inputName = "proxy"

	defaultMeasurement = "proxy"

	sampleCfg = `
[[inputs.proxy]]
    bind = "0.0.0.0"
    port = 9530
`
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}

type Input struct {
	Bind string `toml:"bind"`
	Port int    `toml:"port"`
	Path string `toml:"path"`
}

func (*Input) SampleConfig() string {
	return sampleCfg
}

func (*Input) Catalog() string {
	return "proxy"
}

func (h *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("http proxy input started...")

	listen := fmt.Sprintf("%s:%v", h.Bind, h.Port)
	l.Info("datakit proxy server start...", h.Port)

	// server
	srv := http.Server{
		Addr:    listen,
		Handler: goproxy.NewProxyHttpServer(),
	}

	go func() {
		<-datakit.Exit.Wait()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); nil != err {
			l.Errorf("server shutdown failed, err: %v\n", err)
			return
		}
		l.Info("proxy server gracefully shutdown")
	}()

	err := srv.ListenAndServe()
	if http.ErrServerClosed != err {
		l.Errorf("proxy server not gracefully shutdown, err :%v\n", err)
	}
}
