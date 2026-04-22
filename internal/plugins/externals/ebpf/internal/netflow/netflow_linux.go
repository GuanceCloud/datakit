//go:build linux
// +build linux

// Package netflow collects eBPF-network netflow metrics
package netflow

import (
	"errors"
	"time"
	"unsafe"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/cilium/ebpf"
	"github.com/shirou/gopsutil/host"
	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
	dkct "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/conntrack"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/exporter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/procwatch"
	"golang.org/x/net/context"
)

const connExpirationInterval = 6 * 3600 // 6 * 3600s

const closedEventSize = int(unsafe.Sizeof(ConncetionClosedInfoC{})) //nolint:gosec

const (
	srcNameM     = "netflow"
	transportTCP = "tcp"
	transportUDP = "udp"
	inputName    = "ebpf-net/netflow"
	componentID  = "netflow"
)

var enableUDP bool

func SetEnableUDP(on bool) {
	enableUDP = on
}

type NetFlowTracer struct {
	connStatsRecord *ConnStatsRecord
	closedEventCh   chan *ConncetionClosedInfo
	catalog         *procwatch.Catalog
}

func NewNetFlowTracer(catalog *procwatch.Catalog) *NetFlowTracer {
	return &NetFlowTracer{
		connStatsRecord: newConnStatsRecord(),
		closedEventCh:   make(chan *ConncetionClosedInfo, 64),
		catalog:         catalog,
	}
}

func fillConntrackNATFallback(conn *ConnectionInfo) {
	if conn == nil {
		return
	}
	if conn.NATDport != 0 || (conn.NATDaddr[0]|conn.NATDaddr[1]|conn.NATDaddr[2]|conn.NATDaddr[3]) != 0 {
		return
	}

	natAddr, natPort, ok := dkct.LookupDNATTuple(conn.Saddr, conn.Daddr, conn.Sport, conn.Dport, conn.Netns)
	if !ok {
		return
	}

	conn.NATDaddr = natAddr
	conn.NATDport = natPort
}

func (tracer *NetFlowTracer) Run(ctx context.Context, runtime *bpfutil.Runtime,
	gTags map[string]string, interval time.Duration,
) error {
	connStatsMap, err := runtime.LookupMap("bpfmap_conn_stats")
	if err != nil {
		return err
	}

	tcpStatsMap, err := runtime.LookupMap("bpfmap_conn_tcp_stats")
	if err != nil {
		return err
	}

	tcpSegmentsMap, err := runtime.LookupMap("bpfmap_conn_tcp_segments")
	if err != nil {
		return err
	}

	go tracer.connCollectHanllder(ctx, connStatsMap, tcpStatsMap, tcpSegmentsMap,
		interval, gTags)
	return nil
}

func (tracer *NetFlowTracer) ClosedEventHandler(cpu int, data []byte,
	stream *bpfutil.PerfStream, runtime *bpfutil.Runtime,
) {
	if len(data) < closedEventSize {
		l.Debugf("drop short closed event: got %d want >= %d", len(data), closedEventSize)
		exporter.IncBPFEventDrop(componentID, "conn_closed", "short_payload")
		return
	}

	eventC := (*ConncetionClosedInfoC)(unsafe.Pointer(&data[0])) //nolint:gosec
	event := &ConncetionClosedInfo{
		Info: ConnectionInfo{
			Saddr:    (*(*[4]uint32)(unsafe.Pointer(&eventC.conn_info.saddr))), //nolint:gosec
			Daddr:    (*(*[4]uint32)(unsafe.Pointer(&eventC.conn_info.daddr))), //nolint:gosec
			Sport:    uint32(eventC.conn_info.sport),
			Dport:    uint32(eventC.conn_info.dport),
			Pid:      uint32(eventC.conn_info.pid),
			Netns:    uint32(eventC.conn_info.netns),
			Meta:     uint32(eventC.conn_info.meta),
			NATDaddr: (*(*[4]uint32)(unsafe.Pointer(&eventC.conn_stats.nat_daddr))), //nolint:gosec
			NATDport: uint32(eventC.conn_stats.nat_dport),
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
			ConnectAttempts:  uint32(eventC.conn_tcp_stats.connect_attempts),
			ConnectFailures:  uint32(eventC.conn_tcp_stats.connect_failures),
			CloseWait:        uint32(eventC.conn_tcp_stats.close_wait),
			LastAck:          uint32(eventC.conn_tcp_stats.last_ack),
			TimeWait:         uint32(eventC.conn_tcp_stats.time_wait),
		},
	}
	fillConntrackNATFallback(&event.Info)

	if tracer.catalog != nil {
		if v, ok := tracer.catalog.Lookup(int(event.Info.Pid)); ok {
			event.Info.ProcessName = v.Name()
		} else {
			tracer.catalog.ResolveLater(int(event.Info.Pid))
		}
	}

	SrcIPPortRecorder.InsertAndUpdate(event.Info.Saddr)
	if IPPortFilterIn(&event.Info) {
		select {
		case tracer.closedEventCh <- event:
		default:
			l.Debug("drop closed event: queue full")
			exporter.IncBPFEventDrop(componentID, "conn_closed", "queue_full")
		}
	}
}

