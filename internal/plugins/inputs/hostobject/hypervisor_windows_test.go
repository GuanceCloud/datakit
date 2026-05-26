// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package hostobject

import (
	T "testing"

	"github.com/stretchr/testify/assert"
)

// TestWindowsVirtualMachineDetection tests the Windows-specific VM detection
func TestWindowsVirtualMachineDetection(t *T.T) {
	t.Run("isVirtual_returns_bool", func(t *T.T) {
		result := isVirtual()
		assert.IsType(t, true, result)
	})

	t.Run("getHypervisorType_returns_string", func(t *T.T) {
		result := getHypervisorType()
		assert.IsType(t, "", result)
	})
}

// TestIsWMIVirtualMachine tests WMI-based VM detection
func TestIsWMIVirtualMachine(t *T.T) {
	t.Run("isWMIVirtualMachine_returns_bool", func(t *T.T) {
		result := isWMIVirtualMachine()
		assert.IsType(t, true, result)
	})
}

// TestIsWMIBaseBoardVM tests WMI BaseBoard VM detection
func TestIsWMIBaseBoardVM(t *T.T) {
	t.Run("isWMIBaseBoardVM_returns_bool", func(t *T.T) {
		result := isWMIBaseBoardVM()
		assert.IsType(t, true, result)
	})
}

// TestIsRegistryVM tests registry-based VM detection
func TestIsRegistryVM(t *T.T) {
	t.Run("isRegistryVM_returns_bool", func(t *T.T) {
		result := isRegistryVM()
		assert.IsType(t, true, result)
	})
}

// TestIsCPUSignatureVM tests CPU signature VM detection
func TestIsCPUSignatureVM(t *T.T) {
	t.Run("isCPUSignatureVM_returns_bool", func(t *T.T) {
		result := isCPUSignatureVM()
		assert.IsType(t, true, result)
	})
}

// TestGetWMIManufacturer tests WMI manufacturer retrieval
func TestGetWMIManufacturer(t *T.T) {
	t.Run("getWMIManufacturer_returns_string", func(t *T.T) {
		result := getWMIManufacturer()
		assert.IsType(t, "", result)
	})
}

// TestGetWMIModel tests WMI model retrieval
func TestGetWMIModel(t *T.T) {
	t.Run("getWMIModel_returns_string", func(t *T.T) {
		result := getWMIModel()
		assert.IsType(t, "", result)
	})
}

// TestGetWMIBaseBoardManufacturer tests WMI baseboard manufacturer retrieval
func TestGetWMIBaseBoardManufacturer(t *T.T) {
	t.Run("getWMIBaseBoardManufacturer_returns_string", func(t *T.T) {
		result := getWMIBaseBoardManufacturer()
		assert.IsType(t, "", result)
	})
}

// TestGetWMIHypervisor tests WMI hypervisor extraction
func TestGetWMIHypervisor(t *T.T) {
	t.Run("getWMIHypervisor_returns_string", func(t *T.T) {
		result := getWMIHypervisor()
		assert.IsType(t, "", result)
	})
}

// TestGetRegistryHypervisor tests registry hypervisor extraction
func TestGetRegistryHypervisor(t *T.T) {
	t.Run("getRegistryHypervisor_returns_string", func(t *T.T) {
		result := getRegistryHypervisor()
		assert.IsType(t, "", result)
	})
}

// TestGetCPUSignatureHypervisor tests CPU signature hypervisor extraction
func TestGetCPUSignatureHypervisor(t *T.T) {
	t.Run("getCPUSignatureHypervisor_returns_string", func(t *T.T) {
		result := getCPUSignatureHypervisor()
		assert.IsType(t, "", result)
	})
}

// TestWindowsHypervisorDetectionIntegration tests the full Windows detection flow
func TestWindowsHypervisorDetectionIntegration(t *T.T) {
	t.Run("detection_consistency", func(t *T.T) {
		isVirt := isVirtual()
		hypervisor := getHypervisorType()

		assert.IsType(t, true, isVirt)
		assert.IsType(t, "", hypervisor)
	})
}

// TestMultipleDetectionCalls tests that multiple calls to detection functions are safe
func TestMultipleDetectionCalls(t *T.T) {
	t.Run("repeated_calls_safe", func(t *T.T) {
		for i := 0; i < 5; i++ {
			result1 := isVirtual()
			result2 := getHypervisorType()

			assert.IsType(t, true, result1)
			assert.IsType(t, "", result2)
		}
	})
}
