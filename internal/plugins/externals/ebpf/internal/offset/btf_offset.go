//go:build linux
// +build linux

package offset

// #include "../c/offset_guess/offset.h"
import "C"

import (
	"debug/dwarf"
	"debug/elf"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cilium/ebpf/btf"
	"golang.org/x/sys/unix"

	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
)

const KernelBTFPathEnv = "DK_EBPF_BTF_PATH"
const DisableKernelBTFEnv = "DK_EBPF_DISABLE_BTF_OFFSETS"

type BTFSource struct {
	Path   string
	Method string
}

type btfPathStep struct {
	Name  string
	Index *int
}

func field(name string) btfPathStep {
	return btfPathStep{Name: name}
}

func index(idx int) btfPathStep {
	return btfPathStep{Index: &idx}
}

func (s BTFSource) String() string {
	if s.Method == "" {
		return s.Path
	}
	return fmt.Sprintf("%s (%s)", s.Path, s.Method)
}

func kernelBTFDisabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(DisableKernelBTFEnv)))
	switch v {
	case "", "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func FindKernelBTF() (*BTFSource, error) {
	if kernelBTFDisabled() {
		return nil, fmt.Errorf("kernel BTF disabled by %s", DisableKernelBTFEnv)
	}

	if path := strings.TrimSpace(os.Getenv(KernelBTFPathEnv)); path != "" {
		if _, err := os.Stat(path); err != nil {
			return nil, fmt.Errorf("stat BTF path from %s failed: %w", KernelBTFPathEnv, err)
		}
		return &BTFSource{
			Path:   path,
			Method: "env",
		}, nil
	}

	for _, path := range kernelBTFCandidatePaths() {
		if _, err := os.Stat(path); err == nil {
			method := "file"
			if path == "/sys/kernel/btf/vmlinux" {
				method = "sysfs"
			}
			return &BTFSource{
				Path:   path,
				Method: method,
			}, nil
		}
	}

	return nil, fmt.Errorf("kernel BTF not found; set %s or install matching vmlinux/BTF for %s",
		KernelBTFPathEnv, kernelRelease())
}

func LoadKernelBTF() (*btf.Spec, *BTFSource, error) {
	if kernelBTFDisabled() {
		return nil, nil, fmt.Errorf("kernel BTF disabled by %s", DisableKernelBTFEnv)
	}

	if path := strings.TrimSpace(os.Getenv(KernelBTFPathEnv)); path != "" {
		spec, err := btf.LoadSpec(path)
		if err != nil {
			return nil, nil, fmt.Errorf("load kernel BTF from %s=%s failed: %w", KernelBTFPathEnv, path, err)
		}
		return spec, &BTFSource{
			Path:   path,
			Method: "env",
		}, nil
	}

	var firstErr error
	for _, path := range kernelBTFCandidatePaths() {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		spec, err := btf.LoadSpec(path)
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("load kernel BTF from %s failed: %w", path, err)
			}
			continue
		}
		method := "file"
		if path == "/sys/kernel/btf/vmlinux" {
			method = "sysfs"
		}
		return spec, &BTFSource{
			Path:   path,
			Method: method,
		}, nil
	}

	if firstErr != nil {
		return nil, nil, firstErr
	}

	return nil, nil, fmt.Errorf("kernel BTF not found; set %s or install matching vmlinux/BTF for %s",
		KernelBTFPathEnv, kernelRelease())
}

func GetTCPOffsetFromBTF() (map[string]uint64, error) {
	spec, source, err := LoadKernelBTF()
	if err != nil {
		return nil, err
	}

	l.Debugf("load kernel offsets from BTF: %s", source)

	return getKernelOffsetsFromSpec(spec)
}

func GetKernelOffsetGuessFromBTF() (*OffsetGuessC, *BTFSource, error) {
	spec, source, err := LoadKernelBTF()
	if err != nil {
		return nil, nil, err
	}

	offsets, missing, err := getKernelOffsetsFromSpecPartial(spec)
	if err != nil {
		return nil, nil, err
	}
	if len(missing) > 0 {
		l.Warnf("kernel BTF %s missing offset fields: %s",
			source, strings.Join(missing, ", "))
	}

	return ApplyBTFOffsetToGuess(offsets), source, nil
}

