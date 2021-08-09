package proxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/elazarl/goproxy"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName    = "proxy"
	sampleConfig = `
[[inputs.proxy]]
  ## statistic interval
  interval = "10s"
  ## default bind ip address
  bind = "0.0.0.0"
  ## default bind port
  port = 9530
  ## datakit clients' ip allowed to send requests to proxy server
  allowed_client_list = []
  ## allowed URL to send to proxy server
  allowed_destination_list = []

  ## extra customer headers for proxy server to add in and forward
  [inputs.proxy.headers]
    # "key1" = "value1"
    # "key2" = "value2"
    # ...
`
	l = logger.DefaultSLogger(inputName)
)

type statistic struct {
	delivered int64
	previous  int64
	reqcount  int
}

type Input struct {
	Interval          datakit.Duration  `toml:"interval"`
	Bind              string            `toml:"bind"`
	Port              int               `toml:"port"`
	AllowedClientList []string          `toml:"allowed_client_list"`
	AllowedDstList    []string          `toml:"allowed_destination_list"`
	Headers           map[string]string `toml:"headers"`
	stats             map[string]*statistic
	sync.Mutex
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) Catalog() string {
	return inputName
}

func (h *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("http proxy input started...")

	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest(goproxy.ReqConditionFunc(h.inAllowedClients), goproxy.ReqConditionFunc(h.inAllowdDestinations)).DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		for k, v := range h.Headers {
			req.Header.Set(k, v)
		}

		if req.Header.Get("X-Forwarded-For") == "" {
			if ip, _ := net.RemoteAddr(req); ip != "" {
				req.Header.Set("X-Forwarded-For", ip)
			}
		}

		if c, err := io.Copy(io.Discard, req.Body); err != nil {
			l.Error("get body size failed")
		} else {
			h.Lock()
			defer h.Unlock()

			k := req.URL.Hostname()
			stat, ok := h.stats[k]
			if !ok {
				stat = &statistic{}
				h.stats[k] = stat
			}
			if stat.delivered += c; stat.delivered < 0 {
				stat.delivered = c
			}
			if stat.reqcount += 1; stat.reqcount < 0 {
				stat.reqcount = 1
			}
		}

		return req, nil
	})

	proxysrv := http.Server{
		Addr:    fmt.Sprintf("%s:%v", h.Bind, h.Port),
		Handler: proxy,
	}

	go func() {
		l.Infof("http proxy server listening on %s", proxysrv.Addr)
		if err := proxysrv.ListenAndServe(); err != http.ErrServerClosed {
			l.Errorf("proxy server not gracefully shutdown, err :%v\n", err)
		}
	}()

	tick := time.NewTicker(h.Interval.Duration)
	for {
		select {
		case <-tick.C:
			var ms []inputs.Measurement
			for k, v := range h.stats {
				ms = append(ms, &measurement{
					name: inputName,
					tags: map[string]string{"request_host": k},
					fields: map[string]interface{}{
						"delivered":     v.delivered,
						"increment":     v.delivered - v.previous,
						"request_count": v.reqcount,
					},
					ts: time.Now(),
				})
				v.previous = v.delivered
			}
			if err := inputs.FeedMeasurement(inputName, datakit.Metric, ms, nil); err != nil {
				l.Error(err)
				dkio.FeedLastError(inputName, err.Error())
			}
		case <-datakit.Exit.Wait():
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			if err := proxysrv.Shutdown(ctx); nil != err {
				l.Errorf("server shutdown failed, err: %v\n", err)
			} else {
				l.Info("proxy server gracefully shutdown")
			}

			return
		}
	}

}

func (h *Input) inAllowedClients(req *http.Request, ctx *goproxy.ProxyCtx) bool {
	if len(h.AllowedClientList) == 0 {
		return true
	}

	clntIp, _ := net.RemoteAddr(req)
	for _, allowed := range h.AllowedClientList {
		if allowed == clntIp {
			return true
		}
	}

	return false
}

func (h *Input) inAllowdDestinations(req *http.Request, ctx *goproxy.ProxyCtx) bool {
	if len(h.AllowedDstList) == 0 {
		return true
	}

	for _, allowed := range h.AllowedDstList {
		if allowed == req.URL.Host {
			return true
		}
	}

	return false
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{stats: make(map[string]*statistic)}
	})
}
