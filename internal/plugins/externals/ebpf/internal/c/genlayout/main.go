//go:build linux
// +build linux

package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

type structSpec struct {
	CName  string
	GoName string
	Fields []fieldSpec
}

type fieldSpec struct {
	CName  string
	GoName string
	GoType string
}

type structLayout struct {
	Size   int
	Align  int
	Fields map[string]fieldLayout
}

type fieldLayout struct {
	Offset int
	Size   int
}

type goField struct {
	Name string
	Type string
}

func main() {
	target := flag.String("target", "netflow", "layout target")
	out := flag.String("out", "", "output file")
	flag.Parse()

	if err := run(*target, *out); err != nil {
		fmt.Fprintf(os.Stderr, "genlayout: %v\n", err)
		os.Exit(1)
	}
}

func run(target, out string) error {
	switch target {
	case "bashhistory":
		return generateBashHistory(out)
	case "netflow":
		return generateNetflow(out)
	case "offset":
		return generateOffset(out)
	case "procwatch":
		return generateProcwatch(out)
	case "l7flow":
		return generateL7Flow(out)
	default:
		return fmt.Errorf("unknown target %q", target)
	}
}

func cFields(goType string, names ...string) []fieldSpec {
	fields := make([]fieldSpec, 0, len(names))
	for _, name := range names {
		fields = append(fields, fieldSpec{CName: name, GoName: name, GoType: goType})
	}
	return fields
}

