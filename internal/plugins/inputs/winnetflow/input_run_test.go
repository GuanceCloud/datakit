// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows

package winnetflow

import (
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

// mockFeeder implements dkio.Feeder for testing the full Input.Run path.
type mockFeeder struct {
	mu     sync.Mutex
	points []*point.Point
}

func (m *mockFeeder) Feed(cat point.Category, pts []*point.Point, opts ...dkio.FeedOption) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.points = append(m.points, pts...)
	return nil
}

func (m *mockFeeder) FeedLastError(err string, opts ...metrics.LastErrorOption) {}

func (m *mockFeeder) getPoints() []*point.Point {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]*point.Point(nil), m.points...)
}

func waitInputStarted(t *testing.T, ipt *Input) {
	t.Helper()
	select {
	case <-ipt.testStarted:
	case <-time.After(10 * time.Second):
		t.Fatal("Input.Run did not start within 10s")
	}
	if ipt.collector == nil {
		t.Fatal("L4 ETW collector did not start; check privileges and conflicting ETW sessions")
	}
	if ipt.EnableHTTPFlow && ipt.httpCollector == nil {
		t.Fatal("HTTP ETW collector did not start; check privileges and conflicting ETW sessions")
	}
}

func listenNonLoopbackIPv4(t *testing.T) net.Listener {
	t.Helper()
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Fatalf("list interfaces: %v", err)
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil || ip.To4() == nil || ip.IsLoopback() || ip.IsUnspecified() {
				continue
			}
			ln, err := net.Listen("tcp4", net.JoinHostPort(ip.String(), "0"))
			if err == nil {
				return ln
			}
		}
	}
	t.Fatal("no usable non-loopback IPv4 interface")
	return nil
}

func exchangeOnce(t *testing.T, ln net.Listener) {
	t.Helper()
	accepted := make(chan struct{})
	go func() {
		defer close(accepted)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_, _ = conn.Write([]byte("winnetflow-e2e"))
	}()
	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial %s: %v", ln.Addr(), err)
	}
	_, _ = conn.Read(make([]byte, 64))
	_ = conn.Close()
	<-accepted
}

