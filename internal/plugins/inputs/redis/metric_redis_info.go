// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/go-redis/redis/v8"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type infoMeasurement struct {
	cli                *redis.Client
	name               string
	tags               map[string]string
	lastCollect        *redisCPUUsage
	latencyPercentiles bool
}

// Info get metric info
//
// See also: official docs: https://redis.io/commands/info/
//
// See also: https://github.com/redis/redis/blob/unstable/src/server.c
//
//nolint:lll,funlen
func (m *infoMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: redisInfoM,
		Type: "metric",
		Fields: map[string]interface{}{
			// The fields is copy from official docs, include "Not number" field.

			// # Server
			// Not number, discard. "redis_version":      "Version of the Redis server."},
			// Not number, discard. "redis_git_sha1":     "Git SHA1."},
			// Not number, discard. "redis_git_dirty":    "Git dirty flag."},
			// Not number, discard. "redis_build_id":     "The build id."},
			// Not number, discard. "redis_mode":         "The server's mode (`standalone`, `sentinel` or `cluster`)."},
			// Not number, discard. "os":                 "Operating system hosting the Redis server."},
			"arch_bits": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Architecture (32 or 64 bits)."},
			// Not number, discard. "multiplexing_api":   "Event loop mechanism used by Redis."},
			// Not number, discard. "atomicvar_api":      "Atomicvar API used by Redis."},
			// Not number, discard. "gcc_version":        "Version of the GCC compiler used to compile the Redis server."},
			// Not number, discard. "process_id":         "PID of the server process."},
			// Not number, discard. "process_supervised": "Supervised system (`upstart`, `systemd`, `unknown` or `no`)."},
			// Not number, discard. "run_id":             "Random value identifying the Redis server (to be used by Sentinel and Cluster)."},
			"tcp_port":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "TCP/IP listen port."},
			"server_time_usec":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Epoch-based system time with microsecond precision."},
			"uptime_in_seconds": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Number of seconds since Redis server start."},
			"uptime_in_days":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationDay, Desc: "Same value expressed in days."},
			"hz":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The server's current frequency setting."},
			"configured_hz":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The server's configured frequency setting."},
			"lru_clock":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Clock incrementing every minute, for LRU management."},
			// Not number, discard. "executable":         "The path to the server's executable."},
			// Not number, discard. "config_file":        "The path to the config file."},
			"io_threads_active":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Flag indicating if I/O threads are active."},
			"shutdown_in_milliseconds": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The maximum time remaining for replicas to catch up the replication before completing the shutdown sequence. This field is only present during shutdown."},

			// # Clients
			"connected_clients":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: " Number of client connections (excluding connections from replicas)"},
			"cluster_connections":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "An approximation of the number of sockets used by the cluster's bus."},
			"maxclients":                      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The value of the `maxclients` configuration directive. This is the upper limit for the sum of connected_clients, connected_slaves and cluster_connections."},
			"client_recent_max_input_buffer":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Biggest input buffer among current client connections."},
			"client_recent_max_output_buffer": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Biggest output buffer among current client connections."},
			"blocked_clients":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of clients pending on a blocking call (`BLPOP/BRPOP/BRPOPLPUSH/BLMOVE/BZPOPMIN/BZPOPMAX`)"},
			"tracking_clients":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of clients being tracked (CLIENT TRACKING)."},
			"clients_in_timeout_table":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of clients in the clients timeout table."},
			"total_blocking_keys":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of blocking keys."},
			"total_blocking_keys_on_nokey":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of blocking keys that one or more clients that would like to be unblocked when the key is deleted."},

			// # Memory
			"used_memory": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total number of bytes allocated by Redis using its allocator (either standard libc, jemalloc, or an alternative allocator such as tcmalloc)"},
			// Same with above "used_memory_human": "Human readable representation of previous value."},
			"used_memory_rss": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Number of bytes that Redis allocated as seen by the operating system (a.k.a resident set size)"},
			// Same with above "used_memory_rss_human": "Human readable representation of previous value."},
			"used_memory_peak": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Peak memory consumed by Redis (in bytes)"},
			// Same with above "used_memory_peak_human": "Human readable representation of previous value."},
			"used_memory_peak_perc":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The percentage of used_memory_peak out of used_memory."},
			"used_memory_overhead":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The sum in bytes of all overheads that the server allocated for managing its internal data structures"},
			"used_memory_startup":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Initial amount of memory consumed by Redis at startup in bytes"},
			"used_memory_dataset":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The size in bytes of the dataset (used_memory_overhead subtracted from used_memory)."},
			"used_memory_dataset_perc": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The percentage of used_memory_dataset out of the net memory usage (used_memory minus used_memory_startup)."},
			"total_system_memory":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total amount of memory that the Redis host has."},
			// Same with above "total_system_memory_human": "Human readable representation of previous value."},
			"used_memory_lua": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Number of bytes used by the Lua engine"},
			// Same with above "used_memory_lua_human": "Human readable representation of previous value."},
			"used_memory_scripts": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Number of bytes used by cached Lua scripts."},
			// Same with above "used_memory_scripts_human": "Human readable representation of previous value."},
			"maxmemory": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The value of the Max Memory configuration directive"},
			// Same with above "maxmemory_human": "Human readable representation of previous value."},
			// Not number, discard. "maxmemory_policy": "The value of the maxmemory-policy configuration directive."},
			"mem_fragmentation_ratio":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownType, Desc: "Ratio between used_memory_rss and used_memory"},
			"mem_fragmentation_bytes":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Delta between used_memory_rss and used_memory. Note that when the total fragmentation bytes is low (few megabytes), a high ratio (e.g. 1.5 and above) is not an indication of an issue.."},
			"allocator_frag_ratio":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownType, Desc: "Ratio between allocator_active and allocator_allocated. This is the true (external) fragmentation metric (not mem_fragmentation_ratio).."},
			"allocator_frag_bytes":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Delta between allocator_active and allocator_allocated. See note about mem_fragmentation_bytes.."},
			"allocator_rss_ratio":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownType, Desc: "Ratio between allocator_resident and allocator_active. This usually indicates pages that the allocator can and probably will soon release back to the OS.."},
			"allocator_rss_bytes":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Delta between allocator_resident and allocator_active."},
			"rss_overhead_ratio":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownType, Desc: "Ratio between used_memory_rss (the process RSS) and allocator_resident. This includes RSS overheads that are not allocator or heap related.."},
			"rss_overhead_bytes":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Delta between used_memory_rss (the process RSS) and allocator_resident."},
			"allocator_allocated":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total bytes allocated form the allocator, including internal-fragmentation. Normally the same as used_memory.."},
			"allocator_active":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total bytes in the allocator active pages, this includes external-fragmentation.."},
			"allocator_resident":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total bytes resident (RSS) in the allocator, this includes pages that can be released to the OS (by MEMORY PURGE, or just waiting).."},
			"mem_not_counted_for_evict":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Used memory that's not counted for key eviction. This is basically transient replica and AOF buffers.."},
			"mem_clients_slaves":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Memory used by replica clients - Starting Redis 7.0, replica buffers share memory with the replication backlog, so this field can show 0 when replicas don't trigger an increase of memory usage.."},
			"mem_clients_normal":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Memory used by normal clients."},
			"mem_cluster_links":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Memory used by links to peers on the cluster bus when cluster mode is enabled.."},
			"mem_aof_buffer":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Transient memory used for AOF and AOF rewrite buffers."},
			"mem_replication_backlog":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Memory used by replication backlog."},
			"mem_total_replication_buffers": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total memory consumed for replication buffers - Added in Redis 7.0.."},
			// Not number, discard. "mem_allocator": "Memory allocator, chosen at compile time.."},
			"active_defrag_running":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Flag indicating if active defragmentation is active"},
			"lazyfree_pending_objects": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of objects waiting to be freed (as a result of calling UNLINK, or `FLUSHDB` and `FLUSHALL` with the ASYNC option)."},
			"lazyfreed_objects":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of objects that have been lazy freed.."},

			// # Persistence
			"loading":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Flag indicating if the load of a dump file is on-going."},
			"async_loading":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Currently loading replication data-set asynchronously while serving old data. This means `repl-diskless-load` is enabled and set to `swapdb`. Added in Redis 7.0.."},
			"current_cow_peak":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The peak size in bytes of copy-on-write memory while a child fork is running."},
			"current_cow_size":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The size in bytes of copy-on-write memory while a child fork is running."},
			"current_cow_size_age":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "The age, in seconds, of the current_cow_size value.."},
			"current_fork_perc":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The percentage of progress of the current fork process. For AOF and RDB forks it is the percentage of current_save_keys_processed out of current_save_keys_total.."},
			"current_save_keys_processed": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of keys processed by the current save operation."},
			"current_save_keys_total":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of keys at the beginning of the current save operation."},
			"rdb_changes_since_last_save": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Refers to the number of operations that produced some kind of changes in the dataset since the last time either `SAVE` or `BGSAVE` was called."},
			"rdb_bgsave_in_progress":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Flag indicating a RDB save is on-going"},
			"rdb_last_save_time":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.TimestampSec, Desc: "Epoch-based timestamp of last successful RDB save."},
			// Not number, discard. "rdb_last_bgsave_status": "Status of the last RDB save operation."},
			"rdb_last_bgsave_time_sec":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Duration of the last RDB save operation in seconds"},
			"rdb_current_bgsave_time_sec":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Duration of the on-going RDB save operation if any."},
			"rdb_last_cow_size":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The size in bytes of copy-on-write memory during the last RDB save operation."},
			"rdb_last_load_keys_expired":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of volatile keys deleted during the last RDB loading. Added in Redis 7.0.."},
			"rdb_last_load_keys_loaded":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of keys loaded during the last RDB loading. Added in Redis 7.0.."},
			"aof_enabled":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Flag indicating AOF logging is activated."},
			"aof_rewrite_in_progress":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Flag indicating a AOF rewrite operation is on-going"},
			"aof_rewrite_scheduled":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Flag indicating an AOF rewrite operation will be scheduled once the on-going RDB save is complete.."},
			"aof_last_rewrite_time_sec":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Duration of the last AOF rewrite operation in seconds"},
			"aof_current_rewrite_time_sec": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Duration of the on-going AOF rewrite operation if any."},
			// Not number, discard. "aof_last_bgrewrite_status": "Status of the last AOF rewrite operation."},
			// Not number, discard. "aof_last_write_status": "Status of the last write operation to the AOF."},
			"aof_last_cow_size":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The size in bytes of copy-on-write memory during the last AOF rewrite operation."},
			"module_fork_in_progress":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Flag indicating a module fork is on-going."},
			"module_fork_last_cow_size": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The size in bytes of copy-on-write memory during the last module fork operation."},
			"aof_rewrites":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of AOF rewrites performed since startup."},
			"rdb_saves":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of RDB snapshots performed since startup."},
			"aof_current_size":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "AOF current file size"},
			"aof_base_size":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "AOF file size on latest startup or rewrite."},
			"aof_pending_rewrite":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Flag indicating an AOF rewrite operation will be scheduled once the on-going RDB save is complete.."},
			"aof_buffer_length":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Size of the AOF buffer"},
			"aof_rewrite_buffer_length": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Size of the AOF rewrite buffer. Note this field was removed in Redis 7.0."},
			"aof_pending_bio_fsync":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of fsync pending jobs in background I/O queue."},
			"aof_delayed_fsync":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Delayed fsync counter."},
			"loading_start_time":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.TimestampSec, Desc: "Epoch-based timestamp of the start of the load operation."},
			"loading_total_bytes":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total file size."},
			"loading_rdb_used_mem":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The memory usage of the server that had generated the RDB file at the time of the file's creation."},
			"loading_loaded_bytes":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Number of bytes already loaded"},
			"loading_loaded_perc":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Same value expressed as a percentage"},
			"loading_eta_seconds":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "ETA in seconds for the load to be complete"},

			// # Stats
			"total_connections_received":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of connections accepted by the server."},
			"total_commands_processed":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of commands processed by the server."},
			"instantaneous_ops_per_sec":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of commands processed per second."},
			"total_net_input_bytes":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total number of bytes read from the network"},
			"total_net_output_bytes":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total number of bytes written to the network"},
			"total_net_repl_input_bytes":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total number of bytes read from the network for replication purposes."},
			"total_net_repl_output_bytes":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total number of bytes written to the network for replication purposes."},
			"instantaneous_input_kbps":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec, Desc: "The network's read rate per second in KB/sec."},
			"instantaneous_output_kbps":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec, Desc: "The network's write rate per second in KB/sec."},
			"instantaneous_input_repl_kbps":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec, Desc: "The network's read rate per second in KB/sec for replication purposes."},
			"instantaneous_output_repl_kbps":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec, Desc: "The network's write rate per second in KB/sec for replication purposes."},
			"rejected_connections":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of connections rejected because of Max-Clients limit"},
			"sync_full":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of full `resyncs` with replicas."},
			"sync_partial_ok":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of accepted partial resync requests."},
			"sync_partial_err":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of denied partial resync requests."},
			"expired_keys":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of key expiration events"},
			"expired_stale_perc":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The percentage of keys probably expired."},
			"expired_time_cap_reached_count":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The count of times that active expiry cycles have stopped early."},
			"expire_cycle_cpu_milliseconds":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The cumulative amount of time spent on active expiry cycles."},
			"evicted_keys":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of evicted keys due to Max-Memory limit"},
			"evicted_clients":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of evicted clients due to `maxmemory-clients` limit. Added in Redis 7.0.."},
			"total_eviction_exceeded_time":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time used_memory was greater than `maxmemory` since server startup, in milliseconds."},
			"current_eviction_exceeded_time":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The time passed since used_memory last rose above `maxmemory`, in milliseconds."},
			"keyspace_hits":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of successful lookup of keys in the main dictionary"},
			"keyspace_misses":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of failed lookup of keys in the main dictionary"},
			"pubsub_channels":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Global number of pub/sub channels with client subscriptions"},
			"pubsub_patterns":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Global number of pub/sub pattern with client subscriptions"},
			"pubsubshard_channels":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Global number of pub/sub shard channels with client subscriptions. Added in Redis 7.0.3."},
			"latest_fork_usec":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "Duration of the latest fork operation in microseconds"},
			"total_forks":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of fork operations since the server start."},
			"migrate_cached_sockets":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of sockets open for MIGRATE purposes."},
			"slave_expires_tracked_keys":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of keys tracked for expiry purposes (applicable only to writable replicas)."},
			"active_defrag_hits":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of value reallocations performed by active the defragmentation process"},
			"active_defrag_misses":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of aborted value reallocations started by the active defragmentation process"},
			"active_defrag_key_hits":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of keys that were actively defragmented"},
			"active_defrag_key_misses":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of keys that were skipped by the active defragmentation process"},
			"total_active_defrag_time":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Total time memory fragmentation was over the limit, in milliseconds."},
			"current_active_defrag_time":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The time passed since memory fragmentation last was over the limit, in milliseconds."},
			"tracking_total_keys":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of keys being tracked by the server."},
			"tracking_total_items":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of items, that is the sum of clients number for each key, that are being tracked."},
			"tracking_total_prefixes":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of tracked prefixes in server's prefix table (only applicable for broadcast mode)."},
			"unexpected_error_replies":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of unexpected error replies, that are types of errors from an AOF load or replication."},
			"total_error_replies":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of issued error replies, that is the sum of rejected commands (errors prior command execution) and failed commands (errors within the command execution)."},
			"dump_payload_sanitizations":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of dump payload deep integrity validations (see sanitize-dump-payload config).."},
			"total_reads_processed":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of read events processed."},
			"total_writes_processed":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of write events processed."},
			"io_threaded_reads_processed":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of read events processed by the main and I/O threads."},
			"io_threaded_writes_processed":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of write events processed by the main and I/O threads."},
			"stat_reply_buffer_shrinks":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of output buffer shrinks."},
			"stat_reply_buffer_expands":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of output buffer expands."},
			"eventloop_cycles":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of `eventloop` cycles."},
			"eventloop_duration_sum":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "Total time spent in the `eventloop` in microseconds (including I/O and command processing)."},
			"eventloop_duration_cmd_sum":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "Total time spent on executing commands in microseconds."},
			"instantaneous_eventloop_cycles_per_sec": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Number of `eventloop` cycles per second."},
			"instantaneous_eventloop_duration_usec":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "Average time spent in a single `eventloop` cycle in microseconds."},
			"acl_access_denied_auth":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of authentication failures."},
			"acl_access_denied_cmd":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of commands rejected because of access denied to the command."},
			"acl_access_denied_key":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of commands rejected because of access denied to a key."},
			"acl_access_denied_channel":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of commands rejected because of access denied to a channel."},

			// # Replication
			// Not number, discard. "role": "Value is `master` if the instance is replica of no one, or `slave` if the instance is a replica of some master instance."},
			// Not number, discard. "master_failover_state": "The state of an ongoing failover, if any.."},
			// Not number, discard. "master_replid": "The replication ID of the Redis server.."},
			// Not number, discard. "master_replid2": "The secondary replication ID, used for PSYNC after a failover.."},
			"master_repl_offset":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The server's current replication offset"},
			"second_repl_offset":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The offset up to which replication IDs are accepted."},
			"repl_backlog_active":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Flag indicating replication backlog is active."},
			"repl_backlog_size":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total size in bytes of the replication backlog buffer."},
			"repl_backlog_first_byte_offset": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The master offset of the replication backlog buffer."},
			"repl_backlog_histlen":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Size in bytes of the data in the replication backlog buffer"},
			// Not number, discard. "master_host":        "Host or IP address of the master."},
			// Not number, discard. "master_port":        "Master listening TCP port."},
			// Not number, discard. "master_link_status": "Status of the link (up/down)."},
			"master_last_io_seconds_ago":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Number of seconds since the last interaction with master"},
			"master_sync_in_progress":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Indicate the master is syncing to the replica"},
			"slave_read_repl_offset":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The read replication offset of the replica instance.."},
			"slave_repl_offset":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The replication offset of the replica instance"},
			"slave_priority":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The priority of the instance as a candidate for failover."},
			"slave_read_only":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Flag indicating if the replica is read-only."},
			"replica_announced":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Flag indicating if the replica is announced by Sentinel.."},
			"master_sync_total_bytes":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total number of bytes that need to be transferred. this may be 0 when the size is unknown (for example, when the `repl-diskless-sync` configuration directive is used)."},
			"master_sync_read_bytes":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Number of bytes already transferred."},
			"master_sync_left_bytes":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Number of bytes left before syncing is complete (may be negative when master_sync_total_bytes is 0)"},
			"master_sync_perc":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The percentage master_sync_read_bytes from master_sync_total_bytes, or an approximation that uses loading_rdb_used_mem when master_sync_total_bytes is 0."},
			"master_sync_last_io_seconds_ago": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Number of seconds since last transfer I/O during a SYNC operation."},
			"master_link_down_since_seconds":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Number of seconds since the link is down."},
			"connected_slaves":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of connected replicas"},
			"min_slaves_good_slaves":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of replicas currently considered good."},
			// Not number, discard. "slaveXXX": "id, IP address, port, state, offset, lag."},

			// # CPU
			"used_cpu_sys":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationSecond, Desc: "System CPU consumed by the Redis server, which is the sum of system CPU consumed by all threads of the server process (main thread and background threads)."},
			"used_cpu_user":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationSecond, Desc: "User CPU consumed by the Redis server, which is the sum of user CPU consumed by all threads of the server process (main thread and background threads)."},
			"used_cpu_sys_children":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationSecond, Desc: "System CPU consumed by the background processes."},
			"used_cpu_user_children":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationSecond, Desc: "User CPU consumed by the background processes."},
			"used_cpu_sys_main_thread":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationSecond, Desc: "System CPU consumed by the Redis server main thread."},
			"used_cpu_user_main_thread": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationSecond, Desc: "User CPU consumed by the Redis server main thread."},

			// # Modules

			// # Commandstats
			// already in redis_command_stat

			// # Errorstats
			"errorstat": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Track of the different errors that occurred within Redis."},

			// # Latencystats
			"latency_percentiles_usec": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Latency percentile distribution statistics based on the command type."},

			// # Cluster
			"cluster_enabled": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Bool, Desc: "Indicate Redis cluster is enabled."},

			// # Keyspace
			// already in redis_db

			// # Calculate metric
			"info_latency_ms":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "The latency of the redis INFO command."},
			"used_cpu_sys_percent":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.Percent, Desc: "System CPU percentage consumed by the Redis server, which is the sum of system CPU consumed by all threads of the server process (main thread and background threads)"},
			"used_cpu_user_percent": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Rate, Unit: inputs.Percent, Desc: "User CPU percentage consumed by the Redis server, which is the sum of user CPU consumed by all threads of the server process (main thread and background threads)"},

			// # Old version metric
			"client_biggest_input_buf":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Biggest input buffer among current client connections"},
			"client_longest_output_list": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Longest output list among current client connections"},
		},
		Tags: map[string]interface{}{
			"host":          &inputs.TagInfo{Desc: "Hostname."},
			"redis_version": &inputs.TagInfo{Desc: "Version of the Redis server."},
			"server":        &inputs.TagInfo{Desc: "Server addr."},
			"service_name":  &inputs.TagInfo{Desc: "Service name."},
			"command_type":  &inputs.TagInfo{Desc: "Command type."},
			"error_type":    &inputs.TagInfo{Desc: "Error type."},
			"quantile":      &inputs.TagInfo{Desc: "Histogram `quantile`."},
		},
	}
}

