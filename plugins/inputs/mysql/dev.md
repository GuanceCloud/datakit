## 简介
mysql指标采集，参考datadog提供的指标，提供默认指标收集和用户自定义查询

## 配置
再扩展自定义sql,数据采集功能
```
[[inputs.mysqlMonitor]]
    ## @param host - string - optional
    ## MySQL host to connect to.
    ## NOTE: Even if the host is "localhost", the agent connects to MySQL using TCP/IP, unless you also
    ## provide a value for the sock key (below).
    #
    host = "localhost"

    ## @param user - string - optional
    ## Username used to connect to MySQL.
    #
    user = "cc_monitor"

    ## @param pass - string - optional
    ## Password associated to the MySQL user.
    #
    pass = "<PASS>"

    ## @param port - number - optional - default: 3306
    ## Port to use when connecting to MySQL.
    #
    port = 3306

    ## @param sock - string - optional
    ## Path to a Unix Domain Socket to use when connecting to MySQL (instead of a TCP socket).
    ## If you specify a socket you dont need to specify a host.
    #
    # sock = "/tmp/mysql.sock"

    ## @param charset - string - optional
    ## Charset you want to use.
    #
    # charset = "utf8"

    ## @param connect_timeout - number - optional - default: 10
    ## Maximum number of seconds to wait before timing out when connecting to MySQL.
    #
    # connect_timeout = 10

    ## @param min_collection_interval - number - optional - default: 15
    ## This changes the collection interval of the check. For more information, see:
    #
    interval = 15

    ## @param ssl - mapping - optional
    ## Use this section to configure a TLS connection between the Agent and MySQL.
    ##
    ## The following fields are supported:
    ##
    ## key: Path to a key file.
    ## cert: Path to a cert file.
    ## ca: Path to a CA bundle file.
    #
    ## Optional TLS Config
    # [inputs.mysqlMonitor.tls]
    # tls_key = "/tmp/peer.key"
    # tls_cert = "/tmp/peer.crt"
    # tls_ca = "/tmp/ca.crt"

    ## @param service - string - optional
    ## Attach the tag `service:<SERVICE>` to every metric.
    #
    # service = "<SERVICE_NAME>""

    ## @param tags - list of strings - optional
    ## A list of tags to attach to every metric and service check emitted by this instance.
    #
    # [inputs.mysqlMonitor.tags]
    #   KEY_1 = "VALUE_1"
    #   KEY_2 = "VALUE_2"

    ## Enable options to collect extra metrics from your MySQL integration.
    #
    [inputs.mysqlMonitor.options]
        ## @param replication - boolean - optional - default: false
        ## Set to `true` to collect replication metrics.
        #
        replication = false

        ## @param replication_channel - string - optional
        ## If using multiple sources, set the channel name to monitor.
        #
        # replication_channel: <REPLICATION_CHANNEL>

        ## @param replication_non_blocking_status - boolean - optional - default: false
        ## Set to `true` to grab slave count in a non-blocking manner (requires `performance_schema`);
        #
        replication_non_blocking_status = false

        ## @param galera_cluster - boolean - optional - default: false
        ## Set to `true` to collect Galera cluster metrics.
        #
        galera_cluster = false

        ## @param extra_status_metrics - boolean - optional - default: false
        ## Set to `true` to enable extra status metrics.
        ##
        #
        extra_status_metrics = false

        ## @param extra_innodb_metrics - boolean - optional - default: false
        ## Set to `true` to enable extra InnoDB metrics.
        ##
        #
        extra_innodb_metrics = false

        ## @param disable_innodb_metrics - boolean - optional - default: false
        ## Set to `true` only if experiencing issues with older (unsupported) versions of MySQL
        ## that do not run or have InnoDB engine support.
        ##
        ## If this flag is enabled, you will only receive a small subset of metrics.
        ##
        #
        disable_innodb_metrics = false

        ## @param schema_size_metrics - boolean - optional - default: false
        ## Set to `true` to collect schema size metrics.
        ##
        ## Note that this runs a heavy query against your database to compute the relevant metrics
        ## for all your existing schemas. Due to the nature of these calls, if you have a
        ## high number of tables and schemas, this may have a negative impact on your database performance.
        ##
        #
        schema_size_metrics = false

        ## @param extra_performance_metrics - boolean - optional - default: false
        ## These metrics are reported if `performance_schema` is enabled in the MySQL instance
        ## and if the version for that instance is >= 5.6.0.
        ##
        ## Note that this runs a heavy query against your database to compute the relevant metrics
        ## for all your existing schemas. Due to the nature of these calls, if you have a
        ## high number of tables and schemas, this may have a negative impact on your database performance.
        ##
        ## Metrics provided by the options:
        ##   - mysql.info.schema.size (per schema)
        ##   - mysql.performance.query_run_time.avg (per schema)
        ##   - mysql.performance.digest_95th_percentile.avg_us
        ##
        ## Note that some of these require the `user` defined for this instance
        ## to have PROCESS and SELECT privileges. 
        #
        extra_performance_metrics = false 
```

