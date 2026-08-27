// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows

package winnetflow

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf16"
	"unsafe"
)

// Synthetic HttpService event builders. Layouts verified against
// Windows 10.0.26200 (see dev.md and the probe capture used to derive them).

func httpTestCollector(maxReq int) (*httpCollector, *httpAggregator) {
	agg := newHTTPAggregator(1024)
	agg.includeLoopback = true
	c, err := newHTTPCollector(agg, defaultETWConfig(), defaultHTTPPathLimit, maxReq)
	if err != nil {
		panic(err)
	}
	return c.(*httpCollector), agg
}

func httpRaw(id uint16, act, rel string, data []byte) *httpRawEvent {
	raw := &httpRawEvent{
		eventID:    id,
		ts:         time.Now(),
		pid:        4242,
		activityID: parseGUID(act),
		userData:   data,
	}
	if rel != "" {
		raw.hasRelated = true
		raw.relatedID = parseGUID(rel)
	}
	return raw
}

func trafficPID(t *testing.T, out []byte, key string) uint32 {
	t.Helper()
	prefix := key + "="
	for _, field := range strings.Fields(string(out)) {
		if value, ok := strings.CutPrefix(field, prefix); ok {
			pid, err := strconv.ParseUint(value, 10, 32)
			if err != nil || pid == 0 {
				t.Fatalf("invalid %s in traffic output %q", key, strings.TrimSpace(string(out)))
			}
			return uint32(pid)
		}
	}
	t.Fatalf("missing %s in traffic output %q", key, strings.TrimSpace(string(out)))
	return 0
}

func httpSockAddr(family uint16, port uint16, ip [4]byte) []byte {
	b := make([]byte, 16)
	binary.LittleEndian.PutUint16(b[0:2], family)
	binary.BigEndian.PutUint16(b[2:4], port)
	copy(b[4:8], ip[:])
	return b
}

// httpEvent21 builds a ConnConnect event with the given local/remote IPv4
// endpoints (48 bytes, verified layout).
func httpEvent21(connID string, localPort, remotePort uint16) *httpRawEvent {
	data := make([]byte, 48)
	binary.LittleEndian.PutUint64(data[0:8], 0xFFFFBD0B13F41500)
	binary.LittleEndian.PutUint32(data[8:12], 16)
	copy(data[12:28], httpSockAddr(2, localPort, [4]byte{127, 0, 0, 1}))
	binary.LittleEndian.PutUint32(data[28:32], 16)
	copy(data[32:48], httpSockAddr(2, remotePort, [4]byte{127, 0, 0, 1}))
	return httpRaw(21, connID, "", data)
}

func httpEvent1(connID, reqID string, remotePort uint16) *httpRawEvent {
	data := make([]byte, 36)
	binary.LittleEndian.PutUint64(data[0:8], 0x02000040090000ff)
	binary.LittleEndian.PutUint64(data[8:16], 0x01000030090000ff)
	binary.LittleEndian.PutUint32(data[16:20], 16)
	copy(data[20:36], httpSockAddr(2, remotePort, [4]byte{127, 0, 0, 1}))
	return httpRaw(1, connID, reqID, data)
}

func httpEvent2(reqID string, verb uint32, uri string) *httpRawEvent {
	u16 := utf16.Encode([]rune(uri + "\x00"))
	raw := make([]byte, 12+len(u16)*2)
	binary.LittleEndian.PutUint64(raw[0:8], 0xFFFFBD0B65F64010)
	binary.LittleEndian.PutUint32(raw[8:12], verb)
	for i, u := range u16 {
		binary.LittleEndian.PutUint16(raw[12+i*2:], u)
	}
	return httpRaw(2, reqID, "", raw)
}

func httpEvent3(reqID string, siteID uint32, queue, uri string, status uint32) *httpRawEvent {
	queue16 := utf16.Encode([]rune(queue + "\x00"))
	uri16 := utf16.Encode([]rune(uri + "\x00"))
	raw := make([]byte, 20+len(queue16)*2+len(uri16)*2+4)
	binary.LittleEndian.PutUint64(raw[0:8], 0xFFFFBD0B65F64010)
	binary.LittleEndian.PutUint64(raw[8:16], 0x02000040090000ff)
	binary.LittleEndian.PutUint32(raw[16:20], siteID)
	off := 20
	for _, u := range queue16 {
		binary.LittleEndian.PutUint16(raw[off:], u)
		off += 2
	}
	for _, u := range uri16 {
		binary.LittleEndian.PutUint16(raw[off:], u)
		off += 2
	}
	binary.LittleEndian.PutUint32(raw[off:], status)
	return httpRaw(3, reqID, "", raw)
}

