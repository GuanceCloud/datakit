// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
)

var (
	_ TaskChild = (*WebsocketTask)(nil)
	_ ITask     = (*WebsocketTask)(nil)
)

type WebsocketResponseTime struct {
	IsContainDNS bool   `json:"is_contain_dns"`
	Target       string `json:"target"`

	targetTime time.Duration
}

type WebsocketSuccess struct {
	ResponseTime    []*WebsocketResponseTime    `json:"response_time,omitempty"`
	ResponseMessage []*SuccessOption            `json:"response_message,omitempty"`
	Header          map[string][]*SuccessOption `json:"header,omitempty"`
}

type WebsocketOptRequest struct {
	Timeout string            `json:"timeout,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type WebsocketOptAuth struct {
	// basic auth
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type WebsocketAdvanceOption struct {
	RequestOptions *WebsocketOptRequest `json:"request_options,omitempty"`
	Auth           *WebsocketOptAuth    `json:"auth,omitempty"`
}

type WebsocketTask struct {
	*Task
	URL              string                  `json:"url"`
	Message          string                  `json:"message"`
	SuccessWhen      []*WebsocketSuccess     `json:"success_when"`
	AdvanceOptions   *WebsocketAdvanceOption `json:"advance_options,omitempty"`
	SuccessWhenLogic string                  `json:"success_when_logic"`

	reqCost         time.Duration
	reqDNSCost      time.Duration
	responseMessage string
	resp            *http.Response
	parsedURL       *url.URL
	hostname        string
	reqError        string
	timeout         time.Duration

	rawTask *WebsocketTask
}

func (t *WebsocketTask) init() error {
	t.timeout = 30 * time.Second
	if t.AdvanceOptions != nil {
		if t.AdvanceOptions.RequestOptions != nil && len(t.AdvanceOptions.RequestOptions.Timeout) > 0 {
			if timeout, err := time.ParseDuration(t.AdvanceOptions.RequestOptions.Timeout); err != nil {
				return err
			} else {
				t.timeout = timeout
			}
		}
	}

	if len(t.SuccessWhen) == 0 {
		return errors.New(`no any check rule`)
	}

	for _, checker := range t.SuccessWhen {
		if checker.ResponseTime != nil {
			for _, v := range checker.ResponseTime {
				du, err := time.ParseDuration(v.Target)
				if err != nil {
					return err
				}
				v.targetTime = du
			}
		}

		for _, vs := range checker.Header {
			for _, v := range vs {
				err := genReg(v)
				if err != nil {
					return err
				}
			}
		}

		for _, v := range checker.ResponseMessage {
			err := genReg(v)
			if err != nil {
				return err
			}
		}
	}

	if parsedURL, err := url.Parse(t.URL); err != nil {
		return err
	} else {
		if parsedURL.Port() == "" {
			port := ""
			if parsedURL.Scheme == "wss" {
				port = "443"
			} else if parsedURL.Scheme == "ws" {
				port = "80"
			}
			parsedURL.Host = net.JoinHostPort(parsedURL.Host, port)
		}
		t.parsedURL = parsedURL
		t.hostname = parsedURL.Hostname()
	}

	return nil
}

func (t *WebsocketTask) check() error {
	if len(t.URL) == 0 {
		return errors.New("URL should not be empty")
	}

	return nil
}

func (t *WebsocketTask) checkResult() (reasons []string, succFlag bool) {
	for _, chk := range t.SuccessWhen {
		// check response time
		if chk.ResponseTime != nil {
			for _, v := range chk.ResponseTime {
				reqCost := t.reqCost
				if v.IsContainDNS {
					reqCost += t.reqDNSCost
				}

				if reqCost > v.targetTime && v.targetTime > 0 {
					reasons = append(reasons,
						fmt.Sprintf("response time(%v) larger than %v", reqCost, v.targetTime))
				} else if v.targetTime > 0 {
					succFlag = true
				}
			}
		}

		// check message
		if chk.ResponseMessage != nil {
			for _, v := range chk.ResponseMessage {
				if err := v.check(t.responseMessage, "response message"); err != nil {
					reasons = append(reasons, err.Error())
				} else {
					succFlag = true
				}
			}
		}

		// check header
		if t.resp != nil {
			for k, vs := range chk.Header {
				for _, v := range vs {
					if err := v.check(t.resp.Header.Get(k), fmt.Sprintf("Websocket header `%s'", k)); err != nil {
						reasons = append(reasons, err.Error())
					} else {
						succFlag = true
					}
				}
			}
		}
	}

	return reasons, succFlag
}

