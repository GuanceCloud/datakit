// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows

package winnetflow

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"

	gopsnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

const sessionName = "datakit-winnetflow"

// Microsoft-Windows-TCPIP provider GUID
// {2f07e2ee-15db-40f1-90ef-9d7ba282188a}
var tcpipProviderGUID = parseGUID("2f07e2ee-15db-40f1-90ef-9d7ba282188a")

// TCPIP event keywords. Only these keyword families are enabled so the event
// volume stays proportional to actual traffic.
const (
	kwConnect = 0x8000000400000484 // connect complete/failure, rundown
	kwClose   = 0x8000001000000484 // disconnect/abort/timeout
	kwSend    = 0x8000000100000000 // TCP send / RTT sample / UDP send
	kwRecv    = 0x8000000200000000 // TCP recv / UDP recv
	kwRetrans = 0x8000000100000080 // retransmit round
	kwState   = 0x8000000000000404 // TCP state change
	kwRst     = 0x8000000600000080 // connection terminated by RST

	matchAnyKeywords = kwConnect | kwClose | kwSend | kwRecv | kwRetrans | kwState | kwRst
)

// TCPIP event IDs the collector is interested in. They are also passed to
// EnableTraceEx2 as an EVENT_FILTER_TYPE_EVENT_ID filter so the kernel drops
// everything else before it reaches us.
var wantedEventIDs = []uint16{
	1017, // TcpAcceptTcbComplete
	1033, // TcpConnectTcbComplete
	1034, // TcpConnectTcbFailure
	1040, // TcpAbortTcbComplete
	1043, // TcpDisconnectTcbComplete
	1045, // TcpConnectTcbTimeout
	1051, // TcpTcbStateChange
	1073, // TcpDataTransferSend (legacy)
	1074, // TcpDataTransferReceive
	1169, // UdpEndpointSendMessages
	1170, // UdpEndpointReceiveMessages
	1184, // TcpConnectionTerminatedRcvdRst
	1300, // TcpConnectionRundown
	1332, // TcpDataTransferSend (current)
	1341, // TcpDataTransferRttSample
	1351, // TcpDataTransferRetransmitRound
}

var wantedEventSet = func() map[uint16]struct{} {
	m := make(map[uint16]struct{}, len(wantedEventIDs))
	for _, id := range wantedEventIDs {
		m[id] = struct{}{}
	}
	return m
}()

// flowEventBatch is the allocation-free decode result. A TCP event produces
// at most two normalized events: an optional pending-data replay followed by
// the current event.
type flowEventBatch struct {
	events [2]flowEvent
	count  uint8
}

func oneFlowEvent(ev flowEvent) flowEventBatch {
	return flowEventBatch{events: [2]flowEvent{ev}, count: 1}
}

func (b *flowEventBatch) append(ev flowEvent) {
	if int(b.count) >= len(b.events) {
		return
	}
	b.events[b.count] = ev
	b.count++
}

func (b *flowEventBatch) len() int { return int(b.count) }

func (b *flowEventBatch) at(i int) *flowEvent {
	if i < 0 || i >= int(b.count) {
		return nil
	}
	return &b.events[i]
}

func defaultETWConfig() etwConfig {
	return etwConfig{
		bufferSizeKB: defaultEtwBufferSizeKB,
		minBuffers:   defaultEtwMinBuffers,
		maxBuffers:   defaultEtwMaxBuffers,
	}
}

// rawEvent is the minimal event copy made on the ETW callback thread.
type rawEvent struct {
	eventID   uint16
	version   uint8
	timestamp int64
	userData  []byte
}

// tcbInfo remembers the tuple and PID associated with a TCP control block.
// TCP data events usually carry only the Tcb pointer, so this map is the
// bridge to the connection key.
type tcbInfo struct {
	srcIP     string
	srcPort   uint32
	dstIP     string
	dstPort   uint32
	pid       uint32
	family    string
	direction string
	lastSeen  time.Time
}

// pendingStats buffers flow deltas for a TCB whose tuple is not known yet.
// Once the tuple is discovered (connect/rundown/close event or inline
// addresses), the totals are replayed into the aggregator as one burst event.
type pendingStats struct {
	bytesSent   uint64
	bytesRecv   uint64
	packetsSent uint64
	packetsRecv uint64
	retrans     uint64
	rttSum      uint64
	rttVarSum   uint64
	rttCount    uint64
	lastSeen    time.Time
}

// listenSnapshot holds bound/listening local endpoints and is used to infer
// the direction of events that predate collector startup or lack a connect
// event. Separate TCP and UDP instances are maintained.
type listenSnapshot struct {
	mu        sync.RWMutex
	endpoints map[string]struct{}
}

type etwDecoder struct {
	mu             sync.Mutex
	tcbs           map[uint64]*tcbInfo
	tcbLimit       int
	tcbTTL         time.Duration
	tcbEvicted     uint64
	pending        map[uint64]*pendingStats
	pendingEvicted uint64
	pendingLimit   int
	pendingTTL     time.Duration
	udpSrv         *listenSnapshot
	tcpSrv         *listenSnapshot
	parseErrors    *atomic.Uint64
}

type etwCollector struct {
	agg       *flowAggregator
	session   *etwSession
	decoder   *etwDecoder
	udpSrv    *listenSnapshot
	tcpSrv    *listenSnapshot
	cfg       etwConfig
	rawCh     chan rawEvent
	dropCount atomic.Uint64
	decoded   atomic.Uint64
	parseErr  atomic.Uint64
	onFatal   func(string)
	stopCh    chan struct{}
	traceDone chan struct{}
	stopping  atomic.Bool
	stopOnce  sync.Once
	wg        sync.WaitGroup
	bufPool   sync.Pool
}

// recordHandler is implemented by every ETW consumer that wants to receive
// EVENT_RECORD callbacks from a real-time session. Each collector (TCPIP flows,
// HttpService httpflow, ...) owns one session and one handler.
type recordHandler interface {
	handleRecord(*eventRecord)
}

// handlerBox is the UserContext passed to OpenTrace. Native code only sees a
// pointer, so the box keeps the interface value alive for the session's
// lifetime and the trampoline can dispatch safely.
type handlerBox struct {
	h recordHandler
}

// TCP correlation is order-sensitive: connect must be processed before data,
// and data before close. Keep a single consumer so events copied by the ETW
// callback cannot overtake each other while updating the TCB map.
const decodeWorkers = 1

const (
	defaultTCBLimit       = 131072
	defaultTCBTTL         = 24 * time.Hour
	defaultPendingTTL     = 10 * time.Minute
	tcbPruneInterval      = 10 * time.Minute
	pooledEventBufferSize = 512
)

func newETWCollector(agg *flowAggregator, cfg etwConfig) (flowSource, error) {
	udpSrv := &listenSnapshot{endpoints: make(map[string]struct{})}
	tcpSrv := &listenSnapshot{endpoints: make(map[string]struct{})}
	c := &etwCollector{
		agg:       agg,
		udpSrv:    udpSrv,
		tcpSrv:    tcpSrv,
		cfg:       cfg,
		rawCh:     make(chan rawEvent, 16384),
		stopCh:    make(chan struct{}),
		traceDone: make(chan struct{}),
	}
	c.decoder = &etwDecoder{
		tcbs:         make(map[uint64]*tcbInfo),
		tcbLimit:     defaultTCBLimit,
		tcbTTL:       defaultTCBTTL,
		pending:      make(map[uint64]*pendingStats),
		pendingLimit: 65536,
		pendingTTL:   defaultPendingTTL,
		udpSrv:       udpSrv,
		tcpSrv:       tcpSrv,
		parseErrors:  &c.parseErr,
	}
	c.bufPool.New = func() interface{} {
		return new(eventBuffer)
	}
	return c, nil
}

