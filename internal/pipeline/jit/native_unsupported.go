// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !pipeline_jit || !linux || (!amd64 && !arm64)

package jit

import "fmt"

func openRuntime(string, *HostCompat) (Runtime, error) {
	return nil, fmt.Errorf("JIT runtime is not available in this binary; build with -tags pipeline_jit on linux/amd64 or linux/arm64")
}
