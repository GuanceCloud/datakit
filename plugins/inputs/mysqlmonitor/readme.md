### 简介
mysql指标采集，参考datadog提供的指标，提供默认指标收集和用户自定义查询

### 配置
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

###  收集指标
1. 获取数据结果到resData集中
```
SHOW /*!50002 GLOBAL */ STATUS;
SHOW GLOBAL VARIABLES;
```

2. disable_innodb_metrics and 存储引擎innodb
```
# check innoDB sql
SELECT engine
FROM information_schema.ENGINES
WHERE engine='InnoDB' and support != 'no' and support != 'disabled'
```

extra_innodb_metrics check


```
# 获取innnodb metric
SHOW /*!50000 ENGINE*/ INNODB STATUS

# 算法
for line in innodb_status_text.splitlines():
            line = line.strip()
            row = re.split(" +", line)
            row = [item.strip(',') for item in row]
            row = [item.strip(';') for item in row]
            row = [item.strip('[') for item in row]
            row = [item.strip(']') for item in row]

            if line.startswith('---BUFFER POOL'):
                buffer_id = long(row[2])

            # SEMAPHORES
            if line.find('Mutex spin waits') == 0:
                # Mutex spin waits 79626940, rounds 157459864, OS waits 698719
                # Mutex spin waits 0, rounds 247280272495, OS waits 316513438
                results['Innodb_mutex_spin_waits'] = long(row[3])
                results['Innodb_mutex_spin_rounds'] = long(row[5])
                results['Innodb_mutex_os_waits'] = long(row[8])
            elif line.find('RW-shared spins') == 0 and line.find(';') > 0:
                # RW-shared spins 3859028, OS waits 2100750; RW-excl spins
                # 4641946, OS waits 1530310
                results['Innodb_s_lock_spin_waits'] = long(row[2])
                results['Innodb_x_lock_spin_waits'] = long(row[8])
                results['Innodb_s_lock_os_waits'] = long(row[5])
                results['Innodb_x_lock_os_waits'] = long(row[11])
            elif line.find('RW-shared spins') == 0 and line.find('; RW-excl spins') == -1:
                # Post 5.5.17 SHOW ENGINE INNODB STATUS syntax
                # RW-shared spins 604733, rounds 8107431, OS waits 241268
                results['Innodb_s_lock_spin_waits'] = long(row[2])
                results['Innodb_s_lock_spin_rounds'] = long(row[4])
                results['Innodb_s_lock_os_waits'] = long(row[7])
            elif line.find('RW-excl spins') == 0:
                # Post 5.5.17 SHOW ENGINE INNODB STATUS syntax
                # RW-excl spins 604733, rounds 8107431, OS waits 241268
                results['Innodb_x_lock_spin_waits'] = long(row[2])
                results['Innodb_x_lock_spin_rounds'] = long(row[4])
                results['Innodb_x_lock_os_waits'] = long(row[7])
            elif line.find('seconds the semaphore:') > 0:
                # --Thread 907205 has waited at handler/ha_innodb.cc line 7156 for 1.00 seconds the semaphore:
                results['Innodb_semaphore_waits'] += 1
                results['Innodb_semaphore_wait_time'] += long(float(row[9])) * 1000

            # TRANSACTIONS
            elif line.find('Trx id counter') == 0:
                # The beginning of the TRANSACTIONS section: start counting
                # transactions
                # Trx id counter 0 1170664159
                # Trx id counter 861B144C
                txn_seen = True
            elif line.find('History list length') == 0:
                # History list length 132
                results['Innodb_history_list_length'] = long(row[3])
            elif txn_seen and line.find('---TRANSACTION') == 0:
                # ---TRANSACTION 0, not started, process no 13510, OS thread id 1170446656
                results['Innodb_current_transactions'] += 1
                if line.find('ACTIVE') > 0:
                    results['Innodb_active_transactions'] += 1
            elif line.find('read views open inside InnoDB') > 0:
                # 1 read views open inside InnoDB
                results['Innodb_read_views'] = long(row[0])
            elif line.find('mysql tables in use') == 0:
                # mysql tables in use 2, locked 2
                results['Innodb_tables_in_use'] += long(row[4])
                results['Innodb_locked_tables'] += long(row[6])
            elif txn_seen and line.find('lock struct(s)') > 0:
                # 23 lock struct(s), heap size 3024, undo log entries 27
                # LOCK WAIT 12 lock struct(s), heap size 3024, undo log entries 5
                # LOCK WAIT 2 lock struct(s), heap size 368
                if line.find('LOCK WAIT') == 0:
                    results['Innodb_lock_structs'] += long(row[2])
                    results['Innodb_locked_transactions'] += 1
                elif line.find('ROLLING BACK') == 0:
                    # ROLLING BACK 127539 lock struct(s), heap size 15201832,
                    # 4411492 row lock(s), undo log entries 1042488
                    results['Innodb_lock_structs'] += long(row[2])
                else:
                    results['Innodb_lock_structs'] += long(row[0])

            # FILE I/O
            elif line.find(' OS file reads, ') > 0:
                # 8782182 OS file reads, 15635445 OS file writes, 947800 OS
                # fsyncs
                results['Innodb_os_file_reads'] = long(row[0])
                results['Innodb_os_file_writes'] = long(row[4])
                results['Innodb_os_file_fsyncs'] = long(row[8])
            elif line.find('Pending normal aio reads:') == 0:
                try:
                    if len(row) == 8:
                        # (len(row) == 8)  Pending normal aio reads: 0, aio writes: 0,
                        results['Innodb_pending_normal_aio_reads'] = long(row[4])
                        results['Innodb_pending_normal_aio_writes'] = long(row[7])
                    elif len(row) == 14:
                        # (len(row) == 14) Pending normal aio reads: 0 [0, 0] , aio writes: 0 [0, 0] ,
                        results['Innodb_pending_normal_aio_reads'] = long(row[4])
                        results['Innodb_pending_normal_aio_writes'] = long(row[10])
                    elif len(row) == 16:
                        # (len(row) == 16) Pending normal aio reads: [0, 0, 0, 0] , aio writes: [0, 0, 0, 0] ,
                        if _are_values_numeric(row[4:8]) and _are_values_numeric(row[11:15]):
                            results['Innodb_pending_normal_aio_reads'] = (
                                long(row[4]) + long(row[5]) + long(row[6]) + long(row[7])
                            )
                            results['Innodb_pending_normal_aio_writes'] = (
                                long(row[11]) + long(row[12]) + long(row[13]) + long(row[14])
                            )

                        # (len(row) == 16) Pending normal aio reads: 0 [0, 0, 0, 0] , aio writes: 0 [0, 0] ,
                        elif _are_values_numeric(row[4:9]) and _are_values_numeric(row[12:15]):
                            results['Innodb_pending_normal_aio_reads'] = long(row[4])
                            results['Innodb_pending_normal_aio_writes'] = long(row[12])
                        else:
                            self.log.warning("Can't parse result line %s", line)
                    elif len(row) == 18:
                        # (len(row) == 18) Pending normal aio reads: 0 [0, 0, 0, 0] , aio writes: 0 [0, 0, 0, 0] ,
                        results['Innodb_pending_normal_aio_reads'] = long(row[4])
                        results['Innodb_pending_normal_aio_writes'] = long(row[12])
                    elif len(row) == 22:
                        # (len(row) == 22)
                        # Pending normal aio reads: 0 [0, 0, 0, 0, 0, 0, 0, 0] , aio writes: 0 [0, 0, 0, 0] ,
                        results['Innodb_pending_normal_aio_reads'] = long(row[4])
                        results['Innodb_pending_normal_aio_writes'] = long(row[16])
                except ValueError as e:
                    self.log.warning("Can't parse result line %s: %s", line, e)
            elif line.find('ibuf aio reads') == 0:
                #  ibuf aio reads: 0, log i/o's: 0, sync i/o's: 0
                #  or ibuf aio reads:, log i/o's:, sync i/o's:
                if len(row) == 10:
                    results['Innodb_pending_ibuf_aio_reads'] = long(row[3])
                    results['Innodb_pending_aio_log_ios'] = long(row[6])
                    results['Innodb_pending_aio_sync_ios'] = long(row[9])
                elif len(row) == 7:
                    results['Innodb_pending_ibuf_aio_reads'] = 0
                    results['Innodb_pending_aio_log_ios'] = 0
                    results['Innodb_pending_aio_sync_ios'] = 0
            elif line.find('Pending flushes (fsync)') == 0:
                # Pending flushes (fsync) log: 0; buffer pool: 0
                results['Innodb_pending_log_flushes'] = long(row[4])
                results['Innodb_pending_buffer_pool_flushes'] = long(row[7])

            # INSERT BUFFER AND ADAPTIVE HASH INDEX
            elif line.find('Ibuf for space 0: size ') == 0:
                # Older InnoDB code seemed to be ready for an ibuf per tablespace.  It
                # had two lines in the output.  Newer has just one line, see below.
                # Ibuf for space 0: size 1, free list len 887, seg size 889, is not empty
                # Ibuf for space 0: size 1, free list len 887, seg size 889,
                results['Innodb_ibuf_size'] = long(row[5])
                results['Innodb_ibuf_free_list'] = long(row[9])
                results['Innodb_ibuf_segment_size'] = long(row[12])
            elif line.find('Ibuf: size ') == 0:
                # Ibuf: size 1, free list len 4634, seg size 4636,
                results['Innodb_ibuf_size'] = long(row[2])
                results['Innodb_ibuf_free_list'] = long(row[6])
                results['Innodb_ibuf_segment_size'] = long(row[9])

                if line.find('merges') > -1:
                    results['Innodb_ibuf_merges'] = long(row[10])
            elif line.find(', delete mark ') > 0 and prev_line.find('merged operations:') == 0:
                # Output of show engine innodb status has changed in 5.5
                # merged operations:
                # insert 593983, delete mark 387006, delete 73092
                results['Innodb_ibuf_merged_inserts'] = long(row[1])
                results['Innodb_ibuf_merged_delete_marks'] = long(row[4])
                results['Innodb_ibuf_merged_deletes'] = long(row[6])
                results['Innodb_ibuf_merged'] = (
                    results['Innodb_ibuf_merged_inserts']
                    + results['Innodb_ibuf_merged_delete_marks']
                    + results['Innodb_ibuf_merged_deletes']
                )
            elif line.find(' merged recs, ') > 0:
                # 19817685 inserts, 19817684 merged recs, 3552620 merges
                results['Innodb_ibuf_merged_inserts'] = long(row[0])
                results['Innodb_ibuf_merged'] = long(row[2])
                results['Innodb_ibuf_merges'] = long(row[5])
            elif line.find('Hash table size ') == 0:
                # In some versions of InnoDB, the used cells is omitted.
                # Hash table size 4425293, used cells 4229064, ....
                # Hash table size 57374437, node heap has 72964 buffer(s) <--
                # no used cells
                results['Innodb_hash_index_cells_total'] = long(row[3])
                results['Innodb_hash_index_cells_used'] = long(row[6]) if line.find('used cells') > 0 else 0

            # LOG
            elif line.find(" log i/o's done, ") > 0:
                # 3430041 log i/o's done, 17.44 log i/o's/second
                # 520835887 log i/o's done, 17.28 log i/o's/second, 518724686
                # syncs, 2980893 checkpoints
                results['Innodb_log_writes'] = long(row[0])
            elif line.find(" pending log writes, ") > 0:
                # 0 pending log writes, 0 pending chkp writes
                results['Innodb_pending_log_writes'] = long(row[0])
                results['Innodb_pending_checkpoint_writes'] = long(row[4])
            elif line.find("Log sequence number") == 0:
                # This number is NOT printed in hex in InnoDB plugin.
                # Log sequence number 272588624
                results['Innodb_lsn_current'] = long(row[3])
            elif line.find("Log flushed up to") == 0:
                # This number is NOT printed in hex in InnoDB plugin.
                # Log flushed up to   272588624
                results['Innodb_lsn_flushed'] = long(row[4])
            elif line.find("Last checkpoint at") == 0:
                # Last checkpoint at  272588624
                results['Innodb_lsn_last_checkpoint'] = long(row[3])

            # BUFFER POOL AND MEMORY
            elif line.find("Total memory allocated") == 0 and line.find("in additional pool allocated") > 0:
                # Total memory allocated 29642194944; in additional pool allocated 0
                # Total memory allocated by read views 96
                results['Innodb_mem_total'] = long(row[3])
                results['Innodb_mem_additional_pool'] = long(row[8])
            elif line.find('Adaptive hash index ') == 0:
                #   Adaptive hash index 1538240664     (186998824 + 1351241840)
                results['Innodb_mem_adaptive_hash'] = long(row[3])
            elif line.find('Page hash           ') == 0:
                #   Page hash           11688584
                results['Innodb_mem_page_hash'] = long(row[2])
            elif line.find('Dictionary cache    ') == 0:
                #   Dictionary cache    145525560      (140250984 + 5274576)
                results['Innodb_mem_dictionary'] = long(row[2])
            elif line.find('File system         ') == 0:
                #   File system         313848         (82672 + 231176)
                results['Innodb_mem_file_system'] = long(row[2])
            elif line.find('Lock system         ') == 0:
                #   Lock system         29232616       (29219368 + 13248)
                results['Innodb_mem_lock_system'] = long(row[2])
            elif line.find('Recovery system     ') == 0:
                #   Recovery system     0      (0 + 0)
                results['Innodb_mem_recovery_system'] = long(row[2])
            elif line.find('Threads             ') == 0:
                #   Threads             409336         (406936 + 2400)
                results['Innodb_mem_thread_hash'] = long(row[1])
            elif line.find("Buffer pool size ") == 0:
                # The " " after size is necessary to avoid matching the wrong line:
                # Buffer pool size        1769471
                # Buffer pool size, bytes 28991012864
                if buffer_id == -1:
                    results['Innodb_buffer_pool_pages_total'] = long(row[3])
            elif line.find("Free buffers") == 0:
                # Free buffers            0
                if buffer_id == -1:
                    results['Innodb_buffer_pool_pages_free'] = long(row[2])
            elif line.find("Database pages") == 0:
                # Database pages          1696503
                if buffer_id == -1:
                    results['Innodb_buffer_pool_pages_data'] = long(row[2])

            elif line.find("Modified db pages") == 0:
                # Modified db pages       160602
                if buffer_id == -1:
                    results['Innodb_buffer_pool_pages_dirty'] = long(row[3])
            elif line.find("Pages read ahead") == 0:
                # Must do this BEFORE the next test, otherwise it'll get fooled by this
                # line from the new plugin:
                # Pages read ahead 0.00/s, evicted without access 0.06/s
                pass
            elif line.find("Pages read") == 0:
                # Pages read 15240822, created 1770238, written 21705836
                if buffer_id == -1:
                    results['Innodb_pages_read'] = long(row[2])
                    results['Innodb_pages_created'] = long(row[4])
                    results['Innodb_pages_written'] = long(row[6])

            # ROW OPERATIONS
            elif line.find('Number of rows inserted') == 0:
                # Number of rows inserted 50678311, updated 66425915, deleted
                # 20605903, read 454561562
                results['Innodb_rows_inserted'] = long(row[4])
                results['Innodb_rows_updated'] = long(row[6])
                results['Innodb_rows_deleted'] = long(row[8])
                results['Innodb_rows_read'] = long(row[10])
            elif line.find(" queries inside InnoDB, ") > 0:
                # 0 queries inside InnoDB, 0 queries in queue
                results['Innodb_queries_inside'] = long(row[0])
                results['Innodb_queries_queued'] = long(row[4])

            prev_line = line

        # We need to calculate this metric separately
        try:
            results['Innodb_checkpoint_age'] = results['Innodb_lsn_current'] - results['Innodb_lsn_last_checkpoint']
        except KeyError as e:
            self.log.error("Not all InnoDB LSN metrics available, unable to compute: %s", e)

        # Finally we change back the metrics values to string to make the values
        # consistent with how they are reported by SHOW GLOBAL STATUS
        for metric, value in list(iteritems(results)):
            results[metric] = str(value)
```

