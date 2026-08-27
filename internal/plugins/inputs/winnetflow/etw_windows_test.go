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
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf16"
)

func TestParseSockAddr(t *testing.T) {
	v4 := []byte{
		0x02, 0x00, // AF_INET
		0x01, 0xBB, // port 443
		0xC0, 0xA8, 0x01, 0x01, // 192.168.1.1
		0, 0, 0, 0, 0, 0, 0, 0,
	}
	sa, ok := parseSockAddr(v4)
	if !ok {
		t.Fatal("parse v4 failed")
	}
	if sa.ip != "192.168.1.1" || sa.port != 443 || sa.family != "IPv4" {
		t.Fatalf("unexpected v4 result: %+v", sa)
	}

	v6 := make([]byte, 28)
	v6[0], v6[1] = 0x17, 0x00 // AF_INET6
	v6[2], v6[3] = 0x00, 0x35 // port 53
	copy(v6[8:24], net.ParseIP("::1").To16())
	sa6, ok := parseSockAddr(v6)
	if !ok {
		t.Fatal("parse v6 failed")
	}
	if sa6.ip != "::1" || sa6.port != 53 || sa6.family != "IPv6" {
		t.Fatalf("unexpected v6 result: %+v", sa6)
	}

	if _, ok := parseSockAddr(v6[:6]); ok {
		t.Fatal("short buffer should fail")
	}
}

func TestTCBCorrelationMapBoundedAndPruned(t *testing.T) {
	now := time.Now()
	d := &etwDecoder{
		tcbs:     make(map[uint64]*tcbInfo),
		tcbLimit: 2,
		tcbTTL:   time.Minute,
	}
	d.mu.Lock()
	d.rememberTCBLocked(1, &tcbInfo{}, now)
	d.rememberTCBLocked(2, &tcbInfo{}, now.Add(time.Second))
	d.rememberTCBLocked(3, &tcbInfo{}, now.Add(2*time.Second))
	d.mu.Unlock()

	if len(d.tcbs) != 2 {
		t.Fatalf("TCB map size = %d, want 2", len(d.tcbs))
	}
	if _, ok := d.tcbs[3]; !ok {
		t.Fatal("new TCB was not inserted after capacity eviction")
	}
	if d.tcbEvicted != 1 {
		t.Fatalf("evicted = %d, want 1", d.tcbEvicted)
	}

	d.pruneTCBs(now.Add(2 * time.Minute))
	if len(d.tcbs) != 0 {
		t.Fatalf("expired TCBs remain: %d", len(d.tcbs))
	}
}

func TestPendingCorrelationPruned(t *testing.T) {
	now := time.Now()
	d := &etwDecoder{
		tcbs:       make(map[uint64]*tcbInfo),
		pending:    map[uint64]*pendingStats{1: {lastSeen: now}},
		pendingTTL: time.Minute,
	}
	d.pruneTCBs(now.Add(2 * time.Minute))
	if len(d.pending) != 0 || d.pendingEvicted != 1 {
		t.Fatalf("pending size=%d evicted=%d, want 0/1", len(d.pending), d.pendingEvicted)
	}
}

func TestProcessNameCacheBounded(t *testing.T) {
	procNameCache.Lock()
	old := procNameCache.m
	procNameCache.m = make(map[uint32]cachedProcessName)
	procNameCache.Unlock()
	t.Cleanup(func() {
		procNameCache.Lock()
		procNameCache.m = old
		procNameCache.Unlock()
	})

	now := time.Now()
	for i := 0; i < procNameCacheLimit+10; i++ {
		cacheProcessName(uint32(i+1), "test", now.Add(time.Duration(i)*time.Millisecond))
	}
	procNameCache.Lock()
	got := len(procNameCache.m)
	procNameCache.Unlock()
	if got != procNameCacheLimit {
		t.Fatalf("process cache size = %d, want %d", got, procNameCacheLimit)
	}
}

func TestProcessNameRefreshReplacesStalePIDEntry(t *testing.T) {
	pid := uint32(os.Getpid())
	procNameCache.Lock()
	old := procNameCache.m
	procNameCache.m = map[uint32]cachedProcessName{
		pid: {name: "stale-process.exe", ts: time.Now()},
	}
	procNameCache.Unlock()
	t.Cleanup(func() {
		procNameCache.Lock()
		procNameCache.m = old
		procNameCache.Unlock()
	})

	name := refreshProcessNameCache(pid, time.Now())
	if name == "" || name == "stale-process.exe" {
		t.Fatalf("refreshed process name = %q", name)
	}
	if got := lookupProcessName(pid); got != name {
		t.Fatalf("cached refreshed name = %q, want %q", got, name)
	}
}

func TestSessionSemaphoreOwnership(t *testing.T) {
	name := utf16.Encode([]rune(fmt.Sprintf("Global\\datakit-winnetflow-test-%d\x00", os.Getpid())))
	first, err := acquireSessionSemaphore(&name[0])
	if err != nil {
		t.Fatalf("acquire first mutex: %v", err)
	}
	defer func() {
		if first != 0 {
			releaseSessionSemaphore(first)
		}
	}()

	if second, err := acquireSessionSemaphore(&name[0]); err == nil {
		releaseSessionSemaphore(second)
		t.Fatal("second owner unexpectedly acquired active semaphore")
	}

	releaseSessionSemaphore(first)
	first = 0
	third, err := acquireSessionSemaphore(&name[0])
	if err != nil {
		t.Fatalf("reacquire released mutex: %v", err)
	}
	releaseSessionSemaphore(third)
}

