//go:build linux
// +build linux

package l4log

import (
	"strconv"
	"testing"
)

func TestHTTPLogNewElemHonorsLimit(t *testing.T) {
	h := &HTTPLog{elemLimit: 1}

	first := h.newElem()
	if first == nil {
		t.Fatal("first http elem was not created")
	}
	if second := h.newElem(); second != nil {
		t.Fatal("second http elem was created after limit was reached")
	}

	first.hFinished = true
	second := h.newElem()
	if second == nil {
		t.Fatal("second http elem was not created after finished elem cleanup")
	}
	if len(h.elems) != 1 || h.elems[0] != second {
		t.Fatalf("unexpected http elem state: %+v", h.elems)
	}
}

func TestHTTPElemLimitEnv(t *testing.T) {
	t.Setenv(httpElemLimitEnv, "bad")
	if got := httpElemLimit(); got != defaultHTTPElemLimit {
		t.Fatalf("invalid env limit = %d, want %d", got, defaultHTTPElemLimit)
	}

	t.Setenv(httpElemLimitEnv, strconv.Itoa(maxHTTPElemLimit+1))
	if got := httpElemLimit(); got != maxHTTPElemLimit {
		t.Fatalf("clamped env limit = %d, want %d", got, maxHTTPElemLimit)
	}
}

func TestHTTPLogDropsPendingPacketsAtLimit(t *testing.T) {
	h := &HTTPLog{}
	ln := &PktTCPHdr{Seq: 100, TS: 1}

	if err := h.Handle(nil, directionTX, []byte("GET /"), 5, ln, nil, 0, 1); err != nil {
		t.Fatalf("first handle failed: %v", err)
	}
	if got := len(h.txState.pendingPackets); got != 1 {
		t.Fatalf("pending packets = %d, want 1", got)
	}
	h.txState.maxPendingPackets = 1

	ln.Seq += 5
	if err := h.Handle(nil, directionTX, []byte("x"), 1, ln, nil, 0, 2); err != nil {
		t.Fatalf("second handle failed: %v", err)
	}
	if len(h.txState.pendingPackets) != 0 || h.txState.header.active {
		t.Fatalf("expected pending HTTP state to be dropped, got packets=%d active=%v",
			len(h.txState.pendingPackets), h.txState.header.active)
	}
}

func TestHTTPPendingPacketLimitEnv(t *testing.T) {
	t.Setenv(httpPendingPacketLimitEnv, "bad")
	if got := httpPendingPacketLimit(); got != defaultHTTPPendingPackets {
		t.Fatalf("invalid pending packet limit = %d, want %d", got, defaultHTTPPendingPackets)
	}

	t.Setenv(httpPendingPacketLimitEnv, strconv.Itoa(maxHTTPPendingPackets+1))
	if got := httpPendingPacketLimit(); got != maxHTTPPendingPackets {
		t.Fatalf("clamped pending packet limit = %d, want %d", got, maxHTTPPendingPackets)
	}
}
