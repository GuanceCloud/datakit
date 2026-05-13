//go:build linux
// +build linux

package l4log

import (
	"sync"
	"testing"
)

func TestConnMapShrinksAfterDeletes(t *testing.T) {
	cm := newConnMap()

	for i := 0; i < 512; i++ {
		cm.insert(PMeta{SrcPort: uint16(i + 1)}, &PValue{})
	}

	for i := 0; i < 205; i++ {
		cm.delete(PMeta{SrcPort: uint16(i + 1)})
	}

	if got, want := len(cm.m), 307; got != want {
		t.Fatalf("unexpected map length: got %d want %d", got, want)
	}
	if cm.insertCount != len(cm.m) {
		t.Fatalf("expected insertCount to shrink with map, got insertCount=%d len=%d", cm.insertCount, len(cm.m))
	}
	if cm.deleteCount != 0 {
		t.Fatalf("expected deleteCount to reset after shrink, got %d", cm.deleteCount)
	}
}

func TestTCPLogTrimsChunksByDefaultLimit(t *testing.T) {
	t.Setenv(envNetlogMaxChunksPerConn, "")
	netlogMemoryConfigOnce = sync.Once{}

	var log TCPLog
	for i := 0; i < defaultNetlogMaxChunksPerConn+3; i++ {
		log.GetPktChunk(true, true)
	}

	if got, want := len(log.chunk), defaultNetlogMaxChunksPerConn; got != want {
		t.Fatalf("unexpected chunk count: got %d want %d", got, want)
	}
	if got := log.chunk[0].ChunkID; got != 4 {
		t.Fatalf("expected oldest retained chunk id to be 4, got %d", got)
	}
}

func BenchmarkConnMapInsertDelete(b *testing.B) {
	keys := make([]PMeta, 1024)
	for i := range keys {
		keys[i] = PMeta{
			SrcIP:   "10.0.0.1",
			DstIP:   "10.0.0.2",
			SrcPort: uint16(10000 + i),
			DstPort: 80,
		}
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cm := newConnMap()
		for j := range keys {
			cm.insert(keys[j], &PValue{})
		}
		for j := 0; j < len(keys)/2; j++ {
			cm.delete(keys[j])
		}
	}
}

func BenchmarkConnMapSteadyStateChurn(b *testing.B) {
	const window = 2048

	keys := make([]PMeta, window)
	for i := range keys {
		keys[i] = PMeta{
			SrcIP:   "10.0.0.1",
			DstIP:   "10.0.0.2",
			SrcPort: uint16(10000 + i),
			DstPort: 80,
		}
	}

	cm := newConnMap()
	val := &PValue{}
	for i := range keys {
		cm.insert(keys[i], val)
	}

	nextPort := 10000 + window
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		slot := i % window
		cm.delete(keys[slot])
		keys[slot] = PMeta{
			SrcIP:   "10.0.0.1",
			DstIP:   "10.0.0.2",
			SrcPort: uint16(nextPort),
			DstPort: 80,
		}
		nextPort++
		cm.insert(keys[slot], val)
	}
}
