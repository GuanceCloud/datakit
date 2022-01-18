package dialtesting

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
)

type TcpSuccess struct {
	ResponseTime string `json:"response_time,omitempty"`
	respTime     time.Duration
}

type TcpTask struct {
	Host            string            `json:"dest_host"`
	Port            string            `json:"dest_port"`
	Timeout         string            `json:"timeout"`
	SuccessWhen     []*TcpSuccess     `json:"success_when"`
	ExternalID      string            `json:"external_id"`
	Name            string            `json:"name"`
	AK              string            `json:"access_key"`
	PostURL         string            `json:"post_url"`
	CurStatus       string            `json:"status"`
	Frequency       string            `json:"frequency"`
	Region          string            `json:"region"`
	OwnerExternalID string            `json:"owner_external_id"`
	Tags            map[string]string `json:"tags,omitempty"`
	Labels          []string          `json:"labels,omitempty"`
	UpdateTime      int64             `json:"update_time,omitempty"`

	tcpAddr  string
	reqCost  time.Duration
	reqError string
	timeout  time.Duration
	ticker   *time.Ticker
}

func (t *TcpTask) Init() error {

	if len(t.Timeout) == 0 {
		t.timeout = 10 * time.Second
	} else {
		if timeout, err := time.ParseDuration(t.Timeout); err != nil {
			return err
		} else {
			t.timeout = timeout
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

	t.tcpAddr = net.JoinHostPort(t.Host, t.Port)

	if strings.ToLower(t.CurStatus) == StatusStop {
		return nil
	}

	if len(t.SuccessWhen) == 0 {
		return fmt.Errorf(`no any check rule`)
	}

	for _, checker := range t.SuccessWhen {
		if checker.ResponseTime != "" {
			du, err := time.ParseDuration(checker.ResponseTime)
			if err != nil {
				return err
			}
			checker.respTime = du
		}

	}

	return nil
}

func (t *TcpTask) Check() error {
	if t.ExternalID == "" {
		return fmt.Errorf("external ID missing")
	}

	if len(t.Host) == 0 {
		return fmt.Errorf("host should not be empty")
	}

	if len(t.Port) == 0 {
		return fmt.Errorf("port should not be empty")
	}

	return t.Init()
}

func (t *TcpTask) CheckResult() (reasons []string) {
	for _, chk := range t.SuccessWhen {
		// check response time
		if t.reqCost > chk.respTime && chk.respTime > 0 {
			reasons = append(reasons,
				fmt.Sprintf("HTTP response time(%v) larger than %v", t.reqCost, chk.respTime))
		}
	}

	return
}

func (t *TcpTask) GetResults() (tags map[string]string, fields map[string]interface{}) {
	tags = map[string]string{
		"name":      t.Name,
		"dest_host": t.Host,
		"dest_port": t.Port,
		"status":    "FAIL",
	}

	responseTime := int64(t.reqCost) / 1000 // us

	fields = map[string]interface{}{
		"response_time": responseTime,
		"success":       int64(-1),
	}

	for k, v := range t.Tags {
		tags[k] = v
	}

	message := map[string]interface{}{}

	reasons := t.CheckResult()
	if t.reqError != "" {
		reasons = append(reasons, t.reqError)
	}

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

func (t *TcpTask) MetricName() string {
	return `tcp_dial_testing`
}

func (t *TcpTask) Clear() {
	t.reqCost = 0
	t.reqError = ""
}

func (t *TcpTask) Run() error {
	t.Clear()

	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
	defer cancel()
	start := time.Now()
	conn, err := d.DialContext(ctx, "tcp", t.tcpAddr)

	if err != nil {
		t.reqError = err.Error()
		return err
	}

	defer conn.Close()

	t.reqCost = time.Since(start)

	return nil
}

func (t *TcpTask) Stop() error {
	return nil
}

func (t *TcpTask) UpdateTimeUs() int64 {
	return t.UpdateTime
}

func (t *TcpTask) ID() string {
	if t.ExternalID == `` {
		return cliutils.XID("dtst_")
	}
	return fmt.Sprintf("%s_%s", t.AK, t.ExternalID)
}

func (t *TcpTask) GetOwnerExternalID() string {
	return t.OwnerExternalID
}

func (t *TcpTask) SetOwnerExternalID(exid string) {
	t.OwnerExternalID = exid
}

func (t *TcpTask) SetRegionId(regionId string) {
	t.Region = regionId
}

func (t *TcpTask) SetAk(ak string) {
	t.AK = ak
}

func (t *TcpTask) SetStatus(status string) {
	t.CurStatus = status
}

func (t *TcpTask) SetUpdateTime(ts int64) {
	t.UpdateTime = ts
}

func (t *TcpTask) Status() string {
	return t.CurStatus
}

func (t *TcpTask) Ticker() *time.Ticker {
	return t.ticker
}

func (t *TcpTask) Class() string {
	return ClassTCP
}

func (t *TcpTask) GetFrequency() string {
	return t.Frequency
}

func (t *TcpTask) GetLineData() string {
	return ""
}

func (t *TcpTask) RegionName() string {
	return t.Region
}

func (t *TcpTask) PostURLStr() string {
	return t.PostURL
}

func (t *TcpTask) AccessKey() string {
	return t.AK
}