func httpEvent8(reqID string, status uint16, verb string) *httpRawEvent {
	const trailerSize = 10 // HeaderLength u32, EntityChunkCount u16, CachePolicy u32.
	raw := make([]byte, 18+len(verb)+1+trailerSize)
	binary.LittleEndian.PutUint64(raw[0:8], 0x02000040090000ff)
	binary.LittleEndian.PutUint64(raw[8:16], 0x01000030090000ff)
	binary.LittleEndian.PutUint16(raw[16:18], status)
	copy(raw[18:], verb)
	trailer := 18 + len(verb) + 1
	binary.LittleEndian.PutUint16(raw[trailer+4:], 1)
	return httpRaw(8, reqID, "", raw)
}

func httpEvent12(reqID string, status uint16) *httpRawEvent {
	raw := make([]byte, 10)
	binary.LittleEndian.PutUint64(raw[0:8], 0x02000040090000ff)
	binary.LittleEndian.PutUint16(raw[8:10], status)
	return httpRaw(12, reqID, "", raw)
}

func httpEvent16(reqID string, bytesSent uint32) *httpRawEvent {
	raw := make([]byte, 16)
	binary.LittleEndian.PutUint64(raw[0:8], 0xFFFFBD0B65F64010)
	binary.LittleEndian.PutUint32(raw[8:12], 0)
	binary.LittleEndian.PutUint32(raw[12:16], bytesSent)
	return httpRaw(16, reqID, "", raw)
}

func httpEvent24(connID string) *httpRawEvent {
	raw := make([]byte, 8)
	binary.LittleEndian.PutUint64(raw[0:8], 0xFFFFBD0B13F41500)
	return httpRaw(24, connID, "", raw)
}

const (
	testConnID = "11111111-2222-3333-4444-555555555555"
	testReqID  = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
)

func TestUTF16AtRequiresTerminator(t *testing.T) {
	terminated := []byte{'A', 0, 0, 0}
	if got, next, ok := utf16At(terminated, 0); !ok || got != "A" || next != len(terminated) {
		t.Fatalf("terminated UTF-16 = %q next=%d ok=%t", got, next, ok)
	}
	for _, malformed := range [][]byte{
		{'A', 0},
		{'A', 0, 'B'},
	} {
		if got, _, ok := utf16At(malformed, 0); ok {
			t.Fatalf("unterminated UTF-16 decoded as %q", got)
		}
	}
}

func TestHTTPParseMalformedUTF16CountsError(t *testing.T) {
	c, _ := httpTestCollector(16)
	c.decode(httpEvent21(testConnID, 8080, 50000))
	c.decode(httpEvent1(testConnID, testReqID, 50000))
	raw := make([]byte, 15)
	binary.LittleEndian.PutUint32(raw[8:12], 4) // GET
	raw[12], raw[13], raw[14] = 'A', 0, 'B'     // odd and unterminated
	c.decode(httpRaw(evHTTPParse, testReqID, "", raw))
	if got := c.parseErr.Load(); got != 1 {
		t.Fatalf("parse errors = %d, want 1", got)
	}
}

func TestHTTPDrainProcessesQueuedRequest(t *testing.T) {
	c, agg := httpTestCollector(16)
	for _, raw := range []*httpRawEvent{
		httpEvent21(testConnID, 8080, 50000),
		httpEvent1(testConnID, testReqID, 50000),
		httpEvent2(testReqID, 4, "/drain"),
		httpEvent8(testReqID, 200, "GET"),
		httpEvent12(testReqID, 200),
	} {
		c.rawCh <- *raw
	}
	c.drainRawEvents()
	if got := c.decoded.Load(); got != 5 {
		t.Fatalf("decoded = %d, want 5", got)
	}
	pts := agg.flush()
	if len(pts) != 1 || pts[0].Fields().Get("path").GetS() != "/drain" {
		t.Fatalf("drained points = %v", pts)
	}
}

func TestHTTPCollectorStatsIncludeFilterCounts(t *testing.T) {
	agg := newHTTPAggregator(1024)
	agg.onEvent(&httpEvent{srcIP: "0.0.0.0", srcPort: 1, dstIP: "10.0.0.1", dstPort: 2})
	agg.onEvent(&httpEvent{srcIP: "127.0.0.1", srcPort: 1, dstIP: "127.0.0.1", dstPort: 2})
	st := (&httpCollector{agg: agg}).stats()
	if st.invalidFiltered != 1 || st.loopbackFiltered != 1 {
		t.Fatalf("filter stats = invalid:%d loopback:%d, want 1/1", st.invalidFiltered, st.loopbackFiltered)
	}
}

