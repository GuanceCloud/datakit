package exporter

import (
	"sort"
	"strings"
	"sync"
	"time"
)

type KernelFunctionStatus struct {
	Component string    `json:"component"`
	Program   string    `json:"program"`
	Symbol    string    `json:"symbol"`
	Status    string    `json:"status"`
	Reason    string    `json:"reason,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

var (
	kernelFnMu    sync.RWMutex
	kernelFnState = map[string]KernelFunctionStatus{}
)

func RecordKernelFunctionStatus(component, program, symbol, status, reason string) {
	if component == "" || status == "" {
		return
	}

	key := strings.Join([]string{component, program, symbol}, "\x00")
	kernelFnMu.Lock()
	defer kernelFnMu.Unlock()
	kernelFnState[key] = KernelFunctionStatus{
		Component: component,
		Program:   program,
		Symbol:    symbol,
		Status:    status,
		Reason:    trimKernelFunctionReason(reason),
		UpdatedAt: time.Now(),
	}
}

func SnapshotKernelFunctionStatus() []KernelFunctionStatus {
	kernelFnMu.RLock()
	defer kernelFnMu.RUnlock()

	out := make([]KernelFunctionStatus, 0, len(kernelFnState))
	for _, item := range kernelFnState {
		out = append(out, item)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Component != out[j].Component {
			return out[i].Component < out[j].Component
		}
		if out[i].Status != out[j].Status {
			return out[i].Status < out[j].Status
		}
		if out[i].Symbol != out[j].Symbol {
			return out[i].Symbol < out[j].Symbol
		}
		return out[i].Program < out[j].Program
	})

	return out
}

func FailedKernelFunctionStatus(items []KernelFunctionStatus) []KernelFunctionStatus {
	failed := make([]KernelFunctionStatus, 0, len(items))
	for _, item := range items {
		if isKernelFunctionFailure(item.Status) {
			failed = append(failed, item)
		}
	}
	return failed
}

func isKernelFunctionFailure(status string) bool {
	switch status {
	case "available", "ok":
		return false
	default:
		return true
	}
}

func trimKernelFunctionReason(reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return ""
	}
	if idx := strings.IndexByte(reason, '\n'); idx >= 0 {
		reason = reason[:idx]
	}
	const maxLen = 240
	if len(reason) > maxLen {
		return reason[:maxLen] + "..."
	}
	return reason
}
