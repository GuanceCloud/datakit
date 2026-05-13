//go:build linux
// +build linux

// Package netflow collects eBPF-network netflow metrics
package netflow

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

	netflowClosedEventQueueSizeEnv = "DK_EBPF_NETFLOW_CLOSED_EVENT_QUEUE_SIZE"
	defaultClosedEventQueueSize    = 4096
	maxClosedEventQueueSize        = 65536
	closedEventDrainLimit          = 8192

	netflowListenPortScanIntervalEnv = "DK_EBPF_NETFLOW_LISTEN_PORT_SCAN_INTERVAL"
	defaultListenPortScanInterval    = time.Minute
	minListenPortScanInterval        = time.Second

	mapConnStats         = "bpfmap_conn_stats"
	mapConnTCPStats      = "bpfmap_conn_tcp_stats"
	mapConnTCPSegments   = "bpfmap_conn_tcp_segments"
	mapNetflowUpdateFail = "bpfmap_netflow_update_fail"
	mapPortBind          = "bpfmap_port_bind"
	mapPortBindProc      = "bpfmap_port_bind_proc"
	mapUDPPortBind       = "bpfmap_udp_port_bind"

	portListening = uint8(1)
)

var enableUDP bool

func SetEnableUDP(on bool) {
	enableUDP = on
}

type NetFlowTracer struct {
	connStatsRecord        *ConnStatsRecord
	closedEventCh          chan *ConncetionClosedInfo
	catalog                *procwatch.Catalog
	seededListenPorts      map[tcpListenPort]struct{}
	listenPorts            map[tcpListenPort]struct{}
	lastListenPortScan     time.Time
	listenPortScanInterval time.Duration
}

func NewNetFlowTracer(catalog *procwatch.Catalog) *NetFlowTracer {
	return &NetFlowTracer{
		connStatsRecord:        newConnStatsRecord(),
		closedEventCh:          make(chan *ConncetionClosedInfo, netflowClosedEventQueueSize()),
		catalog:                catalog,
		seededListenPorts:      make(map[tcpListenPort]struct{}),
		listenPortScanInterval: netflowListenPortScanInterval(),
	}
}

type tcpListenPort struct {
	Netns uint32
	Port  uint16
	Pad   uint16
}

func netflowClosedEventQueueSize() int {
	raw := strings.TrimSpace(os.Getenv(netflowClosedEventQueueSizeEnv))
	if raw == "" {
		return defaultClosedEventQueueSize
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		l.Warnf("invalid %s=%q, use default %d", netflowClosedEventQueueSizeEnv, raw, defaultClosedEventQueueSize)
		return defaultClosedEventQueueSize
	}
	if n > maxClosedEventQueueSize {
		return maxClosedEventQueueSize
	}
	return n
}

