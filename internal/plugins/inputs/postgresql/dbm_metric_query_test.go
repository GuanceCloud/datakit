// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

type mockDBMFeeder struct {
	points map[point.Category][]*point.Point
	errors []string
}

func (m *mockDBMFeeder) Feed(category point.Category, pts []*point.Point, _ ...dkio.FeedOption) error {
	if m.points == nil {
		m.points = make(map[point.Category][]*point.Point)
	}
	m.points[category] = append(m.points[category], pts...)
	return nil
}

func (m *mockDBMFeeder) FeedLastError(err string, _ ...metrics.LastErrorOption) {
	m.errors = append(m.errors, err)
}

func findMetricPointByField(pts []*point.Point, field string) *point.Point {
	for _, pt := range pts {
		if pt.Fields().Get(field) != nil {
			return pt
		}
	}
	return nil
}

func TestBuildDbmMetricPoints(t *testing.T) {
	ipt := defaultInput()
	ipt.mergedTags = map[string]string{
		"server":            "127.0.0.1:5432",
		"database_instance": "pg-test",
	}

	baseline := []dbmMetricRow{
		{
			db:             "app",
			rolname:        "alice",
			queryID:        "42",
			querySignature: "query-1",
			metrics: map[string]float64{
				"calls":           10,
				"rows":            20,
				"total_exec_time": 100,
				"total_plan_time": 40,
			},
			deltas: map[string]float64{},
		},
	}

	pts, reportRows := ipt.buildDbmMetricPoints(baseline, time.Unix(1700000000, 0), 0)
	if assert.Len(t, pts, 1) {
		totalPt := findMetricPointByField(pts, "total_calls")
		if assert.NotNil(t, totalPt) {
			fields := totalPt.Fields()
			if totalCalls := fields.Get("total_calls"); totalCalls != nil {
				assert.Equal(t, int64(10), totalCalls.Raw())
			}
			assert.Nil(t, fields.Get("delta_total_calls"))
			assert.Nil(t, fields.Get("dbm_qps"))
		}
	}
	assert.Empty(t, reportRows)

	current := []dbmMetricRow{
		{
			db:             "app",
			rolname:        "alice",
			queryID:        "42",
			querySignature: "query-1",
			metrics: map[string]float64{
				"calls":           13,
				"rows":            29,
				"total_exec_time": 160,
				"total_plan_time": 55,
			},
			deltas: map[string]float64{},
		},
	}

	pts, reportRows = ipt.buildDbmMetricPoints(current, time.Unix(1700000060, 0), 0)
	if assert.Len(t, pts, 2) {
		pt := findMetricPointByField(pts, "delta_calls")
		if assert.NotNil(t, pt) {
			tags := pt.Tags()
			fields := pt.Fields()

			assert.Equal(t, dbmMetricMeasurementInfo.Name, pt.Name())
			assert.Equal(t, "127.0.0.1:5432", tags.GetTag("server"))
			assert.Equal(t, "pg-test", tags.GetTag("database_instance"))
			assert.Equal(t, "app", tags.GetTag("db"))
			assert.Equal(t, "alice", tags.GetTag("rolname"))
			assert.Equal(t, "42", tags.GetTag("queryid"))
			assert.Equal(t, "query-1", tags.GetTag("query_signature"))

			if deltaCalls := fields.Get("delta_calls"); deltaCalls != nil {
				assert.Equal(t, float64(3), deltaCalls.Raw())
			}
			if deltaRows := fields.Get("delta_rows"); deltaRows != nil {
				assert.Equal(t, float64(9), deltaRows.Raw())
			}
			if deltaExec := fields.Get("delta_total_exec_time"); deltaExec != nil {
				assert.Equal(t, float64(60), deltaExec.Raw())
			}
			if deltaPlan := fields.Get("delta_total_plan_time"); deltaPlan != nil {
				assert.Equal(t, float64(15), deltaPlan.Raw())
			}
			if avgExec := fields.Get("avg_total_exec_time"); avgExec != nil {
				assert.Equal(t, float64(20), avgExec.Raw())
			}
			if avgPlan := fields.Get("avg_total_plan_time"); avgPlan != nil {
				assert.Equal(t, float64(5), avgPlan.Raw())
			}
		}

		totalPt := findMetricPointByField(pts, "total_calls")
		if assert.NotNil(t, totalPt) {
			fields := totalPt.Fields()
			if totalCalls := fields.Get("total_calls"); totalCalls != nil {
				assert.Equal(t, int64(13), totalCalls.Raw())
			}
			if deltaTotalCalls := fields.Get("delta_total_calls"); deltaTotalCalls != nil {
				assert.Equal(t, int64(3), deltaTotalCalls.Raw())
			}
		}
	}
	assert.Len(t, reportRows, 1)
}

