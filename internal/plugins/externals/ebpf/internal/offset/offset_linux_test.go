//go:build linux
// +build linux

package offset

import (
	"net"
	"testing"
)

type testAddr string

func (a testAddr) Network() string { return "tcp" }
func (a testAddr) String() string  { return string(a) }

func TestPortFromAddr(t *testing.T) {
	for _, tc := range []struct {
		name string
		addr net.Addr
		want uint16
	}{
		{name: "ipv4", addr: testAddr("127.0.0.1:12345"), want: 12345},
		{name: "ipv6", addr: testAddr("[::1]:23456"), want: 23456},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := portFromAddr(tc.addr)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, got)
			}
		})
	}
}

func TestPortFromAddrRejectsInvalidAddress(t *testing.T) {
	if _, err := portFromAddr(testAddr("127.0.0.1")); err == nil {
		t.Fatal("expected invalid address error")
	}
	if _, err := portFromAddr(nil); err == nil {
		t.Fatal("expected nil address error")
	}
}

func TestProbeServersRejectNilContext(t *testing.T) {
	if _, err := runTCPServer(nil, "tcp4", "127.0.0.1"); err == nil {
		t.Fatal("expected TCP server to reject nil context")
	}
	if _, err := runUDPServer(nil, "udp4", "127.0.0.1"); err == nil {
		t.Fatal("expected UDP server to reject nil context")
	}
}
