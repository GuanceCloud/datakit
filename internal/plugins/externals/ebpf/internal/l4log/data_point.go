//go:build linux
// +build linux

package l4log

import (
	"strconv"
	"time"

	"github.com/goccy/go-json"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/exporter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
)

const chunkTimeoutDuration = int64(time.Second * 30)

func formatUintPair(a, b uint64) string {
	var buf [41]byte
	dst := strconv.AppendUint(buf[:0], a, 10)
	dst = append(dst, '_')
	dst = strconv.AppendUint(dst, b, 10)
	return string(dst)
}

type httpLogTCPSummary struct {
	TXBytes   int64 `json:"tx_bytes"`
	RXBytes   int64 `json:"rx_bytes"`
	TXPackets int64 `json:"tx_packets"`
	RXPackets int64 `json:"rx_packets"`
	TXRetrans int   `json:"tx_retrans"`
	RXRetrans int   `json:"rx_retrans"`
}

type httpLogMessage struct {
	L4Proto string            `json:"l4_proto"`
	L7Proto string            `json:"l7_proto"`
	HTTP    *HTTPLogElem      `json:"http"`
	TCP     httpLogTCPSummary `json:"tcp"`
}

type http2LogMessage struct {
	L4Proto string        `json:"l4_proto"`
	L7Proto string        `json:"l7_proto"`
	HTTP2   *HTTP2LogElem `json:"http2"`
}

func (conns *TCPConns) netlogConv2Point(k *PMeta, v *PValue,
	opt []point.Option, rm bool, nicIPList []string,
) ([]*point.Point, error) {
	pts := []*point.Point{}
	tsnow := ntp.Now().UnixNano()

	if !enableNetlog {
		trimHTTPLogState(v, rm)
		trimTCPLogState(v, rm, tsnow)
		return nil, nil
	}

	baseKVs := buildCommKVs(k, v, conns)

	{ // http log and metric
		var feedHTTPElem []*HTTPLogElem
		var keeplastHTTPElem []*HTTPLogElem
		if !rm {
			hReqLen := len(v.httpInfo.elems)
			if hReqLen > 0 && !v.httpInfo.elems[hReqLen-1].hFinished {
				feedHTTPElem = v.httpInfo.elems[:hReqLen-1]
				keeplastHTTPElem = append(keeplastHTTPElem, v.httpInfo.elems[hReqLen-1])
			} else {
				feedHTTPElem = v.httpInfo.elems
			}
			v.httpInfo.elems = keeplastHTTPElem
		} else {
			feedHTTPElem = v.httpInfo.elems
			v.httpInfo.elems = nil
		}

		for _, elem := range feedHTTPElem {
			if elem.hState == 0 {
				continue
			}

			if kvs, reqTS, ok, err := buildHTTPLog(k, v, elem, baseKVs,
				&conns.aggHTTP, conns.nsUID, nicIPList); err != nil {
				log.Errorf("build http log failed: %s", err.Error())
			} else if ok && enableNetlog {
				pts = append(pts, point.NewPoint("bpf_net_l7_log", kvs, append(
					opt, point.WithTime(time.Unix(0, reqTS)))...))
			}
		}
	}

	{ // tcp log
		chunkCount := len(v.tcpInfo.chunk)
		cCur := 0
		for _, chunk := range v.tcpInfo.chunk {
			if cCur >= chunkCount-1 && !rm {
				chunkElemLen := len(chunk.TCPSreries)
				if chunkElemLen > 0 {
					lastTS := chunk.TCPSreries[chunkElemLen-1].TS
					dur0NotTimeout := (tsnow - (lastTS + chunk.TimePos)) < chunkTimeoutDuration
					dur1NotTimeout := true
					if chunkElemLen >= 2 {
						dur1NotTimeout = (lastTS - chunk.TCPSreries[0].TS) < chunkTimeoutDuration
					}
					if dur0NotTimeout && dur1NotTimeout {
						break
					}
				} else {
					break
				}
			}
			cCur++

			kvs, ts, ok, err := buildTCPLog(chunk, tsnow, baseKVs, v)
			if err != nil {
				log.Errorf("build tcp log failed: %s", err.Error())
			} else if ok && enableNetlog {
				pts = append(pts, point.NewPoint("bpf_net_l4_log", kvs, append(
					opt, point.WithTime(time.Unix(0, ts)))...))
			}
		}

		if cCur <= chunkCount-1 {
			v.tcpInfo.chunk = v.tcpInfo.chunk[cCur:]
		} else {
			v.tcpInfo.chunk = nil
		}
	}
	return pts, nil
}

