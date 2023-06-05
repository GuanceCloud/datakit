//go:build (linux && amd64 && ebpf) || (linux && arm64 && ebpf)
// +build linux,amd64,ebpf linux,arm64,ebpf

package httpflow

import (
	"math"
	"strconv"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/k8sinfo"
	dknetflow "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/netflow"
)

type aggKey struct {
	dknetflow.BaseKey

	sType string
	dType string

	httpVersion string
	method      string
	path        string
	statusCode  int

	family    string
	direction string

	pid int64

	pathTrunc bool
}

type aggValue struct {
	latency []int
	count   int
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
	pidMap map[int][2]string,
) (*client.Point, error) {
	tags := map[string]string{
		"family": key.family,

		"direction": key.direction,
		"transport": key.Transport,

		"src_ip": key.SAddr,
		"dst_ip": key.DAddr,

		"pid": strconv.FormatInt(key.pid, 10),

		"src_ip_type": key.sType,
		"dst_ip_type": key.dType,
	}

	if key.DNATAddr != "" && key.DNATPort != 0 {
		tags["dst_nat_ip"] = key.DNATAddr
		tags["dst_nat_port"] = strconv.FormatInt(int64(key.DNATPort), 10)
	} else {
		tags["dst_nat_ip"] = NoValue
		tags["dst_nat_port"] = NoValue
	}

	if procName, ok := pidMap[int(key.pid)]; ok {
		tags["process_name"] = procName[0]
	} else {
		tags["process_name"] = NoValue
	}

	if key.SPort == math.MaxUint32 {
		tags["src_port"] = "*"
	} else {
		tags["src_port"] = cast.ToString(key.SPort)
	}

	if key.DPort == math.MaxUint32 {
		tags["dst_port"] = "*"
	} else {
		tags["dst_port"] = cast.ToString(key.DPort)
	}

	for k, v := range addTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	fields := map[string]any{
		"method":       key.method,
		"http_version": key.httpVersion,
		"path":         key.path,

		"status_code": key.statusCode,
		"latency":     calLatency(value.latency),

		"truncated": key.pathTrunc,

		"count": value.count,
	}

	tags = dknetflow.AddK8sTags2Map(k8sNetInfo, &key.BaseKey, tags)
	return client.NewPoint(srcNameM, tags, fields, pTime)
}

type FlowAgg struct {
	data map[aggKey]*aggValue
}

func (agg *FlowAgg) Len() int {
	return len(agg.data)
}

func (agg *FlowAgg) Append(httpFinReq *HTTPReqFinishedInfo) error {
	if !ConnNotNeedToFilter(httpFinReq.ConnInfo) {
		return nil
	}

	// cliPid := httpFinReq.HTTPStats.CSPid >> 32
	// svcPid := httpFinReq.HTTPStats.CSPid & 0xFFFF_FFFF

	if agg.data == nil {
		agg.data = map[aggKey]*aggValue{}
	}

	var key aggKey

	key.pid = int64(httpFinReq.ConnInfo.Pid)

	// direction
	key.direction = httpFinReq.HTTPStats.Direction

	// url
	key.path = httpFinReq.HTTPStats.Path
	key.pathTrunc = false

	// http version, method, status_code
	key.httpVersion = ParseHTTPVersion(httpFinReq.HTTPStats.HTTPVersion)
	key.method = HTTPMethodInt(int(httpFinReq.HTTPStats.ReqMethod))
	key.statusCode = int(httpFinReq.HTTPStats.RespCode)

	// family
	isV6 := !dknetflow.ConnAddrIsIPv4(httpFinReq.ConnInfo.Meta)

	if httpFinReq.ConnInfo.Saddr[0] == 0 && httpFinReq.ConnInfo.Saddr[1] == 0 &&
		httpFinReq.ConnInfo.Daddr[0] == 0 && httpFinReq.ConnInfo.Daddr[1] == 0 {
		if httpFinReq.ConnInfo.Saddr[2] == 0xffff0000 && httpFinReq.ConnInfo.Daddr[2] == 0xffff0000 {
			isV6 = false
		} else if httpFinReq.ConnInfo.Saddr[2] == 0 && httpFinReq.ConnInfo.Daddr[2] == 0 &&
			httpFinReq.ConnInfo.Saddr[3] > 1 && httpFinReq.ConnInfo.Daddr[3] > 1 {
			isV6 = false
		}
	}

	// ip type
	if isV6 {
		key.sType = dknetflow.ConnIPv6Type(httpFinReq.ConnInfo.Saddr)
		key.dType = dknetflow.ConnIPv6Type(httpFinReq.ConnInfo.Daddr)
		key.family = "IPv6"
	} else {
		key.sType = dknetflow.ConnIPv4Type(httpFinReq.ConnInfo.Saddr[3])
		key.dType = dknetflow.ConnIPv4Type(httpFinReq.ConnInfo.Daddr[3])
		key.family = "IPv4"
	}

	// saddr, daddr, sport, dport, transport
	key.SAddr = dknetflow.U32BEToIP(httpFinReq.ConnInfo.Saddr, isV6).String()
	key.DAddr = dknetflow.U32BEToIP(httpFinReq.ConnInfo.Daddr, isV6).String()

	if httpFinReq.ConnInfo.NATDport != 0 && (httpFinReq.ConnInfo.NATDaddr[0]|
		httpFinReq.ConnInfo.NATDaddr[1]|httpFinReq.ConnInfo.NATDaddr[2]|httpFinReq.ConnInfo.NATDaddr[3]) != 0 {
		key.DNATPort = httpFinReq.ConnInfo.NATDport
		key.DNATAddr = dknetflow.U32BEToIP(httpFinReq.ConnInfo.NATDaddr, isV6).String()
	}

	key.SPort = httpFinReq.ConnInfo.Sport
	key.DPort = httpFinReq.ConnInfo.Dport

	switch key.direction {
	case DirectionOutgoing:
		if dknetflow.IsEphemeralPort(key.SPort) {
			key.SPort = math.MaxUint32
		}
	case DirectionIncoming:
		if dknetflow.IsEphemeralPort(key.DPort) {
			key.DPort = math.MaxUint32
		}
	}

	// transport
	if dknetflow.ConnProtocolIsTCP(httpFinReq.ConnInfo.Meta) {
		key.Transport = "tcp"
	} else {
		key.Transport = "udp"
	}

	// agg latency and count ++
	if v, ok := agg.data[key]; ok {
		v.count++
		v.latency = append(v.latency,
			int(httpFinReq.HTTPStats.RespTS-httpFinReq.HTTPStats.ReqTS))
	} else {
		agg.data[key] = &aggValue{
			count: 1,
			latency: []int{
				int(httpFinReq.HTTPStats.RespTS - httpFinReq.HTTPStats.ReqTS),
			},
		}
	}

	return nil
}

func (agg *FlowAgg) ToPoint(tags map[string]string, k8sInfo *k8sinfo.K8sNetInfo,
	pidMap map[int][2]string,
) []*client.Point {
	var result []*client.Point

	pTime := time.Now()
	for k, v := range agg.data {
		if pt, err := kv2point(&k, v, pTime, tags, k8sInfo, pidMap); err != nil {
			l.Debug(err)
		} else {
			l.Debug(pt)
			result = append(result, pt)
		}
	}

	return result
}

func (agg *FlowAgg) Clean() {
	agg.data = make(map[aggKey]*aggValue)
}
