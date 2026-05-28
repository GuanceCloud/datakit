//go:build linux
// +build linux

package protodec

import "testing"

func TestH2DecPipeStreamLimit(t *testing.T) {
	t.Setenv(h2StreamLimitEnv, "1")

	dec := newH2DecPipe(ProtoHTTP2).(*h2DecPipe)
	if dec.GetElem(1) == nil {
		t.Fatal("first stream should be tracked")
	}
	if dec.GetElem(2) != nil {
		t.Fatal("second stream was created after limit was reached")
	}
	if got := len(dec.elems); got != 1 {
		t.Fatalf("stream entries = %d, want 1", got)
	}
}

func TestH2StreamLimitEnv(t *testing.T) {
	t.Setenv(h2StreamLimitEnv, "bad")
	if got := h2StreamLimit(); got != defaultH2StreamLimit {
		t.Fatalf("invalid env limit = %d, want %d", got, defaultH2StreamLimit)
	}

	t.Setenv(h2StreamLimitEnv, "9999999")
	if got := h2StreamLimit(); got != maxH2StreamLimit {
		t.Fatalf("clamped env limit = %d, want %d", got, maxH2StreamLimit)
	}
}
