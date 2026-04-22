//go:build linux
// +build linux

package offset

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"os"
	goruntime "runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/cilium/ebpf"
	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/c"
	dkconntrack "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/conntrack"
	"golang.org/x/sys/unix"
)

const (
	maxConntrackGuessRounds      = 256
	conntrackInactiveProbeRounds = 8
	conntrackGuessRemotePort     = 33434
	conntrackNetnsSeedWindow     = 16
	conntrackModernKernel        = uint64(0x0005000000000000)
	conntrackLegacyKernel        = uint64(0x0004000f00000000)
	conntrackTupleOriginSeed     = uint64(32)
	conntrackTupleReplyDelta     = uint64(56)
)

func conntrackSeedOffsets(kernelVersion uint64) (ctNet, nsInum uint64) {
	switch {
	case kernelVersion >= conntrackModernKernel:
		return 136, 168
	case kernelVersion >= conntrackLegacyKernel:
		return 17, 112
	default:
		return 17, 112
	}
}

func conntrackSeedTupleOffsets() (origin, reply uint64) {
	return conntrackTupleOriginSeed, conntrackTupleOriginSeed + conntrackTupleReplyDelta
}

func conntrackGuessTarget() (string, [4]uint32) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return listenIPv4, listenIPv4Arr
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}

			ip4 := ip.To4()
			if ip4 == nil || ip4.IsLoopback() {
				continue
			}

			var arr [4]uint32
			arr[3] = binary.LittleEndian.Uint32(ip4)
			return ip4.String(), arr
		}
	}

	return listenIPv4, listenIPv4Arr
}

func ipv4ToConnaddr(ip net.IP) [4]uint32 {
	ip4 := ip.To4()
	if ip4 == nil {
		return [4]uint32{}
	}

	var arr [4]uint32
	arr[3] = binary.LittleEndian.Uint32(ip4)
	return arr
}

func conntrackGuessRemoteTarget() (string, [4]uint32, uint16, bool) {
	data, err := os.ReadFile("/proc/net/route")
	if err != nil {
		return "", [4]uint32{}, 0, false
	}

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 || fields[1] != "00000000" {
			continue
		}

		gw, err := strconv.ParseUint(fields[2], 16, 32)
		if err != nil || gw == 0 {
			continue
		}

		var raw [4]byte
		binary.LittleEndian.PutUint32(raw[:], uint32(gw))
		ip := net.IPv4(raw[0], raw[1], raw[2], raw[3]).To4()
		if ip == nil || ip.IsLoopback() || ip.Equal(net.IPv4zero) {
			continue
		}

		return ip.String(), ipv4ToConnaddr(ip), conntrackGuessRemotePort, true
	}

	return "", [4]uint32{}, 0, false
}

func newOffsetConntrackRuntime() (*bpfutil.Runtime, error) {
	hooks, err := dkconntrack.ResolveConntrackHookSelection()
	if err != nil {
		return nil, err
	}
	if hooks.InsertSymbol == "" {
		return nil, fmt.Errorf("conntrack insert kernel symbols unavailable")
	}
	insertPrograms := dkconntrack.ConntrackInsertProgramNames(hooks.InsertSymbols)
	if len(insertPrograms) == 0 {
		return nil, fmt.Errorf("conntrack insert programs unavailable")
	}

	m := &bpfutil.Runtime{
		Probes: make([]*bpfutil.HookSpec, 0, len(insertPrograms)),
	}
	for _, insertProgram := range insertPrograms {
		m.Probes = append(m.Probes, &bpfutil.HookSpec{
			ID: bpfutil.HookID{
				Program: insertProgram,
			},
		})
	}
	loadSpec := bpfutil.LoadSpec{
		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},
	}
	buf, err := dkebpf.OffsetConntrackBin()
	if err != nil {
		return nil, fmt.Errorf("offset_conntrack.o: %w", err)
	}

	if err := m.LoadFromReader((bytes.NewReader(buf)), loadSpec); err != nil {
		return nil, fmt.Errorf("init offset conntrack guess: %w", err)
	}
	return m, nil
}

func bpfMapGuessConntrackInit(runtime *bpfutil.Runtime) (*ebpf.Map, error) {
	bpfmapOffsetConntrack, err := runtime.LookupMap("bpfmap_offset_conntrack")
	if err != nil {
		return nil, fmt.Errorf("lookup bpf map bpfmap_offset_conntrack: %w", err)
	}

	zero := uint64(0)
	status := newGuessConntrack()
	if err := bpfmapOffsetConntrack.Update(zero, unsafe.Pointer(&status), //nolint:gosec
		ebpf.UpdateAny); err != nil {
		return nil, err
	}
	time.Sleep(time.Millisecond * 5)
	return bpfmapOffsetConntrack, nil
}

