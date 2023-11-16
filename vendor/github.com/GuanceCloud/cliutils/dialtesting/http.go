// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

// HTTP dialer testing
// auth: tanb
// date: Fri Feb  5 13:17:00 CST 2021

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
)

type HTTPTask struct {
	ExternalID       string             `json:"external_id"`
	Name             string             `json:"name"`
	AK               string             `json:"access_key"`
	Method           string             `json:"method"`
	URL              string             `json:"url"`
	PostURL          string             `json:"post_url"`
	CurStatus        string             `json:"status"`
	Frequency        string             `json:"frequency"`
	Region           string             `json:"region"` // 冗余进来，便于调试
	OwnerExternalID  string             `json:"owner_external_id"`
	SuccessWhenLogic string             `json:"success_when_logic"`
	SuccessWhen      []*HTTPSuccess     `json:"success_when"`
	Tags             map[string]string  `json:"tags,omitempty"`
	Labels           []string           `json:"labels,omitempty"`
	AdvanceOptions   *HTTPAdvanceOption `json:"advance_options,omitempty"`
	UpdateTime       int64              `json:"update_time,omitempty"`
	Option           map[string]string

	ticker   *time.Ticker
	cli      *http.Client
	resp     *http.Response
	req      *http.Request
	respBody []byte
	reqStart time.Time
	reqCost  time.Duration
	reqError string

	dnsParseTime   float64
	connectionTime float64
	sslTime        float64
	ttfbTime       float64
	downloadTime   float64

	destIP string
}

const MaxMsgSize = 15 * 1024 * 1024

func (t *HTTPTask) UpdateTimeUs() int64 {
	return t.UpdateTime
}

func (t *HTTPTask) Clear() {
	t.dnsParseTime = 0.0
	t.connectionTime = 0.0
	t.sslTime = 0.0
	t.downloadTime = 0.0
	t.ttfbTime = 0.0
	t.reqCost = 0

	t.resp = nil
	t.respBody = []byte(``)
	t.reqError = ""
}

func (t *HTTPTask) ID() string {
	if t.ExternalID == `` {
		return cliutils.XID("dtst_")
	}
	return fmt.Sprintf("%s_%s", t.AK, t.ExternalID)
}

func (t *HTTPTask) GetOwnerExternalID() string {
	return t.OwnerExternalID
}

func (t *HTTPTask) SetOwnerExternalID(exid string) {
	t.OwnerExternalID = exid
}

func (t *HTTPTask) SetRegionID(regionID string) {
	t.Region = regionID
}

func (t *HTTPTask) SetAk(ak string) {
	t.AK = ak
}

func (t *HTTPTask) SetStatus(status string) {
	t.CurStatus = status
}

func (t *HTTPTask) SetUpdateTime(ts int64) {
	t.UpdateTime = ts
}

