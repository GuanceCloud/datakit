// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package doris

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

// See also https://doris.apache.org/docs/admin-manual/maint-monitor/monitor-metrics/metrics

type feMeasurement struct{}

func (*feMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "doris_fe",
		Cat:  point.Metric,
		//nolint:lll
		Fields: map[string]interface{}{
			"cache_added":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative value of the number."},
			"cache_hit":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count of cache hits."},
			"connection_total":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of FE MySQL port connections."},
			"counter_hit_sql_block_rule":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of queries blocked by SQL BLOCK RULE."},
			"edit_log_clean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of times the historical metadata log was cleared."},
			"edit_log":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Value of metadata log."},
			"editlog_write_latency_ms":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "metadata log write latency . For example, {quantile=0.75} indicates the 75th percentile write latency ."},
			"image_clean":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of times cleaning of historical metadata image files."},
			"image_push":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of times cleaning of historical metadata image files."},
			"image_write":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The Number of to generate metadata image files."},
			"job":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current count of different job types and different job statuses. For example, {job=load, type=INSERT, state=LOADING} represents an import job of type INSERT and the number of jobs in the LOADING state."},
			"max_journal_id":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The maximum metadata log ID of the current FE node . If it is Master FE , it is the maximum ID currently written , if it is a non- Master FE , represents the maximum ID of the metadata log currently being played back."},
			"max_tablet_compaction_score": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The largest compaction score value among all BE nodes."},
			"qps":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec, Desc: "Current number of FE queries per second ( only query requests are counted )."},
			"query_err":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Value of error query."},
			"query_err_rate":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec, Desc: "Error queries per second."},
			"query_latency_ms":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Percentile statistics of query request latency. For example, {quantile=0.75} indicates the query delay at the 75th percentile."},
			"query_latency_ms_db":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Percentile statistics of query request delay of each DB . For example, {quantile=0.75,db=test} indicates the query delay of the 75th percentile of DB test."},
			"query_olap_table":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The statistics of the number of requests for the internal table ( `OlapTable` )."},
			"query_total":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "All query requests."},
			"report_queue_size":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The queue length of various periodic reporting tasks of BE on the FE side."},
			"request_total":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "All operation requests received through the MySQL port (including queries and other statements )."},
			"routine_load_error_rows":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count the total number of error rows for all Routine Load jobs in the cluster."},
			"routine_load_receive_bytes":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "The amount of data received by all Routine Load jobs in the cluster."},
			"routine_load_rows":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Count the number of data rows received by all Routine Load jobs in the cluster."},
			"rps":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of FE requests per second (including queries and other types of statements )."},
			"scheduled_tablet_num":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Tablets being scheduled by the Master FE node . Includes replicas being repaired and replicas being balanced."},
			"tablet_max_compaction_score": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The compaction core reported by each BE node . For example, { backend=172.21.0.1:9556} represents the reported value of BE 172.21.0.1:9556."},
			"tablet_num":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current total number of tablets on each BE node . For example, {backend=172.21.0.1:9556} indicates the current number of tablets of the BE 172.21.0.1:9556."},
			"tablet_status_count":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Statistics Master FE node The cumulative value of the number of tablets scheduled by the tablet scheduler."},
			"thread_pool":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Count the number of working threads and queuing status of various thread pools . active_thread_num Indicates the number of tasks being executed . pool_size Indicates the total number of threads in the thread pool . task_in_queue Indicates the number of tasks being queued."},
			"txn_counter":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Value of the number of imported transactions in each status."},
			"txn_status":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Count the number of import transactions currently in various states. For example, {type=committed} indicates the number of transactions in the committed state."},
			"query_instance_num":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Specifies the fragment that the user is currently requesting Number of instances . For example, {user=test_u} represents the user test_u The number of instances currently being requested."},
			"query_instance_begin":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Specify the fragment where the user request starts Number of instances . For example, {user=test_u} represents the user test_u Number of instances to start requesting."},
			"query_rpc_total":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Of RPCs sent to the specified BE . For example, { be=192.168.10.1} indicates the number of RPCs sent to BE with IP address 192.168.10.1."},
			"query_rpc_failed":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "RPC failures sent to the specified BE . For example, { be=192.168.10.1} indicates the number of RPC failures sent to BE with IP address 192.168.10.1."},
			"query_rpc_size":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Specify the RPC data size of BE . For example, { be=192.168.10.1} indicates the number of RPC data bytes sent to BE with IP address 192.168.10.1."},
			"txn_exec_latency_ms":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Percentile statistics of transaction execution time. For example, {quantile=0.75} indicates the 75th percentile transaction execution time."},
			"txn_publish_latency_ms":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "Percentile statistics of transaction publish time. For example, {quantile=0.75} indicates that the 75th percentile transaction publish time is."},
			"txn_num":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Specifies the number of transactions being performed by the DB . For example, { db =test} indicates the number of transactions currently being executed by DB test."},
			"publish_txn_num":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Specify the number of transactions being published by the DB . For example, { db =test} indicates the number of transactions currently being published by DB test."},
			"txn_replica_num":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Specifies the number of replicas opened by the transaction being executed by the DB . For example, { db =test} indicates the number of copies opened by the transaction currently being executed by DB test."},
			"thrift_rpc_total":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "RPC requests received by each method of the FE thrift interface . For example, {method=report} indicates the number of RPC requests received by the report method."},
			"thrift_rpc_latency_ms":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Summary, Unit: inputs.DurationMS, Desc: "The RPC requests received by each method of the FE thrift interface take time. For example, {method=report} indicates that the RPC request received by the report method takes time."},
			"external_schema_cache":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "SpecifyExternal Catalog _ The number of corresponding schema caches."},
			"hive_meta_cache":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Specify External Hive Meta store Catalog The number of corresponding partition caches."},
		},
		Tags: map[string]interface{}{
			"host":     inputs.NewTagInfo("Host name."),
			"instance": inputs.NewTagInfo("Instance endpoint."),
			"job":      inputs.NewTagInfo("Job type."),
			"method":   inputs.NewTagInfo("Method type."),
			"name":     inputs.NewTagInfo("Metric name."),
			"quantile": inputs.NewTagInfo("quantile."),
			"state":    inputs.NewTagInfo("State."),
			"type":     inputs.NewTagInfo("Metric type."),
			"catalog":  inputs.NewTagInfo("Catalog."),
		},
	}
}

