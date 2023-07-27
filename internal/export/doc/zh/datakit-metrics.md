# Datakit 自身指标

---

为便于 Datakit 自身可观测性，我们在开发 Datakit 过程中，给相关的业务模块增加了很多 Prometheus 指标暴露，通过暴露这些指标，我们能方便的排查 Datakit 运行过程中的一些问题。

## 指标列表 {#metrics}

自 Datakit [1.5.9 版本](changelog.md#cl-1.5.9)以来，通过访问 `http://localhost:9529/metrics` 即可获取当前的指标列表，不同 Datakit 版本可能会对一些相关指标做调整，或者增删一些指标。

这些指标，在 [Datakit monitor](datakit-monitor.md) 展示中也会用到，只是 monitor 中为了展示上的友好型，做了一些优化处理。如果要查看原始的指标（或者 monitor 上没有展示出来的指标），我们可以通过 `curl` 和 `watch` 命令的组合来查看，比如获取 Datakit 进程 CPU 的使用情况：

```shell
# 每隔 3s 获取一次 CPU 使用率指标
$ watch -n 3 'curl -s http://localhost:9529/metrics | grep datakit_cpu_usage'

# HELP datakit_cpu_usage Datakit CPU usage(%)
# TYPE datakit_cpu_usage gauge
datakit_cpu_usage 4.9920266849857144
```

其它指标也能通过类似方式来观察，目前已有的指标如下（当前版本 {{ .Version }}）：

<!-- 以下这些指标，通过执行 make show_metrics 方式能获取 -->

``` not-set
NAME                                              TYPE                HELP
datakit_cpu_cores                                 GAUGE               Datakit CPU cores
datakit_cpu_usage                                 GAUGE               Datakit CPU usage(%)
datakit_data_overuse                              GAUGE               Does current workspace's data(metric/logging) usage(if 0 not beyond, or with a unix timestamp when overuse occurred)
datakit_dialtesting_dataway_send_failed_number    GAUGE               The number of failed sending for each dataway
datakit_dialtesting_pull_cost_seconds             SUMMARY             Time cost to pull tasks
datakit_dialtesting_task_check_cost_seconds       SUMMARY             Task check time
datakit_dialtesting_task_invalid_total            COUNTER             Invalid task number
datakit_dialtesting_task_number                   GAUGE               The number of tasks
datakit_dialtesting_task_run_cost_seconds         SUMMARY             Task run time
datakit_dialtesting_task_synchronized_total       COUNTER             Task synchronized number
datakit_dialtesting_worker_cached_points_number   GAUGE               The number of cached points
datakit_dialtesting_worker_job_chan_number        GAUGE               The number of the chan for the jobs
datakit_dialtesting_worker_send_points_number     GAUGE               The number of the points which have been sent
datakit_dns_cost_seconds                          SUMMARY             DNS IP lookup cost
datakit_dns_domain_total                          COUNTER             DNS watched domain counter
datakit_dns_ip_updated_total                      COUNTER             Domain IP updated counter
datakit_dns_watch_run_total                       COUNTER             Watch run counter
datakit_election_inputs                           GAUGE               Datakit election input count
datakit_election_pause_total                      COUNTER             Input paused count when election failed
datakit_election_resume_total                     COUNTER             Input resume count when election OK
datakit_election_seconds                          SUMMARY             Election latency
datakit_election_status                           GAUGE               Datakit election status, if metric = 0, meas not elected, or the elected time(unix timestamp second)
datakit_error_total                               COUNTER             Total errors, only count on error source, not include error message
datakit_filter_last_update_timestamp_seconds      GAUGE               Filter last update time
datakit_filter_latency_seconds                    SUMMARY             Filter latency of these filters
datakit_filter_parse_error                        GAUGE               Filter parse error
datakit_filter_point_dropped_total                COUNTER             Dropped points of filters
datakit_filter_point_total                        COUNTER             Filter points of filters
datakit_filter_pull_latency_seconds               SUMMARY             Filter pull(remote) latency
datakit_filter_update_total                       COUNTER             Filters(remote) updated count
datakit_goroutine_alive                           GAUGE               Alive Goroutines
datakit_goroutine_cost_seconds                    SUMMARY             Goroutine running duration
datakit_goroutine_groups                          GAUGE               Goroutine group count
datakit_goroutine_stopped_total                   COUNTER             Stopped Goroutines
datakit_goroutines                                GAUGE               Goroutine count within Datakit
datakit_heap_alloc_bytes                          GAUGE               Datakit memory heap bytes
datakit_http_api_elapsed_seconds                  SUMMARY             API request cost
datakit_http_api_req_size_bytes                   SUMMARY             API request body size
datakit_http_api_total                            COUNTER             API request counter
datakit_httpcli_conn_idle_time_seconds            SUMMARY             HTTP connection idle time
datakit_httpcli_conn_reused_from_idle_total       COUNTER             HTTP connection reused from idle count
datakit_httpcli_dns_cost_seconds                  SUMMARY             HTTP DNS cost
datakit_httpcli_got_first_resp_byte_cost_seconds  SUMMARY             Got first response byte cost
datakit_httpcli_http_connect_cost_seconds         SUMMARY             HTTP connect cost
datakit_httpcli_tcp_conn_total                    COUNTER             HTTP TCP connection count
datakit_httpcli_tls_handshake_seconds             SUMMARY             HTTP TLS handshake cost
datakit_input_collect_latency_seconds             SUMMARY             Input collect latency
datakit_inputs_crash_total                        COUNTER             Input crash count
datakit_inputs_instance                           GAUGE               Input instance count
datakit_io_chan_capacity                          GAUGE               IO channel capacity
datakit_io_chan_usage                             GAUGE               IO channel usage(length of the channel)
datakit_io_dataway_api_latency_seconds            SUMMARY             Dataway HTTP request latency partitioned by HTTP API(method@url) and HTTP status
datakit_io_dataway_not_sink_point_total           COUNTER             Dataway not-Sinked points(condition or category not match)
datakit_io_dataway_point_bytes_total              COUNTER             Dataway uploaded points bytes, partitioned by category and pint send status(HTTP status)
datakit_io_dataway_point_total                    COUNTER             Dataway uploaded points, partitioned by category and send status(HTTP status)
datakit_io_dataway_sink_point_total               COUNTER             Dataway Sinked points, partitioned by category and point send status(ok/failed/dropped)
datakit_io_dataway_sink_total                     COUNTER             Dataway Sinked count, partitioned by category.
datakit_io_feed_cost_seconds                      SUMMARY             IO feed waiting(on block mode) seconds
datakit_io_feed_drop_point_total                  COUNTER             IO feed drop(on non-block mode) points
datakit_io_feed_point_total                       COUNTER             Input feed point total
datakit_io_feed_total                             COUNTER             Input feed total
datakit_io_flush_failcache_bytes                  SUMMARY             IO flush fail-cache bytes(in gzip) summary
datakit_io_flush_total                            COUNTER             IO flush total
datakit_io_flush_workers                          GAUGE               IO flush workers
datakit_io_http_retry_total                       COUNTER             Dataway HTTP retried count
datakit_io_input_filter_point_total               COUNTER             Input filtered point total
datakit_io_last_feed_timestamp_seconds            GAUGE               Input last feed time(according to Datakit local time)
datakit_io_queue_points                           GAUGE               IO module queued(cached) points
datakit_kafkamq_consumer_message_total            COUNTER             Kafka consumer message numbers from Datakit start
datakit_kafkamq_group_election_total              COUNTER             Kafka group election count
datakit_last_err                                  GAUGE               Datakit errors(when error occurred), these errors come from inputs or any sub modules
datakit_open_files                                GAUGE               Datakit open files(only available on Linux)
datakit_pipeline_cost_seconds                     SUMMARY             Pipeline total running time
datakit_pipeline_drop_point_total                 COUNTER             Pipeline total dropped points
datakit_pipeline_error_point_total                COUNTER             Pipeline processed total error points
datakit_pipeline_last_update_timestamp_seconds    GAUGE               Pipeline last update time
datakit_pipeline_point_total                      COUNTER             Pipeline processed total points
datakit_process_ctx_switch_total                  COUNTER             Datakit process context switch count(Linux only)
datakit_process_io_bytes_total                    COUNTER             Datakit process IO bytes count
datakit_process_io_count_total                    COUNTER             Datakit process IO count
datakit_prom_collect_points                       SUMMARY             Total number of prom collection points
datakit_prom_http_get_bytes                       SUMMARY             HTTP get bytes
datakit_prom_http_latency_in_second               SUMMARY             HTTP latency(in second)
datakit_rum_loaded_zip_cnt                        GAUGE               RUM source map currently loaded zip archive count
datakit_rum_locate_statistics_total               COUNTER             locate by ip addr statistics
datakit_rum_source_map_duration_seconds           SUMMARY             statistics elapsed time in RUM source map(unit: second)
datakit_rum_source_map_total                      COUNTER             source map result statistics
datakit_sys_alloc_bytes                           GAUGE               Datakit memory system bytes
datakit_uptime_seconds                            GAUGE               Datakit uptime
```
