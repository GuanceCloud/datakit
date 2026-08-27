// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mongodb

import (
	"strings"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/util"
)

func mustNewQueryMetricsState(t *testing.T) *queryMetricsState {
	t.Helper()
	state, err := newQueryMetricsState()
	require.NoError(t, err)
	return state
}

func testPtr[T any](value T) *T {
	return &value
}

func TestDefaultQueryMetricsConfig(t *testing.T) {
	ipt := defaultInput()

	require.NotNil(t, ipt.Dbm)
	assert.False(t, ipt.Dbm.Enabled)
	assert.Empty(t, ipt.Dbm.Databases)
	require.NotNil(t, ipt.Dbm.QueryMetrics)
	assert.True(t, ipt.Dbm.QueryMetrics.Enabled)
	assert.Equal(t, 60*time.Second, ipt.Dbm.QueryMetrics.Interval.Duration)
	assert.EqualValues(t, 10000, ipt.Dbm.QueryMetrics.Limit)
	assert.Equal(t, 512, ipt.Dbm.QueryMetrics.QueryTextMaxBytes)
	require.NotNil(t, ipt.Dbm.Activity)
	assert.True(t, ipt.Dbm.Activity.Enabled)
	assert.Equal(t, 10*time.Second, ipt.Dbm.Activity.Interval.Duration)
	require.NotNil(t, ipt.Dbm.SlowOperations)
	assert.False(t, ipt.Dbm.SlowOperations.Enabled)
	assert.Equal(t, 10*time.Second, ipt.Dbm.SlowOperations.Interval.Duration)
	assert.EqualValues(t, 1000, ipt.Dbm.SlowOperations.MaxOperations)
}

func TestDecodeQueryMetricsConfig(t *testing.T) {
	ipt := defaultInput()
	_, err := toml.Decode(`
[dbm]
  enabled = true
  databases = ["loadtest", "orders"]

[dbm.query_metrics]
  enabled = true
  interval = "30s"
  limit = 5000
  query_text_max_bytes = 768

[dbm.activity]
  enabled = true
  interval = "15s"

[dbm.slow_operations]
  enabled = true
  interval = "20s"
  max_operations = 500
`, ipt)
	require.NoError(t, err)

	require.NotNil(t, ipt.Dbm)
	assert.True(t, ipt.Dbm.Enabled)
	assert.Equal(t, []string{"loadtest", "orders"}, ipt.Dbm.Databases)
	require.NotNil(t, ipt.Dbm.QueryMetrics)
	assert.True(t, ipt.Dbm.QueryMetrics.Enabled)
	assert.Equal(t, 30*time.Second, ipt.Dbm.QueryMetrics.Interval.Duration)
	assert.EqualValues(t, 5000, ipt.Dbm.QueryMetrics.Limit)
	assert.Equal(t, 768, ipt.Dbm.QueryMetrics.QueryTextMaxBytes)
	require.NotNil(t, ipt.Dbm.Activity)
	assert.True(t, ipt.Dbm.Activity.Enabled)
	assert.Equal(t, 15*time.Second, ipt.Dbm.Activity.Interval.Duration)
	require.NotNil(t, ipt.Dbm.SlowOperations)
	assert.True(t, ipt.Dbm.SlowOperations.Enabled)
	assert.Equal(t, 20*time.Second, ipt.Dbm.SlowOperations.Interval.Duration)
	assert.EqualValues(t, 500, ipt.Dbm.SlowOperations.MaxOperations)

	ipt = defaultInput()
	_, err = toml.Decode("[dbm]\nenabled = true", ipt)
	require.NoError(t, err)
	require.NotNil(t, ipt.Dbm.Activity)
	assert.True(t, ipt.Dbm.Activity.Enabled)
	require.NotNil(t, ipt.Dbm.SlowOperations)
	assert.False(t, ipt.Dbm.SlowOperations.Enabled)
}

func TestSampleConfigQueryMetrics(t *testing.T) {
	var cfg struct {
		Inputs struct {
			MongoDB []*Input `toml:"mongodb"`
		} `toml:"inputs"`
	}

	_, err := toml.Decode(sampleConfig, &cfg)
	require.NoError(t, err)
	require.Len(t, cfg.Inputs.MongoDB, 1)

	ipt := cfg.Inputs.MongoDB[0]
	assert.True(t, ipt.Election)
	assert.True(t, ipt.Object.Enable)
	require.NotNil(t, ipt.Dbm)
	assert.False(t, ipt.Dbm.Enabled)
	assert.Empty(t, ipt.Dbm.Databases)
	require.NotNil(t, ipt.Dbm.QueryMetrics)
	assert.True(t, ipt.Dbm.QueryMetrics.Enabled)
	assert.Equal(t, 60*time.Second, ipt.Dbm.QueryMetrics.Interval.Duration)
	assert.EqualValues(t, 10000, ipt.Dbm.QueryMetrics.Limit)
	assert.Equal(t, 512, ipt.Dbm.QueryMetrics.QueryTextMaxBytes)
	require.NotNil(t, ipt.Dbm.Activity)
	assert.True(t, ipt.Dbm.Activity.Enabled)
	assert.Equal(t, 10*time.Second, ipt.Dbm.Activity.Interval.Duration)
	require.NotNil(t, ipt.Dbm.SlowOperations)
	assert.False(t, ipt.Dbm.SlowOperations.Enabled)
	assert.Equal(t, 10*time.Second, ipt.Dbm.SlowOperations.Interval.Duration)
	assert.EqualValues(t, 1000, ipt.Dbm.SlowOperations.MaxOperations)
}

func TestQueryTextMaxBytes(t *testing.T) {
	tests := []struct {
		name       string
		configured int
		want       int
	}{
		{name: "default", configured: 512, want: 512},
		{name: "configured", configured: 768, want: 768},
		{name: "zero", configured: 0, want: 512},
		{name: "over maximum", configured: 1025, want: 512},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ipt := defaultInput()
			ipt.Dbm.QueryMetrics.QueryTextMaxBytes = test.configured
			assert.Equal(t, test.want, ipt.queryTextMaxBytes())
		})
	}
}

func TestQueryMetricsInterval(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		want     time.Duration
	}{
		{name: "default", want: 60 * time.Second},
		{name: "minimum", interval: time.Second, want: 5 * time.Second},
		{name: "configured", interval: 30 * time.Second, want: 30 * time.Second},
		{name: "maximum", interval: 30 * time.Minute, want: 10 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := defaultInput()
			ipt.Dbm.QueryMetrics.Interval.Duration = tt.interval
			assert.Equal(t, tt.want, ipt.queryMetricsInterval())
		})
	}
}

