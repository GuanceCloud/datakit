// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package proxy used to proxy HTTP request for no-public-network datakits.
package proxy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/elazarl/goproxy"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var (
	inputName    = "proxy"
	sampleConfig = `
[[inputs.proxy]]
  ## default bind ip address
  bind = "0.0.0.0"
  ## default bind port
  port = 9530

  # verbose mode will show more info about during proxying.
  verbose = false

  # mitm: man-in-the-middle mode
  mitm = false
`
	log = logger.DefaultSLogger("input-proxy")
)

type proxyLogger struct{}

func (pl *proxyLogger) Printf(format string, v ...interface{}) {
	log.Infof(format, v...)
}

type Input struct {
	Bind    string `toml:"bind"`
	Port    int    `toml:"port"`
	Verbose bool   `toml:"verbose"`
	MITM    bool   `toml:"mitm"`

	semStop *cliutils.Sem // start stop signal

	proxyServer *http.Server
	proxy       *goproxy.ProxyHttpServer

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
	return datakit.AllOS
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return nil
}

func (ipt *Input) HandleConnect(req string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
	cliIP := "not-set"
	if addr, err := net.ResolveTCPAddr("tcp", ctx.Req.RemoteAddr); err != nil {
		log.Warnf("HandleConnect: invalid client addr(%q), ignored", ctx.Req.RemoteAddr)
	} else {
		cliIP = addr.IP.String()
	}

	log.Debugf("handle connect from %s...", cliIP)

	proxyConnectVec.WithLabelValues(
		cliIP,
	).Inc()

	if ipt.MITM {
		return goproxy.MitmConnect, req
	} else {
		return goproxy.OkConnect, req
	}
}

func (ipt *Input) doInitProxy() {
	p := goproxy.NewProxyHttpServer()
	p.Verbose = ipt.Verbose
	p.Logger = &proxyLogger{}

	p.OnRequest().HandleConnect(ipt)
	p.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			if ctx.Error != nil {
				log.Warnf("on request got error from proxy context: %s", ctx.Error.Error())
			}

			r.Header.Add("X-Proxy-Time", fmt.Sprintf("%d", time.Now().UnixNano()))
			proxyReqVec.WithLabelValues(
				r.URL.Path,
				r.Method,
			).Inc()

			return r, nil
		})

	p.OnResponse().DoFunc(
		func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			if ctx.Error != nil {
				log.Warnf("on response got error from proxy context: %s", ctx.Error.Error())
			}

			if ctx.Req == nil {
				log.Warnf("empty request")
				return resp
			}

			status := "nil response"
			if ctx.Resp != nil {
				status = http.StatusText(ctx.Resp.StatusCode)
			}

			if ctx.Error != nil {
				log.Warnf("%s: %s", status, ctx.Error.Error())
			}

			pt := ctx.Req.Header.Get("X-Proxy-Time")
			if nsec, err := strconv.ParseInt(pt, 10, 64); err == nil {
				proxyReqLatencyVec.WithLabelValues(ctx.Req.URL.Path,
					ctx.Req.Method,
					status,
				).Observe(float64(time.Since(time.Unix(0, nsec))) / float64(time.Second))
			} else {
				log.Warnf("invalid X-Proxy-Time: %q", pt)
			}

			return resp
		})

	ipt.proxyServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%v", ipt.Bind, ipt.Port),
		Handler: p,
	}

	ipt.proxy = p
}

func (ipt *Input) Run() {
	log = logger.SLogger("input-proxy")
	log.Infof("HTTP proxy input started...")

	ipt.doInitProxy()

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_proxy"})

	g.Go(func(ctx context.Context) error {
		log.Infof("http proxy server listening on %s", ipt.proxyServer.Addr)
		for {
			if err := ipt.proxyServer.ListenAndServe(); err != nil {
				if errors.Is(err, http.ErrServerClosed) {
					log.Info("proxy server closed")
					break
				} else {
					log.Warnf("ListenAndServe: %s, retry...", err.Error())
					time.Sleep(time.Second)
				}
			}
		}
		return nil
	})

	for {
		select {
		case <-datakit.Exit.Wait():
			ipt.stop()
			return

		case <-ipt.semStop.Wait():
			ipt.stop()
			return
		}
	}
}

func (ipt *Input) stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := ipt.proxyServer.Shutdown(ctx); nil != err {
		log.Warnf("Shutdown: %s, ignored", err.Error())
	}

	log.Info("proxy server gracefully shutdown")
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
