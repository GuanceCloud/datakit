//nolint:lll
package elasticsearch

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

//
var elasticsearchMeasurementFields = map[string]interface{}{
	"active_shards_percent_as_number": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "active shards percent"},
	"active_primary_shards":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "active primary shards"},
	"status":                          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "status"},
	"timed_out":                       &inputs.FieldInfo{DataType: inputs.Bool, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "timed_out"},
}

// nodeStats.
var nodeStatsTags = map[string]interface{}{
	"cluster_name":                     inputs.NewTagInfo("Name of the cluster, based on the Cluster name setting setting."),
	"node_attribute_ml.enabled":        inputs.NewTagInfo("Set to true (default) to enable machine learning APIs on the node."),
	"node_attribute_ml.machine_memory": inputs.NewTagInfo("The machine’s memory that machine learning may use for running analytics processes."),
	"node_attribute_ml.max_open_jobs":  inputs.NewTagInfo("The maximum number of jobs that can run simultaneously on a node."),
	"node_attribute_xpack.installed":   inputs.NewTagInfo("Show whether xpack is installed."),
	"node_host":                        inputs.NewTagInfo("Network host for the node, based on the network.host setting."),
	"node_id":                          inputs.NewTagInfo("The id for the node."),
	"node_name":                        inputs.NewTagInfo("Human-readable identifier for the node."),
}

var nodeStatsFields = map[string]interface{}{
	"transport_rx_size_in_bytes":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Size of RX packets received by the node during internal cluster communication."},
	"transport_tx_size_in_bytes":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Size of TX packets sent by the node during internal cluster communication."},
	"http_current_open":                                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of open HTTP connections for the node."},
	"indices_fielddata_evictions":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of evictions from the field data cache across all shards assigned to selected nodes."},
	"indices_fielddata_memory_size_in_bytes":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total amount, in bytes, of memory used for the field data cache across all shards assigned to selected nodes."},
	"indices_get_missing_total":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of failed get operations."},
	"indices_get_missing_time_in_millis":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time in milliseconds spent performing failed get operations."},
	"jvm_gc_collectors_old_collection_count":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of JVM garbage collectors that collect old generation objects."},
	"jvm_gc_collectors_old_collection_time_in_millis":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time in milliseconds spent by JVM collecting old generation objects."},
	"jvm_gc_collectors_young_collection_count":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of JVM garbage collectors that collect young generation objects."},
	"jvm_gc_collectors_young_collection_time_in_millis": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time in milliseconds spent by JVM collecting young generation objects."},
	"jvm_mem_heap_committed_in_bytes":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Amount of memory, in bytes, available for use by the heap."},
	"jvm_mem_heap_used_percent":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Percentage of memory currently in use by the heap."},
	"os_cpu_load_average_15m":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Fifteen-minute load average on the system (field is not present if fifteen-minute load average is not available)."},
	"os_cpu_load_average_1m":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "One-minute load average on the system (field is not present if one-minute load average is not available)."},
	"os_cpu_load_average_5m":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: " Five-minute load average on the system (field is not present if five-minute load average is not available)."},
	"os_cpu_percent":                                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Recent CPU usage for the whole system, or -1 if not supported."},
	"os_mem_total_in_bytes":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total amount of physical memory in bytes."},
	"os_mem_used_in_bytes":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Amount of used physical memory in bytes."},
	"os_mem_used_percent":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Percentage of used memory."},
	"process_open_file_descriptors":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of opened file descriptors associated with the current or -1 if not supported."},
	"thread_pool_search_queue":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of tasks in queue for the thread pool"},
	"thread_pool_rollup_indexing_queue":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of tasks in queue for the thread pool"},
	"thread_pool_force_merge_queue":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of tasks in queue for the thread pool"},
	"thread_pool_transform_indexing_queue":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of tasks in queue for the thread pool"},
	"thread_pool_rollup_indexing_rejected":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of tasks rejected by the thread pool executor."},
	"thread_pool_force_merge_rejected":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of tasks rejected by the thread pool executor."},
	"thread_pool_transform_indexing_rejected":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of tasks rejected by the thread pool executor."},
	"thread_pool_search_rejected":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of tasks rejected by the thread pool executor."},
	"fs_data_0_available_in_gigabytes":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total number of gigabytes available to this Java virtual machine on this file store."},
	"fs_data_0_free_in_gigabytes":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total number of unallocated gigabytes in the file store."},
	"fs_data_0_total_in_gigabytes":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total size (in gigabytes) of the file store."},
	"fs_io_stats_devices_0_operations":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of read and write operations for the device completed since starting Elasticsearch."},
	"fs_io_stats_devices_0_read_kilobytes":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of kilobytes read for the device since starting Elasticsearch."},
	"fs_io_stats_devices_0_read_operations":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of read operations for the device completed since starting Elasticsearch."},
	"fs_io_stats_devices_0_write_kilobytes":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of kilobytes written for the device since starting Elasticsearch."},
	"fs_io_stats_devices_0_write_operations":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of write operations for the device completed since starting Elasticsearch."},
	"fs_io_stats_total_operations":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of read and write operations across all devices used by Elasticsearch completed since starting Elasticsearch."},
	"fs_io_stats_total_read_kilobytes":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of kilobytes read across all devices used by Elasticsearch since starting Elasticsearch."},
	"fs_io_stats_total_read_operations":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of read operations for across all devices used by Elasticsearch completed since starting Elasticsearch."},
	"fs_io_stats_total_write_kilobytes":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of kilobytes written across all devices used by Elasticsearch since starting Elasticsearch."},
	"fs_io_stats_total_write_operations":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of write operations across all devices used by Elasticsearch completed since starting Elasticsearch."},
	"fs_timestamp":                                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.TimestampMS, Desc: "Last time the file stores statistics were refreshed. Recorded in milliseconds since the Unix Epoch."},
	"fs_total_available_in_gigabytes":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total number of gigabytes available to this Java virtual machine on all file stores."},
	"fs_total_free_in_gigabytes":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total number of unallocated gigabytes in all file stores."},
	"fs_total_total_in_gigabytes":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total size (in gigabytes) of all file stores."},
}