func TestQueryStatsPipeline(t *testing.T) {
	tests := []struct {
		name      string
		databases []string
		limit     int64
		want      mongo.Pipeline
	}{
		{
			name:  "exclude admin by default",
			limit: 10000,
			want: mongo.Pipeline{
				bson.D{{Key: "$queryStats", Value: bson.D{}}},
				bson.D{{Key: "$match", Value: bson.D{
					{Key: "key.queryShape.cmdNs.db", Value: bson.D{{Key: "$ne", Value: "admin"}}},
				}}},
				bson.D{{Key: "$limit", Value: int64(10000)}},
			},
		},
		{
			name:      "use configured databases including admin",
			databases: []string{"loadtest", "admin"},
			limit:     5000,
			want: mongo.Pipeline{
				bson.D{{Key: "$queryStats", Value: bson.D{}}},
				bson.D{{Key: "$match", Value: bson.D{
					{Key: "key.queryShape.cmdNs.db", Value: bson.D{{Key: "$in", Value: []string{"loadtest", "admin"}}}},
				}}},
				bson.D{{Key: "$limit", Value: int64(5000)}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, queryStatsPipeline(tt.databases, tt.limit))
		})
	}
}

func TestActivityInterval(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		want     time.Duration
	}{
		{name: "default", want: 10 * time.Second},
		{name: "minimum", interval: time.Second, want: 5 * time.Second},
		{name: "configured", interval: 30 * time.Second, want: 30 * time.Second},
		{name: "maximum", interval: 30 * time.Minute, want: 10 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := defaultInput()
			ipt.Dbm.Activity.Interval.Duration = tt.interval
			assert.Equal(t, tt.want, ipt.activityInterval())
		})
	}
}

func TestSlowOperationsConfig(t *testing.T) {
	tests := []struct {
		name      string
		interval  time.Duration
		limit     int64
		want      time.Duration
		wantLimit int64
	}{
		{name: "default", want: 10 * time.Second, wantLimit: 1000},
		{name: "minimum", interval: time.Second, limit: -1, want: 5 * time.Second, wantLimit: 1000},
		{name: "configured", interval: 30 * time.Second, limit: 500, want: 30 * time.Second, wantLimit: 500},
		{name: "maximum", interval: 30 * time.Minute, limit: 2000, want: 10 * time.Minute, wantLimit: 2000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := defaultInput()
			ipt.Dbm.SlowOperations.Interval.Duration = tt.interval
			ipt.Dbm.SlowOperations.MaxOperations = tt.limit
			assert.Equal(t, tt.want, ipt.slowOperationsInterval())
			assert.Equal(t, tt.wantLimit, ipt.slowOperationsLimit())
		})
	}
}

func TestQueryMetricsVersionSupported(t *testing.T) {
	tests := []struct {
		version   string
		supported bool
		wantErr   bool
	}{
		{version: "", wantErr: true},
		{version: "7.0.0"},
		{version: "8.0.0-rc0"},
		{version: "8.0", supported: true},
		{version: "8.0.1", supported: true},
		{version: "9.0.0", supported: true},
		{version: "unknown", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			supported, err := queryMetricsVersionSupported(test.version)
			if test.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.supported, supported)
		})
	}
}

func TestDBMNodeStatusCollectible(t *testing.T) {
	tests := []struct {
		name        string
		status      dbmNodeStatus
		collectible bool
	}{
		{name: "standalone", collectible: true},
		{name: "mongos", status: dbmNodeStatus{IsWritablePrimary: true}, collectible: true},
		{name: "replica set primary", status: dbmNodeStatus{SetName: "rs0", IsWritablePrimary: true}, collectible: true},
		{name: "replica set secondary", status: dbmNodeStatus{SetName: "rs0", Secondary: true}, collectible: true},
		{name: "arbiter", status: dbmNodeStatus{SetName: "rs0", ArbiterOnly: true}},
		{name: "recovering", status: dbmNodeStatus{SetName: "rs0"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.collectible, test.status.collectible())
		})
	}
}

func TestNormalizeMongoDBActivityRow(t *testing.T) {
	commandObfuscator := newMongoDBCommandObfuscator()
	defer commandObfuscator.Stop()

	raw := mongoDBActivityRawRow{
		Type:                "op",
		Description:         "conn42",
		Active:              testPtr(true),
		CurrentOpTime:       primitive.NewDateTimeFromTime(time.Date(2026, time.August, 21, 10, 0, 0, 0, time.UTC)),
		EffectiveUsers:      []mongoDBActivityUser{{User: "datakit"}},
		OperationID:         int64(42),
		MicrosecondsRunning: testPtr[int64](1200),
		Operation:           "query",
		Namespace:           "loadtest.users",
		QueryShapeHash:      "shape-hash",
		Command: bson.D{
			{Key: "find", Value: "users"},
			{Key: "filter", Value: bson.D{{Key: "email", Value: bson.D{{Key: "$eq", Value: "secret@example.com"}}}}},
			{Key: "projection", Value: bson.D{{Key: "email", Value: true}}},
			{Key: "sort", Value: bson.D{{Key: "created_at", Value: -1}}},
			{Key: "comment", Value: "trace-id"},
			{Key: "lsid", Value: bson.D{{Key: "id", Value: primitive.Binary{Subtype: 4, Data: []byte{1, 2, 3}}}}},
			{Key: "$clusterTime", Value: bson.D{{Key: "clusterTime", Value: primitive.Timestamp{T: 1, I: 2}}}},
			{Key: "$db", Value: "loadtest"},
		},
		PlanSummary:          "IXSCAN",
		QueryFramework:       "sbe",
		PrepareReadConflicts: testPtr[int64](1),
		WriteConflicts:       testPtr[int64](2),
		NumYields:            testPtr[int64](3),
		WaitingForLock:       testPtr(true),
		Locks:                bson.D{{Key: "FeatureCompatibilityVersion", Value: "r"}},
		LockStats: bson.D{{Key: "Global", Value: bson.D{
			{Key: "acquireCount", Value: bson.D{{Key: "r", Value: int64(1)}}},
		}}},
		WaitingForFlowControl: testPtr(true),
		FlowControlStats:      bson.D{{Key: "acquireCount", Value: int64(1)}},
		WaitingForLatch:       bson.D{{Key: "captureName", Value: "FutureResolution"}},
		Cursor: bson.D{
			{Key: "cursorId", Value: int64(99)},
			{Key: "originatingCommand", Value: bson.D{
				{Key: "find", Value: "users"},
				{Key: "filter", Value: bson.D{{Key: "token", Value: "cursor-secret"}}},
				{Key: "comment", Value: "cursor-comment"},
				{Key: "$db", Value: "loadtest"},
			}},
		},
		Transaction: bson.D{
			{Key: "parameters", Value: bson.D{{Key: "txnNumber", Value: int64(7)}}},
			{Key: "timeOpenMicros", Value: int64(200)},
		},
		LogicalSessionID: bson.D{{Key: "id", Value: primitive.Binary{Subtype: 4, Data: []byte{1, 2, 3}}}},
		Client:           "127.0.0.1:51000",
		ClientMetadata: bson.D{
			{Key: "driver", Value: bson.D{{Key: "name", Value: "mongo-go-driver"}, {Key: "version", Value: "1.17.0"}}},
			{Key: "os", Value: bson.D{{Key: "type", Value: "darwin"}}},
			{Key: "platform", Value: "go1.25"},
		},
		Application: "load-generator",
	}

	row, include, err := normalizeMongoDBActivityRow(
		raw,
		map[string]struct{}{"loadtest": {}},
		commandObfuscator,
	)
	require.NoError(t, err)
	require.True(t, include)
	assert.Equal(t, "loadtest", row.databaseName)
	assert.Equal(t, "users", row.collection)
	assert.Equal(t, "find", row.commandType)
	assert.Equal(t, "not_truncated", row.queryTruncated)
	assert.Equal(t, "datakit", row.user)
	assert.Equal(t, "42", row.operationID)
	assert.Equal(t, "shape-hash", row.queryShapeHash)
	assert.Equal(t, generateMongoDBQuerySignature("loadtest", "users", "shape-hash"), row.querySignature)
	assert.Contains(t, row.message, `"find":"users"`)
	assert.Contains(t, row.message, `"email":{"$eq":"?"}`)
	assert.NotContains(t, row.message, "secret@example.com")
	assert.NotContains(t, row.message, "trace-id")
	assert.NotContains(t, row.message, "lsid")
	assert.NotContains(t, row.message, "$clusterTime")
	assert.Contains(t, row.client, `"hostname":"127.0.0.1:51000"`)
	assert.Equal(t, "127.0.0.1", row.clientAddress)
	assert.Contains(t, row.locks, "feature_compatibility_version")
	assert.Contains(t, row.lockStats, "acquire_count")
	assert.Contains(t, row.cursor, `"cursor_id":99`)
	assert.NotContains(t, row.cursor, "cursor-comment")
	assert.NotContains(t, row.cursor, "cursor-secret")
	assert.Contains(t, row.transaction, `"txn_number":7`)
	assert.Equal(t, `{"id":"010203"}`, row.logicalSessionID)

	expectedNormalizedQueryHash, err := computeMongoDBNormalizedQueryHash(row.message)
	require.NoError(t, err)
	assert.Equal(t, expectedNormalizedQueryHash, row.normalizedQueryHash)
}

