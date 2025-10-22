// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll // Metric descriptions are intentionally long for clarity
package redis

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	TagGroupCommon   = "common"
	TagGroupInfo     = "info"
	TagGroupCommand  = "command"
	TagGroupReplica  = "replica"
	TagGroupDatabase = "db"
)

type redisMeasurement struct{}

//nolint:funlen // Info function contains all metric definitions
func (m *redisMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "redis",
		Desc:   "Metric set including Redis info, commands, replication, database, and cluster statistics, unified in v2",
		DescZh: "指标集包含 Redis info、command、replication、database 和 cluster 相关指标，v2 版本统一",
		Cat:    point.Metric,
		Tags:   m.getTags(),
		Fields: m.getFields(),
	}
}

func (m *redisMeasurement) getTags() map[string]interface{} {
	return mergeMaps(
		m.getCommonTags(),
		m.getInfoTags(),
		m.getCommandTags(),
		m.getReplicaTags(),
		m.getDatabaseTags(),
	)
}

func (m *redisMeasurement) getCommonTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["host"] = &inputs.TagInfo{Desc: "Hostname."}
	tags["server"] = &inputs.TagInfo{Desc: "Server addr."}
	return tags
}

func (m *redisMeasurement) getInfoTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["redis_version"] = &inputs.TagInfo{Desc: "Version of the Redis server."}
	tags["command_type"] = &inputs.TagInfo{Desc: "Command type."}
	tags["error_type"] = &inputs.TagInfo{Desc: "Error type."}
	tags["quantile"] = &inputs.TagInfo{Desc: "Histogram `quantile`."}
	tags["role"] = &inputs.TagInfo{
		Desc: "Value is `master` if the instance is replica of no one, or `slave` if the instance is a replica of some master instance.",
	}
	tags["redis_mode"] = &inputs.TagInfo{Desc: "Mode of the Redis server."}
	tags["os"] = &inputs.TagInfo{Desc: "Operating system of the Redis server."}
	tags["maxmemory_policy"] = &inputs.TagInfo{Desc: "The value of the maxmemory-policy configuration directive."}
	return tags
}

func (m *redisMeasurement) getCommandTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["method"] = &inputs.TagInfo{Desc: "Command type"}
	return tags
}

func (m *redisMeasurement) getReplicaTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["slave_id"] = &inputs.TagInfo{Desc: "Slave ID, only collected for master redis."}
	tags["slave_addr"] = &inputs.TagInfo{Desc: "Slave addr, only collected for master redis."}
	tags["slave_state"] = &inputs.TagInfo{Desc: "Slave state, only collected for master redis."}
	tags["master_addr"] = &inputs.TagInfo{Desc: "Master addr, only collected for slave redis."}
	return tags
}

func (m *redisMeasurement) getDatabaseTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["db_name"] = &inputs.TagInfo{Desc: "DB name."}
	return tags
}

func (m *redisMeasurement) getFields() map[string]interface{} {
	return mergeMaps(
		m.getInfoFields(),
		m.getCommandFields(),
		m.getReplicaFields(),
		m.getDatabaseFields(),
		m.getClusterFields(),
	)
}

func mergeMaps(fieldMaps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, fields := range fieldMaps {
		for k, v := range fields {
			result[k] = v
		}
	}
	return result
}

