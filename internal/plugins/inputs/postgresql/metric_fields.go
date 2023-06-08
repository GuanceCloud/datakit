// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"

	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var relationMetrics = []struct {
	name            string
	query           string
	measurementInfo *inputs.MeasurementInfo
	schemaField     string
}{
	{
		name: "lock metrics",
		query: `
	SELECT mode,
				locktype,
				pn.nspname AS schema,
				pd.datname AS db,
				pc.relname AS table,
				count(*) AS lock_count
	FROM pg_locks l
	JOIN pg_database pd ON (l.database = pd.oid)
	JOIN pg_class pc ON (l.relation = pc.oid)
	LEFT JOIN pg_namespace pn ON (pn.oid = pc.relnamespace)
	WHERE %s 
		AND l.mode IS NOT NULL
		AND pc.relname NOT LIKE 'pg^_%%%%' ESCAPE '^'
	GROUP BY pd.datname, pc.relname, pn.nspname, locktype, mode
		`,
		measurementInfo: lockMeasurement{}.Info(),
		schemaField:     "nspname",
	},
	{
		name: "stat metrics",
		query: `
SELECT relname AS table,schemaname AS schema, *
FROM pg_stat_user_tables 
WHERE %s 
		`,
		schemaField:     "schemaname",
		measurementInfo: statMeasurement{}.Info(),
	},
	{
		name: "index metrics",
		query: `
SELECT relname AS table,
			schemaname AS schema,
			indexrelname AS index,
			idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
WHERE %s 
		`,
		schemaField:     "schemaname",
		measurementInfo: indexMeasurement{}.Info(),
	},
	{
		name: "size metrics",
		query: `
SELECT
	N.nspname AS schema,
	relname AS table,
	pg_table_size(C.oid) as table_size, 
	pg_indexes_size(C.oid) as index_size, 
	pg_total_relation_size(C.oid) as total_size
FROM pg_class C
LEFT JOIN pg_namespace N ON (N.oid = C.relnamespace)
WHERE nspname NOT IN ('pg_catalog', 'information_schema') AND
	nspname !~ '^pg_toast' AND
	relkind = 'r' AND
	%s
		`,
		schemaField:     "nspname",
		measurementInfo: sizeMeasurement{}.Info(),
	},
	{
		name: "statio metrics",
		query: `
SELECT relname AS table,
			schemaname AS schema,
			heap_blks_read, heap_blks_hit, idx_blks_read, idx_blks_hit, toast_blks_read, toast_blks_hit, tidx_blks_read, tidx_blks_hit
FROM pg_statio_user_tables
WHERE %s 
		`,
		schemaField:     "schemaname",
		measurementInfo: statIOMeasurement{}.Info(),
	},
}

type inputMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	ts       time.Time
	election bool
}

func (m inputMeasurement) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(m.name, m.tags, m.fields, dkpt.MOptElectionV2(m.election))
}

// Point implement MeasurementV2.
func (m *inputMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()

	if m.election {
		opts = append(opts, point.WithExtraTags(dkpt.GlobalElectionTags()))
	}

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m inputMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Fields: map[string]interface{}{
			"numbackends":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of active connections to this database."},
			"xact_commit":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of transactions that have been committed in this database."},
			"xact_rollback":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of transactions that have been rolled back in this database."},
			"blks_read":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of disk blocks read in this database."},
			"blks_hit":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of times disk blocks were found in the buffer cache, preventing the need to read from the database."},
			"tup_returned":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of rows returned by queries in this database."},
			"tup_fetched":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of rows fetched by queries in this database."},
			"tup_inserted":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of rows inserted by queries in this database."},
			"tup_updated":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of rows updated by queries in this database."},
			"tup_deleted":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of rows deleted by queries in this database."},
			"deadlocks":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of deadlocks detected in this database."},
			"temp_bytes":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The amount of data written to temporary files by queries in this database."},
			"temp_files":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of temporary files created by queries in this database."},
			"database_size":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The disk space used by this database."},
			"wraparound":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of transactions that can occur until a transaction wraparound."},
			"session_time":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Time spent by database sessions in this database, in milliseconds."},
			"active_time":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Time spent executing SQL statements in this database, in milliseconds."},
			"idle_in_transaction_time": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Time spent idling while in a transaction in this database, in milliseconds."},
			"sessions":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of sessions established to this database."},
			"sessions_abandoned":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of database sessions to this database that were terminated because connection to the client was lost."},
			"sessions_fatal":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of database sessions to this database that were terminated by fatal errors."},
			"sessions_killed":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of database sessions to this database that were terminated by operator intervention."},
		},
		Tags: map[string]interface{}{
			"server": inputs.NewTagInfo("The server address"),
			"db":     inputs.NewTagInfo("The database name"),
		},
	}
}