func TestBuildDbmMetricPointsSkipsCounterReset(t *testing.T) {
	ipt := defaultInput()
	ipt.dbmMetricValueCache = map[string]*dbmMetricValueCache{
		"query-1": {
			metrics: map[string]float64{
				"calls":           10,
				"total_exec_time": 100,
			},
		},
	}

	pts, reportRows := ipt.buildDbmMetricPoints([]dbmMetricRow{
		{
			db:             "app",
			rolname:        "alice",
			querySignature: "query-1",
			metrics: map[string]float64{
				"calls":           9,
				"total_exec_time": 90,
			},
			deltas: map[string]float64{},
		},
	}, time.Unix(1700000000, 0), 0)

	if assert.Len(t, pts, 1) {
		totalPt := findMetricPointByField(pts, "total_calls")
		if assert.NotNil(t, totalPt) {
			fields := totalPt.Fields()
			if totalCalls := fields.Get("total_calls"); totalCalls != nil {
				assert.Equal(t, int64(9), totalCalls.Raw())
			}
			assert.Nil(t, fields.Get("delta_total_calls"))
		}
	}
	assert.Empty(t, reportRows)
}

func TestBuildDbmMetricPointsTotalCalls(t *testing.T) {
	ipt := defaultInput()
	ipt.mergedTags = map[string]string{
		"server":            "127.0.0.1:5432",
		"database_instance": "pg-test",
	}
	ipt.DbmMetric.Interval.Duration = time.Minute

	firstPts, reportRows := ipt.buildDbmMetricPoints([]dbmMetricRow{
		{
			querySignature: "query-1",
			metrics: map[string]float64{
				"calls": 10,
			},
		},
		{
			querySignature: "query-2",
			metrics: map[string]float64{
				"calls": 5,
			},
		},
	}, time.Unix(1700000000, 0), time.Minute)

	assert.Empty(t, reportRows)
	if assert.Len(t, firstPts, 1) {
		first := findMetricPointByField(firstPts, "total_calls")
		if assert.NotNil(t, first) {
			fields := first.Fields()
			if totalCalls := fields.Get("total_calls"); totalCalls != nil {
				assert.Equal(t, int64(15), totalCalls.Raw())
			}
			assert.Nil(t, fields.Get("delta_total_calls"))
			assert.Nil(t, fields.Get("dbm_qps"))
		}
	}

	secondPts, reportRows := ipt.buildDbmMetricPoints([]dbmMetricRow{
		{
			querySignature: "query-1",
			metrics: map[string]float64{
				"calls": 12,
			},
			deltas: map[string]float64{},
		},
		{
			querySignature: "query-2",
			metrics: map[string]float64{
				"calls": 8,
			},
			deltas: map[string]float64{},
		},
	}, time.Unix(1700000060, 0), time.Minute)

	if assert.Len(t, reportRows, 2) {
		assert.Equal(t, "query-1", reportRows[0].querySignature)
		assert.Equal(t, "query-2", reportRows[1].querySignature)
	}
	if assert.Len(t, secondPts, 3) {
		second := findMetricPointByField(secondPts, "total_calls")
		if assert.NotNil(t, second) {
			fields := second.Fields()
			if totalCalls := fields.Get("total_calls"); totalCalls != nil {
				assert.Equal(t, int64(20), totalCalls.Raw())
			}
			if deltaTotalCalls := fields.Get("delta_total_calls"); deltaTotalCalls != nil {
				assert.Equal(t, int64(5), deltaTotalCalls.Raw())
			}
			if dbmQPS := fields.Get("dbm_qps"); dbmQPS != nil {
				assert.Equal(t, float64(5.0/60.0), dbmQPS.Raw())
			}
		}
	}
}

