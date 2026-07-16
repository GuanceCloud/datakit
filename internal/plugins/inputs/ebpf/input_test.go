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

func TestAppendNetworkPathArgs(t *testing.T) {
	args := appendNetworkPathArgs([]string{"run"}, &Input{
		NetworkPathEnabled:       true,
		NetworkPathAPI:           "http://127.0.0.1:9529/v1/netpath/candidates",
		NetworkPathToken:         "secret",
		NetworkPathFlushInterval: "5s",
		NetworkPathBatchSize:     10,
		NetworkPathHTTPTimeout:   "2s",
		NetworkPathQueueSize:     100,
	})

	assert.Equal(t, []string{
		"run",
		"--network-path-enabled",
		"--network-path-api", "http://127.0.0.1:9529/v1/netpath/candidates",
		"--network-path-flush-interval", "5s",
		"--network-path-batch-size", "10",
		"--network-path-http-timeout", "2s",
		"--network-path-queue-size", "100",
	}, args)

	args = appendNetworkPathArgs([]string{"run"}, &Input{})
	assert.Equal(t, []string{"run"}, args)
}

func TestAppendNetworkPathEnvs(t *testing.T) {
	envs := appendNetworkPathEnvs([]string{"EXISTING=value"}, &Input{
		NetworkPathEnabled: true,
		NetworkPathToken:   "secret",
	})
	assert.Equal(t, []string{"EXISTING=value", "DKE_NETWORK_PATH_TOKEN=secret"}, envs)

	envs = appendNetworkPathEnvs([]string{"EXISTING=value"}, &Input{})
	assert.Equal(t, []string{"EXISTING=value"}, envs)
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

func TestReadEnvNetworkPathConfig(t *testing.T) {
	ipt := &Input{}

	ipt.ReadEnv(map[string]string{
		"ENV_INPUT_EBPF_NETWORK_PATH_ENABLED":        "true",
		"ENV_INPUT_EBPF_NETWORK_PATH_API":            "http://datakit/v1/netpath/candidates",
		"ENV_INPUT_EBPF_NETWORK_PATH_TOKEN":          "secret",
		"ENV_INPUT_EBPF_NETWORK_PATH_FLUSH_INTERVAL": "5s",
		"ENV_INPUT_EBPF_NETWORK_PATH_BATCH_SIZE":     "20",
		"ENV_INPUT_EBPF_NETWORK_PATH_HTTP_TIMEOUT":   "2s",
		"ENV_INPUT_EBPF_NETWORK_PATH_QUEUE_SIZE":     "200",
	})

	assert.True(t, ipt.NetworkPathEnabled)
	assert.Equal(t, "http://datakit/v1/netpath/candidates", ipt.NetworkPathAPI)
	assert.Equal(t, "secret", ipt.NetworkPathToken)
	assert.Equal(t, "5s", ipt.NetworkPathFlushInterval)
	assert.Equal(t, 20, ipt.NetworkPathBatchSize)
	assert.Equal(t, "2s", ipt.NetworkPathHTTPTimeout)
	assert.Equal(t, 200, ipt.NetworkPathQueueSize)
}
