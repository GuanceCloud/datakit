// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mongodb

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/DataDog/datadog-agent/pkg/obfuscate"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongodbDBMSlowQueryName         = "mongodb_dbm_slow_query"
	profilingDatabaseStatusCacheTTL = time.Hour
)

type slowOperationsState struct {
	cursors             map[string]time.Time
	profilingDatabases  map[string]profilingDatabaseStatus
	discoveredDatabases []string
	databasesCheckedAt  time.Time
}

type profilingDatabaseStatus struct {
	enabled   bool
	checkedAt time.Time
}

type mongoDBSlowOperationRawRow struct {
	Timestamp          time.Time `bson:"ts"`
	Operation          string    `bson:"op"`
	Namespace          string    `bson:"ns"`
	Command            bson.D    `bson:"command"`
	QueryShapeHash     string    `bson:"queryShapeHash"`
	QueryHash          string    `bson:"queryHash"`
	PlanCacheShapeHash string    `bson:"planCacheShapeHash"`
	PlanCacheKey       string    `bson:"planCacheKey"`
	PlanSummary        string    `bson:"planSummary"`
	QueryFramework     string    `bson:"queryFramework"`
	User               string    `bson:"user"`
	Application        string    `bson:"appName"`
	Client             string    `bson:"client"`
	DurationMillis     *int64    `bson:"millis"`
	WorkingMillis      *int64    `bson:"workingMillis"`
	NumYields          *int64    `bson:"numYield"`
	ResponseLength     *int64    `bson:"responseLength"`
	DocsReturned       *int64    `bson:"nreturned"`
	DocsMatched        *int64    `bson:"nMatched"`
	DocsModified       *int64    `bson:"nModified"`
	DocsInserted       *int64    `bson:"ninserted"`
	DocsDeleted        *int64    `bson:"ndeleted"`
	KeysExamined       *int64    `bson:"keysExamined"`
	DocsExamined       *int64    `bson:"docsExamined"`
	KeysInserted       *int64    `bson:"keysInserted"`
	WriteConflicts     *int64    `bson:"writeConflicts"`
	CPUNanos           *int64    `bson:"cpuNanos"`
	PlanningTimeMicros *int64    `bson:"planningTimeMicros"`
	CursorExhausted    *bool     `bson:"cursorExhausted"`
	HasSortStage       *bool     `bson:"hasSortStage"`
	UsedDisk           *bool     `bson:"usedDisk"`
	FromMultiPlanner   *bool     `bson:"fromMultiPlanner"`
	Replanned          *bool     `bson:"replanned"`
}

type mongoDBSlowOperationRow struct {
	timestamp           time.Time
	databaseName        string
	collection          string
	commandType         string
	operation           string
	planSummary         string
	queryShapeHash      string
	queryHash           string
	querySignature      string
	normalizedQueryHash string
	user                string
	application         string
	message             string
	namespace           string
	planCacheKey        string
	queryFramework      string
	client              string
	queryTruncated      string
	durationMillis      *int64
	workingMillis       *int64
	numYields           *int64
	responseLength      *int64
	docsReturned        *int64
	docsMatched         *int64
	docsModified        *int64
	docsInserted        *int64
	docsDeleted         *int64
	keysExamined        *int64
	docsExamined        *int64
	keysInserted        *int64
	writeConflicts      *int64
	cpuNanos            *int64
	planningTimeMicros  *int64
	cursorExhausted     *bool
	hasSortStage        *bool
	usedDisk            *bool
	fromMultiPlanner    *bool
	replanned           *bool
}

func newSlowOperationsState() *slowOperationsState {
	return &slowOperationsState{
		cursors:            make(map[string]time.Time),
		profilingDatabases: make(map[string]profilingDatabaseStatus),
	}
}

