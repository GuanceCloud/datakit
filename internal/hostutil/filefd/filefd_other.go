// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !linux
// +build !linux

package filefd

func collect() (map[string]int64, error) {
	info := make(map[string]int64)
	info["allocated"] = -1
	info["maximum"] = -1
	return info, nil
}
