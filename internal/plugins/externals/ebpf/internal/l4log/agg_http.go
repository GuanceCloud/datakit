//go:build linux
// +build linux

package l4log

import (
	"math"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/k8sinfo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
)

type aggHTTPKey struct {
	netflow.BaseKey

	method string
	path   string

	statusCode int

	family    string
	direction string
}

type aggHTTPValue struct {
	latency   int64
	count     int
	recvBytes int64
	sendBytes int64
}

type FlowAggHTTP struct {
	data map[aggHTTPKey]*aggHTTPValue
}

func (agg *FlowAggHTTP) Len() int {
	return len(agg.data)
}

func (agg *FlowAggHTTP) Append(info *PMeta, stats *HTTPLogElem, netns string,
	v6, macEQ bool, nicIPList []string, waitTS int64,
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

	var fm string
	if v6 {
		fm = "IPv6"
	} else {
		fm = "IPv4"
	}

	key := aggHTTPKey{
		BaseKey: netflow.BaseKey{
			SAddr:     info.SrcIP,
			DAddr:     info.DstIP,
			SPort:     uint32(info.SrcPort),
			DPort:     uint32(info.DstPort),
			Transport: "tcp",
			NetNS:     netns,
		},
		method:     stats.Method,
		path:       stats.Path,
		statusCode: stats.StatusCode,
		family:     fm,
		direction:  stats.Direction,
	}

	switch key.direction { //nolint:exhaustive
	case DIncoming:
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

	if agg.data == nil {
		agg.data = map[aggHTTPKey]*aggHTTPValue{}
	}

	var value *aggHTTPValue
	if v, ok := agg.data[key]; ok {
		value = v
	} else {
		agg.data[key] = &aggHTTPValue{}
		value = agg.data[key]
	}

	value.latency += waitTS
	value.count++

	value.recvBytes += stats.rxBytes
	value.sendBytes += stats.txBytes
}

func (agg *FlowAggHTTP) ToPoint(tags map[string]string,
	k8sInfo *k8sinfo.K8sNetInfo,
) []*point.Point {
	var result []*point.Point

	pTime := time.Now()
	for k, v := range agg.data {
		if pt, err := kv2pointHTTP(&k, v, pTime, tags, k8sInfo); err != nil {
			log.Debug(err)
		} else {
			result = append(result, pt)
		}
	}

	return result
}

func (agg *FlowAggHTTP) Clean() {
	agg.data = make(map[aggHTTPKey]*aggHTTPValue)
}

func kv2pointHTTP(key *aggHTTPKey, value *aggHTTPValue, pTime time.Time,
	addTags map[string]string, k8sNetInfo *k8sinfo.K8sNetInfo,
) (*point.Point, error) {
	tags := map[string]string{
		"family": key.family,

		"direction": key.direction,
		"transport": key.Transport,

		"src_ip": key.SAddr,
		"dst_ip": key.DAddr,

		"netns": key.NetNS,
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
		"method": key.method,
		"path":   key.path,

		"status_code": key.statusCode,

		"bytes_read":    value.recvBytes,
		"bytes_written": value.sendBytes,

		"count": value.count,
	}

	if value.count != 0 {
		fields["latency"] = value.latency / int64(value.count)
	}

	tags = netflow.AddK8sTags2Map(k8sNetInfo, &key.BaseKey, tags)

	kvs := point.NewTags(tags)
	kvs = append(kvs, point.NewKVs(fields)...)
	pt := point.NewPointV2("httpflow", kvs, append(
		point.CommonLoggingOptions(), point.WithTime(pTime))...)
	return pt, nil
}
