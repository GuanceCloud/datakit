//go:build linux
// +build linux

//nolint:gosec,lll
package offset

//go:generate go run ../c/genlayout -target offset

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/cilium/ebpf"
	"github.com/goccy/go-json"
	"golang.org/x/sys/unix"

	"github.com/GuanceCloud/cliutils/logger"
	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/c"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/exporter"
)

//nolint:unused
type OffsetCheck struct {
	skNumOk           uint64
	inetSportOk       uint64
	skFamilyOk        uint64
	skRcvSaddrOk      uint64
	skDaddrOk         uint64
	skV6RcvSaddrOk    uint64
	skV6DaddrOk       uint64
	skDportOk         uint64
	tcpSkSrttUsOk     uint64
	tcpSkMdevUsOk     uint64
	flowi4SaddrOk     uint64
	flowi4DaddrOk     uint64
	flowi4SportOk     uint64
	flowi4DportOk     uint64
	flowi6SaddrOk     uint64
	flowi6DaddrOk     uint64
	flowi6SportOk     uint64
	flowi6DportOk     uint64
	skaddrSinPortOk   uint64
	skaddr6Sin6PortOk uint64
	sknetOk           uint64
	netnsInumOk       uint64
	socketSkOK        uint64

	ctOriginTupleOk uint64
	ctReplyTupleOk  uint64
	ctNetOk         uint64
}

const KernelTaskCommLen = 16 // Maximum length of process(thread task) name

const (
	offsetCacheVersion       = "4"
	offsetCacheLegacyVersion = "3"
)

//nolint:stylecheck
const (
	GUESS_SK_NUM = iota + 1
	GUESS_INET_SPORT
	GUESS_SK_FAMILY
	GUESS_SK_RCV_SADDR
	GUESS_SK_DADDR
	GUESS_SK_V6_RCV_SADDR
	GUESS_SK_V6_DADDR
	GUESS_SK_DPORT
	GUESS_TCP_SK_SRTT_US
	GUESS_TCP_SK_MDEV_US
	GUESS_FLOWI4_SADDR
	GUESS_FLOWI4_DADDR
	GUESS_FLOWI4_SPORT
	GUESS_FLOWI4_DPORT
	GUESS_FLOWI6_SADDR
	GUESS_FLOWI6_DADDR
	GUESS_FLOWI6_SPORT
	GUESS_FLOWI6_DPORT
	GUESS_SKADDR_SIN_PORT
	GUESS_SKADRR6_SIN6_PORT
	GUESS_SK_NET
	GUESS_NS_COMMON_INUM
	GUESS_SOCKET_SK

	GUESS_CONNTRACK_TUPLE_ORIGIN
	GUESS_CONNTRACK_TUPLE_REPLY
)

//nolint:stylecheck
const (
	ERR_G_NOERROR = 0
	ERR_G_SK_NET  = 19
)

var l = logger.DefaultSLogger("net_ebpf")

func SetLogger(nl *logger.Logger) {
	l = nl
}

func NewGuessRuntime(guessed *OffsetGuessC, ipv6Disabled bool) (*bpfutil.Runtime, error) {
	probes := make([]*bpfutil.HookSpec, 0, 5)

	if kernelGuessNeedsTCP4(guessed) {
		probes = append(probes, &bpfutil.HookSpec{
			ID: bpfutil.HookID{
				EBPFFuncName: "kprobe__tcp_getsockopt",
			},
		})
	}

	if kernelGuessNeedsUDP4(guessed) {
		probes = append(probes, &bpfutil.HookSpec{
			ID: bpfutil.HookID{
				EBPFFuncName: "kprobe__ip_make_skb",
			},
		})
	}

	if kernelGuessNeedsSocket(guessed) {
		probes = append(probes, &bpfutil.HookSpec{
			ID: bpfutil.HookID{
				EBPFFuncName: "kprobe__sock_common_getsockopt",
			},
		})
	}

	if !ipv6Disabled && kernelGuessNeedsTCP6(guessed) {
		probes = append(probes,
			&bpfutil.HookSpec{
				ID: bpfutil.HookID{
					EBPFFuncName: "kprobe__tcp_v6_connect",
				},
			},
			&bpfutil.HookSpec{
				ID: bpfutil.HookID{
					EBPFFuncName: "kretprobe__tcp_v6_connect",
				},
			},
		)
	}

	if len(probes) == 0 {
		return nil, fmt.Errorf("no active offset probes required")
	}

	filtered := make([]*bpfutil.HookSpec, 0, len(probes))
	for _, probe := range probes {
		if probe == nil {
			continue
		}

		symbol, ok := bpfutil.KernelProbeSymbol(*probe)
		if !ok {
			filtered = append(filtered, probe)
			continue
		}

		found, err := bpfutil.HasKernelSymbol(symbol)
		if err != nil {
			exporter.RecordKernelFunctionStatus("offset", probe.ID.Program, symbol, "detect_error", err.Error())
			l.Warnf("detect kernel symbol %q for offset probe %q failed: %v",
				symbol, probe.ID.Program, err)
			filtered = append(filtered, probe)
			continue
		}
		if !found {
			exporter.RecordKernelFunctionStatus("offset", probe.ID.Program, symbol, "missing", "kernel symbol not found")
			l.Warnf("skip offset probe %q: kernel symbol %q not found",
				probe.ID.Program, symbol)
			continue
		}

		exporter.RecordKernelFunctionStatus("offset", probe.ID.Program, symbol, "available", "")
		filtered = append(filtered, probe)
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no offset probes available on current kernel")
	}

	m := &bpfutil.Runtime{
		Probes: filtered,
	}
	loadSpec := bpfutil.LoadSpec{
		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},
	}
	binLoader := dkebpf.OffsetGuessBin
	binName := "offset_guess.o"
	if useHashMaps, kernelVersion, err := bpfutil.UseHashMapObjects(); err != nil {
		l.Warnf("detect LRU hash map support for offset guess failed: %v", err)
	} else if useHashMaps {
		binLoader = dkebpf.OffsetGuessLegacyBin
		binName = "offset_guess_legacy.o"
		l.Infof("kernel %#x loading legacy offset guess object without LRU hash maps", kernelVersion)
	}
	if buf, err := binLoader(); err != nil {
		return nil, fmt.Errorf("%s: %w", binName, err)
	} else if err := m.LoadFromReader((bytes.NewReader(buf)), loadSpec); err != nil {
		return nil, fmt.Errorf("init offset guess: %w", err)
	}
	return m, nil
}