func trimHTTPLogState(v *PValue, rm bool) {
	if rm {
		v.httpInfo.elems = nil
		return
	}

	hReqLen := len(v.httpInfo.elems)
	if hReqLen > 0 && !v.httpInfo.elems[hReqLen-1].hFinished {
		v.httpInfo.elems = v.httpInfo.elems[hReqLen-1:]
		return
	}

	v.httpInfo.elems = nil
}

func trimTCPLogState(v *PValue, rm bool, tsnow int64) {
	chunkCount := len(v.tcpInfo.chunk)
	cCur := 0

	for _, chunk := range v.tcpInfo.chunk {
		if cCur >= chunkCount-1 && !rm {
			chunkElemLen := len(chunk.TCPSreries)
			if chunkElemLen > 0 {
				lastTS := chunk.TCPSreries[chunkElemLen-1].TS
				dur0NotTimeout := (tsnow - (lastTS + chunk.TimePos)) < chunkTimeoutDuration
				dur1NotTimeout := true
				if chunkElemLen >= 2 {
					dur1NotTimeout = (lastTS - chunk.TCPSreries[0].TS) < chunkTimeoutDuration
				}
				if dur0NotTimeout && dur1NotTimeout {
					break
				}
			} else {
				break
			}
		}
		cCur++
	}

	if cCur <= chunkCount-1 {
		v.tcpInfo.chunk = v.tcpInfo.chunk[cCur:]
	} else {
		v.tcpInfo.chunk = nil
	}
}

const (
	maxFeedCount = 128
)

const baseTagCacheTTL = time.Minute

func needConnMapFullScan(pool *connMap, timeout time.Duration, force bool) bool {
	if force || pool == nil {
		return true
	}
	if time.Since(pool.tn) >= timeout {
		return true
	}
	if enableNetlog && time.Since(pool.lastFullScan) >= time.Duration(chunkTimeoutDuration) {
		return true
	}
	return false
}

func (conns *TCPConns) feedNetworkLog(pool *connMap,
	cal2mslDelete bool, forceDelete bool, fullScan bool, nicIPList []string,
) {
	tn := ntp.Now()
	ts := tn.UnixNano()
	pts := make([]*point.Point, 0, maxFeedCount)
	count := 0
	var logOpt []point.Option

	if pool == nil {
		return
	}

	if enableNetlog {
		logOpt = append(point.CommonLoggingOptions(), point.WithTime(tn))
	}

	processEntry := func(k PMeta, v *PValue) {
		if v == nil {
			return
		}

		var removeConn bool

		switch {
		case forceDelete:
			removeConn = true
			// force delete and do not swap
			delete(pool.m, k)
			// pool.delete(k)
		case cal2mslDelete:
			if v.reuseByNxt {
				removeConn = true
				pool.delete(k)
			} else if ts-v.lastGetTS >= twoMSL.Nanoseconds() {
				removeConn = true
				pool.delete(k)
			}
		default:
			if v.tcpInfo.Closed() {
				removeConn = true
				pool.delete(k)
			} else if ts-v.lastGetTS >= defaultTCPKeepAlive.Nanoseconds() {
				removeConn = true
				pool.delete(k)
			}
		}

		v.tcpInfo.metric.RTT = v.tcpInfo.rtt.getRTT() / int64(time.Microsecond)
		if removeConn {
			if !v.tcpInfo.metric.recClose[0] || v.tcpInfo.metric.recClose[1] {
				v.tcpInfo.metric.recClose[0] = true
				v.tcpInfo.metric.recClose[1] = true
				conns.agg.Append(&k, &v.tcpInfo.metric, conns.nsUID,
					v.tcpInfo.direction, v.v6, v.sMACEQ, nicIPList)
			}
		} else {
			conns.agg.Append(&k, &v.tcpInfo.metric, conns.nsUID,
				v.tcpInfo.direction, v.v6, v.sMACEQ, nicIPList)
		}

		if ptsGot, err := conns.netlogConv2Point(&k, v, logOpt, removeConn,
			nicIPList); err == nil && len(ptsGot) > 0 {
			count += len(ptsGot)
			pts = append(pts, ptsGot...)
		} else if err != nil {
			log.Errorf("conv metric and event to point failed: %w", err)
		}
		if count >= maxFeedCount {
			if len(pts) > 0 && enableNetlog {
				if err := exporter.FeedPoint("bpf-netlog/netlog", point.Logging, pts); err != nil {
					log.Errorf("feed point(toatl %d) failed: %w", len(pts), err)
				}
			}
			pts = make([]*point.Point, 0, maxFeedCount)
			count = 0
		}
	}

	if fullScan {
		for k, v := range pool.m {
			processEntry(k, v)
		}
		pool.finishFullScan(tn)
	} else {
		for _, k := range pool.drainDirty() {
			v, ok := pool.get(k)
			if !ok {
				continue
			}
			processEntry(k, v)
		}
	}

	if len(pts) > 0 && enableNetlog {
		if err := exporter.FeedPoint("bpf-netlog/netlog", point.Logging, pts); err != nil {
			log.Errorf("feed point(toatl %d) failed: %w", len(pts), err)
		}
	}
}

