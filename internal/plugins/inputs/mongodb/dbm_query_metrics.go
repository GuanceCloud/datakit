// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mongodb

import (
	"context"
	"fmt"
	"maps"
	"regexp"
	"strconv"
	"time"

	"github.com/DataDog/datadog-agent/pkg/obfuscate"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
	lru "github.com/hashicorp/golang-lru/v2"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongodbDBMMetricName      = "mongodb_dbm_metric"
	mongodbDBMQueryObjectName = "db_query"
)

var (
	queryStatsTypePattern = regexp.MustCompile(`^\?(string|number|date|bool|objectId|array|object|binData|null|regex|timestamp)$`)
	queryStatsCopyFields  = []string{
		"filter", "projection", "sort", "pipeline", "hint", "limit", "skip", "batchSize",
		"let", "collation", "arrayFilters", "key",
	}
	queryStatsMetricFields = []queryStatsMetricField{
		{mongoName: "execCount", name: "exec_count", extraction: queryStatsDirect},
		{mongoName: "totalExecMicros", name: "total_exec_micros_sum", extraction: queryStatsSum},
		{mongoName: "firstResponseExecMicros", name: "first_response_exec_micros_sum", extraction: queryStatsSum},
		{mongoName: "keysExamined", name: "keys_examined_sum", extraction: queryStatsSum},
		{mongoName: "docsExamined", name: "docs_examined_sum", extraction: queryStatsSum},
		{mongoName: "docsReturned", name: "docs_returned_sum", extraction: queryStatsSum},
		{mongoName: "bytesRead", name: "bytes_read_sum", extraction: queryStatsSum},
		{mongoName: "cpuNanos", name: "cpu_nanos_sum", extraction: queryStatsSum},
		{mongoName: "usedDisk", name: "used_disk_count", extraction: queryStatsTrueCount},
		{mongoName: "hasSortStage", name: "has_sort_stage_count", extraction: queryStatsTrueCount},
		{mongoName: "readTimeMicros", name: "read_time_micros_sum", extraction: queryStatsSum},
		{mongoName: "workingTimeMillis", name: "working_time_millis_sum", extraction: queryStatsSum},
	}
)

type queryStatsExtraction uint8

const (
	queryStatsDirect queryStatsExtraction = iota
	queryStatsSum
	queryStatsTrueCount
)

type queryStatsMetricField struct {
	mongoName  string
	name       string
	extraction queryStatsExtraction
}

type queryStatsKey struct {
	QueryShape bson.D `bson:"queryShape"`
}

type queryStatsRawRow struct {
	Key            queryStatsKey `bson:"key"`
	QueryShapeHash string        `bson:"queryShapeHash"`
	Metrics        bson.M        `bson:"metrics"`
}

type queryMetricRow struct {
	databaseName        string
	collection          string
	commandType         string
	querySignature      string
	queryShapeHash      string
	normalizedQueryHash string
	obfuscatedCommand   string
	queryText           string
	queryTextTruncated  bool
	metrics             map[string]int64
	deltaMetrics        map[string]int64
}

type queryMetricsState struct {
	previous         map[string]map[string]int64
	queryObjectCache *lru.Cache[string, time.Time]
	obfuscator       *obfuscate.Obfuscator
}

func newQueryMetricsState() (*queryMetricsState, error) {
	queryObjectCache, err := lru.New[string, time.Time](queryObjectCacheSize)
	if err != nil {
		return nil, fmt.Errorf("create MongoDB query object cache: %w", err)
	}

	return &queryMetricsState{
		previous:         make(map[string]map[string]int64),
		queryObjectCache: queryObjectCache,
		obfuscator:       newMongoDBCommandObfuscator(),
	}, nil
}

