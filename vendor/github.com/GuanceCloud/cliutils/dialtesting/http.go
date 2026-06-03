// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

// HTTP dialer testing
// auth: tanb
// date: Fri Feb  5 13:17:00 CST 2021

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/net/http2"
)

var (
	_ TaskChild = (*HTTPTask)(nil)
	_ ITask     = (*HTTPTask)(nil)
)

const (
	MaxBodySize           = 10 * 1024
	DefaultHTTPTimeout    = 60 * time.Second
	HTTP3HandshakeTimeout = 5 * time.Second

	ProtocolAuto      = "auto"
	ProtocolHTTP11    = "http/1.1"
	ProtocolHTTP2     = "http/2"
	ProtocolHTTP2Only = "http/2-only"
	ProtocolHTTP3     = "http/3"
)

type HTTPTask struct {
	*Task
	URL              string             `json:"url"`
	Method           string             `json:"method"`
	PostScript       string             `json:"post_script,omitempty"`
	SuccessWhenLogic string             `json:"success_when_logic"`
	SuccessWhen      []*HTTPSuccess     `json:"success_when"`
	AdvanceOptions   *HTTPAdvanceOption `json:"advance_options,omitempty"`

	cli                *http.Client
	resp               *http.Response
	req                *http.Request
	reqHeader          map[string]string
	reqBody            *HTTPOptBody
	respBody           []byte
	reqStart           time.Time
	reqCost            time.Duration
	reqError           string
	postScriptResult   *ScriptResult
	reqBodyBytesBuffer *bytes.Buffer

	dnsParseTime   float64
	connectionTime float64
	sslTime        float64
	ttfbTime       float64
	downloadTime   float64

	destIP           string
	rawTask          *HTTPTask
	sslCertNotBefore int64
	sslCertNotAfter  int64
	protocol         string
	httpTimeout      time.Duration
	tlsConfig        *tls.Config
}

func (t *HTTPTask) clear() {
	t.dnsParseTime = 0.0
	t.connectionTime = 0.0
	t.sslTime = 0.0
	t.downloadTime = 0.0
	t.ttfbTime = 0.0
	t.reqCost = 0

	t.resp = nil
	t.respBody = []byte(``)
	t.reqError = ""
	t.reqBodyBytesBuffer = nil
	t.sslCertNotBefore = 0
	t.sslCertNotAfter = 0
	t.destIP = ""

	if t.reqBody != nil {
		t.reqBody.bodyType = t.reqBody.BodyType
	}
}

func (t *HTTPTask) stop() {
	t.closeHTTP3Transport()
	if t.cli != nil {
		t.cli.CloseIdleConnections()
	}
}

func (t *HTTPTask) class() string {
	return ClassHTTP
}

func (t *HTTPTask) metricName() string {
	return `http_dial_testing`
}

