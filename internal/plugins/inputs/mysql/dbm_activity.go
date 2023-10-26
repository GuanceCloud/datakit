// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"database/sql"
	"encoding/json"
	"sort"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const maxPayloadBytes = 19e6

const activityQuerySQL = `
SELECT
    thread_a.thread_id,
    thread_a.processlist_id,
    thread_a.processlist_user,
    thread_a.processlist_host,
    thread_a.processlist_db,
    thread_a.processlist_command,
    thread_a.processlist_state,
    thread_a.processlist_info AS sql_text,
    statement.timer_start AS event_timer_start,
    statement.timer_end AS event_timer_end,
	statement.timer_wait AS event_timer_wait,
    statement.lock_time,
    statement.current_schema,
    COALESCE(
        IF(thread_a.processlist_state = 'User sleep', 'User sleep',
        IF(waits_a.event_id = waits_a.end_event_id, 'CPU', waits_a.event_name)), 'CPU') AS wait_event,
    waits_a.event_id,
    waits_a.end_event_id,
    waits_a.event_name,
    waits_a.timer_start AS wait_timer_start,
    waits_a.timer_end AS wait_timer_end,
    waits_a.object_schema,
    waits_a.object_name,
    waits_a.index_name,
    waits_a.object_type,
    waits_a.source,
    socket.ip,
    socket.port,
    socket.event_name AS socket_event_name
FROM
    performance_schema.threads AS thread_a
    -- events_waits_current can have multiple rows per thread, thus we use EVENT_ID to identify the row we want to use.
    -- Additionally, we want the row with the highest EVENT_ID which reflects the most recent and current wait.
    LEFT JOIN performance_schema.events_waits_current AS waits_a ON waits_a.thread_id = thread_a.thread_id AND
    waits_a.event_id IN(
        SELECT
            MAX(waits_b.EVENT_ID)
        FROM performance_schema.threads AS thread_b
            LEFT JOIN performance_schema.events_waits_current AS waits_b ON waits_b.thread_id = thread_b.thread_id
        WHERE
            thread_b.processlist_state IS NOT NULL AND
            thread_b.processlist_command != 'Sleep' AND
            thread_b.processlist_id != connection_id()
        GROUP BY thread_b.thread_id)
    LEFT JOIN performance_schema.events_statements_current AS statement ON statement.thread_id = thread_a.thread_id
    LEFT JOIN performance_schema.socket_instances AS socket ON socket.thread_id = thread_a.thread_id
WHERE
    thread_a.processlist_state IS NOT NULL AND
    thread_a.processlist_command != 'Sleep' AND
    thread_a.processlist_id != CONNECTION_ID()
`