//nolint:funlen
func (m *redisMeasurement) getInfoFields() map[string]interface{} {
	fields := make(map[string]interface{})
	// info fields
	// The fields is copy from official docs, include "Not number" field.

	// # Server
	// Not number, discard. "redis_version": "Version of the Redis server."
	// Not number, discard. "redis_git_sha1": "Git SHA1."
	// Not number, discard. "redis_git_dirty": "Git dirty flag."
	// Not number, discard. "redis_build_id": "The build id."
	// Not number, discard. "redis_mode": "The server's mode (`standalone`, `sentinel` or `cluster`)."
	// Not number, discard. "os": "Operating system hosting the Redis server."
	fields["arch_bits"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.EnumType,
		Unit:     inputs.NoUnit,
		Desc:     "Architecture (32 or 64 bits).",
	}
	// Not number, discard. "multiplexing_api": "Event loop mechanism used by Redis."
	// Not number, discard. "atomicvar_api": "Atomicvar API used by Redis."
	// Not number, discard. "gcc_version": "Version of the GCC compiler used to compile the Redis server."
	// Not number, discard. "process_id": "PID of the server process."
	// Not number, discard. "process_supervised": "Supervised system (`upstart`, `systemd`, `unknown` or `no`)."
	// Not number, discard. "run_id": "Random value identifying the Redis server (to be used by Sentinel and
	// Cluster)."
	fields["server_time_usec"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.TimestampUS,
		Desc:     "Epoch-based system time with microsecond precision.",
	}
	fields["uptime_in_seconds"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationSecond,
		Desc:     "Number of seconds since Redis server start.",
	}
	fields["uptime_in_days"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationDay,
		Desc:     "Same value expressed in days.",
	}
	fields["hz"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.FrequencyHz,
		Desc:     "The server's current frequency setting.",
	}
	fields["configured_hz"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.FrequencyHz,
		Desc:     "The server's configured frequency setting.",
	}
	fields["lru_clock"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationMS,
		Desc:     "Clock incrementing every minute, for LRU management.",
	}
	// Not number, discard. "executable": "The path to the server's executable."},
	// Not number, discard. "config_file": "The path to the config file."},
	fields["io_threads_active"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.Bool,
		Desc:     "Flag indicating if I/O threads are active.",
	}
	fields["shutdown_in_milliseconds"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationMS,
		Desc:     "The maximum time remaining for replicas to catch up the replication before completing the shutdown sequence. This field is only present during shutdown.",
	}

	// # Clients
	fields["connected_clients"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of client connections (excluding connections from replicas).",
	}
	fields["cluster_connections"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "An approximation of the number of sockets used by the cluster's bus.",
	}
	fields["maxclients"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The value of the `maxclients` configuration directive. This is the upper limit for the sum of connected_clients, connected_slaves and cluster_connections.",
	}
	fields["client_recent_max_input_buffer"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Biggest input buffer among current client connections.",
	}
	fields["client_recent_max_output_buffer"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Biggest output buffer among current client connections.",
	}
	fields["blocked_clients"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of clients pending on a blocking call (`BLPOP/BRPOP/BRPOPLPUSH/BLMOVE/BZPOPMIN/BZPOPMAX`).",
	}
	fields["tracking_clients"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of clients being tracked (CLIENT TRACKING).",
	}
	fields["clients_in_timeout_table"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of clients in the clients timeout table.",
	}
	fields["total_blocking_keys"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of blocking keys.",
	}
	fields["total_blocking_keys_on_nokey"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of blocking keys that one or more clients that would like to be unblocked when the key is deleted.",
	}

	// # Memory
	fields["used_memory"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Total number of bytes allocated by Redis using its allocator (either standard libc, jemalloc, or an alternative allocator such as tcmalloc).",
	}
	// Same with above "used_memory_human": "Human readable representation of previous value."
	fields["used_memory_rss"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Number of bytes that Redis allocated as seen by the operating system (a.k.a *Resident Set Size*).",
	}
	// Same with above "used_memory_rss_human": "Human readable representation of previous value."
	fields["used_memory_peak"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Peak memory consumed by Redis (in bytes).",
	}
	// Same with above "used_memory_peak_human": "Human readable representation of previous value."
	fields["used_memory_peak_perc"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     "The percentage of `used_memory_peak` out of used_memory.",
	}
	fields["used_memory_overhead"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The sum in bytes of all overheads that the server allocated for managing its internal data structures.",
	}
	fields["used_memory_startup"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Initial amount of memory consumed by Redis at startup in bytes.",
	}
	fields["used_memory_dataset"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The size in bytes of the dataset (used_memory_overhead subtracted from `used_memory`).",
	}
	fields["used_memory_dataset_perc"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     "The percentage of used_memory_dataset out of the net memory usage (`used_memory` minus `used_memory_startup`).",
	}
	fields["total_system_memory"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The total amount of memory that the Redis host has.",
	}
	// Same with above "total_system_memory_human": "Human readable representation of previous value."
	fields["used_memory_lua"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Number of bytes used by the Lua engine.",
	}
	// Same with above "used_memory_lua_human": "Human readable representation of previous value."
	fields["used_memory_scripts"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Number of bytes used by cached Lua scripts.",
	}
	// Same with above "used_memory_scripts_human": "Human readable representation of previous value."
	fields["maxmemory"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The value of the Max Memory configuration directive.",
	}
	// Same with above "maxmemory_human": "Human readable representation of previous value."
	// Not number, discard. "maxmemory_policy": "The value of the maxmemory-policy configuration directive."
	fields["mem_fragmentation_ratio"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.PercentDecimal,
		Desc:     "Ratio between `used_memory_rss` and `used_memory`.",
	}
	fields["mem_fragmentation_bytes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Delta between used_memory_rss and used_memory. Note that when the total fragmentation bytes is low (few megabytes), a high ratio (e.g. 1.5 and above) is not an indication of an issue.",
	}
	fields["allocator_frag_ratio"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.PercentDecimal,
		Desc:     "Ratio between allocator_active and allocator_allocated. This is the true (external) fragmentation metric (not `mem_fragmentation_ratio`).",
	}
	fields["allocator_frag_bytes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Delta between allocator_active and allocator_allocated. See note about mem_fragmentation_bytes.",
	}
	fields["allocator_rss_ratio"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.PercentDecimal,
		Desc:     "Ratio between `allocator_resident` and `allocator_active`. This usually indicates pages that the allocator can and probably will soon release back to the OS.",
	}
	fields["allocator_rss_bytes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Delta between allocator_resident and allocator_active.",
	}
	fields["rss_overhead_ratio"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.PercentDecimal,
		Desc:     "Ratio between used_memory_rss (the process RSS) and allocator_resident. This includes RSS overheads that are not allocator or heap related.",
	}
	fields["rss_overhead_bytes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Delta between used_memory_rss (the process RSS) and allocator_resident.",
	}
	fields["allocator_allocated"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Total bytes allocated form the allocator, including internal-fragmentation. Normally the same as used_memory.",
	}
	fields["allocator_active"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Total bytes in the allocator active pages, this includes external-fragmentation.",
	}
	fields["allocator_resident"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Total bytes resident (RSS) in the allocator, this includes pages that can be released to the OS (by MEMORY PURGE, or just waiting).",
	}
	fields["mem_not_counted_for_evict"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Used memory that's not counted for key eviction. This is basically transient replica and AOF buffers.",
	}
	fields["mem_clients_slaves"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Memory used by replica clients - Starting Redis 7.0, replica buffers share memory with the replication backlog, so this field can show 0 when replicas don't trigger an increase of memory usage.",
	}
	fields["mem_clients_normal"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Memory used by normal clients.",
	}
	fields["mem_cluster_links"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Memory used by links to peers on the cluster bus when cluster mode is enabled.",
	}
	fields["mem_aof_buffer"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Transient memory used for AOF and AOF rewrite buffers.",
	}
	fields["mem_replication_backlog"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Memory used by replication backlog.",
	}
	fields["mem_total_replication_buffers"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Total memory consumed for replication buffers - Added in Redis 7.0.",
	}
	// Not number, discard. "mem_allocator": "Memory allocator, chosen at compile time.."
	fields["active_defrag_running"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.Bool,
		Desc:     "Flag indicating if active defragmentation is active.",
	}
	fields["lazyfree_pending_objects"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The number of objects waiting to be freed (as a result of calling UNLINK, or `FLUSHDB` and `FLUSHALL` with the ASYNC option).",
	}
	fields["lazyfreed_objects"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of objects that have been lazy freed.",
	}

	// # Persistence
	fields["loading"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.Bool,
		Desc:     "Flag indicating if the load of a dump file is on-going.",
	}
	fields["async_loading"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.Bool,
		Desc:     "Currently loading replication data-set asynchronously while serving old data. This means `repl-diskless-load` is enabled and set to `swapdb`. Added in Redis 7.0.",
	}
	fields["current_cow_peak"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The peak size in bytes of copy-on-write memory while a child fork is running.",
	}
	fields["current_cow_size"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The size in bytes of copy-on-write memory while a child fork is running.",
	}
	fields["current_cow_size_age"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationSecond,
		Desc:     "The age, in seconds, of the current_cow_size value.",
	}
	fields["current_fork_perc"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     "The percentage of progress of the current fork process. For AOF and RDB forks it is the percentage of current_save_keys_processed out of current_save_keys_total.",
	}
	fields["current_save_keys_processed"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of keys processed by the current save operation.",
	}
	fields["current_save_keys_total"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of keys at the beginning of the current save operation.",
	}
	fields["rdb_changes_since_last_save"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Refers to the number of operations that produced some kind of changes in the dataset since the last time either `SAVE` or `BGSAVE` was called.",
	}
	fields["rdb_bgsave_in_progress"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.Bool,
		Desc:     "Flag indicating a RDB save is on-going.",
	}
	fields["rdb_last_bgsave_status"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.EnumType,
		Unit:     inputs.NoUnit,
		Desc:     "Status of the last RDB save operation (0 is `ok`, -1 is `err`).",
	}
	fields["rdb_last_save_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.TimestampSec,
		Desc:     "Epoch-based timestamp of last successful RDB save.",
	}
	fields["rdb_last_bgsave_time_sec"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationSecond,
		Desc:     "Duration of the last RDB save operation in seconds. -1 if never.",
	}
	fields["rdb_current_bgsave_time_sec"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationSecond,
		Desc:     "Duration of the on-going RDB save operation if any. -1 if not in progress.",
	}
	fields["rdb_last_cow_size"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The size in bytes of copy-on-write memory during the last RDB save operation.",
	}
	fields["rdb_last_load_keys_expired"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of volatile keys deleted during the last RDB loading. Added in Redis 7.0.",
	}
	fields["rdb_last_load_keys_loaded"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of keys loaded during the last RDB loading. Added in Redis 7.0.",
	}
	fields["aof_enabled"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.Bool,
		Desc:     "Flag indicating AOF logging is activated.",
	}
	fields["aof_rewrite_in_progress"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.Bool,
		Desc:     "Flag indicating a AOF rewrite operation is on-going.",
	}
	fields["aof_rewrite_scheduled"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.Bool,
		Desc:     "Flag indicating an AOF rewrite operation will be scheduled once the on-going RDB save is complete.",
	}
	fields["aof_last_rewrite_time_sec"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationSecond,
		Desc:     "Duration of the last AOF rewrite operation in seconds. -1 if never.",
	}
	fields["aof_current_rewrite_time_sec"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationSecond,
		Desc:     "Duration of the on-going AOF rewrite operation if any. -1 if not in progress.",
	}
	fields["aof_last_bgrewrite_status"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.EnumType,
		Unit:     inputs.NoUnit,
		Desc:     "Status of the last AOF rewrite operation (0 is `ok`, and -1 is `err`).",
	}
	fields["aof_last_write_status"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.EnumType,
		Unit:     inputs.NoUnit,
		Desc:     "Status of the last write operation to the AOF (0 is `ok`, and -1 is `err`).",
	}
	fields["aof_last_cow_size"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The size in bytes of copy-on-write memory during the last AOF rewrite operation.",
	}
	fields["module_fork_in_progress"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.Bool,
		Desc:     "Flag indicating a module fork is on-going.",
	}
	fields["module_fork_last_cow_size"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The size in bytes of copy-on-write memory during the last module fork operation.",
	}
	fields["aof_rewrites"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of AOF rewrites performed since startup.",
	}
	fields["rdb_saves"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of RDB snapshots performed since startup.",
	}
	fields["aof_current_size"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "AOF current file size.",
	}
	fields["aof_base_size"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "AOF file size on latest startup or rewrite.",
	}
	fields["aof_pending_rewrite"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.Bool,
		Desc:     "Flag indicating an AOF rewrite operation will be scheduled once the on-going RDB save is complete.",
	}
	fields["aof_buffer_length"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Size of the AOF buffer.",
	}
	fields["aof_rewrite_buffer_length"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Size of the AOF rewrite buffer. Note this field was removed in Redis 7.0.",
	}
	fields["aof_pending_bio_fsync"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of fsync pending jobs in background I/O queue.",
	}
	fields["aof_delayed_fsync"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Delayed fsync counter.",
	}
	fields["loading_start_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.TimestampSec,
		Desc:     "Epoch-based timestamp of the start of the load operation.",
	}
	fields["loading_total_bytes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Total file size.",
	}
	fields["loading_rdb_used_mem"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "The memory usage of the server that had generated the RDB file at the time of the file's creation.",
	}
	fields["loading_loaded_bytes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Number of bytes already loaded.",
	}
	fields["loading_loaded_perc"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     "Same value expressed as a percentage.",
	}
	fields["loading_eta_seconds"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationSecond,
		Desc:     "ETA in seconds for the load to be complete.",
	}

	// # Stats
	fields["total_connections_received"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Total number of connections accepted by the server.",
	}
	fields["total_commands_processed"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Total number of commands processed by the server.",
	}
	fields["instantaneous_ops_per_sec"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of commands processed per second.",
	}
	fields["total_net_input_bytes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.SizeByte,
		Desc:     "The total number of bytes read from the network.",
	}
	fields["total_net_output_bytes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.SizeByte,
		Desc:     "The total number of bytes written to the network.",
	}
	fields["total_net_repl_input_bytes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.SizeByte,
		Desc:     "The total number of bytes read from the network for replication purposes.",
	}
	fields["total_net_repl_output_bytes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.SizeByte,
		Desc:     "The total number of bytes written to the network for replication purposes.",
	}
	fields["instantaneous_input_kbps"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.KBytesPerSec,
		Desc:     "The network's read rate per second in KB/sec.",
	}
	fields["instantaneous_output_kbps"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.KBytesPerSec,
		Desc:     "The network's write rate per second in KB/sec.",
	}
	fields["instantaneous_input_repl_kbps"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.KBytesPerSec,
		Desc:     "The network's read rate per second in KB/sec for replication purposes.",
	}
	fields["instantaneous_output_repl_kbps"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.KBytesPerSec,
		Desc:     "The network's write rate per second in KB/sec for replication purposes.",
	}
	fields["rejected_connections"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of connections rejected because of Max-Clients limit.",
	}
	fields["sync_full"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of full `resyncs` with replicas.",
	}
	fields["sync_partial_ok"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of accepted partial resync requests.",
	}
	fields["sync_partial_err"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The number of denied partial resync requests.",
	}
	fields["expired_keys"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Total number of key expiration events.",
	}
	fields["expired_stale_perc"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     "The percentage of keys probably expired.",
	}
	fields["expired_time_cap_reached_count"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "The count of times that active expiry cycles have stopped early.",
	}
	fields["expire_cycle_cpu_milliseconds"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.DurationMS,
		Desc:     "The cumulative amount of time spent on active expiry cycles.",
	}
	fields["evicted_keys"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of evicted keys due to Max-Memory limit.",
	}
	fields["evicted_clients"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of evicted clients due to `maxmemory-clients` limit. Added in Redis 7.0.",
	}
	fields["total_eviction_exceeded_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.DurationMS,
		Desc:     "Total time used_memory was greater than `maxmemory` since server startup, in milliseconds.",
	}
	fields["current_eviction_exceeded_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationMS,
		Desc:     "The time passed since used_memory last rose above `maxmemory`, in milliseconds.",
	}
	fields["keyspace_hits"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of successful lookup of keys in the main dictionary.",
	}
	fields["keyspace_misses"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of failed lookup of keys in the main dictionary.",
	}
	fields["pubsub_channels"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Global number of pub/sub channels with client subscriptions.",
	}
	fields["pubsub_patterns"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Global number of pub/sub pattern with client subscriptions.",
	}
	fields["pubsubshard_channels"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Global number of pub/sub shard channels with client subscriptions. Added in Redis 7.0.3.",
	}
	fields["latest_fork_usec"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationUS,
		Desc:     "Duration of the latest fork operation in microseconds.",
	}
	fields["total_forks"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Total number of fork operations since the server start.",
	}
	fields["migrate_cached_sockets"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The number of sockets open for MIGRATE purposes.",
	}
	fields["slave_expires_tracked_keys"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The number of keys tracked for expiry purposes (applicable only to writable replicas).",
	}
	fields["active_defrag_hits"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of value reallocations performed by active the defragmentation process.",
	}
	fields["active_defrag_misses"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of aborted value reallocations started by the active defragmentation process.",
	}
	fields["active_defrag_key_hits"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of keys that were actively defragmented.",
	}
	fields["active_defrag_key_misses"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of keys that were skipped by the active defragmentation process.",
	}
	fields["total_active_defrag_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.DurationMS,
		Desc:     "Total time memory fragmentation was over the limit, in milliseconds.",
	}
	fields["current_active_defrag_time"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationMS,
		Desc:     "The time passed since memory fragmentation last was over the limit, in milliseconds.",
	}
	fields["tracking_total_keys"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of keys being tracked by the server.",
	}
	fields["tracking_total_items"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of items, that is the sum of clients number for each key, that are being tracked.",
	}
	fields["tracking_total_prefixes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of tracked prefixes in server's prefix table (only applicable for broadcast mode).",
	}
	fields["unexpected_error_replies"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of unexpected error replies, that are types of errors from an AOF load or replication.",
	}
	fields["total_error_replies"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Total number of issued error replies, that is the sum of rejected commands (errors prior command execution) and failed commands (errors within the command execution).",
	}
	fields["dump_payload_sanitizations"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Total number of dump payload deep integrity validations (see sanitize-dump-payload config).",
	}
	fields["total_reads_processed"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Total number of read events processed.",
	}
	fields["total_writes_processed"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Total number of write events processed.",
	}
	fields["io_threaded_reads_processed"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of read events processed by the main and I/O threads.",
	}
	fields["io_threaded_writes_processed"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of write events processed by the main and I/O threads.",
	}
	fields["stat_reply_buffer_shrinks"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Total number of output buffer shrinks.",
	}
	fields["stat_reply_buffer_expands"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Total number of output buffer expands.",
	}
	fields["eventloop_cycles"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Total number of `eventloop` cycles.",
	}
	fields["eventloop_duration_sum"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.DurationUS,
		Desc:     "Total time spent in the `eventloop` in microseconds (including I/O and command processing).",
	}
	fields["eventloop_duration_cmd_sum"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.DurationUS,
		Desc:     "Total time spent on executing commands in microseconds.",
	}
	fields["instantaneous_eventloop_cycles_per_sec"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of `eventloop` cycles per second.",
	}
	fields["instantaneous_eventloop_duration_usec"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationUS,
		Desc:     "Average time spent in a single `eventloop` cycle in microseconds.",
	}
	fields["acl_access_denied_auth"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of authentication failures.",
	}
	fields["acl_access_denied_cmd"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of commands rejected because of access denied to the command.",
	}
	fields["acl_access_denied_key"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of commands rejected because of access denied to a key.",
	}
	fields["acl_access_denied_channel"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of commands rejected because of access denied to a channel.",
	}

	// # Replication
	// Not number, discard. "role": "Value is `master` if the instance is replica of no one, or `slave` if the
	// instance is a replica of some
	// master instance."
	// Not number, discard. "master_failover_state": "The state of an ongoing failover, if any.."
	// Not number, discard. "master_replid": "The replication ID of the Redis server.."
	// Not number, discard. "master_replid2": "The secondary replication ID, used for PSYNC after a failover.."
	fields["master_repl_offset"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The server's current replication offset.",
	}
	fields["second_repl_offset"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The offset up to which replication IDs are accepted. -1 if not PSYNC2.",
	}
	fields["repl_backlog_active"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.Bool,
		Desc:     "Flag indicating replication backlog is active.",
	}
	fields["repl_backlog_size"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Total size in bytes of the replication backlog buffer.",
	}
	fields["repl_backlog_first_byte_offset"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The master offset of the replication backlog buffer.",
	}
	fields["repl_backlog_histlen"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Size in bytes of the data in the replication backlog buffer.",
	}
	// Not number, discard. "master_host": "Host or IP address of the master."
	// Not number, discard. "master_port": "Master listening TCP port."
	// Not number, discard. "master_link_status": "Status of the link (up/down)."
	fields["master_last_io_seconds_ago"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationSecond,
		Desc:     "Number of seconds since the last interaction with master.",
	}
	fields["master_sync_in_progress"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.Bool,
		Desc:     "Indicate the master is syncing to the replica.",
	}
	fields["slave_read_repl_offset"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The read replication offset of the replica instance.",
	}
	fields["slave_repl_offset"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The replication offset of the replica instance.",
	}
	fields["slave_priority"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The priority of the instance as a candidate for failover.",
	}
	fields["slave_read_only"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Flag indicating if the replica is read-only.",
	}
	fields["replica_announced"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Flag indicating if the replica is announced by Sentinel.",
	}
	fields["master_sync_total_bytes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Total number of bytes that need to be transferred. this may be 0 when the size is unknown (for example, when the `repl-diskless-sync` configuration directive is used).",
	}
	fields["master_sync_read_bytes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Number of bytes already transferred.",
	}
	fields["master_sync_left_bytes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Number of bytes left before syncing is complete (may be negative when master_sync_total_bytes is 0).",
	}
	fields["master_sync_perc"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     "The percentage master_sync_read_bytes from master_sync_total_bytes, or an approximation that uses loading_rdb_used_mem when master_sync_total_bytes is 0.",
	}
	fields["master_sync_last_io_seconds_ago"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationSecond,
		Desc:     "Number of seconds since last transfer I/O during a SYNC operation.",
	}
	fields["master_link_down_since_seconds"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationSecond,
		Desc:     "Number of seconds since the link is down.",
	}
	fields["connected_slaves"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of connected replicas.",
	}
	fields["min_slaves_good_slaves"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of replicas currently considered good.",
	}
	// Not number, discard. "slaveXXX": "id, IP address, port, state, offset, lag."

	// # CPU
	fields["used_cpu_sys"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Count,
		Unit:     inputs.DurationSecond,
		Desc:     "System CPU consumed by the Redis server, which is the sum of system CPU consumed by all threads of the server process (main thread and background threads).",
	}
	fields["used_cpu_user"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Count,
		Unit:     inputs.DurationSecond,
		Desc:     "User CPU consumed by the Redis server, which is the sum of user CPU consumed by all threads of the server process (main thread and background threads).",
	}
	fields["used_cpu_sys_children"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Count,
		Unit:     inputs.DurationSecond,
		Desc:     "System CPU consumed by the background processes.",
	}
	fields["used_cpu_user_children"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Count,
		Unit:     inputs.DurationSecond,
		Desc:     "User CPU consumed by the background processes.",
	}
	fields["used_cpu_sys_main_thread"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Count,
		Unit:     inputs.DurationSecond,
		Desc:     "System CPU consumed by the Redis server main thread.",
	}
	fields["used_cpu_user_main_thread"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Count,
		Unit:     inputs.DurationSecond,
		Desc:     "User CPU consumed by the Redis server main thread.",
	}

	// # Modules

	// # Commandstats
	// already in redis_command_stat

	// # Errorstats
	fields["errorstat"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Track of the different errors that occurred within Redis.",
	}

	// # Latencystats
	fields["latency_percentiles_usec"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationUS,
		Desc:     "Latency percentile distribution statistics based on the command type.",
	}

	// # Cluster
	fields["cluster_enabled"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.Bool,
		Desc:     "Indicate Redis cluster is enabled.",
	}

	// # Keyspace
	// already in redis_db
	// # Calculate metric
	fields["info_latency_ms"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationMS,
		Desc:     "The latency of the redis INFO command.",
	}
	fields["used_cpu_sys_percent"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     "System CPU percentage consumed by the Redis server, which is the sum of system CPU consumed by all threads of the server process (main thread and background threads).",
	}
	fields["used_cpu_user_percent"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     "User CPU percentage consumed by the Redis server, which is the sum of user CPU consumed by all threads of the server process (main thread and background threads).",
	}

	// # Old version metric
	fields["client_biggest_input_buf"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Biggest input buffer among current client connections.",
	}
	fields["client_longest_output_list"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Longest output list among current client connections.",
	}

	// client fields
	fields["max_idle"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationSecond,
		Desc:     "Max idle time of the connection in seconds of all these clients",
	}
	fields["max_multi"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Max number of commands in a MULTI/EXEC context of all these clients",
	}
	fields["max_obl"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Max output buffer length of all these clients",
	}
	fields["max_qbuf"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Max query buffer length (0 means no query pending) of all these clients",
	}
	fields["max_totmem"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Max total memory consumed by various buffers of all these clients",
	}
	fields["multi_avg"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Avg number of commands in a MULTI/EXEC context of all these clients",
	}
	fields["multi_total"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Sum of all client's total number of commands in a MULTI/EXEC context",
	}
	fields["total_cmds"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Sum of all client's total count of commands this client executed.",
	}
	fields["total_netin"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Sum of all client's total network input bytes read from the client",
	}
	fields["total_netout"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Sum of all client's total network output bytes send to the client",
	}
	fields["total_obl"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Sum of all client's output buffer length of all these clients",
	}
	fields["total_psub"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Sum of all client's number of pattern matching subscriptions",
	}
	fields["total_ssub"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Sum of all client's number of shard channel subscriptions.(Redis >= 7.0.3)",
	}
	fields["total_qbuf"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Sum of all client's query buffer length",
	}
	fields["total_sub"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Sum of all client's number of channel subscriptions",
	}
	fields["total_totmem"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     "Sum of all client's total memory consumed in various buffers",
	}

	m.addTaggedbyToFields(fields, TagGroupInfo)
	return fields
}

