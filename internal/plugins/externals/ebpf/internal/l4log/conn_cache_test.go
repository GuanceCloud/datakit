//go:build linux
// +build linux

package l4log

import (
	"strconv"
	"testing"
	"time"
)

func TestTCPConnsGetValDropsNewConnectionWhenCacheLimitReached(t *testing.T) {
	conns := NewTCPConns(nil, "ctr", "ns", [2]string{"eth0", "aa:bb:cc:dd:ee:ff"}, &portListen{}, nil, nil)
	conns.connLimit = 1

	ts := time.Now().UnixNano()
	key1 := PMeta{SrcIP: "10.0.0.1", DstIP: "10.0.0.2", SrcPort: 12345, DstPort: 80}
	key2 := PMeta{SrcIP: "10.0.0.3", DstIP: "10.0.0.4", SrcPort: 12346, DstPort: 80}

	conns.conns.Lock()
	defer conns.conns.Unlock()

	if v, _, ok := conns.getVal(&key1, ts, false); !ok || v == nil {
		t.Fatal("first connection was not inserted")
	}
	if v, _, ok := conns.getVal(&key1, ts, false); !ok || v == nil {
		t.Fatal("existing connection should still be returned at the limit")
	}
	if v, _, ok := conns.getVal(&key2, ts, false); ok || v != nil {
		t.Fatal("new connection was inserted after cache limit was reached")
	}
	if got := conns.connCacheEntries(); got != 1 {
		t.Fatalf("cache entries = %d, want 1", got)
	}
}

func TestTCPConnsUpdateDropsNegativePayloadSize(t *testing.T) {
	conns := NewTCPConns(nil, "ctr", "ns", [2]string{"eth0", "aa:bb:cc:dd:ee:ff"}, &portListen{}, nil, nil)
	key := &PMeta{SrcIP: "10.0.0.1", DstIP: "10.0.0.2", SrcPort: 12345, DstPort: 80}
	ln := &PktTCPHdr{Seq: 1, TS: time.Now().UnixNano()}

	conns.update(directionTX, key, ln, 60, -1, nil, 0, false)

	conns.conns.Lock()
	defer conns.conns.Unlock()
	if got := conns.connCacheEntries(); got != 0 {
		t.Fatalf("connection entries = %d, want 0", got)
	}
}

func TestL4logConnCacheLimitEnv(t *testing.T) {
	t.Setenv(l4logConnCacheLimitEnv, "bad")
	if got := l4logConnCacheLimit(); got != defaultL4logConnCacheLimit {
		t.Fatalf("invalid env limit = %d, want %d", got, defaultL4logConnCacheLimit)
	}

	t.Setenv(l4logConnCacheLimitEnv, strconv.Itoa(maxL4logConnCacheLimit+1))
	if got := l4logConnCacheLimit(); got != maxL4logConnCacheLimit {
		t.Fatalf("clamped env limit = %d, want %d", got, maxL4logConnCacheLimit)
	}
}

func TestBlacklistCacheLimitEnv(t *testing.T) {
	t.Setenv(blacklistCacheLimitEnv, "bad")
	if got := blacklistCacheLimit(); got != defaultBlacklistCacheLimit {
		t.Fatalf("invalid env limit = %d, want %d", got, defaultBlacklistCacheLimit)
	}

	t.Setenv(blacklistCacheLimitEnv, strconv.Itoa(maxBlacklistCacheLimit+1))
	if got := blacklistCacheLimit(); got != maxBlacklistCacheLimit {
		t.Fatalf("clamped env limit = %d, want %d", got, maxBlacklistCacheLimit)
	}
}
