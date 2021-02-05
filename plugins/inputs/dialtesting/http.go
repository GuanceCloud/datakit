package dialtesting

// HTTP dialer testing
// auth: tanb
// date: Fri Feb  5 13:17:00 CST 2021

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

type httpTask struct {
	TID     string `json:"id"`
	Method  string `json:"method"`
	URL     string `json:"url"`
	PostURL string `json:"post_url"`
	Name    string `json:"name"`

	CurStatus string `json:"status,omitempty"`

	Frequency string `json:"frequency"`
	ticker    *time.Ticker

	Locations   []string       `json:"locations"`
	SuccessWhen []*httpSuccess `json:"success_when"`

	Tags           map[string]string    `json:"tags,omitempty"`
	AdvanceOptions []*httpAdvanceOption `json:"advance_options,omitempty"`

	cli      *http.Client
	resp     *http.Response
	respBody []byte
	reqStart time.Time
	reqCost  time.Duration
}

func (t *httpTask) ID() string {
	return t.TID
}

func (t *httpTask) Stop() error {
	t.cli.CloseIdleConnections()
	return nil
}

func (t *httpTask) Status() string {
	return t.CurStatus
}

func (t *httpTask) Ticker() *time.Ticker {
	return t.ticker
}

type httpSuccess struct {
	Body string `json:"body"`

	ResponseTime string `json:"response_time"`
	respTime     time.Duration

	Header     map[string]*successOption `json:"header"`
	StatusCode *successOption            `json:"status_code"`
}

type httpOptAuth struct {
	// basic auth
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	// TODO: 支持更多的 auth 选项
}

type httpOptRequest struct {
	FollowRedirect bool              `json:"follow_redirect"`
	Headers        map[string]string `json:"headers"`
	Cookies        string            `json:"cookies"`
	Auth           *httpOptAuth      `json:"auth"`
}

type httpOptBody struct {
	BodyType string `json:"body_type"`
	Body     string `json:"body"`
}

type httpOptCertificate struct {
	IgnoreServerCertificateError bool   `json:ignore_server_certificate_error`
	PrivateKey                   string `json:"private_key"`
	Certificate                  string `json:"certificate"`
}

type httpOptProxy struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

type httpAdvanceOption struct {
	RequestsOptions *httpOptRequest     `json:"requests_options,omitempty"`
	RequestBody     *httpOptBody        `json:"request_body,omitempty"`
	Certificate     *httpOptCertificate `json:"certificate,omitempty"`
	Proxy           *httpOptProxy       `json:"proxy,omitempty"`
}

func (t *httpTask) Run() error {
	reqURL, err := url.Parse(t.URL)
	if err != nil {
		l.Errorf("url.Parse(%s): %s", t.URL, err.Error())
		return err
	}

	req, err := http.NewRequest(t.Method, reqURL.String(), nil)
	if err != nil {
		l.Errorf("http.NewRequest(): %s", err.Error())
		return err
	}

	// advance options
	if err := t.setupAdvanceOpts(req); err != nil {
		l.Errorf("setupAdvanceOpts(): %s", err.Error())
		return err
	}

	t.reqStart = time.Now()
	t.resp, err = t.cli.Do(req)
	if err != nil {
		l.Info("cli.Do(): %s", err)
		return err
	}

	t.respBody, err = ioutil.ReadAll(t.resp.Body)
	if err != nil {
		l.Errorf("ioutil.ReadAll(): %s", err)
		return err
	}
	defer t.resp.Body.Close()
	t.reqCost = time.Since(t.reqStart)

	return nil
}

func (t *httpTask) CheckResult() (reasons []string) {

	for _, chk := range t.SuccessWhen {
		// check headers
		for k, v := range chk.Header {
			if err := v.check(t.resp.Header.Get(k), fmt.Sprintf("HTTP header %s", k)); err != nil {
				reasons = append(reasons, err.Error())
			}
		}

		// check body
		if chk.Body != "" {
			if chk.Body != string(t.respBody) {
				reasons = append(reasons, "body not match: %s <> %s", chk.Body, string(t.respBody))
			}
		}

		// check status code
		if chk.StatusCode != nil {
			if err := chk.StatusCode.check(fmt.Sprintf("%s", t.resp.StatusCode), "HTTP status"); err != nil {
				reasons = append(reasons, err.Error())
			}
		}

		// check response time
		if t.reqCost > chk.respTime {
			reasons = append(reasons,
				fmt.Sprintf("HTTP response time(%v) larger than %v", t.reqCost, chk.respTime))
		}
	}

	return
}

