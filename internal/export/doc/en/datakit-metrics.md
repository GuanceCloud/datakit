# DataKit Exported Metrics

---

Foe better self-observability, DataKit has exported many Prometheus metrics among various submodules. We can use these metrics to trouble shooting errors or bugs during DataKit running.

## Exported Metrics {#metrics}

Since DataKit [Version 1.5.9](changelog.md#cl-1.5.9), we can visit `http://localhost:9529/metrics` to see these metrics(Different release may got some new or renamed metrics).

We have used these metrics in [DataKit monitor](datakit-monitor.md), and we can *playing* these metrics with some tricks like this(refresh CPU usage of DataKit):

```shell
# play CPU usage of DataKit on every 3 seconds
$ watch -n 3 'curl -s http://localhost:9529/metrics | grep -a datakit_cpu_usage'

# HELP datakit_cpu_usage DataKit CPU usage(%)
# TYPE datakit_cpu_usage gauge
datakit_cpu_usage 4.9920266849857144
```

We can also playing other metrics too(change the `grep` string), all available metrics list below(current DataKit version is {{ .Version }}):

<!-- we can run `make metrics` go export all these metrics -->
| TYPE    | NAME                                                               | LABELS                                                                                            | HELP                                                                                                                 |
| ---     | ---                                                                | ---                                                                                               | ---                                                                                                                  |
| GAUGE   | `datakit_config_datakit_ulimit`                                    | `status`                                                                                          | DataKit ulimit                                                                                                       |
| GAUGE   | `datakit_kv_last_update_timestamp_seconds`                         | `N/A`                                                                                             | KV last update time                                                                                                  |
| SUMMARY | `datakit_kv_pull_latency_seconds`                                  | `status`                                                                                          | KV pull latency                                                                                                      |
| COUNTER | `datakit_kv_update_total`                                          | `name`                                                                                            | KV updated count                                                                                                     |
| GAUGE   | `datakit_kv_input_last_reload_timestamp_seconds`                   | `source`                                                                                          | Input last reload timestamp                                                                                          |
| COUNTER | `datakit_kv_input_reload_total`                                    | `source`                                                                                          | Input reload count                                                                                                   |
| COUNTER | `datakit_dns_domain_total`                                         | `N/A`                                                                                             | DNS watched domain counter                                                                                           |
| COUNTER | `datakit_dns_ip_updated_total`                                     | `domain`                                                                                          | Domain IP updated counter                                                                                            |
| COUNTER | `datakit_dns_watch_run_total`                                      | `interval`                                                                                        | Watch run counter                                                                                                    |
| SUMMARY | `datakit_dns_cost_seconds`                                         | `domain,status`                                                                                   | DNS IP lookup cost                                                                                                   |
| COUNTER | `datakit_election_pause_total`                                     | `id,namespace`                                                                                    | Input paused count when election failed                                                                              |
| COUNTER | `datakit_election_resume_total`                                    | `id,namespace`                                                                                    | Input resume count when election OK                                                                                  |
| GAUGE   | `datakit_election_status`                                          | `elected_id,id,namespace,status`                                                                  | DataKit election status, if metric = 0, meas not elected, or the elected time(unix timestamp second)                 |
| GAUGE   | `datakit_election_inputs`                                          | `namespace`                                                                                       | DataKit election input count                                                                                         |
| SUMMARY | `datakit_election_seconds`                                         | `namespace,status`                                                                                | Election latency                                                                                                     |
| GAUGE   | `datakit_goroutine_alive`                                          | `name`                                                                                            | Alive Goroutine count                                                                                                |
| COUNTER | `datakit_goroutine_recover_total`                                  | `name`                                                                                            | Recovered Goroutine count                                                                                            |
| COUNTER | `datakit_goroutine_stopped_total`                                  | `name`                                                                                            | Stopped Goroutine count                                                                                              |
| COUNTER | `datakit_goroutine_crashed_total`                                  | `name`                                                                                            | Crashed goroutines count                                                                                             |
| GAUGE   | `datakit_goroutine_groups`                                         | `N/A`                                                                                             | Goroutine group count                                                                                                |
| SUMMARY | `datakit_goroutine_cost_seconds`                                   | `name`                                                                                            | Goroutine running duration                                                                                           |
| SUMMARY | `datakit_http_api_elapsed_seconds`                                 | `api,method,status`                                                                               | API request cost                                                                                                     |
| SUMMARY | `datakit_http_api_req_size_bytes`                                  | `api,method,status`                                                                               | API request body size                                                                                                |
| COUNTER | `datakit_http_api_total`                                           | `api,method,status`                                                                               | API request counter                                                                                                  |
| GAUGE   | `datakit_http_api_global_tags_last_updated`                        | `api,method,status`                                                                               | Global tag updated timestamp, in second                                                                              |
| SUMMARY | `datakit_httpcli_got_first_resp_byte_cost_seconds`                 | `from`                                                                                            | Got first response byte cost                                                                                         |
| COUNTER | `datakit_httpcli_tcp_conn_total`                                   | `from,remote,type`                                                                                | HTTP TCP connection count                                                                                            |
| COUNTER | `datakit_httpcli_conn_reused_from_idle_total`                      | `from`                                                                                            | HTTP connection reused from idle count                                                                               |
| SUMMARY | `datakit_httpcli_conn_idle_time_seconds`                           | `from`                                                                                            | HTTP connection idle time                                                                                            |
| SUMMARY | `datakit_httpcli_dns_cost_seconds`                                 | `from`                                                                                            | HTTP DNS cost                                                                                                        |
| SUMMARY | `datakit_httpcli_tls_handshake_seconds`                            | `from`                                                                                            | HTTP TLS handshake cost                                                                                              |
| SUMMARY | `datakit_httpcli_http_connect_cost_seconds`                        | `from`                                                                                            | HTTP connect cost                                                                                                    |
| SUMMARY | `datakit_io_build_body_batch_bytes`                                | `category,encoding,type`                                                                          | Batch HTTP body size                                                                                                 |
| SUMMARY | `datakit_io_build_body_batch_points`                               | `category,encoding`                                                                               | Batch HTTP body points                                                                                               |
| SUMMARY | `datakit_io_dataway_wal_flush`                                     | `category,gzip,queue`                                                                             | Dataway WAL worker flushed bytes                                                                                     |
| COUNTER | `datakit_io_dataway_skipped_point_total`                           | `category`                                                                                        | Skipped point count during encoding(Protobuf) point                                                                  |
| COUNTER | `datakit_io_dataway_point_total`                                   | `category,status`                                                                                 | Dataway uploaded points, partitioned by category and send status(HTTP status)                                        |
| COUNTER | `datakit_io_dataway_body_total`                                    | `from,op,type`                                                                                    | Dataway total body                                                                                                   |
| COUNTER | `datakit_io_wal_point_total`                                       | `category,status`                                                                                 | WAL queued points                                                                                                    |
| SUMMARY | `datakit_io_wal_put_retried`                                       | `category`                                                                                        | WAL put retried on disk full                                                                                         |
| COUNTER | `datakit_io_dataway_point_bytes_total`                             | `category,enc,status`                                                                             | Dataway uploaded points bytes, partitioned by category and pint send status(HTTP status)                             |
| COUNTER | `datakit_io_dataway_http_drop_point_total`                         | `category,error`                                                                                  | Dataway write drop points                                                                                            |
| SUMMARY | `datakit_io_dataway_api_latency_seconds`                           | `api,status`                                                                                      | Dataway HTTP request latency partitioned by HTTP API(method@url) and HTTP status                                     |
| COUNTER | `datakit_io_http_retry_total`                                      | `api,status`                                                                                      | Dataway HTTP retried count                                                                                           |
| SUMMARY | `datakit_io_grouped_request`                                       | `category`                                                                                        | Grouped requests under sinker                                                                                        |
| GAUGE   | `datakit_io_dataway_wal_mem_len`                                   | `category`                                                                                        | Dataway WAL's memory queue length                                                                                    |
| COUNTER | `datakit_io_flush_drop_pkg_total`                                  | `category`                                                                                        | WAL flush dropped packages count due to expiration                                                                   |
| SUMMARY | `datakit_io_flush_failcache_bytes`                                 | `category`                                                                                        | IO flush fail-cache bytes(in gzip) summary                                                                           |
| SUMMARY | `datakit_io_build_body_cost_seconds`                               | `category,encoding,stage`                                                                         | Build point HTTP body cost                                                                                           |
| SUMMARY | `datakit_io_build_body_batches`                                    | `category,encoding`                                                                               | Batch HTTP body batches                                                                                              |
| SUMMARY | `datakit_io_build_body_points`                                     | `category,encoding`                                                                               | Point count for single compact                                                                                       |
| COUNTER | `datakit_filter_update_total`                                      | `N/A`                                                                                             | Filters(remote) updated count                                                                                        |
| GAUGE   | `datakit_filter_last_update_timestamp_seconds`                     | `N/A`                                                                                             | Filter last update time                                                                                              |
| COUNTER | `datakit_filter_point_total`                                       | `category,filters,source`                                                                         | Filter points of filters                                                                                             |
| GAUGE   | `datakit_filter_parse_error`                                       | `error,filters`                                                                                   | Filter parse error                                                                                                   |
| COUNTER | `datakit_filter_point_dropped_total`                               | `category,filters,source`                                                                         | Dropped points of filters                                                                                            |
| SUMMARY | `datakit_filter_pull_latency_seconds`                              | `status`                                                                                          | Filter pull(remote) latency                                                                                          |
| SUMMARY | `datakit_filter_latency_seconds`                                   | `category,filters,source`                                                                         | Filter latency of these filters                                                                                      |
| COUNTER | `datakit_io_point_time_adjusted_total`                             | `category,name`                                                                                   | Point's time has been adjusted due to invalid timestamp(larger than 2h)                                              |
| GAUGE   | `datakit_io_queue_points`                                          | `category`                                                                                        | IO module queued(cached) points                                                                                      |
| COUNTER | `datakit_io_input_filter_point_total`                              | `name,category`                                                                                   | Input filtered point total                                                                                           |
| COUNTER | `datakit_io_feed_total`                                            | `name,category`                                                                                   | Input feed total                                                                                                     |
| GAUGE   | `datakit_io_last_feed_timestamp_seconds`                           | `name,category`                                                                                   | Input last feed time(according to DataKit local time)                                                                |
| SUMMARY | `datakit_input_collect_latency_seconds`                            | `name,category`                                                                                   | Input collect latency                                                                                                |
| GAUGE   | `datakit_io_chan_usage`                                            | `category`                                                                                        | IO channel usage(length of the channel)                                                                              |
| GAUGE   | `datakit_io_chan_capacity`                                         | `category`                                                                                        | IO channel capacity                                                                                                  |
| SUMMARY | `datakit_io_feed_cost_seconds`                                     | `category,from`                                                                                   | IO feed waiting(on block mode) seconds                                                                               |
| SUMMARY | `datakit_io_feed_point`                                            | `name,category`                                                                                   | Input feed point                                                                                                     |
| GAUGE   | `datakit_io_flush_workers`                                         | `category`                                                                                        | IO flush workers                                                                                                     |
| COUNTER | `datakit_io_flush_total`                                           | `category`                                                                                        | IO flush total                                                                                                       |
| SUMMARY | `datakit_input_tailer_scanner_cost_seconds`                        | `pattern`                                                                                         | Scanning costs seconds                                                                                               |
| SUMMARY | `datakit_input_tailer_scanner_files`                               | `pattern`                                                                                         | Total number of scanned files                                                                                        |
| COUNTER | `datakit_error_total`                                              | `source,category`                                                                                 | Total errors, only count on error source, not include error message                                                  |
| GAUGE   | `datakit_goroutines`                                               | `N/A`                                                                                             | Goroutine count within DataKit                                                                                       |
| GAUGE   | `datakit_mem_stat`                                                 | `type`                                                                                            | DataKit memory system bytes                                                                                          |
| GAUGE   | `datakit_heap_alloc_bytes`                                         | `N/A`                                                                                             | DataKit memory heap bytes(Deprecated by `datakit_golang_mem_usage`)                                                  |
| GAUGE   | `datakit_sys_alloc_bytes`                                          | `N/A`                                                                                             | DataKit memory system bytes(Deprecated by `datakit_golang_mem_usage`)                                                |
| GAUGE   | `datakit_golang_mem_usage`                                         | `type`                                                                                            | DataKit golang memory usage stats                                                                                    |
| GAUGE   | `datakit_cpu_usage`                                                | `N/A`                                                                                             | DataKit CPU usage(%)                                                                                                 |
| GAUGE   | `datakit_open_files`                                               | `N/A`                                                                                             | DataKit open files(only available on Linux)                                                                          |
| GAUGE   | `datakit_cpu_cores`                                                | `N/A`                                                                                             | DataKit CPU cores                                                                                                    |
| GAUGE   | `datakit_uptime_seconds`                                           | `auto_update,docker,hostname,lite,elinker,resource_limit,version=?,build_at=?,branch=?,os_arch=?` | DataKit uptime                                                                                                       |
| GAUGE   | `datakit_data_overuse`                                             | `N/A`                                                                                             | Does current workspace's data(metric/logging) usage(if 0 not beyond, or with a unix timestamp when overuse occurred) |
| COUNTER | `datakit_process_ctx_switch_total`                                 | `type`                                                                                            | DataKit process context switch count(Linux only)                                                                     |
| COUNTER | `datakit_process_io_count_total`                                   | `type`                                                                                            | DataKit process IO count                                                                                             |
| COUNTER | `datakit_process_io_bytes_total`                                   | `type`                                                                                            | DataKit process IO bytes count                                                                                       |
| COUNTER | `datakit_ntp_sync_total`                                           | `N/A`                                                                                             | Total count synced with remote NTP server                                                                            |
| SUMMARY | `datakit_ntp_time_diff`                                            | `N/A`                                                                                             | Time difference(seconds) between remote NTP server                                                                   |
| COUNTER | `datakit_pipeline_offload_point_total`                             | `category,exporter,remote`                                                                        | Pipeline offload processed total points                                                                              |
| COUNTER | `datakit_pipeline_offload_error_point_total`                       | `category,exporter,remote`                                                                        | Pipeline offload processed total error points                                                                        |
| SUMMARY | `datakit_pipeline_offload_cost_seconds`                            | `category,exporter,remote`                                                                        | Pipeline offload total cost                                                                                          |
| COUNTER | `dkebpf_exporter_points_total`                                     | `name,category`                                                                                   | The number of data points processed by the exporter                                                                  |
| COUNTER | `datakit_input_container_kubernetes_resources_change_total`        | `resource,type`                                                                                   | Total number of objects changed for Kubernetes resources.                                                            |
| SUMMARY | `datakit_input_container_kubernetes_collect_cost_seconds`          | `category`                                                                                        | Total time (in seconds) spent collecting metrics from Kubernetes                                                     |
| COUNTER | `datakit_input_container_kubernetes_collect_pts_total`             | `category`                                                                                        | Total number of points collected from Kubernetes resources                                                           |
| SUMMARY | `datakit_input_container_kubernetes_collect_resource_cost_seconds` | `category,name`                                                                                   | Total time (in seconds) spent collecting resource metrics from Kubernetes                                            |
| COUNTER | `datakit_input_container_kubernetes_pod_metrics_query_total`       | `target`                                                                                          | Total number of metric queries made to Kubernetes Pods                                                               |
| SUMMARY | `datakit_input_container_kubernetes_pod_annotation_prom_count`     | `name`                                                                                            | The number of Prometheus-related annotations found in Kubernetes Pods                                                |
| SUMMARY | `datakit_input_container_collect_cost_seconds`                     | `category`                                                                                        | Total time (in seconds) spent collecting container metrics                                                           |
| COUNTER | `datakit_input_container_collect_pts_total`                        | `category`                                                                                        | Total number of points collected from containers                                                                     |
| COUNTER | `datakit_input_ddtrace_truncated_spans_total`                      | `input`                                                                                           | Truncated trace spans                                                                                                |
| COUNTER | `datakit_input_ddtrace_dropped_trace_total`                        | `url`                                                                                             | Dropped illegal traces                                                                                               |
| SUMMARY | `datakit_input_ddtrace_trace_spans`                                | `input`                                                                                           | Trace spans(include truncated spans)                                                                                 |
| SUMMARY | `datakit_dialtesting_task_run_cost_seconds`                        | `region,protocol`                                                                                 | Task run time                                                                                                        |
| SUMMARY | `datakit_dialtesting_task_exec_time_interval_seconds`              | `region,protocol`                                                                                 | Task execution time interval                                                                                         |
| GAUGE   | `datakit_dialtesting_task_max_icmp_concurrency`                    | `N/A`                                                                                             | The max number of icmp packets sent at one time                                                                      |
| GAUGE   | `datakit_dialtesting_task_icmp_concurrency`                       | `N/A`                                                                                             | The current number of icmp packets sending                                                                           |
| GAUGE   | `datakit_dialtesting_worker_job_chan_number`                       | `type`                                                                                            | The number of the channel for the jobs                                                                               |
| GAUGE   | `datakit_dialtesting_worker_job_number`                            | `N/A`                                                                                             | The number of the jobs to send data in parallel                                                                      |
| GAUGE   | `datakit_dialtesting_worker_cached_points_number`                  | `region,protocol`                                                                                 | The number of cached points                                                                                          |
| GAUGE   | `datakit_dialtesting_worker_dropped_points_number`                  | `region,protocol`                                                                                 | The number of cached points has been dropped                                                                                          |
| GAUGE   | `datakit_dialtesting_worker_send_points_number`                    | `region,protocol,status`                                                                          | The number of the points which have been sent                                                                        |
| SUMMARY | `datakit_dialtesting_worker_send_cost_seconds`                     | `region,protocol`                                                                                 | Time cost to send points                                                                                             |
| GAUGE   | `datakit_dialtesting_task_number`                                  | `region,protocol`                                                                                 | The number of tasks                                                                                                  |
| GAUGE   | `datakit_dialtesting_dataway_send_failed_number`                   | `region,protocol`                                                                                 | The number of failed sending                                                                                         |
| SUMMARY | `datakit_dialtesting_pull_cost_seconds`                            | `region,is_first`                                                                                 | Time cost to pull tasks                                                                                              |
| COUNTER | `datakit_dialtesting_task_synchronized_total`                      | `region,protocol`                                                                                 | Task synchronized number                                                                                             |
| COUNTER | `datakit_dialtesting_task_invalid_total`                           | `region,protocol,fail_reason`                                                                     | Invalid task number                                                                                                  |
| SUMMARY | `datakit_dialtesting_task_check_cost_seconds`                      | `region,protocol,status`                                                                          | Task check time                                                                                                      |
| GAUGE   | `datakit_input_graphite_metric_mapper_cache_length`                | `N/A`                                                                                             | The count of unique metrics currently cached.                                                                        |
| COUNTER | `datakit_input_graphite_metric_cache_gets_total`                   | `N/A`                                                                                             | The count of total metric cache gets.                                                                                |
| COUNTER | `datakit_input_graphite_metric_mapper_cache_hits_total`            | `N/A`                                                                                             | The count of total metric cache hits.                                                                                |
| COUNTER | `datakit_input_graphite_tag_parse_failures_total`                  | `N/A`                                                                                             | Total count of samples with invalid tags                                                                             |
| GAUGE   | `datakit_input_graphite_last_processed_timestamp_seconds`          | `N/A`                                                                                             | Unix timestamp of the last processed graphite metric.                                                                |
| GAUGE   | `datakit_input_graphite_sample_expiry_seconds`                     | `N/A`                                                                                             | How long in seconds a metric sample is valid for.                                                                    |
| COUNTER | `datakit_input_kafkamq_consumer_message_total`                     | `topic,partition,status`                                                                          | Kafka consumer message numbers from DataKit start                                                                    |
| COUNTER | `datakit_input_kafkamq_group_election_total`                       | `N/A`                                                                                             | Kafka group election count                                                                                           |
| SUMMARY | `datakit_input_kafkamq_process_message_nano`                       | `topic`                                                                                           | kafkamq process message nanoseconds duration                                                                         |
| COUNTER | `datakit_input_kubernetesprometheus_collect_pts_total`             | `role,name`                                                                                       | The number of the points which have been sent                                                                        |
| SUMMARY | `datakit_input_kubernetesprometheus_collect_cost_seconds`          | `role,name,url`                                                                                   | The collect cost in seconds                                                                                          |
| GAUGE   | `datakit_input_kubernetesprometheus_scraper_number`                | `role,name`                                                                                       | The number of the scraper                                                                                            |
| GAUGE   | `datakit_input_kubernetesprometheus_task_number`                   | `worker`                                                                                          | The number of the task                                                                                               |
| GAUGE   | `datakit_inputs_instance`                                          | `input`                                                                                           | Input instance count                                                                                                 |
| COUNTER | `datakit_inputs_crash_total`                                       | `input`                                                                                           | Input crash count                                                                                                    |
| SUMMARY | `datakit_input_mysql_sql_query_cost_seconds`                       | `metric_name,sql_name`                                                                            | Time cost to query SQL                                                                                               |
| SUMMARY | `datakit_input_oracle_sql_query_cost_seconds`                      | `metric_name,sql_name`                                                                            | Time cost to query SQL                                                                                               |
| GAUGE   | `datakit_input_ploffload_chan_capacity`                            | `channel_name`                                                                                    | PlOffload channel capacity                                                                                           |
| GAUGE   | `datakit_input_ploffload_chan_usage`                               | `channel_name`                                                                                    | PlOffload channel usage                                                                                              |
| COUNTER | `datakit_input_ploffload_point_total`                              | `category`                                                                                        | PlOffload processed total points                                                                                     |
| SUMMARY | `datakit_input_postgresql_sql_query_cost_seconds`                  | `metric_name,sql_name`                                                                            | Time cost to query SQL                                                                                               |
| SUMMARY | `datakit_input_promremote_collect_points`                          | `source`                                                                                          | Total number of promremote collection points                                                                         |
| SUMMARY | `datakit_input_promremote_time_diff_in_second`                     | `source`                                                                                          | Time diff with local time                                                                                            |
| COUNTER | `datakit_input_promremote_no_time_points_total`                    | `source`                                                                                          | Total number of promremote collection no time points                                                                 |
| SUMMARY | `datakit_input_promv2_scrape_points`                               | `source,remote`                                                                                   | The number of points scrape from endpoint                                                                            |
| GAUGE   | `api_elapsed_seconds`                                              | `N/A`                                                                                             | Proxied API elapsed seconds                                                                                          |
| COUNTER | `api_post_bytes_total`                                             | `api,status`                                                                                      | Proxied API post bytes total                                                                                         |
| SUMMARY | `api_latency_seconds`                                              | `api,status`                                                                                      | Proxied API latency                                                                                                  |
| COUNTER | `datakit_input_proxy_connect_total`                                | `client_ip`                                                                                       | Proxy connect(method CONNECT)                                                                                        |
| COUNTER | `datakit_input_proxy_api_total`                                    | `api,method`                                                                                      | Proxy API total                                                                                                      |
| SUMMARY | `datakit_input_proxy_api_latency_seconds`                          | `api,method,status`                                                                               | Proxy API latency                                                                                                    |
| COUNTER | `datakit_input_proxy_reject_total`                                 | `type`                                                                                            | Proxy rejected request                                                                                               |
| COUNTER | `datakit_input_rum_session_replay_drop_total`                      | `app_id,env,version,service`                                                                      | statistics the total count of session replay points which have been filtered by rules                                |
| COUNTER | `datakit_input_rum_session_replay_drop_bytes_total`                | `app_id,env,version,service`                                                                      | statistics the total bytes of session replay points which have been filtered by rules                                |
| COUNTER | `datakit_input_rum_locate_statistics_total`                        | `app_id,ip_status,locate_status`                                                                  | locate by ip addr statistics                                                                                         |
| COUNTER | `datakit_input_rum_source_map_total`                               | `app_id,sdk_name,status,remark`                                                                   | source map result statistics                                                                                         |
| GAUGE   | `datakit_input_rum_loaded_zips`                                    | `platform`                                                                                        | RUM source map currently loaded zip archive count                                                                    |
| SUMMARY | `datakit_input_rum_source_map_duration_seconds`                    | `sdk_name,app_id,env,version`                                                                     | statistics elapsed time in RUM source map(unit: second)                                                              |
| SUMMARY | `datakit_input_rum_session_replay_upload_latency_seconds`          | `app_id,env,version,service,status_code`                                                          | statistics elapsed time in session replay uploading                                                                  |
| COUNTER | `datakit_input_rum_session_replay_upload_failure_total`            | `app_id,env,version,service,status_code`                                                          | statistics count of session replay points which which have unsuccessfully uploaded                                   |
| COUNTER | `datakit_input_rum_session_replay_upload_failure_bytes_total`      | `app_id,env,version,service,status_code`                                                          | statistics the total bytes of session replay points which have unsuccessfully uploaded                               |
| SUMMARY | `datakit_input_rum_session_replay_read_body_delay_seconds`         | `app_id,env,version,service`                                                                      | statistics the duration of reading session replay body                                                               |
| SUMMARY | `datakit_input_snmp_discovery_cost`                                | `profile_type`                                                                                    | Discovery cost(in second)                                                                                            |
| SUMMARY | `datakit_input_snmp_collect_cost`                                  | `N/A`                                                                                             | Every loop collect cost(in second)                                                                                   |
| SUMMARY | `datakit_input_snmp_device_collect_cost`                           | `class`                                                                                           | Device collect cost(in second)                                                                                       |
| GAUGE   | `datakit_input_snmp_alive_devices`                                 | `class`                                                                                           | Alive devices                                                                                                        |
| SUMMARY | `datakit_input_sqlserver_sql_query_cost_seconds`                   | `metric_name,sql_name`                                                                            | Time cost to query SQL                                                                                               |
| GAUGE   | `datakit_windowsremote_reachable_ip`                               | `N/A`                                                                                             | Total number of IPs reachable through tcp/upd port verification.                                                     |
| COUNTER | `datakit_windowsremote_wmi_conn_total`                             | `status`                                                                                          | The WMI connects count                                                                                               |
| COUNTER | `datakit_input_zabbix_exporter_collect_metric_total`               | `object`                                                                                          | exporter metric count number from start                                                                              |
| COUNTER | `datakit_input_zabbix_exporter_collect_file_total`                 | `object`                                                                                          | The files number of exporter file                                                                                    |
| SUMMARY | `datakit_input_zabbix_exporter_request_api`                        | `status`                                                                                          | The time of success or failed API requests                                                                           |
| SUMMARY | `datakit_input_prom_collect_points`                                | `mode,source`                                                                                     | Total number of prom collection points                                                                               |
| SUMMARY | `datakit_input_prom_http_get_bytes`                                | `mode,source`                                                                                     | HTTP get bytes                                                                                                       |
| SUMMARY | `datakit_input_prom_http_latency_in_second`                        | `mode,source`                                                                                     | HTTP latency(in second)                                                                                              |
| GAUGE   | `datakit_input_prom_stream_size`                                   | `mode,source`                                                                                     | Stream size                                                                                                          |
| SUMMARY | `datakit_remote_job_jvm_dump`                                      | `name,status`                                                                                     | JVM dump job execution time statistics                                                                               |
| SUMMARY | `datakit_input_statsd_collect_points`                              | `N/A`                                                                                             | Total number of statsd collection points                                                                             |
| SUMMARY | `datakit_input_statsd_accept_bytes`                                | `N/A`                                                                                             | Accept bytes from network                                                                                            |
| COUNTER | `datakit_tailer_socket_connect_status_total`                       | `network,status`                                                                                  | Total number of socket connection status events                                                                      |
| COUNTER | `datakit_tailer_socket_feed_message_count_total`                   | `network`                                                                                         | Total number of messages fed to the socket                                                                           |
| SUMMARY | `datakit_tailer_socket_log_length`                                 | `network`                                                                                         | Length of the log for socket communication                                                                           |
| COUNTER | `datakit_tailer_receive_create_event_total`                        | `source,type`                                                                                     | Total number of received create events                                                                               |
| COUNTER | `datakit_tailer_discard_log_total`                                 | `source,filepath`                                                                                 | Total number of discarded based on the whitelist                                                                     |
| GAUGE   | `datakit_tailer_open_files`                                        | `source,max`                                                                                      | Total number of currently open files                                                                                 |
| COUNTER | `datakit_tailer_file_rotate_total`                                 | `source,filepath`                                                                                 | Total number of file rotations performed                                                                             |
| COUNTER | `datakit_tailer_parse_fail_total`                                  | `source,filepath,mode`                                                                            | Total number of failed parse attempts                                                                                |
| COUNTER | `datakit_tailer_multiline_state_total`                             | `source,state`                                                                                    | Total number of multiline states encountered                                                                         |
| COUNTER | `datakit_input_tracing_total`                                      | `input,service`                                                                                   | The total links number of Trace processed by the trace module                                                        |
| COUNTER | `datakit_input_sampler_total`                                      | `input,service`                                                                                   | The sampler number of Trace processed by the trace module                                                            |
| SUMMARY | `datakit_grpc_trace_payload_bytes`                                 | `method`                                                                                          | The payload size of gRPC request send to DataKit                                                                     |
| SUMMARY | `diskcache_dropped_data`                                           | `path,reason`                                                                                     | Dropped data during Put() when capacity reached.                                                                     |
| COUNTER | `diskcache_rotate_total`                                           | `path`                                                                                            | Cache rotate count, mean file rotate from data to data.0000xxx                                                       |
| COUNTER | `diskcache_remove_total`                                           | `path`                                                                                            | Removed file count, if some file read EOF, remove it from un-read list                                               |
| COUNTER | `diskcache_wakeup_total`                                           | `path`                                                                                            | Wakeup count on sleeping write file                                                                                  |
| COUNTER | `diskcache_pos_updated_total`                                      | `op,path`                                                                                         | .pos file updated count                                                                                              |
| COUNTER | `diskcache_seek_back_total`                                        | `path`                                                                                            | Seek back when Get() got any error                                                                                   |
| GAUGE   | `diskcache_capacity`                                               | `path`                                                                                            | Current capacity(in bytes)                                                                                           |
| GAUGE   | `diskcache_max_data`                                               | `path`                                                                                            | Max data to Put(in bytes), default 0                                                                                 |
| GAUGE   | `diskcache_batch_size`                                             | `path`                                                                                            | Data file size(in bytes)                                                                                             |
| GAUGE   | `diskcache_size`                                                   | `path`                                                                                            | Current cache size(in bytes)                                                                                         |
| GAUGE   | `diskcache_open_time`                                              | `no_fallback_on_error,no_lock,no_pos,no_sync,path`                                                | Current cache Open time in unix timestamp(second)                                                                    |
| GAUGE   | `diskcache_last_close_time`                                        | `path`                                                                                            | Current cache last Close time in unix timestamp(second)                                                              |
| GAUGE   | `diskcache_datafiles`                                              | `path`                                                                                            | Current un-read data files                                                                                           |
| SUMMARY | `diskcache_stream_put`                                             | `path`                                                                                            | Stream put times                                                                                                     |
| SUMMARY | `diskcache_get_latency`                                            | `path`                                                                                            | Get() cost seconds                                                                                                   |
| SUMMARY | `diskcache_put_latency`                                            | `path`                                                                                            | Put() cost seconds                                                                                                   |
| SUMMARY | `diskcache_put_bytes`                                              | `path`                                                                                            | Cache Put() bytes                                                                                                    |
| SUMMARY | `diskcache_get_bytes`                                              | `path`                                                                                            | Cache Get() bytes                                                                                                    |
| COUNTER | `pointpool_chan_get_total`                                         | `N/A`                                                                                             | Get count from reserved channel                                                                                      |
| COUNTER | `pointpool_chan_put_total`                                         | `N/A`                                                                                             | Put count to reserved channel                                                                                        |
| COUNTER | `pointpool_pool_get_total`                                         | `N/A`                                                                                             | Get count from reserved channel                                                                                      |
| COUNTER | `pointpool_pool_put_total`                                         | `N/A`                                                                                             | Put count to reserved channel                                                                                        |
| GAUGE   | `pointpool_reserved_capacity`                                      | `N/A`                                                                                             | Reserved capacity of the pool                                                                                        |
| COUNTER | `pointpool_malloc_total`                                           | `N/A`                                                                                             | New object malloc from pool                                                                                          |
| COUNTER | `pointpool_escaped_total`                                          | `N/A`                                                                                             | Points that not comes from pool                                                                                      |

### Golang Runtime Metrics {#go-runtime-metrics}

On API `/metrics`, DataKit exported Golang runtime metrics, it seems like this:

```not-set
go_cgo_go_to_c_calls_calls_total 8447
go_gc_cycles_automatic_gc_cycles_total 10
go_gc_cycles_forced_gc_cycles_total 0
go_gc_cycles_total_gc_cycles_total 10
go_gc_duration_seconds{quantile="0"} 3.4709e-05
go_gc_duration_seconds{quantile="0.25"} 3.9917e-05
go_gc_duration_seconds{quantile="0.5"} 0.000138459
go_gc_duration_seconds{quantile="0.75"} 0.000211333
go_gc_duration_seconds{quantile="1"} 0.000693833
go_gc_duration_seconds_sum 0.001920708
go_gc_duration_seconds_count 10
go_gc_heap_allocs_by_size_bytes_bucket{le="8.999999999999998"} 16889
go_gc_heap_allocs_by_size_bytes_bucket{le="24.999999999999996"} 221293
go_gc_heap_allocs_by_size_bytes_bucket{le="64.99999999999999"} 365672
go_gc_heap_allocs_by_size_bytes_bucket{le="144.99999999999997"} 475633
go_gc_heap_allocs_by_size_bytes_bucket{le="320.99999999999994"} 507361
go_gc_heap_allocs_by_size_bytes_bucket{le="704.9999999999999"} 516511
go_gc_heap_allocs_by_size_bytes_bucket{le="1536.9999999999998"} 521176
go_gc_heap_allocs_by_size_bytes_bucket{le="3200.9999999999995"} 522802
go_gc_heap_allocs_by_size_bytes_bucket{le="6528.999999999999"} 524529
go_gc_heap_allocs_by_size_bytes_bucket{le="13568.999999999998"} 525164
go_gc_heap_allocs_by_size_bytes_bucket{le="27264.999999999996"} 525269
go_gc_heap_allocs_by_size_bytes_bucket{le="+Inf"} 525421
go_gc_heap_allocs_by_size_bytes_sum 7.2408264e+07
go_gc_heap_allocs_by_size_bytes_count 525421
go_gc_heap_allocs_bytes_total 7.2408264e+07
go_gc_heap_allocs_objects_total 525421
go_gc_heap_frees_by_size_bytes_bucket{le="8.999999999999998"} 11081
go_gc_heap_frees_by_size_bytes_bucket{le="24.999999999999996"} 168291
go_gc_heap_frees_by_size_bytes_bucket{le="64.99999999999999"} 271749
go_gc_heap_frees_by_size_bytes_bucket{le="144.99999999999997"} 352424
go_gc_heap_frees_by_size_bytes_bucket{le="320.99999999999994"} 378481
go_gc_heap_frees_by_size_bytes_bucket{le="704.9999999999999"} 385700
go_gc_heap_frees_by_size_bytes_bucket{le="1536.9999999999998"} 389443
go_gc_heap_frees_by_size_bytes_bucket{le="3200.9999999999995"} 390591
go_gc_heap_frees_by_size_bytes_bucket{le="6528.999999999999"} 392069
go_gc_heap_frees_by_size_bytes_bucket{le="13568.999999999998"} 392565
go_gc_heap_frees_by_size_bytes_bucket{le="27264.999999999996"} 392636
go_gc_heap_frees_by_size_bytes_bucket{le="+Inf"} 392747
go_gc_heap_frees_by_size_bytes_sum 5.3304296e+07
go_gc_heap_frees_by_size_bytes_count 392747
go_gc_heap_frees_bytes_total 5.3304296e+07
go_gc_heap_frees_objects_total 392747
go_gc_heap_goal_bytes 3.6016864e+07
go_gc_heap_objects_objects 132674
go_gc_heap_tiny_allocs_objects_total 36033
go_gc_limiter_last_enabled_gc_cycle 0
go_gc_pauses_seconds_bucket{le="9.999999999999999e-10"} 0
go_gc_pauses_seconds_bucket{le="9.999999999999999e-09"} 0
go_gc_pauses_seconds_bucket{le="9.999999999999998e-08"} 0
go_gc_pauses_seconds_bucket{le="1.0239999999999999e-06"} 0
go_gc_pauses_seconds_bucket{le="1.0239999999999999e-05"} 1
go_gc_pauses_seconds_bucket{le="0.00010239999999999998"} 15
go_gc_pauses_seconds_bucket{le="0.0010485759999999998"} 20
go_gc_pauses_seconds_bucket{le="0.010485759999999998"} 20
go_gc_pauses_seconds_bucket{le="0.10485759999999998"} 20
go_gc_pauses_seconds_bucket{le="+Inf"} 20
go_gc_pauses_seconds_sum 0.000656384
go_gc_pauses_seconds_count 20
go_gc_stack_starting_size_bytes 4096
go_goroutines 102
go_info{version="go1.19.5"} 1
go_memory_classes_heap_free_bytes 8.839168e+06
go_memory_classes_heap_objects_bytes 1.9103968e+07
go_memory_classes_heap_released_bytes 3.530752e+06
go_memory_classes_heap_stacks_bytes 2.4576e+06
go_memory_classes_heap_unused_bytes 8.011552e+06
go_memory_classes_metadata_mcache_free_bytes 3600
go_memory_classes_metadata_mcache_inuse_bytes 12000
go_memory_classes_metadata_mspan_free_bytes 77472
go_memory_classes_metadata_mspan_inuse_bytes 426960
go_memory_classes_metadata_other_bytes 6.201928e+06
go_memory_classes_os_stacks_bytes 0
go_memory_classes_other_bytes 1.931459e+06
go_memory_classes_profiling_buckets_bytes 1.489565e+06
go_memory_classes_total_bytes 5.2086024e+07
go_memstats_alloc_bytes 1.9103968e+07
go_memstats_alloc_bytes_total 7.2408264e+07
go_memstats_buck_hash_sys_bytes 1.489565e+06
go_memstats_frees_total 428780
go_memstats_gc_sys_bytes 6.201928e+06
go_memstats_heap_alloc_bytes 1.9103968e+07
go_memstats_heap_idle_bytes 1.236992e+07
go_memstats_heap_inuse_bytes 2.711552e+07
go_memstats_heap_objects 132674
go_memstats_heap_released_bytes 3.530752e+06
go_memstats_heap_sys_bytes 3.948544e+07
go_memstats_last_gc_time_seconds 1.6992580814748092e+09
go_memstats_lookups_total 0
go_memstats_mallocs_total 561454
go_memstats_mcache_inuse_bytes 12000
go_memstats_mcache_sys_bytes 15600
go_memstats_mspan_inuse_bytes 426960
go_memstats_mspan_sys_bytes 504432
go_memstats_next_gc_bytes 3.6016864e+07
go_memstats_other_sys_bytes 1.931459e+06
go_memstats_stack_inuse_bytes 2.4576e+06
go_memstats_stack_sys_bytes 2.4576e+06
go_memstats_sys_bytes 5.2086024e+07
go_sched_gomaxprocs_threads 10
go_sched_goroutines_goroutines 102
go_sched_latencies_seconds_bucket{le="9.999999999999999e-10"} 4886
go_sched_latencies_seconds_bucket{le="9.999999999999999e-09"} 4886
go_sched_latencies_seconds_bucket{le="9.999999999999998e-08"} 5883
go_sched_latencies_seconds_bucket{le="1.0239999999999999e-06"} 6669
go_sched_latencies_seconds_bucket{le="1.0239999999999999e-05"} 7191
go_sched_latencies_seconds_bucket{le="0.00010239999999999998"} 7531
go_sched_latencies_seconds_bucket{le="0.0010485759999999998"} 7567
go_sched_latencies_seconds_bucket{le="0.010485759999999998"} 7569
go_sched_latencies_seconds_bucket{le="0.10485759999999998"} 7569
go_sched_latencies_seconds_bucket{le="+Inf"} 7569
go_sched_latencies_seconds_sum 0.00988825
go_sched_latencies_seconds_count 7569
go_threads 16
```