func (m *redisMeasurement) getCommandFields() map[string]interface{} {
	fields := make(map[string]interface{})
	// command fields
	fields["calls"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The number of calls that reached command execution.",
	}
	fields["usec"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationUS,
		Desc:     "The total CPU time consumed by these commands.",
	}
	fields["usec_per_call"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationUS,
		Desc:     "The average CPU consumed per command execution.",
	}
	fields["rejected_calls"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The number of rejected calls (errors prior command execution).",
	}
	fields["failed_calls"] = &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The number of failed calls (errors within the command execution).",
	}

	m.addTaggedbyToFields(fields, TagGroupCommand)
	return fields
}

func (m *redisMeasurement) getReplicaFields() map[string]interface{} {
	fields := make(map[string]interface{})

	// replica fields
	fields["master_link_status"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Desc:     "Status of the link (up/down), `1` for up, `0` for down, only collected for slave redis.",
	}
	fields["slave_offset"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Desc:     "Slave offset, only collected for master redis.",
	}
	fields["slave_lag"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Desc:     "Slave lag, only collected for master redis.",
	}

	m.addTaggedbyToFields(fields, TagGroupReplica)
	return fields
}

func (m *redisMeasurement) getDatabaseFields() map[string]interface{} {
	fields := make(map[string]interface{})
	// db fields
	fields["keys"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Desc:     "Key.",
	}
	fields["expires"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Desc:     "expires time.",
	}
	fields["avg_ttl"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Desc:     "Average ttl.",
	}

	m.addTaggedbyToFields(fields, TagGroupDatabase)
	return fields
}