func NewOffsetHTTPFlowRuntime() (*bpfutil.Runtime, error) {
	useLegacyConsts, _, _ := bpfutil.UseLegacyConstObjects()
	m := &bpfutil.Runtime{
		Probes: []*bpfutil.HookSpec{
			{
				ID: bpfutil.HookID{
					EBPFFuncName: "kprobe__sock_common_getsockopt",
				},
			},
			{
				ID: bpfutil.HookID{
					EBPFFuncName: "kpretrobe__sock_common_getsockopt",
				},
			},
		},
	}
	loadSpec := bpfutil.LoadSpec{
		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},
		LegacyConstants: useLegacyConsts,
	}
	binLoader := dkebpf.OffsetHttpflowBin
	binName := "offset_httpflow.o"
	if useLegacyConsts {
		binLoader = dkebpf.OffsetHttpflowLegacyBin
		binName = "offset_httpflow_legacy.o"
	}
	loadRuntime := func(name string, loader func() ([]byte, error), legacyConstants bool) error {
		loadSpec.LegacyConstants = legacyConstants
		buf, err := loader()
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		if err := m.LoadFromReader((bytes.NewReader(buf)), loadSpec); err != nil {
			return fmt.Errorf("init offset httpflow guess: %w", err)
		}
		return nil
	}
	if err := loadRuntime(binName, binLoader, useLegacyConsts); err != nil {
		if useLegacyConsts {
			return nil, err
		}
		l.Warnf("load modern offset httpflow object failed, fallback to legacy object without LRU hash maps: %v", err)
		if err := loadRuntime("offset_httpflow_legacy.o", dkebpf.OffsetHttpflowLegacyBin, true); err != nil {
			return nil, err
		}
	}
	return m, nil
}

func readMapGuessStatus(m *ebpf.Map) (*OffsetGuessC, error) {
	status := OffsetGuessC{}
	zero := uint64(0)
	if err := m.Lookup(&zero, unsafe.Pointer(&status)); err != nil {
		return nil, err
	} else {
		return &status, err
	}
}

func readMapGuessConntrack(m *ebpf.Map) (*OffsetConntrackC, error) {
	status := OffsetConntrackC{}
	var zero uint64 = 0
	if err := m.Lookup(&zero, unsafe.Pointer(&status)); err != nil {
		return nil, err
	} else {
		return &status, err
	}
}

func updateMapGuessStatus(m *ebpf.Map, status *OffsetGuessC) error {
	zero := uint64(0)
	status.daddr = [4]_Ctype_uint{}
	status.saddr = [4]_Ctype_uint{}
	status.daddr_skt = [4]_Ctype_uint{}
	status.dport = 0
	status.sport = 0
	status.dport_skt = 0
	status.sport_skt = 0
	status.family_skt = 0
	status.rtt = 0
	status.rtt_var = 0
	status.netns = 0
	status.netns_skt = 0
	status.err = 0
	status.state = 0

	return m.Update(&zero, unsafe.Pointer(status), ebpf.UpdateAny)
}

func BpfMapGuessInit(runtime *bpfutil.Runtime) (*ebpf.Map, error) {
	bpfmapOffsetGuess, err := runtime.LookupMap("bpfmap_offset_guess")
	if err != nil {
		return nil, fmt.Errorf("lookup bpf map bpfmap_offset_guess: %w", err)
	}

	zero := uint64(0)
	status := newGuessStatus()
	err = bpfmapOffsetGuess.Update(zero, unsafe.Pointer(&status), ebpf.UpdateAny)
	if err != nil {
		return nil, err
	}
	time.Sleep(time.Millisecond * 5)
	return bpfmapOffsetGuess, nil
}

func BpfMapGuessHTTPInit(runtime *bpfutil.Runtime) (*ebpf.Map, error) {
	bpfmapOffsetHTTP, err := runtime.LookupMap("bpf_map_offset_httpflow")
	if err != nil {
		return nil, fmt.Errorf("lookup bpf map bpf_map_offset_httpflow: %w", err)
	}

	key := uint64(0)
	value := newGuessHTTP()

	err = bpfmapOffsetHTTP.Update(key, unsafe.Pointer(&value), ebpf.UpdateAny)
	if err != nil {
		return nil, err
	}

	time.Sleep(time.Millisecond * 5)

	return bpfmapOffsetHTTP, nil
}

func newGuessStatus() OffsetGuessC {
	procName := filepath.Base(os.Args[0])
	if len(procName) > KernelTaskCommLen-1 {
		procName = procName[:KernelTaskCommLen-1]
	}

	procNameC := [KernelTaskCommLen]_Ctype_uchar{}
	for i := 0; i < KernelTaskCommLen-1 && i < len(procName); i++ {
		procNameC[i] = procName[i]
	}

	status := OffsetGuessC{
		process_name: procNameC,
		pid_tgid:     uint64(unix.Getpid())<<32 | uint64(unix.Gettid()),
	}

	return status
}

