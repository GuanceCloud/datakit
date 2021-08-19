package dataway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/ddtrace/tracer"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
	dktracer "gitlab.jiagouyun.com/cloudcare-tools/datakit/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

var (
	apis = []string{
		datakit.MetricDeprecated,
		datakit.Metric,
		datakit.KeyEvent,
		datakit.Object,
		datakit.CustomObject,
		datakit.Logging,
		datakit.LogFilter,
		datakit.Tracing,
		datakit.Rum,
		datakit.Security,
		datakit.HeartBeat,
		datakit.Election,
		datakit.ElectionHeartbeat,
		datakit.QueryRaw,
		datakit.ListDataWay,
		datakit.ObjectLabel,
	}

	ExtraHeaders      = map[string]string{}
	AvailableDataways = []string{}
	l                 = logger.DefaultSLogger("dataway")
)

type DataWayCfg struct {
	DeprecatedURL string   `toml:"url,omitempty"`
	URLs          []string `toml:"urls"`

	HTTPTimeout     string        `toml:"timeout"`
	TimeoutDuration time.Duration `toml:"-"`

	Proxy     bool   `toml:"proxy,omitempty"`
	HttpProxy string `toml:"http_proxy"`

	dataWayClients []*dataWayClient
	httpCli        *http.Client

	Hostname string `toml:"-"`

	MaxFails int `toml:"max_fail"`
	ontest   bool

	DeprecatedHost   string `toml:"host,omitempty"`
	DeprecatedScheme string `toml:"scheme,omitempty"`
	DeprecatedToken  string `toml:"token,omitempty"`
}

type Option func(cnf *DataWayCfg)

type dataWayClient struct {
	url         string
	host        string
	scheme      string
	proxy       string
	urlValues   url.Values
	categoryURL map[string]string
	ontest      bool
	fails       int
}

func (dw *DataWayCfg) String() string {
	arr := []string{fmt.Sprintf("dataways: [%s]", strings.Join(dw.URLs, ","))}

	for _, x := range dw.dataWayClients {
		arr = append(arr, "---------------------------------")
		for k, v := range x.categoryURL {
			arr = append(arr, fmt.Sprintf("% 24s: %s", k, v))
		}
	}

	return strings.Join(arr, "\n")
}

func (dc *dataWayClient) send(cli *http.Client, category string, data []byte, gz bool) error {
	dktracer.GlobalTracer.Start(tracer.WithLogger(&tracer.SimpleLogger{}))
	defer dktracer.GlobalTracer.Stop()

	requrl, ok := dc.categoryURL[category]
	if !ok {
		// for dialtesting, there are user-defined url to post
		if x, err := url.ParseRequestURI(category); err != nil {
			l.Error(err)

			return fmt.Errorf("invalid url %s", category)
		} else {
			l.Debugf("try use URL %+#v", x)
			requrl = category
		}
	}
	l.Debugf("request %s", requrl)

	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(data))
	if err != nil {
		l.Error(err)

		return err
	}

	if gz {
		req.Header.Set("Content-Encoding", "gzip")
	}
	// append extra headers
	for k, v := range ExtraHeaders {
		req.Header.Set(k, v)
	}

	if dc.ontest {
		l.Debug("Datakit client on test")

		return nil
	}

	// start trace span from request context
	span, _ := dktracer.GlobalTracer.StartSpanFromContext(req.Context(), "datakit.dataway.send", req.RequestURI, ext.SpanTypeHTTP)
	defer dktracer.GlobalTracer.FinishSpan(span, tracer.WithFinishTime(time.Now()))

	// inject span into http header
	dktracer.GlobalTracer.Inject(span, req.Header)

	resp, err := cli.Do(req)
	if err != nil {
		dktracer.GlobalTracer.SetTag(span, "http_client_do_error", err.Error())
		l.Errorf("request url %s failed(proxy: %s): %s", requrl, dc.proxy, err)
		dc.fails++

		return err
	}
	defer resp.Body.Close()

	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		dktracer.GlobalTracer.SetTag(span, "io_read_request_body_error", err.Error())
		l.Error(err)

		return err
	}

	postbeg := time.Now()
	switch resp.StatusCode / 100 {
	case 2:
		dc.fails = 0
		l.Debugf("post %d to %s ok(gz: %v), cost %v, response: %s",
			len(data), requrl, gz, time.Since(postbeg), string(respbody))
		return nil

	case 4:
		dc.fails = 0
		dktracer.GlobalTracer.SetTag(span, "http_request_400_error", fmt.Errorf("%d: %s", resp.StatusCode, http.StatusText(resp.StatusCode)))
		l.Warnf("post %d to %s failed(HTTP: %s): %s, cost %v, data dropped",
			len(data),
			requrl,
			resp.Status,
			string(respbody),
			time.Since(postbeg))
		return nil

	case 5:
		dc.fails++
		dktracer.GlobalTracer.SetTag(span, "http_request_500_error", fmt.Errorf("%d: %s", resp.StatusCode, http.StatusText(resp.StatusCode)))
		l.Errorf("[%d] post %d to %s failed(HTTP: %s): %s, cost %v", dc.fails,
			len(data),
			requrl,
			resp.Status,
			string(respbody),
			time.Since(postbeg))
		return fmt.Errorf("dataway internal error")
	}

	return nil
}

