//go:build linux
// +build linux

package dnsflow

import (
	"math"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	dknetflow "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/cli"
)

type aggKey struct {
	dknetflow.BaseKey

	sType string
	dType string

	rcode int

	family    string
	direction string
}

type aggValue struct {
	latencyMax int
	latency    []int
	count      int
}

func calLatency(l []int) int {
	if len(l) == 0 {
		return 0
	} else {
		t := 0
		for _, v := range l {
			t += v
		}
		return t / len(l)
	}
}

func kv2point(key *aggKey, value *aggValue, pTime time.Time,
	addTags map[string]string, k8sNetInfo *cli.K8sInfo,
) (*point.Point, error) {
	tags := map[string]string{
		"family": key.family,

		"direction": key.direction,
		"transport": key.Transport,

		"src_ip": key.SAddr,
		"dst_ip": key.DAddr,

		"src_ip_type": key.sType,
		"dst_ip_type": key.dType,
	}
	sPort := strconv.FormatInt(int64(key.SPort), 10)
	dPort := strconv.FormatInt(int64(key.DPort), 10)

	if key.SPort == math.MaxUint32 {
		sPort = "*"
		tags["src_port"] = "*"
	} else {
		tags["src_port"] = sPort
	}

	if key.DPort == math.MaxUint32 {
		dPort = "*"
		tags["dst_port"] = "*"
	} else {
		tags["dst_port"] = dPort
	}

	switch key.direction {
	case dknetflow.DirectionIncoming:
		tags["client_port"] = dPort
		tags["server_port"] = sPort
		tags["client_ip"] = key.DAddr
		tags["server_ip"] = key.SAddr
		tags["conn_side"] = "server"
	case dknetflow.DirectionOutgoing:
		tags["client_port"] = sPort
		tags["server_port"] = dPort
		tags["client_ip"] = key.SAddr
		tags["server_ip"] = key.DAddr
		tags["conn_side"] = "client"
	}

	for k, v := range addTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	fields := map[string]any{
		"rcode":       key.rcode,
		"latency":     calLatency(value.latency),
		"latency_max": value.latencyMax,
		"count":       value.count,
	}

	tags = dknetflow.AddK8sTags2Map(k8sNetInfo, &key.BaseKey, tags)

	kvs := point.NewTags(tags)
	kvs = append(kvs, point.NewKVs(fields)...)

	pt := point.NewPointV2(srcNameM, kvs, append(point.CommonLoggingOptions(), point.WithTime(pTime))...)
	return pt, nil
}

type FlowAgg struct {
	data map[aggKey]*aggValue
}

func (agg *FlowAgg) Len() int {
	return len(agg.data)
}

func (agg *FlowAgg) Append(dnsKey DNSQAKey, stats DNSStats) error {
	if agg.data == nil {
		agg.data = map[aggKey]*aggValue{}
	}

	var key aggKey

	key.rcode = stats.RCODE
	// transport
	if dnsKey.IsUDP {
		key.Transport = "udp"
	} else {
		key.Transport = "tcp"
	}

	// direction
	key.direction = dknetflow.DirectionOutgoing

	// port
	key.SPort = uint32(dnsKey.ClientPort)
	key.DPort = uint32(dnsKey.ServerPort)
	if key.SPort != 53 {
		key.SPort = math.MaxUint32
	}
	if key.DPort != 53 {
		key.DPort = math.MaxUint32
	}

	// ip
	key.SAddr = dknetflow.U32BEToIP(dnsKey.ClientIP, !dnsKey.IsV4).String()
	key.DAddr = dknetflow.U32BEToIP(dnsKey.ServerIP, !dnsKey.IsV4).String()

	// ip type
	if dnsKey.IsV4 {
		key.family = "IPv4"
		key.sType = dknetflow.ConnIPv4Type(dnsKey.ClientIP[3])
		key.dType = dknetflow.ConnIPv4Type(dnsKey.ServerIP[3])
	} else {
		key.family = "IPv6"
		key.sType = dknetflow.ConnIPv6Type(dnsKey.ClientIP)
		key.dType = dknetflow.ConnIPv6Type(dnsKey.ServerIP)
	}

	if key.sType == "loopback" || key.dType == "loopback" {
		return nil
	}

	_, err := dknetflow.SrcIPPortRecorder.Query(dnsKey.ServerIP)
	if err == nil {
		// swap ip type
		key.sType, key.dType = key.dType, key.sType
		// swap ip addr
		key.SAddr, key.DAddr = key.DAddr, key.SAddr
		// swap port
		key.SPort, key.DPort = key.DPort, key.SPort
	}

	// agg latency and count ++
	if v, ok := agg.data[key]; ok {
		v.count++
		latency := int(stats.RespTime.Nanoseconds())
		v.latency = append(v.latency, latency)
		if latency > v.latencyMax {
			v.latencyMax = latency
		}
	} else {
		agg.data[key] = &aggValue{
			count: 1,
			latency: []int{
				int(stats.RespTime.Nanoseconds()),
			},
		}
	}

	return nil
}

func (agg *FlowAgg) ToPoint(tags map[string]string, k8sInfo *cli.K8sInfo) []*point.Point {
	var result []*point.Point

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
