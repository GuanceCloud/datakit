// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mongodb

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"
	"unicode"

	"github.com/DataDog/datadog-agent/pkg/obfuscate"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const mongodbDBMActivityName = "mongodb_dbm_activity"

var mongoDBActivityCollectionKeys = []string{
	"collection", "find", "aggregate", "update", "insert", "delete", "findAndModify", "distinct", "count",
}

type mongoDBActivityUser struct {
	User string `bson:"user"`
}

type mongoDBActivityRawRow struct {
	Type                  string                `bson:"type"`
	Shard                 string                `bson:"shard"`
	Description           string                `bson:"desc"`
	Active                *bool                 `bson:"active"`
	CurrentOpTime         interface{}           `bson:"currentOpTime"`
	EffectiveUsers        []mongoDBActivityUser `bson:"effectiveUsers"`
	OperationID           interface{}           `bson:"opid"`
	MicrosecondsRunning   *int64                `bson:"microsecs_running"`
	Operation             string                `bson:"op"`
	Namespace             string                `bson:"ns"`
	Command               bson.D                `bson:"command"`
	QueryShapeHash        string                `bson:"queryShapeHash"`
	PlanSummary           string                `bson:"planSummary"`
	QueryFramework        string                `bson:"queryFramework"`
	PrepareReadConflicts  *int64                `bson:"prepareReadConflicts"`
	WriteConflicts        *int64                `bson:"writeConflicts"`
	NumYields             *int64                `bson:"numYields"`
	WaitingForLock        *bool                 `bson:"waitingForLock"`
	Locks                 bson.D                `bson:"locks"`
	LockStats             bson.D                `bson:"lockStats"`
	WaitingForFlowControl *bool                 `bson:"waitingForFlowControl"`
	FlowControlStats      bson.D                `bson:"flowControlStats"`
	WaitingForLatch       bson.D                `bson:"waitingForLatch"`
	Cursor                bson.D                `bson:"cursor"`
	Transaction           bson.D                `bson:"transaction"`
	LogicalSessionID      bson.D                `bson:"lsid"`
	Client                string                `bson:"client"`
	ClientFromMongos      string                `bson:"client_s"`
	ClientMetadata        bson.D                `bson:"clientMetadata"`
	Application           string                `bson:"appName"`
}

type mongoDBActivityRow struct {
	databaseName          string
	collection            string
	commandType           string
	querySignature        string
	queryShapeHash        string
	normalizedQueryHash   string
	message               string
	activityType          string
	operation             string
	shard                 string
	application           string
	user                  string
	description           string
	operationID           string
	namespace             string
	planSummary           string
	queryFramework        string
	currentOpTime         string
	queryTruncated        string
	client                string
	clientAddress         string
	locks                 string
	lockStats             string
	flowControlStats      string
	waitingForLatch       string
	cursor                string
	transaction           string
	logicalSessionID      string
	active                *bool
	microsecondsRunning   *int64
	prepareReadConflicts  *int64
	writeConflicts        *int64
	numYields             *int64
	waitingForLock        *bool
	waitingForFlowControl *bool
}