func (ipt *Input) runSlowOperations(ctx context.Context) {
	duration := ipt.slowOperationsInterval()
	tick := time.NewTicker(duration)
	defer tick.Stop()

	commandObfuscator := newMongoDBCommandObfuscator()
	defer commandObfuscator.Stop()

	if ipt.slowOperationsState == nil {
		ipt.slowOperationsState = newSlowOperationsState()
	}

	for {
		if ipt.pause.Load() {
			log.Debugf("not leader, MongoDB slow operations collection skipped")
		} else {
			collectCtx, cancel := context.WithTimeout(ctx, duration)
			start := time.Now()
			points, cursors, err := ipt.collectMongoDBSlowOperations(collectCtx, commandObfuscator)
			cancel()
			if err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(inputName),
					metrics.WithLastErrorCategory(point.Logging),
				)
				log.Errorf("collect MongoDB slow operations failed: %s", err)
			} else if len(points) == 0 {
				ipt.slowOperationsState.updateCursors(cursors)
			} else if err := ipt.feeder.Feed(point.Logging, points,
				dkio.WithCollectCost(time.Since(start)),
				dkio.WithElection(ipt.Election),
				dkio.WithSource(dbmFeedName),
				dkio.WithInput(inputName),
			); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(inputName),
					metrics.WithLastErrorCategory(point.Logging),
				)
				log.Errorf("feed MongoDB slow operations failed: %s", err)
			} else {
				ipt.slowOperationsState.updateCursors(cursors)
			}
		}

		select {
		case <-ctx.Done():
			log.Info("MongoDB slow operations collection exit")
			return
		case <-datakit.Exit.Wait():
			log.Info("MongoDB slow operations collection exit")
			return
		case <-ipt.semStop.Wait():
			log.Info("MongoDB slow operations collection return")
			return
		case <-tick.C:
		}
	}
}

func (ipt *Input) collectMongoDBSlowOperations(
	ctx context.Context,
	commandObfuscator *obfuscate.Obfuscator,
) ([]*point.Point, map[string]time.Time, error) {
	if len(ipt.mgoSvrs) == 0 {
		return nil, nil, nil
	}
	if ipt.slowOperationsState == nil {
		ipt.slowOperationsState = newSlowOperationsState()
	}

	svr := ipt.mgoSvrs[0]
	if !svr.canCollectDBM(ctx) {
		return nil, nil, nil
	}

	databases, err := ipt.slowOperationDatabases(ctx, svr)
	if err != nil {
		return nil, nil, err
	}

	collectionTime := ntp.Now()
	remaining := ipt.slowOperationsLimit()
	points := make([]*point.Point, 0)
	nextCursors := make(map[string]time.Time)
	for _, databaseName := range databases {
		if remaining <= 0 {
			break
		}
		if !ipt.slowOperationsState.profilingEnabled(ctx, svr, databaseName) {
			continue
		}

		from := ipt.slowOperationsState.cursors[databaseName]
		if from.IsZero() {
			nextCursors[databaseName] = collectionTime
			continue
		}
		rows, err := loadMongoDBSlowOperations(ctx, svr, databaseName, from, collectionTime, remaining)
		if err != nil {
			log.Warnf("query MongoDB slow operations from database %s on %s failed: %s", databaseName, svr.host, err)
			continue
		}

		nextCursors[databaseName] = collectionTime
		for _, rawRow := range rows {
			row, ok, err := normalizeMongoDBSlowOperation(rawRow, databaseName, commandObfuscator)
			if err != nil {
				log.Warnf("normalize MongoDB slow operation from database %s failed: %s", databaseName, err)
				continue
			}
			if !ok {
				continue
			}
			points = append(points, buildMongoDBSlowOperationPoint(svr, row))
		}
		remaining -= int64(len(rows))
	}

	return points, nextCursors, nil
}

func (ipt *Input) slowOperationDatabases(ctx context.Context, svr *MongodbServer) ([]string, error) {
	if len(ipt.Dbm.Databases) > 0 {
		result := append([]string(nil), ipt.Dbm.Databases...)
		sort.Strings(result)
		return result, nil
	}
	if databases, ok := ipt.slowOperationsState.cachedDatabases(time.Now()); ok {
		return databases, nil
	}

	databases, err := svr.cli.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		if databases, ok := ipt.slowOperationsState.staleDatabases(); ok {
			log.Warnf("refresh MongoDB databases for slow operations on %s failed: %s; using cached databases", svr.host, err)
			return databases, nil
		}
		return nil, fmt.Errorf("list MongoDB databases for slow operations on %s: %w", svr.host, err)
	}
	result := databases[:0]
	for _, databaseName := range databases {
		if databaseName != "admin" {
			result = append(result, databaseName)
		}
	}
	sort.Strings(result)
	if len(result) > defaultSlowMaxDiscoveredDatabases {
		log.Warnf("discovered %d MongoDB databases on %s; slow operations collection is limited to %d databases",
			len(result), svr.host, defaultSlowMaxDiscoveredDatabases)
		result = result[:defaultSlowMaxDiscoveredDatabases]
	}
	ipt.slowOperationsState.setDiscoveredDatabases(result, time.Now())
	return append([]string(nil), result...), nil
}