func (t *HTTPTask) getResults() (tags map[string]string, fields map[string]interface{}) {
	tags = map[string]string{
		"name":    t.Name,
		"url":     t.URL,
		"status":  "FAIL",
		"method":  t.Method,
		"dest_ip": t.destIP,
	}

	if t.rawTask != nil {
		tags["url"] = t.rawTask.URL
	}

	if t.resp != nil {
		tags["proto"] = t.resp.Proto
	} else if t.req != nil {
		tags["proto"] = t.req.Proto
	}

	fields = map[string]interface{}{
		"response_time":      int64(t.reqCost) / 1000, // unit: us
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

	message := map[string]interface{}{
		"request_body":   t.reqBody,
		"request_header": t.reqHeader,
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
		respBody := string(t.respBody)
		if len(respBody) > MaxBodySize {
			respBody = respBody[:MaxBodySize] + "..."
		}
		message[`response_body`] = respBody
	}

	fields[`response_dns`] = t.dnsParseTime
	fields[`response_connection`] = t.connectionTime
	fields[`response_ssl`] = t.sslTime
	fields[`response_ttfb`] = t.ttfbTime
	fields[`response_download`] = t.downloadTime

	if t.sslCertNotAfter > 0 {
		fields[`ssl_cert_not_after`] = t.sslCertNotAfter
		fields[`ssl_cert_expires_in_days`] = (t.sslCertNotAfter - time.Now().UnixMicro()) / (24 * time.Hour).Microseconds()
	}

	message["status"] = tags["status"]
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
	// TODO: support more auth options
}

type HTTPOptRequest struct {
	FollowRedirect bool              `json:"follow_redirect,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	Cookies        string            `json:"cookies,omitempty"`
	Auth           *HTTPOptAuth      `json:"auth,omitempty"`
}

type HTTPOptBody struct {
	BodyType string            `json:"body_type,omitempty"`
	Body     string            `json:"body,omitempty"`
	Files    []HTTPOptBodyFile `json:"files,omitempty"`
	Form     map[string]string `json:"form,omitempty"`

	bodyType string `json:"-"` // used for multipart/form-data
}

type HTTPOptBodyFile struct {
	Name             string `json:"name"`                // field name
	Content          string `json:"content"`             // file content in base64
	Type             string `json:"type"`                // file type, e.g. image/jpeg
	Size             int64  `json:"size"`                // Content size
	Encoding         string `json:"encoding"`            // Content encoding, base64 only
	OriginalFileName string `json:"original_file_name"`  // Original file name
	FilePath         string `json:"file_path,omitempty"` // file path in storage

	Hash string `json:"_"`
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
	RequestTimeout string              `json:"request_timeout,omitempty"`
	Protocol       string              `json:"protocol,omitempty"` // "auto", "http/1.1", "http/2", "http/3"
}

type HTTPSecret struct {
	NoSaveResponseBody bool `json:"not_save,omitempty"`
}

func (opt *HTTPAdvanceOption) getProtocol() string {
	if opt == nil {
		return ProtocolAuto
	}
	switch strings.ToLower(opt.Protocol) {
	case "http/1.1", "http1.1", "1.1", "http/1.1 only":
		return ProtocolHTTP11
	case "http/2", "http2", "2", "http/2 fallback", "http/2 fallback to http/1.1":
		return ProtocolHTTP2
	case "http/2-only", "http/2 only", "http2-only":
		return ProtocolHTTP2Only
	case "http/3", "http3", "3":
		return ProtocolHTTP3
	default:
		return ProtocolAuto
	}
}

func (t *HTTPTask) run() error {
	var (
		t1,
		connect,
		dns,
		tlsHandshake time.Time
		body io.Reader
	)

	trace := &httptrace.ClientTrace{
		DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
		DNSDone: func(ddi httptrace.DNSDoneInfo) {
			t.dnsParseTime = float64(time.Since(dns)) / float64(time.Microsecond)
		},

		TLSHandshakeStart: func() { tlsHandshake = time.Now() },
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			t.sslTime = float64(time.Since(tlsHandshake)) / float64(time.Microsecond)
			// Extract SSL certificate validity information only if TLS handshake was successful
			if err == nil {
				t.extractSSLCertificateValidity(cs)
			}
		},

		ConnectStart: func(network, addr string) { connect = time.Now() },
		ConnectDone: func(network, addr string, err error) {
			t.connectionTime = float64(time.Since(connect)) / float64(time.Microsecond)
			if host, _, err := net.SplitHostPort(addr); err == nil {
				t.destIP = host
			} else {
				t.destIP = addr
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

	body, err = t.getRequestBody()
	if err != nil {
		goto result
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

	if t.protocol != ProtocolHTTP2Only && t.protocol != ProtocolHTTP3 {
		t.req.Header.Add("Connection", "close")
	}

	if agentInfo, ok := t.GetOption()["userAgent"]; ok {
		t.req.Header.Add("User-Agent", agentInfo)
	}

	if t.protocol == ProtocolHTTP3 {
		t.resetHTTP3Client()
	}

	t.reqStart = time.Now()
	t.resp, err = t.cli.Do(t.req)
	if t.protocol == ProtocolHTTP3 && t.resp != nil && t1.IsZero() {
		// For HTTP/3, response_ttfb is response header arrival measured by http.Client.Do return.
		t1 = time.Now()
		t.ttfbTime = float64(time.Since(t.reqStart)) / float64(time.Microsecond)
	}
	if t.protocol == ProtocolHTTP3 && t.resp != nil && t.resp.TLS != nil && t.sslCertNotAfter == 0 {
		t.extractSSLCertificateValidity(*t.resp.TLS)
	}
	if t.resp != nil {
		defer t.resp.Body.Close() //nolint:errcheck
	}

	if err != nil {
		goto result
	}

	if t.protocol == ProtocolHTTP2Only && t.resp.Proto != "HTTP/2.0" {
		t.reqError = fmt.Sprintf("expected HTTP/2, but got %s", t.resp.Proto)
		goto result
	}

	t.respBody, err = io.ReadAll(t.resp.Body)
	t.reqCost = time.Since(t.reqStart)
	if err != nil {
		goto result
	}

	if t.PostScript != "" {
		if result, err := postScriptDo(t.PostScript, t.respBody, t.resp); err != nil {
			t.reqError = err.Error()
			goto result
		} else {
			t.postScriptResult = result
		}
	}

	t.downloadTime = float64(time.Since(t1)) / float64(time.Microsecond)

result:
	if err != nil {
		t.reqError = err.Error()
	}

	return nil
}

func (t *HTTPTask) getRequestBody() (io.Reader, error) {
	if t.AdvanceOptions == nil || t.AdvanceOptions.RequestBody == nil {
		return nil, nil
	}

	if t.reqBodyBytesBuffer != nil {
		return t.reqBodyBytesBuffer, nil
	}

	var body *bytes.Buffer = &bytes.Buffer{}
	requestBody := t.AdvanceOptions.RequestBody

	if requestBody.BodyType == "multipart/form-data" {
		buf := &bytes.Buffer{}
		writer := multipart.NewWriter(buf)
		for k, v := range requestBody.Form {
			if err := writer.WriteField(k, v); err != nil {
				return nil, fmt.Errorf("failed to write form field %s: %w", k, err)
			}
		}

		for _, v := range requestBody.Files {
			if v.Encoding != "base64" {
				return nil, fmt.Errorf("only base64 encoding is supported for file encoding")
			}

			if fileBytes, err := base64.StdEncoding.DecodeString(v.Content); err != nil {
				return nil, fmt.Errorf("failed to decode base64 file content: %w", err)
			} else {
				if part, err := writer.CreateFormFile(v.Name, v.OriginalFileName); err != nil {
					return nil, fmt.Errorf("failed to create form file %s: %w", v.Name, err)
				} else if _, err := io.Copy(part, bytes.NewReader(fileBytes)); err != nil {
					return nil, fmt.Errorf("failed to copy file content to form file %s: %w", v.Name, err)
				}
			}
		}
		writer.Close() // nolint: errcheck,gosec
		requestBody.bodyType = writer.FormDataContentType()
		body = buf
	} else if requestBody.Body != "" {
		body = bytes.NewBufferString(requestBody.Body)
	}

	t.reqBodyBytesBuffer = body
	return body, nil
}

func (t *HTTPTask) check() error {
	if t.reqBody != nil {
		for _, f := range t.reqBody.Files {
			if f.Encoding != "base64" {
				return fmt.Errorf("only base64 encoding is supported for file encoding")
			}
		}
	}
	return nil
}

func (t *HTTPTask) checkResult() (reasons []string, succFlag bool) {
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

	if t.postScriptResult != nil {
		if t.postScriptResult.Result.IsFailed {
			reasons = append(reasons, t.postScriptResult.Result.ErrorMessage)
		} else {
			succFlag = true
		}
	}

	return reasons, succFlag
}

func (t *HTTPTask) setupAdvanceOpts(req *http.Request) error {
	opt := t.AdvanceOptions
	t.reqBody = &HTTPOptBody{}
	t.reqHeader = make(map[string]string)

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

			t.reqHeader[k] = v
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
			req.Header.Add("Content-Type", opt.RequestBody.bodyType)
			t.reqHeader["Content-Type"] = opt.RequestBody.BodyType
		}
		t.reqBody = opt.RequestBody
	}

	// proxy headers
	if opt.Proxy != nil { // see https://stackoverflow.com/a/14663620/342348
		for k, v := range opt.Proxy.Headers {
			req.Header.Add(k, v)
		}
	}

	return nil
}

func (t *HTTPTask) init() error {
	httpTimeout := DefaultHTTPTimeout

	// advance options
	opt := t.AdvanceOptions

	if opt != nil && opt.RequestTimeout != "" {
		du, err := time.ParseDuration(opt.RequestTimeout)
		if err != nil {
			return err
		}

		httpTimeout = du
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if opt != nil && opt.Certificate != nil {
		if opt.Certificate.IgnoreServerCertificateError {
			tlsConfig.InsecureSkipVerify = true //nolint:gosec
		} else if opt.Certificate.CaCert != "" {
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM([]byte(opt.Certificate.CaCert))
			tlsConfig.RootCAs = caCertPool

			cert, err := tls.X509KeyPair([]byte(opt.Certificate.Certificate), []byte(opt.Certificate.PrivateKey))
			if err != nil {
				return err
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}

	protocol := strings.ToLower(opt.getProtocol())
	t.protocol = protocol
	t.httpTimeout = httpTimeout
	t.tlsConfig = tlsConfig.Clone()

	switch protocol {
	case ProtocolHTTP3:
		t.cli = &http.Client{
			Timeout:   httpTimeout,
			Transport: t.newHTTP3RoundTripper(tlsConfig, httpTimeout),
		}
	case ProtocolHTTP2Only:
		if isPlainHTTP(t.URL) {
			t.cli = &http.Client{
				Timeout: httpTimeout,
				Transport: &http2.Transport{
					AllowHTTP: true,
					DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
						var dialer net.Dialer
						return dialer.DialContext(ctx, network, addr)
					},
				},
			}
		} else {
			http2OnlyTLS := tlsConfig.Clone()
			http2OnlyTLS.NextProtos = []string{"h2"}
			t.cli = &http.Client{
				Timeout: httpTimeout,
				Transport: &http.Transport{
					TLSClientConfig:   http2OnlyTLS,
					ForceAttemptHTTP2: true,
				},
			}
		}
	case ProtocolHTTP2:
		t.cli = &http.Client{
			Timeout: httpTimeout,
			Transport: &http.Transport{
				TLSClientConfig:   tlsConfig,
				ForceAttemptHTTP2: true,
			},
		}
	case ProtocolHTTP11:
		t.cli = &http.Client{
			Timeout: httpTimeout,
			Transport: &http.Transport{
				TLSClientConfig:   tlsConfig,
				ForceAttemptHTTP2: false,
			},
		}
	default:
		t.cli = &http.Client{
			Timeout: httpTimeout,
			Transport: &http.Transport{
				TLSClientConfig:   tlsConfig,
				ForceAttemptHTTP2: true,
			},
		}
	}

	if opt != nil {
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
			case "text/plain", "application/json", "text/xml", "application/x-www-form-urlencoded":
			case "text/html", "multipart/form-data", "", "None": // do nothing
			default:
				return fmt.Errorf("invalid body type: `%s'", opt.RequestBody.BodyType)
			}

			opt.RequestBody.bodyType = opt.RequestBody.BodyType
		}

		if protocol == ProtocolHTTP3 && opt.Proxy != nil {
			return fmt.Errorf("HTTP/3 does not support proxy configuration")
		}

		if opt.Proxy != nil { // see https://stackoverflow.com/a/14663620/342348
			proxyURL, err := url.Parse(opt.Proxy.URL)
			if err != nil {
				return err
			}

			if t.cli.Transport == nil {
				t.cli.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
			} else if transport, ok := t.cli.Transport.(*http.Transport); ok {
				transport.Proxy = http.ProxyURL(proxyURL)
			}
		}
	}

	if len(t.SuccessWhen) == 0 && t.PostScript == "" {
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

	return nil
}

