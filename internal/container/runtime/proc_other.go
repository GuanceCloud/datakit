// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !linux
// +build !linux

package runtime

import (
	"fmt"
)

var errNotSupported = fmt.Errorf("non-linux systems are not supported")

func newCPUInfo(_ string) (cpuInfo, error) {
	return nil, errNotSupported
}

func newMemInfo(_ string) (memInfo, error) {
	return nil, errNotSupported
}

func newNetworkStat(_ string, _ int) (networkStat, error) {
	return nil, errNotSupported
}
