//go:build linux
// +build linux

package netflow

import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
)

// #include "../c/netflow/conn_stats.h"
import "C"

type ConnectionInfoC C.struct_connection_info

type ConnectionStatsC C.struct_connection_stats

type ConnectionTCPStatsC C.struct_connection_tcp_stats

type ConncetionClosedInfoC C.struct_connection_closed_info

type ConnectionInfo struct {
	Saddr [4]uint32
	Daddr [4]uint32
	Sport uint32
	Dport uint32
	Pid   uint32
	Netns uint32
	Meta  uint32

	NATDaddr [4]uint32
	NATDport uint32

	ProcessName string
}

func ReadConnInfo(conn *ConnectionInfoC, dnatAddr [4]uint32, dnatPort uint32) ConnectionInfo {
	return ConnectionInfo{
		Saddr:    (*(*[4]uint32)(unsafe.Pointer(&conn.saddr))), //nolint:gosec
		Daddr:    (*(*[4]uint32)(unsafe.Pointer(&conn.daddr))), //nolint:gosec
		Sport:    uint32(conn.sport),
		Dport:    uint32(conn.dport),
		Pid:      uint32(conn.pid),
		Netns:    uint32(conn.netns),
		Meta:     uint32(conn.meta),
		NATDaddr: dnatAddr, //nolint:gosec
		NATDport: dnatPort,
	}
}

func (conn ConnectionInfo) String() string {
	return fmt.Sprintf("%s:%d -> %s:%d, pid:%d, tcp:%t", U32BEToIP(conn.Saddr, true), conn.Sport, U32BEToIP(conn.Daddr, true), conn.Dport, conn.Pid, ConnProtocolIsTCP(conn.Meta))
}

type ConnectionStats struct {
	SentBytes   uint64
	RecvBytes   uint64
	SentPackets uint64
	RecvPackets uint64
	Timestamp   uint64
	Flags       uint32

	NATDaddr [4]uint32
	NATDport uint32

	Direction uint8
}

type ConnectionTCPStats struct {
	StateTransitions uint16
	Retransmits      int32
	Rtt              uint32
	RttVar           uint32
}

type ConncetionClosedInfo struct {
	Info     ConnectionInfo
	Stats    ConnectionStats
	TCPStats ConnectionTCPStats
}

const (
	ConnL3Mask uint32 = 0xFF
	ConnL3IPv4 uint32 = 0x00 // 0x00
	ConnL3IPv6 uint32 = 0x01 // 0x01

	ConnL4Mask uint32 = 0xFF00
	ConnL4TCP  uint32 = 0x0000 // 0x00 << 8
	ConnL4UDP  uint32 = 0x0100 // 0x01 << 8
)

const (
	ConnDirectionAuto = iota
	ConnDirectionIncoming
	ConnDirectionOutgoing
	ConnDirectionUnknown
)

//nolint:stylecheck
const (
	TCP_ESTABLISHED = iota + 1
	TCP_SYN_SENT
	TCP_SYN_RECV
	TCP_FIN_WAIT1
	TCP_FIN_WAIT2
	TCP_TIME_WAIT
	TCP_CLOSE
	TCP_CLOSE_WAIT
	TCP_LAST_ACK
	TCP_LISTEN
	TCP_CLOSING
	TCP_NEW_SYN_RECV
	TCP_MAX_STATES
)

func ConnProtocolIsTCP(meta uint32) bool {
	return (meta & ConnL4Mask) == ConnL4TCP
}

func ConnAddrIsIPv4(meta uint32) bool {
	return (meta & ConnL3Mask) == ConnL3IPv4
}

func ConnDirection2Str(direction uint8) string {
	switch direction {
	case ConnDirectionIncoming:
		return DirectionIncoming
	case ConnDirectionOutgoing:
		return DirectionOutgoing
	default:
		// kprobe__tcp_close in netflow.c does not judge the in/out traffic,
		// If the connection only triggers this function, the direction value is 0.
		return DirectionOutgoing
	}
}