type jvmMeasurement struct{}

func (*jvmMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "doris_jvm",
		Cat:  point.Metric,
		//nolint:lll
		Fields: map[string]interface{}{
			"heap_size_bytes":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "JVM memory metrics. The tags include max, used, committed , corresponding to the maximum value, used and requested memory respectively."},
			"non_heap_size_bytes": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "JVM off-heap memory statistics."},
			"old_gc":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Cumulative value of GC."},
			"old_size_bytes":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "JVM old generation memory statistics."},
			"thread":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "JVM thread count statistics."},
			"young_size_bytes":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "JVM new generation memory statistics."},
		},
		Tags: map[string]interface{}{
			"host":     inputs.NewTagInfo("Host name."),
			"instance": inputs.NewTagInfo("Instance endpoint."),
			"type":     inputs.NewTagInfo("Metric type."),
		},
	}
}

type commonMeasurement struct{}

func (*commonMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "doris_common",
		Cat:  point.Metric,
		//nolint:lll
		Fields: map[string]interface{}{
			"system_meminfo": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "FE node machines. Collected from /proc/meminfo . include buffers , cached , memory_available , memory_free , memory_total."},
			"system_snmp":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "FE node machines. Collected from /proc/net/ SNMP."},
			"node_info":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Node_number."},
		},
		Tags: map[string]interface{}{
			"host":     inputs.NewTagInfo("Host name."),
			"instance": inputs.NewTagInfo("Instance endpoint."),
			"name":     inputs.NewTagInfo("Metric name."),
			"type":     inputs.NewTagInfo("Metric type."),
			"state":    inputs.NewTagInfo("Metric state."),
		},
	}
}

type beMeasurement struct{}