func (t *WebsocketTask) getResults() (tags map[string]string, fields map[string]interface{}) {
	tags = map[string]string{
		"name":   t.Name,
		"url":    t.URL,
		"status": "FAIL",
		"proto":  "websocket",
	}

	responseTime := int64(t.reqCost+t.reqDNSCost) / 1000        // us
	responseTimeWithDNS := int64(t.reqCost+t.reqDNSCost) / 1000 // us

	fields = map[string]interface{}{
		"response_time":          responseTime,
		"response_time_with_dns": responseTimeWithDNS,
		"response_message":       t.responseMessage,
		"sent_message":           t.Message,
		"success":                int64(-1),
	}

	for k, v := range t.Tags {
		tags[k] = v
	}

	message := map[string]interface{}{}

	reasons, succFlag := t.checkResult()
	if t.reqError != "" {
		reasons = append(reasons, t.reqError)
	}

	switch t.SuccessWhenLogic {
	case "or":
		if succFlag && t.reqError == "" {
			tags["status"] = "OK"
			fields["success"] = int64(1)
			message["response_time"] = responseTime
		} else {
			message[`fail_reason`] = strings.Join(reasons, `;`)
			fields[`fail_reason`] = strings.Join(reasons, `;`)
		}
	default:
		if len(reasons) != 0 {
			message[`fail_reason`] = strings.Join(reasons, `;`)
			fields[`fail_reason`] = strings.Join(reasons, `;`)
		} else {
			message["response_time"] = responseTime
		}

		if t.reqError == "" && len(reasons) == 0 {
			tags["status"] = "OK"
			fields["success"] = int64(1)
		}
	}

	if v, ok := fields[`fail_reason`]; ok && len(v.(string)) != 0 && t.resp != nil {
		message[`response_header`] = t.resp.Header
	}

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

func (t *WebsocketTask) metricName() string {
	return `websocket_dial_testing`
}

func (t *WebsocketTask) clear() {
	t.reqCost = 0
	t.reqError = ""
}

func (t *WebsocketTask) run() error {
	ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
	defer cancel()

	hostIP := net.ParseIP(t.hostname)

	if hostIP == nil { // host name
		start := time.Now()
		if ips, err := net.LookupIP(t.hostname); err != nil {
			t.reqError = err.Error()
			return nil
		} else {
			if len(ips) == 0 {
				err := fmt.Errorf("invalid host: %s, found no ip record", t.hostname)
				t.reqError = err.Error()
				return nil
			} else {
				t.reqDNSCost = time.Since(start)
				hostIP = ips[0] // TODO: support mutiple ip for one host
			}
		}
	}

	header := t.getHeader()

	if len(header.Get("Host")) == 0 {
		// set default Host
		header.Add("Host", t.hostname)
	}

	t.parsedURL.Host = net.JoinHostPort(hostIP.String(), t.parsedURL.Port())

	if t.parsedURL.Scheme == "wss" {
		websocket.DefaultDialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // nolint:gosec
	}

	start := time.Now()

	c, resp, err := websocket.DefaultDialer.DialContext(ctx, t.parsedURL.String(), header)
	if err != nil {
		t.reqError = err.Error()
		t.reqDNSCost = 0
		return nil
	}

	t.reqCost = time.Since(start)
	defer func() {
		if err := c.Close(); err != nil {
			_ = err // pass
		}
	}()

	t.resp = resp

	t.getMessage(c)
	return nil
}

func (t *WebsocketTask) getMessage(c *websocket.Conn) {
	err := c.WriteMessage(websocket.TextMessage, []byte(t.Message))
	if err != nil {
		t.reqError = err.Error()
		return
	}

	if _, message, err := c.ReadMessage(); err != nil {
		t.reqError = err.Error()
		return
	} else {
		t.responseMessage = string(message)
	}

	// close error ignore
	_ = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

func (t *WebsocketTask) getHeader() http.Header {
	var header http.Header = make(http.Header)

	if t.AdvanceOptions != nil {
		if t.AdvanceOptions.RequestOptions != nil {
			for k, v := range t.AdvanceOptions.RequestOptions.Headers {
				header[k] = []string{v}
			}

			if t.AdvanceOptions.Auth != nil && len(t.AdvanceOptions.Auth.Username) > 0 && len(t.AdvanceOptions.Auth.Password) > 0 {
				header["Authorization"] = []string{"Basic " + basicAuth(t.AdvanceOptions.Auth.Username, t.AdvanceOptions.Auth.Password)}
			}
		}
	}

	return header
}

func (t *WebsocketTask) stop() {}

func (t *WebsocketTask) class() string {
	return ClassWebsocket
}

func (t *WebsocketTask) getHostName() ([]string, error) {
	if hostName, err := getHostName(t.URL); err != nil {
		return nil, err
	} else {
		return []string{hostName}, nil
	}
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (t *WebsocketTask) getVariableValue(variable Variable) (string, error) {
	return "", errors.New("not support")
}

func (t *WebsocketTask) getRawTask(taskString string) (string, error) {
	task := WebsocketTask{}

	if err := json.Unmarshal([]byte(taskString), &task); err != nil {
		return "", fmt.Errorf("unmarshal websocket task failed: %w", err)
	}

	task.Task = nil

	bytes, _ := json.Marshal(task)
	return string(bytes), nil
}

func (t *WebsocketTask) initTask() {
	if t.Task == nil {
		t.Task = &Task{}
	}
}

func (t *WebsocketTask) renderTemplate(fm template.FuncMap) error {
	if t.rawTask == nil {
		task := &WebsocketTask{}
		if err := t.NewRawTask(task); err != nil {
			return fmt.Errorf("new raw task failed: %w", err)
		}
		t.rawTask = task
	}

	task := t.rawTask
	if task == nil {
		return errors.New("raw task is nil")
	}

	// url
	if url, err := t.GetParsedString(task.URL, fm); err != nil {
		return fmt.Errorf("render url failed: %w", err)
	} else {
		t.URL = url
	}

	// message
	if message, err := t.GetParsedString(task.Message, fm); err != nil {
		return fmt.Errorf("render message failed: %w", err)
	} else {
		t.Message = message
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

func (t *WebsocketTask) renderAdvanceOptions(task *WebsocketTask, fm template.FuncMap) error {
	if task == nil || task.AdvanceOptions == nil {
		return nil
	}

	opt := task.AdvanceOptions

	// request options
	if err := t.renderRequestOptions(opt.RequestOptions, fm); err != nil {
		return fmt.Errorf("render request options failed: %w", err)
	}

	// auth
	if err := t.renderAuth(opt.Auth, fm); err != nil {
		return fmt.Errorf("render auth failed: %w", err)
	}

	return nil
}

func (t *WebsocketTask) renderAuth(auth *WebsocketOptAuth, fm template.FuncMap) error {
	if auth == nil {
		return nil
	}

	// username
	if text, err := t.GetParsedString(auth.Username, fm); err != nil {
		return fmt.Errorf("render auth username failed: %w", err)
	} else {
		t.AdvanceOptions.Auth.Username = text
	}

	// password
	if text, err := t.GetParsedString(auth.Password, fm); err != nil {
		return fmt.Errorf("render auth password failed: %w", err)
	} else {
		t.AdvanceOptions.Auth.Password = text
	}

	return nil
}

func (t *WebsocketTask) renderRequestOptions(requestOpt *WebsocketOptRequest, fm template.FuncMap) error {
	if requestOpt != nil {
		for k, v := range requestOpt.Headers {
			if text, err := t.GetParsedString(v, fm); err != nil {
				return fmt.Errorf("render header failed: %w", err)
			} else {
				t.AdvanceOptions.RequestOptions.Headers[k] = text
			}
		}
	}
	return nil
}

func (t *WebsocketTask) setReqError(err string) {
	t.reqError = err
}

func (t *WebsocketTask) renderSuccessWhen(task *WebsocketTask, fm template.FuncMap) error {
	if task == nil || task.SuccessWhen == nil {
		return nil
	}

	for index, success := range task.SuccessWhen {
		if success == nil {
			continue
		}

		for msgIndex, msg := range success.ResponseMessage {
			if msg == nil {
				continue
			}

			if err := t.renderSuccessOption(msg, t.SuccessWhen[index].ResponseMessage[msgIndex], fm); err != nil {
				return fmt.Errorf("render success when failed: %w", err)
			}
		}

		// header
		for headerIndex, v := range success.Header {
			for header, option := range v {
				if err := t.renderSuccessOption(option, t.SuccessWhen[index].Header[headerIndex][header], fm); err != nil {
					return fmt.Errorf("render header failed: %w", err)
				}
			}
		}
	}
	return nil
}