func (tracer *NetFlowTracer) bpfMapCleanup(cl []ConnectionInfo, connStatsMap, tcpStatsMap, tcpSegmentsMap *ebpf.Map) {
	var okCount float64
	var errCount float64
	var tcpOKCount float64
	var tcpErrCount float64
	var segOKCount float64
	var segErrCount float64

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
		if err != nil && !errors.Is(err, ebpf.ErrKeyNotExist) {
			l.Warn(err)
			errCount++
		} else {
			okCount++
		}

		c.pid = 0
		err = tcpStatsMap.Delete(unsafe.Pointer(&c)) //nolint:gosec
		if err != nil && !errors.Is(err, ebpf.ErrKeyNotExist) {
			tcpErrCount++
		} else {
			tcpOKCount++
		}

		err = tcpSegmentsMap.Delete(unsafe.Pointer(&c)) //nolint:gosec
		if err != nil && !errors.Is(err, ebpf.ErrKeyNotExist) {
			segErrCount++
		} else {
			segOKCount++
		}
	}

	exporter.AddBPFMapCleanup(componentID, "bpfmap_conn_stats", "ok", okCount)
	exporter.AddBPFMapCleanup(componentID, "bpfmap_conn_stats", "error", errCount)
	exporter.AddBPFMapCleanup(componentID, "bpfmap_conn_tcp_stats", "ok", tcpOKCount)
	exporter.AddBPFMapCleanup(componentID, "bpfmap_conn_tcp_stats", "error", tcpErrCount)
	exporter.AddBPFMapCleanup(componentID, "bpfmap_conn_tcp_segments", "ok", segOKCount)
	exporter.AddBPFMapCleanup(componentID, "bpfmap_conn_tcp_segments", "error", segErrCount)
}

func countTCPStatsEntries(tcpStatsMap *ebpf.Map) (uint32, error) {
	var (
		connInfoC ConnectionInfoC
		tcpStatsC ConnectionTCPStatsC
		count     uint32
	)

	iter := tcpStatsMap.Iterate()
	for iter.Next(unsafe.Pointer(&connInfoC), unsafe.Pointer(&tcpStatsC)) { //nolint:gosec
		count++
	}

	if err := iter.Err(); err != nil {
		return 0, err
	}

	return count, nil
}

const KernelTaskCommLen = 16