func getKernelOffsetsFromSpec(spec *btf.Spec) (map[string]uint64, error) {
	offsets, missing, err := getKernelOffsetsFromSpecPartial(spec)
	if err != nil {
		return nil, err
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("resolve kernel offsets from BTF failed: missing %s",
			strings.Join(missing, ", "))
	}
	return offsets, nil
}

func getKernelOffsetsFromSpecPartial(spec *btf.Spec) (map[string]uint64, []string, error) {
	fields := []struct {
		name     string
		typeName string
		path     []btfPathStep
	}{
		{name: "tcp_sk_srtt_us", typeName: "tcp_sock", path: []btfPathStep{field("srtt_us")}},
		{name: "tcp_sk_mdev_us", typeName: "tcp_sock", path: []btfPathStep{field("mdev_us")}},
		{name: "sk_num", typeName: "sock_common", path: []btfPathStep{field("skc_num")}},
		{name: "inet_sport", typeName: "inet_sock", path: []btfPathStep{field("inet_sport")}},
		{name: "sk_family", typeName: "sock_common", path: []btfPathStep{field("skc_family")}},
		{name: "sk_rcv_saddr", typeName: "sock_common", path: []btfPathStep{field("skc_rcv_saddr")}},
		{name: "sk_daddr", typeName: "sock_common", path: []btfPathStep{field("skc_daddr")}},
		{name: "sk_v6_rcv_saddr", typeName: "sock_common", path: []btfPathStep{field("skc_v6_rcv_saddr")}},
		{name: "sk_v6_daddr", typeName: "sock_common", path: []btfPathStep{field("skc_v6_daddr")}},
		{name: "sk_dport", typeName: "sock_common", path: []btfPathStep{field("skc_dport")}},
		{name: "flowi4_saddr", typeName: "flowi4", path: []btfPathStep{field("saddr")}},
		{name: "flowi4_daddr", typeName: "flowi4", path: []btfPathStep{field("daddr")}},
		{name: "flowi4_sport", typeName: "flowi4", path: []btfPathStep{field("uli"), field("ports"), field("sport")}},
		{name: "flowi4_dport", typeName: "flowi4", path: []btfPathStep{field("uli"), field("ports"), field("dport")}},
		{name: "flowi6_saddr", typeName: "flowi6", path: []btfPathStep{field("saddr")}},
		{name: "flowi6_daddr", typeName: "flowi6", path: []btfPathStep{field("daddr")}},
		{name: "flowi6_sport", typeName: "flowi6", path: []btfPathStep{field("uli"), field("ports"), field("sport")}},
		{name: "flowi6_dport", typeName: "flowi6", path: []btfPathStep{field("uli"), field("ports"), field("dport")}},
		{name: "sk_net", typeName: "sock_common", path: []btfPathStep{field("skc_net")}},
		{name: "ns_common_inum", typeName: "ns_common", path: []btfPathStep{field("inum")}},
		{name: "socket_sk", typeName: "socket", path: []btfPathStep{field("sk")}},
	}

	result := make(map[string]uint64, len(fields))
	missing := make([]string, 0)
	for _, f := range fields {
		val, err := getPathOffset(spec, f.typeName, f.path...)
		if err != nil {
			missing = append(missing, f.name)
			continue
		}
		result[f.name] = val
	}

	if len(result) == 0 {
		sort.Strings(missing)
		return nil, nil, fmt.Errorf("resolve kernel offsets from BTF failed: missing %s",
			strings.Join(missing, ", "))
	}

	sort.Strings(missing)
	return result, missing, nil
}

func GetHTTPFlowOffsetFromBTF() (*OffsetHTTPFlowC, *BTFSource, error) {
	spec, source, err := LoadKernelBTF()
	if err != nil {
		return nil, nil, err
	}

	taskFiles, err := getPathOffset(spec, "task_struct", field("files"))
	if err != nil {
		return nil, nil, fmt.Errorf("resolve task_struct.files from BTF failed: %w", err)
	}

	filesFDT, err := getPathOffset(spec, "files_struct", field("fdt"))
	if err != nil {
		return nil, nil, fmt.Errorf("resolve files_struct.fdt from BTF failed: %w", err)
	}

	socketFile, err := getPathOffset(spec, "socket", field("file"))
	if err != nil {
		return nil, nil, fmt.Errorf("resolve socket.file from BTF failed: %w", err)
	}

	filePrivateData, err := getPathOffset(spec, "file", field("private_data"))
	if err != nil {
		return nil, nil, fmt.Errorf("resolve file.private_data from BTF failed: %w", err)
	}

	return &OffsetHTTPFlowC{
		offset_task_struct_files: C.__s32(taskFiles),
		offset_files_struct_fdt:  C.__s32(filesFDT),
		offset_socket_file:       C.__s32(socketFile),
		offset_file_private_data: C.__s32(filePrivateData),
	}, source, nil
}

