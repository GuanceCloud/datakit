// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !(windows && 386)
// +build !windows !386

package refertable

import (
	"database/sql"
	"reflect"
	"testing"

	_ "modernc.org/sqlite"
)

func Test_buildCreateTableStmt(t *testing.T) {
	type args struct {
		table referTable
	}
	tests := []struct {
		name string
		in   args
		want string
	}{
		{
			name: "normal",
			in: args{referTable{
				TableName:  "test_table",
				ColumnName: []string{"c1", "c2"},
				ColumnType: []string{columnTypeStr, columnTypeInt},
			}},
			want: "CREATE TABLE test_table (c1 TEXT, c2 INTEGER)",
		},
		{
			name: "normal",
			in: args{referTable{
				TableName:  "employee",
				ColumnName: []string{"c1", "c2"},
				ColumnType: []string{columnTypeFloat, columnTypeBool},
			}},
			want: "CREATE TABLE employee (c1 REAL, c2 NUMERIC)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildCreateTableStmt(&tt.in.table); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildCreateTableStmt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildInsertIntoStmts(t *testing.T) {
	type args struct {
		table referTable
	}
	tests := []struct {
		name string
		in   args
		want string
	}{
		{
			name: "normal",
			in: args{referTable{
				TableName:  "test_table",
				ColumnName: []string{"c1", "c2"},
				ColumnType: []string{columnTypeStr, columnTypeInt},
				RowData:    [][]any{{"fiona", 20}, {"michael", 25}},
			}},
			want: "INSERT INTO test_table (c1, c2) VALUES (?, ?)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildInsertIntoStmts(&tt.in.table); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildInsertIntoStmts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildSelectStmt(t *testing.T) {
	type args struct {
		tableName string
		colName   []string
		colValue  []any
		kGet      []string
	}
	tests := []struct {
		name string
		in   args
		want string
	}{
		{
			name: "normal",
			in: args{
				tableName: "test_table",
				colName:   []string{},
				colValue:  []any{},
				kGet:      []string{"id"},
			},
			want: "SELECT id FROM test_table",
		},
		{
			name: "normal",
			in: args{
				tableName: "test_table",
				colName:   []string{"weight"},
				colValue:  []any{50.5},
				kGet:      []string{"id"},
			},
			want: "SELECT id FROM test_table WHERE weight = ?",
		},
		{
			name: "normal",
			in: args{
				tableName: "test_table",
				colName:   []string{"name", "age"},
				colValue:  []any{"fiona", 21},
				kGet:      []string{"id", "name", "age"},
			},
			want: "SELECT id, name, age FROM test_table WHERE name = ? AND age = ?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildSelectStmt(tt.in.tableName, tt.in.colName, tt.in.kGet); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildSelectStmt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlReferTablesSqlite_query(t *testing.T) {
	p := &PlReferTablesSqlite{}
	d, err := sql.Open("sqlite", ":memory:")
	p.db = d
	if err != nil {
		t.Fatal(err)
	}
	if err := p.db.Ping(); err != nil {
		t.Fatal(err)
	}
	r := p.db.QueryRow("select sqlite_version()")
	var ver string
	r.Scan(&ver)
	t.Logf("sqlite version: %s", ver)
	if _, err := p.db.Exec("CREATE TABLE test(id INTEGER, name TEXT, age INTEGER, alive NUMERIC, grade REAL)"); err != nil {
		t.Fatal(err)
	}
	if _, err := p.db.Exec("INSERT INTO test(id, name, age, alive, grade) VALUES(23, \"Michael\", 25, true, 99.5)"); err != nil {
		t.Fatal(err)
	}
	if _, err := p.db.Exec("INSERT INTO test(id, name) VALUES(23, \"Jimmy\")"); err != nil {
		t.Fatal(err)
	}
	if _, err := p.db.Exec("INSERT INTO test(id, name, age, alive, grade) VALUES(8, \"Kobe\", 40, false, 99.0)"); err != nil {
		t.Fatal(err)
	}

	type fields struct {
		tableNames []string
	}
	type args struct {
		tableName string
		colName   []string
		colValue  []any
		kGet      []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]any
		want1  bool
	}{
		{
			name: "select by id",
			args: args{
				tableName: "test",
				colName:   []string{"id"},
				colValue:  []any{8},
				kGet:      []string{"id", "name"},
			},
			want:  map[string]any{"id": int64(8), "name": "Kobe"},
			want1: true,
		},
		{
			name: "select by id and name",
			args: args{
				tableName: "test",
				colName:   []string{"id", "name"},
				colValue:  []any{23, "Michael"},
				kGet:      []string{},
			},
			want:  map[string]any{"id": int64(23), "name": "Michael", "age": int64(25), "alive": int64(1), "grade": 99.5},
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := p.query(tt.args.tableName, tt.args.colName, tt.args.colValue, tt.args.kGet)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PlReferTablesSqlite.query() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("PlReferTablesSqlite.query() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
