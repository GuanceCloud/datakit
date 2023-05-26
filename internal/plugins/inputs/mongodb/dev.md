# summary

`mongodb` 运行状态采集共五个指标集

- `mongodb` mongodb 基础运行状态
- `mongodb_db_stats` mongodb 业务数据库运行状态
- `mongodb_col_stats` mongodb collection 运行状态
- `mongodb_shard_stats` mongodb sharding 运行状态
- `mongodb_top_stats` mongodb top 命令
- `mongod_log` mongod log

# config sample

```
[[inputs.mongodb]]
  ## Gathering interval
  # interval = "` + defInterval.UnitString(time.Second) + `"

  ## An array of URLs of the form:
  ##   "mongodb://" [user ":" pass "@"] host [ ":" port]
  ## For example:
  ##   mongodb://user:auth_key@10.10.3.30:27017,
  ##   mongodb://10.10.3.33:18832,
  # servers = ["` + defMongoUrl + `"]

  ## When true, collect replica set stats
  # gather_replica_set_stats = false

  ## When true, collect cluster stats
  ## Note that the query that counts jumbo chunks triggers a COLLSCAN, which may have an impact on performance.
  # gather_cluster_stats = false

  ## When true, collect per database stats
  # gather_per_db_stats = true

  ## When true, collect per collection stats
  # gather_per_col_stats = true

  ## List of db where collections stats are collected, If empty, all dbs are concerned.
  # col_stats_dbs = []

  ## When true, collect top command stats.
  # gather_top_stat = true

  ## Optional TLS Config, enabled if true.
  # enable_tls = false

	## Optional local Mongod log input config, enabled if true.
	# enable_mongod_log = false

  ## TLS connection config
  [inputs.mongodb.tlsconf]
    # ca_certs = ["` + defTlsCaCert + `"]
    # cert = "` + defTlsCert + `"
    # cert_key = "` + defTlsCertKey + `"
    ## Use TLS but skip chain & host verification
    # insecure_skip_verify = false
    # server_name = ""

	## MongoD log
	[inputs.mongodb.log]
		## Log file path check your mongodb config path usually under /var/log/mongodb/mongod.log
		# files = ["` + defMongodLogPath + `"]
		## Grok pipeline script path
		# pipeline = "` + defPipeline + `"

  ## Customer tags, if set will be seen with every metric.
  [inputs.mongodb.tags]
    # "key1" = "value1"
    # "key2" = "value2"
```

# pipeline config

```
json_all()
rename("command", c)
rename("severity", s)
rename("context", ctx)
drop_key(id)
```

# tls config

config mongodb to enable connect to mongod instance using encryption and client side identification.

## prerequisite

run `openssl version` if nothing output run `sudo apt install openssl -y` to install openssl lib.

## encryption with tls option

use openssl to generate .pem file for mongod server tls encryption, run command

```
sudo openssl req -x509 -newkey rsa:2048 -days 365 -keyout mongod.key.pem -out mongod.cert.pem -nodes -subj '/CN=\<mongod server you want to connect to\>'
```

after run this command above you will find mongod certificate file `mongod.cert.pem` and mongod certificate key file `mongod.key.pem` under path, we need to merge these two files, run command

```
sudo bash -c "cat mongod.cert.pem mongod.key.pem >>mongod.pem"
```

after merge add this config into your mongod config file `/etc/mongod.conf`

```
# TLS config
net:
  tls:
    mode: requireTLS
    certificateKeyFile: /etc/ssl/mongod.pem
```

start mongod

```
sudo mongod --config /etc/mongod.conf
```

copy `mongod.cert.pem` file into your client side and test connect with tls, run command

```
mongo --tls --host <your mongod url> --tlsCAFile /etc/ssl/certs/mongod.cert.pem
```

## client identification with tls option

use openssl to generate .pem file for mongo client side identification using tls, run command

```
sudo openssl req -x509 -newkey rsa:2048 -days 365 -keyout mongo.key.pem -out mongo.cert.pem -nodes
```

after run this command above you will find mongo certificate file `mongo.cert.pem` and mongo certificate key file `mongo.key.pem` under path, we need to copy `mongo.cert.pem` file into your mongodb server, and config `/etc/mongod.config` file

```
# Tls config
net:
  tls:
    mode: requireTLS
    certificateKeyFile: /etc/ssl/mongod.pem
    CAFile: /etc/ssl/mongo.cert.pem
```

start mongod server, run command

