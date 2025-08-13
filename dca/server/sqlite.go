// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package server

import (
	"database/sql/driver"
	"fmt"
	"regexp"

	sqlite3 "modernc.org/sqlite"
)

func NewDB() *DB {
	return &DB{}
}

//nolint:gochecknoinits
func init() {
	sqlite3.MustRegisterDeterministicScalarFunction(
		"regexp",
		2,
		func(ctx *sqlite3.FunctionContext, args []driver.Value) (driver.Value, error) {
			var s1 string
			var s2 string

			switch arg0 := args[0].(type) {
			case string:
				s1 = arg0
			default:
				s1 = ""
			}

			switch arg1 := args[1].(type) {
			case string:
				s2 = arg1
			default:
				s2 = ""
			}

			matched, err := regexp.MatchString(s1, s2)
			if err != nil {
				return nil, fmt.Errorf("bad regular expression: %w", err)
			}

			return matched, nil
		},
	)
}
