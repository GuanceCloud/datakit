package datakit

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

type DataWayCfg struct {
	URL              string `toml:"url"`
	Proxy            bool   `toml:"proxy,omitempty"`
	Timeout          string `toml:"timeout"`
	DeprecatedHost   string `toml:"host,omitempty"`
	DeprecatedScheme string `toml:"scheme,omitempty"`
	DeprecatedToken  string `toml:"token,omitempty"`
	paramUrls        []*paramUrl
}

type paramUrl struct {
	host      string
	scheme    string
	urlValues url.Values
}

func (dc *DataWayCfg) DeprecatedMetricURL() []string {
	var resUrls []string
	if dc.Proxy {
		for _, paramUrl := range dc.paramUrls {
			itemUrl := fmt.Sprintf("%s://%s%s?%s", paramUrl.scheme, paramUrl.host, "/proxy", "category=/v1/write/metric")
			resUrls = append(resUrls, itemUrl)
		}
		return resUrls
	}

	for _, paramUrl := range dc.paramUrls {
		itemUrl := fmt.Sprintf("%s://%s%s?%s", paramUrl.scheme, paramUrl.host, "/v1/write/metrics", paramUrl.urlValues.Encode())
		resUrls = append(resUrls, itemUrl)
	}
	return resUrls
}

func (dc *DataWayCfg) MetricURL() []string {
	var resUrls []string
	if dc.Proxy {
		for _, paramUrl := range dc.paramUrls {
			itemUrl := fmt.Sprintf("%s://%s%s?%s", paramUrl.scheme, paramUrl.host, "/proxy", "category=/v1/write/metric")
			resUrls = append(resUrls, itemUrl)
		}
		return resUrls
	}

	for _, paramUrl := range dc.paramUrls {
		itemUrl := fmt.Sprintf("%s://%s%s?%s", paramUrl.scheme, paramUrl.host, "/v1/write/metric", paramUrl.urlValues.Encode())
		resUrls = append(resUrls, itemUrl)
	}
	return resUrls
}

func (dc *DataWayCfg) ObjectURL() []string {
	var resUrls []string
	if dc.Proxy {
		for _, paramUrl := range dc.paramUrls {
			itemUrl := fmt.Sprintf("%s://%s%s?%s", paramUrl.scheme, paramUrl.host, "/proxy", "category=/v1/write/object")
			resUrls = append(resUrls, itemUrl)
		}
		return resUrls
	}

	for _, paramUrl := range dc.paramUrls {
		itemUrl := fmt.Sprintf("%s://%s%s?%s", paramUrl.scheme, paramUrl.host, "/v1/write/object", paramUrl.urlValues.Encode())
		resUrls = append(resUrls, itemUrl)
	}
	return resUrls
}

func (dc *DataWayCfg) LoggingURL() []string {
	var resUrls []string
	if dc.Proxy {
		for _, paramUrl := range dc.paramUrls {
			itemUrl := fmt.Sprintf("%s://%s%s?%s", paramUrl.scheme, paramUrl.host, "/proxy", "category=/v1/write/logging")
			resUrls = append(resUrls, itemUrl)
		}
		return resUrls
	}

	for _, paramUrl := range dc.paramUrls {
		itemUrl := fmt.Sprintf("%s://%s%s?%s", paramUrl.scheme, paramUrl.host, "/v1/write/logging", paramUrl.urlValues.Encode())
		resUrls = append(resUrls, itemUrl)
	}
	return resUrls
}

func (dc *DataWayCfg) TracingURL() []string {
	var resUrls []string
	if dc.Proxy {
		for _, paramUrl := range dc.paramUrls {
			itemUrl := fmt.Sprintf("%s://%s%s?%s", paramUrl.scheme, paramUrl.host, "/proxy", "category=/v1/write/tracing")
			resUrls = append(resUrls, itemUrl)
		}
		return resUrls
	}

	for _, paramUrl := range dc.paramUrls {
		itemUrl := fmt.Sprintf("%s://%s%s?%s", paramUrl.scheme, paramUrl.host, "/v1/write/tracing", paramUrl.urlValues.Encode())
		resUrls = append(resUrls, itemUrl)
	}
	return resUrls
}

