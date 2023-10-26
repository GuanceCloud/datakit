//go:build linux
// +build linux

package offset

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	manager "github.com/DataDog/ebpf-manager"
	"github.com/cilium/ebpf"
	"golang.org/x/sys/unix"
)

func GuessOffsetHTTPFlow(status *OffsetGuessC) ([]manager.ConstantEditor, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	bpfManager, err := NewOffsetHTTPFlow()
	if err != nil {
		return nil, err
	}

	if err := bpfManager.Start(); err != nil {
		return nil, err
	}
	defer bpfManager.Stop(manager.CleanAll) //nolint:errcheck

	m, err := BpfMapGuessHTTPInit(bpfManager)
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
	// Write the offset of pid and socket structure member file (sk - 8)
	offsetHTTP.fd = _Ctype_int(connFile.Fd())
	offsetHTTP.offset_socket_file = _Ctype_int(status.offset_socket_sk) - 8

	const start = 1000
	const end = 3500

	offsetHTTP.offset_task_struct_files = start

	err = updateMapGuessHTTP(m, offsetHTTP)
	if err != nil {
		return nil, err
	}

	skipCount := 0
	for i := start; i < end && skipCount < 20; i++ {
		_, err = unix.GetsockoptTCPInfo(int(connFile.Fd()), syscall.SOL_TCP, syscall.TCP_INFO)
		if err != nil {
			return nil, err
		}
		time.Sleep(time.Millisecond * 5)
		offsetTmp, err := readMapGuessHTTP(m)
		if err != nil {
			return nil, err
		}

		if offsetTmp.times == 0 {
			skipCount++
			i--
			continue
		} else {
			skipCount = 0
		}

		if offsetTmp.state == 0b11 {
			break
		}

		offsetTmp.times = 0
		err = updateMapGuessHTTP(m, offsetTmp)
		if err != nil {
			return nil, err
		}
	}

	if skipCount >= 20 {
		return nil, fmt.Errorf("skipCount >= 20")
	}

	offsetHTTP, err = readMapGuessHTTP(m)
	if err != nil {
		return nil, err
	}

	if offsetHTTP.state != 0b11 {
		return nil, fmt.Errorf("offset httpflow: failed")
	}

	if err = connFile.Close(); err != nil {
		return nil, err
	}

	if err = conn.Close(); err != nil {
		return nil, err
	}

	return NewConstHTTPEditor(offsetHTTP), nil
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