func buildCommTags(k *PMeta, v *PValue, conns *TCPConns) map[string]string {
	tags := make(map[string]string, 18)
	tags["src_ip"] = k.SrcIP
	tags["dst_ip"] = k.DstIP
	tags["src_port"] = strconv.FormatInt(int64(k.SrcPort), 10)
	tags["dst_port"] = strconv.FormatInt(int64(k.DstPort), 10)
	tags["l4_proto"] = "tcp"
	tags["nic_mac"] = conns.ifaceNameMAC[1]
	tags["nic_name"] = conns.ifaceNameMAC[0]
	tags["nic_traceid"] = formatUintPair(uint64(v.tcpInfo.synSeq), uint64(v.tcpInfo.synAckSeq))
	tags["netns"] = conns.nsUID
	tags["vni_id"] = strconv.FormatInt(int64(k.VNIID), 10)
	tags["vxlan_packet"] = strconv.FormatBool(k.VXLAN)

	tags["host_network"] = strconv.FormatBool(conns.hostNetwork)
	tags["virtual_nic"] = strconv.FormatBool(conns.virtualNIC)

	if v.connTraceID != nil {
		tags["inner_traceid"] = v.connTraceID.StringHex()
	}
	tags = netflow.AddK8sTags2Map(k8sNetInfo, &netflow.BaseKey{
		SAddr:     k.SrcIP,
		DAddr:     k.DstIP,
		SPort:     uint32(k.SrcPort),
		DPort:     uint32(k.DstPort),
		Transport: "tcp",
	}, tags)

	direction := netflow.NormalizeDirectionByPorts(v.tcpInfo.direction.String(),
		uint32(k.SrcPort), uint32(k.DstPort))

	switch direction {
	case netflow.DirectionOutgoing:
		tags["conn_side"] = "client"
		tags["client_ip"] = tags["src_ip"]
		tags["client_port"] = tags["src_port"]
		tags["server_ip"] = tags["dst_ip"]
		tags["server_port"] = tags["dst_port"]
	case netflow.DirectionIncoming:
		tags["conn_side"] = "server"

		tags["client_ip"] = tags["dst_ip"]
		tags["client_port"] = tags["dst_port"]
		tags["server_ip"] = tags["src_ip"]
		tags["server_port"] = tags["src_port"]
	default:
		tags["conn_side"] = "unknown"
		tags["client_ip"] = tags["src_ip"]
		tags["client_port"] = tags["src_port"]
		tags["server_ip"] = tags["dst_ip"]
		tags["server_port"] = tags["dst_port"]
	}

	tags["direction"] = direction

	return tags
}