func newGuessHTTP() OffsetHTTPFlowC {
	procName := filepath.Base(os.Args[0])
	if len(procName) > KernelTaskCommLen-1 {
		procName = procName[:KernelTaskCommLen-1]
	}

	procNameC := [KernelTaskCommLen]_Ctype_uchar{}
	for i := 0; i < KernelTaskCommLen-1 && i < len(procName); i++ {
		procNameC[i] = procName[i]
	}

	httpOffset := OffsetHTTPFlowC{
		process_name: procNameC,
		pid_tgid:     uint64(unix.Getpid())<<32 | uint64(unix.Gettid()),
	}

	return httpOffset
}

func newGuessConntrack() OffsetConntrackC {
	procName := filepath.Base(os.Args[0])
	if len(procName) > KernelTaskCommLen-1 {
		procName = procName[:KernelTaskCommLen-1]
	}

	procNameC := [KernelTaskCommLen]_Ctype_uchar{}
	for i := 0; i < KernelTaskCommLen-1 && i < len(procName); i++ {
		procNameC[i] = procName[i]
	}

	offset := OffsetConntrackC{
		process_name: procNameC,
		pid_tgid:     uint64(unix.Getpid())<<32 | uint64(unix.Gettid()),
	}

	return offset
}

func newGuessTCPSeq() OffsetTCPSeqC {
	procName := filepath.Base(os.Args[0])
	if len(procName) > KernelTaskCommLen-1 {
		procName = procName[:KernelTaskCommLen-1]
	}

	procNameC := [KernelTaskCommLen]_Ctype_uchar{}
	for i := 0; i < KernelTaskCommLen-1 && i < len(procName); i++ {
		procNameC[i] = procName[i]
	}

	offset := OffsetTCPSeqC{
		process_name: procNameC,
		pid_tgid:     uint64(unix.Getpid())<<32 | uint64(unix.Gettid()),
	}

	return offset
}

func copyOffsetCT(src, dst *OffsetConntrackC) {
	dst.offset_ct_origin_tuple = src.offset_ct_origin_tuple
	dst.offset_ct_reply_tuple = src.offset_ct_reply_tuple
	dst.offset_ct_net = src.offset_ct_net
	dst.offset_ct_ns_common_inum = src.offset_ct_ns_common_inum
}

func copyOffsetCTRuntime(src, dst *OffsetConntrackC) {
	dst.err = src.err
	dst.state = src.state
	dst.pid_tgid = src.pid_tgid
	dst.origin = src.origin
	dst.reply = src.reply
	dst.netns = src.netns
}

func copyOffset(src *OffsetGuessC, dst *OffsetGuessC) {
	dst.offset_sk_num = src.offset_sk_num
	dst.offset_inet_sport = src.offset_inet_sport
	dst.offset_sk_family = src.offset_sk_family
	dst.offset_sk_rcv_saddr = src.offset_sk_rcv_saddr
	dst.offset_sk_daddr = src.offset_sk_daddr
	dst.offset_sk_v6_rcv_saddr = src.offset_sk_v6_rcv_saddr
	dst.offset_sk_v6_daddr = src.offset_sk_v6_daddr
	dst.offset_sk_dport = src.offset_sk_dport
	dst.offset_tcp_sk_srtt_us = src.offset_tcp_sk_srtt_us
	dst.offset_tcp_sk_mdev_us = src.offset_tcp_sk_mdev_us

	dst.offset_flowi4_saddr = src.offset_flowi4_saddr
	dst.offset_flowi4_daddr = src.offset_flowi4_daddr
	dst.offset_flowi4_sport = src.offset_flowi4_sport
	dst.offset_flowi4_dport = src.offset_flowi4_dport

	dst.offset_flowi6_saddr = src.offset_flowi6_saddr
	dst.offset_flowi6_daddr = src.offset_flowi6_daddr
	dst.offset_flowi6_sport = src.offset_flowi6_sport
	dst.offset_flowi6_dport = src.offset_flowi6_dport

	dst.offset_skaddr_sin_port = src.offset_skaddr_sin_port
	dst.offset_skaddr6_sin6_port = src.offset_skaddr6_sin6_port

	dst.offset_sk_net = src.offset_sk_net
	dst.offset_ns_common_inum = src.offset_ns_common_inum

	dst.offset_socket_sk = src.offset_socket_sk
}

func copySupplementalOffsets(src *OffsetGuessC, dst *OffsetGuessC) {
	dst.offset_task_struct_files = src.offset_task_struct_files
	dst.offset_files_struct_fdt = src.offset_files_struct_fdt
	dst.offset_socket_file = src.offset_socket_file
	dst.offset_file_private_data = src.offset_file_private_data
	dst.offset_ct_net = src.offset_ct_net
	dst.offset_ct_ns_common_inum = src.offset_ct_ns_common_inum
	dst.offset_origin_tuple = src.offset_origin_tuple
	dst.offset_reply_tuple = src.offset_reply_tuple
}

func guessProcessName(status *OffsetGuessC) string {
	if status == nil {
		return ""
	}

	var procName []byte
	for _, ch := range status.process_name {
		if ch == 0 {
			break
		}
		procName = append(procName, ch)
	}

	return string(procName)
}

func guessErrorString(errCode int64) string {
	switch errCode {
	case ERR_G_NOERROR:
		return "ok"
	case ERR_G_SK_NET:
		return "ns_inum"
	default:
		return fmt.Sprintf("unknown(%d)", errCode)
	}
}

