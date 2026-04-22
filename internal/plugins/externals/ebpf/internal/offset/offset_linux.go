//go:build linux
// +build linux

// Package offset guess c struct offset
package offset

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"reflect"
	goruntime "runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/cilium/ebpf"
	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
	"golang.org/x/sys/unix"
)

const (
	ConnL3Mask = 0xFF // 0xFF
	ConnL3IPv4 = 0x00 // 0x00
	ConnL3IPv6 = 0x01 // 0x01

	ConnL4Mask                     = 0xFF00 // 0xFF00
	ConnL4TCP                      = 0x0000 // 0x00 << 8
	ConnL4UDP                      = 0x0100 // 0x01 << 8
	MAXOFFSET                      = 2048
	MINSUCCESS                     = 5
	maxKernelGuessRounds           = 128
	kernelGuessOptionalSlackRounds = 24
)

var (
	listenIPv4    = "127.0.0.2"
	listenIPv4Arr = [4]uint32{0, 0, 0, 0x0200007F}
)

type Conninfo struct {
	Saddr [4]uint32
	Daddr [4]uint32

	Sport uint16
	Dport uint16

	Meta uint32

	NetNS uint32

	Rtt    uint32
	RttVar uint32
}

const minKernelVersionB16 = 0x0004000400000000

func NewConstHTTPEditor(offsetHTTP *OffsetHTTPFlowC) []bpfutil.ConstantPatch {
	return []bpfutil.ConstantPatch{
		{
			Name:  "offset_task_struct_files",
			Value: uint64(offsetHTTP.offset_task_struct_files),
		},
		{
			Name:  "offset_files_struct_fdt",
			Value: uint64(offsetHTTP.offset_files_struct_fdt),
		},
		{
			Name:  "offset_socket_file",
			Value: uint64(offsetHTTP.offset_socket_file),
		},
		{
			Name:  "offset_file_private_data",
			Value: uint64(offsetHTTP.offset_file_private_data),
		},
	}
}

func HTTPFlowPatchesFromGuess(offset *OffsetGuessC) []bpfutil.ConstantPatch {
	if offset == nil {
		return nil
	}

	patches := []bpfutil.ConstantPatch{
		{
			Name:  "offset_task_struct_files",
			Value: uint64(offset.offset_task_struct_files),
		},
		{
			Name:  "offset_files_struct_fdt",
			Value: uint64(offset.offset_files_struct_fdt),
		},
		{
			Name:  "offset_socket_file",
			Value: uint64(offset.offset_socket_file),
		},
		{
			Name:  "offset_file_private_data",
			Value: uint64(offset.offset_file_private_data),
		},
	}

	if offset.offset_socket_sk != 0 {
		patches = append(patches, bpfutil.ConstantPatch{
			Name:  "offset_socket_sk",
			Value: uint64(offset.offset_socket_sk),
		})
	}

	return patches
}

func ApplyConstantPatches(dst *OffsetGuessC, patches []bpfutil.ConstantPatch) {
	if dst == nil || len(patches) == 0 {
		return
	}

	for _, patch := range patches {
		value, ok := patchUint64(patch.Value)
		if !ok {
			continue
		}

		switch patch.Name {
		case "offset_socket_sk":
			dst.offset_socket_sk = _Ctype_ulonglong(value)
		case "offset_task_struct_files":
			dst.offset_task_struct_files = _Ctype_ulonglong(value)
		case "offset_files_struct_fdt":
			dst.offset_files_struct_fdt = _Ctype_ulonglong(value)
		case "offset_socket_file":
			dst.offset_socket_file = _Ctype_ulonglong(value)
		case "offset_file_private_data":
			dst.offset_file_private_data = _Ctype_ulonglong(value)
		case "offset_ct_net":
			dst.offset_ct_net = _Ctype_ulonglong(value)
		case "offset_ct_ns_common_inum":
			dst.offset_ct_ns_common_inum = _Ctype_ulonglong(value)
		case "offset_ct_origin_tuple":
			dst.offset_origin_tuple = _Ctype_ulonglong(value)
		case "offset_ct_reply_tuple":
			dst.offset_reply_tuple = _Ctype_ulonglong(value)
		}
	}
}