func (t *httpTask) setupAdvanceOpts(req *http.Request) error {
	for _, opt := range t.AdvanceOptions {
		// request options
		if opt.RequestsOptions != nil {
			// headers
			for k, v := range opt.RequestsOptions.Headers {
				req.Header.Add(k, v)
			}

			// cookie
			if opt.RequestsOptions.Cookies != "" {
				req.Header.Add("Cookie", opt.RequestsOptions.Cookies)
			}

			// auth
			// TODO: add more auth options
			if opt.RequestsOptions.Auth != nil {
				req.SetBasicAuth(opt.RequestsOptions.Auth.Username, opt.RequestsOptions.Auth.Password)
			}
		}

		// body options
		if opt.RequestBody != nil {
			switch opt.RequestBody.BodyType {
			case "text/plain", "application/json", "text/xml":
				req.Header.Add("Content-Type", opt.RequestBody.BodyType)
			case "": // do nothing
			default:
				return fmt.Errorf("invalid body type: `%s'", opt.RequestBody.BodyType)
			}

			// setup body
			req.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(opt.RequestBody.Body)))
		}

		// proxy headers
		if opt.Proxy != nil { // see https://stackoverflow.com/a/14663620/342348
			for k, v := range opt.Proxy.Headers {
				req.Header.Add(k, v)
			}
		}
	}

	return nil
}

func (t *httpTask) Init() error {

	if t.CurStatus == StatusStop {
		return nil
	}

	// setup HTTP client
	t.cli = &http.Client{
		Timeout: 30 * time.Second, // default timeout
	}

	// setup frequency
	du, err := time.ParseDuration(t.Frequency)
	if err != nil {
		return err
	}
	if t.ticker != nil {
		t.ticker.Stop()
	}
	t.ticker = time.NewTicker(du)

	// check FollowRedirect
	for _, opt := range t.AdvanceOptions {
		if opt.RequestsOptions != nil {
			if !opt.RequestsOptions.FollowRedirect { // see https://stackoverflow.com/a/38150816/342348
				t.cli.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}
			}
		}

		// TLS opotions
		if opt.Certificate != nil { // see https://venilnoronha.io/a-step-by-step-guide-to-mtls-in-go
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM([]byte(opt.Certificate.Certificate))

			cert, err := tls.X509KeyPair([]byte(opt.Certificate.Certificate), []byte(opt.Certificate.PrivateKey))
			if err != nil {
				return err
			}

			t.cli.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:            caCertPool,
					Certificates:       []tls.Certificate{cert},
					InsecureSkipVerify: opt.Certificate.IgnoreServerCertificateError},
			}
		}

		// proxy options
		if opt.Proxy != nil { // see https://stackoverflow.com/a/14663620/342348
			proxyURL, err := url.Parse(opt.Proxy.URL)
			if err != nil {
				return err
			}

			if t.cli.Transport == nil {
				t.cli.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
			} else {
				t.cli.Transport.(*http.Transport).Proxy = http.ProxyURL(proxyURL)
			}
		}
	}

	// init success checker
	for _, checker := range t.SuccessWhen {
		if checker.ResponseTime != "" {
			du, err := time.ParseDuration(checker.ResponseTime)
			if err != nil {
				return err
			}
			checker.respTime = du
		}

		for _, v := range checker.Header {
			if v.MatchRegex != "" {
				if re, err := regexp.Compile(v.MatchRegex); err != nil {
					return err
				} else {
					v.matchRe = re
				}
			}

			if v.NotMatchRegex != "" {
				if re, err := regexp.Compile(v.NotMatchRegex); err != nil {
					return err
				} else {
					v.notMatchRe = re
				}
			}
		}
	}

	// TODO: more checking on task validity

	return nil
}
