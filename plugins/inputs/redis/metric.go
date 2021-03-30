package redis

import (
	"github.com/spf13/cast"
)

type MetricItem struct {
	name string
	value interface{}
	disable bool
	desc string
	parse parse
	plan string
}

type MetricType struct{
	metricSet map[string]*MetricItem // 指标集
	tags map[string]string
	disable bool // 状态
	desc string //描述
}

type parse func(val interface{}) (interface{})

func parseInt(val interface{}) (interface{}) {
	return cast.ToInt64(val)
}

func parseStr(val interface{}) (interface{}) {
	return cast.ToString(val)
}

func parseFloat64(val interface{}) (interface{}) {
	return cast.ToFloat64(val)
}

var metrics = map[string]*MetricType{
	"info": &MetricType{
		metricSet: map[string]*MetricItem{
			"info_latency_ms": &MetricItem{
				name: "info_latency_ms",
				plan: "result",
				parse: parseFloat64,
			},
	        "active_defrag_running": &MetricItem{
				name: "active_defrag_running",
				plan: "result",
				parse: parseInt,
			},
			"redis_version": &MetricItem{
				name: "redis_version",
				plan: "result",
				parse: parseStr,
			},
	        "active_defrag_hits": &MetricItem{
				name: "active_defrag_hits",
				plan: "result",
				parse: parseInt,
			},
	        "active_defrag_misses": &MetricItem{
				name: "active_defrag_misses",
				plan: "result",
				parse: parseInt,
			},
	        "active_defrag_key_hits": &MetricItem{
				name: "active_defrag_key_hits",
				plan: "result",
				parse: parseInt,
			},
	        "active_defrag_key_misses": &MetricItem{
				name: "active_defrag_key_misses",
				plan: "result",
				parse: parseInt,
			},
	        "aof_last_rewrite_time_sec": &MetricItem{
				name: "aof_last_rewrite_time_sec",
				plan: "result",
				parse: parseInt,
			},
	        "aof_rewrite_in_progress": &MetricItem{
				name: "aof_rewrite_in_progress",
				plan: "result",
				parse: parseInt,
			},
	        "aof_current_size": &MetricItem{
				name: "aof_current_size",
				plan: "todo",
				parse: parseInt,
				desc: "aof未开启",
			},
	        "aof_buffer_length": &MetricItem{
				name: "aof_buffer_length",
				plan: "todo",
				parse: parseInt,
				desc: "aof未开启",
			},
	        "loading_total_bytes": &MetricItem{
				name: "loading_total_bytes",
				plan: "todo",
				parse: parseInt,
				desc: "aof未开启",
			},
	        "loading_loaded_bytes": &MetricItem{
				name: "loading_loaded_bytes",
				plan: "todo",
				parse: parseInt,
				desc: "aof未开启",
			},
	        "loading_loaded_perc": &MetricItem{
				name: "loading_loaded_perc",
				plan: "todo",
				parse: parseInt,
				desc: "aof未开启",
			},
	        "loading_eta_seconds": &MetricItem{
				name: "loading_eta_seconds",
				plan: "todo",
				parse: parseInt,
				desc: "aof未开启",
			},
	        "connected_clients": &MetricItem{
				name: "connected_clients",
				plan: "result",
				parse: parseInt,
			},
	        "connected_slaves": &MetricItem{
				name: "connected_slaves",
				plan: "result",
				parse: parseInt,
			},
	        "rejected_connections": &MetricItem{
				name: "rejected_connections",
				plan: "result",
				parse: parseInt,
			},
	        "blocked_clients": &MetricItem{
				name: "blocked_clients",
				plan: "result",
				parse: parseInt,
			},
	        "client_biggest_input_buf": &MetricItem{
				name: "client_biggest_input_buf",
				plan: "todo",
				parse: parseInt,
			},
	        "client_longest_output_list": &MetricItem{
				name: "client_longest_output_list",
				plan: "todo",
				parse: parseInt,
			},
	        "evicted_keys": &MetricItem{
				name: "evicted_keys",
				plan: "result",
				parse: parseInt,
			},
	        "expired_keys": &MetricItem{
				name: "expired_keys",
				plan: "result",
				parse: parseInt,
			},
	        "latest_fork_usec": &MetricItem{
				name: "latest_fork_usec",
				plan: "result",
				parse: parseInt,
			},
	        "bytes_received_per_sec": &MetricItem{
				name: "bytes_received_per_sec",
				plan: "todo",
				parse: parseInt,
			},
	        "bytes_sent_per_sec": &MetricItem{
				name: "bytes_sent_per_sec",
				plan: "todo",
				parse: parseInt,
			},
	        "pubsub_channels": &MetricItem{
				name: "pubsub_channels",
				plan: "result",
				parse: parseInt,
			},
	        "pubsub_patterns": &MetricItem{
				name: "pubsub_patterns",
				plan: "result",
				parse: parseInt,
			},
	        "rdb_bgsave_in_progress": &MetricItem{
				name: "rdb_bgsave_in_progress",
				plan: "result",
				parse: parseInt,
			},
	        "rdb_changes_since_last_save": &MetricItem{
				name: "rdb_changes_since_last_save",
				plan: "result",
				parse: parseInt,
			},
	        "rdb_last_bgsave_time_sec": &MetricItem{
				name: "rdb_last_bgsave_time_sec",
				plan: "result",
				parse: parseInt,
			},
	        "mem_fragmentation_ratio": &MetricItem{
				name: "mem_fragmentation_ratio",
				plan: "result",
				parse: parseFloat64,
			},
	        "used_memory": &MetricItem{
				name: "used_memory",
				plan: "result",
				parse: parseInt,
			},
	        "used_memory_lua": &MetricItem{
				name: "used_memory_lua",
				plan: "result",
				parse: parseInt,
			},
	        "used_memory_peak": &MetricItem{
				name: "used_memory_peak",
				plan: "result",
				parse: parseInt,
			},
	        "used_memory_rss": &MetricItem{
				name: "used_memory_rss",
				plan: "result",
				parse: parseInt,
			},
	        "used_memory_startup": &MetricItem{
				name: "used_memory_startup",
				plan: "result",
				parse: parseInt,
			},
	        "used_memory_overhead": &MetricItem{
				name: "used_memory_overhead",
				plan: "result",
				parse: parseInt,
			},
	        "maxmemory": &MetricItem{
				name: "maxmemory",
				plan: "result",
				parse: parseInt,
			},
	        "master_last_io_seconds_ago": &MetricItem{
				name: "master_last_io_seconds_ago",
				plan: "todo",
				parse: parseInt,
				desc: "主从未配置",
			},
	        "master_sync_in_progress": &MetricItem{
				name: "master_sync_in_progress",
				plan: "todo",
				parse: parseInt,
				desc: "主从未配置",
			},
	        "master_sync_left_bytes": &MetricItem{
				name: "master_sync_left_bytes",
				plan: "todo",
				parse: parseInt,
				desc: "主从未配置",
			},
	        "repl_backlog_histlen": &MetricItem{
				name: "repl_backlog_histlen",
				plan: "result",
				parse: parseInt,
			},
	        "master_repl_offset": &MetricItem{
				name: "master_repl_offset",
				plan: "result",
				parse: parseInt,
			},
	        "slave_repl_offset": &MetricItem{
				name: "slave_repl_offset",
				plan: "todo",
				parse: parseInt,
				desc: "主从未配置",
			},
	        "used_cpu_sys": &MetricItem{
				name: "used_cpu_sys",
				plan: "result",
				parse: parseFloat64,
			},
	        "used_cpu_sys_children": &MetricItem{
				name: "used_cpu_sys_children",
				plan: "result",
				parse: parseFloat64,
			},
	        "used_cpu_user": &MetricItem{
				name: "used_cpu_user",
				plan: "result",
				parse: parseFloat64,
			},
	        "used_cpu_user_children": &MetricItem{
				name: "used_cpu_user_children",
				plan: "result",
				parse: parseFloat64,
			},
	        "keyspace_hits": &MetricItem{
				name: "keyspace_hits",
				plan: "result",
				parse: parseInt,
			},
	        "keyspace_misses": &MetricItem{
				name: "keyspace_misses",
				plan: "result",
				parse: parseInt,
			},
		},
	},
	"keyspace": &MetricType{
		metricSet: map[string]*MetricItem{
	        "keys": &MetricItem{
				name: "key_count",
				plan: "todo",
				parse: parseInt,
			},
	        "expires": &MetricItem{
				name: "expires",
				plan: "todo",
				parse: parseInt,
			},
			"avg_ttl": &MetricItem{
				name: "avg_ttl",
				plan: "todo",
				parse: parseInt,
			},
		},
	},
}


