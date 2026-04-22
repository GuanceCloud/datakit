//go:build linux
// +build linux

package offset

import (
	"fmt"

	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
)

type KernelOffsetPlan struct {
	Guess   *OffsetGuessC
	Patches []bpfutil.ConstantPatch
}

type ConntrackOffsetPlan struct {
	Guess   *OffsetConntrackC
	Patches []bpfutil.ConstantPatch
}

type HTTPFlowOffsetPlan struct {
	Patches []bpfutil.ConstantPatch
}

func ResolveKernelOffsets(saved *OffsetGuessC, wantTCPSeq bool, ipv6Disabled bool) (*KernelOffsetPlan, error) {
	merged := newGuessStatus()
	kernelBTFReady := false
	seqBTFReady := !wantTCPSeq
	if saved != nil {
		copyOffset(saved, &merged)
		copySupplementalOffsets(saved, &merged)
		SetTCPSeqOffset(&merged, GetTCPSeqOffset(saved))
	}

	if kernelOffsetsReady(&merged, wantTCPSeq, ipv6Disabled) {
		l.Infof("resolved kernel offsets from cached guess")
		patches := NewConstEditor(&merged)
		if wantTCPSeq {
			patches = append(patches, NewConstEditorTCPSeq(GetTCPSeqOffset(&merged))...)
		}
		return &KernelOffsetPlan{
			Guess:   &merged,
			Patches: patches,
		}, nil
	}

	if guess, source, err := GetKernelOffsetGuessFromBTF(); err == nil {
		mergeKernelOffsets(&merged, guess)
		kernelBTFReady = true
		l.Infof("resolved kernel offsets from BTF: %s", source)
	} else {
		l.Debugf("load kernel offsets from BTF failed: %v", err)
	}

	if wantTCPSeq {
		if _, seq, err := TryGetTCPSeqOffsetFromBTF(); err == nil {
			SetTCPSeqOffset(&merged, seq)
			seqBTFReady = true
		} else {
			l.Debugf("load TCP seq offsets from BTF failed: %v", err)
		}
	}

	if kernelBTFReady && seqBTFReady && kernelOffsetsReady(&merged, wantTCPSeq, ipv6Disabled) {
		patches := NewConstEditor(&merged)
		if wantTCPSeq {
			patches = append(patches, NewConstEditorTCPSeq(GetTCPSeqOffset(&merged))...)
		}
		return &KernelOffsetPlan{
			Guess:   &merged,
			Patches: patches,
		}, nil
	}

	runtime, err := NewGuessRuntime(&merged, ipv6Disabled)
	if err != nil {
		return nil, fmt.Errorf("new offset runtime: %w", err)
	}
	if err := runtime.StartRuntime(); err != nil {
		return nil, fmt.Errorf("start offset runtime: %w", err)
	}
	defer runtime.Shutdown() //nolint:errcheck

	guess, err := GuessOffset(runtime, &merged, ipv6Disabled)
	if err != nil {
		return nil, err
	}

	patches := NewConstEditor(guess)
	if wantTCPSeq {
		_, seqOffset, err := GuessOffsetTCPSeq(patches)
		if err != nil {
			return nil, err
		}
		SetTCPSeqOffset(guess, seqOffset)
		patches = append(patches, NewConstEditorTCPSeq(seqOffset)...)
	}

	return &KernelOffsetPlan{
		Guess:   guess,
		Patches: patches,
	}, nil
}

func ResolveConntrackOffsets(saved *OffsetConntrackC) (*ConntrackOffsetPlan, error) {
	merged := newGuessConntrack()
	if saved != nil {
		copyOffsetCT(saved, &merged)
	}
	seedConntrackTupleOffsets(&merged)

	if conntrackOffsetsReady(&merged) {
		l.Infof("resolved conntrack offsets from cached guess")
		return &ConntrackOffsetPlan{
			Guess:   &merged,
			Patches: newConntrackConstEditor(&merged),
		}, nil
	}

	if btfOffset, source, err := GetConntrackOffsetFromBTF(); err == nil {
		mergeConntrackOffsets(&merged, btfOffset)
		l.Infof("resolved conntrack offsets from BTF: %s", source)
		if conntrackOffsetsReady(&merged) {
			return &ConntrackOffsetPlan{
				Guess:   &merged,
				Patches: newConntrackConstEditor(&merged),
			}, nil
		}
	} else {
		l.Debugf("load conntrack offsets from BTF failed: %v", err)
	}

	patches, guess, err := GuessOffsetConntrack(&merged)
	if err != nil {
		if conntrackTupleOffsetsReady(&merged) {
			l.Warnf("using conntrack tuple-only fallback: %v", err)
			return &ConntrackOffsetPlan{
				Guess:   &merged,
				Patches: newConntrackConstEditor(&merged),
			}, nil
		}
		return nil, err
	}
	return &ConntrackOffsetPlan{
		Guess:   guess,
		Patches: patches,
	}, nil
}

