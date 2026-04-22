//go:build linux
// +build linux

package l4log

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"

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
	default:
		key.direction = netflow.DirectionOutgoing
	}

	key.direction, key.SPort, key.DPort = netflow.NormalizeDirectionAndPorts(key.direction, key.SPort, key.DPort)

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