func (ipt *Input) runActivity(ctx context.Context) {
	duration := ipt.activityInterval()
	tick := time.NewTicker(duration)
	defer tick.Stop()

	commandObfuscator := newMongoDBCommandObfuscator()
	defer commandObfuscator.Stop()

	ptsTime := ntp.Now()
	for {
		if ipt.pause.Load() {
			log.Debugf("not leader, MongoDB activity collection skipped")
		} else {
			collectCtx, cancel := context.WithTimeout(ctx, duration)
			start := time.Now()
			activityPoints, operationPoints, err := ipt.collectMongoDBActivity(collectCtx, commandObfuscator, ptsTime)
			cancel()
			if err != nil {
				log.Errorf("collect MongoDB activity failed: %s", err)
			} else {
				if len(activityPoints) > 0 {
					if err := ipt.feeder.Feed(point.Logging, activityPoints,
						dkio.WithCollectCost(time.Since(start)),
						dkio.WithElection(ipt.Election),
						dkio.WithSource(dbmFeedName),
						dkio.WithInput(inputName),
					); err != nil {
						ipt.feeder.FeedLastError(err.Error(),
							metrics.WithLastErrorInput(inputName),
							metrics.WithLastErrorCategory(point.Logging),
						)
						log.Errorf("feed MongoDB activity failed: %s", err)
					}
				}
				if len(operationPoints) > 0 {
					if err := ipt.feeder.Feed(point.Metric, operationPoints,
						dkio.WithCollectCost(time.Since(start)),
						dkio.WithElection(ipt.Election),
						dkio.WithSource(dbmFeedName),
						dkio.WithInput(inputName),
					); err != nil {
						ipt.feeder.FeedLastError(err.Error(),
							metrics.WithLastErrorInput(inputName),
							metrics.WithLastErrorCategory(point.Metric),
						)
						log.Errorf("feed MongoDB operation metrics failed: %s", err)
					}
				}
			}
		}

		select {
		case <-ctx.Done():
			log.Info("MongoDB activity collection exit")
			return
		case <-datakit.Exit.Wait():
			log.Info("MongoDB activity collection exit")
			return
		case <-ipt.semStop.Wait():
			log.Info("MongoDB activity collection return")
			return
		case tt := <-tick.C:
			ptsTime = inputs.AlignTime(tt, ptsTime, duration)
		}
	}
}

func (ipt *Input) collectMongoDBActivity(
	ctx context.Context,
	commandObfuscator *obfuscate.Obfuscator,
	ptsTime time.Time,
) ([]*point.Point, []*point.Point, error) {
	if len(ipt.mgoSvrs) == 0 {
		return nil, nil, nil
	}

	// The servers option is legacy. DBM belongs to one database instance, so
	// collect only from the first successfully initialized server.
	svr := ipt.mgoSvrs[0]
	if !svr.canCollectDBM(ctx) {
		return nil, nil, nil
	}

	databaseFilter := make(map[string]struct{}, len(ipt.Dbm.Databases))
	for _, database := range ipt.Dbm.Databases {
		databaseFilter[database] = struct{}{}
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$currentOp", Value: bson.D{{Key: "allUsers", Value: true}}}},
	}
	aggregateOptions := options.Aggregate()
	if deadline, ok := ctx.Deadline(); ok {
		if maxTime := time.Until(deadline); maxTime > 0 {
			aggregateOptions.SetMaxTime(maxTime)
		}
	}
	cursor, err := svr.cli.Database("admin").Aggregate(ctx, pipeline, aggregateOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("query $currentOp on %s: %w", svr.host, err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Warnf("close MongoDB activity cursor failed: %s", err)
		}
	}()

	var activityPoints []*point.Point
	var activityRows []mongoDBActivityRow
	for cursor.Next(ctx) {
		var rawRow mongoDBActivityRawRow
		if err := cursor.Decode(&rawRow); err != nil {
			log.Warnf("decode MongoDB activity row failed: %s", err)
			continue
		}
		row, include, err := normalizeMongoDBActivityRow(rawRow, databaseFilter, commandObfuscator)
		if err != nil {
			log.Warnf("normalize MongoDB activity row failed: %s", err)
			continue
		}
		if !include {
			continue
		}
		activityPoints = append(activityPoints, buildMongoDBActivityPoint(svr, row, ptsTime))
		activityRows = append(activityRows, row)
	}
	if err := cursor.Err(); err != nil {
		return nil, nil, fmt.Errorf("read $currentOp results from %s: %w", svr.host, err)
	}

	operationPoints := buildMongoDBOperationPoints(svr, aggregateMongoDBOperations(activityRows), ptsTime)
	return activityPoints, operationPoints, nil
}

