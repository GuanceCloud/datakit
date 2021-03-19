package mysqlmonitor

type MetricType int

const (
	COUNT MetricType = iota
	GAUGE
	RATE
	MONOTONIC
)

type MetricItem struct {
	name string,
	metricType MetricType,
	disable bool
}

type CollectType struct{
	metric map[string]*MetricItem,
	disable bool,
}

var metric = map[string]*Metric{
	"STATUS_VARS": &CollectType{
		# Command Metrics
	    metric: map[string]&MetricItem{
	    	"Slow_queries": &MetricItem{
	    		name: "mysql.performance.slow_queries",
	    		metricType: RATE,
	    	},
	    	"Questions": &MetricItem{
	    		name: "mysql.performance.questions",
	    		metricType: RATE,
	    	},
	    	"Queries": &MetricItem{
	    		name: "mysql.performance.queries",
	    		metricType: RATE,
	    	},
	    	"Com_select": &MetricItem{
		    	name: "mysql.performance.com_select",
		    	metricType: RATE,
	    	},
		    "Com_insert": &MetricItem{
		    	name: "mysql.performance.com_insert",
		    	metricType: RATE,
		    },
		    "Com_update": &MetricItem{
		    	name: "mysql.performance.com_update",
		    	metricType: RATE,
		    },
		    "Com_delete": &MetricItem{
		    	name: "mysql.performance.com_delete",
		    	metricType: RATE,
		    },
		    "Com_replace": &MetricItem{
		    	name: "mysql.performance.com_replace",
		    	metricType: RATE,
		    },
		    "Com_load": &MetricItem{
		    	name: "mysql.performance.com_load",
		    	metricType: RATE,
		    },
		    "Com_insert_select": &MetricItem{
		    	name: "mysql.performance.com_insert_select",
		    	metricType: RATE,
		    },
		    "Com_update_multi": &MetricItem{
		    	name: "mysql.performance.com_update_multi",
		    	metricType: RATE,
		    },
		    "Com_delete_multi": &MetricItem{
		    	name: "mysql.performance.com_delete_multi",
		    	metricType: RATE,
		    },
		    "Com_replace_select": &MetricItem{
		    	name: "mysql.performance.com_replace_select",
		    	metricType: RATE,
		    },
		    # Connection Metrics
		    "Connections": &MetricItem{
		    	name: "mysql.net.connections",
		    	metricType: RATE,
		    },
		    "Max_used_connections": &MetricItem{
		    	name: "mysql.net.max_connections",
		    	metricType: GAUGE,
		    },
		    "Aborted_clients": &MetricItem{
		    	name: "mysql.net.aborted_clients",
		    	metricType: RATE,
		    },
		    "Aborted_connects": &MetricItem{
		    	name: "mysql.net.aborted_connects",
		    	metricType: RATE,
		    },
		    # Table Cache Metrics
		    "Open_files": &MetricItem{
		    	name: "mysql.performance.open_files",
		    	metricType: GAUGE,
		    },
		    "Open_tables": &MetricItem{
		    	name: "mysql.performance.open_tables",
		    	metricType: GAUGE,
		    },
		    # Network Metrics
		    "Bytes_sent": &MetricItem{
		    	name: "mysql.performance.bytes_sent",
		    	metricType: RATE,
		    },
		    "Bytes_received": &MetricItem{
		    	name: "mysql.performance.bytes_received",
		    	metricType: RATE,
		    },
		    # Query Cache Metrics
		    "Qcache_hits": &MetricItem{
		    	name: "mysql.performance.qcache_hits",
		    	metricType: RATE,
		    },
		    "Qcache_inserts": &MetricItem{
		    	name: "mysql.performance.qcache_inserts",
		    	metricType: RATE,
		    },
		    "Qcache_lowmem_prunes": &MetricItem{
		    	name: "mysql.performance.qcache_lowmem_prunes",
		    	metricType: RATE,
		    },
		    # Table Lock Metrics
		    "Table_locks_waited": &MetricItem{
		    	name: "mysql.performance.table_locks_waited",
		    	metricType: GAUGE,
		    },
		    "Table_locks_waited_rate": &MetricItem{
		    	name: "mysql.performance.table_locks_waited.rate",
		    	metricType: RATE,
		    },
		    # Temporary Table Metrics
		    "Created_tmp_tables": &MetricItem{
		    	name: "mysql.performance.created_tmp_tables",
		    	metricType: RATE,
		    },
		    "Created_tmp_disk_tables": &MetricItem{
		    	name: "mysql.performance.created_tmp_disk_tables",
		    	metricType: RATE,
		    },
		    "Created_tmp_files": &MetricItem{
		    	name: "mysql.performance.created_tmp_files",
		    	metricType: RATE,
		    },
		    # Thread Metrics
		    "Threads_connected": &MetricItem{
		    	name: "mysql.performance.threads_connected",
		    	metricType: GAUGE,
		    },
		    "Threads_running": &MetricItem{
		    	name: "mysql.performance.threads_running",
		    	metricType: GAUGE,
		    },
		    # MyISAM Metrics
		    "Key_buffer_bytes_unflushed": &MetricItem{
		    	name: "mysql.myisam.key_buffer_bytes_unflushed",
		    	metricType: GAUGE,
		    },
		    "Key_buffer_bytes_used": &MetricItem{
		    	name: "mysql.myisam.key_buffer_bytes_used",
		    	metricType: GAUGE,
		    },
		    "Key_read_requests": &MetricItem{
		    	name: "mysql.myisam.key_read_requests",
		    	metricType: RATE,
		    },
		    "Key_reads": &MetricItem{
		    	name: "mysql.myisam.key_reads",
		    	metricType: RATE,
		    },
		    "Key_write_requests": &MetricItem{
		    	name: "mysql.myisam.key_write_requests",
		    	metricType: RATE,
		    },
		    "Key_writes": &MetricItem{
		    	name: "mysql.myisam.key_writes",
		    	metricType: RATE,
		    },
		},
		disable: false,
	},
	"VARIABLES_VARS": &CollectType{
		metric: map[string]&MetricItem{
			"Key_buffer_size": &MetricItem{
				name: "mysql.myisam.key_buffer_size",
				metricType: GAUGE,
			},
		    "Key_cache_utilization": &MetricItem{
		    	name: "mysql.performance.key_cache_utilization",
		    	metricType: GAUGE,
		    },
		    "max_connections": &MetricItem{
		    	name: "mysql.net.max_connections_available",
		    	metricType: GAUGE,
		    },
		    "query_cache_size": &MetricItem{
		    	name: "mysql.performance.qcache_size",
		    	metricType: GAUGE,
		    },
		    "table_open_cache": &MetricItem{
		    	name: "mysql.performance.table_open_cache",
		    	metricType: GAUGE,
		    },
		    "thread_cache_size": &MetricItem{
		    	name: "mysql.performance.thread_cache_size",
		    	metricType: GAUGE,
		    },
		},
		disable: true,
	},
	"INNODB_VARS": &CollectType{
		metric: map[string]&MetricItem{
			"Innodb_data_reads": &MetricItem{
				name: "mysql.innodb.data_reads",
				metricType: RATE,
			},
		    "Innodb_data_writes": &MetricItem{
		    	name: "mysql.innodb.data_writes",
		    	metricType: RATE,
		    },
		    "Innodb_os_log_fsyncs": &MetricItem{
		    	name: "mysql.innodb.os_log_fsyncs",
		    	metricType: RATE,
		    },
		    "Innodb_mutex_spin_waits": &MetricItem{
		    	name: "mysql.innodb.mutex_spin_waits",
		    	metricType: RATE,
		    },
		    "Innodb_mutex_spin_rounds": &MetricItem{
		    	name: "mysql.innodb.mutex_spin_rounds",
		    	metricType: RATE,
		    },
		    "Innodb_mutex_os_waits": &MetricItem{
		    	name: "mysql.innodb.mutex_os_waits",
		    	metricType: RATE,
		    },
		    "Innodb_row_lock_waits": &MetricItem{
		    	name: "mysql.innodb.row_lock_waits",
		    	metricType: RATE,
		    },
		    "Innodb_row_lock_time": &MetricItem{
		    	name: "mysql.innodb.row_lock_time",
		    	metricType: RATE,
		    },
		    "Innodb_row_lock_current_waits": &MetricItem{
		    	name: "mysql.innodb.row_lock_current_waits",
		    	metricType: GAUGE,
		    },
		    "Innodb_current_row_locks": &MetricItem{
		    	name: "mysql.innodb.current_row_locks",
		    	metricType: GAUGE,
		    },
		    "Innodb_buffer_pool_bytes_dirty": &MetricItem{
		    	name: "mysql.innodb.buffer_pool_dirty",
		    	metricType: GAUGE,
		    },
		    "Innodb_buffer_pool_bytes_free": &MetricItem{
		    	name: "mysql.innodb.buffer_pool_free",
		    	metricType: GAUGE,
		    },
		    "Innodb_buffer_pool_bytes_used": &MetricItem{
		    	name: "mysql.innodb.buffer_pool_used",
		    	metricType: GAUGE,
		    },
		    "Innodb_buffer_pool_bytes_total": &MetricItem{
		    	name: "mysql.innodb.buffer_pool_total",
		    	metricType: GAUGE,
		    },
		    "Innodb_buffer_pool_read_requests": &MetricItem{
		    	name: "mysql.innodb.buffer_pool_read_requests",
		    	metricType: RATE,
		    },
		    "Innodb_buffer_pool_reads": &MetricItem{
		    	name: "mysql.innodb.buffer_pool_reads",
		    	metricType: RATE,
		    },
		    "Innodb_buffer_pool_pages_utilization": &MetricItem{
		    	name: "mysql.innodb.buffer_pool_utilization",
		    	metricType: GAUGE,
		    },
		},
		disable: true,
	},
	"BINLOG_VARS": &CollectType{
		metric: map[string]&MetricItem{
			"Binlog_space_usage_bytes": &MetricItem{
				name: "mysql.binlog.disk_use",
				metricType: GAUGE,
			},
			disable: true,
		},
	},
	"OPTIONAL_STATUS_VARS": &CollectType{
		metric: map[string]&MetricItem{
			"Binlog_cache_disk_use": &MetricItem{
				name: "mysql.binlog.cache_disk_use",
				metricType: GAUGE
			},
		    "Binlog_cache_use": &MetricItem{
		    	name: "mysql.binlog.cache_use",
		    	metricType: GAUGE
		    },
		    "Handler_commit": &MetricItem{
		    	name: "mysql.performance.handler_commit",
		    	metricType: RATE
		    },
		    "Handler_delete": &MetricItem{
		    	name: "mysql.performance.handler_delete",
		    	metricType: RATE
		    },
		    "Handler_prepare": &MetricItem{
		    	name: "mysql.performance.handler_prepare",
		    	metricType: RATE
		    },
		    "Handler_read_first": &MetricItem{
		    	name: "mysql.performance.handler_read_first",
		    	metricType: RATE
		    },
		    "Handler_read_key": &MetricItem{
		    	name: "mysql.performance.handler_read_key",
		    	metricType: RATE
		    },
		    "Handler_read_next": &MetricItem{
		    	name: "mysql.performance.handler_read_next",
		    	metricType: RATE
		    },
		    "Handler_read_prev": &MetricItem{
		    	name: "mysql.performance.handler_read_prev",
		    	metricType: RATE
		    },
		    "Handler_read_rnd": &MetricItem{
		    	name: "mysql.performance.handler_read_rnd",
		    	metricType: RATE
		    },
		    "Handler_read_rnd_next": &MetricItem{
		    	name: "mysql.performance.handler_read_rnd_next",
		    	metricType: RATE
		    },
		    "Handler_rollback": &MetricItem{
		    	name: "mysql.performance.handler_rollback",
		    	metricType: RATE
		    },
		    "Handler_update": &MetricItem{
		    	name: "mysql.performance.handler_update",
		    	metricType: RATE
		    },
		    "Handler_write": &MetricItem{
		    	name: "mysql.performance.handler_write",
		    	metricType: RATE
		    },
		    "Opened_tables": &MetricItem{
		    	name: "mysql.performance.opened_tables",
		    	metricType: RATE
		    },
		    "Qcache_total_blocks": &MetricItem{
		    	name: "mysql.performance.qcache_total_blocks",
		    	metricType: GAUGE
		    },
		    "Qcache_free_blocks": &MetricItem{
		    	name: "mysql.performance.qcache_free_blocks",
		    	metricType: GAUGE
		    },
		    "Qcache_free_memory": &MetricItem{
		    	name: "mysql.performance.qcache_free_memory",
		    	metricType: GAUGE
		    },
		    "Qcache_not_cached": &MetricItem{
		    	name: "mysql.performance.qcache_not_cached",
		    	metricType: RATE
		    },
		    "Qcache_queries_in_cache": &MetricItem{
		    	name: "mysql.performance.qcache_queries_in_cache",
		    	metricType: GAUGE
		    },
		    "Select_full_join": &MetricItem{
		    	name: "mysql.performance.select_full_join",
		    	metricType: RATE
		    },
		    "Select_full_range_join": &MetricItem{
		    	name: "mysql.performance.select_full_range_join",
		    	metricType: RATE
		    },
		    "Select_range": &MetricItem{
		    	name: "mysql.performance.select_range",
		    	metricType: RATE
		    },
		    "Select_range_check": &MetricItem{
		    	name: "mysql.performance.select_range_check",
		    	metricType: RATE
		    },
		    "Select_scan": &MetricItem{
		    	name: "mysql.performance.select_scan",
		    	metricType: RATE
		    },
		    "Sort_merge_passes": &MetricItem{
		    	name: "mysql.performance.sort_merge_passes",
		    	metricType: RATE
		    },
		    "Sort_range": &MetricItem{
		    	name: "mysql.performance.sort_range",
		    	metricType: RATE
		    },
		    "Sort_rows": &MetricItem{
		    	name: "mysql.performance.sort_rows",
		    	metricType: RATE
		    },
		    "Sort_scan": &MetricItem{
		    	name: "mysql.performance.sort_scan",
		    	metricType: RATE
		    },
		    "Table_locks_immediate": &MetricItem{
		    	name: "mysql.performance.table_locks_immediate",
		    	metricType: GAUGE
		    },
		    "Table_locks_immediate_rate": &MetricItem{
		    	name: "mysql.performance.table_locks_immediate.rate",
		    	metricType: RATE
		    },
		    "Threads_cached": &MetricItem{
		    	name: "mysql.performance.threads_cached",
		    	metricType: GAUGE
		    },
		    "Threads_created": &MetricItem{
		    	name: "mysql.performance.threads_created",
		    	metricType: MONOTONIC
		    },
		},
	    disable: true,
	"OPTIONAL_STATUS_VARS_5_6_6": &CollectType{
		metric: map[string]&MetricItem{
		    "Table_open_cache_hits": &MetricItem{
		    	name: "mysql.performance.table_cache_hits",
		    	metricType: RATE,
		    },
		    "Table_open_cache_misses": &MetricItem{
		    	name: "mysql.performance.table_cache_misses",
		    	metricType: RATE,
		    },
		},
	    disable: true,
	},
	"OPTIONAL_INNODB_VARS": {
		metric: map[string]&MetricItem{
		    "Innodb_active_transactions": &MetricItem{
		    	name: "mysql.innodb.active_transactions",
		    	metricType: GAUGE,
		    },
		    "Innodb_buffer_pool_bytes_data": &MetricItem{
		    	name: "mysql.innodb.buffer_pool_data",
		    	metricType: GAUGE,
		    },
		    "Innodb_buffer_pool_pages_data": &MetricItem{
		    	name: "mysql.innodb.buffer_pool_pages_data",
		    	metricType: GAUGE,
		    },
		    "Innodb_buffer_pool_pages_dirty": &MetricItem{
		    	name: "mysql.innodb.buffer_pool_pages_dirty",
		    	metricType: GAUGE,
		    },
		    "Innodb_buffer_pool_pages_flushed": &MetricItem{
		    	name: "mysql.innodb.buffer_pool_pages_flushed",
		    	metricType: RATE,
		    },
		    "Innodb_buffer_pool_pages_free": &MetricItem{
		    	name: "mysql.innodb.buffer_pool_pages_free",
		    	metricType: GAUGE,
		    },
		    "Innodb_buffer_pool_pages_total": &MetricItem{
		    	name: "mysql.innodb.buffer_pool_pages_total",
		    	metricType: GAUGE,
		    },
		    "Innodb_buffer_pool_read_ahead": &MetricItem{
		    	name: "mysql.innodb.buffer_pool_read_ahead",
		    	metricType: RATE,
		    },
		    "Innodb_buffer_pool_read_ahead_evicted": &MetricItem{
		    	name: "mysql.innodb.buffer_pool_read_ahead_evicted",
		    	metricType: RATE,
		    },
		    "Innodb_buffer_pool_read_ahead_rnd": &MetricItem{
		    	name: "mysql.innodb.buffer_pool_read_ahead_rnd",
		    	metricType: GAUGE,
		    },
		    "Innodb_buffer_pool_wait_free": &MetricItem{
		    	name: "mysql.innodb.buffer_pool_wait_free",
		    	metricType: MONOTONIC,
		    },
		    "Innodb_buffer_pool_write_requests": &MetricItem{
		    	name: "mysql.innodb.buffer_pool_write_requests",
		    	metricType: RATE,
		    },
		    "Innodb_checkpoint_age": &MetricItem{
		    	name: "mysql.innodb.checkpoint_age",
		    	metricType: GAUGE,
		    },
		    "Innodb_current_transactions": &MetricItem{
		    	name: "mysql.innodb.current_transactions",
		    	metricType: GAUGE,
		    },
		    "Innodb_data_fsyncs": &MetricItem{
		    	name: "mysql.innodb.data_fsyncs",
		    	metricType: RATE,
		    },
		    "Innodb_data_pending_fsyncs": &MetricItem{
		    	name: "mysql.innodb.data_pending_fsyncs",
		    	metricType: GAUGE,
		    },
		    "Innodb_data_pending_reads": &MetricItem{
		    	name: "mysql.innodb.data_pending_reads",
		    	metricType: GAUGE,
		    },
		    "Innodb_data_pending_writes": &MetricItem{
		    	name: "mysql.innodb.data_pending_writes",
		    	metricType: GAUGE,
		    },
		    "Innodb_data_read": &MetricItem{
		    	name: "mysql.innodb.data_read",
		    	metricType: RATE,
		    },
		    "Innodb_data_written": &MetricItem{
		    	name: "mysql.innodb.data_written",
		    	metricType: RATE,
		    },
		    "Innodb_dblwr_pages_written": &MetricItem{
		    	name: "mysql.innodb.dblwr_pages_written",
		    	metricType: RATE,
		    },
		    "Innodb_dblwr_writes": &MetricItem{
		    	name: "mysql.innodb.dblwr_writes",
		    	metricType: RATE,
		    },
		    "Innodb_hash_index_cells_total": &MetricItem{
		    	name: "mysql.innodb.hash_index_cells_total",
		    	metricType: GAUGE,
		    },
		    "Innodb_hash_index_cells_used": &MetricItem{
		    	name: "mysql.innodb.hash_index_cells_used",
		    	metricType: GAUGE,
		    },
		    "Innodb_history_list_length": &MetricItem{
		    	name: "mysql.innodb.history_list_length",
		    	metricType: GAUGE,
		    },
		    "Innodb_ibuf_free_list": &MetricItem{
		    	name: "mysql.innodb.ibuf_free_list",
		    	metricType: GAUGE,
		    },
		    "Innodb_ibuf_merged": &MetricItem{
		    	name: "mysql.innodb.ibuf_merged",
		    	metricType: RATE,
		    },
		    "Innodb_ibuf_merged_delete_marks": &MetricItem{
		    	name: "mysql.innodb.ibuf_merged_delete_marks",
		    	metricType: RATE,
		    },
		    "Innodb_ibuf_merged_deletes": &MetricItem{
		    	name: "mysql.innodb.ibuf_merged_deletes",
		    	metricType: RATE,
		    },
		    "Innodb_ibuf_merged_inserts": &MetricItem{
		    	name: "mysql.innodb.ibuf_merged_inserts",
		    	metricType: RATE,
		    },
		    "Innodb_ibuf_merges": &MetricItem{
		    	name: "mysql.innodb.ibuf_merges",
		    	metricType: RATE,
		    },
		    "Innodb_ibuf_segment_size": &MetricItem{
		    	name: "mysql.innodb.ibuf_segment_size",
		    	metricType: GAUGE,
		    },
		    "Innodb_ibuf_size": &MetricItem{
		    	name: "mysql.innodb.ibuf_size",
		    	metricType: GAUGE,
		    },
		    "Innodb_lock_structs": &MetricItem{
		    	name: "mysql.innodb.lock_structs",
		    	metricType: RATE,
		    },
		    "Innodb_locked_tables": &MetricItem{
		    	name: "mysql.innodb.locked_tables",
		    	metricType: GAUGE,
		    },
		    "Innodb_locked_transactions": &MetricItem{
		    	name: "mysql.innodb.locked_transactions",
		    	metricType: GAUGE,
		    },
		    "Innodb_log_waits": &MetricItem{
		    	name: "mysql.innodb.log_waits",
		    	metricType: RATE,
		    },
		    "Innodb_log_write_requests": &MetricItem{
		    	name: "mysql.innodb.log_write_requests",
		    	metricType: RATE,
		    },
		    "Innodb_log_writes": &MetricItem{
		    	name: "mysql.innodb.log_writes",
		    	metricType: RATE,
		    },
		    "Innodb_lsn_current": &MetricItem{
		    	name: "mysql.innodb.lsn_current",
		    	metricType: RATE,
		    },
		    "Innodb_lsn_flushed": &MetricItem{
		    	name: "mysql.innodb.lsn_flushed",
		    	metricType: RATE,
		    },
		    "Innodb_lsn_last_checkpoint": &MetricItem{
		    	name: "mysql.innodb.lsn_last_checkpoint",
		    	metricType: RATE,
		    },
		    "Innodb_mem_adaptive_hash": &MetricItem{
		    	name: "mysql.innodb.mem_adaptive_hash",
		    	metricType: GAUGE,
		    },
		    "Innodb_mem_additional_pool": &MetricItem{
		    	name: "mysql.innodb.mem_additional_pool",
		    	metricType: GAUGE,
		    },
		    "Innodb_mem_dictionary": &MetricItem{
		    	name: "mysql.innodb.mem_dictionary",
		    	metricType: GAUGE,
		    },
		    "Innodb_mem_file_system": &MetricItem{
		    	name: "mysql.innodb.mem_file_system",
		    	metricType: GAUGE,
		    },
		    "Innodb_mem_lock_system": &MetricItem{
		    	name: "mysql.innodb.mem_lock_system",
		    	metricType: GAUGE,
		    },
		    "Innodb_mem_page_hash": &MetricItem{
		    	name: "mysql.innodb.mem_page_hash",
		    	metricType: GAUGE,
		    },
		    "Innodb_mem_recovery_system": &MetricItem{
		    	name: "mysql.innodb.mem_recovery_system",
		    	metricType: GAUGE,
		    },
		    "Innodb_mem_thread_hash": &MetricItem{
		    	name: "mysql.innodb.mem_thread_hash",
		    	metricType: GAUGE,
		    },
		    "Innodb_mem_total": &MetricItem{
		    	name: "mysql.innodb.mem_total",
		    	metricType: GAUGE,
		    },
		    "Innodb_os_file_fsyncs": &MetricItem{
		    	name: "mysql.innodb.os_file_fsyncs",
		    	metricType: RATE,
		    },
		    "Innodb_os_file_reads": &MetricItem{
		    	name: "mysql.innodb.os_file_reads",
		    	metricType: RATE,
		    },
		    "Innodb_os_file_writes": &MetricItem{
		    	name: "mysql.innodb.os_file_writes",
		    	metricType: RATE,
		    },
		    "Innodb_os_log_pending_fsyncs": &MetricItem{
		    	name: "mysql.innodb.os_log_pending_fsyncs",
		    	metricType: GAUGE,
		    },
		    "Innodb_os_log_pending_writes": &MetricItem{
		    	name: "mysql.innodb.os_log_pending_writes",
		    	metricType: GAUGE,
		    },
		    "Innodb_os_log_written": &MetricItem{
		    	name: "mysql.innodb.os_log_written",
		    	metricType: RATE,
		    },
		    "Innodb_pages_created": &MetricItem{
		    	name: "mysql.innodb.pages_created",
		    	metricType: RATE,
		    },
		    "Innodb_pages_read": &MetricItem{
		    	name: "mysql.innodb.pages_read",
		    	metricType: RATE,
		    },
		    "Innodb_pages_written": &MetricItem{
		    	name: "mysql.innodb.pages_written",
		    	metricType: RATE,
		    },
		    "Innodb_pending_aio_log_ios": &MetricItem{
		    	name: "mysql.innodb.pending_aio_log_ios",
		    	metricType: GAUGE,
		    },
		    "Innodb_pending_aio_sync_ios": &MetricItem{
		    	name: "mysql.innodb.pending_aio_sync_ios",
		    	metricType: GAUGE,
		    },
		    "Innodb_pending_buffer_pool_flushes": &MetricItem{
		    	name: "mysql.innodb.pending_buffer_pool_flushes",
		    	metricType: GAUGE,
		    },
		    "Innodb_pending_checkpoint_writes": &MetricItem{
		    	name: "mysql.innodb.pending_checkpoint_writes",
		    	metricType: GAUGE,
		    },
		    "Innodb_pending_ibuf_aio_reads": &MetricItem{
		    	name: "mysql.innodb.pending_ibuf_aio_reads",
		    	metricType: GAUGE,
		    },
		    "Innodb_pending_log_flushes": &MetricItem{
		    	name: "mysql.innodb.pending_log_flushes",
		    	metricType: GAUGE,
		    },
		    "Innodb_pending_log_writes": &MetricItem{
		    	name: "mysql.innodb.pending_log_writes",
		    	metricType: GAUGE,
		    },
		    "Innodb_pending_normal_aio_reads": &MetricItem{
		    	name: "mysql.innodb.pending_normal_aio_reads",
		    	metricType: GAUGE,
		    },
		    "Innodb_pending_normal_aio_writes": &MetricItem{
		    	name: "mysql.innodb.pending_normal_aio_writes",
		    	metricType: GAUGE,
		    },
		    "Innodb_queries_inside": &MetricItem{
		    	name: "mysql.innodb.queries_inside",
		    	metricType: GAUGE,
		    },
		    "Innodb_queries_queued": &MetricItem{
		    	name: "mysql.innodb.queries_queued",
		    	metricType: GAUGE,
		    },
		    "Innodb_read_views": &MetricItem{
		    	name: "mysql.innodb.read_views",
		    	metricType: GAUGE,
		    },
		    "Innodb_rows_deleted": &MetricItem{
		    	name: "mysql.innodb.rows_deleted",
		    	metricType: RATE,
		    },
		    "Innodb_rows_inserted": &MetricItem{
		    	name: "mysql.innodb.rows_inserted",
		    	metricType: RATE,
		    },
		    "Innodb_rows_read": &MetricItem{
		    	name: "mysql.innodb.rows_read",
		    	metricType: RATE,
		    },
		    "Innodb_rows_updated": &MetricItem{
		    	name: "mysql.innodb.rows_updated",
		    	metricType: RATE,
		    },
		    "Innodb_s_lock_os_waits": &MetricItem{
		    	name: "mysql.innodb.s_lock_os_waits",
		    	metricType: RATE,
		    },
		    "Innodb_s_lock_spin_rounds": &MetricItem{
		    	name: "mysql.innodb.s_lock_spin_rounds",
		    	metricType: RATE,
		    },
		    "Innodb_s_lock_spin_waits": &MetricItem{
		    	name: "mysql.innodb.s_lock_spin_waits",
		    	metricType: RATE,
		    },
		    "Innodb_semaphore_wait_time": &MetricItem{
		    	name: "mysql.innodb.semaphore_wait_time",
		    	metricType: GAUGE,
		    },
		    "Innodb_semaphore_waits": &MetricItem{
		    	name: "mysql.innodb.semaphore_waits",
		    	metricType: GAUGE,
		    },
		    "Innodb_tables_in_use": &MetricItem{
		    	name: "mysql.innodb.tables_in_use",
		    	metricType: GAUGE,
		    },
		    "Innodb_x_lock_os_waits": &MetricItem{
		    	name: "mysql.innodb.x_lock_os_waits",
		    	metricType: RATE,
		    },
		    "Innodb_x_lock_spin_rounds": &MetricItem{
		    	name: "mysql.innodb.x_lock_spin_rounds",
		    	metricType: RATE,
		    },
		    "Innodb_x_lock_spin_waits": &MetricItem{
		    	name: "mysql.innodb.x_lock_spin_waits",
		    	metricType: RATE,
		    },
		},
	    disable: true,
	},
	"GALERA_VARS": &CollectType{
		metric: map[string]&MetricItem{
			"wsrep_cluster_size": &MetricItem{
		    	name: "mysql.galera.wsrep_cluster_size",
		    	metricType: GAUGE,
		    },
		    "wsrep_local_recv_queue_avg": &MetricItem{
		    	name: "mysql.galera.wsrep_local_recv_queue_avg",
		    	metricType: GAUGE,
		    },
		    "wsrep_flow_control_paused": &MetricItem{
		    	name: "mysql.galera.wsrep_flow_control_paused",
		    	metricType: GAUGE,
		    },
		    "wsrep_flow_control_paused_ns": &MetricItem{
		    	name: "mysql.galera.wsrep_flow_control_paused_ns",
		    	metricType: MONOTONIC,
		    },
		    "wsrep_flow_control_recv": &MetricItem{
		    	name: "mysql.galera.wsrep_flow_control_recv",
		    	metricType: MONOTONIC,
		    },
		    "wsrep_flow_control_sent": &MetricItem{
		    	name: "mysql.galera.wsrep_flow_control_sent",
		    	metricType: MONOTONIC,
		    },
		    "wsrep_cert_deps_distance": &MetricItem{
		    	name: "mysql.galera.wsrep_cert_deps_distance",
		    	metricType: GAUGE,
		    },
		    "wsrep_local_send_queue_avg": &MetricItem{
		    	name: "mysql.galera.wsrep_local_send_queue_avg",
		    	metricType: GAUGE,
		    },
		},
		disable: true,
	},
	"PERFORMANCE_VARS": &CollectType{
		metric: map[string]&MetricItem{
			"query_run_time_avg": &MetricItem{
		    	name: "mysql.performance.query_run_time.avg",
		    	metricType: GAUGE,
		    },
		    "perf_digest_95th_percentile_avg_us": &MetricItem{
		    	name: "mysql.performance.digest_95th_percentile.avg_us",
		    	metricType: GAUGE,
		    },
		},
		disable: true,
	},
	"SCHEMA_VARS": {
		metric: map[string]&MetricItem{
			"information_schema_size": &MetricItem{
				name: "mysql.info.schema.size",
				metricType: GAUGE,
			},
		},
		disable: true,
	},
	"REPLICA_VARS": {
		metric: map[string]&MetricItem{
			"Seconds_Behind_Master": &MetricItem{
	    		name: "mysql.replication.seconds_behind_master",
	    		metricType: GAUGE,
	    	},
	    	"Replicas_connected": &MetricItem{
	        	name: "mysql.replication.replicas_connected",
	        	metricType: GAUGE,
	        },
	    },
        disable: true,
	},
}


