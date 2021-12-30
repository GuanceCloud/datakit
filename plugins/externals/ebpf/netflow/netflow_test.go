// +build linux,ignore ebpf

package netflow

import (
	"fmt"
	"math"
	"net"
	"sort"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

type caseConnT struct {
	connStats ConnFullStats
	conn      ConnectionInfo
	result    bool
}

func TestConnFilter(t *testing.T) {
	cases := []caseConnT{
		{
			conn: ConnectionInfo{
				Saddr: [4]uint32{0, 0, 0, 0x0100007F},
				Daddr: [4]uint32{0, 0, 0, 0x0100007F},
				Sport: 1, Dport: 1,
				Meta: ConnL3IPv4,
			},
			connStats: ConnFullStats{
				Stats: ConnectionStats{
					RecvBytes: 1,
					SentBytes: 1,
				},
			},
			result: false,
		},
		{
			conn: ConnectionInfo{
				Saddr: [4]uint32{0, 0, 0, 0x0101007F},
				Sport: 1, Dport: 1,
				Daddr: [4]uint32{0, 0, 0, 0x0100007F},
				Meta:  ConnL3IPv4,
			},
			connStats: ConnFullStats{
				Stats: ConnectionStats{
					RecvBytes: 0,
					SentBytes: 0,
				},
			},
			result: false,
		},
		{
			conn: ConnectionInfo{
				Saddr: [4]uint32{0, 0, 0, 0x01010080},
				Daddr: [4]uint32{0, 0, 0, 0x0100007F},
				Sport: 1, Dport: 1,
				Meta: ConnL3IPv4,
			},
			connStats: ConnFullStats{
				Stats: ConnectionStats{
					RecvBytes: 1,
					SentBytes: 0,
				},
			},
			result: true,
		},
		{
			conn: ConnectionInfo{
				Saddr: [4]uint32{0, 0, 0xffff0000, 0x0100007F},
				Daddr: [4]uint32{0, 0, 0xffff0000, 0x0100007F},
				Sport: 1, Dport: 1,
				Meta: ConnL3IPv6,
			},
			connStats: ConnFullStats{
				Stats: ConnectionStats{
					RecvBytes: 1,
					SentBytes: 1,
				},
			},
			result: false,
		},
		{
			conn: ConnectionInfo{
				Saddr: [4]uint32{0, 0, 0xffff0000, 0x0101008F},
				Daddr: [4]uint32{0, 0, 0xffff0000, 0x0100007F},
				Sport: 1, Dport: 1,
				Meta: ConnL3IPv6,
			},
			connStats: ConnFullStats{
				Stats: ConnectionStats{
					RecvBytes: 0,
					SentBytes: 0,
				},
			},
			result: false,
		},
		{
			conn: ConnectionInfo{
				Saddr: [4]uint32{0, 0, 0xffff0000, 0x0101008F},
				Daddr: [4]uint32{0, 0, 0xffff0000, 0x0100007F},
				Sport: 1, Dport: 1,
				Meta: ConnL3IPv6,
			},
			connStats: ConnFullStats{
				Stats: ConnectionStats{
					RecvBytes: 1,
					SentBytes: 0,
				},
			},
			result: true,
		},
		{
			conn: ConnectionInfo{
				Saddr: [4]uint32{0, 0, 0, 0},
				Daddr: [4]uint32{0, 0, 0xffff0000, 0x0100007F},
				Meta:  ConnL3IPv6,
			},
			connStats: ConnFullStats{
				Stats: ConnectionStats{
					RecvBytes: 1,
					SentBytes: 0,
				},
			},
			result: false,
		},
	}

	for k := 0; k < len(cases); k++ {
		if cases[k].result != ConnNotNeedToFilter(cases[k].conn, cases[k].connStats) {
			t.Errorf("test case %d", k)
		}
	}
}

type caseConvConn2M struct {
	conn      ConnectionInfo
	connStats ConnFullStats
	name      string
	tags      map[string]string
	ts        time.Time
	result    measurement
}

const testServiceName = "netflow"

func TestConvConn2M(t *testing.T) {
	ts := time.Now()
	connR := ConnResult{
		result: make(map[ConnectionInfo]ConnFullStats),
		tags:   make(map[string]string),
		ts:     ts,
	}
	cases := []caseConvConn2M{
		{
			conn: ConnectionInfo{
				Saddr: [4]uint32{0, 0, 0, 0x0101007D},
				Daddr: [4]uint32{0, 0, 0, 0x0100007D},
				Sport: 8080,
				Dport: 23456,
				Pid:   1222,
				Meta:  ConnL4TCP | ConnL3IPv4,
			},
			connStats: ConnFullStats{
				Stats: ConnectionStats{
					SentBytes: 1,
					RecvBytes: 1,
					Direction: ConnDirectionIncoming,
				},
				TCPStats: ConnectionTCPStats{
					Retransmits: 0,
					Rtt:         189000,
					RttVar:      20000,
				},
				TotalClosed:      1,
				TotalEstablished: 0,
			},
			tags: map[string]string{"host": "abc", "service": testServiceName},
			ts:   ts,
			result: measurement{
				tags: map[string]string{
					"host":        "abc",
					"service":     testServiceName,
					"source":      testServiceName,
					"status":      "info",
					"pid":         "1222",
					"src_ip":      "125.0.1.1",
					"src_port":    "8080",
					"src_ip_type": "other",
					"dst_ip":      "125.0.0.1",
					"dst_port":    "23456",
					"dst_ip_type": "other",
					"transport":   "tcp",
					"direction":   "incoming",
					"family":      "IPv4",
				},
				fields: map[string]interface{}{
					"bytes_written":   int64(1),
					"bytes_read":      int64(1),
					"retransmits":     int64(0),
					"rtt":             int64(189000),
					"rtt_var":         int64(20000),
					"tcp_closed":      int64(1),
					"tcp_established": int64(0),
				},
			},
		},
		{
			conn: ConnectionInfo{
				Saddr: [4]uint32{0, 0, 0, 0x0101007D},
				Daddr: [4]uint32{0, 0, 0, 0x0100007D},
				Sport: 8080,
				Dport: 23456,
				Pid:   1222,
				Meta:  ConnL4UDP | ConnL3IPv4,
			},
			connStats: ConnFullStats{
				Stats: ConnectionStats{
					SentBytes: 1,
					RecvBytes: 1,
					Direction: ConnDirectionIncoming,
				},
				TCPStats: ConnectionTCPStats{
					Retransmits: 0,
					Rtt:         189000,
					RttVar:      20000,
				},
				TotalClosed:      1,
				TotalEstablished: 0,
			},
			tags: map[string]string{"host": "abc", "service": testServiceName},
			ts:   ts,
			result: measurement{
				tags: map[string]string{
					"host":        "abc",
					"service":     testServiceName,
					"source":      testServiceName,
					"status":      "info",
					"pid":         "1222",
					"src_ip":      "125.0.1.1",
					"src_port":    "8080",
					"src_ip_type": "other",
					"dst_ip":      "125.0.0.1",
					"dst_port":    "23456",
					"dst_ip_type": "other",
					"transport":   "udp",
					"direction":   "incoming",
					"family":      "IPv4",
				},
				fields: map[string]interface{}{
					"bytes_written": int64(1),
					"bytes_read":    int64(1),
				},
			},
		},
		{
			conn: ConnectionInfo{
				Saddr: [4]uint32{0, 0, 0, 0x0101007D},
				Daddr: [4]uint32{0, 0, 0, 0x0100007D},
				Sport: math.MaxUint32,
				Dport: 23456,
				Pid:   1222,
				Meta:  ConnL4UDP | ConnL3IPv4,
			},
			connStats: ConnFullStats{
				Stats: ConnectionStats{
					SentBytes: 1,
					RecvBytes: 1,
					Direction: ConnDirectionIncoming,
				},
				TCPStats: ConnectionTCPStats{
					Retransmits: 0,
					Rtt:         189000,
					RttVar:      20000,
				},
				TotalClosed:      1,
				TotalEstablished: 0,
			},
			tags: map[string]string{"host": "abc", "service": testServiceName},
			ts:   ts,
			result: measurement{
				tags: map[string]string{
					"host":        "abc",
					"service":     testServiceName,
					"source":      testServiceName,
					"status":      "info",
					"pid":         "1222",
					"src_ip":      "125.0.1.1",
					"src_port":    "*",
					"src_ip_type": "other",
					"dst_ip":      "125.0.0.1",
					"dst_port":    "23456",
					"dst_ip_type": "other",
					"transport":   "udp",
					"direction":   "incoming",
					"family":      "IPv4",
				},
				fields: map[string]interface{}{
					"bytes_written": int64(1),
					"bytes_read":    int64(1),
				},
			},
		},
		{
			conn: ConnectionInfo{
				Saddr: [4]uint32{0, 0, 0xffff0000, 0x0101007F},
				Daddr: [4]uint32{0, 0, 0xffff0000, 0x0100007F},
				Sport: 8080,
				Dport: 23456,
				Pid:   1222,
				Meta:  ConnL4TCP | ConnL3IPv6,
			},
			connStats: ConnFullStats{
				Stats: ConnectionStats{
					SentBytes: 1,
					RecvBytes: 1,
					Direction: ConnDirectionIncoming,
				},
				TCPStats: ConnectionTCPStats{
					Retransmits: 0,
					Rtt:         189000,
					RttVar:      20000,
				},
				TotalClosed:      1,
				TotalEstablished: 0,
			},
			tags: map[string]string{"host": "abc", "service": testServiceName},
			ts:   ts,
			result: measurement{
				tags: map[string]string{
					"host":        "abc",
					"service":     testServiceName,
					"source":      testServiceName,
					"status":      "info",
					"pid":         "1222",
					"src_ip":      "127.0.1.1",
					"src_port":    "8080",
					"src_ip_type": "other",
					"dst_ip":      "127.0.0.1",
					"dst_port":    "23456",
					"dst_ip_type": "other",
					"transport":   "tcp",
					"direction":   "incoming",
					"family":      "IPv6",
				},
				fields: map[string]interface{}{
					"bytes_written":   int64(1),
					"bytes_read":      int64(1),
					"retransmits":     int64(0),
					"rtt":             int64(189000),
					"rtt_var":         int64(20000),
					"tcp_closed":      int64(1),
					"tcp_established": int64(0),
				},
			},
		},
	}
	for _, v := range cases {
		connR.result[v.conn] = v.connStats
		m, ok := ConvConn2M(v.conn, v.connStats, v.name, v.tags, v.ts).(*measurement)
		if !ok {
			t.Error("conv failed")
			continue
		}
		delete(m.fields, "message")
		if len(m.fields) != len(v.result.fields) {
			t.Error("fields length not equal")
		}
		delete(m.tags, "dst_domain")
		if len(m.tags) != len(v.result.tags) {
			t.Error("tags length not equal")
		}
		for eK, eV := range v.result.fields {
			if aV, ok := m.fields[eK]; ok {
				assert.Equal(t, eV, aV, eK)
			} else {
				t.Errorf("cannot find key %s in result fields", eK)
			}
		}
		for eK, eV := range v.result.tags {
			if aV, ok := m.tags[eK]; ok {
				assert.Equal(t, eV, aV, eK)
			} else {
				t.Errorf("cannot find key %s in result tags", eK)
			}
		}
	}
	assert.Equal(t, len(cases), len(ConvertConn2Measurement(&connR, testServiceName)))
}

type caseStatsOp struct {
	fullStats ConnFullStats
	connStats ConnectionStats
	tcpStats  ConnectionTCPStats
	resultMap map[string]ConnFullStats
}

func TestStatsOp(t *testing.T) {
	cases := caseStatsOp{
		fullStats: ConnFullStats{
			Stats: ConnectionStats{
				SentBytes:   1,
				RecvBytes:   1,
				SentPackets: 1,
				RecvPackets: 1,
				Direction:   ConnDirectionUnknown,
			},
			TCPStats: ConnectionTCPStats{
				Retransmits: 1,
				Rtt:         189000,
				RttVar:      20000,
			},
			TotalClosed:      1,
			TotalEstablished: 0,
		},
		connStats: ConnectionStats{
			SentBytes:   10,
			RecvBytes:   20,
			SentPackets: 10,
			RecvPackets: 20,
			Direction:   ConnDirectionIncoming,
		},
		tcpStats: ConnectionTCPStats{
			Retransmits: 2,
			Rtt:         180000,
			RttVar:      30000,
		},
		resultMap: map[string]ConnFullStats{
			"+": {
				Stats: ConnectionStats{
					SentBytes:   11,
					RecvBytes:   21,
					SentPackets: 11,
					RecvPackets: 21,
					Direction:   ConnDirectionIncoming,
				},
				TCPStats: ConnectionTCPStats{
					Retransmits: 3,
					Rtt:         180000,
					RttVar:      30000,
				},
			},
			"-": {
				Stats: ConnectionStats{
					SentBytes:   9,
					RecvBytes:   19,
					SentPackets: 9,
					RecvPackets: 19,
					Direction:   ConnDirectionIncoming,
				},
				TCPStats: ConnectionTCPStats{
					Retransmits: 1,
					Rtt:         180000,
					RttVar:      30000,
				},
			},
		},
	}

	for k, v := range cases.resultMap {
		r := StatsTCPOp(k, cases.fullStats, cases.connStats, cases.tcpStats)
		assert.Equal(t, v.Stats.Direction, r.Stats.Direction, "direction", k)
		assert.Equal(t, v.Stats.RecvBytes, r.Stats.RecvBytes, "recv_bytes", k)
		assert.Equal(t, v.Stats.SentBytes, r.Stats.SentBytes, "sent_bytes", k)
		assert.Equal(t, v.Stats.RecvPackets, r.Stats.RecvPackets, "recv_packets", k)
		assert.Equal(t, v.Stats.SentPackets, r.Stats.SentPackets, "sent_packets", k)

		assert.Equal(t, v.TCPStats.Retransmits, r.TCPStats.Retransmits, "retransmits", k)
		assert.Equal(t, v.TCPStats.Rtt, r.TCPStats.Rtt, "rtt", k)
		assert.Equal(t, v.TCPStats.RttVar, r.TCPStats.RttVar, "rtt_var", k)
	}
}

func TestRecord(t *testing.T) {
	netflowTracer := NewNetFlowTracer()
	conninfo := ConnectionInfo{
		Saddr: [4]uint32{0, 0, 0, 0x0101006F},
		Daddr: [4]uint32{0, 0, 0, 0x0100006F},
		Sport: 8080,
		Dport: 23456,
		Pid:   1222,
		Meta:  ConnL4TCP | ConnL3IPv4,
	}
	conninfo2 := ConnectionInfo{
		Saddr: [4]uint32{0, 0, 0, 0x0101006F},
		Daddr: [4]uint32{0, 0, 0, 0x0101016F},
		Sport: 8080,
		Dport: 23456,
		Pid:   1222,
		Meta:  ConnL4TCP | ConnL3IPv4,
	}
	conninfo3 := ConnectionInfo{
		Saddr: [4]uint32{0, 0, 0, 0x0101006F},
		Daddr: [4]uint32{0, 0, 0, 0x0101016F},
		Sport: 8088,
		Dport: 3456,
		Pid:   1233,
		Meta:  ConnL4TCP | ConnL3IPv4,
	}

	connFullStats := ConnFullStats{
		Stats: ConnectionStats{
			SentBytes: 1,
			RecvBytes: 1,
			Direction: ConnDirectionIncoming,
		},
		TCPStats: ConnectionTCPStats{
			Retransmits: 0,
			Rtt:         189000,
			RttVar:      20000,
		},
		TotalClosed:      0,
		TotalEstablished: 1,
	}

	connFullStatsResult := ConnFullStats{
		Stats: ConnectionStats{
			SentBytes: 1,
			RecvBytes: 1,
			Direction: ConnDirectionIncoming,
		},
		TCPStats: ConnectionTCPStats{
			Retransmits: 0,
			Rtt:         189000,
			RttVar:      20000,
		},
		TotalClosed:      0,
		TotalEstablished: 0,
	}

	// test updateLastActive, 设定上一周期存在两个未关闭的连接
	netflowTracer.connStatsRecord.updateLastActive(conninfo, connFullStats)
	assert.Equal(t, netflowTracer.connStatsRecord.lastActiveConns[conninfo], connFullStatsResult)
	netflowTracer.connStatsRecord.updateLastActive(conninfo2, connFullStats)
	assert.Equal(t, netflowTracer.connStatsRecord.lastActiveConns[conninfo], connFullStatsResult)
	assert.Equal(t, 2, len(netflowTracer.connStatsRecord.lastActiveConns))

	// ==================================================================
	// 存在一个上一周期未关闭的连接，接收到一个 closed event，调用 ClosedEventHandler

	closedEvent := ConncetionClosedInfoC{
		conn_info: _Ctype_struct_connection_info{
			saddr: [4]_Ctype_uint{0, 0, 0, 0x0101006F},
			daddr: [4]_Ctype_uint{0, 0, 0, 0x0100006F},
			sport: 8080,
			dport: 23456,
			pid:   1222,
			meta:  _Ctype_uint(ConnL4TCP | ConnL3IPv4),
		},
		conn_stats: _Ctype_struct_connection_stats{
			sent_bytes: 1,
			recv_bytes: 1,
			direction:  ConnDirectionIncoming,
		},
		conn_tcp_stats: _Ctype_struct_connection_tcp_stats{
			retransmits: 0,
			rtt:         189000,
			rtt_var:     20000,
		},
	}
	connClosedFullStatsResult := ConnFullStats{
		Stats: ConnectionStats{
			SentBytes: 0,
			RecvBytes: 0,
			Direction: ConnDirectionIncoming,
		},
		TCPStats: ConnectionTCPStats{
			Retransmits: 0,
			Rtt:         189000,
			RttVar:      20000,
		},
		TotalClosed:      1,
		TotalEstablished: 0,
	}
	eventStructMock := struct {
		addr uintptr
		len  int
		cap  int
	}{
		addr: uintptr(unsafe.Pointer(&closedEvent)),
		len:  int(unsafe.Sizeof(closedEvent)),
		cap:  int(unsafe.Sizeof(closedEvent)),
	}
	data := *(*[]byte)(unsafe.Pointer(&eventStructMock))

	netflowTracer.ClosedEventHandler(1, data, nil, nil)
	event := <-netflowTracer.closedEventCh
	netflowTracer.connStatsRecord.updateClosedUseEvent(event)
	assert.Equal(t, 1, len(netflowTracer.connStatsRecord.lastActiveConns))
	assert.Equal(t, 1, len(netflowTracer.connStatsRecord.closedConns))
	connInfo := ConnectionInfo{
		Saddr: [4]uint32{0, 0, 0, 0x0101006F},
		Daddr: [4]uint32{0, 0, 0, 0x0100006F},
		Sport: 8080,
		Dport: 23456,
		Pid:   1222,
		Meta:  ConnL4TCP | ConnL3IPv4,
	}
	assert.Equal(t, connClosedFullStatsResult, netflowTracer.connStatsRecord.closedConns[connInfo])

	// ===================================
	// 一个已关闭连接的再次建立，并被关闭，接收 closed event，调用 closedEventHandler
	netflowTracer.ClosedEventHandler(1, data, nil, nil)
	event = <-netflowTracer.closedEventCh
	netflowTracer.connStatsRecord.updateClosedUseEvent(event)
	connClosedFullStatsResult2 := ConnFullStats{
		Stats: ConnectionStats{
			SentBytes: 1,
			RecvBytes: 1,
			Direction: ConnDirectionIncoming,
		},
		TCPStats: ConnectionTCPStats{
			Retransmits: 0,
			Rtt:         189000,
			RttVar:      20000,
		},
		TotalClosed:      2,
		TotalEstablished: 1,
	}
	assert.Equal(t, 1, len(netflowTracer.connStatsRecord.lastActiveConns))
	assert.Equal(t, 1, len(netflowTracer.connStatsRecord.closedConns))
	assert.Equal(t, connClosedFullStatsResult2, netflowTracer.connStatsRecord.closedConns[connInfo])

	// =================================
	// 一个本周期内建立后关闭的连接，调用 closedEventHandler, 首次记录

	closedEvent = ConncetionClosedInfoC{
		conn_info: _Ctype_struct_connection_info{
			saddr: [4]_Ctype_uint{0, 0, 0, 0x0101006F},
			daddr: [4]_Ctype_uint{0, 0, 0, 0x0200006F},
			sport: 8080,
			dport: 23456,
			pid:   1222,
			meta:  _Ctype_uint(ConnL4TCP | ConnL3IPv4),
		},
		conn_stats: _Ctype_struct_connection_stats{
			sent_bytes: 1,
			recv_bytes: 1,
			direction:  ConnDirectionIncoming,
		},
		conn_tcp_stats: _Ctype_struct_connection_tcp_stats{
			retransmits: 0,
			rtt:         189000,
			rtt_var:     20000,
		},
	}

	connClosedFullStatsResult = ConnFullStats{
		Stats: ConnectionStats{
			SentBytes: 1,
			RecvBytes: 1,
			Direction: ConnDirectionIncoming,
		},
		TCPStats: ConnectionTCPStats{
			Retransmits: 0,
			Rtt:         189000,
			RttVar:      20000,
		},
		TotalClosed:      1,
		TotalEstablished: 1,
	}

	connInfo = ConnectionInfo{
		Saddr: [4]uint32{0, 0, 0, 0x0101006F},
		Daddr: [4]uint32{0, 0, 0, 0x0200006F},
		Sport: 8080,
		Dport: 23456,
		Pid:   1222,
		Meta:  ConnL4TCP | ConnL3IPv4,
	}
	eventStructMock = struct {
		addr uintptr
		len  int
		cap  int
	}{
		addr: uintptr(unsafe.Pointer(&closedEvent)),
		len:  int(unsafe.Sizeof(closedEvent)),
		cap:  int(unsafe.Sizeof(closedEvent)),
	}
	data = *(*[]byte)(unsafe.Pointer(&eventStructMock))
	netflowTracer.ClosedEventHandler(1, data, nil, nil)
	event = <-netflowTracer.closedEventCh
	netflowTracer.connStatsRecord.updateClosedUseEvent(event)
	assert.Equal(t, 1, len(netflowTracer.connStatsRecord.lastActiveConns))
	assert.Equal(t, 2, len(netflowTracer.connStatsRecord.closedConns))
	assert.Equal(t, connClosedFullStatsResult, netflowTracer.connStatsRecord.closedConns[connInfo])

	// ================================
	// 模拟从 bpfmap 中获取当前 active 的连接，并 merge 记录的 lastActive、closed

	// 存在于 lastActiveConns, stats op = "-"
	connFullStats.Stats.RecvBytes += 1
	ar := netflowTracer.connStatsRecord.mergeWithClosedLastActive(conninfo2, connFullStats)
	er := ConnFullStats{
		Stats: ConnectionStats{
			SentBytes: 0,
			RecvBytes: 1,
			Direction: ConnDirectionIncoming,
		},
		TCPStats: ConnectionTCPStats{
			Retransmits: 0,
			Rtt:         189000,
			RttVar:      20000,
		},
		TotalClosed:      0,
		TotalEstablished: 0,
	}
	assert.Equal(t, er, ar)
	connFullStats.Stats.RecvBytes -= 1

	// =================

	// 存在于 closedConns, stats op = "+"
	ar = netflowTracer.connStatsRecord.mergeWithClosedLastActive(conninfo, connFullStats)
	er = ConnFullStats{
		Stats: ConnectionStats{
			SentBytes: 2,
			RecvBytes: 2,
			Direction: ConnDirectionIncoming,
		},
		TCPStats: ConnectionTCPStats{
			Retransmits: 0,
			Rtt:         189000,
			RttVar:      20000,
		},
		TotalClosed:      2,
		TotalEstablished: 2,
	}
	assert.Equal(t, er, ar)

	// ================
	// 首次建立
	ar = netflowTracer.connStatsRecord.mergeWithClosedLastActive(conninfo3, connFullStats)
	er = connFullStats
	assert.Equal(t, er, ar)
}

func TestConnMeta(t *testing.T) {
	var meta uint32
	meta = ConnL3IPv4 | ConnL4TCP
	assert.Equal(t, true, connAddrIsIPv4(meta))
	assert.Equal(t, true, connProtocolIsTCP(meta))

	meta = meta&(^ConnL3Mask) | ConnL3IPv6
	assert.Equal(t, false, connAddrIsIPv4(meta))
	meta = meta&(^ConnL4Mask) | ConnL4UDP
	assert.Equal(t, false, connProtocolIsTCP(meta))
}

func TestDirection(t *testing.T) {
	assert.Equal(t, "incoming", connDirection2Str(ConnDirectionIncoming))
	assert.Equal(t, "outgoing", connDirection2Str(ConnDirectionOutgoing))
	assert.Equal(t, "outgoing", connDirection2Str(ConnDirectionAuto))
	assert.Equal(t, "outgoing", connDirection2Str(ConnDirectionUnknown))
}

func TestIPv4Type(t *testing.T) {
	cases := map[uint32]string{
		// 172.16
		uint32(0x10AC): "private",
		// 172.31
		uint32(0x1FAC): "private",
		// 10.
		uint32(0x0A): "private",
		// 192.168
		uint32(0xA8C0): "private",
		// 224.
		uint32(0xE0): "multicast",
		// 239.
		uint32(0xEF): "multicast",
		// 127.
		uint32(0x7F): "loopback",
		// 101.2.3.4
		uint32(0x04030265): "other",
	}
	for k, v := range cases {
		assert.Equal(t, v, connIPv4Type(k))
	}
}

func TestU32BE2NETIp(t *testing.T) {
	cases := map[uint32]string{
		uint32(0x10AC):     "172.16.0.0",
		uint32(0x7F):       "127.0.0.0",
		uint32(0x04030265): "101.2.3.4",
	}
	for k, v := range cases {
		addr := [4]uint32{0, 0, 0, k}
		netip := U32BEToIP(addr, false)
		assert.Equal(t, netip.String(), v)
	}

	casesv6 := map[[4]uint32]string{
		{0x11aa00fe, 0, 0, 0}:      "fe00:aa11::",
		{0xef00, 0, 0, 0xaabbfeda}: "ef::dafe:bbaa",
	}

	for k, v := range casesv6 {
		netip := U32BEToIP(k, true)
		assert.Equal(t, netip.String(), v)
	}
}

func TestIPv6Type(t *testing.T) {
	cases := map[[4]uint32]string{
		// fc01::
		{0x000001fc, 0, 0, 0}: "private",
		// fd00::
		{0x000000fd, 0, 0, 0}: "private",
		// ff01::
		{0x000001ff, 0, 0, 0}: "multicast",
		// ::1
		{0, 0, 0, 0x01000000}: "loopback",
		// 2004:0400::
		{0x00040420, 0, 0, 0}: "other",
	}
	for k, v := range cases {
		assert.Equal(t, v, connIPv6Type(k))
	}
}

func TestConnMerge(t *testing.T) {
	result := ConnResult{
		result: map[ConnectionInfo]ConnFullStats{
			{
				Saddr: [4]uint32{0, 0, 0, 0x01},
				Sport: 1000,
				Daddr: [4]uint32{0, 0, 0, 0x01},
				Dport: 101,
				Meta:  ConnL3IPv4 | ConnL4TCP,
				Pid:   10000,
			}: {
				Stats: ConnectionStats{
					SentBytes: 10,
					RecvBytes: 20,
				},
				TCPStats: ConnectionTCPStats{
					Rtt:    100 * 1000,
					RttVar: 200 * 1000,
				},
				TotalClosed:      1,
				TotalEstablished: 2,
			}, { // tcp sport>
				Saddr: [4]uint32{0, 0, 0, 0x01},
				Sport: math.MaxUint32,
				Daddr: [4]uint32{0, 0, 0, 0x01},
				Dport: 101,
				Meta:  ConnL3IPv4 | ConnL4TCP,
				Pid:   10000,
			}: {
				Stats: ConnectionStats{
					SentBytes: 20,
					RecvBytes: 40,
				},
				TCPStats: ConnectionTCPStats{
					Rtt:    100 * 1000,
					RttVar: 200 * 1000,
				},
				TotalClosed:      2,
				TotalEstablished: 3,
			}, { // ipv4 udp
				Saddr: [4]uint32{0, 0, 0, 0x02},
				Sport: 42234,
				Daddr: [4]uint32{0, 0, 0, 0x02},
				Dport: 201,
				Meta:  ConnL3IPv4 | ConnL4UDP,
				Pid:   10001,
			}: {
				Stats: ConnectionStats{
					SentBytes: 10,
					RecvBytes: 20,
				},
			}, { // ipv6 udp
				Saddr: [4]uint32{0x01, 0, 0, 0x02},
				Sport: math.MaxUint32,
				Daddr: [4]uint32{0x01, 0, 0, 0x02},
				Dport: 201,
				Meta:  ConnL3IPv6 | ConnL4UDP,
				Pid:   10001,
			}: {
				Stats: ConnectionStats{
					SentBytes: 20,
					RecvBytes: 40,
				},
			},
		},
	}
	preResult := ConnResult{
		result: map[ConnectionInfo]ConnFullStats{
			{ // tcp sport>
				Saddr: [4]uint32{0, 0, 0, 0x01},
				Sport: 41234,
				Daddr: [4]uint32{0, 0, 0, 0x01},
				Dport: 101,
				Meta:  ConnL3IPv4 | ConnL4TCP,
				Pid:   10000,
			}: {
				Stats: ConnectionStats{
					SentBytes: 10,
					RecvBytes: 20,
				},
				TCPStats: ConnectionTCPStats{
					Rtt:    100 * 1000,
					RttVar: 200 * 1000,
				},
				TotalClosed:      1,
				TotalEstablished: 1,
			}, {
				Saddr: [4]uint32{0, 0, 0, 0x01},
				Sport: 41235,
				Daddr: [4]uint32{0, 0, 0, 0x01},
				Dport: 101,
				Meta:  ConnL3IPv4 | ConnL4TCP,
				Pid:   10000,
			}: {
				Stats: ConnectionStats{
					SentBytes: 10,
					RecvBytes: 20,
				},
				TCPStats: ConnectionTCPStats{
					Rtt:    100 * 1000,
					RttVar: 200 * 1000,
				},
				TotalClosed:      1,
				TotalEstablished: 2,
			}, {
				Saddr: [4]uint32{0, 0, 0, 0x01},
				Sport: 1000,
				Daddr: [4]uint32{0, 0, 0, 0x01},
				Dport: 101,
				Meta:  ConnL3IPv4 | ConnL4TCP,
				Pid:   10000,
			}: {
				Stats: ConnectionStats{
					SentBytes: 10,
					RecvBytes: 20,
				},
				TCPStats: ConnectionTCPStats{
					Rtt:    100 * 1000,
					RttVar: 200 * 1000,
				},
				TotalClosed:      1,
				TotalEstablished: 2,
			}, { // ipv4 udp
				Saddr: [4]uint32{0, 0, 0, 0x02},
				Sport: 42234,
				Daddr: [4]uint32{0, 0, 0, 0x02},
				Dport: 201,
				Meta:  ConnL3IPv4 | ConnL4UDP,
				Pid:   10001,
			}: {
				Stats: ConnectionStats{
					SentBytes: 10,
					RecvBytes: 20,
				},
			}, { // ipv6 udp
				Saddr: [4]uint32{0x01, 0, 0, 0x02},
				Sport: 42234,
				Daddr: [4]uint32{0x01, 0, 0, 0x02},
				Dport: 201,
				Meta:  ConnL3IPv6 | ConnL4UDP,
				Pid:   10001,
			}: {
				Stats: ConnectionStats{
					SentBytes: 10,
					RecvBytes: 20,
				},
			},
			{ // ipv6 udp
				Saddr: [4]uint32{0x01, 0, 0, 0x02},
				Sport: 42235,
				Daddr: [4]uint32{0x01, 0, 0, 0x02},
				Dport: 201,
				Meta:  ConnL3IPv6 | ConnL4UDP,
				Pid:   10001,
			}: {
				Stats: ConnectionStats{
					SentBytes: 10,
					RecvBytes: 20,
				},
			},
		},
	}
	MergeConns(&preResult)
	if len(preResult.result) != len(result.result) {
		t.Error("len not equal")
	}
	for k, v := range result.result {
		if vp, ok := preResult.result[k]; !ok {
			t.Error("conn not find")
		} else {
			assert.Equal(t, v, vp)
		}
	}
}

func TestConnSort(t *testing.T) {
	connListResult := ConnInfoList{}
	conn, _ := newConn(ConnL3IPv4|ConnL4TCP, "1.1.1.2", "5.2.2.2", 22, 31122, 111)
	connListResult = append(connListResult, *conn)
	conn, _ = newConn(ConnL3IPv4|ConnL4TCP, "1.1.1.2", "5.2.2.2", 22, 31122, 222)
	connListResult = append(connListResult, *conn)
	conn, _ = newConn(ConnL3IPv4|ConnL4TCP, "1.1.1.2", "5.2.2.2", 22, 31122, 333)
	connListResult = append(connListResult, *conn)
	conn, _ = newConn(ConnL3IPv4|ConnL4TCP, "2.2.2.3", "7.1.1.1", 80, 51121, 345)
	connListResult = append(connListResult, *conn)
	conn, _ = newConn(ConnL3IPv4|ConnL4TCP, "2.2.2.3", "8.1.1.1", 80, 53322, 345)
	connListResult = append(connListResult, *conn)
	conn, _ = newConn(ConnL3IPv4|ConnL4TCP, "2.2.2.3", "8.2.1.1", 80, 53322, 345)
	connListResult = append(connListResult, *conn)
	conn, _ = newConn(ConnL3IPv4|ConnL4UDP, "2.2.2.3", "8.2.1.1", 5353, 5353, 3456)
	connListResult = append(connListResult, *conn)
	conn, _ = newConn(ConnL3IPv6|ConnL4TCP, "fe80::", "::ff", 80, 53322, 3457)
	connListResult = append(connListResult, *conn)
	conn, _ = newConn(ConnL3IPv6|ConnL4UDP, "fe80::", "::ff", 5353, 5353, 3456)
	connListResult = append(connListResult, *conn)

	connList := connListResult // swap item
	connList[4], connList[1] = connList[1], connList[4]
	connList[2], connList[3] = connList[3], connList[2]
	connList[5], connList[6] = connList[6], connList[5]
	connList[0], connList[8] = connList[8], connList[0]

	sort.Sort(connList) // sort
	for k, conn := range connList {
		if connListResult[k] != conn {
			t.Errorf(conn.String())
		}
	}
}

func TestMultiPidConns(t *testing.T) {
	connList := []ConnectionInfo{}
	fullStatsList := []ConnFullStats{}

	connBase, _ := newConn(ConnL3IPv4|ConnL4TCP, "1.1.1.1", "2.2.2.2", 80, 52211, 1)
	fullStatsBase := newFullStats(ConnL3IPv4|ConnL4TCP, 0, 0, 0, 0)

	// case 1:
	// 0s: e(pid1), e(pid2)
	// 60s: -(pid2)
	// 120s: e/c(pid3)

	conn := *connBase
	fullStats := *fullStatsBase

	conn.Pid = 1
	fullStats.TotalEstablished = 1
	fullStats.TotalClosed = 0
	connList = append(connList, conn)
	fullStatsList = append(fullStatsList, fullStats)

	conn.Pid = 2
	fullStats.TotalEstablished = 1
	fullStats.TotalClosed = 0
	connList = append(connList, conn)
	fullStatsList = append(fullStatsList, fullStats)

	conn.Pid = 2
	fullStats.TotalEstablished = 0
	fullStats.TotalClosed = 0
	connList = append(connList, conn)
	fullStatsList = append(fullStatsList, fullStats)

	conn.Pid = 3
	fullStats.TotalEstablished = 1
	fullStats.TotalClosed = 1
	connList = append(connList, conn)
	fullStatsList = append(fullStatsList, fullStats)

	ecResult := [][2]int64{ // i0:e , i1:c
		{1, 0}, {0, 0}, {0, 0}, {0, 1},
	}
	if len(ecResult) != len(fullStatsList) || len(fullStatsList) != len(connList) {
		t.Error("Check your code")
	}
	connTCPWithoutPid := newConnTCPWithoutPid()
	i := 0
	for k := range connList {
		v := connTCPWithoutPid.Update(connList[k], fullStatsList[k])
		assert.Equal(t, v.TotalEstablished, ecResult[i][0], "established")
		assert.Equal(t, v.TotalClosed, ecResult[i][1], "closed")
		i++
	}

	// case 2:
	// 0s: e(pid1), e(pid2), e/c(pid3), e(pid4)
	connList = []ConnectionInfo{}
	fullStatsList = []ConnFullStats{}
	conn.Pid = 1
	fullStats.TotalEstablished = 1
	fullStats.TotalClosed = 0
	connList = append(connList, conn)
	fullStatsList = append(fullStatsList, fullStats)

	conn.Pid = 2
	fullStats.TotalEstablished = 1
	fullStats.TotalClosed = 0
	connList = append(connList, conn)
	fullStatsList = append(fullStatsList, fullStats)

	conn.Pid = 3
	fullStats.TotalEstablished = 1
	fullStats.TotalClosed = 1
	connList = append(connList, conn)
	fullStatsList = append(fullStatsList, fullStats)

	conn.Pid = 4
	fullStats.TotalEstablished = 1
	fullStats.TotalClosed = 0
	connList = append(connList, conn)
	fullStatsList = append(fullStatsList, fullStats)

	ecResult = [][2]int64{ // i0:e , i1:c
		{1, 0}, {0, 0}, {0, 1}, {0, 0},
	}
	if len(ecResult) != len(fullStatsList) || len(fullStatsList) != len(connList) {
		t.Error("Check your code")
	}
	connTCPWithoutPid.CleanupConns()
	i = 0
	for k := range connList {
		v := connTCPWithoutPid.Update(connList[k], fullStatsList[k])
		assert.Equal(t, v.TotalEstablished, ecResult[i][0], "established")
		assert.Equal(t, v.TotalClosed, ecResult[i][1], "closed")
		i++
	}

	// case 3:
	// 0s: e(pid1)
	// 60s: 0(pid1)
	// 120s: e2/c2(pid1)
	// 180s: c1(pid1)
	connList = []ConnectionInfo{}
	fullStatsList = []ConnFullStats{}
	conn.Pid = 1

	fullStats.TotalEstablished = 1
	fullStats.TotalClosed = 0
	connList = append(connList, conn)
	fullStatsList = append(fullStatsList, fullStats)

	fullStats.TotalEstablished = 0
	fullStats.TotalClosed = 0
	connList = append(connList, conn)
	fullStatsList = append(fullStatsList, fullStats)

	fullStats.TotalEstablished = 2
	fullStats.TotalClosed = 2
	connList = append(connList, conn)
	fullStatsList = append(fullStatsList, fullStats)

	fullStats.TotalEstablished = 0
	fullStats.TotalClosed = 1
	connList = append(connList, conn)
	fullStatsList = append(fullStatsList, fullStats)

	ecResult = [][2]int64{ // i0:e , i1:c
		{1, 0}, {0, 0}, {2, 2}, {0, 1},
	}
	if len(ecResult) != len(fullStatsList) || len(fullStatsList) != len(connList) {
		t.Error("Check your code")
	}
	connTCPWithoutPid.CleanupConns()
	i = 0
	for k := range connList {
		v := connTCPWithoutPid.Update(connList[k], fullStatsList[k])
		assert.Equal(t, v.TotalEstablished, ecResult[i][0], "established")
		assert.Equal(t, v.TotalClosed, ecResult[i][1], "closed")
		i++
	}
	connTCPWithoutPid.CleanupConns()
	if len(connTCPWithoutPid.Conns) != 0 {
		t.Error("len(connTCPWithoutPid.Conns)", len(connTCPWithoutPid.Conns))
	}
}

func TestIPPortFilterIn(t *testing.T) {
	cases := map[ConnectionInfo]bool{}
	c, _ := newConn(ConnL3IPv4, "127.1.0.1", "1.1.1.1", 1, 1, 1)
	cases[*c] = false
	c, _ = newConn(ConnL3IPv4, "12.1.0.1", "127.0.0.1", 1, 1, 1)
	cases[*c] = false
	c, _ = newConn(ConnL3IPv4, "12.1.0.1", "12.0.0.1", 0, 1, 1)
	cases[*c] = false
	c, _ = newConn(ConnL3IPv4, "12.1.0.1", "12.0.0.1", 1, 0, 1)
	cases[*c] = false
	c, _ = newConn(ConnL3IPv4, "12.1.0.1", "12.0.0.1", 1, 1, 1)
	cases[*c] = true
	c, _ = newConn(ConnL3IPv6, "2::1", "::1", 1, 1, 1)
	cases[*c] = false
	c, _ = newConn(ConnL3IPv6, "::1", "2::1", 1, 1, 1)
	cases[*c] = false
	c, _ = newConn(ConnL3IPv6, "2::1", "2::1", 1, 1, 1)
	cases[*c] = true
	for k, v := range cases {
		if IPPortFilterIn(&k) != v {
			t.Error(k)
		}
	}
}

func newConn(meta uint32, sip, dip string, sport, dport uint32, pid uint32) (*ConnectionInfo, error) {
	srcip := net.ParseIP(sip).To16() // 16 bytes
	dstip := net.ParseIP(dip).To16()
	if srcip == nil || dstip == nil {
		return nil, fmt.Errorf("parse ip failed")
	}

	conn := ConnectionInfo{
		Saddr: *(*[4]uint32)(unsafe.Pointer(&srcip[0])),
		Daddr: *(*[4]uint32)(unsafe.Pointer(&dstip[0])),
		Sport: sport,
		Dport: dport,
		Pid:   pid,
		Meta:  meta,
	}

	return &conn, nil
}

func newFullStats(meta uint32, sent_bytes, recv_bytes uint64, tcp_established, tcp_closed int64) *ConnFullStats {
	fullStats := ConnFullStats{
		Stats: ConnectionStats{
			SentBytes: sent_bytes,
			RecvBytes: recv_bytes,
		},
	}

	if connProtocolIsTCP(meta) {
		fullStats.TotalClosed = tcp_closed
		fullStats.TotalEstablished = tcp_established
	}
	return &fullStats
}
