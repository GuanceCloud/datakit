package datakit

import (
	"fmt"
	"net"
	"net/url"
	"time"
)

type DataWayCfg struct {
	URL     string `toml:"url"`
	Proxy   bool   `toml:"proxy,omitempty"`
	Timeout string `toml:"timeout"`

	DeprecatedHost   string `toml:"host,omitempty"`
	DeprecatedScheme string `toml:"scheme,omitempty"`
	DeprecatedToken  string `toml:"token,omitempty"`

	host      string
	scheme    string
	urlValues url.Values
}

func (dc *DataWayCfg) DeprecatedMetricURL() string {
	if dc.Proxy {
		return fmt.Sprintf("%s://%s%s?%s",
			dc.scheme,
			dc.host,
			"/proxy",
			"category=/v1/write/metric")
	}

	return fmt.Sprintf("%s://%s%s?%s",
		dc.scheme,
		dc.host,
		"/v1/write/metrics",
		dc.urlValues.Encode())
}

func (dc *DataWayCfg) MetricURL() string {

	if dc.Proxy {
		return fmt.Sprintf("%s://%s%s?%s",
			dc.scheme,
			dc.host,
			"/proxy",
			"category=/v1/write/metric")
	}

	return fmt.Sprintf("%s://%s%s?%s",
		dc.scheme,
		dc.host,
		"/v1/write/metric",
		dc.urlValues.Encode())
}

func (dc *DataWayCfg) ObjectURL() string {

	if dc.Proxy {
		return fmt.Sprintf("%s://%s%s?%s",
			dc.scheme,
			dc.host,
			"/proxy",
			"category=/v1/write/object")
	}

	return fmt.Sprintf("%s://%s%s?%s",
		dc.scheme,
		dc.host,
		"/v1/write/object",
		dc.urlValues.Encode())
}

func (dc *DataWayCfg) LoggingURL() string {

	if dc.Proxy {
		return fmt.Sprintf("%s://%s%s?%s",
			dc.scheme,
			dc.host,
			"/proxy",
			"category=/v1/write/logging")
	}

	return fmt.Sprintf("%s://%s%s?%s",
		dc.scheme,
		dc.host,
		"/v1/write/logging",
		dc.urlValues.Encode())
}

func (dc *DataWayCfg) TracingURL() string {
	if dc.Proxy {
		return fmt.Sprintf("%s://%s%s?%s",
			dc.scheme,
			dc.host,
			"/proxy",
			"category=/v1/write/tracing")
	}

	return fmt.Sprintf("%s://%s%s?%s",
		dc.scheme,
		dc.host,
		"/v1/write/tracing",
		dc.urlValues.Encode())
}

func (dc *DataWayCfg) RumURL() string {
	if dc.Proxy {
		return fmt.Sprintf("%s://%s%s?%s",
			dc.scheme,
			dc.host,
			"/proxy",
			"category=/v1/write/rum")
	}

	return fmt.Sprintf("%s://%s%s?%s",
		dc.scheme,
		dc.host,
		"/v1/write/rum",
		dc.urlValues.Encode())
}

func (dc *DataWayCfg) SecurityURL() string {
	if dc.Proxy {
		return fmt.Sprintf("%s://%s%s?%s",
			dc.scheme,
			dc.host,
			"/proxy",
			"category=/v1/write/security")
	}

	return fmt.Sprintf("%s://%s%s?%s",
		dc.scheme,
		dc.host,
		"/v1/write/security",
		dc.urlValues.Encode())
}

func (dc *DataWayCfg) KeyEventURL() string {

	if dc.Proxy {
		return fmt.Sprintf("%s://%s%s?%s",
			dc.scheme,
			dc.host,
			"/proxy",
			"category=/v1/write/keyevent")
	}

	return fmt.Sprintf("%s://%s%s?%s",
		dc.scheme,
		dc.host,
		"/v1/write/keyevent",
		dc.urlValues.Encode())
}

func (dc *DataWayCfg) HeartBeatURL() string {
	if dc.Proxy {
		return fmt.Sprintf("%s://%s%s?%s",
			dc.scheme,
			dc.host,
			"/proxy",
			"category=/v1/write/heartbeat")
	}

	return fmt.Sprintf("%s://%s%s?%s",
		dc.scheme,
		dc.host,
		"/v1/write/heartbeat",
		dc.urlValues.Encode())
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

	httpaddr, err := dc.tcpaddr(dc.scheme, dc.host)
	if err != nil {
		return err
	}

	conn, err := net.DialTimeout("tcp", httpaddr, time.Second*5)
	if err != nil {
		l.Errorf("TCP dial host `%s' failed: %s", dc.host, err.Error())
		return err
	}

	if err := conn.Close(); err != nil {
		l.Errorf("Close(): %s, ignored", err.Error())
	}

	return nil
}

func (dc *DataWayCfg) addToken(tkn string) {
	if dc.urlValues == nil {
		dc.urlValues = url.Values{}
	}

	if dc.urlValues.Get("token") == "" {
		l.Debugf("use old token %s", dc.DeprecatedToken)
		dc.urlValues.Set("token", dc.DeprecatedToken)
	}
}

func (dc *DataWayCfg) GetToken() string {
	if dc.urlValues == nil {
		dc.addToken("")
	}
	return dc.urlValues.Get("token")
}

func ParseDataway(httpurl string) (*DataWayCfg, error) {
	dwcfg := &DataWayCfg{
		Timeout: "30s",
	}
	if httpurl == "" {
		return nil, fmt.Errorf("empty dataway HTTP endpoint")
	}
	u, err := url.Parse(httpurl)
	if err == nil {
		dwcfg.scheme = u.Scheme
		dwcfg.urlValues = u.Query()
		dwcfg.host = u.Host
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

	return dwcfg, nil
}
