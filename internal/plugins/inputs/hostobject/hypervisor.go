// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux || windows
// +build linux windows

package hostobject

// IsVirtual detects whether the current host is a virtual machine.
// Returns true if running in a virtual machine, false if it's a physical machine.
// If detection fails, returns false (defaults to physical machine).
func IsVirtual() bool {
	return isVirtual()
}

// GetHypervisorType returns the type of hypervisor if running in a virtual machine.
// Returns empty string if running on physical machine or detection fails.
// Possible values: "kvm", "vmware", "virtualbox", "xen", "hyperv", "openvz", "unknown", etc.
func GetHypervisorType() string {
	return getHypervisorType()
}