func buildCommKVs(k *PMeta, v *PValue, conns *TCPConns) point.KVs {
	now := ntp.Now().UnixNano()
	dir := v.tcpInfo.direction.String()
	runtimeTags := conns.runtimeTags()

	if cached := v.baseKVsCache; len(cached) > 0 &&
		v.baseTagsCacheDir == dir &&
		now-v.baseTagsCacheTS < baseTagCacheTTL.Nanoseconds() {
		return appendCommDynamicKVs(cloneBaseKVs(cached, len(runtimeTags)+10), k, conns.nsUID, runtimeTags)
	}

	direction := netflow.NormalizeDirectionByPorts(dir,
		uint32(k.SrcPort), uint32(k.DstPort))

	kvs := make(point.KVs, 0, 18)
	kvs = appendStringTagKVFast(kvs, "src_ip", k.SrcIP)
	kvs = appendStringTagKVFast(kvs, "dst_ip", k.DstIP)
	kvs = appendStringTagKVFast(kvs, "src_port", strconv.FormatInt(int64(k.SrcPort), 10))
	kvs = appendStringTagKVFast(kvs, "dst_port", strconv.FormatInt(int64(k.DstPort), 10))
	kvs = appendStringTagKVFast(kvs, "l4_proto", "tcp")
	kvs = appendStringTagKVFast(kvs, "nic_mac", conns.ifaceNameMAC[1])
	kvs = appendStringTagKVFast(kvs, "nic_name", conns.ifaceNameMAC[0])
	kvs = appendStringTagKVFast(kvs, "nic_traceid", formatUintPair(uint64(v.tcpInfo.synSeq), uint64(v.tcpInfo.synAckSeq)))
	kvs = appendStringTagKVFast(kvs, "netns", conns.nsUID)
	kvs = appendStringTagKVFast(kvs, "vni_id", strconv.FormatInt(int64(k.VNIID), 10))
	kvs = appendStringTagKVFast(kvs, "vxlan_packet", strconv.FormatBool(k.VXLAN))
	kvs = appendStringTagKVFast(kvs, "host_network", strconv.FormatBool(conns.hostNetwork))
	kvs = appendStringTagKVFast(kvs, "virtual_nic", strconv.FormatBool(conns.virtualNIC))

	if v.connTraceID != nil {
		kvs = appendStringTagKVFast(kvs, "inner_traceid", v.connTraceID.StringHex())
	}

	switch direction {
	case netflow.DirectionOutgoing:
		kvs = appendStringTagKVFast(kvs, "conn_side", "client")
		kvs = appendStringTagKVFast(kvs, "client_ip", k.SrcIP)
		kvs = appendStringTagKVFast(kvs, "client_port", strconv.FormatInt(int64(k.SrcPort), 10))
		kvs = appendStringTagKVFast(kvs, "server_ip", k.DstIP)
		kvs = appendStringTagKVFast(kvs, "server_port", strconv.FormatInt(int64(k.DstPort), 10))
	case netflow.DirectionIncoming:
		kvs = appendStringTagKVFast(kvs, "conn_side", "server")
		kvs = appendStringTagKVFast(kvs, "client_ip", k.DstIP)
		kvs = appendStringTagKVFast(kvs, "client_port", strconv.FormatInt(int64(k.DstPort), 10))
		kvs = appendStringTagKVFast(kvs, "server_ip", k.SrcIP)
		kvs = appendStringTagKVFast(kvs, "server_port", strconv.FormatInt(int64(k.SrcPort), 10))
	default:
		kvs = appendStringTagKVFast(kvs, "conn_side", "unknown")
		kvs = appendStringTagKVFast(kvs, "client_ip", k.SrcIP)
		kvs = appendStringTagKVFast(kvs, "client_port", strconv.FormatInt(int64(k.SrcPort), 10))
		kvs = appendStringTagKVFast(kvs, "server_ip", k.DstIP)
		kvs = appendStringTagKVFast(kvs, "server_port", strconv.FormatInt(int64(k.DstPort), 10))
	}

	kvs = appendStringTagKVFast(kvs, "direction", direction)

	v.baseKVsCache = kvs
	v.baseTagsCacheTS = now
	v.baseTagsCacheDir = dir
	return appendCommDynamicKVs(cloneBaseKVs(kvs, len(runtimeTags)+10), k, conns.nsUID, runtimeTags)
}

