// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package neo4j

import (
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *Measurement) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(m.name, m.tags, m.fields, dkpt.MOptElection())
}

func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "neo4j",
		Fields: getFields(),
		Tags: map[string]interface{}{
			"host":       inputs.NewTagInfo("Host name."),
			"instance":   inputs.NewTagInfo("Instance endpoint."),
			"quantile":   inputs.NewTagInfo("Histogram `quantile`."),
			"db":         inputs.NewTagInfo("Database name."),
			"database":   inputs.NewTagInfo("Database name."),
			"pool":       inputs.NewTagInfo("Pool name."),
			"bufferpool": inputs.NewTagInfo("Pool name."),
			"gc":         inputs.NewTagInfo("Garbage collection name."),
		},
	}
}

func getFields() map[string]interface{} {
	fields := map[string]interface{}{}
	for _, field := range fieldNames {
		if field.fieldType != "" {
			fields[field.name] = field.field
		}
	}

	return fields
}

type fieldName struct {
	name      string
	prefix    string
	tag       []string // tag:[]string{"foo","bar"} will convert _abc_def_ to tag foo->abc and bar->def
	suffix    string
	fieldType string
	field     interface{}
}

// see also https://neo4j.com/docs/operations-manual/current/monitoring/metrics/essential/
// see also https://neo4j.com/docs/operations-manual/current/monitoring/metrics/reference/
// see also https://neo4j.com/docs/operations-manual/4.4/monitoring/metrics/reference/
// see also https://neo4j.com/docs/operations-manual/3.5/monitoring/metrics/reference/
//
//nolint:lll
var fieldNames []fieldName = []fieldName{
	// Table 1. Bolt metrics
	{name: "dbms_bolt_connections_opened_total", prefix: "dbms_bolt_connections_opened_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of Bolt connections opened since startup. This includes both succeeded and failed connections. Useful for monitoring load via the Bolt drivers in combination with other metrics."}},
	{name: "dbms_bolt_connections_closed_total", prefix: "dbms_bolt_connections_closed_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of Bolt connections closed since startup. This includes both properly and abnormally ended connections. Useful for monitoring load via Bolt drivers in combination with other metrics."}},
	{name: "dbms_bolt_connections_running", prefix: "dbms_bolt_connections_running", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of Bolt connections that are currently executing `Cypher` and returning results. Useful to track the overall load on Bolt connections. This is limited to the number of Bolt worker threads that have been configured via `dbms.connector.bolt.thread_pool_max_size`. Reaching this maximum indicated the server is running at capacity."}},
	{name: "dbms_bolt_connections_idle", prefix: "dbms_bolt_connections_idle", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of Bolt connections that are not currently executing `Cypher` or returning results."}},
	{name: "dbms_bolt_messages_received_total", prefix: "dbms_bolt_messages_received_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of messages received via Bolt since startup. Useful to track general message activity in combination with other metrics."}},
	{name: "dbms_bolt_messages_started_total", prefix: "dbms_bolt_messages_started_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of messages that have started processing since being received. A received message may have begun processing until a Bolt worker thread becomes available. A large gap observed between `bolt.messages_received` and `bolt.messages_started` could indicate the server is running at capacity."}},
	{name: "dbms_bolt_messages_done_total", prefix: "dbms_bolt_messages_done_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of Bolt messages that have completed processing whether successfully or unsuccessfully. Useful for tracking overall load."}},
	{name: "dbms_bolt_messages_failed_total", prefix: "dbms_bolt_messages_failed_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of messages that have failed while processing. A high number of failures may indicate an issue with the server and further investigation of the logs is recommended."}},
	{name: "dbms_bolt_accumulated_queue_time_total", prefix: "dbms_bolt_accumulated_queue_time_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "(unsupported feature) When `internal.server.bolt.thread_pool_queue_size` is enabled,  the total time in milliseconds that a Bolt message waits in the processing queue before a Bolt worker thread becomes available to process it. Sharp increases in this value indicate that the server is running at capacity. If `internal.server.bolt.thread_pool_queue_size` is disabled, the value should be `0`, meaning that messages are directly handed off to worker threads."}},
	{name: "dbms_bolt_accumulated_processing_time_total", prefix: "dbms_bolt_accumulated_processing_time_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "The total amount of time in milliseconds that worker threads have been processing messages. Useful for monitoring load via Bolt drivers in combination with other metrics."}},
	// Table 2. Database checkpointing metrics
	{name: "database_check_point_events_total", prefix: "database_", tag: []string{"db"}, suffix: "_check_point_events_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of checkpoint events executed so far."}},
	{name: "database_check_point_total_time_total", prefix: "database_", tag: []string{"db"}, suffix: "_check_point_total_time_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "The total time, in milliseconds, spent in `checkpointing` so far."}},
	{name: "database_check_point_duration", prefix: "database_", tag: []string{"db"}, suffix: "_check_point_duration", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The duration, in milliseconds, of the last checkpoint event. Checkpoints should generally take several seconds to several minutes. Long checkpoints can be an issue, as these are invoked when the database stops, when a hot backup is taken, and periodically as well. Values over `30` minutes or so should be cause for some investigation."}},
	{name: "database_check_point_flushed_bytes", prefix: "database_", tag: []string{"db"}, suffix: "_check_point_flushed_bytes", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.SizeByte, Unit: inputs.NCount, Desc: "The accumulated number of bytes flushed during the last checkpoint event."}},
	{name: "database_check_point_pages_flushed", prefix: "database_", tag: []string{"db"}, suffix: "_check_point_pages_flushed", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of pages that were flushed during the last checkpoint event."}},
	{name: "database_check_point_io_performed", prefix: "database_", tag: []string{"db"}, suffix: "_check_point_io_performed", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of IOs from Neo4j perspective performed during the last check point event."}},
	{name: "database_check_point_io_limit", prefix: "database_", tag: []string{"db"}, suffix: "_check_point_io_limit", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The IO limit used during the last checkpoint event."}},
	{name: "database_check_point_limit_millis", prefix: "database_", tag: []string{"db"}, suffix: "_check_point_limit_millis", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The time in milliseconds of limit used during the last checkpoint."}},
	{name: "database_check_point_limit_times", prefix: "database_", tag: []string{"db"}, suffix: "_check_point_limit_times", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The times limit used during the last checkpoint."}},
	// Table 3. Cypher metrics
	{name: "database_cypher_replan_events_total", prefix: "database_", tag: []string{"db"}, suffix: "_cypher_replan_events_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of times `Cypher` has decided to re-plan a query. Neo4j caches 1000 plans by default. Seeing sustained replanning events or large spikes could indicate an issue that needs to be investigated."}},
	{name: "database_cypher_replan_wait_time_total", prefix: "database_", tag: []string{"db"}, suffix: "_cypher_replan_wait_time_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationSecond, Desc: "The total number of seconds waited between query replans."}},
	// Table 4. Database data count metrics
	{name: "database_count_relationship", prefix: "database_", tag: []string{"db"}, suffix: "_count_relationship", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of relationships in the database."}},
	{name: "database_count_node", prefix: "database_", tag: []string{"db"}, suffix: "_count_node", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of nodes in the database. A rough metric of how big your graph is. And if you are running a bulk insert operation you can see this tick up."}},
	// Table 5. Database neo4j pools metrics, for db
	// database_system_pool_transaction_system_used_heap -> database_pool_used_heap. tags: db->system, pool->transaction, database->system
	{name: "database_pool_used_heap", prefix: "database_", tag: []string{"db", "", "pool", "database"}, suffix: "_used_heap", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Used or reserved heap memory in bytes."}},
	{name: "database_pool_used_native", prefix: "database_", tag: []string{"db", "", "pool", "database"}, suffix: "_used_native", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Used or reserved native memory in bytes."}},
	{name: "database_pool_total_used", prefix: "database_", tag: []string{"db", "", "pool", "database"}, suffix: "_total_used", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Sum total used heap and native memory in bytes."}},
	{name: "database_pool_total_size", prefix: "database_", tag: []string{"db", "", "pool", "database"}, suffix: "_total_size", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Sum total size of capacity of the heap and/or native memory pool."}},
	{name: "database_pool_free", prefix: "database_", tag: []string{"db", "", "pool", "database"}, suffix: "_free", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Available unused memory in the pool, in bytes."}},
	// Table 5. Database neo4j pools metrics, for dbms
	{name: "dbms_pool_used_heap", prefix: "dbms_pool_", tag: []string{"pool"}, suffix: "_used_heap", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Used or reserved heap memory in bytes."}},
	{name: "dbms_pool_used_native", prefix: "dbms_pool_", tag: []string{"pool"}, suffix: "_used_native", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Used or reserved native memory in bytes."}},
	{name: "dbms_pool_total_used", prefix: "dbms_pool_", tag: []string{"pool"}, suffix: "_total_used", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Sum total used heap and native memory in bytes."}},
	{name: "dbms_pool_total_size", prefix: "dbms_pool_", tag: []string{"pool"}, suffix: "_total_size", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Sum total size of capacity of the heap and/or native memory pool."}},
	{name: "dbms_pool_free", prefix: "dbms_pool_", tag: []string{"pool"}, suffix: "_free", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Available unused memory in the pool, in bytes."}},
	// Table 6. Database operation count metrics
	{name: "db_operation_count_create_total", prefix: "db_operation_count_create_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of successful database create operations."}},
	{name: "db_operation_count_start_total", prefix: "db_operation_count_start_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of successful database start operations."}},
	{name: "db_operation_count_stop_total", prefix: "db_operation_count_stop_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of successful database stop operations."}},
	{name: "db_operation_count_drop_total", prefix: "db_operation_count_drop_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of successful database drop operations."}},
	{name: "db_operation_count_failed_total", prefix: "db_operation_count_failed_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of failed database operations."}},
	{name: "db_operation_count_recovered_total", prefix: "db_operation_count_recovered_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of database operations that failed previously but have recovered."}},
	// Table 7. Database state metrics
	{name: "db_state_count_hosted", prefix: "db_state_count_hosted", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Databases hosted on this server. Databases in states `started`, `store copying`, or `draining` are considered hosted."}},
	{name: "db_state_count_failed", prefix: "db_state_count_failed", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Databases in a failed state on this server."}},
	{name: "db_state_count_desired_started", prefix: "db_state_count_desired_started", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Databases that desire to be started on this server."}},
	// Table 8. Database data metrics
	{name: "database_ids_in_use_relationship_type", prefix: "database_", tag: []string{"db"}, suffix: "_ids_in_use_relationship_type", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of different relationship types stored in the database. Informational, not an indication of any issue. Spikes or large increases indicate large data loads, which could correspond with some behavior you are investigating."}},
	{name: "database_ids_in_use_property", prefix: "database_", tag: []string{"db"}, suffix: "_ids_in_use_property", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of different property names used in the database. Informational, not an indication of any issue. Spikes or large increases indicate large data loads, which could correspond with some behavior you are investigating."}},
	{name: "database_ids_in_use_relationship", prefix: "database_", tag: []string{"db"}, suffix: "_ids_in_use_relationship", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of relationships stored in the database. Informational, not an indication of any issue. Spikes or large increases indicate large data loads, which could correspond with some behavior you are investigating."}},
	{name: "database_ids_in_use_node", prefix: "database_", tag: []string{"db"}, suffix: "_ids_in_use_node", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of nodes stored in the database. Informational, not an indication of any issue. Spikes or large increases indicate large data loads, which could correspond with some behavior you are investigating."}},
	// Table 9. Global neo4j pools metrics
	{name: "dbms_pool_used_heap", prefix: "dbms_pool_", tag: []string{"pool"}, suffix: "_used_heap", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Used or reserved heap memory in bytes."}},
	{name: "dbms_pool_used_native", prefix: "dbms_pool_", tag: []string{"pool"}, suffix: "_used_native", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Used or reserved native memory in bytes."}},
	{name: "dbms_pool_total_used", prefix: "dbms_pool_", tag: []string{"pool"}, suffix: "_total_used", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Sum total used heap and native memory in bytes."}},
	{name: "dbms_pool_total_size", prefix: "dbms_pool_", tag: []string{"pool"}, suffix: "_total_size", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Sum total size of the capacity of the heap and/or native memory pool."}},
	{name: "dbms_pool_free", prefix: "dbms_pool__", tag: []string{"pool"}, suffix: "_free", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Available unused memory in the pool, in bytes."}},
	// Table 10. Database page cache metrics
	{name: "dbms_page_cache_eviction_exceptions_total", prefix: "dbms_page_cache_eviction_exceptions_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of exceptions seen during the eviction process in the page cache."}},
	{name: "dbms_page_cache_flushes_total", prefix: "dbms_page_cache_flushes_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of page flushes executed by the page cache."}},
	{name: "dbms_page_cache_merges_total", prefix: "dbms_page_cache_merges_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of page merges executed by the page cache."}},
	{name: "dbms_page_cache_unpins_total", prefix: "dbms_page_cache_unpins_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of page unpins executed by the page cache."}},
	{name: "dbms_page_cache_pins_total", prefix: "dbms_page_cache_pins_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of page pins executed by the page cache."}},
	{name: "dbms_page_cache_evictions_total", prefix: "dbms_page_cache_evictions_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of page evictions executed by the page cache."}},
	{name: "dbms_page_cache_evictions_cooperative_total", prefix: "dbms_page_cache_evictions_cooperative_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of cooperative page evictions executed by the page cache due to low available pages."}},
	{name: "dbms_page_cache_page_faults_total", prefix: "dbms_page_cache_page_faults_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of page faults happened in the page cache. If this continues to rise over time, it could be an indication that more page cache is needed."}},
	{name: "dbms_page_cache_page_fault_failures_total", prefix: "dbms_page_cache_page_fault_failures_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of failed page faults happened in the page cache."}},
	{name: "dbms_page_cache_page_canceled_faults_total", prefix: "dbms_page_cache_page_canceled_faults_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of canceled page faults happened in the page cache."}},
	{name: "dbms_page_cache_hits_total", prefix: "dbms_page_cache_hits_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of page hits happened in the page cache."}},
	{name: "dbms_page_cache_hit_ratio", prefix: "dbms_page_cache_hit_ratio", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The ratio of hits to the total number of lookups in the page cache. Performance relies on efficiently using the page cache, so this metric should be in the 98-100% range consistently. If it is much lower than that, then the database is going to disk too often."}},
	{name: "dbms_page_cache_usage_ratio", prefix: "dbms_page_cache_usage_ratio", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The ratio of number of used pages to total number of available pages. This metric shows what percentage of the allocated page cache is actually being used. If it is 100%, then it is likely that the hit ratio will start dropping, and you should consider allocating more RAM to page cache."}},
	{name: "dbms_page_cache_bytes_read_total", prefix: "dbms_page_cache_bytes_read_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "The total number of bytes read by the page cache."}},
	{name: "dbms_page_cache_bytes_written_total", prefix: "dbms_page_cache_bytes_written_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "The total number of bytes written by the page cache."}},
	{name: "dbms_page_cache_iops_total", prefix: "dbms_page_cache_iops_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of IO operations performed by page cache."}},
	{name: "dbms_page_cache_throttled_times_total", prefix: "dbms_page_cache_throttled_times_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of times page cache flush IO limiter was throttled during ongoing IO operations."}},
	{name: "dbms_page_cache_throttled_millis_total", prefix: "dbms_page_cache_throttled_millis_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of `millis` page cache flush IO limiter was throttled during ongoing IO operations."}},
	{name: "dbms_page_cache_pages_copied_total", prefix: "dbms_page_cache_pages_copied_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of page copies happened in the page cache."}},
	// Table 11. Query execution metrics
	{name: "database_db_query_execution_success_total", prefix: "database_", tag: []string{"db"}, suffix: "_db_query_execution_success_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of successful queries executed."}},
	{name: "database_db_query_execution_failure_total", prefix: "database_", tag: []string{"db"}, suffix: "_db_query_execution_failure_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of failed queries executed."}},
	{name: "database_db_query_execution_latency_millis", prefix: "database_", tag: []string{"db"}, suffix: "_db_query_execution_latency_millis", fieldType: "summary", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Execution time in milliseconds of queries executed successfully."}},
	{name: "database_db_query_execution_latency_millis_count", prefix: "database_", tag: []string{"db"}, suffix: "_db_query_execution_latency_millis_count"},
	{name: "database_db_query_execution_latency_millis_sum", prefix: "database_", tag: []string{"db"}, suffix: "_db_query_execution_latency_millis_sum"},
	// additional
	{name: "dbms_routing_query_count_local_total", prefix: "dbms_routing_query_count_local_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of queries executed locally."}},
	{name: "dbms_routing_query_count_remote_internal_total", prefix: "dbms_routing_query_count_remote_internal_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of queries executed remotely to a member of the same cluster."}},
	{name: "dbms_routing_query_count_remote_external_total", prefix: "dbms_routing_query_count_remote_external_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of queries executed remotely to a member of a different cluster."}},
	// Table 12. Database store size metrics
	{name: "database_store_size_total", prefix: "database_", tag: []string{"db"}, suffix: "_store_size_total", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total size of the database and transaction logs, in bytes. The total size of the database helps determine how much cache page is required. It also helps compare the total disk space used by the data store and how much is available."}},
	{name: "database_store_size_database", prefix: "database_", tag: []string{"db"}, suffix: "_store_size_database", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The size of the database, in bytes. The total size of the database helps determine how much cache page is required. It also helps compare the total disk space used by the data store and how much is available."}},
	// Table 13. Database transaction log metrics
	{name: "log_rotation_events_total", prefix: "log_rotation_events_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of transaction log rotations executed so far."}},
	{name: "log_rotation_total_time_total", prefix: "log_rotation_total_time_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "The total time, in milliseconds, spent in rotating transaction logs so far."}},
	{name: "log_rotation_duration", prefix: "log_rotation_duration", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The duration, in milliseconds, of the last log rotation event."}},
	{name: "log_appended_bytes_total", prefix: "log_appended_bytes_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "The total number of bytes appended to the transaction log."}},
	{name: "log_flushes_total", prefix: "log_flushes_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of transaction log flushes."}},
	{name: "log_append_batch_size", prefix: "log_append_batch_size", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The size of the last transaction append batch."}},
	// Table 14. Database transaction metrics
	{name: "database_transaction_started_total", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_started_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of started transactions."}},
	{name: "database_transaction_peak_concurrent_total", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_peak_concurrent_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The highest peak of concurrent transactions. This is a useful value to understand. It can help you with the design for the highest load scenarios and whether the Bolt thread settings should be altered."}},
	{name: "database_transaction_active", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_active", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of currently active transactions. Informational, not an indication of any issue. Spikes or large increases could indicate large data loads or just high read load."}},
	{name: "database_transaction_active_read", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_active_read", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of currently active read transactions."}},
	{name: "database_transaction_active_write", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_active_write", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of currently active write transactions."}},
	{name: "database_transaction_committed_total", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_committed_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of committed transactions. Informational, not an indication of any issue. Spikes or large increases indicate large data loads or just high read load."}},
	{name: "database_transaction_committed_read_total", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_committed_read_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of committed read transactions. Informational, not an indication of any issue. Spikes or large increases indicate high read load."}},
	{name: "database_transaction_committed_write_total", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_committed_write_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of committed write transactions. Informational, not an indication of any issue. Spikes or large increases indicate large data loads, which could correspond with some behavior you are investigating."}},
	{name: "database_transaction_rollbacks_total", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_rollbacks_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of rolled back transactions."}},
	{name: "database_transaction_rollbacks_read_total", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_rollbacks_read_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of rolled back read transactions."}},
	{name: "database_transaction_rollbacks_write_total", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_rollbacks_write_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of rolled back write transactions.  Seeing a lot of writes rolled back may indicate various issues with locking, transaction timeouts, etc."}},
	{name: "database_transaction_terminated_total", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_terminated_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of terminated transactions."}},
	{name: "database_transaction_terminated_read_total", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_terminated_read_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of terminated read transactions."}},
	{name: "database_transaction_terminated_write_total", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_terminated_write_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of terminated write transactions."}},
	{name: "database_transaction_last_committed_tx_id_total", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_last_committed_tx_id_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The ID of the last committed transaction. Track this for each instance. (Cluster) Track this for each primary, and each secondary. Might break into separate charts. It should show one line, ever increasing, and if one of the lines levels off or falls behind, it is clear that this member is no longer replicating data, and action is needed to rectify the situation."}},
	{name: "database_transaction_last_closed_tx_id_total", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_last_closed_tx_id_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The ID of the last closed transaction."}},
	{name: "database_transaction_tx_size_heap", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_tx_size_heap", fieldType: "summary", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.SizeByte, Desc: "The transactions' size on heap in bytes."}},
	{name: "database_transaction_tx_size_heap_count", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_tx_size_heap_count"},
	{name: "database_transaction_tx_size_heap_sum", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_tx_size_heap_sum"},
	{name: "database_transaction_tx_size_native", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_tx_size_native", fieldType: "summary", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.SizeByte, Desc: "The transactions' size in native memory in bytes."}},
	{name: "database_transaction_tx_size_native_count", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_tx_size_native_count"},
	{name: "database_transaction_tx_size_native_sum", prefix: "database_", tag: []string{"db"}, suffix: "_transaction_tx_size_native_sum"},
	// Table 15. Database index metrics
	{name: "database_index_fulltext_queried_total", prefix: "database_", tag: []string{"db"}, suffix: "_index_fulltext_queried_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "The total number of times `fulltext` indexes have been queried."}},
	{name: "database_index_fulltext_populated_total", prefix: "database_", tag: []string{"db"}, suffix: "_index_fulltext_populated_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of `fulltext` index population jobs that have been completed."}},
	{name: "database_index_lookup_queried_total", prefix: "database_", tag: []string{"db"}, suffix: "_index_lookup_queried_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of times lookup indexes have been queried."}},
	{name: "database_index_lookup_populated_total", prefix: "database_", tag: []string{"db"}, suffix: "_index_lookup_populated_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of lookup index population jobs that have been completed."}},
	{name: "database_index_text_queried_total", prefix: "database_", tag: []string{"db"}, suffix: "_index_text_queried_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of times text indexes have been queried."}},
	{name: "database_index_text_populated_total", prefix: "database_", tag: []string{"db"}, suffix: "_index_text_populated_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of text index population jobs that have been completed."}},
	{name: "database_index_range_queried_total", prefix: "database_", tag: []string{"db"}, suffix: "_index_range_queried_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of times range indexes have been queried."}},
	{name: "database_index_range_populated_total", prefix: "database_", tag: []string{"db"}, suffix: "_index_range_populated_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of range index population jobs that have been completed."}},
	{name: "database_index_point_queried_total", prefix: "database_", tag: []string{"db"}, suffix: "_index_point_queried_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of times point indexes have been queried."}},
	{name: "database_index_point_populated_total", prefix: "database_", tag: []string{"db"}, suffix: "_index_point_populated_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of point index population jobs that have been completed."}},
	{name: "database_index_vector_queried_total", prefix: "database_", tag: []string{"db"}, suffix: "_index_vector_queried_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of times vector indexes have been queried."}},
	{name: "database_index_vector_populated_total", prefix: "database_", tag: []string{"db"}, suffix: "_index_vector_populated_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of vector index population jobs that have been completed."}},
	// Table 16. Server metrics
	{name: "server_threads_jetty_idle", prefix: "server_threads_jetty_idle", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of idle threads in the jetty pool."}},
	{name: "server_threads_jetty_all", prefix: "server_threads_jetty_all", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of threads (both idle and busy) in the jetty pool."}},
	// Table 17. CatchUp Metrics
	{name: "database_cluster_catchup_tx_pull_requests_received_total", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_catchup_tx_pull_requests_received_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "TX pull requests received from secondaries."}},
	// Table 18. Discovery database primary metrics
	{name: "database_cluster_discovery_replicated_data", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_discovery_replicated_data", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Size of replicated data structures."}},
	{name: "database_cluster_discovery_cluster_members", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_discovery_cluster_members", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Discovery cluster member size."}},
	{name: "database_cluster_discovery_cluster_unreachable", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_discovery_cluster_unreachable", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Discovery cluster unreachable size."}},
	{name: "database_cluster_discovery_cluster_converged", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_discovery_cluster_converged", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Discovery cluster convergence."}},
	// Table 19. Raft database primary metrics
	{name: "database_cluster_raft_append_index", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_append_index", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The append index of the Raft log. Each index represents a write transaction (possibly internal) proposed for commitment. The values mostly increase, but sometimes they can decrease as a consequence of leader changes. The append index should always be less than or equal to the commit index."}},
	{name: "database_cluster_raft_commit_index", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_commit_index", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The commit index of the Raft log. Represents the commitment of previously appended entries. Its value increases monotonically if you do not unbind the cluster state. The commit index should always be bigger than or equal to the appended index."}},
	{name: "database_cluster_raft_applied_index", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_applied_index", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The applied index of the Raft log. Represents the application of the committed Raft log entries to the database and internal state. The applied index should always be bigger than or equal to the commit index. The difference between this and the commit index can be used to monitor how up-to-date the follower database is."}},
	{name: "database_cluster_raft_term", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_term", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The Raft Term of this server. It increases monotonically if you do not unbind the cluster state."}},
	{name: "database_cluster_raft_tx_retries_total", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_tx_retries_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Transaction retries."}},
	{name: "database_cluster_raft_is_leader", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_is_leader", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Is this server the leader? Track this for each database primary in the cluster. It reports `0` if it is not the leader and `1` if it is the leader. The sum of all of these should always be `1`. However, there are transient periods in which the sum can be more than `1` because more than one member thinks it is the leader. Action may be needed if the metric shows `0` for more than 30 seconds."}},
	{name: "database_cluster_raft_in_flight_cache_total_bytes", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_in_flight_cache_total_bytes", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "In-flight cache total bytes."}},
	{name: "database_cluster_raft_in_flight_cache_max_bytes", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_in_flight_cache_max_bytes", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "In-flight cache max bytes."}},
	{name: "database_cluster_raft_in_flight_cache_element_count", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_in_flight_cache_element_count", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "In-flight cache element count."}},
	{name: "database_cluster_raft_in_flight_cache_max_elements", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_in_flight_cache_max_elements", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "In-flight cache maximum elements."}},
	{name: "database_cluster_raft_in_flight_cache_hits_total", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_in_flight_cache_hits_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "In-flight cache hits."}},
	{name: "database_cluster_raft_in_flight_cache_misses_total", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_in_flight_cache_misses_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "In-flight cache misses."}},
	{name: "database_cluster_raft_raft_log_entry_prefetch_buffer_lag", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_raft_log_entry_prefetch_buffer_lag", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Raft Log Entry Prefetch Lag."}},
	{name: "database_cluster_raft_raft_log_entry_prefetch_buffer_bytes", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_raft_log_entry_prefetch_buffer_bytes", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Raft Log Entry Prefetch total bytes."}},
	{name: "database_cluster_raft_raft_log_entry_prefetch_buffer_size", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_raft_log_entry_prefetch_buffer_size", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Raft Log Entry Prefetch buffer size."}},
	{name: "database_cluster_raft_raft_log_entry_prefetch_buffer_async_put", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_raft_log_entry_prefetch_buffer_async_put", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Raft Log Entry Prefetch buffer async puts."}},
	{name: "database_cluster_raft_raft_log_entry_prefetch_buffer_sync_put", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_raft_log_entry_prefetch_buffer_sync_put", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Raft Log Entry Prefetch buffer sync puts."}},
	{name: "database_cluster_raft_message_processing_delay", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_message_processing_delay", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The time the Raft message stays in the queue after being received and before being processed, i.e. append and commit to the store. The higher the value, the longer the messages wait to be processed. This metric is closely linked to disk write performance.(gauge)"}},
	{name: "database_cluster_raft_message_processing_timer", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_message_processing_timer", fieldType: "summary", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.NCount, Desc: "Timer for Raft message processing."}},
	{name: "database_cluster_raft_message_processing_timer_count", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_message_processing_timer_count"},
	{name: "database_cluster_raft_message_processing_timer_sum", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_message_processing_timer_sum"},
	{name: "database_cluster_raft_replication_new_total", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_replication_new_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of Raft replication requests. It increases with write transactions (possibly internal) activity."}},
	{name: "database_cluster_raft_replication_attempt_total", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_replication_attempt_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of Raft replication requests attempts. It is bigger or equal to the replication requests."}},
	{name: "database_cluster_raft_replication_fail_total", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_replication_fail_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of Raft replication attempts that have failed."}},
	{name: "database_cluster_raft_replication_maybe_total", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_replication_maybe_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Raft Replication maybe count."}},
	{name: "database_cluster_raft_replication_success_total", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_replication_success_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of Raft replication requests that have succeeded."}},
	{name: "database_cluster_raft_last_leader_message", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_raft_last_leader_message", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The time elapsed since the last message from a leader in milliseconds. Should reset periodically."}},
	// Table 20. Database secondary Metrics
	{name: "database_cluster_store_copy_pull_updates_total", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_store_copy_pull_updates_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of pull requests made by this instance."}},
	{name: "database_cluster_store_copy_pull_update_highest_tx_id_requested_total", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_store_copy_pull_update_highest_tx_id_requested_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The highest transaction id requested in a pull update by this instance."}},
	{name: "database_cluster_store_copy_pull_update_highest_tx_id_received_total", prefix: "database_", tag: []string{"db"}, suffix: "_cluster_store_copy_pull_update_highest_tx_id_received_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The highest transaction id that has been pulled in the last pull updates by this instance."}},
	// Table 21. JVM file descriptor metrics.
	{name: "dbms_vm_file_descriptors_count", prefix: "dbms_vm_file_descriptors_count", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The current number of open file descriptors."}},
	{name: "dbms_vm_file_descriptors_maximum", prefix: "dbms_vm_file_descriptors_maximum", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "(OS setting) The maximum number of open file descriptors. It is recommended to be set to 40K file handles, because of the native and `Lucene` indexing Neo4j uses. If this metric gets close to the limit, you should consider raising it."}},
	// Table 22. GC metrics. Tag gc have some _
	// dbms_vm_gc_time_g1_old_generation_total->dbms_vm_gc_time_total. tags: gc->g1_old_generation
	{name: "dbms_vm_gc_time_total", prefix: "dbms_vm_gc_time_", tag: []string{"gc"}, suffix: "total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "Accumulated garbage collection time in milliseconds. Long GCs can be an indication of performance issues or potential instability. If this approaches the heartbeat timeout in a cluster, it may cause unwanted leader switches."}},
	{name: "dbms_vm_gc_count_total", prefix: "dbms_vm_gc_count_", tag: []string{"gc"}, suffix: "total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of garbage collections."}},
	// Table 23. JVM Heap metrics.
	{name: "dbms_vm_heap_committed", prefix: "dbms_vm_heap_committed", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Amount of memory (in bytes) guaranteed to be available for use by the JVM."}},
	{name: "dbms_vm_heap_used", prefix: "dbms_vm_heap_used", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Amount of memory (in bytes) currently used. This is the amount of heap space currently used at a given point in time. Monitor this to identify if you are maxing out consistently, in which case, you should increase the initial and max heap size, or if you are `underutilizing`, you should decrease the initial and max heap sizes."}},
	{name: "dbms_vm_heap_max", prefix: "dbms_vm_heap_max", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Maximum amount of heap memory (in bytes) that can be used. This is the amount of heap space currently used at a given point in time. Monitor this to identify if you are maxing out consistently, in which case, you should increase the initial and max heap size, or if you are `underutilizing`, you should decrease the initial and max heap sizes."}},
	// Table 24. JVM memory buffers metrics.
	{name: "dbms_vm_memory_buffer_count", prefix: "dbms_vm_memory_buffer_", tag: []string{"bufferpool"}, suffix: "_count", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Estimated number of buffers in the pool."}},
	{name: "dbms_vm_memory_buffer_used", prefix: "dbms_vm_memory_buffer_", tag: []string{"bufferpool"}, suffix: "_used", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Estimated amount of memory used by the pool."}},
	{name: "dbms_vm_memory_buffer_capacity", prefix: "dbms_vm_memory_buffer_", tag: []string{"bufferpool"}, suffix: "_capacity", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Estimated total capacity of buffers in the pool."}},
	// Table 25. JVM memory pools metrics.
	// dbms_vm_memory_pool_g1_eden_space->dbms_vm_memory_pool. tags: pool->g1_eden_space
	{name: "dbms_vm_memory_pool", prefix: "dbms_vm_memory_pool_", tag: []string{"pool"}, suffix: "", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Estimated amount of memory in bytes used by the pool."}},
	// Table 26. JVM pause time metrics.
	{name: "dbms_vm_pause_time_total", prefix: "dbms_vm_pause_time_total", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "Accumulated detected VM pause time."}},
	// Table 27. JVM threads metrics.
	{name: "dbms_vm_threads", prefix: "dbms_vm_threads", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of live threads including daemon and non-daemon threads."}},

	// v4.4.0
	// see also https://neo4j.com/docs/operations-manual/4.4/monitoring/metrics/reference/

	// Table 1. Bolt metrics
	{name: "bolt_connections_opened_total", prefix: "bolt_connections_opened_total"},
	{name: "bolt_connections_closed_total", prefix: "bolt_connections_closed_total"},
	{name: "bolt_connections_running", prefix: "bolt_connections_running"},
	{name: "bolt_connections_idle", prefix: "bolt_connections_idle"},
	{name: "bolt_messages_received_total", prefix: "bolt_messages_received_total"},
	{name: "bolt_messages_started_total", prefix: "bolt_messages_started_total"},
	{name: "bolt_messages_done_total", prefix: "bolt_messages_done_total"},
	{name: "bolt_messages_failed_total", prefix: "bolt_messages_failed_total"},
	{name: "bolt_accumulated_queue_time_total", prefix: "bolt_accumulated_queue_time_total"},
	{name: "bolt_accumulated_processing_time_total", prefix: "bolt_accumulated_processing_time_total"},
	// Table 2. Database checkpointing metrics
	{name: "check_point_events_total", prefix: "", tag: []string{"db"}, suffix: "_check_point_events_total"},
	{name: "check_point_total_time_total", prefix: "", tag: []string{"db"}, suffix: "_check_point_total_time_total"},
	{name: "check_point_duration", prefix: "", tag: []string{"db"}, suffix: "_check_point_duration"},
	// Table 3. Cypher metrics
	{name: "cypher_replan_events_total", prefix: "", tag: []string{"db"}, suffix: "_cypher_replan_events_total"},
	{name: "cypher_replan_wait_time_total", prefix: "", tag: []string{"db"}, suffix: "_cypher_replan_wait_time_total"},
	// Table 4. Database data count metrics
	{name: "count_relationship", prefix: "", tag: []string{"db"}, suffix: "_count_relationship"},
	{name: "count_node", prefix: "", tag: []string{"db"}, suffix: "_count_node"},
	// Table 5. Database neo4j pools metrics, for db
	// database_system_pool_transaction_system_used_heap -> database_pool_used_heap. tags: db->system, pool->transaction, database->system
	{name: "pool_used_heap", prefix: "", tag: []string{"db", "", "pool", "database"}, suffix: "_used_heap"},
	{name: "pool_used_native", prefix: "", tag: []string{"db", "", "pool", "database"}, suffix: "_used_native"},
	{name: "pool_total_used", prefix: "", tag: []string{"db", "", "pool", "database"}, suffix: "_total_used"},
	{name: "pool_total_size", prefix: "", tag: []string{"db", "", "pool", "database"}, suffix: "_total_size"},
	{name: "pool_free", prefix: "", tag: []string{"db", "", "pool", "database"}, suffix: "_free"},
	// Table 8->7. Database data metrics
	{name: "ids_in_use_relationship_type", prefix: "", tag: []string{"db"}, suffix: "_ids_in_use_relationship_type"},
	{name: "ids_in_use_property", prefix: "", tag: []string{"db"}, suffix: "_ids_in_use_property"},
	{name: "ids_in_use_relationship", prefix: "", tag: []string{"db"}, suffix: "_ids_in_use_relationship"},
	{name: "ids_in_use_node", prefix: "", tag: []string{"db"}, suffix: "_ids_in_use_node"},
	// Table 10->9. Database page cache metrics
	{name: "page_cache_eviction_exceptions_total", prefix: "page_cache_eviction_exceptions_total"},
	{name: "page_cache_flushes_total", prefix: "page_cache_flushes_total"},
	{name: "page_cache_merges_total", prefix: "page_cache_merges_total"},
	{name: "page_cache_unpins_total", prefix: "page_cache_unpins_total"},
	{name: "page_cache_pins_total", prefix: "page_cache_pins_total"},
	{name: "page_cache_evictions_total", prefix: "page_cache_evictions_total"},
	{name: "page_cache_evictions_cooperative_total", prefix: "page_cache_evictions_cooperative_total"},
	{name: "page_cache_page_faults_total", prefix: "page_cache_page_faults_total"},
	{name: "page_cache_page_fault_failures_total", prefix: "page_cache_page_fault_failures_total"},
	{name: "page_cache_page_canceled_faults_total", prefix: "page_cache_page_canceled_faults_total"},
	{name: "page_cache_hits_total", prefix: "page_cache_hits_total"},
	{name: "page_cache_hit_ratio", prefix: "page_cache_hit_ratio"},
	{name: "page_cache_usage_ratio", prefix: "page_cache_usage_ratio"},
	{name: "page_cache_bytes_read_total", prefix: "page_cache_bytes_read_total"},
	{name: "page_cache_bytes_written_total", prefix: "page_cache_bytes_written_total"},
	{name: "page_cache_iops_total", prefix: "page_cache_iops_total"},
	{name: "page_cache_throttled_times_total", prefix: "page_cache_throttled_times_total"},
	{name: "page_cache_throttled_millis_total", prefix: "page_cache_throttled_millis_total"},
	{name: "page_cache_pages_copied_total", prefix: "page_cache_pages_copied_total"},
	// Table 11->10. Query execution metrics
	{name: "db_query_execution_success_total", prefix: "", tag: []string{"db"}, suffix: "_db_query_execution_success_total"},
	{name: "db_query_execution_failure_total", prefix: "", tag: []string{"db"}, suffix: "_db_query_execution_failure_total"},
	{name: "db_query_execution_latency_millis", prefix: "", tag: []string{"db"}, suffix: "_db_query_execution_latency_millis"},
	{name: "db_query_execution_latency_millis_count", prefix: "", tag: []string{"db"}, suffix: "_db_query_execution_latency_millis_count"},
	{name: "db_query_execution_latency_millis_sum", prefix: "", tag: []string{"db"}, suffix: "_db_query_execution_latency_millis_sum"},
	// Table 12->11. Database store size metrics
	{name: "store_size_total", prefix: "", tag: []string{"db"}, suffix: "_store_size_total"},
	{name: "store_size_database", prefix: "", tag: []string{"db"}, suffix: "_store_size_database"},
	// Table 14->13. Database transaction metrics
	{name: "transaction_started_total", prefix: "", tag: []string{"db"}, suffix: "_transaction_started_total"},
	{name: "transaction_peak_concurrent_total", prefix: "", tag: []string{"db"}, suffix: "_transaction_peak_concurrent_total"},
	{name: "transaction_active", prefix: "", tag: []string{"db"}, suffix: "_transaction_active"},
	{name: "transaction_active_read", prefix: "", tag: []string{"db"}, suffix: "_transaction_active_read"},
	{name: "transaction_active_write", prefix: "", tag: []string{"db"}, suffix: "_transaction_active_write"},
	{name: "transaction_committed_total", prefix: "", tag: []string{"db"}, suffix: "_transaction_committed_total"},
	{name: "transaction_committed_read_total", prefix: "", tag: []string{"db"}, suffix: "_transaction_committed_read_total"},
	{name: "transaction_committed_write_total", prefix: "", tag: []string{"db"}, suffix: "_transaction_committed_write_total"},
	{name: "transaction_rollbacks_total", prefix: "", tag: []string{"db"}, suffix: "_transaction_rollbacks_total"},
	{name: "transaction_rollbacks_read_total", prefix: "", tag: []string{"db"}, suffix: "_transaction_rollbacks_read_total"},
	{name: "transaction_rollbacks_write_total", prefix: "", tag: []string{"db"}, suffix: "_transaction_rollbacks_write_total"},
	{name: "transaction_terminated_total", prefix: "", tag: []string{"db"}, suffix: "_transaction_terminated_total"},
	{name: "transaction_terminated_read_total", prefix: "", tag: []string{"db"}, suffix: "_transaction_terminated_read_total"},
	{name: "transaction_terminated_write_total", prefix: "", tag: []string{"db"}, suffix: "_transaction_terminated_write_total"},
	{name: "transaction_last_committed_tx_id_total", prefix: "", tag: []string{"db"}, suffix: "_transaction_last_committed_tx_id_total"},
	{name: "transaction_last_closed_tx_id_total", prefix: "", tag: []string{"db"}, suffix: "_transaction_last_closed_tx_id_total"},
	{name: "transaction_tx_size_heap", prefix: "", tag: []string{"db"}, suffix: "_transaction_tx_size_heap"},
	{name: "transaction_tx_size_heap_count", prefix: "", tag: []string{"db"}, suffix: "_transaction_tx_size_heap_count"},
	{name: "transaction_tx_size_heap_sum", prefix: "", tag: []string{"db"}, suffix: "_transaction_tx_size_heap_sum"},
	{name: "transaction_tx_size_native", prefix: "", tag: []string{"db"}, suffix: "_transaction_tx_size_native"},
	{name: "transaction_tx_size_native_count", prefix: "", tag: []string{"db"}, suffix: "_transaction_tx_size_native_count"},
	{name: "transaction_tx_size_native_sum", prefix: "", tag: []string{"db"}, suffix: "_transaction_tx_size_native_sum"},
	// Table 17->15. CatchUp Metrics
	{name: "cluster_catchup_tx_pull_requests_received_total", prefix: "", tag: []string{"db"}, suffix: "_cluster_catchup_tx_pull_requests_received_total"},
	// Table 18->16. Discovery database primary metrics -> Discovery core metrics
	// database_cluster_ -> causal_clustering_core_
	{name: "causal_clustering_core_discovery_replicated_data", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_discovery_replicated_data"},
	{name: "causal_clustering_core_discovery_cluster_members", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_discovery_cluster_members"},
	{name: "causal_clustering_core_discovery_cluster_unreachable", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_discovery_cluster_unreachable"},
	{name: "causal_clustering_core_discovery_cluster_converged", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_discovery_cluster_converged"},
	// Table 19->17. Raft database primary metrics -> Raft core metrics
	// database_cluster_ -> causal_clustering_core_
	{name: "causal_clustering_core_raft_append_index", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_append_index"},
	{name: "causal_clustering_core_raft_commit_index", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_commit_index"},
	{name: "causal_clustering_core_raft_applied_index", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_applied_index"},
	{name: "causal_clustering_core_raft_term", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_term"},
	{name: "causal_clustering_core_raft_tx_retries_total", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_tx_retries_total"},
	{name: "causal_clustering_core_raft_is_leader", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_is_leader"},
	{name: "causal_clustering_core_raft_in_flight_cache_total_bytes", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_in_flight_cache_total_bytes"},
	{name: "causal_clustering_core_raft_in_flight_cache_max_bytes", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_in_flight_cache_max_bytes"},
	{name: "causal_clustering_core_raft_in_flight_cache_element_count", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_in_flight_cache_element_count"},
	{name: "causal_clustering_core_raft_in_flight_cache_max_elements", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_in_flight_cache_max_elements"},
	{name: "causal_clustering_core_raft_in_flight_cache_hits_total", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_in_flight_cache_hits_total"},
	{name: "causal_clustering_core_raft_in_flight_cache_misses_total", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_in_flight_cache_misses_total"},
	{name: "causal_clustering_core_raft_raft_log_entry_prefetch_buffer_lag", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_raft_log_entry_prefetch_buffer_lag"},
	{name: "causal_clustering_core_raft_raft_log_entry_prefetch_buffer_bytes", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_raft_log_entry_prefetch_buffer_bytes"},
	{name: "causal_clustering_core_raft_raft_log_entry_prefetch_buffer_size", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_raft_log_entry_prefetch_buffer_size"},
	{name: "causal_clustering_core_raft_raft_log_entry_prefetch_buffer_async_put", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_raft_log_entry_prefetch_buffer_async_put"},
	{name: "causal_clustering_core_raft_raft_log_entry_prefetch_buffer_sync_put", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_raft_log_entry_prefetch_buffer_sync_put"},
	{name: "causal_clustering_core_raft_message_processing_delay", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_message_processing_delay"},
	{name: "causal_clustering_core_raft_message_processing_timer", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_message_processing_timer"},
	{name: "causal_clustering_core_raft_message_processing_timer_count", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_message_processing_timer_count"},
	{name: "causal_clustering_core_raft_message_processing_timer_sum", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_message_processing_timer_sum"},
	{name: "causal_clustering_core_raft_replication_new_total", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_replication_new_total"},
	{name: "causal_clustering_core_raft_replication_attempt_total", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_replication_attempt_total"},
	{name: "causal_clustering_core_raft_replication_fail_total", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_replication_fail_total"},
	{name: "causal_clustering_core_raft_replication_maybe_total", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_replication_maybe_total"},
	{name: "causal_clustering_core_raft_replication_success_total", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_replication_success_total"},
	{name: "causal_clustering_core_raft_last_leader_message", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_core_raft_last_leader_message"},
	// Table 20->18. Database secondary Metrics -> Read Replica Metrics
	// database_cluster_store_copy_ -> causal_clustering_read_replica_
	{name: "causal_clustering_read_replica_pull_updates_total", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_read_replica_pull_updates_total"},
	{name: "causal_clustering_read_replica_pull_update_highest_tx_id_requested_total", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_read_replica_pull_update_highest_tx_id_requested_total"},
	{name: "causal_clustering_read_replica_pull_update_highest_tx_id_received_total", prefix: "", tag: []string{"db"}, suffix: "_causal_clustering_read_replica_pull_update_highest_tx_id_received_total"},
	// Table 21->19. JVM file descriptor metrics
	{name: "vm_file_descriptors_count", prefix: "vm_file_descriptors_count"},
	{name: "vm_file_descriptors_maximum", prefix: "vm_file_descriptors_maximum"},
	// Table 22->20. GC metrics. Tag gc have some _
	// dbms_vm_gc_time_g1_old_generation_total->dbms_vm_gc_time_total. tags: gc->g1_old_generation
	{name: "vm_gc_time_total", prefix: "vm_gc_time_", tag: []string{"gc"}, suffix: "total"},
	{name: "vm_gc_count_total", prefix: "vm_gc_count_", tag: []string{"gc"}, suffix: "total"},
	// Table 23->21. JVM Heap metrics.
	{name: "vm_heap_committed", prefix: "vm_heap_committed"},
	{name: "vm_heap_used", prefix: "vm_heap_used"},
	{name: "vm_heap_max", prefix: "vm_heap_max"},
	// Table 24->22. JVM memory buffers metrics.
	{name: "vm_memory_buffer_count", prefix: "vm_memory_buffer_", tag: []string{"bufferpool"}, suffix: "_count"},
	{name: "vm_memory_buffer_used", prefix: "vm_memory_buffer_", tag: []string{"bufferpool"}, suffix: "_used"},
	{name: "vm_memory_buffer_capacity", prefix: "vm_memory_buffer_", tag: []string{"bufferpool"}, suffix: "_capacity"},
	// Table 25->23. JVM memory pools metrics.
	// dbms_vm_memory_pool_g1_eden_space->dbms_vm_memory_pool. tags: pool->g1_eden_space
	{name: "vm_memory_pool", prefix: "vm_memory_pool_", tag: []string{"pool"}, suffix: ""},
	// Table 26->24. JVM pause time metrics.
	{name: "vm_pause_time", prefix: "vm_pause_time"},
	// Table 27. JVM threads metrics. (have new field name)
	{name: "vm_thread_count", prefix: "vm_thread_count", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "(only for neo4j.v4) Estimated number of active threads in the current thread group."}},
	{name: "vm_thread_total", prefix: "vm_thread_total", fieldType: "gauge", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "(only for neo4j.v4) The total number of live threads including daemon and non-daemon threads."}},

	// v3.4.0.
	// see also https://neo4j.com/docs/operations-manual/3.5/monitoring/metrics/reference/

	// Table 1->7. Bolt metrics. (have new field  name)
	{name: "bolt_sessions_started", prefix: "bolt_sessions_started", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "(only for neo4j.v3) The total number of Bolt sessions started since this instance started."}},
	{name: "bolt_connections_opened", prefix: "bolt_connections_opened"},
	{name: "bolt_connections_closed", prefix: "bolt_connections_closed"},
	{name: "bolt_connections_running", prefix: "bolt_connections_running"},
	{name: "bolt_connections_idle", prefix: "bolt_connections_idle"},
	{name: "bolt_messages_received", prefix: "bolt_messages_received"},
	{name: "bolt_messages_started", prefix: "bolt_messages_started"},
	{name: "bolt_messages_done", prefix: "bolt_messages_done"},
	{name: "bolt_messages_failed", prefix: "bolt_messages_failed"},
	{name: "bolt_accumulated_queue_time", prefix: "bolt_accumulated_queue_time"},
	{name: "bolt_accumulated_processing_time", prefix: "bolt_accumulated_processing_time"},
	// Table 2->1. Database checkpointing metrics
	{name: "check_point_events", prefix: "check_point_events"},
	{name: "check_point_total_time", prefix: "check_point_total_time"},
	{name: "check_point_duration", prefix: "check_point_duration"},
	// Table 3->5. Cypher metrics
	{name: "cypher_replan_events", prefix: "cypher_replan_events"},
	{name: "cypher_replan_wait_time", prefix: "cypher_replan_wait_time"},
	// Table 8->2. Database data metrics
	{name: "ids_in_use_relationship_type", prefix: "ids_in_use_relationship_type"},
	{name: "ids_in_use_property", prefix: "ids_in_use_property"},
	{name: "ids_in_use_relationship", prefix: "ids_in_use_relationship"},
	{name: "ids_in_use_node", prefix: "ids_in_use_node"},
	// Table 10->3. Database page cache metrics
	{name: "page_cache_eviction_exceptions", prefix: "page_cache_eviction_exceptions"},
	{name: "page_cache_flushes", prefix: "page_cache_flushes"},
	{name: "page_cache_unpins", prefix: "page_cache_unpins"},
	{name: "page_cache_pins", prefix: "page_cache_pins"},
	{name: "page_cache_evictions", prefix: "page_cache_evictions"},
	{name: "page_cache_page_faults", prefix: "page_cache_page_faults"},
	{name: "page_cache_hits", prefix: "page_cache_hits"},
	{name: "page_cache_hit_ratio", prefix: "page_cache_hit_ratio"},
	{name: "page_cache_usage_ratio", prefix: "page_cache_usage_ratio"},
	// Table 13->6. Database transaction log metrics
	{name: "log_rotation_events", prefix: "log_rotation_events"},
	{name: "log_rotation_total_time", prefix: "log_rotation_total_time"},
	{name: "log_rotation_duration", prefix: "log_rotation_duration"},
	// Table 14->4. Database transaction metrics
	{name: "transaction_started", prefix: "transaction_started"},
	{name: "transaction_peak_concurrent", prefix: "transaction_peak_concurrent"},
	{name: "transaction_active", prefix: "transaction_active"},
	{name: "transaction_active_read", prefix: "transaction_active_read"},
	{name: "transaction_active_write", prefix: "transaction_active_write"},
	{name: "transaction_committed", prefix: "transaction_committed"},
	{name: "transaction_committed_read", prefix: "transaction_committed_read"},
	{name: "transaction_committed_write", prefix: "transaction_committed_write"},
	{name: "transaction_rollbacks", prefix: "transaction_rollbacks"},
	{name: "transaction_rollbacks_read", prefix: "transaction_rollbacks_read"},
	{name: "transaction_rollbacks_write", prefix: "transaction_rollbacks_write"},
	{name: "transaction_terminated", prefix: "transaction_terminated"},
	{name: "transaction_terminated_read", prefix: "transaction_terminated_read"},
	{name: "transaction_terminated_write", prefix: "transaction_terminated_write"},
	{name: "transaction_last_committed_tx_id", prefix: "transaction_last_committed_tx_id"},
	{name: "transaction_last_closed_tx_id", prefix: "transaction_last_closed_tx_id"},
	// Table 16->8. Server metrics. Same as v5.11.0
	{name: "server_threads_jetty_idle", prefix: "server_threads_jetty_idle"},
	{name: "server_threads_jetty_all", prefix: "server_threads_jetty_all"},
	// Table 19->9. Raft database primary metrics -> Core metrics
	{name: "causal_clustering_core_raft_append_index", prefix: "causal_clustering_core_raft_append_index"},
	{name: "causal_clustering_core_raft_commit_index", prefix: "causal_clustering_core_raft_commit_index"},
	{name: "causal_clustering_core_raft_term", prefix: "causal_clustering_core_raft_term"},
	{name: "causal_clustering_core_raft_tx_retries", prefix: "causal_clustering_core_raft_tx_retries"},
	{name: "causal_clustering_core_raft_is_leader", prefix: "causal_clustering_core_raft_is_leader"},
	{name: "causal_clustering_core_raft_in_flight_cache_total_bytes", prefix: "causal_clustering_core_raft_in_flight_cache_total_bytes"},
	{name: "causal_clustering_core_raft_in_flight_cache_max_bytes", prefix: "causal_clustering_core_raft_in_flight_cache_max_bytes"},
	{name: "causal_clustering_core_raft_in_flight_cache_element_count", prefix: "causal_clustering_core_raft_in_flight_cache_element_count"},
	{name: "causal_clustering_core_raft_in_flight_cache_max_elements", prefix: "causal_clustering_core_raft_in_flight_cache_max_elements"},
	{name: "causal_clustering_core_raft_in_flight_cache_hits", prefix: "causal_clustering_core_raft_in_flight_cache_hits"},
	{name: "causal_clustering_core_raft_in_flight_cache_misses", prefix: "causal_clustering_core_raft_in_flight_cache_misses"},
	{name: "causal_clustering_core_raft_message_processing_delay", prefix: "causal_clustering_core_raft_message_processing_delay"},
	{name: "causal_clustering_core_raft_message_processing_timer", prefix: "causal_clustering_core_raft_message_processing_timer"},
	{name: "causal_clustering_core_raft_replication_new", prefix: "causal_clustering_core_raft_replication_new"},
	{name: "causal_clustering_core_raft_replication_attempt", prefix: "causal_clustering_core_raft_replication_attempt"},
	{name: "causal_clustering_core_raft_replication_fail", prefix: "causal_clustering_core_raft_replication_fail"},
	{name: "causal_clustering_core_raft_replication_success", prefix: "causal_clustering_core_raft_replication_success"},
	// Table 20->10. Database secondary Metrics -> Read Replica Metrics
	{name: "causal_clustering_read_replica_pull_updates", prefix: "causal_clustering_read_replica_pull_updates"},
	{name: "causal_clustering_read_replica_pull_update_highest_tx_id_requested", prefix: "causal_clustering_read_replica_pull_update_highest_tx_id_requested"},
	{name: "causal_clustering_read_replica_pull_update_highest_tx_id_received", prefix: "causal_clustering_read_replica_pull_update_highest_tx_id_received"},
	// Table 22->v3.4 useful
	{name: "vm_gc_time", prefix: "vm_gc_time_", tag: []string{"gc"}, suffix: ""},
	{name: "vm_gc_count", prefix: "vm_gc_count_", tag: []string{"gc"}, suffix: ""},
	// Table v3.4 new table. Network
	{name: "network_master_network_store_writes", prefix: "network_master_network_store_writes", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "(only for neo4j.v3) The master network store writes."}},
	{name: "network_master_network_tx_writes", prefix: "network_master_network_tx_writes", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "(only for neo4j.v3) The master network transaction writes."}},
	{name: "network_slave_network_tx_writes", prefix: "network_slave_network_tx_writes", fieldType: "counter", field: &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "(only for neo4j.v3)  The slave network transaction writes."}},
}