func patchUint64(v interface{}) (uint64, bool) {
	switch value := v.(type) {
	case uint64:
		return value, true
	case uint32:
		return uint64(value), true
	case uint16:
		return uint64(value), true
	case uint8:
		return uint64(value), true
	case uint:
		return uint64(value), true
	case int64:
		if value < 0 {
			return 0, false
		}
		return uint64(value), true
	case int32:
		if value < 0 {
			return 0, false
		}
		return uint64(value), true
	case int:
		if value < 0 {
			return 0, false
		}
		return uint64(value), true
	default:
		rv := reflect.ValueOf(v)
		if !rv.IsValid() {
			return 0, false
		}
		kind := rv.Kind()
		if kind >= reflect.Uint && kind <= reflect.Uint64 {
			return rv.Uint(), true
		}
		if kind >= reflect.Int && kind <= reflect.Int64 {
			if rv.Int() < 0 {
				return 0, false
			}
			return uint64(rv.Int()), true
		}
		return 0, false
	}
}

func NewConstEditor(offsetGuess *OffsetGuessC) []bpfutil.ConstantPatch {
	kernelVersion, err := getLinuxKernelVesion()
	if err != nil {
		l.Error(err)
		kernelVersion = minKernelVersionB16
	}
	return []bpfutil.ConstantPatch{
		{
			Name:  "kernel_version",
			Value: kernelVersion,
		},
		{
			Name:  "offset_sk_num",
			Value: uint64(offsetGuess.offset_sk_num),
		},
		{
			Name:  "offset_inet_sport",
			Value: uint64(offsetGuess.offset_inet_sport),
		},
		{
			Name:  "offset_sk_family",
			Value: uint64(offsetGuess.offset_sk_family),
		},
		{
			Name:  "offset_sk_rcv_saddr",
			Value: uint64(offsetGuess.offset_sk_rcv_saddr),
		},
		{
			Name:  "offset_sk_daddr",
			Value: uint64(offsetGuess.offset_sk_daddr),
		},
		{
			Name:  "offset_sk_v6_rcv_saddr",
			Value: uint64(offsetGuess.offset_sk_v6_rcv_saddr),
		},
		{
			Name:  "offset_sk_v6_daddr",
			Value: uint64(offsetGuess.offset_sk_v6_daddr),
		},
		{
			Name:  "offset_sk_dport",
			Value: uint64(offsetGuess.offset_sk_dport),
		},
		{
			Name:  "offset_tcp_sk_srtt_us",
			Value: uint64(offsetGuess.offset_tcp_sk_srtt_us),
		},
		{
			Name:  "offset_tcp_sk_mdev_us",
			Value: uint64(offsetGuess.offset_tcp_sk_mdev_us),
		},
		{
			Name:  "offset_flowi4_saddr",
			Value: uint64(offsetGuess.offset_flowi4_saddr),
		},
		{
			Name:  "offset_flowi4_daddr",
			Value: uint64(offsetGuess.offset_flowi4_daddr),
		},
		{
			Name:  "offset_flowi4_sport",
			Value: uint64(offsetGuess.offset_flowi4_sport),
		},
		{
			Name:  "offset_flowi4_dport",
			Value: uint64(offsetGuess.offset_flowi4_dport),
		},
		{
			Name:  "offset_flowi6_saddr",
			Value: uint64(offsetGuess.offset_flowi6_saddr),
		},
		{
			Name:  "offset_flowi6_daddr",
			Value: uint64(offsetGuess.offset_flowi6_daddr),
		},
		{
			Name:  "offset_flowi6_sport",
			Value: uint64(offsetGuess.offset_flowi6_sport),
		},
		{
			Name:  "offset_flowi6_dport",
			Value: uint64(offsetGuess.offset_flowi6_dport),
		},
		{
			Name:  "offset_sk_net",
			Value: uint64(offsetGuess.offset_sk_net),
		},
		{
			Name:  "offset_ns_common_inum",
			Value: uint64(offsetGuess.offset_ns_common_inum),
		},
		{
			Name:  "offset_socket_sk",
			Value: uint64(offsetGuess.offset_socket_sk),
		},
	}
}