func appendCommDynamicKVs(kvs point.KVs, k *PMeta, nsUID string, runtimeTags map[string]string) point.KVs {
	if k != nil {
		if domain := netflow.LookupPeerDomain(k.DstIP, uint32(k.DstPort), "tcp", nsUID); domain != "" {
			kvs = appendStringTagKVFast(kvs, "dst_domain", domain)
			kvs = appendStringTagKVFast(kvs, "server_domain", domain)
		}
		kvs = appendTagMapKVs(kvs, netflow.AddK8sTags2Map(k8sNetInfo, &netflow.BaseKey{
			SAddr:     k.SrcIP,
			DAddr:     k.DstIP,
			SPort:     uint32(k.SrcPort),
			DPort:     uint32(k.DstPort),
			Transport: "tcp",
		}, nil), map[string]struct{}{"direction": {}})
	}
	return appendTagMapKVs(kvs, runtimeTags, nil)
}

func cloneBaseKVs(base point.KVs, extra int) point.KVs {
	dst := make(point.KVs, len(base), len(base)+extra)
	copy(dst, base)
	return dst
}

func appendTagMapKVs(base point.KVs, tags map[string]string, skip map[string]struct{}) point.KVs {
	if len(tags) == 0 {
		return base
	}

	dst := base
	for key, val := range tags {
		if skip != nil {
			if _, ok := skip[key]; ok {
				continue
			}
		}
		if containsKVKey(dst, key) {
			continue
		}
		dst = appendStringTagKVFast(dst, key, val)
	}
	return dst
}

func containsKVKey(kvs point.KVs, key string) bool {
	for _, kv := range kvs {
		if kv != nil && kv.Key == key {
			return true
		}
	}
	return false
}

func appendStringTagKVFast(kvs point.KVs, key, val string) point.KVs {
	return append(kvs, &point.Field{
		Key:   key,
		Val:   &point.Field_S{S: val},
		IsTag: true,
	})
}

func appendStringFieldKVFast(kvs point.KVs, key, val string) point.KVs {
	return append(kvs, &point.Field{
		Key: key,
		Val: &point.Field_S{S: val},
	})
}

func appendBoolFieldKVFast(kvs point.KVs, key string, val bool) point.KVs {
	return append(kvs, &point.Field{
		Key: key,
		Val: &point.Field_B{B: val},
	})
}

func appendFloatFieldKVFast(kvs point.KVs, key string, val float64) point.KVs {
	return append(kvs, &point.Field{
		Key: key,
		Val: &point.Field_F{F: val},
	})
}

func appendIntFieldKVFast(kvs point.KVs, key string, val int64) point.KVs {
	return append(kvs, &point.Field{
		Key: key,
		Val: &point.Field_I{I: val},
	})
}

func appendUintFieldKVFast(kvs point.KVs, key string, val uint64) point.KVs {
	return append(kvs, &point.Field{
		Key: key,
		Val: &point.Field_U{U: val},
	})
}

