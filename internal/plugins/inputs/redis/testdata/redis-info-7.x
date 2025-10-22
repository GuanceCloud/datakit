# Server
redis_version:7.0.13
redis_git_sha1:00000000
redis_git_dirty:0
redis_build_id:e9f5dfc882060196
redis_mode:standalone
os:Linux 3.10.0-957.el7.x86_64 x86_64
arch_bits:64
monotonic_clock:POSIX clock_gettime
multiplexing_api:epoll
atomicvar_api:c11-builtin
gcc_version:12.2.0
process_id:1
process_supervised:no
run_id:85339d49fa59b92c47a3da041dd85b104af97f27
tcp_port:6379
server_time_usec:1698652602514848
uptime_in_seconds:1048
uptime_in_days:0
hz:10
configured_hz:10
lru_clock:4153786
executable:/data/redis-server
config_file:/etc/redis/redis.conf
io_threads_active:1

# Clients
connected_clients:1
cluster_connections:0
maxclients:10000
client_recent_max_input_buffer:8
client_recent_max_output_buffer:0
blocked_clients:0
tracking_clients:0
clients_in_timeout_table:0

# Memory
used_memory:1086408
used_memory_human:1.04M
used_memory_rss:6877184
used_memory_rss_human:6.56M
used_memory_peak:1115248
used_memory_peak_human:1.06M
used_memory_peak_perc:97.41%
used_memory_overhead:864264
used_memory_startup:862272
used_memory_dataset:222144
used_memory_dataset_perc:99.11%
allocator_allocated:1738944
allocator_active:2080768
allocator_resident:4939776
total_system_memory:16826019840
total_system_memory_human:15.67G
used_memory_lua:31744
used_memory_vm_eval:31744
used_memory_lua_human:31.00K
used_memory_scripts_eval:0
number_of_cached_scripts:0
number_of_functions:0
number_of_libraries:0
used_memory_vm_functions:32768
used_memory_vm_total:64512
used_memory_vm_total_human:63.00K
used_memory_functions:184
used_memory_scripts:184
used_memory_scripts_human:184B
maxmemory:0
maxmemory_human:0B
maxmemory_policy:noeviction
allocator_frag_ratio:1.20
allocator_frag_bytes:341824
allocator_rss_ratio:2.37
allocator_rss_bytes:2859008
rss_overhead_ratio:1.39
rss_overhead_bytes:1937408
mem_fragmentation_ratio:6.46
mem_fragmentation_bytes:5813192
mem_not_counted_for_evict:8
mem_replication_backlog:0
mem_total_replication_buffers:0
mem_clients_slaves:0
mem_clients_normal:1800
mem_cluster_links:0
mem_aof_buffer:8
mem_allocator:jemalloc-5.2.1
active_defrag_running:0
lazyfree_pending_objects:0
lazyfreed_objects:0

# Persistence
loading:0
async_loading:0
current_cow_peak:0
current_cow_size:0
current_cow_size_age:0
current_fork_perc:0.00
current_save_keys_processed:0
current_save_keys_total:0
rdb_changes_since_last_save:0
rdb_bgsave_in_progress:0
rdb_last_save_time:1698651554
rdb_last_bgsave_status:ok
rdb_last_bgsave_time_sec:-1
rdb_current_bgsave_time_sec:-1
rdb_saves:0
rdb_last_cow_size:0
rdb_last_load_keys_expired:0
rdb_last_load_keys_loaded:0
aof_enabled:1
aof_rewrite_in_progress:0
aof_rewrite_scheduled:0
aof_last_rewrite_time_sec:-1
aof_current_rewrite_time_sec:-1
aof_last_bgrewrite_status:ok
aof_rewrites:0
aof_rewrites_consecutive_failures:0
aof_last_write_status:ok
aof_last_cow_size:0
module_fork_in_progress:0
module_fork_last_cow_size:0
aof_current_size:89
aof_base_size:89
aof_pending_rewrite:0
aof_buffer_length:0
aof_pending_bio_fsync:0
aof_delayed_fsync:0