func TestHTTPResponseWarmsServerProcessBeforeCompletion(t *testing.T) {
	c, agg := httpTestCollector(16)
	var warmed []uint32
	agg.warmProcessName = func(pid uint32) { warmed = append(warmed, pid) }
	c.decode(httpEvent21(testConnID, 8080, 50000))
	if len(warmed) != 0 {
		t.Fatalf("client connection PID was warmed as server: %v", warmed)
	}
	c.decode(httpEvent1(testConnID, testReqID, 50000))
	c.decode(httpEvent2(testReqID, 4, "/warm"))
	c.decode(httpEvent8(testReqID, 200, "GET"))
	if len(warmed) != 1 || warmed[0] != 4242 {
		t.Fatalf("response warmed PIDs = %v, want [4242]", warmed)
	}
	c.decode(httpEvent12(testReqID, 200))
	if len(warmed) != 1 {
		t.Fatalf("process warmed %d times in one interval, want 1", len(warmed))
	}
}

func TestHTTPCallbackWarmsServerProcessBeforeDecode(t *testing.T) {
	c, agg := httpTestCollector(16)
	var warmed []uint32
	agg.warmProcessName = func(pid uint32) { warmed = append(warmed, pid) }
	payload := []byte{1}
	rec := eventRecord{
		EventHeader: eventHeader{
			ProcessID:  31337,
			ProviderID: httpProviderGUID,
			Descriptor: eventDescriptor{ID: evHTTPFastResp},
		},
		UserDataLength: uint16(len(payload)),
		UserData:       unsafe.Pointer(&payload[0]),
	}
	c.handleRecord(&rec)
	if len(warmed) != 1 || warmed[0] != rec.EventHeader.ProcessID {
		t.Fatalf("callback warmed PIDs = %v, want [%d]", warmed, rec.EventHeader.ProcessID)
	}
	raw := <-c.rawCh
	releaseEventBuffer(&c.bufPool, raw.userData)
}

func TestHTTPConnectionCachesShortLivedProcess(t *testing.T) {
	procNameCache.Lock()
	oldCache := procNameCache.m
	procNameCache.m = make(map[uint32]cachedProcessName)
	procNameCache.Unlock()
	t.Cleanup(func() {
		procNameCache.Lock()
		procNameCache.m = oldCache
		procNameCache.Unlock()
	})

	cmd := exec.Command("powershell.exe", "-NoProfile", "-Command", "Start-Sleep -Seconds 30")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start short-lived process: %v", err)
	}
	processExited := false
	t.Cleanup(func() {
		if !processExited {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	})

	c, agg := httpTestCollector(16)
	warmer := newProcessNameWarmer()
	defer warmer.stop()
	agg.warmProcessName = warmer.enqueue
	pid := uint32(cmd.Process.Pid)
	c.decode(httpEvent21(testConnID, 8080, 50000))
	c.decode(httpEvent1(testConnID, testReqID, 50000))
	c.decode(httpEvent2(testReqID, 4, "/short-lived"))
	response := httpEvent8(testReqID, 200, "GET")
	response.pid = pid
	c.decode(response)

	var cached string
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		procNameCache.Lock()
		cached = procNameCache.m[pid].name
		procNameCache.Unlock()
		if cached != "" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if cached == "" {
		t.Fatal("response event did not cache process name while process was alive")
	}

	if err := cmd.Process.Kill(); err != nil {
		t.Fatalf("stop short-lived process: %v", err)
	}
	_ = cmd.Wait()
	processExited = true
	if got := processName(pid); got != cached {
		t.Fatalf("cached process after exit = %q, want %q", got, cached)
	}
}

// testReqIDn returns a valid, distinct GUID for the n-th request.
func testReqIDn(n int) string {
	return fmt.Sprintf("aaaaaaaa-bbbb-cccc-dddd-%012x", n)
}

func TestHTTPConnConnectParse(t *testing.T) {
	c, _ := httpTestCollector(1024)
	c.decode(httpEvent21(testConnID, 8080, 51234))

	c.mu.Lock()
	conn, ok := c.conns[parseGUID(testConnID)]
	c.mu.Unlock()
	if !ok {
		t.Fatal("connection not registered")
	}
	if conn.srcIP != "127.0.0.1" || conn.srcPort != 8080 ||
		conn.dstIP != "127.0.0.1" || conn.dstPort != 51234 {
		t.Fatalf("unexpected tuple: %+v", conn)
	}
	if conn.family != "IPv4" {
		t.Fatalf("unexpected family: %+v", conn)
	}
}