func NewConstEditorTCPSeq(offset *OffsetTCPSeqC) []bpfutil.ConstantPatch {
	return []bpfutil.ConstantPatch{
		{
			Name:  "offset_copied_seq",
			Value: uint64(offset.offset_copied_seq),
		},
		{
			Name:  "offset_write_seq",
			Value: uint64(offset.offset_write_seq),
		},
	}
}

func SetTCPSeqOffset(dst *OffsetGuessC, src *OffsetTCPSeqC) {
	dst.offset_copied_seq = _Ctype_ulonglong(src.offset_copied_seq)
	dst.offset_write_seq = _Ctype_ulonglong(src.offset_write_seq)
}

func GetTCPSeqOffset(offset *OffsetGuessC) *OffsetTCPSeqC {
	return &OffsetTCPSeqC{
		offset_copied_seq: _Ctype_int(offset.offset_copied_seq),
		offset_write_seq:  _Ctype_int(offset.offset_write_seq),
	}
}

func kernelGuessNeedsTCP4(offset *OffsetGuessC) bool {
	if offset == nil {
		return true
	}
	return offset.offset_inet_sport == 0 ||
		offset.offset_sk_dport == 0 ||
		offset.offset_tcp_sk_srtt_us == 0 ||
		offset.offset_tcp_sk_mdev_us == 0 ||
		offset.offset_sk_daddr == 0 ||
		offset.offset_sk_net == 0 ||
		offset.offset_ns_common_inum == 0 ||
		offset.offset_sk_family == 0
}

func kernelGuessNeedsSocket(offset *OffsetGuessC) bool {
	if offset == nil {
		return true
	}
	return offset.offset_socket_sk == 0
}

func kernelGuessNeedsUDP4(offset *OffsetGuessC) bool {
	if offset == nil {
		return true
	}
	return offset.offset_flowi4_daddr == 0 ||
		offset.offset_flowi4_saddr == 0 ||
		offset.offset_flowi4_dport == 0 ||
		offset.offset_flowi4_sport == 0
}

func kernelGuessNeedsTCP6(offset *OffsetGuessC) bool {
	if offset == nil {
		return true
	}
	return offset.offset_sk_v6_daddr == 0 ||
		offset.offset_sk_v6_rcv_saddr == 0 ||
		offset.offset_sk_family == 0
}

