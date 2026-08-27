// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows

package winnetflow

import (
	"encoding/binary"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf16"
	"unsafe"
)

const httpSessionName = "datakit-winnetflow-http"

// Microsoft-Windows-HttpService provider GUID
// {dd5ef90a-6398-47a4-ad34-4dcecdef795f}
var httpProviderGUID = parseGUID("dd5ef90a-6398-47a4-ad34-4dcecdef795f")

// HttpService event keywords (from the OS manifest). 0x136 is the union of
// every keyword family we consume:
//   - 0x002  HTTP_KEYWORD_REQUEST  (events 1, 2, 3)
//   - 0x004  HTTP_KEYWORD_RESPONSE (events 4, 8, 11, 12, 16)
//   - 0x010  HTTP_KEYWORD_CONNECTION (events 21, 24)
//   - 0x020  HTTP_KEYWORD_CACHE (events 11, 16)
//   - 0x100  HTTP_KEYWORD_REQUEST_QUEUE (events 1, 3)
const httpMatchAnyKeywords = 0x136

const httpLevel = 4 // TRACE_LEVEL_INFORMATION

// HttpService event IDs the collector consumes. They are also passed to
// EnableTraceEx2 as an EVENT_FILTER_TYPE_EVENT_ID filter so everything else is
// dropped by the kernel before reaching us.
const (
	evHTTPConnConnect   = 21 // New connection (local/remote addresses, ActivityID)
	evHTTPConnCleanup   = 24 // Connection cleanup (always last)
	evHTTPRecvReq       = 1  // Request received (RelatedActivityID -> request scope)
	evHTTPParse         = 2  // Verb + URL
	evHTTPDeliver       = 3  // SiteId, request queue name (app pool), URL
	evHTTPRecvResp      = 4  // Response with status code + verb
	evHTTPFastResp      = 8  // Fast response with status code + verb
	evHTTPSendComplete  = 10 // Send complete (terminal)
	evHTTPCachedAndSend = 11 // Cached and send (terminal)
	evHTTPFastSend      = 12 // Fast send (terminal)
	evHTTPSrvdFrmCache  = 16 // Served from cache (terminal, BytesSent)
)

var httpEventIDs = []uint16{
	evHTTPConnConnect,
	evHTTPConnCleanup,
	evHTTPRecvReq,
	evHTTPParse,
	evHTTPDeliver,
	evHTTPRecvResp,
	evHTTPFastResp,
	evHTTPSendComplete,
	evHTTPCachedAndSend,
	evHTTPFastSend,
	evHTTPSrvdFrmCache,
}

var httpEventSet = func() map[uint16]struct{} {
	m := make(map[uint16]struct{}, len(httpEventIDs))
	for _, id := range httpEventIDs {
		m[id] = struct{}{}
	}
	return m
}()

// httpRawEvent is the minimal event copy made on the ETW callback thread. The
// ActivityID / RelatedActivityID pair is the correlation key between the
// connection and its requests.
type httpRawEvent struct {
	eventID    uint16
	version    uint8
	ts         time.Time
	pid        uint32
	activityID guid
	hasRelated bool
	relatedID  guid
	userData   []byte
}

// httpConnInfo remembers the server-side (local) and client-side (remote)
// endpoints of an HTTP.sys connection. The server PID arrives later on response
// events; the connection event's header PID belongs to the HTTP client.
type httpConnInfo struct {
	srcIP     string
	srcPort   uint32
	dstIP     string
	dstPort   uint32
	family    string
	connected time.Time
	pending   map[guid]struct{}
}

// httpReqInfo tracks one in-flight request, keyed by the request ActivityID.
type httpReqInfo struct {
	connID       guid
	start        time.Time
	method       string
	path         string
	status       uint16
	bytesWritten uint64
	pid          uint32
	appPool      string
	siteID       uint32
	truncated    bool
	done         bool
}

