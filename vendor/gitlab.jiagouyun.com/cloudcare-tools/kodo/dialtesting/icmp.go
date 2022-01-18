package dialtesting

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
)

const (
	MAX_PG = 2000 // max packet size
)

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
	Host            string            `json:"dest_host"`
	PacketSize      int               `json:"packet_size"`
	PacketCount     int               `json:"packet_count"`
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
	ticker            *time.Ticker
}

func (t *IcmpTask) Init() error {
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

	// check packet size
	if t.PacketSize == 0 {
		t.PacketSize = 48
	} else if t.PacketSize > MAX_PG {
		t.PacketSize = MAX_PG
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

	if t.PacketSize > MAX_PG {
		return fmt.Errorf("packet size is larger than %v", MAX_PG)
	}

	return t.Init()
}

func (t *IcmpTask) CheckResult() (reasons []string) {
	for _, chk := range t.SuccessWhen {
		// check response time
		if chk.respTime > 0 && t.avgRoundTripTime > float64(chk.respTime.Milliseconds()) {
			reasons = append(reasons,
				fmt.Sprintf("ICMP average round-trip time (%v) larger than %v", t.avgRoundTripTime, chk.respTime))
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
	t.avgRoundTripTime = 0
	t.minRoundTripTime = 0
	t.maxRoundTripTime = 0
	t.packetLossPercent = 0
	t.reqError = ""
}

func (t *IcmpTask) Run() error {
	t.Clear()

	var (
		icmp                                                 ICMP
		laddr                                                = net.IPAddr{IP: net.ParseIP("0.0.0.0")}
		raddr, _                                             = net.ResolveIPAddr("ip", t.Host)
		maxRoundTripTime, minRoundTripTime, avgRoundTripTime float64
	)

	ipType := "ip4"
	if raddr.IP.To4() == nil {
		ipType = "ip6"
	}

	conn, err := net.DialIP(ipType+":icmp", &laddr, raddr)

	if err != nil {
		t.reqError = err.Error()
		return err
	}

	defer conn.Close()

	icmp = ICMP{8, 0, 0, 0, 0}

	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, icmp)
	binary.Write(&buffer, binary.BigEndian, t.originBytes[0:t.PacketSize])
	b := buffer.Bytes()

	binary.BigEndian.PutUint16(b[2:], t.CheckSum(b))

	recv := make([]byte, 1024)
	roundTripTimeList := []float64{}

	dropPackCount := 0.0
	maxRoundTripTime = 0.0
	minRoundTripTime = 0.0
	avgRoundTripTime = 0.0

	for i := t.PacketCount; i > 0; i-- {
		if _, err := conn.Write(buffer.Bytes()); err != nil {
			dropPackCount++
			time.Sleep(time.Second)
			continue
		}
		startTime := time.Now()
		conn.SetReadDeadline((time.Now().Add(time.Second * 3)))
		_, err := conn.Read(recv)

		if err != nil {
			dropPackCount++
			time.Sleep(time.Second)
			continue
		}
		endTime := time.Now()
		dur := float64(endTime.Sub(startTime).Nanoseconds()) / 1e6 // ms
		roundTripTimeList = append(roundTripTimeList, dur)

		if dur > maxRoundTripTime {
			maxRoundTripTime = dur
		}

		if minRoundTripTime == 0 {
			minRoundTripTime = dur
		} else if dur < minRoundTripTime {
			minRoundTripTime = dur
		}
		time.Sleep(time.Second)
	}
	t.packetLossPercent = dropPackCount / float64(t.PacketCount)
	if len(roundTripTimeList) == 0 {
		avgRoundTripTime = 3000.0
	} else {
		sum := 0.0
		for _, n := range roundTripTimeList {
			sum += n
		}
		avgRoundTripTime = sum / float64(len(roundTripTimeList))
	}
	t.minRoundTripTime = t.round(minRoundTripTime, 3)
	t.avgRoundTripTime = t.round(avgRoundTripTime, 3)
	t.maxRoundTripTime = t.round(maxRoundTripTime, 3)

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