// GuessOffset guess the offset of the structure field, such as tcp_sock.srtt_us.
func GuessOffset(rt *bpfutil.Runtime, guessed *OffsetGuessC, ipv6Disabled bool) (*OffsetGuessC, error) {
	goruntime.LockOSThread()
	defer goruntime.UnlockOSThread()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ebpfMapGuess, err := BpfMapGuessInit(rt)
	if err != nil {
		return nil, err
	}

	var s syscall.Stat_t
	var netns uint32 = 0
	if err := syscall.Stat("/proc/self/ns/net", &s); err != nil {
		l.Error(err)
	} else {
		netns = uint32(s.Ino)
	}
	status := newGuessStatus()
	if guessed != nil {
		copyOffset(guessed, &status)
		copySupplementalOffsets(guessed, &status)
	}
	status.pid_tgid = _Ctype_ulonglong(uint64(unix.Getpid())<<32 | uint64(unix.Gettid()))

	needTCP4 := kernelGuessNeedsTCP4(&status) || kernelGuessNeedsSocket(&status)
	needUDP4 := kernelGuessNeedsUDP4(&status)
	needTCP6 := !ipv6Disabled && kernelGuessNeedsTCP6(&status)

	var (
		serverAddr    string
		serverAddrUDP string
		serverAddr6   string
		conninfo      Conninfo
		conninfoUDP   Conninfo
		conninfo6     Conninfo
	)

	if needTCP4 {
		tcp4ServerPort, err := runTCPServer(ctx, "tcp4", listenIPv4)
		if err != nil {
			return nil, err
		}
		serverAddr = fmt.Sprintf("%s:%d", listenIPv4, tcp4ServerPort)
		conninfo = Conninfo{
			Dport: tcp4ServerPort,
			Daddr: listenIPv4Arr,
			Meta:  ConnL4TCP | ConnL3IPv4,
			NetNS: netns,
		}
		status.meta = _Ctype_uint(conninfo.Meta)
	}

	if needUDP4 {
		udp4ServerPort, err := runUDPServer(ctx, "udp4", listenIPv4)
		if err != nil {
			return nil, err
		}
		serverAddrUDP = fmt.Sprintf("%s:%d", listenIPv4, udp4ServerPort)
		conninfoUDP = Conninfo{
			Dport: udp4ServerPort,
			Daddr: listenIPv4Arr,
			Meta:  ConnL4UDP | ConnL3IPv4,
		}
	}

	if needTCP6 {
		daddr6, ip6 := generateRandomIPv6Address()
		conninfo6 = Conninfo{
			Dport: 57391,
			Daddr: daddr6,
			Meta:  ConnL3IPv6 | ConnL4TCP,
		}
		serverAddr6 = fmt.Sprintf("[%s]:%d", ip6.String(), conninfo6.Dport)
	}

	if err := updateMapGuessStatus(ebpfMapGuess, &status); err != nil {
		return nil, err
	}

	offsetCheck := OffsetCheck{}
	requiredReadyRound := -1
	for round := 0; round < maxKernelGuessRounds; round++ {
		if needTCP4 {
			err := guessTCP4(serverAddr, conninfo, ebpfMapGuess, &offsetCheck, &status)
			if err != nil {
				return nil, err
			}
		}

		if needTCP6 {
			err = guessTCP6(serverAddr6, conninfo6, ebpfMapGuess, &offsetCheck, &status)
			if err != nil {
				return nil, err
			}
		}

		if needUDP4 {
			err = guessUDP4(serverAddrUDP, conninfoUDP, ebpfMapGuess, &offsetCheck, &status)
			if err != nil {
				return nil, err
			}
		}

		if kernelGuessRequiredReady(&offsetCheck, needTCP4, needUDP4, needTCP6) {
			if requiredReadyRound < 0 {
				requiredReadyRound = round
			}

			if kernelGuessOptionalReady(&offsetCheck, needTCP4, needTCP6) ||
				round-requiredReadyRound >= kernelGuessOptionalSlackRounds {
				newStatus := finalizeKernelGuess(&status)
				if !kernelGuessOptionalReady(&offsetCheck, needTCP4, needTCP6) {
					l.Warnf(
						"kernel offset guess returning with unresolved optional fields "+
							"after %d rounds: progress[%s] "+
							"status[inet_sport=%d tcp_srtt=%d tcp_mdev=%d socket_sk=%d sk_v6_daddr=%d]",
						round+1,
						kernelGuessProgress(&offsetCheck),
						newStatus.offset_inet_sport,
						newStatus.offset_tcp_sk_srtt_us,
						newStatus.offset_tcp_sk_mdev_us,
						newStatus.offset_socket_sk,
						newStatus.offset_sk_v6_daddr,
					)
				}
				return newStatus, nil
			}
		}
	}

	l.Warnf(
		"kernel offset guess exhausted after %d rounds: progress[%s] "+
			"status[state=%d err=%s(%d) pid_tgid=%d proc=%q "+
			"inet_sport=%d sk_dport=%d sk_daddr=%d sk_family=%d sk_net=%d "+
			"ns_inum=%d flowi4_saddr=%d flowi4_daddr=%d flowi4_dport=%d socket_sk=%d]",
		maxKernelGuessRounds,
		kernelGuessProgress(&offsetCheck),
		status.state,
		guessErrorString(int64(status.err)),
		status.err,
		status.pid_tgid,
		guessProcessName(&status),
		status.offset_inet_sport,
		status.offset_sk_dport,
		status.offset_sk_daddr,
		status.offset_sk_family,
		status.offset_sk_net,
		status.offset_ns_common_inum,
		status.offset_flowi4_saddr,
		status.offset_flowi4_daddr,
		status.offset_flowi4_dport,
		status.offset_socket_sk,
	)

	return nil, fmt.Errorf("guess kernel offsets: exceeded %d rounds", maxKernelGuessRounds)
}

