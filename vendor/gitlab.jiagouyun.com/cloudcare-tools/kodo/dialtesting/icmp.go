package dialtesting

import (
	"encoding/json"
	"fmt"
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

type IcmpSuccess struct {
	PacketLossPercent float64 `json:"packet_loss_percent,omitempty"`
	ResponseTime      string  `json:"response_time,omitempty"`
	respTime          time.Duration
}

type IcmpTask struct {
	Host            string            `json:"host"`
	PacketCount     int               `json:"packet_count"`
	Timeout         string            `json:"timeout"`
	SuccessWhen     []*IcmpSuccess    `json:"success_when"`
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

	packetLossPercent float64
	avgRoundTripTime  float64 // ms
	minRoundTripTime  float64
	maxRoundTripTime  float64
	originBytes       []byte
	reqError          string
	timeout           time.Duration
	ticker            *time.Ticker
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
		if checker.ResponseTime != "" {
			du, err := time.ParseDuration(checker.ResponseTime)
			if err != nil {
				return err
			}
			checker.respTime = du
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

func (t *IcmpTask) CheckResult() (reasons []string) {
	for _, chk := range t.SuccessWhen {
		// check response time
		if chk.respTime > 0 && t.avgRoundTripTime > float64(chk.respTime.Milliseconds()) {
			reasons = append(reasons,
				fmt.Sprintf("ICMP average round-trip time (%v ms) larger than %v", t.avgRoundTripTime, chk.respTime))
		}

		// check packet loss
		if chk.PacketLossPercent > 0 && t.packetLossPercent > chk.PacketLossPercent {
			reasons = append(reasons,
				fmt.Sprintf("ICMP packet loss %v, larger than %v", t.packetLossPercent, chk.PacketLossPercent))
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
		"max_round_trip_time_in_millis":     t.maxRoundTripTime,
		"packet_loss_percent":               t.packetLossPercent,
		"success":                           int64(-1),
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
		message["average_round_trip_time_in_millis"] = t.avgRoundTripTime
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

func (t *IcmpTask) MetricName() string {
	return `icmp_dial_testing`
}

func (t *IcmpTask) Clear() {
	if t.timeout == 0 {
		t.timeout = PING_TIMEOUT
	}

	timeout := float64(t.timeout.Milliseconds())
	t.avgRoundTripTime = timeout
	t.minRoundTripTime = timeout
	t.maxRoundTripTime = timeout

	t.packetLossPercent = 100
	t.reqError = ""
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

	pinger.Timeout = t.timeout

	pinger.OnFinish = func(stats *ping.Statistics) {
		if stats.PacketLoss != 100 {
			t.packetLossPercent = stats.PacketLoss
			t.minRoundTripTime = t.round(float64(stats.MinRtt.Nanoseconds())/1e6, 3)
			t.avgRoundTripTime = t.round(float64(stats.AvgRtt.Nanoseconds())/1e6, 3)
			t.maxRoundTripTime = t.round(float64(stats.MaxRtt.Nanoseconds())/1e6, 3)
		}

	}

	if err := pinger.Run(); err != nil {
		t.reqError = err.Error()
		return err
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
