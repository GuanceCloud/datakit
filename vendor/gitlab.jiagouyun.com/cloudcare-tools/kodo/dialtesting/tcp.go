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

type TcpResponseTime struct {
	IsContainDNS bool   `json:"is_contain_dns"`
	Target       string `json:"target"`

	targetTime time.Duration
}

type TcpSuccess struct {
	ResponseTime []*TcpResponseTime `json:"response_time,omitempty"`
	Hops         []*ValueSuccess    `json:"hops,omitempty"`
}

type TcpTask struct {
	Host             string            `json:"host"`
	Port             string            `json:"port"`
	Timeout          string            `json:"timeout"`
	EnableTraceroute bool              `json:"enable_traceroute"`
	TracerouteConfig *TracerouteOption `json:"traceroute_config"`
	SuccessWhen      []*TcpSuccess     `json:"success_when"`
	SuccessWhenLogic string            `json:"success_when_logic"`
	ExternalID       string            `json:"external_id"`
	Name             string            `json:"name"`
	AK               string            `json:"access_key"`
	PostURL          string            `json:"post_url"`
	CurStatus        string            `json:"status"`
	Frequency        string            `json:"frequency"`
	Region           string            `json:"region"`
	OwnerExternalID  string            `json:"owner_external_id"`
	Tags             map[string]string `json:"tags,omitempty"`
	Labels           []string          `json:"labels,omitempty"`
	UpdateTime       int64             `json:"update_time,omitempty"`

	reqCost    time.Duration
	reqDnsCost time.Duration
	reqError   string
	timeout    time.Duration
	ticker     *time.Ticker
	traceroute []*Route
}

func (t *TcpTask) InitDebug() error {

	return t.init(true)
}

func (t *TcpTask) init(debug bool) error {
	if len(t.Timeout) == 0 {
		t.timeout = 10 * time.Second
	} else {
		if timeout, err := time.ParseDuration(t.Timeout); err != nil {
			return err
		} else {
			t.timeout = timeout
		}
	}

	if !debug {
		du, err := time.ParseDuration(t.Frequency)
		if err != nil {
			return err
		}
		if t.ticker != nil {
			t.ticker.Stop()
		}
		t.ticker = time.NewTicker(du)
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

		// if [checker.Hops] is not nil, set traceroute to be true
		if checker.Hops != nil {
			t.EnableTraceroute = true
		}

	}

	return nil
}

func (t *TcpTask) Init() error {

	return t.init(false)
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

func (t *TcpTask) CheckResult() (reasons []string, succFlag bool) {
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
						fmt.Sprintf("TCP response time(%v) larger than %v", reqCost, v.targetTime))
				} else if v.targetTime > 0 {
					succFlag = true
				}
			}
		}

		// check traceroute
		if t.EnableTraceroute {
			hops := float64(len(t.traceroute))

			if hops == 0 {
				reasons = append(reasons, "traceroute failed with no hops")
			} else {
				for _, v := range chk.Hops {
					if err := v.check(hops); err != nil {
						reasons = append(reasons, fmt.Sprintf("traceroute hops check failed: %s", err.Error()))
					} else {
						succFlag = true
					}
				}
			}

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
		"proto":     "tcp",
	}

	responseTime := int64(t.reqCost) / 1000                     // us
	responseTimeWithDNS := int64(t.reqCost+t.reqDnsCost) / 1000 // us

	fields = map[string]interface{}{
		"response_time":          responseTime,
		"response_time_with_dns": responseTimeWithDNS,
		"success":                int64(-1),
	}

	if t.EnableTraceroute {
		fields["hops"] = 0
		if t.traceroute == nil {
			fields["traceroute"] = "[]"
		} else {
			tracerouteData, err := json.Marshal(t.traceroute)
			if err == nil && len(tracerouteData) > 0 {
				fields["traceroute"] = string(tracerouteData)
				fields["hops"] = len(t.traceroute)
			} else {
				fields["traceroute"] = "[]"
			}
		}
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

func (t *TcpTask) MetricName() string {
	return `tcp_dial_testing`
}

func (t *TcpTask) Clear() {
	t.reqCost = 0
	t.reqError = ""
	t.traceroute = nil
}

func (t *TcpTask) Run() error {
	t.Clear()

	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
	defer cancel()

	hostIP := net.ParseIP(t.Host)

	if hostIP == nil { // host name
		start := time.Now()
		if ips, err := net.LookupIP(t.Host); err != nil {
			t.reqError = err.Error()
			return err
		} else {
			if len(ips) == 0 {
				err := fmt.Errorf("invalid host: %s, found no ip record", t.Host)
				t.reqError = err.Error()
				return err
			} else {
				t.reqDnsCost = time.Since(start)
				hostIP = ips[0] // TODO: support mutiple ip for one host
			}
		}
	}

	tcpIPAddr := net.JoinHostPort(hostIP.String(), t.Port)

	start := time.Now()
	conn, err := d.DialContext(ctx, "tcp", tcpIPAddr)

	if err != nil {
		t.reqError = err.Error()
		t.reqDnsCost = 0
	} else {
		t.reqCost = time.Since(start)
		conn.Close()
	}

	if t.EnableTraceroute {
		routes, err := TracerouteIP(hostIP.String(), t.TracerouteConfig)
		if err != nil {
			t.reqError = err.Error()
		} else {
			t.traceroute = routes
		}
	}

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