type lockMeasurement struct {
	inputMeasurement
}

func (m lockMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "postgresql_lock",
		Fields: map[string]interface{}{
			"lock_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of locks active for this database.",
			},
		},
		Tags: map[string]interface{}{
			"db":       inputs.NewTagInfo("The database name"),
			"server":   inputs.NewTagInfo("The server address"),
			"table":    inputs.NewTagInfo("The table name"),
			"schema":   inputs.NewTagInfo("The schema name"),
			"locktype": inputs.NewTagInfo("The lock type"),
			"mode":     inputs.NewTagInfo("The lock mode"),
		},
	}
}

type statMeasurement struct {
	inputMeasurement
}

func (m statMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "postgresql_stat",
		Fields: map[string]interface{}{
			"seq_scan": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of sequential scans initiated on this table.",
			},
			"seq_tup_read": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of live rows fetched by sequential scans.",
			},
			"idx_scan": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of index scans initiated on this table, tagged by index.",
			},
			"idx_tup_fetch": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of live rows fetched by index scans.",
			},
			"n_tup_ins": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of rows inserted by queries in this database.",
			},
			"n_tup_upd": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of rows updated by queries in this database.",
			},
			"n_tup_del": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of rows deleted by queries in this database.",
			},
			"n_tup_hot_upd": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of rows HOT updated, meaning no separate index update was needed.",
			},
			"n_live_tup": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The estimated number of live rows.",
			},
			"n_dead_tup": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The estimated number of dead rows.",
			},
			"vacuum_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of times this table has been manually vacuumed.",
			},
			"autovacuum_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of times this table has been vacuumed by the `autovacuum` daemon.",
			},
			"analyze_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of times this table has been manually analyzed.",
			},
			"autoanalyze_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of times this table has been analyzed by the `autovacuum` daemon.",
			},
		},
		Tags: map[string]interface{}{
			"server": inputs.NewTagInfo("The server address"),
			"db":     inputs.NewTagInfo("The database name"),
			"table":  inputs.NewTagInfo("The table name"),
			"schema": inputs.NewTagInfo("The schema name"),
		},
	}
}

type indexMeasurement struct {
	inputMeasurement
}

func (m indexMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "postgresql_index",
		Fields: map[string]interface{}{
			"idx_scan": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of index scans initiated on this table, tagged by index.",
			},
			"idx_tup_read": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of index entries returned by scans on this index.",
			},
			"idx_tup_fetch": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of live rows fetched by index scans.",
			},
		},
		Tags: map[string]interface{}{
			"table":  inputs.NewTagInfo("The table name"),
			"db":     inputs.NewTagInfo("The database name"),
			"server": inputs.NewTagInfo("The server address"),
			"schema": inputs.NewTagInfo("The schema name"),
			"index":  inputs.NewTagInfo("The index name"),
		},
	}
}

type sizeMeasurement struct {
	inputMeasurement
}

func (m sizeMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "postgresql_size",
		Fields: map[string]interface{}{
			"table_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "The total disk space used by the specified table. Includes TOAST, free space map, and visibility map. Excludes indexes.",
			},
			"index_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "The total disk space used by indexes attached to the specified table.",
			},
			"total_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "The total disk space used by the table, including indexes and TOAST data.",
			},
		},
		Tags: map[string]interface{}{
			"db":     inputs.NewTagInfo("The database name"),
			"server": inputs.NewTagInfo("The server address"),
			"table":  inputs.NewTagInfo("The table name"),
			"schema": inputs.NewTagInfo("The schema name"),
		},
	}
}

type statIOMeasurement struct {
	inputMeasurement
}

func (m statIOMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "postgresql_statio",
		Fields: map[string]interface{}{
			"heap_blks_read": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of disk blocks read from this table.",
			},
			"heap_blks_hit": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of buffer hits in this table.",
			},
			"idx_blks_read": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of disk blocks read from all indexes on this table.",
			},
			"idx_blks_hit": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of buffer hits in all indexes on this table.",
			},
			"toast_blks_read": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of disk blocks read from this table's TOAST table.",
			},
			"toast_blks_hit": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of buffer hits in this table's TOAST table.",
			},
			"tidx_blks_read": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of disk blocks read from this table's TOAST table index.",
			},
			"tidx_blks_hit": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of buffer hits in this table's TOAST table index.",
			},
		},
		Tags: map[string]interface{}{
			"db":     inputs.NewTagInfo("The database name"),
			"server": inputs.NewTagInfo("The server address"),
			"table":  inputs.NewTagInfo("The table name"),
			"schema": inputs.NewTagInfo("The schema name"),
		},
	}
}

type replicationMeasurement struct {
	inputMeasurement
}