func (m *infoMeasurement) getData() ([]*point.Point, error) {
	start := time.Now()
	ctx := context.Background()

	info, err := m.cli.Info(ctx, "ALL").Result()
	if err != nil {
		l.Error("redis exec command `All`, happen error,", err)
		return nil, err
	}
	elapsed := time.Since(start)

	nextTS := start.Add(elapsed / 2)

	latencyMs := Round(float64(elapsed)/float64(time.Millisecond), 2)

	return m.parseInfoData(info, latencyMs, nextTS)
}

func (m *infoMeasurement) parseInfoData(info string, latencyMs float64, nextTS time.Time) ([]*point.Point, error) {
	collectCache := []*point.Point{}
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(time.Now()))

	var kvs point.KVs

	rdr := strings.NewReader(info)
	scanner := bufio.NewScanner(rdr)
	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		val := strings.TrimSpace(parts[1])

		if strings.HasPrefix(key, "errorstat_") {
			collectCache = append(collectCache, m.getErrorPoints(key, val)...)
			continue
		}
		if strings.HasPrefix(key, "latency_percentiles_usec_") {
			collectCache = append(collectCache, m.getLatencyPoints(key, val)...)
			continue
		}

		val = strings.TrimSuffix(val, "%")

		float, err := strconv.ParseFloat(val, 64)
		if err != nil {
			if key == "redis_version" {
				if val == "" {
					val = "unknown"
				}
				kvs = kvs.AddTag("redis_version", val)
				m.tags["redis_version"] = val
			}
			continue
		}

		if _, has := infoFieldMap[key]; has {
			// key in the MeasurementInfo
			if strings.HasSuffix(key, "_kbps") {
				float *= 1000
			}
			kvs = kvs.Add(key, float, false, false)
		}
	}

	kvs = kvs.Add("info_latency_ms", latencyMs, false, false)

	// Calculate CPU usage.
	if usedCPUSys := kvs.Get("used_cpu_sys"); usedCPUSys != nil {
		usedCPUSysFloat := usedCPUSys.GetF()
		totElapsed := nextTS.Sub(m.lastCollect.usedCPUSysTS)
		if !m.lastCollect.usedCPUSysTS.IsZero() {
			kvs = kvs.Add("used_cpu_sys_percent", (usedCPUSysFloat-m.lastCollect.usedCPUSys)/totElapsed.Seconds(), false, false)
		}
		m.lastCollect.usedCPUSys = usedCPUSysFloat
		m.lastCollect.usedCPUSysTS = nextTS
	}

	if usedCPUUser := kvs.Get("used_cpu_user"); usedCPUUser != nil {
		usedCPUUserFloat := usedCPUUser.GetF()
		totElapsed := nextTS.Sub(m.lastCollect.usedCPUUserTS)
		if !m.lastCollect.usedCPUUserTS.IsZero() {
			kvs = kvs.Add("used_cpu_user_percent", (usedCPUUserFloat-m.lastCollect.usedCPUUser)/totElapsed.Seconds(), false, false)
		}
		m.lastCollect.usedCPUUser = usedCPUUserFloat
		m.lastCollect.usedCPUUserTS = nextTS
	}

	for k, v := range m.tags {
		kvs = kvs.AddTag(k, v)
	}
	collectCache = append(collectCache, point.NewPointV2(m.name, kvs, opts...))

	return collectCache, nil
}

