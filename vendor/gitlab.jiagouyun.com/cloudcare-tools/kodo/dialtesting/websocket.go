package dialtesting

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
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
	URL              string                  `json:"url"`
	Message          string                  `json:"message"`
	SuccessWhen      []*WebsocketSuccess     `json:"success_when"`
	AdvanceOptions   *WebsocketAdvanceOption `json:"advance_options,omitempty"`
	SuccessWhenLogic string                  `json:"success_when_logic"`
	ExternalID       string                  `json:"external_id"`
	Name             string                  `json:"name"`
	AK               string                  `json:"access_key"`
	PostURL          string                  `json:"post_url"`
	CurStatus        string                  `json:"status"`
	Frequency        string                  `json:"frequency"`
	Region           string                  `json:"region"`
	OwnerExternalID  string                  `json:"owner_external_id"`
	Tags             map[string]string       `json:"tags,omitempty"`
	Labels           []string                `json:"labels,omitempty"`
	UpdateTime       int64                   `json:"update_time,omitempty"`

	reqCost         time.Duration
	reqDnsCost      time.Duration
	responseMessage string
	resp            *http.Response
	parsedURL       *url.URL
	hostname        string
	reqError        string
	timeout         time.Duration
	ticker          *time.Ticker
}

func (t *WebsocketTask) InitDebug() error {
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

	if strings.ToLower(t.CurStatus) == StatusStop {
		return nil
	}

	if len(t.SuccessWhen) == 0 {
		return fmt.Errorf(`no any check rule`)
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

	}

	if parsedURL, err := url.Parse(t.URL); err != nil {
		return err
	} else {
		t.parsedURL = parsedURL
		t.hostname = parsedURL.Hostname()
	}

	return nil
}

func (t *WebsocketTask) Init() error {
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

	if len(t.SuccessWhen) == 0 {
		return fmt.Errorf(`no any check rule`)
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

func (t *WebsocketTask) Check() error {
	if t.ExternalID == "" {
		return fmt.Errorf("external ID missing")
	}

	if len(t.URL) == 0 {
		return fmt.Errorf("URL should not be empty")
	}

	return t.Init()
}

func (t *WebsocketTask) CheckResult() (reasons []string, succFlag bool) {
	for _, chk := range t.SuccessWhen {
		// check response time
		if chk.ResponseTime != nil {
			for _, v := range chk.ResponseTime {
				reqCost := t.reqCost
				if v.IsContainDNS {
					reqCost += t.reqDnsCost
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

	return
}

func (t *WebsocketTask) GetResults() (tags map[string]string, fields map[string]interface{}) {
	tags = map[string]string{
		"name":   t.Name,
		"url":    t.URL,
		"status": "FAIL",
		"proto":  "websocket",
	}

	responseTime := int64(t.reqCost+t.reqDnsCost) / 1000        // us
	responseTimeWithDNS := int64(t.reqCost+t.reqDnsCost) / 1000 // us

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
		} else {
			message["response_time_in_micros"] = responseTime
		}

		if t.reqError == "" && len(reasons) == 0 {
			tags["status"] = "OK"
			fields["success"] = int64(1)
		}
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

	return
}

func (t *WebsocketTask) MetricName() string {
	return `websocket_dial_testing`
}

func (t *WebsocketTask) Clear() {
	t.reqCost = 0
	t.reqError = ""
}

func (t *WebsocketTask) Run() error {
	t.Clear()

	ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
	defer cancel()

	hostIP := net.ParseIP(t.hostname)

	if hostIP == nil { // host name
		start := time.Now()
		if ips, err := net.LookupIP(t.hostname); err != nil {
			t.reqError = err.Error()
			return err
		} else {
			if len(ips) == 0 {
				err := fmt.Errorf("invalid host: %s, found no ip record", t.hostname)
				t.reqError = err.Error()
				return err
			} else {
				t.reqDnsCost = time.Since(start)
				hostIP = ips[0] // TODO: support mutiple ip for one host
			}
		}
	}

	header := t.getHeader()

	t.parsedURL.Host = net.JoinHostPort(hostIP.String(), t.parsedURL.Port())

	//ingore tls verify
	if t.parsedURL.Scheme == "wss" {
		websocket.DefaultDialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	start := time.Now()

	c, resp, err := websocket.DefaultDialer.DialContext(ctx, t.parsedURL.String(), header)

	if err != nil {
		t.reqError = err.Error()
		t.reqDnsCost = 0
		return err
	}

	t.reqCost = time.Since(start)
	defer c.Close()

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

func (t *WebsocketTask) Stop() error {
	return nil
}

func (t *WebsocketTask) UpdateTimeUs() int64 {
	return t.UpdateTime
}

func (t *WebsocketTask) ID() string {
	if t.ExternalID == `` {
		return cliutils.XID("dtst_")
	}
	return fmt.Sprintf("%s_%s", t.AK, t.ExternalID)
}

func (t *WebsocketTask) GetOwnerExternalID() string {
	return t.OwnerExternalID
}

func (t *WebsocketTask) SetOwnerExternalID(exid string) {
	t.OwnerExternalID = exid
}

func (t *WebsocketTask) SetRegionId(regionId string) {
	t.Region = regionId
}

func (t *WebsocketTask) SetAk(ak string) {
	t.AK = ak
}

func (t *WebsocketTask) SetStatus(status string) {
	t.CurStatus = status
}

func (t *WebsocketTask) SetUpdateTime(ts int64) {
	t.UpdateTime = ts
}

func (t *WebsocketTask) Status() string {
	return t.CurStatus
}

func (t *WebsocketTask) Ticker() *time.Ticker {
	return t.ticker
}

func (t *WebsocketTask) Class() string {
	return ClassWebsocket
}

func (t *WebsocketTask) GetFrequency() string {
	return t.Frequency
}

func (t *WebsocketTask) GetLineData() string {
	return ""
}

func (t *WebsocketTask) RegionName() string {
	return t.Region
}

func (t *WebsocketTask) PostURLStr() string {
	return t.PostURL
}

func (t *WebsocketTask) AccessKey() string {
	return t.AK
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