func kernelGuessProgress(check *OffsetCheck) string {
	if check == nil {
		return ""
	}

	return fmt.Sprintf(
		"inet_sport=%d sk_dport=%d sk_daddr=%d sk_family=%d tcp_srtt=%d tcp_mdev=%d flowi4_saddr=%d flowi4_daddr=%d flowi4_dport=%d sk_net=%d netns=%d socket_sk=%d sk_v6_daddr=%d",
		check.inetSportOk,
		check.skDportOk,
		check.skDaddrOk,
		check.skFamilyOk,
		check.tcpSkSrttUsOk,
		check.tcpSkMdevUsOk,
		check.flowi4SaddrOk,
		check.flowi4DaddrOk,
		check.flowi4DportOk,
		check.sknetOk,
		check.netnsInumOk,
		check.socketSkOK,
		check.skV6DaddrOk,
	)
}

func kernelGuessRequiredReady(check *OffsetCheck, needTCP4, needUDP4, needTCP6 bool) bool {
	if check == nil {
		return false
	}

	if needTCP4 && !(check.skDaddrOk > MINSUCCESS &&
		check.skDportOk > MINSUCCESS &&
		check.skFamilyOk > MINSUCCESS) {
		return false
	}

	if needUDP4 && !(check.flowi4DaddrOk > MINSUCCESS &&
		check.flowi4DportOk > MINSUCCESS &&
		check.flowi4SaddrOk > MINSUCCESS) {
		return false
	}

	if needTCP6 && !(check.skV6DaddrOk > MINSUCCESS) {
		return false
	}

	return true
}

func kernelGuessOptionalReady(check *OffsetCheck, needTCP4, needTCP6 bool) bool {
	if check == nil {
		return false
	}

	if needTCP4 && !(check.inetSportOk > MINSUCCESS &&
		check.tcpSkSrttUsOk > MINSUCCESS &&
		check.tcpSkMdevUsOk > MINSUCCESS &&
		check.sknetOk > MINSUCCESS &&
		check.netnsInumOk > MINSUCCESS &&
		check.socketSkOK > MINSUCCESS) {
		return false
	}

	if needTCP6 && !(check.skV6DaddrOk > MINSUCCESS) {
		return false
	}

	return true
}

func finalizeKernelGuess(status *OffsetGuessC) *OffsetGuessC {
	newStatus := newGuessStatus()
	copyOffset(status, &newStatus)
	copySupplementalOffsets(status, &newStatus)

	if newStatus.offset_flowi4_daddr > newStatus.offset_flowi4_saddr {
		// + sizeof(flowi_common)
		newStatus.offset_flowi6_daddr = newStatus.offset_flowi4_saddr
	} else {
		newStatus.offset_flowi6_daddr = newStatus.offset_flowi4_daddr
	}
	newStatus.offset_flowi6_saddr = newStatus.offset_flowi6_daddr + 16 // +128bit
	newStatus.offset_flowi6_dport = newStatus.offset_flowi6_daddr + 36 // +256bit + 32bit
	newStatus.offset_flowi6_sport = newStatus.offset_flowi6_daddr + 38 // +256bit + 32bit +16bit

	return &newStatus
}

func tryGuessConntrack(status *OffsetConntrackC, check *OffsetCheck, conn *Conninfo,
	guessWhich int,
) bool {
	switch guessWhich {
	case GUESS_CONNTRACK_TUPLE_ORIGIN:
		if conn.Sport != status.origin.src_port ||
			conn.Saddr != *(*[4]uint32)(unsafe.Pointer(&status.origin.src_ip)) ||
			conn.Dport != status.origin.dst_port ||
			conn.Daddr != *(*[4]uint32)(unsafe.Pointer(&status.origin.dst_ip)) {
			status.offset_ct_origin_tuple++
			check.ctOriginTupleOk = 0
			return false
		} else {
			check.ctOriginTupleOk++
		}
	case GUESS_CONNTRACK_TUPLE_REPLY:
		if conn.Dport != status.reply.src_port ||
			conn.Daddr != *(*[4]uint32)(unsafe.Pointer(&status.reply.src_ip)) ||
			conn.Sport != status.reply.dst_port ||
			conn.Saddr != *(*[4]uint32)(unsafe.Pointer(&status.reply.dst_ip)) {
			status.offset_ct_reply_tuple++
			check.ctReplyTupleOk = 0
			return false
		} else {
			check.ctReplyTupleOk++
		}
	case GUESS_NS_COMMON_INUM:
		if status.err == ERR_G_SK_NET {
			status.offset_ct_net++
			status.offset_ct_ns_common_inum = 0
			check.ctNetOk = 0
			check.netnsInumOk = 0
			return false
		} else {
			if conn.NetNS != status.netns {
				status.offset_ct_ns_common_inum++
				check.ctNetOk = 0
				check.netnsInumOk = 0
				return false
			} else {
				check.netnsInumOk++
				check.ctNetOk++
			}
		}
	}

	return true
}