func TestUnexpectedSuccessfulTraceExitIsFatal(t *testing.T) {
	var got string
	reportUnexpectedTraceExit("trace stopped", nil, func(message string) {
		got = message
	})
	if got != "trace stopped" {
		t.Fatalf("fatal message = %q, want %q", got, "trace stopped")
	}
}

func TestTCPDecodeWorkerCountPreservesOrder(t *testing.T) {
	if decodeWorkers != 1 {
		t.Fatalf("decode workers = %d, want 1", decodeWorkers)
	}
}

func TestRawEventPreservesTimestamp(t *testing.T) {
	raw := rawEvent{
		eventID:   1332,
		version:   5,
		timestamp: 132537600000000000,
		userData:  []byte{1},
	}
	rec := rawToEventRecord(&raw)
	if rec.EventHeader.TimeStamp != raw.timestamp {
		t.Fatalf("timestamp = %d, want %d", rec.EventHeader.TimeStamp, raw.timestamp)
	}
}

func TestRawEventWithEmptyPayload(t *testing.T) {
	rec := rawToEventRecord(&rawEvent{eventID: 1332})
	if rec.UserData != nil || rec.UserDataLength != 0 {
		t.Fatalf("empty payload produced UserData=%p length=%d", rec.UserData, rec.UserDataLength)
	}
}

func TestEventBufferPoolDoesNotRetainLargeBuffers(t *testing.T) {
	var allocations atomic.Uint64
	pool := sync.Pool{New: func() interface{} {
		allocations.Add(1)
		return new(eventBuffer)
	}}

	large := acquireEventBuffer(&pool, pooledEventBufferSize+1)
	releaseEventBuffer(&pool, large)
	pooled := acquireEventBuffer(&pool, 1)
	if cap(pooled) != pooledEventBufferSize {
		t.Fatalf("pooled capacity = %d, want %d", cap(pooled), pooledEventBufferSize)
	}
	if allocations.Load() != 1 {
		t.Fatalf("pool allocations = %d, want 1; oversized buffer was retained", allocations.Load())
	}
}

func TestEventBufferPoolHotPathDoesNotAllocate(t *testing.T) {
	pool := sync.Pool{New: func() interface{} { return new(eventBuffer) }}
	pool.Put(new(eventBuffer))
	allocs := testing.AllocsPerRun(1000, func() {
		buf := acquireEventBuffer(&pool, 128)
		releaseEventBuffer(&pool, buf)
	})
	if allocs != 0 {
		t.Fatalf("buffer pool hot-path allocations = %.2f, want 0", allocs)
	}
}

func TestL4DrainProcessesQueuedEvents(t *testing.T) {
	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	source, err := newETWCollector(agg, defaultETWConfig())
	if err != nil {
		t.Fatalf("new collector: %v", err)
	}
	c := source.(*etwCollector)
	payload := build1332(5, 0x1234, 4096, 2500, 100,
		v4SockAddr("127.0.0.1", 50000), v4SockAddr("10.0.0.1", 443))
	buf := acquireEventBuffer(&c.bufPool, len(payload))
	copy(buf, payload)
	c.rawCh <- rawEvent{
		eventID:   1332,
		version:   5,
		timestamp: 132537600000000000,
		userData:  buf,
	}

	c.drainRawEvents()
	if got := c.decoded.Load(); got != 1 {
		t.Fatalf("decoded = %d, want 1", got)
	}
	pts := agg.flush()
	if len(pts) != 1 || pts[0].Fields().Get("bytes_written").GetI() != 4096 {
		t.Fatalf("drained points = %v", pts)
	}
}

func TestL4CollectorStatsIncludeFilterCounts(t *testing.T) {
	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	agg.onEvent(&flowEvent{srcIP: "0.0.0.0", srcPort: 1, dstIP: "10.0.0.1", dstPort: 2})
	agg.onEvent(&flowEvent{srcIP: "127.0.0.1", srcPort: 1, dstIP: "127.0.0.1", dstPort: 2})
	st := (&etwCollector{agg: agg}).stats()
	if st.invalidFiltered != 1 || st.loopbackFiltered != 1 {
		t.Fatalf("filter stats = invalid:%d loopback:%d, want 1/1", st.invalidFiltered, st.loopbackFiltered)
	}
}

func BenchmarkFastTCPSendDecode(b *testing.B) {
	d := testDecoder()
	registerTCP(d, 0x1234)
	rec := testRec(5)
	rec.EventHeader.TimeStamp = 132537600000000000
	data := build1332(5, 0x1234, 4096, 2500, 100,
		v4SockAddr("127.0.0.1", 50000), v4SockAddr("10.0.0.1", 443))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if evs := d.onTCPSendFast(rec, data); evs.len() != 1 {
			b.Fatal("decode failed")
		}
	}
}