func normalizeMongoDBActivityRow(
	rawRow mongoDBActivityRawRow,
	databaseFilter map[string]struct{},
	commandObfuscator *obfuscate.Obfuscator,
) (mongoDBActivityRow, bool, error) {
	namespaceDatabase, namespaceCollection := mongoDBActivityNamespace(rawRow.Namespace)
	if namespaceDatabase == "" {
		return mongoDBActivityRow{}, false, nil
	}
	if len(databaseFilter) == 0 {
		if namespaceDatabase == "admin" {
			return mongoDBActivityRow{}, false, nil
		}
	} else {
		if _, ok := databaseFilter[namespaceDatabase]; !ok {
			return mongoDBActivityRow{}, false, nil
		}
	}
	if len(rawRow.Command) == 0 {
		return mongoDBActivityRow{}, false, nil
	}
	if rawRow.Operation == "command" {
		if _, ok := queryStatsDocumentValue(rawRow.Command, "hello"); ok {
			return mongoDBActivityRow{}, false, nil
		}
	}

	databaseName := namespaceDatabase
	if value, ok := queryStatsDocumentValue(rawRow.Command, "$db"); ok {
		if commandDatabase, ok := value.(string); ok && commandDatabase != "" {
			databaseName = commandDatabase
		}
	}
	collection := mongoDBActivityCollection(rawRow.Command, namespaceCollection)
	commandType := mongoDBActivityCommandType(rawRow.Command)

	command := mongoDBActivityCommandWithoutMetadata(rawRow.Command)
	message, err := obfuscateMongoDBCommand(commandObfuscator, command)
	if err != nil {
		return mongoDBActivityRow{}, false, fmt.Errorf("obfuscate command: %w", err)
	}

	normalizedCommand := mongoDBCommandForNormalizedHash(rawRow.Command)
	normalizedMessage, err := obfuscateMongoDBCommand(commandObfuscator, normalizedCommand)
	if err != nil {
		return mongoDBActivityRow{}, false, fmt.Errorf("obfuscate normalized command: %w", err)
	}
	normalizedQueryHash, err := computeMongoDBNormalizedQueryHash(normalizedMessage)
	if err != nil {
		return mongoDBActivityRow{}, false, fmt.Errorf("compute normalized query hash: %w", err)
	}

	client, err := mongoDBActivityClient(rawRow)
	if err != nil {
		return mongoDBActivityRow{}, false, fmt.Errorf("marshal client metadata: %w", err)
	}
	cursor, err := mongoDBActivityCursor(rawRow.Cursor, commandObfuscator)
	if err != nil {
		return mongoDBActivityRow{}, false, fmt.Errorf("normalize cursor: %w", err)
	}

	row := mongoDBActivityRow{
		databaseName:          databaseName,
		collection:            collection,
		commandType:           commandType,
		queryShapeHash:        rawRow.QueryShapeHash,
		normalizedQueryHash:   normalizedQueryHash,
		message:               message,
		activityType:          rawRow.Type,
		operation:             rawRow.Operation,
		shard:                 rawRow.Shard,
		application:           rawRow.Application,
		description:           rawRow.Description,
		operationID:           mongoDBActivityString(rawRow.OperationID),
		namespace:             rawRow.Namespace,
		planSummary:           rawRow.PlanSummary,
		queryFramework:        rawRow.QueryFramework,
		currentOpTime:         mongoDBActivityString(rawRow.CurrentOpTime),
		queryTruncated:        mongoDBActivityTruncationState(rawRow.Command),
		client:                client,
		clientAddress:         mongoDBActivityClientAddress(rawRow),
		locks:                 mongoDBActivityDocumentJSON(rawRow.Locks, true),
		lockStats:             mongoDBActivityDocumentJSON(rawRow.LockStats, true),
		flowControlStats:      mongoDBActivityDocumentJSON(rawRow.FlowControlStats, true),
		waitingForLatch:       mongoDBActivityDocumentJSON(rawRow.WaitingForLatch, true),
		cursor:                cursor,
		transaction:           mongoDBActivityTransaction(rawRow.Transaction),
		logicalSessionID:      mongoDBActivityLogicalSessionID(rawRow.LogicalSessionID),
		active:                rawRow.Active,
		microsecondsRunning:   rawRow.MicrosecondsRunning,
		prepareReadConflicts:  rawRow.PrepareReadConflicts,
		writeConflicts:        rawRow.WriteConflicts,
		numYields:             rawRow.NumYields,
		waitingForLock:        rawRow.WaitingForLock,
		waitingForFlowControl: rawRow.WaitingForFlowControl,
	}
	if row.queryShapeHash != "" {
		row.querySignature = generateMongoDBQuerySignature(databaseName, collection, row.queryShapeHash)
	}
	if len(rawRow.EffectiveUsers) > 0 {
		row.user = rawRow.EffectiveUsers[0].User
	}

	return row, true, nil
}

