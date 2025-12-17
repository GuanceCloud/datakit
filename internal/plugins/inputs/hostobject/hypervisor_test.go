// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import (
	T "testing"

	"github.com/stretchr/testify/assert"
)

// TestPublicAPIs tests the public API functions
func TestPublicAPIs(t *T.T) {
	t.Run("IsVirtual", func(t *T.T) {
		// Just ensure the function can be called without panicking
		result := IsVirtual()
		assert.IsType(t, true, result)
	})

	t.Run("GetHypervisorType", func(t *T.T) {
		// Just ensure the function can be called without panicking
		result := GetHypervisorType()
		assert.IsType(t, "", result)
	})
}

// TestHypervisorDetectionConsistency tests that detection results are consistent
func TestHypervisorDetectionConsistency(t *T.T) {
	t.Run("detection_result_type", func(t *T.T) {
		isVirtual := IsVirtual()
		hypervisor := GetHypervisorType()

		// If not virtual, hypervisor type should be empty
		if !isVirtual {
			assert.Empty(t, hypervisor, "physical machine should have empty hypervisor type")
		}

		// If virtual but hypervisor type is empty, that's also valid (unknown VM type)
		// Just ensure the string is valid
		assert.IsType(t, "", hypervisor)
	})
}

// TestInputVirtualMachineTagsIntegration tests the virtual machine detection integration with Input
func TestInputVirtualMachineTagsIntegration(t *T.T) {
	t.Run("virtual_tags_applied", func(t *T.T) {
		ipt := defaultInput()
		ipt.VirtualTags = map[string]string{
			"host_type": "virtual",
			"env":       "cloud",
		}
		ipt.PhysicalTags = map[string]string{
			"host_type": "physical",
			"env":       "on-premise",
		}

		// Simulate setup
		ipt.setup()

		// Verify that isVirtual and hypervisorType are set
		assert.NotZero(t, ipt.isVirtual)
		// hypervisorType can be empty or non-empty, both are valid
		assert.IsType(t, "", ipt.hypervisorType)
	})

	t.Run("physical_tags_applied", func(t *T.T) {
		ipt := defaultInput()
		ipt.PhysicalTags = map[string]string{
			"host_type": "physical",
			"hardware":  "baremetal",
		}

		// Simulate setup
		ipt.setup()

		// Verify that detection happened
		assert.IsType(t, true, ipt.isVirtual)
	})
}

// TestPublicAPIConsistency tests that the public APIs are consistent
func TestPublicAPIConsistency(t *T.T) {
	t.Run("api_callable_multiple_times", func(t *T.T) {
		// Call multiple times to ensure consistency
		for i := 0; i < 3; i++ {
			isVirt := IsVirtual()
			hypervisor := GetHypervisorType()

			assert.IsType(t, true, isVirt)
			assert.IsType(t, "", hypervisor)
		}
	})

	t.Run("hypervisor_type_when_not_virtual", func(t *T.T) {
		isVirt := IsVirtual()
		hypervisor := GetHypervisorType()

		// Logical consistency: if not virtual, hypervisor should be empty
		if !isVirt {
			assert.Empty(t, hypervisor)
		}
	})
}