func GuessOffsetConntrack(guessed *OffsetConntrackC) ([]bpfutil.ConstantPatch, *OffsetConntrackC, error) {
	goruntime.LockOSThread()
	defer goruntime.UnlockOSThread()

	rt, err := newOffsetConntrackRuntime()
	if err != nil {
		return nil, nil, err
	}

	if err := rt.StartRuntime(); err != nil {
		return nil, nil, err
	}

	defer rt.Shutdown() //nolint:errcheck

	bpfmap, err := bpfMapGuessConntrackInit(rt)
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serverAddr := ""

	var s syscall.Stat_t
	var netns uint32 = 0
	if err := syscall.Stat("/proc/self/ns/net", &s); err != nil {
		l.Error(err)
	} else {
		netns = uint32(s.Ino)
	}

	var conninfo Conninfo
	guessWithUDP := false
	if remoteAddr, remoteArr, remotePort, ok := conntrackGuessRemoteTarget(); ok {
		serverAddr = fmt.Sprintf("%s:%d", remoteAddr, remotePort)
		conninfo = Conninfo{
			Dport: remotePort,
			Daddr: remoteArr,
			Meta:  ConnL4UDP | ConnL3IPv4,
			NetNS: netns,
		}
		guessWithUDP = true
	} else {
		guessListenIPv4, guessListenIPv4Arr := conntrackGuessTarget()
		tcp4ServerPort, err := runTCPServer(ctx, "tcp4", guessListenIPv4)
		if err != nil {
			return nil, nil, err
		}
		serverAddr = fmt.Sprintf("%s:%d", guessListenIPv4, tcp4ServerPort)
		conninfo = Conninfo{
			Dport: tcp4ServerPort,
			Daddr: guessListenIPv4Arr,
			Meta:  ConnL4TCP | ConnL3IPv4,
			NetNS: netns,
		}
	}

	status := newGuessConntrack()
	if guessed != nil {
		copyOffsetCT(guessed, &status)
	}
	seedOrigin, seedReply := conntrackSeedTupleOffsets()
	if uint64(status.offset_ct_origin_tuple) == 0 {
		status.offset_ct_origin_tuple = _Ctype_ulonglong(seedOrigin)
	}
	if uint64(status.offset_ct_reply_tuple) == 0 {
		status.offset_ct_reply_tuple = _Ctype_ulonglong(seedReply)
	}
	if kernelVersion, err := bpfutil.CurrentKernelVersion(); err == nil {
		seedCTNet, seedNSInum := conntrackSeedOffsets(kernelVersion)
		if uint64(status.offset_ct_net) == 0 ||
			(kernelVersion >= conntrackModernKernel && uint64(status.offset_ct_net) < seedCTNet/2) {
			status.offset_ct_net = _Ctype_ulonglong(seedCTNet)
		}
		if uint64(status.offset_ct_ns_common_inum) == 0 ||
			(kernelVersion >= conntrackModernKernel && uint64(status.offset_ct_ns_common_inum) < 128) {
			status.offset_ct_ns_common_inum = _Ctype_ulonglong(seedNSInum)
		}
	}
	seedNSCommonInum := uint64(status.offset_ct_ns_common_inum)
	offsetCheck := OffsetCheck{}
	zeroStateRounds := 0
	var (
		lastState  uint64 = ^uint64(0)
		lastOrigin uint64 = ^uint64(0)
		lastReply  uint64 = ^uint64(0)
		lastCTNet  uint64 = ^uint64(0)
		lastNSInum uint64 = ^uint64(0)
		lastErr    int64  = -1
	)
	for round := 0; round < maxConntrackGuessRounds; round++ {
		if seedNSCommonInum != 0 && status.offset_ct_ns_common_inum == 0 {
			status.offset_ct_ns_common_inum = _Ctype_ulonglong(seedNSCommonInum)
		}

		var err error
		if guessWithUDP {
			err = guessConntrackUDP(serverAddr, conninfo, bpfmap, &offsetCheck, &status)
		} else {
			err = guessConntrack(serverAddr, conninfo, bpfmap, &offsetCheck, &status)
		}
		if err != nil {
			return nil, nil, err
		}
		if seedNSCommonInum != 0 &&
			offsetCheck.ctNetOk == 0 &&
			offsetCheck.netnsInumOk == 0 &&
			uint64(status.offset_ct_ns_common_inum) > seedNSCommonInum+conntrackNetnsSeedWindow {
			status.offset_ct_net++
			status.offset_ct_ns_common_inum = _Ctype_ulonglong(seedNSCommonInum)
			l.Debugf(
				"conntrack offset guess advance ct_net after netns seed window miss: round=%d next_ct_net=%d reset_ns_inum=%d",
				round+1,
				uint64(status.offset_ct_net),
				seedNSCommonInum,
			)
		}
		if status.state == 0 {
			zeroStateRounds++
			if zeroStateRounds >= conntrackInactiveProbeRounds {
				return nil, nil, fmt.Errorf("conntrack insert hooks inactive for probe traffic after %d rounds", zeroStateRounds)
			}
		} else {
			zeroStateRounds = 0
		}
		if uint64(status.state) != lastState ||
			uint64(status.offset_ct_origin_tuple) != lastOrigin ||
			uint64(status.offset_ct_reply_tuple) != lastReply ||
			uint64(status.offset_ct_net) != lastCTNet ||
			uint64(status.offset_ct_ns_common_inum) != lastNSInum ||
			int64(status.err) != lastErr ||
			round%16 == 15 {
			l.Debugf(
				"conntrack offset guess progress: round=%d state=%d err=%s(%d) "+
					"origin=%d reply=%d ct_net=%d ns_inum=%d "+
					"hits[origin=%d reply=%d ct_net=%d netns=%d]",
				round+1,
				uint64(status.state),
				guessErrorString(int64(status.err)),
				int64(status.err),
				uint64(status.offset_ct_origin_tuple),
				uint64(status.offset_ct_reply_tuple),
				uint64(status.offset_ct_net),
				uint64(status.offset_ct_ns_common_inum),
				offsetCheck.ctOriginTupleOk,
				offsetCheck.ctReplyTupleOk,
				offsetCheck.ctNetOk,
				offsetCheck.netnsInumOk,
			)
			lastState = uint64(status.state)
			lastOrigin = uint64(status.offset_ct_origin_tuple)
			lastReply = uint64(status.offset_ct_reply_tuple)
			lastCTNet = uint64(status.offset_ct_net)
			lastNSInum = uint64(status.offset_ct_ns_common_inum)
			lastErr = int64(status.err)
		}

		if offsetCheck.ctOriginTupleOk > MINSUCCESS &&
			offsetCheck.ctNetOk > MINSUCCESS &&
			offsetCheck.netnsInumOk > MINSUCCESS {
			replyConfirmed := offsetCheck.ctReplyTupleOk > MINSUCCESS
			replyDerived := uint64(status.offset_ct_reply_tuple) ==
				uint64(status.offset_ct_origin_tuple)+conntrackTupleReplyDelta
			if !replyConfirmed && !replyDerived {
				continue
			}
			newstatus := newGuessConntrack()
			copyOffsetCT(&status, &newstatus)
			if !replyConfirmed && replyDerived {
				l.Infof("conntrack reply tuple offset derived from origin offset: origin=%d reply=%d",
					uint64(newstatus.offset_ct_origin_tuple),
					uint64(newstatus.offset_ct_reply_tuple),
				)
			}
			l.Infof("conntrack offsets guessed: origin=%d reply=%d ct_net=%d ns_inum=%d",
				uint64(newstatus.offset_ct_origin_tuple),
				uint64(newstatus.offset_ct_reply_tuple),
				uint64(newstatus.offset_ct_net),
				uint64(newstatus.offset_ct_ns_common_inum),
			)

			return newConntrackConstEditor(&newstatus), &newstatus, nil
		}
	}

	l.Warnf(
		"conntrack offset guess exhausted after %d rounds: state=%d err=%s(%d) "+
			"offsets[origin=%d reply=%d ct_net=%d ns_inum=%d] "+
			"hits[origin=%d reply=%d ct_net=%d netns=%d]",
		maxConntrackGuessRounds,
		uint64(status.state),
		guessErrorString(int64(status.err)),
		int64(status.err),
		uint64(status.offset_ct_origin_tuple),
		uint64(status.offset_ct_reply_tuple),
		uint64(status.offset_ct_net),
		uint64(status.offset_ct_ns_common_inum),
		offsetCheck.ctOriginTupleOk,
		offsetCheck.ctReplyTupleOk,
		offsetCheck.ctNetOk,
		offsetCheck.netnsInumOk,
	)

	return nil, nil, fmt.Errorf("guess conntrack offsets: exceeded %d rounds", maxConntrackGuessRounds)
}

