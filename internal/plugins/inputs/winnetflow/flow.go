// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package winnetflow

import (
	"net"
	"net/netip"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

// Windows TCP state values, see MIB_TCP_STATE.
const (
	tcpStateClosed      = 1
	tcpStateListen      = 2
	tcpStateSynSent     = 3
	tcpStateSynRcvd     = 4
	tcpStateEstablished = 5
	tcpStateFinWait1    = 6
	tcpStateFinWait2    = 7
	tcpStateCloseWait   = 8
	tcpStateLastAck     = 9
	tcpStateTimeWait    = 10
	tcpStateDeleteTcb   = 11
)

const (
	directionIncoming = "incoming"
	directionOutgoing = "outgoing"

	// Align client-port rollup with the eBPF netflow collector. Starting at
	// 32768 also covers common Linux ephemeral ports observed by Windows
	// servers, while the known direction keeps server ports intact.
	clientPortRollupStart uint32 = 32768
	normalizedClientPort  uint32 = ^uint32(0)
)

type endpointFilterReason uint8

const (
	endpointAccepted endpointFilterReason = iota
	endpointInvalid
	endpointLoopback
)

// Flow-map bounds. The default mirrors eBPF; the hard maximum limits transient
// memory while flush holds both the detached snapshot and the next hot map.
const (
	defaultMaxFlows    = 65536
	minMaxFlows        = 1024
	maxMaxFlows        = 262144
	stateTrackingTTL   = 24 * time.Hour
	statePruneInterval = 10 * time.Minute
)

// eventKind describes what an ETW flow event represents.
type eventKind int

const (
	evTCPEstablished eventKind = iota
	evTCPConnectFail
	evTCPClosed
	evTCPDataSend
	evTCPDataRecv
	evTCPRetransmit
	evTCPStateChange
	evRTT
	evUDPDataSend
	evUDPDataRecv
	// evBurstReplay carries cumulative totals that were buffered while a
	// connection's tuple was still unknown.
	evBurstReplay
)

// flowEvent is a normalized per-connection event produced by the ETW layer.
// src/dst always follow the local/remote side of the owning socket.
type flowEvent struct {
	ts        time.Time
	kind      eventKind
	transport string // tcp | udp
	family    string // IPv4 | IPv6
	direction string // incoming | outgoing
	pid       uint32
	srcIP     string
	srcPort   uint32
	dstIP     string
	dstPort   uint32

	bytes   uint64 // payload bytes of this event
	packets uint64 // packets/datagrams of this event (0 if unknown)
	rtt     uint64 // microseconds
	rttVar  uint64 // microseconds
	retrans uint64

	oldState uint32
	newState uint32

	// Burst replay totals (only meaningful for evBurstReplay).
	cumBytesSend   uint64
	cumBytesRecv   uint64
	cumPacketsSend uint64
	cumPacketsRecv uint64
	cumRetrans     uint64
	cumRTT         uint64
	cumRTTVar      uint64
	cumRTTCount    uint64
}

type flowKey struct {
	transport string
	family    string
	direction string
	pid       uint32
	srcIP     string
	srcPort   uint32
	dstIP     string
	dstPort   uint32
}

type flowStats struct {
	bytesRead    uint64
	bytesWritten uint64
	packetsRead  uint64
	packetsWrite uint64

	retransmits uint64
	rttSum      uint64
	rttVarSum   uint64
	rttCount    uint64

	tcpEstablished     uint64
	tcpClosed          uint64
	tcpConnectAttempts uint64
	tcpConnectFailures uint64
	tcpCloseWait       uint64
	tcpLastAck         uint64
	tcpTimeWait        uint64
}

// tcpLifecycleSeen de-duplicates overlapping TCPIP provider signals. Windows
// can report one transition through both a dedicated connect/close event and a
// generic state-change event. The state survives aggregation flushes because
// the two signals can straddle an interval boundary.
type tcpLifecycleSeen struct {
	attempted   bool
	established bool
	closed      bool
	lastSeen    time.Time
}

// resolveProcessName maps a PID to a process name. It is installed by the
// Windows-specific acquisition layer; on other platforms it stays nil.
var resolveProcessName func(pid uint32) string
var refreshProcessName func(pid uint32) string

func processName(pid uint32) string {
	if resolveProcessName != nil {
		if name := resolveProcessName(pid); name != "" {
			return name
		}
	}
	return "unknown"
}

type flowAggregator struct {
	mu               sync.Mutex
	data             map[flowKey]*flowStats
	maxFlows         int
	flowsSkipped     uint64
	invalidFiltered  atomic.Uint64
	loopbackFiltered atomic.Uint64
	includeLoopback  bool // test-only escape hatch for local ETW smoke tests
	// stateSince remembers the TCP states a flow passed through, so transitions
	// into CLOSE_WAIT / LAST_ACK / TIME_WAIT can be counted.
	stateSince      map[flowKey]map[uint32]time.Time
	lifecycle       map[flowKey]*tcpLifecycleSeen
	lastStatePrune  time.Time
	interval        time.Duration
	extraTags       map[string]string
	warmProcessName func(uint32)
	warmPIDs        map[uint32]struct{}
}

// etwSessionStats carries live ETW session counters.
type etwSessionStats struct {
	numberOfBuffers     uint32
	freeBuffers         uint32
	eventsLost          uint32
	realTimeBuffersLost uint32
}

// collectorStats is the collector self-telemetry snapshot.
type collectorStats struct {
	decoded          uint64
	dropped          uint64
	parseErrors      uint64
	tcbEvicted       uint64
	pendingEvicted   uint64
	flowsSkipped     uint64
	invalidFiltered  uint64
	loopbackFiltered uint64
	session          etwSessionStats
}

// etwConfig carries tunable ETW session parameters.
type etwConfig struct {
	bufferSizeKB uint32
	minBuffers   uint32
	maxBuffers   uint32
}

func newFlowAggregator(interval time.Duration, maxFlows int) *flowAggregator {
	return &flowAggregator{
		data:           make(map[flowKey]*flowStats),
		stateSince:     make(map[flowKey]map[uint32]time.Time),
		lifecycle:      make(map[flowKey]*tcpLifecycleSeen),
		warmPIDs:       make(map[uint32]struct{}),
		lastStatePrune: time.Now(),
		interval:       interval,
		maxFlows:       maxFlows,
	}
}

func (a *flowAggregator) onEvent(ev *flowEvent) {
	if ev == nil {
		return
	}
	switch filterEndpoints(ev.srcIP, ev.srcPort, ev.dstIP, ev.dstPort, a.includeLoopback) {
	case endpointAccepted:
	case endpointInvalid:
		a.invalidFiltered.Add(1)
		return
	case endpointLoopback:
		a.loopbackFiltered.Add(1)
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	key := flowKey{
		transport: ev.transport,
		family:    ev.family,
		direction: ev.direction,
		pid:       ev.pid,
		srcIP:     ev.srcIP,
		srcPort:   ev.srcPort,
		dstIP:     ev.dstIP,
		dstPort:   ev.dstPort,
	}

	st := a.data[key]
	newIdentity := false
	if st == nil && key.pid != 0 {
		st = a.promoteUnknownPIDLocked(key)
		newIdentity = st != nil
	}
	if st == nil {
		if a.maxFlows > 0 && len(a.data) >= a.maxFlows {
			// Bound memory under connection storms: drop new flows instead of
			// growing without limit, and count them for observability.
			a.flowsSkipped++
			return
		}
		st = &flowStats{}
		a.data[key] = st
		newIdentity = true
	}
	if newIdentity {
		a.warmPIDLocked(ev.pid)
	}

	switch ev.kind {
	case evTCPEstablished:
		lc := a.lifecycleFor(key, ev.ts)
		if lc.closed {
			lc.reset(ev.ts)
		}
		if ev.direction == directionOutgoing && !lc.attempted {
			st.tcpConnectAttempts++
			lc.attempted = true
		}
		if !lc.established {
			st.tcpEstablished++
			lc.established = true
		}
	case evTCPConnectFail:
		lc := a.lifecycleFor(key, ev.ts)
		if lc.closed {
			lc.reset(ev.ts)
		}
		if !lc.attempted {
			st.tcpConnectAttempts++
			lc.attempted = true
		}
		st.tcpConnectFailures++
		lc.closed = true
	case evTCPClosed:
		lc := a.lifecycleFor(key, ev.ts)
		if !lc.closed {
			st.tcpClosed++
			lc.closed = true
		}
		a.accumulateAllStates(key)
	case evTCPDataSend:
		st.bytesWritten += ev.bytes
		st.packetsWrite += ev.packets
		a.accumulateRTT(st, ev)
	case evTCPDataRecv:
		st.bytesRead += ev.bytes
		st.packetsRead += ev.packets
		a.accumulateRTT(st, ev)
	case evTCPRetransmit:
		st.retransmits += ev.retrans
	case evTCPStateChange:
		a.accumulateState(key, ev.oldState)
		lc := a.lifecycleFor(key, ev.ts)
		switch ev.newState {
		case tcpStateSynSent:
			// SYN_SENT starts a new active connection. This also handles fast
			// tuple reuse after a previous connection closed.
			if lc.closed || lc.established {
				lc.reset(ev.ts)
			}
			if !lc.attempted {
				st.tcpConnectAttempts++
				lc.attempted = true
			}
		case tcpStateEstablished:
			if lc.closed {
				lc.reset(ev.ts)
			}
			if !lc.established {
				st.tcpEstablished++
				lc.established = true
			}
		case tcpStateClosed, tcpStateDeleteTcb:
			if !lc.closed {
				st.tcpClosed++
				lc.closed = true
			}
		}
		a.setState(key, ev.ts, ev.newState)
	case evRTT:
		a.accumulateRTT(st, ev)
	case evUDPDataSend:
		st.bytesWritten += ev.bytes
		st.packetsWrite += ev.packets
	case evUDPDataRecv:
		st.bytesRead += ev.bytes
		st.packetsRead += ev.packets
	case evBurstReplay:
		st.bytesWritten += ev.cumBytesSend
		st.bytesRead += ev.cumBytesRecv
		st.packetsWrite += ev.cumPacketsSend
		st.packetsRead += ev.cumPacketsRecv
		st.retransmits += ev.cumRetrans
		if ev.cumRTT > 0 {
			st.rttSum += ev.cumRTT
			st.rttVarSum += ev.cumRTTVar
			st.rttCount += ev.cumRTTCount
		}
	}
}

// promoteUnknownPIDLocked moves early data that arrived before a control event
// supplied the owning PID onto the authoritative flow key. TCPIP events can be
// delivered from different CPUs, so a data event with an inline tuple may
// occasionally precede the accept/connect event for the same TCB.
func (a *flowAggregator) promoteUnknownPIDLocked(key flowKey) *flowStats {
	unknown := key
	unknown.pid = 0
	st := a.data[unknown]
	if st == nil {
		if unknown.direction == directionIncoming {
			unknown.direction = directionOutgoing
		} else {
			unknown.direction = directionIncoming
		}
		st = a.data[unknown]
	}
	if st == nil {
		return nil
	}
	delete(a.data, unknown)
	a.data[key] = st

	if old := a.lifecycle[unknown]; old != nil {
		delete(a.lifecycle, unknown)
		if current := a.lifecycle[key]; current == nil {
			a.lifecycle[key] = old
		} else {
			current.attempted = current.attempted || old.attempted
			current.established = current.established || old.established
			current.closed = current.closed || old.closed
			if old.lastSeen.After(current.lastSeen) {
				current.lastSeen = old.lastSeen
			}
		}
	}
	if old := a.stateSince[unknown]; old != nil {
		delete(a.stateSince, unknown)
		current := a.stateSince[key]
		if current == nil {
			a.stateSince[key] = old
		} else {
			for state, since := range old {
				if existing, ok := current[state]; !ok || since.Before(existing) {
					current[state] = since
				}
			}
		}
	}
	return st
}

func (a *flowAggregator) warmPIDLocked(pid uint32) {
	if a.warmProcessName == nil || pid == 0 {
		return
	}
	if _, warmed := a.warmPIDs[pid]; warmed {
		return
	}
	a.warmPIDs[pid] = struct{}{}
	a.warmProcessName(pid)
}

func (a *flowAggregator) lifecycleFor(key flowKey, ts time.Time) *tcpLifecycleSeen {
	lc := a.lifecycle[key]
	if lc == nil {
		if a.maxFlows > 0 && len(a.lifecycle) >= a.maxFlows {
			for staleKey := range a.lifecycle {
				delete(a.lifecycle, staleKey)
				break
			}
		}
		lc = &tcpLifecycleSeen{}
		a.lifecycle[key] = lc
	}
	if ts.IsZero() {
		ts = time.Now()
	}
	lc.lastSeen = ts
	return lc
}

func (s *tcpLifecycleSeen) reset(ts time.Time) {
	*s = tcpLifecycleSeen{lastSeen: ts}
}

func (a *flowAggregator) accumulateRTT(st *flowStats, ev *flowEvent) {
	if ev.rtt > 0 {
		st.rttSum += ev.rtt
		st.rttVarSum += ev.rttVar
		st.rttCount++
	}
}

func (a *flowAggregator) setState(key flowKey, ts time.Time, state uint32) {
	if state != tcpStateCloseWait && state != tcpStateLastAck && state != tcpStateTimeWait {
		return
	}
	m := a.stateSince[key]
	if m == nil {
		if a.maxFlows > 0 && len(a.stateSince) >= a.maxFlows {
			for staleKey := range a.stateSince {
				delete(a.stateSince, staleKey)
				break
			}
		}
		m = make(map[uint32]time.Time)
		a.stateSince[key] = m
	}
	m[state] = ts
}

func (a *flowAggregator) pruneStatesLocked(now time.Time) {
	for key, states := range a.stateSince {
		for state, since := range states {
			if now.Sub(since) >= stateTrackingTTL {
				delete(states, state)
			}
		}
		if len(states) == 0 {
			delete(a.stateSince, key)
		}
	}
	for key, state := range a.lifecycle {
		if now.Sub(state.lastSeen) >= stateTrackingTTL {
			delete(a.lifecycle, key)
		}
	}
}

// accumulateState counts the transition out of oldState into the matching
// counter (schema-compatible with ebpf-net/netflow, where tcp_close_wait /
// tcp_last_ack / tcp_time_wait are transition counts) and clears the entry.
func (a *flowAggregator) accumulateState(key flowKey, oldState uint32) {
	m := a.stateSince[key]
	if m == nil {
		return
	}
	if _, ok := m[oldState]; !ok {
		delete(m, oldState)
		return
	}
	st := a.data[key]
	if st == nil {
		delete(m, oldState)
		return
	}
	switch oldState {
	case tcpStateCloseWait:
		st.tcpCloseWait++
	case tcpStateLastAck:
		st.tcpLastAck++
	case tcpStateTimeWait:
		st.tcpTimeWait++
	}
	delete(m, oldState)
}

// accumulateAllStates counts every tracked TCP state once (used when a
// connection closes without a final state transition event, so states the
// connection passed through are still reflected in the transition counters).
func (a *flowAggregator) accumulateAllStates(key flowKey) {
	m := a.stateSince[key]
	if m == nil {
		return
	}
	for state := range m {
		st := a.data[key]
		if st == nil {
			continue
		}
		switch state {
		case tcpStateCloseWait:
			st.tcpCloseWait++
		case tcpStateLastAck:
			st.tcpLastAck++
		case tcpStateTimeWait:
			st.tcpTimeWait++
		}
	}
	delete(a.stateSince, key)
}

func avg(sum uint64, count uint64) int64 {
	if count == 0 {
		return 0
	}
	return int64(sum / count)
}

// flush snapshots and resets the aggregator, producing netflow points with the
// same field/tag schema as the Linux eBPF netflow collector.
func (a *flowAggregator) flush() []*point.Point {
	a.mu.Lock()
	ts := time.Now()
	if ts.Sub(a.lastStatePrune) >= statePruneInterval {
		a.pruneStatesLocked(ts)
		a.lastStatePrune = ts
	}
	if len(a.data) == 0 {
		a.mu.Unlock()
		return nil
	}

	data := a.data
	a.data = make(map[flowKey]*flowStats, boundedMapHint(len(a.data), defaultMaxFlows))
	a.warmPIDs = make(map[uint32]struct{}, processNameCacheHint(len(a.warmPIDs)))
	a.mu.Unlock()
	data = aggregateNormalizedFlows(data)

	pts := make([]*point.Point, 0, len(data))
	processNames := make(map[uint32]string, processNameCacheHint(len(data)))
	for key, st := range data {
		fields := map[string]any{
			"bytes_read":      int64(st.bytesRead),
			"bytes_written":   int64(st.bytesWritten),
			"packets_read":    int64(st.packetsRead),
			"packets_written": int64(st.packetsWrite),
		}
		if key.transport == "tcp" {
			fields["retransmits"] = int64(st.retransmits)
			fields["rtt"] = avg(st.rttSum, st.rttCount)
			fields["rtt_var"] = avg(st.rttVarSum, st.rttCount)
			fields["tcp_closed"] = int64(st.tcpClosed)
			fields["tcp_established"] = int64(st.tcpEstablished)
			fields["tcp_connect_attempts"] = int64(st.tcpConnectAttempts)
			fields["tcp_connect_failures"] = int64(st.tcpConnectFailures)
			fields["tcp_close_wait"] = int64(st.tcpCloseWait)
			fields["tcp_last_ack"] = int64(st.tcpLastAck)
			fields["tcp_time_wait"] = int64(st.tcpTimeWait)
		}

		tags := map[string]string{
			"family":       key.family,
			"direction":    key.direction,
			"transport":    key.transport,
			"pid":          strconv.FormatUint(uint64(key.pid), 10),
			"src_ip":       key.srcIP,
			"dst_ip":       key.dstIP,
			"src_port":     flowPortTag(key.srcPort),
			"dst_port":     flowPortTag(key.dstPort),
			"process_name": flushProcessName(processNames, key.pid),
			"src_ip_type":  endpointIPType(key.srcIP),
			"dst_ip_type":  endpointIPType(key.dstIP),
			"dst_nat_ip":   "N/A",
			"dst_nat_port": "N/A",
		}
		addClientServerInfo(tags, fields)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(ts))
		if len(a.extraTags) > 0 {
			opts = append(opts, point.WithExtraTags(a.extraTags))
		}
		pts = append(pts, point.NewPoint(metricName,
			append(point.NewTags(tags), point.NewKVs(fields)...), opts...))
	}

	return pts
}

// aggregateNormalizedFlows runs after the exact per-connection snapshot has
// been detached. Keeping exact ports on the hot path preserves independent
// TCP lifecycle state for concurrent connections; normalizing only here
// safely reduces the number and cardinality of emitted points.
func aggregateNormalizedFlows(data map[flowKey]*flowStats) map[flowKey]*flowStats {
	result := make(map[flowKey]*flowStats, boundedMapHint(len(data), defaultMaxFlows))
	for key, st := range data {
		key.srcPort, key.dstPort = normalizeClientPort(key.direction, key.srcPort, key.dstPort)
		if current := result[key]; current != nil {
			mergeFlowStats(current, st)
			continue
		}
		result[key] = st
	}
	return result
}

func mergeFlowStats(dst, src *flowStats) {
	dst.bytesRead += src.bytesRead
	dst.bytesWritten += src.bytesWritten
	dst.packetsRead += src.packetsRead
	dst.packetsWrite += src.packetsWrite
	dst.retransmits += src.retransmits
	dst.rttSum += src.rttSum
	dst.rttVarSum += src.rttVarSum
	dst.rttCount += src.rttCount
	dst.tcpEstablished += src.tcpEstablished
	dst.tcpClosed += src.tcpClosed
	dst.tcpConnectAttempts += src.tcpConnectAttempts
	dst.tcpConnectFailures += src.tcpConnectFailures
	dst.tcpCloseWait += src.tcpCloseWait
	dst.tcpLastAck += src.tcpLastAck
	dst.tcpTimeWait += src.tcpTimeWait
}

func normalizeClientPort(direction string, srcPort, dstPort uint32) (uint32, uint32) {
	switch direction {
	case directionOutgoing:
		if srcPort >= clientPortRollupStart {
			srcPort = normalizedClientPort
		}
	case directionIncoming:
		if dstPort >= clientPortRollupStart {
			dstPort = normalizedClientPort
		}
	}
	return srcPort, dstPort
}

func flowPortTag(port uint32) string {
	if port == normalizedClientPort {
		return "*"
	}
	return strconv.FormatUint(uint64(port), 10)
}

func filterEndpoints(srcIP string, srcPort uint32, dstIP string, dstPort uint32, includeLoopback bool) endpointFilterReason {
	if srcPort == 0 || dstPort == 0 {
		return endpointInvalid
	}
	src, srcErr := netip.ParseAddr(srcIP)
	dst, dstErr := netip.ParseAddr(dstIP)
	if srcErr != nil || dstErr != nil || src.IsUnspecified() || dst.IsUnspecified() {
		return endpointInvalid
	}
	if !includeLoopback && src.IsLoopback() && dst.IsLoopback() {
		return endpointLoopback
	}
	return endpointAccepted
}

// addClientServerInfo mirrors the endpoint-role normalization used by the
// Linux eBPF netflow/httpflow exporters. src/dst remain local/remote, while
// client/server remain stable across incoming and outgoing flows.
func addClientServerInfo(tags map[string]string, fields map[string]any) {
	if tags["direction"] == directionIncoming {
		tags["conn_side"] = "server"
		tags["server_ip"] = tags["src_ip"]
		tags["server_ip_type"] = tags["src_ip_type"]
		tags["server_port"] = tags["src_port"]
		tags["client_ip"] = tags["dst_ip"]
		tags["client_ip_type"] = tags["dst_ip_type"]
		tags["client_port"] = tags["dst_port"]
		fields["client_sent"] = fields["bytes_read"]
		fields["server_sent"] = fields["bytes_written"]
		return
	}

	tags["conn_side"] = "client"
	tags["server_ip"] = tags["dst_ip"]
	tags["server_ip_type"] = tags["dst_ip_type"]
	tags["server_port"] = tags["dst_port"]
	tags["client_ip"] = tags["src_ip"]
	tags["client_ip_type"] = tags["src_ip_type"]
	tags["client_port"] = tags["src_port"]
	fields["client_sent"] = fields["bytes_written"]
	fields["server_sent"] = fields["bytes_read"]
}

func endpointIPType(raw string) string {
	ip := net.ParseIP(raw)
	if ip == nil {
		return "other"
	}
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

func processNameCacheHint(entries int) int {
	if entries < 1024 {
		return entries
	}
	return 1024
}

func boundedMapHint(entries, limit int) int {
	if entries < 0 {
		return 0
	}
	if entries > limit {
		return limit
	}
	return entries
}

func flushProcessName(cache map[uint32]string, pid uint32) string {
	if name, ok := cache[pid]; ok {
		return name
	}
	name := processName(pid)
	cache[pid] = name
	return name
}

// processNameWarmer resolves a process as soon as its first flow arrives,
// while the process is likely still alive. Resolution stays off the ordered
// ETW decoder path and the bounded queue prevents connection storms from
// creating unbounded work.
type processNameWarmer struct {
	ch      chan uint32
	stopCh  chan struct{}
	stopped chan struct{}
	mu      sync.Mutex
	pending map[uint32]struct{}
}

func newProcessNameWarmer() *processNameWarmer {
	w := &processNameWarmer{
		ch:      make(chan uint32, 1024),
		stopCh:  make(chan struct{}),
		stopped: make(chan struct{}),
		pending: make(map[uint32]struct{}),
	}
	go w.run()
	return w
}

func (w *processNameWarmer) enqueue(pid uint32) {
	if pid == 0 {
		return
	}
	w.mu.Lock()
	if _, ok := w.pending[pid]; ok {
		w.mu.Unlock()
		return
	}
	w.pending[pid] = struct{}{}
	w.mu.Unlock()

	select {
	case w.ch <- pid:
	default:
		w.mu.Lock()
		delete(w.pending, pid)
		w.mu.Unlock()
	}
}

func (w *processNameWarmer) run() {
	defer close(w.stopped)
	for {
		select {
		case <-w.stopCh:
			return
		case pid := <-w.ch:
			if refreshProcessName != nil {
				refreshProcessName(pid)
			} else {
				processName(pid)
			}
			w.mu.Lock()
			delete(w.pending, pid)
			w.mu.Unlock()
		}
	}
}

func (w *processNameWarmer) stop() {
	close(w.stopCh)
	<-w.stopped
}

func (a *flowAggregator) flowsSkippedCount() uint64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.flowsSkipped
}

func (a *flowAggregator) filterCounts() (invalid, loopback uint64) {
	return a.invalidFiltered.Load(), a.loopbackFiltered.Load()
}
