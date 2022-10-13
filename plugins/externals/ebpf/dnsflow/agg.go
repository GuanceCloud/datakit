//go:build (linux && amd64 && ebpf) || (linux && arm64 && ebpf)
// +build linux,amd64,ebpf linux,arm64,ebpf

package dnsflow

import (
	"math"
	"strconv"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/ebpf/k8sinfo"
	dknetflow "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/ebpf/netflow"
)

type aggKey struct {
	sAddr string
	dAddr string

	sPort uint32
	dPort uint32

	sType string
	dType string

	rcode int

	family    string
	transport string
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
	addTags map[string]string, k8sNetInfo *k8sinfo.K8sNetInfo,
) (*client.Point, error) {
	tags := map[string]string{
		"family": key.family,

		"direction": key.direction,
		"transport": key.transport,

		"src_ip": key.sAddr,
		"dst_ip": key.dAddr,

		"src_ip_type": key.sType,
		"dst_ip_type": key.dType,
	}

	if key.sPort == math.MaxUint32 {
		tags["src_port"] = "*"
	} else {
		tags["src_port"] = strconv.FormatInt(int64(key.sPort), 10)
	}

	if key.dPort == math.MaxUint32 {
		tags["dst_port"] = "*"
	} else {
		tags["dst_port"] = strconv.FormatInt(int64(key.dPort), 10)
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

	tags = dknetflow.AddK8sTags2Map(k8sNetInfo, key.sAddr, key.dAddr,
		key.sPort, key.dPort, key.transport, tags)
	return client.NewPoint(srcNameM, tags, fields, pTime)
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
		key.transport = "udp"
	} else {
		key.transport = "tcp"
	}

	// direction
	key.direction = dknetflow.DirectionOutgoing

	// port
	key.sPort = uint32(dnsKey.ClientPort)
	key.dPort = uint32(dnsKey.ServerPort)
	if key.sPort != 53 {
		key.sPort = math.MaxUint32
	}
	if key.dPort != 53 {
		key.dPort = math.MaxUint32
	}

	// ip
	key.sAddr = dknetflow.U32BEToIP(dnsKey.ClientIP, !dnsKey.IsV4).String()
	key.dAddr = dknetflow.U32BEToIP(dnsKey.ServerIP, !dnsKey.IsV4).String()

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
		key.sAddr, key.dAddr = key.dAddr, key.sAddr
		// swap port
		key.sPort, key.dPort = key.dPort, key.sPort
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

func (agg *FlowAgg) ToPoint(tags map[string]string, k8sInfo *k8sinfo.K8sNetInfo) []*point.Point {
	var result []*client.Point

	pTime := time.Now()
	for k, v := range agg.data {
		if pt, err := kv2point(&k, v, pTime, tags, k8sInfo); err != nil {
			l.Debug(err)
		} else {
			result = append(result, pt)
		}
	}

	return point.WrapPoint(result)
}

func (agg *FlowAgg) Clean() {
	agg.data = make(map[aggKey]*aggValue)
}
