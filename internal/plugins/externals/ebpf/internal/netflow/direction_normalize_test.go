//go:build linux
// +build linux

package netflow

import (
	"math"
	"os"
	"path/filepath"
	"strings"
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

func TestDirectionFromTCPListenPorts(t *testing.T) {
	listenPorts := map[tcpListenPort]struct{}{
		{Netns: 42, Port: 50051}: {},
	}

	incoming := directionFromTCPListenPorts(DirectionOutgoing, ConnectionInfo{
		Saddr: [4]uint32{0, 0, 0, 0x0101006F},
		Daddr: [4]uint32{0, 0, 0, 0x0100006F},
		Sport: 50051,
		Dport: 45678,
		Netns: 42,
		Meta:  ConnL4TCP | ConnL3IPv4,
	}, listenPorts)
	if incoming != DirectionIncoming {
		t.Fatalf("expected source listen port to force incoming, got %q", incoming)
	}

	outgoing := directionFromTCPListenPorts(DirectionIncoming, ConnectionInfo{
		Saddr: [4]uint32{0, 0, 0, 0x0101006F},
		Daddr: [4]uint32{0, 0, 0, 0x0100006F},
		Sport: 45678,
		Dport: 50051,
		Netns: 42,
		Meta:  ConnL4TCP | ConnL3IPv4,
	}, listenPorts)
	if outgoing != DirectionOutgoing {
		t.Fatalf("expected destination listen port to force outgoing, got %q", outgoing)
	}
}

func TestParseTCPListenPorts(t *testing.T) {
	input := `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 00000000:C383 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 1 1 0000000000000000 100 0 0 10 0
   1: 0100007F:1F90 0100007F:C001 01 00000000:00000000 00:00000000 00000000     0        0 1 1 0000000000000000 100 0 0 10 0
`
	ports := make(map[tcpListenPort]struct{})
	if err := parseTCPListenPorts(strings.NewReader(input), 42, ports); err != nil {
		t.Fatal(err)
	}
	if _, ok := ports[tcpListenPort{Netns: 42, Port: 50051}]; !ok {
		t.Fatalf("expected listen port 50051, got %#v", ports)
	}
	if len(ports) != 1 {
		t.Fatalf("expected one listen port, got %#v", ports)
	}
}

func TestScanTCPListenPortsDedupsNetns(t *testing.T) {
	procRoot := t.TempDir()
	tcp := `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 00000000:C383 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 1 1 0000000000000000 100 0 0 10 0
`

	writeProcTCP := func(pid, netns string, content string) {
		pidRoot := filepath.Join(procRoot, pid)
		if err := os.MkdirAll(filepath.Join(pidRoot, "ns"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(pidRoot, "net"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(netns, filepath.Join(pidRoot, "ns", "net")); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(pidRoot, "net", "tcp"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeProcNetnsOnly := func(pid, netns string) {
		pidRoot := filepath.Join(procRoot, pid)
		if err := os.MkdirAll(filepath.Join(pidRoot, "ns"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(netns, filepath.Join(pidRoot, "ns", "net")); err != nil {
			t.Fatal(err)
		}
	}

	writeProcTCP("100", "net:[42]", tcp)
	writeProcTCP("101", "net:[42]", strings.ReplaceAll(tcp, "C383", "1F90"))
	writeProcTCP("200", "net:[43]", tcp)
	writeProcNetnsOnly("300", "net:[44]")
	writeProcTCP("301", "net:[44]", strings.ReplaceAll(tcp, "C383", "2382"))

	ports, err := scanTCPListenPorts(procRoot)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ports[tcpListenPort{Netns: 42, Port: 50051}]; !ok {
		t.Fatalf("expected netns 42 listen port 50051, got %#v", ports)
	}
	if _, ok := ports[tcpListenPort{Netns: 43, Port: 50051}]; !ok {
		t.Fatalf("expected netns 43 listen port 50051, got %#v", ports)
	}
	if _, ok := ports[tcpListenPort{Netns: 42, Port: 8080}]; ok {
		t.Fatalf("expected duplicate netns pid to be skipped, got %#v", ports)
	}
	if _, ok := ports[tcpListenPort{Netns: 44, Port: 9090}]; !ok {
		t.Fatalf("expected scan to fall back to a live pid in netns 44, got %#v", ports)
	}
	if len(ports) != 3 {
		t.Fatalf("expected three namespace-scoped ports, got %#v", ports)
	}
}
