//go:build linux
// +build linux

package netflow

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netpath"
)

type capturePathScheduler struct {
	candidates []netpath.Candidate
}

func (s *capturePathScheduler) Schedule(candidate netpath.Candidate) bool {
	s.candidates = append(s.candidates, candidate)
	return true
}

func TestSchedulePathCandidateUsesOutgoingTCP(t *testing.T) {
	scheduler := &capturePathScheduler{}
	tracer := NewNetFlowTracer(nil)
	tracer.SetPathScheduler(scheduler)

	tracer.schedulePathCandidate(ConnectionInfo{
		Saddr:       [4]uint32{0, 0, 0, 0x0101A8C0},
		Daddr:       [4]uint32{0, 0, 0, 0x0A7100CB},
		Sport:       50000,
		Dport:       443,
		Pid:         1234,
		Netns:       42,
		Meta:        ConnL3IPv4 | ConnL4TCP,
		ProcessName: "curl",
		ServiceName: "checkout",
	}, ConnFullStats{
		Stats: ConnectionStats{
			Direction: ConnDirectionOutgoing,
			SentBytes: 128,
			RecvBytes: 256,
		},
		TCPStats: ConnectionTCPStats{
			Rtt:             1200,
			ConnectAttempts: 1,
		},
	})

	if len(scheduler.candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(scheduler.candidates))
	}
	got := scheduler.candidates[0]
	if got.TargetIP != "203.0.113.10" {
		t.Fatalf("unexpected target ip %q", got.TargetIP)
	}
	if got.Port != 443 {
		t.Fatalf("unexpected port %d", got.Port)
	}
	if got.Source.ProcessName != "curl" || got.Source.ServiceName != "checkout" {
		t.Fatalf("unexpected source %#v", got.Source)
	}
	if got.Source.IP != "192.168.1.1" || got.Source.Port != 50000 {
		t.Fatalf("unexpected source tuple %#v", got.Source)
	}
	if got.DstIP != "203.0.113.10" || got.DstPort != 443 {
		t.Fatalf("unexpected destination tuple %s:%d", got.DstIP, got.DstPort)
	}
	if len(got.Tags) != 0 {
		t.Fatalf("eBPF candidate must not duplicate structured attributes in tags: %#v", got.Tags)
	}
}

func TestSchedulePathCandidateUsesCanonicalNATTags(t *testing.T) {
	scheduler := &capturePathScheduler{}
	tracer := NewNetFlowTracer(nil)
	tracer.SetPathScheduler(scheduler)

	tracer.schedulePathCandidate(ConnectionInfo{
		Saddr:    [4]uint32{0, 0, 0, 0x0101A8C0},
		Daddr:    [4]uint32{0, 0, 0, 0x0A7100CB},
		Dport:    443,
		NATDaddr: [4]uint32{0, 0, 0, 0x146433C6},
		NATDport: 8443,
		Meta:     ConnL3IPv4 | ConnL4TCP,
	}, ConnFullStats{
		Stats: ConnectionStats{Direction: ConnDirectionOutgoing},
	})

	if len(scheduler.candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(scheduler.candidates))
	}
	got := scheduler.candidates[0]
	if got.TargetIP != "198.51.100.20" || got.Port != 8443 {
		t.Fatalf("unexpected NAT target %s:%d", got.TargetIP, got.Port)
	}
	if got.DstIP != "203.0.113.10" || got.DstPort != 443 {
		t.Fatalf("unexpected original destination %s:%d", got.DstIP, got.DstPort)
	}
	if len(got.Tags) != 0 {
		t.Fatalf("eBPF candidate must not duplicate NAT attributes in tags: %#v", got.Tags)
	}
}

func TestSchedulePathCandidateSkipsIncoming(t *testing.T) {
	scheduler := &capturePathScheduler{}
	tracer := NewNetFlowTracer(nil)
	tracer.SetPathScheduler(scheduler)

	tracer.schedulePathCandidate(ConnectionInfo{
		Saddr: [4]uint32{0, 0, 0, 0x0101A8C0},
		Daddr: [4]uint32{0, 0, 0, 0x0A7100CB},
		Dport: 443,
		Meta:  ConnL3IPv4 | ConnL4TCP,
	}, ConnFullStats{
		Stats: ConnectionStats{Direction: ConnDirectionIncoming},
	})

	if len(scheduler.candidates) != 0 {
		t.Fatalf("expected no candidates, got %d", len(scheduler.candidates))
	}
}
