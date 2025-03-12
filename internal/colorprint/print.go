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

// Infof same as fmt.Printf but with green color.
func Infof(fmtstr string, args ...interface{}) {
	color.Set(color.FgGreen)
	Printf(fmtstr, args...)
	color.Unset()
}

// Warnf same as fmt.Printf but with yellow color.
func Warnf(fmtstr string, args ...interface{}) {
	color.Set(color.FgYellow)
	Printf(fmtstr, args...)
	color.Unset()
}

// Errorf same as fmt.Printf but with red color.
func Errorf(fmtstr string, args ...interface{}) {
	color.Set(color.FgRed)
	Printf(fmtstr, args...)
	color.Unset()
}

// Printf same as fmt.Printf.
func Printf(fmtstr string, args ...any) {
	fmt.Printf(fmtstr, args...)
}

// Println same as fmt.Println.
func Println(args ...any) {
	fmt.Println(args...)
}
