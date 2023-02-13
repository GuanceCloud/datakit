// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !linux
// +build !linux

package net

// VirtualInterfaces returns virtual network interfaces existing in the system.
func VirtualInterfaces(mockData ...string) (map[string]bool, error) {
	return nil, nil
}
