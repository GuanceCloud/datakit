//go:build linux && arm64
// +build linux,arm64

// Package ebpf wraps eBPF-network's CGO extensions
package ebpf

import (
	"embed"
)

//go:embed elf/linux_arm64
var binData embed.FS

func HTTPFlowBin() ([]byte, error) {
	return binData.ReadFile("elf/linux_arm64/httpflow.o")
}

func NetFlowBin() ([]byte, error) {
	return binData.ReadFile("elf/linux_arm64/netflow.o")
}

func ConntrackBin() ([]byte, error) {
	return binData.ReadFile("elf/linux_arm64/conntrack.o")
}

func ProcessSchedBin() ([]byte, error) {
	return binData.ReadFile("elf/linux_arm64/process_sched.o")
}

func OffsetGuessBin() ([]byte, error) {
	return binData.ReadFile("elf/linux_arm64/offset_guess.o")
}

func OffsetHttpflowBin() ([]byte, error) {
	return binData.ReadFile("elf/linux_arm64/offset_httpflow.o")
}

func OffsetConntrackBin() ([]byte, error) {
	return binData.ReadFile("elf/linux_arm64/offset_conntrack.o")
}

func OffsetTCPSeqBin() ([]byte, error) {
	return binData.ReadFile("elf/linux_arm64/offset_tcp_seq.o")
}

func BashHistoryBin() ([]byte, error) {
	return binData.ReadFile("elf/linux_arm64/bash_history.o")
}