// clusterStats.
var clusterStatsTags = map[string]interface{}{
	"cluster_name": inputs.NewTagInfo("Name of the cluster, based on the cluster.name setting."),
	"node_name":    inputs.NewTagInfo("Name of the node."),
	"status":       inputs.NewTagInfo("Health status of the cluster, based on the state of its primary and replica shards."),
}

var clusterStatsFields = map[string]interface{}{
	"nodes_process_open_file_descriptors_avg": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Average number of concurrently open file descriptors. Returns -1 if not supported."},
}

// clusterHealth.
var clusterHealthTags = map[string]interface{}{
	"name": inputs.NewTagInfo("Name of the cluster."),
}

var clusterHealthFields = map[string]interface{}{
	"active_primary_shards":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of active primary shards in the cluster."},
	"active_shards":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of active shards in the cluster."},
	"initializing_shards":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of shards that are currently initializing."},
	"number_of_data_nodes":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of data nodes in the cluster."},
	"number_of_pending_tasks":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of pending tasks."},
	"relocating_shards":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of shards that are relocating from one node to another."},
	"status":                        &inputs.FieldInfo{DataType: inputs.String, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "The cluster status: red, yellow, green."},
	"unassigned_shards":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of shards that are unassigned to a node."},
	"indices_lifecycle_error_count": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of indices that are managed by ILM and are in an error state."},
}

// clusterHealthIndices.
var clusterHealthIndicesTags = map[string]interface{}{
	"name":  inputs.NewTagInfo("Name of the cluster."),
	"index": inputs.NewTagInfo("Name of the index."),
}

var clusterHealthIndicesFields = map[string]interface{}{
	"active_primary_shards": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of active primary shards in the cluster."},
	"active_shards":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of active shards in the cluster."},
	"initializing_shards":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of shards that are currently initializing."},
	"number_of_replicas":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of replica in the target index."},
	"number_of_shards":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of shards in the target index."},
	"relocating_shards":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of shards that are relocating from one node to another."},
	"status":                &inputs.FieldInfo{DataType: inputs.String, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "The status: red, yellow, green."},
	"status_code":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The health as a number: red = 0, yellow = 1, green = 2."},
	"unassigned_shards":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of shards that are unassigned to a node."},
}

// indicesStatsShardsTotal
// NOTE: no tags.
var indicesStatsShardsTotalFields = map[string]interface{}{
	"failed":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of failed indexing operations"},
	"successful": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of successful operations"},
	"total":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of total operations"},
}

// indicesStats.
var indicesStatsTags = map[string]interface{}{
	"cluster_name": inputs.NewTagInfo("Name of the cluster, based on the Cluster name setting setting."),
	"index_name":   inputs.NewTagInfo("Name of the index. The name '_all' target all data streams and indices in a cluster."),
}

