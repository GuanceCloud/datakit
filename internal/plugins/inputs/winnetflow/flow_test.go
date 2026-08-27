// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package winnetflow

import (
	"testing"
	"time"
)

func TestFlowAggregatorTCP(t *testing.T) {
	agg := newFlowAggregator(time.Second, defaultMaxFlows)

	ts := time.Now()
	base := flowEvent{
		ts:        ts,
		transport: "tcp",
		family:    "IPv4",
		direction: directionOutgoing,
		pid:       42,
		srcIP:     "127.0.0.1",
		srcPort:   50000,
		dstIP:     "10.0.0.1",
		dstPort:   443,
	}

	ev := base
	ev.kind = evTCPEstablished
	agg.onEvent(&ev)

	ev = base
	ev.kind = evTCPStateChange
	ev.oldState = tcpStateSynRcvd
	ev.newState = tcpStateEstablished
	agg.onEvent(&ev)

	ev = base
	ev.kind = evTCPDataSend
	ev.bytes = 1000
	ev.rtt = 500
	ev.rttVar = 50
	agg.onEvent(&ev)

	ev = base
	ev.kind = evTCPDataRecv
	ev.bytes = 2000
	ev.packets = 4
	agg.onEvent(&ev)

	ev = base
	ev.kind = evTCPRetransmit
	ev.retrans = 2
	agg.onEvent(&ev)

	ev = base
	ev.kind = evTCPStateChange
	ev.oldState = tcpStateEstablished
	ev.newState = tcpStateCloseWait
	agg.onEvent(&ev)

	ev = base
	ev.kind = evTCPStateChange
	ev.oldState = tcpStateCloseWait
	ev.newState = tcpStateLastAck
	ev.ts = ev.ts.Add(10 * time.Millisecond)
	agg.onEvent(&ev)

	ev = base
	ev.kind = evTCPClosed
	ev.ts = ts.Add(11 * time.Millisecond)
	agg.onEvent(&ev)

	pts := agg.flush()
	if len(pts) != 1 {
		t.Fatalf("expected 1 point, got %d", len(pts))
	}
	pt := pts[0]

	checkField := func(name string, want int64) {
		t.Helper()
		f := pt.Fields().Get(name)
		if f == nil {
			t.Fatalf("field %s missing", name)
		}
		if got := f.GetI(); got != want {
			t.Fatalf("field %s = %d, want %d", name, got, want)
		}
	}

	checkField("bytes_written", 1000)
	checkField("bytes_read", 2000)
	checkField("packets_read", 4)
	checkField("retransmits", 2)
	checkField("rtt", 500)
	checkField("rtt_var", 50)
	checkField("tcp_established", 1)
	checkField("tcp_connect_attempts", 1)
	checkField("tcp_closed", 1)
	checkField("tcp_close_wait", 1)
	checkField("tcp_last_ack", 1)

	if pt.Tags().Get("direction").GetS() != directionOutgoing {
		t.Fatalf("unexpected direction tag: %v", pt.Tags().Get("direction"))
	}
	if pt.Tags().Get("process_name").GetS() == "" {
		t.Fatal("process_name tag missing")
	}
}

func TestFlowAggregatorDeduplicatesTCPProviderSignals(t *testing.T) {
	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	base := flowEvent{
		ts: time.Now(), transport: "tcp", family: "IPv4",
		direction: directionOutgoing, pid: 42,
		srcIP: "127.0.0.1", srcPort: 50000,
		dstIP: "10.0.0.1", dstPort: 443,
	}

	state := base
	state.kind = evTCPStateChange
	state.oldState = tcpStateSynSent
	state.newState = tcpStateEstablished
	agg.onEvent(&state)

	connected := base
	connected.kind = evTCPEstablished
	agg.onEvent(&connected)

	closed := base
	closed.kind = evTCPClosed
	agg.onEvent(&closed)

	state = base
	state.kind = evTCPStateChange
	state.oldState = tcpStateClosed
	state.newState = tcpStateDeleteTcb
	agg.onEvent(&state)

	pts := agg.flush()
	if len(pts) != 1 {
		t.Fatalf("points = %d, want 1", len(pts))
	}
	for name, want := range map[string]int64{
		"tcp_connect_attempts": 1,
		"tcp_established":      1,
		"tcp_closed":           1,
	} {
		if got := pts[0].Fields().Get(name).GetI(); got != want {
			t.Fatalf("%s = %d, want %d", name, got, want)
		}
	}
}