// example data: errorstat_ERR:count=188
// want point: `redis_info,count_type=count,error_type=ERR errorstat=188i timestamp`.
func (m *infoMeasurement) getErrorPoints(k, v string) []*point.Point {
	pts := make([]*point.Point, 0)
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(time.Now()))
	var kvs point.KVs

	s := v[strings.LastIndex(v, "=")+1:]
	value, err := strconv.Atoi(s)
	if err != nil {
		l.Debugf("getErrorPoints string: %s , err: %s", s, err)
		return pts
	}
	kvs = kvs.Add("errorstat", value, false, true)

	s = strings.TrimPrefix(k, "errorstat_")
	kvs = kvs.AddTag("error_type", s)

	for k, v := range m.tags {
		kvs = kvs.AddTag(k, v)
	}
	pts = append(pts, point.NewPointV2(m.name, kvs, opts...))

	return pts
}

// example data: latency_percentiles_usec_client|list:p50=23.039,p99=70.143,p99.9=70.143
// example data: latency_percentiles_usec_ping:p50=1.003,p99=2.007,p99.9=2.007
//
// want point: `redis_info,command_type=client|list,quantile=0.5 latency_percentiles_usec=23.039 timestamp`
// want point: `redis_info,command_type=client|list,quantile=0.99 latency_percentiles_usec=70.143 timestamp`
// want point: `redis_info,command_type=client|list,quantile=0.999 latency_percentiles_usec=70.143 timestamp`
// ...
func (m *infoMeasurement) getLatencyPoints(k, v string) []*point.Point {
	if !m.latencyPercentiles {
		return make([]*point.Point, 0)
	}

	pts := make([]*point.Point, 0)
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(time.Now()))

	// get tag in k
	// example data: latency_percentiles_usec_client|list
	commandType := "unknow"
	s := strings.TrimPrefix(k, "latency_percentiles_usec_")
	if s != "" {
		commandType = s // client|list / ping ...
	}

	// get tags & fields in v
	// example data: p50=23.039,p99=70.143,p99.9=70.143
	tagField := getLatencyTagField(v)

	// make points
	for key, value := range tagField {
		var kvs point.KVs
		kvs = kvs.Add("latency_percentiles_usec", value, false, true)
		kvs = kvs.AddTag("command_type", commandType)
		kvs = kvs.AddTag("quantile", key)

		for k, v := range m.tags {
			kvs = kvs.AddTag(k, v)
		}
		pts = append(pts, point.NewPointV2(m.name, kvs, opts...))
	}

	return pts
}

