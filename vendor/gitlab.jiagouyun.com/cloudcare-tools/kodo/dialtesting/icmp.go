package dialtesting

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/go-ping/ping"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
)

const PING_TIMEOUT = 3 * time.Second

type ICMP struct {
	Type        uint8
	Code        uint8
	Checksum    uint16
	Identifier  uint16
	SequenceNum uint16
}

type ResponseTimeSucess struct {
	Func   string `json:"func,omitempty"`
	Op     string `json:"op,omitempty"`
	Target string `json:"target,omitempty"`

	target float64
}

type IcmpSuccess struct {
	PacketLossPercent []*ValueSuccess       `json:"packet_loss_percent,omitempty"`
	ResponseTime      []*ResponseTimeSucess `json:"response_time,omitempty"`
	Hops              []*ValueSuccess       `json:"hops,omitempty"`
	Packets           []*ValueSuccess       `json:"packets,omitempty"`
}

type IcmpTask struct {
	Host             string            `json:"host"`
	PacketCount      int               `json:"packet_count"`
	Timeout          string            `json:"timeout"`
	EnableTraceroute bool              `json:"enable_traceroute"`
	TracerouteConfig *TracerouteOption `json:"traceroute_config"`
	SuccessWhen      []*IcmpSuccess    `json:"success_when"`
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

	packetLossPercent float64
	avgRoundTripTime  float64 // ms
	minRoundTripTime  float64
	maxRoundTripTime  float64
	stdRoundTripTime  float64
	originBytes       []byte
	reqError          string
	sentPackets       int
	recvPackets       int
	timeout           time.Duration
	ticker            *time.Ticker
	traceroute        []*Route
}

func (t *IcmpTask) InitDebug() error {
	if len(t.Timeout) == 0 {
		t.timeout = PING_TIMEOUT
	} else {
		if timeout, err := time.ParseDuration(t.Timeout); err != nil {
			return err
		} else {
			t.timeout = timeout
		}
	}

	if strings.ToLower(t.CurStatus) == StatusStop {
		return nil
	}

	if len(t.SuccessWhen) == 0 {
		return fmt.Errorf(`no any check rule`)
	}

	if t.PacketCount <= 0 {
		t.PacketCount = 3
	}

	for _, checker := range t.SuccessWhen {
		if checker.ResponseTime != nil {
			for _, resp := range checker.ResponseTime {
				du, err := time.ParseDuration(resp.Target)
				if err != nil {
					return err
				}
				resp.target = float64(du.Nanoseconds() / 1e6) // ms
			}
		}

	}

	t.originBytes = make([]byte, 2000)

	return nil
}