// TestInputRunEndToEnd exercises the production path: Input.Run starts both
// real ETW sessions, real loopback TCP + HttpListener traffic is generated,
// and the ticker flushes netflow/httpflow points through the feeder with the
// merged (host + user) tags applied. Requires real ETW sessions; skipped
// unless DK_WINNETFLOW_SMOKE=1.
func TestInputRunEndToEnd(t *testing.T) {
	if os.Getenv("DK_WINNETFLOW_SMOKE") != "1" {
		t.Skip("set DK_WINNETFLOW_SMOKE=1 to run the ETW smoke test")
	}

	feeder := &mockFeeder{}
	ipt := NewInput()
	ipt.includeLoopback = true
	ipt.Interval = "60s"
	ipt.testInterval = time.Second
	ipt.testStarted = make(chan struct{})
	ipt.feeder = feeder
	ipt.tagger = testutils.NewTaggerHost()
	ipt.Tags = map[string]string{"env": "prod"}

	done := make(chan struct{})
	go func() {
		defer close(done)
		ipt.Run()
	}()
	waitInputStarted(t, ipt)
	defer func() {
		ipt.Terminate()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatal("Input.Run did not stop after Terminate")
		}
	}()

	// TCP loopback traffic, mirroring TestETWSmokeCollectTCP.
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
			go func(c net.Conn) {
				defer c.Close()
				_, _ = c.Write([]byte("hello from winnetflow test"))
			}(conn)
		}
	}()
	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	buf := make([]byte, 128)
	_, _ = conn.Read(buf)
	conn.Close()
	tcpPort := strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)

	// HTTP.sys traffic via the HttpListener script.
	script, err := filepath.Abs(filepath.Join("verify", "gen-httpsvc-traffic.ps1"))
	if err != nil {
		t.Fatalf("resolve traffic script: %v", err)
	}
	port := 18534
	out, err := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass",
		"-File", script, "-Port", strconv.Itoa(port)).CombinedOutput()
	if err != nil {
		t.Fatalf("HTTP traffic generation failed: %v\n%s", err, out)
	}

	deadline := time.Now().Add(30 * time.Second)
	var gotNetflow, gotHTTPflow bool
	for time.Now().Before(deadline) {
		for _, pt := range feeder.getPoints() {
			if pt.Name() == metricName &&
				(pt.Tags().Get("src_port").GetS() == tcpPort || pt.Tags().Get("dst_port").GetS() == tcpPort) {
				if process := pt.Tags().Get("process_name").GetS(); process == "" || process == "unknown" {
					continue
				}
				gotNetflow = true
				if pt.Tags().Get("host").GetS() != "HOST" {
					t.Fatalf("netflow host tag = %q, want HOST", pt.Tags().Get("host").GetS())
				}
				if pt.Tags().Get("env").GetS() != "prod" {
					t.Fatalf("netflow env tag = %q, want prod", pt.Tags().Get("env").GetS())
				}
			}
			if pt.Name() == httpflowMetricName && pt.Tags().Get("src_port").GetS() == strconv.Itoa(port) {
				if process := pt.Tags().Get("process_name").GetS(); process == "" || process == "unknown" {
					continue
				}
				gotHTTPflow = true
				if m := pt.Fields().Get("method").GetS(); m == "" {
					t.Fatal("httpflow method is empty")
				}
				if pt.Tags().Get("host").GetS() != "HOST" {
					t.Fatalf("httpflow host tag = %q, want HOST", pt.Tags().Get("host").GetS())
				}
				if pt.Tags().Get("env").GetS() != "prod" {
					t.Fatalf("httpflow env tag = %q, want prod", pt.Tags().Get("env").GetS())
				}
			}
		}
		if gotNetflow && gotHTTPflow {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if !gotNetflow {
		t.Fatal("no target netflow point with process attribution reached the feeder within 30s")
	}
	if !gotHTTPflow {
		t.Fatal("no target httpflow point with process attribution reached the feeder within 30s")
	}
	t.Logf("end-to-end ok: netflow+httpflow points fed with merged tags (traffic: %s)",
		strings.TrimSpace(string(out)))
}

// TestInputRunDefaultFiltering exercises the production filtering mode with
// real ETW events: loopback L4/HTTP.sys traffic must be filtered while a TCP
// connection through a non-loopback host interface must still reach feeder.
func TestInputRunDefaultFiltering(t *testing.T) {
	if os.Getenv("DK_WINNETFLOW_SMOKE") != "1" {
		t.Skip("set DK_WINNETFLOW_SMOKE=1 to run the ETW smoke test")
	}

	feeder := &mockFeeder{}
	ipt := NewInput()
	ipt.Interval = "60s"
	ipt.testInterval = time.Second
	ipt.testStarted = make(chan struct{})
	ipt.feeder = feeder
	ipt.tagger = testutils.NewTaggerHost()
	done := make(chan struct{})
	go func() {
		defer close(done)
		ipt.Run()
	}()
	defer func() {
		ipt.Terminate()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatal("Input.Run did not stop after Terminate")
		}
	}()
	waitInputStarted(t, ipt)

	loopback, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen loopback: %v", err)
	}
	defer loopback.Close()
	exchangeOnce(t, loopback)
	loopbackPort := strconv.Itoa(loopback.Addr().(*net.TCPAddr).Port)

	nonLoopback := listenNonLoopbackIPv4(t)
	defer nonLoopback.Close()
	exchangeOnce(t, nonLoopback)
	nonLoopbackPort := strconv.Itoa(nonLoopback.Addr().(*net.TCPAddr).Port)

	script, err := filepath.Abs(filepath.Join("verify", "gen-httpsvc-traffic.ps1"))
	if err != nil {
		t.Fatalf("resolve traffic script: %v", err)
	}
	const httpPort = 18542
	if out, err := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass",
		"-File", script, "-Port", strconv.Itoa(httpPort), "-Requests", "2",
		"-PostTrafficDelayMS", "500").CombinedOutput(); err != nil {
		t.Fatalf("HTTP traffic generation failed: %v\n%s", err, out)
	}

	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		l4 := ipt.collector.(interface{ stats() collectorStats }).stats()
		http := ipt.httpCollector.(interface{ stats() httpStats }).stats()
		foundNonLoopback := false
		for _, pt := range feeder.getPoints() {
			if pt.Name() == metricName &&
				(pt.Tags().Get("src_port").GetS() == nonLoopbackPort || pt.Tags().Get("dst_port").GetS() == nonLoopbackPort) {
				foundNonLoopback = true
			}
		}
		if foundNonLoopback && l4.loopbackFiltered > 0 && http.loopbackFiltered > 0 {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}

	foundNonLoopback := false
	for _, pt := range feeder.getPoints() {
		if pt.Name() == metricName &&
			(pt.Tags().Get("src_port").GetS() == nonLoopbackPort || pt.Tags().Get("dst_port").GetS() == nonLoopbackPort) {
			foundNonLoopback = true
		}
		if pt.Name() == metricName &&
			(pt.Tags().Get("src_port").GetS() == loopbackPort || pt.Tags().Get("dst_port").GetS() == loopbackPort) {
			t.Fatalf("loopback L4 point reached feeder: %v", pt.Tags())
		}
		if pt.Name() == httpflowMetricName && pt.Tags().Get("src_port").GetS() == strconv.Itoa(httpPort) {
			t.Fatalf("loopback HTTP point reached feeder: %v", pt.Tags())
		}
	}
	if !foundNonLoopback {
		t.Fatal("non-loopback flow did not reach feeder")
	}
	if st := ipt.collector.(interface{ stats() collectorStats }).stats(); st.loopbackFiltered == 0 {
		t.Fatal("L4 loopback filter counter did not increase")
	}
	if st := ipt.httpCollector.(interface{ stats() httpStats }).stats(); st.loopbackFiltered == 0 {
		t.Fatal("HTTP loopback filter counter did not increase")
	}
}