func (c *etwCollector) start() error {
	s := &etwSession{name: sessionName}
	if err := s.start(c.cfg, &tcpipProviderGUID, 4, matchAnyKeywords, wantedEventIDs, c); err != nil {
		return fmt.Errorf("start ETW session: %w", err)
	}
	c.session = s

	c.udpSrv.refresh("udp")
	c.tcpSrv.refresh("tcp")
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer close(c.traceDone)
		err := processTrace(&s.traceHandle, 1)
		if !c.stopping.Load() {
			reportUnexpectedTraceExit("winnetflow ETW session terminated unexpectedly", err, c.onFatal)
		}
	}()
	c.wg.Add(1)
	go c.refreshListenSnapshots()
	for i := 0; i < decodeWorkers; i++ {
		c.wg.Add(1)
		go c.decodeLoop()
	}
	return nil
}

func reportUnexpectedTraceExit(message string, err error, onFatal func(string)) {
	if err != nil {
		message += ": " + err.Error()
	}
	l.Error(message)
	if onFatal != nil {
		onFatal(message)
	}
}

func (c *etwCollector) stop() {
	c.stopOnce.Do(func() {
		c.stopping.Store(true)
		if c.session != nil {
			c.session.stop()
		}
		<-c.traceDone
		close(c.stopCh)
		c.wg.Wait()
	})
}

func (c *etwCollector) stats() collectorStats {
	s := collectorStats{
		decoded:     c.decoded.Load(),
		dropped:     c.dropCount.Load(),
		parseErrors: c.parseErr.Load(),
	}
	if c.agg != nil {
		s.flowsSkipped = c.agg.flowsSkippedCount()
		s.invalidFiltered, s.loopbackFiltered = c.agg.filterCounts()
	}
	if c.decoder != nil {
		c.decoder.mu.Lock()
		s.pendingEvicted = c.decoder.pendingEvicted
		s.tcbEvicted = c.decoder.tcbEvicted
		c.decoder.mu.Unlock()
	}
	if c.session != nil {
		if st, err := c.session.stats(); err == nil {
			s.session = st
		}
	}
	return s
}

func (c *etwCollector) setOnFatal(fn func(string)) {
	c.onFatal = fn
}

func (c *etwCollector) decodeLoop() {
	defer c.wg.Done()
	for {
		select {
		case <-c.stopCh:
			c.drainRawEvents()
			return
		case raw := <-c.rawCh:
			c.processRawEvent(&raw)
		}
	}
}

func (c *etwCollector) processRawEvent(raw *rawEvent) {
	rec := rawToEventRecord(raw)
	batch := c.decoder.decode(rec, raw.userData)
	for i := 0; i < batch.len(); i++ {
		c.agg.onEvent(batch.at(i))
	}
	c.decoded.Add(1)
	releaseEventBuffer(&c.bufPool, raw.userData)
}

// drainRawEvents processes events already copied by the ETW callback. stop()
// stops ProcessTrace and waits for its producer before closing stopCh, so the
// channel cannot receive new events while this drain is running.
func (c *etwCollector) drainRawEvents() {
	for {
		select {
		case raw := <-c.rawCh:
			c.processRawEvent(&raw)
		default:
			return
		}
	}
}

func (c *etwCollector) refreshListenSnapshots() {
	defer c.wg.Done()
	udpTick := time.NewTicker(10 * time.Second)
	tcpTick := time.NewTicker(30 * time.Second)
	pruneTick := time.NewTicker(tcbPruneInterval)
	defer udpTick.Stop()
	defer tcpTick.Stop()
	defer pruneTick.Stop()
	for {
		select {
		case <-c.stopCh:
			return
		case <-udpTick.C:
			c.udpSrv.refresh("udp")
		case <-tcpTick.C:
			c.tcpSrv.refresh("tcp")
		case now := <-pruneTick.C:
			c.decoder.pruneTCBs(now)
		}
	}
}

// etwEventRecordCB is the ProcessTrace callback trampoline. It must stay alive
// for the lifetime of the process. The UserContext points at a handlerBox
// owned by the session that started the trace.
var etwEventRecordCB = syscall.NewCallback(etwEventRecordCallback)

func etwEventRecordCallback(recPtr uintptr) uintptr {
	// uintptr -> pointer round-trip via the argument address is the documented
	// pattern for syscall.NewCallback trampolines.
	rec := *(**eventRecord)(unsafe.Pointer(&recPtr))
	if rec == nil || rec.UserContext == nil {
		return 0
	}
	box := (*handlerBox)(rec.UserContext)
	if box == nil || box.h == nil {
		return 0
	}
	box.h.handleRecord(rec)
	return 0
}

func (c *etwCollector) handleRecord(rec *eventRecord) {
	if rec.EventHeader.ProviderID != tcpipProviderGUID {
		return
	}
	id := rec.EventHeader.Descriptor.ID
	if _, ok := wantedEventSet[id]; !ok {
		return
	}
	if rec.UserData == nil || rec.UserDataLength == 0 {
		return
	}

	raw := rawEvent{
		eventID:   id,
		version:   rec.EventHeader.Descriptor.Version,
		timestamp: rec.EventHeader.TimeStamp,
	}

	raw.userData = acquireEventBuffer(&c.bufPool, int(rec.UserDataLength))
	copy(raw.userData, unsafe.Slice((*byte)(rec.UserData), rec.UserDataLength))

	select {
	case c.rawCh <- raw:
	default:
		c.dropCount.Add(1)
		releaseEventBuffer(&c.bufPool, raw.userData)
	}
}

func acquireEventBuffer(pool *sync.Pool, size int) []byte {
	if size <= pooledEventBufferSize {
		return pool.Get().(*eventBuffer)[:size]
	}
	return make([]byte, size)
}

func releaseEventBuffer(pool *sync.Pool, buf []byte) {
	if cap(buf) == pooledEventBufferSize {
		pool.Put((*eventBuffer)(buf[:pooledEventBufferSize]))
	}
}

type eventBuffer [pooledEventBufferSize]byte

// rawToEventRecord rebuilds a minimal EVENT_RECORD from the copied data; TDH
// only needs the descriptor (provider/id/version) and UserData.
func rawToEventRecord(raw *rawEvent) *eventRecord {
	rec := &eventRecord{}
	rec.EventHeader.ProviderID = tcpipProviderGUID
	rec.EventHeader.Descriptor.ID = raw.eventID
	rec.EventHeader.Descriptor.Version = raw.version
	rec.EventHeader.TimeStamp = raw.timestamp
	rec.UserDataLength = uint16(len(raw.userData))
	if len(raw.userData) > 0 {
		rec.UserData = unsafe.Pointer(&raw.userData[0])
	}
	return rec
}