# Stats
total_connections_received:10
total_commands_processed:211
instantaneous_ops_per_sec:0
total_net_input_bytes:6445
total_net_output_bytes:407193
total_net_repl_input_bytes:0
total_net_repl_output_bytes:0
instantaneous_input_kbps:0.00
instantaneous_output_kbps:0.00
instantaneous_input_repl_kbps:0.00
instantaneous_output_repl_kbps:0.00
rejected_connections:0
sync_full:0
sync_partial_ok:0
sync_partial_err:0
expired_keys:0
expired_stale_perc:0.00
expired_time_cap_reached_count:0
expire_cycle_cpu_milliseconds:29
evicted_keys:0
evicted_clients:0
total_eviction_exceeded_time:0
current_eviction_exceeded_time:0
keyspace_hits:0
keyspace_misses:0
pubsub_channels:0
pubsub_patterns:0
pubsubshard_channels:0
latest_fork_usec:0
total_forks:0
migrate_cached_sockets:0
slave_expires_tracked_keys:0
active_defrag_hits:0
active_defrag_misses:0
active_defrag_key_hits:0
active_defrag_key_misses:0
total_active_defrag_time:0
current_active_defrag_time:0
tracking_total_keys:0
tracking_total_items:0
tracking_total_prefixes:0
unexpected_error_replies:0
total_error_replies:8
dump_payload_sanitizations:0
total_reads_processed:221
total_writes_processed:213
io_threaded_reads_processed:0
io_threaded_writes_processed:0
reply_buffer_shrinks:28
reply_buffer_expands:26

# Replication
role:master
connected_slaves:0
master_failover_state:no-failover
master_replid:b853cdaf3b3ea32e2c4064f4fe6bb75ec9278ce7
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:0
second_repl_offset:-1
repl_backlog_active:0
repl_backlog_size:1048576
repl_backlog_first_byte_offset:0
repl_backlog_histlen:0

# CPU
used_cpu_sys:1.300455
used_cpu_user:1.717573
used_cpu_sys_children:0.004007
used_cpu_user_children:0.007766
used_cpu_sys_main_thread:1.296392
used_cpu_user_main_thread:1.714453

# Modules

# Commandstats
cmdstat_ping:calls=1,usec=2,usec_per_call=2.00,rejected_calls=0,failed_calls=0
cmdstat_auth:calls=9,usec=549,usec_per_call=61.00,rejected_calls=0,failed_calls=8
cmdstat_latency|latest:calls=28,usec=173,usec_per_call=6.18,rejected_calls=0,failed_calls=0
cmdstat_command|docs:calls=1,usec=1829,usec_per_call=1829.00,rejected_calls=0,failed_calls=0
cmdstat_client|list:calls=28,usec=1035,usec_per_call=36.96,rejected_calls=0,failed_calls=0
cmdstat_info:calls=113,usec=19830,usec_per_call=175.49,rejected_calls=0,failed_calls=0
cmdstat_slowlog|get:calls=28,usec=169,usec_per_call=6.04,rejected_calls=0,failed_calls=0
cmdstat_acl|setuser:calls=3,usec=226,usec_per_call=75.33,rejected_calls=0,failed_calls=0

# Errorstats
errorstat_WRONGPASS:count=8

# Latencystats
latency_percentiles_usec_ping:p50=2.007,p99=2.007,p99.9=2.007
latency_percentiles_usec_auth:p50=51.199,p99=114.175,p99.9=114.175
latency_percentiles_usec_latency|latest:p50=4.015,p99=30.079,p99.9=30.079
latency_percentiles_usec_command|docs:p50=1835.007,p99=1835.007,p99.9=1835.007
latency_percentiles_usec_client|list:p50=35.071,p99=88.063,p99.9=88.063
latency_percentiles_usec_info:p50=97.279,p99=581.631,p99.9=708.607
latency_percentiles_usec_slowlog|get:p50=5.023,p99=24.063,p99.9=24.063
latency_percentiles_usec_acl|setuser:p50=78.335,p99=103.423,p99.9=103.423

# Cluster
cluster_enabled:0

# Keyspace