func buildHTTPLog(k *PMeta, v *PValue, elem *HTTPLogElem, baseKVs point.KVs,
	agg *FlowAggHTTP, nsUID string, nicIPList []string,
) (point.KVs, int64, bool, error) {
	kvs := cloneBaseKVs(baseKVs, 24)
	if k != nil && elem.Host != "" {
		netflow.RecordPeerDomain(k.DstIP, uint32(k.DstPort), "tcp", nsUID, elem.Host)
		if !containsKVKey(kvs, "dst_domain") {
			kvs = appendStringTagKVFast(kvs, "dst_domain", elem.Host)
		}
		if !containsKVKey(kvs, "server_domain") {
			kvs = appendStringTagKVFast(kvs, "server_domain", elem.Host)
		}
	}

	// tags
	kvs = appendStringTagKVFast(kvs, "trace_id", elem.TraceID)
	kvs = appendStringTagKVFast(kvs, "parent_id", elem.ParentID)
	kvs = appendStringTagKVFast(kvs, "l7_proto", "http")
	kvs = appendStringTagKVFast(kvs, "http_path", elem.Path)
	kvs = appendIntFieldKVFast(kvs, "http_status_code", int64(elem.StatusCode))
	kvs = appendStringTagKVFast(kvs, "http_method", elem.Method)
	kvs = appendStringFieldKVFast(kvs, "req_seq", strconv.FormatInt(int64(elem.reqSeq), 10))
	kvs = appendStringFieldKVFast(kvs, "resp_seq", strconv.FormatInt(int64(elem.respSeq), 10))
	kvs = appendStringTagKVFast(kvs, "l7_traceid", formatUintPair(uint64(elem.reqSeq), uint64(elem.respSeq)))

	if elem.Direction == DOutging {
		kvs = appendIntFieldKVFast(kvs, "tx_seq", int64(elem.reqSeq))
		kvs = appendIntFieldKVFast(kvs, "rx_seq", int64(elem.respSeq))
	} else {
		kvs = appendIntFieldKVFast(kvs, "tx_seq", int64(elem.respSeq))
		kvs = appendIntFieldKVFast(kvs, "rx_seq", int64(elem.reqSeq))
	}

	var reqDlDur, respDlDur float64
	var reqTS int64
	switch elem.Direction {
	case DOutging:
		reqTS = elem.txFirstByteTS
		if elem.txLastByteTS > 0 && elem.txFirstByteTS > 0 {
			respDlDur = float64(elem.txLastByteTS-elem.txFirstByteTS) / float64(time.Millisecond)
		}
		if elem.rxLastByteTS > 0 && elem.rxFirstByteTS > 0 {
			reqDlDur = float64(elem.rxLastByteTS-elem.rxFirstByteTS) / float64(time.Millisecond)
		}

	case DIncoming:
		reqTS = elem.rxFirstByteTS
		if elem.rxLastByteTS > 0 && elem.rxFirstByteTS > 0 {
			respDlDur = float64(elem.rxLastByteTS-elem.rxFirstByteTS) / float64(time.Millisecond)
		}
		if elem.txLastByteTS > 0 && elem.txFirstByteTS > 0 {
			reqDlDur = float64(elem.txLastByteTS-elem.txFirstByteTS) / float64(time.Millisecond)
		}
	}

	kvs = appendFloatFieldKVFast(kvs, "cost_req_sent", reqDlDur)
	kvs = appendFloatFieldKVFast(kvs, "cost_cnt_dl", respDlDur)

	var waitRespDur int64
	switch {
	case elem.txFirstByteTS > elem.rxLastByteTS:
		waitRespDur = elem.txFirstByteTS - elem.rxLastByteTS
	case elem.rxFirstByteTS > elem.txLastByteTS:
		waitRespDur = elem.rxFirstByteTS - elem.txLastByteTS
	}

	if waitRespDur > int64(time.Hour) {
		waitRespDur = 0
	}
	if agg != nil {
		agg.Append(k, elem, nsUID, v.v6, v.sMACEQ, nicIPList, waitRespDur)
	}

	// conv to millsecond
	kvs = appendFloatFieldKVFast(kvs, "cost_resp_wait", float64(waitRespDur)/float64(time.Millisecond))

	// same as tcp
	kvs = appendIntFieldKVFast(kvs, "tx_packets", elem.txPkts)
	kvs = appendIntFieldKVFast(kvs, "rx_packets", elem.rxPkts)
	kvs = appendIntFieldKVFast(kvs, "tx_bytes", elem.txBytes)
	kvs = appendIntFieldKVFast(kvs, "rx_bytes", elem.rxBytes)
	kvs = appendIntFieldKVFast(kvs, "tx_retrans", int64(elem.txRetransmits))
	kvs = appendIntFieldKVFast(kvs, "rx_retrans", int64(elem.rxRetransmits))

	msg, err := httpLogMessageJSON(elem)
	if err != nil {
		return nil, 0, false, err
	}
	kvs = appendStringFieldKVFast(kvs, "message", msg)
	return kvs, reqTS, true, nil
}

