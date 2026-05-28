//go:build linux
// +build linux

package l4log

import (
	"strconv"
	"testing"
)

func TestHTTP2LogGetElemHonorsStreamLimit(t *testing.T) {
	h2log := &HTTP2Log{streamLimit: 1}

	first := h2log.GetElem(1)
	if first == nil {
		t.Fatal("first stream was not created")
	}
	if second := h2log.GetElem(2); second != nil {
		t.Fatal("second stream was created after stream limit was reached")
	}

	first.hFinished = true
	second := h2log.GetElem(2)
	if second == nil {
		t.Fatal("second stream was not created after finished stream cleanup")
	}
	if len(h2log.elems) != 1 || h2log.elems[0].streamid != 2 {
		t.Fatalf("unexpected stream state: %+v", h2log.elems)
	}
}

func TestHTTP2StreamLimitEnv(t *testing.T) {
	t.Setenv(http2StreamLimitEnv, "bad")
	if got := http2StreamLimit(); got != defaultHTTP2StreamLimit {
		t.Fatalf("invalid env limit = %d, want %d", got, defaultHTTP2StreamLimit)
	}

	t.Setenv(http2StreamLimitEnv, strconv.Itoa(maxHTTP2StreamLimit+1))
	if got := http2StreamLimit(); got != maxHTTP2StreamLimit {
		t.Fatalf("clamped env limit = %d, want %d", got, maxHTTP2StreamLimit)
	}
}