func (t *HTTPTask) Stop() error {
	if t.cli != nil {
		t.cli.CloseIdleConnections()
	}
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

func (t *HTTPTask) GetFrequency() string {
	return t.Frequency
}

func (t *HTTPTask) GetLineData() string {
	return ""
}

func (t *HTTPTask) GetResults() (tags map[string]string, fields map[string]interface{}) {
	tags = map[string]string{
		"name":    t.Name,
		"url":     t.URL,
		"proto":   t.req.Proto,
		"status":  "FAIL",
		"method":  t.Method,
		"dest_ip": t.destIP,
	}

	fields = map[string]interface{}{
		"response_time":      int64(t.reqCost) / 1000, // 单位为us
		"response_body_size": int64(len(t.respBody)),
		"success":            int64(-1),
	}

	if t.resp != nil {
		fields["status_code"] = t.resp.StatusCode
		tags["status_code_string"] = t.resp.Status
		tags["status_code_class"] = fmt.Sprintf(`%dxx`, t.resp.StatusCode/100)
	}

	for k, v := range t.Tags {
		tags[k] = v
	}

	message := map[string]interface{}{}

	if t.req != nil {
		message[`request_body`] = t.req.Body
		message[`request_header`] = t.req.Header
	}

	reasons, succFlag := t.CheckResult()
	if t.reqError != "" {
		reasons = append(reasons, t.reqError)
	}
	switch t.SuccessWhenLogic {
	case "or":
		if succFlag && t.reqError == "" {
			tags["status"] = "OK"
			fields["success"] = int64(1)
		} else {
			message[`fail_reason`] = strings.Join(reasons, `;`)
			fields[`fail_reason`] = strings.Join(reasons, `;`)
		}
	default:
		if len(reasons) != 0 {
			message[`fail_reason`] = strings.Join(reasons, `;`)
			fields[`fail_reason`] = strings.Join(reasons, `;`)
		}

		if t.reqError == "" && len(reasons) == 0 {
			tags["status"] = "OK"
			fields["success"] = int64(1)
		}
	}

	notSave := false
	if t.AdvanceOptions != nil && t.AdvanceOptions.Secret != nil && t.AdvanceOptions.Secret.NoSaveResponseBody {
		notSave = true
	}

	if v, ok := fields[`fail_reason`]; ok && !notSave && len(v.(string)) != 0 && t.resp != nil {
		message[`response_header`] = t.resp.Header
		message[`response_body`] = string(t.respBody)
	}

	fields[`response_dns`] = t.dnsParseTime
	fields[`response_connection`] = t.connectionTime
	fields[`response_ssl`] = t.sslTime
	fields[`response_ttfb`] = t.ttfbTime
	fields[`response_download`] = t.downloadTime

	data, err := json.Marshal(message)
	if err != nil {
		fields[`message`] = err.Error()
	}

	if len(data) > MaxMsgSize {
		fields[`message`] = string(data[:MaxMsgSize])
	} else {
		fields[`message`] = string(data)
	}

	return tags, fields
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

	return t.Init()
}

type HTTPSuccess struct {
	Body []*SuccessOption `json:"body,omitempty"`

	ResponseTime string `json:"response_time,omitempty"`
	respTime     time.Duration

	Header     map[string][]*SuccessOption `json:"header,omitempty"`
	StatusCode []*SuccessOption            `json:"status_code,omitempty"`
}

type HTTPOptAuth struct {
	// basic auth
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	// TODO: 支持更多的 auth 选项
}

type HTTPOptRequest struct {
	FollowRedirect bool              `json:"follow_redirect,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	Cookies        string            `json:"cookies,omitempty"`
	Auth           *HTTPOptAuth      `json:"auth,omitempty"`
}

type HTTPOptBody struct {
	BodyType string `json:"body_type,omitempty"`
	Body     string `json:"body,omitempty"`
}

type HTTPOptCertificate struct {
	IgnoreServerCertificateError bool   `json:"ignore_server_certificate_error,omitempty"`
	PrivateKey                   string `json:"private_key,omitempty"`
	Certificate                  string `json:"certificate,omitempty"`
	CaCert                       string `json:"ca,omitempty"`
}

type HTTPOptProxy struct {
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type HTTPAdvanceOption struct {
	RequestOptions *HTTPOptRequest     `json:"request_options,omitempty"`
	RequestBody    *HTTPOptBody        `json:"request_body,omitempty"`
	Certificate    *HTTPOptCertificate `json:"certificate,omitempty"`
	Proxy          *HTTPOptProxy       `json:"proxy,omitempty"`
	Secret         *HTTPSecret         `json:"secret,omitempty"`
}

type HTTPSecret struct {
	NoSaveResponseBody bool `json:"not_save,omitempty"`
}

func (t *HTTPTask) Run() error {
	t.Clear()

	var t1, connect, dns, tlsHandshake time.Time
	var body io.Reader = nil

	trace := &httptrace.ClientTrace{
		DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
		DNSDone: func(ddi httptrace.DNSDoneInfo) {
			t.dnsParseTime = float64(time.Since(dns)) / float64(time.Microsecond)
		},

		TLSHandshakeStart: func() { tlsHandshake = time.Now() },
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			t.sslTime = float64(time.Since(tlsHandshake)) / float64(time.Microsecond)
		},

		ConnectStart: func(network, addr string) { connect = time.Now() },
		ConnectDone: func(network, addr string, err error) {
			t.connectionTime = float64(time.Since(connect)) / float64(time.Microsecond)
			addrParts := strings.Split(addr, ":")
			if len(addrParts) > 0 {
				t.destIP = addrParts[0]
			}
		},

		GotFirstResponseByte: func() {
			t1 = time.Now()
			t.ttfbTime = float64(time.Since(t.reqStart)) / float64(time.Microsecond)
		},
	}

	reqURL, err := url.Parse(t.URL)
	if err != nil {
		goto result
	}

	if t.AdvanceOptions != nil && t.AdvanceOptions.RequestBody != nil && t.AdvanceOptions.RequestBody.Body != "" {
		body = strings.NewReader(t.AdvanceOptions.RequestBody.Body)
	}

	t.req, err = http.NewRequest(t.Method, reqURL.String(), body)
	if err != nil {
		goto result
	}

	// advance options
	if err := t.setupAdvanceOpts(t.req); err != nil {
		goto result
	}

	t.req = t.req.WithContext(httptrace.WithClientTrace(t.req.Context(), trace))

	t.req.Header.Add("Connection", "close")

	if agentInfo, ok := t.Option["userAgent"]; ok {
		t.req.Header.Add("User-Agent", agentInfo)
	}

	t.reqStart = time.Now()
	t.resp, err = t.cli.Do(t.req)
	if t.resp != nil {
		defer t.resp.Body.Close() //nolint:errcheck
	}

	if err != nil {
		goto result
	}

	t.reqCost = time.Since(t.reqStart)
	t.respBody, err = io.ReadAll(t.resp.Body)
	if err != nil {
		goto result
	}

	t.downloadTime = float64(time.Since(t1)) / float64(time.Microsecond)

result:
	if err != nil {
		t.reqError = err.Error()
	}

	return err
}

func (t *HTTPTask) CheckResult() (reasons []string, succFlag bool) {
	if t.resp == nil {
		return nil, true
	}

	for _, chk := range t.SuccessWhen {
		// check headers

		for k, vs := range chk.Header {
			for _, v := range vs {
				if err := v.check(t.resp.Header.Get(k), fmt.Sprintf("HTTP header `%s'", k)); err != nil {
					reasons = append(reasons, err.Error())
				} else {
					succFlag = true
				}
			}
		}

		// check body
		if chk.Body != nil {
			for _, v := range chk.Body {
				if err := v.check(string(t.respBody), "response body"); err != nil {
					reasons = append(reasons, err.Error())
				} else {
					succFlag = true
				}
			}
		}

		// check status code
		if chk.StatusCode != nil {
			for _, v := range chk.StatusCode {
				if err := v.check(fmt.Sprintf(`%d`, t.resp.StatusCode), "HTTP status"); err != nil {
					reasons = append(reasons, err.Error())
				} else {
					succFlag = true
				}
			}
		}

		// check response time
		if t.reqCost > chk.respTime && chk.respTime > 0 {
			reasons = append(reasons,
				fmt.Sprintf("HTTP response time(%v) larger than %v", t.reqCost, chk.respTime))
		} else if chk.respTime > 0 {
			succFlag = true
		}
	}

	return reasons, succFlag
}

