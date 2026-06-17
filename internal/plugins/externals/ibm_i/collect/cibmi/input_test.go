// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cibmi

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ibm_i/collect/ccommon"
)

func TestNewInputUsesPasswordEnvironmentAndODBCFlags(t *testing.T) {
	t.Setenv(passwordEnv, "password-from-env")

	ipt, err := NewInput(&ccommon.Option{
		Host:                 "ibmi.example.com",
		Username:             "datakit",
		Password:             "password-from-args",
		Driver:               "IBM i Access ODBC Driver 64-bit",
		Interval:             "60s",
		QueryTimeout:         "10s",
		JobQueryTimeout:      "120s",
		SystemMQQueryTimeout: "70s",
		MetricEnabled:        "false",
		Queries:              []string{queryCPUUsage, queryMessageQueue},
		SeverityThreshold:    70,
		MessageQueues:        []string{"qsysopr"},
		Log:                  filepath.Join(t.TempDir(), "ibm_i.log"),
	})

	require.NoError(t, err)
	require.Equal(t, "password-from-env", ipt.Password)
	require.False(t, ipt.Metric.Enabled)
	require.Equal(t, 10*time.Second, ipt.QueryTimeout)
	require.Equal(t, 120*time.Second, ipt.JobQueryTimeout)
	require.Equal(t, 70*time.Second, ipt.SystemMQQueryTimeout)
	require.Equal(t, []string{queryCPUUsage, queryMessageQueue}, ipt.Queries)
	require.Equal(t, []string{"QSYSOPR"}, ipt.MessageQueues)
	require.Equal(t, "Driver={IBM i Access ODBC Driver 64-bit};System=ibmi.example.com;UID=datakit;PWD=password-from-env;", ipt.connectionString())
}

func TestSetupDefaultsAndConnectionString(t *testing.T) {
	ipt := &Input{
		DSN:    "HOSTNAME=ibmi.example.com;UID=datakit;PWD=secret;",
		Metric: MetricConfig{Enabled: true},
	}

	require.NoError(t, ipt.setup())
	require.Equal(t, "ibmi.example.com", ipt.Host)
	require.Equal(t, defaultDriver, ipt.Driver)
	require.Equal(t, defaultODBCDriver, ipt.ODBCDriver)
	require.Equal(t, 50, ipt.SeverityThreshold)
	require.Equal(t, 30*time.Second, ipt.QueryTimeout)
	require.Equal(t, 240*time.Second, ipt.JobQueryTimeout)
	require.Equal(t, 80*time.Second, ipt.SystemMQQueryTimeout)
	require.Equal(t, defaultQueries, ipt.Queries)
	require.Equal(t, "HOSTNAME=ibmi.example.com;UID=datakit;PWD=secret;", ipt.connectionString())
}

func TestQuerySelectionAndMessageQueueSQL(t *testing.T) {
	queries, err := normalizeQueries([]string{queryCPUUsage, queryCPUUsage, queryMessageQueue})
	require.NoError(t, err)
	require.Equal(t, []string{queryCPUUsage, queryMessageQueue}, queries)

	_, err = normalizeQueries([]string{"unknown_custom_query"})
	require.Error(t, err)

	query, args := messageQueueSQL(80, []string{"QSYSOPR", "QSYSMSG"})
	require.Contains(t, query, "SEVERITY >= 80")
	require.Contains(t, query, "MESSAGE_QUEUE_NAME IN (?,?)")
	require.NotContains(t, query, "HISTORY_LOG_INFO")
	require.NotContains(t, query, "JOBLOG_INFO")
	require.Equal(t, []any{"QSYSOPR", "QSYSMSG"}, args)
}