func ConnIPv4Type(addr uint32) string {
	ip := U32IP4(addr)

	if ip.IsPrivate() {
		return "private"
	}

	if ip.IsLoopback() {
		return "loopback"
	}

	if ip.IsMulticast() {
		return "multicast"
	}
	return "other"
}

func ConnIPv6Type(addr [4]uint32) string {
	ip := U32IP6(addr)

	if ip.IsPrivate() {
		return "private"
	}

	if ip.IsLoopback() {
		return "loopback"
	}

	if ip.IsMulticast() {
		return "multicast"
	}

	return "other"
}

type ConnFullStats struct {
	Stats    ConnectionStats
	TCPStats ConnectionTCPStats

	TotalClosed      int64
	TotalEstablished int64
	// RttCount         int64
	// RttAvgNs         float64
}

type ConnStatsRecord struct {
	sync.Mutex
	closedConns     map[ConnectionInfo]ConnFullStats
	lastActiveConns map[ConnectionInfo]ConnFullStats

	lastTS time.Time // UTC
}

func newConnStatsRecord() *ConnStatsRecord {
	return &ConnStatsRecord{
		closedConns:     make(map[ConnectionInfo]ConnFullStats),
		lastActiveConns: make(map[ConnectionInfo]ConnFullStats),
		lastTS:          ntp.Now(),
	}
}

func (c *ConnStatsRecord) clearClosedConnsCache() {
	c.closedConns = make(map[ConnectionInfo]ConnFullStats)
}

func (c *ConnStatsRecord) updateClosedUseEvent(closedEvents *ConncetionClosedInfo) {
	if connLastActive, ok := c.lastActiveConns[closedEvents.Info]; ok {
		// Connections that were not closed during the last collection cycle.
		if ConnProtocolIsTCP(closedEvents.Info.Meta) {
			connLastActive.TotalEstablished = 0
			connLastActive.TotalClosed = 1
			connLastActive = StatsTCPOp("-", connLastActive, closedEvents.Stats, closedEvents.TCPStats)
		} else {
			connLastActive = StatsOp("-", connLastActive, closedEvents.Stats)
		}
		c.deleteLastActive(closedEvents.Info)
		// Save to closedConns.

		c.closedConns[closedEvents.Info] = connLastActive
	} else if connClosed, ok := c.closedConns[closedEvents.Info]; ok {
		// A connection that has been closed, has been recorded.

		if ConnProtocolIsTCP(closedEvents.Info.Meta) {
			connClosed.TotalClosed += 1
			connClosed.TotalEstablished += 1
			connClosed = StatsTCPOp("+", connClosed, closedEvents.Stats, closedEvents.TCPStats)
		} else {
			connClosed = StatsOp("+", connClosed, closedEvents.Stats)
		}
		c.closedConns[closedEvents.Info] = connClosed
	} else {
		// Connections established and closed during the current cycle, the first record.
		connF := ConnFullStats{
			Stats:    closedEvents.Stats,
			TCPStats: closedEvents.TCPStats,
		}
		if ConnProtocolIsTCP((closedEvents.Info.Meta)) {
			connF.TotalClosed = 1
			connF.TotalEstablished = 1
		}
		c.closedConns[closedEvents.Info] = connF
	}
}

func (c *ConnStatsRecord) updateLastActive(activeConnInfo ConnectionInfo, activeConnFullStats ConnFullStats) {
	activeConnFullStats.TotalEstablished = 0
	activeConnFullStats.TotalClosed = 0
	c.lastActiveConns[activeConnInfo] = activeConnFullStats
}

func (c *ConnStatsRecord) readLastActive(conninfo ConnectionInfo) (ConnFullStats, bool) {
	v, ok := c.lastActiveConns[conninfo]
	return v, ok
}

func (c *ConnStatsRecord) deleteLastActive(conninfo ConnectionInfo) {
	delete(c.lastActiveConns, conninfo)
}

