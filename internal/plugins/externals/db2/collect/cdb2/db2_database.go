// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux && amd64

package cdb2

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/db2/collect/ccommon"
)

type databaseMetrics struct {
	x collectParameters
}

var _ ccommon.DBMetricsCollector = (*databaseMetrics)(nil)

func newDatabaseMetrics(opts ...collectOption) *databaseMetrics {
	m := &databaseMetrics{}

	for _, opt := range opts {
		if opt != nil {
			opt(&m.x)
		}
	}

	return m
}

func (m *databaseMetrics) Collect() (*point.Point, error) {
	l.Debug("Collect entry")

	tf, err := m.collect()
	if err != nil {
		return nil, err
	}

	if tf.IsEmpty() {
		return nil, fmt.Errorf("process empty data")
	}

	opt := &ccommon.BuildPointOpt{
		TF:         tf,
		MetricName: m.x.MetricName,
		Tags:       m.x.Ipt.tags,
		Host:       m.x.Ipt.host,
	}
	return ccommon.BuildPoint(l, opt), nil
}

//nolint:stylecheck
type databaseRowDB struct {
	ApplsCurCons      sql.NullInt64  `db:"APPLS_CUR_CONS"`
	Appls_in_db2      sql.NullInt64  `db:"APPLS_IN_DB2"`
	Connections_top   sql.NullInt64  `db:"CONNECTIONS_TOP"`
	Current_time      time.Time      `db:"CURRENT_TIME"`
	DB_status         sql.NullString `db:"DB_STATUS"`
	Deadlocks         sql.NullInt64  `db:"DEADLOCKS"`
	Last_backup       time.Time      `db:"LAST_BACKUP"`
	Lock_list_in_use  sql.NullInt64  `db:"LOCK_LIST_IN_USE"`
	Lock_timeouts     sql.NullInt64  `db:"LOCK_TIMEOUTS"`
	Lock_wait_time    sql.NullInt64  `db:"LOCK_WAIT_TIME"`
	Lock_waits        sql.NullInt64  `db:"LOCK_WAITS"`
	Num_locks_held    sql.NullInt64  `db:"NUM_LOCKS_HELD"`
	Num_locks_waiting sql.NullInt64  `db:"NUM_LOCKS_WAITING"`
	Rows_modified     sql.NullInt64  `db:"ROWS_MODIFIED"`
	Rows_read         sql.NullInt64  `db:"ROWS_READ"`
	Rows_returned     sql.NullInt64  `db:"ROWS_RETURNED"`
	Total_cons        sql.NullInt64  `db:"TOTAL_CONS"`
}

// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.sql.rtn.doc/doc/r0060769.html
// SELECT appls_cur_cons,appls_in_db2,connections_top,current timestamp AS current_time,db_status,deadlocks,last_backup,lock_list_in_use,lock_timeouts,lock_wait_time,lock_waits,num_locks_held,num_locks_waiting,rows_modified,rows_read,rows_returned,total_cons FROM TABLE(MON_GET_DATABASE(-1)) //nolint:lll
//
//nolint:lll,stylecheck
var (
	DATABASE_TABLE_COLUMNS = []string{
		"appls_cur_cons",
		"appls_in_db2",
		"connections_top",
		"current timestamp AS current_time",
		"db_status",
		"deadlocks",
		"last_backup",
		"lock_list_in_use",
		"lock_timeouts",
		"lock_wait_time",
		"lock_waits",
		"num_locks_held",
		"num_locks_waiting",
		"rows_modified",
		"rows_read",
		"rows_returned",
		"total_cons",
	}
	DATABASE_TABLE = "SELECT " + strings.Join(DATABASE_TABLE_COLUMNS, ",") + " FROM TABLE(MON_GET_DATABASE(-1))"
)

const (
	OK       = 0
	WARNING  = 1
	CRITICAL = 2
	UNKNOWN  = 3
)

var dbStatusMap = map[string]int{
	"ACTIVE":         OK,
	"QUIESCE_PEND":   WARNING,
	"QUIESCED":       CRITICAL,
	"ROLLFWD":        WARNING,
	"ACTIVE_STANDBY": OK,
	"STANDBY":        OK,
}

//nolint:lll,stylecheck
func (m *databaseMetrics) collect() (*ccommon.TagField, error) {
	tf := ccommon.NewTagField()
	rows := []databaseRowDB{}

	err := selectWrapper(m.x.Ipt, &rows, DATABASE_TABLE)
	if err != nil {
		return nil, fmt.Errorf("failed to collect database: %w", err)
	}

	// Only 1 database
	for _, r := range rows {
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001156.html
		// self.service_check(self.SERVICE_CHECK_STATUS, status_to_service_check(r.db_status.Int64), nil)
		if v, ok := dbStatusMap[r.DB_status.String]; ok {
			tf.AddField("status", v, nil)
		} else {
			tf.AddField("status", UNKNOWN, nil)
		}

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001201.html
		tf.AddField("application_active", r.ApplsCurCons.Int64, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001202.html
		tf.AddField("application_executing", r.Appls_in_db2.Int64, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0002225.html
		tf.AddField("connection_max", r.Connections_top.Int64, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001200.html
		tf.AddField("connection_total", r.Total_cons.Int64, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001283.html
		tf.AddField("lock_dead", r.Deadlocks.Int64, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001290.html
		tf.AddField("lock_timeouts", r.Lock_timeouts.Int64, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001281.html
		tf.AddField("lock_active", r.Num_locks_held.Int64, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001296.html
		tf.AddField("lock_waiting", r.Num_locks_waiting.Int64, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001294.html
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001293.html
		var average_lock_wait float64
		if r.Lock_waits.Int64 > 0 {
			average_lock_wait = float64(r.Lock_wait_time.Int64) / float64(r.Lock_waits.Int64)
		} else {
			average_lock_wait = 0
		}
		tf.AddField("lock_wait", average_lock_wait, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001282.html
		// https://www.ibm.com/support/knowledgecenter/en/SSEPGG_11.1.0/com.ibm.db2.luw.admin.config.doc/doc/r0000267.html
		tf.AddField("lock_pages", r.Lock_list_in_use.Int64/4096, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001160.html
		last_backup := r.Last_backup
		var seconds_since_last_backup float64
		if !last_backup.IsZero() {
			seconds_since_last_backup = r.Current_time.Sub(last_backup).Seconds()
		} else {
			seconds_since_last_backup = -1
		}
		tf.AddField("backup_latest", seconds_since_last_backup, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0051568.html
		tf.AddField("row_modified_total", r.Rows_modified.Int64, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001317.html
		tf.AddField("row_reads_total", r.Rows_read.Int64, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0051569.html
		tf.AddField("row_returned_total", r.Rows_returned.Int64, nil)
	}

	return tf, nil
}
