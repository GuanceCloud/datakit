//go:build linux
// +build linux

package bpfutil

import (
	"fmt"
	"math"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/asm"
)

const legacyConstMagic uint64 = 0x54A17EAD00000000

var legacyConstSentinels = map[string]uint64{
	"kernel_version":           legacyConstMagic | 0x001,
	"offset_sk_num":            legacyConstMagic | 0x002,
	"offset_inet_sport":        legacyConstMagic | 0x003,
	"offset_sk_family":         legacyConstMagic | 0x004,
	"offset_sk_rcv_saddr":      legacyConstMagic | 0x005,
	"offset_sk_daddr":          legacyConstMagic | 0x006,
	"offset_sk_v6_rcv_saddr":   legacyConstMagic | 0x007,
	"offset_sk_v6_daddr":       legacyConstMagic | 0x008,
	"offset_sk_dport":          legacyConstMagic | 0x009,
	"offset_tcp_sk_srtt_us":    legacyConstMagic | 0x00A,
	"offset_tcp_sk_mdev_us":    legacyConstMagic | 0x00B,
	"offset_flowi4_saddr":      legacyConstMagic | 0x00C,
	"offset_flowi4_daddr":      legacyConstMagic | 0x00D,
	"offset_flowi4_sport":      legacyConstMagic | 0x00E,
	"offset_flowi4_dport":      legacyConstMagic | 0x00F,
	"offset_flowi6_saddr":      legacyConstMagic | 0x010,
	"offset_flowi6_daddr":      legacyConstMagic | 0x011,
	"offset_flowi6_sport":      legacyConstMagic | 0x012,
	"offset_flowi6_dport":      legacyConstMagic | 0x013,
	"offset_sk_net":            legacyConstMagic | 0x014,
	"offset_ns_common_inum":    legacyConstMagic | 0x015,
	"offset_socket_sk":         legacyConstMagic | 0x016,
	"offset_socket_file":       legacyConstMagic | 0x017,
	"offset_task_struct_files": legacyConstMagic | 0x018,
	"offset_files_struct_fdt":  legacyConstMagic | 0x019,
	"offset_file_private_data": legacyConstMagic | 0x01A,
	"offset_copied_seq":        legacyConstMagic | 0x01B,
	"offset_write_seq":         legacyConstMagic | 0x01C,
	"offset_ct_net":            legacyConstMagic | 0x01D,
	"offset_ct_ns_common_inum": legacyConstMagic | 0x01E,
	"offset_ct_origin_tuple":   legacyConstMagic | 0x01F,
	"offset_ct_reply_tuple":    legacyConstMagic | 0x020,
	"apiflow_min_capture_size": legacyConstMagic | 0x021,
}

func rewriteLegacyConstants(spec *ebpf.CollectionSpec, constants []Constant) error {
	if len(constants) == 0 {
		return nil
	}

	replacements := make(map[int64]int64, len(constants))
	for _, constant := range constants {
		sentinel, ok := legacyConstSentinels[constant.Name]
		if !ok {
			return fmt.Errorf("legacy constant %q is not declared", constant.Name)
		}

		value, err := legacyConstantValue(constant.Value)
		if err != nil {
			return fmt.Errorf("legacy constant %q: %w", constant.Name, err)
		}
		replacements[int64(sentinel)] = int64(value)
	}

	for progName, prog := range spec.Programs {
		if prog == nil {
			continue
		}
		if err := rewriteLegacyInstructions(prog.Instructions, replacements); err != nil {
			return fmt.Errorf("rewrite legacy constants for program %q: %w", progName, err)
		}
	}
	return nil
}

func rewriteLegacyInstructions(insns asm.Instructions, replacements map[int64]int64) error {
	for idx := range insns {
		ins := &insns[idx]
		if !ins.IsConstantLoad(asm.DWord) {
			continue
		}
		if value, ok := replacements[ins.Constant]; ok {
			ins.Constant = value
		}
	}
	return nil
}

func legacyConstantValue(v interface{}) (uint64, error) {
	switch value := v.(type) {
	case uint64:
		return value, nil
	case uint32:
		return uint64(value), nil
	case uint16:
		return uint64(value), nil
	case uint8:
		return uint64(value), nil
	case uint:
		return uint64(value), nil
	case int64:
		if value < 0 {
			return 0, fmt.Errorf("negative values are unsupported")
		}
		return uint64(value), nil
	case int32:
		if value < 0 {
			return 0, fmt.Errorf("negative values are unsupported")
		}
		return uint64(value), nil
	case int:
		if value < 0 {
			return 0, fmt.Errorf("negative values are unsupported")
		}
		return uint64(value), nil
	default:
		return 0, fmt.Errorf("unsupported value type %T", v)
	}
}

func kernelSupportsReadOnlyMaps(kernelVersion uint64) bool {
	const minReadOnlyMapKernel = uint64(0x0005000200000000)
	return kernelVersion >= minReadOnlyMapKernel
}

func shouldUseLegacyConstants(kernelVersion uint64) bool {
	return kernelVersion != 0 && !kernelSupportsReadOnlyMaps(kernelVersion)
}

func UseLegacyConstObjects() (bool, uint64, error) {
	kernelVersion, err := CurrentKernelVersion()
	if err != nil {
		return false, 0, err
	}
	return shouldUseLegacyConstants(kernelVersion), kernelVersion, nil
}

func sentinelConstant(name string) (int64, error) {
	sentinel, ok := legacyConstSentinels[name]
	if !ok {
		return 0, fmt.Errorf("unknown legacy constant %q", name)
	}
	if sentinel > math.MaxInt64 {
		return 0, fmt.Errorf("legacy sentinel %q exceeds int64", name)
	}
	return int64(sentinel), nil
}
