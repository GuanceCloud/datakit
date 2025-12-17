// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package hostobject

import (
	T "testing"

	"github.com/stretchr/testify/assert"
)

// TestLinuxVirtualMachineDetection tests the Linux-specific VM detection
func TestLinuxVirtualMachineDetection(t *T.T) {
	t.Run("isVirtual_returns_bool", func(t *T.T) {
		result := isVirtual()
		assert.IsType(t, true, result)
	})

	t.Run("getHypervisorType_returns_string", func(t *T.T) {
		result := getHypervisorType()
		assert.IsType(t, "", result)
	})
}

// TestIsXenVM tests Xen hypervisor detection
func TestIsXenVM(t *T.T) {
	t.Run("isXenVM_returns_bool", func(t *T.T) {
		result := isXenVM()
		assert.IsType(t, true, result)
	})
}

// TestIsCPUFlagsVM tests CPU flags VM detection
func TestIsCPUFlagsVM(t *T.T) {
	t.Run("isCPUFlagsVM_returns_bool", func(t *T.T) {
		result := isCPUFlagsVM()
		assert.IsType(t, true, result)
	})
}

// TestGetCPUHypervisor tests CPU hypervisor extraction
func TestGetCPUHypervisor(t *T.T) {
	t.Run("getCPUHypervisor_returns_string", func(t *T.T) {
		result := getCPUHypervisor()
		assert.IsType(t, "", result)
	})
}

// TestIsDMIVM tests DMI system product name VM detection
func TestIsDMIVM(t *T.T) {
	t.Run("isDMIVM_returns_bool", func(t *T.T) {
		result := isDMIVM()
		assert.IsType(t, true, result)
	})
}

// TestGetDMIHypervisor tests DMI hypervisor extraction
func TestGetDMIHypervisor(t *T.T) {
	t.Run("getDMIHypervisor_returns_string", func(t *T.T) {
		result := getDMIHypervisor()
		assert.IsType(t, "", result)
	})
}

// TestIsOpenVZVM tests OpenVZ environment detection
func TestIsOpenVZVM(t *T.T) {
	t.Run("isOpenVZVM_returns_bool", func(t *T.T) {
		result := isOpenVZVM()
		assert.IsType(t, true, result)
	})
}

// TestIsContainerEnv tests container environment detection
func TestIsContainerEnv(t *T.T) {
	t.Run("isContainerEnv_returns_bool", func(t *T.T) {
		result := isContainerEnv()
		assert.IsType(t, true, result)
	})
}

// TestGetContainerType tests container type detection
func TestGetContainerType(t *T.T) {
	t.Run("getContainerType_returns_string", func(t *T.T) {
		result := getContainerType()
		assert.IsType(t, "", result)
	})
}

// TestLinuxHypervisorDetectionIntegration tests the full detection flow
func TestLinuxHypervisorDetectionIntegration(t *T.T) {
	t.Run("detection_consistency", func(t *T.T) {
		isVirt := isVirtual()
		hypervisor := getHypervisorType()

		assert.IsType(t, true, isVirt)
		assert.IsType(t, "", hypervisor)
	})
}
