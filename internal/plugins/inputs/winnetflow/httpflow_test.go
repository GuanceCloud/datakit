// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package winnetflow

import (
	"testing"
	"time"
	"unicode/utf8"
)

func testHTTPEvent() *httpEvent {
	return &httpEvent{
		ts:           time.Now(),
		start:        time.Now().Add(-10 * time.Millisecond),
		family:       "IPv4",
		pid:          4242,
		srcIP:        "10.0.0.1",
		srcPort:      8080,
		dstIP:        "10.0.0.2",
		dstPort:      51234,
		method:       "GET",
		path:         "/api/users",
		status:       200,
		bytesWritten: 512,
	}
}

func TestHTTPAggregatorFlush(t *testing.T) {
	a := newHTTPAggregator(1024)
	a.onEvent(testHTTPEvent())

	pts := a.flush()
	if len(pts) != 1 {
		t.Fatalf("expected 1 point, got %d", len(pts))
	}
	pt := pts[0]
	if pt.Name() != "httpflow" {
		t.Fatalf("measurement = %s, want httpflow", pt.Name())
	}
	if pt.Tags().Get("src_ip").GetS() != "10.0.0.1" ||
		pt.Tags().Get("dst_ip").GetS() != "10.0.0.2" ||
		pt.Tags().Get("src_port").GetS() != "8080" ||
		pt.Tags().Get("dst_port").GetS() != "*" {
		t.Fatalf("unexpected tags: %v", pt.Tags())
	}
	if pt.Tags().Get("direction").GetS() != directionIncoming ||
		pt.Tags().Get("transport").GetS() != "tcp" {
		t.Fatalf("unexpected direction/transport tags: %v", pt.Tags())
	}
	if pt.Tags().Get("conn_side").GetS() != "server" ||
		pt.Tags().Get("server_port").GetS() != "8080" ||
		pt.Tags().Get("client_port").GetS() != "*" {
		t.Fatalf("unexpected client/server tags: %v", pt.Tags())
	}
	if pt.Fields().Get("method").GetS() != "GET" ||
		pt.Fields().Get("path").GetS() != "/api/users" ||
		pt.Fields().Get("status_code").GetI() != 200 {
		t.Fatalf("unexpected fields: %v", pt.Fields())
	}
	if pt.Fields().Get("count").GetI() != 1 {
		t.Fatalf("count = %d, want 1", pt.Fields().Get("count").GetI())
	}
	if pt.Fields().Get("bytes_written").GetI() != 512 {
		t.Fatalf("bytes_written = %d, want 512", pt.Fields().Get("bytes_written").GetI())
	}
	if pt.Fields().Get("server_sent").GetI() != 512 ||
		pt.Fields().Get("client_sent").GetI() != 0 {
		t.Fatalf("unexpected client/server byte fields: %v", pt.Fields())
	}
	if pt.Fields().Get("truncated").GetB() {
		t.Fatalf("truncated should be false")
	}
	if pt.Fields().Get("http_version").GetS() != "" {
		t.Fatalf("http_version should be empty, got %q", pt.Fields().Get("http_version").GetS())
	}
	// Latency should be ~10ms in ns.
	lat := pt.Fields().Get("latency").GetI()
	if lat < 9_000_000 || lat > 11_000_000 {
		t.Fatalf("latency = %d ns, want ~10ms", lat)
	}

	if pts2 := a.flush(); len(pts2) != 0 {
		t.Fatalf("aggregator should reset after flush, got %d points", len(pts2))
	}
}

func TestHTTPAggregatorGrouping(t *testing.T) {
	a := newHTTPAggregator(1024)
	base := testHTTPEvent()

	// Same key: aggregated count and average latency.
	a.onEvent(base)
	dup := testHTTPEvent()
	dup.start = dup.start.Add(-2 * time.Millisecond) // 12ms total
	a.onEvent(dup)

	// Different method -> separate group.
	post := testHTTPEvent()
	post.method = "POST"
	a.onEvent(post)

	// Different path -> separate group.
	other := testHTTPEvent()
	other.path = "/other"
	a.onEvent(other)

	pts := a.flush()
	if len(pts) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(pts))
	}
	for _, pt := range pts {
		if pt.Fields().Get("method").GetS() == "GET" && pt.Fields().Get("path").GetS() == "/api/users" {
			if pt.Fields().Get("count").GetI() != 2 {
				t.Fatalf("GET /api/users count = %d, want 2", pt.Fields().Get("count").GetI())
			}
			lat := pt.Fields().Get("latency").GetI()
			if lat < 10_000_000 || lat > 12_000_000 {
				t.Fatalf("average latency = %d, want ~11ms", lat)
			}
		}
	}
}

func TestHTTPAggregatorTruncationFlag(t *testing.T) {
	a := newHTTPAggregator(1024)
	ev := testHTTPEvent()
	ev.truncated = true
	a.onEvent(ev)
	pts := a.flush()
	if len(pts) != 1 || !pts[0].Fields().Get("truncated").GetB() {
		t.Fatalf("truncated flag not propagated: %v", pts)
	}
}

func TestHTTPAggregatorLimit(t *testing.T) {
	a := newHTTPAggregator(2)
	ev := testHTTPEvent()
	for i := 0; i < 3; i++ {
		e := *ev
		e.path = "/p" + string(rune('a'+i))
		a.onEvent(&e)
	}
	pts := a.flush()
	if len(pts) != 2 {
		t.Fatalf("expected 2 points under limit, got %d", len(pts))
	}
	if a.requestsSkippedCount() != 1 {
		t.Fatalf("requestsSkipped = %d, want 1", a.requestsSkippedCount())
	}
}

