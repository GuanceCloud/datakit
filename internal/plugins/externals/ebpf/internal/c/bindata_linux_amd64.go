//go:build linux && amd64
// +build linux,amd64

// Package ebpf wraps eBPF-network's CGO extensions
package ebpf

import (
	"embed"
)

//go:embed elf/linux_amd64
var binData embed.FS

func HTTPFlowBin() ([]byte, error) {
	return binData.ReadFile("elf/linux_amd64/httpflow.o")
}

func NetFlowBin() ([]byte, error) {
	return binData.ReadFile("elf/linux_amd64/netflow.o")
}

func ConntrackBin() ([]byte, error) {
	return binData.ReadFile("elf/linux_amd64/conntrack.o")
}

func ProcessSchedBin() ([]byte, error) {
	return binData.ReadFile("elf/linux_amd64/process_sched.o")
}

func OffsetGuessBin() ([]byte, error) {
	return binData.ReadFile("elf/linux_amd64/offset_guess.o")
}

func OffsetHttpflowBin() ([]byte, error) {
	return binData.ReadFile("elf/linux_amd64/offset_httpflow.o")
}

func OffsetConntrackBin() ([]byte, error) {
	return binData.ReadFile("elf/linux_amd64/offset_conntrack.o")
}

func OffsetTCPSeqBin() ([]byte, error) {
	return binData.ReadFile("elf/linux_amd64/offset_tcp_seq.o")
}

func BashHistoryBin() ([]byte, error) {
	return binData.ReadFile("elf/linux_amd64/bash_history.o")
}