type dbmActivity struct {
	Enabled bool `toml:"enabled"`
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

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m *dbmActivityMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Desc: "Collect the waiting event of the current thread",
		Name: "mysql_dbm_activity",
		Type: "logging",
		Fields: map[string]interface{}{
			"query_signature": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The hash value computed from SQL text",
			},
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The text of the normalized SQL text",
			},
			"thread_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The thread ID",
			},
			"processlist_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The process list ID",
			},
			"processlist_user": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The user associated with a thread",
			},
			"processlist_host": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The host name of the client with a thread",
			},
			"processlist_db": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The default database for the thread, or NULL if none has been selected",
			},
			"processlist_command": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The command of the thread",
			},
			"processlist_state": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The state of the thread",
			},
			"sql_text": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The statement the thread is executing",
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
			"event_timer_wait": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationNS,
				Desc:     "The time the event has elapsed so far",
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
				Unit:     inputs.UnknownUnit,
				Desc:     "The default database for the statement, NULL if there is none",
			},
			"wait_event": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The name of the wait event",
			},
			"event_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The event id",
			},
			"end_event_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The thread current event number when the event ends",
			},
			"event_name": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The name of the instrument that produced the event",
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
			"object_schema": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The schema of th object being acted on",
			},
			"object_name": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The name of the object being acted on",
			},
			"index_name": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The name of the index used",
			},
			"object_type": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The type of the object being acted on",
			},
			"event_source": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The name of the source file",
			},
			"ip": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The client IP address",
			},
			"port": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The TCP/IP port number, in the range from 0 to 65535",
			},
			"socket_event_name": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The name of the wait/io/socket/* instrument that produced the event",
			},
			"connections": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total number of the connection",
			},
		},
		Tags: map[string]interface{}{
			"host":    &inputs.TagInfo{Desc: "The server host address"},
			"service": &inputs.TagInfo{Desc: "The service name and the value is 'mysql'"},
			"server":  &inputs.TagInfo{Desc: "The server address"},
		},
	}
}

// get mysql dbm activity.
func (ipt *Input) metricCollectMysqlDbmActivity() ([]*point.Point, error) {
	ms := []inputs.MeasurementV2{}

	// get connections
	connections := getActiveConnections(ipt)
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
		tags := map[string]string{
			"service": "mysql",
			"host":    ipt.Host,
			"status":  "info",
		}
		for key, value := range ipt.Tags {
			tags[key] = value
		}

		message := ""

		if len(activity.ProcesslistState.String) > 0 {
			message += "processlist_state: " + activity.ProcesslistState.String
		}

		if len(activity.SQLText.String) > 0 {
			message += "\nsql_text: " + activity.SQLText.String
		}

		fields := map[string]interface{}{
			"query_signature":     activity.QuerySignature,
			"message":             message,
			"thread_id":           activity.ThreadID.String,
			"processlist_id":      activity.ProcesslistID.String,
			"processlist_user":    activity.ProcesslistUser.String,
			"processlist_host":    activity.ProcesslistHost.String,
			"processlist_db":      activity.ProcesslistDB.String,
			"processlist_command": activity.ProcesslistCommand.String,
			"processlist_state":   activity.ProcesslistState.String,
			"sql_text":            activity.SQLText.String,
			"event_timer_start":   activity.EventTimerStart.Int64 / 1000,
			"event_timer_end":     activity.EventTimerEnd.Int64 / 1000,
			"event_timer_wait":    activity.EventTimerWait.Int64 / 1000,
			"lock_time":           activity.LockTime.Int64 / 1000,
			"current_schema":      activity.CurrentSchema.String,
			"wait_event":          activity.WaitEvent.String,
			"event_id":            activity.EventID.String,
			"end_event_id":        activity.EndEventID.String,
			"event_name":          activity.EventName.String,
			"wait_timer_start":    activity.WaitTimerStart.Int64 / 1000,
			"wait_timer_end":      activity.WaitTimerEnd.Int64 / 1000,
			"object_schema":       activity.ObjectSchema.String,
			"object_name":         activity.ObjectName.String,
			"index_name":          activity.IndexName.String,
			"object_type":         activity.ObjectType.String,
			"event_source":        activity.Source.String,
			"ip":                  activity.IP.String,
			"port":                activity.Port.String,
			"socket_event_name":   activity.SocketEventName.String,
			"connections":         0,
		}
		key := activity.ProcesslistDB.String + activity.ProcesslistHost.String + activity.ProcesslistUser.String + activity.ProcesslistState.String
		if connections, ok := connectionsMap[key]; ok {
			fields["connections"] = connections
		}

		m := &dbmActivityMeasurement{
			name:     "mysql_dbm_activity",
			tags:     tags,
			fields:   fields,
			election: ipt.Election,
		}

		ms = append(ms, m)
	}

	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)

		return pts, nil
	}

	return []*point.Point{}, nil
}

type connectionRow struct {
	processlistUser  sql.NullString
	processlistHost  sql.NullString
	processlistDB    sql.NullString
	processlistState sql.NullString
	connections      sql.NullInt64
}

func getActiveConnections(i *Input) (connectionRows []connectionRow) {
	rows := i.q(connectionsQuerySQL)

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
	QuerySignature     string         `json:"query_signature"`
	ThreadID           sql.NullString `json:"thread_id"`
	ProcesslistID      sql.NullString `json:"processlist_id"`
	ProcesslistUser    sql.NullString `json:"processlist_user"`
	ProcesslistHost    sql.NullString `json:"processlist_host"`
	ProcesslistDB      sql.NullString `json:"processlist_db"`
	ProcesslistCommand sql.NullString `json:"processlist_command"`
	ProcesslistState   sql.NullString `json:"processlist_state"`
	SQLText            sql.NullString `json:"sql_text"`
	EventTimerStart    sql.NullInt64  `json:"event_timer_start"`
	EventTimerEnd      sql.NullInt64  `json:"event_timer_end"`
	EventTimerWait     sql.NullInt64  `json:"event_timer_wait"`
	LockTime           sql.NullInt64  `json:"lock_time"`
	CurrentSchema      sql.NullString `json:"current_schema"`
	WaitEvent          sql.NullString `json:"wait_event"`
	EventID            sql.NullString `json:"event_id"`
	EndEventID         sql.NullString `json:"end_event_id"`
	EventName          sql.NullString `json:"event_name"`
	WaitTimerStart     sql.NullInt64  `json:"wait_timer_start"`
	WaitTimerEnd       sql.NullInt64  `json:"wait_timer_end"`
	ObjectSchema       sql.NullString `json:"object_schema"`
	ObjectName         sql.NullString `json:"object_name"`
	IndexName          sql.NullString `json:"index_name"`
	ObjectType         sql.NullString `json:"object_type"`
	Source             sql.NullString `json:"source"`
	IP                 sql.NullString `json:"ip"`
	Port               sql.NullString `json:"port"`
	SocketEventName    sql.NullString `json:"socket_event_name"`
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

func getActivityRows(i *Input) (activityRows []activityRow) {
	rows := i.q(activityQuerySQL)
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
			&row.EventTimerStart,
			&row.EventTimerEnd,
			&row.EventTimerWait,
			&row.LockTime,
			&row.CurrentSchema,
			&row.WaitEvent,
			&row.EventID,
			&row.EndEventID,
			&row.EventName,
			&row.WaitTimerStart,
			&row.WaitTimerEnd,
			&row.ObjectSchema,
			&row.ObjectName,
			&row.IndexName,
			&row.ObjectType,
			&row.Source,
			&row.IP,
			&row.Port,
			&row.SocketEventName,
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
	size := 0
	normalizedRows := activityRowSlice{}
	for _, row := range rows {
		obfuscatedRow := obfuscateRow(row)

		size += getEstimatedRowSizeBytes(obfuscatedRow)

		if size > maxPayloadBytes {
			return normalizedRows
		}

		normalizedRows = append(normalizedRows, obfuscatedRow)
	}
	return normalizedRows
}

func obfuscateRow(row activityRow) activityRow {
	if row.SQLText.Valid && len(row.SQLText.String) > 0 {
		row.SQLText.String = obfuscateSQL(row.SQLText.String)
		row.QuerySignature = computeSQLSignature(row.SQLText.String)
	}

	return row
}

func getEstimatedRowSizeBytes(row activityRow) int {
	if bytes, err := json.Marshal(row); err != nil {
		return 0
	} else {
		return len(bytes)
	}
}
