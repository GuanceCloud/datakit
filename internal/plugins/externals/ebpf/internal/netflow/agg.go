//go:build linux
// +build linux

package netflow

import (
	"math"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/cli"
)

type BaseKey struct {
	SAddr string
	DAddr string

	SPort uint32
	DPort uint32

	Transport string

	DNATAddr string
	DNATPort uint32

	PID   int
	NetNS string
}

type aggKey struct {
	sAddr [4]uint32
	dAddr [4]uint32

	sPort uint32
	dPort uint32

	transport string

	dnatAddr [4]uint32
	dnatPort uint32

	pid   int
	netns uint32

	isIPv6 bool

	sType string
	dType string

	family      string
	direction   string
	processName string
}

type aggValue struct {
	bytesRead    int64
	bytesWritten int64
	packetsRead  int64
	packetsWrite int64

	retransmits    int64
	rtt            int64
	rttVar         int64
	tcpClosed      int64
	tcpEstablished int64
	tcpConnects    int64
	tcpFailures    int64
	tcpCloseWait   int64
	tcpLastAck     int64
	tcpTimeWait    int64

	count int
}

func avgLatency(total int64, count int) int64 {
	if count == 0 {
		return 0
	}
	return total / int64(count)
}

func kv2point(key *aggKey, value *aggValue, pTime time.Time,
	addTags map[string]string, k8sNetInfo *cli.K8sInfo,
) (*point.Point, error) {
	baseKey := BaseKey{
		SAddr:     U32BEToIP(key.sAddr, key.isIPv6).String(),
		DAddr:     U32BEToIP(key.dAddr, key.isIPv6).String(),
		SPort:     key.sPort,
		DPort:     key.dPort,
		Transport: key.transport,
		DNATPort:  key.dnatPort,
		PID:       key.pid,
		NetNS:     strconv.FormatUint(uint64(key.netns), 10),
	}
	if key.dnatPort != 0 && (key.dnatAddr[0]|key.dnatAddr[1]|key.dnatAddr[2]|key.dnatAddr[3]) != 0 {
		baseKey.DNATAddr = U32BEToIP(key.dnatAddr, key.isIPv6).String()
	}

	tags := map[string]string{
		"family": key.family,

		"direction": key.direction,
		"transport": baseKey.Transport,

		"src_ip": baseKey.SAddr,
		"dst_ip": baseKey.DAddr,

		"src_ip_type": key.sType,
		"dst_ip_type": key.dType,

		"netns": baseKey.NetNS,
	}

	if baseKey.DNATAddr != "" && baseKey.DNATPort != 0 {
		tags["dst_nat_ip"] = baseKey.DNATAddr
		tags["dst_nat_port"] = strconv.FormatInt(int64(baseKey.DNATPort), 10)
		l.Debugf("netflow NAT point: dst=%s:%d nat=%s:%d transport=%s pid=%d netns=%s",
			baseKey.DAddr, baseKey.DPort, baseKey.DNATAddr, baseKey.DNATPort,
			baseKey.Transport, baseKey.PID, baseKey.NetNS)
	} else {
		tags["dst_nat_ip"] = NoValue
		tags["dst_nat_port"] = NoValue
	}

	tags["process_name"] = key.processName

	if baseKey.SPort == math.MaxUint32 {
		tags["src_port"] = "*"
	} else {
		tags["src_port"] = strconv.Itoa(int(baseKey.SPort))
	}

	if baseKey.DPort == math.MaxUint32 {
		tags["dst_port"] = "*"
	} else {
		tags["dst_port"] = strconv.Itoa(int(baseKey.DPort))
	}

	if domain := LookupPeerDomain(baseKey.DAddr, baseKey.DPort, baseKey.Transport, baseKey.NetNS); domain != "" {
		tags["dst_domain"] = domain
	} else if dnsRecord != nil {
		tags["dst_domain"] = dnsRecord.LookupAddr(baseKey.DAddr)
	}

	for k, v := range addTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	var fields map[string]any

	if baseKey.Transport == transportTCP {
		fields = map[string]any{
			"bytes_read":           value.bytesRead,
			"bytes_written":        value.bytesWritten,
			"packets_read":         value.packetsRead,
			"packets_written":      value.packetsWrite,
			"retransmits":          value.retransmits,
			"rtt":                  avgLatency(value.rtt, value.count),
			"rtt_var":              avgLatency(value.rttVar, value.count),
			"tcp_closed":           value.tcpClosed,
			"tcp_established":      value.tcpEstablished,
			"tcp_connect_attempts": value.tcpConnects,
			"tcp_connect_failures": value.tcpFailures,
			"tcp_close_wait":       value.tcpCloseWait,
			"tcp_last_ack":         value.tcpLastAck,
			"tcp_time_wait":        value.tcpTimeWait,
		}
	} else {
		fields = map[string]any{
			"bytes_read":      value.bytesRead,
			"bytes_written":   value.bytesWritten,
			"packets_read":    value.packetsRead,
			"packets_written": value.packetsWrite,
		}
	}

	tags = AddK8sTags2Map(k8sNetInfo, &baseKey, tags)

	tags, fields = AddClientServerInf(tags, fields)

	kvs := point.NewTags(tags)
	kvs = append(kvs, point.NewKVs(fields)...)
	pt := point.NewPoint(srcNameM, kvs, append(
		point.CommonLoggingOptions(), point.WithTime(pTime))...)

	return pt, nil
}