func isPlainHTTP(rawURL string) bool {
	u, err := url.Parse(rawURL)
	return err == nil && strings.EqualFold(u.Scheme, "http")
}

func (t *HTTPTask) resetHTTP3Client() {
	t.closeHTTP3Transport()
	t.cli = &http.Client{
		Timeout:   t.httpTimeout,
		Transport: t.newHTTP3RoundTripper(t.tlsConfig, t.httpTimeout),
	}
	if t.AdvanceOptions != nil && t.AdvanceOptions.RequestOptions != nil && !t.AdvanceOptions.RequestOptions.FollowRedirect {
		t.cli.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
}

func (t *HTTPTask) closeHTTP3Transport() {
	if t.protocol != ProtocolHTTP3 || t.cli == nil || t.cli.Transport == nil {
		return
	}
	if closer, ok := t.cli.Transport.(interface{ Close() error }); ok {
		_ = closer.Close()
	}
}

func (t *HTTPTask) newHTTP3RoundTripper(tlsConfig *tls.Config, httpTimeout time.Duration) http.RoundTripper {
	return &http3.Transport{
		TLSClientConfig: tlsConfig,
		Dial: func(ctx context.Context, addr string, tlsCfg *tls.Config, cfg *quic.Config) (*quic.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			resolveStart := time.Now()
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			t.dnsParseTime = float64(time.Since(resolveStart)) / float64(time.Microsecond)
			if err != nil {
				return nil, err
			}
			if len(ips) == 0 {
				return nil, fmt.Errorf("no IP addresses found for %q", host)
			}
			ips = preferIPv4(ips)

			dialTLSConfig := tlsCfg
			if dialTLSConfig == nil {
				dialTLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
			} else {
				dialTLSConfig = dialTLSConfig.Clone()
			}
			if dialTLSConfig.ServerName == "" {
				dialTLSConfig.ServerName = host
			}
			dialQUICConfig := limitHTTP3HandshakeTimeout(cfg, httpTimeout)

			var lastErr error
			for _, ip := range ips {
				resolvedAddr := net.JoinHostPort(ip.IP.String(), port)
				connectStart := time.Now()
				conn, err := quic.DialAddrEarly(ctx, resolvedAddr, dialTLSConfig, dialQUICConfig)
				// For HTTP/3, response_connection is QUIC connection setup time, not TCP connect time.
				t.connectionTime = float64(time.Since(connectStart)) / float64(time.Microsecond)
				if err != nil {
					lastErr = err
					continue
				}

				handshakeStart := time.Now()
				select {
				case <-conn.HandshakeComplete():
					// For HTTP/3, response_ssl is QUIC/TLS handshake time, not standalone TCP TLS handshake time.
					t.sslTime = float64(time.Since(handshakeStart)) / float64(time.Microsecond)
					t.destIP = ip.IP.String()
					return conn, nil
				case <-ctx.Done():
					_ = conn.CloseWithError(quic.ApplicationErrorCode(0), ctx.Err().Error())
					return nil, ctx.Err()
				case <-conn.Context().Done():
					lastErr = conn.Context().Err()
					continue
				}
			}

			return nil, lastErr
		},
	}
}

func limitHTTP3HandshakeTimeout(cfg *quic.Config, httpTimeout time.Duration) *quic.Config {
	var dialQUICConfig *quic.Config
	if cfg == nil {
		dialQUICConfig = &quic.Config{}
	} else {
		dialQUICConfig = cfg.Clone()
	}

	handshakeTimeout := HTTP3HandshakeTimeout
	if httpTimeout > 0 && httpTimeout < HTTP3HandshakeTimeout {
		handshakeTimeout = httpTimeout
	} else if dialQUICConfig.HandshakeIdleTimeout == 0 {
		return dialQUICConfig
	}
	if dialQUICConfig.HandshakeIdleTimeout > handshakeTimeout {
		dialQUICConfig.HandshakeIdleTimeout = handshakeTimeout
	}
	return dialQUICConfig
}

func preferIPv4(ips []net.IPAddr) []net.IPAddr {
	if len(ips) <= 1 {
		return ips
	}

	sorted := make([]net.IPAddr, 0, len(ips))
	for _, ip := range ips {
		if ip.IP.To4() != nil {
			sorted = append(sorted, ip)
		}
	}
	for _, ip := range ips {
		if ip.IP.To4() == nil {
			sorted = append(sorted, ip)
		}
	}
	return sorted
}

func (t *HTTPTask) getHostName() ([]string, error) {
	if hostName, err := getHostName(t.URL); err != nil {
		return nil, err
	} else {
		return []string{hostName}, nil
	}
}

func (t *HTTPTask) getVariableValue(variable Variable) (string, error) {
	if variable.PostScript == "" && t.PostScript == "" {
		return "", fmt.Errorf("post_script is empty")
	}

	if variable.TaskVarName == "" {
		return "", fmt.Errorf("task variable name is empty")
	}

	if t.respBody == nil || t.resp == nil {
		return "", fmt.Errorf("response body or response is empty")
	}

	var result *ScriptResult
	var err error
	if variable.PostScript == "" { // use task post script
		result = t.postScriptResult
	} else { // use task variable post script
		if result, err = postScriptDo(variable.PostScript, t.respBody, t.resp); err != nil {
			return "", fmt.Errorf("run pipeline failed: %w", err)
		}
	}

	if result == nil {
		return "", fmt.Errorf("pipeline result is empty")
	}

	value, ok := result.Vars[variable.TaskVarName]
	if !ok {
		return "", fmt.Errorf("task variable name not found")
	} else {
		return fmt.Sprintf("%v", value), nil
	}
}

func (t *HTTPTask) getRawTask(taskString string) (string, error) {
	task := HTTPTask{}

	if err := json.Unmarshal([]byte(taskString), &task); err != nil {
		return "", fmt.Errorf("unmarshal http task failed: %w", err)
	}

	task.Task = nil

	bytes, _ := json.Marshal(task)
	return string(bytes), nil
}

func (t *HTTPTask) initTask() {
	if t.Task == nil {
		t.Task = &Task{}
	}
}

func (t *HTTPTask) renderTemplate(fm template.FuncMap) error {
	if t.rawTask == nil {
		task := &HTTPTask{}
		if err := t.NewRawTask(task); err != nil {
			return fmt.Errorf("new raw task failed: %w", err)
		}
		t.rawTask = task
	}

	task := t.rawTask
	if task == nil {
		return fmt.Errorf("raw task is nil")
	}

	// url
	if url, err := t.GetParsedString(task.URL, fm); err != nil {
		return fmt.Errorf("render url failed: %w", err)
	} else {
		t.URL = url
	}

	// success when
	if err := t.renderSuccessWhen(task, fm); err != nil {
		return fmt.Errorf("render success when failed: %w", err)
	}

	// advance options
	if err := t.renderAdvanceOptions(task, fm); err != nil {
		return fmt.Errorf("render advance options failed: %w", err)
	}

	return nil
}

func (t *HTTPTask) renderAdvanceOptions(task *HTTPTask, fm template.FuncMap) error {
	if task == nil || task.AdvanceOptions == nil {
		return nil
	}

	opt := task.AdvanceOptions

	// request options
	if err := t.renderRequestOptions(opt.RequestOptions, fm); err != nil {
		return fmt.Errorf("render request options failed: %w", err)
	}

	// request body
	if err := t.renderRequestBody(opt.RequestBody, fm); err != nil {
		return fmt.Errorf("render request body failed: %w", err)
	}

	return nil
}

func (t *HTTPTask) renderRequestBody(requestBody *HTTPOptBody, fm template.FuncMap) error {
	if requestBody == nil {
		return nil
	}

	// body
	if text, err := t.GetParsedString(requestBody.Body, fm); err != nil {
		return fmt.Errorf("render request body failed: %w", err)
	} else {
		t.AdvanceOptions.RequestBody.Body = text
	}

	// form
	for k, v := range requestBody.Form {
		key, err := t.GetParsedString(k, fm)
		if err != nil {
			return fmt.Errorf("render form failed: %w", err)
		}
		value, err := t.GetParsedString(v, fm)
		if err != nil {
			return fmt.Errorf("render form failed: %w", err)
		}

		delete(t.AdvanceOptions.RequestBody.Form, k)
		t.AdvanceOptions.RequestBody.Form[key] = value
	}

	return nil
}

func (t *HTTPTask) renderRequestOptions(requestOpt *HTTPOptRequest, fm template.FuncMap) error {
	if requestOpt != nil {
		// header
		for k, v := range requestOpt.Headers {
			if text, err := t.GetParsedString(v, fm); err != nil {
				return fmt.Errorf("render header failed: %w", err)
			} else {
				t.AdvanceOptions.RequestOptions.Headers[k] = text
			}
		}

		// cookies
		if text, err := t.GetParsedString(requestOpt.Cookies, fm); err != nil {
			return fmt.Errorf("render cookies failed: %w", err)
		} else {
			t.AdvanceOptions.RequestOptions.Cookies = text
		}

		// auth
		if requestOpt.Auth != nil {
			if text, err := t.GetParsedString(requestOpt.Auth.Username, fm); err != nil {
				return fmt.Errorf("render auth username failed: %w", err)
			} else {
				t.AdvanceOptions.RequestOptions.Auth.Username = text
			}

			if text, err := t.GetParsedString(requestOpt.Auth.Password, fm); err != nil {
				return fmt.Errorf("render auth password failed: %w", err)
			} else {
				t.AdvanceOptions.RequestOptions.Auth.Password = text
			}
		}
	}
	return nil
}

func (t *HTTPTask) renderSuccessWhen(task *HTTPTask, fm template.FuncMap) error {
	if task == nil {
		return nil
	}

	if task.SuccessWhen != nil {
		for index, checker := range task.SuccessWhen {
			// body
			for bodyIndex, v := range checker.Body {
				if err := t.renderSuccessOption(v, t.SuccessWhen[index].Body[bodyIndex], fm); err != nil {
					return fmt.Errorf("render body failed: %w", err)
				}
			}

			// response time
			if text, err := t.GetParsedString(checker.ResponseTime, fm); err != nil {
				return fmt.Errorf("render response time failed: %w", err)
			} else {
				t.SuccessWhen[index].ResponseTime = text
			}

			// header
			for headerIndex, v := range checker.Header {
				for header, option := range v {
					if err := t.renderSuccessOption(option, t.SuccessWhen[index].Header[headerIndex][header], fm); err != nil {
						return fmt.Errorf("render header failed: %w", err)
					}
				}
			}
		}
	}

	return nil
}

func (t *HTTPTask) setReqError(err string) {
	t.reqError = err
}

// extractSSLCertificateValidity extracts SSL certificate validity information from the given connection state.
func (t *HTTPTask) extractSSLCertificateValidity(cs tls.ConnectionState) {
	if len(cs.PeerCertificates) > 0 {
		cert := cs.PeerCertificates[0] // Use the first certificate in the chain (server certificate)
		t.sslCertNotBefore = cert.NotBefore.UnixMicro()
		t.sslCertNotAfter = cert.NotAfter.UnixMicro()
	}
}