func TestFlowAggregatorDeduplicatesTCPProviderSignalsAcrossFlush(t *testing.T) {
	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	base := flowEvent{
		ts: time.Now(), transport: "tcp", family: "IPv4",
		direction: directionOutgoing, pid: 42,
		srcIP: "127.0.0.1", srcPort: 50000,
		dstIP: "10.0.0.1", dstPort: 443,
	}
	connected := base
	connected.kind = evTCPEstablished
	agg.onEvent(&connected)
	if got := agg.flush()[0].Fields().Get("tcp_established").GetI(); got != 1 {
		t.Fatalf("first interval tcp_established = %d, want 1", got)
	}

	state := base
	state.kind = evTCPStateChange
	state.oldState = tcpStateSynRcvd
	state.newState = tcpStateEstablished
	agg.onEvent(&state)
	if got := agg.flush()[0].Fields().Get("tcp_established").GetI(); got != 0 {
		t.Fatalf("duplicate next-interval tcp_established = %d, want 0", got)
	}

	closed := base
	closed.kind = evTCPClosed
	agg.onEvent(&closed)
	if got := agg.flush()[0].Fields().Get("tcp_closed").GetI(); got != 1 {
		t.Fatalf("first close tcp_closed = %d, want 1", got)
	}
	state.oldState = tcpStateClosed
	state.newState = tcpStateDeleteTcb
	agg.onEvent(&state)
	if got := agg.flush()[0].Fields().Get("tcp_closed").GetI(); got != 0 {
		t.Fatalf("duplicate next-interval tcp_closed = %d, want 0", got)
	}
}

func TestFlowAggregatorConnectFailureCountsAttempt(t *testing.T) {
	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	agg.onEvent(&flowEvent{
		ts: time.Now(), kind: evTCPConnectFail, transport: "tcp", family: "IPv4",
		direction: directionOutgoing, pid: 42,
		srcIP: "127.0.0.1", srcPort: 50000,
		dstIP: "10.0.0.1", dstPort: 443,
	})
	pt := agg.flush()[0]
	if got := pt.Fields().Get("tcp_connect_attempts").GetI(); got != 1 {
		t.Fatalf("tcp_connect_attempts = %d, want 1", got)
	}
	if got := pt.Fields().Get("tcp_connect_failures").GetI(); got != 1 {
		t.Fatalf("tcp_connect_failures = %d, want 1", got)
	}
}

func TestFlowAggregatorUDP(t *testing.T) {
	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	agg.includeLoopback = true

	send := flowEvent{
		ts:        time.Now(),
		kind:      evUDPDataSend,
		transport: "udp",
		family:    "IPv6",
		direction: directionOutgoing,
		pid:       7,
		srcIP:     "::1",
		srcPort:   50000,
		dstIP:     "::1",
		dstPort:   53,
		bytes:     100,
		packets:   2,
	}
	recv := send
	recv.kind = evUDPDataRecv
	recv.bytes = 300
	recv.packets = 3

	agg.onEvent(&send)
	agg.onEvent(&recv)

	pts := agg.flush()
	if len(pts) != 1 {
		t.Fatalf("expected 1 point, got %d", len(pts))
	}
	pt := pts[0]

	if got := pt.Fields().Get("bytes_written").GetI(); got != int64(100) {
		t.Fatalf("bytes_written = %v, want 100", got)
	}
	if got := pt.Fields().Get("bytes_read").GetI(); got != int64(300) {
		t.Fatalf("bytes_read = %v, want 300", got)
	}
	if got := pt.Fields().Get("packets_written").GetI(); got != int64(2) {
		t.Fatalf("packets_written = %v, want 2", got)
	}
	if pt.Fields().Get("rtt") != nil {
		t.Fatal("UDP point should not have rtt field")
	}
}

