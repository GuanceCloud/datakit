// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !linux && !windows
// +build !linux,!windows

package resourcelimit

import (
	"fmt"
	"runtime"
)

func run(opt *ResourceLimitOptions) error {
	return fmt.Errorf("not implemented at os: %s", runtime.GOOS)
}

func info() string {
	return "-"
}