func (m replicationMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "postgresql_replication",
		Fields: map[string]interface{}{
			"replication_delay": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				Desc:     "The current replication delay in seconds. Only available with `postgresql` 9.1 and newer.",
			},
			"replication_delay_bytes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "The current replication delay in bytes. Only available with `postgresql` 9.2 and newer.",
			},
		},
		Tags: map[string]interface{}{
			"db":     inputs.NewTagInfo("The database name"),
			"server": inputs.NewTagInfo("The server address"),
		},
	}
}

type slruMeasurement struct {
	inputMeasurement
}

//nolint:lll
func (m slruMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "postgresql_slru",
		Fields: map[string]interface{}{
			"blks_zeroed":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of blocks zeroed during initializations of `SLRU` (simple least-recently-used) cache."},
			"blks_hit":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of times disk blocks were found already in the `SLRU` (simple least-recently-used.)"},
			"blks_read":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of disk blocks read for this `SLRU` (simple least-recently-used) cache. `SLRU` caches are created with a fixed number of pages."},
			"blks_written": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of disk blocks written for this `SLRU` (simple least-recently-used) cache."},
			"blks_exists":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of blocks checked for existence for this `SLRU` (simple least-recently-used) cache."},
			"flushes":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of flush of dirty data for this `SLRU` (simple least-recently-used) cache."},
			"truncates":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of truncates for this `SLRU` (simple least-recently-used) cache."},
		},
		Tags: map[string]interface{}{
			"db":     inputs.NewTagInfo("The database name"),
			"server": inputs.NewTagInfo("The server address"),
			"name":   inputs.NewTagInfo("The name of the `SLRU`"),
		},
	}
}

type bgwriterMeasurement struct {
	inputMeasurement
}

//nolint:lll
func (m bgwriterMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "postgresql_bgwriter",
		Fields: map[string]interface{}{
			"checkpoints_timed":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of scheduled checkpoints that were performed."},
			"checkpoints_req":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of requested checkpoints that were performed."},
			"buffers_checkpoint":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of buffers written during checkpoints."},
			"buffers_clean":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of buffers written by the background writer."},
			"maxwritten_clean":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of times the background writer stopped a cleaning scan due to writing too many buffers."},
			"buffers_backend":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of buffers written directly by a backend."},
			"buffers_alloc":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of buffers allocated"},
			"buffers_backend_fsync": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The of times a backend had to execute its own fsync call instead of the background writer."},
			"checkpoint_write_time": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "The total amount of checkpoint processing time spent writing files to disk."},
			"checkpoint_sync_time":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "The total amount of checkpoint processing time spent synchronizing files to disk."},
		},
		Tags: map[string]interface{}{
			"server": inputs.NewTagInfo("The server address"),
			"db":     inputs.NewTagInfo("The database name"),
		},
	}
}

type connectionMeasurement struct {
	inputMeasurement
}

//nolint:lll
func (m connectionMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "postgresql_connection",
		Fields: map[string]interface{}{
			"max_connections":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The maximum number of client connections allowed to this database."},
			"percent_usage_connections": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of connections to this database as a fraction of the maximum number of allowed connections."},
		},
		Tags: map[string]interface{}{
			"server": inputs.NewTagInfo("The server address"),
			"db":     inputs.NewTagInfo("The database name"),
		},
	}
}

type conflictMeasurement struct {
	inputMeasurement
}

//nolint:lll
func (m conflictMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "postgresql_conflict",
		Fields: map[string]interface{}{
			"confl_tablespace": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of queries in this database that have been canceled due to dropped tablespaces. This will occur when a `temp_tablespace` is dropped while being used on a standby."},
			"confl_lock":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of queries in this database that have been canceled due to dropped tablespaces. This will occur when a `temp_tablespace` is dropped while being used on a standby."},
			"confl_snapshot":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of queries in this database that have been canceled due to old snapshots."},
			"confl_bufferpin":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of queries in this database that have been canceled due to pinned buffers."},
			"confl_deadlock":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of queries in this database that have been canceled due to deadlocks."},
		},
		Tags: map[string]interface{}{
			"server": inputs.NewTagInfo("The server address"),
			"db":     inputs.NewTagInfo("The database name"),
		},
	}
}

type archiverMeasurement struct {
	inputMeasurement
}

//nolint:lll
func (m archiverMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "postgresql_archiver",
		Fields: map[string]interface{}{
			// archiver metric
			"archived_count":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of WAL files that have been successfully archived."},
			"archived_failed_count": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of failed attempts for archiving WAL files."},
		},
		Tags: map[string]interface{}{
			"server": inputs.NewTagInfo("The server address"),
			"db":     inputs.NewTagInfo("The database name"),
		},
	}
}