func TestFlowAggregatorFlushResets(t *testing.T) {
	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	agg.includeLoopback = true
	ev := flowEvent{
		ts:        time.Now(),
		kind:      evTCPDataSend,
		transport: "tcp",
		family:    "IPv4",
		direction: directionOutgoing,
		pid:       1,
		srcIP:     "127.0.0.1",
		srcPort:   1,
		dstIP:     "127.0.0.1",
		dstPort:   2,
		bytes:     10,
	}
	agg.onEvent(&ev)
	if pts := agg.flush(); len(pts) != 1 {
		t.Fatalf("first flush: expected 1 point, got %d", len(pts))
	}
	if pts := agg.flush(); len(pts) != 0 {
		t.Fatalf("second flush: expected 0 points, got %d", len(pts))
	}
}

func TestFlowAggregatorStateSurvivesFlush(t *testing.T) {
	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	agg.includeLoopback = true
	base := flowEvent{
		ts:        time.Now(),
		kind:      evTCPStateChange,
		transport: "tcp",
		family:    "IPv4",
		direction: directionOutgoing,
		pid:       1,
		srcIP:     "127.0.0.1",
		srcPort:   50000,
		dstIP:     "127.0.0.1",
		dstPort:   443,
		oldState:  tcpStateEstablished,
		newState:  tcpStateCloseWait,
	}
	agg.onEvent(&base)
	if pts := agg.flush(); len(pts) != 1 {
		t.Fatalf("first flush: expected 1 point, got %d", len(pts))
	}

	exit := base
	exit.ts = exit.ts.Add(time.Second)
	exit.oldState = tcpStateCloseWait
	exit.newState = tcpStateLastAck
	agg.onEvent(&exit)
	pts := agg.flush()
	if len(pts) != 1 {
		t.Fatalf("second flush: expected 1 point, got %d", len(pts))
	}
	if got := pts[0].Fields().Get("tcp_close_wait"); got == nil || got.GetI() != 1 {
		t.Fatalf("tcp_close_wait = %v, want 1", got)
	}
}

func TestFlowAggregatorStateTrackingBounded(t *testing.T) {
	agg := newFlowAggregator(time.Second, 2)
	for i := 0; i < 3; i++ {
		agg.setState(flowKey{pid: uint32(i + 1)}, time.Now(), tcpStateCloseWait)
	}
	if got := len(agg.stateSince); got != 2 {
		t.Fatalf("tracked state flows = %d, want 2", got)
	}
}

func TestFlowAggregatorExtraTags(t *testing.T) {
	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	agg.extraTags = map[string]string{"host": "test-host", "k8s_namespace": "ns1"}

	agg.onEvent(&flowEvent{
		ts:        time.Now(),
		kind:      evTCPEstablished,
		transport: "tcp",
		family:    "IPv4",
		direction: directionOutgoing,
		pid:       42,
		srcIP:     "127.0.0.1",
		srcPort:   50000,
		dstIP:     "10.0.0.1",
		dstPort:   443,
	})

	pts := agg.flush()
	if len(pts) != 1 {
		t.Fatalf("expected 1 point, got %d", len(pts))
	}
	if got := pts[0].Tags().Get("host").GetS(); got != "test-host" {
		t.Fatalf("host tag = %q, want test-host", got)
	}
	if got := pts[0].Tags().Get("k8s_namespace").GetS(); got != "ns1" {
		t.Fatalf("k8s_namespace tag = %q, want ns1", got)
	}
}