func ResolveHTTPFlowOffsets(status *OffsetGuessC) (*HTTPFlowOffsetPlan, error) {
	if httpFlowOffsetsReady(status) {
		l.Infof("resolved HTTP flow offsets from cached guess")
		return &HTTPFlowOffsetPlan{Patches: HTTPFlowPatchesFromGuess(status)}, nil
	}

	if btfOffset, source, err := GetHTTPFlowOffsetFromBTF(); err == nil {
		l.Infof("resolved HTTP flow offsets from BTF: %s", source)
		return &HTTPFlowOffsetPlan{Patches: NewConstHTTPEditor(btfOffset)}, nil
	} else {
		l.Debugf("load HTTP flow offsets from BTF failed: %v", err)
	}

	patches, err := GuessOffsetHTTPFlow(status)
	if err != nil {
		return nil, err
	}
	return &HTTPFlowOffsetPlan{Patches: patches}, nil
}

func httpFlowOffsetsReady(offset *OffsetGuessC) bool {
	if offset == nil {
		return false
	}

	return offset.offset_task_struct_files != 0 &&
		offset.offset_files_struct_fdt != 0 &&
		offset.offset_socket_file != 0 &&
		offset.offset_file_private_data != 0
}

func ConntrackGuessFromOffset(offset *OffsetGuessC) *OffsetConntrackC {
	if offset == nil {
		return nil
	}

	guess := newGuessConntrack()
	guess.offset_ct_net = offset.offset_ct_net
	guess.offset_ct_origin_tuple = offset.offset_origin_tuple
	guess.offset_ct_reply_tuple = offset.offset_reply_tuple
	guess.offset_ct_ns_common_inum = offset.offset_ct_ns_common_inum
	if guess.offset_ct_net == 0 {
		guess.offset_ct_net = offset.offset_sk_net
	}
	if guess.offset_ct_ns_common_inum == 0 {
		guess.offset_ct_ns_common_inum = offset.offset_ns_common_inum
	}
	seedConntrackTupleOffsets(&guess)

	if guess.offset_ct_net == 0 &&
		guess.offset_ct_origin_tuple == 0 &&
		guess.offset_ct_reply_tuple == 0 &&
		guess.offset_ct_ns_common_inum == 0 {
		return nil
	}

	return &guess
}

func mergeKernelOffsets(dst, src *OffsetGuessC) {
	if dst == nil || src == nil {
		return
	}
	copyOffset(src, dst)
	SetTCPSeqOffset(dst, GetTCPSeqOffset(src))
}

func mergeConntrackOffsets(dst, src *OffsetConntrackC) {
	if dst == nil || src == nil {
		return
	}
	copyOffsetCT(src, dst)
}

func seedConntrackTupleOffsets(dst *OffsetConntrackC) {
	if dst == nil {
		return
	}

	origin, reply := conntrackSeedTupleOffsets()
	if dst.offset_ct_origin_tuple == 0 {
		dst.offset_ct_origin_tuple = _Ctype_ulonglong(origin)
	}
	if dst.offset_ct_reply_tuple == 0 {
		dst.offset_ct_reply_tuple = _Ctype_ulonglong(reply)
	}
}

func kernelOffsetsReady(offset *OffsetGuessC, wantTCPSeq bool, ipv6Disabled bool) bool {
	if offset == nil {
		return false
	}

	required := []uint64{
		uint64(offset.offset_sk_num),
		uint64(offset.offset_sk_family),
		uint64(offset.offset_sk_rcv_saddr),
		uint64(offset.offset_sk_dport),
		uint64(offset.offset_flowi4_saddr),
		uint64(offset.offset_flowi4_daddr),
		uint64(offset.offset_flowi4_sport),
		uint64(offset.offset_flowi4_dport),
	}

	if !ipv6Disabled {
		required = append(required,
			uint64(offset.offset_sk_v6_rcv_saddr),
			uint64(offset.offset_sk_v6_daddr),
			uint64(offset.offset_flowi6_saddr),
			uint64(offset.offset_flowi6_daddr),
			uint64(offset.offset_flowi6_sport),
			uint64(offset.offset_flowi6_dport),
		)
	}

	if wantTCPSeq {
		required = append(required,
			uint64(offset.offset_copied_seq),
			uint64(offset.offset_write_seq),
		)
	}

	for _, value := range required {
		if value == 0 {
			return false
		}
	}

	return true
}

func conntrackOffsetsReady(offset *OffsetConntrackC) bool {
	if offset == nil {
		return false
	}

	return offset.offset_ct_origin_tuple != 0 &&
		offset.offset_ct_reply_tuple != 0 &&
		offset.offset_ct_net != 0 &&
		offset.offset_ct_ns_common_inum != 0
}

func conntrackTupleOffsetsReady(offset *OffsetConntrackC) bool {
	if offset == nil {
		return false
	}

	return offset.offset_ct_origin_tuple != 0 &&
		offset.offset_ct_reply_tuple != 0
}