func BenchmarkFastTCPSendPipeline(b *testing.B) {
	d := testDecoder()
	registerTCP(d, 0x1234)
	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	rec := testRec(5)
	rec.EventHeader.TimeStamp = 132537600000000000
	data := build1332(5, 0x1234, 4096, 2500, 100,
		v4SockAddr("127.0.0.1", 50000), v4SockAddr("10.0.0.1", 443))
	first := d.onTCPSendFast(rec, data)
	agg.onEvent(first.at(0))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batch := d.onTCPSendFast(rec, data)
		for j := 0; j < batch.len(); j++ {
			agg.onEvent(batch.at(j))
		}
	}
}

func BenchmarkEventBufferPool(b *testing.B) {
	pool := sync.Pool{New: func() interface{} { return new(eventBuffer) }}
	pool.Put(new(eventBuffer))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := acquireEventBuffer(&pool, 128)
		releaseEventBuffer(&pool, buf)
	}
}

func TestParseGUID(t *testing.T) {
	g := parseGUID("2f07e2ee-15db-40f1-90ef-9d7ba282188a")
	want := [16]byte{
		0xee, 0xe2, 0x07, 0x2f,
		0xdb, 0x15,
		0xf1, 0x40,
		0x90, 0xef, 0x9d, 0x7b, 0xa2, 0x82, 0x18, 0x8a,
	}
	if g != (guid)(want) {
		t.Fatalf("guid mismatch: %x, want %x", g, want)
	}
}

func TestPendingAccumulateAndReplay(t *testing.T) {
	d := &etwDecoder{
		tcbs:    make(map[uint64]*tcbInfo),
		pending: make(map[uint64]*pendingStats),
	}

	d.accumulatePending(1, 100, 0, 0, 0, 500, 50, 0)
	d.accumulatePending(1, 0, 200, 0, 3, 0, 0, 2)
	d.accumulatePending(1, 0, 0, 4, 0, 0, 0, 0)

	info := &tcbInfo{
		srcIP: "10.0.0.1", srcPort: 50000,
		dstIP: "10.0.0.2", dstPort: 80,
		pid: 42, family: "IPv4", direction: directionOutgoing,
	}

	d.mu.Lock()
	burst, ok := d.replayPendingLocked(1, info, time.Now())
	d.mu.Unlock()

	if !ok {
		t.Fatal("expected burst event")
	}
	if burst.kind != evBurstReplay {
		t.Fatalf("unexpected kind %v", burst.kind)
	}
	if burst.cumBytesSend != 100 || burst.cumBytesRecv != 200 {
		t.Fatalf("unexpected bytes: send=%d recv=%d", burst.cumBytesSend, burst.cumBytesRecv)
	}
	if burst.cumPacketsSend != 4 || burst.cumPacketsRecv != 3 {
		t.Fatalf("unexpected packets: send=%d recv=%d", burst.cumPacketsSend, burst.cumPacketsRecv)
	}
	if burst.cumRetrans != 2 {
		t.Fatalf("unexpected retrans %d", burst.cumRetrans)
	}
	if burst.cumRTT != 500 || burst.cumRTTVar != 50 || burst.cumRTTCount != 1 {
		t.Fatalf("unexpected rtt: sum=%d var=%d count=%d", burst.cumRTT, burst.cumRTTVar, burst.cumRTTCount)
	}
	if burst.pid != 42 || burst.direction != directionOutgoing {
		t.Fatalf("unexpected attribution: pid=%d direction=%s", burst.pid, burst.direction)
	}

	d.mu.Lock()
	_, ok = d.replayPendingLocked(1, info, time.Now())
	d.mu.Unlock()
	if ok {
		t.Fatal("pending should have been cleared after replay")
	}
}

func TestPendingEviction(t *testing.T) {
	d := testDecoder()
	d.pendingLimit = 2
	for i := uint64(1); i <= 3; i++ {
		d.accumulatePending(i, 10, 0, 0, 0, 0, 0, 0)
	}
	d.mu.Lock()
	left := len(d.pending)
	evicted := d.pendingEvicted
	d.mu.Unlock()
	if left != 2 || evicted != 1 {
		t.Fatalf("left=%d evicted=%d, want 2/1", left, evicted)
	}
}

func TestUDPSnapshotDirection(t *testing.T) {
	s := &listenSnapshot{endpoints: map[string]struct{}{
		"0.0.0.0:53":      {},
		"192.168.1.1:161": {},
	}}
	if !s.isIncoming("10.0.0.5", 53, true) {
		t.Fatal("wildcard 0.0.0.0:53 should match any local ip")
	}
	if !s.isIncoming("192.168.1.1", 161, true) {
		t.Fatal("exact bound endpoint should match")
	}
	if s.isIncoming("192.168.1.2", 161, true) {
		t.Fatal("non-matching ip should not match")
	}
	if s.isIncoming("10.0.0.5", 50000, true) {
		t.Fatal("ephemeral port should not be incoming")
	}
	if s.isIncoming("10.0.0.5", 0, true) {
		t.Fatal("port 0 should not be incoming")
	}
}

func TestTCPSnapshotDirectionAllowsEphemeralListener(t *testing.T) {
	s := &listenSnapshot{endpoints: map[string]struct{}{
		"0.0.0.0:50000": {},
	}}
	d := &etwDecoder{tcpSrv: s}
	if got := d.inferTCPDirection("10.0.0.5", 50000); got != directionIncoming {
		t.Fatalf("listener direction = %q, want incoming", got)
	}
	if got := d.inferTCPDirection("10.0.0.5", 50001); got != directionOutgoing {
		t.Fatalf("non-listener direction = %q, want outgoing", got)
	}
}

