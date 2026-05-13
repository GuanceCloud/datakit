//go:build linux
// +build linux

package l4log

import (
	"bytes"
)

// var _ L7ProtoEventAndMetric = (*HTTPLog)(nil)

// const maxTCPHDRCountLimit = 320

type HTTPLog struct {
	elems  []*HTTPLogElem
	isHTTP bool

	flow *tcpFlowTracker

	txState httpHTTPStreamState
	rxState httpHTTPStreamState

	probePackets   uint8
	probeBytes     int
	probeExhausted bool
}

const (
	httpProbePacketBudget = 8
	httpProbeByteBudget   = 4 * 1024
)

func (h *HTTPLog) Handle(v any, txrx int8, cnt []byte,
	cntSize int64, ln *PktTCPHdr, k *PMeta, pktState int8, chunkid int64,
) error {
	st := h.ensureStreamState(txrx)
	needPendingReplay := st != nil && h.activeElem() == nil && cntSize > 0 &&
		(st.header.active || maybeHTTPStart(cnt))
	if needPendingReplay && st != nil {
		st.pendingPackets = append(st.pendingPackets, httpPacketObservation{
			ln:       *ln,
			cntSize:  cntSize,
			pktState: pktState,
			chunkid:  chunkid,
		})
	}

	replayed := false
	for _, evt := range h.feedEvents(txrx, ln, cnt) {
		elem := h.applyEvent(txrx, ln, evt)
		if elem != nil && st != nil && len(st.pendingPackets) > 0 {
			h.replayPendingPackets(elem, txrx, st.pendingPackets)
			st.pendingPackets = st.pendingPackets[:0]
			replayed = true
		}
	}

	elem := h.activeElem()
	if elem == nil {
		return nil
	}

	if replayed {
		return nil
	}

	h.recordPacket(elem, txrx, pktState, cntSize, ln, chunkid)
	return nil
}

func (h *HTTPLog) replayPendingPackets(elem *HTTPLogElem, txrx int8, packets []httpPacketObservation) {
	for _, pkt := range packets {
		pktLn := pkt.ln
		h.recordPacket(elem, txrx, pkt.pktState, pkt.cntSize, &pktLn, pkt.chunkid)
	}
}

func (h *HTTPLog) recordPacket(elem *HTTPLogElem, txrx int8, pktState int8, cntSize int64, ln *PktTCPHdr, chunkid int64) {
	elem.markMessageDirty()
	if elem.ChunkRange[0] == 0 || elem.ChunkRange[0] > chunkid {
		elem.ChunkRange[0] = chunkid
	}
	if elem.ChunkRange[1] < chunkid {
		elem.ChunkRange[1] = chunkid
	}

	elem.recReqRespTS(txrx, cntSize, ln)

	if pktState == 1 {
		switch txrx {
		case directionRX:
			elem.rxRetransmits++
		case directionTX:
			elem.txRetransmits++
		}
		return
	}

	if cntSize > 0 && elem.hState > 0 && !elem.hFinished {
		switch txrx {
		case directionRX:
			elem.rxPkts++
			elem.rxBytes += cntSize
		case directionTX:
			elem.txPkts++
			elem.txBytes += cntSize
		}
	}
}

const (
	DIncoming = "incoming"
	DOutging  = "outgoing"
)

type HTTPLogElem struct {
	Direction string `json:"direction"`
	// tcp seq

	ChunkRange [2]int64 `json:"pkt_chunk_range"`

	reqSeq  uint32
	respSeq uint32

	// fist packet arrive time
	// ReqTS  int64 `json:"req_first_arrive_ts"`
	// RespTS int64 `json:"resp_first_arrive_ts"`

	txFirstByteTS int64
	rxFirstByteTS int64

	txNxtAckSeq uint32
	txAcked     bool

	rxNxtAckSeq uint32
	rxAcked     bool
	// 如果是 tx 则通过 rx 的 ack 或 tx 含 payload 的包的时间计算
	txLastByteTS int64

	rxLastByteTS int64

	// tcp packets
	txPkts  int64
	rxPkts  int64
	txBytes int64
	rxBytes int64

	txRetransmits int
	rxRetransmits int

	// req/resp content size (tcp payload <http>)
	// Send int64 `json:"send_bytes"`
	// Recv int64 `json:"recv_bytes"`

	TraceID  string `json:"trace_id"`
	ParentID string `json:"parent_id"`

	// HTTPVersion string
	// ReqHeaders map[string][]string `json:"req_headers"`
	// RespHeaders map[string][]string `json:"resp_headers"`

	// URL
	Path string `json:"path"`
	Host string `json:"host,omitempty"`
	// Param string `json:"param"`

	Method string `json:"method"`

	// response
	StatusCode int `json:"status_code"`

	hState    int8 // 1: req, 2: resp
	hFinished bool

	messageCache string
	messageDirty bool
}

