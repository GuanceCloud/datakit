package dataway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/tracer"
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
	}

	ExtraHeaders = map[string]string{}
	l            = logger.DefaultSLogger("dataway")
)

type DataWayCfg struct {
	DeprecatedURL string   `toml:"url,omitempty"`
	URLs          []string `toml:"urls"`

	DeprecatedHost   string `toml:"host,omitempty"`
	DeprecatedScheme string `toml:"scheme,omitempty"`
	DeprecatedToken  string `toml:"token,omitempty"`

	HTTPTimeout     string        `toml:"timeout"`
	TimeoutDuration time.Duration `toml:"-"`

	Proxy     bool   `toml:"proxy,omitempty"`
	HttpProxy string `toml:"http_proxy"`

	dataWayClients []*dataWayClient
	httpCli        *http.Client

	Hostname string `toml:"-"`

	ontest bool
}

type Option func(cnf *DataWayCfg)

type dataWayClient struct {
	url         string
	host        string
	scheme      string
	urlValues   url.Values
	categoryURL map[string]string
	ontest      bool
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
	tracer.GlobalTracer.Start(tracer.WithLogger(tracer.DDLog{}))
	defer tracer.GlobalTracer.Stop()

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

	span := tracer.GlobalTracer.StartSpan("send")
	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(data))
	if err != nil {
		l.Error(err)
		tracer.GlobalTracer.FinishSpan(span, tracer.WithError(err))

		return err
	}
	tracer.GlobalTracer.Inject(span, req.Header)

	if gz {
		req.Header.Set("Content-Encoding", "gzip")
	}

	// append extra headers
	for k, v := range ExtraHeaders {
		req.Header.Set(k, v)
	}

	postbeg := time.Now()

	l.Debugf("request %s", requrl)
	if dc.ontest {
		return nil
	}

	resp, err := cli.Do(req)
	if err != nil {
		l.Errorf("request url %s failed: %s", requrl, err)
		return err
	}

	defer resp.Body.Close()
	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Error(err)
		return err
	}

	switch resp.StatusCode / 100 {
	case 2:
		l.Debugf("post %d to %s ok(gz: %v), cost %v, response: %s",
			len(data), requrl, gz, time.Since(postbeg), string(respbody))
		return nil

	case 4:
		l.Warnf("post %d to %s failed(HTTP: %s): %s, cost %v, data dropped",
			len(data),
			requrl,
			resp.Status,
			string(respbody),
			time.Since(postbeg))
		return nil

	case 5:
		l.Errorf("post %d to %s failed(HTTP: %s): %s, cost %v",
			len(data),
			requrl,
			resp.Status,
			string(respbody),
			time.Since(postbeg))
		return fmt.Errorf("dataway internal error")
	}

	return nil
}

// func (dc *dataWayClient) sendWithTracing(cli *http.Client, category string, data []byte, gz bool) error {
// 	l.Info("send data with tracing")

// 	tracer.GlobalTracer.Start(tracer.WithLogger(tracer.DDLog{}))
// 	defer tracer.GlobalTracer.Stop()

// 	span := tracer.GlobalTracer.StartSpan("send")
// 	err := dc.send(cli, category, data, gz)
// 	span.Finish(tracer.WithFinishTime(time.Now()), tracer.WithError(err))

// 	return err
// }

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
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
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

func (dw *DataWayCfg) Election(id string) ([]byte, error) {
	if len(dw.dataWayClients) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.dataWayClients[0] // 选举相关接口只只发送给第一个 dataway

	requrl, ok := dc.categoryURL[datakit.Election]
	if !ok {
		return nil, fmt.Errorf("no election URL available")
	}

	if strings.Contains(requrl, "?token") {
		requrl += ("&id=" + id)
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

func (dw *DataWayCfg) ElectionHeartbeat(id string) ([]byte, error) {
	if len(dw.dataWayClients) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.dataWayClients[0] // 选举相关接口只只发送给第一个 dataway

	requrl, ok := dc.categoryURL[datakit.ElectionHeartbeat]
	if !ok {
		return nil, fmt.Errorf("no election URL available")
	}

	if strings.Contains(requrl, "?token") {
		requrl += ("&id=" + id)
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
		l.Debugf("send to %dth dataway", i)
		if err := dc.send(dw.httpCli, category, data, gz); err != nil {
			return err
		}
		// if tracer.GlobalTracer != nil {
		// 	if err := dc.sendWithTracing(dw.httpCli, category, data, gz); err != nil {
		// 		return err
		// 	}
		// } else {
		// 	if err := dc.send(dw.httpCli, category, data, gz); err != nil {
		// 		return err
		// 	}
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

	timeout, err := time.ParseDuration(dw.HTTPTimeout)
	if err != nil {
		return err
	}

	dw.TimeoutDuration = timeout

	if err := dw.initHttp(); err != nil {
		return err
	}

	for _, httpurl := range dw.URLs {
		u, err := url.ParseRequestURI(httpurl)
		if err != nil {
			l.Errorf("parse dataway url %s failed: %s", httpurl, err.Error())
			return err
		}

		cli := &dataWayClient{
			url:         httpurl,
			scheme:      u.Scheme,
			urlValues:   u.Query(),
			host:        u.Host,
			categoryURL: map[string]string{},
			ontest:      dw.ontest,
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

		dw.dataWayClients = append(dw.dataWayClients, cli)
	}

	return nil
}

func (dw *DataWayCfg) initHttp() error {
	dw.httpCli = &http.Client{
		Timeout: dw.TimeoutDuration,
	}

	if dw.HttpProxy != "" {
		uri, err := url.ParseRequestURI(dw.HttpProxy)
		if err != nil {
			l.Error("parse url error: ", err)
			return err
		}

		tr := &http.Transport{
			Proxy: http.ProxyURL(uri),
		}

		dw.httpCli.Transport = tr
	}

	return nil
}