var indicesStatsFields = map[string]interface{}{
	"total_indexing_index_current":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of indexing operations currently running."},
	"total_get_missing_total":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of failed get operations."},
	"total_indexing_index_time_in_millis": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time in milliseconds spent performing indexing operations."},
	"total_indexing_index_total":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of indexing operations."},
	"total_merges_total":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of merge operations."},
	"total_merges_total_docs":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of merged documents."},
	"total_merges_current_docs":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of document merges currently running."},
	"total_merges_total_time_in_millis":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time in milliseconds spent performing merge operations."},
	"total_flush_total":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of flush operations."},
	"total_flush_total_time_in_millis":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time in milliseconds spent performing flush operations."},
	"total_refresh_total":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of refresh operations."},
	"total_refresh_total_time_in_millis":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time in milliseconds spent performing refresh operations."},
	"total_search_fetch_current":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of fetch operations currently running."},
	"total_search_fetch_time_in_millis":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time in milliseconds spent performing fetch operations."},
	"total_search_fetch_total":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of fetch operations."},
	"total_search_query_current":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of query operations currently running."},
	"total_search_query_time_in_millis":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time in milliseconds spent performing query operations."},
	"total_search_query_total":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of query operations."},
	"total_store_size_in_bytes":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total size, in bytes, of all shards assigned to selected nodes."},
}

// indicesStatsShards.
var indicesStatsShardsTags = map[string]interface{}{
	"index_name": inputs.NewTagInfo("Name of the index."),
	"node_name":  inputs.NewTagInfo("Name of the node."),
	"shard_name": inputs.NewTagInfo("Name of the shard."),
	"type":       inputs.NewTagInfo("Type of the shard."),
}

