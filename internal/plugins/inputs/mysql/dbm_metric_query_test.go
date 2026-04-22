// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"testing"
	"time"

	gcPoint "github.com/GuanceCloud/cliutils/point"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/stretchr/testify/assert"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

type mockDBMFeeder struct {
	points map[gcPoint.Category][]*gcPoint.Point
	errors []string
}

func (m *mockDBMFeeder) Feed(category gcPoint.Category, pts []*gcPoint.Point, _ ...dkio.FeedOption) error {
	if m.points == nil {
		m.points = make(map[gcPoint.Category][]*gcPoint.Point)
	}
	m.points[category] = append(m.points[category], pts...)
	return nil
}

func (m *mockDBMFeeder) FeedLastError(err string, _ ...metrics.LastErrorOption) {
	m.errors = append(m.errors, err)
}

func TestMergeDuplicateRows(t *testing.T) {
	rows := []dbmRow{
		{
			schemaName:     "app",
			digest:         "digest-1",
			querySignature: "sig-1",
			countStar:      10,
			sumTimerWait:   200,
			sumRowsSent:    4,
		},
		{
			schemaName:     "app",
			digest:         "digest-1",
			querySignature: "sig-1",
			countStar:      5,
			sumTimerWait:   100,
			sumRowsSent:    2,
		},
		{
			schemaName:     "app",
			digest:         "digest-2",
			querySignature: "sig-2",
			countStar:      7,
			sumTimerWait:   70,
		},
	}

	merged := mergeDuplicateRows(rows)
	if assert.Len(t, merged, 2) {
		mergedBySignature := map[string]dbmRow{}
		for _, row := range merged {
			mergedBySignature[row.querySignature] = row
		}

		assert.Equal(t, uint64(15), mergedBySignature["sig-1"].countStar)
		assert.Equal(t, uint64(300), mergedBySignature["sig-1"].sumTimerWait)
		assert.Equal(t, uint64(6), mergedBySignature["sig-1"].sumRowsSent)
		assert.Equal(t, uint64(7), mergedBySignature["sig-2"].countStar)
	}
}

func TestGetMetricRowsSkipsCounterReset(t *testing.T) {
	var snapshot map[string]dbmMetricCache

	_, snapshot = getMetricRows([]dbmRow{
		{
			schemaName:         "app",
			querySignature:     "sig-1",
			countStar:          10,
			sumTimerWait:       20000,
			sumLockTime:        1000,
			sumErrors:          2,
			sumRowsAffected:    3,
			sumRowsSent:        4,
			sumRowsExamined:    5,
			sumSelectScan:      1,
			sumSelectFullJoin:  1,
			sumNoIndexUsed:     1,
			sumNoGoodIndexUsed: 1,
		},
	}, snapshot)

	got, snapshot := getMetricRows([]dbmRow{
		{
			schemaName:         "app",
			querySignature:     "sig-1",
			countStar:          9,
			sumTimerWait:       19000,
			sumLockTime:        900,
			sumErrors:          1,
			sumRowsAffected:    2,
			sumRowsSent:        3,
			sumRowsExamined:    4,
			sumSelectScan:      0,
			sumSelectFullJoin:  0,
			sumNoIndexUsed:     0,
			sumNoGoodIndexUsed: 0,
		},
	}, snapshot)

	assert.Empty(t, got)
	assert.Contains(t, snapshot, getRowKey("app", "sig-1"))
}

func TestBuildMysqlDbmMetric(t *testing.T) {
	rows := []dbmRow{
		{
			schemaName:           "app",
			digest:               "digest-1",
			querySignature:       "sig-1",
			countStar:            12,
			sumTimerWait:         36000,
			sumLockTime:          9000,
			sumErrors:            1,
			sumRowsAffected:      8,
			sumRowsSent:          9,
			sumRowsExamined:      10,
			sumSelectScan:        2,
			sumSelectFullJoin:    3,
			sumNoIndexUsed:       1,
			sumNoGoodIndexUsed:   2,
			deltaCountStar:       3,
			deltaTimerWait:       6000,
			deltaLockTime:        3000,
			deltaErrors:          1,
			deltaRowsAffected:    2,
			deltaRowsSent:        4,
			deltaRowsExamined:    5,
			deltaSelectScan:      1,
			deltaSelectFullJoin:  2,
			deltaNoIndexUsed:     1,
			deltaNoGoodIndexUsed: 1,
		},
	}

	ipt := &Input{
		mergedTags: map[string]string{
			"server":            "127.0.0.1:3306",
			"database_instance": "mysql-test",
		},
	}

	pts, err := ipt.buildMysqlDbmMetric(rows, time.Unix(1700000000, 0))
	assert.NoError(t, err)
	if assert.Len(t, pts, 1) {
		pt := pts[0]
		tags := pt.Tags()
		fields := pt.Fields()

		assert.Equal(t, metricNameMySQLDbmMetric, pt.Name())
		assert.Equal(t, "127.0.0.1:3306", tags.GetTag("server"))
		assert.Equal(t, "mysql-test", tags.GetTag("database_instance"))
		assert.Equal(t, "app", tags.GetTag("schema_name"))
		assert.Equal(t, "digest-1", tags.GetTag("digest"))
		assert.Equal(t, "sig-1", tags.GetTag("query_signature"))

		if countStar := fields.Get("count_star"); countStar != nil {
			assert.Equal(t, uint64(12), countStar.Raw())
		}
		if deltaCount := fields.Get("delta_count_star"); deltaCount != nil {
			assert.Equal(t, uint64(3), deltaCount.Raw())
		}
		if avgTimerWait := fields.Get("avg_timer_wait"); avgTimerWait != nil {
			assert.Equal(t, uint64(2000), avgTimerWait.Raw())
		}
	}
}