// ---------------------------------------------------------------------------
// Fast-path layout tests: synthetic user data per event version. These pin the
// field offsets that the hot decoder relies on, so OS version regressions show
// up in CI before they hit production.
// ---------------------------------------------------------------------------

func appendU32(b []byte, v ...uint32) []byte {
	for _, x := range v {
		var tmp [4]byte
		binary.LittleEndian.PutUint32(tmp[:], x)
		b = append(b, tmp[:]...)
	}
	return b
}

func appendU64(b []byte, v uint64) []byte {
	var tmp [8]byte
	binary.LittleEndian.PutUint64(tmp[:], v)
	return append(b, tmp[:]...)
}

func v4SockAddr(ip string, port uint16) []byte {
	b := make([]byte, 16)
	binary.LittleEndian.PutUint16(b[0:2], 2) // AF_INET
	binary.BigEndian.PutUint16(b[2:4], port)
	copy(b[4:8], net.ParseIP(ip).To4())
	return b
}

func build1332(v uint8, tcb uint64, bytesSent, srtt, rttVar uint32, local, remote []byte) []byte {
	b := appendU64(nil, tcb)
	b = appendU32(b, 0) // Cwnd
	b = appendU32(b, 0) // SndWnd
	b = appendU32(b, bytesSent)
	b = appendU32(b, 0) // SeqNo
	b = appendU32(b, srtt)
	b = appendU32(b, rttVar)
	b = appendU32(b, 0) // RTO
	if v >= 1 {
		b = appendU32(b, 0) // RcvWnd
	}
	if v >= 2 {
		b = appendU32(b, 0) // PacingRate
	}
	if v >= 3 {
		b = appendU32(b, 0, 0, 0, 0, 0) // TcpState CongestionState SndUna SndMax RecoveryMax
	}
	if v >= 4 {
		b = appendU32(b, 0, 0) // RcvBufSet MaxRcvBuf
	}
	if v >= 5 {
		b = appendU32(b, uint32(len(local)))
		b = append(b, local...)
		b = appendU32(b, uint32(len(remote)))
		b = append(b, remote...)
	}
	return b
}

func testRec(version uint8) *eventRecord {
	rec := &eventRecord{}
	rec.EventHeader.ProviderID = tcpipProviderGUID
	rec.EventHeader.Descriptor.Version = version
	return rec
}

func testDecoder() *etwDecoder {
	return &etwDecoder{
		tcbs:    make(map[uint64]*tcbInfo),
		pending: make(map[uint64]*pendingStats),
	}
}

func registerTCP(d *etwDecoder, tcb uint64) {
	d.tcbs[tcb] = &tcbInfo{
		srcIP: "127.0.0.1", srcPort: 50000,
		dstIP: "10.0.0.1", dstPort: 443,
		pid: 42, family: "IPv4", direction: directionOutgoing,
	}
}

func TestFastDecode1332AllVersions(t *testing.T) {
	versions := []struct {
		v        uint8
		withAddr bool
	}{
		{0, false}, {1, false}, {2, false}, {3, false}, {4, false}, {5, true},
	}
	local := v4SockAddr("127.0.0.1", 50000)
	remote := v4SockAddr("10.0.0.1", 443)

	for _, tc := range versions {
		t.Run(fmt.Sprintf("v%d", tc.v), func(t *testing.T) {
			d := testDecoder()
			data := build1332(tc.v, 0x1234, 4096, 2500, 100, local, remote)
			if !tc.withAddr {
				registerTCP(d, 0x1234)
			}

			evs := d.onTCPSendFast(testRec(tc.v), data)
			if evs.len() != 1 {
				t.Fatalf("expected 1 event, got %d", evs.len())
			}
			ev := evs.at(0)
			if ev.bytes != 4096 || ev.rtt != 2500 || ev.rttVar != 100 {
				t.Fatalf("wrong fields: bytes=%d rtt=%d rttVar=%d", ev.bytes, ev.rtt, ev.rttVar)
			}
			if ev.srcIP != "127.0.0.1" || ev.dstIP != "10.0.0.1" ||
				ev.srcPort != 50000 || ev.dstPort != 443 {
				t.Fatalf("wrong tuple: %s:%d -> %s:%d", ev.srcIP, ev.srcPort, ev.dstIP, ev.dstPort)
			}
			// v5 must register the TCB from inline addresses.
			if tc.withAddr {
				d.mu.Lock()
				_, registered := d.tcbs[0x1234]
				d.mu.Unlock()
				if !registered {
					t.Fatal("v5 should register tcb from inline addresses")
				}
			}
		})
	}
}

func TestFastDecode1074(t *testing.T) {
	for _, tc := range []struct {
		v       uint8
		packets uint32
	}{
		{0, 0},
		{1, 7},
	} {
		t.Run(fmt.Sprintf("v%d", tc.v), func(t *testing.T) {
			d := testDecoder()
			registerTCP(d, 0x99)
			data := appendU64(nil, 0x99)
			data = appendU32(data, 2048) // NumBytes
			data = appendU32(data, 0)    // SeqNo
			if tc.v >= 1 {
				data = appendU32(data, tc.packets) // NumPkt
			}
			evs := d.onTCPRecvFast(testRec(tc.v), data)
			if evs.len() != 1 {
				t.Fatalf("expected 1 event, got %d", evs.len())
			}
			ev := evs.at(0)
			if ev.bytes != 2048 || ev.packets != uint64(tc.packets) {
				t.Fatalf("wrong fields: bytes=%d packets=%d", ev.bytes, ev.packets)
			}
		})
	}
}

