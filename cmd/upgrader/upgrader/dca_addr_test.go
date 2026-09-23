// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package upgrader

import (
	"net"
	"testing"
	"time"
)

// TestDialableListenAddr: datakit's listen address may be a wildcard, which is
// not a valid connect target (on Windows the dial is refused), so the upgrader
// must fall back to loopback.
func TestDialableListenAddr(t *testing.T) {
	cases := map[string]string{
		"0.0.0.0:9529":      "127.0.0.1:9529",
		"[::]:9529":         "[::1]:9529",
		":9529":             "127.0.0.1:9529",
		"127.0.0.1:9529":    "127.0.0.1:9529",
		"localhost:9529":    "localhost:9529",
		"192.168.1.10:9529": "192.168.1.10:9529",
		"9529":              "9529", // not a host:port pair, kept untouched
		"":                  "",
	}

	for in, want := range cases {
		if got := dialableListenAddr(in); got != want {
			t.Errorf("dialableListenAddr(%q) = %q, want %q", in, got, want)
		}
	}
}

// DataKit explicitly uses tcp6 for IPv6 listen addresses, not a dual-stack socket.
func TestDialableListenAddrIPv6(t *testing.T) {
	listener, err := net.Listen("tcp6", "[::]:0")
	if err != nil {
		t.Skipf("IPv6 listener unavailable: %s", err)
	}
	defer listener.Close() //nolint:errcheck
	conn, err := net.DialTimeout("tcp", dialableListenAddr(listener.Addr().String()), time.Second)
	if err != nil {
		t.Fatalf("upgrader cannot dial IPv6 DataKit listener: %s", err)
	}
	defer conn.Close() //nolint:errcheck
}
