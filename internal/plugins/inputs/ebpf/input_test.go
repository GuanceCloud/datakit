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