func TestMongoDBCommandObfuscatorObfuscatesComment(t *testing.T) {
	commandObfuscator := newMongoDBCommandObfuscator()
	defer commandObfuscator.Stop()

	command, err := obfuscateMongoDBCommand(commandObfuscator, bson.D{
		{Key: "find", Value: "users"},
		{Key: "comment", Value: "sensitive-comment"},
	})
	require.NoError(t, err)
	assert.Contains(t, command, `"find":"users"`)
	assert.NotContains(t, command, "sensitive-comment")
}

func TestMongoDBActivityFilters(t *testing.T) {
	commandObfuscator := newMongoDBCommandObfuscator()
	defer commandObfuscator.Stop()

	base := mongoDBActivityRawRow{
		Namespace: "loadtest.users",
		Operation: "query",
		Command: bson.D{
			{Key: "find", Value: "users"},
			{Key: "$db", Value: "loadtest"},
		},
	}
	databaseFilter := map[string]struct{}{"loadtest": {}}

	tests := []struct {
		name string
		row  mongoDBActivityRawRow
		want bool
	}{
		{name: "included", row: base, want: true},
		{name: "missing namespace", row: func() mongoDBActivityRawRow { row := base; row.Namespace = ""; return row }()},
		{name: "admin database", row: func() mongoDBActivityRawRow { row := base; row.Namespace = "admin.$cmd"; return row }()},
		{name: "filtered database", row: func() mongoDBActivityRawRow { row := base; row.Namespace = "other.users"; return row }()},
		{name: "missing command", row: func() mongoDBActivityRawRow { row := base; row.Command = nil; return row }()},
		{name: "hello command", row: mongoDBActivityRawRow{
			Namespace: "loadtest.$cmd",
			Operation: "command",
			Command:   bson.D{{Key: "hello", Value: 1}, {Key: "$db", Value: "loadtest"}},
		}},
		{name: "getMore", row: mongoDBActivityRawRow{
			Namespace: "loadtest.users",
			Operation: "getmore",
			Command: bson.D{
				{Key: "getMore", Value: int64(99)},
				{Key: "collection", Value: "users"},
				{Key: "$db", Value: "loadtest"},
			},
		}, want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			row, include, err := normalizeMongoDBActivityRow(test.row, databaseFilter, commandObfuscator)
			require.NoError(t, err)
			assert.Equal(t, test.want, include)
			if test.name == "getMore" {
				assert.Equal(t, "getMore", row.commandType)
				assert.Equal(t, "users", row.collection)
			}
		})
	}

	adminRow := base
	adminRow.Namespace = "admin.system.users"
	adminRow.Command = bson.D{{Key: "find", Value: "system.users"}, {Key: "$db", Value: "admin"}}
	_, include, err := normalizeMongoDBActivityRow(adminRow, nil, commandObfuscator)
	require.NoError(t, err)
	assert.False(t, include)
	_, include, err = normalizeMongoDBActivityRow(adminRow, map[string]struct{}{"admin": {}}, commandObfuscator)
	require.NoError(t, err)
	assert.True(t, include)
}

func TestMongoDBActivityTruncationState(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		present bool
		want    string
	}{
		{name: "missing", want: "not_truncated"},
		{name: "boolean false", value: false, present: true, want: "not_truncated"},
		{name: "boolean true", value: true, present: true, want: "truncated"},
		{name: "empty string", value: "", present: true, want: "not_truncated"},
		{name: "string summary", value: `{ find: "users", filter: { ... } }`, present: true, want: "truncated"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := bson.D{}
			if test.present {
				command = append(command, bson.E{Key: "$truncated", Value: test.value})
			}
			assert.Equal(t, test.want, mongoDBActivityTruncationState(command))
		})
	}
}

func TestMongoDBActivitySnakeCase(t *testing.T) {
	assert.Equal(t, "feature_compatibility_version", mongoDBActivitySnakeCase("FeatureCompatibilityVersion"))
	assert.Equal(t, "rstl", mongoDBActivitySnakeCase("RSTL"))
	assert.Equal(t, "acquire_count", mongoDBActivitySnakeCase("acquireCount"))
}