func (t *HTTPTask) setupAdvanceOpts(req *http.Request) error {
	opt := t.AdvanceOptions

	if opt == nil {
		return nil
	}

	// request options
	if opt.RequestOptions != nil {
		// headers
		for k, v := range opt.RequestOptions.Headers {
			if k == "Host" || k == "host" {
				req.Host = v
			} else {
				req.Header.Add(k, v)
			}
		}

		// cookie
		if opt.RequestOptions.Cookies != "" {
			req.Header.Add("Cookie", opt.RequestOptions.Cookies)
		}

		// auth
		// TODO: add more auth options
		if opt.RequestOptions.Auth != nil {
			if !(opt.RequestOptions.Auth.Username == "" && opt.RequestOptions.Auth.Password == "") {
				req.SetBasicAuth(opt.RequestOptions.Auth.Username, opt.RequestOptions.Auth.Password)
			}
		}
	}

	// body options
	if opt.RequestBody != nil {
		if opt.RequestBody.BodyType != "" {
			req.Header.Add("Content-Type", opt.RequestBody.BodyType)
		}
	}

	// proxy headers
	if opt.Proxy != nil { // see https://stackoverflow.com/a/14663620/342348
		for k, v := range opt.Proxy.Headers {
			req.Header.Add(k, v)
		}
	}

	return nil
}

func (t *HTTPTask) InitDebug() error {
	return t.init(true)
}

func (t *HTTPTask) Init() error {
	return t.init(false)
}

func (t *HTTPTask) init(debug bool) error {
	if !debug {
		// setup frequency
		du, err := time.ParseDuration(t.Frequency)
		if err != nil {
			return err
		}
		if t.ticker != nil {
			t.ticker.Stop()
		}
		t.ticker = time.NewTicker(du)
	}

	if t.Option == nil {
		t.Option = map[string]string{}
	}

	if strings.EqualFold(t.CurStatus, StatusStop) {
		return nil
	}

	// setup HTTP client
	t.cli = &http.Client{
		Timeout: 30 * time.Second, // default timeout
	}

	// advance options
	opt := t.AdvanceOptions
	if opt != nil && opt.RequestOptions != nil {
		// check FollowRedirect
		if !opt.RequestOptions.FollowRedirect { // see https://stackoverflow.com/a/38150816/342348
			t.cli.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
		}
	}

	if opt != nil && opt.RequestBody != nil {
		switch opt.RequestBody.BodyType {
		case "text/plain", "application/json", "text/xml", "application/x-www-form-urlencoded":
		case "text/html", "multipart/form-data", "", "None": // do nothing
		default:
			return fmt.Errorf("invalid body type: `%s'", opt.RequestBody.BodyType)
		}
	}

	// TLS opotions
	if opt != nil && opt.Certificate != nil { // see https://venilnoronha.io/a-step-by-step-guide-to-mtls-in-go
		if opt.Certificate.IgnoreServerCertificateError {
			t.cli.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: opt.Certificate.IgnoreServerCertificateError, //nolint:gosec
				},
			}
		} else {
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM([]byte(opt.Certificate.CaCert))

			cert, err := tls.X509KeyPair([]byte(opt.Certificate.Certificate), []byte(opt.Certificate.PrivateKey))
			if err != nil {
				return err
			}

			t.cli.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{ //nolint:gosec
					RootCAs:      caCertPool,
					Certificates: []tls.Certificate{cert},
				},
			}
		}
	}

	// proxy options
	if opt != nil && opt.Proxy != nil { // see https://stackoverflow.com/a/14663620/342348
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

		for _, vs := range checker.Header {
			for _, v := range vs {
				err := genReg(v)
				if err != nil {
					return err
				}
			}
		}

		// body
		for _, v := range checker.Body {
			err := genReg(v)
			if err != nil {
				return err
			}
		}

		// status_code
		for _, v := range checker.StatusCode {
			err := genReg(v)
			if err != nil {
				return err
			}
		}
	}

	// TODO: more checking on task validity

	return nil
}
