//go:build linux
// +build linux

package bpfutil

import "testing"

func TestPerfRingBufferPagesHonorsMemoryBudget(t *testing.T) {
	if got := perfRingBufferPages(32, 16*1024*1024, 4, 4096, 1); got != 32 {
		t.Fatalf("normal perf pages = %d, want 32", got)
	}
	if got := perfRingBufferPages(32, 16*1024*1024, 4, 4096, 1024); got != 4 {
		t.Fatalf("min-budget perf pages = %d, want 4", got)
	}
	if got := perfRingBufferPages(32, 16*1024*1024, 4, 4096, 4096); got != 1 {
		t.Fatalf("over-min-budget perf pages = %d, want 1", got)
	}
}

func TestPerfRingBufferPagesFallbacks(t *testing.T) {
	if got := perfRingBufferPages(0, 0, 0, 4096, 1); got != 1 {
		t.Fatalf("zero config perf pages = %d, want 1", got)
	}
	if got := perfRingBufferPages(2, 0, 4, 4096, 1); got != 4 {
		t.Fatalf("no budget perf pages = %d, want min 4", got)
	}
}