// TestInputRunFinalFlush uses an interval that cannot tick during the test and
// proves Terminate drains copied ETW events and publishes the partial interval.
func TestInputRunFinalFlush(t *testing.T) {
	if os.Getenv("DK_WINNETFLOW_SMOKE") != "1" {
		t.Skip("set DK_WINNETFLOW_SMOKE=1 to run the ETW smoke test")
	}

	feeder := &mockFeeder{}
	ipt := NewInput()
	ipt.includeLoopback = true
	ipt.Interval = "5m"
	ipt.testStarted = make(chan struct{})
	ipt.feeder = feeder
	ipt.tagger = testutils.NewTaggerHost()
	done := make(chan struct{})
	go func() {
		defer close(done)
		ipt.Run()
	}()
	stopped := false
	defer func() {
		if !stopped {
			ipt.Terminate()
			select {
			case <-done:
			case <-time.After(10 * time.Second):
				t.Fatal("Input.Run did not stop after Terminate")
			}
		}
	}()

	waitInputStarted(t, ipt)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	acceptDone := make(chan struct{})
	go func() {
		defer close(acceptDone)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 1024)
		n, _ := conn.Read(buf)
		if n > 0 {
			_, _ = conn.Write(buf[:n])
		}
	}()
	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	payload := make([]byte, 1024)
	_, _ = conn.Write(payload)
	_, _ = readFull(conn, make([]byte, len(payload)))
	conn.Close()
	<-acceptDone

	script, err := filepath.Abs(filepath.Join("verify", "gen-httpsvc-traffic.ps1"))
	if err != nil {
		t.Fatalf("resolve traffic script: %v", err)
	}
	const httpPort = 18543
	if out, err := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass",
		"-File", script, "-Port", strconv.Itoa(httpPort), "-Requests", "2",
		"-PostTrafficDelayMS", "500").CombinedOutput(); err != nil {
		t.Fatalf("HTTP traffic generation failed: %v\n%s", err, out)
	}

	time.Sleep(500 * time.Millisecond)
	ipt.Terminate()
	stopped = true
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Input.Run did not stop after Terminate")
	}

	port := strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
	foundL4, foundHTTP := false, false
	for _, pt := range feeder.getPoints() {
		if pt.Name() == metricName &&
			(pt.Tags().Get("src_port").GetS() == port || pt.Tags().Get("dst_port").GetS() == port) &&
			(pt.Fields().Get("bytes_read").GetI() > 0 || pt.Fields().Get("bytes_written").GetI() > 0) {
			foundL4 = true
		}
		if pt.Name() == httpflowMetricName && pt.Tags().Get("src_port").GetS() == strconv.Itoa(httpPort) {
			foundHTTP = true
		}
	}
	if !foundL4 || !foundHTTP {
		t.Fatalf("final partial interval missing: netflow=%t httpflow=%t", foundL4, foundHTTP)
	}
}

