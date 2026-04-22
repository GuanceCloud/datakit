//go:build linux
// +build linux

package offset

import "testing"

func TestTaskStructFilesGuessSequenceForKernel3xPrefersCentOSAnchor(t *testing.T) {
	got := taskStructFilesGuessSequenceForKernel(3 << 48)
	if len(got) == 0 {
		t.Fatal("empty guess sequence")
	}
	if got[0] != 1880 {
		t.Fatalf("expected first 3.x anchor 1880, got %d", got[0])
	}
}

func TestTaskStructFilesGuessSequenceForKernel415PrefersUbuntuAnchor(t *testing.T) {
	got := taskStructFilesGuessSequenceForKernel(4<<48 | 15<<32)
	if len(got) == 0 {
		t.Fatal("empty guess sequence")
	}
	if got[0] != 2704 {
		t.Fatalf("expected first 4.15 anchor 2704, got %d", got[0])
	}
}

func TestTaskStructFilesGuessSequenceCoversLinearFallbackWithoutDuplicates(t *testing.T) {
	got := taskStructFilesGuessSequenceForKernel(0)
	if len(got) == 0 {
		t.Fatal("empty guess sequence")
	}

	seen := make(map[int32]struct{}, len(got))
	for _, candidate := range got {
		if candidate < httpTaskFilesGuessStart || candidate >= httpTaskFilesGuessEnd {
			t.Fatalf("candidate out of range: %d", candidate)
		}
		if candidate%httpTaskFilesGuessStep != 0 {
			t.Fatalf("candidate not aligned: %d", candidate)
		}
		if _, ok := seen[candidate]; ok {
			t.Fatalf("duplicate candidate: %d", candidate)
		}
		seen[candidate] = struct{}{}
	}

	for candidate := int32(httpTaskFilesGuessStart); candidate < httpTaskFilesGuessEnd; candidate += httpTaskFilesGuessStep {
		if _, ok := seen[candidate]; !ok {
			t.Fatalf("missing fallback candidate: %d", candidate)
		}
	}
}