var indicesStatsShardsFields = map[string]interface{}{
	"commit_generation":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "The commit generation"},
	"commit_num_docs":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of docs"},
	"completion_size_in_bytes":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total bytes used for completion."},
	"docs_count":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of non-deleted documents."},
	"docs_deleted":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of deleted documents."},
	"fielddata_evictions":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of evictions from the field data cache."},
	"fielddata_memory_size_in_bytes":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total amount, in bytes, of memory used for the field data cache."},
	"flush_periodic":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of flush periodic operations."},
	"flush_total":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of flush operations."},
	"flush_total_time_in_millis":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time in milliseconds spent performing flush operations."},
	"get_current":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of get operations currently running."},
	"get_exists_time_in_millis":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time in milliseconds spent performing successful get operations."},
	"get_exists_total":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of successful get operations."},
	"get_missing_time_in_millis":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time in milliseconds spent performing failed get operations."},
	"get_missing_total":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of failed get operations."},
	"get_time_in_millis":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time in milliseconds spent performing get operations."},
	"get_total":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of get operations."},
	"indexing_delete_current":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of deletion operations currently running."},
	"indexing_delete_time_in_millis":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time in milliseconds spent performing deletion operations."},
	"indexing_delete_total":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of deletion operations."},
	"indexing_index_current":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of indexing operations currently running."},
	"indexing_index_failed":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of failed indexing operations."},
	"indexing_index_time_in_millis":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time in milliseconds spent performing indexing operations."},
	"indexing_index_total":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of indexing operations."},
	"indexing_is_throttled":                  &inputs.FieldInfo{DataType: inputs.Bool, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of times operations were throttled."},
	"indexing_noop_update_total":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of noop operations."},
	"indexing_throttle_time_in_millis":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time in milliseconds spent throttling operations."},
	"merges_current":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of merge operations currently running."},
	"merges_current_docs":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of document merges currently running."},
	"merges_current_size_in_bytes":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Memory, in bytes, used performing current document merges."},
	"merges_total":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of merge operations."},
	"merges_total_auto_throttle_in_bytes":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Size, in bytes, of automatically throttled merge operations."},
	"merges_total_docs":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of merged documents."},
	"merges_total_size_in_bytes":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total size of document merges in bytes."},
	"merges_total_stopped_time_in_millis":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time in milliseconds spent stopping merge operations."},
	"merges_total_throttled_time_in_millis":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time in milliseconds spent throttling merge operations."},
	"merges_total_time_in_millis":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time in milliseconds spent performing merge operations."},
	"query_cache_cache_count":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Count of queries in the query cache."},
	"query_cache_cache_size":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Size, in bytes, of the query cache."},
	"query_cache_evictions":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of query cache evictions."},
	"query_cache_hit_count":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of query cache hits."},
	"query_cache_memory_size_in_bytes":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total amount of memory, in bytes, used for the query cache across all shards assigned to the node."},
	"query_cache_miss_count":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of query cache misses."},
	"query_cache_total_count":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total count of hits, misses, and cached queries in the query cache."},
	"recovery_current_as_source":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of recoveries that used an index shard as a source."},
	"recovery_current_as_target":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of recoveries that used an index shard as a target."},
	"recovery_throttle_time_in_millis":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time in milliseconds recovery operations were delayed due to throttling."},
	"refresh_external_total":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of external refresh operations."},
	"refresh_external_total_time_in_millis":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time in milliseconds spent performing external operations."},
	"refresh_listeners":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of refresh listeners."},
	"refresh_total":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of refresh operations."},
	"refresh_total_time_in_millis":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time in milliseconds spent performing refresh operations."},
	"request_cache_evictions":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of request cache operations."},
	"request_cache_hit_count":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of request cache hits."},
	"request_cache_memory_size_in_bytes":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Memory, in bytes, used by the request cache."},
	"request_cache_miss_count":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of request cache misses."},
	"retention_leases_primary_term":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "retention_leases_primary_term"},
	"retention_leases_version":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "retention_leases_version"},
	"routing_state":                          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "routing_state"},
	"search_fetch_current":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of fetch operations currently running."},
	"search_fetch_time_in_millis":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time in milliseconds spent performing fetch operations."},
	"search_fetch_total":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of fetch operations."},
	"search_open_contexts":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of open search contexts."},
	"search_query_current":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of query operations currently running."},
	"search_query_time_in_millis":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time in milliseconds spent performing query operations."},
	"search_query_total":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of query operations."},
	"search_scroll_current":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of scroll operations currently running."},
	"search_scroll_time_in_millis":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time in milliseconds spent performing scroll operations."},
	"search_scroll_total":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of scroll operations."},
	"search_suggest_current":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of suggest operations currently running."},
	"search_suggest_time_in_millis":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Time in milliseconds spent performing suggest operations."},
	"search_suggest_total":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of suggest operations."},
	"segments_count":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of segments."},
	"segments_doc_values_memory_in_bytes":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total amount of memory, in bytes, used for doc values across all shards assigned to the node."},
	"segments_fixed_bit_set_memory_in_bytes": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total amount of memory, in bytes, used by fixed bit sets across all shards assigned to the node."},
	"segments_index_writer_memory_in_bytes":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total amount of memory, in bytes, used by all index writers across all shards assigned to the node."},
	"segments_max_unsafe_auto_id_timestamp":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: " Unix timestamp, in milliseconds, of the most recently retried indexing request."},
	"segments_memory_in_bytes":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total amount, in bytes, of memory used for segments across all shards assigned to selected nodes."},
	"segments_norms_memory_in_bytes":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total amount, in bytes, of memory used for normalization factors across all shards assigned to selected nodes."},
	"segments_points_memory_in_bytes":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total amount, in bytes, of memory used for points across all shards assigned to selected nodes."},
	"segments_stored_fields_memory_in_bytes": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total amount, in bytes, of memory used for stored fields across all shards assigned to selected nodes."},
	"segments_term_vectors_memory_in_bytes":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "term_vectors_memory_in_bytes"},
	"segments_terms_memory_in_bytes":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total amount, in bytes, of memory used for terms across all shards assigned to selected nodes."},
	"segments_version_map_memory_in_bytes":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total amount, in bytes, of memory used by all version maps across all shards assigned to selected nodes."},
	"seq_no_global_checkpoint":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "seq_no_global_checkpoint"},
	"seq_no_local_checkpoint":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "seq_no_local_checkpoint"},
	"seq_no_max_seq_no":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "seq_no_max_seq_no"},
	"shard_path_is_custom_data_path":         &inputs.FieldInfo{DataType: inputs.Bool, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "shard_path_is_custom_data_path"},
	"store_size_in_bytes":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total size, in bytes, of all shards assigned to selected nodes."},
	"translog_earliest_last_modified_age":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Earliest last modified age for the transaction log."},
	"translog_operations":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of operations in the transaction log."},
	"translog_size_in_bytes":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The size of the transaction log."},
	"translog_uncommitted_operations":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of uncommitted operations in the transaction log."},
	"translog_uncommitted_size_in_bytes":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total amount, in bytes, of uncommitted operations in the transaction log."},
	"warmer_current":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of active index warmers."},
	"warmer_total":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of index warmers."},
	"warmer_total_time_in_millis":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total time in milliseconds spent performing index warming operations."},
}