```
sudo mongod --config /etc/mongod.conf
```

on your client side, merge `mongo.cert.pem` and `mongo.key.pem` into `mongo.pem`, run command

```
sudo bash -c "cat mongo.cert.pem mongo.key.pem >>mongo.pem"
```

connect to mongod server with tls, run command

```
mongo --tls --host <your mongod url> --tlsCAFile /etc/ssl/certs/mongod.cert.pem --tlsCertificateKeyFile /etc/ssl/certs/mongo.pem
```

# metrics

## mongodb

| 标签名    | 描述                     |
| --------- | ------------------------ |
| hostname  | mongodb host             |
| node_type | node type in replica set |
| rs_name   | replica set name         |

| 指标                                      | 类型          | 指标源       | 单位          | 描述                                                                                                                                                                                                                                                                                                                        |
| ----------------------------------------- | ------------- | ------------ | ------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| active_reads                              | inputs.Int    | inputs.Gauge | inputs.NCount | The number of the active client connections performing read operations.                                                                                                                                                                                                                                                     |
| active_writes                             | inputs.Int    | inputs.Gauge | inputs.NCount | The number of active client connections performing write operations.                                                                                                                                                                                                                                                        |
| aggregate_command_failed                  | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times that 'aggregate' command failed on this mongod                                                                                                                                                                                                                                                          |
| aggregate_command_total                   | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times that 'aggregate' command executed on this mongod.                                                                                                                                                                                                                                                       |
| assert_msg                                | inputs.Int    | inputs.Gauge | inputs.NCount | The number of message assertions raised since the MongoDB process started. Check the log file for more information about these messages.                                                                                                                                                                                    |
| assert_regular                            | inputs.Int    | inputs.Gauge | inputs.NCount | The number of regular assertions raised since the MongoDB process started. Check the log file for more information about these messages.                                                                                                                                                                                    |
| assert_rollovers                          | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times that the rollover counters have rolled over since the last time the MongoDB process started. The counters will rollover to zero after 2 30 assertions. Use this value to provide context to the other values in the asserts data structure.                                                             |
| assert_user                               | inputs.Int    | inputs.Gauge | inputs.NCount | The number of "user asserts" that have occurred since the last time the MongoDB process started. These are errors that user may generate, such as out of disk space or duplicate key. You can prevent these assertions by fixing a problem with your application or deployment. Check the MongoDB log for more information. |
| assert_warning                            | inputs.Int    | inputs.Gauge | inputs.NCount | Changed in version 4.0. Starting in MongoDB 4.0, the field returns zero 0. In earlier versions, the field returns the number of warnings raised since the MongoDB process started.                                                                                                                                          |
| available_reads                           | inputs.Int    | inputs.Gauge | inputs.NCount | The number of concurrent of read transactions allowed into the WiredTiger storage engine                                                                                                                                                                                                                                    |
| available_writes                          | inputs.Int    | inputs.Gauge | inputs.NCount | The number of concurrent of write transactions allowed into the WiredTiger storage engine                                                                                                                                                                                                                                   |
| commands                                  | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of commands issued to the database since the mongod instance last started. opcounters.command counts all commands except the write commands: insert, update, and delete.                                                                                                                                   |
| connections_available                     | inputs.Int    | inputs.Gauge | inputs.NCount | The number of unused incoming connections available.                                                                                                                                                                                                                                                                        |
| connections_current                       | inputs.Int    | inputs.Gauge | inputs.NCount | The number of incoming connections from clients to the database server.                                                                                                                                                                                                                                                     |
| connections_total_created                 | inputs.Int    | inputs.Gauge | inputs.NCount | Count of all incoming connections created to the server. This number includes connections that have since closed.                                                                                                                                                                                                           |
| count_command_failed                      | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times that 'count' command failed on this mongod                                                                                                                                                                                                                                                              |
| count_command_total                       | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times that 'count' command executed on this mongod                                                                                                                                                                                                                                                            |
| cursor_no_timeout_count                   | inputs.Int    | inputs.Gauge | inputs.NCount | The number of open cursors with the option DBQuery.Option.noTimeout set to prevent timeout after a period of inactivity                                                                                                                                                                                                     |
| cursor_pinned_count                       | inputs.Int    | inputs.Gauge | inputs.NCount | The number of "pinned" open cursors.                                                                                                                                                                                                                                                                                        |
| cursor_timed_out_count                    | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of cursors that have timed out since the server process started. If this number is large or growing at a regular rate, this may indicate an application error.                                                                                                                                             |
| cursor_total_count                        | inputs.Int    | inputs.Gauge | inputs.NCount | The number of cursors that MongoDB is maintaining for clients. Because MongoDB exhausts unused cursors, typically this value small or zero. However, if there is a queue, stale tailable cursors, or a large number of operations this value may rise.                                                                      |
| delete_command_failed                     | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times that 'delete' command failed on this mongod                                                                                                                                                                                                                                                             |
| delete_command_total                      | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times that 'delete' command executed on this mongod                                                                                                                                                                                                                                                           |
| deletes                                   | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of delete operations since the mongod instance last started.                                                                                                                                                                                                                                               |
| distinct_command_failed                   | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times that 'distinct' command failed on this mongod                                                                                                                                                                                                                                                           |
| distinct_command_total                    | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times that 'distinct' command executed on this mongod                                                                                                                                                                                                                                                         |
| document_deleted                          | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of documents deleted.                                                                                                                                                                                                                                                                                      |
| document_inserted                         | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of documents inserted.                                                                                                                                                                                                                                                                                     |
| document_returned                         | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of documents returned by queries.                                                                                                                                                                                                                                                                          |
| document_updated                          | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of documents updated.                                                                                                                                                                                                                                                                                      |
| find_and_modify_command_failed            | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times that 'find' and 'modify' commands failed on this mongod                                                                                                                                                                                                                                                 |
| find_and_modify_command_total             | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times that 'find' and 'modify' commands executed on this mongod                                                                                                                                                                                                                                               |
| find_command_failed                       | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times that 'find' command failed on this mongod                                                                                                                                                                                                                                                               |
| find_command_total                        | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times that 'find' command executed on this mongod                                                                                                                                                                                                                                                             |
| flushes                                   | inputs.Int    | inputs.Gauge | inputs.NCount | The number of transaction checkpoints                                                                                                                                                                                                                                                                                       |
| flushes_total_time_ns                     | inputs.Int    | inputs.Gauge | inputs.NCount | The transaction checkpoint total time (msecs)                                                                                                                                                                                                                                                                               |
| get_more_command_failed                   | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times that 'get more' command failed on this mongod                                                                                                                                                                                                                                                           |
| get_more_command_total                    | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times that 'get more' command executed on this mongod                                                                                                                                                                                                                                                         |
| getmores                                  | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of getMore operations since the mongod instance last started. This counter can be high even if the query count is low. Secondary nodes send getMore operations as part of the replication process.                                                                                                         |
| insert_command_failed                     | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times that 'insert' command failed on this mongod                                                                                                                                                                                                                                                             |
| insert_command_total                      | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times that 'insert' command executed on this mongod                                                                                                                                                                                                                                                           |
| inserts                                   | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of insert operations received since the mongod instance last started.                                                                                                                                                                                                                                      |
| jumbo_chunks                              | inputs.Int    | inputs.Gauge | inputs.NCount | Count jumbo flags in cluster chunk.                                                                                                                                                                                                                                                                                         |
| latency_commands                          | inputs.Int    | inputs.Gauge | inputs.NCount | The total combined latency in microseconds of latency statistics for database command.                                                                                                                                                                                                                                      |
| latency_commands_count                    | inputs.Int    | inputs.Gauge | inputs.NCount | The total combined latency of operations performed on the collection for database command.                                                                                                                                                                                                                                  |
| latency_reads                             | inputs.Int    | inputs.Gauge | inputs.NCount | The total combined latency in microseconds of latency statistics for read request.                                                                                                                                                                                                                                          |
| latency_reads_count                       | inputs.Int    | inputs.Gauge | inputs.NCount | The total combined latency of operations performed on the collection for read request.                                                                                                                                                                                                                                      |
| latency_writes                            | inputs.Int    | inputs.Gauge | inputs.NCount | The total combined latency in microseconds of latency statistics for write request.                                                                                                                                                                                                                                         |
| latency_writes_count                      | inputs.Int    | inputs.Gauge | inputs.NCount | The total combined latency of operations performed on the collection for write request.                                                                                                                                                                                                                                     |
| member_status                             | inputs.String | inputs.Gauge | inputs.NCount | The state of ndoe in replica members.                                                                                                                                                                                                                                                                                       |
| net_in_bytes_count                        | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of bytes that the server has received over network connections initiated by clients or other mongod instances.                                                                                                                                                                                             |
| net_out_bytes_count                       | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of bytes that the server has sent over network connections initiated by clients or other mongod instances.                                                                                                                                                                                                 |
| open_connections                          | inputs.Int    | inputs.Gauge | inputs.NCount | The number of incoming connections from clients to the database server.                                                                                                                                                                                                                                                     |
| operation_scan_and_order                  | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of queries that return sorted numbers that cannot perform the sort operation using an index.                                                                                                                                                                                                               |
| operation_write_conflicts                 | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of queries that encountered write conflicts.                                                                                                                                                                                                                                                               |
| page_faults                               | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of page faults.                                                                                                                                                                                                                                                                                            |
| percent_cache_dirty                       | inputs.Float  | inputs.Gauge | inputs.NCount | Size in bytes of the dirty data in the cache. This value should be less than the bytes currently in the cache value.                                                                                                                                                                                                        |
| percent_cache_used                        | inputs.Float  | inputs.Gauge | inputs.NCount | Size in byte of the data currently in cache. This value should not be greater than the maximum bytes configured value.                                                                                                                                                                                                      |
| queries                                   | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of queries received since the mongod instance last started.                                                                                                                                                                                                                                                |
| queued_reads                              | inputs.Int    | inputs.Gauge | inputs.NCount | The number of operations that are currently queued and waiting for the read lock. A consistently small read-queue, particularly of shorter operations, should cause no concern.                                                                                                                                             |
| queued_writes                             | inputs.Int    | inputs.Gauge | inputs.NCount | The number of operations that are currently queued and waiting for the write lock. A consistently small write-queue, particularly of shorter operations, is no cause for concern.                                                                                                                                           |
| repl_apply_batches_num                    | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of batches applied across all databases.                                                                                                                                                                                                                                                                   |
| repl_apply_batches_total_millis           | inputs.Int    | inputs.Gauge | inputs.NCount | The total amount of time in milliseconds the mongod has spent applying operations from the oplog.                                                                                                                                                                                                                           |
| repl_apply_ops                            | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of oplog operations applied. metrics.repl.apply.ops is incremented after each operation.                                                                                                                                                                                                                   |
| repl_buffer_count                         | inputs.Int    | inputs.Gauge | inputs.NCount | The current number of operations in the oplog buffer.                                                                                                                                                                                                                                                                       |
| repl_buffer_size_bytes                    | inputs.Int    | inputs.Gauge | inputs.NCount | The current size of the contents of the oplog buffer.                                                                                                                                                                                                                                                                       |
| repl_commands                             | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of replicated commands issued to the database since the mongod instance last started.                                                                                                                                                                                                                      |
| repl_deletes                              | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of replicated delete operations since the mongod instance last started.                                                                                                                                                                                                                                    |
| repl_executor_pool_in_progress_count      | inputs.Int    | inputs.Gauge | inputs.NCount | inputs.TODO                                                                                                                                                                                                                                                                                                                 |
| repl_executor_queues_network_in_progress  | inputs.Int    | inputs.Gauge | inputs.NCount | inputs.TODO                                                                                                                                                                                                                                                                                                                 |
| repl_executor_queues_sleepers             | inputs.Int    | inputs.Gauge | inputs.NCount | inputs.TODO                                                                                                                                                                                                                                                                                                                 |
| repl_executor_unsignaled_events           | inputs.Int    | inputs.Gauge | inputs.NCount | inputs.TODO                                                                                                                                                                                                                                                                                                                 |
| repl_getmores                             | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of getMore operations since the mongod instance last started.                                                                                                                                                                                                                                              |
| repl_inserts                              | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of replicated insert operations since the mongod instance last started.                                                                                                                                                                                                                                    |
| repl_lag                                  | inputs.Int    | inputs.Gauge | inputs.NCount | inputs.TODO                                                                                                                                                                                                                                                                                                                 |
| repl_network_bytes                        | inputs.Int    | inputs.Gauge | inputs.NCount | The total amount of data read from the replication sync source.                                                                                                                                                                                                                                                             |
| repl_network_getmores_num                 | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of getmore operations, which are operations that request an additional set of operations from the replication sync source.                                                                                                                                                                                 |
| repl_network_getmores_total_millis        | inputs.Int    | inputs.Gauge | inputs.NCount | The total amount of time required to collect data from getmore operations.                                                                                                                                                                                                                                                  |
| repl_network_ops                          | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of operations read from the replication source.                                                                                                                                                                                                                                                            |
| repl_oplog_window_sec                     | inputs.Int    | inputs.Gauge | inputs.NCount | The second window of replication oplog.                                                                                                                                                                                                                                                                                     |
| repl_queries                              | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of replicated queries since the mongod instance last started.                                                                                                                                                                                                                                              |
| repl_state                                | inputs.Int    | inputs.Gauge | inputs.NCount | The node state of replication member.                                                                                                                                                                                                                                                                                       |
| repl_updates                              | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of replicated update operations since the mongod instance last started.                                                                                                                                                                                                                                    |
| resident_megabytes                        | inputs.Int    | inputs.Gauge | inputs.NCount | The value of mem.resident is roughly equivalent to the amount of RAM, in mebibyte (MiB), currently used by the database process.                                                                                                                                                                                            |
| state                                     | inputs.String | inputs.Gauge | inputs.NCount | The replication state.                                                                                                                                                                                                                                                                                                      |
| storage_freelist_search_bucket_exhausted  | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times that mongod has checked the free list without finding a suitably large record allocation.                                                                                                                                                                                                               |
| storage_freelist_search_requests          | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times mongod has searched for available record allocations.                                                                                                                                                                                                                                                   |
| storage_freelist_search_scanned           | inputs.Int    | inputs.Gauge | inputs.NCount | The number of available record allocations mongod has searched.                                                                                                                                                                                                                                                             |
| tcmalloc_central_cache_free_bytes         | inputs.Int    | inputs.Gauge | inputs.NCount | Number of free bytes in the central cache that have been assigned to size classes. They always count towards virtual memory usage, and unless the underlying memory is swapped.                                                                                                                                             |
| tcmalloc_current_allocated_bytes          | inputs.Int    | inputs.Gauge | inputs.NCount | Number of bytes currently allocated by application.                                                                                                                                                                                                                                                                         |
| tcmalloc_current_total_thread_cache_bytes | inputs.Int    | inputs.Gauge | inputs.NCount | Number of bytes used across all thread caches.                                                                                                                                                                                                                                                                              |
| tcmalloc_heap_size                        | inputs.Int    | inputs.Gauge | inputs.NCount | Number of bytes in the heap.                                                                                                                                                                                                                                                                                                |
| tcmalloc_max_total_thread_cache_bytes     | inputs.Int    | inputs.Gauge | inputs.NCount | Upper limit on total number of bytes stored across all per-thread caches. Default: 16MB.                                                                                                                                                                                                                                    |
| tcmalloc_pageheap_commit_count            | inputs.Int    | inputs.Gauge | inputs.NCount | Number of virtual memory commits.                                                                                                                                                                                                                                                                                           |
| tcmalloc_pageheap_committed_bytes         | inputs.Int    | inputs.Gauge | inputs.NCount | Bytes committed, always <= system*bytes*.                                                                                                                                                                                                                                                                                   |
| tcmalloc_pageheap_decommit_count          | inputs.Int    | inputs.Gauge | inputs.NCount | Number of virtual memory decommits.                                                                                                                                                                                                                                                                                         |
| tcmalloc_pageheap_free_bytes              | inputs.Int    | inputs.Gauge | inputs.NCount | Number of bytes in free, mapped pages in page heap.                                                                                                                                                                                                                                                                         |
| tcmalloc_pageheap_reserve_count           | inputs.Int    | inputs.Gauge | inputs.NCount | Number of virtual memory reserves.                                                                                                                                                                                                                                                                                          |
| tcmalloc_pageheap_scavenge_count          | inputs.Int    | inputs.Gauge | inputs.NCount | Number of times scavagened flush pages.                                                                                                                                                                                                                                                                                     |
| tcmalloc_pageheap_total_commit_bytes      | inputs.Int    | inputs.Gauge | inputs.NCount | Bytes committed in lifetime of process.                                                                                                                                                                                                                                                                                     |
| tcmalloc_pageheap_total_decommit_bytes    | inputs.Int    | inputs.Gauge | inputs.NCount | Bytes decommitted in lifetime of process.                                                                                                                                                                                                                                                                                   |
| tcmalloc_pageheap_total_reserve_bytes     | inputs.Int    | inputs.Gauge | inputs.NCount | Number of virtual memory reserves.                                                                                                                                                                                                                                                                                          |
| tcmalloc_pageheap_unmapped_bytes          | inputs.Int    | inputs.Gauge | inputs.NCount | Total bytes on returned freelists.                                                                                                                                                                                                                                                                                          |
| tcmalloc_spinlock_total_delay_ns          | inputs.Int    | inputs.Gauge | inputs.NCount | inputs.TODO                                                                                                                                                                                                                                                                                                                 |
| tcmalloc_thread_cache_free_bytes          | inputs.Int    | inputs.Gauge | inputs.NCount | Bytes in thread caches.                                                                                                                                                                                                                                                                                                     |
| tcmalloc_total_free_bytes                 | inputs.Int    | inputs.Gauge | inputs.NCount | Total bytes on normal freelists.                                                                                                                                                                                                                                                                                            |
| tcmalloc_transfer_cache_free_bytes        | inputs.Int    | inputs.Gauge | inputs.NCount | Bytes in central transfer cache.                                                                                                                                                                                                                                                                                            |
| total_available                           | inputs.Int    | inputs.Gauge | inputs.NCount | The number of connections available from the mongos to the config servers, replica sets, and standalone mongod instances in the cluster.                                                                                                                                                                                    |
| total_created                             | inputs.Int    | inputs.Gauge | inputs.NCount | The number of connections the mongos has ever created to other members of the cluster.                                                                                                                                                                                                                                      |
| total_docs_scanned                        | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of index items scanned during queries and query-plan evaluation.                                                                                                                                                                                                                                           |
| total_in_use                              | inputs.Int    | inputs.Gauge | inputs.NCount | Reports the total number of outgoing connections from the current mongod/mongos instance to other members of the sharded cluster or replica set that are currently in use.                                                                                                                                                  |
| total_keys_scanned                        | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of index items scanned during queries and query-plan evaluation.                                                                                                                                                                                                                                           |
| total_refreshing                          | inputs.Int    | inputs.Gauge | inputs.NCount | Reports the total number of outgoing connections from the current mongod/mongos instance to other members of the sharded cluster or replica set that are currently being refreshed.                                                                                                                                         |
| total_tickets_reads                       | inputs.Int    | inputs.Gauge | inputs.NCount | A document that returns information on the number of concurrent of read transactions allowed into the WiredTiger storage engine.                                                                                                                                                                                            |
| total_tickets_writes                      | inputs.Int    | inputs.Gauge | inputs.NCount | A document that returns information on the number of concurrent of write transactions allowed into the WiredTiger storage engine.                                                                                                                                                                                           |
| ttl_deletes                               | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of documents deleted from collections with a ttl index.                                                                                                                                                                                                                                                    |
| ttl_passes                                | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times the background process removes documents from collections with a ttl index.                                                                                                                                                                                                                             |
| update_command_failed                     | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times that 'update' command failed on this mongod                                                                                                                                                                                                                                                             |
| update_command_total                      | inputs.Int    | inputs.Gauge | inputs.NCount | The number of times that 'update' command executed on this mongod                                                                                                                                                                                                                                                           |
| updates                                   | inputs.Int    | inputs.Gauge | inputs.NCount | The total number of update operations received since the mongod instance last started.                                                                                                                                                                                                                                      |
| uptime_ns                                 | inputs.Int    | inputs.Gauge | inputs.NCount | The total upon time of mongod in nano seconds.                                                                                                                                                                                                                                                                              |
| version                                   | inputs.String | inputs.Gauge | inputs.NCount | Mongod version                                                                                                                                                                                                                                                                                                              |
| vsize_megabytes                           | inputs.Int    | inputs.Gauge | inputs.NCount | mem.virtual displays the quantity, in mebibyte (MiB), of virtual memory used by the mongod process.                                                                                                                                                                                                                         |
| wtcache_app_threads_page_read_count       | inputs.Int    | inputs.Gauge | inputs.NCount | inputs.TODO                                                                                                                                                                                                                                                                                                                 |
| wtcache_app_threads_page_read_time        | inputs.Int    | inputs.Gauge | inputs.NCount | inputs.TODO                                                                                                                                                                                                                                                                                                                 |
| wtcache_app_threads_page_write_count      | inputs.Int    | inputs.Gauge | inputs.NCount | inputs.TODO                                                                                                                                                                                                                                                                                                                 |
| wtcache_bytes_read_into                   | inputs.Int    | inputs.Gauge | inputs.NCount | inputs.TODO                                                                                                                                                                                                                                                                                                                 |
| wtcache_bytes_written_from                | inputs.Int    | inputs.Gauge | inputs.NCount | inputs.TODO                                                                                                                                                                                                                                                                                                                 |
| wtcache_current_bytes                     | inputs.Int    | inputs.Gauge | inputs.NCount | inputs.TODO                                                                                                                                                                                                                                                                                                                 |
| wtcache_internal_pages_evicted            | inputs.Int    | inputs.Gauge | inputs.NCount | inputs.TODO                                                                                                                                                                                                                                                                                                                 |
| wtcache_max_bytes_configured              | inputs.Int    | inputs.Gauge | inputs.NCount | Maximum cache size.                                                                                                                                                                                                                                                                                                         |
| wtcache_modified_pages_evicted            | inputs.Int    | inputs.Gauge | inputs.NCount | inputs.TODO                                                                                                                                                                                                                                                                                                                 |
| wtcache_pages_evicted_by_app_thread       | inputs.Int    | inputs.Gauge | inputs.NCount | inputs.TODO                                                                                                                                                                                                                                                                                                                 |
| wtcache_pages_queued_for_eviction         | inputs.Int    | inputs.Gauge | inputs.NCount | inputs.TODO                                                                                                                                                                                                                                                                                                                 |
| wtcache_pages_read_into                   | inputs.Int    | inputs.Gauge | inputs.NCount | Number of pages read into the cache.                                                                                                                                                                                                                                                                                        |
| wtcache_pages_requested_from              | inputs.Int    | inputs.Gauge | inputs.NCount | Number of pages request from the cache.                                                                                                                                                                                                                                                                                     |
| wtcache_server_evicting_pages             | inputs.Int    | inputs.Gauge | inputs.NCount | inputs.TODO                                                                                                                                                                                                                                                                                                                 |
| wtcache_tracked_dirty_bytes               | inputs.Int    | inputs.Gauge | inputs.NCount | inputs.TODO                                                                                                                                                                                                                                                                                                                 |
| wtcache_unmodified_pages_evicted          | inputs.Int    | inputs.Gauge | inputs.NCount | Main statistics for page eviction.                                                                                                                                                                                                                                                                                          |
| wtcache_worker_thread_evictingpages       | inputs.Int    | inputs.Gauge | inputs.NCount | inputs.TODO                                                                                                                                                                                                                                                                                                                 |