// Lock resource connStatsRecord while scanning connStatMap.
func (tracer *NetFlowTracer) connCollectHanllder(ctx context.Context, connStatsMap, tcpStatsMap, tcpSegmentsMap *ebpf.Map,
	interval time.Duration, gTags map[string]string,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	agg := FlowAgg{}

	for {
		select {
		case event := <-tracer.closedEventCh:
			tracer.connStatsRecord.updateClosedUseEvent(event)
		case <-ticker.C:
			var connInfoC ConnectionInfoC

			var connStatsC ConnectionStatsC

			var tcpStatsC ConnectionTCPStatsC

			iter := connStatsMap.Iterate()
			var connEntries uint32

			connsNeedCleanup := []ConnectionInfo{}
			uptime, err := host.Uptime()
			if err != nil {
				l.Error(err)
			}

			// Collect unclosed connection information and merge it with recorded closed connections
			// and unclosed connections in the previous collection cycle.
			for iter.Next(unsafe.Pointer(&connInfoC), unsafe.Pointer(&connStatsC)) { //nolint:gosec
				connEntries++
				connInfo := ConnectionInfo{
					Saddr:    (*(*[4]uint32)(unsafe.Pointer(&connInfoC.saddr))), //nolint:gosec
					Daddr:    (*(*[4]uint32)(unsafe.Pointer(&connInfoC.daddr))), //nolint:gosec
					Sport:    uint32(connInfoC.sport),
					Dport:    uint32(connInfoC.dport),
					Pid:      uint32(connInfoC.pid),
					Netns:    uint32(connInfoC.netns),
					Meta:     uint32(connInfoC.meta),
					NATDaddr: (*(*[4]uint32)(unsafe.Pointer(&connStatsC.nat_daddr))), //nolint:gosec
					NATDport: uint32(connStatsC.nat_dport),
				}
				fillConntrackNATFallback(&connInfo)

				SrcIPPortRecorder.InsertAndUpdate(connInfo.Saddr)

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
				if tracer.catalog != nil {
					if v, ok := tracer.catalog.Lookup(int(connInfoC.pid)); ok {
						connInfo.ProcessName = v.Name()
					} else {
						tracer.catalog.ResolveLater(int(connInfoC.pid))
					}
				}

				connFullStats := ConnFullStats{
					Stats:            connStats,
					TotalClosed:      0,
					TotalEstablished: 0,
				}
				if ConnProtocolIsTCP(connInfo.Meta) {
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
							ConnectAttempts:  uint32(tcpStatsC.connect_attempts),
							ConnectFailures:  uint32(tcpStatsC.connect_failures),
							CloseWait:        uint32(tcpStatsC.close_wait),
							LastAck:          uint32(tcpStatsC.last_ack),
							TimeWait:         uint32(tcpStatsC.time_wait),
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
				err := agg.Append(connInfo, connFullStats)
				if err != nil {
					l.Debug(err)
				}
			}
			exporter.ObserveBPFMap(componentID, "bpfmap_conn_stats", connEntries, connStatsMap.MaxEntries())
			if err := iter.Err(); err != nil {
				l.Warnf("iterate bpfmap_conn_stats failed: %s", err)
				exporter.IncBPFMapObserveError(componentID, "bpfmap_conn_stats", "iterate")
			}

			tcpEntries, err := countTCPStatsEntries(tcpStatsMap)
			if err != nil {
				l.Warnf("iterate bpfmap_conn_tcp_stats failed: %s", err)
				exporter.IncBPFMapObserveError(componentID, "bpfmap_conn_tcp_stats", "iterate")
			} else {
				exporter.ObserveBPFMap(componentID, "bpfmap_conn_tcp_stats", tcpEntries, tcpStatsMap.MaxEntries())
			}

			if len(connsNeedCleanup) > 0 {
				for _, conn := range connsNeedCleanup {
					tracer.connStatsRecord.deleteLastActive(conn)
				}
				tracer.bpfMapCleanup(connsNeedCleanup, connStatsMap, tcpStatsMap, tcpSegmentsMap)
			}
			// Collect connections that are closed for the current cycle.
			for k, v := range tracer.connStatsRecord.closedConns {
				err := agg.Append(k, v)
				if err != nil {
					l.Debug(err)
				}
			}
			tracer.connStatsRecord.clearClosedConnsCache()

			exporter.ObserveAggEntries(componentID, agg.Len())
			flushStart := time.Now()
			pts := agg.ToPoint(gTags, k8sNetInfo)
			agg.Clean()
			exporter.ObserveAggEntries(componentID, 0)
			if err := tracer.feedHandler(inputName, point.Network, pts); err != nil {
				exporter.ObserveAggFlush(componentID, len(pts), time.Since(flushStart), "error")
			} else {
				exporter.ObserveAggFlush(componentID, len(pts), time.Since(flushStart), "ok")
			}
		case <-ctx.Done():
			return
		}
	}
}

// Receive all connections collected in one cycle and send them to DataKit.
func (tracer *NetFlowTracer) feedHandler(name string, cat point.Category, pts []*point.Point) error {
	if err := exporter.FeedPoint(name, cat, pts); err != nil {
		l.Debug(err)
		return err
	}
	return nil
}