// httpStats carries httpflow self-telemetry.
type httpStats struct {
	decoded          uint64
	dropped          uint64
	parseErrors      uint64
	completed        uint64
	missedConn       uint64
	missedReq        uint64
	evictedReq       uint64
	droppedReq       uint64
	requestsSkipped  uint64
	invalidFiltered  uint64
	loopbackFiltered uint64
	session          etwSessionStats
}

type httpCollector struct {
	agg         *httpAggregator
	session     *etwSession
	cfg         etwConfig
	pathLimit   int
	maxRequests int

	rawCh chan httpRawEvent

	mu    sync.Mutex
	conns map[guid]*httpConnInfo
	reqs  map[guid]*httpReqInfo

	dropCount  atomic.Uint64
	decoded    atomic.Uint64
	parseErr   atomic.Uint64
	completed  atomic.Uint64
	missedConn atomic.Uint64
	missedReq  atomic.Uint64
	evictedReq atomic.Uint64
	droppedReq atomic.Uint64

	onFatal   func(string)
	stopCh    chan struct{}
	traceDone chan struct{}
	stopping  atomic.Bool
	stopOnce  sync.Once
	wg        sync.WaitGroup
	bufPool   sync.Pool
}

// httpDecodeWorkers is 1 because HTTP correlation is strictly order-sensitive:
// a connection event must be registered before any request event on that
// connection, and the request before its parse/response/terminal events.
// Parallel consumers of rawCh would let a later event overtake an earlier one,
// which produced missed correlations under load. HTTP event volume is modest
// (a handful per request), so a single consumer keeps up.
const httpDecodeWorkers = 1

func newHTTPCollector(agg *httpAggregator, cfg etwConfig, pathLimit, maxRequests int) (flowSource, error) {
	if maxRequests <= 0 {
		maxRequests = defaultMaxHTTPRequests
	}
	if pathLimit <= 0 {
		pathLimit = defaultHTTPPathLimit
	}
	c := &httpCollector{
		agg:         agg,
		cfg:         cfg,
		pathLimit:   pathLimit,
		maxRequests: maxRequests,
		rawCh:       make(chan httpRawEvent, 16384),
		conns:       make(map[guid]*httpConnInfo),
		reqs:        make(map[guid]*httpReqInfo),
		stopCh:      make(chan struct{}),
		traceDone:   make(chan struct{}),
	}
	c.bufPool.New = func() interface{} {
		return new(eventBuffer)
	}
	return c, nil
}

func (c *httpCollector) start() error {
	s := &etwSession{name: httpSessionName}
	if err := s.start(c.cfg, &httpProviderGUID, httpLevel, httpMatchAnyKeywords, httpEventIDs, c); err != nil {
		return fmt.Errorf("start HTTP.sys ETW session: %w", err)
	}
	c.session = s

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer close(c.traceDone)
		err := processTrace(&s.traceHandle, 1)
		if !c.stopping.Load() {
			reportUnexpectedTraceExit("winnetflow HTTP.sys ETW session terminated unexpectedly", err, c.onFatal)
		}
	}()
	for i := 0; i < httpDecodeWorkers; i++ {
		c.wg.Add(1)
		go c.decodeLoop()
	}
	return nil
}

