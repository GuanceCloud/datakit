//go:build linux
// +build linux

package offset

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"net"
	"reflect"
	goruntime "runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/cilium/ebpf"
	"github.com/google/gopacket/afpacket"
	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
	"golang.org/x/sys/unix"

	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/c"
)

func newOffsetTCPSeqRuntime(skFd int, cnt []bpfutil.ConstantPatch) (*bpfutil.Runtime, error) {
	useLegacyConsts, _, err := bpfutil.UseLegacyConstObjects()
	if err != nil {
		return nil, err
	}
	m := &bpfutil.Runtime{
		Probes: []*bpfutil.HookSpec{
			{
				ID: bpfutil.HookID{
					EBPFFuncName: "socket__packet_tcp_header",
				},
				SocketFD: skFd,
			},
			{
				ID: bpfutil.HookID{
					EBPFFuncName: "kprobe__tcp_getsockopt",
					UID:          "tcp_getsockopt_tcp_seq",
				},
			},
		},
	}

	loadSpec := bpfutil.LoadSpec{
		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},
		Constants:       cnt,
		LegacyConstants: useLegacyConsts,
	}

	bufLoader := dkebpf.OffsetTCPSeqBin
	binName := "offset_tcp_seq.o"
	if useLegacyConsts {
		bufLoader = dkebpf.OffsetTCPSeqLegacyBin
		binName = "offset_tcp_seq_legacy.o"
	}

	buf, err := bufLoader()
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", binName, err)
	}

	if err := m.LoadFromReader(bytes.NewReader(buf), loadSpec); err != nil {
		return nil, fmt.Errorf("init offset tcp seq guess: %w", err)
	}

	return m, nil
}

func bpfMapGuessTCPSeqInit(runtime *bpfutil.Runtime) (*ebpf.Map, error) {
	bpfmapTCPSeq, err := runtime.LookupMap("bpfmap_offset_tcp_seq")
	if err != nil {
		return nil, fmt.Errorf("lookup bpf map bpfmap_offset_tcp_seq: %w", err)
	}
	zero := uint64(0)
	status := newGuessTCPSeq()

	//nolint:gosec
	if err := bpfmapTCPSeq.Update(unsafe.Pointer(&zero), unsafe.Pointer(&status), ebpf.UpdateAny); err != nil {
		return nil, err
	}

	s := OffsetTCPSeqC{}
	//nolint:gosec
	if err := bpfmapTCPSeq.Lookup(unsafe.Pointer(&zero), unsafe.Pointer(&s)); err != nil {
		return nil, err
	}

	time.Sleep(time.Millisecond * 5)
	return bpfmapTCPSeq, nil
}

func updateSeqOffsetMap(m *ebpf.Map, status *OffsetTCPSeqC) error {
	key := uint64(0)
	// status.state = 0
	return m.Update(&key, status, ebpf.UpdateAny)
}

func readSeqOffset(m *ebpf.Map) (*OffsetTCPSeqC, error) {
	status := OffsetTCPSeqC{}
	var zero uint64 = 0

	//nolint:gosec
	if err := m.Lookup(&zero, unsafe.Pointer(&status)); err != nil {
		return nil, err
	} else {
		return &status, err
	}
}

func GuessOffsetTCPSeq(netflowOffset []bpfutil.ConstantPatch) ([]bpfutil.ConstantPatch, *OffsetTCPSeqC, error) {
	seqEditor, seqOffset, err := TryGetTCPSeqOffsetFromBTF()
	if err == nil && seqEditor != nil {
		l.Info("TCP seq offset obtained from BTF successfully")
		return seqEditor, seqOffset, nil
	}
	l.Debugf("get TCP seq offset from BTF failed: %v, fallback to guess", err)

	// current netns

	rawSocket, err := afpacket.NewTPacket()
	if err != nil {
		return nil, nil, fmt.Errorf("error creating raw socket: %w", err)
	}
	defer rawSocket.Close()

	// The underlying socket file descriptor is private, hence the use of reflection
	skFd := int(reflect.ValueOf(rawSocket).Elem().FieldByName("fd").Int())

	rt, err := newOffsetTCPSeqRuntime(skFd, netflowOffset)
	if err != nil {
		return nil, nil, err
	}

	if err := rt.StartRuntime(); err != nil {
		return nil, nil, err
	}

	defer rt.Shutdown() //nolint:errcheck

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tcp4ServerPort, err := runTCPServer(ctx, "tcp4", listenIPv4)
	if err != nil {
		return nil, nil, err
	}

	serverAddr := fmt.Sprintf("%s:%d", listenIPv4, tcp4ServerPort)

	goruntime.LockOSThread()
	defer goruntime.UnlockOSThread()

	bpfmapTCPSeq, err := rt.LookupMap("bpfmap_offset_tcp_seq")
	if err != nil {
		return nil, nil, fmt.Errorf("lookup bpf map bpfmap_offset_tcp_seq: %w", err)
	}

	_, err = bpfMapGuessTCPSeqInit(rt)
	if err != nil {
		return nil, nil, err
	}

	// stauts := newGuessTCPSeq()
	offset := OffsetTCPSeqC{}

	var okTimes int
	status := newGuessTCPSeq()
	for i := 0; i < 2048; i++ {
		if err := updateSeqOffsetMap(bpfmapTCPSeq, &status); err != nil {
			return nil, nil, err
		}
		if err := guessTCPSeq(serverAddr); err != nil {
			return nil, nil, err
		}
		s, err := readSeqOffset(bpfmapTCPSeq)
		if err != nil {
			return nil, nil, err
		}
		if s.state&0b11 == 0b11 {
			offset.offset_copied_seq = s.offset_copied_seq
			offset.offset_write_seq = s.offset_write_seq
			okTimes++
			break
		} else {
			status.offset_copied_seq = s.offset_copied_seq
			status.offset_write_seq = s.offset_write_seq
			status.state = s.state
		}
	}

	if okTimes == 0 {
		l.Warn("guess tcp seq offset failed, trying BTF")
		return TryGetTCPSeqOffsetFromBTF()
	}

	seqConstEditor := NewConstEditorTCPSeq(&offset)

	return seqConstEditor, &offset, nil
}

func guessTCPSeq(svc string) error {
	conn, err := net.Dial("tcp4", svc)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	defer conn.Close() //nolint:errcheck

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("conv conn to tcp conn")
	}

	connFile, err := tcpConn.File()
	if err != nil {
		return fmt.Errorf("get tcp file failed: %w", err)
	}

	time.Sleep(time.Millisecond * 15)

	// just used for call kernel func tcp_getsockopt
	_, err = unix.GetsockoptTCPInfo(int(connFile.Fd()), syscall.SOL_TCP, syscall.TCP_INFO)
	if err != nil {
		return err
	}

	_ = connFile.Close()

	return nil
}