func (m *redisMeasurement) getClusterFields() map[string]interface{} {
	fields := make(map[string]interface{})
	// cluster fields
	fields["cluster_state"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.EnumValue,
		Desc:     "State is 1(ok) if the node is able to receive queries. 0(fail) if there is at least one hash slot which is unbound (no node associated), in error state (node serving it is flagged with FAIL flag), or if the majority of masters can't be reached by this node.",
	}
	fields["cluster_slots_assigned"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     " Number of slots which are associated to some node (not unbound). This number should be 16384 for the node to work properly, which means that each hash slot should be mapped to a node.",
	}
	fields["cluster_slots_ok"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of hash slots mapping to a node not in `FAIL` or `PFAIL` state.",
	}
	fields["cluster_slots_pfail"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of hash slots mapping to a node in `PFAIL` state. Note that those hash slots still work correctly, as long as the `PFAIL` state is not promoted to FAIL by the failure detection algorithm. `PFAIL` only means that we are currently not able to talk with the node, but may be just a transient error.",
	}
	fields["cluster_slots_fail"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Number of hash slots mapping to a node in FAIL state. If this number is not zero the node is not able to serve queries unless cluster-require-full-coverage is set to no in the configuration.",
	}
	fields["cluster_known_nodes"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The total number of known nodes in the cluster, including nodes in HANDSHAKE state that may not currently be proper members of the cluster.",
	}
	fields["cluster_size"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "The number of master nodes serving at least one hash slot in the cluster.",
	}
	fields["cluster_current_epoch"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NoUnit,
		Desc:     "The local Current Epoch variable. This is used in order to create unique increasing version numbers during fail overs.",
	}
	fields["cluster_my_epoch"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NoUnit,
		Desc:     "The Config Epoch of the node we are talking with. This is the current configuration version assigned to this node.",
	}
	fields["cluster_stats_messages_sent"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of messages sent via the cluster node-to-node binary bus.",
	}
	fields["cluster_stats_messages_received"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Number of messages received via the cluster node-to-node binary bus.",
	}
	fields["total_cluster_links_buffer_limit_exceeded"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Accumulated count of cluster links freed due to exceeding the `cluster-link-sendbuf-limit` configuration.",
	}
	fields["cluster_stats_messages_ping_sent"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Cluster bus send PING (not to be confused with the client command PING).",
	}
	fields["cluster_stats_messages_ping_received"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Cluster bus received PING (not to be confused with the client command PING).",
	}
	fields["cluster_stats_messages_pong_sent"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "PONG send (reply to PING).",
	}
	fields["cluster_stats_messages_pong_received"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "PONG received (reply to PING).",
	}
	fields["cluster_stats_messages_meet_sent"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Handshake message sent to a new node, either through gossip or CLUSTER MEET.",
	}
	fields["cluster_stats_messages_meet_received"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Handshake message received from a new node, either through gossip or CLUSTER MEET.",
	}
	fields["cluster_stats_messages_fail_sent"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Mark node xxx as failing send.",
	}
	fields["cluster_stats_messages_fail_received"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Mark node xxx as failing received.",
	}
	fields["cluster_stats_messages_publish_sent"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Pub/Sub Publish propagation send.",
	}
	fields["cluster_stats_messages_publish_received"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Pub/Sub Publish propagation received.",
	}
	fields["cluster_stats_messages_auth_req_sent"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Replica initiated leader election to replace its master.",
	}
	fields["cluster_stats_messages_auth_req_received"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Replica initiated leader election to replace its master.",
	}
	fields["cluster_stats_messages_auth_ack_sent"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Message indicating a vote during leader election.",
	}
	fields["cluster_stats_messages_auth_ack_received"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Message indicating a vote during leader election.",
	}
	fields["cluster_stats_messages_update_sent"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Another node slots configuration.",
	}
	fields["cluster_stats_messages_update_received"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Another node slots configuration.",
	}
	fields["cluster_stats_messages_mfstart_sent"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Pause clients for manual failover.",
	}
	fields["cluster_stats_messages_mfstart_received"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Pause clients for manual failover.",
	}
	fields["cluster_stats_messages_module_sent"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Module cluster API message.",
	}
	fields["cluster_stats_messages_module_received"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     "Module cluster API message.",
	}
	fields["cluster_stats_messages_publishshard_sent"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Pub/Sub Publish shard propagation, see Sharded Pubsub.",
	}
	fields["cluster_stats_messages_publishshard_received"] = &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     "Pub/Sub Publish shard propagation, see Sharded Pubsub.",
	}
	return fields
}

func (m *redisMeasurement) addTaggedbyToFields(fields map[string]interface{}, tagGroup string) {
	var tags map[string]interface{}

	// Select corresponding getTags function based on tagGroup
	switch tagGroup {
	case TagGroupInfo:
		tags = m.getInfoTags()
	case TagGroupCommand:
		tags = m.getCommandTags()
	case TagGroupReplica:
		tags = m.getReplicaTags()
	case TagGroupDatabase:
		tags = m.getDatabaseTags()
	default:
		return
	}

	// Extract tag keys
	taggedBy := make([]string, 0, len(tags))
	for tag := range tags {
		taggedBy = append(taggedBy, tag)
	}

	// Add Taggedby to each field
	for _, field := range fields {
		if fieldInfo, ok := field.(*inputs.FieldInfo); ok {
			fieldInfo.Taggedby = taggedBy
		}
	}
}