func TestBuildDbmMetricPointsMergesDuplicateQuerySignatures(t *testing.T) {
	ipt := defaultInput()
	ipt.mergedTags = map[string]string{
		"server":            "127.0.0.1:5432",
		"database_instance": "pg-test",
	}

	baseline := []dbmMetricRow{
		{
			db:             "app",
			rolname:        "alice",
			queryID:        "101",
			querySignature: "query-dup",
			message:        "SHOW cluster_name",
			metrics: map[string]float64{
				"calls":           10,
				"rows":            20,
				"total_exec_time": 100,
			},
			deltas: map[string]float64{},
		},
		{
			db:             "app",
			rolname:        "alice",
			queryID:        "202",
			querySignature: "query-dup",
			message:        "SHOW cluster_name",
			metrics: map[string]float64{
				"calls":           5,
				"rows":            7,
				"total_exec_time": 30,
			},
			deltas: map[string]float64{},
		},
	}

	pts, reportRows := ipt.buildDbmMetricPoints(baseline, time.Unix(1700000000, 0), 0)
	if assert.Len(t, pts, 1) {
		totalPt := findMetricPointByField(pts, "total_calls")
		if assert.NotNil(t, totalPt) {
			fields := totalPt.Fields()
			if totalCalls := fields.Get("total_calls"); totalCalls != nil {
				assert.Equal(t, int64(15), totalCalls.Raw())
			}
		}
	}
	assert.Empty(t, reportRows)

	current := []dbmMetricRow{
		{
			db:             "app",
			rolname:        "alice",
			queryID:        "101",
			querySignature: "query-dup",
			message:        "SHOW cluster_name",
			metrics: map[string]float64{
				"calls":           12,
				"rows":            24,
				"total_exec_time": 120,
			},
			deltas: map[string]float64{},
		},
		{
			db:             "app",
			rolname:        "alice",
			queryID:        "202",
			querySignature: "query-dup",
			message:        "SHOW cluster_name",
			metrics: map[string]float64{
				"calls":           6,
				"rows":            10,
				"total_exec_time": 35,
			},
			deltas: map[string]float64{},
		},
	}

	pts, reportRows = ipt.buildDbmMetricPoints(current, time.Unix(1700000060, 0), 0)
	if assert.Len(t, pts, 2) {
		pt := findMetricPointByField(pts, "delta_calls")
		if assert.NotNil(t, pt) {
			tags := pt.Tags()
			fields := pt.Fields()

			assert.Equal(t, "app", tags.GetTag("db"))
			assert.Equal(t, "alice", tags.GetTag("rolname"))
			assert.Equal(t, "101", tags.GetTag("queryid"))
			assert.Equal(t, "query-dup", tags.GetTag("query_signature"))

			if deltaCalls := fields.Get("delta_calls"); deltaCalls != nil {
				assert.Equal(t, float64(7), deltaCalls.Raw())
			}
			if deltaRows := fields.Get("delta_rows"); deltaRows != nil {
				assert.Equal(t, float64(17), deltaRows.Raw())
			}
			if deltaExec := fields.Get("delta_total_exec_time"); deltaExec != nil {
				assert.Equal(t, float64(90), deltaExec.Raw())
			}
			if avgExec := fields.Get("avg_total_exec_time"); avgExec != nil {
				assert.Equal(t, float64(90.0/7.0), avgExec.Raw())
			}
		}

		totalPt := findMetricPointByField(pts, "total_calls")
		if assert.NotNil(t, totalPt) {
			fields := totalPt.Fields()
			if totalCalls := fields.Get("total_calls"); totalCalls != nil {
				assert.Equal(t, int64(18), totalCalls.Raw())
			}
			if deltaTotalCalls := fields.Get("delta_total_calls"); deltaTotalCalls != nil {
				assert.Equal(t, int64(3), deltaTotalCalls.Raw())
			}
		}
	}
	if assert.Len(t, reportRows, 1) {
		assert.Equal(t, "101", reportRows[0].queryID)
		assert.Equal(t, "SHOW cluster_name", reportRows[0].message)
	}
}

func TestCollectPostgreSQLDbmQueriesDeduplicates(t *testing.T) {
	feeder := &mockDBMFeeder{}
	ipt := defaultInput()
	ipt.feeder = feeder
	ipt.Object.name = "pg-server"
	ipt.databaseInstance = "pg-test"
	ipt.mergedTags = map[string]string{
		"database_instance": "pg-test",
	}

	rows := []dbmMetricRow{
		{
			db:             "app",
			rolname:        "alice",
			queryID:        "42",
			querySignature: "query-1",
			message:        "SELECT $1",
		},
		{
			db:             "app",
			rolname:        "alice",
			queryID:        "43",
			querySignature: "query-2",
			message:        "",
		},
	}

	ipt.collectDbmQueries(rows, time.Unix(1700000000, 0))
	ipt.collectDbmQueries(rows, time.Unix(1700000060, 0))

	objectPts := feeder.points[point.Object]
	if assert.Len(t, objectPts, 1) {
		pt := objectPts[0]
		tags := pt.Tags()
		fields := pt.Fields()

		assert.Equal(t, dbmQueryObjectMeasurementID, pt.Name())
		assert.Equal(t, "pg-server-pg-test-query-1", tags.GetTag("name"))
		assert.Equal(t, "pg-server", tags.GetTag("server"))
		assert.Equal(t, "pg-test", tags.GetTag("database_instance"))
		assert.Equal(t, "PostgreSQL", tags.GetTag("database_type"))
		assert.Equal(t, "app", tags.GetTag("db"))
		assert.Equal(t, "alice", tags.GetTag("rolname"))
		assert.Equal(t, "42", tags.GetTag("queryid"))
		assert.Equal(t, "query-1", tags.GetTag("query_signature"))

		if message := fields.Get("message"); message != nil {
			assert.Equal(t, "SELECT $1", message.Raw())
		}
	}
}

