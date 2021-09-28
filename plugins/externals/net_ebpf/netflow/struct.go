// +build linux

package netflow

// #include "../c/netflow/conn_stats.h"
import "C"

type ConnectionInfoC C.struct_connection_info

type ConnectionStatsC C.struct_connection_stats

type ConnectionTcpStatsC C.struct_connection_tcp_stats

type ConncetionClosedInfoC C.struct_connection_closed_info

type ConnectionInfo struct {
	Saddr [4]uint32
	Daddr [4]uint32
	Sport uint32
	Dport uint32
	Pid   uint32
	Netns uint32
	Meta  uint32
}

type ConnectionStats struct {
	Sent_bytes   uint64
	Recv_bytes   uint64
	Sent_packets uint64
	Recv_packets uint64
	Flags        uint32
	Direction    uint8
	Timestamp    uint64
}

type ConnectionTcpStats struct {
	State_transitions uint16
	Retransmits       int32
	Rtt               uint32
	Rtt_var           uint32
}

type ConncetionClosedInfo struct {
	Info      ConnectionInfo
	Stats     ConnectionStats
	Tcp_stats ConnectionTcpStats
}