func mongoDBActivityNamespace(namespace string) (string, string) {
	database, collection, found := strings.Cut(namespace, ".")
	if !found {
		return database, ""
	}
	return database, collection
}

func mongoDBActivityCollection(command bson.D, namespaceCollection string) string {
	if namespaceCollection != "" && namespaceCollection != "$cmd" {
		return namespaceCollection
	}
	for _, key := range mongoDBActivityCollectionKeys {
		if value, ok := queryStatsDocumentValue(command, key); ok {
			if collection, ok := value.(string); ok {
				return collection
			}
		}
	}
	return ""
}

func mongoDBActivityCommandType(command bson.D) string {
	for _, element := range command {
		switch element.Key {
		case "comment", "lsid", "$clusterTime", "$db", "$readPreference", "$truncated":
			continue
		default:
			return element.Key
		}
	}
	return "unknown"
}

func mongoDBActivityCommandWithoutMetadata(command bson.D) bson.D {
	result := make(bson.D, 0, len(command))
	for _, element := range command {
		switch element.Key {
		case "comment", "lsid", "$clusterTime":
			continue
		default:
			result = append(result, element)
		}
	}
	return result
}

func mongoDBCommandForNormalizedHash(command bson.D) bson.D {
	result := mongoDBActivityCommandWithoutMetadata(command)
	for index := range result {
		if result[index].Key == "getMore" {
			result[index].Value = "?"
			break
		}
	}
	return result
}

func mongoDBActivityTruncationState(command bson.D) string {
	value, ok := queryStatsDocumentValue(command, "$truncated")
	if !ok {
		return "not_truncated"
	}
	switch truncated := value.(type) {
	case bool:
		if truncated {
			return "truncated"
		}
	case string:
		if truncated != "" {
			return "truncated"
		}
	}
	return "not_truncated"
}

func mongoDBActivityClient(rawRow mongoDBActivityRawRow) (string, error) {
	hostname := mongoDBActivityClientEndpoint(rawRow)
	client := bson.D{{Key: "hostname", Value: hostname}}
	for _, key := range []string{"driver", "os", "platform"} {
		if value, ok := queryStatsDocumentValue(rawRow.ClientMetadata, key); ok {
			client = append(client, bson.E{Key: key, Value: value})
		}
	}
	data, err := bson.MarshalExtJSON(client, false, false)
	return string(data), err
}

func mongoDBActivityClientAddress(rawRow mongoDBActivityRawRow) string {
	endpoint := mongoDBActivityClientEndpoint(rawRow)
	if host, _, err := net.SplitHostPort(endpoint); err == nil {
		return host
	}
	return endpoint
}

func mongoDBActivityClientEndpoint(rawRow mongoDBActivityRawRow) string {
	if rawRow.Client != "" {
		return rawRow.Client
	}
	return rawRow.ClientFromMongos
}

func mongoDBActivityCursor(cursor bson.D, commandObfuscator *obfuscate.Obfuscator) (string, error) {
	if len(cursor) == 0 {
		return "", nil
	}

	result := bson.D{}
	fields := []struct {
		mongoName string
		name      string
	}{
		{mongoName: "cursorId", name: "cursor_id"},
		{mongoName: "createdDate", name: "created_date"},
		{mongoName: "lastAccessDate", name: "last_access_date"},
		{mongoName: "nDocsReturned", name: "n_docs_returned"},
		{mongoName: "nBatchesReturned", name: "n_batches_returned"},
		{mongoName: "noCursorTimeout", name: "no_cursor_timeout"},
		{mongoName: "tailable", name: "tailable"},
		{mongoName: "awaitData", name: "await_data"},
		{mongoName: "planSummary", name: "plan_summary"},
		{mongoName: "operationUsingCursorId", name: "operation_using_cursor_id"},
	}
	for _, field := range fields {
		if value, ok := queryStatsDocumentValue(cursor, field.mongoName); ok {
			if field.mongoName == "createdDate" || field.mongoName == "lastAccessDate" {
				value = mongoDBActivityString(value)
			}
			result = append(result, bson.E{Key: field.name, Value: value})
		}
	}

	if value, ok := queryStatsDocumentValue(cursor, "originatingCommand"); ok {
		if originatingCommand, ok := value.(bson.D); ok {
			obfuscatedCommand, err := obfuscateMongoDBCommand(
				commandObfuscator,
				mongoDBActivityCommandWithoutMetadata(originatingCommand),
			)
			if err != nil {
				return "", err
			}
			result = append(result, bson.E{Key: "originating_command", Value: obfuscatedCommand})
		}
	}

	data, err := bson.MarshalExtJSON(result, false, false)
	return string(data), err
}

