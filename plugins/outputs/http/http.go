package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/outputs"
)

var (
	DefaultURL = "http://127.0.0.1:8080/telegraf"
)

var sampleConfig = `
  ## URL is the address to send metrics to
  # url = "http://127.0.0.1:8080/telegraf"
  # timeout = "5s"
  # method = "POST"
  # data_format = "influx"
  # content_encoding = "gzip"

  ## Additional HTTP headers
  # [outputs.http.headers]
  #   # Should be set manually to "application/json" for json data_format
  #   Content-Type = "text/plain; charset=utf-8"
`

const (
	defaultClientTimeout = 5 * time.Second
	defaultContentType   = `text/plain; charset=utf-8`
	defaultMethod        = http.MethodPost
)

type HTTP struct {
	URL     string            `toml:"url"`
	Timeout internal.Duration `toml:"timeout"`
	Method  string            `toml:"method"`
	//Username string            `toml:"username"`
	//Password string            `toml:"password"`
	Headers map[string]string `toml:"headers"`
	//ClientID        string            `toml:"client_id"`
	//ClientSecret    string            `toml:"client_secret"`
	//TokenURL        string            `toml:"token_url"`
	//Scopes          []string          `toml:"scopes"`
	ContentEncoding string `toml:"content_encoding"`
	//tls.ClientConfig

	client     *http.Client
	serializer serializers.Serializer
}

func (h *HTTP) SetSerializer(serializer serializers.Serializer) {
	h.serializer = serializer
}

func (h *HTTP) createClient(ctx context.Context) (*http.Client, error) {
	// tlsCfg, err := h.ClientConfig.TLSConfig()
	// if err != nil {
	// 	return nil, err
	// }

	client := &http.Client{
		Transport: &http.Transport{
			//TLSClientConfig: tlsCfg,
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: h.Timeout.Duration,
	}

	// if h.ClientID != "" && h.ClientSecret != "" && h.TokenURL != "" {
	// 	oauthConfig := clientcredentials.Config{
	// 		ClientID:     h.ClientID,
	// 		ClientSecret: h.ClientSecret,
	// 		TokenURL:     h.TokenURL,
	// 		Scopes:       h.Scopes,
	// 	}
	// 	ctx = context.WithValue(ctx, oauth2.HTTPClient, client)
	// 	client = oauthConfig.Client(ctx)
	// }

	return client, nil
}

func (h *HTTP) Connect() error {
	if h.Method == "" {
		h.Method = http.MethodPost
	}
	h.Method = strings.ToUpper(h.Method)
	if h.Method != http.MethodPost && h.Method != http.MethodPut {
		return fmt.Errorf("invalid method [%s] %s", h.URL, h.Method)
	}

	if h.Timeout.Duration == 0 {
		h.Timeout.Duration = defaultClientTimeout
	}

	ctx := context.Background()
	client, err := h.createClient(ctx)
	if err != nil {
		return err
	}

	h.client = client

	return nil
}

func (h *HTTP) Close() error {
	return nil
}

func (h *HTTP) Description() string {
	return "A plugin that can transmit metrics over HTTP"
}

func (h *HTTP) SampleConfig() string {
	return sampleConfig
}

func (h *HTTP) Write(metrics []telegraf.Metric) error {
	reqBody, err := h.serializer.SerializeBatch(metrics)
	if err != nil {
		return err
	}

	if err := h.write(reqBody); err != nil {
		return err
	}

	return nil
}

func (h *HTTP) write(reqBody []byte) error {
	var reqBodyBuffer io.Reader = bytes.NewBuffer(reqBody)

	//log.Printf("D! ftdataway: %s", h.URL)

	var err error
	if h.ContentEncoding == "gzip" {
		rc, err := internal.CompressWithGzip(reqBodyBuffer)
		if err != nil {
			return err
		}
		defer rc.Close()
		reqBodyBuffer = rc
	}

	req, err := http.NewRequest(h.Method, h.URL, reqBodyBuffer)
	if err != nil {
		return err
	}

	// if h.Username != "" || h.Password != "" {
	// 	req.SetBasicAuth(h.Username, h.Password)
	// }

	//req.Header.Set("User-Agent", "Telegraf/"+internal.Version())
	req.Header.Set("Content-Type", defaultContentType)
	if h.ContentEncoding == "gzip" {
		req.Header.Set("Content-Encoding", "gzip")
	}
	for k, v := range h.Headers {
		if strings.ToLower(k) == "host" {
			req.Host = v
		}
		req.Header.Set(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("when writing to [%s] received status code: %d, body: %s", h.URL, resp.StatusCode, string(body))
	}

	return nil
}

func NewHttpOutput() *HTTP {
	return &HTTP{
		Timeout: internal.Duration{Duration: defaultClientTimeout},
		Method:  defaultMethod,
		URL:     DefaultURL,
	}
}

func init() {
	outputs.Add("http", func() telegraf.Output {
		return &HTTP{
			Timeout: internal.Duration{Duration: defaultClientTimeout},
			Method:  defaultMethod,
			URL:     DefaultURL,
		}
	})
}