func generateOffset(out string) error {
	root, err := findRepoRoot()
	if err != nil {
		return err
	}

	offsetGuessFields := cFields("uint64",
		"offset_sk_num",
		"offset_inet_sport",
		"offset_sk_family",
		"offset_sk_rcv_saddr",
		"offset_sk_daddr",
		"offset_sk_v6_rcv_saddr",
		"offset_sk_v6_daddr",
		"offset_sk_dport",
		"offset_tcp_sk_srtt_us",
		"offset_tcp_sk_mdev_us",
		"offset_flowi4_saddr",
		"offset_flowi4_daddr",
		"offset_flowi4_sport",
		"offset_flowi4_dport",
		"offset_flowi6_saddr",
		"offset_flowi6_daddr",
		"offset_flowi6_sport",
		"offset_flowi6_dport",
		"offset_skaddr_sin_port",
		"offset_skaddr6_sin6_port",
		"offset_sk_net",
		"offset_ns_common_inum",
		"offset_socket_sk",
		"offset_copied_seq",
		"offset_write_seq",
		"offset_task_struct_files",
		"offset_files_struct_fdt",
		"offset_socket_file",
		"offset_file_private_data",
		"offset_ct_net",
		"offset_ct_ns_common_inum",
		"offset_origin_tuple",
		"offset_reply_tuple",
	)
	offsetGuessFields = append(offsetGuessFields,
		fieldSpec{CName: "process_name", GoName: "process_name", GoType: "[16]uint8"},
		fieldSpec{CName: "err", GoName: "err", GoType: "int64"},
		fieldSpec{CName: "state", GoName: "state", GoType: "uint64"},
		fieldSpec{CName: "pid_tgid", GoName: "pid_tgid", GoType: "uint64"},
		fieldSpec{CName: "conn_type", GoName: "conn_type", GoType: "uint32"},
	)
	offsetGuessFields = append(offsetGuessFields, cFields("uint16",
		"sport",
		"dport",
		"sport_skt",
		"dport_skt",
		"family_skt",
		"_pad1",
	)...)
	offsetGuessFields = append(offsetGuessFields,
		fieldSpec{CName: "saddr", GoName: "saddr", GoType: "[4]uint32"},
		fieldSpec{CName: "daddr", GoName: "daddr", GoType: "[4]uint32"},
		fieldSpec{CName: "daddr_skt", GoName: "daddr_skt", GoType: "[4]uint32"},
	)
	offsetGuessFields = append(offsetGuessFields, cFields("uint32",
		"netns",
		"netns_skt",
		"meta",
		"rtt",
		"rtt_var",
		"_pad",
	)...)

	specs := []structSpec{
		{
			CName:  "offset_guess",
			GoName: "OffsetGuessC",
			Fields: offsetGuessFields,
		},
		{
			CName:  "offset_httpflow",
			GoName: "OffsetHTTPFlowC",
			Fields: []fieldSpec{
				{CName: "process_name", GoName: "process_name", GoType: "[16]uint8"},
				{CName: "pid_tgid", GoName: "pid_tgid", GoType: "uint64"},
				{CName: "offset_task_struct_files", GoName: "offset_task_struct_files", GoType: "int32"},
				{CName: "offset_files_struct_fdt", GoName: "offset_files_struct_fdt", GoType: "int32"},
				{CName: "offset_fdtable_fd", GoName: "offset_fdtable_fd", GoType: "int32"},
				{CName: "offset_socket_file", GoName: "offset_socket_file", GoType: "int32"},
				{CName: "offset_file_private_data", GoName: "offset_file_private_data", GoType: "int32"},
				{CName: "times", GoName: "times", GoType: "int32"},
				{CName: "state", GoName: "state", GoType: "int32"},
				{CName: "fd", GoName: "fd", GoType: "int32"},
				{CName: "sport", GoName: "sport", GoType: "uint16"},
				{CName: "dport", GoName: "dport", GoType: "uint16"},
				{CName: "saddr", GoName: "saddr", GoType: "[4]uint32"},
				{CName: "daddr", GoName: "daddr", GoType: "[4]uint32"},
				{CName: "offset_socket_sk", GoName: "offset_socket_sk", GoType: "int32"},
			},
		},
		{
			CName:  "offset_tcp_seq",
			GoName: "OffsetTCPSeqC",
			Fields: []fieldSpec{
				{CName: "process_name", GoName: "process_name", GoType: "[16]uint8"},
				{CName: "pid_tgid", GoName: "pid_tgid", GoType: "uint64"},
				{CName: "gs_rtt", GoName: "gs_rtt", GoType: "int32"},
				{CName: "offset_copied_seq", GoName: "offset_copied_seq", GoType: "int32"},
				{CName: "offset_write_seq", GoName: "offset_write_seq", GoType: "int32"},
				{CName: "state", GoName: "state", GoType: "int32"},
			},
		},
		{
			CName:  "nf_conn_tuple",
			GoName: "CTConnC",
			Fields: []fieldSpec{
				{CName: "src_ip", GoName: "src_ip", GoType: "[4]uint32"},
				{CName: "dst_ip", GoName: "dst_ip", GoType: "[4]uint32"},
				{CName: "src_port", GoName: "src_port", GoType: "uint16"},
				{CName: "dst_port", GoName: "dst_port", GoType: "uint16"},
				{CName: "l3num", GoName: "l3num", GoType: "uint16"},
				{CName: "l4proto", GoName: "l4proto", GoType: "uint8"},
				{CName: "_pad", GoName: "_pad", GoType: "uint8"},
			},
		},
		{
			CName:  "offset_conntrack",
			GoName: "OffsetConntrackC",
			Fields: []fieldSpec{
				{CName: "process_name", GoName: "process_name", GoType: "[16]uint8"},
				{CName: "err", GoName: "err", GoType: "int64"},
				{CName: "state", GoName: "state", GoType: "uint64"},
				{CName: "pid_tgid", GoName: "pid_tgid", GoType: "uint64"},
				{CName: "offset_ct_origin_tuple", GoName: "offset_ct_origin_tuple", GoType: "uint64"},
				{CName: "offset_ct_reply_tuple", GoName: "offset_ct_reply_tuple", GoType: "uint64"},
				{CName: "offset_ct_net", GoName: "offset_ct_net", GoType: "uint64"},
				{CName: "offset_ct_ns_common_inum", GoName: "offset_ct_ns_common_inum", GoType: "uint64"},
				{CName: "origin", GoName: "origin", GoType: "CTConnC"},
				{CName: "reply", GoName: "reply", GoType: "CTConnC"},
				{CName: "netns", GoName: "netns", GoType: "uint32"},
				{CName: "_pad", GoName: "_pad", GoType: "uint32"},
			},
		},
	}

	header := filepath.Join(root, "internal/plugins/externals/ebpf/internal/c/offset_guess/offset.h")
	layouts, err := readCLayouts(header, specs)
	if err != nil {
		return err
	}

	if out == "" {
		out = filepath.Join(root, "internal/plugins/externals/ebpf/internal/offset/offset_layout_gen.go")
	}

	const aliases = `type _Ctype_uchar = uint8
type _Ctype_ushort = uint16
type _Ctype_uint = uint32
type _Ctype_ulonglong = uint64
type _Ctype_longlong = int64
type _Ctype_int = int32
type _Ctype_struct_nf_conn_tuple = CTConnC
`

	src, err := renderLayouts("offset", "linux", "linux", aliases, layouts, specs)
	if err != nil {
		return err
	}

	return writeGeneratedFile(out, src)
}