3. binlog统计
- check binlog是否开启
```
SHOW BINARY LOGS;

# 算法
for value in itervalues(master_logs):
    binary_log_space += value
```

4. cache utilization
```
# Compute key cache utilization metric
key_blocks_unused = collect_scalar('Key_blocks_unused', results)
key_cache_block_size = collect_scalar('key_cache_block_size', results)
key_buffer_size = collect_scalar('key_buffer_size', results)
results['Key_buffer_size'] = key_buffer_size

try:
    # can be null if the unit is missing in the user config (4 instead of 4G for eg.)
    if key_buffer_size != 0:
        key_cache_utilization = 1 - ((key_blocks_unused * key_cache_block_size) / key_buffer_size)
        results['Key_cache_utilization'] = key_cache_utilization

    results['Key_buffer_bytes_used'] = collect_scalar('Key_blocks_used', results) * key_cache_block_size
    results['Key_buffer_bytes_unflushed'] = (
        collect_scalar('Key_blocks_not_flushed', results) * key_cache_block_size
    )
```

5. extra_status_metrics
```
if is_affirmative(self._config.options.get('extra_status_metrics', False)):
    self.log.debug("Collecting Extra Status Metrics")
    metrics.update(OPTIONAL_STATUS_VARS)

    if self.version.version_compatible((5, 6, 6)):
        metrics.update(OPTIONAL_STATUS_VARS_5_6_6)
```