func TestHTTPRequestLifecycle(t *testing.T) {
	c, agg := httpTestCollector(1024)
	events := []*httpRawEvent{
		httpEvent21(testConnID, 8080, 51234),
		httpEvent1(testConnID, testReqID, 51234),
		httpEvent2(testReqID, 4, "http://127.0.0.1:8080/api/users?x=1"),
		httpEvent3(testReqID, 0, "<<unnamed>>", "http://127.0.0.1:8080/api/users?x=1", 0),
		httpEvent8(testReqID, 200, "GET"),
		httpEvent12(testReqID, 200),
	}
	var got []*httpEvent
	for _, ev := range events {
		got = append(got, c.decode(ev)...)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 completed http event, got %d", len(got))
	}
	for _, h := range got {
		agg.onEvent(h)
	}
	h := got[0]
	if h.method != "GET" || h.path != "/api/users" || h.status != 200 {
		t.Fatalf("unexpected http event: %+v", h)
	}
	if h.srcIP != "127.0.0.1" || h.srcPort != 8080 ||
		h.dstIP != "127.0.0.1" || h.dstPort != 51234 {
		t.Fatalf("unexpected tuple: %+v", h)
	}
	if h.family != "IPv4" || h.pid != 4242 {
		t.Fatalf("unexpected family/pid: %+v", h)
	}

	c.mu.Lock()
	reqCount := len(c.reqs)
	c.mu.Unlock()
	if reqCount != 0 {
		t.Fatalf("request map should be empty after completion, got %d", reqCount)
	}

	pts := agg.flush()
	if len(pts) != 1 {
		t.Fatalf("aggregator should produce 1 point, got %d", len(pts))
	}
	if pts[0].Name() != "httpflow" {
		t.Fatalf("measurement = %s, want httpflow", pts[0].Name())
	}
}

func TestHTTPKeepAliveTwoRequests(t *testing.T) {
	c, _ := httpTestCollector(1024)
	var got []*httpEvent
	for i := 0; i < 2; i++ {
		c.decode(httpEvent21(testConnID, 8080, 51234))
		reqID := testReqIDn(i + 1)
		got = append(got, c.decode(httpEvent1(testConnID, reqID, 51234))...)
		got = append(got, c.decode(httpEvent2(reqID, 4, "http://127.0.0.1:8080/p"))...)
		got = append(got, c.decode(httpEvent3(reqID, 0, "<<unnamed>>", "http://127.0.0.1:8080/p", 0))...)
		got = append(got, c.decode(httpEvent8(reqID, 200, "GET"))...)
		got = append(got, c.decode(httpEvent12(reqID, 200))...)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 completed events, got %d", len(got))
	}
	c.mu.Lock()
	connCount := len(c.conns)
	c.mu.Unlock()
	if connCount != 1 {
		t.Fatalf("keep-alive connection should remain, got %d conns", connCount)
	}
}

func TestHTTPRequestWithoutConnection(t *testing.T) {
	c, _ := httpTestCollector(1024)
	got := c.decode(httpEvent1("99999999-0000-0000-0000-000000000000", testReqID, 51234))
	if len(got) != 0 {
		t.Fatalf("request without connection should not complete, got %d events", len(got))
	}
	if c.missedConn.Load() == 0 {
		t.Fatal("missed connection should be counted")
	}
}

func TestHTTPParseBeforeRecvReq(t *testing.T) {
	c, _ := httpTestCollector(1024)
	c.decode(httpEvent2(testReqID, 4, "http://h/p"))
	if c.missedReq.Load() == 0 {
		t.Fatal("out-of-order parse should count a missed request")
	}
}

func TestHTTPCacheFlow(t *testing.T) {
	c, _ := httpTestCollector(1024)
	events := []*httpRawEvent{
		httpEvent21(testConnID, 443, 51234),
		httpEvent1(testConnID, testReqID, 51234),
		httpEvent2(testReqID, 4, "http://127.0.0.1:443/static/app.js"),
		httpEvent3(testReqID, 0, "<<unnamed>>", "http://127.0.0.1:443/static/app.js", 0),
		httpEvent8(testReqID, 200, "GET"),
		httpEvent16(testReqID, 4096),
	}
	var got []*httpEvent
	for _, ev := range events {
		got = append(got, c.decode(ev)...)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 completed event, got %d", len(got))
	}
	if got[0].bytesWritten != 4096 {
		t.Fatalf("bytes_written = %d, want 4096", got[0].bytesWritten)
	}
	if got[0].status != 200 {
		t.Fatalf("status = %d, want 200", got[0].status)
	}
}

