//go:build linux
// +build linux

package offset

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	goruntime "runtime"
	"syscall"
	"unsafe"

	"github.com/cilium/ebpf"
	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
	"golang.org/x/sys/unix"
)

const (
	httpTaskFilesGuessStart = 1024
	httpTaskFilesGuessEnd   = 4096
	httpTaskFilesGuessStep  = 8
)

func GuessOffsetHTTPFlow(status *OffsetGuessC) ([]bpfutil.ConstantPatch, error) {
	goruntime.LockOSThread()
	defer goruntime.UnlockOSThread()
	_ = status

	rt, err := NewOffsetHTTPFlowRuntime()
	if err != nil {
		return nil, err
	}

	if err := rt.StartRuntime(); err != nil {
		return nil, err
	}
	defer rt.Shutdown() //nolint:errcheck

	m, err := BpfMapGuessHTTPInit(rt)
	if err != nil {
		return nil, err
	}

	offsetHTTP, err := readMapGuessHTTP(m)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tcp4ServerPort, err := runTCPServer(ctx, "tcp4", listenIPv4)
	if err != nil {
		return nil, err
	}

	serverAddr := fmt.Sprintf("%s:%d", listenIPv4, tcp4ServerPort)

	conn, err := net.Dial("tcp4", serverAddr)
	if err != nil {
		return nil, err
	}

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return nil, fmt.Errorf("conv conn to tcp conn")
	}

	connFile, err := tcpConn.File()
	if err != nil {
		return nil, fmt.Errorf("get tcp file failed: %w", err)
	}
	offsetHTTP.fd = _Ctype_int(connFile.Fd())
	if laddr, ok := conn.LocalAddr().(*net.TCPAddr); ok {
		offsetHTTP.sport = _Ctype_ushort(laddr.Port)
		if ip4 := laddr.IP.To4(); ip4 != nil {
			offsetHTTP.saddr[3] = _Ctype_uint(binary.BigEndian.Uint32(ip4))
		}
	}
	if raddr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		offsetHTTP.dport = _Ctype_ushort(raddr.Port)
		if ip4 := raddr.IP.To4(); ip4 != nil {
			offsetHTTP.daddr[3] = _Ctype_uint(binary.BigEndian.Uint32(ip4))
		}
	}

	taskFilesGuesses := taskStructFilesGuessSequence()
	if len(taskFilesGuesses) == 0 {
		return nil, fmt.Errorf("no task_struct.files guess candidates")
	}

	offsetHTTP.offset_task_struct_files = _Ctype_int(taskFilesGuesses[0])

	err = updateMapGuessHTTP(m, offsetHTTP)
	if err != nil {
		return nil, err
	}

	l.Debugf("start HTTP flow offset guess: socket_file=%d task_struct_files=%d fd=%d saddr=%#x daddr=%#x sport=%d dport=%d",
		int32(offsetHTTP.offset_socket_file), int32(offsetHTTP.offset_task_struct_files), int32(offsetHTTP.fd),
		uint32(offsetHTTP.saddr[3]), uint32(offsetHTTP.daddr[3]), int32(offsetHTTP.sport), int32(offsetHTTP.dport))

	skipCount := 0
	candidateIdx := 0
	lastState := int32(-1)
	lastTaskFiles := int32(offsetHTTP.offset_task_struct_files)
	lastFilesFDT := int32(offsetHTTP.offset_files_struct_fdt)
	lastPrivateData := int32(offsetHTTP.offset_file_private_data)
	for round := 0; round < len(taskFilesGuesses)+32 && skipCount < 20; round++ {
		if round > 0 {
			offsetHTTP, err = readMapGuessHTTP(m)
			if err != nil {
				return nil, err
			}

			if offsetHTTP.state&0b10 == 0 {
				candidateIdx = 0
			} else if offsetHTTP.state&0b1 == 0 && candidateIdx+1 < len(taskFilesGuesses) {
				candidateIdx++
			}

			offsetHTTP.offset_task_struct_files = _Ctype_int(taskFilesGuesses[candidateIdx])
			offsetHTTP.times = 0
			if err := updateMapGuessHTTP(m, offsetHTTP); err != nil {
				return nil, err
			}
		}

		_, err = unix.GetsockoptTCPInfo(int(connFile.Fd()), syscall.SOL_TCP, syscall.TCP_INFO)
		if err != nil {
			return nil, err
		}
		offsetTmp, err := readMapGuessHTTP(m)
		if err != nil {
			return nil, err
		}

		if offsetTmp.times == 0 {
			skipCount++
			continue
		} else {
			skipCount = 0
		}

		if int32(offsetTmp.state) != lastState ||
			int32(offsetTmp.offset_task_struct_files) != lastTaskFiles ||
			int32(offsetTmp.offset_files_struct_fdt) != lastFilesFDT ||
			int32(offsetTmp.offset_file_private_data) != lastPrivateData {
			l.Debugf(
				"HTTP flow offset guess progress: round=%d state=%03b "+
					"task_struct_files=%d files_struct_fdt=%d "+
					"file_private_data=%d socket_sk=%d times=%d",
				round+1, int32(offsetTmp.state), int32(offsetTmp.offset_task_struct_files), int32(offsetTmp.offset_files_struct_fdt),
				int32(offsetTmp.offset_file_private_data), int32(offsetTmp.offset_socket_sk), int32(offsetTmp.times),
			)
			lastState = int32(offsetTmp.state)
			lastTaskFiles = int32(offsetTmp.offset_task_struct_files)
			lastFilesFDT = int32(offsetTmp.offset_files_struct_fdt)
			lastPrivateData = int32(offsetTmp.offset_file_private_data)
		}

		if offsetTmp.state&0b11 == 0b11 {
			break
		}

		offsetTmp.times = 0
		err = updateMapGuessHTTP(m, offsetTmp)
		if err != nil {
			return nil, err
		}
	}

	if skipCount >= 20 {
		l.Warnf("HTTP flow offset guess stalled: task_struct_files=%d files_struct_fdt=%d socket_file=%d file_private_data=%d socket_sk=%d state=%03b",
			int32(offsetHTTP.offset_task_struct_files), int32(offsetHTTP.offset_files_struct_fdt), int32(offsetHTTP.offset_socket_file),
			int32(offsetHTTP.offset_file_private_data), int32(offsetHTTP.offset_socket_sk), int32(offsetHTTP.state))
		return nil, fmt.Errorf("skipCount >= 20")
	}

	offsetHTTP, err = readMapGuessHTTP(m)
	if err != nil {
		return nil, err
	}

	if offsetHTTP.state&0b11 != 0b11 {
		l.Warnf(
			"HTTP flow offset guess failed: state=%03b task_struct_files=%d "+
				"files_struct_fdt=%d socket_file=%d file_private_data=%d "+
				"socket_sk=%d times=%d fd=%d",
			int32(offsetHTTP.state), int32(offsetHTTP.offset_task_struct_files), int32(offsetHTTP.offset_files_struct_fdt),
			int32(offsetHTTP.offset_socket_file), int32(offsetHTTP.offset_file_private_data), int32(offsetHTTP.offset_socket_sk),
			int32(offsetHTTP.times), int32(offsetHTTP.fd),
		)
		return nil, fmt.Errorf("offset httpflow: failed")
	}

	l.Infof("HTTP flow offsets guessed: task_struct_files=%d files_struct_fdt=%d socket_file=%d file_private_data=%d socket_sk=%d",
		int32(offsetHTTP.offset_task_struct_files), int32(offsetHTTP.offset_files_struct_fdt),
		int32(offsetHTTP.offset_socket_file), int32(offsetHTTP.offset_file_private_data), int32(offsetHTTP.offset_socket_sk))

	if err = connFile.Close(); err != nil {
		return nil, err
	}

	if err = conn.Close(); err != nil {
		return nil, err
	}

	patches := NewConstHTTPEditor(offsetHTTP)
	switch {
	case offsetHTTP.offset_socket_sk > 0:
		patches = append(patches, bpfutil.ConstantPatch{
			Name:  "offset_socket_sk",
			Value: uint64(int32(offsetHTTP.offset_socket_sk)),
		})
	case offsetHTTP.offset_socket_file > 0:
		socketSk := uint64(int32(offsetHTTP.offset_socket_file)) + uint64(unsafe.Sizeof(uintptr(0))) //nolint:gosec
		l.Warnf(
			"HTTP flow socket.sk offset did not converge; "+
				"derive from socket.file fallback: socket_file=%d socket_sk=%d",
			int32(offsetHTTP.offset_socket_file), socketSk)
		patches = append(patches, bpfutil.ConstantPatch{
			Name:  "offset_socket_sk",
			Value: socketSk,
		})
	default:
		l.Warnf("HTTP flow socket.sk offset did not converge; keep using kernel offset fallback=%d",
			uint64(status.offset_socket_sk))
	}

	return patches, nil
}