func TestBuildMongoDBActivityPoint(t *testing.T) {
	ipt := defaultInput()
	svr := &MongodbServer{
		host:             "127.0.0.1:27017",
		databaseInstance: "mongo-test",
		ipt:              ipt,
	}
	row := mongoDBActivityRow{
		databaseName:        "loadtest",
		collection:          "users",
		commandType:         "find",
		querySignature:      "query-signature",
		queryShapeHash:      "shape-hash",
		normalizedQueryHash: "normalized-hash",
		message:             `{"find":"users","$db":"loadtest"}`,
		activityType:        "op",
		operation:           "query",
		application:         "load-generator",
		user:                "datakit",
		namespace:           "loadtest.users",
		active:              testPtr(true),
		microsecondsRunning: testPtr[int64](100),
		writeConflicts:      testPtr[int64](2),
		waitingForLock:      testPtr(true),
		queryTruncated:      "not_truncated",
		locks:               `{"global":"r"}`,
	}

	activityPoint := buildMongoDBActivityPoint(svr, row, time.Unix(100, 0))
	assert.Equal(t, mongodbDBMActivityName, activityPoint.Name())
	assert.Equal(t, "loadtest", activityPoint.GetTag("database_name"))
	assert.Equal(t, "users", activityPoint.GetTag("collection"))
	assert.Equal(t, "mongo-test", activityPoint.GetTag("database_instance"))
	assert.Equal(t, "query-signature", activityPoint.GetTag("query_signature"))
	assert.Equal(t, "shape-hash", activityPoint.GetTag("query_shape_hash"))
	assert.Equal(t, "normalized-hash", activityPoint.GetTag("normalized_query_hash"))
	message, ok := activityPoint.GetS("message")
	require.True(t, ok)
	assert.Equal(t, row.message, message)
	active, ok := activityPoint.GetB("active")
	require.True(t, ok)
	assert.True(t, active)
	duration, ok := activityPoint.GetI("microsecs_running")
	require.True(t, ok)
	assert.Equal(t, int64(100), duration)
	writeConflicts, ok := activityPoint.GetI("write_conflicts")
	require.True(t, ok)
	assert.Equal(t, int64(2), writeConflicts)
	_, ok = activityPoint.GetI("prepare_read_conflicts")
	assert.False(t, ok)
	_, ok = activityPoint.GetS("plan_summary")
	assert.False(t, ok)

	info := (&mongodbDBMActivityMeasurement{}).Info()
	assert.Equal(t, mongodbDBMActivityName, info.Name)
	assert.Contains(t, info.Tags, "normalized_query_hash")
	assert.Contains(t, info.Fields, "cursor")
}

func TestMongoDBActivityOptionalFields(t *testing.T) {
	encoded, err := bson.Marshal(bson.D{
		{Key: "active", Value: false},
		{Key: "microsecs_running", Value: int64(0)},
		{Key: "waitingForLock", Value: false},
	})
	require.NoError(t, err)

	var raw mongoDBActivityRawRow
	require.NoError(t, bson.Unmarshal(encoded, &raw))
	require.NotNil(t, raw.Active)
	require.NotNil(t, raw.MicrosecondsRunning)
	require.NotNil(t, raw.WaitingForLock)
	assert.Nil(t, raw.WriteConflicts)

	pt := buildMongoDBActivityPoint(&MongodbServer{}, mongoDBActivityRow{
		message:             `{}`,
		namespace:           "loadtest.users",
		active:              raw.Active,
		microsecondsRunning: raw.MicrosecondsRunning,
		waitingForLock:      raw.WaitingForLock,
	}, time.Unix(100, 0))

	active, ok := pt.GetB("active")
	require.True(t, ok)
	assert.False(t, active)
	duration, ok := pt.GetI("microsecs_running")
	require.True(t, ok)
	assert.Zero(t, duration)
	waitingForLock, ok := pt.GetB("waiting_for_lock")
	require.True(t, ok)
	assert.False(t, waitingForLock)
	_, ok = pt.GetI("write_conflicts")
	assert.False(t, ok)
	_, ok = pt.GetS("description")
	assert.False(t, ok)
}

func TestMongoDBOperationMetrics(t *testing.T) {
	baseRow := mongoDBActivityRow{
		namespace:      "loadtest.users",
		databaseName:   "loadtest",
		collection:     "users",
		commandType:    "find",
		operation:      "query",
		user:           "datakit",
		clientAddress:  "127.0.0.1",
		application:    "load-generator",
		active:         testPtr(true),
		waitingForLock: testPtr(true),
	}
	secondRow := baseRow
	secondRow.waitingForLock = testPtr(false)
	getMoreRow := baseRow
	getMoreRow.commandType = "getMore"
	getMoreRow.operation = "getmore"
	getMoreRow.waitingForLock = testPtr(false)
	inactiveRow := baseRow
	inactiveRow.active = testPtr(false)

	rows := aggregateMongoDBOperations([]mongoDBActivityRow{baseRow, secondRow, getMoreRow, inactiveRow})
	require.Len(t, rows, 2)

	ipt := defaultInput()
	svr := &MongodbServer{
		host:             "127.0.0.1:27017",
		databaseInstance: "mongo-test",
		ipt:              ipt,
	}
	points := buildMongoDBOperationPoints(svr, rows, time.Unix(100, 0))
	require.Len(t, points, 2)

	for _, pt := range points {
		assert.Equal(t, mongodbDBMOperationName, pt.Name())
		assert.Equal(t, "loadtest.users", pt.GetTag("namespace"))
		assert.Equal(t, "loadtest", pt.GetTag("database_name"))
		assert.Empty(t, pt.GetTag("collection"))
		assert.Equal(t, "load-generator", pt.GetTag("application"))
		assert.Equal(t, "127.0.0.1", pt.GetTag("client_address"))

		activeCount, ok := pt.GetI("active_operation_count")
		require.True(t, ok)
		waitingCount, ok := pt.GetI("waiting_for_lock_count")
		require.True(t, ok)
		if pt.GetTag("command_type") == "find" {
			assert.Equal(t, int64(2), activeCount)
			assert.Equal(t, int64(1), waitingCount)
		} else {
			assert.Equal(t, "getMore", pt.GetTag("command_type"))
			assert.Equal(t, int64(1), activeCount)
			assert.Equal(t, int64(0), waitingCount)
		}
	}

	info := (&mongodbDBMOperationMeasurement{}).Info()
	assert.Equal(t, mongodbDBMOperationName, info.Name)
	assert.Contains(t, info.Tags, "namespace")
	assert.Contains(t, info.Tags, "client_address")
	assert.NotContains(t, info.Tags, "collection")
	assert.Contains(t, info.Tags, "application")
	assert.Contains(t, info.Fields, "active_operation_count")
	assert.Contains(t, info.Fields, "waiting_for_lock_count")
}