//nolint:gocyclo
func tryGuess(status *OffsetGuessC, check *OffsetCheck, conn *Conninfo, guessWhich int) bool {
	switch guessWhich {
	case GUESS_INET_SPORT:
		if conn.Sport != status.sport {
			status.offset_inet_sport++
			check.inetSportOk = 0
			return false
		} else {
			check.inetSportOk++
		}
	case GUESS_SK_FAMILY:
		if status.meta&ConnL3Mask != conn.Meta&ConnL3Mask {
			status.offset_sk_family++
			check.skFamilyOk = 0
			return false
		} else {
			check.skFamilyOk++
		}
	case GUESS_SK_DADDR:
		if conn.Daddr != *(*[4]uint32)(unsafe.Pointer(&status.daddr)) {
			status.offset_sk_daddr++
			check.skDaddrOk = 0
			return false
		} else {
			status.offset_sk_rcv_saddr = status.offset_sk_daddr + 4 // +32bit
			check.skDaddrOk++
		}
	case GUESS_SK_DPORT:
		if conn.Dport != status.dport {
			status.offset_sk_dport++
			check.skDportOk = 0
			return false
		} else {
			status.offset_sk_num = status.offset_sk_dport + 2 // +sizeof(__be16)
			check.skDportOk++
		}
	case GUESS_SK_V6_DADDR:
		if conn.Daddr != *(*[4]uint32)(unsafe.Pointer(&status.daddr)) {
			status.offset_sk_v6_daddr++
			check.skV6DaddrOk = 0
			return false
		} else {
			status.offset_sk_v6_rcv_saddr = status.offset_sk_v6_daddr + 16 // +128bit
			check.skV6DaddrOk++
		}
	case GUESS_TCP_SK_SRTT_US:
		if conn.Rtt != status.rtt {
			status.offset_tcp_sk_srtt_us++
			check.tcpSkSrttUsOk = 0
			return false
		} else {
			check.tcpSkSrttUsOk++
		}
	case GUESS_TCP_SK_MDEV_US:
		if conn.RttVar != status.rtt_var {
			status.offset_tcp_sk_mdev_us++
			check.tcpSkMdevUsOk = 0
			return false
		} else {
			check.tcpSkMdevUsOk++
		}
	case GUESS_SOCKET_SK:
		if !(check.skDportOk > MINSUCCESS &&
			check.skDaddrOk > MINSUCCESS &&
			check.skFamilyOk > MINSUCCESS) {
			return false
		}

		candidateMeta := socketCandidateMeta(status.family_skt)
		candidateDaddr := *(*[4]uint32)(unsafe.Pointer(&status.daddr_skt))
		netnsMatch := true
		if check.netnsInumOk > MINSUCCESS {
			netnsMatch = conn.NetNS == status.netns_skt
		}

		if !(conn.Sport == status.sport_skt &&
			conn.Dport == status.dport_skt &&
			conn.Daddr == candidateDaddr &&
			(conn.Meta&ConnL3Mask) == candidateMeta &&
			netnsMatch) {
			status.offset_socket_sk++
			check.socketSkOK = 0
			return false
		} else {
			check.socketSkOK++
		}
	case GUESS_FLOWI4_SADDR:
		if conn.Saddr != *(*[4]uint32)(unsafe.Pointer(&status.saddr)) {
			status.offset_flowi4_saddr++
			check.flowi4SaddrOk = 0
			return false
		} else {
			check.flowi4SaddrOk++
		}
	case GUESS_FLOWI4_DADDR:
		if conn.Daddr != *(*[4]uint32)(unsafe.Pointer(&status.daddr)) {
			status.offset_flowi4_daddr++
			check.flowi4DaddrOk = 0
			return false
		} else {
			check.flowi4DaddrOk++
		}
	case GUESS_FLOWI4_SPORT:
	case GUESS_FLOWI4_DPORT:
		if conn.Dport != status.dport {
			status.offset_flowi4_dport++
			check.flowi4DportOk = 0
			return false
		} else {
			status.offset_flowi4_sport = status.offset_flowi4_dport + 2 // +sizeof(__be16)
			check.flowi4DportOk++
		}
	case GUESS_FLOWI6_SADDR:
	case GUESS_FLOWI6_DADDR:
	case GUESS_FLOWI6_SPORT:
	case GUESS_FLOWI6_DPORT:
	case GUESS_SKADDR_SIN_PORT:
	case GUESS_NS_COMMON_INUM:
		if status.err == ERR_G_SK_NET {
			status.offset_sk_net++
			status.offset_ns_common_inum = 0
			check.sknetOk = 0
			return false
		} else {
			check.sknetOk++

			if conn.NetNS != status.netns {
				status.offset_ns_common_inum++
				check.netnsInumOk = 0
				return false
			} else {
				check.netnsInumOk++
			}
		}
	}
	return true
}

func socketCandidateMeta(family uint16) uint32 {
	switch family {
	case unix.AF_INET:
		return ConnL3IPv4
	case unix.AF_INET6:
		return ConnL3IPv6
	default:
		return ConnL3Mask
	}
}

// Based on github.com/weaveworks/tcptracer-bpf.
func generateRandomIPv6Address() ([4]uint32, net.IP) {
	// multicast (ff00::/8) or link-local (fe80::/10) addresses don't work for
	// our purposes so let's choose a "random number" for the first 32 bits.
	addr := [4]uint32{}
	addr[0] = 0x87586031
	addr[1] = rand.Uint32() //nolint:gosec
	addr[2] = rand.Uint32() //nolint:gosec
	addr[3] = rand.Uint32() //nolint:gosec

	ip := net.IP{}
	for x := 0; x < 4; x++ {
		for y := 0; y < 4; y++ {
			ip = append(ip,
				byte((addr[x]&(0xff<<(8*y)))>>(8*y)),
			)
		}
	}
	return addr, ip
}

func getLinuxKernelVesion() (uint64, error) {
	return bpfutil.CurrentKernelVersion()
}

func currentOffsetCacheKernelVersion() (string, error) {
	kernelVersion, err := getLinuxKernelVesion()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%#x", kernelVersion), nil
}

type OffsetCacheMeta struct {
	KernelVersion string `json:"kernel_version"`
}

