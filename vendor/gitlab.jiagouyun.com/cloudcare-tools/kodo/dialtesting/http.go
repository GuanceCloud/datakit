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
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
)

type HTTPTask struct {
	ExternalID     string               `json:"external_id"`
	Name           string               `json:"name"`
	AK             string               `json:"access_key"`
	Method         string               `json:"method"`
	URL            string               `json:"url"`
	PostURL        string               `json:"post_url"`
	CurStatus      string               `json:"status"`
	Frequency      string               `json:"frequency"`
	Region         string               `json:"region"` // 冗余进来，便于调试
	SuccessWhen    []*HTTPSuccess       `json:"success_when"`
	Tags           map[string]string    `json:"tags,omitempty"`
	AdvanceOptions []*HTTPAdvanceOption `json:"advance_options,omitempty"`
	UpdateTime     int64                `json:"update_time,omitempty"`

	ticker   *time.Ticker
	cli      *http.Client
	resp     *http.Response
	req      *http.Request
	respBody []byte
	reqStart time.Time
	reqCost  time.Duration
	reqError string
}

func (t *HTTPTask) UpdateTimeUs() int64 {
	return t.UpdateTime
}

func (t *HTTPTask) ID() string {
	if t.ExternalID == `` {
		return cliutils.XID("dtst_")
	}
	return fmt.Sprintf("%s_%s", t.AK, t.ExternalID)
}

func (t *HTTPTask) Stop() error {
	t.cli.CloseIdleConnections()
	return nil
}

func (t *HTTPTask) Status() string {
	return t.CurStatus
}

func (t *HTTPTask) Ticker() *time.Ticker {
	return t.ticker
}

func (t *HTTPTask) Class() string {
	return "HTTP"
}

func (t *HTTPTask) MetricName() string {
	return `http_dial_testing`
}

func (t *HTTPTask) PostURLStr() string {
	return t.PostURL
}

func (t *HTTPTask) GetResults() (tags map[string]string, fields map[string]interface{}) {
	tags = map[string]string{
		"name":   t.Name,
		"url":    t.URL,
		"region": t.Region,
		"proto":  t.req.Proto,
		"result": "FAIL",
	}

	if t.resp != nil {
		tags["status_code"] = fmt.Sprintf(`%v`, t.resp.StatusCode)
	}

	for k, v := range t.Tags {
		tags[k] = v
	}

	fields = map[string]interface{}{
		"response_time":  int64(t.reqCost) / 1000000, // 单位为ms
		"content_length": int64(len(t.respBody)),
		"success":        int64(-1),
	}

	if t.reqError != "" {
		fields[`failed_reason`] = t.reqError
	}

	return
}

func (t *HTTPTask) RegionName() string {
	return t.Region
}

func (t *HTTPTask) AccessKey() string {
	return t.AK
}

func (t *HTTPTask) Check() error {
	// TODO: check task validity
	if t.ExternalID == "" {
		return fmt.Errorf("external ID missing")
	}
	return nil
}

type HTTPSuccess struct {
	Body string `json:"body,omitempty"`

	ResponseTime string `json:"response_time,omitempty"`
	respTime     time.Duration

	Header     map[string]*SuccessOption `json:"header,omitempty"`
	StatusCode *SuccessOption            `json:"status_code,omitempty"`
}

type HTTPOptAuth struct {
	// basic auth
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	// TODO: 支持更多的 auth 选项
}

type HTTPOptRequest struct {
	FollowRedirect bool              `json:"follow_redirect"`
	Headers        map[string]string `json:"headers"`
	Cookies        string            `json:"cookies"`
	Auth           *HTTPOptAuth      `json:"auth"`
}

type HTTPOptBody struct {
	BodyType string `json:"body_type"`
	Body     string `json:"body"`
}

type HTTPOptCertificate struct {
	IgnoreServerCertificateError bool   `json:ignore_server_certificate_error`
	PrivateKey                   string `json:"private_key"`
	Certificate                  string `json:"certificate"`
}

type HTTPOptProxy struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

