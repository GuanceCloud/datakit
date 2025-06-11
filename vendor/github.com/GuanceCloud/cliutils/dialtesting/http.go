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
	"time"
)

var (
	_ TaskChild = (*HTTPTask)(nil)
	_ ITask     = (*HTTPTask)(nil)
)

const (
	MaxBodySize        = 10 * 1024
	DefaultHTTPTimeout = 60 * time.Second
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
	rawURL         string

	destIP string
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

	if t.reqBody != nil {
		t.reqBody.bodyType = t.reqBody.BodyType
	}
}

func (t *HTTPTask) stop() {
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
		"url":     t.rawURL,
		"status":  "FAIL",
		"method":  t.Method,
		"dest_ip": t.destIP,
	}

	if t.req != nil {
		tags["proto"] = t.req.Proto
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
	// TODO: 支持更多的 auth 选项
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
}

type HTTPSecret struct {
	NoSaveResponseBody bool `json:"not_save,omitempty"`
}

func (t *HTTPTask) run() error {
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

	t.req.Header.Add("Connection", "close")

	if agentInfo, ok := t.GetOption()["userAgent"]; ok {
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
		writer.Close()
		requestBody.bodyType = writer.FormDataContentType()
		body = buf
	} else {
		if requestBody.Body != "" {
			body = bytes.NewBufferString(requestBody.Body)
		}
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

	// setup HTTP client
	t.cli = &http.Client{
		Timeout: httpTimeout,
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

		// TLS opotions
		if opt.Certificate != nil { // see https://venilnoronha.io/a-step-by-step-guide-to-mtls-in-go
			if opt.Certificate.IgnoreServerCertificateError {
				t.cli.Transport = &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: opt.Certificate.IgnoreServerCertificateError, //nolint:gosec
					},
				}
			} else if opt.Certificate.CaCert != "" {
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

func (t *HTTPTask) beforeFirstRender() {
	t.rawURL = t.URL
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