func mongoDBActivityTransaction(transaction bson.D) string {
	if len(transaction) == 0 {
		return ""
	}
	result := bson.D{}
	if value, ok := queryStatsDocumentValue(transaction, "parameters"); ok {
		if parameters, ok := value.(bson.D); ok {
			for _, field := range []struct {
				mongoName string
				name      string
			}{
				{mongoName: "txnNumber", name: "txn_number"},
				{mongoName: "txnRetryCounter", name: "txn_retry_counter"},
			} {
				if parameter, ok := queryStatsDocumentValue(parameters, field.mongoName); ok {
					result = append(result, bson.E{Key: field.name, Value: parameter})
				}
			}
		}
	}
	for _, field := range []struct {
		mongoName string
		name      string
	}{
		{mongoName: "timeOpenMicros", name: "time_open_micros"},
		{mongoName: "timeActiveMicros", name: "time_active_micros"},
		{mongoName: "timeInactiveMicros", name: "time_inactive_micros"},
	} {
		if value, ok := queryStatsDocumentValue(transaction, field.mongoName); ok {
			result = append(result, bson.E{Key: field.name, Value: value})
		}
	}
	return mongoDBActivityDocumentJSON(result, false)
}

func mongoDBActivityLogicalSessionID(lsid bson.D) string {
	value, ok := queryStatsDocumentValue(lsid, "id")
	if !ok {
		return ""
	}
	binary, ok := value.(primitive.Binary)
	if !ok {
		return ""
	}
	return mongoDBActivityDocumentJSON(bson.D{{Key: "id", Value: hex.EncodeToString(binary.Data)}}, false)
}

func mongoDBActivityDocumentJSON(document bson.D, formatKeys bool) string {
	if len(document) == 0 {
		return "{}"
	}
	if formatKeys {
		document, _ = formatMongoDBActivityValue(document).(bson.D)
	}
	data, err := bson.MarshalExtJSON(document, false, false)
	if err != nil {
		log.Warnf("marshal MongoDB activity document failed: %s", err)
		return ""
	}
	return string(data)
}

func formatMongoDBActivityValue(value interface{}) interface{} {
	switch value := value.(type) {
	case bson.D:
		result := make(bson.D, 0, len(value))
		for _, element := range value {
			result = append(result, bson.E{
				Key:   mongoDBActivitySnakeCase(element.Key),
				Value: formatMongoDBActivityValue(element.Value),
			})
		}
		return result
	case bson.A:
		result := make(bson.A, len(value))
		for index, item := range value {
			result[index] = formatMongoDBActivityValue(item)
		}
		return result
	default:
		return value
	}
}

func mongoDBActivitySnakeCase(value string) string {
	var result strings.Builder
	characters := []rune(value)
	for index, character := range characters {
		if unicode.IsUpper(character) {
			if index > 0 && (unicode.IsLower(characters[index-1]) || unicode.IsDigit(characters[index-1]) ||
				(index+1 < len(characters) && unicode.IsLower(characters[index+1]))) {
				result.WriteByte('_')
			}
			result.WriteRune(unicode.ToLower(character))
			continue
		}
		result.WriteRune(character)
	}
	return result.String()
}

