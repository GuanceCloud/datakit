// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !linux
// +build !linux

package process

import (
	pr "github.com/shirou/gopsutil/v3/process"
)

// Not supported on non-linux systems.
func getContainerID(ps *pr.Process) string {
	return ""
}
