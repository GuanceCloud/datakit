// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type innodbMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

// Point implement MeasurementV2.
func (m *innodbMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()

	if m.election {
		opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	}

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll,funlen
func (m *innodbMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameMySQLInnodb,
		Type: "metric",
		Fields: map[string]interface{}{
			"lock_deadlocks": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of deadlocks",
			},
			"lock_timeouts": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of lock timeouts",
			},
			"lock_row_lock_current_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of row locks currently being waited for (innodb_row_lock_current_waits)",
			},
			"lock_row_lock_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Time spent in acquiring row locks, in milliseconds (innodb_row_lock_time)",
			},
			"lock_row_lock_time_max": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "The maximum time to acquire a row lock, in milliseconds (innodb_row_lock_time_max)",
			},
			"lock_row_lock_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of times a row lock had to be waited for (innodb_row_lock_waits)",
			},
			"lock_row_lock_time_avg": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "The average time to acquire a row lock, in milliseconds (innodb_row_lock_time_avg)",
			},
			"buffer_pool_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Server buffer pool size (all buffer pools) in bytes",
			},
			"buffer_pool_reads": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of reads directly from disk (innodb_buffer_pool_reads)",
			},
			"buffer_pool_read_requests": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of logical read requests (innodb_buffer_pool_read_requests)",
			},
			"buffer_pool_write_requests": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of write requests (innodb_buffer_pool_write_requests)",
			},
			"buffer_pool_wait_free": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of times waited for free buffer (innodb_buffer_pool_wait_free)",
			},
			"buffer_pool_read_ahead": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of pages read as read ahead (innodb_buffer_pool_read_ahead)",
			},
			"buffer_pool_read_ahead_evicted": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Read-ahead pages evicted without being accessed (innodb_buffer_pool_read_ahead_evicted)",
			},
			"buffer_pool_pages_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Total buffer pool size in pages (innodb_buffer_pool_pages_total)",
			},
			"buffer_pool_pages_misc": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Buffer pages for misc use such as row locks or the adaptive hash index (innodb_buffer_pool_pages_misc)",
			},
			"buffer_pool_pages_data": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Buffer pages containing data (innodb_buffer_pool_pages_data)",
			},
			"buffer_pool_pages_dirty": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Buffer pages currently dirty (innodb_buffer_pool_pages_dirty)",
			},
			"buffer_pool_bytes_dirty": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Buffer bytes containing data (innodb_buffer_pool_bytes_data)",
			},
			"buffer_pool_pages_free": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Buffer pages currently free (innodb_buffer_pool_pages_free)",
			},
			"buffer_pages_created": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of pages created (innodb_pages_created)",
			},
			"buffer_pages_written": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of pages written (innodb_pages_written)",
			},
			"buffer_pages_read": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of pages read (innodb_pages_read)",
			},
			"buffer_data_reads": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Amount of data read in bytes (innodb_data_reads)",
			},
			"buffer_data_written": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Amount of data written in bytes (innodb_data_written)",
			},
			"os_data_reads": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of reads initiated (innodb_data_reads)",
			},
			"os_data_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of writes initiated (innodb_data_writes)",
			},
			"os_data_fsyncs": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of fsync() calls (innodb_data_fsyncs)",
			},
			"os_log_bytes_written": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Bytes of log written (innodb_os_log_written)",
			},
			"os_log_fsyncs": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of fsync log writes (innodb_os_log_fsyncs)",
			},
			"os_log_pending_fsyncs": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of pending fsync write (innodb_os_log_pending_fsyncs)",
			},
			"os_log_pending_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of pending log file writes (innodb_os_log_pending_writes)",
			},
			"trx_rseg_history_len": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Length of the TRX_RSEG_HISTORY list",
			},
			"log_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of log waits due to small log buffer (innodb_log_waits)",
			},
			"log_write_requests": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of log write requests (innodb_log_write_requests)",
			},
			"log_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of log writes (innodb_log_writes)",
			},

			"log_padded": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Bytes of log padded for log write ahead",
			},
			"adaptive_hash_searches": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of successful searches using Adaptive Hash Index",
			},
			"adaptive_hash_searches_btree": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of searches using B-tree on an index search",
			},
			"file_num_open_files": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of files currently open (innodb_num_open_files)",
			},
			"ibuf_merges_insert": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of inserted records merged by change buffering",
			},
			"ibuf_merges_delete_mark": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of deleted records merged by change buffering",
			},
			"ibuf_merges_delete": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of purge records merged by change buffering",
			},
			"ibuf_merges_discard_insert": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of insert merged operations discarded",
			},
			"ibuf_merges_discard_delete_mark": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of deleted merged operations discarded",
			},
			"ibuf_merges_discard_delete": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of purge merged  operations discarded",
			},
			"ibuf_merges": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of change buffer merges",
			},
			"ibuf_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Change buffer size in pages",
			},
			"innodb_activity_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current server activity count",
			},
			"innodb_dblwr_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of `doublewrite` operations that have been performed (innodb_dblwr_writes)",
			},
			"innodb_dblwr_pages_written": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of pages that have been written for `doublewrite` operations (innodb_dblwr_pages_written)",
			},
			"innodb_page_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "InnoDB page size in bytes (innodb_page_size)",
			},
			"innodb_rwlock_s_spin_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rwlock spin waits due to shared latch request",
			},
			"innodb_rwlock_x_spin_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rwlock spin waits due to exclusive latch request",
			},
			"innodb_rwlock_sx_spin_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rwlock spin waits due to sx latch request",
			},
			"innodb_rwlock_s_spin_rounds": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rwlock spin loop rounds due to shared latch request",
			},
			"innodb_rwlock_x_spin_rounds": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rwlock spin loop rounds due to exclusive latch request",
			},
			"innodb_rwlock_sx_spin_rounds": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rwlock spin loop rounds due to sx latch request",
			},
			"innodb_rwlock_s_os_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of OS waits due to shared latch request",
			},
			"innodb_rwlock_x_os_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of OS waits due to exclusive latch request",
			},
			"innodb_rwlock_sx_os_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of OS waits due to sx latch request",
			},
			"dml_inserts": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rows inserted",
			},
			"dml_deletes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rows deleted",
			},
			"dml_updates": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rows updated",
			},
			"active_transactions": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of active transactions on InnoDB tables",
			},
			"buffer_pool_bytes_data": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "The total number of bytes in the InnoDB buffer pool containing data. The number includes both dirty and clean pages.",
			},
			"buffer_pool_pages_flushed": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of requests to flush pages from the InnoDB buffer pool",
			},
			"buffer_pool_read_ahead_rnd": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of random `read-aheads` initiated by InnoDB. This happens when a query scans a large portion of a table but in random order.",
			},
			"checkpoint_age": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Checkpoint age as shown in the LOG section of the `SHOW ENGINE INNODB STATUS` output",
			},
			"current_transactions": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Current `InnoDB` transactions",
			},
			"data_fsyncs": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of fsync() operations per second.",
			},
			"data_pending_fsyncs": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The current number of pending fsync() operations.",
			},
			"data_pending_reads": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The current number of pending reads.",
			},
			"data_pending_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The current number of pending writes.",
			},
			"data_read": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "The amount of data read per second.",
			},
			"data_written": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "The amount of data written per second.",
			},
			"dblwr_pages_written": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of pages written per second to the `doublewrite` buffer.",
			},
			"dblwr_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "The number of `doublewrite` operations performed per second.",
			},
			"hash_index_cells_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Total number of cells of the adaptive hash index",
			},
			"hash_index_cells_used": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of used cells of the adaptive hash index",
			},
			"history_list_length": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "History list length as shown in the TRANSACTIONS section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"ibuf_free_list": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Insert buffer free list, as shown in the INSERT BUFFER AND ADAPTIVE HASH INDEX section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"ibuf_merged": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Insert buffer and `adaptative` hash index merged",
			},
			"ibuf_merged_delete_marks": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Insert buffer and `adaptative` hash index merged delete marks",
			},
			"ibuf_merged_deletes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Insert buffer and `adaptative` hash index merged delete",
			},
			"ibuf_merged_inserts": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Insert buffer and `adaptative` hash index merged inserts",
			},
			"lock_structs": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Lock `structs`",
			},
			"locked_tables": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Locked tables",
			},
			"locked_transactions": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Locked transactions",
			},
			"lsn_current": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Log sequence number as shown in the LOG section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"lsn_flushed": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Flushed up to log sequence number as shown in the LOG section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"lsn_last_checkpoint": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Log sequence number last checkpoint as shown in the LOG section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"mem_adaptive_hash": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "As shown in the BUFFER POOL AND MEMORY section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"mem_additional_pool": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "As shown in the BUFFER POOL AND MEMORY section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"mem_dictionary": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "As shown in the BUFFER POOL AND MEMORY section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"mem_file_system": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "As shown in the BUFFER POOL AND MEMORY section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"mem_lock_system": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the BUFFER POOL AND MEMORY section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"mem_page_hash": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the BUFFER POOL AND MEMORY section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"mem_recovery_system": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the BUFFER POOL AND MEMORY section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"mem_thread_hash": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the BUFFER POOL AND MEMORY section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"mem_total": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the BUFFER POOL AND MEMORY section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"os_file_fsyncs": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "(Delta) The total number of fsync() operations performed by InnoDB.",
			},
			"os_file_reads": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "(Delta) The total number of files reads performed by read threads within InnoDB.",
			},
			"os_file_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "(Delta) The total number of file writes performed by write threads within InnoDB.",
			},
			"os_log_written": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "Number of bytes written to the InnoDB log.",
			},
			"pages_created": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of InnoDB pages created.",
			},
			"pages_read": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of InnoDB pages read.",
			},
			"pages_written": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of InnoDB pages written.",
			},
			"pending_aio_log_ios": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the FILE I/O section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"pending_aio_sync_ios": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the FILE I/O section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"pending_buffer_pool_flushes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the FILE I/O section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"pending_checkpoint_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the FILE I/O section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"pending_ibuf_aio_reads": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the FILE I/O section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"pending_log_flushes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the FILE I/O section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"pending_log_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the FILE I/O section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"pending_normal_aio_reads": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the FILE I/O section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"pending_normal_aio_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the FILE I/O section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"queries_inside": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the FILE I/O section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"queries_queued": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the FILE I/O section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"read_views": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the FILE I/O section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"rows_deleted": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rows deleted from InnoDB tables.",
			},
			"rows_inserted": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rows inserted into InnoDB tables.",
			},
			"rows_read": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rows read from InnoDB tables.",
			},
			"rows_updated": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of rows updated in InnoDB tables.",
			},
			"s_lock_os_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the SEMAPHORES section of the `SHOW ENGINE INNODB STATUS` output",
			},
			"s_lock_spin_rounds": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the SEMAPHORES section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"s_lock_spin_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the SEMAPHORES section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"semaphore_wait_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Semaphore wait time",
			},
			"semaphore_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number semaphore currently being waited for by operations on InnoDB tables.",
			},
			"tables_in_use": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Tables in use",
			},
			"x_lock_os_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the SEMAPHORES section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"x_lock_spin_rounds": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the SEMAPHORES section of the `SHOW ENGINE INNODB STATUS` output.",
			},
			"x_lock_spin_waits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "As shown in the SEMAPHORES section of the `SHOW ENGINE INNODB STATUS` output.",
			},
		},
		Tags: map[string]interface{}{
			"server": &inputs.TagInfo{
				Desc: "Server addr",
			},
			"host": &inputs.TagInfo{
				Desc: "The server host address",
			},
		},
	}
}
