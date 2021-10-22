// +build linux

package netflow

import (
	"time"
	"unsafe"

	"github.com/DataDog/ebpf"
	"github.com/DataDog/ebpf/manager"
	"github.com/shirou/gopsutil/host"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/net_ebpf/dns"
	"golang.org/x/net/context"
)

type ConnResult struct {
	result map[ConnectionInfo]ConnFullStats
	tags   map[string]string
	ts     time.Time
}

const connExpirationInterval = 8 * 3600 // 8 * 3600s

var resultCh = make(chan *ConnResult)

var connStatsRecord = ConnStatsRecord{}

const (
	inputName = "netflow"
)

func Run(ctx context.Context, bpfManger *manager.Manager, datakitPostURL string, gTags map[string]string, interval time.Duration) error {
	connStatsMap, found, err := bpfManger.GetMap("bpfmap_conn_stats")
	if err != nil || !found {
		return err
	}

	tcpStatsMap, found, err := bpfManger.GetMap("bpfmap_conn_tcp_stats")
	if err != nil || !found {
		return err
	}

	go feedHandler(ctx, datakitPostURL)
	go connCollectHanllder(ctx, connStatsMap, tcpStatsMap, interval, gTags)
	go bpfMapCleanupHandler(ctx, connStatsMap, tcpStatsMap)
	if tp, err := dns.NewTPacketDNS(); err != nil {
		return err
	} else {
		go dnsRecord.Gather(ctx, tp)
	}
	return nil
}

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

// 在扫描 connStatMap 时锁定资源 connStatsRecord.
func connCollectHanllder(ctx context.Context, connStatsMap *ebpf.Map, tcpStatsMap *ebpf.Map, interval time.Duration, gTags map[string]string) {
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

			connsNeedCleanup := []ConnectionInfo{}
			uptime, err := host.Uptime()
			if err != nil {
				l.Error(err)
			}

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
				if int(uptime)-int(connFullStats.Stats.Timestamp/1000000000) > connExpirationInterval {
					if connFullStats.TotalClosed == 0 && connFullStats.TotalEstablished == 0 &&
						connFullStats.Stats.Recv_bytes == 0 && connFullStats.Stats.Sent_bytes == 0 {
						connsNeedCleanup = append(connsNeedCleanup, connInfo)
						continue
					}
				}
				connResult.result[connInfo] = connFullStats
			}
			if len(connsNeedCleanup) > 0 {

				for _, conn := range connsNeedCleanup {
					connStatsRecord.deleteLastActive(conn)
				}
				cleanupCh <- &connsNeedCleanup
			}
			// 收集当前周期处于关闭状态的连接
			for k, v := range connStatsRecord.closedConns {
				connResult.result[k] = v
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
func feedHandler(ctx context.Context, datakitPostURL string) {
	for {
		select {
		case result := <-resultCh:
			MergeConns(result)
			collectCache := ConvertConn2Measurement(result, inputName)
			if err := FeedMeasurement(collectCache, datakitPostURL); err != nil {
				l.Error(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

var cleanupCh = make(chan *[]ConnectionInfo)

func bpfMapCleanupHandler(ctx context.Context, connStatsMap *ebpf.Map, tcpStatsMap *ebpf.Map) {
	for {
		select {
		case cl := <-cleanupCh:
			for _, v := range *cl {
				c := ConnectionInfoC{
					saddr: (*(*[4]_Ctype_uint)(unsafe.Pointer(&v.Saddr))),
					daddr: (*(*[4]_Ctype_uint)(unsafe.Pointer(&v.Daddr))),
					sport: _Ctype_ushort(v.Sport),
					dport: _Ctype_ushort(v.Dport),
					pid:   _Ctype_uint(v.Pid),
					netns: _Ctype_uint(v.Netns),
					meta:  _Ctype_uint(v.Meta),
				}
				err := connStatsMap.Delete(unsafe.Pointer(&c))
				if err != nil {
					l.Error(err)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func init() {
	connStatsRecord.initCache()
}
