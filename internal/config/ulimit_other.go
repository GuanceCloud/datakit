// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package config

func setUlimit(n uint64) error {
	return nil
}

func getUlimit() (soft, hard uint64, err error) {
	return 0, 0, nil
}