##  收集指标
| 指标 | 单位 | 类型 | 标签 | 描述 | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | --- |
| mysql.info.schema.size |  |  |  | Size of schemas in MiB |  |  |
| mysql.galera.wsrep_cluster_size |  |  |  | The current number of nodes in the Galera cluster. |  |  |
| Innodb_buffer_pool_bytes_free |  |  |  | The number of free pages in the InnoDB Buffer Pool. |  |  |
| Innodb_buffer_pool_bytes_total |  |  |  | The total number of pages in the InnoDB Buffer Pool. |  |  |
| Innodb_buffer_pool_bytes_used |  |  |  | The number of used pages in the InnoDB Buffer Pool. |  |  |
| Innodb_buffer_pool_pages_utilization |  |  |  | The utilization of the InnoDB Buffer Pool. |  |  |
| Innodb_current_row_locks |  |  |  | The number of current row locks. |  |  |
| Innodb_data_reads |  |  |  | The rate of data reads. |  |  |
| Innodb_data_writes |  |  |  | The rate of data writes. |  |  |
| Innodb_mutex_os_waits |  |  |  | The rate of mutex OS waits. |  |  |
| Innodb_mutex_spin_rounds |  |  |  | The rate of mutex spin rounds. |  |  |
| Innodb_mutex_spin_waits |  |  |  | The rate of mutex spin waits. |  |  |
| Innodb_os_log_fsyncs |  |  |  | The rate of fsync writes to the log file. |  |  |
| Innodb_row_lock_time |  |  |  | Fraction of time spent (ms/s) acquiring row locks. |  |  |
| Innodb_row_lock_waits |  |  |  | The number of times per second a row lock had to be waited for. |  |  |
| connections |  |  |  | The rate of connections to the server. |  |  |
| max_connections |  |  |  | The maximum number of connections that have been in use simultaneously since the server started. |  |  |
| max_connections_available |  |  |  | The maximum permitted number of simultaneous client connections. |  |  |
| com_delete |  |  |  | The rate of delete statements. |  |  |
| com_delete_multi |  |  |  | The rate of delete-multi statements. |  |  |
| com_insert |  |  |  | The rate of insert statements. |  |  |
| com_insert_select |  |  |  | The rate of insert-select statements. |  |  |
| com_replace_select |  |  |  | The rate of replace-select statements. |  |  |
| com_select |  |  |  | The rate of select statements. |  |  |
| com_update |  |  |  | The rate of update statements. |  |  |
| com_update_multi |  |  |  | The rate of update-multi. |  |  |
| created_tmp_disk_tables |  |  |  | The rate of internal on-disk temporary tables created by second by the server while executing statements. |  |  |
| created_tmp_files |  |  |  | The rate of temporary files created by second. |  |  |
| created_tmp_tables |  |  |  | The rate of internal temporary tables created by second by the server while executing statements. |  |  |
| key_cache_utilization |  |  |  | The key cache utilization ratio. |  |  |
| open_files |  |  |  | The number of open files. |  |  |
| open_tables |  |  |  | The number of of tables that are open. |  |  |
| qcache_hits |  |  |  | The rate of query cache hits. |  |  |
| questions |  |  |  | The rate of statements executed by the server. |  |  |
| slow_queries |  |  |  | The rate of slow queries. |  |  |
| table_locks_waited |  |  |  | The total number of times that a request for a table lock could not be granted immediately and a wait was needed. |  |  |
| threads_connected |  |  |  | The number of currently open connections. |  |  |
| threads_running |  |  |  | The number of threads that are not sleeping. |  |  |
| user_time |  |  |  | Percentage of CPU time spent in user space by MySQL. |  |  |
| seconds_behind_master |  |  |  | The lag in seconds between the master and the slave. |  |  |
| slave_running |  |  |  | Deprecated. Use service check mysql.replication.replica_running instead. A boolean showing if this server is a replication slave / master that is running. |  |  |
| slaves_connected |  |  |  | Deprecated. Use mysql.replication.replicas_connected instead. Number of slaves connected to a replication master. |  |  |
| replicas_connected |  |  |  | Number of replicas connected to a replication source. |  |  |
| queries |  |  |  | The rate of queries. |  |  |
| com_replace |  |  |  | The rate of replace statements. |  |  |
| com_load |  |  |  | The rate of load statements. |  |  |
| aborted_clients |  |  |  | The number of connections that were aborted because the client died without closing the connection properly. |  |  |
| aborted_connects |  |  |  | The number of failed attempts to connect to the MySQL server. |  |  |
| bytes_sent |  |  |  | The number of bytes sent to all clients. |  |  |
| bytes_received |  |  |  | The number of bytes received from all clients. |  |  |
| qcache_inserts |  |  |  | The number of queries added to the query cache. |  |  |
| qcache_lowmem_prunes |  |  |  | The number of queries that were deleted from the query cache because of low memory. |  |  |
| key_read_requests |  |  |  | The number of requests to read a key block from the MyISAM key cache. |  |  |
| key_reads |  |  |  | The number of physical reads of a key block from disk into the MyISAM key cache. If Key_reads is large, then your key_buffer_size value is probably too small. The cache miss rate can be calculated as Key_reads/Key_read_requests. |  |  |
| key_write_requests |  |  |  | The number of requests to write a key block to the MyISAM key cache. |  |  |
| key_writes |  |  |  | The number of physical writes of a key block from the MyISAM key cache to disk. |  |  |
| key_buffer_size |  |  |  | Size of the buffer used for index blocks. |  |  |
| qcache_size |  |  |  | The amount of memory allocated for caching query results. |  |  |
| table_open_cache |  |  |  | The number of open tables for all threads. Increasing this value increases the number of file descriptors that mysqld requires. |  |  |
| thread_cache_size |  |  |  | How many threads the server should cache for reuse. When a client disconnects, the client's threads are put in the cache if there are fewer than thread_cache_size threads there. |  |  |
| Innodb_row_lock_current_waits |  |  |  | The number of row locks currently being waited for by operations on InnoDB tables. |  |  |
| Innodb_buffer_pool_bytes_dirty |  |  |  | The total current number of bytes held in dirty pages in the InnoDB buffer pool. |  |  |
| Innodb_buffer_pool_read_requests |  |  |  | The number of logical read requests. |  |  |
| Innodb_buffer_pool_reads |  |  |  | The number of logical reads that InnoDB could not satisfy from the buffer pool, and had to read directly from disk. |  |  |
| cache_disk_use |  |  |  | The number of transactions that used the temporary binary log cache but that exceeded the value of binlog_cache_size and used a temporary file to store statements from the transaction. |  |  |
| cache_use |  |  |  | The number of transactions that used the binary log cache. |  |  |
| handler_commit |  |  |  | The number of internal COMMIT statements. |  |  |
| handler_delete |  |  |  | The number of internal DELETE statements. |  |  |
| handler_prepare |  |  |  | The number of internal PREPARE statements. |  |  |
| handler_read_first |  |  |  | The number of internal READ_FIRST statements. |  |  |
| handler_read_key |  |  |  | The number of internal READ_KEY statements. |  |  |
| handler_read_next |  |  |  | The number of internal READ_NEXT statements. |  |  |
| handler_read_prev |  |  |  | The number of internal READ_PREV statements. |  |  |
| handler_read_rnd |  |  |  | The number of internal READ_RND statements. |  |  |
| handler_read_rnd_next |  |  |  | The number of internal READ_RND_NEXT statements. |  |  |
| handler_rollback |  |  |  | The number of internal ROLLBACK statements. |  |  |
| handler_update |  |  |  | The number of internal UPDATE statements. |  |  |
| handler_write |  |  |  | The number of internal WRITE statements. |  |  |
| opened_tables |  |  |  | The number of tables that have been opened. If Opened_tables is big, your table_open_cache value is probably too small. |  |  |
| qcache_total_blocks |  |  |  | The total number of blocks in the query cache. |  |  |
| qcache_free_blocks |  |  |  | The number of free memory blocks in the query cache. |  |  |
| qcache_free_memory |  |  |  | The amount of free memory for the query cache. |  |  |
| qcache_not_cached |  |  |  | The number of noncached queries (not cacheable, or not cached due to the query_cache_type setting). |  |  |
| qcache_queries_in_cache |  |  |  | The number of queries registered in the query cache. |  |  |
| select_full_join |  |  |  | The number of joins that perform table scans because they do not use indexes. If this value is not 0, you should carefully check the indexes of your tables. |  |  |
| select_full_range_join |  |  |  | The number of joins that used a range search on a reference table. |  |  |
| select_range |  |  |  | The number of joins that used ranges on the first table. This is normally not a critical issue even if the value is quite large. |  |  |
| select_range_check |  |  |  | The number of joins without keys that check for key usage after each row. If this is not 0, you should carefully check the indexes of your tables. |  |  |
| select_scan |  |  |  | The number of joins that did a full scan of the first table. |  |  |
| sort_merge_passes |  |  |  | The number of merge passes that the sort algorithm has had to do. If this value is large, you should consider increasing the value of the sort_buffer_size system variable. |  |  |
| sort_range |  |  |  | The number of sorts that were done using ranges. |  |  |
| sort_rows |  |  |  | The number of sorted rows. |  |  |
| sort_scan |  |  |  | The number of sorts that were done by scanning the table. |  |  |
| table_locks_immediate |  |  |  | The number of times that a request for a table lock could be granted immediately. |  |  |
| threads_cached |  |  |  | The number of threads in the thread cache. |  |  |
| threads_created |  |  |  | The number of threads created to handle connections. If Threads_created is big, you may want to increase the thread_cache_size value. |  |  |
| Table_open_cache_hits |  |  |  | The number of hits for open tables cache lookups. |  |  |
| Table_open_cache_misses |  |  |  | The number of misses for open tables cache lookups. |  |  |
| innodb指标 |  |  |  |  |  |  |
| buffer_pool_data |  |  |  | The total number of bytes in the InnoDB buffer pool containing data. The number includes both dirty and clean pages. |  |  |
| buffer_pool_pages_data |  |  |  | The number of pages in the InnoDB buffer pool containing data. The number includes both dirty and clean pages. |  |  |
| buffer_pool_pages_dirty |  |  |  | The current number of dirty pages in the InnoDB buffer pool. |  |  |
| buffer_pool_pages_flushed |  |  |  | The number of requests to flush pages from the InnoDB buffer pool. |  |  |
| buffer_pool_pages_free |  |  |  | The number of free pages in the InnoDB buffer pool. |  |  |
| buffer_pool_pages_total |  |  |  | The total size of the InnoDB buffer pool, in pages. |  |  |
| buffer_pool_read_ahead |  |  |  | The number of pages read into the InnoDB buffer pool by the read-ahead background thread. |  |  |
| buffer_pool_read_ahead_evicted |  |  |  | The number of pages read into the InnoDB buffer pool by the read-ahead background thread that were subsequently evicted without having been accessed by queries. |  |  |
| buffer_pool_read_ahead_rnd |  |  |  | The number of random read-aheads initiated by InnoDB. This happens when a query scans a large portion of a table but in random order. |  |  |
| buffer_pool_wait_free |  |  |  | When InnoDB needs to read or create a page and no clean pages are available, InnoDB flushes some dirty pages first and waits for that operation to finish. This counter counts instances of these waits. |  |  |
| buffer_pool_write_requests |  |  |  | The number of writes done to the InnoDB buffer pool. |  |  |
| current_transactions |  |  |  | Current InnoDB transactions |  |  |
| data_fsyncs |  |  |  | The number of fsync() operations per second. |  |  |
| data_pending_fsyncs |  |  |  | The current number of pending fsync() operations. |  |  |
| data_pending_reads |  |  |  | The current number of pending reads. |  |  |
| data_pending_writes |  |  |  | The current number of pending writes. |  |  |
| data_read |  |  |  | The amount of data read per second. |  |  |
| data_written |  |  |  | The amount of data written per second. |  |  |
| dblwr_pages_written |  |  |  | The number of pages written per second to the doublewrite buffer. |  |  |
| dblwr_writes |  |  |  | The number of doublewrite operations performed per second. |  |  |
| history_list_length |  |  |  | History list length as shown in the TRANSACTIONS section of the SHOW ENGINE INNODB STATUS output. |  |  |
| log_waits |  |  |  | The number of times that the log buffer was too small and a wait was required for it to be flushed before continuing. |  |  |
| log_write_requests |  |  |  | The number of write requests for the InnoDB redo log. |  |  |
| log_writes |  |  |  | The number of physical writes to the InnoDB redo log file. |  |  |
| hash_index_cells_total |  |  |  | Total number of cells of the adaptive hash index |  |  |
| hash_index_cells_used |  |  |  | Number of used cells of the adaptive hash index |  |  |
| ibuf_free_list |  |  |  | Insert buffer free list, as shown in the INSERT BUFFER AND ADAPTIVE HASH INDEX section of the SHOW ENGINE INNODB STATUS output. |  |  |
| ibuf_segment_size |  |  |  | Insert buffer segment size, as shown in the INSERT BUFFER AND ADAPTIVE HASH INDEX section of the SHOW ENGINE INNODB STATUS output. |  |  |
| ibuf_size |  |  |  | Insert buffer size, as shown in the INSERT BUFFER AND ADAPTIVE HASH INDEX section of the SHOW ENGINE INNODB STATUS output. |  |  |
| pending_aio_log_ios |  |  |  | As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output. |  |  |
| pending_aio_sync_ios |  |  |  | As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output. |  |  |
| pending_ibuf_aio_reads |  |  |  | As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output. |  |  |
| pending_normal_aio_reads |  |  |  | As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output. |  |  |
| pending_normal_aio_writes |  |  |  | As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output. |  |  |
| pending_buffer_pool_flushes |  |  |  | As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output. |  |  |
| pending_log_flushes |  |  |  | As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output. |  |  |
| pending_log_writes |  |  |  | As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output. |  |  |
| checkpoint_age |  |  |  | Checkpoint age as shown in the LOG section of the SHOW ENGINE INNODB STATUS output. |  |  |
| lsn_current |  |  |  | Log sequence number as shown in the LOG section of the SHOW ENGINE INNODB STATUS output. |  |  |
| lsn_flushed |  |  |  | Flushed up to log sequence number as shown in the LOG section of the SHOW ENGINE INNODB STATUS output. |  |  |
| lsn_last_checkpoint |  |  |  | Log sequence number last checkpoint as shown in the LOG section of the SHOW ENGINE INNODB STATUS output. |  |  |
| pending_checkpoint_writes |  |  |  | As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output. |  |  |
| queries_inside |  |  |  | As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output. |  |  |
| queries_queued |  |  |  | As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output. |  |  |
| read_views |  |  |  | As shown in the FILE I/O section of the SHOW ENGINE INNODB STATUS output. |  |  |
| key_buffer_bytes_unflushed |  |  |  | MyISAM key buffer bytes unflushed. |  |  |
| key_buffer_bytes_used |  |  |  | MyISAM key buffer bytes used. |  |  |
| mem_adaptive_hash |  |  |  | As shown in the BUFFER POOL AND MEMORY section of the SHOW ENGINE INNODB STATUS output. |  |  |
| innodb.mem_dictionary |  |  |  | As shown in the BUFFER POOL AND MEMORY section of the SHOW ENGINE INNODB STATUS output. |  |  |
| mem_total |  |  |  | As shown in the BUFFER POOL AND MEMORY section of the SHOW ENGINE INNODB STATUS output. |  |  |
| mem_additional_pool |  |  |  | As shown in the BUFFER POOL AND MEMORY section of the SHOW ENGINE INNODB STATUS output. |  |  |
| mem_file_system |  |  |  | As shown in the BUFFER POOL AND MEMORY section of the SHOW ENGINE INNODB STATUS output. |  |  |
| mem_lock_system |  |  |  | As shown in the BUFFER POOL AND MEMORY section of the SHOW ENGINE INNODB STATUS output. |  |  |
| mem_page_hash |  |  |  | As shown in the BUFFER POOL AND MEMORY section of the SHOW ENGINE INNODB STATUS output. |  |  |
| mem_recovery_system |  |  |  | As shown in the BUFFER POOL AND MEMORY section of the SHOW ENGINE INNODB STATUS output. |  |  |
| os_log_pending_fsyncs |  |  |  | Number of pending InnoDB log fsync (sync-to-disk) requests. |  |  |
| os_log_pending_writes |  |  |  | Number of pending InnoDB log writes. |  |  |
| os_log_written |  |  |  | Number of bytes written to the InnoDB log. |  |  |
| pages_created |  |  |  | Number of InnoDB pages created. |  |  |
| pages_read |  |  |  | Number of InnoDB pages read. |  |  |
| pages_written |  |  |  | Number of InnoDB pages written. |  |  |
| rows_deleted |  |  |  | Number of rows deleted from InnoDB tables. |  |  |
| rows_inserted |  |  |  | Number of rows inserted into InnoDB tables. |  |  |
| rows_read |  |  |  | Number of rows read from InnoDB tables. |  |  |
| rows_updated |  |  |  | Number of rows updated in InnoDB tables. |  |  |
| s_lock_os_waits |  |  |  | As shown in the SEMAPHORES section of the SHOW ENGINE INNODB STATUS output |  |  |
| s_lock_spin_rounds |  |  |  | As shown in the SEMAPHORES section of the SHOW ENGINE INNODB STATUS output. |  |  |
| s_lock_spin_waits |  |  |  | As shown in the SEMAPHORES section of the SHOW ENGINE INNODB STATUS output. |  |  |
| x_lock_os_waits |  |  |  | As shown in the SEMAPHORES section of the SHOW ENGINE INNODB STATUS output. |  |  |
| x_lock_spin_rounds |  |  |  | As shown in the SEMAPHORES section of the SHOW ENGINE INNODB STATUS output. |  |  |
| x_lock_spin_waits |  |  |  | As shown in the SEMAPHORES section of the SHOW ENGINE INNODB STATUS output. |  |  |
| mysql.galera.wsrep_local_recv_queue_avg |  |  |  | Shows the average size of the local received queue since the last status query. |  |  |
| mysql.galera.wsrep_flow_control_paused |  |  |  | Shows the fraction of the time, since FLUSH STATUS was last called, that the node paused due to Flow Control. |  |  |
| mysql.galera.wsrep_flow_control_paused_ns |  |  |  | Shows the pause time due to Flow Control, in nanoseconds. |  |  |
| mysql.galera.wsrep_flow_control_recv |  |  |  | Shows the number of times the galera node has received a pausing Flow Control message from others |  |  |
| mysql.galera.wsrep_flow_control_sent |  |  |  | Shows the number of times the galera node has sent a pausing Flow Control message to others |  |  |
| mysql.galera.wsrep_cert_deps_distance |  |  |  | Shows the average distance between the lowest and highest sequence number, or seqno, values that the node can possibly apply in parallel. |  |  |
| mysql.galera.wsrep_local_send_queue_avg |  |  |  | Show an average for the send queue length since the last FLUSH STATUS query. |  |  |
| mysql.performance.qcache.utilization |  |  |  | Fraction of the query cache memory currently being used. |  |  |
| mysql.performance.digest_95th_percentile.avg_us |  |  |  | Query response time 95th percentile per schema. |  |  |
| mysql.performance.query_run_time.avg |  |  |  | Avg query response time per schema. |  |  |
| disk_use |  |  |  |  |  |  |
| mysql.queries.count |  |  |  | The total count of executed queries per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA. |  |  |
| mysql.queries.errors |  |  |  | The total count of queries run with an error per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA. |  |  |
| Shown as error |  |  |  |  |  |  |
| mysql.queries.time |  |  |  | The total query execution time per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA. |  |  |
| mysql.queries.select_scan |  |  |  | The total count of full table scans on the first table per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA. |  |  |
| mysql.queries.select_full_join |  |  |  | The total count of full table scans on a joined table per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA. |  |  |
| mysql.queries.no_index_used |  |  |  | The total count of queries which do not use an index per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA. |  |  |
| mysql.queries.no_good_index_used |  |  |  | The total count of queries which used a sub-optimal index per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA. |  |  |
| mysql.queries.lock_time |  |  |  | The total time spent waiting on locks per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA. |  |  |
| mysql.queries.rows_affected |  |  |  | The number of rows mutated per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA. |  |  |
| mysql.queries.rows_sent |  |  |  | The number of rows sent per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA. |  |  |
| mysql.queries.rows_examined |  |  |  | The number of rows examined per query per schema. This metric is only available as part of the Deep Database Monitoring ALPHA. |  |  |