func GetConntrackOffsetFromBTF() (*OffsetConntrackC, *BTFSource, error) {
	spec, source, err := LoadKernelBTF()
	if err != nil {
		return nil, nil, err
	}

	originTuple, err := getPathOffset(spec, "nf_conn", field("tuplehash"), index(0), field("tuple"))
	if err != nil {
		return nil, nil, fmt.Errorf("resolve nf_conn.tuplehash[0].tuple from BTF failed: %w", err)
	}

	replyTuple, err := getPathOffset(spec, "nf_conn", field("tuplehash"), index(1), field("tuple"))
	if err != nil {
		return nil, nil, fmt.Errorf("resolve nf_conn.tuplehash[1].tuple from BTF failed: %w", err)
	}

	ctNet, err := getPathOffset(spec, "nf_conn", field("ct_net"))
	if err != nil {
		return nil, nil, fmt.Errorf("resolve nf_conn.ct_net from BTF failed: %w", err)
	}

	nsCommonInum, err := getPathOffset(spec, "ns_common", field("inum"))
	if err != nil {
		return nil, nil, fmt.Errorf("resolve ns_common.inum from BTF failed: %w", err)
	}

	return &OffsetConntrackC{
		offset_ct_origin_tuple:   C.__u64(originTuple),
		offset_ct_reply_tuple:    C.__u64(replyTuple),
		offset_ct_net:            C.__u64(ctNet),
		offset_ct_ns_common_inum: C.__u64(nsCommonInum),
	}, source, nil
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
		offset_flowi4_saddr:    C.__u64(offsets["flowi4_saddr"]),
		offset_flowi4_daddr:    C.__u64(offsets["flowi4_daddr"]),
		offset_flowi4_sport:    C.__u64(offsets["flowi4_sport"]),
		offset_flowi4_dport:    C.__u64(offsets["flowi4_dport"]),
		offset_flowi6_saddr:    C.__u64(offsets["flowi6_saddr"]),
		offset_flowi6_daddr:    C.__u64(offsets["flowi6_daddr"]),
		offset_flowi6_sport:    C.__u64(offsets["flowi6_sport"]),
		offset_flowi6_dport:    C.__u64(offsets["flowi6_dport"]),
		offset_sk_net:          C.__u64(offsets["sk_net"]),
		offset_ns_common_inum:  C.__u64(offsets["ns_common_inum"]),
		offset_socket_sk:       C.__u64(offsets["socket_sk"]),
	}
}

func TryGetTCPSeqOffsetFromBTF() ([]bpfutil.ConstantPatch, *OffsetTCPSeqC, error) {
	spec, source, err := LoadKernelBTF()
	if err == nil {
		offset, err := getTCPSeqOffsetFromSpec(spec)
		if err != nil {
			return nil, nil, err
		}

		l.Infof("TCP seq offsets obtained from BTF %s: copied_seq=%d, write_seq=%d",
			source, offset.offset_copied_seq, offset.offset_write_seq)

		return NewConstEditorTCPSeq(offset), offset, nil
	}

	dw, dwSource, dwErr := LoadKernelDWARF()
	if dwErr != nil {
		return nil, nil, fmt.Errorf(
			"load tcp seq offsets from BTF failed: %w; load DWARF fallback failed: %v",
			err,
			dwErr,
		)
	}

	offset, err := getTCPSeqOffsetFromDWARF(dw)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve tcp seq offsets from DWARF %s failed: %w", dwSource, err)
	}

	l.Infof("TCP seq offsets obtained from DWARF %s: copied_seq=%d, write_seq=%d",
		dwSource, offset.offset_copied_seq, offset.offset_write_seq)

	return NewConstEditorTCPSeq(offset), offset, nil
}