func TestDbmPlanObjectMeasurementName(t *testing.T) {
	info := (&dbmPlanObjectMeasurement{}).Info()

	assert.Equal(t, dbmPlanObjectName, info.Name)
	assert.Equal(t, point.Object, info.Cat)
	assert.Contains(t, info.Fields, "statement")
}

func TestCanExplainStatement(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want bool
	}{
		{name: "select is supported", sql: "SELECT * FROM t", want: true},
		{name: "with is supported", sql: "WITH cte AS (SELECT 1) SELECT * FROM cte", want: true},
		{name: "leading spaces are trimmed", sql: "  update t set id = 1", want: true},
		{name: "autovacuum is skipped", sql: "autovacuum: VACUUM public.t", want: false},
		{name: "helper call is skipped", sql: "SELECT datakit.explain_statement('select 1')", want: false},
		{name: "ddl is not supported", sql: "ALTER TABLE t ADD COLUMN c int", want: false},
		{name: "empty string is not supported", sql: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, canExplainStatement(tt.sql))
		})
	}
}

func TestCheckPlanSignatureRate(t *testing.T) {
	ipt := defaultInput()
	ipt.DbmSample.PlanCacheTTL.Duration = time.Minute

	assert.False(t, checkPlanSignatureRate(ipt, ""))
	assert.True(t, checkPlanSignatureRate(ipt, "plan-1"))

	ipt.recordReportedPlanSignature("plan-1")
	assert.False(t, checkPlanSignatureRate(ipt, "plan-1"))
	assert.True(t, checkPlanSignatureRate(ipt, "plan-2"))
}

func TestShouldCollectPGActivity(t *testing.T) {
	tests := []struct {
		name string
		row  map[string]any
		want bool
	}{
		{
			name: "skip idle client backend",
			row: map[string]any{
				"backend_type": "client backend",
				"state":        "idle",
			},
			want: false,
		},
		{
			name: "keep idle in transaction client backend",
			row: map[string]any{
				"backend_type": "client backend",
				"state":        "idle in transaction",
			},
			want: true,
		},
		{
			name: "keep non client backend idle",
			row: map[string]any{
				"backend_type": "autovacuum worker",
				"state":        "idle",
			},
			want: true,
		},
		{
			name: "missing backend type defaults to client backend",
			row: map[string]any{
				"state": "idle",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, shouldCollectPGActivity(tt.row))
		})
	}
}

func TestGetPGActivityWaitGroup(t *testing.T) {
	tests := []struct {
		name string
		row  map[string]any
		want string
	}{
		{
			name: "active without wait event maps to cpu",
			row: map[string]any{
				"state": "active",
			},
			want: waitGroupCPU,
		},
		{
			name: "blocked lock maps to lock",
			row: map[string]any{
				"state":           "active",
				"wait_event_type": "Lock",
				"wait_event":      "transactionid",
				"blocking_pids":   "42",
			},
			want: waitGroupLock,
		},
		{
			name: "idle in transaction client read maps to network",
			row: map[string]any{
				"state":           "idle in transaction",
				"wait_event_type": "Client",
				"wait_event":      "ClientRead",
			},
			want: waitGroupNetwork,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, getPGActivityWaitGroup(tt.row))
		})
	}
}

