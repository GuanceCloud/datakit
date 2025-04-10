//go:build linux
// +build linux

// Package protodec implements the protocol decoder
package protodec

import (
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/cli"
)

type HTTPAggP struct {
	data map[aggKey]*aggValue

	sync.RWMutex
}

type aggKey struct {
	netflow.BaseKey

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
	latency   int64
	count     int
	recvBytes int64
	sendBytes int64
}

func ConnNotNeedToFilter(conn *comm.ConnectionInfo) bool {
	if (conn.Saddr[0]|conn.Saddr[1]|conn.Saddr[2]|conn.Saddr[3]) == 0 ||
		(conn.Daddr[0]|conn.Daddr[1]|conn.Daddr[2]|conn.Daddr[3]) == 0 ||
		conn.Sport == 0 || conn.Dport == 0 {
		return false
	}
	if netflow.ConnAddrIsIPv4(conn.Meta) { // IPv4
		if (conn.Saddr[3]&0xff) == 127 && (conn.Daddr[3]&0xff) == 127 {
			return false
		}
	} else { // IPv6
		if conn.Saddr[2] == 0xffff0000 && conn.Daddr[2] == 0xffff0000 {
			if (conn.Saddr[3]&0xff) == 127 && (conn.Daddr[3]&0xff) == 127 {
				return false
			}
		} else if (conn.Saddr[0]|conn.Saddr[1]|conn.Saddr[2]) == 0 && conn.Saddr[3] == 1 &&
			(conn.Daddr[0]|conn.Daddr[1]|conn.Daddr[2]) == 0 && conn.Daddr[3] == 1 {
			return false
		}
	}
	return true
}

func (agg *HTTPAggP) Obs(conn *comm.ConnectionInfo, data *ProtoData) {
	agg.Lock()
	defer agg.Unlock()
	if agg.data == nil {
		agg.data = map[aggKey]*aggValue{}
	}

	if !ConnNotNeedToFilter(conn) {
		return
	}

	var key aggKey
	key.NetNS = strconv.FormatUint(uint64(conn.Netns), 10)
	key.processName = conn.ProcessName
	key.pid = int64(conn.Pid)

	// direction
	key.direction = data.Direction.String()

	// family
	isV6 := !netflow.ConnAddrIsIPv4(conn.Meta)
	if conn.Saddr[0] == 0 && conn.Saddr[1] == 0 &&
		conn.Daddr[0] == 0 && conn.Daddr[1] == 0 {
		if conn.Saddr[2] == 0xffff0000 && conn.Daddr[2] == 0xffff0000 {
			isV6 = false
		} else if conn.Saddr[2] == 0 && conn.Daddr[2] == 0 &&
			conn.Saddr[3] > 1 && conn.Daddr[3] > 1 {
			isV6 = false
		}
	}

	// ip type
	if isV6 {
		key.sType = netflow.ConnIPv6Type(conn.Saddr)
		key.dType = netflow.ConnIPv6Type(conn.Daddr)
		key.family = "IPv6"
	} else {
		key.sType = netflow.ConnIPv4Type(conn.Saddr[3])
		key.dType = netflow.ConnIPv4Type(conn.Daddr[3])
		key.family = "IPv4"
	}

	// saddr, daddr, sport, dport, transport
	key.SAddr = netflow.U32BEToIP(conn.Saddr, isV6).String()
	key.DAddr = netflow.U32BEToIP(conn.Daddr, isV6).String()

	if conn.NATDport != 0 && (conn.NATDaddr[0]|
		conn.NATDaddr[1]|conn.NATDaddr[2]|conn.NATDaddr[3]) != 0 {
		key.DNATPort = conn.NATDport
		key.DNATAddr = netflow.U32BEToIP(conn.NATDaddr, isV6).String()
	}

	key.SPort = conn.Sport
	key.DPort = conn.Dport

	switch key.direction {
	case DirectionOutgoing:
		if netflow.IsEphemeralPort(key.SPort) {
			key.SPort = math.MaxUint32
		}
	case DirectionIncoming:
		if netflow.IsEphemeralPort(key.DPort) {
			key.DPort = math.MaxUint32
		}
	}

	// transport
	if netflow.ConnProtocolIsTCP(conn.Meta) {
		key.Transport = "tcp"
	} else {
		key.Transport = "udp"
	}
	// path, method, status_code, http_version
	key.path = data.KVs.Get(comm.FieldHTTPRoute).GetS()
	key.method = data.KVs.Get(comm.FieldHTTPMethod).GetS()
	key.statusCode, _ = strconv.Atoi(data.KVs.Get(comm.FieldHTTPStatusCode).GetS())
	key.httpVersion = data.KVs.Get(comm.FieldHTTPVersion).GetS()
	rcv := data.KVs.Get(comm.FieldBytesRead).GetI()
	snd := data.KVs.Get(comm.FieldBytesWritten).GetI()

	// agg latency and count ++
	if v, ok := agg.data[key]; ok {
		v.count++
		v.latency += data.Cost
		v.recvBytes += rcv
		v.sendBytes += snd
	} else {
		agg.data[key] = &aggValue{
			count:     1,
			latency:   data.Cost,
			recvBytes: snd,
			sendBytes: rcv,
		}
	}
}

func (agg *HTTPAggP) Proto() L7Protocol {
	return ProtoHTTP
}

func (agg *HTTPAggP) Export(tags map[string]string, k8sInfo *cli.K8sInfo) []*point.Point {
	agg.RLock()
	defer agg.RUnlock()
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

func (agg *HTTPAggP) Cleanup() {
	agg.Lock()
	defer agg.Unlock()

	agg.data = map[aggKey]*aggValue{}
}

func newHTTPAggP(p L7Protocol) AggPool {
	return &HTTPAggP{}
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

		"pid": strconv.FormatInt(key.pid, 10),

		"src_ip_type": key.sType,
		"dst_ip_type": key.dType,

		"netns": key.NetNS,
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
		"latency":     value.latency / int64(value.count),

		"bytes_read":    value.recvBytes,
		"bytes_written": value.sendBytes,

		"truncated": key.pathTrunc,

		"count": value.count,
	}

	tags = netflow.AddK8sTags2Map(k8sNetInfo, &key.BaseKey, tags)

	kvs := point.NewTags(tags)
	kvs = append(kvs, point.NewKVs(fields)...)
	pt := point.NewPointV2("httpflow", kvs, append(
		point.CommonLoggingOptions(), point.WithTime(pTime))...)
	return pt, nil
}
