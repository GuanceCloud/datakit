//go:build linux
// +build linux

package l4log

import (
	"math"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/k8sinfo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
)

type aggKey struct {
	netflow.BaseKey
	pid int

	family      string
	direction   string
	processName string
}

type aggValue struct {
	bytesRead    int64
	bytesWritten int64

	retransmits    int64
	rtt            int64
	rttVar         int64
	tcpClosed      int64
	tcpEstablished int64

	count int
}

func kv2point(key *aggKey, value *aggValue, pTime time.Time,
	addTags map[string]string, k8sNetInfo *k8sinfo.K8sNetInfo,
) (*point.Point, error) {
	tags := map[string]string{
		"family": key.family,

		"direction": key.direction,
		"transport": key.Transport,

		"src_ip": key.SAddr,
		"dst_ip": key.DAddr,

		"pid": strconv.FormatInt(int64(key.pid), 10),

		"netns": key.NetNS,
	}

	tags["process_name"] = key.processName

	if key.SPort == math.MaxUint32 {
		tags["src_port"] = "*"
	} else {
		tags["src_port"] = strconv.FormatInt(int64(key.SPort), 10)
	}

	if key.DPort == math.MaxUint32 {
		tags["dst_port"] = "*"
	} else {
		tags["dst_port"] = strconv.FormatInt(int64(key.DPort), 10)
	}

	for k, v := range addTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	fields := map[string]any{
		"bytes_read":      value.bytesRead,
		"bytes_written":   value.bytesWritten,
		"retransmits":     value.retransmits,
		"tcp_closed":      value.tcpClosed,
		"tcp_established": value.tcpEstablished,
	}

	if value.count != 0 {
		fields["rtt"] = value.rtt / int64(value.count)
		// fields["rtt_var"] = value.rttVar / int64(value.count)
	}

	tags = netflow.AddK8sTags2Map(k8sNetInfo, &key.BaseKey, tags)

	kvs := point.NewTags(tags)
	kvs = append(kvs, point.NewKVs(fields)...)
	pt := point.NewPointV2("netflow", kvs, append(
		point.CommonLoggingOptions(), point.WithTime(pTime))...)

	return pt, nil
}

type FlowAggTCP struct {
	data map[aggKey]*aggValue
}

func (agg *FlowAggTCP) Len() int {
	return len(agg.data)
}

func (agg *FlowAggTCP) Append(info *PMeta, stats *TCPMetrics, netns string,
	dir conndirection, v6, macEQ bool, nicIPList []string,
) {
	if info == nil || !macEQ {
		return
	}

	var keep bool
	for _, v := range nicIPList {
		if info.SrcIP == v {
			keep = true
			break
		}
	}
	if !keep {
		return
	}

	if stats.BytesRead == 0 && stats.BytesWritten == 0 &&
		!stats.recEstab && !stats.recClose[1] {
		return
	}

	if agg.data == nil {
		agg.data = map[aggKey]*aggValue{}
	}

	var fm string
	if v6 {
		fm = "IPv6"
	} else {
		fm = "IPv4"
	}
	key := aggKey{
		BaseKey: netflow.BaseKey{
			SAddr:     info.SrcIP,
			DAddr:     info.DstIP,
			SPort:     uint32(info.SrcPort),
			DPort:     uint32(info.DstPort),
			Transport: "tcp",
			NetNS:     netns,
		},
		family: fm,
	}

	switch dir { //nolint:exhaustive
	case directionIncoming:
		key.direction = netflow.DirectionIncoming
		if netflow.IsEphemeralPort(key.DPort) {
			key.DPort = math.MaxUint32
		}

	default:
		key.direction = "outgoing"
		if netflow.IsEphemeralPort(key.SPort) {
			key.SPort = math.MaxUint32
		}
	}

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

	if stats != nil {
		value.bytesRead += int64(stats.BytesRead)
		value.bytesWritten += int64(stats.BytesWritten)

		value.rtt += stats.RTT
		value.rttVar += stats.RTTVar
		value.retransmits += int64(stats.Retransmits)

		if stats.recEstab {
			value.tcpEstablished++
			stats.recEstab = false
		}

		if stats.recClose[1] {
			value.tcpClosed++
			stats.recClose[1] = false
		}

		// cleanup stats, for next duration
		stats.BytesRead = 0
		stats.BytesWritten = 0
		stats.Retransmits = 0
	}
}

func (agg *FlowAggTCP) ToPoint(tags map[string]string,
	k8sInfo *k8sinfo.K8sNetInfo,
) []*point.Point {
	var result []*point.Point

	pTime := time.Now()
	for k, v := range agg.data {
		if pt, err := kv2point(&k, v, pTime, tags, k8sInfo); err != nil {
			log.Debug(err)
		} else {
			result = append(result, pt)
		}
	}

	return result
}

func (agg *FlowAggTCP) Clean() {
	agg.data = make(map[aggKey]*aggValue)
}
