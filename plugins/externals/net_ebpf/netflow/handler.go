// +build linux

package netflow

import (
	"time"
	"unsafe"

	"github.com/DataDog/ebpf"
	"github.com/DataDog/ebpf/manager"
	"golang.org/x/net/context"
)

type ConnResult struct {
	result map[ConnectionInfo]ConnFullStats
	tags   map[string]string
	ts     time.Time
}

var resultCh = make(chan *ConnResult)

var connStatsRecord = ConnStatsRecord{}

const (
	inputName = "netflow"
)

func closedEventHandler(cpu int, data []byte, perfmap *manager.PerfMap, manager *manager.Manager) {
	eventC := (*ConncetionClosedInfoC)(unsafe.Pointer(&data[0]))
	event := ConncetionClosedInfo{
		Info: ConnectionInfo{
			Saddr: (*(*[4]uint32)(unsafe.Pointer(&eventC.conn_info.saddr))),
			Daddr: (*(*[4]uint32)(unsafe.Pointer(&eventC.conn_info.daddr))),
			Sport: uint32(eventC.conn_info.sport),
			Dport: uint32(eventC.conn_info.dport),
			Pid:   uint32(eventC.conn_info.pid),
			Netns: uint32(eventC.conn_info.netns),
			Meta:  uint32(eventC.conn_info.meta),
		},
		Stats: ConnectionStats{
			Sent_bytes: uint64(eventC.conn_stats.sent_bytes),
			Recv_bytes: uint64(eventC.conn_stats.recv_bytes),
			Flags:      uint32(eventC.conn_stats.flags),
			Direction:  uint8(eventC.conn_stats.direction),
			Timestamp:  uint64(eventC.conn_stats.timestamp),
		},
		Tcp_stats: ConnectionTcpStats{
			State_transitions: uint16(eventC.conn_tcp_stats.state_transitions),
			Retransmits:       int32(eventC.conn_tcp_stats.retransmits),
			Rtt:               uint32(eventC.conn_tcp_stats.rtt),
			Rtt_var:           uint32(eventC.conn_tcp_stats.rtt_var),
		},
	}
	connStatsRecord.Lock()
	defer connStatsRecord.Unlock()
	connStatsRecord.updateClosedUseEvent(event)
}

// 在扫描 connStatMap 时锁定资源 connStatsRecord
// duration 介于 10s ～ 30min, 若非，默认设为 30s.
func ConnCollectHanllder(ctx context.Context, connStatsMap *ebpf.Map, tcpStatsMap *ebpf.Map,
	interval time.Duration, gTags map[string]string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			connStatsRecord.Lock()

			connResult := ConnResult{
				result: make(map[ConnectionInfo]ConnFullStats),
				tags:   gTags,
				ts:     time.Now(),
			}

			var connInfoC ConnectionInfoC

			var connStatsC ConnectionStatsC

			var tcpStatsC ConnectionTcpStatsC

			iter := connStatsMap.IterateFrom(connInfoC)

			// 收集未关闭的连接信息, 并与记录的关闭连接和上一采集周期未关闭的连接进行合并
			for iter.Next(unsafe.Pointer(&connInfoC), unsafe.Pointer(&connStatsC)) {
				connInfo := ConnectionInfo{
					Saddr: (*(*[4]uint32)(unsafe.Pointer(&connInfoC.saddr))),
					Daddr: (*(*[4]uint32)(unsafe.Pointer(&connInfoC.daddr))),
					Sport: uint32(connInfoC.sport),
					Dport: uint32(connInfoC.dport),
					Pid:   uint32(connInfoC.pid),
					Netns: uint32(connInfoC.netns),
					Meta:  uint32(connInfoC.meta),
				}

				connStats := ConnectionStats{
					Sent_bytes:   uint64(connStatsC.sent_bytes),
					Recv_bytes:   uint64(connStatsC.recv_bytes),
					Sent_packets: uint64(connStatsC.sent_packets),
					Recv_packets: uint64(connStatsC.recv_packets),
					Flags:        uint32(connStatsC.flags),
					Direction:    uint8(connStatsC.direction),
					Timestamp:    uint64(connStatsC.timestamp),
				}
				connFullStats := ConnFullStats{
					Stats:            connStats,
					TotalClosed:      0,
					TotalEstablished: 0,
				}
				if connProtocolIsTCP(connInfo.Meta) {
					pid := connInfoC.pid
					connInfoC.pid = _Ctype_uint(0)
					if err := tcpStatsMap.Lookup(unsafe.Pointer(&connInfoC), unsafe.Pointer(&tcpStatsC)); err == nil {
						connFullStats.TcpStats = ConnectionTcpStats{
							State_transitions: uint16(tcpStatsC.state_transitions),
							Retransmits:       int32(tcpStatsC.retransmits),
							Rtt:               uint32(tcpStatsC.rtt),
							Rtt_var:           uint32(tcpStatsC.rtt_var),
						}
					}
					connInfoC.pid = pid
				}

				connFullStats = connStatsRecord.mergeWithClosedLastActive(connInfo, connFullStats)
				if ConnFilter(connInfo, connFullStats) {
					connResult.result[connInfo] = connFullStats
				}
			}
			// 收集当前周期处于关闭状态的连接
			for k, v := range connStatsRecord.closedConns {
				if ConnFilter(k, v) {
					connResult.result[k] = v
				}
			}
			connStatsRecord.clearClosedConnsCache()
			connStatsRecord.Unlock()

			resultCh <- &connResult
		case <-ctx.Done():
			return
		}
	}
}

// 接收一个周期内采集的全部连接, 并发送至 DataKit.
func FeedHandler(ctx context.Context, datakitPostURL string) {
	for {
		select {
		case result := <-resultCh:
			connMerge(result)
			collectCache := convertConn2Measurement(result, inputName)
			FeedMeasurement(collectCache, datakitPostURL)
		case <-ctx.Done():
			return
		}
	}
}

func init() {
	connStatsRecord.initCache()
}
