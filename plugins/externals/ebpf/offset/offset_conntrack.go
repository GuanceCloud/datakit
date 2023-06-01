//go:build (linux && amd64 && ebpf) || (linux && arm64 && ebpf)
// +build linux,amd64,ebpf linux,arm64,ebpf

package offset

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"net"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/DataDog/ebpf"
	"github.com/DataDog/ebpf/manager"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/ebpf/c"
	"golang.org/x/sys/unix"
)

func newOffsetConntrackManger() (*manager.Manager, error) {
	m := &manager.Manager{
		Probes: []*manager.Probe{
			{
				Section: "kprobe/__nf_conntrack_hash_insert",
			},
		},
	}
	mOpts := manager.Options{
		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},
	}
	buf, err := dkebpf.OffsetConntrackBin()
	if err != nil {
		return nil, fmt.Errorf("offset_conntrack.o: %w", err)
	}

	if err := m.InitWithOptions((bytes.NewReader(buf)), mOpts); err != nil {
		return nil, fmt.Errorf("init offset conntrack guess: %w", err)
	}
	return m, nil
}

func bpfMapGuessConntrackInit(m *manager.Manager) (*ebpf.Map, error) {
	bpfmapOffsetConntrack, found, err := m.GetMap("bpfmap_offset_conntrack")
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fmt.Errorf("bpf map bpfmap_offset_conntrack not found")
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

func GuessOffsetConntrack(guessed *OffsetConntrackC) ([]manager.ConstantEditor, *OffsetConntrackC, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	bpfManager, err := newOffsetConntrackManger()
	if err != nil {
		return nil, nil, err
	}

	if err := bpfManager.Start(); err != nil {
		return nil, nil, err
	}

	defer bpfManager.Stop(manager.CleanAll) //nolint:errcheck

	bpfmap, err := bpfMapGuessConntrackInit(bpfManager)
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tcp4ServerPort, err := runTCPServer(ctx, "tcp4", listenIPv4)
	if err != nil {
		return nil, nil, err
	}

	serverAddr := fmt.Sprintf("%s:%d", listenIPv4, tcp4ServerPort)

	var s syscall.Stat_t
	var netns uint32 = 0
	if err := syscall.Stat("/proc/self/ns/net", &s); err != nil {
		l.Error(err)
	} else {
		netns = uint32(s.Ino)
	}
	conninfo := Conninfo{
		Dport: tcp4ServerPort,
		Daddr: listenIPv4Arr,
		Meta:  ConnL4TCP | ConnL3IPv4,
		NetNS: netns,
	}

	status := newGuessConntrack()
	if guessed != nil {
		copyOffsetCT(guessed, status)
	}
	offsetCheck := OffsetCheck{}
	for {
		if err := guessConntrack(serverAddr, conninfo, bpfmap, &offsetCheck, status); err != nil {
			return nil, nil, err
		}

		if offsetCheck.ctOriginTupleOk > MINSUCCESS &&
			offsetCheck.ctReplyTupleOk > MINSUCCESS &&
			offsetCheck.ctNetOk > MINSUCCESS &&
			offsetCheck.netnsInumOk > MINSUCCESS {
			newstatus := newGuessConntrack()
			copyOffsetCT(status, newstatus)

			return newConntrackConstEditor(newstatus), newstatus, nil
		}
	}
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
		l.Warn(statusAct.pid_tgid)
		time.Sleep(time.Millisecond * 20)
	}

	tryGuessConntrack(statusAct, offsetCk, &conninfo, GUESS_CONNTRACK_TUPLE_ORIGIN)
	tryGuessConntrack(statusAct, offsetCk, &conninfo, GUESS_CONNTRACK_TUPLE_REPLY)
	tryGuessConntrack(statusAct, offsetCk, &conninfo, GUESS_NS_COMMON_INUM)

	if status.offset_origin_tuple > 512 ||
		status.offset_reply_tuple > 512 ||
		status.offset_net > 512 {
		return fmt.Errorf("guess conntrack: offset > 512")
	}

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

func newConntrackConstEditor(offset *OffsetConntrackC) []manager.ConstantEditor {
	return []manager.ConstantEditor{
		{
			Name:  "offset_net",
			Value: uint64(offset.offset_net),
		},
		{
			Name:  "offset_ns_common_inum",
			Value: uint64(offset.offset_ns_common_inum),
		},
		{
			Name:  "offset_origin_tuple",
			Value: uint64(offset.offset_origin_tuple),
		},
		{
			Name:  "offset_reply_tuple",
			Value: uint64(offset.offset_reply_tuple),
		},
	}
}