func getTCPSeqOffsetFromSpec(spec *btf.Spec) (*OffsetTCPSeqC, error) {
	copiedSeq, err := getPathOffset(spec, "tcp_sock", field("copied_seq"))
	if err != nil {
		return nil, fmt.Errorf("resolve tcp_sock.copied_seq from BTF failed: %w", err)
	}

	writeSeq, err := getPathOffset(spec, "tcp_sock", field("write_seq"))
	if err != nil {
		return nil, fmt.Errorf("resolve tcp_sock.write_seq from BTF failed: %w", err)
	}

	offset := &OffsetTCPSeqC{
		offset_copied_seq: C.__s32(copiedSeq),
		offset_write_seq:  C.__s32(writeSeq),
	}

	return offset, nil
}

func LoadKernelDWARF() (*dwarf.Data, *BTFSource, error) {
	if kernelBTFDisabled() {
		return nil, nil, fmt.Errorf("kernel BTF disabled by %s", DisableKernelBTFEnv)
	}

	if path := strings.TrimSpace(os.Getenv(KernelBTFPathEnv)); path != "" {
		dw, err := loadDWARFFromELF(path)
		if err != nil {
			return nil, nil, fmt.Errorf("load kernel DWARF from %s=%s failed: %w", KernelBTFPathEnv, path, err)
		}
		return dw, &BTFSource{
			Path:   path,
			Method: "env-dwarf",
		}, nil
	}

	var firstErr error
	for _, path := range kernelBTFCandidatePaths() {
		if _, err := os.Stat(path); err != nil {
			continue
		}

		dw, err := loadDWARFFromELF(path)
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("load kernel DWARF from %s failed: %w", path, err)
			}
			continue
		}

		return dw, &BTFSource{
			Path:   path,
			Method: "dwarf",
		}, nil
	}

	if firstErr != nil {
		return nil, nil, firstErr
	}

	return nil, nil, fmt.Errorf("kernel DWARF not found; set %s to matching vmlinux/debug image for %s",
		KernelBTFPathEnv, kernelRelease())
}

func loadDWARFFromELF(path string) (*dwarf.Data, error) {
	f, err := elf.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck

	return f.DWARF()
}

func getTCPSeqOffsetFromDWARF(data *dwarf.Data) (*OffsetTCPSeqC, error) {
	copiedSeq, err := getStructFieldOffsetFromDWARF(data, "tcp_sock", "copied_seq")
	if err != nil {
		return nil, fmt.Errorf("resolve tcp_sock.copied_seq from DWARF failed: %w", err)
	}

	writeSeq, err := getStructFieldOffsetFromDWARF(data, "tcp_sock", "write_seq")
	if err != nil {
		return nil, fmt.Errorf("resolve tcp_sock.write_seq from DWARF failed: %w", err)
	}

	return &OffsetTCPSeqC{
		offset_copied_seq: C.__s32(copiedSeq),
		offset_write_seq:  C.__s32(writeSeq),
	}, nil
}

func getStructFieldOffsetFromDWARF(data *dwarf.Data, structName, fieldName string) (uint64, error) {
	r := data.Reader()
	for {
		ent, err := r.Next()
		if err != nil {
			return 0, err
		}
		if ent == nil {
			break
		}

		if ent.Tag == dwarf.TagStructType {
			if name, _ := ent.Val(dwarf.AttrName).(string); name == structName {
				return getStructFieldOffsetFromDWARFEntry(r, ent, fieldName)
			}
		}

		if ent.Children {
			r.SkipChildren()
		}
	}

	return 0, fmt.Errorf("struct %q not found", structName)
}

func getStructFieldOffsetFromDWARFEntry(r *dwarf.Reader, ent *dwarf.Entry, fieldName string) (uint64, error) {
	if !ent.Children {
		return 0, fmt.Errorf("struct has no children")
	}

	for {
		child, err := r.Next()
		if err != nil {
			return 0, err
		}
		if child == nil || child.Tag == 0 {
			break
		}

		if child.Tag == dwarf.TagMember {
			if name, _ := child.Val(dwarf.AttrName).(string); name == fieldName {
				return dwarfFieldOffset(child)
			}
		}

		if child.Children {
			r.SkipChildren()
		}
	}

	return 0, fmt.Errorf("field %q not found", fieldName)
}

func dwarfFieldOffset(ent *dwarf.Entry) (uint64, error) {
	loc := ent.Val(dwarf.AttrDataMemberLoc)
	switch v := loc.(type) {
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("negative offset %d", v)
		}
		return uint64(v), nil
	case uint64:
		return v, nil
	default:
		return 0, fmt.Errorf("unsupported DWARF member location type %T", loc)
	}
}

