### 简介
oracle指标采集，参考datadog提供的指标

### 配置
```
[[inputs.oracle]]
    ## @param server - string - required
    ## The IP address or hostname of the Oracle Database Server.
    #
    server = "localhost:1521"

    ## @param service_name - string - required
    ## The Oracle Database service name. To view the services available on your server,
    ## run the following query: `SELECT value FROM v$parameter WHERE name='service_names'`
    #
    service_name = "<SERVICE_NAME>"

    ## @param user - string - required
    ## The username for the Datadog user account.
    #
    user = "<USER>"

    ## @param password - string - required
    ## The password for the Datadog user account.
    #
    password = "<PASSWORD>"

    ## @param service - string - optional
    ## Attach the tag `service:<SERVICE>` to every metric, event, and service check emitted by this integration.
    ##
    #
    service = "<SERVICE>"

    ## @param interval - number - optional - default: 15
    ## This changes the collection interval of the check. For more information, see:
    #
    interval = "15m"
```

### [Metrics]

system metrics

| 指标  | 单位 | 类型 | 标签 | 描述  | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| **buffer_cachehit_ratio** | - |  |addr  | Ratio of buffer cache hits |
| **cursor_cachehit_ratio** | - |  |addr  | Ratio of cursor cache hits |
| **library_cachehit_ratio** | - |   |addr  | Ratio of library cache hits |
| **shared_pool_free** |  - |  |addr  | shared pool free memory % |
| **physical_reads** | - |  |addr  | physical reads per sec |
| **physical_writes** | - |  |addr  | physical writes per sec |
| **enqueue_timeouts** | - |  |addr  | enqueue timeouts per sec |
| **gc_cr_block_received** | - |  |addr  | GC CR block received |
| **cache_blocks_corrupt** | - |  |addr  | corrupt cache blocks |
| **cache_blocks_lost** | - |  |addr  | lost cache blocks |
| **logons** | - |  |addr  | number of logon attempts |
| **active_sessions** | - |  |addr  | number of active sessions |
| **long_table_scans** | - |  |addr  | number of long table scans per sec |
| **service_response_time** | - |  |addr  | service response time |
| **user_rollbacks** | - |  |addr  | number of user rollbacks |
| **sorts_per_user_call** | - |  |addr  | sorts per user call |
| **rows_per_sort** | - |  |addr  | rows per sort |
| **disk_sorts** | - |  |addr  | disk sorts per second |
| **memory_sorts_ratio** | - |  |addr  | memory sorts ratio |
| **database_wait_time_ratio** | - |  |addr  | memory sorts per second |
| **session_limit_usage** | - |  |addr  | session limit usage |
| **session_count** | - |  |addr  | session count |

process metrics

| 指标  | 单位 | 类型 | 标签 | 描述  | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| **process_pga_used_memory** | - |  |addr, program  | PGA memory used by process |
| **process_pga_allocated_memory** | - |  |addr, program  | PGA memory allocated by process |
| **process_pga_freeable_memory** | - |  |addr, program  | PGA memory freeable by process |
| **process_pga_maximum_memory** | - |  |addr, program  | PGA maximum memory ever allocated by process |

tableSpace metrics

| 指标  | 单位 | 类型 | 标签 | 描述  | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| **temp_space_used** | - |  |addr, tablespace  | temp space used |
| **tablespace_used** | - |  |addr, tablespace  | tablespace used |
| **tablespace_size** | - |  |addr, tablespace  | tablespace size |
| **tablespace_in_use** | - |  | addr, tablespace | tablespace in-use |
| **tablespace_offline** | - |  |addr, tablespace  | tablespace offline |