func generateProcwatch(out string) error {
	root, err := findRepoRoot()
	if err != nil {
		return err
	}

	specs := []structSpec{
		{
			CName:  "rec_process_sched_status",
			GoName: "schedEvent",
			Fields: []fieldSpec{
				{CName: "status", GoName: "status", GoType: "int32"},
				{CName: "prv_pid", GoName: "prv_pid", GoType: "int32"},
				{CName: "nxt_pid", GoName: "nxt_pid", GoType: "int32"},
				{CName: "__pad", GoName: "__pad", GoType: "int32"},
				{CName: "comm", GoName: "comm", GoType: "[16]uint8"},
			},
		},
		{
			CName:  "proc_filter_info",
			GoName: "procFilterInfo",
			Fields: []fieldSpec{
				{CName: "disable", GoName: "disable", GoType: "uint8"},
				{CName: "pad0", GoName: "pad0", GoType: "uint8"},
				{CName: "pad1", GoName: "pad1", GoType: "uint16"},
				{CName: "pad2", GoName: "pad2", GoType: "uint32"},
			},
		},
		{
			CName:  "proc_inject",
			GoName: "procInjectInfo",
			Fields: []fieldSpec{
				{CName: "offset_go_runtime_g_goid", GoName: "offset_go_runtime_g_goid", GoType: "uint64"},
				{CName: "go_use_register", GoName: "go_use_register", GoType: "uint64"},
			},
		},
	}

	header := filepath.Join(root, "internal/plugins/externals/ebpf/internal/c/process_sched/process_sched.h")
	layouts, err := readCLayouts(header, specs)
	if err != nil {
		return err
	}

	if out == "" {
		out = filepath.Join(root, "internal/plugins/externals/ebpf/internal/procwatch/runtime_layout_gen.go")
	}

	src, err := renderLayouts("procwatch", "linux", "linux", "", layouts, specs)
	if err != nil {
		return err
	}

	return writeGeneratedFile(out, src)
}

func generateL7Flow(out string) error {
	root, err := findRepoRoot()
	if err != nil {
		return err
	}

	specs := []structSpec{
		{
			CName:  "connection_info",
			GoName: "ConnectionInfoC",
			Fields: []fieldSpec{
				{CName: "saddr", GoName: "saddr", GoType: "[4]uint32"},
				{CName: "daddr", GoName: "daddr", GoType: "[4]uint32"},
				{CName: "sport", GoName: "sport", GoType: "uint16"},
				{CName: "dport", GoName: "dport", GoType: "uint16"},
				{CName: "pid", GoName: "pid", GoType: "uint32"},
				{CName: "netns", GoName: "netns", GoType: "uint32"},
				{CName: "meta", GoName: "meta", GoType: "uint32"},
			},
		},
		{
			CName:  "id_generator",
			GoName: "CUniID",
			Fields: []fieldSpec{
				{CName: "init", GoName: "init", GoType: "uint8"},
				{CName: "_pad", GoName: "_pad", GoType: "uint8"},
				{CName: "cpu_id", GoName: "cpu_id", GoType: "uint16"},
				{CName: "id", GoName: "id", GoType: "uint32"},
				{CName: "ktime", GoName: "ktime", GoType: "uint64"},
			},
		},
		{
			CName:  "sk_inf",
			GoName: "cSkInf",
			Fields: []fieldSpec{
				{CName: "uni_id", GoName: "uni_id", GoType: "CUniID"},
				{CName: "index", GoName: "index", GoType: "uint64"},
				{CName: "skptr", GoName: "skptr", GoType: "uint64"},
				{CName: "conn", GoName: "conn", GoType: "ConnectionInfoC"},
			},
		},
		{
			CName:  "netdata_meta",
			GoName: "cNetdataMeta",
			Fields: []fieldSpec{
				{CName: "ts", GoName: "ts", GoType: "uint64"},
				{CName: "ts_tail", GoName: "ts_tail", GoType: "uint64"},
				{CName: "tid_utid", GoName: "tid_utid", GoType: "uint64"},
				{CName: "comm", GoName: "comm", GoType: "[16]uint8"},
				{CName: "sk_inf", GoName: "sk_inf", GoType: "cSkInf"},
				{CName: "tcp_seq", GoName: "tcp_seq", GoType: "uint32"},
				{CName: "_pad0", GoName: "_pad0", GoType: "uint16"},
				{CName: "func_id", GoName: "func_id", GoType: "uint16"},
				{CName: "original_size", GoName: "original_size", GoType: "int32"},
				{CName: "capture_size", GoName: "capture_size", GoType: "int32"},
			},
		},
		{
			CName:  "event_rec",
			GoName: "cEventRec",
			Fields: []fieldSpec{
				{CName: "num", GoName: "num", GoType: "uint32"},
				{CName: "bytes", GoName: "bytes", GoType: "uint32"},
			},
		},
		{
			CName:  "network_events",
			GoName: "CNetEvents",
			Fields: []fieldSpec{
				{CName: "rec", GoName: "rec", GoType: "cEventRec"},
				{CName: "payload", GoName: "payload", GoType: "bytes"},
			},
		},
		{
			CName:  "net_event_comm",
			GoName: "CNetEventComm",
			Fields: []fieldSpec{
				{CName: "rec", GoName: "rec", GoType: "cEventRec"},
				{CName: "meta", GoName: "meta", GoType: "cNetdataMeta"},
			},
		},
	}

	header := filepath.Join(root, "internal/plugins/externals/ebpf/internal/c/apiflow/l7_stats.h")
	layouts, err := readCLayouts(header, specs)
	if err != nil {
		return err
	}

	if out == "" {
		out = filepath.Join(root, "internal/plugins/externals/ebpf/internal/l7flow/l7_layout_gen.go")
	}

	src, err := renderLayouts("l7flow", "linux", "linux", "", layouts, specs)
	if err != nil {
		return err
	}

	return writeGeneratedFile(out, src)
}

