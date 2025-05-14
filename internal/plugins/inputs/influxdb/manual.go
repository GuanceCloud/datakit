// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package influxdb

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	influxdbCq            = "influxdb_cq"
	influxdbHttpd         = "influxdb_httpd"
	influxdbMemstats      = "influxdb_memstats"
	influxdbQueryExecutor = "influxdb_queryExecutor"
	influxdbRuntime       = "influxdb_runtime"
	influxdbSubscriber    = "influxdb_subscriber"
	influxdbWrite         = "influxdb_write"
)

////////////////////////////////////////////////////////////////////////////////

type measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     int64
}

// Point implement MeasurementV2.
func (m *measurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m *measurement) Info() *inputs.MeasurementInfo { return nil }

////////////////////////////////////////////////////////////////////////////////

type InfluxdbCqM measurement

// Point implement MeasurementV2.
func (m *InfluxdbCqM) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *InfluxdbCqM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNamePrefix + "cq",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "System hostname."},
		},
		Fields: map[string]interface{}{
			"query_fail": nFIFloatUnknown("The total number of continuous queries that executed but failed."),
			"query_ok":   nFIFloatUnknown("The total number of continuous queries that executed successfully."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

type InfluxdbHttpdM measurement

// Point implement MeasurementV2.
func (m *InfluxdbHttpdM) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *InfluxdbHttpdM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNamePrefix + "httpd",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "System hostname."},
			"bind": &inputs.TagInfo{Desc: "Bind port."},
		},
		Fields: map[string]interface{}{
			"auth_fail":                  nFIFloatUnknown("The number of HTTP requests that were aborted due to authentication being required, but not supplied or incorrect."),
			"client_error":               nFIFloatUnknown("The number of HTTP responses due to client errors, with a 4XX HTTP status code."),
			"flux_query_req":             nFIFloatUnknown("The number of Flux query requests served."),
			"flux_query_req_duration_ns": nFIFloatDurationNs("The duration (wall-time), in nanoseconds, spent executing Flux query requests."),
			"ping_req":                   nFIFloatUnknown("The number of times InfluxDB HTTP server served the /ping HTTP endpoint."),
			"points_written_dropped":     nFIFloatUnknown("The number of points dropped by the storage engine."),
			"points_written_fail":        nFIFloatUnknown("The number of points accepted by the HTTP /write endpoint, but unable to be persisted."),
			"points_written_ok":          nFIFloatUnknown("The number of points successfully accepted and persisted by the HTTP /write endpoint."),
			"prom_read_req":              nFIFloatUnknown("The number of read requests to the Prometheus /read endpoint."),
			"prom_write_req":             nFIFloatUnknown("The number of write requests to the Prometheus /write endpoint."),
			"query_req":                  nFIFloatUnknown("The number of query requests."),
			"query_req_duration_ns":      nFIFloatDurationNs("The total query request duration, in nanosecond (ns)."),
			"query_resp_bytes":           nFIFloatBytes("The total number of bytes returned in query responses."),
			"recovered_panics":           nFIFloatUnknown("The total number of panics recovered by the HTTP handler."),
			"req":                        nFIFloatUnknown("The total number of HTTP requests served."),
			"req_active":                 nFIFloatUnknown("The number of currently active requests."),
			"req_duration_ns":            nFIFloatDurationNs("The duration (wall time), in nanoseconds, spent inside HTTP requests."),
			"server_error":               nFIFloatUnknown("The number of HTTP responses due to server errors."),
			"status_req":                 nFIFloatUnknown("The number of status requests served using the HTTP /status endpoint."),
			"values_written_ok":          nFIFloatUnknown("The number of values (fields) successfully accepted and persisted by the HTTP /write endpoint."),
			"write_req":                  nFIFloatUnknown("The number of write requests served using the HTTP /write endpoint."),
			"write_req_active":           nFIFloatUnknown("The number of currently active write requests."),
			"write_req_bytes":            nFIFloatBytes("The total number of bytes of line protocol data received by write requests, using the HTTP /write endpoint."),
			"write_req_duration_ns":      nFIFloatDurationNs("The duration (wall time), in nanoseconds, of write requests served using the /write HTTP endpoint."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

type InfluxdbMemstatsM measurement

// Point implement MeasurementV2.
func (m *InfluxdbMemstatsM) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *InfluxdbMemstatsM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNamePrefix + "memstats",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "System hostname."},
		},
		Fields: map[string]interface{}{
			"alloc":           nFIFloatBytes("The currently allocated number of bytes of heap objects."),
			"total_alloc":     nFIFloatBytes("The cumulative bytes allocated for heap objects."),
			"sys":             nFIFloatBytes("The cumulative bytes allocated for heap objects."),
			"lookups":         nFIFloatUnknown("The number of pointer lookups performed by the runtime."),
			"mallocs":         nFIFloatUnknown("The total number of heap objects allocated."),
			"frees":           nFIFloatUnknown("The cumulative number of freed (live) heap objects."),
			"heap_alloc":      nFIFloatBytes("The size, in bytes, of all heap objects."),
			"heap_sys":        nFIFloatBytes("The number of bytes of heap memory obtained from the OS."),
			"heap_idle":       nFIFloatBytes("The number of bytes of idle heap objects."),
			"heap_inuse":      nFIFloatBytes("The number of bytes in in-use spans."),
			"heap_released":   nFIFloatBytes("The number of bytes of physical memory returned to the OS."),
			"heap_objects":    nFIFloatUnknown("The number of allocated heap objects."),
			"stack_inuse":     nFIFloatBytes("The number of bytes in in-use stacks."),
			"stack_sys":       nFIFloatBytes("The total number of bytes of memory obtained from the stack in use."),
			"mspan_inuse":     nFIFloatBytes("The bytes of allocated mcache structures."),
			"mspan_sys":       nFIFloatBytes("The bytes of memory obtained from the OS for `mspan`."),
			"mcache_inuse":    nFIFloatBytes("The bytes of allocated mcache structures."),
			"mcache_sys":      nFIFloatBytes("The bytes of memory obtained from the OS for mcache structures."),
			"buck_hash_sys":   nFIFloatBytes("The bytes of memory in profiling bucket hash tables."),
			"gc_sys":          nFIFloatBytes("The bytes of memory in garbage collection metadata."),
			"other_sys":       nFIFloatBytes("The number of bytes of memory used other than `heap_sys/stacks_sys/mspan_sys/mcache_sys/buckhash_sys/gc_sys`."),
			"next_gc":         nFIFloatUnknown("The target heap size of the next garbage collection cycle."),
			"last_gc":         nFIFloatTimestampNs("Time the last garbage collection finished, as nanoseconds since 1970 (the UNIX epoch)."),
			"pause_total_ns":  nFIFloatDurationNs("The total time garbage collection cycles are paused in nanoseconds."),
			"pause_ns":        nFIFloatDurationNs("The time garbage collection cycles are paused in nanoseconds."),
			"num_gc":          nFIFloatUnknown("The number of completed garbage collection cycles."),
			"num_forced_gc":   nFIFloatUnknown("The number of GC cycles that were forced by the application calling the GC function."),
			"gc_cpu_fraction": nFIFloatUnknown("The fraction of CPU time used by the garbage collection cycle."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

type InfluxdbQueryExecutorM measurement

// Point implement MeasurementV2.
func (m *InfluxdbQueryExecutorM) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *InfluxdbQueryExecutorM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNamePrefix + "queryExecutor",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "System hostname."},
		},
		Fields: map[string]interface{}{
			"queries_active":    nFIFloatUnknown("The number of active queries currently being handled."),
			"queries_executed":  nFIFloatUnknown("The number of queries executed (started)."),
			"queries_finished":  nFIFloatUnknown("The number of queries that have finished executing."),
			"query_duration_ns": nFIFloatDurationNs("The duration (wall time), in nanoseconds, of every query executed. "),
			"recovered_panics":  nFIFloatUnknown("The number of panics recovered by the Query Executor."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

type InfluxdbRuntimeM measurement

// Point implement MeasurementV2.
func (m *InfluxdbRuntimeM) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *InfluxdbRuntimeM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNamePrefix + "runtime",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "System hostname."},
		},
		Fields: map[string]interface{}{
			"alloc":          nFIFloatBytes("The currently allocated number of bytes of heap objects."),
			"frees":          nFIFloatUnknown("The cumulative number of freed (live) heap objects."),
			"heap_alloc":     nFIFloatBytes("The size, in bytes, of all heap objects."),
			"heap_idle":      nFIFloatBytes("The number of bytes of idle heap objects."),
			"heap_inuse":     nFIFloatBytes("The number of bytes in in-use spans."),
			"heap_objects":   nFIFloatUnknown("The number of allocated heap objects."),
			"heap_released":  nFIFloatBytes("The number of bytes of physical memory returned to the OS."),
			"heap_sys":       nFIFloatBytes("The number of bytes of heap memory obtained from the OS."),
			"lookups":        nFIFloatUnknown("The number of pointer lookups performed by the runtime."),
			"mallocs":        nFIFloatUnknown("The total number of heap objects allocated."),
			"num_gc":         nFIFloatUnknown("The number of completed garbage collection cycles."),
			"num_goroutine":  nFIFloatUnknown("The total number of Go routines."),
			"pause_total_ns": nFIFloatDurationNs("The total time garbage collection cycles are paused in nanoseconds."),
			"sys":            nFIFloatBytes("The cumulative bytes allocated for heap objects."),
			"total_alloc":    nFIFloatBytes("The cumulative bytes allocated for heap objects."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

type InfluxdbSubscriberM measurement

// Point implement MeasurementV2.
func (m *InfluxdbSubscriberM) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *InfluxdbSubscriberM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNamePrefix + "subscriber",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "System hostname."},
		},
		Fields: map[string]interface{}{
			"create_failures": nFIFloatUnknown("The number of subscriptions that failed to be created."),
			"points_written":  nFIFloatUnknown("The total number of points that were successfully written to subscribers."),
			"write_failures":  nFIFloatUnknown("The total number of batches that failed to be written to subscribers."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

type InfluxdbWriteM measurement

// Point implement MeasurementV2.
func (m *InfluxdbWriteM) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *InfluxdbWriteM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNamePrefix + "write",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "System hostname."},
		},
		Fields: map[string]interface{}{
			"point_req":       nFIFloatUnknown("The total number of every point requested to be written to this data node."),
			"point_req_local": nFIFloatUnknown("The total number of point requests that have been attempted to be written into a shard on the same (local) node."),
			"req":             nFIFloatUnknown("The total number of batches of points requested to be written to this node."),
			"sub_write_drop":  nFIFloatUnknown("The total number of batches of points that failed to be sent to the subscription dispatcher."),
			"sub_write_ok":    nFIFloatUnknown("The total number of batches of points that were successfully sent to the subscription dispatcher."),
			"write_drop":      nFIFloatUnknown("The total number of write requests for points that have been dropped due to timestamps not matching any existing retention policies."),
			"write_error":     nFIFloatUnknown("The total number of batches of points that were not successfully written, due to a failure to write to a local or remote shard."),
			"write_ok":        nFIFloatUnknown("The total number of batches of points written at the requested consistency level."),
			"write_timeout":   nFIFloatUnknown("The total number of write requests that failed to complete within the default write timeout duration."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

type InfluxdbDatabaseM measurement

// Point implement MeasurementV2.
func (m *InfluxdbDatabaseM) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *InfluxdbDatabaseM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNamePrefix + "database",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "System hostname."},
			"database": &inputs.TagInfo{Desc: "Database name."},
		},
		Fields: map[string]interface{}{
			"num_measurements": nFIFloatUnknown("The current number of measurements in the specified database."),
			"num_series":       nFIFloatUnknown("The current series cardinality of the specified database. "),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

type InfluxdbShardM measurement

// Point implement MeasurementV2.
func (m *InfluxdbShardM) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *InfluxdbShardM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNamePrefix + "shard",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"host":             &inputs.TagInfo{Desc: "System hostname."},
			"database":         &inputs.TagInfo{Desc: "Database name."},
			"engine":           &inputs.TagInfo{Desc: "Engine."},
			"id":               &inputs.TagInfo{Desc: "ID."},
			"index_type":       &inputs.TagInfo{Desc: "Index type."},
			"path":             &inputs.TagInfo{Desc: "Path."},
			"retention_policy": &inputs.TagInfo{Desc: "Retention policy."},
			"wal_path":         &inputs.TagInfo{Desc: "Wal path."},
		},
		Fields: map[string]interface{}{
			"disk_bytes":           nFIFloatBytes("The size, in bytes, of the shard, including the size of the data directory and the WAL directory."),
			"fields_create":        nFIFloatUnknown("The number of fields created."),
			"series_create":        nFIFloatUnknown("Then number of series created."),
			"write_bytes":          nFIFloatBytes("The number of bytes written to the shard."),
			"write_points_dropped": nFIFloatUnknown("The number of requests to write points t dropped from a write."),
			"write_points_err":     nFIFloatUnknown("The number of requests to write points that failed to be written due to errors."),
			"write_points_ok":      nFIFloatUnknown("The number of points written successfully."),
			"write_req":            nFIFloatUnknown("The total number of write requests."),
			"write_req_err":        nFIFloatUnknown("The total number of write requests that failed due to errors."),
			"write_req_ok":         nFIFloatUnknown("The total number of successful write requests."),
			"write_values_ok":      nFIFloatUnknown("The number of write values successfully."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

type InfluxdbTsm1EngineM measurement

// Point implement MeasurementV2.
func (m *InfluxdbTsm1EngineM) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *InfluxdbTsm1EngineM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNamePrefix + "tsm1_engine",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"host":             &inputs.TagInfo{Desc: "System hostname."},
			"database":         &inputs.TagInfo{Desc: "Database name."},
			"engine":           &inputs.TagInfo{Desc: "Engine."},
			"id":               &inputs.TagInfo{Desc: "ID."},
			"index_type":       &inputs.TagInfo{Desc: "Index type."},
			"path":             &inputs.TagInfo{Desc: "Path."},
			"retention_policy": &inputs.TagInfo{Desc: "Retention policy."},
			"wal_path":         &inputs.TagInfo{Desc: "Wal path."},
		},
		Fields: map[string]interface{}{
			"cache_compaction_duration": nFIFloatDurationNs("The duration (wall time), in nanoseconds, spent in cache compactions."),
			"cache_compaction_err":      nFIFloatUnknown("The number of cache compactions that have failed due to errors."),
			"cache_compactions":         nFIFloatUnknown("The total number of cache compactions that have ever run."),
			"cache_compactions_active":  nFIFloatUnknown("The number of cache compactions that are currently running."),

			"tsm_full_compaction_duration": nFIFloatUnknown("The duration (wall time), in nanoseconds, spent in full compactions."),
			"tsm_full_compaction_err":      nFIFloatUnknown("The total number of TSM full compactions that have failed due to errors."),
			"tsm_full_compaction_queue":    nFIFloatUnknown("The current number of pending TMS Full compactions."),
			"tsm_full_compactions":         nFIFloatUnknown("The total number of TSM full compactions that have ever run."),
			"tsm_full_compactions_active":  nFIFloatUnknown("The number of TSM full compactions currently running."),

			"tsm_level1_compaction_duration": nFIFloatDurationNs("The duration (wall time), in nanoseconds, spent in TSM level 1 compactions."),
			"tsm_level1_compaction_err":      nFIFloatUnknown("The total number of TSM level 1 compactions that have failed due to errors."),
			"tsm_level1_compaction_queue":    nFIFloatUnknown("The current number of pending TSM level 1 compactions."),
			"tsm_level1_compactions":         nFIFloatUnknown("The total number of TSM level 1 compactions that have ever run."),
			"tsm_level1_compactions_active":  nFIFloatUnknown("The number of TSM level 1 compactions that are currently running."),

			"tsm_level2_compaction_duration": nFIFloatDurationNs("The duration (wall time), in nanoseconds, spent in TSM level 2 compactions."),
			"tsm_level2_compaction_err":      nFIFloatUnknown("The number of TSM level 2 compactions that have failed due to errors."),
			"tsm_level2_compaction_queue":    nFIFloatUnknown("The current number of pending TSM level 2 compactions."),
			"tsm_level2_compactions":         nFIFloatUnknown("The total number of TSM level 2 compactions that have ever run."),
			"tsm_level2_compactions_active":  nFIFloatUnknown("The number of TSM level 2 compactions that are currently running."),

			"tsm_level3_compaction_duration": nFIFloatDurationNs("The duration (wall time), in nanoseconds, spent in TSM level 3 compactions."),
			"tsm_level3_compaction_err":      nFIFloatUnknown("The number of TSM level 3 compactions that have failed due to errors."),
			"tsm_level3_compaction_queue":    nFIFloatUnknown("The current number of pending TSM level 3 compactions."),
			"tsm_level3_compactions":         nFIFloatUnknown("The total number of TSM level 3 compactions that have ever run."),
			"tsm_level3_compactions_active":  nFIFloatUnknown("The number of TSM level 3 compactions that are currently running."),

			"tsm_optimize_compaction_duration": nFIFloatDurationNs("The duration (wall time), in nanoseconds, spent during TSM optimize compactions."),
			"tsm_optimize_compaction_err":      nFIFloatUnknown("The total number of TSM optimize compactions that have failed due to errors."),
			"tsm_optimize_compaction_queue":    nFIFloatUnknown("The current number of pending TSM optimize compactions."),
			"tsm_optimize_compactions":         nFIFloatUnknown("The total number of TSM optimize compactions that have ever run."),
			"tsm_optimize_compactions_active":  nFIFloatUnknown("The number of TSM optimize compactions that are currently running."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

type InfluxdbTsm1CacheM measurement

// Point implement MeasurementV2.
func (m *InfluxdbTsm1CacheM) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *InfluxdbTsm1CacheM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNamePrefix + "tsm1_cache",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"host":             &inputs.TagInfo{Desc: "System hostname."},
			"database":         &inputs.TagInfo{Desc: "Database name."},
			"engine":           &inputs.TagInfo{Desc: "Engine."},
			"id":               &inputs.TagInfo{Desc: "ID."},
			"index_type":       &inputs.TagInfo{Desc: "Index type."},
			"path":             &inputs.TagInfo{Desc: "Path."},
			"retention_policy": &inputs.TagInfo{Desc: "Retention policy."},
			"wal_path":         &inputs.TagInfo{Desc: "Wal path."},
		},
		Fields: map[string]interface{}{
			"wal_compaction_time_ms": nFIFloatDurationMs("The duration, in milliseconds, that the commit lock is held while compacting snapshots."),
			"cache_age_ms":           nFIFloatDurationMs("The duration, in milliseconds, since the cache was last snapshotted at sample time."),
			"cached_bytes":           nFIFloatBytes("The total number of bytes that have been written into snapshots."),
			"disk_bytes":             nFIFloatBytes("The size, in bytes, of on-disk snapshots."),
			"mem_bytes":              nFIFloatBytes("The size, in bytes, of in-memory cache."),
			"snapshot_count":         nFIFloatUnknown("The current level (number) of active snapshots."),
			"write_dropped":          nFIFloatUnknown("The total number of writes dropped due to timeouts."),
			"write_err":              nFIFloatUnknown("The total number of writes that failed."),
			"write_ok":               nFIFloatUnknown("The total number of successful writes."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

type InfluxdbTsm1FilestoreM measurement

// Point implement MeasurementV2.
func (m *InfluxdbTsm1FilestoreM) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *InfluxdbTsm1FilestoreM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNamePrefix + "tsm1_filestore",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"host":             &inputs.TagInfo{Desc: "System hostname."},
			"database":         &inputs.TagInfo{Desc: "Database name."},
			"engine":           &inputs.TagInfo{Desc: "Engine."},
			"id":               &inputs.TagInfo{Desc: "ID."},
			"index_type":       &inputs.TagInfo{Desc: "Index type."},
			"path":             &inputs.TagInfo{Desc: "Path."},
			"retention_policy": &inputs.TagInfo{Desc: "Retention policy."},
			"wal_path":         &inputs.TagInfo{Desc: "Wal path."},
		},
		Fields: map[string]interface{}{
			"disk_bytes": nFIFloatBytes("The size, in bytes, of disk usage by the TSM file store."),
			"num_files":  nFIFloatUnknown("The total number of files in the TSM file store."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

type InfluxdbTsm1WalM measurement

// Point implement MeasurementV2.
func (m *InfluxdbTsm1WalM) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *InfluxdbTsm1WalM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNamePrefix + "tsm1_wal",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"host":             &inputs.TagInfo{Desc: "System hostname."},
			"database":         &inputs.TagInfo{Desc: "Database name."},
			"engine":           &inputs.TagInfo{Desc: "Engine."},
			"id":               &inputs.TagInfo{Desc: "ID."},
			"index_type":       &inputs.TagInfo{Desc: "Index type."},
			"path":             &inputs.TagInfo{Desc: "Path."},
			"retention_policy": &inputs.TagInfo{Desc: "Retention policy."},
			"wal_path":         &inputs.TagInfo{Desc: "Wal path."},
		},
		Fields: map[string]interface{}{
			"current_segment_disk_bytes": nFIFloatBytes("The current size, in bytes, of the segment disk."),
			"old_segments_disk_bytes":    nFIFloatBytes("The size, in bytes, of the segment disk."),
			"write_err":                  nFIFloatUnknown("The number of writes that failed due to errors."),
			"write_ok":                   nFIFloatUnknown("The number of writes that succeeded."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

func nFIFloatUnknown(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.UnknownUnit,
		Desc:     desc,
	}
}

func nFIFloatBytes(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     desc,
	}
}

func nFIFloatDurationNs(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationNS,
		Desc:     desc,
	}
}

func nFIFloatDurationMs(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationMS,
		Desc:     desc,
	}
}

func nFIFloatTimestampNs(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.TimestampNS,
		Desc:     desc,
	}
}