type FlowAgg struct {
	data map[aggKey]*aggValue
}

func (agg *FlowAgg) Len() int {
	return len(agg.data)
}

func (agg *FlowAgg) Append(info ConnectionInfo, stats ConnFullStats) error {
	if !ConnNotNeedToFilter(&info, &stats) {
		return nil
	}

	if agg.data == nil {
		agg.data = map[aggKey]*aggValue{}
	}

	var key aggKey

	// family
	isV6 := !ConnAddrIsIPv4(info.Meta)

	if info.Saddr[0] == 0 && info.Saddr[1] == 0 &&
		info.Daddr[0] == 0 && info.Daddr[1] == 0 {
		if info.Saddr[2] == 0xffff0000 && info.Daddr[2] == 0xffff0000 {
			isV6 = false
		} else if info.Saddr[2] == 0 && info.Daddr[2] == 0 &&
			info.Saddr[3] > 1 && info.Daddr[3] > 1 {
			isV6 = false
		}
	}

	// ip type
	if isV6 {
		key.sType = ConnIPv6Type(info.Saddr)
		key.dType = ConnIPv6Type(info.Daddr)
		key.family = "IPv6"
	} else {
		key.sType = ConnIPv4Type(info.Saddr[3])
		key.dType = ConnIPv4Type(info.Daddr[3])
		key.family = "IPv4"
	}

	// saddr, daddr
	key.isIPv6 = isV6
	key.sAddr = info.Saddr
	key.dAddr = info.Daddr
	if info.NATDport != 0 && (info.NATDaddr[0]|
		info.NATDaddr[1]|info.NATDaddr[2]|info.NATDaddr[3]) != 0 {
		key.dnatPort = info.NATDport
		key.dnatAddr = info.NATDaddr
	}

	key.netns = info.Netns
	// sport, dport
	key.sPort = info.Sport
	key.dPort = info.Dport

	// transport
	if ConnProtocolIsTCP(info.Meta) {
		key.transport = transportTCP
	} else {
		key.transport = transportUDP
	}

	// direction
	key.direction = ConnDirection2Str(stats.Stats.Direction)

	key.processName = info.ProcessName
	key.pid = int(info.Pid)

	if k8sNetInfo != nil && IsIncomingFromK8s(k8sNetInfo, key.pid, U32BEToIP(key.sAddr, key.isIPv6).String(),
		key.sPort, key.transport) {
		key.direction = DirectionIncoming
	}

	key.direction, key.sPort, key.dPort = NormalizeDirectionAndPorts(key.direction, key.sPort, key.dPort)

	var value *aggValue
	// agg latency and count ++
	if v, ok := agg.data[key]; ok {
		v.count++
		value = v
	} else {
		value = &aggValue{
			count: 1,
		}
		agg.data[key] = value
	}

	value.bytesRead += int64(stats.Stats.RecvBytes)
	value.bytesWritten += int64(stats.Stats.SentBytes)
	value.packetsRead += int64(stats.Stats.RecvPackets)
	value.packetsWrite += int64(stats.Stats.SentPackets)

	if key.transport == transportTCP {
		value.rtt += int64(stats.TCPStats.Rtt)
		value.rttVar += int64(stats.TCPStats.RttVar)
		value.retransmits += int64(stats.TCPStats.Retransmits)
		value.tcpClosed += stats.TotalClosed
		value.tcpEstablished += stats.TotalEstablished
		value.tcpConnects += int64(stats.TCPStats.ConnectAttempts)
		value.tcpFailures += int64(stats.TCPStats.ConnectFailures)
		value.tcpCloseWait += int64(stats.TCPStats.CloseWait)
		value.tcpLastAck += int64(stats.TCPStats.LastAck)
		value.tcpTimeWait += int64(stats.TCPStats.TimeWait)
	}

	return nil
}

func (agg *FlowAgg) ToPoint(tags map[string]string,
	k8sInfo *cli.K8sInfo,
) []*point.Point {
	result := make([]*point.Point, 0, len(agg.data))

	pTime := ntp.Now()
	for k, v := range agg.data {
		if pt, err := kv2point(&k, v, pTime, tags, k8sInfo); err != nil {
			l.Debug(err)
		} else {
			result = append(result, pt)
		}
	}

	return result
}

func (agg *FlowAgg) Clean() {
	agg.data = make(map[aggKey]*aggValue)
}