type OffsetCacheKernel struct {
	SkNum           uint64 `json:"offset_sk_num"`
	InetSport       uint64 `json:"offset_inet_sport"`
	SkFamily        uint64 `json:"offset_sk_family"`
	SkRcvSaddr      uint64 `json:"offset_sk_rcv_saddr"`
	SkDaddr         uint64 `json:"offset_sk_daddr"`
	SkV6RcvSaddr    uint64 `json:"offset_sk_v6_rcv_saddr"`
	SkV6Daddr       uint64 `json:"offset_sk_v6_daddr"`
	SkDport         uint64 `json:"offset_sk_dport"`
	TCPSkSrttUs     uint64 `json:"offset_tcp_sk_srtt_us"`
	TCPSkMdevUs     uint64 `json:"offset_tcp_sk_mdev_us"`
	Flowi4Saddr     uint64 `json:"offset_flowi4_saddr"`
	Flowi4Daddr     uint64 `json:"offset_flowi4_daddr"`
	Flowi4SPort     uint64 `json:"offset_flowi4_sport"`
	Flowi4DPort     uint64 `json:"offset_flowi4_dport"`
	Flowi6SAddr     uint64 `json:"offset_flowi6_saddr"`
	Flowi6DAddr     uint64 `json:"offset_flowi6_daddr"`
	Flowi6SPort     uint64 `json:"offset_flowi6_sport"`
	Flowi6Dport     uint64 `json:"offset_flowi6_dport"`
	SkAddrSinPort   uint64 `json:"offset_skaddr_sin_port"`
	SkAddr6Sin6Port uint64 `json:"offset_skaddr6_sin6_port"`
	SkNet           uint64 `json:"offset_sk_net"`
	NsCommonInum    uint64 `json:"offset_ns_common_inum"`
	SocketSk        uint64 `json:"offset_socket_sk"`
}

type OffsetCacheTCPSeq struct {
	CopiedSeq uint64 `json:"offset_copied_seq"`
	WriteSeq  uint64 `json:"offset_write_seq"`
}

type OffsetCacheHTTPFlow struct {
	TaskFiles       uint64 `json:"offset_task_struct_files"`
	FileFDT         uint64 `json:"offset_files_struct_fdt"`
	SocketFile      uint64 `json:"offset_socket_file"`
	FilePrivateData uint64 `json:"offset_file_private_data"`
}

type OffsetCacheConntrack struct {
	CTNet          uint64 `json:"offset_ct_net"`
	CTNsCommonInum uint64 `json:"offset_ct_ns_common_inum"`
	CTOriginTuple  uint64 `json:"offset_origin_tuple"`
	CTReplyTuple   uint64 `json:"offset_reply_tuple"`
}

type OffsetCache struct {
	Version   string               `json:"version"`
	Meta      OffsetCacheMeta      `json:"meta"`
	Kernel    OffsetCacheKernel    `json:"kernel"`
	TCPSeq    OffsetCacheTCPSeq    `json:"tcp_seq"`
	HTTPFlow  OffsetCacheHTTPFlow  `json:"httpflow"`
	Conntrack OffsetCacheConntrack `json:"conntrack"`
}

type OffsetCacheLegacy struct {
	Version         string `json:"version"`
	KernelVersion   string `json:"kernel_version"`
	SkNum           uint64 `json:"offset_sk_num"`
	InetSport       uint64 `json:"offset_inet_sport"`
	SkFamily        uint64 `json:"offset_sk_family"`
	SkRcvSaddr      uint64 `json:"offset_sk_rcv_saddr"`
	SkDaddr         uint64 `json:"offset_sk_daddr"`
	SkV6RcvSaddr    uint64 `json:"offset_sk_v6_rcv_saddr"`
	SkV6Daddr       uint64 `json:"offset_sk_v6_daddr"`
	SkDport         uint64 `json:"offset_sk_dport"`
	TCPSkSrttUs     uint64 `json:"offset_tcp_sk_srtt_us"`
	TCPSkMdevUs     uint64 `json:"offset_tcp_sk_mdev_us"`
	Flowi4Saddr     uint64 `json:"offset_flowi4_saddr"`
	Flowi4Daddr     uint64 `json:"offset_flowi4_daddr"`
	Flowi4SPort     uint64 `json:"offset_flowi4_sport"`
	Flowi4DPort     uint64 `json:"offset_flowi4_dport"`
	Flowi6SAddr     uint64 `json:"offset_flowi6_saddr"`
	Flowi6DAddr     uint64 `json:"offset_flowi6_daddr"`
	Flowi6SPort     uint64 `json:"offset_flowi6_sport"`
	Flowi6Dport     uint64 `json:"offset_flowi6_dport"`
	SkAddrSinPort   uint64 `json:"offset_skaddr_sin_port"`
	SkAddr6Sin6Port uint64 `json:"offset_skaddr6_sin6_port"`
	SkNet           uint64 `json:"offset_sk_net"`
	NsCommonInum    uint64 `json:"offset_ns_common_inum"`
	SocketSk        uint64 `json:"offset_socket_sk"`
	CopiedSeq       uint64 `json:"offset_copied_seq"`
	WriteSeq        uint64 `json:"offset_write_seq"`
	TaskFiles       uint64 `json:"offset_task_struct_files"`
	FileFDT         uint64 `json:"offset_files_struct_fdt"`
	SocketFile      uint64 `json:"offset_socket_file"`
	FilePrivateData uint64 `json:"offset_file_private_data"`
	CTNet           uint64 `json:"offset_ct_net"`
	CTNsCommonInum  uint64 `json:"offset_ct_ns_common_inum"`
	CTOriginTuple   uint64 `json:"offset_origin_tuple"`
	CTReplyTuple    uint64 `json:"offset_reply_tuple"`
}