func (ipt *Input) collectQueryMetrics(ctx context.Context, ptsTime time.Time) {
	if ipt.queryMetricsState == nil {
		state, err := newQueryMetricsState()
		if err != nil {
			log.Errorf("init MongoDB query metrics state: %s", err)
			return
		}
		ipt.queryMetricsState = state
	}

	queryTextMaxBytes := ipt.queryTextMaxBytes()

	if len(ipt.mgoSvrs) == 0 {
		return
	}

	// The servers option is legacy. DBM belongs to one database instance, so
	// collect only from the first successfully initialized server.
	svr := ipt.mgoSvrs[0]
	if !svr.canCollectDBM(ctx) {
		return
	}

	start := time.Now()
	rawRows, err := loadQueryStats(ctx, svr.cli.Database("admin"), ipt.Dbm.Databases, ipt.queryMetricsLimit())
	if err != nil {
		log.Errorf("query MongoDB $queryStats on %s failed: %s", svr.host, err)
		return
	}

	normalizedRows := make([]queryMetricRow, 0, len(rawRows))
	for _, rawRow := range rawRows {
		row, err := ipt.queryMetricsState.normalizeRow(rawRow, queryTextMaxBytes)
		if err != nil {
			log.Warnf("normalize MongoDB query metrics row failed: %s", err)
			continue
		}
		normalizedRows = append(normalizedRows, row)
	}

	deltaRows := ipt.queryMetricsState.computeDeltas(normalizedRows)
	if len(deltaRows) == 0 {
		return
	}

	metricPoints := buildMongoDBQueryMetricPoints(svr, deltaRows, ptsTime)
	if err := ipt.feeder.Feed(point.Metric, metricPoints,
		dkio.WithCollectCost(time.Since(start)),
		dkio.WithElection(ipt.Election),
		dkio.WithSource(dbmFeedName),
		dkio.WithInput(inputName),
	); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(point.Metric),
		)
		log.Errorf("feed MongoDB query metrics failed: %s", err)
	}

	queryObjects, querySignatures := ipt.queryMetricsState.buildQueryObjectPoints(svr, deltaRows, ptsTime)
	if len(queryObjects) > 0 {
		if err := ipt.feeder.Feed(point.Object, queryObjects,
			dkio.WithCollectCost(time.Since(start)),
			dkio.WithElection(ipt.Election),
			dkio.WithSource(dbmFeedName),
			dkio.WithInput(inputName),
		); err != nil {
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorCategory(point.Object),
			)
			log.Errorf("feed MongoDB query objects failed: %s", err)
		} else {
			reportedAt := time.Now()
			for _, querySignature := range querySignatures {
				ipt.queryMetricsState.queryObjectCache.Add(querySignature, reportedAt)
			}
		}
	}
}

func loadQueryStats(
	ctx context.Context,
	database *mongo.Database,
	databases []string,
	limit int64,
) ([]queryStatsRawRow, error) {
	pipeline := queryStatsPipeline(databases, limit)
	aggregateOptions := options.Aggregate()
	if deadline, ok := ctx.Deadline(); ok {
		if maxTime := time.Until(deadline); maxTime > 0 {
			aggregateOptions.SetMaxTime(maxTime)
		}
	}
	cursor, err := database.Aggregate(ctx, pipeline, aggregateOptions)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Warnf("close MongoDB query metrics cursor failed: %s", err)
		}
	}()

	var rows []queryStatsRawRow
	for cursor.Next(ctx) {
		var row queryStatsRawRow
		if err := cursor.Decode(&row); err != nil {
			log.Warnf("decode MongoDB query metrics row failed: %s", err)
			continue
		}
		rows = append(rows, row)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return rows, nil
}

func queryStatsPipeline(databases []string, limit int64) mongo.Pipeline {
	databaseMatch := bson.D{{Key: "$ne", Value: "admin"}}
	if len(databases) > 0 {
		databaseMatch = bson.D{{Key: "$in", Value: databases}}
	}

	return mongo.Pipeline{
		bson.D{{Key: "$queryStats", Value: bson.D{}}},
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "key.queryShape.cmdNs.db", Value: databaseMatch},
		}}},
		bson.D{{Key: "$limit", Value: limit}},
	}
}

