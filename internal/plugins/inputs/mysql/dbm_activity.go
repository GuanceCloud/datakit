// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

	"github.com/DataDog/datadog-agent/pkg/obfuscate"
)

const maxPayloadBytes = 19e6

var (
	activityQuery5_7 = `
SELECT
    thread_a.thread_id,
    thread_a.processlist_id,
    thread_a.processlist_user,
    thread_a.processlist_host,
    thread_a.processlist_db,
    thread_a.processlist_command,
    thread_a.processlist_state,
    COALESCE(statement.sql_text, thread_a.PROCESSLIST_info) AS sql_text,
    statement.digest_text as digest_text,
    statement.timer_start AS event_timer_start,
    statement.timer_end AS event_timer_end,
    statement.lock_time,
    statement.current_schema,
    waits_a.event_id AS event_id,
    waits_a.end_event_id,
    IF(waits_a.thread_id IS NULL,
        'other',
        COALESCE(
            IF(thread_a.processlist_state = 'User sleep', 'User sleep',
            IF(waits_a.event_id = waits_a.end_event_id, 'CPU', waits_a.event_name)), 'CPU'
        )
    ) AS wait_event,
    waits_a.operation,
    waits_a.timer_start AS wait_timer_start,
    waits_a.timer_end AS wait_timer_end,
    waits_a.object_schema,
    waits_a.object_name,
    waits_a.index_name,
    waits_a.object_type,
    waits_a.source,
    blocking_thread.thread_id AS blocking_thread_id,
    blocking_thread.processlist_id AS blocking_processlist_id
FROM
    performance_schema.threads AS thread_a
    LEFT JOIN performance_schema.events_waits_current AS waits_a ON waits_a.thread_id = thread_a.thread_id
    LEFT JOIN performance_schema.events_statements_current AS statement ON statement.thread_id = thread_a.thread_id
    LEFT JOIN information_schema.INNODB_TRX AS trx ON thread_a.processlist_id = trx.trx_mysql_thread_id
    LEFT JOIN information_schema.INNODB_LOCK_WAITS AS lock_waits ON trx.trx_id = lock_waits.requesting_trx_id
    LEFT JOIN information_schema.INNODB_TRX AS blocking_trx ON lock_waits.blocking_trx_id = blocking_trx.trx_id
    LEFT JOIN performance_schema.threads AS blocking_thread ON blocking_trx.trx_mysql_thread_id = blocking_thread.processlist_id
WHERE
    (
        thread_a.processlist_state IS NOT NULL
        AND thread_a.processlist_id != CONNECTION_ID()
        AND thread_a.PROCESSLIST_COMMAND != 'Daemon'
        AND thread_a.processlist_command != 'Sleep'
        AND (waits_a.EVENT_NAME != 'idle' OR waits_a.EVENT_NAME IS NULL)
        AND (waits_a.operation != 'idle' OR waits_a.operation IS NULL)
        -- events_waits_current can have multiple rows per thread, thus we use EVENT_ID to identify the row
        -- we want to use. Additionally, we want the row with the highest EVENT_ID which reflects the most recent wait.
        AND (
            waits_a.event_id = (
            SELECT
                MAX(waits_b.EVENT_ID)
            FROM  performance_schema.events_waits_current AS waits_b
            Where waits_b.thread_id = thread_a.thread_id
        ) OR waits_a.event_id is NULL)
        -- We ignore rows without SQL text because there will be rows for background operations that do not have
        -- SQL text associated with it.
        AND COALESCE(statement.sql_text, thread_a.PROCESSLIST_info) != ''
    )
    OR
        -- Include idle sessions that are blocking others
        thread_a.processlist_id IN (
            SELECT blocking_trx.trx_mysql_thread_id
            FROM information_schema.INNODB_LOCK_WAITS AS lock_waits
            JOIN information_schema.INNODB_TRX AS blocking_trx ON lock_waits.blocking_trx_id = blocking_trx.trx_id
     )
LIMIT %d
`
	activityQuery8_0 = `
SELECT
    thread_a.thread_id,
    thread_a.processlist_id,
    thread_a.processlist_user,
    thread_a.processlist_host,
    thread_a.processlist_db,
    thread_a.processlist_command,
    thread_a.processlist_state,
    COALESCE(statement.sql_text, thread_a.PROCESSLIST_info) AS sql_text,
    statement.digest_text as digest_text,
    statement.timer_start AS event_timer_start,
    statement.timer_end AS event_timer_end,
    statement.lock_time,
    statement.current_schema,
    waits_a.event_id AS event_id,
    waits_a.end_event_id,
    IF(waits_a.thread_id IS NULL,
        'other',
        COALESCE(
            IF(thread_a.processlist_state = 'User sleep', 'User sleep',
            IF(waits_a.event_id = waits_a.end_event_id, 'CPU', waits_a.event_name)), 'CPU'
        )
    ) AS wait_event,
    waits_a.operation,
    waits_a.timer_start AS wait_timer_start,
    waits_a.timer_end AS wait_timer_end,
    waits_a.object_schema,
    waits_a.object_name,
    waits_a.index_name,
    waits_a.object_type,
    waits_a.source,
    blocking_thread.thread_id AS blocking_thread_id,
    blocking_thread.processlist_id AS blocking_processlist_id
FROM
    performance_schema.threads AS thread_a
    LEFT JOIN performance_schema.events_waits_current AS waits_a ON waits_a.thread_id = thread_a.thread_id
    LEFT JOIN performance_schema.events_statements_current AS statement ON statement.thread_id = thread_a.thread_id
    LEFT JOIN performance_schema.data_lock_waits AS lock_waits ON thread_a.thread_id = lock_waits.requesting_thread_id
    LEFT JOIN performance_schema.threads AS blocking_thread ON lock_waits.blocking_thread_id = blocking_thread.thread_id
WHERE
    (
        thread_a.processlist_state IS NOT NULL
        AND thread_a.processlist_id != CONNECTION_ID()
        AND thread_a.PROCESSLIST_COMMAND != 'Daemon'
        AND thread_a.processlist_command != 'Sleep'
        AND (waits_a.EVENT_NAME != 'idle' OR waits_a.EVENT_NAME IS NULL)
        AND (waits_a.operation != 'idle' OR waits_a.operation IS NULL)
        -- events_waits_current can have multiple rows per thread, thus we use EVENT_ID to identify the row
        -- we want to use. Additionally, we want the row with the highest EVENT_ID which reflects the most recent wait.
        AND (
            waits_a.event_id = (
            SELECT
                MAX(waits_b.EVENT_ID)
            FROM  performance_schema.events_waits_current AS waits_b
            Where waits_b.thread_id = thread_a.thread_id
        ) OR waits_a.event_id is NULL)
        -- We ignore rows without SQL text because there will be rows for background operations that do not have
        -- SQL text associated with it.
        AND COALESCE(statement.sql_text, thread_a.PROCESSLIST_info) != ''
    )
    OR
        -- Include idle sessions that are blocking others
        thread_a.thread_id IN (
            SELECT blocking_thread_id
            FROM performance_schema.data_lock_waits
    )
LIMIT %d
`
)