## mongodb_db_stats

| 标签名   | 描述          |
| -------- | ------------- |
| db_name  | database name |
| hostname | mongodb host  |

| 指标         | 类型          | 指标源       | 单位          | 描述                                                                                                             |
| ------------ | ------------- | ------------ | ------------- | ---------------------------------------------------------------------------------------------------------------- |
| avg_obj_size | inputs.Float  | inputs.Gauge | inputs.NCount | The average size of each document in bytes.                                                                      |
| collections  | inputs.Int    | inputs.Gauge | inputs.NCount | Contains a count of the number of collections in that database.                                                  |
| data_size    | inputs.Int    | inputs.Gauge | inputs.NCount | The total size of the uncompressed data held in this database. The dataSize decreases when you remove documents. |
| index_size   | inputs.Int    | inputs.Gauge | inputs.NCount | The total size of all indexes created on this database.                                                          |
| indexes      | inputs.Int    | inputs.Gauge | inputs.NCount | Contains a count of the total number of indexes across all collections in the database.                          |
| objects      | inputs.Int    | inputs.Gauge | inputs.NCount | Contains a count of the number of objects (i.e. documents) in the database across all collections.               |
| ok           | inputs.Int    | inputs.Gauge | inputs.NCount | Command execute state.                                                                                           |
| storage_size | inputs.Int    | inputs.Gauge | inputs.NCount | The total amount of space allocated to collections in this database for document storage.                        |
| type         | inputs.String | inputs.Gauge | inputs.NCount | Metric type.                                                                                                     |

