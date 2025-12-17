// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package jolokia provides a client for interacting with Jolokia JMX HTTP bridge.
package jolokia

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/telegraf/plugins/common/tls"
)

type Client struct {
	url        string
	httpClient *http.Client
	config     *Config
}

// Config represents the client configuration.
type Config struct {
	URL             string           // Jolokia server URL, e.g., "http://localhost:8080/jolokia"
	Username        string           // Basic auth username (optional)
	Password        string           // Basic auth password (optional)
	ResponseTimeout time.Duration    // Response timeout
	TLS             tls.ClientConfig // TLS configuration (optional)
	ProxyConfig     *ProxyConfig     // Proxy configuration (optional, for proxy mode)
	Input           string           // Input name
}

// ProxyConfig represents proxy mode configuration.
type ProxyConfig struct {
	Username string // Target username
	Password string // Target password
	URL      string // Target URL, e.g., "service:jmx:rmi:///jndi/rmi://target:9010/jmxrmi"
}

// Request represents a Jolokia request.
type Request struct {
	Type      string        `json:"type"`                // Request type: read, write, exec, search, list, version
	Mbean     string        `json:"mbean,omitempty"`     // MBean object name
	Attribute interface{}   `json:"attribute,omitempty"` // Attribute name (string or []string)
	Path      string        `json:"path,omitempty"`      // Attribute path
	Value     interface{}   `json:"value,omitempty"`     // Value to write (for write operation)
	Operation string        `json:"operation,omitempty"` // Operation name (for exec operation)
	Arguments []interface{} `json:"arguments,omitempty"` // Operation arguments (for exec operation)
	Target    *Target       `json:"target,omitempty"`    // Proxy target (for proxy mode)
}

