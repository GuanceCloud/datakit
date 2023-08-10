//go:build (!linux && ebpf) || (linux && !amd64 && !arm64 && ebpf) || !ebpf
// +build !linux,ebpf linux,!amd64,!arm64,ebpf !ebpf

// Package ebpf wraps eBPF-network's CGO extensions
package ebpf

import (
	"fmt"
	"runtime"
)

func HTTPFlowBin() ([]byte, error) {
	return nil, fmt.Errorf("unsupportd platform: %s, %s", runtime.GOOS, runtime.GOARCH)
}

func NetFlowBin() ([]byte, error) {
	return nil, fmt.Errorf("unsupportd platform: %s, %s", runtime.GOOS, runtime.GOARCH)
}

func OffsetGuessBin() ([]byte, error) {
	return nil, fmt.Errorf("unsupportd platform: %s, %s", runtime.GOOS, runtime.GOARCH)
}

func OffsetHttpflowBin() ([]byte, error) {
	return nil, fmt.Errorf("unsupportd platform: %s, %s", runtime.GOOS, runtime.GOARCH)
}

func OffsetConntrackBin() ([]byte, error) {
	return nil, fmt.Errorf("unsupportd platform: %s, %s", runtime.GOOS, runtime.GOARCH)
}

func BashHistoryBin() ([]byte, error) {
	return nil, fmt.Errorf("unsupportd platform: %s, %s", runtime.GOOS, runtime.GOARCH)
}