func guessConntrack(svc string, conninfo Conninfo, ebpfMap *ebpf.Map,
	offsetCk *OffsetCheck, status *OffsetConntrackC,
) error {
	if err := updateMapConntrack(ebpfMap, status); err != nil {
		return err
	}

	conn, err := net.Dial("tcp4", svc)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("conv conn to tcp conn")
	}

	defer tcpConn.Close() //nolint:errcheck

	if err := tcpConn.SetLinger(0); err != nil {
		return fmt.Errorf(err.Error())
	}

	clientAddr := strings.Split(tcpConn.LocalAddr().String(), ":")
	sport, err := strconv.Atoi(clientAddr[1])
	if err != nil {
		return err
	}
	saddr := net.ParseIP(clientAddr[0]).To4()
	conninfo.Saddr = [4]uint32{0}
	if len(saddr) == 4 {
		for i := range saddr {
			conninfo.Saddr[3] = conninfo.Saddr[3]<<8 + uint32(saddr[3-i])
		}
	}

	conninfo.Sport = uint16(sport)

	statusAct, err := readMapGuessConntrack(ebpfMap)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	if statusAct.state == 0 { // lost
		copyOffsetCTRuntime(statusAct, status)
		copyOffsetCT(statusAct, status)
		return nil
	}

	tryGuessConntrack(statusAct, offsetCk, &conninfo, GUESS_CONNTRACK_TUPLE_ORIGIN)
	tryGuessConntrack(statusAct, offsetCk, &conninfo, GUESS_CONNTRACK_TUPLE_REPLY)
	tryGuessConntrack(statusAct, offsetCk, &conninfo, GUESS_NS_COMMON_INUM)

	if status.offset_ct_origin_tuple > 512 ||
		status.offset_ct_reply_tuple > 512 ||
		status.offset_ct_net > 512 {
		return fmt.Errorf("guess conntrack: offset > 512")
	}

	copyOffsetCTRuntime(statusAct, status)
	copyOffsetCT(statusAct, status)

	return nil
}