func TestFlowFlushResolvesProcessNameOncePerPID(t *testing.T) {
	oldResolver := resolveProcessName
	t.Cleanup(func() { resolveProcessName = oldResolver })
	calls := 0
	resolveProcessName = func(pid uint32) string {
		calls++
		return "process"
	}

	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	for i := 0; i < 2; i++ {
		agg.onEvent(&flowEvent{
			ts:        time.Now(),
			kind:      evTCPDataSend,
			transport: "tcp",
			family:    "IPv4",
			direction: directionOutgoing,
			pid:       42,
			srcIP:     "127.0.0.1",
			srcPort:   uint32(50000 + i),
			dstIP:     "10.0.0.1",
			dstPort:   443,
			bytes:     1,
		})
	}
	if pts := agg.flush(); len(pts) != 1 {
		t.Fatalf("points = %d, want 1 after dynamic-port aggregation", len(pts))
	}
	if calls != 1 {
		t.Fatalf("process-name lookups = %d, want 1", calls)
	}
}

func TestProcessNameWarmerDeduplicatesPendingPID(t *testing.T) {
	oldRefresher := refreshProcessName
	t.Cleanup(func() { refreshProcessName = oldRefresher })
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})
	refreshProcessName = func(pid uint32) string {
		close(started)
		<-release
		close(done)
		return "short-lived.exe"
	}

	w := newProcessNameWarmer()
	w.enqueue(42)
	<-started
	for i := 0; i < 100; i++ {
		w.enqueue(42)
	}
	close(release)
	<-done
	w.stop()

	w.mu.Lock()
	pending := len(w.pending)
	w.mu.Unlock()
	if pending != 0 {
		t.Fatalf("pending process lookups = %d, want 0", pending)
	}
}

func TestFlowWarmsProcessOncePerPIDPerFlush(t *testing.T) {
	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	calls := 0
	agg.warmProcessName = func(uint32) { calls++ }
	for i := 0; i < 10; i++ {
		agg.onEvent(&flowEvent{
			ts: time.Now(), kind: evTCPDataSend, transport: "tcp", family: "IPv4",
			direction: directionOutgoing, pid: 42,
			srcIP: "127.0.0.1", srcPort: uint32(50000 + i),
			dstIP: "10.0.0.1", dstPort: 443, bytes: 1,
		})
	}
	if calls != 1 {
		t.Fatalf("process warm calls = %d, want 1", calls)
	}
	agg.flush()
	agg.onEvent(&flowEvent{
		ts: time.Now(), kind: evTCPDataSend, transport: "tcp", family: "IPv4",
		direction: directionOutgoing, pid: 42,
		srcIP: "127.0.0.1", srcPort: 51000,
		dstIP: "10.0.0.1", dstPort: 443, bytes: 1,
	})
	if calls != 2 {
		t.Fatalf("process warm calls after flush = %d, want 2", calls)
	}
}

func TestFlowAggregatorSkipsEmptyTuple(t *testing.T) {
	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	ev := flowEvent{ts: time.Now(), kind: evTCPClosed, transport: "tcp"}
	agg.onEvent(&ev) // must not panic
	if pts := agg.flush(); pts != nil {
		t.Fatalf("expected no points, got %d", len(pts))
	}
}

func TestFlowAggregatorFlowCap(t *testing.T) {
	agg := newFlowAggregator(time.Second, 4)
	for i := 0; i < 8; i++ {
		ev := flowEvent{
			ts:        time.Now(),
			kind:      evTCPDataSend,
			transport: "tcp",
			family:    "IPv4",
			direction: directionOutgoing,
			pid:       uint32(i),
			srcIP:     "10.0.0.1",
			srcPort:   uint32(1000 + i),
			dstIP:     "10.0.0.2",
			dstPort:   80,
			bytes:     100,
		}
		agg.onEvent(&ev)
	}
	if got := agg.flowsSkippedCount(); got != 4 {
		t.Fatalf("flows skipped = %d, want 4", got)
	}
	if pts := agg.flush(); len(pts) != 4 {
		t.Fatalf("tracked flows = %d, want 4", len(pts))
	}
}