func TestNormalizeQueryStatsValue(t *testing.T) {
	input := bson.D{
		{Key: "filter", Value: bson.D{
			{Key: "$and", Value: bson.A{
				bson.D{{Key: "status", Value: bson.D{{Key: "$eq", Value: "?string"}}}},
				bson.D{{Key: "amount", Value: bson.D{{Key: "$gt", Value: "?number"}}}},
			}},
		}},
		{Key: "regular", Value: "?unknown"},
	}

	assert.Equal(t, bson.D{
		{Key: "filter", Value: bson.D{
			{Key: "$and", Value: bson.A{
				bson.D{{Key: "status", Value: bson.D{{Key: "$eq", Value: "?"}}}},
				bson.D{{Key: "amount", Value: bson.D{{Key: "$gt", Value: "?"}}}},
			}},
		}},
		{Key: "regular", Value: "?unknown"},
	}, normalizeQueryStatsValue(input))
}

func TestReconstructQueryStatsCommand(t *testing.T) {
	queryShape := bson.D{
		{Key: "cmdNs", Value: bson.D{{Key: "db", Value: "loadtest"}, {Key: "coll", Value: "users"}}},
		{Key: "command", Value: "find"},
		{Key: "filter", Value: bson.D{{Key: "status", Value: bson.D{{Key: "$eq", Value: "?string"}}}}},
		{Key: "limit", Value: "?number"},
	}

	command := reconstructQueryStatsCommand(queryShape)
	assert.Equal(t, bson.D{
		{Key: "find", Value: "users"},
		{Key: "$db", Value: "loadtest"},
		{Key: "filter", Value: bson.D{{Key: "status", Value: bson.D{{Key: "$eq", Value: "?"}}}}},
		{Key: "limit", Value: "?"},
	}, command)
}

func TestNormalizeQueryMetricsRow(t *testing.T) {
	state := mustNewQueryMetricsState(t)
	raw := queryStatsRawRow{
		Key: queryStatsKey{
			QueryShape: bson.D{
				{Key: "cmdNs", Value: bson.D{{Key: "db", Value: "loadtest"}, {Key: "coll", Value: "users"}}},
				{Key: "command", Value: "find"},
				{Key: "filter", Value: bson.D{{Key: "email", Value: bson.D{{Key: "$eq", Value: "?string"}}}}},
			},
		},
		QueryShapeHash: "shape-hash",
		Metrics: bson.M{
			"execCount":               int64(7),
			"totalExecMicros":         bson.M{"sum": int64(350)},
			"firstResponseExecMicros": bson.M{"sum": int64(210)},
			"keysExamined":            bson.M{"sum": int64(14)},
			"docsExamined":            bson.M{"sum": int64(21)},
			"docsReturned":            bson.M{"sum": int64(7)},
			"usedDisk":                bson.M{"true": int64(2)},
			"hasSortStage":            bson.M{"true": int64(3)},
		},
	}

	row, err := state.normalizeRow(raw, defaultQueryTextMaxBytes)
	require.NoError(t, err)
	assert.Equal(t, "loadtest", row.databaseName)
	assert.Equal(t, "users", row.collection)
	assert.Equal(t, "find", row.commandType)
	assert.Equal(t, generateMongoDBQuerySignature("loadtest", "users", "shape-hash"), row.querySignature)
	assert.Equal(t, "shape-hash", row.queryShapeHash)
	assert.Equal(t, `{"find":"users","$db":"loadtest","filter":{"email":{"$eq":"?"}}}`, row.obfuscatedCommand)
	assert.Equal(t,
		util.ComputeNormalizedSQLHash(`{"$db":"loadtest","filter":{"email":{"$eq":"?"}},"find":"users"}`),
		row.normalizedQueryHash,
	)
	assert.NotContains(t, row.obfuscatedCommand, "?string")
	assert.Equal(t, row.obfuscatedCommand, row.queryText)
	assert.False(t, row.queryTextTruncated)
	assert.Equal(t, int64(7), row.metrics["exec_count"])
	assert.Equal(t, int64(350), row.metrics["total_exec_micros_sum"])
	assert.Equal(t, int64(210), row.metrics["first_response_exec_micros_sum"])
	assert.Equal(t, int64(14), row.metrics["keys_examined_sum"])
	assert.Equal(t, int64(21), row.metrics["docs_examined_sum"])
	assert.Equal(t, int64(7), row.metrics["docs_returned_sum"])
	assert.Equal(t, int64(2), row.metrics["used_disk_count"])
	assert.Equal(t, int64(3), row.metrics["has_sort_stage_count"])
}

func TestQueryMetricsDeltas(t *testing.T) {
	state := mustNewQueryMetricsState(t)
	baseline := queryMetricRow{
		databaseName:   "loadtest",
		collection:     "users",
		querySignature: "signature",
		metrics: map[string]int64{
			"exec_count":            5,
			"total_exec_micros_sum": 100,
			"docs_examined_sum":     20,
		},
	}

	assert.Empty(t, state.computeDeltas([]queryMetricRow{baseline}))

	current := baseline
	current.metrics = map[string]int64{
		"exec_count":            8,
		"total_exec_micros_sum": 190,
		"docs_examined_sum":     32,
	}
	deltas := state.computeDeltas([]queryMetricRow{current})
	require.Len(t, deltas, 1)
	assert.Equal(t, int64(8), deltas[0].metrics["exec_count"])
	assert.Equal(t, int64(190), deltas[0].metrics["total_exec_micros_sum"])
	assert.Equal(t, int64(32), deltas[0].metrics["docs_examined_sum"])
	assert.Equal(t, int64(3), deltas[0].deltaMetrics["exec_count"])
	assert.Equal(t, int64(90), deltas[0].deltaMetrics["total_exec_micros_sum"])
	assert.Equal(t, int64(12), deltas[0].deltaMetrics["docs_examined_sum"])

	unchangedExecutionCount := current
	unchangedExecutionCount.metrics = map[string]int64{
		"exec_count":            8,
		"total_exec_micros_sum": 200,
		"docs_examined_sum":     32,
	}
	assert.Empty(t, state.computeDeltas([]queryMetricRow{unchangedExecutionCount}))

	reset := baseline
	reset.metrics = map[string]int64{
		"exec_count":            1,
		"total_exec_micros_sum": 10,
		"docs_examined_sum":     2,
	}
	assert.Empty(t, state.computeDeltas([]queryMetricRow{reset}))

	afterReset := reset
	afterReset.metrics = map[string]int64{
		"exec_count":            2,
		"total_exec_micros_sum": 30,
		"docs_examined_sum":     6,
	}
	deltas = state.computeDeltas([]queryMetricRow{afterReset})
	require.Len(t, deltas, 1)
	assert.Equal(t, int64(2), deltas[0].metrics["exec_count"])
	assert.Equal(t, int64(30), deltas[0].metrics["total_exec_micros_sum"])
	assert.Equal(t, int64(1), deltas[0].deltaMetrics["exec_count"])
	assert.Equal(t, int64(20), deltas[0].deltaMetrics["total_exec_micros_sum"])
}

