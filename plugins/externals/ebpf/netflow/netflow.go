//go:build (linux && ignore) || ebpf
// +build linux,ignore ebpf

package netflow

import (
	"time"
	"unsafe"

	"github.com/DataDog/ebpf"
	"github.com/DataDog/ebpf/manager"
	"github.com/shirou/gopsutil/host"
	dkfeed "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/ebpf/feed"
	"golang.org/x/net/context"
)

type ConnResult struct {
	result map[ConnectionInfo]ConnFullStats
	tags   map[string]string
	ts     time.Time
}

const connExpirationInterval = 8 * 3600 // 8 * 3600s

const (
	srcNameM = "netflow"
)

type NetFlowTracer struct {
	connStatsRecord *ConnStatsRecord
	resultCh        chan *ConnResult
	closedEventCh   chan *ConncetionClosedInfo
}

func NewNetFlowTracer() *NetFlowTracer {
	return &NetFlowTracer{
		connStatsRecord: newConnStatsRecord(),
		resultCh:        make(chan *ConnResult, 64),
		closedEventCh:   make(chan *ConncetionClosedInfo, 64),
	}
}

func (tracer *NetFlowTracer) Run(ctx context.Context, bpfManger *manager.Manager,
	datakitPostURL string, gTags map[string]string, interval time.Duration) error {
	connStatsMap, found, err := bpfManger.GetMap("bpfmap_conn_stats")
	if err != nil || !found {
		return err
	}

	tcpStatsMap, found, err := bpfManger.GetMap("bpfmap_conn_tcp_stats")
	if err != nil || !found {
		return err
	}

	go tracer.feedHandler(ctx, datakitPostURL)
	go tracer.connCollectHanllder(ctx, connStatsMap, tcpStatsMap, interval, gTags)
	return nil
}

func (tracer *NetFlowTracer) ClosedEventHandler(cpu int, data []byte,
	perfmap *manager.PerfMap, manager *manager.Manager) {
	eventC := (*ConncetionClosedInfoC)(unsafe.Pointer(&data[0])) //nolint:gosec
	event := ConncetionClosedInfo{
		Info: ConnectionInfo{
			Saddr: (*(*[4]uint32)(unsafe.Pointer(&eventC.conn_info.saddr))), //nolint:gosec
			Daddr: (*(*[4]uint32)(unsafe.Pointer(&eventC.conn_info.daddr))), //nolint:gosec
			Sport: uint32(eventC.conn_info.sport),
			Dport: uint32(eventC.conn_info.dport),
			Pid:   uint32(eventC.conn_info.pid),
			Netns: uint32(eventC.conn_info.netns),
			Meta:  uint32(eventC.conn_info.meta),
		},
		Stats: ConnectionStats{
			SentBytes: uint64(eventC.conn_stats.sent_bytes),
			RecvBytes: uint64(eventC.conn_stats.recv_bytes),
			Flags:     uint32(eventC.conn_stats.flags),
			Direction: uint8(eventC.conn_stats.direction),
			Timestamp: uint64(eventC.conn_stats.timestamp),
		},
		TCPStats: ConnectionTCPStats{
			StateTransitions: uint16(eventC.conn_tcp_stats.state_transitions),
			Retransmits:      int32(eventC.conn_tcp_stats.retransmits),
			Rtt:              uint32(eventC.conn_tcp_stats.rtt),
			RttVar:           uint32(eventC.conn_tcp_stats.rtt_var),
		},
	}
	if IPPortFilterIn(&event.Info) {
		tracer.closedEventCh <- &event
	}
}

func (tracer *NetFlowTracer) bpfMapCleanup(cl []ConnectionInfo, connStatsMap *ebpf.Map) {
	for _, v := range cl {
		c := ConnectionInfoC{
			saddr: (*(*[4]_Ctype_uint)(unsafe.Pointer(&v.Saddr))), //nolint:gosec
			daddr: (*(*[4]_Ctype_uint)(unsafe.Pointer(&v.Daddr))), //nolint:gosec
			sport: _Ctype_ushort(v.Sport),
			dport: _Ctype_ushort(v.Dport),
			pid:   _Ctype_uint(v.Pid),
			netns: _Ctype_uint(v.Netns),
			meta:  _Ctype_uint(v.Meta),
		}
		err := connStatsMap.Delete(unsafe.Pointer(&c)) //nolint:gosec
		if err != nil {
			l.Error(err)
		}
	}
}