func TestCollectMetricPointsMapsDatadogQueries(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ipt := defaultInput()
	ipt.Host = "ibmi.example.com"
	ipt.User = "datakit"
	ipt.Tags = map[string]string{"env": "prod"}
	ipt.MessageQueues = []string{"QSYSOPR"}
	ipt.db = sqlx.NewDb(db, "sqlmock")
	require.NoError(t, ipt.setup())

	ts := time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC)
	mock.ExpectQuery("ENV_SYS_INFO").
		WillReturnRows(sqlmock.NewRows([]string{"HOST_NAME", "OS_VERSION", "OS_RELEASE"}).
			AddRow("LPAR1", "7", "4"))
	mock.ExpectQuery("SYSDISKSTAT").
		WillReturnRows(sqlmock.NewRows([]string{
			"ASP_NUMBER", "UNIT_NUMBER", "UNIT_TYPE", "SERIAL_NUMBER", "UNIT_STORAGE_CAPACITY",
			"UNIT_SPACE_AVAILABLE", "PERCENT_USED",
		}).AddRow("1", "1", "SSD", "SN1", 1000, 450, 55.0))
	mock.ExpectQuery("SYSDISKSTAT").
		WillReturnRows(sqlmock.NewRows([]string{
			"ASP_NUMBER", "UNIT_NUMBER", "UNIT_TYPE", "SERIAL_NUMBER", "RESOURCE_NAME",
			"ELAPSED_PERCENT_BUSY", "ELAPSED_IO_REQUESTS",
		}).AddRow("1", "1", "SSD", "SN1", "DD001", 12.5, 42.0))
	mock.ExpectQuery("SYSTEM_STATUS").
		WillReturnRows(sqlmock.NewRows([]string{
			"AVERAGE_CPU_UTILIZATION", "CONFIGURED_CPUS", "CURRENT_CPU_CAPACITY", "PARTITION_ID",
			"ELAPSED_CPU_SHARED", "NORMALIZED_CPU_USAGE",
		}).AddRow(8.5, 2.0, 2.0, "1", 0.5, 17.0))
	mock.ExpectQuery("JOB_INFO").
		WillReturnRows(sqlmock.NewRows([]string{
			"JOB_ID", "JOB_USER", "JOB_NAME", "JOB_SUBSYSTEM", "JOB_STATUS", "JOB_QUEUE_LIBRARY",
			"JOB_QUEUE_NAME", "JOB_QUEUE_STATUS", "STATUS", "JOBQ_DURATION",
		}).AddRow("123456", "DKUSER", "DKJOB", "QBATCH", "JOBQ", "QGPL", "QBATCH", "RELEASED", 1, 30.0))
	mock.ExpectQuery("ACTIVE_JOB_INFO").
		WillReturnRows(sqlmock.NewRows([]string{
			"JOB_ID", "JOB_USER", "JOB_NAME", "SUBSYSTEM", "JOB_STATUS", "JOB_ACTIVE_STATUS",
			"STATUS", "CPU_USAGE", "CPU_USAGE_PCT", "ACTIVE_DURATION",
		}).AddRow("123456", "DKUSER", "DKJOB", "QSYSWRK", "ACTIVE", "RUN", 1, 0.1, 10.5, 120.0))
	mock.ExpectQuery("ACTIVE_JOB_INFO").
		WillReturnRows(sqlmock.NewRows([]string{
			"JOB_ID", "JOB_USER", "JOB_NAME", "SUBSYSTEM", "JOB_STATUS", "MEMORY_POOL", "TEMPORARY_STORAGE",
		}).AddRow("123456", "DKUSER", "DKJOB", "QSYSWRK", "RUN", "BASE", 2048))
	mock.ExpectQuery("MEMORY_POOL_INFO").
		WillReturnRows(sqlmock.NewRows([]string{
			"POOL_NAME", "SUBSYSTEM_NAME", "CURRENT_SIZE", "RESERVED_SIZE", "DEFINED_SIZE",
		}).AddRow("BASE", "QSYSWRK", 886.15, 12.5, 1024.75))
	mock.ExpectQuery("SUBSYSTEM_INFO").
		WillReturnRows(sqlmock.NewRows([]string{
			"SUBSYSTEM_DESCRIPTION", "ACTIVE", "CURRENT_ACTIVE_JOBS",
		}).AddRow("QSYSWRK", 1, 20))
	mock.ExpectQuery("JOB_QUEUE_INFO").
		WillReturnRows(sqlmock.NewRows([]string{
			"JOB_QUEUE_NAME", "JOB_QUEUE_STATUS", "SUBSYSTEM_NAME", "NUMBER_OF_JOBS",
			"RELEASED_JOBS", "SCHEDULED_JOBS", "HELD_JOBS",
		}).AddRow("QBATCH", "RELEASED", "QBATCH", 5, 3, 1, 1))
	mock.ExpectQuery("MESSAGE_QUEUE_INFO").
		WithArgs("QSYSOPR").
		WillReturnRows(sqlmock.NewRows([]string{
			"MESSAGE_QUEUE_NAME", "MESSAGE_QUEUE_LIBRARY", "MESSAGE_QUEUE_SIZE", "CRITICAL_MESSAGE_QUEUE_SIZE",
		}).AddRow("QSYSOPR", "QSYS", 7, 2))

	pts, logPts, err := ipt.collectMetricPoints(context.Background(), ts)
	require.NoError(t, err)
	require.Empty(t, logPts)
	require.Len(t, pts, 11)
	require.Equal(t, map[string]int{measurementName: 10, "collector": 1}, pointCounts(pts))

	system := firstPointWithField(t, pts, measurementName, "system_cpu_usage")
	require.Equal(t, "prod", system.Tags().Get("env").GetS())
	require.Equal(t, "ibmi.example.com", system.Tags().Get("host").GetS())
	require.InDelta(t, 8.5, system.Fields().Get("system_cpu_usage").GetF(), 0.001)
	require.InDelta(t, 2.0, system.Fields().Get("system_configured_cpus").GetF(), 0.001)

	job := firstPointWithField(t, pts, measurementName, "job_queue_duration")
	require.Equal(t, "123456", job.Tags().Get("job_id").GetS())
	require.Equal(t, int64(1), job.Fields().Get("job_status_value").GetI())

	mq := firstPointWithField(t, pts, measurementName, "message_queue_critical_size")
	require.Equal(t, "QSYSOPR", mq.Tags().Get("message_queue_name").GetS())
	require.Equal(t, int64(2), mq.Fields().Get("message_queue_critical_size").GetI())

	pool := firstPointWithField(t, pts, measurementName, "pool_size")
	require.Equal(t, "BASE", pool.Tags().Get("pool_name").GetS())
	require.InDelta(t, 886.15, pool.Fields().Get("pool_size").GetF(), 0.001)
	require.InDelta(t, 12.5, pool.Fields().Get("pool_reserved_size").GetF(), 0.001)
	require.InDelta(t, 1024.75, pool.Fields().Get("pool_defined_size").GetF(), 0.001)

	require.Equal(t, int64(1), firstPoint(t, pts, "collector").Fields().Get("up").GetI())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCollectMetricPointsSkips73OnlyQueriesOn72(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ipt := defaultInput()
	ipt.Host = "ibmi.example.com"
	ipt.User = "datakit"
	ipt.Queries = []string{queryDiskUsage, querySubsystem}
	ipt.db = sqlx.NewDb(db, "sqlmock")
	require.NoError(t, ipt.setup())

	mock.ExpectQuery("ENV_SYS_INFO").
		WillReturnRows(sqlmock.NewRows([]string{"HOST_NAME", "OS_VERSION", "OS_RELEASE"}).
			AddRow("LPAR1", "7", "2"))
	mock.ExpectQuery("SYSDISKSTAT").
		WillReturnRows(sqlmock.NewRows([]string{
			"ASP_NUMBER", "UNIT_NUMBER", "UNIT_TYPE", "UNIT_STORAGE_CAPACITY",
			"UNIT_SPACE_AVAILABLE", "PERCENT_USED",
		}).AddRow("1", "1", "SSD", 1000, 450, 55.0))

	pts, _, err := ipt.collectMetricPoints(context.Background(), time.Now())
	require.NoError(t, err)
	require.Equal(t, map[string]int{measurementName: 1, "collector": 1}, pointCounts(pts))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCollectorUpIsZeroOnQueryFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ipt := defaultInput()
	ipt.Host = "ibmi.example.com"
	ipt.User = "datakit"
	ipt.Queries = []string{queryCPUUsage}
	ipt.db = sqlx.NewDb(db, "sqlmock")
	require.NoError(t, ipt.setup())

	mock.ExpectQuery("ENV_SYS_INFO").
		WillReturnRows(sqlmock.NewRows([]string{"HOST_NAME", "OS_VERSION", "OS_RELEASE"}).
			AddRow("LPAR1", "7", "4"))
	mock.ExpectQuery("SYSTEM_STATUS").WillReturnError(errors.New("query failed"))

	pts, _, err := ipt.collectMetricPoints(context.Background(), time.Now())
	require.Error(t, err)
	require.Len(t, pts, 1)
	require.Equal(t, "collector", pts[0].Name())
	require.Equal(t, int64(0), pts[0].Fields().Get("up").GetI())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCollectorUpStaysOneOnOptionalQueryFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ipt := defaultInput()
	ipt.Host = "ibmi.example.com"
	ipt.User = "datakit"
	ipt.Queries = []string{queryCPUUsage, queryMessageQueue}
	ipt.db = sqlx.NewDb(db, "sqlmock")
	require.NoError(t, ipt.setup())

	mock.ExpectQuery("ENV_SYS_INFO").
		WillReturnRows(sqlmock.NewRows([]string{"HOST_NAME", "OS_VERSION", "OS_RELEASE"}).
			AddRow("LPAR1", "7", "4"))
	mock.ExpectQuery("SYSTEM_STATUS").
		WillReturnRows(sqlmock.NewRows([]string{
			"AVERAGE_CPU_UTILIZATION", "CONFIGURED_CPUS", "CURRENT_CPU_CAPACITY", "PARTITION_ID",
			"ELAPSED_CPU_SHARED", "NORMALIZED_CPU_USAGE",
		}).AddRow(8.5, 2.0, 2.0, "1", 0.5, 17.0))
	mock.ExpectQuery("MESSAGE_QUEUE_INFO").WillReturnError(errors.New("query failed"))

	pts, _, err := ipt.collectMetricPoints(context.Background(), time.Now())
	require.Error(t, err)
	require.Len(t, pts, 2)
	require.Equal(t, int64(1), firstPoint(t, pts, "collector").Fields().Get("up").GetI())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConnectUntilReadyRetries(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	ipt := defaultInput()
	attempts := 0
	ipt.retryPeriod = time.Millisecond
	ipt.connectDBFunc = func() (*sqlx.DB, error) {
		attempts++
		if attempts == 1 {
			return nil, errors.New("IBM i is not ready")
		}
		return sqlx.NewDb(db, "sqlmock"), nil
	}
	stop := make(chan os.Signal)

	require.True(t, ipt.connectUntilReady(stop))
	require.Equal(t, 2, attempts)
	mock.ExpectClose()
	ipt.closeDB()
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportFailureReconnectsAfterLostConnection(t *testing.T) {
	oldDB, oldMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	newDB, newMock, err := sqlmock.New()
	require.NoError(t, err)

	ipt := defaultInput()
	ipt.db = sqlx.NewDb(oldDB, "sqlmock")
	oldMock.ExpectPing().WillReturnError(errors.New("connection lost"))
	oldMock.ExpectClose()
	ipt.connectDBFunc = func() (*sqlx.DB, error) {
		return sqlx.NewDb(newDB, "sqlmock"), nil
	}

	ipt.reportFailure("collect metric", errors.New("query failed"))
	require.Nil(t, ipt.db)
	require.True(t, ipt.ensureConnection())
	require.NotNil(t, ipt.db)

	newMock.ExpectClose()
	ipt.closeDB()
	require.NoError(t, oldMock.ExpectationsWereMet())
	require.NoError(t, newMock.ExpectationsWereMet())
}

func TestHelpers(t *testing.T) {
	require.Equal(t, "123456/DKUSER/DKJOB", joinJobID("123456", "DKUSER", "DKJOB"))
	require.Equal(t, "ibmi.example.com", connValue("HOSTNAME=ibmi.example.com;PORT=446", "hostname"))
	require.Equal(t, "prod", parseTags("env=prod;team=ops")["env"])
	require.Equal(t, queryCPUUsage, queryNameFromSQL(cpuUsageSQL))
	require.Equal(t, queryMessageQueue, queryNameFromSQL("SELECT * FROM QSYS2.MESSAGE_QUEUE_INFO"))
	require.Equal(t, "custom", queryNameFromSQL("SELECT 1"))

	rows := []cpuUsageRow{{}, {}}
	require.Equal(t, 2, selectedRows(&rows))
	require.Equal(t, 0, selectedRows(rows))
}

func pointCounts(pts []*point.Point) map[string]int {
	out := map[string]int{}
	for _, pt := range pts {
		out[pt.Name()]++
	}
	return out
}

func firstPoint(t *testing.T, pts []*point.Point, name string) *point.Point {
	t.Helper()
	for _, pt := range pts {
		if pt.Name() == name {
			return pt
		}
	}
	require.Failf(t, "point not found", "name=%s", name)
	return nil
}

func firstPointWithField(t *testing.T, pts []*point.Point, name, field string) *point.Point {
	t.Helper()
	for _, pt := range pts {
		if pt.Name() == name && pt.Fields().Has(field) {
			return pt
		}
	}
	require.Failf(t, "point not found", "name=%s field=%s", name, field)
	return nil
}