func (d *etwDecoder) decode(rec *eventRecord, data []byte) flowEventBatch {
	if rec == nil {
		return flowEventBatch{}
	}
	version := rec.EventHeader.Descriptor.Version
	switch rec.EventHeader.Descriptor.ID {
	case 1017:
		return d.onTCPAccept(rec)
	case 1033:
		return d.onTCPConnect(rec)
	case 1034, 1045:
		return d.onTCPConnectFail(rec)
	case 1040, 1043:
		return d.onTCPClose(rec)
	case 1184:
		return d.onTCPRST(rec)
	case 1051:
		if version == 0 {
			return d.onTCPStateChangeFast(rec, data)
		}
		return d.onTCPStateChangeTDH(rec)
	case 1073:
		if version == 0 {
			return d.onTCPSendLegacyFast(rec, data)
		}
		return d.onTCPSendLegacyTDH(rec)
	case 1332:
		if version <= 5 {
			return d.onTCPSendFast(rec, data)
		}
		return d.onTCPSendTDH(rec)
	case 1074:
		if version <= 1 {
			return d.onTCPRecvFast(rec, data)
		}
		return d.onTCPRecvTDH(rec)
	case 1341:
		if version == 0 {
			return d.onRTTFast(rec, data)
		}
		return d.onRTTTDH(rec)
	case 1351:
		if version <= 1 {
			return d.onRetransmitFast(rec, data)
		}
		return d.onRetransmitTDH(rec)
	case 1169:
		if version <= 1 {
			return d.onUDPFast(rec, data, evUDPDataSend)
		}
		return d.onUDPTDH(rec, evUDPDataSend)
	case 1170:
		if version <= 1 {
			return d.onUDPFast(rec, data, evUDPDataRecv)
		}
		return d.onUDPTDH(rec, evUDPDataRecv)
	case 1300:
		return d.onTCPRundown(rec)
	default:
		return flowEventBatch{}
	}
}

// bumpParseErr records a malformed/unexpected event so operators can observe
// decode health. It is safe to call from any decode worker.
func (d *etwDecoder) bumpParseErr() {
	if d.parseErrors != nil {
		d.parseErrors.Add(1)
	}
}

// Fast path decoders below read the user data directly from the known TCPIP
// template layouts instead of going through TDH (which costs two syscalls per
// property). Data events dominate the event stream, so this removes nearly all
// syscall overhead from the hot path.

func (d *etwDecoder) onTCPSendFast(rec *eventRecord, data []byte) flowEventBatch {
	tcb, ok := u64At(data, 0)
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	bytes, ok := u32At(data, 16) // BytesSent
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	srtt, _ := u32At(data, 24) // SRtt
	rttVar, _ := u32At(data, 28)

	info, burst, hasBurst := d.lookupOrRegisterTcbFast(rec.EventHeader.Descriptor.Version, data, tcb)
	if info == nil {
		d.accumulatePending(tcb, uint64(bytes), 0, 0, 0, uint64(srtt), uint64(rttVar), 0)
		return flowEventBatch{}
	}
	return d.dataEvents(rec, evTCPDataSend, info, uint64(bytes), 0, uint64(srtt), uint64(rttVar), 0, burst, hasBurst)
}

func (d *etwDecoder) onTCPSendLegacyFast(rec *eventRecord, data []byte) flowEventBatch {
	tcb, ok := u64At(data, 0)
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	bytes, ok := u32At(data, 20) // NumBytes
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	srtt, _ := u32At(data, 36) // SRTT
	if srtt == 0 {
		srtt, _ = u32At(data, 16) // RttSample
	}

	info, burst, hasBurst := d.lookupOrRegisterTcbFast(rec.EventHeader.Descriptor.Version, data, tcb)
	if info == nil {
		d.accumulatePending(tcb, uint64(bytes), 0, 0, 0, uint64(srtt), 0, 0)
		return flowEventBatch{}
	}
	return d.dataEvents(rec, evTCPDataSend, info, uint64(bytes), 0, uint64(srtt), 0, 0, burst, hasBurst)
}

func (d *etwDecoder) onTCPRecvFast(rec *eventRecord, data []byte) flowEventBatch {
	tcb, ok := u64At(data, 0)
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	bytes, ok := u32At(data, 8) // NumBytes
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	packets, _ := u32At(data, 16) // NumPkt, version >= 1

	info := d.lookupTcb(tcb)
	if info == nil {
		d.accumulatePending(tcb, 0, uint64(bytes), 0, uint64(packets), 0, 0, 0)
		return flowEventBatch{}
	}
	return d.dataEvents(rec, evTCPDataRecv, info, uint64(bytes), uint64(packets), 0, 0, 0, flowEvent{}, false)
}

func (d *etwDecoder) onRTTFast(rec *eventRecord, data []byte) flowEventBatch {
	tcb, ok := u64At(data, 0)
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	rttVar, _ := u32At(data, 12)
	srtt, _ := u32At(data, 16) // SRTT
	if srtt == 0 {
		srtt, _ = u32At(data, 8) // RttSample
	}

	info := d.lookupTcb(tcb)
	if info == nil {
		d.accumulatePending(tcb, 0, 0, 0, 0, uint64(srtt), uint64(rttVar), 0)
		return flowEventBatch{}
	}
	return d.dataEvents(rec, evRTT, info, 0, 0, uint64(srtt), uint64(rttVar), 0, flowEvent{}, false)
}

func (d *etwDecoder) onRetransmitFast(rec *eventRecord, data []byte) flowEventBatch {
	tcb, ok := u64At(data, 0)
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	retrans, _ := u32At(data, 12) // RexmitCount

	info := d.lookupTcb(tcb)
	if info == nil {
		d.accumulatePending(tcb, 0, 0, 0, 0, 0, 0, uint64(retrans))
		return flowEventBatch{}
	}
	return d.dataEvents(rec, evTCPRetransmit, info, 0, 0, 0, 0, uint64(retrans), flowEvent{}, false)
}

func (d *etwDecoder) onTCPStateChangeFast(rec *eventRecord, data []byte) flowEventBatch {
	oldState, _ := u32At(data, 0)
	newState, _ := u32At(data, 4)
	tcb, ok := u64At(data, 12)
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}

	info := d.lookupTcb(tcb)
	if info == nil {
		return flowEventBatch{}
	}
	return oneFlowEvent(flowEvent{
		ts:        filetimeToTime(rec.EventHeader.TimeStamp),
		kind:      evTCPStateChange,
		transport: "tcp",
		family:    info.family,
		direction: info.direction,
		pid:       info.pid,
		srcIP:     info.srcIP,
		srcPort:   info.srcPort,
		dstIP:     info.dstIP,
		dstPort:   info.dstPort,
		oldState:  oldState,
		newState:  newState,
	})
}

