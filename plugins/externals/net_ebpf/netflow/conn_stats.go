// +build linux

package netflow

import (
	"sync"
	"time"
)

const (
	CONN_L3_MASK uint32 = 0xFF
	CONN_L3_IPv4 uint32 = 0x00 // 0x00
	CONN_L3_IPv6 uint32 = 0x01 // 0x01

	CONN_L4_MASK uint32 = 0xFF00
	CONN_L4_TCP  uint32 = 0x0000 // 0x00 << 8
	CONN_L4_UDP  uint32 = 0x0100 // 0x01 << 8
)

const (
	CONN_DIRECTION_AUTO = iota
	CONN_DIRECTION_INCOMING
	CONN_DIRECTION_OUTGOING
	CONN_DIRECTION_UNKNOWN
)

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

func connProtocolIsTCP(meta uint32) bool {
	return (meta & CONN_L4_MASK) == CONN_L4_TCP
}

func connAddrIsIPv4(meta uint32) bool {
	return (meta & CONN_L3_MASK) == CONN_L3_IPv4
}

func connDirection2Str(direction uint8) string {
	switch direction {
	case CONN_DIRECTION_INCOMING:
		return "incoming"
	case CONN_DIRECTION_OUTGOING:
		return "outgoing"
	default:
		// netflow.c 中的 kprobe__tcp_close 不判断进出口流量，
		// 若该连接只触发此函数，direction 值为 0.
		return "outgoing"
	}
}

func connIPv4Type(addr uint32) string {
	ip := U32BEToIPv4Array(addr)

	if (ip[0] == 10) ||
		((ip[0] == 172) && (ip[1] >= 16) && (ip[1] <= 31)) ||
		((ip[0] == 192) && (ip[1] == 168)) {
		// 10.0.0.0/8; 172.16.0.0/12; 192.168.0.0/16
		return "private"
	}
	if ip[0] > 223 && ip[0] <= 239 {
		return "multicast"
	}
	if ip[0] == 127 {
		// 127.0.0.0/8
		return "loopback"
	}

	return "other"
}

func connIPv6Type(addr [4]uint32) string {
	ip := U32BEToIPv6Array(addr)

	if (ip[0]|ip[1]|ip[2]|ip[3]|ip[4]|ip[5]|ip[6]) == 0 &&
		ip[7] == 1 { // ::1/128
		return "loopback"
	}
	if ip[0]&0xfe00 == 0xfc00 { // fc00::/7
		return "private"
	}
	if ip[0]&0xff00 == 0xff00 { // ff00::/8
		return "multicast"
	}

	return "other"
}

type ConnFullStats struct {
	Stats    ConnectionStats
	TcpStats ConnectionTcpStats

	TotalClosed      int64
	TotalEstablished int64
	// RttCount         int64
	// RttAvgNs         float64
}

type ConnStatsRecord struct {
	sync.Mutex
	closedConns     map[ConnectionInfo]ConnFullStats
	lastActiveConns map[ConnectionInfo]ConnFullStats

	lastTs time.Time // UTC
}

func (c *ConnStatsRecord) initCache() {
	c.closedConns = make(map[ConnectionInfo]ConnFullStats)
	c.lastActiveConns = make(map[ConnectionInfo]ConnFullStats)
	c.lastTs = time.Now()
}

func newConnStatsRecord() *ConnStatsRecord {
	return &ConnStatsRecord{
		closedConns:     make(map[ConnectionInfo]ConnFullStats),
		lastActiveConns: make(map[ConnectionInfo]ConnFullStats),
		lastTs:          time.Now(),
	}
}

func (c *ConnStatsRecord) clearClosedConnsCache() {
	c.closedConns = make(map[ConnectionInfo]ConnFullStats)
}

