//go:build linux && amd64 && ebpf
// +build linux,amd64,ebpf

// Package ebpf wraps eBPF-network's CGO extensions
package ebpf

import (
	"embed"
)

//go:embed bin/amd64
var binData embed.FS

func HTTPFlowBin() ([]byte, error) {
	return binData.ReadFile("bin/amd64/httpflow.o")
}

func NetFlowBin() ([]byte, error) {
	return binData.ReadFile("bin/amd64/netflow.o")
}

func ConntrackBin() ([]byte, error) {
	return binData.ReadFile("bin/amd64/conntrack.o")
}

func OffsetGuessBin() ([]byte, error) {
	return binData.ReadFile("bin/amd64/offset_guess.o")
}

func OffsetHttpflowBin() ([]byte, error) {
	return binData.ReadFile("bin/amd64/offset_httpflow.o")
}

func OffsetConntrackBin() ([]byte, error) {
	return binData.ReadFile("bin/amd64/offset_conntrack.o")
}

func BashHistoryBin() ([]byte, error) {
	return binData.ReadFile("bin/amd64/bash_history.o")
}