func guessTCP4(serverAddr string, conninfo Conninfo, ebpfMapGuess *ebpf.Map,
	offsetCheck *OffsetCheck, status *OffsetGuessC,
) error {
	if offsetCheck.tcpSkSrttUsOk > MINSUCCESS &&
		offsetCheck.tcpSkMdevUsOk > MINSUCCESS &&
		offsetCheck.inetSportOk > MINSUCCESS &&
		offsetCheck.skDportOk > MINSUCCESS &&
		offsetCheck.skDaddrOk > MINSUCCESS &&
		offsetCheck.skFamilyOk > MINSUCCESS &&
		offsetCheck.netnsInumOk > MINSUCCESS &&
		offsetCheck.sknetOk > MINSUCCESS &&
		offsetCheck.socketSkOK > MINSUCCESS {
		return nil
	}

	status.conn_type = ConnL3IPv4 | ConnL4TCP
	if err := updateMapGuessStatus(ebpfMapGuess, status); err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 15)
	conn, err := net.Dial("tcp4", serverAddr)
	if err != nil {
		return err
	}
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("conv conn to tcp conn")
	}

	sport, err := strconv.Atoi(strings.Split(tcpConn.LocalAddr().String(), ":")[1])
	if err != nil {
		return err
	}
	conninfo.Sport = uint16(sport)
	connFile, err := tcpConn.File()
	if err != nil {
		return err
	}

	tcpInfo, err := unix.GetsockoptTCPInfo(int(connFile.Fd()), syscall.SOL_TCP, syscall.TCP_INFO)
	if err != nil {
		return err
	}
	conninfo.Rtt = tcpInfo.Rtt
	conninfo.RttVar = tcpInfo.Rttvar

	if err = connFile.Close(); err != nil {
		return err
	}
	if err = conn.Close(); err != nil {
		return err
	}

	statusAct, err := readMapGuessStatus(ebpfMapGuess)
	if err != nil {
		return err
	}
	if statusAct.state == 0 { // lost
		l.Warnf(
			"kernel offset guess tcp4 probe missed event: pid_tgid=%d proc=%q conn=%s "+
				"err=%s(%d) offsets[inet_sport=%d sk_dport=%d sk_daddr=%d "+
				"sk_family=%d sk_net=%d ns_inum=%d socket_sk=%d]",
			statusAct.pid_tgid,
			guessProcessName(statusAct),
			serverAddr,
			guessErrorString(int64(statusAct.err)),
			statusAct.err,
			status.offset_inet_sport,
			status.offset_sk_dport,
			status.offset_sk_daddr,
			status.offset_sk_family,
			status.offset_sk_net,
			status.offset_ns_common_inum,
			status.offset_socket_sk,
		)
		time.Sleep(time.Millisecond * 20)
		return nil
	}

	tryGuess(statusAct, offsetCheck, &conninfo, GUESS_INET_SPORT)
	tryGuess(statusAct, offsetCheck, &conninfo, GUESS_SK_DPORT)
	tryGuess(statusAct, offsetCheck, &conninfo, GUESS_TCP_SK_SRTT_US)
	tryGuess(statusAct, offsetCheck, &conninfo, GUESS_TCP_SK_MDEV_US)
	tryGuess(statusAct, offsetCheck, &conninfo, GUESS_SK_DADDR)
	tryGuess(statusAct, offsetCheck, &conninfo, GUESS_NS_COMMON_INUM)
	tryGuess(statusAct, offsetCheck, &conninfo, GUESS_SK_FAMILY)
	tryGuess(statusAct, offsetCheck, &conninfo, GUESS_SOCKET_SK)

	copyOffset(statusAct, status)
	if status.offset_tcp_sk_srtt_us > MAXOFFSET ||
		status.offset_tcp_sk_mdev_us > MAXOFFSET ||
		status.offset_inet_sport > MAXOFFSET ||
		status.offset_sk_dport > MAXOFFSET ||
		status.offset_socket_sk > MAXOFFSET ||
		status.offset_sk_daddr > MAXOFFSET ||
		status.offset_sk_net > MAXOFFSET ||
		status.offset_ns_common_inum > MAXOFFSET ||
		status.offset_sk_family > MAXOFFSET {
		l.Error(status)
		return fmt.Errorf("guess tcp4: offset > MAXOFFSET")
	}

	if err = updateMapGuessStatus(ebpfMapGuess, status); err != nil {
		return err
	}

	return nil
}

