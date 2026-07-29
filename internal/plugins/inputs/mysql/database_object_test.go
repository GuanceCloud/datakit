// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestRegexNameFilter(t *testing.T) {
	t.Run("empty includes all", func(t *testing.T) {
		filter, err := newRegexNameFilter(nil, nil)
		require.NoError(t, err)
		require.True(t, filter.allow("orders"))
	})

	t.Run("exclude takes precedence", func(t *testing.T) {
		filter, err := newRegexNameFilter([]string{`^app_`}, []string{`_tmp$`})
		require.NoError(t, err)
		require.True(t, filter.allow("app_orders"))
		require.False(t, filter.allow("app_orders_tmp"))
		require.False(t, filter.allow("system_orders"))
	})

	t.Run("invalid pattern", func(t *testing.T) {
		_, err := newRegexNameFilter([]string{"["}, nil)
		require.Error(t, err)
	})
}

func TestGetMysqlDatabasesAppliesFilterBeforeCollectingTables(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	ipt := defaultInput()
	ipt.db = db
	ipt.timeoutDuration = time.Second
	ipt.Object.CollectSchemas.IncludeDatabases = []string{`^app_db$`}
	require.NoError(t, ipt.Object.CollectSchemas.initFilters())

	databaseRows := sqlmock.NewRows([]string{
		"name",
		"default_character_set_name",
		"default_collation_name",
	}).
		AddRow("app_db", "utf8mb4", "utf8mb4_general_ci").
		AddRow("archive_db", "utf8mb4", "utf8mb4_general_ci")
	mock.ExpectQuery(`FROM information_schema\.SCHEMATA`).WillReturnRows(databaseRows)
	mock.ExpectQuery(`WHERE TABLE_SCHEMA = 'app_db'`).WillReturnRows(
		sqlmock.NewRows([]string{"name", "engine", "row_format", "create_time"}),
	)

	databases, err := ipt.getMysqlDatabases()
	require.NoError(t, err)
	require.Len(t, databases, 1)
	require.Equal(t, "app_db", databases[0].Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetDatabaseTablesFiltersBeforePopulatingMetadata(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	ipt := defaultInput()
	ipt.db = db
	ipt.timeoutDuration = time.Second
	ipt.Interval.Duration = time.Second
	ipt.Object.CollectSchemas.IncludeTables = []string{`^orders$`}
	require.NoError(t, ipt.Object.CollectSchemas.initFilters())

	tableRows := sqlmock.NewRows([]string{"name", "engine", "row_format", "create_time"}).
		AddRow("audit_log", "InnoDB", "Dynamic", "2026-07-27 10:00:00")
	mock.ExpectQuery(`FROM information_schema\.TABLES`).WillReturnRows(tableRows)

	tables, err := ipt.getDatabaseTables("app_db")
	require.NoError(t, err)
	require.Empty(t, tables)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDefaultObjectSchemaCollectionEnabled(t *testing.T) {
	require.True(t, defaultInput().Object.CollectSchemas.Enabled)
}

func TestInvalidObjectSchemaFilterDisablesOnlySchemaCollection(t *testing.T) {
	ipt := defaultInput()
	ipt.Object.CollectSchemas.IncludeTables = []string{"["}

	ipt.initObjectCollectSchemas()

	require.True(t, ipt.Object.Enable)
	require.False(t, ipt.Object.CollectSchemas.Enabled)
}