func (dc *dataWayClient) getLogFilter(cli *http.Client) ([]byte, error) {
	url, ok := dc.categoryURL[datakit.LogFilter]
	if !ok {
		return nil, fmt.Errorf("LogFilter API missing, should not been here")
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getLogFilter failed with status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (dc *dataWayClient) run(method, tp string, cli *http.Client) ([]byte, error) {
	url, ok := dc.categoryURL[tp]
	if !ok {
		return nil, fmt.Errorf(" %s API missing, should not been here", tp)
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	switch resp.StatusCode / 100 {
	case 2:
		l.Debugf(" %s ok", url)
		return body, nil
	default:
		return nil, fmt.Errorf("%s: %s", tp, string(body))
	}
}

func (dc *dataWayClient) heartBeat(cli *http.Client, data []byte) error {
	requrl, ok := dc.categoryURL[datakit.HeartBeat]
	if !ok {
		return fmt.Errorf("HeartBeat API missing, should not been here")
	}

	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(data))

	if dc.ontest {
		return nil
	}

	resp, err := cli.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		err := fmt.Errorf("heart beat resp err: %+#v", resp)
		return err
	}

	return nil
}

func (dw *DataWayCfg) DQLQuery(body []byte) (*http.Response, error) {
	if len(dw.dataWayClients) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.dataWayClients[0]
	requrl, ok := dc.categoryURL[datakit.QueryRaw]
	if !ok {
		return nil, fmt.Errorf("no DQL query URL available")
	}

	defer dw.httpCli.CloseIdleConnections()
	return dw.httpCli.Post(requrl, "application/json", bytes.NewBuffer(body))
}

func (dw *DataWayCfg) Election(namespace, id string) ([]byte, error) {
	if len(dw.dataWayClients) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.dataWayClients[0] // 选举相关接口只只发送给第一个 dataway

	requrl, ok := dc.categoryURL[datakit.Election]
	if !ok {
		return nil, fmt.Errorf("no election URL available")
	}

	if strings.Contains(requrl, "?token") {
		requrl += fmt.Sprintf("&namespace=%s&id=%s", namespace, id)
	} else {
		return nil, fmt.Errorf("token missing")
	}

	defer dw.httpCli.CloseIdleConnections()

	l.Debugf("election sending %s", requrl)
	resp, err := dw.httpCli.Post(requrl, "", nil)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	defer resp.Body.Close()
	switch resp.StatusCode / 100 {
	case 2:
		l.Debugf("election %s ok", requrl)
		return body, nil
	default:
		l.Debugf("election failed: %d", resp.StatusCode)
		return nil, fmt.Errorf("election failed: %s", string(body))
	}
}

func (dw *DataWayCfg) ElectionHeartbeat(namespace, id string) ([]byte, error) {
	if len(dw.dataWayClients) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.dataWayClients[0] // 选举相关接口只只发送给第一个 dataway

	requrl, ok := dc.categoryURL[datakit.ElectionHeartbeat]
	if !ok {
		return nil, fmt.Errorf("no election URL available")
	}

	if strings.Contains(requrl, "?token") {
		requrl += fmt.Sprintf("&namespace=%s&id=%s", namespace, id)
	} else {
		return nil, fmt.Errorf("token missing")
	}

	defer dw.httpCli.CloseIdleConnections()

	l.Debugf("election sending heartbeat %s", requrl)
	resp, err := dw.httpCli.Post(requrl, "", nil)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	defer resp.Body.Close()
	switch resp.StatusCode / 100 {
	case 2:
		return body, nil
	default:
		return nil, fmt.Errorf("election heartbeat failed: %s", string(body))
	}
}

func (dw *DataWayCfg) Send(category string, data []byte, gz bool) error {

	defer dw.httpCli.CloseIdleConnections()

	for i, dc := range dw.dataWayClients {
		l.Debugf("send to %dth dataway, %d:%d", i, dc.fails, dw.MaxFails)
		// 判断fails
		if dc.fails > dw.MaxFails && len(AvailableDataways) > 0 {
			rand.Seed(time.Now().UnixNano())
			index := rand.Intn(len(AvailableDataways))

			var err error
			url := fmt.Sprintf(`%s?%s`, AvailableDataways[index], dc.urlValues.Encode())
			dc, err = dw.initDatawayCli(url)
			if err != nil {
				l.Error(err)
				return err
			}

			dw.dataWayClients[i] = dc
		}

		if err := dc.send(dw.httpCli, category, data, gz); err != nil {
			return err
		}
	}

	return nil
}

func (dw *DataWayCfg) ClientsCount() int {
	return len(dw.dataWayClients)
}

func (dw *DataWayCfg) GetLogFilter() ([]byte, error) {
	if dw.httpCli != nil {
		defer dw.httpCli.CloseIdleConnections()
	}

	if len(dw.dataWayClients) == 0 {
		return nil, fmt.Errorf("[error] dataway url empty")
	}

	return dw.dataWayClients[0].getLogFilter(dw.httpCli)
}

type dataways struct {
	Content []string `json:"content"`
}

func (dw *DataWayCfg) DatawayList() error {
	if dw.httpCli != nil {
		defer dw.httpCli.CloseIdleConnections()
	}

	if len(dw.dataWayClients) == 0 {
		l.Errorf(`dataway url empty`)
		return fmt.Errorf("[error] dataway url empty")
	}

	if dw.httpCli == nil {
		if err := dw.initHttp(); err != nil {
			return err
		}
	}

	res, err := dw.dataWayClients[0].run(`GET`, datakit.ListDataWay, dw.httpCli)
	if err != nil {
		l.Error(err)
		return err
	}

	var dws dataways
	err = json.Unmarshal(res, &dws)
	if err != nil {
		l.Errorf(`%s, body: %s`, err, res)
		return err
	}

	AvailableDataways = dws.Content

	l.Debugf(`avaliable dataways; %+#v`, AvailableDataways)
	return nil
}

func (dw *DataWayCfg) HeartBeat() error {
	if dw.httpCli != nil {
		defer dw.httpCli.CloseIdleConnections()
	}

	body := map[string]interface{}{
		"dk_uuid":   dw.Hostname, // 暂用 hostname 代之, 后将弃用该字段
		"heartbeat": time.Now().Unix(),
		"host":      dw.Hostname,
	}

	if dw.httpCli == nil {
		if err := dw.initHttp(); err != nil {
			return err
		}
	}

	bodyByte, err := json.Marshal(body)
	if err != nil {
		err := fmt.Errorf("[error] heartbeat json marshal err:%s", err.Error())
		return err
	}

	for _, dc := range dw.dataWayClients {
		if err := dc.heartBeat(dw.httpCli, bodyByte); err != nil {
			l.Errorf("heart beat send data error %v", err)
		}
	}

	return nil
}

func (dw *DataWayCfg) GetToken() []string {
	resToken := []string{}
	for _, dataWayClient := range dw.dataWayClients {
		if dataWayClient.urlValues != nil {
			token := dataWayClient.urlValues.Get("token")
			if token != "" {
				resToken = append(resToken, token)
			}
		}
	}

	return resToken
}

func (dw *DataWayCfg) Apply() error {
	l = logger.SLogger("dataway")

	// 如果 env 已传入了 dataway 配置, 则不再追加老的 dataway 配置,
	// 避免俩边配置了同样的 dataway, 造成数据混乱
	if dw.DeprecatedURL != "" && len(dw.URLs) == 0 {
		dw.URLs = []string{dw.DeprecatedURL}
	}

	if len(dw.URLs) == 0 {
		return fmt.Errorf("dataway not set")
	}

	if dw.HTTPTimeout == "" {
		dw.HTTPTimeout = "5s"
	}

	if dw.MaxFails == 0 {
		dw.MaxFails = 20
	}

	timeout, err := time.ParseDuration(dw.HTTPTimeout)
	if err != nil {
		return err
	}

	dw.TimeoutDuration = timeout

	if err := dw.initHttp(); err != nil {
		return err
	}

	dw.dataWayClients = dw.dataWayClients[:0]

	for _, httpurl := range dw.URLs {
		cli, err := dw.initDatawayCli(httpurl)
		if err != nil {
			l.Errorf("init dataway url %s failed: %s", httpurl, err.Error())
			return err
		}

		dw.dataWayClients = append(dw.dataWayClients, cli)
	}

	return nil
}

func (dw *DataWayCfg) initDatawayCli(httpurl string) (*dataWayClient, error) {
	u, err := url.ParseRequestURI(httpurl)
	if err != nil {
		l.Errorf("parse dataway url %s failed: %s", httpurl, err.Error())
		return nil, err
	}

	cli := &dataWayClient{
		url:         httpurl,
		scheme:      u.Scheme,
		urlValues:   u.Query(),
		host:        u.Host,
		categoryURL: map[string]string{},
		ontest:      dw.ontest,
		proxy:       dw.HttpProxy,
	}

	for _, api := range apis {
		if cli.urlValues.Encode() != "" {
			cli.categoryURL[api] = fmt.Sprintf("%s://%s%s?%s",
				cli.scheme,
				cli.host,
				api,
				cli.urlValues.Encode())
		} else {
			cli.categoryURL[api] = fmt.Sprintf("%s://%s%s",
				cli.scheme,
				cli.host,
				api)
		}
	}

	return cli, nil
}

var proxyOnce sync.Once

func (dw *DataWayCfg) initHttp() error {
	proxyOnce.Do(func() {
		if dw.HttpProxy != "" {
			if pxurl, err := url.ParseRequestURI(dw.HttpProxy); err != nil {
				l.Errorf("parse http proxy failed err:", err.Error())
			} else {
				ihttp.DefTransport.Proxy = http.ProxyURL(pxurl)
			}
		}
	})

	dw.httpCli = &http.Client{
		Transport: ihttp.DefTransport,
		Timeout:   dw.TimeoutDuration,
	}

	return nil
}

// UpsertObjectLabels , dw api create or update object labels
func (dw *DataWayCfg) UpsertObjectLabels(tkn string, body []byte) (*http.Response, error) {
	if len(dw.dataWayClients) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.dataWayClients[0]
	requrl, ok := dc.categoryURL[datakit.ObjectLabel]
	if !ok {
		return nil, fmt.Errorf("no object labels URL available")
	}

	defer dw.httpCli.CloseIdleConnections()
	return dw.httpCli.Post(requrl, "application/json", bytes.NewBuffer(body))
}

// DeleteObjectLabels , dw api delete object labels
func (dw *DataWayCfg) DeleteObjectLabels(tkn string, body []byte) (*http.Response, error) {
	if len(dw.dataWayClients) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.dataWayClients[0]
	requrl, ok := dc.categoryURL[datakit.ObjectLabel]
	if !ok {
		return nil, fmt.Errorf("no object labels URL available")
	}

	defer dw.httpCli.CloseIdleConnections()
	rBody := bytes.NewReader(body)
	req, err := http.NewRequest("DELETE", requrl, rBody)
	if err != nil {
		return nil, fmt.Errorf("delete object label error: %s", err.Error())
	}
	return dw.httpCli.Do(req)
}
