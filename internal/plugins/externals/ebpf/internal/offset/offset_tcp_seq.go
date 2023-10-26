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
	"runtime"
	"syscall"
	"time"
	"unsafe"

	manager "github.com/DataDog/ebpf-manager"
	"github.com/cilium/ebpf"
	"github.com/google/gopacket/afpacket"
	"golang.org/x/sys/unix"

	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/c"
)

func newOffsetTCPSeqManager(skFd int, cnt []manager.ConstantEditor) (*manager.Manager, error) {
	m := &manager.Manager{
		Probes: []*manager.Probe{
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "socket__packet_tcp_header",
				},
				SocketFD: skFd,
			},
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kprobe__tcp_getsockopt",
					UID:          "tcp_getsockopt_tcp_seq",
				},
			},
		},
	}

	mOpts := manager.Options{
		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},
		ConstantEditors: cnt,
	}

	buf, err := dkebpf.OffsetTCPSeqBin()
	if err != nil {
		return nil, fmt.Errorf("load bpf prog: %w", err)
	}

	if err := m.InitWithOptions(bytes.NewReader(buf), mOpts); err != nil {
		return nil, fmt.Errorf("init offset tcp seq guess: %w", err)
	}

	return m, nil
}

func bpfMapGuessTCPSeqInit(m *manager.Manager) (*ebpf.Map, error) {
	bpfmapTCPSeq, found, err := m.GetMap("bpfmap_offset_tcp_seq")
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fmt.Errorf("bpf map bpfmap_offset_tcp_seq not found")
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

func GuessOffsetTCPSeq(netflowOffset []manager.ConstantEditor) ([]manager.ConstantEditor, *OffsetTCPSeqC, error) {
	// current netns

	rawSocket, err := afpacket.NewTPacket()
	if err != nil {
		return nil, nil, fmt.Errorf("error creating raw socket: %w", err)
	}
	defer rawSocket.Close()

	// The underlying socket file descriptor is private, hence the use of reflection
	skFd := int(reflect.ValueOf(rawSocket).Elem().FieldByName("fd").Int())

	m, err := newOffsetTCPSeqManager(skFd, netflowOffset)
	if err != nil {
		return nil, nil, err
	}

	if err := m.Start(); err != nil {
		return nil, nil, err
	}

	defer m.Stop(manager.CleanAll) //nolint:errcheck

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tcp4ServerPort, err := runTCPServer(ctx, "tcp4", listenIPv4)
	if err != nil {
		return nil, nil, err
	}

	serverAddr := fmt.Sprintf("%s:%d", listenIPv4, tcp4ServerPort)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	bpfmapTCPSeq, found, err := m.GetMap("bpfmap_offset_tcp_seq")
	if err != nil {
		return nil, nil, err
	}

	if !found {
		return nil, nil, fmt.Errorf("bpf map bpfmap_offset_tcp_seq not found")
	}

	_, err = bpfMapGuessTCPSeqInit(m)
	if err != nil {
		return nil, nil, err
	}

	// stauts := newGuessTCPSeq()
	offset := OffsetTCPSeqC{}

	var okTimes int
	status := newGuessTCPSeq()
	for i := 0; i < 1024; i++ {
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
		return nil, nil, fmt.Errorf("guess tcp seq offset failed")
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
