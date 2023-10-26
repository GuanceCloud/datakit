//go:build (linux && !amd64 && !arm64) || !linux
// +build linux,!amd64,!arm64 !linux

// Package ebpf wraps eBPF-network's CGO extensions
package ebpf

import (
	"fmt"
	"runtime"
)

func notImplemented() ([]byte, error) {
	return nil, fmt.Errorf("unsupported platform: %s, %s", runtime.GOOS, runtime.GOARCH)
}

func HTTPFlowBin() ([]byte, error) {
	return notImplemented()
}

func NetFlowBin() ([]byte, error) {
	return notImplemented()
}

func ConntrackBin() ([]byte, error) {
	return notImplemented()
}

func ProcessSchedBin() ([]byte, error) {
	return notImplemented()
}

func OffsetGuessBin() ([]byte, error) {
	return notImplemented()
}

func OffsetHttpflowBin() ([]byte, error) {
	return notImplemented()
}

func OffsetConntrackBin() ([]byte, error) {
	return notImplemented()
}

func OffsetTCPSeqBin() ([]byte, error) {
	return notImplemented()
}

func BashHistoryBin() ([]byte, error) {
	return notImplemented()
}