// Return the merged result (closed and unclosed in the previous cycle);
// Calling this method will update/delete the elements of Map: lastActiveConns, closedConns in record.
func (c *ConnStatsRecord) mergeWithClosedLastActive(connInfo ConnectionInfo, connFullStats ConnFullStats) ConnFullStats {
	if v, ok := c.closedConns[connInfo]; ok {
		// closed
		if ConnProtocolIsTCP(connInfo.Meta) {
			v = StatsTCPOp("+", v, connFullStats.Stats, connFullStats.TCPStats)
			v.TotalEstablished += 1
		} else {
			v = StatsOp("+", v, connFullStats.Stats)
		}

		c.updateLastActive(connInfo, connFullStats) // Copy current active to lastActiveConns.

		// Remove the information that the current connection is closed after the connection is established in the current cycle.
		delete(c.closedConns, connInfo)

		return v
	} else if v, ok := c.readLastActive(connInfo); ok {
		// last active
		if ConnProtocolIsTCP(connInfo.Meta) {
			v = StatsTCPOp("-", v, connFullStats.Stats, connFullStats.TCPStats)
			v.TotalEstablished = 0
		} else {
			v = StatsOp("-", v, connFullStats.Stats)
		}
		c.updateLastActive(connInfo, connFullStats)
		return v
	} else {
		if ConnProtocolIsTCP(connInfo.Meta) {
			// Connections established before ebpf-net started cannot record TCP_ESTABLISHED.
			// Do not judge whether the connection is established according to TCP_ESTABLISHED,
			// connections that exist in bpfmap_conn_stats are considered unclosed connections
			// if connFullStats.TcpStats.State_transitions>>TCP_ESTABLISHED == 1 .
			connFullStats.TotalEstablished = 1
		}
		c.updateLastActive(connInfo, connFullStats)
		return connFullStats
	}
}

// StatsOp fullConn = connStats op("+", "-", ...) fullConn.
func StatsOp(op string, fullConn ConnFullStats, connStats ConnectionStats) ConnFullStats {
	switch op {
	case "+":
		fullConn.Stats.SentBytes += connStats.SentBytes
		fullConn.Stats.RecvBytes += connStats.RecvBytes
		fullConn.Stats.SentPackets += connStats.SentPackets
		fullConn.Stats.RecvPackets += connStats.RecvPackets
	case "-":
		if connStats.SentBytes >= fullConn.Stats.SentBytes && connStats.RecvBytes >= fullConn.Stats.RecvBytes &&
			connStats.SentPackets >= fullConn.Stats.SentPackets && connStats.RecvPackets >= fullConn.Stats.RecvPackets {
			fullConn.Stats.SentBytes = connStats.SentBytes - fullConn.Stats.SentBytes
			fullConn.Stats.RecvBytes = connStats.RecvBytes - fullConn.Stats.RecvBytes
			fullConn.Stats.SentPackets = connStats.SentPackets - fullConn.Stats.SentPackets
			fullConn.Stats.RecvPackets = connStats.RecvPackets - fullConn.Stats.RecvPackets
		} else {
			fullConn.Stats.SentBytes = 0
			fullConn.Stats.RecvBytes = 0
			fullConn.Stats.SentPackets = 0
			fullConn.Stats.RecvPackets = 0
		}
	}
	fullConn.Stats.Direction = connStats.Direction
	fullConn.Stats.Flags = connStats.Flags
	fullConn.Stats.Timestamp = connStats.Timestamp
	return fullConn
}

// StatsTCPOp op: operator; fullConn: a saved connection statistics; connStat: new connection statistics; tcpstats: TCP statistics.
func StatsTCPOp(op string, fullConn ConnFullStats, connStats ConnectionStats,
	tcpstats ConnectionTCPStats) ConnFullStats {
	fullConn = StatsOp(op, fullConn, connStats)
	switch op {
	case "+":
		fullConn.TCPStats.Retransmits += tcpstats.Retransmits
	case "-":
		if tcpstats.Retransmits >= fullConn.TCPStats.Retransmits {
			fullConn.TCPStats.Retransmits = tcpstats.Retransmits - fullConn.TCPStats.Retransmits
		} else {
			fullConn.TCPStats.Retransmits = 0
		}
	}
	fullConn.TCPStats.Rtt = tcpstats.Rtt
	fullConn.TCPStats.RttVar = tcpstats.RttVar
	fullConn.TCPStats.StateTransitions = tcpstats.StateTransitions
	return fullConn
}