func (state *queryMetricsState) normalizeRow(
	rawRow queryStatsRawRow,
	queryTextMaxBytes int,
) (queryMetricRow, error) {
	queryShape := rawRow.Key.QueryShape
	if len(queryShape) == 0 {
		return queryMetricRow{}, fmt.Errorf("missing queryShape document")
	}
	databaseName, collection, commandType := queryStatsCommandInfo(queryShape)
	if databaseName == "" {
		return queryMetricRow{}, fmt.Errorf("missing database name")
	}

	if commandType == "" {
		commandType = "unknown"
	}
	queryShapeHash := rawRow.QueryShapeHash
	if queryShapeHash == "" {
		return queryMetricRow{}, fmt.Errorf("missing queryShapeHash")
	}
	metricDocument := rawRow.Metrics
	if metricDocument == nil {
		return queryMetricRow{}, fmt.Errorf("missing metrics document")
	}
	metricValues := make(map[string]int64, len(queryStatsMetricFields))
	for _, field := range queryStatsMetricFields {
		if value, ok := extractQueryStatsMetric(metricDocument, field); ok {
			metricValues[field.name] = value
		}
	}
	if _, ok := metricValues["exec_count"]; !ok {
		return queryMetricRow{}, fmt.Errorf("missing execCount metric")
	}

	reconstructedCommand := reconstructQueryStatsCommand(queryShape)
	obfuscatedCommand, err := obfuscateMongoDBCommand(state.obfuscator, reconstructedCommand)
	if err != nil {
		return queryMetricRow{}, fmt.Errorf("obfuscate reconstructed command: %w", err)
	}
	normalizedQueryHash, err := computeMongoDBNormalizedQueryHash(obfuscatedCommand)
	if err != nil {
		return queryMetricRow{}, fmt.Errorf("compute normalized query hash: %w", err)
	}
	queryText, queryTextTruncated := util.TruncateUTF8ByBytes(obfuscatedCommand, queryTextMaxBytes)

	return queryMetricRow{
		databaseName:        databaseName,
		collection:          collection,
		commandType:         commandType,
		querySignature:      generateMongoDBQuerySignature(databaseName, collection, queryShapeHash),
		queryShapeHash:      queryShapeHash,
		normalizedQueryHash: normalizedQueryHash,
		obfuscatedCommand:   obfuscatedCommand,
		queryText:           queryText,
		queryTextTruncated:  queryTextTruncated,
		metrics:             metricValues,
	}, nil
}

func generateMongoDBQuerySignature(databaseName, collection, queryShapeHash string) string {
	return fmt.Sprintf("%016x", xxhash.Sum64String(databaseName+":"+collection+":"+queryShapeHash))
}

func queryStatsDocumentValue(document bson.D, key string) (interface{}, bool) {
	for _, element := range document {
		if element.Key == key {
			return element.Value, true
		}
	}
	return nil, false
}

func queryStatsCommandInfo(queryShape bson.D) (databaseName, collection, commandType string) {
	if commandNamespaceValue, ok := queryStatsDocumentValue(queryShape, "cmdNs"); ok {
		if commandNamespace, ok := commandNamespaceValue.(bson.D); ok {
			if databaseValue, ok := queryStatsDocumentValue(commandNamespace, "db"); ok {
				databaseName, _ = databaseValue.(string)
			}
			if collectionValue, ok := queryStatsDocumentValue(commandNamespace, "coll"); ok {
				collection, _ = collectionValue.(string)
			}
		}
	}
	if commandValue, ok := queryStatsDocumentValue(queryShape, "command"); ok {
		commandType, _ = commandValue.(string)
	}
	return databaseName, collection, commandType
}

func reconstructQueryStatsCommand(queryShape bson.D) bson.D {
	databaseName, collection, commandType := queryStatsCommandInfo(queryShape)
	if commandType == "" {
		commandType = "unknown"
	}

	command := bson.D{
		{Key: commandType, Value: collection},
		{Key: "$db", Value: databaseName},
	}
	for _, field := range queryStatsCopyFields {
		if value, ok := queryStatsDocumentValue(queryShape, field); ok {
			command = append(command, bson.E{Key: field, Value: normalizeQueryStatsValue(value)})
		}
	}

	return command
}

func normalizeQueryStatsValue(value interface{}) interface{} {
	switch value := value.(type) {
	case string:
		if queryStatsTypePattern.MatchString(value) {
			return "?"
		}
		return value
	case bson.D:
		normalized := make(bson.D, 0, len(value))
		for _, element := range value {
			normalized = append(normalized, bson.E{
				Key:   element.Key,
				Value: normalizeQueryStatsValue(element.Value),
			})
		}
		return normalized
	case bson.A:
		normalized := make(bson.A, len(value))
		for index, item := range value {
			normalized[index] = normalizeQueryStatsValue(item)
		}
		return normalized
	default:
		return value
	}
}

func extractQueryStatsMetric(metrics bson.M, field queryStatsMetricField) (int64, bool) {
	value, ok := metrics[field.mongoName]
	if !ok {
		return 0, false
	}

	switch field.extraction {
	case queryStatsDirect:
		return queryStatsInt64(value)
	case queryStatsSum:
		summary, ok := value.(bson.M)
		if !ok {
			return 0, false
		}
		return queryStatsInt64(summary["sum"])
	case queryStatsTrueCount:
		summary, ok := value.(bson.M)
		if !ok {
			return 0, false
		}
		return queryStatsInt64(summary["true"])
	default:
		return 0, false
	}
}

func queryStatsInt64(value interface{}) (int64, bool) {
	switch value := value.(type) {
	case int32:
		return int64(value), true
	case int64:
		return value, true
	default:
		return 0, false
	}
}

