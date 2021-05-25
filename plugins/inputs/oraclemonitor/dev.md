### 简介
oracle指标采集，参考datadog提供的指标

### 配置
```
再扩展自定义sql,数据采集功能
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
tag: addr

| 指标  | 单位 | 类型 | 标签 | 描述  | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| **buffer_cachehit_ratio** | - |  |Ratio of buffer cache hits |
| **cursor_cachehit_ratio** | - |  |Ratio of cursor cache hits |
| **library_cachehit_ratio** | - |   |Ratio of library cache hits |
| **shared_pool_free** |  - |  |shared pool free memory % |
| **physical_reads** | - |  |physical reads per sec |
| **physical_writes** | - |  |physical writes per sec |
| **enqueue_timeouts** | - |  |enqueue timeouts per sec |
| **gc_cr_block_received** | - |  |GC CR block received |
| **cache_blocks_corrupt** | - |  |corrupt cache blocks |
| **cache_blocks_lost** | - |  |lost cache blocks |
| **logons** | - |  |number of logon attempts |
| **active_sessions** | - |  |number of active sessions |
| **long_table_scans** | - |  |number of long table scans per sec |
| **service_response_time** | - |  |service response time |
| **user_rollbacks** | - |  |number of user rollbacks |
| **sorts_per_user_call** | - |  |sorts per user call |
| **rows_per_sort** | - |  |rows per sort |
| **disk_sorts** | - |  |disk sorts per second |
| **memory_sorts_ratio** | - |  |memory sorts ratio |
| **database_wait_time_ratio** | - |  |memory sorts per second |
| **session_limit_usage** | - |  |session limit usage |
| **session_count** | - |  |session count |

process metrics
tag: addr, program  

| 指标  | 单位 | 类型 | 描述  | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | 
| **process_pga_used_memory** | - |  |PGA memory used by process |
| **process_pga_allocated_memory** | - |  |PGA memory allocated by process |
| **process_pga_freeable_memory** | - |  |PGA memory freeable by process |
| **process_pga_maximum_memory** | - |  |PGA maximum memory ever allocated by process |

tableSpace metrics
tag: addr, tablespace

| 指标  | 单位 | 类型 | 描述  | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | 
| **temp_space_used** | - |  | temp space used |
| **tablespace_used** | - |  |tablespace used |
| **tablespace_size** | - |  |tablespace size |
| **tablespace_in_use** | - |  | tablespace in-use |
| **tablespace_offline** | - |  |tablespace offline |





