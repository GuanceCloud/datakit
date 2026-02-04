//go:build linux
// +build linux

package offset

// #include "../c/offset_guess/offset.h"
import "C"

import (
	"fmt"

	manager "github.com/DataDog/ebpf-manager"
	"github.com/cilium/ebpf/btf"
)

func GetTCPOffsetFromBTF() (map[string]uint64, error) {
	spec, err := btf.LoadKernelSpec()
	if err != nil {
		return nil, fmt.Errorf("load kernel BTF failed: %w", err)
	}

	result := make(map[string]uint64)

	fields := []struct {
		name   string
		typ    string
		member string
	}{
		{"tcp_sk_srtt_us", "tcp_sock", "srtt_us"},
		{"tcp_sk_mdev_us", "tcp_sock", "mdev_us"},
		{"sk_num", "sock_common", "skc_num"},
		{"inet_sport", "inet_sock", "inet_sport"},
		{"sk_family", "sock_common", "skc_family"},
		{"sk_rcv_saddr", "sock_common", "skc_rcv_saddr"},
		{"sk_daddr", "sock_common", "skc_daddr"},
		{"sk_v6_rcv_saddr", "sock_common", "skc_v6_rcv_saddr"},
		{"sk_v6_daddr", "sock_common", "skc_v6_daddr"},
		{"sk_dport", "sock_common", "skc_dport"},
		{"sk_net", "sock_common", "skc_net"},
		{"ns_common_inum", "ns_common", "inum"},
		{"socket_sk", "socket_alloc", "sock"},
	}

	for _, f := range fields {
		val, err := getMemberOffset(spec, f.typ, f.member)
		if err != nil {
			l.Debugf("get %s.%s offset failed: %v", f.typ, f.member, err)
			continue
		}
		result[f.name] = val
	}

	return result, nil
}

func getMemberOffset(spec *btf.Spec, typeName, memberName string) (uint64, error) {
	var s *btf.Struct
	if err := spec.TypeByName(typeName, &s); err != nil {
		return 0, fmt.Errorf("type %s not found: %w", typeName, err)
	}

	if s == nil {
		return 0, fmt.Errorf("%s is not a struct", typeName)
	}

	for _, m := range s.Members {
		if m.Name == memberName {
			return uint64(m.Offset.Bytes()), nil
		}
	}

	return 0, fmt.Errorf("member %s not found", memberName)
}

func ApplyBTFOffsetToGuess(offsets map[string]uint64) *OffsetGuessC {
	return &OffsetGuessC{
		offset_tcp_sk_srtt_us:  C.__u64(offsets["tcp_sk_srtt_us"]),
		offset_tcp_sk_mdev_us:  C.__u64(offsets["tcp_sk_mdev_us"]),
		offset_sk_num:          C.__u64(offsets["sk_num"]),
		offset_inet_sport:      C.__u64(offsets["inet_sport"]),
		offset_sk_family:       C.__u64(offsets["sk_family"]),
		offset_sk_rcv_saddr:    C.__u64(offsets["sk_rcv_saddr"]),
		offset_sk_daddr:        C.__u64(offsets["sk_daddr"]),
		offset_sk_v6_rcv_saddr: C.__u64(offsets["sk_v6_rcv_saddr"]),
		offset_sk_v6_daddr:     C.__u64(offsets["sk_v6_daddr"]),
		offset_sk_dport:        C.__u64(offsets["sk_dport"]),
		offset_sk_net:          C.__u64(offsets["sk_net"]),
		offset_ns_common_inum:  C.__u64(offsets["ns_common_inum"]),
		offset_socket_sk:       C.__u64(offsets["socket_sk"]),
	}
}

func TryGetTCPSeqOffsetFromBTF() ([]manager.ConstantEditor, *OffsetTCPSeqC, error) {
	spec, err := btf.LoadKernelSpec()
	if err != nil {
		return nil, nil, fmt.Errorf("load kernel BTF failed: %w", err)
	}

	offset := OffsetTCPSeqC{}

	copiedSeq, err := getMemberOffset(spec, "tcp_sock", "copied_seq")
	if err != nil {
		l.Debugf("get tcp_sock.copied_seq offset failed: %v", err)
		return nil, nil, err
	}

	writeSeq, err := getMemberOffset(spec, "tcp_sock", "write_seq")
	if err != nil {
		l.Debugf("get tcp_sock.write_seq offset failed: %v", err)
		return nil, nil, err
	}

	offset.offset_copied_seq = C.__s32(copiedSeq)
	offset.offset_write_seq = C.__s32(writeSeq)

	l.Infof("TCP seq offsets obtained from BTF: copied_seq=%d, write_seq=%d", copiedSeq, writeSeq)

	seqConstEditor := NewConstEditorTCPSeq(&offset)
	return seqConstEditor, &offset, nil
}
