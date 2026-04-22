//go:build linux
// +build linux

package offset

import "testing"

func TestConntrackSeedOffsets(t *testing.T) {
	t.Run("modern kernel", func(t *testing.T) {
		ctNet, nsInum := conntrackSeedOffsets(0x0006001100000000)
		if ctNet != 136 || nsInum != 168 {
			t.Fatalf("unexpected modern seeds: ct_net=%d ns_inum=%d", ctNet, nsInum)
		}
	})

	t.Run("legacy kernel", func(t *testing.T) {
		ctNet, nsInum := conntrackSeedOffsets(0x0004000f00000000)
		if ctNet != 17 || nsInum != 112 {
			t.Fatalf("unexpected legacy seeds: ct_net=%d ns_inum=%d", ctNet, nsInum)
		}
	})
}

func TestConntrackSeedTupleOffsets(t *testing.T) {
	origin, reply := conntrackSeedTupleOffsets()
	if origin != 32 || reply != 88 {
		t.Fatalf("unexpected tuple seeds: origin=%d reply=%d", origin, reply)
	}
}

func TestNewConntrackConstEditorUsesDedicatedNetnsOffset(t *testing.T) {
	offset := &OffsetConntrackC{
		offset_ct_net:            136,
		offset_ct_ns_common_inum: 168,
		offset_ct_origin_tuple:   32,
		offset_ct_reply_tuple:    88,
	}

	patches := newConntrackConstEditor(offset)
	if len(patches) != 4 {
		t.Fatalf("unexpected patch count: %d", len(patches))
	}

	found := false
	for _, patch := range patches {
		if patch.Name == "offset_ct_ns_common_inum" {
			found = true
			if patch.Value != uint64(168) {
				t.Fatalf("unexpected conntrack netns patch value: %#v", patch.Value)
			}
		}
	}

	if !found {
		t.Fatal("expected conntrack netns patch")
	}
}