func TestCollectSampleActivity(t *testing.T) {
	ipt := defaultInput()
	ipt.mergedTags = map[string]string{
		"server":            "127.0.0.1:5432",
		"database_instance": "pg-test",
	}

	rows := []map[string]any{
		{
			"query_signature":  "query-1",
			"client_hostname":  "client-host",
			"client_port":      "12345",
			"client_addr":      "10.0.0.8",
			"application_name": "psql",
			"usename":          "alice",
			"datname":          "app",
			"state":            "active",
			"pid":              "88",
			"wait_event_type":  "Lock",
			"wait_event":       "transactionid",
			"wait_group":       "Lock",
			"backend_type":     "client backend",
			"statement":        "SELECT * FROM orders WHERE id = $1",
			"blocking_pids":    "99,100",
			"now":              int64(1000),
			"backend_start":    int64(100),
			"query_start":      int64(200),
			"xact_start":       int64(150),
			"state_change":     int64(250),
		},
		{
			"query_signature":  "query-2",
			"application_name": "psql",
			"usename":          "alice",
			"datname":          "app",
			"state":            "idle",
			"pid":              "89",
			"wait_group":       "Other",
			"backend_type":     "client backend",
			"statement":        "SELECT 1",
		},
		{
			"query_signature":  "query-3",
			"application_name": "psql",
			"usename":          "alice",
			"datname":          "app",
			"state":            "idle in transaction",
			"pid":              "90",
			"wait_group":       "Network",
			"backend_type":     "client backend",
			"statement":        "BEGIN",
		},
		{
			"query_signature":  "query-4",
			"application_name": "autovacuum worker",
			"usename":          "postgres",
			"datname":          "app",
			"state":            "idle",
			"pid":              "91",
			"wait_group":       "Other",
			"backend_type":     "autovacuum worker",
			"statement":        "autovacuum: VACUUM public.orders",
		},
	}

	filtered := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		if shouldCollectPGActivity(row) {
			filtered = append(filtered, row)
		}
	}

	pts := ipt.collectSampleActivity(filtered, time.Unix(1700000000, 0))
	if assert.Len(t, pts, 3) {
		signatures := make([]string, 0, len(pts))
		for _, pt := range pts {
			signatures = append(signatures, pt.Tags().GetTag("query_signature"))
		}
		assert.ElementsMatch(t, []string{"query-1", "query-3", "query-4"}, signatures)

		pt := pts[0]
		tags := pt.Tags()
		fields := pt.Fields()

		assert.Equal(t, dbmActivityMeasurementInfo.Name, pt.Name())
		assert.Equal(t, "127.0.0.1:5432", tags.GetTag("server"))
		assert.Equal(t, "pg-test", tags.GetTag("database_instance"))
		assert.Equal(t, "postgresql", tags.GetTag("service"))
		assert.Equal(t, "info", tags.GetTag("status"))
		assert.Equal(t, "query-1", tags.GetTag("query_signature"))
		assert.Equal(t, "transactionid", tags.GetTag("wait_event"))
		assert.Equal(t, "Lock", tags.GetTag("wait_group"))
		assert.Equal(t, "SELECT * FROM orders WHERE id = $1", tags.GetTag("message"))

		if backendStart := fields.Get("backend_start"); backendStart != nil {
			assert.Equal(t, int64(100), backendStart.Raw())
		}
		if queryStart := fields.Get("query_start"); queryStart != nil {
			assert.Equal(t, int64(200), queryStart.Raw())
		}
		if duration := fields.Get("duration"); duration != nil {
			assert.Equal(t, int64(800), duration.Raw())
		}
		if txDuration := fields.Get("tx_duration"); txDuration != nil {
			assert.Equal(t, int64(850), txDuration.Raw())
		}
		if waitDuration := fields.Get("wait_duration"); waitDuration != nil {
			assert.Equal(t, int64(750), waitDuration.Raw())
		}
		if blockingPIDs := fields.Get("blocking_pids"); blockingPIDs != nil {
			assert.Equal(t, "99,100", blockingPIDs.Raw())
		}
	}
}

func TestBuildDbmConnectionPoints(t *testing.T) {
	ipt := defaultInput()
	ipt.mergedTags = map[string]string{
		"server":            "127.0.0.1:5432",
		"database_instance": "pg-test",
	}

	pts := ipt.buildDbmConnectionPoints([]dbmConnectionRow{
		{
			applicationName: "psql",
			state:           "active",
			usename:         "alice",
			datname:         "app",
			connectionCount: 3,
		},
	}, time.Unix(1700000000, 0))
	if assert.Len(t, pts, 1) {
		pt := pts[0]
		tags := pt.Tags()
		fields := pt.Fields()

		assert.Equal(t, metricNamePostgreSQLDbmConnection, pt.Name())
		assert.Equal(t, "127.0.0.1:5432", tags.GetTag("server"))
		assert.Equal(t, "pg-test", tags.GetTag("database_instance"))
		assert.Equal(t, "psql", tags.GetTag("application_name"))
		assert.Equal(t, "active", tags.GetTag("state"))
		assert.Equal(t, "alice", tags.GetTag("usename"))
		assert.Equal(t, "app", tags.GetTag("db"))

		if connectionCount := fields.Get("connection_count"); connectionCount != nil {
			assert.Equal(t, int64(3), connectionCount.Raw())
		}
	}
}
