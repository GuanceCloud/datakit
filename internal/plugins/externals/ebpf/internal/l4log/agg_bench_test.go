//go:build linux
// +build linux

package l4log

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
)

func TestFlowAggTCPAppendAggregates(t *testing.T) {
	agg := &FlowAggTCP{}
	info := &PMeta{
		SrcIP:   "10.20.30.40",
		DstIP:   "10.20.30.50",
		SrcPort: 43210,
		DstPort: 8080,
	}
	stats := &TCPMetrics{
		BytesRead:    64,
		BytesWritten: 128,
		RTT:          100,
		RTTVar:       20,
		Retransmits:  2,
		recEstab:     true,
	}
	stats.recClose[1] = true

	agg.Append(info, stats, "42", directionOutgoing, false, true, []string{"10.20.30.40"})

	if agg.Len() != 1 {
		t.Fatalf("expected 1 flow, got %d", agg.Len())
	}
	if stats.BytesRead != 0 || stats.BytesWritten != 0 || stats.Retransmits != 0 {
		t.Fatalf("expected stats to be reset, got %+v", stats)
	}
	if stats.recEstab || stats.recClose[1] {
		t.Fatalf("expected state flags to be cleared, got %+v", stats)
	}

	for _, value := range agg.data {
		if value.bytesRead != 64 || value.bytesWritten != 128 {
			t.Fatalf("unexpected aggregated bytes: %+v", value)
		}
		if value.tcpEstablished != 1 || value.tcpClosed != 1 {
			t.Fatalf("unexpected tcp counters: %+v", value)
		}
	}
}

func TestFlowAggHTTPAppendAggregates(t *testing.T) {
	agg := &FlowAggHTTP{}
	info := &PMeta{
		SrcIP:   "10.20.30.40",
		DstIP:   "10.20.30.50",
		SrcPort: 43210,
		DstPort: 8080,
	}
	stats := &HTTPLogElem{
		Method:     "GET",
		Path:       "/health",
		StatusCode: 200,
		Direction:  DIncoming,
		txBytes:    128,
		rxBytes:    64,
	}

	agg.Append(info, stats, "42", false, true, []string{"10.20.30.40"}, 15)
	agg.Append(info, stats, "42", false, true, []string{"10.20.30.40"}, 25)

	if agg.Len() != 1 {
		t.Fatalf("expected 1 flow, got %d", agg.Len())
	}

	for _, value := range agg.data {
		if value.count != 2 {
			t.Fatalf("expected 2 requests, got %+v", value)
		}
		if value.latency != 40 || value.recvBytes != 128 || value.sendBytes != 256 {
			t.Fatalf("unexpected aggregated http value: %+v", value)
		}
	}
}

func TestFlowAggTCPToPointAddsDstDomain(t *testing.T) {
	agg := &FlowAggTCP{}
	info := &PMeta{
		SrcIP:   "10.20.30.40",
		DstIP:   "10.20.30.50",
		SrcPort: 43210,
		DstPort: 443,
	}
	stats := &TCPMetrics{
		BytesRead:    64,
		BytesWritten: 128,
	}
	netflow.RecordPeerDomain("10.20.30.50", 443, "tcp", "42", "api.example.com")

	agg.Append(info, stats, "42", directionOutgoing, false, true, []string{"10.20.30.40"})
	pts := agg.ToPoint(nil, nil)
	if len(pts) != 1 {
		t.Fatalf("expected 1 point, got %d", len(pts))
	}
	if got := pts[0].GetTag("dst_domain"); got != "api.example.com" {
		t.Fatalf("unexpected dst_domain %q", got)
	}
	if got := pts[0].GetTag("server_domain"); got != "api.example.com" {
		t.Fatalf("unexpected server_domain %q", got)
	}
}

func TestFlowAggHTTPToPointAddsDstDomain(t *testing.T) {
	agg := &FlowAggHTTP{}
	info := &PMeta{
		SrcIP:   "10.20.30.40",
		DstIP:   "10.20.30.60",
		SrcPort: 43210,
		DstPort: 443,
	}
	stats := &HTTPLogElem{
		Method:     "GET",
		Path:       "/health",
		StatusCode: 200,
		Direction:  DOutging,
		txBytes:    128,
		rxBytes:    64,
	}
	netflow.RecordPeerDomain("10.20.30.60", 443, "tcp", "42", "frontend.example.com")

	agg.Append(info, stats, "42", false, true, []string{"10.20.30.40"}, 15)
	pts := agg.ToPoint(nil, nil)
	if len(pts) != 1 {
		t.Fatalf("expected 1 point, got %d", len(pts))
	}
	if got := pts[0].GetTag("dst_domain"); got != "frontend.example.com" {
		t.Fatalf("unexpected dst_domain %q", got)
	}
}

func BenchmarkFlowAggTCPAppend(b *testing.B) {
	agg := &FlowAggTCP{}
	info := &PMeta{
		SrcIP:   "10.20.30.40",
		DstIP:   "10.20.30.50",
		SrcPort: 43210,
		DstPort: 8080,
	}
	nicIPList := []string{"10.20.30.40"}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		stats := &TCPMetrics{
			BytesRead:    64,
			BytesWritten: 128,
			RTT:          100,
			RTTVar:       20,
			Retransmits:  1,
		}
		agg.Append(info, stats, "42", directionOutgoing, false, true, nicIPList)
	}
}

func BenchmarkFlowAggHTTPAppend(b *testing.B) {
	agg := &FlowAggHTTP{}
	info := &PMeta{
		SrcIP:   "10.20.30.40",
		DstIP:   "10.20.30.50",
		SrcPort: 43210,
		DstPort: 8080,
	}
	nicIPList := []string{"10.20.30.40"}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		stats := &HTTPLogElem{
			Method:     "GET",
			Path:       "/bench",
			StatusCode: 200,
			Direction:  DIncoming,
			txBytes:    128,
			rxBytes:    64,
		}
		agg.Append(info, stats, "42", false, true, nicIPList, 15)
	}
}
