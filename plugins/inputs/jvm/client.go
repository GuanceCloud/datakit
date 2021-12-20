package jvm

import (
	"net/http"
	"time"

	"github.com/influxdata/telegraf/plugins/common/tls"
)

type Client struct {
	URL    string
	client *http.Client
	config *ClientConfig
}

type ClientConfig struct {
	ResponseTimeout time.Duration
	Username        string
	Password        string
	ProxyConfig     *ProxyConfig
	tls.ClientConfig
}

type ProxyConfig struct {
	DefaultTargetUsername string
	DefaultTargetPassword string
	Targets               []ProxyTargetConfig
}

type ProxyTargetConfig struct {
	Username string
	Password string
	URL      string
}

type ReadRequest struct {
	Mbean      string
	Attributes []string
	Path       string
}

type ReadResponse struct {
	Status            int
	Value             interface{}
	RequestMbean      string
	RequestAttributes []string
	RequestPath       string
	RequestTarget     string
}

// Jolokia JSON response object. Example
//  {
//   "request": {
//     "type": "read"
//     "mbean": "java.lang:type=Runtime",
//     "attribute": "Uptime",
//     "target": {
//       "url": "service:jmx:rmi:///jndi/rmi://target:9010/jmxrmi"
//     }
//   },
//   "value": 1214083,
//   "timestamp": 1488059309,
//   "status": 200
// }.

func NewClient(url string, config *ClientConfig) (*Client, error) {
	tlsConfig, err := config.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		ResponseHeaderTimeout: config.ResponseTimeout,
		TLSClientConfig:       tlsConfig,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.ResponseTimeout,
	}

	return &Client{
		URL:    url,
		config: config,
		client: client,
	}, nil
}