func (t *IcmpTask) Init() error {
	if len(t.Timeout) == 0 {
		t.timeout = PING_TIMEOUT
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

	if strings.ToLower(t.CurStatus) == StatusStop {
		return nil
	}

	if len(t.SuccessWhen) == 0 {
		return fmt.Errorf(`no any check rule`)
	}

	if t.PacketCount <= 0 {
		t.PacketCount = 3
	}

	for _, checker := range t.SuccessWhen {
		if checker.ResponseTime != nil {
			for _, resp := range checker.ResponseTime {
				du, err := time.ParseDuration(resp.Target)
				if err != nil {
					return err
				}
				resp.target = float64(du.Nanoseconds() / 1e6) // ms
			}
		}

	}

	t.originBytes = make([]byte, 2000)

	return nil
}

func (t *IcmpTask) Check() error {
	if t.ExternalID == "" {
		return fmt.Errorf("external ID missing")
	}

	if len(t.Host) == 0 {
		return fmt.Errorf("host should not be empty")
	}

	return t.Init()
}

func (t *IcmpTask) CheckResult() (reasons []string, succFlag bool) {
	for _, chk := range t.SuccessWhen {
		// check response time
		for _, v := range chk.ResponseTime {
			vs := &ValueSuccess{
				Op:     v.Op,
				Target: v.target,
			}

			checkVal := float64(0)

			switch v.Func {
			case "avg":
				checkVal = t.avgRoundTripTime
			case "min":
				checkVal = t.minRoundTripTime
			case "max":
				checkVal = t.maxRoundTripTime
			case "std":
				checkVal = t.stdRoundTripTime
			}
			if checkVal > 0 {
				if err := vs.check(checkVal); err != nil {
					reasons = append(reasons,
						fmt.Sprintf("ICMP round-trip(%s) check failed: %s", v.Func, err.Error()))
				} else {
					succFlag = true
				}
			}
		}

		// check packet loss
		for _, v := range chk.PacketLossPercent {
			if err := v.check(t.packetLossPercent); err != nil {
				reasons = append(reasons, fmt.Sprintf("packet_loss_percent check failed: %s", err.Error()))
			} else {
				succFlag = true
			}
		}

		// check packets received
		for _, v := range chk.Packets {
			if err := v.check(float64(t.recvPackets)); err != nil {
				reasons = append(reasons, fmt.Sprintf("packets received check failed: %s", err.Error()))
			} else {
				succFlag = true
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

func (t *IcmpTask) GetResults() (tags map[string]string, fields map[string]interface{}) {
	tags = map[string]string{
		"name":      t.Name,
		"dest_host": t.Host,
		"status":    "FAIL",
	}

	fields = map[string]interface{}{
		"average_round_trip_time_in_millis": t.avgRoundTripTime,
		"min_round_trip_time_in_millis":     t.minRoundTripTime,
		"std_round_trip_time_in_millis":     t.stdRoundTripTime,
		"max_round_trip_time_in_millis":     t.maxRoundTripTime,
		"packet_loss_percent":               t.packetLossPercent,
		"packets_sent":                      t.sentPackets,
		"packets_received":                  t.recvPackets,
		"success":                           int64(-1),
	}

	for k, v := range t.Tags {
		tags[k] = v
	}

	if t.EnableTraceroute {
		if t.traceroute == nil {
			fields["traceroute"] = "[]"
		} else {
			tracerouteData, err := json.Marshal(t.traceroute)
			if err == nil && len(tracerouteData) > 0 {
				fields["traceroute"] = string(tracerouteData)
			} else {
				fields["traceroute"] = "[]"
			}
		}
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
			message["average_round_trip_time_in_millis"] = t.avgRoundTripTime
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

func (t *IcmpTask) MetricName() string {
	return `icmp_dial_testing`
}

func (t *IcmpTask) Clear() {
	if t.timeout == 0 {
		t.timeout = PING_TIMEOUT
	}

	t.avgRoundTripTime = 0
	t.minRoundTripTime = 0
	t.maxRoundTripTime = 0
	t.stdRoundTripTime = 0

	t.recvPackets = 0
	t.sentPackets = 0

	t.packetLossPercent = 100
	t.reqError = ""
	t.traceroute = nil
}

func (t *IcmpTask) Run() error {
	t.Clear()

	pinger, err := ping.NewPinger(t.Host)

	if err != nil {
		t.reqError = err.Error()
		return err
	}

	if t.PacketCount > 0 {
		pinger.Count = t.PacketCount
	} else {
		pinger.Count = 3
	}

	pinger.Interval = 1 * time.Second

	pinger.Timeout = time.Duration(pinger.Count) * pinger.Interval

	pinger.SetPrivileged(true)

	if err := pinger.Run(); err != nil {
		t.reqError = err.Error()
	} else {
		stats := pinger.Statistics()

		t.packetLossPercent = stats.PacketLoss
		t.sentPackets = stats.PacketsSent
		t.recvPackets = stats.PacketsRecv
		t.minRoundTripTime = t.round(float64(stats.MinRtt.Nanoseconds())/1e6, 3)
		t.avgRoundTripTime = t.round(float64(stats.AvgRtt.Nanoseconds())/1e6, 3)
		t.maxRoundTripTime = t.round(float64(stats.MaxRtt.Nanoseconds())/1e6, 3)
		t.stdRoundTripTime = t.round(float64(stats.StdDevRtt.Nanoseconds())/1e6, 3)
	}

	if t.EnableTraceroute {
		hostIP := net.ParseIP(t.Host)
		if hostIP == nil {
			if ips, err := net.LookupIP(t.Host); err != nil {
				t.reqError = err.Error()
				return err
			} else {
				if len(ips) == 0 {
					err := fmt.Errorf("invalid host: %s, found no ip record", t.Host)
					t.reqError = err.Error()
					return err
				} else {
					hostIP = ips[0]
				}
			}
		}
		routes, err := TracerouteIP(hostIP.String(), t.TracerouteConfig)
		if err != nil {
			t.reqError = err.Error()
		} else {
			t.traceroute = routes
		}
	}

	return nil
}

func (t *IcmpTask) round(num float64, n int) float64 {
	s := fmt.Sprintf("%."+strconv.Itoa(n)+"f", num)
	roundNum, _ := strconv.ParseFloat(s, 64)

	return roundNum
}

func (t *IcmpTask) Stop() error {
	return nil
}

func (t *IcmpTask) UpdateTimeUs() int64 {
	return t.UpdateTime
}

func (t *IcmpTask) ID() string {
	if t.ExternalID == `` {
		return cliutils.XID("dtst_")
	}
	return fmt.Sprintf("%s_%s", t.AK, t.ExternalID)
}

func (t *IcmpTask) GetOwnerExternalID() string {
	return t.OwnerExternalID
}

func (t *IcmpTask) SetOwnerExternalID(exid string) {
	t.OwnerExternalID = exid
}

func (t *IcmpTask) SetRegionId(regionId string) {
	t.Region = regionId
}

func (t *IcmpTask) SetAk(ak string) {
	t.AK = ak
}

func (t *IcmpTask) SetStatus(status string) {
	t.CurStatus = status
}

func (t *IcmpTask) SetUpdateTime(ts int64) {
	t.UpdateTime = ts
}

func (t *IcmpTask) Status() string {
	return t.CurStatus
}

func (t *IcmpTask) Ticker() *time.Ticker {
	return t.ticker
}

func (t *IcmpTask) Class() string {
	return ClassICMP
}

func (t *IcmpTask) GetFrequency() string {
	return t.Frequency
}

func (t *IcmpTask) GetLineData() string {
	return ""
}

func (t *IcmpTask) RegionName() string {
	return t.Region
}

func (t *IcmpTask) PostURLStr() string {
	return t.PostURL
}

func (t *IcmpTask) AccessKey() string {
	return t.AK
}

func (t *IcmpTask) CheckSum(data []byte) (rt uint16) {
	var (
		sum    uint32
		length int = len(data)
		index  int
	)
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}
	if length > 0 {
		sum += uint32(data[index]) << 8
	}

	rt = uint16(sum) + uint16(sum>>16)

	return ^rt
}
