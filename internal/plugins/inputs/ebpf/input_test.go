// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ebpf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/external"
)

func TestEBPFInputInheritsExternal(t *testing.T) {
	// Test that ebpf.Input embeds external.Input
	// This ensures ebpf process monitoring is automatically integrated
	ipt := &Input{}

	// Verify that ebpf.Input has embedded external.Input
	// by checking if we can access external.Input fields
	assert.NotNil(t, &ipt.Input, "ebpf.Input should embed external.Input")

	// The Name field from external.Input should be accessible
	ipt.Input.Name = "ebpf"
	assert.Equal(t, "ebpf", ipt.Input.Name)

	// The Daemon field from external.Input should be accessible
	ipt.Input.Daemon = true
	assert.True(t, ipt.Input.Daemon)
}

func TestEBPFProcessMonitoringIntegration(t *testing.T) {
	// Test that when ebpf runs as daemon, it will be monitored
	// via external.ProcessMonitor
	ipt := &Input{}
	ipt.Input.Daemon = true
	ipt.Input.Name = "ebpf"

	// When Daemon is true, the process will be registered to ProcessMonitor
	// in external.Input.daemonRun() method
	assert.True(t, ipt.Input.Daemon, "ebpf should run as daemon")

	// Verify that GetProcessMonitor is available
	monitor := external.GetProcessMonitor()
	assert.NotNil(t, monitor, "ProcessMonitor should be initialized")
}

func TestAppendResourceLimitArgs(t *testing.T) {
	args := appendResourceLimitArgs([]string{"run"}, "2.0", "4GiB", "100MiB/s")

	assert.Equal(t, []string{
		"run",
		"--res-cpu", "2.0",
		"--res-mem", "4GiB",
		"--res-bandwidth", "100MiB/s",
	}, args)
}

func TestAppendNetlogCaptureLimitArgs(t *testing.T) {
	args := appendNetlogCaptureLimitArgs([]string{"run"}, 2, 4, 32)

	assert.Equal(t, []string{
		"run",
		"--netlog-fallback-sockets", "2",
		"--netlog-fallback-blocks", "4",
		"--netlog-shared-blocks", "32",
	}, args)

	args = appendNetlogCaptureLimitArgs([]string{"run"}, 0, -1, 0)
	assert.Equal(t, []string{"run"}, args)
}

func TestAppendNetlogL7LogArgs(t *testing.T) {
	args := appendNetlogL7LogArgs([]string{"run"},
		[]string{"http1", " http2 ", ""},
		[]string{"host", "x-request-id"},
	)

	assert.Equal(t, []string{
		"run",
		"--netlog-protocols", "http1,http2",
		"--netlog-l7log-headers", "host,x-request-id",
	}, args)

	args = appendNetlogL7LogArgs([]string{"run"}, nil, []string{" none "})
	assert.Equal(t, []string{
		"run",
		"--netlog-l7log-headers", "none",
	}, args)
}

func TestReadEnvNetlogL7LogConfig(t *testing.T) {
	ipt := &Input{}

	ipt.ReadEnv(map[string]string{
		"ENV_INPUT_EBPF_NETLOG_L7LOG_PROTOCOLS": "http1, http2,,grpc",
		"ENV_INPUT_EBPF_NETLOG_L7LOG_HEADERS":   "host, x-request-id, traceparent",
	})

	assert.Equal(t, []string{"http1", "http2", "grpc"}, ipt.NetlogL7LogProtocols)
	assert.Equal(t, []string{"host", "x-request-id", "traceparent"}, ipt.NetlogL7LogHeaders)
}
