// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows && 386
// +build windows,386

package refertable

import (
	"database/sql"
)

//nolint:unused
type PlReferTablesSqlite struct {
	tableNames []string
	db         *sql.DB
}

func (p *PlReferTablesSqlite) Query(tableName string, colName []string, colValue []any, kGet []string) (map[string]any, bool) {
	l.Errorf("windows-386 does not support query using SQLite")
	return nil, false
}

func (p *PlReferTablesSqlite) updateAll(tables []referTable) (retErr error) {
	l.Errorf("windows-386 does not support query using SQLite")
	return nil
}

func (p *PlReferTablesSqlite) Stats() *ReferTableStats {
	l.Errorf("windows-386 does not support query using SQLite")
	return nil
}