func TestHTTPDoubleComplete(t *testing.T) {
	c, _ := httpTestCollector(1024)
	c.decode(httpEvent21(testConnID, 8080, 51234))
	c.decode(httpEvent1(testConnID, testReqID, 51234))
	c.decode(httpEvent2(testReqID, 4, "http://h/p"))
	c.decode(httpEvent8(testReqID, 200, "GET"))
	got := c.decode(httpEvent12(testReqID, 200))
	got2 := c.decode(httpEvent12(testReqID, 200))
	if len(got) != 1 || len(got2) != 0 {
		t.Fatalf("second terminal event should be ignored: first=%d second=%d", len(got), len(got2))
	}
}

func TestHTTPConnCleanupDropsPending(t *testing.T) {
	c, agg := httpTestCollector(1024)
	c.decode(httpEvent21(testConnID, 8080, 51234))
	c.decode(httpEvent1(testConnID, testReqID, 51234))
	c.decode(httpEvent2(testReqID, 4, "http://h/p"))
	got := c.decode(httpEvent24(testConnID))
	if len(got) != 0 {
		t.Fatalf("cleanup should not complete requests, got %d", len(got))
	}
	c.mu.Lock()
	conns := len(c.conns)
	reqs := len(c.reqs)
	c.mu.Unlock()
	if conns != 0 || reqs != 0 {
		t.Fatalf("cleanup should remove conn and pending reqs: conns=%d reqs=%d", conns, reqs)
	}
	if c.droppedReq.Load() != 1 {
		t.Fatalf("droppedReq = %d, want 1", c.droppedReq.Load())
	}
	if pts := agg.flush(); len(pts) != 0 {
		t.Fatalf("no points expected, got %d", len(pts))
	}
}

func TestHTTPMethodFromAsciiVerb(t *testing.T) {
	c, _ := httpTestCollector(1024)
	c.decode(httpEvent21(testConnID, 8080, 51234))
	c.decode(httpEvent1(testConnID, testReqID, 51234))
	// HTTP_VERB does not define PATCH -> verb value 0, recovered from event 8.
	c.decode(httpEvent2(testReqID, 0, "http://h/patch"))
	c.decode(httpEvent8(testReqID, 200, "PATCH"))
	got := c.decode(httpEvent12(testReqID, 200))
	if len(got) != 1 || got[0].method != "PATCH" {
		t.Fatalf("expected method PATCH from ASCII verb, got %+v", got)
	}
}

func TestHTTPPathTruncated(t *testing.T) {
	c, _ := httpTestCollector(1024)
	// 300 'x' characters.
	buf := make([]byte, 0, 301)
	buf = append(buf, '/')
	for i := 0; i < 300; i++ {
		buf = append(buf, 'x')
	}
	uri := "http://127.0.0.1:8080" + string(buf)

	c.decode(httpEvent21(testConnID, 8080, 51234))
	c.decode(httpEvent1(testConnID, testReqID, 51234))
	c.decode(httpEvent2(testReqID, 4, uri))
	c.decode(httpEvent8(testReqID, 200, "GET"))
	got := c.decode(httpEvent12(testReqID, 200))
	if len(got) != 1 {
		t.Fatalf("expected 1 event, got %d", len(got))
	}
	if !got[0].truncated {
		t.Fatal("truncated flag should be set")
	}
	if len(got[0].path) != defaultHTTPPathLimit {
		t.Fatalf("path length = %d, want %d", len(got[0].path), defaultHTTPPathLimit)
	}
}

func TestHTTPMaxRequestsEviction(t *testing.T) {
	c, _ := httpTestCollector(2)
	c.decode(httpEvent21(testConnID, 8080, 51234))
	for i := 0; i < 3; i++ {
		reqID := testReqIDn(i + 1)
		c.decode(httpEvent1(testConnID, reqID, 51234))
	}
	c.mu.Lock()
	reqs := len(c.reqs)
	c.mu.Unlock()
	if reqs > 2 {
		t.Fatalf("request map should be bounded, got %d", reqs)
	}
	if c.evictedReq.Load() == 0 {
		t.Fatal("eviction should be counted")
	}
}

