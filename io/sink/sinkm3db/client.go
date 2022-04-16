// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sinkm3db

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	// nolint:staticcheck
	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
)

// Label is a metric label.
type Label struct {
	Name  string
	Value string
}

// TimeSeries are made of labels and a datapoint.
type TimeSeries struct {
	Labels    []Label
	Datapoint Datapoint
}

// Datapoint is a single data value reported at a given time.
type Datapoint struct {
	Timestamp time.Time
	Value     float64
}

// WriteOptions specifies additional write options.
type WriteOptions struct {
	// Headers to append or override the outgoing headers.
	Headers map[string]string
}

// WriteResult returns the successful HTTP status code.
type WriteResult struct {
	StatusCode int
}

// WriteError is an error that can also return the HTTP status code
// if the response is what caused an error.
type WriteError interface {
	error
	StatusCode() int
}

// Config defines the configuration used to construct a client.
type Config struct {
	// WriteURL is the URL which the client uses to write to m3coordinator.
	WriteURL string `yaml:"writeURL"`

	// HTTPClientTimeout is the timeout that is set for the client.
	HTTPClientTimeout time.Duration `yaml:"httpClientTimeout"`

	// If not nil, http client is used instead of constructing one.
	HTTPClient *http.Client

	// UserAgent is the `User-Agent` header in the request.
	UserAgent string `yaml:"userAgent"`
}

// ConfigOption defines a config option that can be used when constructing a client.
type ConfigOption func(*Config)

// NewConfig creates a new Config struct based on options passed to the function.
func NewConfig(opts ...ConfigOption) Config {
	cfg := Config{}
	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}

func (c Config) validate() error {
	if c.HTTPClientTimeout <= 0 {
		return fmt.Errorf("http client timeout should be greater than 0: %d", c.HTTPClientTimeout)
	}

	if c.WriteURL == "" {
		return errors.New("remote write URL should not be blank")
	}

	if c.UserAgent == "" {
		return errors.New("User-Agent should not be blank")
	}

	return nil
}

// WriteURLOption sets the URL which the client uses to write to m3coordinator.
func WriteURLOption(writeURL string) ConfigOption {
	return func(c *Config) {
		c.WriteURL = writeURL
	}
}

// HTTPClientTimeoutOption sets the timeout that is set for the client.
func HTTPClientTimeoutOption(httpClientTimeout time.Duration) ConfigOption {
	return func(c *Config) {
		c.HTTPClientTimeout = httpClientTimeout
	}
}

// UserAgent sets the `User-Agent` header in the request.
func UserAgent(userAgent string) ConfigOption {
	return func(c *Config) {
		c.UserAgent = userAgent
	}
}

type client struct {
	writeURL   string
	httpClient *http.Client
	userAgent  string
}

// NewClient creates a new remote write coordinator client.
func NewClient(c Config) (*client, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Timeout: c.HTTPClientTimeout,
	}

	if c.HTTPClient != nil {
		httpClient = c.HTTPClient
	}

	return &client{
		writeURL:   c.WriteURL,
		httpClient: httpClient,
	}, nil
}

func (c *client) WriteProto(
	ctx context.Context,
	promWR *prompb.WriteRequest,
	opts WriteOptions,
) (WriteResult, WriteError) {
	var result WriteResult
	data, err := proto.Marshal(promWR)
	if err != nil {
		return result, writeError{err: fmt.Errorf("unable to marshal protobuf: %w", err)}
	}

	encoded := snappy.Encode(nil, data)

	body := bytes.NewReader(encoded)
	req, err := http.NewRequest("POST", c.writeURL, body)
	if err != nil {
		return result, writeError{err: err}
	}

	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Content-Encoding", "snappy")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")
	if opts.Headers != nil {
		for k, v := range opts.Headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return result, writeError{err: err}
	}

	result.StatusCode = resp.StatusCode

	defer resp.Body.Close() //nolint:errcheck

	if result.StatusCode/100 != 2 {
		writeErr := writeError{
			err:  fmt.Errorf("expected HTTP 200 status code: actual=%d", resp.StatusCode),
			code: result.StatusCode,
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			writeErr.err = fmt.Errorf("%w, body_read_error=%s", writeErr.err, err)
			return result, writeErr
		}

		writeErr.err = fmt.Errorf("%w, body=%s", writeErr.err, body)
		return result, writeErr
	}

	return result, nil
}

type writeError struct {
	err  error
	code int
}

func (e writeError) Error() string {
	return e.err.Error()
}

// StatusCode returns the HTTP status code of the error if error
// was caused by the response, otherwise it will be just zero.
func (e writeError) StatusCode() int {
	return e.code
}