func TestFastDecode1073(t *testing.T) {
	d := testDecoder()
	registerTCP(d, 0x77)
	// Tcb Cwnd SSThresh RttSample NumBytes SeqNo SndUna Round SRTT RTO DWnd BaseRtt DupAckCount
	data := appendU64(nil, 0x77)
	data = appendU32(data, 0, 0, 1000, 8192, 0, 0, 0, 3000, 0, 0, 0, 0)
	evs := d.onTCPSendLegacyFast(testRec(0), data)
	if evs.len() != 1 {
		t.Fatalf("expected 1 event, got %d", evs.len())
	}
	ev := evs.at(0)
	if ev.bytes != 8192 || ev.rtt != 3000 {
		t.Fatalf("wrong fields: bytes=%d rtt=%d", ev.bytes, ev.rtt)
	}
}

func TestFastDecode1341(t *testing.T) {
	d := testDecoder()
	registerTCP(d, 0x55)
	data := appendU64(nil, 0x55)
	data = appendU32(data, 1000, 50, 2500) // RttSample RttVar SRTT
	evs := d.onRTTFast(testRec(0), data)
	if evs.len() != 1 || evs.at(0).rtt != 2500 || evs.at(0).rttVar != 50 {
		t.Fatalf("wrong rtt event: %+v", evs)
	}
}

func TestFastDecode1351(t *testing.T) {
	d := testDecoder()
	registerTCP(d, 0x44)
	data := appendU64(nil, 0x44)
	data = appendU32(data, 0, 3, 2500, 4000) // SndUna RexmitCount SRTT RTO
	evs := d.onRetransmitFast(testRec(0), data)
	if evs.len() != 1 || evs.at(0).retrans != 3 {
		t.Fatalf("wrong retransmit event: %+v", evs)
	}
}

func TestFastDecode1051(t *testing.T) {
	d := testDecoder()
	registerTCP(d, 0x33)
	data := appendU32(nil, 3, 5, 0) // OldState SYN_SENT NewState ESTABLISHED SndNxt
	data = appendU64(data, 0x33)
	evs := d.onTCPStateChangeFast(testRec(0), data)
	if evs.len() != 1 || evs.at(0).oldState != 3 || evs.at(0).newState != 5 {
		t.Fatalf("wrong state change: %+v", evs)
	}
}

func TestFastDecodeUDP(t *testing.T) {
	local := v4SockAddr("127.0.0.1", 50000)
	remote := v4SockAddr("10.0.0.2", 53)
	data := appendU64(nil, 0x100)  // Endpoint
	data = appendU32(data, 3, 512) // NumMessages NumBytes
	data = appendU32(data, 16)
	data = append(data, local...)
	data = appendU32(data, 16)
	data = append(data, remote...)
	data = appendU32(data, 77) // Pid

	for _, kind := range []eventKind{evUDPDataSend, evUDPDataRecv} {
		d := testDecoder()
		evs := d.onUDPFast(testRec(0), data, kind)
		if evs.len() != 1 {
			t.Fatalf("kind %v: expected 1 event, got %d", kind, evs.len())
		}
		ev := evs.at(0)
		if ev.bytes != 512 || ev.packets != 3 || ev.pid != 77 {
			t.Fatalf("wrong udp fields: %+v", ev)
		}
		if ev.srcIP != "127.0.0.1" || ev.srcPort != 50000 || ev.dstIP != "10.0.0.2" || ev.dstPort != 53 {
			t.Fatalf("wrong udp tuple: %+v", ev)
		}
	}
}

func TestFastDecodeMalformedBumpsError(t *testing.T) {
	var errCnt atomic.Uint64
	d := testDecoder()
	d.parseErrors = &errCnt

	// 1332 truncated: only Tcb present.
	evs := d.onTCPSendFast(testRec(0), appendU64(nil, 0x1))
	if evs.len() != 0 {
		t.Fatal("truncated event should not decode")
	}
	// UDP with invalid address length.
	data := appendU64(nil, 0x1)
	data = appendU32(data, 1, 10, 7) // NumMessages NumBytes LocalSockAddrLength=7 (invalid)
	data = append(data, make([]byte, 7)...)
	if evs := d.onUDPFast(testRec(0), data, evUDPDataSend); evs.len() != 0 {
		t.Fatal("invalid addr length should not decode")
	}
	if errCnt.Load() == 0 {
		t.Fatal("expected parse errors to be counted")
	}
}

