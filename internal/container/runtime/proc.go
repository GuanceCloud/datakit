// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

// The linuxproc cannot be compiled on non-Linux systems, so export the interface.

type cpuInfo interface {
	cores() int
}

type memInfo interface {
	total() int64
}

type networkStat interface {
	rxBytes(skipLoopback bool) int64
	txBytes(skipLoopback bool) int64
}
