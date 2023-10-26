//go:build linux
// +build linux

package netflow

import (
	"fmt"
	"math"
	"net"
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

type measurement struct {
	tags   map[string]string
	fields map[string]interface{}
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
		if cases[k].result != ConnNotNeedToFilter(&cases[k].conn, &cases[k].connStats) {
			t.Errorf("test case %d", k)
		}
	}
}

type caseConvConn2M struct {
	conn      ConnectionInfo
	connStats ConnFullStats
	// name      string
	tags   map[string]string
	ts     time.Time
	result measurement
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

					"process_name": "N/A",
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

					"process_name": "N/A",
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

					"process_name": "N/A",
				},
				fields: map[string]interface{}{
					"bytes_written": int64(1),
					"bytes_read":    int64(1),
				},
			},
		},
		{
			conn: ConnectionInfo{
				Saddr: [4]uint32{0, 0, 0xffff0000, 0x0101005F},
				Daddr: [4]uint32{0, 0, 0xffff0000, 0x0100005F},
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
					"status":      "info",
					"pid":         "1222",
					"src_ip":      "95.0.1.1",
					"src_port":    "8080",
					"src_ip_type": "other",
					"dst_ip":      "95.0.0.1",
					"dst_port":    "23456",
					"dst_ip_type": "other",
					"transport":   "tcp",
					"direction":   "incoming",
					"family":      "IPv4",

					"process_name": "N/A",
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
		pt, err := ConvConn2M(v.conn, v.connStats, srcNameM, v.tags,
			v.ts, nil)
		if err != nil {
			t.Error(err)
			continue
		}
		tags := pt.MapTags()
		fields := pt.InfluxFields()
		delete(fields, "message")
		if len(fields) != len(v.result.fields) {
			t.Error("fields length not equal")
		}
		delete(tags, "dst_domain")
		if len(tags) != len(v.result.tags) {
			t.Error("tags length not equal")
		}
		for eK, eV := range v.result.fields {
			if aV, ok := fields[eK]; ok {
				assert.Equal(t, eV, aV, eK)
			} else {
				t.Errorf("cannot find key %s in result fields", eK)
			}
		}
		for eK, eV := range v.result.tags {
			if aV, ok := tags[eK]; ok {
				assert.Equal(t, eV, aV, eK)
			} else {
				t.Errorf("cannot find key %s in result tags", eK)
			}
		}
	}
	agg := FlowAgg{}
	for k, v := range connR.result {
		_ = agg.Append(k, v)
	}
	pts := agg.ToPoint(connR.tags, nil)
	assert.Equal(t, len(cases), len(pts))
	agg.Clean()
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
	netflowTracer := NewNetFlowTracer(nil)
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

	// test updateLastActive, set two unclosed connections in the previous cycle
	netflowTracer.connStatsRecord.updateLastActive(conninfo, connFullStats)
	assert.Equal(t, netflowTracer.connStatsRecord.lastActiveConns[conninfo], connFullStatsResult)
	netflowTracer.connStatsRecord.updateLastActive(conninfo2, connFullStats)
	assert.Equal(t, netflowTracer.connStatsRecord.lastActiveConns[conninfo], connFullStatsResult)
	assert.Equal(t, 2, len(netflowTracer.connStatsRecord.lastActiveConns))

	// ==================================================================
	// There is a connection that was not closed in the previous cycle, a closed event is received, and the ClosedEventHandler is called.

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
	// A closed connection is re-established and closed, receive closed event, call closedEventHandler
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
	// A connection closed after establishment within this period, calling closedEventHandler, the first record

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
	// Simulate getting the current active connection from bpfmap, and merge the recorded lastActive and closed

	// present in lastActiveConns, stats op = "-"
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

	// present in closedConns, stats op = "+"
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
	// First established
	ar = netflowTracer.connStatsRecord.mergeWithClosedLastActive(conninfo3, connFullStats)
	er = connFullStats
	assert.Equal(t, er, ar)
}

func TestConnMeta(t *testing.T) {
	var meta uint32
	meta = ConnL3IPv4 | ConnL4TCP
	assert.Equal(t, true, ConnAddrIsIPv4(meta))
	assert.Equal(t, true, ConnProtocolIsTCP(meta))

	meta = meta&(^ConnL3Mask) | ConnL3IPv6
	assert.Equal(t, false, ConnAddrIsIPv4(meta))
	meta = meta&(^ConnL4Mask) | ConnL4UDP
	assert.Equal(t, false, ConnProtocolIsTCP(meta))
}

func TestDirection(t *testing.T) {
	assert.Equal(t, "incoming", ConnDirection2Str(ConnDirectionIncoming))
	assert.Equal(t, "outgoing", ConnDirection2Str(ConnDirectionOutgoing))
	assert.Equal(t, "outgoing", ConnDirection2Str(ConnDirectionAuto))
	assert.Equal(t, "outgoing", ConnDirection2Str(ConnDirectionUnknown))
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
		assert.Equal(t, v, ConnIPv4Type(k))
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
		assert.Equal(t, v, ConnIPv6Type(k))
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
