// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !linux && !darwin
// +build !linux,!darwin

package server

import "runtime"

func NewDB() *DB {
	l.Panicf("not supported: %s/%s", runtime.GOOS, runtime.GOARCH)
	return &DB{}
}
