// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

var (
	_ TaskChild = (*ICMPTask)(nil)
	_ ITask     = (*ICMPTask)(nil)
)

const (
	PingTimeout    = 3 * time.Second
	DefaultICMPTTL = 64
)

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

type ICMPSuccess struct {
	PacketLossPercent []*ValueSuccess       `json:"packet_loss_percent,omitempty"`
	ResponseTime      []*ResponseTimeSucess `json:"response_time,omitempty"`
	Hops              []*ValueSuccess       `json:"hops,omitempty"`
	Packets           []*ValueSuccess       `json:"packets,omitempty"`
}

type ICMPTask struct {
	*Task
	Host             string            `json:"host"`
	PacketCount      int               `json:"packet_count"`
	Timeout          string            `json:"timeout"`
	EnableTraceroute bool              `json:"enable_traceroute"`
	TracerouteConfig *TracerouteOption `json:"traceroute_config"`
	SuccessWhen      []*ICMPSuccess    `json:"success_when"`
	SuccessWhenLogic string            `json:"success_when_logic"`

	packetLossPercent float64
	avgRoundTripTime  float64 // us
	minRoundTripTime  float64 // us
	maxRoundTripTime  float64 // us
	stdRoundTripTime  float64 // us
	originBytes       []byte
	reqError          string
	sentPackets       int
	recvPackets       int
	timeout           time.Duration
	traceroute        []*Route

	rawTask *ICMPTask
}

func (t *ICMPTask) init() error {
	if len(t.Timeout) == 0 {
		t.timeout = PingTimeout
	} else {
		if timeout, err := time.ParseDuration(t.Timeout); err != nil {
			return err
		} else {
			t.timeout = timeout
		}
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
				resp.target = float64(du.Microseconds()) // us
			}
		}

		// if [checker.Hops] is not nil, set traceroute to be true
		if checker.Hops != nil {
			t.EnableTraceroute = true
		}

		t.EnableTraceroute = IsEnabledTraceroute(t.EnableTraceroute)
	}

	t.originBytes = make([]byte, 2000)

	return nil
}

func (t *ICMPTask) check() error {
	if len(t.Host) == 0 {
		return fmt.Errorf("host should not be empty")
	}

	return nil
}

