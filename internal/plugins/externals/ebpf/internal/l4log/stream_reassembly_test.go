//go:build linux
// +build linux

package l4log

import (
	"bytes"
	"testing"
)

func TestStreamReassemblerInOrder(t *testing.T) {
	sr := newStreamReassembler(8, 1024, true)

	res := sr.Push(100, 0, []byte("hello"), 1)
	if len(res.Deliveries) != 1 {
		t.Fatalf("deliveries = %d, want 1", len(res.Deliveries))
	}
	if got := string(res.Deliveries[0].Payload); got != "hello" {
		t.Fatalf("payload = %q, want %q", got, "hello")
	}
	if sr.nextSeq != 105 {
		t.Fatalf("nextSeq = %d, want 105", sr.nextSeq)
	}
}

func TestStreamReassemblerOutOfOrderDrain(t *testing.T) {
	sr := newStreamReassembler(8, 1024, true)

	_ = sr.Push(100, 0, []byte("hello"), 1)

	first := sr.Push(110, 0, []byte("world"), 2)
	if !first.Buffered {
		t.Fatalf("first push should buffer out-of-order segment")
	}

	second := sr.Push(105, 0, []byte("_____"), 3)
	if len(second.Deliveries) != 2 {
		t.Fatalf("deliveries = %d, want 2", len(second.Deliveries))
	}
	if got := string(second.Deliveries[0].Payload); got != "_____" {
		t.Fatalf("delivery[0] = %q, want %q", got, "_____")
	}
	if got := string(second.Deliveries[1].Payload); got != "world" {
		t.Fatalf("delivery[1] = %q, want %q", got, "world")
	}
}

func TestStreamReassemblerRetransmit(t *testing.T) {
	sr := newStreamReassembler(8, 1024, true)

	_ = sr.Push(100, 0, []byte("hello"), 1)
	res := sr.Push(100, 0, []byte("hello"), 2)
	if !res.Retransmit {
		t.Fatalf("expected retransmit classification")
	}
	if len(res.Deliveries) != 0 {
		t.Fatalf("unexpected deliveries on retransmit: %d", len(res.Deliveries))
	}
}

func TestStreamReassemblerOverlapTrim(t *testing.T) {
	sr := newStreamReassembler(8, 1024, true)

	_ = sr.Push(100, 0, []byte("hello"), 1)
	res := sr.Push(103, 0, []byte("lo!!!"), 2)
	if len(res.Deliveries) != 1 {
		t.Fatalf("deliveries = %d, want 1", len(res.Deliveries))
	}
	if got := string(res.Deliveries[0].Payload); got != "!!!" {
		t.Fatalf("payload = %q, want %q", got, "!!!")
	}
}

func TestStreamReassemblerGapResync(t *testing.T) {
	sr := newStreamReassembler(1, 8, true)

	_ = sr.Push(100, 0, []byte("hello"), 1)
	res := sr.Push(200, 0, []byte("world"), 2)
	if !res.Gap {
		t.Fatalf("expected gap resync")
	}
	if len(res.Deliveries) != 1 {
		t.Fatalf("deliveries = %d, want 1", len(res.Deliveries))
	}
	if got := string(res.Deliveries[0].Payload); got != "world" {
		t.Fatalf("payload = %q, want %q", got, "world")
	}
}

func TestBufferedSegmentCopiesPayload(t *testing.T) {
	orig := []byte("world")
	seg := bufferedStreamSegment(105, 0, orig, 2, true)
	orig[0] = 'X'

	if !bytes.Equal(seg.payload, []byte("world")) {
		t.Fatalf("buffered payload = %q, want original copy", seg.payload)
	}
}

func TestBufferedSegmentSkipsPayloadWhenDisabled(t *testing.T) {
	seg := bufferedStreamSegment(105, 0, []byte("world"), 2, false)
	if seg.payload != nil {
		t.Fatalf("payload = %q, want nil when payload capture is disabled", seg.payload)
	}
	if seg.endSeq != 110 {
		t.Fatalf("endSeq = %d, want 110", seg.endSeq)
	}
}

func BenchmarkStreamReassemblerPushInOrder(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sr := newStreamReassembler(8, 1024, true)
		seq := uint32(100)
		for j := 0; j < 32; j++ {
			payload := []byte("hello")
			_ = sr.Push(seq, 0, payload, int64(j))
			seq += uint32(len(payload))
		}
	}
}

func BenchmarkStreamReassemblerPushReorder(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sr := newStreamReassembler(8, 1024, true)
		_ = sr.Push(100, 0, []byte("hello"), 1)

		seq := uint32(110)
		for j := 0; j < 8; j++ {
			_ = sr.Push(seq, 0, []byte("world"), int64(j+2))
			seq += 5
		}

		seq = 105
		for j := 0; j < 8; j++ {
			_ = sr.Push(seq, 0, []byte("_____"), int64(j+10))
			seq += 10
		}
	}
}

func BenchmarkStreamReassemblerPushReorderNoPayload(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sr := newStreamReassembler(8, 1024, false)
		_ = sr.Push(100, 0, []byte("hello"), 1)

		seq := uint32(110)
		for j := 0; j < 8; j++ {
			_ = sr.Push(seq, 0, []byte("world"), int64(j+2))
			seq += 5
		}

		seq = 105
		for j := 0; j < 8; j++ {
			_ = sr.Push(seq, 0, []byte("_____"), int64(j+10))
			seq += 10
		}
	}
}