func (dc *DataWayCfg) RumURL() []string {
	var resUrls []string
	if dc.Proxy {
		for _, paramUrl := range dc.paramUrls {
			itemUrl := fmt.Sprintf("%s://%s%s?%s", paramUrl.scheme, paramUrl.host, "/proxy", "category=/v1/write/rum")
			resUrls = append(resUrls, itemUrl)
		}
		return resUrls
	}

	for _, paramUrl := range dc.paramUrls {
		itemUrl := fmt.Sprintf("%s://%s%s?%s", paramUrl.scheme, paramUrl.host, "/v1/write/rum", paramUrl.urlValues.Encode())
		resUrls = append(resUrls, itemUrl)
	}
	return resUrls
}

func (dc *DataWayCfg) SecurityURL() []string {
	var resUrls []string
	if dc.Proxy {
		for _, paramUrl := range dc.paramUrls {
			itemUrl := fmt.Sprintf("%s://%s%s?%s", paramUrl.scheme, paramUrl.host, "/proxy", "category=/v1/write/security")
			resUrls = append(resUrls, itemUrl)
		}
		return resUrls
	}

	for _, paramUrl := range dc.paramUrls {
		itemUrl := fmt.Sprintf("%s://%s%s?%s", paramUrl.scheme, paramUrl.host, "/v1/write/security", paramUrl.urlValues.Encode())
		resUrls = append(resUrls, itemUrl)
	}
	return resUrls
}

func (dc *DataWayCfg) KeyEventURL() []string {
	var resUrls []string
	if dc.Proxy {
		for _, paramUrl := range dc.paramUrls {
			itemUrl := fmt.Sprintf("%s://%s%s?%s", paramUrl.scheme, paramUrl.host, "/proxy", "category=/v1/write/keyevent")
			resUrls = append(resUrls, itemUrl)
		}
		return resUrls
	}

	for _, paramUrl := range dc.paramUrls {
		itemUrl := fmt.Sprintf("%s://%s%s?%s", paramUrl.scheme, paramUrl.host, "/v1/write/keyevent", paramUrl.urlValues.Encode())
		resUrls = append(resUrls, itemUrl)
	}
	return resUrls
}

func (dc *DataWayCfg) HeartBeatURL() []string {
	var resUrls []string
	if dc.Proxy {
		for _, paramUrl := range dc.paramUrls {
			itemUrl := fmt.Sprintf("%s://%s%s?%s", paramUrl.scheme, paramUrl.host, "/proxy", "category=/v1/write/heartbeat")
			resUrls = append(resUrls, itemUrl)
		}
		return resUrls
	}

	for _, paramUrl := range dc.paramUrls {
		itemUrl := fmt.Sprintf("%s://%s%s?%s", paramUrl.scheme, paramUrl.host, "/v1/write/heartbeat", paramUrl.urlValues.Encode())
		resUrls = append(resUrls, itemUrl)
	}
	return resUrls
}

func (dc *DataWayCfg) tcpaddr(scheme, addr string) (string, error) {
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

func (dc *DataWayCfg) Test() error {
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

func (dc *DataWayCfg) GetToken() []string {
	resToken := []string{}
	for _, paramUrl := range dc.paramUrls {
		if paramUrl.urlValues != nil {
			token := paramUrl.urlValues.Get("token")
			resToken = append(resToken, token)
		}
	}

	return resToken
}

func ParseDataway(httpurls string) (*DataWayCfg, error) {
	dwcfg := &DataWayCfg{
		Timeout: "30s",
	}
	if httpurls == "" {
		return nil, fmt.Errorf("empty dataway HTTP endpoint")
	}

	urlArr := strings.Split(httpurls, ";")

	for _, httpurl := range urlArr {
		u, err := url.Parse(httpurl)
		if err == nil {
			paramUrl := &paramUrl{}

			paramUrl.scheme = u.Scheme
			paramUrl.urlValues = u.Query()
			paramUrl.host = u.Host
			if u.Path == "/proxy" {
				l.Debugf("datakit proxied by %s", u.Host)
				dwcfg.Proxy = true
			} else {
				u.Path = ""
			}
		} else {
			l.Errorf("parse url %s failed: %s", httpurl, err.Error())
			return nil, err
		}
		dwcfg.URL = u.String()
	}

	return dwcfg, nil
}
