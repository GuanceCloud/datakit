// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !linux
// +build !linux

// Package utils contains utils
package utils

import (
	"fmt"
	"os"
)

func RunCmd(name string, args []string, stdout, stderr *os.File) (int, error) {
	return 0, fmt.Errorf("unsupported")
}