func (c *httpCollector) stop() {
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

func (c *httpCollector) setOnFatal(fn func(string)) {
	c.onFatal = fn
}

func (c *httpCollector) stats() httpStats {
	st := httpStats{
		decoded:     c.decoded.Load(),
		dropped:     c.dropCount.Load(),
		parseErrors: c.parseErr.Load(),
		completed:   c.completed.Load(),
		missedConn:  c.missedConn.Load(),
		missedReq:   c.missedReq.Load(),
		evictedReq:  c.evictedReq.Load(),
		droppedReq:  c.droppedReq.Load(),
	}
	if c.agg != nil {
		st.requestsSkipped = c.agg.requestsSkippedCount()
		st.invalidFiltered, st.loopbackFiltered = c.agg.filterCounts()
	}
	if c.session != nil {
		if s, err := c.session.stats(); err == nil {
			st.session = s
		}
	}
	return st
}

func (c *httpCollector) bumpParseErr() {
	c.parseErr.Add(1)
}

// extDataItem mirrors EVENT_HEADER_EXTENDED_DATA_ITEM (evntcons.h):
//
//	USHORT    Reserved1;
//	USHORT    ExtType;
//	struct { USHORT Linkage : 1; USHORT Reserved2 : 15; };
//	USHORT    DataSize;
//	ULONGLONG DataPtr;
type extDataItem struct {
	Reserved1 uint16
	ExtType   uint16
	Linkage   uint16
	DataSize  uint16
	DataPtr   uint64
}

// extTypeRelatedActivityID is EVENT_HEADER_EXT_TYPE_RELATED_ACTIVITYID.
const extTypeRelatedActivityID = 0x0001

func relatedActivityID(rec *eventRecord) (*guid, bool) {
	if rec.ExtendedDataCount == 0 || rec.ExtendedData == nil {
		return nil, false
	}
	items := unsafe.Slice((*extDataItem)(rec.ExtendedData), rec.ExtendedDataCount)
	for i := range items {
		if items[i].ExtType == extTypeRelatedActivityID && items[i].DataSize == 16 {
			// DataPtr is a native pointer value read from the extended data
			// item; reinterpret the 8-byte field without a uintptr round-trip
			// (which go vet's unsafeptr check rejects).
			ptr := *(*unsafe.Pointer)(unsafe.Pointer(&items[i].DataPtr))
			return (*guid)(ptr), true
		}
	}
	return nil, false
}

func (c *httpCollector) handleRecord(rec *eventRecord) {
	if rec.EventHeader.ProviderID != httpProviderGUID {
		return
	}
	id := rec.EventHeader.Descriptor.ID
	if _, ok := httpEventSet[id]; !ok {
		return
	}
	if rec.UserData == nil || rec.UserDataLength == 0 {
		return
	}

	raw := httpRawEvent{
		eventID:    id,
		version:    rec.EventHeader.Descriptor.Version,
		ts:         filetimeToTime(rec.EventHeader.TimeStamp),
		pid:        rec.EventHeader.ProcessID,
		activityID: rec.EventHeader.ActivityID,
	}
	// Resolve the server process as soon as HTTP.sys emits its first response.
	// Waiting for the ordered decoder can lose the name when a short-lived
	// process exits immediately after completing its final request. The warmer
	// is non-blocking and de-duplicates PIDs, so the ETW callback stays cheap.
	switch id {
	case evHTTPRecvResp, evHTTPFastResp, evHTTPSendComplete,
		evHTTPCachedAndSend, evHTTPFastSend, evHTTPSrvdFrmCache:
		c.warmServerPID(raw.pid)
	}
	if rai, ok := relatedActivityID(rec); ok {
		raw.hasRelated = true
		raw.relatedID = *rai
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

func (c *httpCollector) decodeLoop() {
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

func (c *httpCollector) processRawEvent(raw *httpRawEvent) {
	evs := c.decode(raw)
	c.decoded.Add(1)
	if len(evs) > 0 {
		c.completed.Add(uint64(len(evs)))
		for _, ev := range evs {
			if c.agg != nil {
				c.agg.onEvent(ev)
			}
		}
	}
	releaseEventBuffer(&c.bufPool, raw.userData)
}

func (c *httpCollector) drainRawEvents() {
	for {
		select {
		case raw := <-c.rawCh:
			c.processRawEvent(&raw)
		default:
			return
		}
	}
}

// decode dispatches one raw HTTP.sys event. Fast paths below use the template
// layouts verified against the OS manifest (see dev.md); unknown versions
// degrade to TDH property extraction.
func (c *httpCollector) decode(raw *httpRawEvent) []*httpEvent {
	switch raw.eventID {
	case evHTTPConnConnect:
		if raw.version == 0 {
			c.onConnConnectFast(raw)
		} else {
			c.onConnConnectTDH(raw)
		}
	case evHTTPRecvReq:
		c.onRecvReq(raw)
	case evHTTPParse:
		if raw.version == 0 {
			c.onParseFast(raw)
		} else {
			c.onParseTDH(raw)
		}
	case evHTTPDeliver:
		if raw.version == 0 {
			c.onDeliverFast(raw)
		} else {
			c.onDeliverTDH(raw)
		}
	case evHTTPRecvResp, evHTTPFastResp:
		if raw.version == 0 {
			c.onRespFast(raw)
		} else {
			c.onRespTDH(raw)
		}
	case evHTTPSendComplete, evHTTPCachedAndSend, evHTTPFastSend:
		if raw.version == 0 {
			return c.onSendCompleteFast(raw, false)
		}
		return c.onSendCompleteTDH(raw, false)
	case evHTTPSrvdFrmCache:
		if raw.version == 0 {
			return c.onSendCompleteFast(raw, true)
		}
		return c.onSendCompleteTDH(raw, true)
	case evHTTPConnCleanup:
		c.onConnCleanup(raw)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Fast-path decoders (layout verified on Windows 10.0.26200, see dev.md)
// ---------------------------------------------------------------------------

// onConnConnectFast parses event 21:
//
//	0  ConnectionObj (8)
//	8  LocalAddrLength u32
//	12 LocalAddr sockaddr (16 or 28 bytes)
//	12+len Local RemoteAddrLength u32
//	12+len Local+4 RemoteAddr sockaddr
func (c *httpCollector) onConnConnectFast(raw *httpRawEvent) {
	data := raw.userData
	localLen, ok := u32At(data, 8)
	if !ok || (localLen != 16 && localLen != 28) {
		c.bumpParseErr()
		return
	}
	localOff := 12
	if int(localLen) > len(data)-localOff {
		c.bumpParseErr()
		return
	}
	remoteLenOff := localOff + int(localLen)
	remoteLen, ok := u32At(data, remoteLenOff)
	if !ok || (remoteLen != 16 && remoteLen != 28) {
		c.bumpParseErr()
		return
	}
	remoteOff := remoteLenOff + 4
	if int(remoteLen) > len(data)-remoteOff {
		c.bumpParseErr()
		return
	}
	local, ok := parseSockAddr(data[localOff : localOff+int(localLen)])
	if !ok {
		c.bumpParseErr()
		return
	}
	remote, ok := parseSockAddr(data[remoteOff : remoteOff+int(remoteLen)])
	if !ok {
		c.bumpParseErr()
		return
	}
	c.addConn(raw, local, remote)
}

// onRecvReq correlates event 1 to its connection via the ActivityID and
// registers the request under the RelatedActivityID (the request scope used
// by events 2, 3, 4, 8, 10, 11, 12 and 16).
func (c *httpCollector) onRecvReq(raw *httpRawEvent) {
	if !raw.hasRelated {
		c.bumpParseErr()
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	conn, ok := c.conns[raw.activityID]
	if !ok {
		c.missedConn.Add(1)
		return
	}
	if c.maxRequests > 0 && len(c.reqs) >= c.maxRequests {
		c.evictOneLocked()
	}
	req := &httpReqInfo{
		connID: raw.activityID,
		start:  raw.ts,
	}
	c.reqs[raw.relatedID] = req
	conn.pending[raw.relatedID] = struct{}{}
}

// onParseFast parses event 2:
//
//	0  RequestObj (8)
//	8  HttpVerb u32 (HTTP_VERB enumeration)
//	12 Url UTF-16 null-terminated
func (c *httpCollector) onParseFast(raw *httpRawEvent) {
	data := raw.userData
	c.mu.Lock()
	req, ok := c.reqs[raw.activityID]
	if !ok {
		c.missedReq.Add(1)
		c.mu.Unlock()
		return
	}
	if verb, ok := u32At(data, 8); ok {
		if m := httpVerbName(verb); m != "" {
			req.method = m
		}
	}
	if uri, _, ok := utf16At(data, 12); ok {
		req.path, req.truncated = httpRequestPath(uri, c.pathLimit)
	} else {
		c.bumpParseErr()
	}
	c.mu.Unlock()
}

// onDeliverFast parses event 3:
//
//	0  RequestObj (8)
//	8  RequestId (8)
//	16 SiteId u32
//	20 RequestQueueName UTF-16 null-terminated
//	then Url UTF-16 null-terminated
//	then Status u32
func (c *httpCollector) onDeliverFast(raw *httpRawEvent) {
	data := raw.userData
	c.mu.Lock()
	req, ok := c.reqs[raw.activityID]
	if !ok {
		c.missedReq.Add(1)
		c.mu.Unlock()
		return
	}
	if siteID, ok := u32At(data, 16); ok {
		req.siteID = siteID
	}
	queue, urlOff, ok := utf16At(data, 20)
	if ok {
		req.appPool = queue
	} else {
		c.bumpParseErr()
		c.mu.Unlock()
		return
	}
	if req.path == "" {
		if uri, _, ok := utf16At(data, urlOff); ok {
			req.path, req.truncated = httpRequestPath(uri, c.pathLimit)
		} else {
			c.bumpParseErr()
		}
	}
	c.mu.Unlock()
}

// onRespFast parses events 4/8:
//
//	0  RequestId (8)
//	8  ConnectionId (8)
//	16 StatusCode u16
//	18 Verb ASCII null-terminated
//	then HeaderLength u32, EntityChunkCount u16, CachePolicy u32
func (c *httpCollector) onRespFast(raw *httpRawEvent) {
	data := raw.userData
	c.mu.Lock()
	req, ok := c.reqs[raw.activityID]
	if !ok {
		c.missedReq.Add(1)
		c.mu.Unlock()
		return
	}
	if raw.pid != 0 {
		req.pid = raw.pid
	}
	if status, ok := u16At(data, 16); ok && status != 0 {
		req.status = status
	}
	// The ASCII verb covers verbs outside the HTTP_VERB enumeration (PATCH,
	// custom verbs) and is authoritative when present.
	if verb, ok := asciiAt(data, 18); ok && verb != "" {
		req.method = verb
	}
	pid := req.pid
	c.mu.Unlock()
	c.warmServerPID(pid)
}

// onSendCompleteFast completes the request. For cache-served responses (event
// 16) it also reads BytesSent; for the send-complete family it supplies the
// status code when the response event was missed.
func (c *httpCollector) onSendCompleteFast(raw *httpRawEvent, fromCache bool) []*httpEvent {
	data := raw.userData
	c.mu.Lock()
	req, ok := c.reqs[raw.activityID]
	if !ok {
		c.missedReq.Add(1)
		c.mu.Unlock()
		return nil
	}
	if req.done {
		c.mu.Unlock()
		return nil
	}
	req.done = true
	if raw.pid != 0 {
		req.pid = raw.pid
	}
	if fromCache {
		if b, ok := u32At(data, 12); ok {
			req.bytesWritten = uint64(b)
		}
	} else if req.status == 0 {
		if s, ok := u16At(data, 8); ok {
			req.status = s
		}
	}
	delete(c.reqs, raw.activityID)
	conn := c.conns[req.connID]
	if conn != nil {
		delete(conn.pending, raw.activityID)
	}

	ev := c.reqToEvent(raw, req, conn)
	pid := req.pid
	c.mu.Unlock()
	c.warmServerPID(pid)
	return []*httpEvent{ev}
}

func (c *httpCollector) onConnCleanup(raw *httpRawEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	conn, ok := c.conns[raw.activityID]
	if !ok {
		return
	}
	delete(c.conns, raw.activityID)
	for id := range conn.pending {
		delete(c.reqs, id)
		c.droppedReq.Add(1)
	}
}

// reqToEvent builds the httpflow event; c.mu must be held by the caller.
func (c *httpCollector) reqToEvent(raw *httpRawEvent, req *httpReqInfo, conn *httpConnInfo) *httpEvent {
	ev := &httpEvent{
		ts:        raw.ts,
		start:     req.start,
		method:    req.method,
		path:      req.path,
		status:    req.status,
		truncated: req.truncated,
	}
	if conn != nil {
		ev.family = conn.family
		ev.pid = req.pid
		ev.srcIP = conn.srcIP
		ev.srcPort = conn.srcPort
		ev.dstIP = conn.dstIP
		ev.dstPort = conn.dstPort
		ev.bytesWritten = req.bytesWritten
	}
	return ev
}

// addConn stores a connection, evicting one old connection when at capacity.
func (c *httpCollector) addConn(raw *httpRawEvent, local, remote sockAddr) {
	c.mu.Lock()
	if c.maxRequests > 0 && len(c.conns) >= c.maxRequests {
		for id := range c.conns {
			old := c.conns[id]
			delete(c.conns, id)
			for reqID := range old.pending {
				delete(c.reqs, reqID)
				c.evictedReq.Add(1)
			}
			break
		}
	}
	c.conns[raw.activityID] = &httpConnInfo{
		srcIP:     local.ip,
		srcPort:   local.port,
		dstIP:     remote.ip,
		dstPort:   remote.port,
		family:    local.family,
		connected: raw.ts,
		pending:   make(map[guid]struct{}),
	}
	c.mu.Unlock()
}

func (c *httpCollector) warmServerPID(pid uint32) {
	if c.agg != nil && pid != 0 {
		c.agg.warmPID(pid)
	}
}

// warmPID starts process attribution at the earliest server response event.
// HTTP test servers and other short-lived HTTP.sys applications can exit before
// the completed request reaches flush.
func (a *httpAggregator) warmPID(pid uint32) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.warmPIDLocked(pid)
}

// evictOneLocked drops one in-flight request when the request map is full.
// c.mu must be held by the caller.
func (c *httpCollector) evictOneLocked() {
	for id, req := range c.reqs {
		delete(c.reqs, id)
		if conn, ok := c.conns[req.connID]; ok {
			delete(conn.pending, id)
		}
		c.evictedReq.Add(1)
		return
	}
}

// ---------------------------------------------------------------------------
// TDH fallbacks for unknown event versions
// ---------------------------------------------------------------------------

func (c *httpCollector) rawToRecord(raw *httpRawEvent) *eventRecord {
	rec := &eventRecord{}
	rec.EventHeader.ProviderID = httpProviderGUID
	rec.EventHeader.Descriptor.ID = raw.eventID
	rec.EventHeader.Descriptor.Version = raw.version
	rec.EventHeader.Descriptor.Level = httpLevel
	rec.UserDataLength = uint16(len(raw.userData))
	if len(raw.userData) > 0 {
		rec.UserData = unsafe.Pointer(&raw.userData[0])
	}
	return rec
}

func (c *httpCollector) onConnConnectTDH(raw *httpRawEvent) {
	rec := c.rawToRecord(raw)
	local, ok := propSockAddr(rec, "LocalAddr")
	if !ok {
		c.bumpParseErr()
		return
	}
	remote, ok := propSockAddr(rec, "RemoteAddr")
	if !ok {
		c.bumpParseErr()
		return
	}
	c.addConn(raw, local, remote)
}

func (c *httpCollector) onParseTDH(raw *httpRawEvent) {
	rec := c.rawToRecord(raw)
	verb, _ := propU32(rec, "HttpVerb")
	c.mu.Lock()
	req, ok := c.reqs[raw.activityID]
	if !ok {
		c.missedReq.Add(1)
		c.mu.Unlock()
		return
	}
	if m := httpVerbName(verb); m != "" {
		req.method = m
	}
	if b, ok := property(rec, "Url"); ok {
		if uri, ok := utf16BytesToString(b); ok {
			req.path, req.truncated = httpRequestPath(uri, c.pathLimit)
		}
	}
	c.mu.Unlock()
}

func (c *httpCollector) onDeliverTDH(raw *httpRawEvent) {
	rec := c.rawToRecord(raw)
	siteID, _ := propU32(rec, "SiteId")
	c.mu.Lock()
	req, ok := c.reqs[raw.activityID]
	if !ok {
		c.missedReq.Add(1)
		c.mu.Unlock()
		return
	}
	req.siteID = siteID
	if b, ok := property(rec, "RequestQueueName"); ok {
		if q, ok := utf16BytesToString(b); ok {
			req.appPool = q
		}
	}
	if req.path == "" {
		if b, ok := property(rec, "Url"); ok {
			if uri, ok := utf16BytesToString(b); ok {
				req.path, req.truncated = httpRequestPath(uri, c.pathLimit)
			}
		}
	}
	c.mu.Unlock()
}

func (c *httpCollector) onRespTDH(raw *httpRawEvent) {
	rec := c.rawToRecord(raw)
	status, _ := propU16(rec, "StatusCode")
	c.mu.Lock()
	req, ok := c.reqs[raw.activityID]
	if !ok {
		c.missedReq.Add(1)
		c.mu.Unlock()
		return
	}
	if raw.pid != 0 {
		req.pid = raw.pid
	}
	if status != 0 {
		req.status = status
	}
	if b, ok := property(rec, "Verb"); ok {
		if verb := asciiBytesToString(b); verb != "" {
			req.method = verb
		}
	}
	pid := req.pid
	c.mu.Unlock()
	c.warmServerPID(pid)
}

func (c *httpCollector) onSendCompleteTDH(raw *httpRawEvent, fromCache bool) []*httpEvent {
	rec := c.rawToRecord(raw)
	c.mu.Lock()
	req, ok := c.reqs[raw.activityID]
	if !ok {
		c.missedReq.Add(1)
		c.mu.Unlock()
		return nil
	}
	if req.done {
		c.mu.Unlock()
		return nil
	}
	req.done = true
	if raw.pid != 0 {
		req.pid = raw.pid
	}
	if fromCache {
		if b, _ := propU32(rec, "BytesSent"); b > 0 {
			req.bytesWritten = uint64(b)
		}
	} else if req.status == 0 {
		if s, ok := propU16(rec, "HttpStatus"); ok {
			req.status = s
		}
	}
	delete(c.reqs, raw.activityID)
	conn := c.conns[req.connID]
	if conn != nil {
		delete(conn.pending, raw.activityID)
	}
	ev := c.reqToEvent(raw, req, conn)
	pid := req.pid
	c.mu.Unlock()
	c.warmServerPID(pid)
	return []*httpEvent{ev}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func u16At(data []byte, off int) (uint16, bool) {
	if off < 0 || off+2 > len(data) {
		return 0, false
	}
	return binary.LittleEndian.Uint16(data[off:]), true
}

// utf16At reads a UTF-16LE null-terminated string at off and returns the
// string plus the offset just past the terminator (for chained fields).
func utf16At(data []byte, off int) (string, int, bool) {
	if off < 0 || off+2 > len(data) {
		return "", off, false
	}
	end := off
	terminated := false
	for end+1 < len(data) {
		if binary.LittleEndian.Uint16(data[end:]) == 0 {
			terminated = true
			break
		}
		end += 2
	}
	if !terminated {
		return "", off, false
	}
	u16 := make([]uint16, 0, (end-off)/2)
	for i := off; i < end; i += 2 {
		u16 = append(u16, binary.LittleEndian.Uint16(data[i:]))
	}
	return string(utf16.Decode(u16)), end + 2, true
}

func asciiAt(data []byte, off int) (string, bool) {
	if off < 0 || off >= len(data) {
		return "", false
	}
	end := off
	for end < len(data) && data[end] != 0 {
		end++
	}
	if end >= len(data) {
		return "", false
	}
	return string(data[off:end]), true
}

func utf16BytesToString(b []byte) (string, bool) {
	if len(b) < 2 {
		return "", false
	}
	u16 := make([]uint16, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		u := binary.LittleEndian.Uint16(b[i:])
		if u == 0 {
			break
		}
		u16 = append(u16, u)
	}
	return string(utf16.Decode(u16)), true
}

func asciiBytesToString(b []byte) string {
	end := 0
	for end < len(b) && b[end] != 0 {
		end++
	}
	return string(b[:end])
}