func guessTCP6(serverAddr6 string, conninfo6 Conninfo, ebpfMapGuess *ebpf.Map,
	offsetCheck *OffsetCheck, status *OffsetGuessC,
) error {
	if offsetCheck.skV6DaddrOk > MINSUCCESS && offsetCheck.skFamilyOk > MINSUCCESS {
		return nil
	}
	status.conn_type = ConnL3IPv6 | ConnL4TCP
	if err := updateMapGuessStatus(ebpfMapGuess, status); err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 10)

	if conn, err := net.DialTimeout("tcp6", serverAddr6, time.Millisecond*10); err == nil {
		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			return fmt.Errorf("conv conn to tcp conn")
		}
		if err := tcpConn.SetLinger(0); err != nil {
			return err
		}
		if err := conn.Close(); err != nil {
			return err
		}
	}

	var err error
	statusAct, err := readMapGuessStatus(ebpfMapGuess)
	if err != nil {
		return err
	}

	if statusAct.state == 0 { // lost
		l.Warnf(
			"kernel offset guess tcp6 probe missed event: pid_tgid=%d proc=%q conn=%s err=%s(%d) offsets[sk_v6_daddr=%d sk_family=%d]",
			statusAct.pid_tgid,
			guessProcessName(statusAct),
			serverAddr6,
			guessErrorString(int64(statusAct.err)),
			statusAct.err,
			status.offset_sk_v6_daddr,
			status.offset_sk_family,
		)
		time.Sleep(time.Millisecond * 20)
		return nil
	}
	tryGuess(statusAct, offsetCheck, &conninfo6, GUESS_SK_V6_DADDR)
	tryGuess(statusAct, offsetCheck, &conninfo6, GUESS_SK_FAMILY)
	copyOffset(statusAct, status)
	if status.offset_sk_v6_daddr > MAXOFFSET ||
		status.offset_sk_family > MAXOFFSET {
		l.Error(status)
		return fmt.Errorf("guesss tcp6: offset > MAXOFFSET")
	}
	if err = updateMapGuessStatus(ebpfMapGuess, status); err != nil {
		return err
	}

	return nil
}

func guessUDP4(serverAddrUDP string, conninfoUDP Conninfo, ebpfMapGuess *ebpf.Map,
	offsetCheck *OffsetCheck, status *OffsetGuessC,
) error {
	if offsetCheck.flowi4DaddrOk > MINSUCCESS &&
		offsetCheck.flowi4SaddrOk > MINSUCCESS &&
		offsetCheck.flowi4DportOk > MINSUCCESS {
		return nil
	}
	status.conn_type = ConnL3IPv4 | ConnL4UDP
	if err := updateMapGuessStatus(ebpfMapGuess, status); err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 10)
	conn, err := net.Dial("udp4", serverAddrUDP)
	if err != nil {
		return err
	}

	srcIP := net.ParseIP(strings.Split(conn.LocalAddr().String(), ":")[0])
	if srcIP != nil {
		srcIP = srcIP.To16()
	}
	if srcIP == nil {
		return fmt.Errorf("src ip: %s", conn.LocalAddr().String())
	}

	ip4arr := *(*[4]uint32)(unsafe.Pointer(&srcIP[0])) //nolint:gosec
	conninfoUDP.Saddr = [4]uint32{0, 0, 0, ip4arr[3]}
	_, err = conn.Write([]byte("guess flowi4"))
	if err != nil {
		return err
	}
	if err = conn.Close(); err != nil {
		return err
	}
	statusAct, err := readMapGuessStatus(ebpfMapGuess)
	if err != nil {
		return err
	}
	if statusAct.state == 0 { // lost
		l.Warnf(
			"kernel offset guess udp4 probe missed event: pid_tgid=%d proc=%q conn=%s err=%s(%d) offsets[flowi4_saddr=%d flowi4_daddr=%d flowi4_dport=%d]",
			statusAct.pid_tgid,
			guessProcessName(statusAct),
			serverAddrUDP,
			guessErrorString(int64(statusAct.err)),
			statusAct.err,
			status.offset_flowi4_saddr,
			status.offset_flowi4_daddr,
			status.offset_flowi4_dport,
		)
		time.Sleep(time.Millisecond * 20)
		return nil
	}

	tryGuess(statusAct, offsetCheck, &conninfoUDP, GUESS_FLOWI4_DADDR)
	tryGuess(statusAct, offsetCheck, &conninfoUDP, GUESS_FLOWI4_SADDR)
	tryGuess(statusAct, offsetCheck, &conninfoUDP, GUESS_FLOWI4_DPORT)
	copyOffset(statusAct, status)
	if status.offset_flowi4_daddr > MAXOFFSET ||
		status.offset_flowi4_saddr > MAXOFFSET ||
		status.offset_flowi4_dport > MAXOFFSET {
		l.Error(status)
		return fmt.Errorf("guess upd4: offset > MAXOFFSET")
	}
	if err = updateMapGuessStatus(ebpfMapGuess, status); err != nil {
		return err
	}
	return nil
}