func netflowListenPortScanInterval() time.Duration {
	raw := strings.TrimSpace(os.Getenv(netflowListenPortScanIntervalEnv))
	if raw == "" {
		return defaultListenPortScanInterval
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		if seconds, convErr := strconv.Atoi(raw); convErr == nil {
			d = time.Duration(seconds) * time.Second
		} else {
			l.Warnf("invalid %s=%q, use default %s", netflowListenPortScanIntervalEnv, raw, defaultListenPortScanInterval)
			return defaultListenPortScanInterval
		}
	}
	if d <= 0 {
		return defaultListenPortScanInterval
	}
	if d < minListenPortScanInterval {
		return minListenPortScanInterval
	}
	return d
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
	connStatsMap, err := runtime.LookupMap(mapConnStats)
	if err != nil {
		return err
	}

	tcpStatsMap, err := runtime.LookupMap(mapConnTCPStats)
	if err != nil {
		return err
	}

	tcpSegmentsMap, err := runtime.LookupMap(mapConnTCPSegments)
	if err != nil {
		return err
	}

	updateFailMap, _ := runtime.LookupMap(mapNetflowUpdateFail)
	portBindMap, _ := runtime.LookupMap(mapPortBindProc)

	go tracer.connCollectHanllder(ctx, connStatsMap, tcpStatsMap, tcpSegmentsMap, portBindMap,
		updateFailMap, interval, gTags)
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
			Pid:      eventC.conn_info.pid,
			Netns:    eventC.conn_info.netns,
			Meta:     eventC.conn_info.meta,
			NATDaddr: (*(*[4]uint32)(unsafe.Pointer(&eventC.conn_stats.nat_daddr))), //nolint:gosec
			NATDport: uint32(eventC.conn_stats.nat_dport),
		},
		Stats: ConnectionStats{
			SentBytes: eventC.conn_stats.sent_bytes,
			RecvBytes: eventC.conn_stats.recv_bytes,
			Flags:     eventC.conn_stats.flags,
			Direction: eventC.conn_stats.direction,
			Timestamp: eventC.conn_stats.timestamp,
		},
		TCPStats: ConnectionTCPStats{
			StateTransitions: eventC.conn_tcp_stats.state_transitions,
			Retransmits:      eventC.conn_tcp_stats.retransmits,
			Rtt:              eventC.conn_tcp_stats.rtt,
			RttVar:           eventC.conn_tcp_stats.rtt_var,
			ConnectAttempts:  eventC.conn_tcp_stats.connect_attempts,
			ConnectFailures:  eventC.conn_tcp_stats.connect_failures,
			CloseWait:        eventC.conn_tcp_stats.close_wait,
			LastAck:          eventC.conn_tcp_stats.last_ack,
			TimeWait:         eventC.conn_tcp_stats.time_wait,
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

func (tracer *NetFlowTracer) drainClosedEvents(limit int) int {
	drained := 0
	for limit <= 0 || drained < limit {
		select {
		case event := <-tracer.closedEventCh:
			tracer.connStatsRecord.updateClosedUseEvent(event)
			drained++
		default:
			return drained
		}
	}
	return drained
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
			pid:   v.Pid,
			netns: v.Netns,
			meta:  v.Meta,
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

func countTCPSegmentEntries(tcpSegmentsMap *ebpf.Map) (uint32, error) {
	var (
		connInfoC ConnectionInfoC
		segments  struct {
			In  uint32
			Out uint32
		}
		count uint32
	)

	iter := tcpSegmentsMap.Iterate()
	for iter.Next(unsafe.Pointer(&connInfoC), unsafe.Pointer(&segments)) { //nolint:gosec
		count++
	}

	if err := iter.Err(); err != nil {
		return 0, err
	}

	return count, nil
}

func observeNetflowUpdateFailures(updateFailMap *ebpf.Map) {
	if updateFailMap == nil {
		return
	}

	names := []string{
		"conn_stats",
		"tcp_stats",
		"tcp_segments",
	}
	for idx, name := range names {
		key := uint32(idx)
		var count uint64
		if err := updateFailMap.Lookup(&key, &count); err != nil {
			exporter.IncBPFMapObserveError(componentID, mapNetflowUpdateFail, "lookup")
			return
		}
		exporter.ObserveCacheEntries(componentID, "update_fail_"+name, uint64ToInt(count))
	}
}

func uint64ToInt(v uint64) int {
	maxInt := int(^uint(0) >> 1)
	if v > uint64(maxInt) {
		return maxInt
	}
	return int(v)
}

func parseNetnsInode(link string) (uint32, bool) {
	start := strings.LastIndex(link, "[")
	end := strings.LastIndex(link, "]")
	if start < 0 || end <= start+1 {
		return 0, false
	}
	v, err := strconv.ParseUint(link[start+1:end], 10, 32)
	if err != nil {
		return 0, false
	}
	return uint32(v), true
}

func parseTCPListenPorts(r io.Reader, netns uint32, dst map[tcpListenPort]struct{}) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 || fields[0] == "sl" || fields[3] != "0A" {
			continue
		}
		local := fields[1]
		colon := strings.LastIndex(local, ":")
		if colon < 0 || colon+1 >= len(local) {
			continue
		}
		port, err := strconv.ParseUint(local[colon+1:], 16, 16)
		if err != nil || port == 0 {
			continue
		}
		dst[tcpListenPort{Netns: netns, Port: uint16(port)}] = struct{}{}
	}
	return scanner.Err()
}

func scanVisibleNetns(procRoot string) (map[uint32][]string, error) {
	entries, err := os.ReadDir(procRoot)
	if err != nil {
		return nil, err
	}

	netnsRoots := make(map[uint32][]string)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := strconv.Atoi(entry.Name()); err != nil {
			continue
		}

		pidRoot := filepath.Join(procRoot, entry.Name())
		link, err := os.Readlink(filepath.Join(pidRoot, "ns/net"))
		if err != nil {
			continue
		}
		netns, ok := parseNetnsInode(link)
		if !ok {
			continue
		}
		netnsRoots[netns] = append(netnsRoots[netns], pidRoot)
	}

	return netnsRoots, nil
}

func scanTCPListenPorts(procRoot string) (map[tcpListenPort]struct{}, error) {
	netnsRoots, err := scanVisibleNetns(procRoot)
	if err != nil {
		return nil, err
	}

	result := make(map[tcpListenPort]struct{})
	for netns, pidRoots := range netnsRoots {
		for _, pidRoot := range pidRoots {
			opened := false
			for _, name := range []string{"tcp", "tcp6"} {
				f, err := os.Open(filepath.Join(pidRoot, "net", name)) //nolint:gosec
				if err != nil {
					continue
				}
				opened = true
				if err := parseTCPListenPorts(f, netns, result); err != nil {
					_ = f.Close()
					return result, err
				}
				_ = f.Close()
			}
			if opened {
				break
			}
		}
	}

	return result, nil
}

