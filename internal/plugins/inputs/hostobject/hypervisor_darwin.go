// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build darwin
// +build darwin

package hostobject

// IsVirtual is a stub implementation for macOS.
// Always returns false (assumes not running in a virtual machine).
func IsVirtual() bool {
	return false
}

// GetHypervisorType is a stub implementation for macOS.
// Always returns empty string (no hypervisor type detected).
func GetHypervisorType() string {
	return ""
}