func runTCPServer(ctx context.Context, network, address string) (uint16, error) {
	netListen, err := net.Listen(network, address+":0")
	if err != nil {
		return 0, err
	}

	addr := netListen.Addr().String()

	l.Debug("start the tcp server to guess the offset: ", addr)
	var serverPort int
	if addr[:1] == "[" {
		serverPort, err = strconv.Atoi(strings.Split(addr, "]")[1][1:])
	} else {
		serverPort, err = strconv.Atoi(strings.Split(addr, ":")[1])
	}
	if err != nil {
		return 0, err
	}

	go func() {
		<-ctx.Done()
		if err := netListen.Close(); err != nil {
			l.Error(err)
		}
	}()

	go func() {
		for {
			conn, err := netListen.Accept()
			if err != nil {
				return
			}

			if tp, ok := conn.(*net.TCPConn); ok {
				// send RST to avoid generating a lot of TIME_WAIT
				if err := tp.SetLinger(0); err != nil {
					return
				}
				// wait for the client to send fin
				_, _ = io.Copy(io.Discard, tp)
			}

			if err = conn.Close(); err != nil {
				l.Error(err)
				return
			}
		}
	}()

	return uint16(serverPort), nil
}

func runUDPServer(ctx context.Context, network, addr string) (uint16, error) {
	netListen, err := net.ListenPacket(network, addr+":0")
	if err != nil {
		return 0, err
	}
	localAddr := netListen.LocalAddr().String()
	var serverPort int
	if localAddr[:1] == "[" {
		serverPort, err = strconv.Atoi(strings.Split(localAddr, "]")[1][1:])
	} else {
		serverPort, err = strconv.Atoi(strings.Split(localAddr, ":")[1])
	}
	if err != nil {
		return 0, err
	}
	go func() {
		<-ctx.Done()
		if err := netListen.Close(); err != nil {
			l.Error(err)
		}
	}()

	go func() {
		for {
			p := []byte{}
			err := netListen.SetReadDeadline(time.Now().Add(time.Microsecond * 20))
			if err != nil {
				l.Error(err)
			}
			_, _, err = netListen.ReadFrom(p)
			if err != nil && !os.IsTimeout(err) {
				return
			}
		}
	}()

	return uint16(serverPort), nil
}

func DumpOffset(dir string, offset *OffsetGuessC) error {
	dirpath := filepath.Join(dir, "externals")
	offsetPath := filepath.Join(dirpath, "datakit-ebpf.offset")

	offsetStr, err := dumpOffset(*offset)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dirpath, 0o750); err != nil {
		return err
	}

	if existing, err := os.ReadFile(offsetPath); err == nil && string(existing) == offsetStr { //nolint:gosec
		return nil
	}

	fp, err := os.CreateTemp(dirpath, "datakit-ebpf.offset.*")
	if err != nil {
		return err
	}
	tmpPath := fp.Name()
	defer func() {
		if err := os.Remove(tmpPath); err != nil && !os.IsNotExist(err) {
			l.Error(err)
		}
	}()

	if _, err := fp.Write([]byte(offsetStr)); err != nil {
		if closeErr := fp.Close(); closeErr != nil {
			l.Error(closeErr)
		}
		return err
	}
	if err := fp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, offsetPath); err != nil {
		return err
	}
	return nil
}

func LoadOffset(dir string) (*OffsetGuessC, error) {
	dirpath := filepath.Join(dir, "externals")
	offsetPath := filepath.Join(dirpath, "datakit-ebpf.offset")
	offsetByte, err := os.ReadFile(offsetPath) //nolint:gosec
	if err != nil {
		return nil, err
	}
	offset, err := loadOffset(string(offsetByte))
	if err != nil {
		return nil, err
	}
	return &offset, err
}