func (state *slowOperationsState) cachedDatabases(now time.Time) ([]string, bool) {
	if state.databasesCheckedAt.IsZero() ||
		now.Sub(state.databasesCheckedAt) >= defaultSlowDatabaseDiscoveryInterval {
		return nil, false
	}
	return append([]string(nil), state.discoveredDatabases...), true
}

func (state *slowOperationsState) staleDatabases() ([]string, bool) {
	if state.databasesCheckedAt.IsZero() {
		return nil, false
	}
	return append([]string(nil), state.discoveredDatabases...), true
}

func (state *slowOperationsState) setDiscoveredDatabases(databases []string, checkedAt time.Time) {
	state.discoveredDatabases = append(state.discoveredDatabases[:0], databases...)
	state.databasesCheckedAt = checkedAt
}

func (state *slowOperationsState) profilingEnabled(
	ctx context.Context,
	svr *MongodbServer,
	databaseName string,
) bool {
	if status, ok := state.cachedProfilingDatabaseStatus(databaseName, time.Now()); ok {
		return status
	}

	var status struct {
		Level int `bson:"was"`
	}
	err := svr.cli.Database(databaseName).
		RunCommand(ctx, bson.D{{Key: "profile", Value: -1}}).
		Decode(&status)
	if err != nil {
		state.profilingDatabases[databaseName] = profilingDatabaseStatus{
			enabled:   false,
			checkedAt: time.Now(),
		}
		log.Warnf("check MongoDB profiling level for database %s on %s failed: %s", databaseName, svr.host, err)
		return false
	}

	enabled := status.Level > 0
	state.profilingDatabases[databaseName] = profilingDatabaseStatus{
		enabled:   enabled,
		checkedAt: time.Now(),
	}
	if !enabled {
		log.Debugf("MongoDB profiling is disabled for database %s on %s; skipping slow operations", databaseName, svr.host)
	}
	return enabled
}

func (state *slowOperationsState) cachedProfilingDatabaseStatus(
	databaseName string,
	now time.Time,
) (bool, bool) {
	status, ok := state.profilingDatabases[databaseName]
	if !ok || now.Sub(status.checkedAt) >= profilingDatabaseStatusCacheTTL {
		return false, false
	}
	return status.enabled, true
}

func (state *slowOperationsState) updateCursors(cursors map[string]time.Time) {
	for databaseName, cursor := range cursors {
		state.cursors[databaseName] = cursor
	}
}

func loadMongoDBSlowOperations(
	ctx context.Context,
	svr *MongodbServer,
	databaseName string,
	from time.Time,
	to time.Time,
	limit int64,
) ([]mongoDBSlowOperationRawRow, error) {
	filter := bson.D{{Key: "ts", Value: bson.D{
		{Key: "$gt", Value: from},
		{Key: "$lte", Value: to},
	}}}
	findOptions := options.Find().
		SetSort(bson.D{{Key: "ts", Value: 1}}).
		SetLimit(limit)
	if deadline, ok := ctx.Deadline(); ok {
		if maxTime := time.Until(deadline); maxTime > 0 {
			findOptions.SetMaxTime(maxTime)
		}
	}

	cursor, err := svr.cli.Database(databaseName).Collection("system.profile").Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Warnf("close MongoDB slow operations cursor failed: %s", err)
		}
	}()

	var rows []mongoDBSlowOperationRawRow
	for cursor.Next(ctx) {
		var row mongoDBSlowOperationRawRow
		if err := cursor.Decode(&row); err != nil {
			log.Warnf("decode MongoDB slow operation from database %s failed: %s", databaseName, err)
			continue
		}
		rows = append(rows, row)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return rows, nil
}

