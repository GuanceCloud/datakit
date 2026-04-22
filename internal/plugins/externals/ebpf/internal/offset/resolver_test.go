//go:build linux
// +build linux

package offset

import "testing"

func TestResolveKernelOffsetsUsesCachedGuess(t *testing.T) {
	saved := &OffsetGuessC{
		offset_sk_num:          14,
		offset_sk_family:       16,
		offset_sk_rcv_saddr:    4,
		offset_sk_dport:        12,
		offset_flowi4_saddr:    20,
		offset_flowi4_daddr:    24,
		offset_flowi4_sport:    30,
		offset_flowi4_dport:    28,
		offset_sk_v6_rcv_saddr: 72,
		offset_sk_v6_daddr:     56,
		offset_flowi6_saddr:    36,
		offset_flowi6_daddr:    20,
		offset_flowi6_sport:    58,
		offset_flowi6_dport:    56,
	}

	plan, err := ResolveKernelOffsets(saved, false, false)
	if err != nil {
		t.Fatalf("ResolveKernelOffsets: %v", err)
	}
	if plan == nil || plan.Guess == nil {
		t.Fatal("expected cached kernel plan")
	}
	if plan.Guess.offset_flowi4_saddr != saved.offset_flowi4_saddr ||
		plan.Guess.offset_flowi6_dport != saved.offset_flowi6_dport {
		t.Fatalf("unexpected cached kernel guess: %+v", plan.Guess)
	}
	if len(plan.Patches) == 0 {
		t.Fatal("expected kernel patches from cached guess")
	}
}

func TestConntrackGuessFromOffset(t *testing.T) {
	offset := &OffsetGuessC{
		offset_ct_net:            64,
		offset_ct_ns_common_inum: 84,
		offset_origin_tuple:      72,
		offset_reply_tuple:       104,
	}

	got := ConntrackGuessFromOffset(offset)
	if got == nil {
		t.Fatal("expected cached conntrack guess")
	}

	if got.offset_ct_net != 64 ||
		got.offset_ct_origin_tuple != 72 ||
		got.offset_ct_reply_tuple != 104 ||
		got.offset_ct_ns_common_inum != 84 {
		t.Fatalf("unexpected conntrack guess: %+v", got)
	}
}

func TestConntrackGuessFromOffsetIncomplete(t *testing.T) {
	offset := &OffsetGuessC{
		offset_ct_ns_common_inum: 84,
	}

	got := ConntrackGuessFromOffset(offset)
	if got == nil {
		t.Fatal("expected partial conntrack guess when ns_common.inum is known")
	}
	if got.offset_ct_ns_common_inum != 84 {
		t.Fatalf("unexpected partial conntrack guess: %+v", got)
	}
	if got.offset_ct_net != 0 || got.offset_ct_origin_tuple != 32 || got.offset_ct_reply_tuple != 88 {
		t.Fatalf("unexpected non-zero partial conntrack fields: %+v", got)
	}
}

func TestConntrackGuessFromOffsetFallsBackToKernelNetOffsets(t *testing.T) {
	offset := &OffsetGuessC{
		offset_sk_net:            17,
		offset_ns_common_inum:    69,
		offset_ct_ns_common_inum: 0,
		offset_ct_net:            0,
	}

	got := ConntrackGuessFromOffset(offset)
	if got == nil {
		t.Fatal("expected conntrack guess from kernel offsets")
	}

	if got.offset_ct_net != 17 {
		t.Fatalf("unexpected conntrack net seed: %d", got.offset_ct_net)
	}
	if got.offset_ct_ns_common_inum != 69 {
		t.Fatalf("unexpected conntrack netns seed: %d", got.offset_ct_ns_common_inum)
	}
	if got.offset_ct_origin_tuple != 32 || got.offset_ct_reply_tuple != 88 {
		t.Fatalf("unexpected tuple seeds: %+v", got)
	}
}

func TestSeedConntrackTupleOffsets(t *testing.T) {
	got := newGuessConntrack()
	seedConntrackTupleOffsets(&got)

	if got.offset_ct_origin_tuple != 32 || got.offset_ct_reply_tuple != 88 {
		t.Fatalf("unexpected conntrack tuple seeds: %+v", got)
	}
}

func TestConntrackTupleOffsetsReady(t *testing.T) {
	if conntrackTupleOffsetsReady(nil) {
		t.Fatal("nil offset should not be ready")
	}

	got := &OffsetConntrackC{}
	if conntrackTupleOffsetsReady(got) {
		t.Fatal("zero offset should not be ready")
	}

	got.offset_ct_origin_tuple = 32
	got.offset_ct_reply_tuple = 88
	if !conntrackTupleOffsetsReady(got) {
		t.Fatal("tuple offsets should be ready")
	}
}