func generateBashHistory(out string) error {
	root, err := findRepoRoot()
	if err != nil {
		return err
	}

	specs := []structSpec{
		{
			CName:  "bash_event",
			GoName: "BashEventC",
			Fields: []fieldSpec{
				{CName: "pid_tgid", GoName: "pid_tgid", GoType: "uint64"},
				{CName: "uid_gid", GoName: "uid_gid", GoType: "uint64"},
				{CName: "line", GoName: "line", GoType: "[128]uint8"},
			},
		},
	}

	header := filepath.Join(root, "internal/plugins/externals/ebpf/internal/c/bash_history/bash_history.h")
	layouts, err := readCLayouts(header, specs)
	if err != nil {
		return err
	}

	if out == "" {
		out = filepath.Join(root, "internal/plugins/externals/ebpf/internal/bashhistory/bash_history_layout_gen.go")
	}

	src, err := renderLayouts("bashhistory", "linux", "linux", "", layouts, specs)
	if err != nil {
		return err
	}

	return writeGeneratedFile(out, src)
}

func generateNetflow(out string) error {
	root, err := findRepoRoot()
	if err != nil {
		return err
	}

	specs := []structSpec{
		{
			CName:  "connection_info",
			GoName: "ConnectionInfoC",
			Fields: []fieldSpec{
				{CName: "saddr", GoName: "saddr", GoType: "[4]uint32"},
				{CName: "daddr", GoName: "daddr", GoType: "[4]uint32"},
				{CName: "sport", GoName: "sport", GoType: "uint16"},
				{CName: "dport", GoName: "dport", GoType: "uint16"},
				{CName: "pid", GoName: "pid", GoType: "uint32"},
				{CName: "netns", GoName: "netns", GoType: "uint32"},
				{CName: "meta", GoName: "meta", GoType: "uint32"},
			},
		},
		{
			CName:  "connection_stats",
			GoName: "ConnectionStatsC",
			Fields: []fieldSpec{
				{CName: "sent_bytes", GoName: "sent_bytes", GoType: "uint64"},
				{CName: "recv_bytes", GoName: "recv_bytes", GoType: "uint64"},
				{CName: "sent_packets", GoName: "sent_packets", GoType: "uint64"},
				{CName: "recv_packets", GoName: "recv_packets", GoType: "uint64"},
				{CName: "timestamp", GoName: "timestamp", GoType: "uint64"},
				{CName: "flags", GoName: "flags", GoType: "uint32"},
				{CName: "nat_daddr", GoName: "nat_daddr", GoType: "[4]uint32"},
				{CName: "nat_dport", GoName: "nat_dport", GoType: "uint16"},
				{CName: "direction", GoName: "direction", GoType: "uint8"},
				{CName: "_pad0", GoName: "_pad0", GoType: "uint8"},
				{CName: "cmd", GoName: "cmd", GoType: "[16]uint8"},
			},
		},
		{
			CName:  "connection_tcp_stats",
			GoName: "ConnectionTCPStatsC",
			Fields: []fieldSpec{
				{CName: "state_transitions", GoName: "state_transitions", GoType: "uint16"},
				{CName: "retransmits", GoName: "retransmits", GoType: "int32"},
				{CName: "rtt", GoName: "rtt", GoType: "uint32"},
				{CName: "rtt_var", GoName: "rtt_var", GoType: "uint32"},
				{CName: "connect_attempts", GoName: "connect_attempts", GoType: "uint32"},
				{CName: "connect_failures", GoName: "connect_failures", GoType: "uint32"},
				{CName: "close_wait", GoName: "close_wait", GoType: "uint32"},
				{CName: "last_ack", GoName: "last_ack", GoType: "uint32"},
				{CName: "time_wait", GoName: "time_wait", GoType: "uint32"},
			},
		},
		{
			CName:  "connection_closed_info",
			GoName: "ConncetionClosedInfoC",
			Fields: []fieldSpec{
				{CName: "conn_info", GoName: "conn_info", GoType: "ConnectionInfoC"},
				{CName: "conn_stats", GoName: "conn_stats", GoType: "ConnectionStatsC"},
				{CName: "conn_tcp_stats", GoName: "conn_tcp_stats", GoType: "ConnectionTCPStatsC"},
			},
		},
	}

	header := filepath.Join(root, "internal/plugins/externals/ebpf/internal/c/netflow/conn_stats.h")
	layouts, err := readCLayouts(header, specs)
	if err != nil {
		return err
	}

	if out == "" {
		out = filepath.Join(root, "internal/plugins/externals/ebpf/internal/netflow/conn_stats_layout_gen.go")
	}

	const aliases = `type _Ctype_uint = uint32
type _Ctype_ushort = uint16
type _Ctype_struct_connection_info = ConnectionInfoC
type _Ctype_struct_connection_stats = ConnectionStatsC
type _Ctype_struct_connection_tcp_stats = ConnectionTCPStatsC
`

	src, err := renderLayouts("netflow", "linux", "linux", aliases, layouts, specs)
	if err != nil {
		return err
	}

	return writeGeneratedFile(out, src)
}

