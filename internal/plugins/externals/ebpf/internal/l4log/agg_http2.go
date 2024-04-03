//go:build linux
// +build linux

package l4log

import (
	"math"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
)

func (agg *FlowAggHTTP) AppendH2(info *PMeta, stats *HTTP2LogElem, netns string,
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