func TestFlowAggregatorBurstReplay(t *testing.T) {
	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	ev := flowEvent{
		ts:             time.Now(),
		kind:           evBurstReplay,
		transport:      "tcp",
		family:         "IPv4",
		direction:      directionIncoming,
		pid:            9,
		srcIP:          "10.0.0.2",
		srcPort:        80,
		dstIP:          "10.0.0.1",
		dstPort:        50000,
		cumBytesSend:   1000,
		cumBytesRecv:   2000,
		cumPacketsSend: 5,
		cumPacketsRecv: 7,
		cumRetrans:     1,
		cumRTT:         3000,
		cumRTTVar:      300,
		cumRTTCount:    6,
	}
	agg.onEvent(&ev)

	pts := agg.flush()
	if len(pts) != 1 {
		t.Fatalf("expected 1 point, got %d", len(pts))
	}
	pt := pts[0]
	checks := map[string]int64{
		"bytes_written":   1000,
		"bytes_read":      2000,
		"packets_written": 5,
		"packets_read":    7,
		"retransmits":     1,
		"rtt":             500, // 3000 / 6
		"rtt_var":         50,
	}
	for name, want := range checks {
		if got := pt.Fields().Get(name).GetI(); got != want {
			t.Fatalf("field %s = %d, want %d", name, got, want)
		}
	}
	if pt.Tags().Get("direction").GetS() != directionIncoming {
		t.Fatalf("direction tag = %v", pt.Tags().Get("direction"))
	}
	if got := pt.Tags().Get("conn_side").GetS(); got != "server" {
		t.Fatalf("conn_side = %q, want server", got)
	}
	if got := pt.Tags().Get("server_ip").GetS(); got != "10.0.0.2" {
		t.Fatalf("server_ip = %q, want local source", got)
	}
	if got := pt.Tags().Get("client_ip").GetS(); got != "10.0.0.1" {
		t.Fatalf("client_ip = %q, want remote destination", got)
	}
	if got := pt.Fields().Get("client_sent").GetI(); got != 2000 {
		t.Fatalf("client_sent = %d, want bytes_read 2000", got)
	}
	if got := pt.Fields().Get("server_sent").GetI(); got != 1000 {
		t.Fatalf("server_sent = %d, want bytes_written 1000", got)
	}
}

func TestFlowAggregatorPromotesUnknownPID(t *testing.T) {
	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	agg.includeLoopback = true
	base := flowEvent{
		ts:        time.Now(),
		transport: "tcp",
		family:    "IPv4",
		direction: directionIncoming,
		srcIP:     "127.0.0.1",
		srcPort:   8080,
		dstIP:     "127.0.0.1",
		dstPort:   50000,
	}
	data := base
	data.kind = evTCPDataSend
	data.bytes = 1024
	// Inline data may be inferred as outgoing before a newly created listener
	// appears in the periodic snapshot. The accept event is authoritative.
	data.direction = directionOutgoing
	agg.onEvent(&data)
	established := base
	established.kind = evTCPEstablished
	established.pid = 42
	agg.onEvent(&established)

	pts := agg.flush()
	if len(pts) != 1 {
		t.Fatalf("points = %d, want 1", len(pts))
	}
	pt := pts[0]
	if got := pt.Tags().Get("pid").GetS(); got != "42" {
		t.Fatalf("pid = %q, want 42", got)
	}
	if got := pt.Fields().Get("bytes_written").GetI(); got != 1024 {
		t.Fatalf("bytes_written = %d, want 1024", got)
	}
	if got := pt.Fields().Get("tcp_established").GetI(); got != 1 {
		t.Fatalf("tcp_established = %d, want 1", got)
	}
}

