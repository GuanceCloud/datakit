//go:build linux
// +build linux

package l4log

import "testing"

func TestFlowAggTCPLimitDropsNewKeysAndResetsStats(t *testing.T) {
	agg := &FlowAggTCP{limit: 1}
	nicIPList := []string{"10.20.30.40"}

	info := &PMeta{SrcIP: "10.20.30.40", DstIP: "10.20.30.50", SrcPort: 12345, DstPort: 80}
	stats := &TCPMetrics{BytesRead: 10, BytesWritten: 20, Retransmits: 1, recEstab: true}
	stats.recClose[1] = true
	agg.Append(info, stats, "42", directionOutgoing, false, true, nicIPList)

	droppedInfo := &PMeta{SrcIP: "10.20.30.40", DstIP: "10.20.30.51", SrcPort: 12346, DstPort: 80}
	droppedStats := &TCPMetrics{BytesRead: 30, BytesWritten: 40, Retransmits: 2, recEstab: true}
	droppedStats.recClose[1] = true
	agg.Append(droppedInfo, droppedStats, "42", directionOutgoing, false, true, nicIPList)

	if got := agg.Len(); got != 1 {
		t.Fatalf("tcp agg entries = %d, want 1", got)
	}
	if droppedStats.BytesRead != 0 || droppedStats.BytesWritten != 0 || droppedStats.Retransmits != 0 ||
		droppedStats.recEstab || droppedStats.recClose[1] {
		t.Fatalf("expected dropped stats to be reset, got %+v", droppedStats)
	}

	nextStats := &TCPMetrics{BytesRead: 50, BytesWritten: 60}
	agg.Append(info, nextStats, "42", directionOutgoing, false, true, nicIPList)
	for _, value := range agg.data {
		if value.count != 2 {
			t.Fatalf("existing key should still aggregate at limit, got %+v", value)
		}
	}
}

func TestFlowAggHTTPLimitDropsNewKeysAndKeepsExisting(t *testing.T) {
	agg := &FlowAggHTTP{limit: 1}
	info := &PMeta{SrcIP: "10.20.30.40", DstIP: "10.20.30.50", SrcPort: 12345, DstPort: 80}
	nicIPList := []string{"10.20.30.40"}

	stats := &HTTPLogElem{Method: "GET", Path: "/a", StatusCode: 200, Direction: DIncoming}
	agg.Append(info, stats, "42", false, true, nicIPList, 10)
	agg.Append(info, &HTTPLogElem{Method: "GET", Path: "/b", StatusCode: 200, Direction: DIncoming},
		"42", false, true, nicIPList, 20)
	agg.Append(info, stats, "42", false, true, nicIPList, 30)

	if got := agg.Len(); got != 1 {
		t.Fatalf("http agg entries = %d, want 1", got)
	}
	for _, value := range agg.data {
		if value.count != 2 || value.latency != 40 {
			t.Fatalf("existing key should still aggregate at limit, got %+v", value)
		}
	}
}

func TestFlowAggHTTP2LimitDropsNewKeysAndKeepsExisting(t *testing.T) {
	agg := &FlowAggHTTP{limit: 1}
	info := &PMeta{SrcIP: "10.20.30.40", DstIP: "10.20.30.50", SrcPort: 12345, DstPort: 443}
	nicIPList := []string{"10.20.30.40"}

	stats := &HTTP2LogElem{Method: "GET", Path: "/a", StatusCode: 200, Direction: DIncoming}
	agg.AppendH2(info, stats, "42", false, true, nicIPList, 10)
	agg.AppendH2(info, &HTTP2LogElem{Method: "GET", Path: "/b", StatusCode: 200, Direction: DIncoming},
		"42", false, true, nicIPList, 20)
	agg.AppendH2(info, stats, "42", false, true, nicIPList, 30)

	if got := agg.Len(); got != 1 {
		t.Fatalf("http2 agg entries = %d, want 1", got)
	}
	for _, value := range agg.data {
		if value.count != 2 || value.latency != 40 {
			t.Fatalf("existing http2 key should still aggregate at limit, got %+v", value)
		}
	}
}

func TestL4logAggLimitEnv(t *testing.T) {
	t.Setenv(l4logNetflowAggLimitEnv, "bad")
	if got := l4logAggLimitFromEnv(l4logNetflowAggLimitEnv, defaultL4logNetflowAggLimit); got != defaultL4logNetflowAggLimit {
		t.Fatalf("invalid netflow agg limit = %d, want %d", got, defaultL4logNetflowAggLimit)
	}

	t.Setenv(l4logHTTPAggLimitEnv, "9999999")
	if got := l4logAggLimitFromEnv(l4logHTTPAggLimitEnv, defaultL4logHTTPAggLimit); got != maxL4logAggLimit {
		t.Fatalf("clamped http agg limit = %d, want %d", got, maxL4logAggLimit)
	}
}
