//nolint:lll
package mongodb

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type mongodbMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *mongodbMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *mongodbMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "mongodb",
		Tags: map[string]interface{}{
			"hostname":  &inputs.TagInfo{Desc: "mongodb host"},
			"node_type": &inputs.TagInfo{Desc: "node type in replica set"},
			"rs_name":   &inputs.TagInfo{Desc: "replica set name"},
		},
		Fields: map[string]interface{}{
			"active_reads":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of the active client connections performing read operations.`},                                                                                                                                                                                                                                                     // (integer)
			"active_writes":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of active client connections performing write operations.`},                                                                                                                                                                                                                                                        // (integer)
			"aggregate_command_failed": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times that 'aggregate' command failed on this mongod`},                                                                                                                                                                                                                                                          // (integer)
			"aggregate_command_total":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times that 'aggregate' command executed on this mongod.`},                                                                                                                                                                                                                                                       // (integer)
			"assert_msg":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of message assertions raised since the MongoDB process started. Check the log file for more information about these messages.`},                                                                                                                                                                                    // (integer)
			"assert_regular":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of regular assertions raised since the MongoDB process started. Check the log file for more information about these messages.`},                                                                                                                                                                                    // (integer)
			"assert_rollovers":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times that the rollover counters have rolled over since the last time the MongoDB process started. The counters will rollover to zero after 2 30 assertions. Use this value to provide context to the other values in the asserts data structure.`},                                                             // (integer)
			"assert_user":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of "user asserts" that have occurred since the last time the MongoDB process started. These are errors that user may generate, such as out of disk space or duplicate key. You can prevent these assertions by fixing a problem with your application or deployment. Check the MongoDB log for more information.`}, // (integer)
			"assert_warning":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Changed in version 4.0. Starting in MongoDB 4.0, the field returns zero 0. In earlier versions, the field returns the number of warnings raised since the MongoDB process started.`},                                                                                                                                          // (integer)
			"available_reads":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of concurrent of read transactions allowed into the WiredTiger storage engine`},                                                                                                                                                                                                                                    // (integer)
			"available_writes":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of concurrent of write transactions allowed into the WiredTiger storage engine`},                                                                                                                                                                                                                                   // (integer)
			"commands":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of commands issued to the database since the mongod instance last started. opcounters.command counts all commands except the write commands: insert, update, and delete.`},                                                                                                                                   // (integer)
			// "commands_per_sec":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``}, // (integer, deprecated in 1.10; use commands))
			"connections_available":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of unused incoming connections available.`},                                                              // (integer)
			"connections_current":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of incoming connections from clients to the database server .`},                                          // (integer)
			"connections_total_created": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Count of all incoming connections created to the server. This number includes connections that have since closed.`}, // (integer)
			"count_command_failed":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times that 'count' command failed on this mongod`},                                                    // (integer)
			"count_command_total":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times that 'count' command executed on this mongod`},                                                  // (integer)
			// "cursor_no_timeout":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``}, // (integer, opened/sec, deprecated in 1.10; use cursor_no_timeout_count))
			"cursor_no_timeout_count": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of open cursors with the option DBQuery.Option.noTimeout set to prevent timeout after a period of inactivity`}, // (integer)
			// "cursor_pinned":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``}, // (integer, opened/sec, deprecated in 1.10; use cursor_pinned_count))
			"cursor_pinned_count": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of "pinned" open cursors.`}, // (integer)
			// "cursor_timed_out":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``}, // (integer, opened/sec, deprecated in 1.10; use cursor_timed_out_count))
			"cursor_timed_out_count": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of cursors that have timed out since the server process started. If this number is large or growing at a regular rate, this may indicate an application error.`}, // (integer)
			// "cursor_total":                              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``},    // (integer, opened/sec, deprecated in 1.10; use cursor_total_count))
			"cursor_total_count":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of cursors that MongoDB is maintaining for clients. Because MongoDB exhausts unused cursors, typically this value small or zero. However, if there is a queue, stale tailable cursors, or a large number of operations this value may rise.`}, // (integer)
			"delete_command_failed": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times that 'delete' command failed on this mongod`},                                                                                                                                                                                        // (integer)
			"delete_command_total":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times that 'delete' command executed on this mongod`},                                                                                                                                                                                      // (integer)
			"deletes":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of delete operations since the mongod instance last started.`},                                                                                                                                                                          // (integer)
			// "deletes_per_sec":                           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``},    // (integer, deprecated in 1.10; use deletes))
			"distinct_command_failed":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times that 'distinct' command failed on this mongod`},             // (integer)
			"distinct_command_total":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times that 'distinct' command executed on this mongod`},           // (integer)
			"document_deleted":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of documents deleted.`},                                        // (integer)
			"document_inserted":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of documents inserted.`},                                       // (integer)
			"document_returned":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of documents returned by queries.`},                            // (integer)
			"document_updated":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of documents updated.`},                                        // (integer)
			"find_and_modify_command_failed": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times that 'find' and 'modify' commands failed on this mongod`},   // (integer)
			"find_and_modify_command_total":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times that 'find' and 'modify' commands executed on this mongod`}, // (integer)
			"find_command_failed":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times that 'find' command failed on this mongod`},                 // (integer)
			"find_command_total":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times that 'find' command executed on this mongod`},               // (integer)
			"flushes":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of transaction checkpoints`},                                         // (integer)
			// "flushes_per_sec":                           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``},    // (integer, deprecated in 1.10; use flushes))
			"flushes_total_time_ns":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The transaction checkpoint total time (msecs)"`},                                                                                                                                                                      // (integer)
			"get_more_command_failed": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times that 'get more' command failed on this mongod`},                                                                                                                                                   // (integer)
			"get_more_command_total":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times that 'get more' command executed on this mongod`},                                                                                                                                                 // (integer)
			"getmores":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of getMore operations since the mongod instance last started. This counter can be high even if the query count is low. Secondary nodes send getMore operations as part of the replication process.`}, // (integer)
			// "getmores_per_sec":                          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``},    // (integer, deprecated in 1.10; use getmores))
			"insert_command_failed": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times that 'insert' command failed on this mongod`},                        // (integer)
			"insert_command_total":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times that 'insert' command executed on this mongod`},                      // (integer)
			"inserts":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of insert operations received since the mongod instance last started.`}, // (integer)
			// "inserts_per_sec":                           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``},    // (integer, deprecated in 1.10; use inserts))
			"jumbo_chunks":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Count jumbo flags in cluster chunk.`},                                                        // (integer)
			"latency_commands":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total combined latency in microseconds of latency statistics for database command.`},     // (integer)
			"latency_commands_count": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total combined latency of operations performed on the collection for database command.`}, // (integer)
			"latency_reads":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total combined latency in microseconds of latency statistics for read request.`},         // (integer)
			"latency_reads_count":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total combined latency of operations performed on the collection for read request.`},     // (integer)
			"latency_writes":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total combined latency in microseconds of latency statistics for write request.`},        // (integer)
			"latency_writes_count":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total combined latency of operations performed on the collection for write request.`},    // (integer)
			"member_status":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The state of ndoe in replica members.`},                                                   // (string)
			// "net_in_bytes":                              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``},    // (integer, bytes/sec, deprecated in 1.10; use net_out_bytes_count))
			"net_in_bytes_count": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of bytes that the server has received over network connections initiated by clients or other mongod instances.`}, // (integer)
			// "net_out_bytes":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``},    // (integer, bytes/sec, deprecated in 1.10; use net_out_bytes_count))
			"net_out_bytes_count":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of bytes that the server has sent over network connections initiated by clients or other mongod instances.`}, // (integer)
			"open_connections":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of incoming connections from clients to the database server.`},                                                     // (integer)
			"operation_scan_and_order":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of queries that return sorted numbers that cannot perform the sort operation using an index.`},               // (integer)
			"operation_write_conflicts": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of queries that encountered write conflicts.`},                                                               // (integer)
			"page_faults":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of page faults.`},                                                                                            // (integer)
			"percent_cache_dirty":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Size in bytes of the dirty data in the cache. This value should be less than the bytes currently in the cache value.`},      // (float)
			"percent_cache_used":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Size in byte of the data currently in cache. This value should not be greater than the maximum bytes configured value.`},    // (float)
			"queries":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of queries received since the mongod instance last started.`},                                                // (integer)
			// "queries_per_sec":                           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``},    // (integer, deprecated in 1.10; use queries))
			"queued_reads":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of operations that are currently queued and waiting for the read lock. A consistently small read-queue, particularly of shorter operations, should cause no concern.`},   // (integer)
			"queued_writes":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of operations that are currently queued and waiting for the write lock. A consistently small write-queue, particularly of shorter operations, is no cause for concern.`}, // (integer)
			"repl_apply_batches_num":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of batches applied across all databases.`},                                                                                                                         // (integer)
			"repl_apply_batches_total_millis": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total amount of time in milliseconds the mongod has spent applying operations from the oplog.`},                                                                                 // (integer)
			"repl_apply_ops":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of oplog operations applied. metrics.repl.apply.ops is incremented after each operation.`},                                                                         // (integer)
			"repl_buffer_count":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The current number of operations in the oplog buffer.`},                                                                                                                             // (integer)
			"repl_buffer_size_bytes":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The current size of the contents of the oplog buffer.`},                                                                                                                             // (integer)
			"repl_commands":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of replicated commands issued to the database since the mongod instance last started.`},                                                                            // (integer)
			// "repl_commands_per_sec":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``},    // (integer, deprecated in 1.10; use repl_commands))
			"repl_deletes": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of replicated delete operations since the mongod instance last started.`}, // (integer)
			// "repl_deletes_per_sec":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``},    // (integer, deprecated in 1.10; use repl_deletes)
			"repl_executor_pool_in_progress_count":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: inputs.TODO},                                                                      // (integer)
			"repl_executor_queues_network_in_progress": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: inputs.TODO},                                                                      // (integer)
			"repl_executor_queues_sleepers":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: inputs.TODO},                                                                      // (integer)
			"repl_executor_unsignaled_events":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: inputs.TODO},                                                                      // (integer)
			"repl_getmores":                            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of getMore operations since the mongod instance last started.`}, // (integer)
			// "repl_getmores_per_sec":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``},    // (integer, deprecated in 1.10; use repl_getmores)
			"repl_inserts": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of replicated insert operations since the mongod instance last started.`}, // (integer)
			// "repl_inserts_per_sec":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``},    // (integer, deprecated in 1.10; use repl_inserts))
			"repl_lag":                           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: inputs.TODO},                                                                                                                                   // (integer)
			"repl_network_bytes":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total amount of data read from the replication sync source.`},                                                                             // (integer)
			"repl_network_getmores_num":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of getmore operations, which are operations that request an additional set of operations from the replication sync source.`}, // (integer)
			"repl_network_getmores_total_millis": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total amount of time required to collect data from getmore operations.`},                                                                  // (integer)
			"repl_network_ops":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of operations read from the replication source.`},                                                                            // (integer)
			"repl_oplog_window_sec":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The second window of replication oplog.`},                                                                                                     // (integer)
			"repl_queries":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of replicated queries since the mongod instance last started.`},                                                              // (integer)
			// "repl_queries_per_sec":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``},    // (integer, deprecated in 1.10; use repl_queries))
			"repl_state":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The node state of replication member.`},                                                    // (integer)
			"repl_updates": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of replicated update operations since the mongod instance last started.`}, // (integer)
			// "repl_updates_per_sec":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``},    // (integer, deprecated in 1.10; use repl_updates))
			"resident_megabytes": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The value of mem.resident is roughly equivalent to the amount of RAM, in mebibyte (MiB), currently used by the database process.`}, // (integer)
			"state":              &inputs.FieldInfo{DataType: inputs.String, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The replication state.`},                                                                                                        // (string)
			"storage_freelist_search_bucket_exhausted":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times that mongod has checked the free list without finding a suitably large record allocation.`},                                                                       // (integer)
			"storage_freelist_search_requests":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times mongod has searched for available record allocations.`},                                                                                                           // (integer)
			"storage_freelist_search_scanned":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of available record allocations mongod has searched.`},                                                                                                                     // (integer)
			"tcmalloc_central_cache_free_bytes":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Number of free bytes in the central cache that have been assigned to size classes. They always count towards virtual memory usage, and unless the underlying memory is swapped.`},     // (integer)
			"tcmalloc_current_allocated_bytes":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Number of bytes currently allocated by application.`},                                                                                                                                 // (integer)
			"tcmalloc_current_total_thread_cache_bytes": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Number of bytes used across all thread caches.`},                                                                                                                                      // (integer)
			"tcmalloc_heap_size":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Number of bytes in the heap.`},                                                                                                                                                        // (integer)
			"tcmalloc_max_total_thread_cache_bytes":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Upper limit on total number of bytes stored across all per-thread caches. Default: 16MB.`},                                                                                            // (integer)
			"tcmalloc_pageheap_commit_count":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Number of virtual memory commits.`},                                                                                                                                                   // (integer)
			"tcmalloc_pageheap_committed_bytes":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Bytes committed, always <= system_bytes_.`},                                                                                                                                           // (integer)
			"tcmalloc_pageheap_decommit_count":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Number of virtual memory decommits.`},                                                                                                                                                 // (integer)
			"tcmalloc_pageheap_free_bytes":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Number of bytes in free, mapped pages in page heap.`},                                                                                                                                 // (integer)
			"tcmalloc_pageheap_reserve_count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Number of virtual memory reserves.`},                                                                                                                                                  // (integer)
			"tcmalloc_pageheap_scavenge_count":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Number of times scavagened flush pages.`},                                                                                                                                             // (integer)
			"tcmalloc_pageheap_total_commit_bytes":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Bytes committed in lifetime of process.`},                                                                                                                                             // (integer)
			"tcmalloc_pageheap_total_decommit_bytes":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Bytes decommitted in lifetime of process.`},                                                                                                                                           // (integer)
			"tcmalloc_pageheap_total_reserve_bytes":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Number of virtual memory reserves.`},                                                                                                                                                  // (integer)
			"tcmalloc_pageheap_unmapped_bytes":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Total bytes on returned freelists.`},                                                                                                                                                  // (integer)
			"tcmalloc_spinlock_total_delay_ns":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: inputs.TODO},                                                                                                                                                                           // (integer)
			"tcmalloc_thread_cache_free_bytes":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Bytes in thread caches.`},                                                                                                                                                             // (integer)
			"tcmalloc_total_free_bytes":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Total bytes on normal freelists.`},                                                                                                                                                    // (integer)
			"tcmalloc_transfer_cache_free_bytes":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Bytes in central transfer cache.`},                                                                                                                                                    // (integer)
			"total_available":                           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of connections available from the mongos to the config servers, replica sets, and standalone mongod instances in the cluster.`},                                            // (integer)
			"total_created":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of connections the mongos has ever created to other members of the cluster.`},                                                                                              // (integer)
			"total_docs_scanned":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of index items scanned during queries and query-plan evaluation.`},                                                                                                   // (integer)
			"total_in_use":                              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Reports the total number of outgoing connections from the current mongod/mongos instance to other members of the sharded cluster or replica set that are currently in use.`},          // (integer)
			"total_keys_scanned":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of index items scanned during queries and query-plan evaluation.`},                                                                                                   // (integer)
			"total_refreshing":                          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Reports the total number of outgoing connections from the current mongod/mongos instance to other members of the sharded cluster or replica set that are currently being refreshed.`}, // (integer)
			"total_tickets_reads":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `A document that returns information on the number of concurrent of read transactions allowed into the WiredTiger storage engine.`},                                                    // (integer)
			"total_tickets_writes":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `A document that returns information on the number of concurrent of write transactions allowed into the WiredTiger storage engine.`},                                                   // (integer)
			"ttl_deletes":                               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of documents deleted from collections with a ttl index.`},                                                                                                            // (integer)
			// "ttl_deletes_per_sec":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``},    // (integer, deprecated in 1.10; use ttl_deletes))
			"ttl_passes": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times the background process removes documents from collections with a ttl index.`}, // (integer)
			// "ttl_passes_per_sec":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``},    // (integer, deprecated in 1.10; use ttl_passes))
			"update_command_failed": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times that 'update' command failed on this mongod`},                        // (integer)
			"update_command_total":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of times that 'update' command executed on this mongod`},                      // (integer)
			"updates":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of update operations received since the mongod instance last started.`}, // (integer)
			// "updates_per_sec":                           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ``},    // (integer, deprecated in 1.10; use updates))
			"uptime_ns":                            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total upon time of mongod in nano seconds.`},                                                      // (integer)
			"version":                              &inputs.FieldInfo{DataType: inputs.String, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Mongod version`},                                                                                   // (string)
			"vsize_megabytes":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `mem.virtual displays the quantity, in mebibyte (MiB), of virtual memory used by the mongod process.`}, // (integer)
			"wtcache_app_threads_page_read_count":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: inputs.TODO},                                                                                           // (integer)
			"wtcache_app_threads_page_read_time":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: inputs.TODO},                                                                                           // (integer)
			"wtcache_app_threads_page_write_count": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: inputs.TODO},                                                                                           // (integer)
			"wtcache_bytes_read_into":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: inputs.TODO},                                                                                           // (integer)
			"wtcache_bytes_written_from":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: inputs.TODO},                                                                                           // (integer)
			"wtcache_current_bytes":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: inputs.TODO},                                                                                           // (integer)
			"wtcache_internal_pages_evicted":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: inputs.TODO},                                                                                           // (integer)
			"wtcache_max_bytes_configured":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Maximum cache size.`},                                                                                 // (integer)
			"wtcache_modified_pages_evicted":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: inputs.TODO},                                                                                           // (integer)
			"wtcache_pages_evicted_by_app_thread":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: inputs.TODO},                                                                                           // (integer)
			"wtcache_pages_queued_for_eviction":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: inputs.TODO},                                                                                           // (integer)
			"wtcache_pages_read_into":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Number of pages read into the cache.`},                                                                // (integer)
			"wtcache_pages_requested_from":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Number of pages request from the cache.`},                                                             // (integer)
			"wtcache_server_evicting_pages":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: inputs.TODO},                                                                                           // (integer)
			"wtcache_tracked_dirty_bytes":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: inputs.TODO},                                                                                           // (integer)
			"wtcache_unmodified_pages_evicted":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Main statistics for page eviction.`},                                                                  // (integer)
			"wtcache_worker_thread_evictingpages":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: inputs.TODO},                                                                                           // (integer)
		},
	}
}

type mongodbDBMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *mongodbDBMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *mongodbDBMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "mongodb_db_stats",
		Tags: map[string]interface{}{
			"db_name":  &inputs.TagInfo{Desc: "database name"},
			"hostname": &inputs.TagInfo{Desc: "mongodb host"},
		},
		Fields: map[string]interface{}{
			"avg_obj_size": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The average size of each document in bytes."},                                                                    // (float)
			"collections":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Contains a count of the number of collections in that database."},                                                  // (integer)
			"data_size":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total size of the uncompressed data held in this database. The dataSize decreases when you remove documents."}, // (integer)
			"index_size":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total size of all indexes created on this database."},                                                          // (integer)
			"indexes":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Contains a count of the total number of indexes across all collections in the database."},                          // (integer)
			"objects":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Contains a count of the number of objects (i.e. documents) in the database across all collections."},               // (integer)
			"ok":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Command execute state."},                                                                                           // (integer)
			"storage_size": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total amount of space allocated to collections in this database for document storage."},                        // (integer)
			"type":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Metrics type."},                                                                                                 // (string)
		},
	}
}

type mongodbColMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *mongodbColMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *mongodbColMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "mongodb_col_stats",
		Tags: map[string]interface{}{
			"collection": &inputs.TagInfo{Desc: "collection name"},
			"db_name":    &inputs.TagInfo{Desc: "database name"},
			"hostname":   &inputs.TagInfo{Desc: "mongodb host"},
		},
		Fields: map[string]interface{}{
			"avg_obj_size":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The average size of an object in the collection. "},                              // (integer)
			"count":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of objects or documents in this collection."},                         // (integer)
			"ok":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Command execute state."},                                                         // (integer)
			"size":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total uncompressed size in memory of all records in a collection."},          // (integer)
			"storage_size":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total amount of storage allocated to this collection for document storage."}, // (integer)
			"total_index_size": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total size of all indexes."},                                                 // (integer)
			"type":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Metrics type."},                                                                  // (string)
		},
	}
}

type mongodbShardMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *mongodbShardMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *mongodbShardMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "mongodb_shard_stats",
		Tags: map[string]interface{}{
			"hostname": &inputs.TagInfo{Desc: "mongodb host"},
		},
		Fields: map[string]interface{}{
			"available":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of connections available for this host to connect to the mongos.`},                                                                                                         // (integer)
			"created":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The number of connections the host has ever created to connect to the mongos.`},                                                                                                       // (integer)
			"in_use":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Reports the total number of outgoing connections from the current mongod/mongos instance to other members of the sharded cluster or replica set that are currently in use.`},          // (integer)
			"refreshing": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `Reports the total number of outgoing connections from the current mongod/mongos instance to other members of the sharded cluster or replica set that are currently being refreshed.`}, // (integer)
		},
	}
}

type mongodbTopMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *mongodbTopMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *mongodbTopMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "mongodb_top_stats",
		Tags: map[string]interface{}{
			"collection": &inputs.TagInfo{Desc: "collection name"},
			"hostname":   &inputs.TagInfo{Desc: "mongodb host"},
		},
		Fields: map[string]interface{}{
			"commands_count":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of "command" event issues.`},                // (integer)
			"commands_time":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The amount of time in microseconds that "command" costs.`},   // (integer)
			"get_more_count":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of "getmore" event issues.`},                // (integer)
			"get_more_time":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The amount of time in microseconds that "getmore" costs.`},   // (integer)
			"insert_count":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of "insert" event issues.`},                 // (integer)
			"insert_time":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The amount of time in microseconds that "insert" costs.`},    // (integer)
			"queries_count":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of "queries" event issues.`},                // (integer)
			"queries_time":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The amount of time in microseconds that "queries" costs.`},   // (integer)
			"read_lock_count":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of "readLock" event issues.`},               // (integer)
			"read_lock_time":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The amount of time in microseconds that "readLock" costs.`},  // (integer)
			"remove_count":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of "remove" event issues.`},                 // (integer)
			"remove_time":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The amount of time in microseconds that "remove" costs.`},    // (integer)
			"total_count":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of "total" event issues.`},                  // (integer)
			"total_time":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The amount of time in microseconds that "total" costs.`},     // (integer)
			"update_count":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of "update" event issues.`},                 // (integer)
			"update_time":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The amount of time in microseconds that "update" costs.`},    // (integer)
			"write_lock_count": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The total number of "writeLock" event issues.`},              // (integer)
			"write_lock_time":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: `The amount of time in microseconds that "writeLock" costs.`}, // (integer)
		},
	}
}