func TestHTTPAggregatorRejectsEmptyTuple(t *testing.T) {
	a := newHTTPAggregator(1024)
	ev := testHTTPEvent()
	ev.srcIP = ""
	a.onEvent(ev)
	if pts := a.flush(); len(pts) != 0 {
		t.Fatalf("empty tuple should be rejected, got %d points", len(pts))
	}
}

func TestHTTPAggregatorFiltersInvalidAndLoopback(t *testing.T) {
	a := newHTTPAggregator(1024)
	invalid := testHTTPEvent()
	invalid.dstPort = 0
	a.onEvent(invalid)
	loopback := testHTTPEvent()
	loopback.srcIP = "127.0.0.1"
	loopback.dstIP = "127.0.0.1"
	a.onEvent(loopback)

	if pts := a.flush(); len(pts) != 0 {
		t.Fatalf("filtered events produced %d points", len(pts))
	}
	invalidCount, loopbackCount := a.filterCounts()
	if invalidCount != 1 || loopbackCount != 1 {
		t.Fatalf("filter counts = invalid:%d loopback:%d, want 1/1", invalidCount, loopbackCount)
	}
}

func TestHTTPAggregatorGroupsClientDynamicPorts(t *testing.T) {
	a := newHTTPAggregator(1024)
	first := testHTTPEvent()
	second := testHTTPEvent()
	now := time.Now()
	first.ts = now
	first.start = now.Add(-10 * time.Millisecond)
	first.bytesWritten = 100
	second.ts = now
	second.start = now.Add(-20 * time.Millisecond)
	second.bytesWritten = 300
	second.truncated = true
	second.dstPort++
	a.onEvent(first)
	a.onEvent(second)

	pts := a.flush()
	if len(pts) != 1 {
		t.Fatalf("dynamic client ports produced %d points, want 1", len(pts))
	}
	if got := pts[0].Fields().Get("count").GetI(); got != 2 {
		t.Fatalf("count = %d, want 2", got)
	}
	if got := pts[0].Tags().Get("client_port").GetS(); got != "*" {
		t.Fatalf("client_port = %q, want *", got)
	}
	if got := pts[0].Fields().Get("latency").GetI(); got != int64(15*time.Millisecond) {
		t.Fatalf("latency = %d, want %d", got, 15*time.Millisecond)
	}
	if got := pts[0].Fields().Get("bytes_written").GetI(); got != 400 {
		t.Fatalf("bytes_written = %d, want 400", got)
	}
	if !pts[0].Fields().Get("truncated").GetB() {
		t.Fatal("truncated should be preserved across aggregation")
	}
}

func TestHTTPAggregatorDynamicPortsKeepOtherDimensions(t *testing.T) {
	a := newHTTPAggregator(1024)
	base := testHTTPEvent()
	a.onEvent(base)
	otherPID := testHTTPEvent()
	otherPID.pid++
	otherPID.dstPort++
	a.onEvent(otherPID)
	otherServer := testHTTPEvent()
	otherServer.srcPort++
	otherServer.dstPort += 2
	a.onEvent(otherServer)
	if pts := a.flush(); len(pts) != 3 {
		t.Fatalf("points = %d, want 3 distinct PID/server dimensions", len(pts))
	}
}

func TestHTTPAggregatorExtraTags(t *testing.T) {
	a := newHTTPAggregator(1024)
	a.extraTags = map[string]string{"host": "test-host", "env": "prod"}
	a.onEvent(testHTTPEvent())

	pts := a.flush()
	if len(pts) != 1 {
		t.Fatalf("expected 1 point, got %d", len(pts))
	}
	if got := pts[0].Tags().Get("host").GetS(); got != "test-host" {
		t.Fatalf("host tag = %q, want test-host", got)
	}
	if got := pts[0].Tags().Get("env").GetS(); got != "prod" {
		t.Fatalf("env tag = %q, want prod", got)
	}
}

func TestHTTPPathLimit(t *testing.T) {
	short := "/a"
	path, trunc := httpRequestPath(short, 256)
	if path != "/a" || trunc {
		t.Fatalf("short path: got %q trunc=%v", path, trunc)
	}

	// 300 'x' characters after the slash.
	buf := make([]byte, 0, 301)
	buf = append(buf, '/')
	for i := 0; i < 300; i++ {
		buf = append(buf, 'x')
	}
	longURI := string(buf)
	path, trunc = httpRequestPath(longURI, 256)
	if len(path) != 256 || !trunc {
		t.Fatalf("long path: len=%d trunc=%v", len(path), trunc)
	}

	// Query strings are stripped to avoid leaking secrets and high-cardinality
	// values into metric fields.
	path, trunc = httpRequestPath("http://h/p?q=1", 256)
	if path != "/p" || trunc {
		t.Fatalf("query path: got %q trunc=%v", path, trunc)
	}

	path, trunc = httpRequestPath("/你好/world", 5)
	if path != "/你" || !trunc || !utf8.ValidString(path) {
		t.Fatalf("UTF-8 truncation: got %q trunc=%v", path, trunc)
	}

	// Fallback for unparseable input.
	path, trunc = httpRequestPath("", 256)
	if path != "/" || trunc {
		t.Fatalf("empty uri: got %q trunc=%v", path, trunc)
	}
}

func TestHTTPVerbName(t *testing.T) {
	cases := map[uint32]string{
		0: "", 1: "", 2: "",
		3: "OPTIONS", 4: "GET", 5: "HEAD", 6: "POST", 7: "PUT",
		8: "DELETE", 9: "TRACE", 10: "CONNECT", 20: "QUERY",
	}
	for v, want := range cases {
		if got := httpVerbName(v); got != want {
			t.Fatalf("httpVerbName(%d) = %q, want %q", v, got, want)
		}
	}
}