db size

| 指标 | 单位 | 类型 | 描述 | 指标源 | 状态 |
| --- | --- | --- | --- | --- | --- | 
| information_schema_size |  | float | db size |  |  |

主从相关



## 数据库监控指标采集
### Metric

**`mysql_dbm_metric`**

- 标签(tag)

| 名称 |描述|
|--|--|
|digest_text| The text of the normalized statement digest.|
|digest| The digest hash value computed from the original normalized statement. |
|query_signature| The MD5 hash value computed from digest_text|
|schema_name|The schema name|
|server| The server address|




- 指标(field)

| 字段名 |类型|说明|
|--|--|--|
|sum_timer_wait|nanosecond | The total query execution time per normalized query and schema.|
|count_star|count | The total count of executed queries per normalized query and schema.|
|sum_errors|count | The total count of queries run with an error per normalized query and schema.|
|sum_lock_time| nanosecond| The total time spent waiting on locks per normalized query and schema.|
|sum_rows_sent|count |The number of rows sent per normalized query and schema. |
|sum_select_scan|count | The total count of full table scans on the first table per normalized query and schema.|
|sum_no_index_used|count | The total count of queries which do not use an index per normalized query and schema.|
|sum_rows_affected|count | The number of rows mutated per normalized query and schema.|
|sum_rows_examined| count| The number of rows examined per normalized query and schema.|
|sum_select_full_join|count |The total count of full table scans on a joined table per normalized query and schema. |
|sum_no_good_index_used|count |The total count of queries which used a sub-optimal index per normalized query and schema. |


