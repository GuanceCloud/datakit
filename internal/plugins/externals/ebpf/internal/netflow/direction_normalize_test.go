//go:build linux
// +build linux

package netflow

import (
	"math"
	"testing"
)

func TestNormalizeDirectionAndPorts(t *testing.T) {
	tcases := []struct {
		name      string
		direction string
		sport     uint32
		dport     uint32
		wantDir   string
		wantSPort uint32
		wantDPort uint32
	}{
		{
			name:      "keep outgoing client ephemeral port rollup",
			direction: DirectionOutgoing,
			sport:     43210,
			dport:     80,
			wantDir:   DirectionOutgoing,
			wantSPort: math.MaxUint32,
			wantDPort: 80,
		},
		{
			name:      "flip to incoming when source looks like server port",
			direction: DirectionOutgoing,
			sport:     80,
			dport:     43210,
			wantDir:   DirectionIncoming,
			wantSPort: 80,
			wantDPort: math.MaxUint32,
		},
		{
			name:      "infer outgoing when direction unknown",
			direction: DirectionUnknown,
			sport:     52345,
			dport:     6379,
			wantDir:   DirectionOutgoing,
			wantSPort: math.MaxUint32,
			wantDPort: 6379,
		},
		{
			name:      "leave stable ports untouched when both look like service ports",
			direction: DirectionUnknown,
			sport:     8080,
			dport:     9090,
			wantDir:   DirectionUnknown,
			wantSPort: 8080,
			wantDPort: 9090,
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			gotDir, gotSPort, gotDPort := NormalizeDirectionAndPorts(tc.direction, tc.sport, tc.dport)
			if gotDir != tc.wantDir || gotSPort != tc.wantSPort || gotDPort != tc.wantDPort {
				t.Fatalf("NormalizeDirectionAndPorts(%q, %d, %d) = (%q, %d, %d), want (%q, %d, %d)",
					tc.direction, tc.sport, tc.dport,
					gotDir, gotSPort, gotDPort,
					tc.wantDir, tc.wantSPort, tc.wantDPort)
			}
		})
	}
}

func TestAddClientServerInfIncomingUsesSourceIPType(t *testing.T) {
	tags, fields := AddClientServerInf(map[string]string{
		"direction":   DirectionIncoming,
		"src_ip":      "10.0.0.1",
		"src_ip_type": "other",
		"src_port":    "80",
		"dst_ip":      "10.0.0.2",
		"dst_ip_type": "private",
		"dst_port":    "*",
	}, map[string]any{
		"bytes_read":    int64(10),
		"bytes_written": int64(20),
	})

	if got := tags["server_ip_type"]; got != "other" {
		t.Fatalf("unexpected server_ip_type %q", got)
	}
	if got := tags["client_port"]; got != "*" {
		t.Fatalf("unexpected client_port %q", got)
	}
	if got := fields["client_sent"]; got != int64(10) {
		t.Fatalf("unexpected client_sent %v", got)
	}
	if got := fields["server_sent"]; got != int64(20) {
		t.Fatalf("unexpected server_sent %v", got)
	}
}