func writeGeneratedFile(out string, src []byte) error {
	// Generated Go sources are committed and should keep normal source permissions.
	return os.WriteFile(out, src, 0o644) //nolint:gosec
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("could not find go.mod")
		}
		dir = parent
	}
}

func readCLayouts(header string, specs []structSpec) (map[string]structLayout, error) {
	tmpDir, err := os.MkdirTemp("", "dk-genlayout-*")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	source := filepath.Join(tmpDir, "layout.c")
	binary := filepath.Join(tmpDir, "layout")
	if err := os.WriteFile(source, []byte(layoutProgram(header, specs)), 0o600); err != nil {
		return nil, err
	}

	cc, err := findCC()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(cc, source, "-o", binary) //nolint:gosec
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s failed: %w: %s", cc, err, strings.TrimSpace(string(out)))
	}

	out, err = exec.Command(binary).CombinedOutput() //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("layout probe failed: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return parseLayoutOutput(out)
}

func findCC() (string, error) {
	for _, cc := range []string{"cc", "clang", "gcc"} {
		path, err := exec.LookPath(cc)
		if err == nil {
			return path, nil
		}
	}
	return "", errors.New("no C compiler found")
}

func layoutProgram(header string, specs []structSpec) string {
	var b strings.Builder

	fmt.Fprintf(&b, "#include <stddef.h>\n")
	fmt.Fprintf(&b, "#include <stdio.h>\n")
	fmt.Fprintf(&b, "#include %q\n\n", header)
	fmt.Fprintf(&b, "int main(void) {\n")
	for _, spec := range specs {
		fmt.Fprintf(&b, "  printf(\"struct %s %%zu %%zu\\n\", sizeof(struct %s), _Alignof(struct %s));\n", spec.CName, spec.CName, spec.CName)
		for _, field := range spec.Fields {
			fmt.Fprintf(&b, "  printf(\"field %s %s %%zu %%zu\\n\", offsetof(struct %s, %s), sizeof(((struct %s*)0)->%s));\n",
				spec.CName, field.CName, spec.CName, field.CName, spec.CName, field.CName)
		}
	}
	fmt.Fprintf(&b, "  return 0;\n}\n")
	return b.String()
}

