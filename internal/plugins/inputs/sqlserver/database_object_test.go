// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

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

func TestGetSchemasAppliesFilterBeforeCollectingTables(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	ipt := defaultInput()
	ipt.db = db
	ipt.timeoutDuration = time.Second
	ipt.Object.CollectSchemas.IncludeSchemas = []string{`^app$`}
	require.NoError(t, ipt.Object.CollectSchemas.initFilters())

	schemaRows := sqlmock.NewRows([]string{"name", "id", "owner_name"}).
		AddRow("app", "1", "dbo").
		AddRow("audit", "2", "dbo")
	mock.ExpectQuery(`(?s)FROM\s+sys\.schemas`).WillReturnRows(schemaRows)
	mock.ExpectQuery(`(?s)FROM\s+sys\.tables\s+WHERE schema_id=1`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name"}),
	)

	schemas, err := ipt.getSchemas("appdb")
	require.NoError(t, err)
	require.Len(t, schemas, 1)
	require.Equal(t, "app", schemas[0].Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTablesAppliesFilterBeforePopulatingMetadata(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	ipt := defaultInput()
	ipt.db = db
	ipt.timeoutDuration = time.Second
	ipt.Object.CollectSchemas.IncludeTables = []string{`^orders$`}
	require.NoError(t, ipt.Object.CollectSchemas.initFilters())

	tableRows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow("1", "audit_log")
	mock.ExpectQuery(`(?s)FROM\s+sys\.tables\s+WHERE schema_id=1`).WillReturnRows(tableRows)

	tables, err := ipt.getTables("appdb", &sqlserverSchema{ID: "1", Name: "app"})
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
