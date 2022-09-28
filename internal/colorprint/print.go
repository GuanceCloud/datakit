// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package colorprint output with colors
package colorprint

import (
	"fmt"

	"github.com/fatih/color"
)

func Infof(fmtstr string, args ...interface{}) {
	color.Set(color.FgGreen)
	Output(fmtstr, args...)
	color.Unset()
}

func Warnf(fmtstr string, args ...interface{}) {
	color.Set(color.FgYellow)
	Output(fmtstr, args...)
	color.Unset()
}

func Errorf(fmtstr string, args ...interface{}) {
	color.Set(color.FgRed)
	Output(fmtstr, args...)
	color.Unset()
}

func Output(fmtstr string, args ...interface{}) {
	fmt.Printf(fmtstr, args...)
}