func TestQueryMetricsMergesDuplicateRows(t *testing.T) {
	state := mustNewQueryMetricsState(t)
	row := queryMetricRow{
		databaseName:   "loadtest",
		collection:     "users",
		querySignature: "signature",
		metrics:        map[string]int64{"exec_count": 1, "docs_returned_sum": 2},
	}
	assert.Empty(t, state.computeDeltas([]queryMetricRow{row, row}))

	row.metrics = map[string]int64{"exec_count": 3, "docs_returned_sum": 5}
	deltas := state.computeDeltas([]queryMetricRow{row, row})
	require.Len(t, deltas, 1)
	assert.Equal(t, int64(6), deltas[0].metrics["exec_count"])
	assert.Equal(t, int64(10), deltas[0].metrics["docs_returned_sum"])
	assert.Equal(t, int64(4), deltas[0].deltaMetrics["exec_count"])
	assert.Equal(t, int64(6), deltas[0].deltaMetrics["docs_returned_sum"])
}

func TestBuildMongoDBQueryMetricPoints(t *testing.T) {
	ipt := defaultInput()
	svr := &MongodbServer{
		host:             "127.0.0.1:27017",
		databaseInstance: "mongo-test",
		ipt:              ipt,
	}
	row := queryMetricRow{
		databaseName:        "loadtest",
		collection:          "users",
		commandType:         "find",
		querySignature:      "abc123",
		queryShapeHash:      "shape-hash",
		normalizedQueryHash: "normalized-hash",
		obfuscatedCommand:   `{"$db":"loadtest","filter":{"email":"?"},"find":"users"}`,
		queryText:           `{"$db":"loadtest","filter":{"email":"?"},"find":"users"}`,
		queryTextTruncated:  true,
		metrics:             map[string]int64{"exec_count": 3, "total_exec_micros_sum": 90},
		deltaMetrics:        map[string]int64{"exec_count": 2, "total_exec_micros_sum": 50},
	}
	ptsTime := time.Date(2026, time.August, 20, 10, 0, 0, 0, time.UTC)

	points := buildMongoDBQueryMetricPoints(svr, []queryMetricRow{row}, ptsTime)
	require.Len(t, points, 1)
	assert.Equal(t, mongodbDBMMetricName, points[0].Name())
	assert.Equal(t, "loadtest", points[0].GetTag("database_name"))
	assert.Equal(t, "users", points[0].GetTag("collection"))
	assert.Equal(t, "abc123", points[0].GetTag("query_signature"))
	assert.Equal(t, "normalized-hash", points[0].GetTag("normalized_query_hash"))
	assert.Equal(t, "mongo-test", points[0].GetTag("database_instance"))
	assert.Equal(t, row.queryText, points[0].GetTag("query_text"))
	assert.Equal(t, "true", points[0].GetTag("query_truncated"))
	execCount, ok := points[0].GetI("exec_count")
	require.True(t, ok)
	assert.Equal(t, int64(3), execCount)
	deltaExecCount, ok := points[0].GetI("delta_exec_count")
	require.True(t, ok)
	assert.Equal(t, int64(2), deltaExecCount)
	deltaExecMicros, ok := points[0].GetI("delta_total_exec_micros_sum")
	require.True(t, ok)
	assert.Equal(t, int64(50), deltaExecMicros)

	state := mustNewQueryMetricsState(t)
	objects, querySignatures := state.buildQueryObjectPoints(svr, []queryMetricRow{row}, ptsTime)
	require.Len(t, objects, 1)
	assert.Equal(t, []string{row.querySignature}, querySignatures)
	assert.Equal(t, mongodbDBMQueryObjectName, objects[0].Name())
	assert.Equal(t, "127.0.0.1:27017-mongo-test-abc123", objects[0].GetTag("name"))
	assert.Equal(t, "normalized-hash", objects[0].GetTag("normalized_query_hash"))
	message, ok := objects[0].GetS("message")
	require.True(t, ok)
	assert.Equal(t, row.obfuscatedCommand, message)
	sameSignature := row
	sameSignature.databaseName = "another-database"
	objects, querySignatures = state.buildQueryObjectPoints(svr, []queryMetricRow{sameSignature}, ptsTime)
	require.Len(t, objects, 1)
	assert.Equal(t, []string{row.querySignature}, querySignatures)
	for _, querySignature := range querySignatures {
		state.queryObjectCache.Add(querySignature, time.Now())
	}
	objects, querySignatures = state.buildQueryObjectPoints(svr, []queryMetricRow{sameSignature}, ptsTime)
	assert.Empty(t, objects)
	assert.Empty(t, querySignatures)
	state.queryObjectCache.Add(row.querySignature, time.Now().Add(-queryObjectCacheTTL))
	objects, querySignatures = state.buildQueryObjectPoints(svr, []queryMetricRow{sameSignature}, ptsTime)
	require.Len(t, objects, 1)
	assert.Equal(t, []string{row.querySignature}, querySignatures)
	assert.Equal(t, 24*time.Hour, queryObjectCacheTTL)
}

func TestGenerateMongoDBQuerySignature(t *testing.T) {
	signature := generateMongoDBQuerySignature("loadtest", "users", "shape-hash")
	assert.Len(t, signature, 16)
	assert.NotEqual(t, signature, generateMongoDBQuerySignature("other", "users", "shape-hash"))
	assert.NotEqual(t, signature, generateMongoDBQuerySignature("loadtest", "orders", "shape-hash"))
	assert.NotEqual(t, signature, generateMongoDBQuerySignature("loadtest", "users", "other-shape"))
}

func TestComputeMongoDBNormalizedQueryHash(t *testing.T) {
	first := `{"find":"users","filter":{"a":"?","b":"?"},"pipeline":[{"$match":{"x":"?","y":"?"}},{"$sort":{"a":1,"b":-1}}]}`
	second := `{"pipeline":[{"$match":{"y":"?","x":"?"}},{"$sort":{"b":-1,"a":1}}],"filter":{"b":"?","a":"?"},"find":"users"}`
	reversedPipeline := `{"find":"users","filter":{"a":"?","b":"?"},"pipeline":[{"$sort":{"a":1,"b":-1}},{"$match":{"x":"?","y":"?"}}]}`

	firstHash, err := computeMongoDBNormalizedQueryHash(first)
	require.NoError(t, err)
	secondHash, err := computeMongoDBNormalizedQueryHash(second)
	require.NoError(t, err)
	reversedPipelineHash, err := computeMongoDBNormalizedQueryHash(reversedPipeline)
	require.NoError(t, err)

	assert.Equal(t, firstHash, secondHash)
	assert.NotEqual(t, firstHash, reversedPipelineHash)
	_, err = computeMongoDBNormalizedQueryHash(`{"find":`)
	assert.Error(t, err)
}