**`mysql_dbm_sample`**

- 标签(tag)

|名称|描述|
|--|--|
|current_schema | The name of the current schema.|
|plan_definition | The plan definition of JSON format.|
|plan_signature | The hash value computed from plan definition.|
|query_signature | The hash value computed from digest_text.|
|resource_hash |The hash value computed from sql text.|
|query_truncated |It indicates whether the query is truncated.|
|network_client_ip | The ip address of the client|
|digest | The digest hash value computed from the original normalized statement. |
|digest_text | The text of the normalized statement digest.|
|processlist_db | The name of the database.|
|processlist_user | The user name of the client.|


- 指标(field)

| 字段名 |类型|说明|
|--|--|--|
|timestamp |millisecond| The timestamp when then the event ends.|
|duration |nanosecond|Value in nanoseconds of the event's duration.|
|lock_time_ns | nanosecond|Time in nanoseconds spent waiting for locks. |
|no_good_index_used |int |0 if a good index was found for the statement, 1 if no good index was found.|
|no_index_used | int  | 0 if the statement performed a table scan with an index, 1 if without an index.|
|rows_affected | count| Number of rows the statement affected.|
|rows_examined | count | Number of rows read during the statement's execution.|
|rows_sent | count |Number of rows returned. |
|select_full_join | count | Number of joins performed by the statement which did not use an index.|
|select_full_range_join | count |Number of joins performed by the statement which used a range search of the int first table. |
|select_range | count |Number of joins performed by the statement which used a range of the first table. |
|select_range_check | count |Number of joins without keys performed by the statement that check for key usage after int each row. |
|select_scan | count | Number of joins performed by the statement which used a full scan of the first table.|
|sort_merge_passes | count |Number of merge passes by the sort algorithm performed by the statement. |
|sort_range | count | Number of sorts performed by the statement which used a range.|
|sort_rows | count |Number of rows sorted by the statement. |
|sort_scan | count | Number of sorts performed by the statement which used a full table scan.|
|timer_wait_ns | nanosecond |Value in nanoseconds of the event's duration |