func (state *queryMetricsState) computeDeltas(rows []queryMetricRow) []queryMetricRow {
	if len(rows) == 0 {
		return nil
	}

	merged := make(map[string]queryMetricRow, len(rows))
	for _, row := range rows {
		key := row.querySignature
		if current, ok := merged[key]; ok {
			for metric, value := range row.metrics {
				current.metrics[metric] += value
			}
			merged[key] = current
			continue
		}
		row.metrics = maps.Clone(row.metrics)
		merged[key] = row
	}

	result := make([]queryMetricRow, 0, len(merged))
	for key, row := range merged {
		previous, ok := state.previous[key]
		if !ok {
			continue
		}
		currentExecCount, ok := row.metrics["exec_count"]
		if !ok {
			continue
		}
		previousExecCount, ok := previous["exec_count"]
		if !ok || currentExecCount <= previousExecCount {
			continue
		}

		deltas := make(map[string]int64, len(row.metrics))
		hasNegative := false
		for metric, currentValue := range row.metrics {
			previousValue, ok := previous[metric]
			if !ok {
				continue
			}
			delta := currentValue - previousValue
			if delta < 0 {
				hasNegative = true
				break
			}
			deltas[metric] = delta
		}
		if hasNegative {
			continue
		}

		row.deltaMetrics = deltas
		result = append(result, row)
	}

	for key := range state.previous {
		if _, ok := merged[key]; !ok {
			delete(state.previous, key)
		}
	}
	for key, row := range merged {
		state.previous[key] = maps.Clone(row.metrics)
	}

	return result
}

func buildMongoDBQueryMetricPoints(svr *MongodbServer, rows []queryMetricRow, ptsTime time.Time) []*point.Point {
	opts := append(point.DefaultMetricOptions(), point.WithTime(ptsTime))
	points := make([]*point.Point, 0, len(rows))
	for _, row := range rows {
		kvs := mongoDBQueryMetricTags(svr, row)
		for _, field := range queryStatsMetricFields {
			if value, ok := row.metrics[field.name]; ok {
				kvs = kvs.Set(field.name, value)
			}
			if value, ok := row.deltaMetrics[field.name]; ok {
				kvs = kvs.Set("delta_"+field.name, value)
			}
		}
		if row.queryText != "" {
			kvs = kvs.AddTag("query_text", row.queryText)
			kvs = kvs.AddTag("query_truncated", strconv.FormatBool(row.queryTextTruncated))
		}
		points = append(points, point.NewPoint(mongodbDBMMetricName, kvs, opts...))
	}

	return points
}

func (state *queryMetricsState) buildQueryObjectPoints(
	svr *MongodbServer,
	rows []queryMetricRow,
	ptsTime time.Time,
) ([]*point.Point, []string) {
	opts := append(point.DefaultObjectOptions(), point.WithTime(ptsTime))
	points := make([]*point.Point, 0, len(rows))
	querySignatures := make([]string, 0, len(rows))
	for _, row := range rows {
		if reportedAt, ok := state.queryObjectCache.Get(row.querySignature); ok {
			if time.Since(reportedAt) < queryObjectCacheTTL {
				continue
			}
			state.queryObjectCache.Remove(row.querySignature)
		}

		kvs := mongoDBQueryMetricTags(svr, row)
		objectName := fmt.Sprintf("%s-%s", svr.host, row.querySignature)
		if svr.databaseInstance != "" {
			objectName = fmt.Sprintf("%s-%s-%s", svr.host, svr.databaseInstance, row.querySignature)
		}
		kvs = kvs.AddTag("name", objectName)
		kvs = kvs.Set("message", row.obfuscatedCommand)
		points = append(points, point.NewPoint(mongodbDBMQueryObjectName, kvs, opts...))
		querySignatures = append(querySignatures, row.querySignature)
	}

	return points, querySignatures
}

func mongoDBQueryMetricTags(svr *MongodbServer, row queryMetricRow) point.KVs {
	var kvs point.KVs
	for key, value := range svr.getDefaultTags() {
		kvs = kvs.AddTag(key, value)
	}
	kvs = kvs.AddTag("database_type", "MongoDB")
	kvs = kvs.AddTag("database_name", row.databaseName)
	kvs = kvs.AddTag("collection", row.collection)
	kvs = kvs.AddTag("command_type", row.commandType)
	kvs = kvs.AddTag("query_signature", row.querySignature)
	kvs = kvs.AddTag("query_shape_hash", row.queryShapeHash)
	if row.normalizedQueryHash != "" {
		kvs = kvs.AddTag("normalized_query_hash", row.normalizedQueryHash)
	}

	return kvs
}