func kernelBTFCandidatePaths() []string {
	return kernelBTFCandidatePathsForRelease(kernelRelease())
}

func kernelBTFCandidatePathsForRelease(release string) []string {
	paths := []string{"/sys/kernel/btf/vmlinux"}
	if release == "" {
		return paths
	}

	candidates := []string{
		filepath.Join("/boot", "vmlinux-"+release),
		filepath.Join("/usr/lib/modules", release, "vmlinux"),
		filepath.Join("/lib/modules", release, "vmlinux"),
		filepath.Join("/usr/lib/debug/boot", "vmlinux-"+release),
		filepath.Join("/usr/lib/debug/boot", "vmlinux-"+release+".debug"),
		filepath.Join("/usr/lib/debug/lib/modules", release, "vmlinux"),
		filepath.Join("/usr/lib/debug/lib/modules", release, "vmlinux.debug"),
	}

	return append(paths, candidates...)
}

func kernelRelease() string {
	var uname unix.Utsname
	if err := unix.Uname(&uname); err != nil {
		return ""
	}

	var buf []byte
	for _, c := range uname.Release {
		if c == 0 {
			break
		}
		buf = append(buf, c)
	}
	return string(buf)
}

func getPathOffset(spec *btf.Spec, typeName string, steps ...btfPathStep) (uint64, error) {
	typ, err := namedBTFType(spec, typeName)
	if err != nil {
		return 0, err
	}

	var total uint64
	current := typ
	for _, step := range steps {
		if step.Name != "" {
			offset, next, err := findMemberOffset(current, step.Name)
			if err != nil {
				return 0, err
			}
			total += offset
			current = next
		}

		if step.Index != nil {
			array, ok := btf.UnderlyingType(current).(*btf.Array)
			if !ok {
				return 0, fmt.Errorf("type %s is not an array", btfTypeName(current))
			}
			if *step.Index < 0 || *step.Index >= int(array.Nelems) {
				return 0, fmt.Errorf("array index %d out of range", *step.Index)
			}
			elemSize, err := btf.Sizeof(array.Type)
			if err != nil {
				return 0, fmt.Errorf("sizeof array element failed: %w", err)
			}
			total += uint64(*step.Index * elemSize)
			current = array.Type
		}
	}

	return total, nil
}

func namedBTFType(spec *btf.Spec, typeName string) (btf.Type, error) {
	var st *btf.Struct
	if err := spec.TypeByName(typeName, &st); err == nil && st != nil {
		return st, nil
	}

	var un *btf.Union
	if err := spec.TypeByName(typeName, &un); err == nil && un != nil {
		return un, nil
	}

	var td *btf.Typedef
	if err := spec.TypeByName(typeName, &td); err == nil && td != nil {
		return td, nil
	}

	return nil, fmt.Errorf("type %s not found in BTF", typeName)
}

func findMemberOffset(typ btf.Type, memberName string) (uint64, btf.Type, error) {
	switch current := btf.UnderlyingType(typ).(type) {
	case *btf.Struct:
		return findMemberOffsetInMembers(current.Members, memberName)
	case *btf.Union:
		return findMemberOffsetInMembers(current.Members, memberName)
	default:
		return 0, nil, fmt.Errorf("type %s does not contain members", btfTypeName(typ))
	}
}

func findMemberOffsetInMembers(members []btf.Member, memberName string) (uint64, btf.Type, error) {
	for _, member := range members {
		if member.Name == memberName {
			return uint64(member.Offset.Bytes()), member.Type, nil
		}
	}

	for _, member := range members {
		if !isAnonymousBTFMember(member) {
			continue
		}
		offset, typ, err := findMemberOffset(member.Type, memberName)
		if err == nil {
			return uint64(member.Offset.Bytes()) + offset, typ, nil
		}
	}

	return 0, nil, fmt.Errorf("member %s not found", memberName)
}

func isAnonymousBTFMember(member btf.Member) bool {
	return member.Name == "" || strings.HasPrefix(member.Name, "(")
}

func btfTypeName(typ btf.Type) string {
	if typ == nil {
		return "<nil>"
	}
	if name := typ.TypeName(); name != "" {
		return name
	}
	return fmt.Sprintf("%T", typ)
}
