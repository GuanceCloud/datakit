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

实例维度指标

| 指标  | 单位 | 类型 | 标签 | 描述  | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| **active_defrag_running** | int | - | addr, customTag |Whether active defragmentation is running or not. | info all | todo |
| **active_defrag_hits** | int | - | addr, customTag | Number of value reallocations performed by the active defragmentation process. | info all | todo |
| **active_defrag_misses** | int | - | addr, customTag | Number of aborted value reallocations started by the active defragmentation process. | info all | todo |
| **active_defrag_key_hits** | int | - | addr, customTag | Number of keys that were actively defragmented. | info all | todo |
| **active_defrag_key_misses** | int | - | addr, customTag | Number of keys that were skipped by. | info all | todo |
| **aof_buffer_length** | int | MB | addr, customTag | Size of the AOF buffer. | info all | todo |
| **aof_last_rewrite_time** | int | - | addr, customTag | Duration of the last AOF rewrite operation in seconds | info all | todo |
| **aof_rewrite** | int | - | addr, customTag | Flag indicating a AOF rewrite operation is on-going. | info all | todo |
| **aof_size** | int | MB | addr, customTag | AOF current file size (aof_current_size). | info all | todo |
| **loading_total_bytes** | int | MB | addr, customTag | The total amount of bytes already loaded. | info all | todo |
| **loading_loaded_bytes** | int | MB | addr, customTag | The amount of bytes to load. | info all | todo |
| **loading_loaded_perc** | - | - | addr, customTag | The percent loaded. | info all | todo |
| **loading_eta_seconds** | - | - | addr, customTag | The estimated amount of time left to load. | info all | todo |
| **client_biggest_input_buf** | - | - | addr, customTag | The biggest input buffer among current client connections. | info all | todo |
| **blocked_clients** | - | - | addr, customTag | The number of connections waiting on a blocking call. | info all | todo |
| **client_longest_output_list** | - | - | addr, customTag | The longest output list among current client connections. | info all | todo |
| **used_cpu_sys** | - | - | addr, customTag | System CPU consumed by the Redis server. | info all | todo |
| **used_cpu_sys_children** | - | - | addr, customTag | System CPU consumed by the background processes. | info all | todo |
| **used_cpu_user** | - | - | addr, customTag | User CPU consumed by the Redis server. | info all | todo |
| **used_cpu_user_children** | - | - | addr, customTag | User CPU consumed by the background processes. | info all | todo |
| **expires** | - | - | addr, customTag | The number of keys with an expiration. | info all | todo |
| **info_latency_ms** | - | - | addr, customTag | The latency of the redis INFO command. | - | todo |
| **keys** | - | - | addr, customTag | The total number of keys. | info all | todo |
| **evicted_keys** | - | - | addr, customTag | The total number of keys evicted due to the maxmemory limit. | info all | todo |
| **expired_keys** | - | - | addr, customTag | The total number of keys expired from the db. | info all | todo |
| **mem_fragmentation_ratio** | - | - | addr, customTag | Ratio between used_memory_rss and used_memory. | info all | todo |
| **used_memory_lua** | - | - | addr, customTag | Amount of memory used by the Lua engine. | info all | todo |
| **maxmemory** | - | - | addr, customTag | Maximum amount of memory allocated to the Redisdb system. | info all | todo |
| **used_memory_peak** | - | - | addr, customTag | The peak amount of memory used by Redis. | info all | todo |
| **used_memory_rss** | - | - | addr, customTag | Amount of memory that Redis allocated as seen by the os. | info all | todo |
| **used_memory** | - | - | addr, customTag | Amount of memory allocated by Redis. | info all | todo |
| **used_memory_startup** | - | - | addr, customTag | Amount of memory consumed by Redis at startup. | info all | todo |
| **used_memory_overhead** | - | - | addr, customTag | Sum of all overheads allocated by Redis for managing its internal datastructures. | info all | todo |
| **connected_clients** | - | - | addr, customTag | The number of connected clients (excluding slaves). | info all | todo |
| **total_commands_processed** | - | - | addr, customTag | The number of commands processed by the server. | info all | todo |
| **instantaneous_ops_per_sec** | - | - | addr, customTag | The number of commands processed by the server per second. | info all | todo |
| **rejected_connections** | - | - | addr, customTag | The number of rejected connections. | info all | todo |
| **connected_slaves** | - | - | addr, customTag | The number of connected slaves. | info all | todo |
| **latest_fork_usec** | - | - | addr, customTag | The duration of the latest fork. | info all | todo |
| **pubsub_channels** | - | - | addr, customTag | The number of active pubsub channels. | info all | todo |
| **pubsub_patterns** | - | - | addr, customTag | The number of active pubsub patterns. | info all | todo |
| **rdb_bgsave_in_progress** | - | - | addr, customTag | One if a bgsave is in progress and zero otherwise. | info all | todo |
| **rdb_changes_since_last_save** | - | - | addr, customTag | The number of changes since the last background save. | info all | todo |
| **rdb_last_bgsave_time_sec** | - | - | addr, customTag | Duration of the last bg_save operation. | info all | todo |
| **repl_backlog_histlen** | - | - | addr, customTag | The amount of data in the backlog sync buffer. | info all | todo |
| **master_last_io_seconds_ago** | - | - | addr, customTag | Amount of time since the last interaction with master. | info all | todo |
| **master_repl_offset** | - | - | addr, customTag | The replication offset reported by the master. | info all | todo |
| **slave_repl_offset** | - | - | addr, customTag | The replication offset reported by the slave. | info all | todo |
| **master_sync_in_progress** | - | - | addr, customTag | One if a sync is in progress and zero otherwise. | info all | todo |
| **master_sync_left_bytes** | - | - | addr, customTag | Amount of data left before syncing is complete. |
| **keyspace_hits** | - | - | addr, customTag | The rate of successful lookups in the main db. | info all | todo |
| **keyspace_misses** | - | - | addr, customTag | The rate of missed lookups in the main db. | info all | todo |


db维度key使用指标

| 指标  | 单位 | 类型 | 标签 | 描述  | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| **persist_keys** | -  | int | addr, customTag, db | The number of keys persisted (keys - expires). |
| **persist_percent** | % | float | addr, customTag, db | Percentage of total keys that are persisted. |
| **expires_percent** | % | float | addr, customTag, db | Percentage of total keys with an expiration. | 

db维度key使用指标

| 指标  | 单位 | 类型 | 标签 | 描述  | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| **command_calls** | - | - | addr, customTag | The number of times a redis command has been called, tagged by 'command', e.g. 'command:append'. Enable in Agent's redisdb.yaml with the command_stats option. |
| **command_usec_per_call** |  - | - | addr, customTag | The CPU time consumed per redis command call, tagged by 'command', e.g. 'command:append'. Enable in Agent's redisdb.yaml with the command_stats option. |

sloglog采集

| 指标  | 单位 | 类型 | 标签 | 描述  | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| **slowlog_micros** | int | - | addr, customTag, command | slow log info |

key length

| 指标  | 单位 | 类型 | 标签 | 描述  | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| **key_length** | - | int | addr, customTag, keyName | The key length, scan find bigkey |

主从指标采集

| 指标  | 单位 | 类型 | 标签 | 描述  | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| **repl_delay** | - | - | addr, customTag, slave_addr, slave_id | The key length, scan find bigkey |
| **master_link_down_since_seconds** | - | - | addr, customTag, slave_addr, slave_id | Amount of time that the master link has been down. | info all | todo |