func TestHTTPConnMapBounded(t *testing.T) {
	c, _ := httpTestCollector(2)
	for i := 0; i < 3; i++ {
		connID := testReqIDn(i + 1)
		c.decode(httpEvent21(connID, 8080, 51234))
	}
	c.mu.Lock()
	conns := len(c.conns)
	c.mu.Unlock()
	if conns > 2 {
		t.Fatalf("conn map should be bounded, got %d", conns)
	}
}

func TestHTTPDeliverSetsAppPool(t *testing.T) {
	c, _ := httpTestCollector(1024)
	c.decode(httpEvent21(testConnID, 8080, 51234))
	c.decode(httpEvent1(testConnID, testReqID, 51234))
	c.decode(httpEvent2(testReqID, 6, "http://127.0.0.1:8080/api"))
	c.decode(httpEvent3(testReqID, 7, "MyAppPool", "http://127.0.0.1:8080/api", 0))

	c.mu.Lock()
	req := c.reqs[parseGUID(testReqID)]
	c.mu.Unlock()
	if req == nil {
		t.Fatal("request not found")
	}
	if req.siteID != 7 || req.appPool != "MyAppPool" {
		t.Fatalf("unexpected app pool/site: %+v", req)
	}
}

func TestHTTPDeliverFillsMissingPath(t *testing.T) {
	c, _ := httpTestCollector(1024)
	c.decode(httpEvent21(testConnID, 8080, 51234))
	c.decode(httpEvent1(testConnID, testReqID, 51234))
	// Event 2 without a parseable URL; event 3 supplies it.
	bad := httpEvent2(testReqID, 4, "http://127.0.0.1:8080/ok")
	bad.userData = bad.userData[:12]
	c.decode(bad)
	c.decode(httpEvent3(testReqID, 0, "<<unnamed>>", "http://127.0.0.1:8080/fromdeliver", 0))
	c.decode(httpEvent8(testReqID, 200, "GET"))
	got := c.decode(httpEvent12(testReqID, 200))
	if len(got) != 1 || got[0].path != "/fromdeliver" {
		t.Fatalf("expected path from deliver event, got %+v", got)
	}
}

// TestHTTPSessionLifecycle verifies the HTTP.sys ETW session can be stopped and
// restarted cleanly. Skipped unless DK_WINNETFLOW_SMOKE=1.
func TestHTTPSessionLifecycle(t *testing.T) {
	if os.Getenv("DK_WINNETFLOW_SMOKE") != "1" {
		t.Skip("set DK_WINNETFLOW_SMOKE=1 to run the ETW smoke test")
	}

	col, err := newHTTPCollector(newHTTPAggregator(1024), defaultETWConfig(),
		defaultHTTPPathLimit, defaultMaxHTTPRequests)
	if err != nil {
		t.Fatalf("newHTTPCollector: %v", err)
	}
	if err := col.start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	col.stop()

	col2, err := newHTTPCollector(newHTTPAggregator(1024), defaultETWConfig(),
		defaultHTTPPathLimit, defaultMaxHTTPRequests)
	if err != nil {
		t.Fatalf("newHTTPCollector: %v", err)
	}
	if err := col2.start(); err != nil {
		t.Fatalf("second start: %v", err)
	}
	col2.stop()
}