func TestFlowAggregatorFiltersInvalidAndLoopback(t *testing.T) {
	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	agg.onEvent(&flowEvent{
		ts: time.Now(), kind: evTCPDataSend, transport: "tcp", family: "IPv4",
		direction: directionOutgoing, srcIP: "0.0.0.0", srcPort: 50000,
		dstIP: "10.0.0.1", dstPort: 443,
	})
	agg.onEvent(&flowEvent{
		ts: time.Now(), kind: evTCPDataSend, transport: "tcp", family: "IPv4",
		direction: directionOutgoing, srcIP: "127.0.0.1", srcPort: 50000,
		dstIP: "127.0.0.1", dstPort: 443,
	})

	if pts := agg.flush(); len(pts) != 0 {
		t.Fatalf("filtered events produced %d points", len(pts))
	}
	invalid, loopback := agg.filterCounts()
	if invalid != 1 || loopback != 1 {
		t.Fatalf("filter counts = invalid:%d loopback:%d, want 1/1", invalid, loopback)
	}
}

func TestFlowAggregatorGroupsDynamicPortsWithoutMergingLifecycle(t *testing.T) {
	agg := newFlowAggregator(time.Second, defaultMaxFlows)
	for i := uint32(0); i < 2; i++ {
		base := flowEvent{
			ts: time.Now(), transport: "tcp", family: "IPv4",
			direction: directionOutgoing, pid: 42,
			srcIP: "10.0.0.1", srcPort: clientPortRollupStart + i,
			dstIP: "10.0.0.2", dstPort: 443,
		}
		connected := base
		connected.kind = evTCPEstablished
		agg.onEvent(&connected)
		sent := base
		sent.kind = evTCPDataSend
		sent.bytes = uint64(100 * (i + 1))
		sent.packets = uint64(i + 1)
		agg.onEvent(&sent)
		received := base
		received.kind = evTCPDataRecv
		received.bytes = uint64(200 * (i + 1))
		received.packets = uint64(2 * (i + 1))
		agg.onEvent(&received)
		rtt := base
		rtt.kind = evRTT
		rtt.rtt = uint64(100 + 200*i)
		rtt.rttVar = uint64(10 + 20*i)
		agg.onEvent(&rtt)
		retransmit := base
		retransmit.kind = evTCPRetransmit
		retransmit.retrans = uint64(i + 1)
		agg.onEvent(&retransmit)
		closed := base
		closed.kind = evTCPClosed
		agg.onEvent(&closed)
	}

	pts := agg.flush()
	if len(pts) != 1 {
		t.Fatalf("dynamic client ports produced %d points, want 1", len(pts))
	}
	pt := pts[0]
	if got := pt.Tags().Get("client_port").GetS(); got != "*" {
		t.Fatalf("client_port = %q, want *", got)
	}
	if got := pt.Fields().Get("tcp_established").GetI(); got != 2 {
		t.Fatalf("tcp_established = %d, want 2", got)
	}
	if got := pt.Fields().Get("tcp_closed").GetI(); got != 2 {
		t.Fatalf("tcp_closed = %d, want 2", got)
	}
	for field, want := range map[string]int64{
		"bytes_written":   300,
		"bytes_read":      600,
		"packets_written": 3,
		"packets_read":    6,
		"retransmits":     3,
		"rtt":             200,
		"rtt_var":         20,
	} {
		if got := pt.Fields().Get(field).GetI(); got != want {
			t.Fatalf("%s = %d, want %d", field, got, want)
		}
	}
}

func TestNormalizeClientPortPreservesServerPort(t *testing.T) {
	src, dst := normalizeClientPort(directionIncoming, 60000, 50000)
	if src != 60000 || dst != normalizedClientPort {
		t.Fatalf("incoming ports = %d/%d, want 60000/*", src, dst)
	}
	src, dst = normalizeClientPort(directionOutgoing, 50000, 60000)
	if src != normalizedClientPort || dst != 60000 {
		t.Fatalf("outgoing ports = %d/%d, want */60000", src, dst)
	}
}

func TestNormalizeClientPortBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name      string
		direction string
		src       uint32
		dst       uint32
		wantSrc   uint32
		wantDst   uint32
	}{
		{name: "outgoing below range", direction: directionOutgoing, src: 32767, dst: 443, wantSrc: 32767, wantDst: 443},
		{name: "outgoing range start", direction: directionOutgoing, src: 32768, dst: 443, wantSrc: normalizedClientPort, wantDst: 443},
		{name: "outgoing max", direction: directionOutgoing, src: 65535, dst: 443, wantSrc: normalizedClientPort, wantDst: 443},
		{name: "incoming range start", direction: directionIncoming, src: 443, dst: 32768, wantSrc: 443, wantDst: normalizedClientPort},
		{name: "unknown direction", direction: "", src: 32768, dst: 32769, wantSrc: 32768, wantDst: 32769},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src, dst := normalizeClientPort(tc.direction, tc.src, tc.dst)
			if src != tc.wantSrc || dst != tc.wantDst {
				t.Fatalf("ports = %d/%d, want %d/%d", src, dst, tc.wantSrc, tc.wantDst)
			}
		})
	}
}

func TestFilterEndpointsMatrix(t *testing.T) {
	for _, tc := range []struct {
		name            string
		srcIP           string
		srcPort         uint32
		dstIP           string
		dstPort         uint32
		includeLoopback bool
		want            endpointFilterReason
	}{
		{name: "IPv4 accepted", srcIP: "10.0.0.1", srcPort: 1234, dstIP: "8.8.8.8", dstPort: 53, want: endpointAccepted},
		{name: "IPv6 accepted", srcIP: "2001:db8::1", srcPort: 1234, dstIP: "2001:db8::2", dstPort: 443, want: endpointAccepted},
		{name: "zero source port", srcIP: "10.0.0.1", dstIP: "10.0.0.2", dstPort: 443, want: endpointInvalid},
		{name: "zero destination port", srcIP: "10.0.0.1", srcPort: 1234, dstIP: "10.0.0.2", want: endpointInvalid},
		{name: "IPv6 unspecified", srcIP: "::", srcPort: 1234, dstIP: "2001:db8::2", dstPort: 443, want: endpointInvalid},
		{name: "malformed source", srcIP: "not-an-ip", srcPort: 1234, dstIP: "10.0.0.2", dstPort: 443, want: endpointInvalid},
		{name: "pure IPv4 loopback", srcIP: "127.0.0.1", srcPort: 1234, dstIP: "127.0.0.1", dstPort: 443, want: endpointLoopback},
		{name: "pure IPv6 loopback", srcIP: "::1", srcPort: 1234, dstIP: "::1", dstPort: 443, want: endpointLoopback},
		{name: "loopback included", srcIP: "127.0.0.1", srcPort: 1234, dstIP: "127.0.0.1", dstPort: 443, includeLoopback: true, want: endpointAccepted},
		{name: "invalid despite loopback override", srcIP: "0.0.0.0", srcPort: 1234, dstIP: "127.0.0.1", dstPort: 443, includeLoopback: true, want: endpointInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := filterEndpoints(tc.srcIP, tc.srcPort, tc.dstIP, tc.dstPort, tc.includeLoopback); got != tc.want {
				t.Fatalf("filterEndpoints() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestEndpointIPType(t *testing.T) {
	for raw, want := range map[string]string{
		"127.0.0.1": "loopback",
		"10.0.0.1":  "private",
		"224.0.0.1": "multicast",
		"8.8.8.8":   "other",
		"bad":       "other",
	} {
		if got := endpointIPType(raw); got != want {
			t.Fatalf("endpointIPType(%q) = %q, want %q", raw, got, want)
		}
	}
}
