package datakit

import (
	"fmt"
	"net"
	"net/url"
	"net/http"
	"time"
	"bytes"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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
	DeprecatedURL    string   `toml:"url"`
	Urls             []string `toml:"urls"`
	Proxy            bool     `toml:"proxy,omitempty"`
	Timeout          string   `toml:"timeout"`
	DeprecatedHost   string   `toml:"host,omitempty"`
	DeprecatedScheme string   `toml:"scheme,omitempty"`
	DeprecatedToken  string   `toml:"token,omitempty"`
	dataWayClients   []*dataWayClient
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
		err := fmt.Errorf("category %s not exist", category)
		l.Errorf(err)
		return err
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
		fmt.Sprintf("%s; %s", datakit.Cfg.MainCfg.Hostname, git.Version))

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
			len(body), url, gz, time.Since(postbeg), string(respbody))
		return nil

	case 4:
		l.Debugf("post %d to %s failed(HTTP: %s): %s, cost %v, data dropped",
			len(body), url, resp.StatusCode, string(respbody), time.Since(postbeg))
		return nil

	case 5:
		l.Errorf("post %d to %s failed(HTTP: %s): %s, cost %v",
			len(body), url, resp.Status, string(respbody), time.Since(postbeg))
		return fmt.Errorf("dataway internal error")
	}

	return nil
}

func (dw *DataWayCfg) Send(category string, data []byte, gz bool) error {
	if dw.httpCli == nil {
		dw.httpCli = http.Client{
			Timeout: x.HTTPTimeout,
		}
	}

	for _, dc := range dw.dataWayClients {
		if err := dc.send(dw.httpCli, category, data); err != nil {
			l.Errorf("send data error %v", err)
		}
	}
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
	if len(dc.paramUrls) > 0 {
		httpaddr, err := dc.tcpaddr(dc.paramUrls[0].scheme, dc.paramUrls[0].host)
		if err != nil {
			return err
		}

		conn, err := net.DialTimeout("tcp", httpaddr, time.Second*5)
		if err != nil {
			l.Errorf("TCP dial host `%s' failed: %s", dc.paramUrls[0].host, err.Error())
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
	for _, dataWayClient := range dc.dataWayClients {
		if dataWayClient.urlValues != nil {
			token := dataWayClient.urlValues.Get("token")
			resToken = append(resToken, token)
		}
	}

	return resToken
}

func ParseDataway(httpurls []string) (*DataWayCfg, error) {
	dwcfg := &DataWayCfg{
		Timeout: "30s",
	}
	if len(httpurls) == "" {
		return nil, fmt.Errorf("empty dataway HTTP endpoint")
	}

	categorys := []string{"MetricDeprecated", "Metric", "KeyEvent", "Object", "Logging", "Tracing", "Rum", "Security", "HeartBeat"}

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
				categoryUrl := fmt.Sprintf("%s%s", baseUrl, category)
				dataWayCli.categoryUrl[category] = fmt.Sprintf("%s://%s%s?%s", paramUrl.scheme, paramUrl.host, categoryUrl, paramUrl.urlValues.Encode())
			}

			if u.Path == "/proxy" {
				l.Debugf("datakit proxied by %s", u.Host)
				dwcfg.Proxy = true

				for _, category := range categorys {
					categoryUrl := fmt.Sprintf("%s%s", baseUrl, category)
					dataWayCli.categoryUrl[category] = fmt.Sprintf("%s://%s%s?%s", paramUrl.scheme, paramUrl.host, "/proxy", categoryUrl)
				}
			} else {
				u.Path = ""
			}

			dwcfg.dataWayClients = append(dwcfg.dataWayClients, dataWayCli)
		} else {
			l.Errorf("parse url %s failed: %s", httpurl, err.Error())
			return nil, err
		}
	}

	return dwcfg, nil
}
