package jvm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
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

// Jolokia JSON request object. Example: {
//   "type": "read",
//   "mbean: "java.lang:type="Runtime",
//   "attribute": "Uptime",
//   "target": {
//     "url: "service:jmx:rmi:///jndi/rmi://target:9010/jmxrmi"
//   }
// }
type jolokiaRequest struct {
	Type      string         `json:"type"`
	Mbean     string         `json:"mbean"`
	Attribute interface{}    `json:"attribute,omitempty"`
	Path      string         `json:"path,omitempty"`
	Target    *jolokiaTarget `json:"target,omitempty"`

	name string
}

type jolokiaTarget struct {
	URL      string `json:"url"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

// Jolokia JSON response object. Example: {
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
// }
type jolokiaResponse struct {
	Request jolokiaRequest `json:"request"`
	Value   interface{}    `json:"value"`
	Status  int            `json:"status"`
}

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

func (j *JVM) createClient(url string) (*Client, error) {
	return NewClient(url, &ClientConfig{
		Username:        j.Username,
		Password:        j.Password,
		ResponseTimeout: j.ResponseTimeout,
		ClientConfig:    j.ClientConfig,
	})
}

func (c *Client) read() ([]jolokiaResponse, error) {
	requestURL, err := formatReadURL(c.URL, c.config.Username, c.config.Password)
	if err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(reqObjs)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("unable to create new request for: '%s'", c.URL)
	}

	req.Header.Add("Content-type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response from url \"%s\" has status code %d (%s), expected %d (%s)",
			c.URL, resp.StatusCode, http.StatusText(resp.StatusCode), http.StatusOK, http.StatusText(http.StatusOK))
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	l.Error(string(responseBody))
	var jResponses []jolokiaResponse
	if err = json.Unmarshal(responseBody, &jResponses); err != nil {
		return nil, fmt.Errorf("decoding JSON response: %s: %s", err, responseBody)
	}

	return jResponses, nil
}

func formatReadURL(configURL, username, password string) (string, error) {
	parsedURL, err := url.Parse(configURL)
	if err != nil {
		return "", err
	}

	readURL := url.URL{
		Host:   parsedURL.Host,
		Scheme: parsedURL.Scheme,
	}

	if username != "" || password != "" {
		readURL.User = url.UserPassword(username, password)
	}

	readURL.Path = path.Join(parsedURL.Path, "read")
	readURL.Query().Add("ignoreErrors", "true")
	return readURL.String(), nil
}