// Target represents a proxy target.
type Target struct {
	URL      string `json:"url"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

// Response represents a Jolokia response.
type Response struct {
	Request   *Request    `json:"request"`              // Original request
	Value     interface{} `json:"value"`                // Response value
	Status    int         `json:"status"`               // Status code (200 = success)
	Timestamp int64       `json:"timestamp,omitempty"`  // Timestamp
	Error     string      `json:"error,omitempty"`      // Error message
	ErrorType string      `json:"error_type,omitempty"` // Error type
}

// NewClient creates a new Jolokia client.
func NewClient(config Config) (*Client, error) {
	tlsConfig, err := config.TLS.TLSConfig()
	if err != nil {
		return nil, fmt.Errorf("tls config: %w", err)
	}

	transport := &http.Transport{
		ResponseHeaderTimeout: config.ResponseTimeout,
		TLSClientConfig:       tlsConfig,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   config.ResponseTimeout,
	}

	return &Client{
		url:        config.URL,
		httpClient: httpClient,
		config:     &config,
	}, nil
}

// URL returns the client's connection URL.
func (c *Client) URL() string {
	return c.url
}

// Execute executes a single request.
func (c *Client) Execute(req *Request) (*Response, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}

	responses, err := c.BatchExecute([]*Request{req})
	if err != nil {
		return nil, err
	}
	if len(responses) == 0 {
		return nil, fmt.Errorf("empty response")
	}
	return responses[0], nil
}

// BatchExecute executes multiple requests in a single batch.
func (c *Client) BatchExecute(requests []*Request) ([]*Response, error) {
	startTime := time.Now()

	inputName := c.config.Input
	if inputName == "" {
		inputName = "unknown"
	}

	defer func() {
		requestTotalVec.WithLabelValues(c.url, inputName).Inc()
		requestLatencyVec.WithLabelValues(c.url, inputName).Observe(time.Since(startTime).Seconds())
	}()

	if len(requests) == 0 {
		return nil, fmt.Errorf("requests is empty")
	}

	for i, req := range requests {
		if req == nil {
			return nil, fmt.Errorf("request at index %d is nil", i)
		}
	}

	requestBody, err := json.Marshal(requests)
	if err != nil {
		requestErrorVec.WithLabelValues(c.url, inputName, "marshal_error").Inc()
		return nil, fmt.Errorf("marshal requests: %w", err)
	}

	requestURL, err := c.formatURL()
	if err != nil {
		requestErrorVec.WithLabelValues(c.url, inputName, "url_format_error").Inc()
		return nil, fmt.Errorf("format url: %w", err)
	}

	httpReq, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(requestBody))
	if err != nil {
		requestErrorVec.WithLabelValues(c.url, inputName, "create_request_error").Inc()
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	if c.config.Username != "" || c.config.Password != "" {
		httpReq.SetBasicAuth(c.config.Username, c.config.Password)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		requestErrorVec.WithLabelValues(c.url, inputName, "http_error").Inc()
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode != http.StatusOK {
		requestErrorVec.WithLabelValues(c.url, inputName, "http_status_error").Inc()
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("http status %d: %s", httpResp.StatusCode, string(body))
	}

	var result []*Response
	decoder := json.NewDecoder(httpResp.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		requestErrorVec.WithLabelValues(c.url, inputName, "decode_error").Inc()
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result, nil
}

// NewReadRequest creates a read request.
func (c *Client) NewReadRequest(mbean string, attributes []string, path string) *Request {
	req := &Request{
		Type:  "read",
		Mbean: mbean,
		Path:  path,
	}

	switch len(attributes) {
	case 0:
		req.Attribute = nil
	case 1:
		req.Attribute = attributes[0]
	default:
		req.Attribute = attributes
	}

	c.setTarget(req)
	return req
}

// NewWriteRequest creates a write request.
func (c *Client) NewWriteRequest(mbean string, attribute string, path string, value interface{}) *Request {
	req := &Request{
		Type:      "write",
		Mbean:     mbean,
		Attribute: attribute,
		Path:      path,
		Value:     value,
	}
	c.setTarget(req)
	return req
}

// NewExecRequest creates an exec request.
func (c *Client) NewExecRequest(mbean string, operation string, arguments []interface{}) *Request {
	req := &Request{
		Type:      "exec",
		Mbean:     mbean,
		Operation: operation,
		Arguments: arguments,
	}
	c.setTarget(req)
	return req
}

// NewSearchRequest creates a search request.
func (c *Client) NewSearchRequest(mbeanPattern string) *Request {
	req := &Request{
		Type:  "search",
		Mbean: mbeanPattern,
	}
	c.setTarget(req)
	return req
}

// NewListRequest creates a list request.
func (c *Client) NewListRequest(mbeanPattern string) *Request {
	req := &Request{
		Type:  "list",
		Mbean: mbeanPattern,
	}
	c.setTarget(req)
	return req
}

// NewVersionRequest creates a version request.
func (c *Client) NewVersionRequest() *Request {
	req := &Request{
		Type: "version",
	}
	c.setTarget(req)
	return req
}

func (c *Client) setTarget(req *Request) {
	if c.config.ProxyConfig == nil {
		return
	}

	req.Target = &Target{
		URL:      c.config.ProxyConfig.URL,
		User:     c.config.ProxyConfig.Username,
		Password: c.config.ProxyConfig.Password,
	}
}

func (c *Client) formatURL() (string, error) {
	parsedURL, err := url.Parse(c.url)
	if err != nil {
		return "", err
	}

	readURL := url.URL{
		Scheme: parsedURL.Scheme,
		Host:   parsedURL.Host,
	}

	readURL.Path = parsedURL.Path
	if readURL.Path == "" {
		readURL.Path = "/jolokia"
	}
	// Ensure path ends with / for POST requests
	if !strings.HasSuffix(readURL.Path, "/") {
		readURL.Path += "/"
	}

	q := readURL.Query()
	q.Set("ignoreErrors", "true")
	readURL.RawQuery = q.Encode()

	return readURL.String(), nil
}
