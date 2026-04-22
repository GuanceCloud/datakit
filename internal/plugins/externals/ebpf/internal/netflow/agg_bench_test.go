//go:build linux
// +build linux

package netflow

import "testing"

func TestFlowAggAppendAggregatesTCPStats(t *testing.T) {
	agg := &FlowAgg{}
	info := ConnectionInfo{
		Saddr:       [4]uint32{0, 0, 0, 0x0101008F},
		Daddr:       [4]uint32{0, 0, 0, 0x0200000A},
		Sport:       43210,
		Dport:       8080,
		Pid:         99,
		Netns:       42,
		Meta:        ConnL3IPv4 | ConnL4TCP,
		ProcessName: "nginx",
	}
	stats := ConnFullStats{
		Stats: ConnectionStats{
			RecvBytes:   100,
			SentBytes:   200,
			RecvPackets: 3,
			SentPackets: 5,
			Direction:   ConnDirectionOutgoing,
		},
		TCPStats: ConnectionTCPStats{
			Retransmits:     1,
			Rtt:             1000,
			RttVar:          100,
			ConnectAttempts: 1,
			ConnectFailures: 1,
			CloseWait:       1,
			LastAck:         1,
			TimeWait:        1,
		},
		TotalClosed:      1,
		TotalEstablished: 1,
	}

	if err := agg.Append(info, stats); err != nil {
		t.Fatalf("append failed: %v", err)
	}
	if agg.Len() != 1 {
		t.Fatalf("expected 1 flow, got %d", agg.Len())
	}

	for _, value := range agg.data {
		if value.bytesRead != 100 || value.bytesWritten != 200 {
			t.Fatalf("unexpected bytes aggregation: %+v", value)
		}
		if value.packetsRead != 3 || value.packetsWrite != 5 {
			t.Fatalf("unexpected packet aggregation: %+v", value)
		}
		if value.retransmits != 1 || value.tcpClosed != 1 || value.tcpEstablished != 1 {
			t.Fatalf("unexpected tcp aggregation: %+v", value)
		}
		if value.rtt != 1000 || value.rttVar != 100 {
			t.Fatalf("unexpected latency aggregation: %+v", value)
		}
		if value.tcpConnects != 1 || value.tcpFailures != 1 || value.tcpCloseWait != 1 ||
			value.tcpLastAck != 1 || value.tcpTimeWait != 1 {
			t.Fatalf("unexpected tcp event aggregation: %+v", value)
		}
	}
}

func BenchmarkFlowAggAppend(b *testing.B) {
	agg := &FlowAgg{}
	info := ConnectionInfo{
		Saddr:       [4]uint32{0, 0, 0, 0x0101008F},
		Daddr:       [4]uint32{0, 0, 0, 0x0200000A},
		Sport:       43210,
		Dport:       8080,
		Pid:         99,
		Netns:       42,
		Meta:        ConnL3IPv4 | ConnL4TCP,
		ProcessName: "nginx",
	}
	stats := ConnFullStats{
		Stats: ConnectionStats{
			RecvBytes:   100,
			SentBytes:   200,
			RecvPackets: 3,
			SentPackets: 5,
			Direction:   ConnDirectionOutgoing,
		},
		TCPStats: ConnectionTCPStats{
			Retransmits:     1,
			Rtt:             1000,
			RttVar:          100,
			ConnectAttempts: 1,
			ConnectFailures: 1,
			CloseWait:       1,
			LastAck:         1,
			TimeWait:        1,
		},
		TotalClosed:      1,
		TotalEstablished: 1,
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := agg.Append(info, stats); err != nil {
			b.Fatal(err)
		}
	}
}