func newOffsetCache(offsetC OffsetGuessC, kernelVersion string) OffsetCache {
	return OffsetCache{
		Version: offsetCacheVersion,
		Meta: OffsetCacheMeta{
			KernelVersion: kernelVersion,
		},
		Kernel: OffsetCacheKernel{
			SkNum:           offsetC.offset_sk_num,
			InetSport:       offsetC.offset_inet_sport,
			SkFamily:        offsetC.offset_sk_family,
			SkRcvSaddr:      offsetC.offset_sk_rcv_saddr,
			SkDaddr:         offsetC.offset_sk_daddr,
			SkV6RcvSaddr:    offsetC.offset_sk_v6_rcv_saddr,
			SkV6Daddr:       offsetC.offset_sk_v6_daddr,
			SkDport:         offsetC.offset_sk_dport,
			TCPSkSrttUs:     offsetC.offset_tcp_sk_srtt_us,
			TCPSkMdevUs:     offsetC.offset_tcp_sk_mdev_us,
			Flowi4Saddr:     offsetC.offset_flowi4_saddr,
			Flowi4Daddr:     offsetC.offset_flowi4_daddr,
			Flowi4SPort:     offsetC.offset_flowi4_sport,
			Flowi4DPort:     offsetC.offset_flowi4_dport,
			Flowi6SAddr:     offsetC.offset_flowi6_saddr,
			Flowi6DAddr:     offsetC.offset_flowi6_daddr,
			Flowi6SPort:     offsetC.offset_flowi6_sport,
			Flowi6Dport:     offsetC.offset_flowi6_dport,
			SkAddrSinPort:   offsetC.offset_skaddr_sin_port,
			SkAddr6Sin6Port: offsetC.offset_skaddr6_sin6_port,
			SkNet:           offsetC.offset_sk_net,
			NsCommonInum:    offsetC.offset_ns_common_inum,
			SocketSk:        offsetC.offset_socket_sk,
		},
		TCPSeq: OffsetCacheTCPSeq{
			CopiedSeq: offsetC.offset_copied_seq,
			WriteSeq:  offsetC.offset_write_seq,
		},
		HTTPFlow: OffsetCacheHTTPFlow{
			TaskFiles:       offsetC.offset_task_struct_files,
			FileFDT:         offsetC.offset_files_struct_fdt,
			SocketFile:      offsetC.offset_socket_file,
			FilePrivateData: offsetC.offset_file_private_data,
		},
		Conntrack: OffsetCacheConntrack{
			CTNet:          offsetC.offset_ct_net,
			CTNsCommonInum: offsetC.offset_ct_ns_common_inum,
			CTOriginTuple:  offsetC.offset_origin_tuple,
			CTReplyTuple:   offsetC.offset_reply_tuple,
		},
	}
}

func newLegacyOffsetCache(offsetC OffsetGuessC, kernelVersion string) OffsetCacheLegacy {
	return OffsetCacheLegacy{
		Version:         offsetCacheLegacyVersion,
		KernelVersion:   kernelVersion,
		SkNum:           offsetC.offset_sk_num,
		InetSport:       offsetC.offset_inet_sport,
		SkFamily:        offsetC.offset_sk_family,
		SkRcvSaddr:      offsetC.offset_sk_rcv_saddr,
		SkDaddr:         offsetC.offset_sk_daddr,
		SkV6RcvSaddr:    offsetC.offset_sk_v6_rcv_saddr,
		SkV6Daddr:       offsetC.offset_sk_v6_daddr,
		SkDport:         offsetC.offset_sk_dport,
		TCPSkSrttUs:     offsetC.offset_tcp_sk_srtt_us,
		TCPSkMdevUs:     offsetC.offset_tcp_sk_mdev_us,
		Flowi4Saddr:     offsetC.offset_flowi4_saddr,
		Flowi4Daddr:     offsetC.offset_flowi4_daddr,
		Flowi4SPort:     offsetC.offset_flowi4_sport,
		Flowi4DPort:     offsetC.offset_flowi4_dport,
		Flowi6SAddr:     offsetC.offset_flowi6_saddr,
		Flowi6DAddr:     offsetC.offset_flowi6_daddr,
		Flowi6SPort:     offsetC.offset_flowi6_sport,
		Flowi6Dport:     offsetC.offset_flowi6_dport,
		SkAddrSinPort:   offsetC.offset_skaddr_sin_port,
		SkAddr6Sin6Port: offsetC.offset_skaddr6_sin6_port,
		SkNet:           offsetC.offset_sk_net,
		NsCommonInum:    offsetC.offset_ns_common_inum,
		SocketSk:        offsetC.offset_socket_sk,
		CopiedSeq:       offsetC.offset_copied_seq,
		WriteSeq:        offsetC.offset_write_seq,
		TaskFiles:       offsetC.offset_task_struct_files,
		FileFDT:         offsetC.offset_files_struct_fdt,
		SocketFile:      offsetC.offset_socket_file,
		FilePrivateData: offsetC.offset_file_private_data,
		CTNet:           offsetC.offset_ct_net,
		CTNsCommonInum:  offsetC.offset_ct_ns_common_inum,
		CTOriginTuple:   offsetC.offset_origin_tuple,
		CTReplyTuple:    offsetC.offset_reply_tuple,
	}
}

func offsetGuessFromLegacyCache(offset OffsetCacheLegacy) OffsetGuessC {
	return OffsetGuessC{
		offset_sk_num:            offset.SkNum,
		offset_inet_sport:        offset.InetSport,
		offset_sk_family:         offset.SkFamily,
		offset_sk_rcv_saddr:      offset.SkRcvSaddr,
		offset_sk_daddr:          offset.SkDaddr,
		offset_sk_v6_rcv_saddr:   offset.SkV6RcvSaddr,
		offset_sk_v6_daddr:       offset.SkV6Daddr,
		offset_sk_dport:          offset.SkDport,
		offset_tcp_sk_srtt_us:    offset.TCPSkSrttUs,
		offset_tcp_sk_mdev_us:    offset.TCPSkMdevUs,
		offset_flowi4_saddr:      offset.Flowi4Saddr,
		offset_flowi4_daddr:      offset.Flowi4Daddr,
		offset_flowi4_sport:      offset.Flowi4SPort,
		offset_flowi4_dport:      offset.Flowi4DPort,
		offset_flowi6_saddr:      offset.Flowi6SAddr,
		offset_flowi6_daddr:      offset.Flowi6DAddr,
		offset_flowi6_sport:      offset.Flowi6SPort,
		offset_flowi6_dport:      offset.Flowi6Dport,
		offset_skaddr_sin_port:   offset.SkAddrSinPort,
		offset_skaddr6_sin6_port: offset.SkAddr6Sin6Port,
		offset_sk_net:            offset.SkNet,
		offset_ns_common_inum:    offset.NsCommonInum,
		offset_socket_sk:         offset.SocketSk,
		offset_copied_seq:        offset.CopiedSeq,
		offset_write_seq:         offset.WriteSeq,
		offset_task_struct_files: offset.TaskFiles,
		offset_files_struct_fdt:  offset.FileFDT,
		offset_socket_file:       offset.SocketFile,
		offset_file_private_data: offset.FilePrivateData,
		offset_ct_net:            offset.CTNet,
		offset_ct_ns_common_inum: offset.CTNsCommonInum,
		offset_origin_tuple:      offset.CTOriginTuple,
		offset_reply_tuple:       offset.CTReplyTuple,
	}
}