// TestETWSmokeCollectHTTP validates the full httpflow pipeline against a real
// HTTP.sys ETW session: start the collector, generate HttpListener traffic,
// and verify aggregated httpflow points. Skipped unless DK_WINNETFLOW_SMOKE=1.
func TestETWSmokeCollectHTTP(t *testing.T) {
	if os.Getenv("DK_WINNETFLOW_SMOKE") != "1" {
		t.Skip("set DK_WINNETFLOW_SMOKE=1 to run the ETW smoke test")
	}

	agg := newHTTPAggregator(defaultMaxHTTPRequests)
	agg.includeLoopback = true
	warmer := newProcessNameWarmer()
	agg.warmProcessName = warmer.enqueue
	defer warmer.stop()
	col, err := newHTTPCollector(agg, defaultETWConfig(), defaultHTTPPathLimit, defaultMaxHTTPRequests)
	if err != nil {
		t.Fatalf("newHTTPCollector: %v", err)
	}
	if err := col.start(); err != nil {
		t.Fatalf("start HTTP.sys ETW collector: %v", err)
	}
	defer col.stop()

	script, err := filepath.Abs(filepath.Join("verify", "gen-httpsvc-traffic.ps1"))
	if err != nil {
		t.Fatalf("resolve traffic script: %v", err)
	}
	port := 18532
	out, err := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass",
		"-File", script, "-Port", strconv.Itoa(port)).CombinedOutput()
	if err != nil {
		t.Fatalf("traffic generation failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "TRAFFIC_OK") {
		t.Fatalf("unexpected traffic output: %s", out)
	}
	t.Logf("traffic: %s", strings.TrimSpace(string(out)))
	serverPID := strconv.FormatUint(uint64(trafficPID(t, out, "server_pid")), 10)
	clientPID := strconv.FormatUint(uint64(trafficPID(t, out, "client_pid")), 10)

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		pts := agg.flush()
		found := false
		for _, pt := range pts {
			if pt.Fields().Get("method").GetS() == "GET" &&
				pt.Fields().Get("status_code").GetI() == 200 {
				pid := pt.Tags().Get("pid").GetS()
				if pid != serverPID || pid == clientPID {
					t.Fatalf("HTTP PID attribution = %s, want server %s (client %s)", pid, serverPID, clientPID)
				}
				process := pt.Tags().Get("process_name").GetS()
				if process == "" || process == "unknown" {
					t.Fatalf("HTTP process attribution unavailable: pid=%s process=%q",
						pt.Tags().Get("pid").GetS(), process)
				}
				t.Logf("http smoke ok: %s:%s -> %s:%s pid=%s process=%s path=%s count=%d",
					pt.Tags().Get("src_ip").GetS(), pt.Tags().Get("src_port").GetS(),
					pt.Tags().Get("dst_ip").GetS(), pt.Tags().Get("dst_port").GetS(),
					pt.Tags().Get("pid").GetS(), pt.Tags().Get("process_name").GetS(),
					pt.Fields().Get("path").GetS(), pt.Fields().Get("count").GetI())
				found = true
				break
			}
		}
		if found {
			// The smoke traffic is controlled (~22 requests over one
			// keep-alive connection), so correlation must succeed almost
			// completely. A systematic failure mode (e.g. request events
			// overtaking connection events) produces misses for every
			// request; allow only a handful of legitimate edge misses.
			var st httpStats
			if c, ok := col.(*httpCollector); ok {
				st = c.stats()
			}
			if st.missedConn+st.missedReq > 5 {
				t.Fatalf("httpflow correlation degraded: missed_conn=%d missed_req=%d decoded=%d",
					st.missedConn, st.missedReq, st.decoded)
			}
			return
		}
		time.Sleep(500 * time.Millisecond)
	}

	var st httpStats
	if c, ok := col.(*httpCollector); ok {
		st = c.stats()
	}
	t.Fatalf("no httpflow point observed within 30s (decoded=%d dropped=%d missed_conn=%d missed_req=%d)",
		st.decoded, st.dropped, st.missedConn, st.missedReq)
}

// TestHTTPTDHPathWithRealEvents validates the TDH fallback decoders (the
// cross-OS safety net) against real events captured on this machine. In
// production the version-0 fast path always wins locally, so this is the only
// way to prove the TDH property names (HttpVerb/Url/SiteId/RequestQueueName/
// StatusCode/Verb/HttpStatus/BytesSent/LocalAddr/RemoteAddr) resolve against a
// real manifest. Skipped unless DK_WINNETFLOW_SMOKE=1.
func TestHTTPTDHPathWithRealEvents(t *testing.T) {
	if os.Getenv("DK_WINNETFLOW_SMOKE") != "1" {
		t.Skip("set DK_WINNETFLOW_SMOKE=1 to run the ETW smoke test")
	}

	agg := newHTTPAggregator(defaultMaxHTTPRequests)
	agg.includeLoopback = true
	c, err := newHTTPCollector(agg, defaultETWConfig(), defaultHTTPPathLimit, defaultMaxHTTPRequests)
	if err != nil {
		t.Fatalf("newHTTPCollector: %v", err)
	}
	hc := c.(*httpCollector)

	// Start the real session but keep the collector's decodeLoop out of the
	// way: drain rawCh into a slice instead.
	s := &etwSession{name: "datakit-winnetflow-http-tdh"}
	if err := s.start(defaultETWConfig(), &httpProviderGUID, httpLevel, httpMatchAnyKeywords, httpEventIDs, hc); err != nil {
		t.Fatalf("start HTTP.sys ETW session: %v", err)
	}
	hc.session = s
	procDone := make(chan struct{})
	go func() {
		defer close(procDone)
		_ = processTrace(&s.traceHandle, 1)
	}()
	defer func() {
		s.stop()
		<-procDone
		hc.session = nil
	}()

	script, err := filepath.Abs(filepath.Join("verify", "gen-httpsvc-traffic.ps1"))
	if err != nil {
		t.Fatalf("resolve traffic script: %v", err)
	}
	port := 18540
	out, err := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass",
		"-File", script, "-Port", strconv.Itoa(port)).CombinedOutput()
	if err != nil {
		t.Fatalf("traffic generation failed: %v\n%s", err, out)
	}

	// Drain the channel until the traffic has been fully delivered.
	raws := make([]*httpRawEvent, 0, 256)
	deadline := time.Now().Add(15 * time.Second)
	quiet := time.Now()