type httpParsedEvent struct {
	reqResp  int8
	seq      uint32
	ts       int64
	method   string
	path     string
	host     string
	traceID  string
	parentID string
	status   int
}

type httpHeaderAssembler struct {
	pending    []byte
	pendingSeq uint32
	pendingTS  int64
	active     bool
	maxPending int
}

type httpPacketObservation struct {
	ln       PktTCPHdr
	cntSize  int64
	pktState int8
	chunkid  int64
}

type httpHTTPStreamState struct {
	header         httpHeaderAssembler
	pendingPackets []httpPacketObservation
}

func httpReqOrResp(cnt []byte) int8 { // 1: req, 2: resp
	line, ok := firstHTTPLine(cnt)
	if !ok {
		return 0
	}

	methodEnd := bytes.IndexByte(line, ' ')
	if methodEnd <= 0 {
		return 0
	}

	if isHTTPRequestMethod(line[:methodEnd]) {
		return 1
	}

	if bytes.HasPrefix(line, []byte("HTTP/")) {
		return 2
	}

	return 0
}

func (h *HTTPLog) DetectProto(cnt []byte) bool {
	return httpReqOrResp(cnt) > 0
}

func (h *HTTPLog) ShouldHandle(cnt []byte) bool {
	if h.isHTTP || h.activeElem() != nil || h.hasPendingState() {
		return true
	}

	if len(cnt) == 0 {
		return false
	}

	if maybeHTTPStart(cnt) {
		return true
	}

	if h.probeExhausted {
		return false
	}

	h.probePackets++
	h.probeBytes += len(cnt)
	if h.probePackets >= httpProbePacketBudget || h.probeBytes >= httpProbeByteBudget {
		h.probeExhausted = true
	}

	return true
}

func (h *HTTPLog) ensureStreamState(txrx int8) *httpHTTPStreamState {
	var st *httpHTTPStreamState

	switch txrx {
	case directionTX:
		st = &h.txState
	case directionRX:
		st = &h.rxState
	default:
		return nil
	}

	if h.flow == nil {
		h.flow = newTCPFlowTracker(8, 64*1024)
	}
	if st.header.maxPending == 0 {
		st.header.maxPending = 16 * 1024
	}

	return st
}

func (h *HTTPLog) hasPendingState() bool {
	return h.txState.header.active || h.rxState.header.active ||
		len(h.txState.pendingPackets) > 0 || len(h.rxState.pendingPackets) > 0
}

func (h *HTTPLog) feedEvents(txrx int8, ln *PktTCPHdr, cnt []byte) []httpParsedEvent {
	if len(cnt) == 0 {
		return nil
	}

	st := h.ensureStreamState(txrx)
	if st == nil {
		return nil
	}

	res := h.flow.Push(txrx, ln.Seq, ln.Flags, cnt, ln.TS)
	events := make([]httpParsedEvent, 0, len(res.Deliveries))
	for _, delivery := range res.Deliveries {
		events = append(events, st.header.Feed(delivery)...)
	}
	if len(events) > 0 {
		h.isHTTP = true
	} else if !st.header.active && len(st.pendingPackets) > 0 {
		st.pendingPackets = st.pendingPackets[:0]
	}

	return events
}

func (h *HTTPLog) lastElem() *HTTPLogElem {
	if len(h.elems) == 0 {
		return nil
	}
	return h.elems[len(h.elems)-1]
}

func (h *HTTPLog) activeElem() *HTTPLogElem {
	elem := h.lastElem()
	if elem == nil || elem.hFinished {
		return nil
	}
	return elem
}

func (h *HTTPLog) newElem() *HTTPLogElem {
	elem := &HTTPLogElem{messageDirty: true}
	h.elems = append(h.elems, elem)
	return elem
}