func (d *etwDecoder) onUDPFast(rec *eventRecord, data []byte, kind eventKind) flowEventBatch {
	messages, _ := u32At(data, 8)
	bytes, _ := u32At(data, 12)
	localLen, ok := u32At(data, 16)
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	if localLen != 16 && localLen != 28 {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	localOff := 20
	if int(localLen) > len(data)-localOff {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	remoteLenOff := localOff + int(localLen)
	remoteLen, ok := u32At(data, remoteLenOff)
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	if remoteLen != 16 && remoteLen != 28 {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	remoteOff := remoteLenOff + 4
	if int(remoteLen) > len(data)-remoteOff {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	pidOff := remoteOff + int(remoteLen)
	pid, _ := u32At(data, pidOff)

	local, ok := parseSockAddr(data[localOff : localOff+int(localLen)])
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	remote, ok := parseSockAddr(data[remoteOff : remoteOff+int(remoteLen)])
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}

	direction := directionOutgoing
	if d.udpSrv != nil && d.udpSrv.isIncoming(local.ip, local.port, true) {
		direction = directionIncoming
	}

	return oneFlowEvent(flowEvent{
		ts:        filetimeToTime(rec.EventHeader.TimeStamp),
		kind:      kind,
		transport: "udp",
		family:    local.family,
		direction: direction,
		pid:       pid,
		srcIP:     local.ip,
		srcPort:   local.port,
		dstIP:     remote.ip,
		dstPort:   remote.port,
		bytes:     uint64(bytes),
		packets:   uint64(messages),
	})
}

// TDH fallbacks for unknown event versions. Slower than the fast path but
// correct regardless of layout changes, so future Windows versions degrade
// gracefully instead of misparsing.

func (d *etwDecoder) onTCPSendTDH(rec *eventRecord) flowEventBatch {
	tcb, ok := propU64(rec, "Tcb")
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	bytes, _ := propU32(rec, "BytesSent")
	srtt, _ := propU32(rec, "SRtt")
	rttVar, _ := propU32(rec, "RttVar")

	info, burst, hasBurst := d.lookupOrRegisterTcbTDH(rec, tcb)
	if info == nil {
		d.accumulatePending(tcb, uint64(bytes), 0, 0, 0, uint64(srtt), uint64(rttVar), 0)
		return flowEventBatch{}
	}
	return d.dataEvents(rec, evTCPDataSend, info, uint64(bytes), 0, uint64(srtt), uint64(rttVar), 0, burst, hasBurst)
}

func (d *etwDecoder) onTCPSendLegacyTDH(rec *eventRecord) flowEventBatch {
	tcb, ok := propU64(rec, "Tcb")
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	bytes, _ := propU32(rec, "NumBytes")
	srtt, _ := propU32(rec, "SRTT")
	if srtt == 0 {
		srtt, _ = propU32(rec, "RttSample")
	}

	info, burst, hasBurst := d.lookupOrRegisterTcbTDH(rec, tcb)
	if info == nil {
		d.accumulatePending(tcb, uint64(bytes), 0, 0, 0, uint64(srtt), 0, 0)
		return flowEventBatch{}
	}
	return d.dataEvents(rec, evTCPDataSend, info, uint64(bytes), 0, uint64(srtt), 0, 0, burst, hasBurst)
}

func (d *etwDecoder) onTCPRecvTDH(rec *eventRecord) flowEventBatch {
	tcb, ok := propU64(rec, "Tcb")
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	bytes, _ := propU32(rec, "NumBytes")
	packets, _ := propU32(rec, "NumPkt")

	info := d.lookupTcb(tcb)
	if info == nil {
		d.accumulatePending(tcb, 0, uint64(bytes), 0, uint64(packets), 0, 0, 0)
		return flowEventBatch{}
	}
	return d.dataEvents(rec, evTCPDataRecv, info, uint64(bytes), uint64(packets), 0, 0, 0, flowEvent{}, false)
}

func (d *etwDecoder) onRTTTDH(rec *eventRecord) flowEventBatch {
	tcb, ok := propU64(rec, "Tcb")
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	srtt, _ := propU32(rec, "SRTT")
	if srtt == 0 {
		srtt, _ = propU32(rec, "RttSample")
	}
	rttVar, _ := propU32(rec, "RttVar")

	info := d.lookupTcb(tcb)
	if info == nil {
		d.accumulatePending(tcb, 0, 0, 0, 0, uint64(srtt), uint64(rttVar), 0)
		return flowEventBatch{}
	}
	return d.dataEvents(rec, evRTT, info, 0, 0, uint64(srtt), uint64(rttVar), 0, flowEvent{}, false)
}

func (d *etwDecoder) onRetransmitTDH(rec *eventRecord) flowEventBatch {
	tcb, ok := propU64(rec, "Tcb")
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	retrans, _ := propU32(rec, "RexmitCount")

	info := d.lookupTcb(tcb)
	if info == nil {
		d.accumulatePending(tcb, 0, 0, 0, 0, 0, 0, uint64(retrans))
		return flowEventBatch{}
	}
	return d.dataEvents(rec, evTCPRetransmit, info, 0, 0, 0, 0, uint64(retrans), flowEvent{}, false)
}

func (d *etwDecoder) onTCPStateChangeTDH(rec *eventRecord) flowEventBatch {
	tcb, ok := propU64(rec, "Tcb")
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	oldState, _ := propU32(rec, "OldState")
	newState, _ := propU32(rec, "NewState")

	info := d.lookupTcb(tcb)
	if info == nil {
		return flowEventBatch{}
	}
	return oneFlowEvent(flowEvent{
		ts:        filetimeToTime(rec.EventHeader.TimeStamp),
		kind:      evTCPStateChange,
		transport: "tcp",
		family:    info.family,
		direction: info.direction,
		pid:       info.pid,
		srcIP:     info.srcIP,
		srcPort:   info.srcPort,
		dstIP:     info.dstIP,
		dstPort:   info.dstPort,
		oldState:  oldState,
		newState:  newState,
	})
}

func (d *etwDecoder) onUDPTDH(rec *eventRecord, kind eventKind) flowEventBatch {
	bytes, _ := propU32(rec, "NumBytes")
	messages, _ := propU32(rec, "NumMessages")
	pid, _ := propU32(rec, "Pid")

	local, ok := propSockAddr(rec, "LocalSockAddr")
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	remote, ok := propSockAddr(rec, "RemoteSockAddr")
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}

	direction := directionOutgoing
	if d.udpSrv != nil && d.udpSrv.isIncoming(local.ip, local.port, true) {
		direction = directionIncoming
	}

	return oneFlowEvent(flowEvent{
		ts:        filetimeToTime(rec.EventHeader.TimeStamp),
		kind:      kind,
		transport: "udp",
		family:    local.family,
		direction: direction,
		pid:       pid,
		srcIP:     local.ip,
		srcPort:   local.port,
		dstIP:     remote.ip,
		dstPort:   remote.port,
		bytes:     uint64(bytes),
		packets:   uint64(messages),
	})
}

// lookupOrRegisterTcbTDH mirrors lookupOrRegisterTcbFast but resolves the tuple
// through TDH, used by the fallback decoders.
func (d *etwDecoder) lookupOrRegisterTcbTDH(rec *eventRecord, tcb uint64) (*tcbInfo, flowEvent, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if info, ok := d.tcbs[tcb]; ok {
		info.lastSeen = time.Now()
		return info, flowEvent{}, false
	}
	info := d.tupleFromRecLocked(rec, tcb, 0)
	if info == nil {
		return nil, flowEvent{}, false
	}
	info.direction = d.inferTCPDirection(info.srcIP, info.srcPort)
	d.rememberTCBLocked(tcb, info, time.Now())
	burst, ok := d.replayPendingLocked(tcb, info, time.Now())
	return info, burst, ok
}

// lookupOrRegisterTcbFast mirrors lookupOrRegisterTcb but parses the inline
// addresses from the raw data (1332 version 5+) instead of using TDH.
func (d *etwDecoder) lookupOrRegisterTcbFast(version uint8, data []byte, tcb uint64) (*tcbInfo, flowEvent, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if info, ok := d.tcbs[tcb]; ok {
		info.lastSeen = time.Now()
		return info, flowEvent{}, false
	}
	if version < 5 {
		return nil, flowEvent{}, false
	}

	localLen, ok := u32At(data, 72)
	if !ok {
		d.bumpParseErr()
		return nil, flowEvent{}, false
	}
	if localLen != 16 && localLen != 28 {
		d.bumpParseErr()
		return nil, flowEvent{}, false
	}
	localOff := 76
	if int(localLen) > len(data)-localOff {
		d.bumpParseErr()
		return nil, flowEvent{}, false
	}
	remoteLenOff := localOff + int(localLen)
	remoteLen, ok := u32At(data, remoteLenOff)
	if !ok {
		d.bumpParseErr()
		return nil, flowEvent{}, false
	}
	if remoteLen != 16 && remoteLen != 28 {
		d.bumpParseErr()
		return nil, flowEvent{}, false
	}
	remoteOff := remoteLenOff + 4
	if int(remoteLen) > len(data)-remoteOff {
		d.bumpParseErr()
		return nil, flowEvent{}, false
	}

	local, ok := parseSockAddr(data[localOff : localOff+int(localLen)])
	if !ok {
		d.bumpParseErr()
		return nil, flowEvent{}, false
	}
	remote, ok := parseSockAddr(data[remoteOff : remoteOff+int(remoteLen)])
	if !ok {
		d.bumpParseErr()
		return nil, flowEvent{}, false
	}

	info := &tcbInfo{
		srcIP:     local.ip,
		srcPort:   local.port,
		dstIP:     remote.ip,
		dstPort:   remote.port,
		family:    local.family,
		direction: d.inferTCPDirection(local.ip, local.port),
	}
	d.rememberTCBLocked(tcb, info, time.Now())
	burst, ok := d.replayPendingLocked(tcb, info, time.Now())
	return info, burst, ok
}

func (d *etwDecoder) onTCPConnect(rec *eventRecord) flowEventBatch {
	return d.onTCPEstablished(rec, directionOutgoing)
}

// onTCPAccept handles TcpAcceptTcbComplete (event 1017), the passive-side
// counterpart of TcpConnectTcbComplete. Without it, server connections can
// carry bytes but miss their establishment count because the ESTABLISHED state
// transition usually precedes the first data event that can reveal the tuple.
func (d *etwDecoder) onTCPAccept(rec *eventRecord) flowEventBatch {
	return d.onTCPEstablished(rec, directionIncoming)
}

func (d *etwDecoder) onTCPEstablished(rec *eventRecord, direction string) flowEventBatch {
	tcb, ok := propU64(rec, "Tcb")
	if !ok {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	pid, _ := propU32(rec, "ProcessId")
	info := d.tupleFromRec(rec, tcb, pid)
	if info == nil {
		d.bumpParseErr()
		return flowEventBatch{}
	}
	info.direction = direction
	d.mu.Lock()
	d.rememberTCBLocked(tcb, info, time.Now())
	burst, hasBurst := d.replayPendingLocked(tcb, info, filetimeToTime(rec.EventHeader.TimeStamp))
	d.mu.Unlock()

	var batch flowEventBatch
	if hasBurst {
		batch.append(burst)
	}
	batch.append(flowEvent{
		ts:        filetimeToTime(rec.EventHeader.TimeStamp),
		kind:      evTCPEstablished,
		transport: "tcp",
		family:    info.family,
		direction: info.direction,
		pid:       info.pid,
		srcIP:     info.srcIP,
		srcPort:   info.srcPort,
		dstIP:     info.dstIP,
		dstPort:   info.dstPort,
	})
	return batch
}

func (d *etwDecoder) onTCPConnectFail(rec *eventRecord) flowEventBatch {
	pid, _ := propU32(rec, "ProcessId")
	info := d.tupleFromRec(rec, 0, pid)
	if info == nil {
		return flowEventBatch{}
	}
	info.direction = directionOutgoing
	return oneFlowEvent(flowEvent{
		ts:        filetimeToTime(rec.EventHeader.TimeStamp),
		kind:      evTCPConnectFail,
		transport: "tcp",
		family:    info.family,
		direction: info.direction,
		pid:       info.pid,
		srcIP:     info.srcIP,
		srcPort:   info.srcPort,
		dstIP:     info.dstIP,
		dstPort:   info.dstPort,
	})
}

func (d *etwDecoder) onTCPClose(rec *eventRecord) flowEventBatch {
	tcb, _ := propU64(rec, "Tcb")
	pid, _ := propU32(rec, "ProcessId")
	info := d.tupleFromRec(rec, tcb, pid)
	if info == nil {
		return flowEventBatch{}
	}
	d.mu.Lock()
	burst, hasBurst := d.replayPendingLocked(tcb, info, filetimeToTime(rec.EventHeader.TimeStamp))
	delete(d.tcbs, tcb)
	d.mu.Unlock()

	var batch flowEventBatch
	if hasBurst {
		batch.append(burst)
	}
	batch.append(flowEvent{
		ts:        filetimeToTime(rec.EventHeader.TimeStamp),
		kind:      evTCPClosed,
		transport: "tcp",
		family:    info.family,
		direction: info.direction,
		pid:       info.pid,
		srcIP:     info.srcIP,
		srcPort:   info.srcPort,
		dstIP:     info.dstIP,
		dstPort:   info.dstPort,
	})
	return batch
}

func (d *etwDecoder) onTCPRST(rec *eventRecord) flowEventBatch {
	tcb, _ := propU64(rec, "Tcb")
	info := d.tupleFromRec(rec, tcb, 0)
	if info == nil {
		return flowEventBatch{}
	}
	d.mu.Lock()
	if known, ok := d.tcbs[tcb]; ok && known.pid != 0 {
		known.lastSeen = time.Now()
		info.pid = known.pid
	}
	burst, hasBurst := d.replayPendingLocked(tcb, info, filetimeToTime(rec.EventHeader.TimeStamp))
	delete(d.tcbs, tcb)
	d.mu.Unlock()

	var batch flowEventBatch
	if hasBurst {
		batch.append(burst)
	}
	batch.append(flowEvent{
		ts:        filetimeToTime(rec.EventHeader.TimeStamp),
		kind:      evTCPClosed,
		transport: "tcp",
		family:    info.family,
		direction: info.direction,
		pid:       info.pid,
		srcIP:     info.srcIP,
		srcPort:   info.srcPort,
		dstIP:     info.dstIP,
		dstPort:   info.dstPort,
	})
	return batch
}

func (d *etwDecoder) onTCPRundown(rec *eventRecord) flowEventBatch {
	tcb, ok := propU64(rec, "Tcb")
	if !ok {
		return flowEventBatch{}
	}
	pid, _ := propU32(rec, "Pid")
	state, _ := propU32(rec, "State")

	info := d.tupleFromRec(rec, tcb, pid)
	if info == nil {
		return flowEventBatch{}
	}
	// Listening sockets are not flows.
	if state == tcpStateListen {
		return flowEventBatch{}
	}
	info.direction = d.inferTCPDirection(info.srcIP, info.srcPort)

	d.mu.Lock()
	d.rememberTCBLocked(tcb, info, time.Now())
	burst, hasBurst := d.replayPendingLocked(tcb, info, filetimeToTime(rec.EventHeader.TimeStamp))
	d.mu.Unlock()

	var batch flowEventBatch
	if hasBurst {
		batch.append(burst)
	}
	// Rundown describes connections that already existed when the trace
	// started. Register them for subsequent data attribution, but do not count
	// an establishment in the current reporting interval.
	return batch
}

func (d *etwDecoder) inferTCPDirection(localIP string, localPort uint32) string {
	if d.tcpSrv != nil && d.tcpSrv.isIncoming(localIP, localPort, false) {
		return directionIncoming
	}
	return directionOutgoing
}

func (d *etwDecoder) dataEvents(rec *eventRecord, kind eventKind, info *tcbInfo,
	bytes, packets, rtt, rttVar, retrans uint64, burst flowEvent, hasBurst bool) flowEventBatch {
	var batch flowEventBatch
	if hasBurst {
		batch.append(burst)
	}
	batch.append(flowEvent{
		ts:        filetimeToTime(rec.EventHeader.TimeStamp),
		kind:      kind,
		transport: "tcp",
		family:    info.family,
		direction: info.direction,
		pid:       info.pid,
		srcIP:     info.srcIP,
		srcPort:   info.srcPort,
		dstIP:     info.dstIP,
		dstPort:   info.dstPort,
		bytes:     bytes,
		packets:   packets,
		rtt:       rtt,
		rttVar:    rttVar,
		retrans:   retrans,
	})
	return batch
}

func (d *etwDecoder) lookupTcb(tcb uint64) *tcbInfo {
	d.mu.Lock()
	defer d.mu.Unlock()
	info := d.tcbs[tcb]
	if info != nil {
		info.lastSeen = time.Now()
	}
	return info
}

// rememberTCBLocked inserts or refreshes a TCB while keeping the correlation
// table bounded. d.mu must be held by the caller.
func (d *etwDecoder) rememberTCBLocked(tcb uint64, info *tcbInfo, now time.Time) {
	if tcb == 0 || info == nil {
		return
	}
	if _, exists := d.tcbs[tcb]; !exists && d.tcbLimit > 0 && len(d.tcbs) >= d.tcbLimit {
		for candidate := range d.tcbs {
			delete(d.tcbs, candidate)
			break
		}
		d.tcbEvicted++
	}
	info.lastSeen = now
	d.tcbs[tcb] = info
}

func (d *etwDecoder) pruneTCBs(now time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.tcbTTL > 0 {
		for tcb, info := range d.tcbs {
			if now.Sub(info.lastSeen) >= d.tcbTTL {
				delete(d.tcbs, tcb)
				d.tcbEvicted++
			}
		}
	}
	if d.pendingTTL > 0 {
		for tcb, pending := range d.pending {
			if now.Sub(pending.lastSeen) >= d.pendingTTL {
				delete(d.pending, tcb)
				d.pendingEvicted++
			}
		}
	}
}

// accumulatePending buffers deltas for a TCB whose tuple is not known yet.
func (d *etwDecoder) accumulatePending(tcb uint64,
	bytesSent, bytesRecv, packetsSent, packetsRecv, rtt, rttVar, retrans uint64) {
	if tcb == 0 {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	p := d.pending[tcb]
	if p == nil {
		p = &pendingStats{}
		d.pending[tcb] = p
	}
	p.lastSeen = time.Now()
	if bytesSent > 0 {
		p.bytesSent += bytesSent
	}
	if bytesRecv > 0 {
		p.bytesRecv += bytesRecv
	}
	if packetsSent > 0 {
		p.packetsSent += packetsSent
	}
	if packetsRecv > 0 {
		p.packetsRecv += packetsRecv
	}
	if retrans > 0 {
		p.retrans += retrans
	}
	if rtt > 0 {
		p.rttSum += rtt
		p.rttVarSum += rttVar
		p.rttCount++
	}
	// Bound memory: if tuples never resolve (e.g. dropped tuple events), evict
	// one buffered TCB so the map cannot grow without limit.
	if d.pendingLimit > 0 && len(d.pending) > d.pendingLimit {
		for tcb := range d.pending {
			delete(d.pending, tcb)
			d.pendingEvicted++
			break
		}
	}
}

// replayPendingLocked converts buffered pending deltas into a burst event and
// clears the buffer. d.mu must be held by the caller.
func (d *etwDecoder) replayPendingLocked(tcb uint64, info *tcbInfo, ts time.Time) (flowEvent, bool) {
	p, ok := d.pending[tcb]
	if !ok {
		return flowEvent{}, false
	}
	delete(d.pending, tcb)
	if p.bytesSent == 0 && p.bytesRecv == 0 && p.retrans == 0 && p.rttCount == 0 {
		return flowEvent{}, false
	}
	return flowEvent{
		ts:             ts,
		kind:           evBurstReplay,
		transport:      "tcp",
		family:         info.family,
		direction:      info.direction,
		pid:            info.pid,
		srcIP:          info.srcIP,
		srcPort:        info.srcPort,
		dstIP:          info.dstIP,
		dstPort:        info.dstPort,
		cumBytesSend:   p.bytesSent,
		cumBytesRecv:   p.bytesRecv,
		cumPacketsSend: p.packetsSent,
		cumPacketsRecv: p.packetsRecv,
		cumRetrans:     p.retrans,
		cumRTT:         p.rttSum,
		cumRTTVar:      p.rttVarSum,
		cumRTTCount:    p.rttCount,
	}, true
}

// tupleFromRec builds tcbInfo from the event's LocalAddress/RemoteAddress
// properties. pid is used when the event carries one; otherwise a previously
// known TCB entry wins.
func (d *etwDecoder) tupleFromRec(rec *eventRecord, tcb uint64, pid uint32) *tcbInfo {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.tupleFromRecLocked(rec, tcb, pid)
}

// tupleFromRecLocked requires d.mu to be held by the caller.
func (d *etwDecoder) tupleFromRecLocked(rec *eventRecord, tcb uint64, pid uint32) *tcbInfo {
	local, ok := propSockAddr(rec, "LocalAddress")
	if !ok {
		return nil
	}
	remote, ok := propSockAddr(rec, "RemoteAddress")
	if !ok {
		return nil
	}

	info := &tcbInfo{
		srcIP:     local.ip,
		srcPort:   local.port,
		dstIP:     remote.ip,
		dstPort:   remote.port,
		family:    local.family,
		direction: d.inferTCPDirection(local.ip, local.port),
	}
	if known, ok := d.tcbs[tcb]; ok {
		known.lastSeen = time.Now()
		info.pid = known.pid
		info.direction = known.direction
	}
	if pid != 0 {
		info.pid = pid
	}
	return info
}

// ---------------------------------------------------------------------------
// ETW session
// ---------------------------------------------------------------------------

type etwSession struct {
	name           string
	handle         uintptr
	traceHandle    uintptr
	ownerSemaphore uintptr
	rawProps       []byte
	nameBytes      []uint16
	ctxBox         *handlerBox
}

// start creates a real-time ETW session, enables the given manifest provider
// (with kernel-side event ID filtering), and opens the trace for callback
// delivery to handler.
func (s *etwSession) start(cfg etwConfig, provider *guid, level uint8, matchAny uint64,
	eventIDs []uint16, handler recordHandler) error {
	name := utf16.Encode([]rune(s.name + "\x00"))
	s.nameBytes = name
	s.ctxBox = &handlerBox{h: handler}
	semaphoreName := utf16.Encode([]rune("Global\\" + s.name + "-owner\x00"))
	ownerSemaphore, err := acquireSessionSemaphore(&semaphoreName[0])
	if err != nil {
		return fmt.Errorf("acquire ETW session ownership: %w", err)
	}
	s.ownerSemaphore = ownerSemaphore
	started := false
	defer func() {
		if !started {
			releaseSessionSemaphore(s.ownerSemaphore)
			s.ownerSemaphore = 0
		}
	}()

	newProps := func() ([]byte, *eventTraceProperties) {
		// Leave extra room so ControlTrace can write back session properties
		// (including the log-file-name slot) without ERROR_MORE_DATA.
		propsSize := int(unsafe.Sizeof(eventTraceProperties{})) + len(name)*2 + 1024
		raw := make([]byte, propsSize)
		props := (*eventTraceProperties)(unsafe.Pointer(&raw[0]))
		props.Wnode.BufferSize = uint32(propsSize)
		props.Wnode.Flags = wnodeFlagTracedGUID
		props.Wnode.ClientContext = 1 // QPC clock
		props.LogFileMode = eventTraceRealTimeMode
		props.BufferSize = cfg.bufferSizeKB
		props.MinimumBuffers = cfg.minBuffers
		props.MaximumBuffers = cfg.maxBuffers
		props.FlushTimer = 1
		props.LoggerNameOffset = uint32(unsafe.Sizeof(eventTraceProperties{}))
		copy(raw[props.LoggerNameOffset:], unsafe.Slice((*byte)(unsafe.Pointer(&name[0])), len(name)*2))
		return raw, props
	}
	raw, props := newProps()

	var handle uintptr
	if err = startTrace(&handle, &name[0], props); err != nil {
		if err == syscall.Errno(errorAlreadyExists) {
			// The ownership semaphore proves no live collector with this session name
			// exists, so the ETW session is an orphan from an unclean shutdown.
			_ = controlTrace(0, &name[0], props, eventTraceControlStop)
			raw, props = newProps() // ControlTrace writes into the properties buffer.
			if err = startTrace(&handle, &name[0], props); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	s.rawProps = raw
	s.handle = handle

	filter := &eventFilterEventID{FilterIn: 1, Count: uint16(len(eventIDs))}
	copy(filter.Events[:], eventIDs)
	fd := eventFilterDescriptor{
		Ptr:  uint64(uintptr(unsafe.Pointer(filter))),
		Size: uint32(unsafe.Sizeof(*filter)),
		Type: eventFilterTypeEventID,
	}
	params := enableTraceParameters{
		Version:          2,
		EnableFilterDesc: &fd,
		FilterDescCount:  1,
	}
	if err := enableTraceEx2(handle, provider, eventControlCodeEnableProvider,
		level, matchAny, 0, 0, &params); err != nil {
		_ = controlTrace(handle, &name[0], props, eventTraceControlStop)
		return err
	}

	logfile := &eventTraceLogfile{
		LoggerName:          &name[0],
		ProcessTraceMode:    processTraceModeRealTime | processTraceModeEventRecord,
		EventRecordCallback: etwEventRecordCB,
		Context:             uintptr(unsafe.Pointer(s.ctxBox)),
	}
	traceHandle, err := openTrace(logfile)
	if err != nil {
		_ = controlTrace(handle, &name[0], props, eventTraceControlStop)
		return err
	}
	s.traceHandle = traceHandle
	started = true
	return nil
}

func (s *etwSession) stop() {
	if s.handle != 0 && len(s.nameBytes) > 0 {
		props := (*eventTraceProperties)(unsafe.Pointer(&s.rawProps[0]))
		if err := controlTrace(s.handle, &s.nameBytes[0], props, 0); err == nil {
			if props.EventsLost > 0 || props.RealTimeBuffersLost > 0 {
				l.Warnf("ETW session %s: events_lost=%d realtime_buffers_lost=%d",
					s.name, props.EventsLost, props.RealTimeBuffersLost)
			}
		}
	}
	if s.traceHandle != 0 {
		_ = closeTrace(s.traceHandle)
		s.traceHandle = 0
	}
	if s.handle != 0 && len(s.nameBytes) > 0 {
		props := (*eventTraceProperties)(unsafe.Pointer(&s.rawProps[0]))
		if err := controlTrace(s.handle, &s.nameBytes[0], props, eventTraceControlStop); err != nil {
			l.Warnf("stop ETW session %s failed: %v", s.name, err)
		}
		s.handle = 0
	}
	releaseSessionSemaphore(s.ownerSemaphore)
	s.ownerSemaphore = 0
}

// stats queries the live session counters via ControlTrace QUERY. A fresh
// properties buffer is used so the running session data is not disturbed.
func (s *etwSession) stats() (etwSessionStats, error) {
	var st etwSessionStats
	if s.handle == 0 {
		return st, fmt.Errorf("session not started")
	}

	propsSize := int(unsafe.Sizeof(eventTraceProperties{})) + len(s.nameBytes)*2 + 1024
	raw := make([]byte, propsSize)
	props := (*eventTraceProperties)(unsafe.Pointer(&raw[0]))
	props.Wnode.BufferSize = uint32(propsSize)
	props.LoggerNameOffset = uint32(unsafe.Sizeof(eventTraceProperties{}))
	copy(raw[props.LoggerNameOffset:], unsafe.Slice((*byte)(unsafe.Pointer(&s.nameBytes[0])), len(s.nameBytes)*2))

	if err := controlTrace(s.handle, &s.nameBytes[0], props, 0); err != nil { // EVENT_TRACE_CONTROL_QUERY
		return st, err
	}
	st.numberOfBuffers = props.NumberOfBuffers
	st.freeBuffers = props.FreeBuffers
	st.eventsLost = props.EventsLost
	st.realTimeBuffersLost = props.RealTimeBuffersLost
	return st, nil
}

// refresh rebuilds the set of bound UDP endpoints or TCP listeners.
func (s *listenSnapshot) refresh(network string) {
	conns, err := gopsnet.Connections(network)
	if err != nil {
		return
	}
	m := make(map[string]struct{}, len(conns))
	for _, c := range conns {
		switch network {
		case "udp":
			if c.Raddr.IP != "" || c.Raddr.Port != 0 {
				continue // connected (client) socket
			}
		case "tcp":
			if !strings.EqualFold(c.Status, "LISTEN") {
				continue
			}
		default:
			return
		}
		if c.Laddr.IP == "" && c.Laddr.Port == 0 {
			continue
		}
		key := net.JoinHostPort(c.Laddr.IP, strconv.Itoa(int(c.Laddr.Port)))
		m[key] = struct{}{}
	}
	s.mu.Lock()
	s.endpoints = m
	s.mu.Unlock()
}

func (s *listenSnapshot) isIncoming(ip string, port uint32, rejectEphemeral bool) bool {
	if port == 0 || (rejectEphemeral && port >= 49152) { // Windows default ephemeral range start
		return false
	}
	portStr := strconv.FormatUint(uint64(port), 10)
	keys := []string{net.JoinHostPort(ip, portStr)}
	if p := net.ParseIP(ip); p != nil && p.To4() != nil {
		keys = append(keys, "0.0.0.0:"+portStr)
	} else {
		keys = append(keys, "[::]:"+portStr)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, k := range keys {
		if _, ok := s.endpoints[k]; ok {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Property extraction (TDH)
// ---------------------------------------------------------------------------

func property(rec *eventRecord, name string) ([]byte, bool) {
	u16 := utf16.Encode([]rune(name + "\x00"))
	desc := propertyDataDescriptor{
		PropertyName: uint64(uintptr(unsafe.Pointer(&u16[0]))),
	}
	size, err := tdhGetPropertySize(rec, &desc)
	if err != nil || size == 0 || size > 4096 {
		return nil, false
	}
	buf := make([]byte, size)
	if err := tdhGetProperty(rec, &desc, buf); err != nil {
		return nil, false
	}
	return buf, true
}

func propU32(rec *eventRecord, name string) (uint32, bool) {
	b, ok := property(rec, name)
	if !ok || len(b) < 4 {
		return 0, false
	}
	return binary.LittleEndian.Uint32(b), true
}

func propU16(rec *eventRecord, name string) (uint16, bool) {
	b, ok := property(rec, name)
	if !ok || len(b) < 2 {
		return 0, false
	}
	return binary.LittleEndian.Uint16(b), true
}

func propU64(rec *eventRecord, name string) (uint64, bool) {
	b, ok := property(rec, name)
	if !ok || len(b) < 8 {
		return 0, false
	}
	return binary.LittleEndian.Uint64(b), true
}

// u32At reads a little-endian uint32 at a fixed offset in the raw event user
// data (fast path, no TDH involved).
func u32At(data []byte, off int) (uint32, bool) {
	if off < 0 || off+4 > len(data) {
		return 0, false
	}
	return binary.LittleEndian.Uint32(data[off:]), true
}

func u64At(data []byte, off int) (uint64, bool) {
	if off < 0 || off+8 > len(data) {
		return 0, false
	}
	return binary.LittleEndian.Uint64(data[off:]), true
}

type sockAddr struct {
	ip     string
	port   uint32
	family string
}

func propSockAddr(rec *eventRecord, name string) (sockAddr, bool) {
	b, ok := property(rec, name)
	if !ok {
		return sockAddr{}, false
	}
	return parseSockAddr(b)
}

func parseSockAddr(b []byte) (sockAddr, bool) {
	if len(b) < 8 {
		return sockAddr{}, false
	}
	switch binary.LittleEndian.Uint16(b[0:2]) {
	case 2: // AF_INET
		if len(b) < 8 {
			return sockAddr{}, false
		}
		return sockAddr{
			ip:     net.IP(b[4:8]).String(),
			port:   uint32(binary.BigEndian.Uint16(b[2:4])),
			family: "IPv4",
		}, true
	case 23: // AF_INET6
		if len(b) < 28 {
			return sockAddr{}, false
		}
		return sockAddr{
			ip:     net.IP(b[8:24]).String(),
			port:   uint32(binary.BigEndian.Uint16(b[2:4])),
			family: "IPv6",
		}, true
	default:
		return sockAddr{}, false
	}
}

func parseGUID(s string) guid {
	var g guid
	clean := strings.ReplaceAll(s, "{", "")
	clean = strings.ReplaceAll(clean, "}", "")
	parts := strings.Split(clean, "-")
	if len(parts) != 5 {
		return g
	}
	data1, _ := strconvUint(parts[0], 16)
	data2, _ := strconvUint(parts[1], 16)
	data3, _ := strconvUint(parts[2], 16)
	binary.LittleEndian.PutUint32(g[0:4], uint32(data1))
	binary.LittleEndian.PutUint16(g[4:6], uint16(data2))
	binary.LittleEndian.PutUint16(g[6:8], uint16(data3))
	data4 := parts[3] + parts[4]
	for i := 0; i < len(data4); i += 2 {
		if 8+i/2 >= len(g) {
			break
		}
		v, _ := strconvUint(data4[i:i+2], 16)
		g[8+i/2] = byte(v)
	}
	return g
}

func strconvUint(s string, base int) (uint64, error) {
	return strconv.ParseUint(s, base, 64)
}

// filetimeToTime converts a 100ns FILETIME (since 1601-01-01) to time.Time.
func filetimeToTime(ft int64) time.Time {
	const unixEpoch100ns = 116444736000000000
	return time.Unix(0, (ft-unixEpoch100ns)*100).UTC()
}

// ---------------------------------------------------------------------------
// Process name resolution
// ---------------------------------------------------------------------------

type cachedProcessName struct {
	name string
	ts   time.Time
}

var procNameCache = struct {
	sync.Mutex
	m map[uint32]cachedProcessName
}{m: make(map[uint32]cachedProcessName)}

const (
	procNameCacheTTL         = maxInterval + time.Minute
	procNameNegativeCacheTTL = 5 * time.Second
	procNameCacheLimit       = 4096
)

func lookupProcessName(pid uint32) string {
	if pid == 0 {
		return ""
	}
	now := time.Now()

	procNameCache.Lock()
	if c, ok := procNameCache.m[pid]; ok {
		ttl := procNameCacheTTL
		if c.name == "" {
			ttl = procNameNegativeCacheTTL
		}
		if now.Sub(c.ts) < ttl {
			name := c.name
			procNameCache.Unlock()
			return name
		}
		delete(procNameCache.m, pid)
	}
	procNameCache.Unlock()
	return refreshProcessNameCache(pid, now)
}

func refreshProcessNameCache(pid uint32, now time.Time) string {
	if pid == 0 {
		return ""
	}
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		cacheProcessName(pid, "", now)
		return ""
	}
	name, err := p.Name()
	if err != nil || name == "" {
		cacheProcessName(pid, "", now)
		return ""
	}

	cacheProcessName(pid, name, now)
	return name
}

func cacheProcessName(pid uint32, name string, now time.Time) {
	procNameCache.Lock()
	defer procNameCache.Unlock()
	if len(procNameCache.m) >= procNameCacheLimit {
		for cachedPID, cached := range procNameCache.m {
			if now.Sub(cached.ts) >= procNameCacheTTL {
				delete(procNameCache.m, cachedPID)
			}
		}
	}
	if len(procNameCache.m) >= procNameCacheLimit {
		var oldestPID uint32
		var oldest time.Time
		for cachedPID, cached := range procNameCache.m {
			if oldest.IsZero() || cached.ts.Before(oldest) {
				oldestPID = cachedPID
				oldest = cached.ts
			}
		}
		delete(procNameCache.m, oldestPID)
	}
	procNameCache.m[pid] = cachedProcessName{name: name, ts: now}
}

func init() { //nolint:gochecknoinits
	resolveProcessName = lookupProcessName
	refreshProcessName = func(pid uint32) string {
		return refreshProcessNameCache(pid, time.Now())
	}
}
