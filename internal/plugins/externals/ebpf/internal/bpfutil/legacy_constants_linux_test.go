//go:build linux
// +build linux

package bpfutil

import (
	"testing"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/asm"
)

func TestRewriteLegacyConstants(t *testing.T) {
	sentinel, err := sentinelConstant("offset_sk_num")
	if err != nil {
		t.Fatalf("sentinelConstant: %v", err)
	}

	spec := &ebpf.CollectionSpec{
		Programs: map[string]*ebpf.ProgramSpec{
			"test_prog": {
				Instructions: asm.Instructions{
					asm.LoadImm(asm.R1, sentinel, asm.DWord),
					asm.Return(),
				},
			},
		},
	}

	err = rewriteLegacyConstants(spec, []Constant{
		{Name: "offset_sk_num", Value: uint64(42)},
	})
	if err != nil {
		t.Fatalf("rewriteLegacyConstants: %v", err)
	}

	got := spec.Programs["test_prog"].Instructions[0].Constant
	if got != 42 {
		t.Fatalf("unexpected patched constant: got %d, want 42", got)
	}
}

func TestSentinelConstantIncludesConntrackNetns(t *testing.T) {
	got, err := sentinelConstant("offset_ct_ns_common_inum")
	if err != nil {
		t.Fatalf("sentinelConstant: %v", err)
	}
	if got == 0 {
		t.Fatal("expected conntrack netns sentinel")
	}
}

func TestSentinelConstantIncludesApiflowMinCaptureSize(t *testing.T) {
	got, err := sentinelConstant("apiflow_min_capture_size")
	if err != nil {
		t.Fatalf("sentinelConstant: %v", err)
	}
	if got == 0 {
		t.Fatal("expected apiflow min capture size sentinel")
	}
}

func TestKernelSupportsReadOnlyMaps(t *testing.T) {
	if kernelSupportsReadOnlyMaps(0x0005000100000000) {
		t.Fatal("kernel 5.1 should not support read-only maps")
	}
	if !kernelSupportsReadOnlyMaps(0x0005000200000000) {
		t.Fatal("kernel 5.2 should support read-only maps")
	}
}

func TestKernelSupportsLRUHashMaps(t *testing.T) {
	if kernelSupportsLRUHashMaps(0x0004000900000000) {
		t.Fatal("kernel 4.9 should not support LRU hash maps")
	}
	if !kernelSupportsLRUHashMaps(0x0004000a00000000) {
		t.Fatal("kernel 4.10 should support LRU hash maps")
	}
}

func TestKernelSupportsRetprobeMaxActiveOverride(t *testing.T) {
	if kernelSupportsRetprobeMaxActiveOverride(0x0004000e00000000) {
		t.Fatal("kernel 4.14 should not support retprobe maxactive override")
	}
	if !kernelSupportsRetprobeMaxActiveOverride(0x0004000f00000000) {
		t.Fatal("kernel 4.15 should support retprobe maxactive override")
	}
}

func TestParseKernelRelease(t *testing.T) {
	got, err := parseKernelRelease("4.15.0-212-generic")
	if err != nil {
		t.Fatalf("parseKernelRelease: %v", err)
	}
	if got != 0x0004000f00000000 {
		t.Fatalf("unexpected kernel version: got %#x", got)
	}
}
