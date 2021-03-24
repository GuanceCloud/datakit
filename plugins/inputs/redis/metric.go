package redis

type Metric struct {
	metricName string
	parse parse
	disable bool
	plan string
}

var metrics = map[string]*Metric{
	// Append-only metrics
		"aof_last_rewrite_time_sec": &Metric{
			metricName: "aof_last_rewrite_time_sec",
			plan: "todo",
		},
		"aof_rewrite_in_progress": &Metric{
			metricName: "aof_rewrite_in_progress",
			plan: "todo",
		},
		"aof_current_size":&Metric{
			metricName: "aof_current_size",
			plan: "todo",
		},
		"aof_buffer_length": &Metric{
			metricName: "aof_buffer_length",
			plan: "todo",
		},

		// Network
		"connected_clients": &Metric{
			metricName: "connected_clients",
			plan: "todo",
		},
		"connected_slaves": &Metric{
			metricName: "connected_slaves",
			plan: "todo",
		},
		"rejected_connections": &Metric{
			metricName: "rejected_connections",
			plan: "todo",
		},

		// clients
		"blocked_clients": &Metric{
			metricName: "blocked_clients",
			plan: "todo",
		},
		"client_biggest_input_buf": &Metric{
			metricName: "client_biggest_input_buf",
			plan: "todo",
		},
		"client_longest_output_list": &Metric{
			metricName: "client_longest_output_list",
			plan: "todo",
		},

		// Keys
		"evicted_keys": &Metric{
			metricName: "evicted_keys",
			plan: "todo",
		},
		"expired_keys": &Metric{
			metricName: "expired_keys",
			plan: "todo",
		},

		// stats
		"latest_fork_usec": &Metric{
			metricName: "latest_fork_usec",
			plan: "todo",
		},

		// pubsub
		"pubsub_channels": &Metric{
			metricName: "pubsub_channels",
			plan: "todo",
		},
		"pubsub_patterns": &Metric{
			metricName: "pubsub_patterns",
			plan: "todo",
		},

		// rdb
		"rdb_bgsave_in_progress": &Metric{
			metricName: "rdb_bgsave_in_progress",
			plan: "todo",
		},
		"rdb_changes_since_last_save": &Metric{
			metricName: "rdb_changes_since_last_save",
			plan: "todo",
		},
		"rdb_last_bgsave_time_sec": &Metric{
			metricName: "rdb_last_bgsave_time_sec",
			plan: "todo",
		},

		// memory
		"mem_fragmentation_ratio": &Metric{
			metricName: "mem_fragmentation_ratio",
			plan: "todo",
		},
		"used_memory": &Metric{
			metricName: "used_memory",
			plan: "todo",
		},
		"used_memory_lua": &Metric{
			metricName: "used_memory_lua",
			plan: "todo",
		},
		"used_memory_peak": &Metric{
			metricName: "used_memory_peak",
			plan: "todo",
		},
		"used_memory_rss": &Metric{
			metricName: "used_memory_rss",
			plan: "todo",
		},

		// replication
		"master_last_io_seconds_ago": &Metric{
			metricName: "master_last_io_seconds_ago",
			plan: "todo",
		},
		"master_sync_in_progress": &Metric{
			metricName: "master_sync_in_progress",
			plan: "todo",
		},
		"master_sync_left_bytes": &Metric{
			metricName: "master_sync_left_bytes",
			plan: "todo",
		},
		"repl_backlog_histlen": &Metric{
			metricName: "repl_backlog_histlen",
			plan: "todo",
		},
		"master_repl_offset": &Metric{
			metricName: "master_repl_offset",
			plan: "todo",
		},
		"slave_repl_offset": &Metric{
			metricName: "slave_repl_offset",
			plan: "todo",
		},
		"used_cpu_sys": &Metric{
			metricName: "used_cpu_sys",
			plan: "todo",
		},
		"used_cpu_sys_children": &Metric{
			metricName: "used_cpu_sys_children",
			plan: "todo",
		},
		"used_cpu_user": &Metric{
			metricName: "used_cpu_user",
			plan: "todo",
		},
		"used_cpu_user_children": &Metric{
			metricName: "used_cpu_user_children",
			plan: "todo",
		},

		// stats
		"total_commands_processed": &Metric{
			metricName: "total_commands_processed",
			plan: "todo",
		},
		"keyspace_hits": &Metric{
			metricName: "keyspace_hits",
			plan: "todo",
		},
		"keyspace_misses": &Metric{
			metricName: "keyspace_misses",
			plan: "todo",
		},
	}