type HTTPAdvanceOption struct {
	RequestOptions *HTTPOptRequest     `json:"requests_options,omitempty"`
	RequestBody    *HTTPOptBody        `json:"request_body,omitempty"`
	Certificate    *HTTPOptCertificate `json:"certificate,omitempty"`
	Proxy          *HTTPOptProxy       `json:"proxy,omitempty"`
}

func (t *HTTPTask) Run() error {
	reqURL, err := url.Parse(t.URL)
	if err != nil {
		goto result
	}

	t.req, err = http.NewRequest(t.Method, reqURL.String(), nil)
	if err != nil {
		goto result
	}

	// advance options
	if err := t.setupAdvanceOpts(t.req); err != nil {
		goto result
	}

	t.reqStart = time.Now()
	t.resp, err = t.cli.Do(t.req)
	if err != nil {
		goto result
	}

	t.respBody, err = ioutil.ReadAll(t.resp.Body)
	if err != nil {
		goto result
	}
	defer t.resp.Body.Close()

result:
	if err != nil {
		t.reqError = err.Error()
	}
	t.reqCost = time.Since(t.reqStart)

	return err
}

func (t *HTTPTask) CheckResult() (reasons []string) {
	if t.resp == nil {
		return nil
	}

	for _, chk := range t.SuccessWhen {
		// check headers

		for k, v := range chk.Header {
			if err := v.check(t.resp.Header.Get(k), fmt.Sprintf("HTTP header `%s'", k)); err != nil {
				reasons = append(reasons, err.Error())
			}
		}

		// check body
		if chk.Body != "" {
			if chk.Body != string(t.respBody) {
				reasons = append(reasons, "body not match: `%s' <> `%s'", chk.Body, string(t.respBody))
			}
		}

		// check status code
		if chk.StatusCode != nil {
			if err := chk.StatusCode.check(fmt.Sprintf("%d", t.resp.StatusCode), "HTTP status"); err != nil {
				reasons = append(reasons, err.Error())
			}
		}

		// check response time
		if t.reqCost > chk.respTime && chk.respTime > 0 {
			reasons = append(reasons,
				fmt.Sprintf("HTTP response time(%v) larger than %v", t.reqCost, chk.respTime))
		}
	}

	return
}

func (t *HTTPTask) setupAdvanceOpts(req *http.Request) error {
	for _, opt := range t.AdvanceOptions {
		// request options
		if opt.RequestOptions != nil {
			// headers
			for k, v := range opt.RequestOptions.Headers {
				req.Header.Add(k, v)
			}

			// cookie
			if opt.RequestOptions.Cookies != "" {
				req.Header.Add("Cookie", opt.RequestOptions.Cookies)
			}

			// auth
			// TODO: add more auth options
			if opt.RequestOptions.Auth != nil {
				req.SetBasicAuth(opt.RequestOptions.Auth.Username, opt.RequestOptions.Auth.Password)
			}
		}

		// body options
		if opt.RequestBody != nil {
			req.Header.Add("Content-Type", opt.RequestBody.BodyType)
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

func (t *HTTPTask) Init() error {

	// setup frequency
	du, err := time.ParseDuration(t.Frequency)
	if err != nil {
		return err
	}
	if t.ticker != nil {
		t.ticker.Stop()
	}
	t.ticker = time.NewTicker(du)

	if strings.ToLower(t.CurStatus) == StatusStop {
		return nil
	}

	// setup HTTP client
	t.cli = &http.Client{
		Timeout: 30 * time.Second, // default timeout
	}

	// advance options
	for _, opt := range t.AdvanceOptions {
		if opt.RequestOptions != nil {
			// check FollowRedirect
			if !opt.RequestOptions.FollowRedirect { // see https://stackoverflow.com/a/38150816/342348
				t.cli.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}
			}
		}

		if opt.RequestBody != nil {
			switch opt.RequestBody.BodyType {
			case "text/plain", "application/json", "text/xml":
			case "": // do nothing
			default:
				return fmt.Errorf("invalid body type: `%s'", opt.RequestBody.BodyType)
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

	if len(t.SuccessWhen) == 0 {
		return fmt.Errorf(`no any check rule`)
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
