package datakit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	baseUrl          = "/v1/write"
	MetricDeprecated = "metrics"
	Metric           = "metric"
	KeyEvent         = "keyevent"
	Object           = "object"
	Logging          = "logging"
	Tracing          = "tracing"
	Rum              = "rum"
	Security         = "security"
	HeartBeat        = "heartbeat"
)

type DataWayCfg struct {
	DeprecatedURL    string   `toml:"url,omitempty"`
	Urls             []string `toml:"urls"`
	Proxy            bool     `toml:"proxy,omitempty"`
	DeprecatedHost   string   `toml:"host,omitempty"`
	DeprecatedScheme string   `toml:"scheme,omitempty"`
	DeprecatedToken  string   `toml:"token,omitempty"`
	dataWayClients   []*dataWayClient
	httpCli          *http.Client
	HTTPTimeout      string        `toml:"timeout"`
	TimeoutDuration  time.Duration `toml:"-"`
	HttpProxy        string        `toml:"http_proxy"`
}

type dataWayClient struct {
	url         string
	host        string
	scheme      string
	urlValues   url.Values
	categoryUrl map[string]string
	httpCli     *http.Client
}

// 发送数据
func (dc *dataWayClient) send(cli *http.Client, category string, data []byte, gz bool) error {
	url, ok := dc.categoryUrl[category]
	if !ok {
		url = category
		// err := fmt.Errorf("category %s not exist", category)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		l.Error(err)
		return err
	}

	if gz {
		req.Header.Set("Content-Encoding", "gzip")
	}

	// append datakit info
	req.Header.Set("X-Datakit-Info",
		fmt.Sprintf("%s; %s", Cfg.Hostname, git.Version))

	postbeg := time.Now()

	dc.httpCli = cli

	resp, err := dc.httpCli.Do(req)
	if err != nil {
		l.Errorf("request url %s failed: %s", url, err)
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
			len(data), url, gz, time.Since(postbeg), string(respbody))
		return nil

	case 4:
		l.Debugf("post %d to %s failed(HTTP: %s): %s, cost %v, data dropped",
			len(data), url, resp.StatusCode, string(respbody), time.Since(postbeg))
		return nil

	case 5:
		l.Errorf("post %d to %s failed(HTTP: %s): %s, cost %v",
			len(data), url, resp.Status, string(respbody), time.Since(postbeg))
		return fmt.Errorf("dataway internal error")
	}

	return nil
}