// 在扫描 connStatMap 时锁定资源 connStatsRecord.
func (tracer *NetFlowTracer) connCollectHanllder(ctx context.Context, connStatsMap *ebpf.Map, tcpStatsMap *ebpf.Map,
	interval time.Duration, gTags map[string]string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case event := <-tracer.closedEventCh:
			tracer.connStatsRecord.updateClosedUseEvent(event)
		case <-ticker.C:
			connResult := ConnResult{
				result: make(map[ConnectionInfo]ConnFullStats),
				tags:   gTags,
				ts:     time.Now(),
			}

			var connInfoC ConnectionInfoC

			var connStatsC ConnectionStatsC

			var tcpStatsC ConnectionTCPStatsC

			iter := connStatsMap.IterateFrom(connInfoC)

			connsNeedCleanup := []ConnectionInfo{}
			uptime, err := host.Uptime()
			if err != nil {
				l.Error(err)
			}

			// 收集未关闭的连接信息, 并与记录的关闭连接和上一采集周期未关闭的连接进行合并
			for iter.Next(unsafe.Pointer(&connInfoC), unsafe.Pointer(&connStatsC)) { //nolint:gosec
				connInfo := ConnectionInfo{
					Saddr: (*(*[4]uint32)(unsafe.Pointer(&connInfoC.saddr))), //nolint:gosec
					Daddr: (*(*[4]uint32)(unsafe.Pointer(&connInfoC.daddr))), //nolint:gosec
					Sport: uint32(connInfoC.sport),
					Dport: uint32(connInfoC.dport),
					Pid:   uint32(connInfoC.pid),
					Netns: uint32(connInfoC.netns),
					Meta:  uint32(connInfoC.meta),
				}

				if !IPPortFilterIn(&connInfo) {
					continue
				}

				connStats := ConnectionStats{
					SentBytes:   uint64(connStatsC.sent_bytes),
					RecvBytes:   uint64(connStatsC.recv_bytes),
					SentPackets: uint64(connStatsC.sent_packets),
					RecvPackets: uint64(connStatsC.recv_packets),
					Flags:       uint32(connStatsC.flags),
					Direction:   uint8(connStatsC.direction),
					Timestamp:   uint64(connStatsC.timestamp),
				}
				connFullStats := ConnFullStats{
					Stats:            connStats,
					TotalClosed:      0,
					TotalEstablished: 0,
				}
				if connProtocolIsTCP(connInfo.Meta) {
					pid := connInfoC.pid
					connInfoC.pid = _Ctype_uint(0)
					if err := tcpStatsMap.Lookup(
						unsafe.Pointer(&connInfoC),               //nolint:gosec
						unsafe.Pointer(&tcpStatsC)); err == nil { //nolint:gosec
						connFullStats.TCPStats = ConnectionTCPStats{
							StateTransitions: uint16(tcpStatsC.state_transitions),
							Retransmits:      int32(tcpStatsC.retransmits),
							Rtt:              uint32(tcpStatsC.rtt),
							RttVar:           uint32(tcpStatsC.rtt_var),
						}
					}
					connInfoC.pid = pid
				}
				connFullStats = tracer.connStatsRecord.mergeWithClosedLastActive(connInfo, connFullStats)
				if int(uptime)-int(connFullStats.Stats.Timestamp/1000000000) > connExpirationInterval {
					if connFullStats.TotalClosed == 0 && connFullStats.TotalEstablished == 0 &&
						connFullStats.Stats.RecvBytes == 0 && connFullStats.Stats.SentBytes == 0 {
						connsNeedCleanup = append(connsNeedCleanup, connInfo)
						continue
					}
				}
				connResult.result[connInfo] = connFullStats
			}
			if len(connsNeedCleanup) > 0 {
				for _, conn := range connsNeedCleanup {
					tracer.connStatsRecord.deleteLastActive(conn)
				}
				tracer.bpfMapCleanup(connsNeedCleanup, connStatsMap)
			}
			// 收集当前周期处于关闭状态的连接
			for k, v := range tracer.connStatsRecord.closedConns {
				connResult.result[k] = v
			}
			tracer.connStatsRecord.clearClosedConnsCache()
			select {
			case tracer.resultCh <- &connResult:
			default:
				l.Error("channel is full, drop data")
			}

		case <-ctx.Done():
			return
		}
	}
}

// 接收一个周期内采集的全部连接, 并发送至 DataKit.
func (tracer *NetFlowTracer) feedHandler(ctx context.Context, datakitPostURL string) {
	for {
		select {
		case result := <-tracer.resultCh:
			MergeConns(result)
			collectCache := ConvertConn2Measurement(result, srcNameM)
			if len(collectCache) == 0 {
				l.Warn("netflow: no data")
			} else if err := dkfeed.FeedMeasurement(collectCache, datakitPostURL); err != nil {
				l.Error(err)
			}
		case <-ctx.Done():
			return
		}
	}
}
