//go:build linux
// +build linux

package l4log

import (
	"golang.org/x/net/bpf"
)

// tcpdump tcp or udp port 8472 or udp port 4789 -s2048 -dd.
func newBPFFilter() []bpf.RawInstruction {
	return []bpf.RawInstruction{
		{Op: 0x28, Jt: 0, Jf: 0, K: 0x0000000c},
		{Op: 0x15, Jt: 0, Jf: 11, K: 0x000086dd},
		{Op: 0x30, Jt: 0, Jf: 0, K: 0x00000014},
		{Op: 0x15, Jt: 22, Jf: 0, K: 0x00000006},
		{Op: 0x15, Jt: 0, Jf: 2, K: 0x0000002c},
		{Op: 0x30, Jt: 0, Jf: 0, K: 0x00000036},
		{Op: 0x15, Jt: 19, Jf: 20, K: 0x00000006},
		{Op: 0x15, Jt: 0, Jf: 19, K: 0x00000011},
		{Op: 0x28, Jt: 0, Jf: 0, K: 0x00000036},
		{Op: 0x15, Jt: 16, Jf: 0, K: 0x00002118},
		{Op: 0x15, Jt: 15, Jf: 0, K: 0x000012b5},
		{Op: 0x28, Jt: 0, Jf: 0, K: 0x00000038},
		{Op: 0x15, Jt: 13, Jf: 12, K: 0x00002118},
		{Op: 0x15, Jt: 0, Jf: 13, K: 0x00000800},
		{Op: 0x30, Jt: 0, Jf: 0, K: 0x00000017},
		{Op: 0x15, Jt: 10, Jf: 0, K: 0x00000006},
		{Op: 0x15, Jt: 0, Jf: 10, K: 0x00000011},
		{Op: 0x28, Jt: 0, Jf: 0, K: 0x00000014},
		{Op: 0x45, Jt: 8, Jf: 0, K: 0x00001fff},
		{Op: 0xb1, Jt: 0, Jf: 0, K: 0x0000000e},
		{Op: 0x48, Jt: 0, Jf: 0, K: 0x0000000e},
		{Op: 0x15, Jt: 4, Jf: 0, K: 0x00002118},
		{Op: 0x15, Jt: 3, Jf: 0, K: 0x000012b5},
		{Op: 0x48, Jt: 0, Jf: 0, K: 0x00000010},
		{Op: 0x15, Jt: 1, Jf: 0, K: 0x00002118},
		{Op: 0x15, Jt: 0, Jf: 1, K: 0x000012b5},
		{Op: 0x6, Jt: 0, Jf: 0, K: 0x00000800},
		{Op: 0x6, Jt: 0, Jf: 0, K: 0x00000000},
	}
}
