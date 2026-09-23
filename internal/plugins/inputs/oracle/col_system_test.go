// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestCollectPGAParameters(t *testing.T) {
	for _, tc := range []struct {
		name    string
		version string
		pdb     string
		rows    *sqlmock.Rows
		err     error
		want    map[string]interface{}
	}{
		{
			name: "PDB values", version: "21.3", pdb: "XEPDB1",
			rows: sqlmock.NewRows([]string{"NAME", "VALUE", "PDB_NAME"}).
				AddRow("pga_aggregate_target", 536870912, "XEPDB1").
				AddRow("pga_aggregate_limit", 2147483648, "XEPDB1"),
			want: map[string]interface{}{"pga_aggregate_target": int64(536870912), "pga_aggregate_limit": int64(2147483648)},
		},
		{
			name: "root zero values", version: "21.3", pdb: "CDB$ROOT",
			rows: sqlmock.NewRows([]string{"NAME", "VALUE", "PDB_NAME"}).
				AddRow("pga_aggregate_target", 0, "CDB$ROOT").
				AddRow("pga_aggregate_limit", 0, "CDB$ROOT"),
			want: map[string]interface{}{"pga_aggregate_target": int64(0), "pga_aggregate_limit": int64(0)},
		},
		{
			name: "11g without limit or container", version: "11.2",
			rows: sqlmock.NewRows([]string{"NAME", "VALUE"}).AddRow("pga_aggregate_target", 536870912),
			want: map[string]interface{}{"pga_aggregate_target": int64(536870912)},
		},
		{name: "empty", version: "21.3", rows: sqlmock.NewRows([]string{"NAME", "VALUE", "PDB_NAME"})},
		{
			name: "null", version: "21.3",
			rows: sqlmock.NewRows([]string{"NAME", "VALUE", "PDB_NAME"}).AddRow("pga_aggregate_target", nil, "XEPDB1"),
		},
		{name: "query failure", version: "21.3", err: errors.New("ORA-00942: table or view does not exist")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			t.Cleanup(func() {
				mock.ExpectClose()
				require.NoError(t, db.Close())
			})
			ipt := &Input{
				db: sqlx.NewDb(db, "sqlmock"), timeoutDuration: time.Second,
				dbVersion: tc.version, fullVersion: tc.version, cdbName: "xe",
				mergedTags: map[string]string{"server": "localhost:1521", "oracle_service": "XE"},
			}
			query := sqlPGAParameters["default"]
			if tc.version == "11.2" {
				query = sqlPGAParameters["11"]
			}
			expect := mock.ExpectQuery(query)
			if tc.err != nil {
				expect.WillReturnError(tc.err)
			} else {
				expect.WillReturnRows(tc.rows)
			}
			kvs := ipt.collectPGAParameters("oracle_system")
			if tc.want == nil {
				require.Nil(t, kvs)
			} else {
				pt := point.NewPoint("oracle_system", kvs, point.DefaultMetricOptions()...)
				require.Equal(t, len(tc.want), kvs.FieldCount())
				for name, value := range tc.want {
					require.Equal(t, value, pt.Get(name))
				}
				require.Empty(t, pt.GetTag("cdb_name"))
				require.Equal(t, tc.version, pt.GetTag("version"))
				require.Equal(t, "localhost:1521", pt.GetTag("server"))
				require.Equal(t, "XE", pt.GetTag("oracle_service"))
				require.Equal(t, tc.pdb, pt.GetTag("pdb_name"))
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