6. galera_cluster
```
if is_affirmative(self._config.options.get('galera_cluster', False)):
    # already in result-set after 'SHOW STATUS' just add vars to collect
    self.log.debug("Collecting Galera Metrics.")
    metrics.update(GALERA_VARS)
```

7. performance_schema
```
 performance_schema_enabled = self._get_variable_enabled(results, 'performance_schema')
above_560 = self.version.version_compatible((5, 6, 0))
if (
    is_affirmative(self._config.options.get('extra_performance_metrics', False))
    and above_560
    and performance_schema_enabled
):
    # report avg query response time per schema to Datadog
    results['perf_digest_95th_percentile_avg_us'] = self._get_query_exec_time_95th_us(db)
    results['query_run_time_avg'] = self._query_exec_time_per_schema(db)
    metrics.update(PERFORMANCE_VARS)
```

8. schema_size_metrics
```
if is_affirmative(self._config.options.get('schema_size_metrics', False)):
    # report avg query response time per schema to Datadog
    results['information_schema_size'] = self._query_size_per_schema(db)
    metrics.update(SCHEMA_VARS)
```

9. replication
```
if is_affirmative(self._config.options.get('replication', False)):
    replication_metrics = self._collect_replication_metrics(db, results, above_560)
    metrics.update(replication_metrics)
    self._check_replication_status(results)
```