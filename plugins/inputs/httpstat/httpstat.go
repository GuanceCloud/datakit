package httpstat

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"golang.org/x/net/http2"
)

// httpstat
type Httpstat struct {
	Url string
}

type TraceTime struct {
	trace0 time.Time
	trace1 time.Time
	trace2 time.Time
	trace3 time.Time
	trace4 time.Time
	trace5 time.Time
	trace6 time.Time
}

const sampleConfig = `
	# get http protocol request time, contain dnsLookup, tcpConnection, tlsHandshake,
	# serverProcessing, contentTransfer, and total time
	# url config set website  domain

	url = "https://www.dataflux.cn/"
`

const description = `stat http protocol request time, contain dnsLookup, tcpConnection, tlsHandshake,
	serverProcessing, contentTransfer, and total time`

func (h *Httpstat) Description() string {
	return description
}

func (h *Httpstat) SampleConfig() string {
	return sampleConfig
}

func (h *Httpstat) Gather(acc telegraf.Accumulator) error {
	h.exec(acc)
	return nil
}

func (h *Httpstat) exec(acc telegraf.Accumulator) {
	var (
		traceTime = new(TraceTime)
	)

	// 解析url
	url := parseURL(h.Url)

	req := newRequest(url)

	req = req.WithContext(httptrace.WithClientTrace(context.Background(), tracer(traceTime)))

	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	switch url.Scheme {
	case "https":
		host, _, err := net.SplitHostPort(req.Host)
		if err != nil {
			host = req.Host
		}

		tr.TLSClientConfig = &tls.Config{
			ServerName: host,
		}

		err = http2.ConfigureTransport(tr)
		if err != nil {
			log.Fatalf("failed to prepare transport for HTTP/2: %v", err)
		}
	}

	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// 执行请求
	_, err := client.Do(req)
	if err != nil {
		log.Fatalf("failed to read response: %v", err)
	}

	trace7 := time.Now()
	if traceTime.trace0.IsZero() {
		traceTime.trace0 = traceTime.trace1
	}

	fields := make(map[string]interface{})
	tags := make(map[string]string)

	if url.Scheme == "https" {
		fields["time_dns_lookup"] = traceTime.trace1.Sub(traceTime.trace0).Microseconds()
		fields["time_tcp_connection"] = traceTime.trace2.Sub(traceTime.trace1).Microseconds()
		fields["time_tls_handshake"] = traceTime.trace6.Sub(traceTime.trace5).Microseconds()
		fields["time_server_processing"] = traceTime.trace4.Sub(traceTime.trace3).Microseconds()
		fields["time_content_transfer"] = trace7.Sub(traceTime.trace4).Microseconds()
		fields["total"] = trace7.Sub(traceTime.trace0).Microseconds()
	} else {
		fields["time_dns_lookup"] = traceTime.trace1.Sub(traceTime.trace0).Microseconds()
		fields["time_tcp_connection"] = traceTime.trace2.Sub(traceTime.trace1).Microseconds()
		fields["time_server_processing"] = traceTime.trace4.Sub(traceTime.trace3).Microseconds()
		fields["time_content_transfer"] = trace7.Sub(traceTime.trace4).Microseconds()
		fields["total"] = trace7.Sub(traceTime.trace0).Microseconds()
	}

	tags["addr"] = h.Url //域名

	acc.AddFields("httpstat", fields, tags)
}

func parseURL(uri string) *url.URL {
	if !strings.Contains(uri, "://") && !strings.HasPrefix(uri, "//") {
		uri = "//" + uri
	}

	url, err := url.Parse(uri)
	if err != nil {
		log.Fatalf("could not parse url %q: %v", uri, err)
	}

	if url.Scheme == "" {
		url.Scheme = "http"
		if !strings.HasSuffix(url.Host, ":80") {
			url.Scheme += "s"
		}
	}
	return url
}

func isRedirect(resp *http.Response) bool {
	return resp.StatusCode > 299 && resp.StatusCode < 400
}

func newRequest(url *url.URL) *http.Request {
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		log.Fatalf("unable to create request: %v", err)
	}

	return req
}

func tracer(r *TraceTime) *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { r.trace0 = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { r.trace1 = time.Now() },
		ConnectStart: func(_, _ string) {
			if r.trace1.IsZero() {
				r.trace1 = time.Now()
			}
		},
		ConnectDone: func(net, addr string, err error) {
			if err != nil {
				log.Fatalf("unable to connect to host %v: %v", addr, err)
			}
			r.trace2 = time.Now()
		},
		GotConn:              func(_ httptrace.GotConnInfo) { r.trace3 = time.Now() },
		GotFirstResponseByte: func() { r.trace4 = time.Now() },
		TLSHandshakeStart:    func() { r.trace5 = time.Now() },
		TLSHandshakeDone:     func(_ tls.ConnectionState, _ error) { r.trace6 = time.Now() },
	}
}

func init() {
	inputs.Add("httpstat", func() telegraf.Input { return &Httpstat{} })
}
