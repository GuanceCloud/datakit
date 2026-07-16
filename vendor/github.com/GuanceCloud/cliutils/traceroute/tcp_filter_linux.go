// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux

package traceroute

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"golang.org/x/net/bpf"
	"golang.org/x/sys/unix"
)

const tcpFilterSnapshotLength = 128

func setTCPPacketFilter(conn *net.IPConn, localIP, targetIP net.IP,
	sourcePort, targetPort uint16,
) error {
	instructions, err := tcpPacketFilter(localIP, targetIP, sourcePort, targetPort)
	if err != nil {
		return err
	}
	rawInstructions, err := bpf.Assemble(instructions)
	if err != nil {
		return fmt.Errorf("assemble TCP traceroute packet filter: %w", err)
	}
	filters := make([]unix.SockFilter, len(rawInstructions))
	for index, instruction := range rawInstructions {
		filters[index] = unix.SockFilter{
			Code: instruction.Op,
			Jt:   instruction.Jt,
			Jf:   instruction.Jf,
			K:    instruction.K,
		}
	}
	program := unix.SockFprog{Len: uint16(len(filters)), Filter: &filters[0]}
	rawConn, err := conn.SyscallConn()
	if err != nil {
		return fmt.Errorf("get TCP traceroute raw socket: %w", err)
	}
	var attachErr error
	if err := rawConn.Control(func(fd uintptr) {
		attachErr = unix.SetsockoptSockFprog(int(fd), unix.SOL_SOCKET, unix.SO_ATTACH_FILTER, &program)
	}); err != nil {
		return fmt.Errorf("control TCP traceroute raw socket: %w", err)
	}
	if attachErr != nil {
		return fmt.Errorf("attach TCP traceroute packet filter: %w", attachErr)
	}
	return nil
}

func tcpPacketFilter(localIP, targetIP net.IP, sourcePort, targetPort uint16) ([]bpf.Instruction, error) {
	local := localIP.To4()
	target := targetIP.To4()
	if local == nil || target == nil {
		return nil, errors.New("TCP traceroute packet filter requires IPv4 addresses")
	}
	checks := []bpf.Instruction{
		bpf.LoadAbsolute{Off: 9, Size: 1},
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: 6},
		bpf.LoadAbsolute{Off: 12, Size: 4},
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: binary.BigEndian.Uint32(target)},
		bpf.LoadAbsolute{Off: 16, Size: 4},
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: binary.BigEndian.Uint32(local)},
		bpf.LoadAbsolute{Off: 6, Size: 2},
		bpf.JumpIf{Cond: bpf.JumpBitsNotSet, Val: 0x1fff},
		bpf.LoadMemShift{Off: 0},
		bpf.LoadIndirect{Off: 0, Size: 2},
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: uint32(targetPort)},
		bpf.LoadIndirect{Off: 2, Size: 2},
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: uint32(sourcePort)},
		bpf.LoadIndirect{Off: 13, Size: 1},
		bpf.JumpIf{Cond: bpf.JumpBitsSet, Val: tcpFlagACK},
		bpf.JumpIf{Cond: bpf.JumpBitsSet, Val: tcpFlagSYN | tcpFlagRST},
	}
	dropIndex := len(checks) + 1
	for index, instruction := range checks {
		jump, ok := instruction.(bpf.JumpIf)
		if !ok {
			continue
		}
		jump.SkipFalse = uint8(dropIndex - index - 1)
		checks[index] = jump
	}
	return append(checks,
		bpf.RetConstant{Val: tcpFilterSnapshotLength},
		bpf.RetConstant{Val: 0},
	), nil
}