## mongodb_col_stats

| 标签名     | 描述            |
| ---------- | --------------- |
| collection | collection name |
| db_name    | database name   |
| hostname   | mongodb host    |

| 指标             | 类型       | 指标源       | 单位          | 描述                                                                          |
| ---------------- | ---------- | ------------ | ------------- | ----------------------------------------------------------------------------- |
| avg_obj_size     | inputs.Int | inputs.Gauge | inputs.NCount | The average size of an object in the collection.                              |
| count            | inputs.Int | inputs.Gauge | inputs.NCount | The number of objects or documents in this collection.                        |
| ok               | inputs.Int | inputs.Gauge | inputs.NCount | Command execute state.                                                        |
| size             | inputs.Int | inputs.Gauge | inputs.NCount | The total uncompressed size in memory of all records in a collection.         |
| storage_size     | inputs.Int | inputs.Gauge | inputs.NCount | The total amount of storage allocated to this collection for document storage |
| total_index_size | inputs.Int | inputs.Gauge | inputs.NCount | The total size of all indexes.                                                |
| type             | inputs.Int | inputs.Gauge | inputs.NCount | Metrics type.                                                                 |

## mongodb_shard_stats

| 标签名   | 描述         |
| -------- | ------------ |
| hostname | mongodb host |

| 指标       | 类型       | 指标源       | 单位          | 描述                                                                                                                                                                                |
| ---------- | ---------- | ------------ | ------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| available  | inputs.Int | inputs.Gauge | inputs.NCount | The number of connections available for this host to connect to the mongos.                                                                                                         |
| created    | inputs.Int | inputs.Gauge | inputs.NCount | The number of connections the host has ever created to connect to the mongos.                                                                                                       |
| in_use     | inputs.Int | inputs.Gauge | inputs.NCount | Reports the total number of outgoing connections from the current mongod/mongos instance to other members of the sharded cluster or replica set that are currently in use.          |
| refreshing | inputs.Int | inputs.Gauge | inputs.NCount | Reports the total number of outgoing connections from the current mongod/mongos instance to other members of the sharded cluster or replica set that are currently being refreshed. |

