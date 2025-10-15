// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kingbase

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestMapToStruct(t *testing.T) {
	type connectionResult struct {
		ActiveConnections int64 `db:"active_connections"`
		IdleConnections   int64 `db:"idle_connections"`
		MaxConnections    int64 `db:"max_connections"`
	}

	data := map[string]any{
		"ACTIVE_CONNECTIONS": 3,
		"Idle_Connections":   7,
		"max_connections":    100,
	}

	var result connectionResult
	err := mapToStruct(data, &result)
	require.NoError(t, err)

	require.Equal(t, int64(3), result.ActiveConnections)
	require.Equal(t, int64(7), result.IdleConnections)
	require.Equal(t, int64(100), result.MaxConnections)
}

func TestKingbaseUppercaseColumns(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	rows := sqlmock.NewRows([]string{
		"ACTIVE_CONNECTIONS", "IDLE_CONNECTIONS", "MAX_CONNECTIONS",
	}).AddRow(3, 7, 100)

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	var result connectionResult

	r, err := sqlxDB.Queryx("SELECT * FROM sys_stat_activity")
	require.NoError(t, err)
	defer r.Close()

	if r.Next() {
		rowMap := map[string]any{}
		err = r.MapScan(rowMap)
		require.NoError(t, err)

		err = mapToStruct(rowMap, &result)
		require.NoError(t, err)
	}

	// t.Logf("修复后结果：Active=%d, Idle=%d, Max=%d",
	// 	result.ActiveConnections, result.IdleConnections, result.MaxConnections)
	require.Equal(t, int64(3), result.ActiveConnections)
	require.Equal(t, int64(7), result.IdleConnections)
	require.Equal(t, int64(100), result.MaxConnections)
}

// 测试当数据库字段为 NULL 时
func TestMapToStructWithNulls(t *testing.T) {
	type testRow struct {
		QueryID        string  `db:"queryid"`
		ActiveConn     int64   `db:"active_connections"`
		IdleConn       int64   `db:"idle_connections"`
		MaxConnections int64   `db:"max_connections"`
		AvgLatency     float64 `db:"avg_latency"`
	}
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	rows := sqlmock.NewRows([]string{
		"QUERYID", "ACTIVE_CONNECTIONS", "IDLE_CONNECTIONS", "MAX_CONNECTIONS", "AVG_LATENCY",
	}).AddRow(nil, 10, nil, 200, nil) // string=NULL, int=NULL, float=NULL

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	rowMap := map[string]any{}
	err = sqlxDB.QueryRowx("SELECT * FROM sys_stat_activity").MapScan(rowMap)
	require.NoError(t, err)

	var result testRow
	err = mapToStruct(rowMap, &result)
	require.NoError(t, err)

	// fmt.Printf("QueryID=%q, Active=%d, Idle=%d, Max=%d, Latency=%.2f\n",
	// 	result.QueryID, result.ActiveConn, result.IdleConn, result.MaxConnections, result.AvgLatency)

	require.Equal(t, "", result.QueryID)           // NULL string → ""
	require.Equal(t, int64(10), result.ActiveConn) // 正常 int
	require.Equal(t, int64(0), result.IdleConn)    // NULL int → 0
	require.Equal(t, int64(200), result.MaxConnections)
	require.Equal(t, 0.0, result.AvgLatency) // NULL float → 0.0
}

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"V008R006C008B0014", "V8R6"},
		{"V008R006C008M060B0003", "V8R6"},
		{"V8R6C008B0014", "V8R6"},
		{"V09R001", "V9R1"},
		{"V0R0", "V0R0"},
		{"invalid", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := extractVersion(tt.input)
			if tt.expected == "" {
				if err == nil {
					t.Errorf("expected error for input %q, got %q", tt.input, got)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tt.input, err)
				} else if got != tt.expected {
					t.Errorf("input %q: expected %q, got %q", tt.input, tt.expected, got)
				}
			}
		})
	}
}