func guessConntrackUDP(svc string, conninfo Conninfo, ebpfMap *ebpf.Map,
	offsetCk *OffsetCheck, status *OffsetConntrackC,
) error {
	if err := updateMapConntrack(ebpfMap, status); err != nil {
		return err
	}

	conn, err := net.Dial("udp4", svc)
	if err != nil {
		return err
	}
	defer conn.Close() //nolint:errcheck

	clientAddr := conn.LocalAddr().(*net.UDPAddr) //nolint:forcetypeassert
	conninfo.Saddr = ipv4ToConnaddr(clientAddr.IP)
	conninfo.Sport = uint16(clientAddr.Port)

	if _, err := conn.Write([]byte("guess conntrack")); err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 10)

	statusAct, err := readMapGuessConntrack(ebpfMap)
	if err != nil {
		return err
	}

	if statusAct.state == 0 { // lost
		copyOffsetCTRuntime(statusAct, status)
		copyOffsetCT(statusAct, status)
		return nil
	}

	tryGuessConntrack(statusAct, offsetCk, &conninfo, GUESS_CONNTRACK_TUPLE_ORIGIN)
	tryGuessConntrack(statusAct, offsetCk, &conninfo, GUESS_CONNTRACK_TUPLE_REPLY)
	tryGuessConntrack(statusAct, offsetCk, &conninfo, GUESS_NS_COMMON_INUM)

	if status.offset_ct_origin_tuple > 512 ||
		status.offset_ct_reply_tuple > 512 ||
		status.offset_ct_net > 512 {
		return fmt.Errorf("guess conntrack: offset > 512")
	}

	copyOffsetCTRuntime(statusAct, status)
	copyOffsetCT(statusAct, status)

	return nil
}

func updateMapConntrack(m *ebpf.Map, status *OffsetConntrackC) error {
	var key uint64 = 0
	status.origin = _Ctype_struct_nf_conn_tuple{}
	status.reply = _Ctype_struct_nf_conn_tuple{}
	status.err = ERR_G_NOERROR
	status.state = 0
	return m.Update(&key, status, ebpf.UpdateAny)
}

func newConntrackConstEditor(offset *OffsetConntrackC) []bpfutil.ConstantPatch {
	return []bpfutil.ConstantPatch{
		{
			Name:  "offset_ct_net",
			Value: uint64(offset.offset_ct_net),
		},
		{
			Name:  "offset_ct_ns_common_inum",
			Value: uint64(offset.offset_ct_ns_common_inum),
		},
		{
			Name:  "offset_ct_origin_tuple",
			Value: uint64(offset.offset_ct_origin_tuple),
		},
		{
			Name:  "offset_ct_reply_tuple",
			Value: uint64(offset.offset_ct_reply_tuple),
		},
	}
}
