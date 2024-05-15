// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cassandra

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	istatsd "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/statsd"
)

type CassandraMeasurement struct{}

// See also https://docs.datadoghq.com/integrations/cassandra/#metrics
// See also https://docs.datadoghq.com/opentelemetry/runtime_metrics/java/
// See also https://docs.datadoghq.com/developers/metrics/types/?tab=count#metric-types

// Info ...
// nolint:lll
func (m *CassandraMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "cassandra",
		Fields: map[string]interface{}{
			"active_tasks":                                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of tasks that the thread pool is actively executing."}, // cassandra metrics
			"bloom_filter_false_ratio":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The ratio of Bloom filter false positives to total checks."},
			"bytes_flushed_count":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of data that was flushed since (re)start."},
			"cas_commit_latency_75th_percentile":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The latency of 'paxos' commit round - p75."},
			"cas_commit_latency_95th_percentile":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The latency of 'paxos' commit round - p95."},
			"cas_commit_latency_one_minute_rate":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec, Desc: "The number of 'paxos' commit round per second."},
			"cas_prepare_latency_75th_percentile":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The latency of 'paxos' prepare round - p75."},
			"cas_prepare_latency_95th_percentile":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The latency of 'paxos' prepare round - p95."},
			"cas_prepare_latency_one_minute_rate":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec, Desc: "The number of 'paxos' prepare round per second."},
			"cas_propose_latency_75th_percentile":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The latency of 'paxos' propose round - p75."},
			"cas_propose_latency_95th_percentile":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The latency of 'paxos' propose round - p95."},
			"cas_propose_latency_one_minute_rate":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec, Desc: "The number of 'paxos' propose round per second."},
			"col_update_time_delta_histogram_75th_percentile": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The column update time delta - p75."},
			"col_update_time_delta_histogram_95th_percentile": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The column update time delta - p95."},
			"col_update_time_delta_histogram_min":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The column update time delta - min."},
			"compaction_bytes_written_count":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of data that was compacted since (re)start."},
			"compression_ratio":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The compression ratio for all SSTables. A low value means a high compression contrary to what the name suggests. Formula used is: 'size of the compressed SSTable / size of original'"},
			"currently_blocked_tasks":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of currently blocked tasks for the thread pool."},
			"currently_blocked_tasks_count":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of currently blocked tasks for the thread pool."},
			"db_droppable_tombstone_ratio":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The estimate of the droppable tombstone ratio."},
			"dropped_one_minute_rate":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The tasks dropped during execution for the thread pool."},
			"exceptions_count":                                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of exceptions thrown from 'Storage' metrics."},
			"key_cache_hit_rate":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The key cache hit rate."},
			"latency_75th_percentile":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The client request latency - p75."},
			"latency_95th_percentile":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The client request latency - p95."},
			"latency_one_minute_rate":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec, Desc: "The number of client requests."},
			"live_disk_space_used_count":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The disk space used by live SSTables (only counts in use files)."},
			"live_ss_table_count":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of live (in use) SSTables."},
			"load_count":                                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The disk space used by live data on a node."},
			"max_partition_size":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The size of the largest compacted partition."},
			"max_row_size":                                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The size of the largest compacted row."},
			"mean_partition_size":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The average size of compacted partition."},
			"mean_row_size":                                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The average size of compacted rows."},
			"net_down_endpoint_count":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of unhealthy nodes in the cluster. They represent each individual node's view of the cluster and thus should not be summed across reporting nodes."},
			"net_up_endpoint_count":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of healthy nodes in the cluster. They represent each individual node's view of the cluster and thus should not be summed across reporting nodes."},
			"pending_compactions":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of pending compactions."},
			"pending_flushes_count":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of pending flushes."},
			"pending_tasks":                                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of pending tasks for the thread pool."},
			"range_latency_75th_percentile":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The local range request latency - p75."},
			"range_latency_95th_percentile":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The local range request latency - p95."},
			"range_latency_one_minute_rate":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec, Desc: "The number of local range requests."},
			"read_latency_75th_percentile":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The local read latency - p75."},
			"read_latency_95th_percentile":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The local read latency - p95."},
			"read_latency_99th_percentile":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The local read latency - p99."},
			"read_latency_one_minute_rate":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec, Desc: "The number of local read requests."},
			"row_cache_hit_out_of_range_count":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of row cache hits that do not satisfy the query filter and went to disk."},
			"row_cache_hit_count":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of row cache hits."},
			"row_cache_miss_count":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of table row cache misses."},
			"snapshots_size":                                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The disk space truly used by snapshots."},
			"ss_tables_per_read_histogram_75th_percentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of SSTable data files accessed per read - p75."},
			"ss_tables_per_read_histogram_95th_percentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of SSTable data files accessed per read - p95."},
			"timeouts_count":                                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Count of requests not acknowledged within configurable timeout window."},
			"timeouts_one_minute_rate":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Recent timeout rate, as an exponentially weighted moving average over a one-minute interval."},
			"tombstone_scanned_histogram_75th_percentile":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of tombstones scanned per read - p75."},
			"tombstone_scanned_histogram_95th_percentile":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of tombstones scanned per read - p95."},
			"total_blocked_tasks":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total blocked tasks"},
			"total_blocked_tasks_count":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total count of blocked tasks"},
			"total_commit_log_size":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The size used on disk by commit logs."},
			"total_disk_space_used_count":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total disk space used by SSTables including obsolete ones waiting to be garbage collected"},
			"view_lock_acquire_time_75th_percentile":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The time taken acquiring a partition lock for materialized view updates - p75."},
			"view_lock_acquire_time_95th_percentile":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The time taken acquiring a partition lock for materialized view updates - p95."},
			"view_lock_acquire_time_one_minute_rate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of requests to acquire a partition lock for materialized view updates."},
			"view_read_time_75th_percentile":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The time taken during the local read of a materialized view update - p75."},
			"view_read_time_95th_percentile":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The time taken during the local read of a materialized view update - p95."},
			"view_read_time_one_minute_rate":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of local reads for materialized view updates."},
			"waiting_on_free_memtable_space_75th_percentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The time spent waiting for free mem table space either on- or off-heap - p75."},
			"waiting_on_free_memtable_space_95th_percentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The time spent waiting for free mem table space either on- or off-heap - p95."},
			"write_latency_75th_percentile":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The local write latency - p75."},
			"write_latency_95th_percentile":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The local write latency - p95."},
			"write_latency_99th_percentile":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The local write latency - p99."},
			"write_latency_one_minute_rate":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec, Desc: "The number of local write requests."},
			"nodetool_status_replication_availability":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Percentage of data available per 'keyspace' times replication factor."},
			"nodetool_status_replication_factor":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Replication factor per 'keyspace'."},
			"nodetool_status_status":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Node status: up (1) or down (0)."},
			"nodetool_status_owns":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Percentage of the data owned by the node per data center times the replication factor."},
			"nodetool_status_load":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Amount of file system data under the 'cassandra' data directory without snapshot content."},
			"metrics_count":                                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Metrics count."},
			"metrics_value":                                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Metrics value."},
			"metrics_95th_percentile":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Metrics - p95."},
			"metrics_75th_percentile":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Metrics - p75."},
			"metrics_one_minute_rate":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of metrics."},
		},
		Tags: map[string]interface{}{
			"host":         inputs.TagInfo{Desc: "Host name."},
			"instance":     inputs.TagInfo{Desc: "Instance name."},
			"jmx_domain":   inputs.TagInfo{Desc: "JMX domain."},
			"metric_type":  inputs.TagInfo{Desc: "Metric type."},
			"name":         inputs.TagInfo{Desc: "Type name."},
			"runtime-id":   inputs.TagInfo{Desc: "Runtime id."},
			"service":      inputs.TagInfo{Desc: "Service name."},
			"type":         inputs.TagInfo{Desc: "Object type."},
			"scope":        inputs.TagInfo{Desc: "scope=ReadStage scope=MutationStage scope=HintsDispatcher scope='MemtableFlushWriter' scope='MemtablePostFlush'"},
			"path":         inputs.TagInfo{Desc: "path=request"},
			"keyspace":     inputs.TagInfo{Desc: "'keyspace'=system 'keyspace'=system_schema "},
			"columnfamily": inputs.TagInfo{Desc: "'columnfamily'=batches 'columnfamily'=built_views 'columnfamily'=columns  'columnfamily'='paxos' 'columnfamily'=peer"},
			"table":        inputs.TagInfo{Desc: "table=IndexInfo,table=available_ranges,table=batches,table=built_views,"},
		},
	}
}

type CassandraJVMMeasurement struct{}

func (m *CassandraJVMMeasurement) Info() *inputs.MeasurementInfo {
	temp := istatsd.JVMMeasurement{}
	info := temp.Info()
	info.Name = "cassandra_jvm"

	return info
}

type CassandraJMXMeasurement struct{}

func (m *CassandraJMXMeasurement) Info() *inputs.MeasurementInfo {
	temp := istatsd.JMXMeasurement{}
	info := temp.Info()
	info.Name = "cassandra_jmx"

	return info
}

type CassandraDDtraceMeasurement struct{}

func (m *CassandraDDtraceMeasurement) Info() *inputs.MeasurementInfo {
	temp := istatsd.DDtraceMeasurement{}
	info := temp.Info()
	info.Name = "cassandra_datadog"

	return info
}