func parseLayoutOutput(out []byte) (map[string]structLayout, error) {
	layouts := map[string]structLayout{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Fields(line)
		switch parts[0] {
		case "struct":
			if len(parts) != 4 {
				return nil, fmt.Errorf("invalid struct line %q", line)
			}
			size, err := strconv.Atoi(parts[2])
			if err != nil {
				return nil, err
			}
			align, err := strconv.Atoi(parts[3])
			if err != nil {
				return nil, err
			}
			layouts[parts[1]] = structLayout{
				Size:   size,
				Align:  align,
				Fields: map[string]fieldLayout{},
			}
		case "field":
			if len(parts) != 5 {
				return nil, fmt.Errorf("invalid field line %q", line)
			}
			layout, ok := layouts[parts[1]]
			if !ok {
				return nil, fmt.Errorf("field before struct %q", line)
			}
			offset, err := strconv.Atoi(parts[3])
			if err != nil {
				return nil, err
			}
			size, err := strconv.Atoi(parts[4])
			if err != nil {
				return nil, err
			}
			layout.Fields[parts[2]] = fieldLayout{Offset: offset, Size: size}
			layouts[parts[1]] = layout
		default:
			return nil, fmt.Errorf("invalid layout line %q", line)
		}
	}
	return layouts, nil
}

func renderLayouts(packageName, goBuild, plusBuild, aliases string, layouts map[string]structLayout, specs []structSpec) ([]byte, error) {
	type renderedStruct struct {
		Name   string
		Size   int
		Align  int
		Fields []goField
	}

	rendered := make([]renderedStruct, 0, len(specs))
	for _, spec := range specs {
		layout, ok := layouts[spec.CName]
		if !ok {
			return nil, fmt.Errorf("missing layout for struct %s", spec.CName)
		}

		fields, err := buildGoFields(spec, layout)
		if err != nil {
			return nil, err
		}
		rendered = append(rendered, renderedStruct{
			Name:   spec.GoName,
			Size:   layout.Size,
			Align:  layout.Align,
			Fields: fields,
		})
	}

	sort.SliceStable(rendered, func(i, j int) bool {
		return rendered[i].Name < rendered[j].Name
	})

	const tpl = `// Code generated by genlayout; DO NOT EDIT.

//go:build {{.GoBuild}}
// +build {{.PlusBuild}}

package {{.PackageName}}

{{.Aliases}}
{{range .Structs}}
type {{.Name}} struct {
{{- range .Fields}}
	{{.Name}} {{.Type}}
{{- end}}
}

{{end -}}
`

	data := struct {
		PackageName string
		GoBuild     string
		PlusBuild   string
		Aliases     string
		Structs     []renderedStruct
	}{
		PackageName: packageName,
		GoBuild:     goBuild,
		PlusBuild:   plusBuild,
		Aliases:     aliases,
		Structs:     rendered,
	}

	var buf bytes.Buffer
	if err := template.Must(template.New("layout").Parse(tpl)).Execute(&buf, data); err != nil {
		return nil, err
	}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated source: %w\n%s", err, buf.String())
	}
	return src, nil
}

func buildGoFields(spec structSpec, layout structLayout) ([]goField, error) {
	fields := make([]goField, 0, len(spec.Fields))
	cur := 0
	padIdx := 0

	for _, field := range spec.Fields {
		fl, ok := layout.Fields[field.CName]
		if !ok {
			return nil, fmt.Errorf("missing field %s.%s", spec.CName, field.CName)
		}
		if fl.Offset < cur {
			return nil, fmt.Errorf("overlapping field %s.%s", spec.CName, field.CName)
		}
		if fl.Offset > cur {
			fields = append(fields, goField{
				Name: fmt.Sprintf("_pad%d", padIdx),
				Type: fmt.Sprintf("[%d]byte", fl.Offset-cur),
			})
			padIdx++
		}

		name := field.GoName
		if name == "_pad0" {
			name = "_c_pad0"
		}
		goType := field.GoType
		if goType == "bytes" {
			goType = fmt.Sprintf("[%d]uint8", fl.Size)
		}
		fields = append(fields, goField{
			Name: name,
			Type: goType,
		})
		cur = fl.Offset + fl.Size
	}

	if layout.Size > cur {
		fields = append(fields, goField{
			Name: fmt.Sprintf("_pad%d", padIdx),
			Type: fmt.Sprintf("[%d]byte", layout.Size-cur),
		})
	}

	return fields, nil
}
