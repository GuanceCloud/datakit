//go:build linux
// +build linux

package netflow

import (
	"testing"

	"github.com/stretchr/testify/require"
	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/exporter"
)

func TestRecordNetflowKprobeCapability(t *testing.T) {
	recordNetflowKprobeCapability(bpfutil.KprobeCapability{
		PMUTypePath: "/sys/bus/event_source/devices/kprobe/type",
	})

	snapshot := exporter.SnapshotKernelFunctionStatus()
	var foundPMU, foundTrace, foundDebug bool
	for _, item := range snapshot {
		switch item.Symbol {
		case "/sys/bus/event_source/devices/kprobe/type":
			foundPMU = true
			require.Equal(t, "available", item.Status)
		case "/sys/kernel/tracing/kprobe_events":
			foundTrace = true
			require.Equal(t, "unknown", item.Status)
		case "/sys/kernel/debug/tracing/kprobe_events":
			foundDebug = true
			require.Equal(t, "unknown", item.Status)
		}
	}

	require.True(t, foundPMU)
	require.True(t, foundTrace)
	require.True(t, foundDebug)
}