func readMapGuessHTTP(m *ebpf.Map) (*OffsetHTTPFlowC, error) {
	value := OffsetHTTPFlowC{}
	key := uint64(0)
	if err := m.Lookup(&key, unsafe.Pointer(&value)); err != nil { //nolint:gosec
		return nil, err
	} else {
		return &value, nil
	}
}

func updateMapGuessHTTP(m *ebpf.Map, offset *OffsetHTTPFlowC) error {
	key := uint64(0)
	return m.Update(&key, offset, ebpf.UpdateAny)
}

func taskStructFilesGuessSequence() []int32 {
	kernelVersion, err := bpfutil.CurrentKernelVersion()
	if err != nil {
		l.Debugf("read kernel version for HTTP guess ordering failed: %v", err)
	}

	return taskStructFilesGuessSequenceForKernel(kernelVersion)
}

func taskStructFilesGuessSequenceForKernel(kernelVersion uint64) []int32 {
	const (
		start = httpTaskFilesGuessStart
		end   = httpTaskFilesGuessEnd
		step  = httpTaskFilesGuessStep
	)

	anchors := []int32{1880, 2704, 2048, 1536}
	switch {
	case kernelVersion != 0 && kernelVersion < (4<<48):
		anchors = []int32{1880, 2048, 1536, 2704}
	case kernelVersion >= (4<<48|15<<32) && kernelVersion < (5<<48):
		anchors = []int32{2704, 2688, 2720, 1880, 2048}
	}

	sequence := make([]int32, 0, (end-start)/step+1)
	seen := make(map[int32]struct{}, (end-start)/step+1)
	add := func(v int32) {
		if v < start || v >= end || v%step != 0 {
			return
		}
		if _, ok := seen[v]; ok {
			return
		}
		seen[v] = struct{}{}
		sequence = append(sequence, v)
	}

	deltas := []int32{0, -8, 8, -16, 16, -24, 24, -32, 32, -64, 64, -128, 128}
	for _, anchor := range anchors {
		for _, delta := range deltas {
			add(anchor + delta)
		}
	}

	for candidate := int32(start); candidate < end; candidate += step {
		add(candidate)
	}

	return sequence
}