func mongoDBActivityString(value interface{}) string {
	switch value := value.(type) {
	case nil:
		return ""
	case string:
		return value
	case primitive.DateTime:
		return value.Time().UTC().Format(time.RFC3339Nano)
	case time.Time:
		return value.UTC().Format(time.RFC3339Nano)
	default:
		return fmt.Sprint(value)
	}
}

func buildMongoDBActivityPoint(svr *MongodbServer, row mongoDBActivityRow, ptsTime time.Time) *point.Point {
	var kvs point.KVs
	for key, value := range svr.getDefaultTags() {
		kvs = kvs.AddTag(key, value)
	}

	// Tags.
	kvs = kvs.AddTag("database_type", "MongoDB")
	kvs = kvs.AddTag("database_name", row.databaseName)
	if row.collection != "" {
		kvs = kvs.AddTag("collection", row.collection)
	}
	kvs = kvs.AddTag("command_type", row.commandType)
	if row.querySignature != "" {
		kvs = kvs.AddTag("query_signature", row.querySignature)
	}
	if row.queryShapeHash != "" {
		kvs = kvs.AddTag("query_shape_hash", row.queryShapeHash)
	}
	kvs = kvs.AddTag("normalized_query_hash", row.normalizedQueryHash)
	if row.activityType != "" {
		kvs = kvs.AddTag("activity_type", row.activityType)
	}
	if row.operation != "" {
		kvs = kvs.AddTag("operation", row.operation)
	}
	if row.shard != "" {
		kvs = kvs.AddTag("shard", row.shard)
	}
	if row.application != "" {
		kvs = kvs.AddTag("application", row.application)
	}
	if row.user != "" {
		kvs = kvs.AddTag("user", row.user)
	}

	// Fields.
	kvs = kvs.Set("message", row.message)
	kvs = kvs.Set("namespace", row.namespace)
	if row.active != nil {
		kvs = kvs.Set("active", *row.active)
	}
	if row.description != "" {
		kvs = kvs.Set("description", row.description)
	}
	if row.operationID != "" {
		kvs = kvs.Set("operation_id", row.operationID)
	}
	if row.planSummary != "" {
		kvs = kvs.Set("plan_summary", row.planSummary)
	}
	if row.queryFramework != "" {
		kvs = kvs.Set("query_framework", row.queryFramework)
	}
	if row.currentOpTime != "" {
		kvs = kvs.Set("current_op_time", row.currentOpTime)
	}
	if row.microsecondsRunning != nil {
		kvs = kvs.Set("microsecs_running", *row.microsecondsRunning)
	}
	if row.prepareReadConflicts != nil {
		kvs = kvs.Set("prepare_read_conflicts", *row.prepareReadConflicts)
	}
	if row.writeConflicts != nil {
		kvs = kvs.Set("write_conflicts", *row.writeConflicts)
	}
	if row.numYields != nil {
		kvs = kvs.Set("num_yields", *row.numYields)
	}
	if row.waitingForLock != nil {
		kvs = kvs.Set("waiting_for_lock", *row.waitingForLock)
	}
	if row.locks != "" {
		kvs = kvs.Set("locks", row.locks)
	}
	if row.lockStats != "" {
		kvs = kvs.Set("lock_stats", row.lockStats)
	}
	if row.waitingForFlowControl != nil {
		kvs = kvs.Set("waiting_for_flow_control", *row.waitingForFlowControl)
	}
	if row.flowControlStats != "" {
		kvs = kvs.Set("flow_control_stats", row.flowControlStats)
	}
	if row.waitingForLatch != "" {
		kvs = kvs.Set("waiting_for_latch", row.waitingForLatch)
	}
	if row.cursor != "" {
		kvs = kvs.Set("cursor", row.cursor)
	}
	if row.transaction != "" {
		kvs = kvs.Set("transaction", row.transaction)
	}
	if row.logicalSessionID != "" {
		kvs = kvs.Set("logical_session_id", row.logicalSessionID)
	}
	if row.client != "" {
		kvs = kvs.Set("client", row.client)
	}
	kvs = kvs.Set("query_truncated", row.queryTruncated)

	opts := append(point.DefaultLoggingOptions(), point.WithTime(ptsTime))
	return point.NewPoint(mongodbDBMActivityName, kvs, opts...)
}