func buildTCPLog(chunk *PktChunk, tsnow int64,
	baseKVs point.KVs, v *PValue,
) (point.KVs, int64, bool, error) {
	kvs := cloneBaseKVs(baseKVs, 20)
	kvs = appendStringTagKVFast(kvs, "l7_proto", v.tcpInfo.l7proto.String())
	kvs = appendIntFieldKVFast(kvs, "chunk_id", chunk.ChunkID)
	kvs = appendUintFieldKVFast(kvs, "tx_seq_min", uint64(chunk.txSeq[0]))
	kvs = appendUintFieldKVFast(kvs, "tx_seq_max", uint64(chunk.txSeq[1]))
	kvs = appendUintFieldKVFast(kvs, "rx_seq_min", uint64(chunk.rxSeq[0]))
	kvs = appendUintFieldKVFast(kvs, "rx_seq_max", uint64(chunk.rxSeq[1]))

	if isSYNChunk(chunk.chunkKind) {
		kvs = appendBoolFieldKVFast(kvs, "chunk_syn", true)
		s0 := v.tcpInfo.synfinTS[0]
		s1 := v.tcpInfo.synfinTS[1]
		if s0 != 0 && s1 != 0 && s1 > s0 {
			kvs = appendFloatFieldKVFast(kvs, "tcp_3whs_cost", float64(s1-s0)/float64(time.Millisecond))
		}
	}
	if isFINChunk(chunk.chunkKind) {
		kvs = appendBoolFieldKVFast(kvs, "chunk_fin", true)
		f0 := v.tcpInfo.synfinTS[2]
		f1 := v.tcpInfo.synfinTS[3]
		if f0 != 0 && f1 != 0 && f1 > f0 {
			kvs = appendFloatFieldKVFast(kvs, "tcp_4whs_cost", float64(f1-f0)/float64(time.Millisecond))
		}
	}

	kvs = appendFloatFieldKVFast(kvs, "tcp_rtt", float64(v.tcpInfo.rtt.getRTT())/float64(time.Millisecond))
	kvs = appendIntFieldKVFast(kvs, "tx_packets", chunk.TXPacket)
	kvs = appendIntFieldKVFast(kvs, "rx_packets", chunk.RXPacket)
	kvs = appendIntFieldKVFast(kvs, "tx_bytes", int64(chunk.TxBytes))
	kvs = appendIntFieldKVFast(kvs, "rx_bytes", int64(chunk.RxBytes))
	kvs = appendIntFieldKVFast(kvs, "tx_retrans", int64(chunk.RetransmitsTx))
	kvs = appendIntFieldKVFast(kvs, "rx_retrans", int64(chunk.RetransmitsRx))
	kvs = appendIntFieldKVFast(kvs, "tcp_syn_retrans", int64(v.tcpInfo.RetransmitsSYN))

	msg, err := tcpLogMessageJSON(chunk)
	if err != nil {
		return nil, 0, false, err
	}
	kvs = appendStringFieldKVFast(kvs, "message", msg)

	if len(chunk.TCPSreries) > 0 {
		return kvs, chunk.TCPSreries[0].TS + chunk.TimePos, true, nil
	} else {
		return kvs, tsnow, true, nil
	}
}

var _ = buildH2Log

