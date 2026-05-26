//go:build linux
// +build linux

package bpfutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKprobeCapabilityHasAnyInterface(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		require.False(t, (KprobeCapability{}).HasAnyInterface())
	})

	t.Run("pmu", func(t *testing.T) {
		require.True(t, (KprobeCapability{PMUTypePath: "/sys/bus/event_source/devices/kprobe/type"}).HasAnyInterface())
	})

	t.Run("tracefs", func(t *testing.T) {
		require.True(t, (KprobeCapability{TraceFSKprobeEvents: "/sys/kernel/tracing/kprobe_events"}).HasAnyInterface())
	})
}

func TestKprobeCapabilityMissingPaths(t *testing.T) {
	caps := KprobeCapability{
		PMUTypePath: "/sys/bus/event_source/devices/kprobe/type",
	}

	require.Equal(t, []string{
		"/sys/kernel/tracing/kprobe_events",
		"/sys/kernel/debug/tracing/kprobe_events",
	}, caps.MissingPaths())
}
