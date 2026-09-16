// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

// A zoned DATE and a timezone-free cursor must not cause unchanged SQL statistics
// to be selected again. The next query must use the newest database clock value.
func TestCollectSlowQueryCursorTimezones(t *testing.T) {
	for _, offset := range []int{0, 7, 8, -5} {
		t.Run(fmt.Sprintf("UTC%+d", offset), func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			t.Cleanup(func() {
				mock.ExpectClose()
				require.NoError(t, db.Close())
			})

			feeder := dkio.NewMockedFeeder()
			ipt := &Input{
				db:              sqlx.NewDb(db, "sqlmock"),
				feeder:          feeder,
				slowQueryTime:   10 * time.Millisecond,
				timeoutDuration: time.Second,
			}
			collectedAt := time.Date(2026, 9, 10, 17, 31, 37, 0, time.UTC)
			zone := time.FixedZone("database", offset*60*60)
			active := time.Date(2026, 9, 10, 16, 31, 20, 0, zone)
			latest := active.Add(10 * time.Second)
			columns := []string{"SQL_ID", "LAST_ACTIVE_TIME", "LAST_ACTIVE_TIME_STR", "EXECUTIONS", "ELAPSED_TIME"}

			mock.ExpectQuery(SQLQueryMaxActive).WillReturnRows(
				sqlmock.NewRows([]string{"MAX_LAST_ACTIVE_TIME"}).AddRow("2026-09-10 16:30:42"))
			ipt.collectSlowQuery(collectedAt)
			require.Equal(t, "2026-09-10 16:30:42", ipt.lastActiveTime)

			// Results need not be ordered, and different SQLs can share an active time.
			mock.ExpectQuery(fmt.Sprintf(SQLSlow, 10000, "2026-09-10 16:30:42")).WillReturnRows(
				sqlmock.NewRows(columns).
					AddRow("newest", latest, "2026-09-10 16:31:30", 1, 100690).
					AddRow("same_time", latest, "2026-09-10 16:31:30", 1, 100690).
					AddRow("older", active, "2026-09-10 16:31:20", 1, 100690))
			ipt.collectSlowQuery(collectedAt.Add(time.Minute))
			require.Equal(t, "2026-09-10 16:31:30", ipt.lastActiveTime)
			pts, err := feeder.AnyPoints(time.Second)
			require.NoError(t, err)
			require.Len(t, pts, 3)
			require.Equal(t, active.Format(time.RFC3339Nano), pts[2].Get("last_active_time"),
				"the reported original timestamp must retain its timezone")

			// An idle cycle must query past the previous maximum, excluding old rows.
			mock.ExpectQuery(fmt.Sprintf(SQLSlow, 10000, "2026-09-10 16:31:30")).
				WillReturnRows(sqlmock.NewRows(columns))
			ipt.collectSlowQuery(collectedAt.Add(2 * time.Minute))
			require.NoError(t, mock.ExpectationsWereMet())
			_, err = feeder.AnyPoints(time.Millisecond)
			require.ErrorIs(t, err, dkio.ErrTimeout)

			// The same SQL is still reported when it becomes active again.
			mock.ExpectQuery(fmt.Sprintf(SQLSlow, 10000, "2026-09-10 16:31:30")).WillReturnRows(
				sqlmock.NewRows(columns).
					AddRow("older", active.Add(2*time.Minute), "2026-09-10 16:33:20", 2, 201380))
			ipt.collectSlowQuery(collectedAt.Add(3 * time.Minute))
			require.Equal(t, "2026-09-10 16:33:20", ipt.lastActiveTime)
			pts, err = feeder.AnyPoints(time.Second)
			require.NoError(t, err)
			require.Len(t, pts, 1)
			require.Equal(t, "older", pts[0].GetTag("sql_id"))
			require.Equal(t, int64(2), pts[0].Get("executions"))
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