## mongodb_top_stats

| 标签名     | 描述            |
| ---------- | --------------- |
| hostname   | mongodb host    |
| collection | collection name |

| 指标             | 类型       | 指标源       | 单位          | 描述                                                       |
| ---------------- | ---------- | ------------ | ------------- | ---------------------------------------------------------- |
| commands_count   | inputs.Int | inputs.Gauge | inputs.NCount | The total number of "command" event issues.                |
| commands_time    | inputs.Int | inputs.Gauge | inputs.NCount | The amount of time in microseconds that "command" costs.   |
| get_more_count   | inputs.Int | inputs.Gauge | inputs.NCount | The total number of "getmore" event issues.                |
| get_more_time    | inputs.Int | inputs.Gauge | inputs.NCount | The amount of time in microseconds that "getmore" costs.   |
| insert_count     | inputs.Int | inputs.Gauge | inputs.NCount | The total number of "insert" event issues.                 |
| insert_time      | inputs.Int | inputs.Gauge | inputs.NCount | The amount of time in microseconds that "insert" costs.    |
| queries_count    | inputs.Int | inputs.Gauge | inputs.NCount | The total number of "queries" event issues.                |
| queries_time     | inputs.Int | inputs.Gauge | inputs.NCount | The amount of time in microseconds that "queries" costs.   |
| read_lock_count  | inputs.Int | inputs.Gauge | inputs.NCount | The total number of "readLock" event issues.               |
| read_lock_time   | inputs.Int | inputs.Gauge | inputs.NCount | The amount of time in microseconds that "readLock" costs.  |
| remove_count     | inputs.Int | inputs.Gauge | inputs.NCount | The total number of "remove" event issues.                 |
| remove_time      | inputs.Int | inputs.Gauge | inputs.NCount | The amount of time in microseconds that "remove" costs.    |
| total_count      | inputs.Int | inputs.Gauge | inputs.NCount | The total number of "total" event issues.                  |
| total_time       | inputs.Int | inputs.Gauge | inputs.NCount | The amount of time in microseconds that "total" costs.     |
| update_count     | inputs.Int | inputs.Gauge | inputs.NCount | The total number of "update" event issues.                 |
| update_time      | inputs.Int | inputs.Gauge | inputs.NCount | The amount of time in microseconds that "update" costs.    |
| write_lock_count | inputs.Int | inputs.Gauge | inputs.NCount | The total number of "writeLock" event issues.              |
| write_lock_time  | inputs.Int | inputs.Gauge | inputs.NCount | The amount of time in microseconds that "writeLock" costs. |

## mongod_log

| 标签名   | 描述                       |
| -------- | -------------------------- |
| filename | The file name to 'tail -f' |
| host     | host name                  |
| service  | service name: mongod_log   |

| 字段名    | 字段值 | 说明                                                           |
| --------- | ------ | -------------------------------------------------------------- |
| message   | string | Log raw data .                                                 |
| component | string | The full component string of the log message                   |
| context   | string | The name of the thread issuing the log statement               |
| msg       | string | The raw log output message as passed from the server or driver |
| severity  | string | The short severity code of the log message                     |
| status    | string | Log level                                                      |
| time      | string | Timestamp                                                      |