func TestBuildMysqlDbmSample(t *testing.T) {
	ipt := &Input{
		Object:       mysqlObject{name: "mysql-server"},
		InstanceName: "mysql-test",
		mergedTags: map[string]string{
			"database_instance": "mysql-test",
		},
		planSignatureCache: expirable.NewLRU[string, struct{}](16, nil, time.Hour),
	}

	pts, err := ipt.buildMysqlDbmSample([]planObj{
		{
			timestamp:       1700000000123,
			duration:        987654,
			currentSchema:   "app",
			planDefinition:  `{"query_block":{"select_id":1}}`,
			planSignature:   "plan-1",
			querySignature:  "query-1",
			statement:       "SELECT * FROM app.orders WHERE id = ?",
			digest:          "digest-1",
			lockTimeNs:      123,
			noGoodIndexUsed: 1,
			noIndexUsed:     0,
			rowsAffected:    2,
			rowsExamined:    3,
			rowsSent:        4,
		},
	}, time.Unix(1700000000, 0))
	assert.NoError(t, err)
	if assert.Len(t, pts, 1) {
		pt := pts[0]
		tags := pt.Tags()
		fields := pt.Fields()
		planKey := generatePlanCacheKey("query-1", "plan-1")

		assert.Equal(t, dbmExecPlanObjectName, pt.Name())
		assert.Equal(t, "mysql-server-mysql-test-"+planKey, tags.GetTag("name"))
		assert.Equal(t, "mysql-server", tags.GetTag("server"))
		assert.Equal(t, "mysql-test", tags.GetTag("database_instance"))
		assert.Equal(t, "MySQL", tags.GetTag("database_type"))
		assert.Equal(t, "JSON", tags.GetTag("plan_type"))
		assert.Equal(t, "app", tags.GetTag("schema_name"))
		assert.Equal(t, "plan-1", tags.GetTag("plan_signature"))
		assert.Equal(t, "query-1", tags.GetTag("query_signature"))
		assert.Equal(t, "digest-1", tags.GetTag("digest"))

		if message := fields.Get("message"); message != nil {
			assert.Equal(t, `{"query_block":{"select_id":1}}`, message.Raw())
		}
		if statement := fields.Get("statement"); statement != nil {
			assert.Equal(t, "SELECT * FROM app.orders WHERE id = ?", statement.Raw())
		}
	}

	_, ok := ipt.planSignatureCache.Get(generatePlanCacheKey("query-1", "plan-1"))
	assert.True(t, ok)
}

func TestCollectMysqlDbmQueriesDeduplicates(t *testing.T) {
	feeder := &mockDBMFeeder{}
	ipt := &Input{
		feeder:       feeder,
		Object:       mysqlObject{name: "mysql-server"},
		InstanceName: "mysql-test",
		mergedTags: map[string]string{
			"database_instance": "mysql-test",
		},
	}

	rows := []dbmRow{
		{
			schemaName:     "app",
			digest:         "digest-1",
			digestText:     "SELECT ?",
			querySignature: "query-1",
		},
		{
			schemaName:     "app",
			digest:         "digest-2",
			digestText:     "SELECT 2",
			querySignature: "",
		},
	}

	ipt.collectDbmQueries(rows, time.Unix(1700000000, 0))
	ipt.collectDbmQueries(rows, time.Unix(1700000060, 0))

	objectPts := feeder.points[gcPoint.Object]
	if assert.Len(t, objectPts, 1) {
		pt := objectPts[0]
		tags := pt.Tags()
		fields := pt.Fields()

		assert.Equal(t, dbmQueryObjectName, pt.Name())
		assert.Equal(t, "mysql-server-mysql-test-query-1", tags.GetTag("name"))
		assert.Equal(t, "mysql-server", tags.GetTag("server"))
		assert.Equal(t, "mysql-test", tags.GetTag("database_instance"))
		assert.Equal(t, "MySQL", tags.GetTag("database_type"))
		assert.Equal(t, "app", tags.GetTag("schema_name"))
		assert.Equal(t, "query-1", tags.GetTag("query_signature"))
		assert.Equal(t, "digest-1", tags.GetTag("digest"))

		if message := fields.Get("message"); message != nil {
			assert.Equal(t, "SELECT ?", message.Raw())
		}
	}
}
