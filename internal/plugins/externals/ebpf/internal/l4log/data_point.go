//go:build linux
// +build linux

package l4log

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/output"
)

var _colsnames = []string{
	"txrx", "src_mac", "dst_mac", "flags", "seq", "ack_seq", "payload_size", "win", "ts",
}

const chunkTimeoutDuration = int64(time.Second * 30)

func (conns *TCPConns) conv2Point(k *PMeta, v *PValue,
	opt []point.Option, rm bool, nicIPList []string,
) ([]*point.Point, error) {
	pts := []*point.Point{}
	tsnow := time.Now().Nanosecond()
	tags := map[string]string{
		"src_ip":       k.SrcIP,
		"dst_ip":       k.DstIP,
		"src_port":     strconv.FormatInt(int64(k.SrcPort), 10),
		"dst_port":     strconv.FormatInt(int64(k.DstPort), 10),
		"l4_proto":     "tcp",
		"nic_mac":      conns.ifaceNameMAC[1],
		"nic_name":     conns.ifaceNameMAC[0],
		"nic_traceid":  fmt.Sprintf("%d_%d", v.tcpInfo.synSeq, v.tcpInfo.synAckSeq),
		"netns":        conns.nsUID,
		"vni_id":       strconv.FormatInt(int64(k.VNIID), 10),
		"vxlan_packet": strconv.FormatBool(k.VXLAN),
	}

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
	delete(tags, "direction")
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

	if v.tcpInfo.tcpRTT < 0 {
		v.tcpInfo.tcpRTT = 0
	}

	for _, elem := range feedHTTPElem {
		if elem.hState == 0 {
			continue
		}
		kvs := point.NewTags(tags)
		// tags
		kvs = kvs.Add("trace_id", elem.TraceID, true, true)
		kvs = kvs.Add("parent_id", elem.ParentID, true, true)
		kvs = kvs.Add("direction", elem.Direction, true, true)
		kvs = kvs.Add("l7_proto", "http", true, true)
		kvs = kvs.Add("http_path", elem.Path, true, true)
		kvs = kvs.Add("http_status_code", elem.StatusCode, true, true)
		kvs = kvs.Add("http_method", elem.Method, true, true)
		kvs = kvs.Add("l7_traceid", fmt.Sprintf("%d_%d",
			elem.ReqSeq1st, elem.RespSeq1st), true, true)

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

		kvs = kvs.Add("cost_req_sent", reqDlDur, false, true)
		kvs = kvs.Add("cost_cnt_dl", respDlDur, false, true)

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
		conns.aggHTTP.Append(k, elem, conns.nsUID, v.v6, v.sMACEQ, nicIPList, waitRespDur)

		// conv to millsecond
		kvs = kvs.Add("cost_resp_wait", float64(waitRespDur)/float64(time.Millisecond), false, true)

		// same as tcp
		kvs = kvs.Add("tx_packets", elem.txPkts, false, true)
		kvs = kvs.Add("rx_packets", elem.rxPkts, false, true)
		kvs = kvs.Add("tx_bytes", elem.txBytes, false, true)
		kvs = kvs.Add("rx_bytes", elem.rxBytes, false, true)
		kvs = kvs.Add("tx_retrans", elem.txRetransmits, false, true)
		kvs = kvs.Add("rx_retrans", elem.rxRetransmits, false, true)

		msg := map[string]any{
			"l4_proto": "tcp",
			"l7_proto": "http",
			"http":     elem,
			"tcp": map[string]any{
				"tx_bytes":         elem.txBytes,
				"rx_bytes":         elem.rxBytes,
				"tx_packets":       elem.txPkts,
				"rx_packets":       elem.rxPkts,
				"tx_first_byte_ts": elem.txFirstByteTS,
				"tx_last_byte_ts":  elem.txLastByteTS,
				"rx_first_byte_ts": elem.rxFirstByteTS,
				"rx_last_byte_ts":  elem.rxLastByteTS,
				"tx_retrans":       elem.txRetransmits,
				"rx_retrans":       elem.rxRetransmits,
			},
		}

		buf, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		kvs = kvs.Add("message", string(buf), false, true)
		if enableNetlog {
			pts = append(pts, point.NewPointV2("bpf_net_l7_log", kvs, append(
				opt, point.WithExtraTags(conns.tags), point.WithTime(time.Unix(0, reqTS)))...))
		}
	}

	chunkCount := len(v.tcpInfo.chunk)
	cCur := 0
	for _, chunk := range v.tcpInfo.chunk {
		if cCur >= chunkCount-1 && !rm {
			chunkElemLen := len(chunk.TCPSreries)
			if chunkElemLen > 0 {
				lastTS := chunk.TCPSreries[chunkElemLen-1].TS
				dur0NotTimeout := (int64(tsnow) - lastTS) < chunkTimeoutDuration
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
		kvs := point.NewTags(tags)
		kvs = kvs.Add("chunk_id", chunk.ChunkID, false, true)
		kvs = kvs.Add("tx_seq_min", chunk.txSeq[0], false, true)
		kvs = kvs.Add("tx_seq_max", chunk.txSeq[1], false, true)
		kvs = kvs.Add("rx_seq_min", chunk.rxSeq[0], false, true)
		kvs = kvs.Add("rx_seq_max", chunk.rxSeq[1], false, true)

		if isSYNChunk(chunk.chunkKind) {
			kvs = kvs.Add("chunk_syn", true, false, true)
			s0 := v.tcpInfo.synfinTS[0]
			s1 := v.tcpInfo.synfinTS[1]
			if s0 != 0 && s1 != 0 && s1 > s0 {
				kvs = kvs.Add("tcp_3whs_cost", float64(s1-s0)/float64(time.Millisecond), false, true)
			}
		}
		if isFINChunk(chunk.chunkKind) {
			kvs = kvs.Add("chunk_fin", true, false, true)
			f0 := v.tcpInfo.synfinTS[2]
			f1 := v.tcpInfo.synfinTS[3]
			if f0 != 0 && f1 != 0 && f1 > f0 {
				kvs = kvs.Add("tcp_4whs_cost", float64(f1-f0)/float64(time.Millisecond), false, true)
			}
		}

		kvs = kvs.Add("tcp_rtt", v.tcpInfo.tcpRTT, false, true)
		kvs = kvs.Add("tx_packets", chunk.TXPacket, false, true)
		kvs = kvs.Add("rx_packets", chunk.RXPacket, false, true)
		kvs = kvs.Add("tx_bytes", chunk.TxBytes, false, true)
		kvs = kvs.Add("rx_bytes", chunk.RxBytes, false, true)
		kvs = kvs.Add("tx_retrans", chunk.RetransmitsTx, false, true)
		kvs = kvs.Add("tx_retrans", chunk.RetransmitsRx, false, true)
		kvs = kvs.Add("tcp_syn_retrans", chunk.RetransmitsSYN, false, true)
		chunk.TCPColName = _colsnames
		m := map[string]any{
			"l4_proto": "tcp",
			"tcp":      chunk,
		}
		buf, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		kvs = kvs.Add("message", string(buf), false, true)

		var ts int64
		if len(chunk.TCPSreries) > 0 {
			ts = chunk.TCPSreries[0].TS
		}

		if enableNetlog {
			pts = append(pts, point.NewPointV2("bpf_net_l4_log", kvs, append(
				opt, point.WithExtraTags(conns.tags), point.WithTime(time.Unix(0, ts)))...))
		}

		chunk.TCPColName = nil
	}

	if cCur <= chunkCount-1 {
		v.tcpInfo.chunk = v.tcpInfo.chunk[cCur:]
	} else {
		v.tcpInfo.chunk = nil
	}

	return pts, nil
}

const (
	maxFeedCount = 128
)

func (conns *TCPConns) feedNetworkLog(pool *connMap,
	cal2mslDelete bool, forceDelete bool, nicNameMAC [2]string,
	nicIPList []string,
) {
	tn := time.Now()
	ts := tn.UnixNano()
	pts := make([]*point.Point, 0, maxFeedCount)
	count := 0

	if pool == nil {
		return
	}

	for k, v := range pool.m {
		if v == nil {
			continue
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

		opt := append(point.CommonLoggingOptions(), point.WithTime(tn))
		if ptsGot, err := conns.conv2Point(&k, v, opt, removeConn, nicIPList); err == nil {
			count += len(ptsGot)
			pts = append(pts, ptsGot...)
		} else {
			log.Errorf("conv metric and event to point failed: %w", err)
		}
		if count >= maxFeedCount {
			if len(pts) > 0 && enableNetlog {
				if err := output.FeedPoint(conns.url, pts, false); err != nil {
					log.Errorf("feed point(toatl %d) failed: %w", len(pts), err)
				}
			}
			pts = make([]*point.Point, 0, maxFeedCount)
			count = 0
		}
	}

	if len(pts) > 0 && enableNetlog {
		if err := output.FeedPoint(conns.url, pts, false); err != nil {
			log.Errorf("feed point(toatl %d) failed: %w", len(pts), err)
		}
	}
}