func offsetGuessFromCache(offset OffsetCache) OffsetGuessC {
	return OffsetGuessC{
		offset_sk_num:            offset.Kernel.SkNum,
		offset_inet_sport:        offset.Kernel.InetSport,
		offset_sk_family:         offset.Kernel.SkFamily,
		offset_sk_rcv_saddr:      offset.Kernel.SkRcvSaddr,
		offset_sk_daddr:          offset.Kernel.SkDaddr,
		offset_sk_v6_rcv_saddr:   offset.Kernel.SkV6RcvSaddr,
		offset_sk_v6_daddr:       offset.Kernel.SkV6Daddr,
		offset_sk_dport:          offset.Kernel.SkDport,
		offset_tcp_sk_srtt_us:    offset.Kernel.TCPSkSrttUs,
		offset_tcp_sk_mdev_us:    offset.Kernel.TCPSkMdevUs,
		offset_flowi4_saddr:      offset.Kernel.Flowi4Saddr,
		offset_flowi4_daddr:      offset.Kernel.Flowi4Daddr,
		offset_flowi4_sport:      offset.Kernel.Flowi4SPort,
		offset_flowi4_dport:      offset.Kernel.Flowi4DPort,
		offset_flowi6_saddr:      offset.Kernel.Flowi6SAddr,
		offset_flowi6_daddr:      offset.Kernel.Flowi6DAddr,
		offset_flowi6_sport:      offset.Kernel.Flowi6SPort,
		offset_flowi6_dport:      offset.Kernel.Flowi6Dport,
		offset_skaddr_sin_port:   offset.Kernel.SkAddrSinPort,
		offset_skaddr6_sin6_port: offset.Kernel.SkAddr6Sin6Port,
		offset_sk_net:            offset.Kernel.SkNet,
		offset_ns_common_inum:    offset.Kernel.NsCommonInum,
		offset_socket_sk:         offset.Kernel.SocketSk,
		offset_copied_seq:        offset.TCPSeq.CopiedSeq,
		offset_write_seq:         offset.TCPSeq.WriteSeq,
		offset_task_struct_files: offset.HTTPFlow.TaskFiles,
		offset_files_struct_fdt:  offset.HTTPFlow.FileFDT,
		offset_socket_file:       offset.HTTPFlow.SocketFile,
		offset_file_private_data: offset.HTTPFlow.FilePrivateData,
		offset_ct_net:            offset.Conntrack.CTNet,
		offset_ct_ns_common_inum: offset.Conntrack.CTNsCommonInum,
		offset_origin_tuple:      offset.Conntrack.CTOriginTuple,
		offset_reply_tuple:       offset.Conntrack.CTReplyTuple,
	}
}

func dumpOffset(offsetC OffsetGuessC) (string, error) {
	kernelVersion, err := currentOffsetCacheKernelVersion()
	if err != nil {
		return "", err
	}

	offset := newOffsetCache(offsetC, kernelVersion)
	buff := []byte{}
	buf := bytes.NewBuffer(buff)
	encoder := json.NewEncoder(buf)
	if err := encoder.Encode(offset); err != nil {
		return "", err
	} else {
		return buf.String(), nil
	}
}

func loadOffset(str string) (OffsetGuessC, error) {
	header := struct {
		Version string `json:"version"`
	}{}
	if err := json.NewDecoder(strings.NewReader(str)).Decode(&header); err != nil {
		return OffsetGuessC{}, err
	}

	kernelVersion, err := currentOffsetCacheKernelVersion()
	if err != nil {
		return OffsetGuessC{}, err
	}

	switch header.Version {
	case offsetCacheVersion:
		offset := &OffsetCache{}
		if err := json.NewDecoder(strings.NewReader(str)).Decode(offset); err != nil {
			return OffsetGuessC{}, err
		}
		if offset.Meta.KernelVersion != kernelVersion {
			return OffsetGuessC{}, fmt.Errorf("offset cache kernel mismatch: cache=%s current=%s",
				offset.Meta.KernelVersion, kernelVersion)
		}
		return offsetGuessFromCache(*offset), nil
	case offsetCacheLegacyVersion:
		offset := &OffsetCacheLegacy{}
		if err := json.NewDecoder(strings.NewReader(str)).Decode(offset); err != nil {
			return OffsetGuessC{}, err
		}
		if offset.KernelVersion != kernelVersion {
			return OffsetGuessC{}, fmt.Errorf("offset cache kernel mismatch: cache=%s current=%s",
				offset.KernelVersion, kernelVersion)
		}
		return offsetGuessFromLegacyCache(*offset), nil
	default:
		return OffsetGuessC{}, fmt.Errorf("unsupported offset cache version %q", header.Version)
	}
}