func TestNormalizeQueryMetricsRowTruncatesQueryText(t *testing.T) {
	state := mustNewQueryMetricsState(t)
	raw := queryStatsRawRow{
		Key: queryStatsKey{
			QueryShape: bson.D{
				{Key: "cmdNs", Value: bson.D{{Key: "db", Value: "loadtest"}, {Key: "coll", Value: strings.Repeat("x", 600)}}},
				{Key: "command", Value: "find"},
			},
		},
		QueryShapeHash: "shape-hash",
		Metrics:        bson.M{"execCount": int64(1)},
	}

	row, err := state.normalizeRow(raw, 64)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(row.queryText), 64)
	assert.True(t, row.queryTextTruncated)
	expectedHash, err := computeMongoDBNormalizedQueryHash(row.obfuscatedCommand)
	require.NoError(t, err)
	assert.Equal(t, expectedHash, row.normalizedQueryHash)
	assert.NotEqual(t, util.ComputeNormalizedSQLHash(row.queryText), row.normalizedQueryHash)
}

func TestQueryTextPreservesBSONOrder(t *testing.T) {
	normalize := func(t *testing.T, sort bson.D) string {
		t.Helper()
		source := bson.D{
			{Key: "key", Value: bson.D{
				{Key: "queryShape", Value: bson.D{
					{Key: "cmdNs", Value: bson.D{{Key: "db", Value: "loadtest"}, {Key: "coll", Value: "users"}}},
					{Key: "command", Value: "find"},
					{Key: "sort", Value: sort},
				}},
			}},
			{Key: "queryShapeHash", Value: "shape-hash"},
			{Key: "metrics", Value: bson.D{{Key: "execCount", Value: int64(1)}}},
		}

		data, err := bson.Marshal(source)
		require.NoError(t, err)
		var raw queryStatsRawRow
		require.NoError(t, bson.Unmarshal(data, &raw))

		row, err := mustNewQueryMetricsState(t).normalizeRow(raw, defaultQueryTextMaxBytes)
		require.NoError(t, err)
		return row.obfuscatedCommand
	}

	first := bson.D{{Key: "created_at", Value: int32(-1)}, {Key: "score", Value: int32(1)}}
	second := bson.D{{Key: "score", Value: int32(1)}, {Key: "created_at", Value: int32(-1)}}
	for i := 0; i < 20; i++ {
		assert.Equal(t, `{"find":"users","$db":"loadtest","sort":{"created_at":-1,"score":1}}`, normalize(t, first))
	}
	assert.Equal(t, `{"find":"users","$db":"loadtest","sort":{"score":1,"created_at":-1}}`, normalize(t, second))
}

func TestNormalizeMongoDBSlowOperation(t *testing.T) {
	commandObfuscator := newMongoDBCommandObfuscator()
	defer commandObfuscator.Stop()

	timestamp := time.Unix(1700000000, int64(123*time.Millisecond)).UTC()
	rawRow := mongoDBSlowOperationRawRow{
		Timestamp: timestamp,
		Operation: "query",
		Namespace: "loadtest.users",
		Command: bson.D{
			{Key: "find", Value: "users"},
			{Key: "filter", Value: bson.D{{Key: "email", Value: "secret@example.com"}}},
			{Key: "$db", Value: "loadtest"},
			{Key: "comment", Value: "trace-id"},
		},
		QueryShapeHash:     "shape-hash",
		PlanCacheShapeHash: "query-hash",
		PlanCacheKey:       "plan-cache-key",
		PlanSummary:        "IXSCAN { email: 1 }",
		QueryFramework:     "sbe",
		User:               "datakit@admin",
		Application:        "datakittest",
		Client:             "127.0.0.1:50000",
		DurationMillis:     testPtr[int64](120),
		WorkingMillis:      testPtr[int64](100),
		NumYields:          testPtr[int64](2),
		ResponseLength:     testPtr[int64](1024),
		DocsReturned:       testPtr[int64](1),
		KeysExamined:       testPtr[int64](1),
		DocsExamined:       testPtr[int64](1),
		CPUNanos:           testPtr[int64](500),
		UsedDisk:           testPtr(true),
	}

	row, ok, err := normalizeMongoDBSlowOperation(rawRow, "fallback", commandObfuscator)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, timestamp, row.timestamp)
	assert.Equal(t, "loadtest", row.databaseName)
	assert.Equal(t, "users", row.collection)
	assert.Equal(t, "find", row.commandType)
	assert.Equal(t, "query-hash", row.queryHash)
	assert.Equal(t, generateMongoDBQuerySignature("loadtest", "users", "shape-hash"), row.querySignature)
	assert.NotEmpty(t, row.normalizedQueryHash)
	assert.Contains(t, row.message, `"find":"users"`)
	assert.Contains(t, row.message, `"email":"?"`)
	assert.NotContains(t, row.message, "secret@example.com")
	assert.NotContains(t, row.message, "trace-id")
	assert.Equal(t, "not_truncated", row.queryTruncated)
}

func TestNormalizeMongoDBSlowOperationGetMoreHash(t *testing.T) {
	commandObfuscator := newMongoDBCommandObfuscator()
	defer commandObfuscator.Stop()

	normalize := func(cursorID int64) mongoDBSlowOperationRow {
		row, ok, err := normalizeMongoDBSlowOperation(mongoDBSlowOperationRawRow{
			Timestamp: time.Unix(1700000000, 0),
			Operation: "getmore",
			Namespace: "loadtest.users",
			Command: bson.D{
				{Key: "getMore", Value: cursorID},
				{Key: "collection", Value: "users"},
				{Key: "batchSize", Value: 100},
				{Key: "$db", Value: "loadtest"},
			},
		}, "loadtest", commandObfuscator)
		require.NoError(t, err)
		require.True(t, ok)
		return row
	}

	first := normalize(1001)
	second := normalize(2002)
	assert.Equal(t, "getMore", first.commandType)
	assert.NotEqual(t, first.message, second.message)
	assert.Equal(t, first.normalizedQueryHash, second.normalizedQueryHash)

	activity, ok, err := normalizeMongoDBActivityRow(mongoDBActivityRawRow{
		Namespace: "loadtest.users",
		Operation: "getmore",
		Command: bson.D{
			{Key: "getMore", Value: int64(3003)},
			{Key: "collection", Value: "users"},
			{Key: "batchSize", Value: 100},
			{Key: "$db", Value: "loadtest"},
		},
	}, map[string]struct{}{"loadtest": {}}, commandObfuscator)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, first.normalizedQueryHash, activity.normalizedQueryHash)
}