func (tracer *NetFlowTracer) refreshTCPListenPorts(mp *ebpf.Map, force bool) map[tcpListenPort]struct{} {
	if !force && !tracer.lastListenPortScan.IsZero() &&
		time.Since(tracer.lastListenPortScan) < tracer.listenPortScanInterval {
		return tracer.listenPorts
	}

	ports, err := scanTCPListenPorts(procwatch.HostProc())
	if err != nil {
		l.Debugf("scan tcp listen ports failed: %v", err)
		exporter.IncBPFMapObserveError(componentID, mapPortBindProc, "scan_listen")
		return tracer.listenPorts
	}
	tracer.lastListenPortScan = time.Now()
	tracer.listenPorts = ports
	exporter.ObserveCacheEntries(componentID, "tcp_listen_ports", len(ports))
	if mp == nil {
		return ports
	}

	for port := range ports {
		if err := seedTCPListenPort(mp, port); err != nil {
			l.Debugf("seed tcp listen port %+v failed: %v", port, err)
			exporter.IncBPFMapObserveError(componentID, mapPortBindProc, "seed_listen")
		}
	}
	for port := range tracer.seededListenPorts {
		if _, ok := ports[port]; ok {
			continue
		}
		if err := unseedTCPListenPort(mp, port); err != nil {
			l.Debugf("delete stale tcp listen port %+v failed: %v", port, err)
			exporter.IncBPFMapObserveError(componentID, mapPortBindProc, "delete_seeded_listen")
		}
	}
	tracer.seededListenPorts = ports
	return ports
}

func seedTCPListenPort(mp *ebpf.Map, port tcpListenPort) error {
	state := portListening
	return mp.Update(&port, &state, ebpf.UpdateAny)
}

func unseedTCPListenPort(mp *ebpf.Map, port tcpListenPort) error {
	if err := mp.Delete(&port); err != nil && !errors.Is(err, ebpf.ErrKeyNotExist) {
		return err
	}
	return nil
}

func directionFromTCPListenPorts(direction string, conn ConnectionInfo, listenPorts map[tcpListenPort]struct{}) string {
	if !ConnProtocolIsTCP(conn.Meta) || len(listenPorts) == 0 {
		return direction
	}
	if _, ok := listenPorts[tcpListenPort{Netns: conn.Netns, Port: uint16(conn.Sport)}]; ok {
		return DirectionIncoming
	}
	if _, ok := listenPorts[tcpListenPort{Netns: conn.Netns, Port: uint16(conn.Dport)}]; ok {
		return DirectionOutgoing
	}
	return direction
}

func directionStringToC(direction string) uint8 {
	switch direction {
	case DirectionIncoming:
		return ConnDirectionIncoming
	case DirectionOutgoing:
		return ConnDirectionOutgoing
	default:
		return ConnDirectionUnknown
	}
}

const KernelTaskCommLen = 16