// TestInputRunSoak runs the full input under sustained TCP + periodic HTTP.sys
// traffic (default 60s, override with DK_WINNETFLOW_SOAK_SECONDS) and asserts
// production invariants: no dropped/parse-error events, no ETW buffer loss,
// clean shutdown and no goroutine leak.
func TestInputRunSoak(t *testing.T) {
	if os.Getenv("DK_WINNETFLOW_SMOKE") != "1" {
		t.Skip("set DK_WINNETFLOW_SMOKE=1 to run the ETW smoke test")
	}

	duration := 60 * time.Second
	if s := os.Getenv("DK_WINNETFLOW_SOAK_SECONDS"); s != "" {
		if d, err := strconv.Atoi(s); err == nil && d > 0 {
			duration = time.Duration(d) * time.Second
		}
	}

	feeder := &mockFeeder{}
	ipt := NewInput()
	ipt.includeLoopback = true
	ipt.Interval = "60s"
	ipt.testInterval = 5 * time.Second
	ipt.testStarted = make(chan struct{})
	ipt.feeder = feeder
	ipt.tagger = testutils.NewTaggerHost()

	done := make(chan struct{})
	go func() {
		defer close(done)
		ipt.Run()
	}()
	waitInputStarted(t, ipt)
	stopped := false
	defer func() {
		if !stopped {
			ipt.Terminate()
			select {
			case <-done:
			case <-time.After(10 * time.Second):
				t.Fatal("Input.Run did not stop after Terminate")
			}
		}
	}()

	// Sustained TCP traffic: open/close connections with payload until stop.
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
			go func(conn net.Conn) {
				defer conn.Close()
				_, _ = conn.Write([]byte("soak payload"))
			}(c)
		}
	}()

	trafficStop := make(chan struct{})
	trafficDone := make(chan struct{})
	go func() {
		defer close(trafficDone)
		for {
			select {
			case <-trafficStop:
				return
			default:
			}
			c, err := net.Dial("tcp", ln.Addr().String())
			if err == nil {
				buf := make([]byte, 64)
				_, _ = c.Read(buf)
				c.Close()
			}
			time.Sleep(20 * time.Millisecond)
		}
	}()

	// Periodic bursts of HTTP.sys traffic, one every 30s for the whole soak.
	script, err := filepath.Abs(filepath.Join("verify", "gen-httpsvc-traffic.ps1"))
	if err != nil {
		t.Fatalf("resolve traffic script: %v", err)
	}
	httpStop := make(chan struct{})
	httpDone := make(chan struct{})
	go func() {
		defer close(httpDone)
		port := 18536
		for {
			select {
			case <-httpStop:
				return
			default:
			}
			out, err := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass",
				"-File", script, "-Port", strconv.Itoa(port)).CombinedOutput()
			if err != nil {
				t.Errorf("HTTP traffic generation failed: %v\n%s", err, out)
				return
			}
			port++
			select {
			case <-httpStop:
				return
			case <-time.After(30 * time.Second):
			}
		}
	}()

	time.Sleep(duration)

	close(trafficStop)
	<-trafficDone
	close(httpStop)
	<-httpDone

	// Assert ETW health before shutdown.
	if c, ok := ipt.collector.(interface{ stats() collectorStats }); ok {
		st := c.stats()
		if st.dropped != 0 || st.parseErrors != 0 {
			t.Fatalf("L4 unhealthy: dropped=%d parse_errors=%d", st.dropped, st.parseErrors)
		}
		if st.session.eventsLost != 0 || st.session.realTimeBuffersLost != 0 {
			t.Fatalf("L4 session loss: events_lost=%d buffers_lost=%d",
				st.session.eventsLost, st.session.realTimeBuffersLost)
		}
	}
	if c, ok := ipt.httpCollector.(interface{ stats() httpStats }); ok {
		st := c.stats()
		if st.dropped != 0 || st.parseErrors != 0 {
			t.Fatalf("L7 unhealthy: dropped=%d parse_errors=%d", st.dropped, st.parseErrors)
		}
		if st.session.eventsLost != 0 || st.session.realTimeBuffersLost != 0 {
			t.Fatalf("L7 session loss: events_lost=%d buffers_lost=%d",
				st.session.eventsLost, st.session.realTimeBuffersLost)
		}
		if st.missedConn+st.missedReq > 20 {
			t.Fatalf("L7 correlation degraded: missed_conn=%d missed_req=%d", st.missedConn, st.missedReq)
		}
	}
	var netflowPoints, httpflowPoints int
	for _, pt := range feeder.getPoints() {
		switch pt.Name() {
		case metricName:
			netflowPoints++
		case httpflowMetricName:
			httpflowPoints++
		}
	}
	if netflowPoints == 0 || httpflowPoints == 0 {
		t.Fatalf("soak produced no periodic output: netflow=%d httpflow=%d", netflowPoints, httpflowPoints)
	}

	gBefore := runtime.NumGoroutine()
	ipt.Terminate()
	stopped = true
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Input.Run did not stop after Terminate")
	}
	time.Sleep(2 * time.Second) // let worker goroutines unwind
	gAfter := runtime.NumGoroutine()
	if gAfter > gBefore+5 {
		t.Fatalf("possible goroutine leak: before=%d after=%d", gBefore, gAfter)
	}
	t.Logf("soak ok (%s): goroutines before=%d after=%d", duration, gBefore, gAfter)
}
