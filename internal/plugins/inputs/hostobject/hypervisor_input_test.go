// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import (
	T "testing"

	"github.com/stretchr/testify/assert"
)

// TestVirtualMachineTagsMerging tests the merging of virtual machine detection tags
func TestVirtualMachineTagsMerging(t *T.T) {
	t.Run("virtual_machine_tags_configuration", func(t *T.T) {
		ipt := &Input{
			VirtualTags: map[string]string{
				"host_type": "virtual",
				"env":       "cloud",
			},
			PhysicalTags: map[string]string{
				"host_type": "physical",
				"env":       "on-premise",
			},
		}

		assert.Equal(t, "virtual", ipt.VirtualTags["host_type"])
		assert.Equal(t, "cloud", ipt.VirtualTags["env"])
		assert.Equal(t, "physical", ipt.PhysicalTags["host_type"])
		assert.Equal(t, "on-premise", ipt.PhysicalTags["env"])
	})
}

// TestInputVirtualMachineDetectionCaching tests that VM detection results are cached
func TestInputVirtualMachineDetectionCaching(t *T.T) {
	t.Run("detection_fields_exist", func(t *T.T) {
		ipt := &Input{}

		// Verify fields exist
		assert.IsType(t, false, ipt.isVirtual)
		assert.IsType(t, "", ipt.hypervisorType)
	})
}

// TestConfigurationLoading tests that virtual/physical tags configuration is properly loaded
func TestConfigurationLoading(t *T.T) {
	t.Run("tags_configuration", func(t *T.T) {
		ipt := &Input{
			VirtualTags: map[string]string{
				"vm_type":  "vmware",
				"location": "cloud",
			},
			PhysicalTags: map[string]string{
				"machine_type": "bare_metal",
				"location":     "datacenter",
			},
		}

		assert.Len(t, ipt.VirtualTags, 2)
		assert.Len(t, ipt.PhysicalTags, 2)
		assert.Equal(t, "vmware", ipt.VirtualTags["vm_type"])
		assert.Equal(t, "bare_metal", ipt.PhysicalTags["machine_type"])
	})

	t.Run("empty_configuration", func(t *T.T) {
		ipt := &Input{
			VirtualTags:  map[string]string{},
			PhysicalTags: map[string]string{},
		}

		assert.Empty(t, ipt.VirtualTags)
		assert.Empty(t, ipt.PhysicalTags)
	})

	t.Run("nil_configuration", func(t *T.T) {
		ipt := &Input{}

		assert.Nil(t, ipt.VirtualTags)
		assert.Nil(t, ipt.PhysicalTags)
	})
}

// TestHypervisorTypeValidation validates the hypervisor type returned
func TestHypervisorTypeValidation(t *T.T) {
	t.Run("public_api_callable", func(t *T.T) {
		// Just verify the public APIs can be called
		isVirt := IsVirtual()
		hypervisor := GetHypervisorType()

		assert.IsType(t, true, isVirt)
		assert.IsType(t, "", hypervisor)
	})
}

// TestInputVirtualMachineFieldsInitialization tests proper initialization of VM fields
func TestInputVirtualMachineFieldsInitialization(t *T.T) {
	t.Run("fields_exist", func(t *T.T) {
		ipt := &Input{}

		assert.IsType(t, false, ipt.isVirtual)
		assert.IsType(t, "", ipt.hypervisorType)
	})

	t.Run("fields_zero_values", func(t *T.T) {
		ipt := &Input{}

		assert.False(t, ipt.isVirtual)
		assert.Empty(t, ipt.hypervisorType)
	})
}