// Lock resource connStatsRecord while scanning connStatMap.
func (tracer *NetFlowTracer) connCollectHanllder(ctx context.Context, connStatsMap, tcpStatsMap, tcpSegmentsMap, portBindMap *ebpf.Map,
	updateFailMap *ebpf.Map, interval time.Duration, gTags map[string]string,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	agg := FlowAgg{}
	tracer.refreshTCPListenPorts(portBindMap, true)

	for {
		select {
		case event := <-tracer.closedEventCh:
			tracer.connStatsRecord.updateClosedUseEvent(event)
			tracer.drainClosedEvents(closedEventDrainLimit)
		case <-ticker.C:
			tracer.drainClosedEvents(closedEventDrainLimit)
			exporter.ObserveCacheEntries(componentID, "closed_event_queue", len(tracer.closedEventCh))
			listenPorts := tracer.refreshTCPListenPorts(portBindMap, false)
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
			seenActive := make(map[ConnectionInfo]struct{}, len(tracer.connStatsRecord.lastActiveConns))

			// Collect unclosed connection information and merge it with recorded closed connections
			// and unclosed connections in the previous collection cycle.
			for iter.Next(unsafe.Pointer(&connInfoC), unsafe.Pointer(&connStatsC)) { //nolint:gosec
				connEntries++
				connInfo := ConnectionInfo{
					Saddr:    (*(*[4]uint32)(unsafe.Pointer(&connInfoC.saddr))), //nolint:gosec
					Daddr:    (*(*[4]uint32)(unsafe.Pointer(&connInfoC.daddr))), //nolint:gosec
					Sport:    uint32(connInfoC.sport),
					Dport:    uint32(connInfoC.dport),
					Pid:      connInfoC.pid,
					Netns:    connInfoC.netns,
					Meta:     connInfoC.meta,
					NATDaddr: (*(*[4]uint32)(unsafe.Pointer(&connStatsC.nat_daddr))), //nolint:gosec
					NATDport: uint32(connStatsC.nat_dport),
				}
				fillConntrackNATFallback(&connInfo)

				SrcIPPortRecorder.InsertAndUpdate(connInfo.Saddr)

				if !IPPortFilterIn(&connInfo) {
					continue
				}
				seenActive[connStatsCacheKey(connInfo)] = struct{}{}

				connStats := ConnectionStats{
					SentBytes:   connStatsC.sent_bytes,
					RecvBytes:   connStatsC.recv_bytes,
					SentPackets: connStatsC.sent_packets,
					RecvPackets: connStatsC.recv_packets,
					Flags:       connStatsC.flags,
					Direction:   connStatsC.direction,
					Timestamp:   connStatsC.timestamp,
				}
				if tracer.catalog != nil {
					if v, ok := tracer.catalog.Lookup(int(connInfoC.pid)); ok {
						connInfo.ProcessName = v.Name()
					} else {
						tracer.catalog.ResolveLater(int(connInfoC.pid))
					}
				}
				connStats.Direction = directionStringToC(directionFromTCPListenPorts(
					ConnDirection2Str(connStats.Direction), connInfo, listenPorts))

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
							StateTransitions: tcpStatsC.state_transitions,
							Retransmits:      tcpStatsC.retransmits,
							Rtt:              tcpStatsC.rtt,
							RttVar:           tcpStatsC.rtt_var,
							ConnectAttempts:  tcpStatsC.connect_attempts,
							ConnectFailures:  tcpStatsC.connect_failures,
							CloseWait:        tcpStatsC.close_wait,
							LastAck:          tcpStatsC.last_ack,
							TimeWait:         tcpStatsC.time_wait,
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
			exporter.ObserveBPFMap(componentID, mapConnStats, connEntries, connStatsMap.MaxEntries())
			iterErr := iter.Err()
			if iterErr != nil {
				l.Warnf("iterate %s failed: %s", mapConnStats, iterErr)
				exporter.IncBPFMapObserveError(componentID, mapConnStats, "iterate")
			}

			tcpEntries, err := countTCPStatsEntries(tcpStatsMap)
			if err != nil {
				l.Warnf("iterate %s failed: %s", mapConnTCPStats, err)
				exporter.IncBPFMapObserveError(componentID, mapConnTCPStats, "iterate")
			} else {
				exporter.ObserveBPFMap(componentID, mapConnTCPStats, tcpEntries, tcpStatsMap.MaxEntries())
			}

			segmentEntries, err := countTCPSegmentEntries(tcpSegmentsMap)
			if err != nil {
				l.Warnf("iterate %s failed: %s", mapConnTCPSegments, err)
				exporter.IncBPFMapObserveError(componentID, mapConnTCPSegments, "iterate")
			} else {
				exporter.ObserveBPFMap(componentID, mapConnTCPSegments, segmentEntries, tcpSegmentsMap.MaxEntries())
			}
			observeNetflowUpdateFailures(updateFailMap)
			if iterErr == nil {
				if removed := tracer.connStatsRecord.pruneLastActiveNotSeen(seenActive); removed > 0 {
					exporter.AddCacheEvictions(componentID, "last_active", "not_seen", removed)
				}
			}

			if len(connsNeedCleanup) > 0 {
				for _, conn := range connsNeedCleanup {
					tracer.connStatsRecord.deleteLastActive(conn)
				}
				exporter.AddCacheEvictions(componentID, "last_active", "expired_zero", len(connsNeedCleanup))
				tracer.bpfMapCleanup(connsNeedCleanup, connStatsMap, tcpStatsMap, tcpSegmentsMap)
			}
			exporter.ObserveCacheEntries(componentID, "last_active", tracer.connStatsRecord.lastActiveLen())
			// Collect connections that are closed for the current cycle.
			for k, v := range tracer.connStatsRecord.closedConns {
				connInfo := k
				if info, ok := tracer.connStatsRecord.closedConnInfo[k]; ok {
					connInfo = info
				}
				v.Stats.Direction = directionStringToC(directionFromTCPListenPorts(
					ConnDirection2Str(v.Stats.Direction), connInfo, listenPorts))
				err := agg.Append(connInfo, v)
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