func (h *HTTPLog) applyEvent(txrx int8, ln *PktTCPHdr, evt httpParsedEvent) *HTTPLogElem {
	switch evt.reqResp {
	case 1:
		elem := h.lastElem()
		if elem != nil && elem.reqSeq != 0 && elem.reqSeq != evt.seq {
			elem.recReqRespTS(txrx, 0, ln)
			elem.finished(true)
			elem = nil
		}
		if elem == nil || elem.hFinished {
			elem = h.newElem()
		}
		elem.markMessageDirty()
		switch txrx {
		case directionRX:
			elem.Direction = DIncoming
		case directionTX:
			elem.Direction = DOutging
		}
		elem.reqSeq = evt.seq
		elem.hState = 1
		elem.Method = evt.method
		elem.Path = evt.path
		elem.Host = evt.host
		elem.TraceID = evt.traceID
		elem.ParentID = evt.parentID
		return elem
	case 2:
		elem := h.lastElem()
		if elem == nil || elem.hFinished {
			elem = h.newElem()
		}
		elem.markMessageDirty()
		switch txrx {
		case directionRX:
			elem.Direction = DOutging
		case directionTX:
			elem.Direction = DIncoming
		}
		elem.respSeq = evt.seq
		elem.hState = 2
		elem.StatusCode = evt.status
		return elem
	}

	return nil
}

func (h *HTTPLogElem) markMessageDirty() {
	h.messageDirty = true
}

func (h *HTTPLogElem) recReqRespTS(txrx int8, cntSize int64, ln *PktTCPHdr) {
	switch txrx {
	case directionTX:
		if cntSize > 0 {
			if h.txFirstByteTS == 0 {
				h.txFirstByteTS = ln.TS
			}
			h.txLastByteTS = ln.TS
			h.txNxtAckSeq = ln.Seq + uint32(cntSize)
			h.txAcked = false
		} else if !h.rxAcked && h.rxNxtAckSeq == ln.AckSeq {
			h.rxLastByteTS = ln.TS
			h.rxAcked = true
		}
	case directionRX:
		if cntSize > 0 {
			if h.rxFirstByteTS == 0 {
				h.rxFirstByteTS = ln.TS
			}
			h.rxNxtAckSeq = ln.Seq + uint32(cntSize)
			h.rxLastByteTS = ln.TS
			h.rxAcked = false
		} else if !h.txAcked && h.txNxtAckSeq == ln.AckSeq {
			h.txLastByteTS = ln.TS
			h.txAcked = true
		}
	}
}

func (h *HTTPLogElem) finished(haveNxtReq bool) bool {
	if haveNxtReq && h.hState != 0 {
		h.hFinished = true
		// todo: do some thing here
	}

	return h.hFinished
}

func (a *httpHeaderAssembler) Feed(delivery streamDelivery) []httpParsedEvent {
	if len(delivery.Payload) == 0 {
		return nil
	}

	if !a.active {
		if !maybeHTTPStart(delivery.Payload) {
			return nil
		}
		a.active = true
		a.pendingSeq = delivery.Seq
		a.pendingTS = delivery.TS
		a.pending = append(a.pending[:0], delivery.Payload...)
	} else {
		a.pending = append(a.pending, delivery.Payload...)
	}

	if a.maxPending > 0 && len(a.pending) > a.maxPending {
		a.reset()
		return nil
	}

	block, ok := httpHeaderBlock(a.pending)
	if !ok {
		return nil
	}

	reqResp := httpReqOrResp(block)
	switch reqResp {
	case 1:
		method, path, host, traceID, parentID, ok := parseHTTPRequestMeta(block)
		if ok {
			evt := httpParsedEvent{
				reqResp:  1,
				seq:      a.pendingSeq,
				ts:       a.pendingTS,
				method:   method,
				path:     path,
				host:     host,
				traceID:  traceID,
				parentID: parentID,
			}
			a.reset()
			return []httpParsedEvent{evt}
		}
	case 2:
		if status, ok := parseHTTPResponseStatus(block); ok {
			evt := httpParsedEvent{
				reqResp: 2,
				seq:     a.pendingSeq,
				ts:      a.pendingTS,
				status:  status,
			}
			a.reset()
			return []httpParsedEvent{evt}
		}
	}

	a.reset()
	return nil
}

func (a *httpHeaderAssembler) reset() {
	a.pending = a.pending[:0]
	a.pendingSeq = 0
	a.pendingTS = 0
	a.active = false
}

func httpHeaderBlock(cnt []byte) ([]byte, bool) {
	if idx := bytes.Index(cnt, []byte("\r\n\r\n")); idx >= 0 {
		return cnt[:idx+4], true
	}
	return nil, false
}