type dbmActivity struct {
	Enabled  bool             `toml:"enabled"`
	Interval datakit.Duration `toml:"interval"`
	Limit    int              `toml:"limit"`
}

type dbmActivityMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

// Point implement MeasurementV2.
func (m *dbmActivityMeasurement) Point() *point.Point {
	opts := point.DefaultLoggingOptions()

	if m.election {
		opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	}

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m *dbmActivityMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Desc: "Collect the waiting event of the current thread",
		Name: metricNameMySQLDbmActivity,
		Cat:  point.Logging,
		Fields: map[string]interface{}{
			"query_signature": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The hash value computed from SQL text",
			},
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The text of the normalized SQL text",
			},
			"thread_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The thread ID",
			},
			"processlist_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The process list ID",
			},
			"processlist_user": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Desc:     "The user associated with a thread",
			},
			"processlist_host": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Desc:     "The host name of the client with a thread",
			},
			"processlist_db": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Desc:     "The default database for the thread, or NULL if none has been selected",
			},
			"processlist_command": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Desc:     "The command of the thread",
			},
			"processlist_state": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Desc:     "The state of the thread",
			},
			"sql_text": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Desc:     "The statement the thread is executing",
			},
			"digest_text": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The normalized digest form of the statement",
			},
			"event_timer_start": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationNS,
				Desc:     "The time when event timing started",
			},
			"event_timer_end": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationNS,
				Desc:     "The time when event timing ended",
			},
			"event_duration": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationNS,
				Desc:     "The execution duration derived from event_timer_end - event_timer_start.",
			},
			"lock_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationNS,
				Desc:     "The time spent waiting for table locks",
			},
			"current_schema": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Desc:     "The default database for the statement, NULL if there is none",
			},
			"wait_event": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Desc:     "The name of the wait event",
			},
			"wait_group": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Desc:     "Datakit unified wait group: Lock, I/O, Concurrency, Memory, Network, CPU, Commit/Log, Other (derived from wait_event).",
			},
			"event_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Desc:     "The event id",
			},
			"end_event_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Desc:     "The thread current event number when the event ends",
			},
			"operation": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The operation of the wait event",
			},
			"wait_timer_start": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationNS,
				Desc:     "The time when the waiting event timing started",
			},
			"wait_timer_end": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationNS,
				Desc:     "The time when the waiting event timing ended",
			},
			"wait_duration": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationNS,
				Desc:     "The waiting event duration derived from wait_timer_end - wait_timer_start.",
			},
			"object_schema": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Desc:     "The schema of th object being acted on",
			},
			"object_name": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Desc:     "The name of the object being acted on",
			},
			"index_name": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Desc:     "The name of the index used",
			},
			"object_type": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Desc:     "The type of the object being acted on",
			},
			"event_source": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Desc:     "The name of the source file",
			},
			"blocking_thread_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The thread ID of the blocking thread",
			},
			"blocking_processlist_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The process list ID of the blocking thread",
			},
			"connections": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total number of the connection",
			},
		},
		Tags: map[string]interface{}{
			"host":              &inputs.TagInfo{Desc: "The server host address"},
			"server":            &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`"},
			"database_instance": &inputs.TagInfo{Desc: "MySQL instance identifier from configured tag or @@server_uuid."},
		},
	}
}

// get mysql dbm activity.
func (ipt *Input) metricCollectMysqlDbmActivity(connections []connectionRow) ([]*point.Point, activityRowSlice) {
	var pts []*point.Point
	opts := ipt.getKVsOpts(point.Logging)

	connectionsMap := map[string]int64{}
	for _, connection := range connections {
		key := connection.processlistDB.String + connection.processlistHost.String +
			connection.processlistUser.String + connection.processlistState.String
		connectionsMap[key] = connection.connections.Int64
	}

	// get activity rows
	activityRows := getActivityRows(ipt)
	activityRows = getNormalLizeActivityRows(activityRows)

	for _, activity := range activityRows {
		kvs := ipt.getKVs()

		message := ""
		if len(activity.DigestText.String) > 0 {
			message = activity.DigestText.String
		} else if len(activity.SQLText.String) > 0 {
			message = activity.SQLText.String
		}

		kvs = kvs.Set("query_signature", activity.QuerySignature)
		kvs = kvs.Set("message", message)
		kvs = kvs.Set("thread_id", activity.ThreadID.String)
		kvs = kvs.Set("processlist_id", activity.ProcesslistID.String)
		kvs = kvs.Set("processlist_user", activity.ProcesslistUser.String)
		kvs = kvs.Set("processlist_host", activity.ProcesslistHost.String)
		kvs = kvs.Set("processlist_db", activity.ProcesslistDB.String)
		kvs = kvs.Set("processlist_command", activity.ProcesslistCommand.String)
		kvs = kvs.Set("processlist_state", activity.ProcesslistState.String)
		kvs = kvs.Set("sql_text", activity.SQLText.String)
		kvs = kvs.Set("digest_text", activity.DigestText.String)
		kvs = kvs.Set("event_timer_start", activity.EventTimerStart.Int64/1000)
		kvs = kvs.Set("event_timer_end", activity.EventTimerEnd.Int64/1000)
		if activity.EventTimerEnd.Int64 >= activity.EventTimerStart.Int64 {
			kvs = kvs.Set("event_duration", (activity.EventTimerEnd.Int64-activity.EventTimerStart.Int64)/1000)
		} else {
			kvs = kvs.Set("event_duration", 0)
		}
		kvs = kvs.Set("lock_time", activity.LockTime.Int64/1000)
		kvs = kvs.Set("current_schema", activity.CurrentSchema.String)
		kvs = kvs.Set("wait_event", activity.WaitEvent.String)
		kvs = kvs.Set("wait_group", activity.WaitGroup)
		kvs = kvs.Set("event_id", activity.EventID.String)
		kvs = kvs.Set("end_event_id", activity.EndEventID.String)
		kvs = kvs.Set("operation", activity.Operation.String)
		kvs = kvs.Set("wait_timer_start", activity.WaitTimerStart.Int64/1000)
		kvs = kvs.Set("wait_timer_end", activity.WaitTimerEnd.Int64/1000)
		if activity.WaitTimerEnd.Int64 >= activity.WaitTimerStart.Int64 {
			kvs = kvs.Set("wait_duration", (activity.WaitTimerEnd.Int64-activity.WaitTimerStart.Int64)/1000)
		} else {
			kvs = kvs.Set("wait_duration", 0)
		}
		kvs = kvs.Set("object_schema", activity.ObjectSchema.String)
		kvs = kvs.Set("object_name", activity.ObjectName.String)
		kvs = kvs.Set("index_name", activity.IndexName.String)
		kvs = kvs.Set("object_type", activity.ObjectType.String)
		kvs = kvs.Set("event_source", activity.Source.String)
		kvs = kvs.Set("blocking_thread_id", activity.BlockingThreadID.String)
		kvs = kvs.Set("blocking_processlist_id", activity.BlockingProcesslistID.String)
		kvs = kvs.Set("connections", 0)

		key := activity.ProcesslistDB.String + activity.ProcesslistHost.String + activity.ProcesslistUser.String + activity.ProcesslistState.String
		if connections, ok := connectionsMap[key]; ok {
			kvs = kvs.Set("connections", connections)
		}

		pts = append(pts, point.NewPoint(metricNameMySQLDbmActivity, kvs, opts...))
	}

	return pts, activityRows
}

type connectionRow struct {
	processlistUser  sql.NullString
	processlistHost  sql.NullString
	processlistDB    sql.NullString
	processlistState sql.NullString
	connections      sql.NullInt64
}

func getActiveConnections(i *Input) (connectionRows []connectionRow) {
	ctx, cancel := context.WithTimeout(context.Background(), i.timeoutDuration)
	defer cancel()

	rows := i.q(ctx, connectionsQuerySQL, getMetricName(metricNameMySQLDbmActivity, "active_connections"))

	if rows == nil {
		return
	}

	defer closeRows(rows)

	for rows.Next() {
		row := connectionRow{}
		if err := rows.Scan(
			&row.processlistUser,
			&row.processlistHost,
			&row.processlistDB,
			&row.processlistState,
			&row.connections,
		); err != nil {
			l.Warnf("Mysql dbm activity connection row scan error: %s\n", err.Error())
		} else {
			connectionRows = append(connectionRows, row)
		}
	}

	return
}

type activityRow struct {
	QuerySignature        string         `json:"query_signature"`
	ThreadID              sql.NullString `json:"thread_id"`
	ProcesslistID         sql.NullString `json:"processlist_id"`
	ProcesslistUser       sql.NullString `json:"processlist_user"`
	ProcesslistHost       sql.NullString `json:"processlist_host"`
	ProcesslistDB         sql.NullString `json:"processlist_db"`
	ProcesslistCommand    sql.NullString `json:"processlist_command"`
	ProcesslistState      sql.NullString `json:"processlist_state"`
	SQLText               sql.NullString `json:"sql_text"`
	DigestText            sql.NullString `json:"digest_text"`
	EventTimerStart       sql.NullInt64  `json:"event_timer_start"`
	EventTimerEnd         sql.NullInt64  `json:"event_timer_end"`
	LockTime              sql.NullInt64  `json:"lock_time"`
	CurrentSchema         sql.NullString `json:"current_schema"`
	EventID               sql.NullString `json:"event_id"`
	EndEventID            sql.NullString `json:"end_event_id"`
	WaitEvent             sql.NullString `json:"wait_event"`
	Operation             sql.NullString `json:"operation"`
	WaitTimerStart        sql.NullInt64  `json:"wait_timer_start"`
	WaitTimerEnd          sql.NullInt64  `json:"wait_timer_end"`
	ObjectSchema          sql.NullString `json:"object_schema"`
	ObjectName            sql.NullString `json:"object_name"`
	IndexName             sql.NullString `json:"index_name"`
	ObjectType            sql.NullString `json:"object_type"`
	Source                sql.NullString `json:"source"`
	BlockingThreadID      sql.NullString `json:"blocking_thread_id"`
	BlockingProcesslistID sql.NullString `json:"blocking_processlist_id"`
	WaitGroup             string         `json:"wait_group"`
}

type activityRowSlice []activityRow

func (r activityRowSlice) Len() int { return len(r) }
func (r activityRowSlice) Less(i, j int) bool {
	nowVal := time.Now().UnixNano() * 1000 // picoseconds

	currentVal := nowVal
	nextVal := nowVal

	if r[i].EventTimerStart.Valid {
		currentVal = r[i].EventTimerStart.Int64
	}

	if r[j].EventTimerStart.Valid {
		nextVal = r[j].EventTimerStart.Int64
	}

	return currentVal < nextVal
}
func (r activityRowSlice) Swap(i, j int) { r[i], r[j] = r[j], r[i] }

func getActivityQuery(i *Input) string {
	if i.Version != nil && i.Version.flavor != strMariaDB && i.Version.versionCompatible([]int{8, 0, 0}) {
		return activityQuery8_0
	}
	return activityQuery5_7
}

func getActivityRows(i *Input) (activityRows []activityRow) {
	limit := i.DbmActivity.Limit
	if limit <= 0 {
		limit = 1000
	}
	query := fmt.Sprintf(getActivityQuery(i), limit)
	ctx, cancel := context.WithTimeout(context.Background(), i.timeoutDuration)
	defer cancel()

	rows := i.q(ctx, query, getMetricName(metricNameMySQLDbmActivity, "activity_rows"))
	if rows == nil {
		return
	}
	defer closeRows(rows)

	for rows.Next() {
		row := activityRow{}
		if err := rows.Scan(
			&row.ThreadID,
			&row.ProcesslistID,
			&row.ProcesslistUser,
			&row.ProcesslistHost,
			&row.ProcesslistDB,
			&row.ProcesslistCommand,
			&row.ProcesslistState,
			&row.SQLText,
			&row.DigestText,
			&row.EventTimerStart,
			&row.EventTimerEnd,
			&row.LockTime,
			&row.CurrentSchema,
			&row.EventID,
			&row.EndEventID,
			&row.WaitEvent,
			&row.Operation,
			&row.WaitTimerStart,
			&row.WaitTimerEnd,
			&row.ObjectSchema,
			&row.ObjectName,
			&row.IndexName,
			&row.ObjectType,
			&row.Source,
			&row.BlockingThreadID,
			&row.BlockingProcesslistID,
		); err != nil {
			l.Warnf("Mysql dbm activity row scan error: %s", err.Error())
		} else {
			activityRows = append(activityRows, row)
		}
	}
	return activityRows
}

func getNormalLizeActivityRows(rows activityRowSlice) activityRowSlice {
	sort.Sort(rows)
	// Create a fresh obfuscator per normalization run to avoid shared state across inputs.
	o := newMySQLSQLObfuscator()

	// Deduplicate rows per thread, keeping only the most recent statement per thread,
	// following the Datadog activity normalization logic.
	seen := make(map[string]int64)       // thread_id -> first seen event_timer_start
	secondPass := make(map[string]int64) // thread_id -> later event_timer_start for more recent row

	size := 0
	normalizedRows := activityRowSlice{}
	for _, row := range rows {
		if row.ThreadID.Valid {
			threadID := row.ThreadID.String
			eventStart := int64(0)
			if row.EventTimerStart.Valid {
				eventStart = row.EventTimerStart.Int64
			}

			if prevStart, ok := seen[threadID]; ok {
				// If this row ended before the first seen event started, skip it as an older statement.
				if row.EventTimerEnd.Valid && row.EventTimerEnd.Int64 < prevStart {
					continue
				}
				// Otherwise, mark for second pass to remove older duplicates.
				if row.EventTimerStart.Valid {
					secondPass[threadID] = eventStart
				}
			} else if row.EventTimerStart.Valid {
				seen[threadID] = eventStart
			}
		}

		obfuscatedRow, ok := obfuscateRow(row, o)
		if !ok {
			continue
		}
		assignActivityRowWaitGroup(&obfuscatedRow)

		size += getEstimatedRowSizeBytes(obfuscatedRow)

		if size > maxPayloadBytes {
			return normalizedRows
		}

		normalizedRows = append(normalizedRows, obfuscatedRow)
	}

	if len(secondPass) > 0 {
		normalizedRows = eliminateDuplicateActivityRows(normalizedRows, secondPass)
	}

	return normalizedRows
}

// eliminateDuplicateActivityRows removes older activity rows for threads that have multiple
// statements, keeping only rows whose event_timer_end is after the newer event_timer_start.
func eliminateDuplicateActivityRows(rows activityRowSlice, secondPass map[string]int64) activityRowSlice {
	filtered := activityRowSlice{}
	for _, row := range rows {
		if row.ThreadID.Valid && row.EventTimerEnd.Valid {
			if start, ok := secondPass[row.ThreadID.String]; ok && row.EventTimerEnd.Int64 < start {
				// Older row, drop it.
				continue
			}
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func obfuscateRow(row activityRow, o *obfuscate.Obfuscator) (activityRow, bool) {
	var obfSQLResult, obfDigestResult *obfuscate.ObfuscatedQuery
	var err error

	if row.SQLText.Valid && len(row.SQLText.String) > 0 {
		obfSQLResult, err = o.ObfuscateSQLString(row.SQLText.String)
		if err != nil {
			l.Warnf("obfuscate sql text failed: %s, sql: %s", err.Error(), row.SQLText.String)
			return row, false
		}
		row.SQLText.String = obfSQLResult.Query
	}

	if row.DigestText.Valid && len(row.DigestText.String) > 0 {
		obfDigestResult, err = o.ObfuscateSQLString(row.DigestText.String)
		if err != nil {
			l.Warnf("obfuscate digest text failed: %s, digest: %s", err.Error(), row.DigestText.String)
			return row, false
		}
		row.DigestText.String = obfDigestResult.Query
	}

	if row.DigestText.Valid && len(row.DigestText.String) > 0 {
		row.QuerySignature = generateQuerySignature(row.CurrentSchema.String, row.DigestText.String)
	} else if row.SQLText.Valid && len(row.SQLText.String) > 0 {
		row.QuerySignature = generateQuerySignature(row.CurrentSchema.String, row.SQLText.String)
	}

	return row, true
}

func getEstimatedRowSizeBytes(row activityRow) int {
	if bytes, err := json.Marshal(row); err != nil {
		return 0
	} else {
		return len(bytes)
	}
}

func assignActivityRowWaitGroup(row *activityRow) {
	w := row.WaitEvent.String
	if !row.WaitEvent.Valid || w == "" {
		w = "other"
	}
	row.WaitGroup = mapMysqlWaitEventToGroup(w)
}

func mapMysqlWaitEventToGroup(waitEvent string) string {
	low := strings.ToLower(strings.TrimSpace(waitEvent))

	if low == "cpu" {
		return "CPU"
	}
	if low == "" || low == "other" || strings.HasPrefix(low, "idle") || strings.Contains(low, "sleep") {
		return "Other"
	}

	switch {
	case strings.HasPrefix(low, "wait/synch/"):
		return "Concurrency"
	case strings.HasPrefix(low, "wait/io/table/sql/handler"):
		return "Concurrency"
	case strings.Contains(low, "/lock/"), strings.Contains(low, "lock"):
		return "Lock"
	case strings.Contains(low, "/socket/"), strings.Contains(low, "sql/net"):
		return "Network"
	case strings.Contains(low, "redo"), strings.Contains(low, "binlog"), strings.Contains(low, "innodb_log"):
		return "Commit/Log"
	case strings.HasPrefix(low, "wait/io/"):
		return "I/O"
	case strings.Contains(low, "/memory/"), strings.Contains(low, "buffer_pool"):
		return "Memory"
	default:
		return "Other"
	}
}