func (dc *dataWayClient) heartBeat(cli *http.Client, data []byte) error {
	req, err := http.NewRequest("POST", dc.categoryUrl[HeartBeat], bytes.NewBuffer(data))

	dc.httpCli = cli

	resp, err := dc.httpCli.Do(req)
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

func (dw *DataWayCfg) Send(category string, data []byte, gz bool) error {
	if dw.httpCli != nil {
		defer dw.httpCli.CloseIdleConnections()
	}

	dw.initHttp()

	for idx, dc := range dw.dataWayClients {
		l.Debugf("post to %d dataway...", idx)
		if err := dc.send(dw.httpCli, category, data, gz); err != nil {
			return err
		}
	}

	return nil
}

func (dw *DataWayCfg) HeartBeat() error {
	if dw.httpCli != nil {
		defer dw.httpCli.CloseIdleConnections()
	}

	body := map[string]interface{}{
		"dk_uuid":   Cfg.UUID,
		"heartbeat": time.Now().Unix(),
		"host":      Cfg.Hostname,
	}

	dw.initHttp()

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

func (dw *DataWayCfg) ElectionURL() []string {
	var resUrl []string
	for _, dc := range dw.dataWayClients {
		electionUrl := dc.categoryUrl["electionUrl"]
		resUrl = append(resUrl, electionUrl)
	}

	return resUrl
}

func (dw *DataWayCfg) ElectionHeartBeatURL() []string {
	var resUrl []string
	for _, dc := range dw.dataWayClients {
		electionBeatUrl := dc.categoryUrl["electionBeatUrl"]
		resUrl = append(resUrl, electionBeatUrl)
	}

	return resUrl
}

func (dw *DataWayCfg) tcpaddr(scheme, addr string) (string, error) {
	tcpaddr := addr
	if _, _, err := net.SplitHostPort(tcpaddr); err != nil {
		switch scheme {
		case "http":
			tcpaddr += ":80"
		case "https":
			tcpaddr += ":443"
		}

		if _, _, err := net.SplitHostPort(tcpaddr); err != nil {
			l.Errorf("net.SplitHostPort(): %s", err)
			return "", err
		}
	}

	return tcpaddr, nil
}

func (dw *DataWayCfg) Test() error {
	if len(dw.dataWayClients) > 0 {
		httpaddr, err := dw.tcpaddr(dw.dataWayClients[0].scheme, dw.dataWayClients[0].host)
		if err != nil {
			return err
		}

		conn, err := net.DialTimeout("tcp", httpaddr, time.Second*5)
		if err != nil {
			l.Errorf("TCP dial host `%s' failed: %s", dw.dataWayClients[0].host, err.Error())
			return err
		}

		if err := conn.Close(); err != nil {
			l.Errorf("Close(): %s, ignored", err.Error())
		}
	}
	return nil
}

func (dw *DataWayCfg) GetToken() []string {
	resToken := []string{}
	for _, dataWayClient := range dw.dataWayClients {
		if dataWayClient.urlValues != nil {
			token := dataWayClient.urlValues.Get("token")
			resToken = append(resToken, token)
		}
	}

	return resToken
}

func ParseDataway(httpurls []string) (*DataWayCfg, error) {
	dw := Cfg.DataWay

	if dw.HTTPTimeout == "" {
		dw.HTTPTimeout = "5s"
	}

	timeout, err := time.ParseDuration(dw.HTTPTimeout)
	if err != nil {
		return nil, err
	}

	dw.TimeoutDuration = timeout

	dw.initHttp()

	if len(httpurls) == 0 {
		return nil, fmt.Errorf("empty dataway HTTP endpoint")
	}

	categorys := []string{MetricDeprecated, Metric, KeyEvent, Object, Logging, Tracing, Rum, Security, HeartBeat}

	for _, httpurl := range httpurls {
		u, err := url.Parse(httpurl)
		if err == nil {
			dataWayCli := &dataWayClient{}

			dataWayCli.url = httpurl
			dataWayCli.scheme = u.Scheme
			dataWayCli.urlValues = u.Query()
			dataWayCli.host = u.Host
			dataWayCli.categoryUrl = make(map[string]string)

			for _, category := range categorys {
				categoryUrl := fmt.Sprintf("%s/%s", baseUrl, category)
				dataWayCli.categoryUrl[category] = fmt.Sprintf("%s://%s%s?%s", dataWayCli.scheme, dataWayCli.host, categoryUrl, dataWayCli.urlValues.Encode())
			}

			// election
			electionUrl := fmt.Sprintf("%s://%s%s?%s", dataWayCli.scheme, dataWayCli.host, "/v1/election", dataWayCli.urlValues.Encode())
			electionBeatUrl := fmt.Sprintf("%s://%s%s?%s", dataWayCli.scheme, dataWayCli.host, "/v1/election/heartbeat", dataWayCli.urlValues.Encode())
			dataWayCli.categoryUrl["electionUrl"] = electionUrl
			dataWayCli.categoryUrl["electionBeatUrl"] = electionBeatUrl

			dw.dataWayClients = append(dw.dataWayClients, dataWayCli)
		} else {
			l.Errorf("parse url %s failed: %s", httpurl, err.Error())
			return nil, err
		}
	}

	return dw, nil
}

func (dw *DataWayCfg) initHttp() {
	if dw.httpCli == nil {
		dw.httpCli = &http.Client{
			Timeout: dw.TimeoutDuration,
		}

		if dw.HttpProxy != "" {
			uri, err := url.Parse(dw.HttpProxy)
			if err != nil {
				l.Error("parse url error: ", err)
			}

			tr := &http.Transport{
				Proxy: http.ProxyURL(uri),
			}

			dw.httpCli.Transport = tr
		}
	}
}