// TestETWSessionLifecycle verifies that a session started by the collector is
// reliably stopped again (no lingering ETW sessions after cleanup). Requires an
// elevated shell; skipped unless DK_WINNETFLOW_SMOKE=1.
func TestETWSessionLifecycle(t *testing.T) {
	if os.Getenv("DK_WINNETFLOW_SMOKE") != "1" {
		t.Skip("set DK_WINNETFLOW_SMOKE=1 to run the ETW smoke test")
	}

	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	agg.includeLoopback = true
	col, err := newETWCollector(agg, defaultETWConfig())
	if err != nil {
		t.Fatalf("newETWCollector: %v", err)
	}
	if err := col.start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	col.stop()

	// Starting again must succeed immediately; a lingering session would
	// surface as ERROR_ALREADY_EXISTS through the stale-session path.
	col2, err := newETWCollector(newFlowAggregator(time.Second, defaultMaxFlows), defaultETWConfig())
	if err != nil {
		t.Fatalf("newETWCollector: %v", err)
	}
	if err := col2.start(); err != nil {
		t.Fatalf("second start: %v", err)
	}
	col2.stop()
}

// TestETWSmokeCollectTCP validates the full pipeline against a real ETW
// session: start the collector, generate loopback TCP traffic, and verify flow
// bytes are aggregated. It requires an elevated shell and is skipped unless
// DK_WINNETFLOW_SMOKE=1.
func TestETWSmokeCollectTCP(t *testing.T) {
	if os.Getenv("DK_WINNETFLOW_SMOKE") != "1" {
		t.Skip("set DK_WINNETFLOW_SMOKE=1 to run the ETW smoke test")
	}

	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	agg.includeLoopback = true
	col, err := newETWCollector(agg, defaultETWConfig())
	if err != nil {
		t.Fatalf("newETWCollector: %v", err)
	}
	if err := col.start(); err != nil {
		t.Fatalf("start ETW collector: %v", err)
	}
	defer col.stop()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	// The collector normally snapshots listeners at startup. This listener was
	// created after start so refresh once to exercise server-side direction
	// attribution in the same smoke test.
	c := col.(*etwCollector)
	c.tcpSrv.refresh("tcp")
	serverPort := strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				buf := make([]byte, 1024)
				for {
					n, err := c.Read(buf)
					if n > 0 {
						_, _ = c.Write(buf[:n])
					}
					if err != nil {
						_ = c.Close()
						return
					}
				}
			}(conn)
		}
	}()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	payload := make([]byte, 4096)
	for i := 0; i < len(payload); i++ {
		payload[i] = byte(i)
	}
	for i := 0; i < 8; i++ {
		if _, err := conn.Write(payload); err != nil {
			t.Fatalf("write: %v", err)
		}
		buf := make([]byte, len(payload))
		if _, err := readFull(conn, buf); err != nil {
			t.Fatalf("read: %v", err)
		}
	}
	_ = conn.Close()

	deadline := time.Now().Add(30 * time.Second)
	foundIncoming, foundOutgoing := false, false
	establishedIncoming, establishedOutgoing := false, false
	for time.Now().Before(deadline) {
		pts := agg.flush()
		for _, pt := range pts {
			written := pt.Fields().Get("bytes_written").GetI()
			read := pt.Fields().Get("bytes_read").GetI()
			direction := pt.Tags().Get("direction").GetS()
			if pt.Tags().Get("src_port").GetS() == serverPort && direction == directionIncoming {
				if written > 0 && read > 0 {
					foundIncoming = true
				}
				if pt.Fields().Get("tcp_established").GetI() > 0 {
					establishedIncoming = true
				}
			}
			if pt.Tags().Get("dst_port").GetS() == serverPort && direction == directionOutgoing {
				if written > 0 && read > 0 {
					foundOutgoing = true
				}
				if pt.Fields().Get("tcp_established").GetI() > 0 {
					establishedOutgoing = true
				}
			}
			if written > 0 && read > 0 {
				t.Logf("smoke flow direction=%s %s:%s -> %s:%s written=%d read=%d",
					direction,
					pt.Tags().Get("src_ip").GetS(), pt.Tags().Get("src_port").GetS(),
					pt.Tags().Get("dst_ip").GetS(), pt.Tags().Get("dst_port").GetS(),
					written, read)
			}
		}
		if foundIncoming && foundOutgoing && establishedIncoming && establishedOutgoing {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	drops := uint64(0)
	if ec, ok := col.(*etwCollector); ok {
		drops = ec.dropCount.Load()
	}
	t.Fatalf("missing flow/lifecycle within 30s (incoming=%t outgoing=%t established_incoming=%t established_outgoing=%t dropped_events=%d)",
		foundIncoming, foundOutgoing, establishedIncoming, establishedOutgoing, drops)
}

// TestL4TDHPathWithRealEvents validates the L4 TDH data-event fallbacks
// (TCPSend/Recv/RTT/Retransmit/StateChange/UDP) against real events captured
// on this machine. On this OS the fast path handles every data-event version,
// so this is the only way to prove the TDH property names
// (Tcb/BytesSent/NumBytes/NumPkt/SRtt/RttVar/RexmitCount/OldState/NewState/
// NumMessages/Pid/LocalSockAddr/RemoteSockAddr) resolve against a real
// manifest. Skipped unless DK_WINNETFLOW_SMOKE=1.
func TestL4TDHPathWithRealEvents(t *testing.T) {
	if os.Getenv("DK_WINNETFLOW_SMOKE") != "1" {
		t.Skip("set DK_WINNETFLOW_SMOKE=1 to run the ETW smoke test")
	}

	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	agg.includeLoopback = true
	col, err := newETWCollector(agg, defaultETWConfig())
	if err != nil {
		t.Fatalf("newETWCollector: %v", err)
	}
	c := col.(*etwCollector)

	s := &etwSession{name: "datakit-winnetflow-tdh"}
	if err := s.start(defaultETWConfig(), &tcpipProviderGUID, 4, matchAnyKeywords, wantedEventIDs, c); err != nil {
		t.Fatalf("start TCPIP ETW session: %v", err)
	}
	c.session = s
	procDone := make(chan struct{})
	go func() {
		defer close(procDone)
		_ = processTrace(&s.traceHandle, 1)
	}()
	defer func() {
		s.stop()
		<-procDone
		c.session = nil
	}()

	// TCP echo traffic.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(cn net.Conn) {
				buf := make([]byte, 1024)
				for {
					n, err := cn.Read(buf)
					if n > 0 {
						_, _ = cn.Write(buf[:n])
					}
					if err != nil {
						_ = cn.Close()
						return
					}
				}
			}(conn)
		}
	}()
	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	payload := make([]byte, 4096)
	for i := range payload {
		payload[i] = byte(i)
	}
	for i := 0; i < 4; i++ {
		if _, err := conn.Write(payload); err != nil {
			t.Fatalf("write: %v", err)
		}
		buf := make([]byte, len(payload))
		if _, err := readFull(conn, buf); err != nil {
			t.Fatalf("read: %v", err)
		}
	}
	_ = conn.Close()

	// UDP traffic so events 1169/1170 are captured.
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("resolve udp: %v", err)
	}
	udpLn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	defer udpLn.Close()
	udpClient, err := net.DialUDP("udp", nil, udpLn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		t.Fatalf("dial udp: %v", err)
	}
	for i := 0; i < 8; i++ {
		_, _ = udpClient.Write([]byte("udp payload"))
		buf := make([]byte, 64)
		_, _, _ = udpLn.ReadFromUDP(buf)
	}
	udpClient.Close()

	// Drain the raw channel with a quiet period.
	raws := make([]*rawEvent, 0, 512)
	deadline := time.Now().Add(15 * time.Second)
