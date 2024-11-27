// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package server

import (
	_ "modernc.org/sqlite"
)

func NewDB() *DB {
	return &DB{}
}
