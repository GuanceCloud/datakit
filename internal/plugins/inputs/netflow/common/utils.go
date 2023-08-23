// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package common

import (
	"net"
)

// MinUint64 returns the min of the two passed number.
func MinUint64(a uint64, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// MaxUint64 returns the max of the two passed number.
func MaxUint64(a uint64, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

// MaxUint32 returns the max of the two passed number.
func MaxUint32(a uint32, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

// MaxUint16 returns the max of the two passed number.
func MaxUint16(a uint16, b uint16) uint16 {
	if a > b {
		return a
	}
	return b
}

// IPBytesToString convert IP in []byte to string.
func IPBytesToString(ip []byte) string {
	if len(ip) == 0 {
		return ""
	}
	return net.IP(ip).String()
}