func buildH2Log(k *PMeta, v *PValue, elem *HTTP2LogElem, baseKVs point.KVs,
	agg *FlowAggHTTP, nsUID string, nicIPList []string,
) (point.KVs, int64, bool, error) {
	kvs := cloneBaseKVs(baseKVs, 16)
	if k != nil && elem.Host != "" {
		netflow.RecordPeerDomain(k.DstIP, uint32(k.DstPort), "tcp", nsUID, elem.Host)
		if !containsKVKey(kvs, "dst_domain") {
			kvs = appendStringTagKVFast(kvs, "dst_domain", elem.Host)
		}
		if !containsKVKey(kvs, "server_domain") {
			kvs = appendStringTagKVFast(kvs, "server_domain", elem.Host)
		}
	}

	// tags
	kvs = appendStringTagKVFast(kvs, "trace_id", elem.TraceID)
	kvs = appendStringTagKVFast(kvs, "parent_id", elem.ParentID)
	kvs = appendStringTagKVFast(kvs, "l7_proto", "http2")
	kvs = appendStringTagKVFast(kvs, "req_seq", strconv.FormatInt(int64(elem.reqSeq), 10))
	kvs = appendStringTagKVFast(kvs, "resp_seq", strconv.FormatInt(int64(elem.respSeq), 10))
	kvs = appendStringTagKVFast(kvs, "l7_traceid", formatUintPair(uint64(elem.reqSeq), uint64(elem.respSeq)))

	// fields
	kvs = appendStringFieldKVFast(kvs, "http_method", elem.Method)
	kvs = appendStringFieldKVFast(kvs, "http_path", elem.Path)
	kvs = appendIntFieldKVFast(kvs, "http_status_code", int64(elem.StatusCode))

	var reqDlDur, respDlDur float64
	var reqTS int64
	switch elem.Direction {
	case DOutging:
		reqTS = elem.txFirstByteTS
		if elem.txLastByteTS > 0 && elem.txFirstByteTS > 0 {
			respDlDur = float64(elem.txLastByteTS-elem.txFirstByteTS) / float64(time.Millisecond)
		}
		if elem.rxLastByteTS > 0 && elem.rxFirstByteTS > 0 {
			reqDlDur = float64(elem.rxLastByteTS-elem.rxFirstByteTS) / float64(time.Millisecond)
		}

	case DIncoming:
		reqTS = elem.rxFirstByteTS
		if elem.rxLastByteTS > 0 && elem.rxFirstByteTS > 0 {
			respDlDur = float64(elem.rxLastByteTS-elem.rxFirstByteTS) / float64(time.Millisecond)
		}
		if elem.txLastByteTS > 0 && elem.txFirstByteTS > 0 {
			reqDlDur = float64(elem.txLastByteTS-elem.txFirstByteTS) / float64(time.Millisecond)
		}
	}

	kvs = appendFloatFieldKVFast(kvs, "cost_req_sent", reqDlDur)
	kvs = appendFloatFieldKVFast(kvs, "cost_cnt_dl", respDlDur)

	var waitRespDur int64
	switch {
	case elem.txFirstByteTS > elem.rxLastByteTS:
		waitRespDur = elem.txFirstByteTS - elem.rxLastByteTS
	case elem.rxFirstByteTS > elem.txLastByteTS:
		waitRespDur = elem.rxFirstByteTS - elem.txLastByteTS
	}

	if waitRespDur > int64(time.Hour) {
		waitRespDur = 0
	}

	// 由于平台的不支持，暂时不分离 grpc 进行数据聚合，且 h2 作为 http 聚合
	if agg != nil {
		agg.AppendH2(k, elem, nsUID, v.v6, v.sMACEQ, nicIPList, waitRespDur)
	}

	// conv to millsecond
	kvs = appendFloatFieldKVFast(kvs, "cost_resp_wait", float64(waitRespDur)/float64(time.Millisecond))

	msg, err := http2LogMessageJSON(elem)
	if err != nil {
		return nil, 0, false, err
	}

	kvs = appendStringFieldKVFast(kvs, "message", msg)
	return kvs, reqTS, true, nil
}

func httpLogMessageJSON(elem *HTTPLogElem) (string, error) {
	if elem != nil && !elem.messageDirty && elem.messageCache != "" {
		return elem.messageCache, nil
	}

	buf, err := json.Marshal(httpLogMessage{
		L4Proto: "tcp",
		L7Proto: "http",
		HTTP:    elem,
		TCP: httpLogTCPSummary{
			TXBytes:   elem.txBytes,
			RXBytes:   elem.rxBytes,
			TXPackets: elem.txPkts,
			RXPackets: elem.rxPkts,
			TXRetrans: elem.txRetransmits,
			RXRetrans: elem.rxRetransmits,
		},
	})
	if err != nil {
		return "", err
	}

	elem.messageCache = string(buf)
	elem.messageDirty = false
	return elem.messageCache, nil
}

func tcpLogMessageJSON(chunk *PktChunk) (string, error) {
	if chunk == nil {
		return "", nil
	}
	if chunk != nil && !chunk.messageDirty && chunk.messageCache != "" {
		return chunk.messageCache, nil
	}

	buf := make([]byte, 0, 160+len(chunk.TCPSreries)*48+chunk.macCount*24+len(chunk.extraMAC)*32)
	buf = append(buf, `{"l4_proto":"tcp","tcp":`...)
	buf = chunk.appendJSON(buf)
	buf = append(buf, '}')

	chunk.messageCache = string(buf)
	chunk.messageDirty = false
	return chunk.messageCache, nil
}

func http2LogMessageJSON(elem *HTTP2LogElem) (string, error) {
	if elem != nil && !elem.messageDirty && elem.messageCache != "" {
		return elem.messageCache, nil
	}

	buf, err := json.Marshal(http2LogMessage{
		L4Proto: "tcp",
		L7Proto: "http2",
		HTTP2:   elem,
	})
	if err != nil {
		return "", err
	}

	elem.messageCache = string(buf)
	elem.messageDirty = false
	return elem.messageCache, nil
}
