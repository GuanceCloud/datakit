//go:build linux
// +build linux

package httpflow

import (
	"math"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/k8sinfo"
	dknetflow "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
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

	processName string

	pathTrunc bool
}

type aggValue struct {
	latency   []int
	count     int
	recvBytes int64
	sendBytes int64
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
) (*point.Point, error) {
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

	tags["process_name"] = key.processName

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

		"bytes_read":    value.recvBytes,
		"bytes_written": value.sendBytes,

		"truncated": key.pathTrunc,

		"count": value.count,
	}

	tags = dknetflow.AddK8sTags2Map(k8sNetInfo, &key.BaseKey, tags)

	kvs := point.NewTags(tags)
	kvs = append(kvs, point.NewKVs(fields)...)
	pt := point.NewPointV2(srcNameM, kvs, append(
		point.CommonLoggingOptions(), point.WithTime(pTime))...)
	return pt, nil
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

	key.processName = httpFinReq.ConnInfo.ProcessName
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
		v.recvBytes += int64(httpFinReq.HTTPStats.Recv)
		v.sendBytes += int64(httpFinReq.HTTPStats.Send)
	} else {
		agg.data[key] = &aggValue{
			count: 1,
			latency: []int{
				int(httpFinReq.HTTPStats.RespTS - httpFinReq.HTTPStats.ReqTS),
			},
			recvBytes: int64(httpFinReq.HTTPStats.Recv),
			sendBytes: int64(httpFinReq.HTTPStats.Send),
		}
	}

	return nil
}

func (agg *FlowAgg) ToPoint(tags map[string]string, k8sInfo *k8sinfo.K8sNetInfo) []*point.Point {
	var result []*point.Point

	pTime := time.Now()
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

func getBaseKey(httpFinReq *HTTPReqFinishedInfo) dknetflow.BaseKey {
	bk := dknetflow.BaseKey{}

	info := httpFinReq.ConnInfo
	isV6 := !dknetflow.ConnAddrIsIPv4(info.Meta)

	if info.Saddr[0] == 0 && info.Saddr[1] == 0 &&
		info.Daddr[0] == 0 && info.Daddr[1] == 0 {
		if info.Saddr[2] == 0xffff0000 && info.Daddr[2] == 0xffff0000 {
			isV6 = false
		} else if info.Saddr[2] == 0 && info.Daddr[2] == 0 &&
			info.Saddr[3] > 1 && info.Daddr[3] > 1 {
			isV6 = false
		}
	}

	// fields["src_ip"]
	bk.SAddr = dknetflow.U32BEToIP(info.Saddr, isV6).String()
	// fields["dst_ip"]
	bk.DAddr = dknetflow.U32BEToIP(info.Daddr, isV6).String()
	if info.NATDport != 0 && (info.NATDaddr[0]|
		info.NATDaddr[1]|info.NATDaddr[2]|info.NATDaddr[3]) != 0 {
		// fields["dst_nat_port"]
		bk.DNATPort = info.NATDport
		// fields["dst_nat_ip"]
		bk.DNATAddr = dknetflow.U32BEToIP(info.NATDaddr, isV6).String()
	}

	// fields["src_port"] =
	bk.SPort = info.Sport
	bk.DPort = info.Dport
	return bk
}
