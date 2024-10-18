// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !(linux && amd64) && !(linux && arm64)
// +build !linux !amd64
// +build !linux !arm64

package utils

import "github.com/GuanceCloud/cliutils/logger"

func Download(log *logger.Logger, opt ...Opt) error {
	return nil
}

func Install(opt ...Opt) error {
	return nil
}

func Uninstall(opt ...Opt) error {
	return nil
}
