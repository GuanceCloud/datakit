// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !linux
// +build !linux

package conntrack

func Collect() *Info {
	info := &Info{
		Current:       -1,
		Limit:         -1,
		Found:         -1,
		Invalid:       -1,
		Ignore:        -1,
		Insert:        -1,
		InsertFailed:  -1,
		Drop:          -1,
		EarlyDrop:     -1,
		SearchRestart: -1,
	}

	return info
}
