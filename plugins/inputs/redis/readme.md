### 简介
redis指标采集，参考datadog提供的指标

### 配置
```
[[inputs.redis]]
    ## @param host - string - required
    ## Enter the host to connect to.
    host = "xxxxx"
    ## @param port - integer - required
    ## Enter the port of the host to connect to.
    port = 6379

    ## @param username - string - optional
    ## The username to use for the connection. Redis 6+ only.
    #
    # username: <USERNAME>

    ## @param password - string - optional
    ## The password to use for the connection.
    #
    # password: <PASSWORD>
```

###  收集指标
| 指标  | 单位 | 类型 | 标签 | 描述  | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| **active_defrag_running** | int | - | server |Whether active defragmentation is running or not. | info all | todo |
| **active_defrag_hits** | int | - | server | Number of value reallocations performed by the active defragmentation process. | info all | todo |
| **active_defrag_misses** | int | - | server | Number of aborted value reallocations started by the active defragmentation process. | info all | todo |
| **active_defrag_key_hits** | int | - | server | Number of keys that were actively defragmented. | info all | todo |
| **active_defrag_key_misses** | int | - | server | Number of keys that were skipped by. | info all | todo |
| **aof_buffer_length** | int | MB | server | Size of the AOF buffer. | info all | todo |
| **aof_last_rewrite_time** | int | - | server | Duration of the last AOF rewrite operation in seconds | info all | todo |
| **aof_rewrite** | int | - | server | Flag indicating a AOF rewrite operation is on-going. | info all | todo |
| **aof_size** | int | MB | server | AOF current file size (aof_current_size). | info all | todo |
| **loading_total_bytes** | int | MB | server | The total amount of bytes already loaded. | info all | todo |
| **loading_loaded_bytes** | int | MB | server | The amount of bytes to load. | info all | todo |
| **loading_loaded_perc** | - | - | server | The percent loaded. | info all | todo |
| **loading_eta_seconds** | - | - | server | The estimated amount of time left to load. | info all | todo |
| **client_biggest_input_buf** | - | - | server | The biggest input buffer among current client connections. | info all | todo |
| **blocked_clients** | - | - | server | The number of connections waiting on a blocking call. | info all | todo |
| **client_longest_output_list** | - | - | server | The longest output list among current client connections. | info all | todo |
| **used_cpu_sys** | - | - | server | System CPU consumed by the Redis server. | info all | todo |
| **used_cpu_sys_children** | - | - | server | System CPU consumed by the background processes. | info all | todo |
| **used_cpu_user** | - | - | server | User CPU consumed by the Redis server. | info all | todo |
| **used_cpu_user_children** | - | - | server | User CPU consumed by the background processes. | info all | todo |
| **expires** | - | - | server | The number of keys with an expiration. | info all | todo |
| **expires_percent** | - | - | server | Percentage of total keys with an expiration. | - | todo |
| **info_latency_ms** | - | - | server | The latency of the redis INFO command. | - | todo |
| **key_length** | - | - | server | The number of elements in a given key, tagged by key, e.g. 'key:mykeyname'. Enable in Agent's redisdb.yaml with the keys option. | - | todo |
| **keys** | - | - | server | The total number of keys. | info all | todo |
| **evicted_keys** | - | - | server | The total number of keys evicted due to the maxmemory limit. | info all | todo |
| **expired_keys** | - | - | server | The total number of keys expired from the db. | info all | todo |
| **mem_fragmentation_ratio** | - | - | server | Ratio between used_memory_rss and used_memory. | info all | todo |
| **used_memory_lua** | - | - | server | Amount of memory used by the Lua engine. | info all | todo |
| **maxmemory** | - | - | server | Maximum amount of memory allocated to the Redisdb system. | info all | todo |
| **used_memory_peak** | - | - | server | The peak amount of memory used by Redis. | info all | todo |
| **used_memory_rss** | - | - | server | Amount of memory that Redis allocated as seen by the os. | info all | todo |
| **used_memory** | - | - | server | Amount of memory allocated by Redis. | info all | todo |
| **used_memory_startup** | - | - | server | Amount of memory consumed by Redis at startup. | info all | todo |
| **used_memory_overhead** | - | - | server | Sum of all overheads allocated by Redis for managing its internal datastructures. | info all | todo |
| **connected_clients** | - | - | server | The number of connected clients (excluding slaves). | info all | todo |
| **total_commands_processed** | - | - | server | The number of commands processed by the server. | info all | todo |
| **instantaneous_ops_per_sec** | - | - | server | The number of commands processed by the server per second. | info all | todo |
| **connections** | - | - | server | The number of connections tagged by client name. |
| **rejected_connections** | - | - | server | The number of rejected connections. | info all | todo |
| **connected_slaves** | - | - | server | The number of connected slaves. | info all | todo |
| **maxclients** | - | - | server | The maximum number of connected clients. |
| **latest_fork_usec** | - | - | server | The duration of the latest fork. | info all | todo |
| **persist_keys** | - | - | server | The number of keys persisted (keys - expires). |
| **persist_percent** | - | - | server | Percentage of total keys that are persisted. |
| **pubsub_channels** | - | - | server | The number of active pubsub channels. | info all | todo |
| **pubsub_patterns** | - | - | server | The number of active pubsub patterns. | info all | todo |
| **rdb_bgsave_in_progress** | - | - | server | One if a bgsave is in progress and zero otherwise. | info all | todo |
| **rdb_changes_since_last_save** | - | - | server | The number of changes since the last background save. | info all | todo |
| **rdb_last_bgsave_time_sec** | - | - | server | Duration of the last bg_save operation. | info all | todo |
| **repl_backlog_histlen** | - | - | server | The amount of data in the backlog sync buffer. | info all | todo |
| **repl_delay** | - | - | server | The replication delay in offsets. |
| **master_last_io_seconds_ago** | - | - | server | Amount of time since the last interaction with master. | info all | todo |
| **master_link_down_since_seconds** | - | - | server | Amount of time that the master link has been down. | info all | todo |
| **master_repl_offset** | - | - | server | The replication offset reported by the master. | info all | todo |
| **slave_repl_offset** | - | - | server | The replication offset reported by the slave. | info all | todo |
| **master_sync_in_progress** | - | - | server | One if a sync is in progress and zero otherwise. | info all | todo |
| **master_sync_left_bytes** | - | - | server | Amount of data left before syncing is complete. |
| **95percentile** | - | - | server | The 95th percentile of the duration of queries reported in the slow log. |
| **micros_avg** | - | - | server | The average duration of queries reported in the slow log. |
| **micros_count** | - | - | server | The rate of queries reported in the slow log. |
| **micros_max** | - | - | server | The maximum duration of queries reported in the slow log. |
| **micros_median** | - | - | server | The median duration of queries reported in the slow log. |
| **keyspace_hits** | - | - | server | The rate of successful lookups in the main db. | info all | todo |
| **keyspace_misses** | - | - | server | The rate of missed lookups in the main db. | info all | todo |
| **command_calls** | - | - | server | The number of times a redis command has been called, tagged by 'command', e.g. 'command:append'. Enable in Agent's redisdb.yaml with the command_stats option. |
| **command_usec_per_call** |  - | - | server | The CPU time consumed per redis command call, tagged by 'command', e.g. 'command:append'. Enable in Agent's redisdb.yaml with the command_stats option. |