func normalizeMongoDBSlowOperation(
	rawRow mongoDBSlowOperationRawRow,
	defaultDatabaseName string,
	commandObfuscator *obfuscate.Obfuscator,
) (mongoDBSlowOperationRow, bool, error) {
	if rawRow.Timestamp.IsZero() || len(rawRow.Command) == 0 {
		return mongoDBSlowOperationRow{}, false, nil
	}

	namespaceDatabase, namespaceCollection := mongoDBActivityNamespace(rawRow.Namespace)
	databaseName := namespaceDatabase
	if databaseName == "" {
		databaseName = defaultDatabaseName
	}
	if value, ok := queryStatsDocumentValue(rawRow.Command, "$db"); ok {
		if commandDatabase, ok := value.(string); ok && commandDatabase != "" {
			databaseName = commandDatabase
		}
	}
	collection := mongoDBActivityCollection(rawRow.Command, namespaceCollection)
	commandType := mongoDBSlowOperationCommandType(rawRow.Operation, rawRow.Command)

	command := mongoDBActivityCommandWithoutMetadata(rawRow.Command)
	message, err := obfuscateMongoDBCommand(commandObfuscator, command)
	if err != nil {
		return mongoDBSlowOperationRow{}, false, fmt.Errorf("obfuscate command: %w", err)
	}

	normalizedCommand := mongoDBCommandForNormalizedHash(rawRow.Command)
	normalizedMessage, err := obfuscateMongoDBCommand(commandObfuscator, normalizedCommand)
	if err != nil {
		return mongoDBSlowOperationRow{}, false, fmt.Errorf("obfuscate normalized command: %w", err)
	}
	normalizedQueryHash, err := computeMongoDBNormalizedQueryHash(normalizedMessage)
	if err != nil {
		return mongoDBSlowOperationRow{}, false, fmt.Errorf("compute normalized query hash: %w", err)
	}

	queryHash := rawRow.QueryHash
	if queryHash == "" {
		queryHash = rawRow.PlanCacheShapeHash
	}
	row := mongoDBSlowOperationRow{
		timestamp:           rawRow.Timestamp,
		databaseName:        databaseName,
		collection:          collection,
		commandType:         commandType,
		operation:           rawRow.Operation,
		planSummary:         rawRow.PlanSummary,
		queryShapeHash:      rawRow.QueryShapeHash,
		queryHash:           queryHash,
		normalizedQueryHash: normalizedQueryHash,
		user:                rawRow.User,
		application:         rawRow.Application,
		message:             message,
		namespace:           rawRow.Namespace,
		planCacheKey:        rawRow.PlanCacheKey,
		queryFramework:      rawRow.QueryFramework,
		client:              rawRow.Client,
		queryTruncated:      mongoDBActivityTruncationState(rawRow.Command),
		durationMillis:      rawRow.DurationMillis,
		workingMillis:       rawRow.WorkingMillis,
		numYields:           rawRow.NumYields,
		responseLength:      rawRow.ResponseLength,
		docsReturned:        rawRow.DocsReturned,
		docsMatched:         rawRow.DocsMatched,
		docsModified:        rawRow.DocsModified,
		docsInserted:        rawRow.DocsInserted,
		docsDeleted:         rawRow.DocsDeleted,
		keysExamined:        rawRow.KeysExamined,
		docsExamined:        rawRow.DocsExamined,
		keysInserted:        rawRow.KeysInserted,
		writeConflicts:      rawRow.WriteConflicts,
		cpuNanos:            rawRow.CPUNanos,
		planningTimeMicros:  rawRow.PlanningTimeMicros,
		cursorExhausted:     rawRow.CursorExhausted,
		hasSortStage:        rawRow.HasSortStage,
		usedDisk:            rawRow.UsedDisk,
		fromMultiPlanner:    rawRow.FromMultiPlanner,
		replanned:           rawRow.Replanned,
	}
	if row.queryShapeHash != "" {
		row.querySignature = generateMongoDBQuerySignature(databaseName, collection, row.queryShapeHash)
	}
	return row, true, nil
}

func mongoDBSlowOperationCommandType(operation string, command bson.D) string {
	commandType := mongoDBActivityCommandType(command)
	switch operation {
	case "update", "insert", "remove", "delete":
		return operation
	case "getmore":
		return "getMore"
	case "query":
		if commandType == "q" || commandType == "query" {
			return "find"
		}
	}
	return commandType
}