draining:
	for time.Now().Before(deadline) {
		select {
		case raw := <-c.rawCh:
			raws = append(raws, &raw)
		default:
			// Wait until TCP data (1332/1074) and UDP data (1169/1170) events
			// have all been seen; UDP events can lag the TCP burst.
			seen := map[uint16]bool{}
			for _, r := range raws {
				seen[r.eventID] = true
			}
			if seen[1017] && seen[1033] && seen[1332] && seen[1074] && seen[1169] && seen[1170] {
				break draining
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	if len(raws) < 10 {
		t.Fatalf("captured too few events: %d", len(raws))
	}
	byID := map[uint16]int{}
	for _, raw := range raws {
		byID[raw.eventID]++
	}
	t.Logf("captured events: %v", byID)

	// Feed every event through the TDH decoders (control events use the same
	// handlers as production; data events are forced down the TDH path).
	for _, raw := range raws {
		rec := rawToEventRecord(raw)
		var evs flowEventBatch
		switch raw.eventID {
		case 1017:
			evs = c.decoder.onTCPAccept(rec)
		case 1033:
			evs = c.decoder.onTCPConnect(rec)
		case 1034, 1045:
			evs = c.decoder.onTCPConnectFail(rec)
		case 1040, 1043:
			evs = c.decoder.onTCPClose(rec)
		case 1184:
			evs = c.decoder.onTCPRST(rec)
		case 1300:
			evs = c.decoder.onTCPRundown(rec)
		case 1051:
			evs = c.decoder.onTCPStateChangeTDH(rec)
		case 1073:
			evs = c.decoder.onTCPSendLegacyTDH(rec)
		case 1332:
			evs = c.decoder.onTCPSendTDH(rec)
		case 1074:
			evs = c.decoder.onTCPRecvTDH(rec)
		case 1341:
			evs = c.decoder.onRTTTDH(rec)
		case 1351:
			evs = c.decoder.onRetransmitTDH(rec)
		case 1169:
			evs = c.decoder.onUDPTDH(rec, evUDPDataSend)
		case 1170:
			evs = c.decoder.onUDPTDH(rec, evUDPDataRecv)
		}
		for i := 0; i < evs.len(); i++ {
			agg.onEvent(evs.at(i))
		}
	}

	if n := c.parseErr.Load(); n > 0 {
		t.Fatalf("L4 TDH path produced %d parse errors (property names may not match the manifest)", n)
	}
	pts := agg.flush()
	foundTCP, foundUDP, foundIncomingEstablished := false, false, false
	for _, pt := range pts {
		written := pt.Fields().Get("bytes_written").GetI()
		read := pt.Fields().Get("bytes_read").GetI()
		switch pt.Tags().Get("transport").GetS() {
		case "tcp":
			if written > 0 && read > 0 {
				foundTCP = true
			}
			if pt.Tags().Get("direction").GetS() == directionIncoming &&
				pt.Fields().Get("tcp_established").GetI() > 0 {
				foundIncomingEstablished = true
			}
		case "udp":
			if written > 0 {
				foundUDP = true
			}
		}
	}
	if !foundTCP {
		t.Fatalf("L4 TDH path produced no TCP flow with bytes (points=%d)", len(pts))
	}
	if !foundUDP {
		t.Fatalf("L4 TDH path produced no UDP flow with bytes (points=%d events=%v)", len(pts), byID)
	}
	if !foundIncomingEstablished {
		t.Fatalf("L4 TDH path produced no incoming TCP establishment (events=%v)", byID)
	}
	t.Logf("L4 TDH path ok: tcp=%v udp=%v incoming_established=%v (events=%d)",
		foundTCP, foundUDP, foundIncomingEstablished, len(raws))
}

func readFull(conn net.Conn, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := conn.Read(buf[total:])
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

// TestETWStressTCP pushes sustained concurrent loopback traffic through the
// collector and reports byte accuracy plus dropped ETW events. It requires an
// elevated shell and is skipped unless DK_WINNETFLOW_STRESS=1.
func TestETWStressTCP(t *testing.T) {
	if os.Getenv("DK_WINNETFLOW_STRESS") != "1" {
		t.Skip("set DK_WINNETFLOW_STRESS=1 to run the ETW stress test")
	}

	const (
		conns      = 200
		chunkSize  = 16384
		chunksPer  = 32 // 512 KB per connection
		udpSockets = 50
		udpMsgs    = 200
		udpSize    = 512
	)

	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	agg.includeLoopback = true
	col, err := newETWCollector(agg, defaultETWConfig())
	if err != nil {
		t.Fatalf("newETWCollector: %v", err)
	}
	if err := col.start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer col.stop()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				buf := make([]byte, 32*1024)
				for {
					n, err := c.Read(buf)
					if n > 0 {
						if _, werr := c.Write(buf[:n]); werr != nil {
							_ = c.Close()
							return
						}
					}
					if err != nil {
						_ = c.Close()
						return
					}
				}
			}(c)
		}
	}()

	payload := make([]byte, chunkSize)
	for i := range payload {
		payload[i] = byte(i)
	}

	start := time.Now()
	expectedSend := int64(conns) * chunkSize * chunksPer
	var wg sync.WaitGroup
	for i := 0; i < conns; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, err := net.Dial("tcp", ln.Addr().String())
			if err != nil {
				return
			}
			defer c.Close()
			rbuf := make([]byte, chunkSize)
			for j := 0; j < chunksPer; j++ {
				if _, err := c.Write(payload); err != nil {
					return
				}
				if _, err := readFull(c, rbuf); err != nil {
					return
				}
			}
		}()
	}

	// UDP burst in parallel.
	udpConn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("udp listen: %v", err)
	}
	defer udpConn.Close()
	udpBuf := make([]byte, udpSize)
	var udpWG sync.WaitGroup
	expectedUDPSend := int64(udpSockets) * udpMsgs * udpSize
	for i := 0; i < udpSockets; i++ {
		udpWG.Add(1)
		go func() {
			defer udpWG.Done()
			c, err := net.Dial("udp", udpConn.LocalAddr().String())
			if err != nil {
				return
			}
			defer c.Close()
			for j := 0; j < udpMsgs; j++ {
				if _, err := c.Write(udpBuf); err != nil {
					return
				}
			}
		}()
	}

	// Poll the aggregator while the load runs, then drain for a few seconds.
	var observedSend, observedRecv int64
	stopPoll := make(chan struct{})
	var pollWG sync.WaitGroup
	pollWG.Add(1)
	go func() {
		defer pollWG.Done()
		for {
			select {
			case <-stopPoll:
				return
			default:
			}
			pts := agg.flush()
			for _, pt := range pts {
				observedSend += pt.Fields().Get("bytes_written").GetI()
				observedRecv += pt.Fields().Get("bytes_read").GetI()
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	wg.Wait()
	udpWG.Wait()
	time.Sleep(5 * time.Second) // drain trailing events

	close(stopPoll)
	pollWG.Wait()
	// Final flush after the poller stopped.
	pts := agg.flush()
	for _, pt := range pts {
		observedSend += pt.Fields().Get("bytes_written").GetI()
		observedRecv += pt.Fields().Get("bytes_read").GetI()
	}

	elapsed := time.Since(start)
	expectedTotal := 4*expectedSend + 2*expectedUDPSend // TCP counted on both endpoints, UDP once per endpoint pair
	observedTotal := observedSend + observedRecv
	ratio := float64(observedTotal) / float64(expectedTotal)
	ec := col.(*etwCollector)
	drops := ec.dropCount.Load()
	decoded := ec.decoded.Load()

	t.Logf("stress: conns=%d sent=%dMB expected_total=%dMB observed_total=%dMB ratio=%.2f%% drops=%d decoded_events=%d elapsed=%s",
		conns, expectedSend>>20, expectedTotal>>20, observedTotal>>20, ratio*100, drops, decoded, elapsed.Round(time.Millisecond))
	t.Logf("stress: observed_send=%d observed_recv=%d", observedSend, observedRecv)

	st := ec.stats()
	if drops > 0 || st.session.eventsLost > 0 || st.session.realTimeBuffersLost > 0 {
		t.Fatalf("stress event loss: channel_dropped=%d events_lost=%d realtime_buffers_lost=%d",
			drops, st.session.eventsLost, st.session.realTimeBuffersLost)
	}
	if ratio < 0.95 {
		t.Fatalf("byte accuracy too low: %.2f%%", ratio*100)
	}
	if ratio > 1.10 {
		t.Fatalf("byte accuracy too high: %.2f%%", ratio*100)
	}
}
