// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mongodb

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

const mongodbDBMOperationName = "mongodb_dbm_operation"

type mongoDBOperationKey struct {
	namespace     string
	databaseName  string
	commandType   string
	operation     string
	user          string
	clientAddress string
	application   string
	shard         string
}

type mongoDBOperationRow struct {
	mongoDBOperationKey
	activeOperationCount int64
	waitingForLockCount  int64
}

func aggregateMongoDBOperations(activityRows []mongoDBActivityRow) []mongoDBOperationRow {
	aggregated := make(map[mongoDBOperationKey]*mongoDBOperationRow)
	for _, activityRow := range activityRows {
		if activityRow.active == nil || !*activityRow.active {
			continue
		}

		key := mongoDBOperationKey{
			namespace:     activityRow.namespace,
			databaseName:  activityRow.databaseName,
			commandType:   activityRow.commandType,
			operation:     activityRow.operation,
			user:          activityRow.user,
			clientAddress: activityRow.clientAddress,
			application:   activityRow.application,
			shard:         activityRow.shard,
		}
		row, ok := aggregated[key]
		if !ok {
			row = &mongoDBOperationRow{mongoDBOperationKey: key}
			aggregated[key] = row
		}
		row.activeOperationCount++
		if activityRow.waitingForLock != nil && *activityRow.waitingForLock {
			row.waitingForLockCount++
		}
	}

	rows := make([]mongoDBOperationRow, 0, len(aggregated))
	for _, row := range aggregated {
		rows = append(rows, *row)
	}
	return rows
}

func buildMongoDBOperationPoints(
	svr *MongodbServer,
	rows []mongoDBOperationRow,
	ptsTime time.Time,
) []*point.Point {
	points := make([]*point.Point, 0, len(rows))
	opts := append(point.DefaultMetricOptions(), point.WithTime(ptsTime))
	for _, row := range rows {
		var kvs point.KVs
		for key, value := range svr.getDefaultTags() {
			kvs = kvs.AddTag(key, value)
		}
		kvs = kvs.AddTag("database_type", "MongoDB")
		kvs = kvs.AddTag("namespace", row.namespace)
		kvs = kvs.AddTag("database_name", row.databaseName)
		if row.commandType != "" {
			kvs = kvs.AddTag("command_type", row.commandType)
		}
		if row.operation != "" {
			kvs = kvs.AddTag("operation", row.operation)
		}
		if row.user != "" {
			kvs = kvs.AddTag("user", row.user)
		}
		if row.clientAddress != "" {
			kvs = kvs.AddTag("client_address", row.clientAddress)
		}
		if row.application != "" {
			kvs = kvs.AddTag("application", row.application)
		}
		if row.shard != "" {
			kvs = kvs.AddTag("shard", row.shard)
		}
		kvs = kvs.Set("active_operation_count", row.activeOperationCount)
		kvs = kvs.Set("waiting_for_lock_count", row.waitingForLockCount)
		points = append(points, point.NewPoint(mongodbDBMOperationName, kvs, opts...))
	}
	return points
}