func (c *ConnStatsRecord) updateClosedUseEvent(closedEvents ConncetionClosedInfo) {
	if connLastActive, ok := c.lastActiveConns[closedEvents.Info]; ok {
		// 上一采集周期内未关闭的连接
		if connProtocolIsTCP(closedEvents.Info.Meta) {
			connLastActive.TotalEstablished = 0
			connLastActive.TotalClosed = 1
			connLastActive = StatsTCPOp("-", connLastActive, closedEvents.Stats, closedEvents.Tcp_stats)
		} else {
			connLastActive = StatsOp("-", connLastActive, closedEvents.Stats)
		}
		c.deleteLastActive(closedEvents.Info)
		// save to closedConns

		c.closedConns[closedEvents.Info] = connLastActive
	} else if connClosed, ok := c.closedConns[closedEvents.Info]; ok {
		// 已经关闭的连接，已被记录的

		if connProtocolIsTCP(closedEvents.Info.Meta) {
			connClosed.TotalClosed += 1
			connClosed.TotalEstablished += 1
			connClosed = StatsTCPOp("+", connClosed, closedEvents.Stats, closedEvents.Tcp_stats)
		} else {
			connClosed = StatsOp("+", connClosed, closedEvents.Stats)
		}
		c.closedConns[closedEvents.Info] = connClosed
	} else {
		// 在当前周期内建立连接，并关闭的连接, 首次记录
		connF := ConnFullStats{
			Stats:    closedEvents.Stats,
			TcpStats: closedEvents.Tcp_stats,
		}
		if connProtocolIsTCP((closedEvents.Info.Meta)) {
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

// 返回合并结果(与已关闭的和上一周期未关闭的);
// 调用此方法将更新/删除 record 中的 Map: lastActiveConns, closedConns 的元素.
func (c *ConnStatsRecord) mergeWithClosedLastActive(connInfo ConnectionInfo, connFullStats ConnFullStats) ConnFullStats {
	if v, ok := c.closedConns[connInfo]; ok {
		// closed
		if connProtocolIsTCP(connInfo.Meta) {
			v = StatsTCPOp("+", v, connFullStats.Stats, connFullStats.TcpStats)
			v.TotalEstablished += 1
		} else {
			v = StatsOp("+", v, connFullStats.Stats)
		}
		c.updateLastActive(connInfo, connFullStats) // 将当前 active 拷贝至 lastActiveConns 中
		delete(c.closedConns, connInfo)             // 移除当前周期内当前连接连接建立后关闭的信息
		return v
	} else if v, ok := c.readLastActive(connInfo); ok {
		// last active
		if connProtocolIsTCP(connInfo.Meta) {
			v = StatsTCPOp("-", v, connFullStats.Stats, connFullStats.TcpStats)
			v.TotalEstablished = 0
		} else {
			v = StatsOp("-", v, connFullStats.Stats)
		}
		c.updateLastActive(connInfo, connFullStats)
		return v
	} else {
		if connProtocolIsTCP(connInfo.Meta) {
			// 在 net_ebpf 启动前建立的连接无法记录 TCP_ESTABLISHED,
			// 不依据 TCP_ESTABLISHED 判断连接是否建立，
			// 存在于 bpfmap_conn_stats 的连接视为未关闭的连接
			// if connFullStats.TcpStats.State_transitions>>TCP_ESTABLISHED == 1 .
			connFullStats.TotalEstablished = 1
		}
		c.updateLastActive(connInfo, connFullStats)
		return connFullStats
	}
}

// fullConn = connStats op("+", "-", ...) fullConn;.
func StatsOp(op string, fullConn ConnFullStats, connStats ConnectionStats) ConnFullStats {
	switch op {
	case "+":
		fullConn.Stats.Sent_bytes += connStats.Sent_bytes
		fullConn.Stats.Recv_bytes += connStats.Recv_bytes
		fullConn.Stats.Sent_packets += connStats.Sent_packets
		fullConn.Stats.Recv_packets += connStats.Recv_packets
	case "-":
		if connStats.Sent_bytes >= fullConn.Stats.Sent_bytes && connStats.Recv_bytes >= fullConn.Stats.Recv_bytes && connStats.Sent_packets >= fullConn.Stats.Sent_packets && connStats.Recv_packets >= fullConn.Stats.Recv_packets {
			fullConn.Stats.Sent_bytes = connStats.Sent_bytes - fullConn.Stats.Sent_bytes
			fullConn.Stats.Recv_bytes = connStats.Recv_bytes - fullConn.Stats.Recv_bytes
			fullConn.Stats.Sent_packets = connStats.Sent_packets - fullConn.Stats.Sent_packets
			fullConn.Stats.Recv_packets = connStats.Recv_packets - fullConn.Stats.Recv_packets
		} else {
			fullConn.Stats.Sent_bytes = 0
			fullConn.Stats.Recv_bytes = 0
			fullConn.Stats.Sent_packets = 0
			fullConn.Stats.Recv_packets = 0
		}
	}
	fullConn.Stats.Direction = connStats.Direction
	fullConn.Stats.Flags = connStats.Flags
	fullConn.Stats.Timestamp = connStats.Timestamp
	return fullConn
}

// op: 操作符; fullConn: 被保存的一个连接统计信息; connStat: 新的连接统计信息; tcpstats: TC统计信息
func StatsTCPOp(op string, fullConn ConnFullStats, connStats ConnectionStats,
	tcpstats ConnectionTcpStats) ConnFullStats {
	fullConn = StatsOp(op, fullConn, connStats)
	switch op {
	case "+":
		fullConn.TcpStats.Retransmits += tcpstats.Retransmits
	case "-":
		if tcpstats.Retransmits >= fullConn.TcpStats.Retransmits {
			fullConn.TcpStats.Retransmits = tcpstats.Retransmits - fullConn.TcpStats.Retransmits
		} else {
			fullConn.TcpStats.Retransmits = 0
		}
	}
	fullConn.TcpStats.Rtt = tcpstats.Rtt
	fullConn.TcpStats.Rtt_var = tcpstats.Rtt_var
	fullConn.TcpStats.State_transitions = tcpstats.State_transitions
	return fullConn
}