func (t *ICMPTask) checkResult() (reasons []string, succFlag bool) {
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

			if t.packetLossPercent == 100 {
				reasons = append(reasons, "all packets lost")
			} else if err := vs.check(checkVal); err != nil {
				reasons = append(reasons,
					fmt.Sprintf("ICMP round-trip(%s) check failed: %s", v.Func, err.Error()))
			} else {
				succFlag = true
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

	return reasons, succFlag
}

func (t *ICMPTask) getResults() (tags map[string]string, fields map[string]interface{}) {
	tags = map[string]string{
		"name":      t.Name,
		"dest_host": t.Host,
		"status":    "FAIL",
		"proto":     "icmp",
	}

	fields = map[string]interface{}{
		"average_round_trip_time_in_millis": t.round(t.avgRoundTripTime/1000, 3),
		"average_round_trip_time":           t.avgRoundTripTime,
		"min_round_trip_time_in_millis":     t.round(t.minRoundTripTime/1000, 3),
		"min_round_trip_time":               t.minRoundTripTime,
		"std_round_trip_time_in_millis":     t.round(t.stdRoundTripTime/1000, 3),
		"std_round_trip_time":               t.stdRoundTripTime,
		"max_round_trip_time_in_millis":     t.round(t.maxRoundTripTime/1000, 3),
		"max_round_trip_time":               t.maxRoundTripTime,
		"packet_loss_percent":               t.packetLossPercent,
		"packets_sent":                      t.sentPackets,
		"packets_received":                  t.recvPackets,
		"success":                           int64(-1),
	}

	for k, v := range t.Tags {
		tags[k] = v
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
			message["average_round_trip_time"] = t.avgRoundTripTime
		} else {
			message[`fail_reason`] = strings.Join(reasons, `;`)
			fields[`fail_reason`] = strings.Join(reasons, `;`)
		}
	default:
		if len(reasons) != 0 {
			message[`fail_reason`] = strings.Join(reasons, `;`)
			fields[`fail_reason`] = strings.Join(reasons, `;`)
		} else {
			message["average_round_trip_time"] = t.avgRoundTripTime
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

	return tags, fields
}

func (t *ICMPTask) metricName() string {
	return `icmp_dial_testing`
}

func (t *ICMPTask) clear() {
	if t.timeout == 0 {
		t.timeout = PingTimeout
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

func (t *ICMPTask) run() error {
	count := 3

	if t.PacketCount > 0 {
		count = t.PacketCount
	}

	interval := 3 * time.Second
	if stats, err := pingTarget(t.Host, count, interval, t.timeout); err != nil {
		t.reqError = err.Error()
	} else {
		t.packetLossPercent = stats.PacketLoss
		t.sentPackets = stats.PacketsSent
		t.recvPackets = stats.PacketsRecv
		t.minRoundTripTime = t.round(float64(stats.MinRtt.Nanoseconds())/1e3, 3)
		t.avgRoundTripTime = t.round(float64(stats.AvgRtt.Nanoseconds())/1e3, 3)
		t.maxRoundTripTime = t.round(float64(stats.MaxRtt.Nanoseconds())/1e3, 3)
		t.stdRoundTripTime = t.round(float64(stats.StdDevRtt.Nanoseconds())/1e3, 3)
	}

	if t.EnableTraceroute {
		hostIP := net.ParseIP(t.Host)
		if hostIP == nil {
			if ips, err := net.LookupIP(t.Host); err != nil {
				t.reqError = err.Error()
				return nil
			} else {
				if len(ips) == 0 {
					err := fmt.Errorf("invalid host: %s, found no ip record", t.Host)
					t.reqError = err.Error()
					return nil
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

func (t *ICMPTask) round(num float64, n int) float64 {
	s := fmt.Sprintf("%."+strconv.Itoa(n)+"f", num)
	roundNum, _ := strconv.ParseFloat(s, 64)

	return roundNum
}

func (t *ICMPTask) stop() {}

func (t *ICMPTask) class() string {
	return ClassICMP
}

func (t *ICMPTask) getHostName() ([]string, error) {
	return []string{t.Host}, nil
}

func (t *ICMPTask) getVariableValue(variable Variable) (string, error) {
	return "", fmt.Errorf("not support")
}

func (t *ICMPTask) getRawTask(taskString string) (string, error) {
	task := ICMPTask{}

	if err := json.Unmarshal([]byte(taskString), &task); err != nil {
		return "", fmt.Errorf("unmarshal icmp task failed: %w", err)
	}

	task.Task = nil

	bytes, _ := json.Marshal(task)
	return string(bytes), nil
}

func (t *ICMPTask) initTask() {
	if t.Task == nil {
		t.Task = &Task{}
	}
}

func (t *ICMPTask) renderTemplate(fm template.FuncMap) error {
	if t.rawTask == nil {
		task := &ICMPTask{}
		if err := t.NewRawTask(task); err != nil {
			return fmt.Errorf("new raw task failed: %w", err)
		}
		t.rawTask = task
	}

	task := t.rawTask
	if task == nil {
		return fmt.Errorf("raw task is nil")
	}

	// host
	if text, err := t.GetParsedString(task.Host, fm); err != nil {
		return fmt.Errorf("render host failed: %w", err)
	} else {
		t.Host = text
	}

	return nil
}

// The following code is copied from https://github.com/prometheus/blackbox_exporter/blob/main/prober/icmp/icmp.go
// and modified to adapt to the dialtesting package.

var (
	icmpID            int
	icmpSequence      uint16
	icmpSequenceMutex sync.Mutex
)

func init() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// PID is typically 1 when running in a container; in that case, set
	// the ICMP echo ID to a random value to avoid potential clashes with
	// other blackbox_exporter instances. See #411.
	if pid := os.Getpid(); pid == 1 {
		icmpID = r.Intn(1 << 16)
	} else {
		icmpID = pid & 0xffff
	}

	// Start the ICMP echo sequence at a random offset to prevent them from
	// being in sync when several blackbox_exporter instances are restarted
	// at the same time. See #411.
	icmpSequence = uint16(r.Intn(1 << 16))
}

func getICMPSequence() uint16 {
	icmpSequenceMutex.Lock()
	defer icmpSequenceMutex.Unlock()
	icmpSequence++
	return icmpSequence
}

func doPing(timeout time.Duration, target string) (rtt time.Duration, err error) {
	// limit icmp concurrent
	isReturnCh := false
	if ICMPConcurrentCh != nil {
		waitCtx, waitCancel := context.WithTimeout(context.Background(), timeout)
		defer waitCancel()
		select {
		case ICMPConcurrentCh <- struct{}{}:
			isReturnCh = true
		case <-waitCtx.Done():
			logger.Errorf("exceed max icmp concurrent %d", len(ICMPConcurrentCh))
			return 0, nil
		}

		defer func() {
			if isReturnCh {
				select {
				case <-ICMPConcurrentCh:
				default:
				}
			}
		}()
	}

	var (
		requestType icmp.Type
		replyType   icmp.Type
		icmpConn    *icmp.PacketConn
		rttStart    time.Time
		rb          []byte
		dstIPAddr   *net.IPAddr
		dst         net.Addr
		idUnknown   bool
		wb          []byte
	)
	{
		dstIPAddr, _, err = chooseProtocol(timeout, "ip4", true, target)
		if err != nil {
			return 0, fmt.Errorf("error resolving address: %w", err)
		}
		srcIP := net.ParseIP("::")
		privileged := true
		// Unprivileged sockets are supported on Darwin and Linux only.
		tryUnprivileged := runtime.GOOS == "darwin" || runtime.GOOS == "linux"

		if dstIPAddr.IP.To4() == nil {
			requestType = ipv6.ICMPTypeEchoRequest
			replyType = ipv6.ICMPTypeEchoReply

			if tryUnprivileged {
				// "udp" here means unprivileged -- not the protocol "udp".
				icmpConn, err = icmp.ListenPacket("udp6", srcIP.String())
				if err == nil {
					privileged = false
				}
			}

			if privileged {
				icmpConn, err = icmp.ListenPacket("ip6:ipv6-icmp", srcIP.String())
				if err != nil {
					return 0, fmt.Errorf("error listening to socket: %w", err)
				}
			}
			defer icmpConn.Close()

			if err := icmpConn.IPv6PacketConn().SetControlMessage(ipv6.FlagHopLimit, true); err != nil {
				logger.Debug("Failed to set Control Message for retrieving Hop Limit", "err", err)
			}
		} else {
			requestType = ipv4.ICMPTypeEcho
			replyType = ipv4.ICMPTypeEchoReply

			if tryUnprivileged {
				icmpConn, err = icmp.ListenPacket("udp4", srcIP.String())
				if err == nil {
					privileged = false
				}
			}

			if privileged {
				icmpConn, err = icmp.ListenPacket("ip4:icmp", srcIP.String())
				if err != nil {
					return 0, fmt.Errorf("error listening to socket: %w", err)
				}
			}
			defer icmpConn.Close()

			if err := icmpConn.IPv4PacketConn().SetControlMessage(ipv4.FlagTTL, true); err != nil {
				logger.Debug("Failed to set Control Message for retrieving TTL", "err", err)
			}
		}

		dst = dstIPAddr
		if !privileged {
			dst = &net.UDPAddr{IP: dstIPAddr.IP, Zone: dstIPAddr.Zone}
		}

		var data []byte = []byte("ICMP testing")

		body := &icmp.Echo{
			ID:   icmpID,
			Seq:  int(getICMPSequence()),
			Data: data,
		}
		wm := icmp.Message{
			Type: requestType,
			Code: 0,
			Body: body,
		}

		wb, err = wm.Marshal(nil)
		if err != nil {
			return 0, fmt.Errorf("error marshalling packet: %w", err)
		}

		rttStart = time.Now()

		_, err = icmpConn.WriteTo(wb, dst)
		if err != nil {
			return 0, fmt.Errorf("error writing to socket: %w", err)
		}

		// Reply should be the same except for the message type and ID if
		// unprivileged sockets were used and the kernel used its own.
		wm.Type = replyType
		// Unprivileged cannot set IDs on Linux.
		idUnknown = !privileged && runtime.GOOS == "linux"
		if idUnknown {
			body.ID = 0
		}
		wb, err = wm.Marshal(nil)
		if err != nil {
			return 0, fmt.Errorf("error marshalling packet: %w", err)
		}

		if idUnknown {
			// If the ID is unknown (due to unprivileged sockets) we also cannot know
			// the checksum in userspace.
			wb[2] = 0
			wb[3] = 0
		}

		rb = make([]byte, 65536)
		deadline := time.Now().Add(timeout)
		err = icmpConn.SetReadDeadline(deadline)
		if err != nil {
			return 0, fmt.Errorf("error setting socket deadline: %w", err)
		}

	}

	for {
		var n int
		var peer net.Addr
		readTime := time.Now()
		if dstIPAddr.IP.To4() == nil {
			n, _, peer, err = icmpConn.IPv6PacketConn().ReadFrom(rb)
		} else {
			n, _, peer, err = icmpConn.IPv4PacketConn().ReadFrom(rb)
		}
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				logger.Debugf("timeout reading from socket: %s", err.Error())
				return 0, nil
			}
			return 0, fmt.Errorf("error reading from socket: %w", err)
		}
		if peer.String() != dst.String() {
			//  if read time is too big, then return concurrent channel
			if time.Since(readTime) > 10*time.Second && isReturnCh {
				select {
				case <-ICMPConcurrentCh:
					isReturnCh = false
				default:
				}
			}
			continue
		}
		if idUnknown {
			// Clear the ID from the packet, as the kernel will have replaced it (and
			// kept track of our packet for us, hence clearing is safe).
			rb[4] = 0
			rb[5] = 0
		}
		if idUnknown || replyType == ipv6.ICMPTypeEchoReply {
			// Clear checksum to make comparison succeed.
			rb[2] = 0
			rb[3] = 0
		}
		if bytes.Equal(rb[:n], wb) {
			rtt = time.Since(rttStart)
			return rtt, nil
		}
	}
}

// Returns the IP for the IPProtocol and lookup time.
func chooseProtocol(timeout time.Duration, IPProtocol string, fallbackIPProtocol bool, target string) (ip *net.IPAddr, lookupTime float64, err error) {
	if IPProtocol == "ip6" || IPProtocol == "" {
		IPProtocol = "ip6"
	} else {
		IPProtocol = "ip4"
	}

	resolveStart := time.Now()

	defer func() {
		lookupTime = time.Since(resolveStart).Seconds()
	}()

	resolver := &net.Resolver{}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if !fallbackIPProtocol {
		ips, err := resolver.LookupIP(ctx, IPProtocol, target)
		if err == nil {
			for _, ip := range ips {
				return &net.IPAddr{IP: ip}, lookupTime, nil
			}
		}
		return nil, 0.0, err
	}

	ips, err := resolver.LookupIPAddr(ctx, target)
	if err != nil {
		return nil, 0.0, err
	}

	// Return the IP in the requested protocol.
	var fallback *net.IPAddr
	for _, ip := range ips {
		switch IPProtocol {
		case "ip4":
			if ip.IP.To4() != nil {
				return &ip, lookupTime, nil
			}

			// ip4 as fallback
			fallback = &ip

		case "ip6":
			if ip.IP.To4() == nil {
				return &ip, lookupTime, nil
			}

			// ip6 as fallback
			fallback = &ip
		}
	}

	// Unable to find ip and no fallback set.
	if fallback == nil || !fallbackIPProtocol {
		return nil, 0.0, fmt.Errorf("unable to find ip; no fallback")
	}

	return fallback, lookupTime, nil
}

type pingStat struct {
	PacketLoss float64
	PacketsSent,
	PacketsRecv int
	MinRtt,
	AvgRtt,
	MaxRtt,
	StdDevRtt time.Duration
}

func pingTarget(target string, count int, interval, timeout time.Duration) (stat *pingStat, err error) {
	if count <= 0 {
		count = 3
	}
	if interval < time.Second {
		interval = 1 * time.Second
	}
	if timeout < time.Second {
		timeout = 3 * time.Second
	}

	stat = &pingStat{}

	rtts := []time.Duration{}
	for i := 0; i < count; i++ {
		err = func() error {
			rtt, err := doPing(timeout, target)
			if err != nil {
				return fmt.Errorf("ping failed: %w", err)
			}
			stat.PacketsSent++
			if rtt > 0 {
				stat.PacketsRecv++
				rtts = append(rtts, rtt)
			} else {
				stat.PacketLoss++
			}

			return nil
		}()
		if err != nil {
			return nil, err
		}
		time.Sleep(interval)
	}

	pktCount := time.Duration(len(rtts))
	if pktCount > 0 {
		stat.MinRtt = rtts[0]
		stat.MaxRtt = rtts[0]
		stat.AvgRtt = rtts[0]
		var m2 float64
		for i, rtt := range rtts {
			if rtt < stat.MinRtt {
				stat.MinRtt = rtt
			}
			if rtt > stat.MaxRtt {
				stat.MaxRtt = rtt
			}

			oldAvg := stat.AvgRtt
			stat.AvgRtt += (rtt - stat.AvgRtt) / time.Duration(i+1)
			m2 += float64((rtt - oldAvg) * (rtt - stat.AvgRtt))
		}
		if pktCount > 1 {
			stat.StdDevRtt = time.Duration(math.Sqrt(m2 / float64(pktCount)))
		} else {
			stat.StdDevRtt = 0
		}
	}

	if stat.PacketsSent > 0 {
		loss := float64(stat.PacketsSent-stat.PacketsRecv) / float64(stat.PacketsSent) * 100
		stat.PacketLoss = loss
	}

	return stat, nil
}

func (t *ICMPTask) setReqError(err string) {
	t.reqError = err
}