func (*beMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "doris_be",
		Cat:  point.Metric,
		//nolint:lll
		Fields: map[string]interface{}{
			// be process metrics
			"active_scan_context_count":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of scanners currently opened directly from the outside."},
			"add_batch_task_queue_size":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "When recording import, the queue size of the thread pool that receives the batch."},
			"agent_task_queue_size":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Display the length of each Agent Task processing queue, such as {type=CREATE_TABLE} Indicates the length of the CREATE_TABLE task queue."},
			"brpc_endpoint_stub_count":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Created _ The number of `brpc` stubs used for interaction between BEs."},
			"brpc_function_endpoint_stub_count":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Created _ The number of `brpc` stubs used to interact with Remote RPC."},
			"cache_capacity":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Record the capacity of the specified LRU Cache."},
			"cache_usage":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Record the usage of the specified LRU Cache."},
			"cache_usage_ratio":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Record the usage of the specified LRU Cache."},
			"cache_lookup_count":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Record the number of times the specified LRU Cache is searched."},
			"cache_hit_count":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Record the number of hits in the specified LRU Cache."},
			"cache_hit_ratio":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Record the hit rate of the specified LRU Cache."},
			"chunk_pool_local_core_alloc_count":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "ChunkAllocator , the number of times memory is allocated from the memory queue of the bound core."},
			"chunk_pool_other_core_alloc_count":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "ChunkAllocator , the number of times memory is allocated from the memory queue of other cores."},
			"chunk_pool_reserved_bytes":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "ChunkAllocator The amount of memory reserved in."},
			"chunk_pool_system_alloc_cost_ns":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationNS, Desc: "SystemAllocator The cumulative value of time spent applying for memory."},
			"chunk_pool_system_alloc_count":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "SystemAllocator Number of times to apply for memory."},
			"chunk_pool_system_free_cost_ns":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationNS, Desc: "SystemAllocator Cumulative value of time taken to release memory."},
			"chunk_pool_system_free_count":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "SystemAllocator The number of times memory is released."},
			"compaction_bytes_total":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Value of the amount of data processed by compaction."},
			"compaction_deltas_total":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Processed by compaction `rowset` The cumulative value of the number."},
			"disks_compaction_num":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Compaction tasks being executed on the specified data directory . like {path=path1} means /path1 The number of tasks being executed on the directory."},
			"disks_compaction_score":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Specifies the number of compaction tokens being executed on the data directory. like {path=path1} means /path1 Number of tokens being executed on the directory."},
			"compaction_used_permits":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of tokens used by the Compaction task."},
			"compaction_waitting_permits":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Compaction tokens awaiting."},
			"data_stream_receiver_count":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of data receiving terminals Receiver."},
			"disks_avail_capacity":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Specify the remaining space on the disk where the specified data directory is located. like {path=path1} express /path1 The remaining space on the disk where the directory is located."},
			"disks_local_used_capacity":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The specified data directory is located."},
			"disks_remote_used_capacity":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The specified data directory is located."},
			"disks_state":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Specifies the disk status of the data directory . 1 means normal. 0 means abnormal."},
			"disks_total_capacity":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Capacity of the disk where the specified data directory is located."},
			"engine_requests_total":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Engine_requests total on BE."},
			"fragment_endpoint_count":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Value of various task execution statuses on BE."},
			"fragment_request_duration_us":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationUS, Desc: "All fragment `intance` The cumulative execution time of."},
			"fragment_requests_total":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The cumulative number of executed fragment instances."},
			"load_channel_count":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of load channels currently open."},
			"local_bytes_read_total":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Depend on LocalFileReader Number of bytes read."},
			"local_bytes_written_total":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Depend on LocalFileWriter Number of bytes written."},
			"local_file_reader_total":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Opened LocalFileReader Cumulative count of."},
			"local_file_open_reading":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Currently open LocalFileReader number."},
			"local_file_writer_total":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Opened LocalFileWriter cumulative count."},
			"mem_consumption":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Specifies the current memory overhead of the module . For example, {type=compaction} represents the current total memory overhead of the compaction module."},
			"memory_allocated_bytes":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "BE process physical memory size, taken from /proc/self/status/ VmRSS."},
			"memory_jemalloc":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Jemalloc stats, taken from `je_mallctl`."},
			"memory_pool_bytes_total":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "all MemPool The size of memory currently occupied. Statistical value, does not represent actual memory usage."},
			"memtable_flush_duration_us":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationUS, Desc: "value of the time taken to write `memtable` to disk."},
			"memtable_flush_total":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "number of `memtable` writes to disk."},
			"meta_request_duration":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationUS, Desc: "Access RocksDB The cumulative time consumption of meta in."},
			"meta_request_total":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Access RocksDB The cumulative number of meta requests."},
			"fragment_instance_count":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of fragment instances currently received."},
			"process_fd_num_limit_hard":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "BE process. pass /proc/ pid /limits collection."},
			"process_fd_num_limit_soft":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "BE process. pass /proc/ pid /limits collection."},
			"process_fd_num_used":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of file handles used by the BE process. pass /proc/ pid /limits collection."},
			"process_thread_num":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "BE process threads. pass /proc/ pid /task collection."},
			"query_cache_memory_total_byte":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Number of bytes occupied by Query Cache."},
			"query_cache_partition_total_count":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Current number of Partition Cache caches."},
			"query_cache_sql_total_count":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of SQL Cache caches."},
			"query_scan_bytes":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Read the cumulative value of the data amount. Here we only count reads `Olap` The amount of data in the table."},
			"query_scan_bytes_per_second":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec, Desc: "According to doris_be_query_scan_bytes Calculated read rate."},
			"query_scan_rows":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Read the cumulative value of the number of rows. Here we only count reads `Olap` The amount of data in the table. and is RawRowsRead (Some data rows may be skipped by the index and not actually read, but will still be recorded in this value )."},
			"result_block_queue_count":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of fragment instances in the current query result cache."},
			"result_buffer_block_count":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of queries in the current query result cache."},
			"routine_load_task_count":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of routine load tasks currently being executed."},
			"rowset_count_generated_and_in_use":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "New and in use since the last startup The number of `rowset` ids."},
			"s3_bytes_read_total":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "S3FileReader The cumulative number."},
			"s3_file_open_reading":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "currently open S3FileReader number."},
			"scanner_thread_pool_queue_size":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "used for `OlapScanner` The current queued number of thread pools."},
			"segment_read":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Value of the number of segments read."},
			"send_batch_thread_pool_queue_size":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of queues in the thread pool used to send data packets when importing."},
			"send_batch_thread_pool_thread_num":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of threads in the thread pool used to send packets when importing."},
			"small_file_cache_count":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Currently cached by BE."},
			"streaming_load_current_processing":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of stream load tasks currently running."},
			"streaming_load_duration_ms":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "The cumulative value of the execution time of all stream load tasks."},
			"streaming_load_requests_total":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Value of the number of stream load tasks."},
			"stream_load_pipe_count":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current stream load data pipelines."},
			"stream_load":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Value of the number received by stream load."},
			"tablet_base_max_compaction_score":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The current largest Base Compaction Score."},
			"tablet_cumulative_max_compaction_score": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Same as above. Current largest Cumulative Compaction Score."},
			"tablet_version_num_distribution":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Histogram, Unit: inputs.NCount, Desc: "The histogram of the number of tablet versions."},
			"thrift_connections_total":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Thrift connections created . like {name=heartbeat} Indicates the cumulative number of connections to the heartbeat service."},
			"thrift_current_connections":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of thrift connections. like {name=heartbeat} Indicates the current number of connections to the heartbeat service."},
			"thrift_opened_clients":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Thrift clients currently open . like {name=frontend} Indicates the number of clients accessing the FE service."},
			"thrift_used_clients":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Thrift clients currently in use . like {name=frontend} Indicates the number of clients being used to access the FE service."},
			"timeout_canceled_fragment_count":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Cumulative value of the number of fragment instances canceled due to timeout."},
			"stream_load_txn_request":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Value of the number of transactions by stream load."},
			"unused_rowsets_count":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of currently abandoned `rowsets`."},
			"upload_fail_count":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative value of `rowset` failed to be uploaded to remote storage."},
			"upload_rowset_count":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of `rowsets` successfully uploaded to remote storage."},
			"upload_total_byte":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Value of `rowset` data successfully uploaded to remote storage."},
			"load_bytes":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Cumulative quantity sent through tablet Sink."},
			"load_rows":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of rows sent through tablet Sink."},
			"fragment_thread_pool_queue_size":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current query execution thread pool waiting queue."},
			"all_rowsets_num":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "All currently `rowset` number of."},
			"all_segments_num":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of all current segments."},
			"heavy_work_max_threads":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of heavy thread pool threads."},
			"light_work_max_threads":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of light thread pool threads."},
			"heavy_work_pool_queue_size":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The maximum length of the heavy thread pool queue will block the submission of work if it exceeds it."},
			"light_work_pool_queue_size":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The maximum length of the light thread pool queue . If it exceeds the maximum length, the submission of work will be blocked."},
			"heavy_work_active_threads":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of active threads in heavy thread pool."},
			"light_work_active_threads":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of active threads in light thread pool."},

			// be machine metrics
			"cpu":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "CPU related metrics metrics, from /proc/stat collection. Each value of each logical core will be collected separately . like {device=cpu0,mode =nice} Indicates the nice value of cpu0."},
			"disk_bytes_read":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "The cumulative value of disk reads. from /proc/ `diskstats` collection. The values of each disk will be collected separately . like {device=vdd} express vvd disk value."},
			"disk_bytes_written":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "The cumulative value of disk writes."},
			"disk_io_time_ms":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "The dis io time."},
			"disk_io_time_weighted":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "The dis io time weighted."},
			"disk_reads_completed":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "The dis reads completed."},
			"disk_read_time_ms":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "The dis reads time."},
			"disk_writes_completed":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "The disk writes completed."},
			"disk_write_time_ms":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "The disk write time."},
			"fd_num_limit":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "System file handle limit upper limit. from /proc/sys/fs/file-nr collection."},
			"fd_num_used":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of file handles used by the system . from /proc/sys/fs/file-nr collection."},
			"file_created_total":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of local file creation times."},
			"load_average":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Machine Load Avg Metric metrics. For example, {mode=15_minutes} is 15 minutes Load Avg."},
			"max_disk_io_util_percent":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "value of the disk with the largest IO UTIL among all disks."},
			"max_network_receive_bytes_rate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec, Desc: "The maximum receive rate calculated among all network cards."},
			"max_network_send_bytes_rate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec, Desc: "The calculated maximum sending rate among all network cards."},
			"memory_pgpgin":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of data written by the system from disk to memory page."},
			"memory_pgpgout":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of data written to disk by system memory pages."},
			"memory_pswpin":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The number of times the system swapped from disk to memory."},
			"memory_pswpout":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The number of times the system swapped from memory to disk."},
			"network_receive_bytes":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "each network card are accumulated. Collected from /proc/net/dev."},
			"network_receive_packets":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "each network card is accumulated. Collected from /proc/net/dev."},
			"network_send_bytes":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "each network card . Collected from /proc/net/dev."},
			"network_send_packets":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of packets sent by each network card is accumulated. Collected from /proc/net/dev."},
			"proc":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of processes currently ."},
			"snmp_tcp_in_errs":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "tcp packet reception errors. Collected from /proc/net/ SNMP."},
			"snmp_tcp_in_segs":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "tcp packets sent . Collected from /proc/net/ SNMP."},
			"snmp_tcp_out_segs":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "tcp packets sent. Collected from /proc/net/ SNMP."},
			"snmp_tcp_retrans_segs":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: "TCP packet retransmissions . Collected from /proc/net/ SNMP."},
		},
		Tags: map[string]interface{}{
			"host":     inputs.NewTagInfo("Host name."),
			"instance": inputs.NewTagInfo("Instance endpoint."),
			"mode":     inputs.NewTagInfo("Metric mode."),
			"name":     inputs.NewTagInfo("Metric name."),
			"path":     inputs.NewTagInfo("File path."),
			"quantile": inputs.NewTagInfo("quantile."),
			"type":     inputs.NewTagInfo("Metric type."),
			"status":   inputs.NewTagInfo("Metric status."),
			"device":   inputs.NewTagInfo("Device name."),
		},
	}
}