func buildMongoDBSlowOperationPoint(svr *MongodbServer, row mongoDBSlowOperationRow) *point.Point {
	var kvs point.KVs
	for key, value := range svr.getDefaultTags() {
		kvs = kvs.AddTag(key, value)
	}

	kvs = kvs.AddTag("database_type", "MongoDB")
	kvs = kvs.AddTag("database_name", row.databaseName)
	if row.collection != "" {
		kvs = kvs.AddTag("collection", row.collection)
	}
	if row.commandType != "" {
		kvs = kvs.AddTag("command_type", row.commandType)
	}
	if row.operation != "" {
		kvs = kvs.AddTag("operation", row.operation)
	}
	if row.planSummary != "" {
		kvs = kvs.AddTag("plan_summary", row.planSummary)
	}
	if row.queryShapeHash != "" {
		kvs = kvs.AddTag("query_shape_hash", row.queryShapeHash)
	}
	if row.queryHash != "" {
		kvs = kvs.AddTag("query_hash", row.queryHash)
	}
	if row.querySignature != "" {
		kvs = kvs.AddTag("query_signature", row.querySignature)
	}
	if row.normalizedQueryHash != "" {
		kvs = kvs.AddTag("normalized_query_hash", row.normalizedQueryHash)
	}
	if row.user != "" {
		kvs = kvs.AddTag("user", row.user)
	}
	if row.application != "" {
		kvs = kvs.AddTag("application", row.application)
	}

	kvs = kvs.Set("message", row.message)
	if row.namespace != "" {
		kvs = kvs.Set("namespace", row.namespace)
	}
	if row.planCacheKey != "" {
		kvs = kvs.Set("plan_cache_key", row.planCacheKey)
	}
	if row.queryFramework != "" {
		kvs = kvs.Set("query_framework", row.queryFramework)
	}
	if row.client != "" {
		kvs = kvs.Set("client", row.client)
	}
	kvs = kvs.Set("query_truncated", row.queryTruncated)
	if row.durationMillis != nil {
		kvs = kvs.Set("duration_ms", *row.durationMillis)
	}
	if row.workingMillis != nil {
		kvs = kvs.Set("working_millis", *row.workingMillis)
	}
	if row.numYields != nil {
		kvs = kvs.Set("num_yields", *row.numYields)
	}
	if row.responseLength != nil {
		kvs = kvs.Set("response_length", *row.responseLength)
	}
	if row.docsReturned != nil {
		kvs = kvs.Set("docs_returned", *row.docsReturned)
	}
	if row.docsMatched != nil {
		kvs = kvs.Set("docs_matched", *row.docsMatched)
	}
	if row.docsModified != nil {
		kvs = kvs.Set("docs_modified", *row.docsModified)
	}
	if row.docsInserted != nil {
		kvs = kvs.Set("docs_inserted", *row.docsInserted)
	}
	if row.docsDeleted != nil {
		kvs = kvs.Set("docs_deleted", *row.docsDeleted)
	}
	if row.keysExamined != nil {
		kvs = kvs.Set("keys_examined", *row.keysExamined)
	}
	if row.docsExamined != nil {
		kvs = kvs.Set("docs_examined", *row.docsExamined)
	}
	if row.keysInserted != nil {
		kvs = kvs.Set("keys_inserted", *row.keysInserted)
	}
	if row.writeConflicts != nil {
		kvs = kvs.Set("write_conflicts", *row.writeConflicts)
	}
	if row.cpuNanos != nil {
		kvs = kvs.Set("cpu_nanos", *row.cpuNanos)
	}
	if row.planningTimeMicros != nil {
		kvs = kvs.Set("planning_time_micros", *row.planningTimeMicros)
	}
	if row.cursorExhausted != nil {
		kvs = kvs.Set("cursor_exhausted", *row.cursorExhausted)
	}
	if row.hasSortStage != nil {
		kvs = kvs.Set("has_sort_stage", *row.hasSortStage)
	}
	if row.usedDisk != nil {
		kvs = kvs.Set("used_disk", *row.usedDisk)
	}
	if row.fromMultiPlanner != nil {
		kvs = kvs.Set("from_multi_planner", *row.fromMultiPlanner)
	}
	if row.replanned != nil {
		kvs = kvs.Set("replanned", *row.replanned)
	}

	opts := append(point.DefaultLoggingOptions(), point.WithTime(row.timestamp))
	return point.NewPoint(mongodbDBMSlowQueryName, kvs, opts...)
}