draining:
	for time.Now().Before(deadline) {
		select {
		case raw := <-hc.rawCh:
			raws = append(raws, &raw)
			quiet = time.Now()
		default:
			if len(raws) > 0 && time.Since(quiet) >= 500*time.Millisecond {
				break draining
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	if len(raws) < 10 {
		t.Fatalf("captured too few events: %d", len(raws))
	}
	serverPID := strconv.FormatUint(uint64(trafficPID(t, out, "server_pid")), 10)
	clientPID := strconv.FormatUint(uint64(trafficPID(t, out, "client_pid")), 10)
	byIDPID := map[uint16]map[uint32]int{}
	for _, raw := range raws {
		if byIDPID[raw.eventID] == nil {
			byIDPID[raw.eventID] = map[uint32]int{}
		}
		byIDPID[raw.eventID][raw.pid]++
	}
	t.Logf("traffic=%s event_header_pids=%v", strings.TrimSpace(string(out)), byIDPID)

	// Feed every captured event through the TDH decoders, in order, exactly as
	// the fallback would on an unknown OS version.
	tdhSrc, err := newHTTPCollector(agg, defaultETWConfig(), hc.pathLimit, hc.maxRequests)
	if err != nil {
		t.Fatalf("newHTTPCollector: %v", err)
	}
	tdhC := tdhSrc.(*httpCollector)
	for _, raw := range raws {
		switch raw.eventID {
		case evHTTPConnConnect:
			tdhC.onConnConnectTDH(raw)
		case evHTTPRecvReq:
			tdhC.onRecvReq(raw)
		case evHTTPParse:
			tdhC.onParseTDH(raw)
		case evHTTPDeliver:
			tdhC.onDeliverTDH(raw)
		case evHTTPRecvResp, evHTTPFastResp:
			tdhC.onRespTDH(raw)
		case evHTTPSendComplete, evHTTPCachedAndSend, evHTTPFastSend:
			for _, ev := range tdhC.onSendCompleteTDH(raw, false) {
				agg.onEvent(ev)
			}
		case evHTTPSrvdFrmCache:
			for _, ev := range tdhC.onSendCompleteTDH(raw, true) {
				agg.onEvent(ev)
			}
		case evHTTPConnCleanup:
			tdhC.onConnCleanup(raw)
		}
	}

	st := tdhC.stats()
	if st.parseErrors > 0 {
		t.Fatalf("TDH path produced parse errors: %d (property names may not match the manifest)", st.parseErrors)
	}
	pts := agg.flush()
	if len(pts) == 0 {
		byID := map[uint16]int{}
		for _, raw := range raws {
			byID[raw.eventID]++
		}
		t.Fatalf("TDH path produced no points (events=%d by_id=%v missed_conn=%d missed_req=%d completed=%d parse_errors=%d)",
			len(raws), byID, st.missedConn, st.missedReq, st.completed, st.parseErrors)
	}
	found := false
	for _, pt := range pts {
		if pt.Fields().Get("method").GetS() != "" && pt.Fields().Get("status_code").GetI() > 0 {
			pid := pt.Tags().Get("pid").GetS()
			if pid != serverPID || pid == clientPID {
				t.Fatalf("TDH PID attribution = %s, want server %s (client %s)", pid, serverPID, clientPID)
			}
			found = true
			t.Logf("TDH path ok: method=%s path=%s status=%d count=%d",
				pt.Fields().Get("method").GetS(), pt.Fields().Get("path").GetS(),
				pt.Fields().Get("status_code").GetI(), pt.Fields().Get("count").GetI())
			break
		}
	}
	if !found {
		for _, pt := range pts {
			t.Logf("point: method=%q path=%q status=%d count=%d",
				pt.Fields().Get("method").GetS(), pt.Fields().Get("path").GetS(),
				pt.Fields().Get("status_code").GetI(), pt.Fields().Get("count").GetI())
		}
		t.Fatalf("TDH path produced points but none completed (events=%d missed_conn=%d missed_req=%d completed=%d parse_errors=%d)",
			len(raws), st.missedConn, st.missedReq, st.completed, st.parseErrors)
	}
}