// example data: p50=23.039,p99=70.143,p99.9=70.143
// want get: map[string]float64{"0.5":23.039,"0.99":70.143,"0.999":70.143}.
func getLatencyTagField(v string) map[string]float64 {
	slic := strings.Split(v, ",")
	tagField := map[string]float64{}

	for _, temp := range slic {
		items := strings.Split(temp, "=")
		if len(items) != 2 || items[0] == "" || items[1] == "" {
			l.Debugf("getLatencyTagField error, string: %s", temp)
			continue
		}

		// quantile
		quantile, err := getQuantile(strings.TrimPrefix(items[0], "p"))
		if err != nil {
			l.Debugf("getLatencyTagField string: %s , err: %s", items, err)
			continue
		}

		// value
		if value, err := strconv.ParseFloat(items[1], 64); err == nil {
			tagField[quantile] = value
		} else {
			l.Debugf("getLatencyTagField string: %s , err: %s", items, err)
		}
	}

	return tagField
}

func getQuantile(s string) (string, error) {
	s = "000" + s
	dotIndex := strings.Index(s, ".")
	if dotIndex == -1 {
		dotIndex = len(s) - 2
	} else {
		dotIndex -= 2
		s = strings.Replace(s, ".", "", 1)
	}
	s = s[:dotIndex] + "." + s[dotIndex:]

	quantile, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return "", err
	}

	q := fmt.Sprintf("%.8f", quantile)
	if strings.Contains(q, ".") {
		q = strings.TrimRight(q, "0")
		r := q[strings.Index(q, ".")+1:]
		_ = r
		if len(q[strings.Index(q, ".")+1:]) > 6 {
			return "", fmt.Errorf("max 6 decimal places %s", q)
		}
	}

	return strings.TrimRight(q, "."), nil
}

var infoFieldMap = map[string]struct{}{}

func getInfoFieldMap() {
	m := infoMeasurement{}
	for k := range m.Info().Fields {
		infoFieldMap[k] = struct{}{}
	}
}

func getInfoField() []string {
	keys := []string{}
	m := infoMeasurement{}
	for k := range m.Info().Fields {
		keys = append(keys, k)
	}
	return keys
}