func maybeHTTPStart(cnt []byte) bool {
	if len(cnt) == 0 {
		return false
	}

	methods := [][]byte{
		[]byte("GET "), []byte("HEAD "), []byte("POST "), []byte("PUT "),
		[]byte("DELETE "), []byte("CONNECT "), []byte("OPTIONS "),
		[]byte("PATCH "), []byte("TRACE "), []byte("HTTP/"),
	}

	for _, m := range methods {
		switch {
		case len(cnt) >= len(m) && bytes.HasPrefix(cnt, m):
			return true
		case len(cnt) < len(m) && bytes.Equal(cnt, m[:len(cnt)]):
			return true
		}
	}

	return false
}

func firstHTTPLine(cnt []byte) ([]byte, bool) {
	if s := bytes.Index(cnt, []byte{'\r', '\n'}); s > 0 {
		return cnt[:s], true
	}

	return nil, false
}

func isHTTPRequestMethod(method []byte) bool {
	switch len(method) {
	case 3:
		return bytes.Equal(method, []byte("GET")) || bytes.Equal(method, []byte("PUT"))
	case 4:
		return bytes.Equal(method, []byte("HEAD")) || bytes.Equal(method, []byte("POST"))
	case 5:
		return bytes.Equal(method, []byte("PATCH")) || bytes.Equal(method, []byte("TRACE"))
	case 6:
		return bytes.Equal(method, []byte("DELETE"))
	case 7:
		return bytes.Equal(method, []byte("CONNECT")) || bytes.Equal(method, []byte("OPTIONS"))
	default:
		return false
	}
}

func nextHTTPLine(cnt []byte, start int) (line []byte, next int, ok bool) {
	if start >= len(cnt) {
		return nil, start, false
	}

	for i := start; i+1 < len(cnt); i++ {
		if cnt[i] == '\r' && cnt[i+1] == '\n' {
			return cnt[start:i], i + 2, true
		}
	}

	return cnt[start:], len(cnt), len(cnt) > start
}

func parseHTTPRequestMeta(cnt []byte) (method, path, host, traceID, parentID string, ok bool) {
	line, next, ok := nextHTTPLine(cnt, 0)
	if !ok {
		return "", "", "", "", "", false
	}

	firstSpace := bytes.IndexByte(line, ' ')
	if firstSpace <= 0 || !isHTTPRequestMethod(line[:firstSpace]) {
		return "", "", "", "", "", false
	}

	rest := line[firstSpace+1:]
	secondSpace := bytes.IndexByte(rest, ' ')
	if secondSpace <= 0 {
		return "", "", "", "", "", false
	}

	method = string(line[:firstSpace])
	target := rest[:secondSpace]
	path = string(target)
	if q := bytes.IndexByte(target, '?'); q >= 0 {
		path = string(target[:q])
	}
	host = extractAbsoluteFormHost(target)

	for next < len(cnt) {
		header, n, ok := nextHTTPLine(cnt, next)
		completeLine := n >= 2 && n <= len(cnt) && cnt[n-2] == '\r' && cnt[n-1] == '\n'
		next = n
		if !ok || len(header) == 0 {
			break
		}

		colon := bytes.IndexByte(header, ':')
		if colon <= 0 {
			continue
		}

		name := header[:colon]
		switch {
		case bytes.EqualFold(name, []byte("traceparent")):
			value := bytes.TrimLeft(header[colon+1:], " \t")
			traceID, parentID = parseTraceparentHeader(value)
		case host == "" && completeLine && bytes.EqualFold(name, []byte("host")):
			host = normalizeHTTPHostBytes(header[colon+1:])
		}
	}

	return method, path, host, traceID, parentID, true
}

func parseHTTPResponseStatus(cnt []byte) (int, bool) {
	line, ok := firstHTTPLine(cnt)
	if !ok || !bytes.HasPrefix(line, []byte("HTTP/")) {
		return 0, false
	}

	firstSpace := bytes.IndexByte(line, ' ')
	if firstSpace <= 0 || firstSpace+4 > len(line) {
		return 0, false
	}

	status := line[firstSpace+1:]
	if len(status) < 3 {
		return 0, false
	}

	code := 0
	for i := 0; i < 3; i++ {
		ch := status[i]
		if ch < '0' || ch > '9' {
			return 0, false
		}
		code = code*10 + int(ch-'0')
	}

	return code, true
}

func parseTraceparentHeader(value []byte) (traceID, parentID string) {
	parts := bytes.SplitN(value, []byte{'-'}, 4)
	if len(parts) != 4 {
		return "", ""
	}

	return string(parts[1]), string(parts[2])
}
