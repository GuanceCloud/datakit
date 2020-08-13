package httpstat

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"regexp"
	"strings"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	l         *logger.Logger
	inputName = "httpstat"
)

// project
type httpPing struct {
	cfg           *Action
	metricName    string
	url           string
	host          string
	method        string
	uAgent        string
	buf           string
	transport     *http.Transport
	rAddr         net.Addr
	nsTime        time.Duration
	kAlive        bool
	compress      bool
	tlsSkipVerify bool
}

// Result holds Ping result
type Result struct {
	Trace Trace
}

// Trace holds trace results
type Trace struct {
	dnsLookupTime    time.Duration
	connectionTime   time.Duration
	toFirstByteTime  time.Duration
	tlsHandshakeTime time.Duration
	totalTime        time.Duration
}

func (h *Httpstat) Description() string {
	return description
}

func (h *Httpstat) SampleConfig() string {
	return httpstatConfigSample
}

func (_ *Httpstat) Catalog() string {
	return "network"
}

func (h *Httpstat) Run() {
	l = logger.SLogger("httpStat")

	l.Info("httpStat input started...")

	if h.Interval != "" {
		du, err := time.ParseDuration(h.Interval)
		if err != nil {
			l.Errorf("bad interval %s: %s, use default: 10s", h.Interval, err.Error())
			return
		} else {
			h.IntervalDuration = du
		}
	}

	if h.MetricName == "" {
		h.MetricName = "httpstat"
	}

	tick := time.NewTicker(h.IntervalDuration)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			// handle
			h.run()
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}
}

func (h *Httpstat) run() {
	for _, c := range h.Actions {
		p := &httpPing{
			cfg:        c,
			metricName: h.MetricName,
		}
		go p.run()
	}
}

func (h *httpPing) run() error {
	// 参数校验
	h.paramCheck()

	// 数据初始化
	h.parse()

	// 执行
	_, err := h.ping()
	if err != nil {
		l.Errorf("Error: '%s'\n", err)
	}

	return nil
}

// 参数check
func (h *httpPing) paramCheck() {
	// 请求方法校验
	if strings.ToUpper(h.cfg.Method) != "GET" && strings.ToUpper(h.cfg.Method) != "POST" && strings.ToUpper(h.cfg.Method) != "HEAD" {
		l.Errorf("Error: Method '%s' not recognized.\n", h.method)
		return
	}

	// url 校验
	URL := Normalize(h.cfg.Url)
	u, err := url.Parse(URL)
	if err != nil {
		l.Errorf("Error: url '%s' not right.\n", h.cfg.Url)
		return
	}

	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		host = u.Host
	}

	ipAddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		l.Errorf("Error: cannot resolve %s: Unknown host. \n", host)
		return
	}

	h.method = h.cfg.Method
	h.method = strings.ToUpper(h.method)
	h.rAddr = ipAddr
	h.url = h.cfg.Url
	h.host = u.Host
}

func (h *httpPing) parse() {
	h.buf = h.cfg.Playload
	h.kAlive = h.cfg.KAlive
	h.tlsSkipVerify = h.cfg.TLSSkipVerify
	h.compress = h.cfg.Compress

	h.setTransport()
}

func (h *httpPing) ping() (Result, error) {
	var (
		r    Result
		resp *http.Response
		req  *http.Request
		err  error
	)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		// Timeout:   "2s",
		Transport: h.transport,
	}

	if h.method == "POST" {
		reader := strings.NewReader(h.buf)
		req, err = http.NewRequest(h.method, h.url, reader)
	} else {
		req, err = http.NewRequest(h.method, h.url, nil)
	}

	if err != nil {
		return r, err
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), tracer(&r)))
	begin := time.Now()
	resp, err = client.Do(req)
	r.Trace.totalTime = time.Since(begin)
	if err != nil {
		return r, err
	}

	defer resp.Body.Close()

	h.uploadData(r)
	return r, nil
}

func (h *httpPing) uploadData(resData Result) {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	tags["host"] = h.host
	tags["url"] = h.url
	tags["addr"] = fmt.Sprintf("%v", h.rAddr)
	fields["dnsLookupTime"] = resData.Trace.dnsLookupTime.Microseconds()
	fields["connectionTime"] = resData.Trace.connectionTime.Microseconds()
	fields["tlsHandshakeTime"] = resData.Trace.tlsHandshakeTime.Microseconds()
	fields["toFirstByteTime"] = resData.Trace.toFirstByteTime.Microseconds()
	fields["totalTime"] = resData.Trace.totalTime.Microseconds()

	pt, _ := influxdb.NewPoint(h.metricName, tags, fields, time.Now())

	io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
}

func tracer(r *Result) *httptrace.ClientTrace {
	var (
		begin             = time.Now()
		dnsStart          time.Time
		connectStart      time.Time
		tlsHandshakeStart time.Time
	)

	return &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			r.Trace.dnsLookupTime = time.Since(dnsStart)
		},
		ConnectStart: func(x, y string) {
			connectStart = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			r.Trace.connectionTime = time.Since(connectStart)
		},
		TLSHandshakeStart: func() {
			tlsHandshakeStart = time.Now()
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			r.Trace.tlsHandshakeTime = time.Since(tlsHandshakeStart)
		},
		GotConn: func(_ httptrace.GotConnInfo) {
			begin = time.Now()
		},
		GotFirstResponseByte: func() {
			r.Trace.toFirstByteTime = time.Since(begin)
		},
	}
}

// setTransport set transport
func (h *httpPing) setTransport() {
	h.transport = &http.Transport{
		DisableKeepAlives:  !h.kAlive,
		DisableCompression: h.compress,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: h.tlsSkipVerify,
		},
		Proxy: http.ProxyFromEnvironment,
	}
}

// Normalize fixes scheme
func Normalize(URL string) string {
	re := regexp.MustCompile(`(?i)https{0,1}://`)
	if !re.MatchString(URL) {
		URL = fmt.Sprintf("http://%s", URL)
	}
	return URL
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Httpstat{}
	})
}