func TestBuildMongoDBSlowOperationPoint(t *testing.T) {
	timestamp := time.Unix(1700000000, int64(123*time.Millisecond)).UTC()
	svr := &MongodbServer{
		host:             "mongo8:27017",
		databaseInstance: "mongo-test",
	}
	row := mongoDBSlowOperationRow{
		timestamp:           timestamp,
		databaseName:        "loadtest",
		collection:          "users",
		commandType:         "find",
		operation:           "query",
		planSummary:         "COLLSCAN",
		queryShapeHash:      "shape-hash",
		queryHash:           "query-hash",
		querySignature:      "query-signature",
		normalizedQueryHash: "normalized-hash",
		user:                "datakit@admin",
		application:         "datakittest",
		message:             `{"find":"users","$db":"loadtest"}`,
		namespace:           "loadtest.users",
		durationMillis:      testPtr[int64](120),
		docsReturned:        testPtr[int64](2),
		keysExamined:        testPtr[int64](10),
		docsExamined:        testPtr[int64](20),
		usedDisk:            testPtr(true),
	}

	pt := buildMongoDBSlowOperationPoint(svr, row)
	assert.Equal(t, mongodbDBMSlowQueryName, pt.Name())
	assert.Equal(t, timestamp.UnixNano(), pt.Time().UnixNano())
	assert.Equal(t, "mongo-test", pt.GetTag("database_instance"))
	assert.Equal(t, "loadtest", pt.GetTag("database_name"))
	assert.Equal(t, "users", pt.GetTag("collection"))
	assert.Equal(t, "COLLSCAN", pt.GetTag("plan_summary"))
	assert.Equal(t, "query-signature", pt.GetTag("query_signature"))
	assert.Equal(t, "normalized-hash", pt.GetTag("normalized_query_hash"))
	message, ok := pt.GetS("message")
	require.True(t, ok)
	assert.Equal(t, row.message, message)
	duration, ok := pt.GetI("duration_ms")
	require.True(t, ok)
	assert.EqualValues(t, 120, duration)
	usedDisk, ok := pt.GetB("used_disk")
	require.True(t, ok)
	assert.True(t, usedDisk)
	_, ok = pt.GetI("docs_modified")
	assert.False(t, ok)
	_, ok = pt.GetB("has_sort_stage")
	assert.False(t, ok)

	info := (&mongodbDBMSlowQueryMeasurement{}).Info()
	assert.Equal(t, point.Logging, info.Cat)
	assert.Contains(t, info.Tags, "plan_summary")
	assert.Contains(t, info.Tags, "normalized_query_hash")
	assert.Contains(t, info.Fields, "message")
	assert.Contains(t, info.Fields, "duration_ms")
	assert.Contains(t, info.Fields, "docs_examined")
}

func TestMongoDBSlowOperationOptionalFields(t *testing.T) {
	encoded, err := bson.Marshal(bson.D{
		{Key: "millis", Value: int64(0)},
		{Key: "nreturned", Value: int64(0)},
		{Key: "usedDisk", Value: false},
	})
	require.NoError(t, err)

	var raw mongoDBSlowOperationRawRow
	require.NoError(t, bson.Unmarshal(encoded, &raw))
	require.NotNil(t, raw.DurationMillis)
	require.NotNil(t, raw.DocsReturned)
	require.NotNil(t, raw.UsedDisk)
	assert.Nil(t, raw.DocsModified)

	pt := buildMongoDBSlowOperationPoint(&MongodbServer{}, mongoDBSlowOperationRow{
		timestamp:      time.Unix(100, 0),
		message:        `{}`,
		durationMillis: raw.DurationMillis,
		docsReturned:   raw.DocsReturned,
		usedDisk:       raw.UsedDisk,
	})

	duration, ok := pt.GetI("duration_ms")
	require.True(t, ok)
	assert.Zero(t, duration)
	docsReturned, ok := pt.GetI("docs_returned")
	require.True(t, ok)
	assert.Zero(t, docsReturned)
	usedDisk, ok := pt.GetB("used_disk")
	require.True(t, ok)
	assert.False(t, usedDisk)
	_, ok = pt.GetI("docs_modified")
	assert.False(t, ok)
	_, ok = pt.GetS("namespace")
	assert.False(t, ok)
}

func TestSlowOperationsStateUpdateCursors(t *testing.T) {
	state := newSlowOperationsState()
	cursor := time.Unix(1700000000, 0)
	state.updateCursors(map[string]time.Time{"loadtest": cursor})
	assert.Equal(t, cursor, state.cursors["loadtest"])
}

func TestSlowOperationsProfilingDatabaseStatusCache(t *testing.T) {
	state := newSlowOperationsState()
	checkedAt := time.Unix(1700000000, 0)
	state.profilingDatabases["enabled"] = profilingDatabaseStatus{enabled: true, checkedAt: checkedAt}
	state.profilingDatabases["disabled"] = profilingDatabaseStatus{enabled: false, checkedAt: checkedAt}

	enabled, ok := state.cachedProfilingDatabaseStatus("enabled", checkedAt.Add(profilingDatabaseStatusCacheTTL-time.Second))
	assert.True(t, ok)
	assert.True(t, enabled)

	enabled, ok = state.cachedProfilingDatabaseStatus("disabled", checkedAt.Add(time.Minute))
	assert.True(t, ok)
	assert.False(t, enabled)

	_, ok = state.cachedProfilingDatabaseStatus("enabled", checkedAt.Add(profilingDatabaseStatusCacheTTL))
	assert.False(t, ok)

	_, ok = state.cachedProfilingDatabaseStatus("missing", checkedAt)
	assert.False(t, ok)
}

func TestSlowOperationsDatabaseDiscoveryCache(t *testing.T) {
	state := newSlowOperationsState()
	checkedAt := time.Unix(1700000000, 0)
	state.setDiscoveredDatabases([]string{"config", "loadtest"}, checkedAt)

	databases, ok := state.cachedDatabases(checkedAt.Add(defaultSlowDatabaseDiscoveryInterval - time.Second))
	require.True(t, ok)
	assert.Equal(t, []string{"config", "loadtest"}, databases)

	// The returned slice must not change the cached database list.
	databases[0] = "changed"
	databases, ok = state.cachedDatabases(checkedAt.Add(time.Minute))
	require.True(t, ok)
	assert.Equal(t, []string{"config", "loadtest"}, databases)

	_, ok = state.cachedDatabases(checkedAt.Add(defaultSlowDatabaseDiscoveryInterval))
	assert.False(t, ok)

	databases, ok = state.staleDatabases()
	require.True(t, ok)
	assert.Equal(t, []string{"config", "loadtest"}, databases)
}
