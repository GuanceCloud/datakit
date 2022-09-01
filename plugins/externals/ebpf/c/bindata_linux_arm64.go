//go:build linux && arm64 && ebpf
// +build linux,arm64,ebpf

// Package ebpf wraps eBPF-network's CGO extensions
package ebpf

import (
	"embed"
)

//go:embed bin/arm64
var binData embed.FS

func HTTPFlowBin() ([]byte, error) {
	return binData.ReadFile("bin/arm64/httpflow.o")
}

func NetFlowBin() ([]byte, error) {
	return binData.ReadFile("bin/arm64/netflow.o")
}

func OffsetGuessBin() ([]byte, error) {
	return binData.ReadFile("bin/arm64/offset_guess.o")
}

func BashHistoryBin() ([]byte, error) {
	return binData.ReadFile("bin/arm64/bash_history.o")
}
