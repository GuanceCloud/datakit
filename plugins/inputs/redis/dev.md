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
tag: addr, customTag

| 指标  | 单位 | 类型 | 标签 | 描述  | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| **active_defrag_running** | int | - |  |Whether active defragmentation is running or not. | info all | todo |
| **active_defrag_hits** | int | - |  | Number of value reallocations performed by the active defragmentation process. | info all | todo |
| **active_defrag_misses** | int | - |  | Number of aborted value reallocations started by the active defragmentation process. | info all | todo |
| **active_defrag_key_hits** | int | - |  | Number of keys that were actively defragmented. | info all | todo |
| **active_defrag_key_misses** | int | - |  | Number of keys that were skipped by. | info all | todo |
| **aof_buffer_length** | int | MB |  | Size of the AOF buffer. | info all | todo |
| **aof_last_rewrite_time** | int | - |  | Duration of the last AOF rewrite operation in seconds | info all | todo |
| **aof_rewrite** | int | - |  | Flag indicating a AOF rewrite operation is on-going. | info all | todo |
| **aof_size** | int | MB |  | AOF current file size (aof_current_size). | info all | todo |
| **loading_total_bytes** | int | MB |  | The total amount of bytes already loaded. | info all | todo |
| **loading_loaded_bytes** | int | MB |  | The amount of bytes to load. | info all | todo |
| **loading_loaded_perc** | - | - |  | The percent loaded. | info all | todo |
| **loading_eta_seconds** | - | - |  | The estimated amount of time left to load. | info all | todo |
| **client_biggest_input_buf** | - | - |  | The biggest input buffer among current client connections. | info all | todo |
| **blocked_clients** | - | - |  | The number of connections waiting on a blocking call. | info all | todo |
| **client_longest_output_list** | - | - |  | The longest output list among current client connections. | info all | todo |
| **used_cpu_sys** | - | - |  | System CPU consumed by the Redis server. | info all | todo |
| **used_cpu_sys_children** | - | - |  | System CPU consumed by the background processes. | info all | todo |
| **used_cpu_user** | - | - |  | User CPU consumed by the Redis server. | info all | todo |
| **used_cpu_user_children** | - | - |  | User CPU consumed by the background processes. | info all | todo |
| **expires** | - | - |  | The number of keys with an expiration. | info all | todo |
| **info_latency_ms** | - | - |  | The latency of the redis INFO command. | - | todo |
| **keys** | - | - |  | The total number of keys. | info all | todo |
| **evicted_keys** | - | - |  | The total number of keys evicted due to the maxmemory limit. | info all | todo |
| **expired_keys** | - | - |  | The total number of keys expired from the db. | info all | todo |
| **mem_fragmentation_ratio** | - | - |  | Ratio between used_memory_rss and used_memory. | info all | todo |
| **used_memory_lua** | - | - |  | Amount of memory used by the Lua engine. | info all | todo |
| **maxmemory** | - | - |  | Maximum amount of memory allocated to the Redisdb system. | info all | todo |
| **used_memory_peak** | - | - |  | The peak amount of memory used by Redis. | info all | todo |
| **used_memory_rss** | - | - |  | Amount of memory that Redis allocated as seen by the os. | info all | todo |
| **used_memory** | - | - |  | Amount of memory allocated by Redis. | info all | todo |
| **used_memory_startup** | - | - |  | Amount of memory consumed by Redis at startup. | info all | todo |
| **used_memory_overhead** | - | - |  | Sum of all overheads allocated by Redis for managing its internal datastructures. | info all | todo |
| **connected_clients** | - | - |  | The number of connected clients (excluding slaves). | info all | todo |
| **total_commands_processed** | - | - |  | The number of commands processed by the server. | info all | todo |
| **instantaneous_ops_per_sec** | - | - |  | The number of commands processed by the server per second. | info all | todo |
| **rejected_connections** | - | - |  | The number of rejected connections. | info all | todo |
| **connected_slaves** | - | - |  | The number of connected slaves. | info all | todo |
| **latest_fork_usec** | - | - |  | The duration of the latest fork. | info all | todo |
| **pubsub_channels** | - | - |  | The number of active pubsub channels. | info all | todo |
| **pubsub_patterns** | - | - |  | The number of active pubsub patterns. | info all | todo |
| **rdb_bgsave_in_progress** | - | - |  | One if a bgsave is in progress and zero otherwise. | info all | todo |
| **rdb_changes_since_last_save** | - | - |  | The number of changes since the last background save. | info all | todo |
| **rdb_last_bgsave_time_sec** | - | - |  | Duration of the last bg_save operation. | info all | todo |
| **repl_backlog_histlen** | - | - |  | The amount of data in the backlog sync buffer. | info all | todo |
| **master_last_io_seconds_ago** | - | - |  | Amount of time since the last interaction with master. | info all | todo |
| **master_repl_offset** | - | - |  | The replication offset reported by the master. | info all | todo |
| **slave_repl_offset** | - | - |  | The replication offset reported by the slave. | info all | todo |
| **master_sync_in_progress** | - | - |  | One if a sync is in progress and zero otherwise. | info all | todo |
| **master_sync_left_bytes** | - | - |  | Amount of data left before syncing is complete. |
| **keyspace_hits** | - | - |  | The rate of successful lookups in the main db. | info all | todo |
| **keyspace_misses** | - | - |  | The rate of missed lookups in the main db. | info all | todo |


db维度key使用指标
tag: db

| 指标  | 单位 | 类型 |  描述  | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | 
| **persist_keys** | -  | int | The number of keys persisted (keys - expires). |
| **persist_percent** | % | float | Percentage of total keys that are persisted. |
| **expires_percent** | % | float | Percentage of total keys with an expiration. | 

command统计
tag: method

| 指标  | 单位 | 类型 |  描述  | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | 
| **command_calls** | - | - | The number of times a redis command has been called, tagged by 'command', e.g. 'command:append'. Enable in Agent's redisdb.yaml with the command_stats option. |
| **command_usec_per_call** |  - | - | The CPU time consumed per redis command call, tagged by 'command', e.g. 'command:append'. Enable in Agent's redisdb.yaml with the command_stats option. |

sloglog采集

| 指标  | 单位 | 类型 |  描述  | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | 
| **slowlog_micros** | int | - | slow log info |

key length
tag: keyName

| 指标  | 单位 | 类型 | 描述  | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | 
| **key_length** | - | int | The key length, scan find bigkey |

主从指标采集
tag: slave_addr, slave_id

| 指标  | 单位 | 类型 | 描述  | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | 
| **repl_delay** | - | - | The key length, scan find bigkey |
| **master_link_down_since_seconds** | - | - | Amount of time that the master link has been down. | info all | todo |



