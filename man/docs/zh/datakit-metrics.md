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

其它指标也能通过类似方式来观察，目前已有的指标如下：

<!-- 以下这些指标，通过执行 make show_metrics 方式能获取，另外稍加编辑，即可将其排版成 markdown 格式-->

| TYPE      | NAME                                          | HELP
| ---       | ---                                           | ---
| SUMMARY   | `datakit_http_api_elapsed`                    | API request cost(in ms)
| HISTOGRAM | `datakit_http_api_elapsed_histogram`          | API request cost(in ms) histogram
| SUMMARY   | `datakit_http_api_req_size`                   | API request body size
| COUNTER   | `datakit_http_api_total`                      | API request counter
| COUNTER   | `datakit_dns_domain_total`                    | DNS watched domain counter
| COUNTER   | `datakit_dns_ip_updated_total`                | Domain IP updated counter
| COUNTER   | `datakit_dns_watch_run_total`                 | Watch run counter
| SUMMARY   | `datakit_dns_cost`                            | DNS IP lookup cost(ms)
| COUNTER   | `datakit_election_pause_total`                | Input paused count when election failed
| COUNTER   | `datakit_election_resume_total`               | Input resume count when election OK
| GAUGE     | `datakit_election_status`                     | Datakit election status, if metric = 0, meas not elected, or the elected time(unix timestamp second)
| GAUGE     | `datakit_election_inputs`                     | Datakit election input count
| SUMMARY   | `datakit_election`                            | Election latency(in millisecond)
| GAUGE     | `datakit_goroutine_alive`                     | Alive Goroutines
| COUNTER   | `datakit_goroutine_stopped_total`             | Stopped Goroutines
| GAUGE     | `datakit_goroutine_groups`                    | Goroutine group count
| SUMMARY   | `datakit_goroutine_cost`                      | Goroutine running time(in nanosecond)
| GAUGE     | `datakit_goroutines`                          | Goroutine count within Datakit
| GAUGE     | `datakit_heap_alloc`                          | Datakit memory heap bytes
| GAUGE     | `datakit_sys_alloc`                           | Datakit memory system bytes
| GAUGE     | `datakit_cpu_usage`                           | Datakit CPU usage(%)
| GAUGE     | `datakit_open_files`                          | Datakit open files(only available on Linux)
| GAUGE     | `datakit_cpu_cores`                           | Datakit CPU cores
| GAUGE     | `datakit_uptime`                              | Datakit uptime(second)
| GAUGE     | `datakit_data_overuse`                        | Does current workspace's data(metric/logging) usage(if 0 not beyond, or with a unix timestamp when overuse occurred)
| COUNTER   | `datakit_process_ctx_switch_total`            | Datakit process context switch count(Linux only)
| COUNTER   | `datakit_process_ctx_switch_total`            | Datakit process context switch count(Linux only)
| COUNTER   | `datakit_process_io_count_total`              | Datakit process IO count
| COUNTER   | `datakit_process_io_count_total`              | Datakit process IO count
| COUNTER   | `datakit_process_io_bytes_total`              | Datakit process IO bytes count
| COUNTER   | `datakit_process_io_bytes_total`              | Datakit process IO bytes count
| COUNTER   | `datakit_io_http_tcp_conn_total`              | Dataway HTTP TCP connection count
| COUNTER   | `datakit_io_http_conn_reused_from_idle_total` | Dataway HTTP connection reused from idle count
| SUMMARY   | `datakit_io_http_conn_idle_time`              | Dataway HTTP connection idle time(ms)
| SUMMARY   | `datakit_io_http_dns_cost`                    | Dataway HTTP DNS cost(ms)
| SUMMARY   | `datakit_io_http_tls_handshake`               | Dataway TLS handshake cost(ms)
| SUMMARY   | `datakit_io_http_connect_cost`                | Dataway HTTP connect cost(ms)
| SUMMARY   | `datakit_io_http_got_first_resp_byte_cost`    | Dataway got first response byte cost(ms)
| SUMMARY   | `datakit_io_flush_failcache_bytes`            | IO flush fail-cache bytes(in gzip) summary
| COUNTER   | `datakit_io_dataway_api_request_total`        | Dataway HTTP request processed, partitioned by status code and HTTP API(url path)
| COUNTER   | `datakit_io_dataway_point_total`              | Dataway uploaded points, partitioned by category and send status(HTTP status)
| COUNTER   | `datakit_io_dataway_point_bytes_total`        | Dataway uploaded points bytes, partitioned by category and pint send status(HTTP status)
| SUMMARY   | `datakit_io_dataway_api_latency`              | Dataway HTTP request latency(ms) partitioned by HTTP API(method@url) and HTTP status
| COUNTER   | `datakit_io_http_retry_total`                 | Dataway HTTP retried count
| COUNTER   | `datakit_io_dataway_sink_total`               | Dataway Sink count, partitioned by category.
| COUNTER   | `datakit_io_dataway_not_sink_point_total`     | Dataway not-Sinked points(condition or category not match)
| COUNTER   | `datakit_io_dataway_sink_point_total`         | Dataway Sinked points, partitioned by category and point send status(ok/failed/dropped)
| GAUGE     | `datakit_filter_last_update`                  | Filter last update time(in unix timestamp second)
| COUNTER   | `datakit_filter_point_total`                  | Filter points of filters
| COUNTER   | `datakit_filter_point_dropped_total`          | Dropped points of filters
| SUMMARY   | `datakit_filter_pull_latency`                 | Filter pull(remote) latency(ms)
| SUMMARY   | `datakit_filter_latency`                      | Filter latency(us) of these filters
| COUNTER   | `datakit_filter_update_total`                 | Filters(remote) updated count
| COUNTER   | `datakit_io_input_filter_point_total`         | Input filtered point total
| COUNTER   | `datakit_io_feed_total`                       | Input feed total
| GAUGE     | `datakit_io_last_feed`                        | Input last feed time(unix timestamp in second)
| SUMMARY   | `datakit_input_collect_latency`               | Input collect latency(us)
| GAUGE     | `datakit_io_chan_usage`                       | IO channel usage(length of the channel)
| GAUGE     | `datakit_io_chan_capacity`                    | IO channel capacity
| GAUGE     | `datakit_io_flush_workers`                    | IO flush workers
| COUNTER   | `datakit_io_flush_total`                      | IO flush total
| GAUGE     | `datakit_io_queue_pts`                        | IO module queued(cached) points
| GAUGE     | `datakit_last_err`                            | Datakit errors(when error occurred), these errors come from inputs or any sub modules
| COUNTER   | `datakit_error_total`                         | Total errors, only count on error source, not include error message
| COUNTER   | `datakit_io_feed_point_total`                 | Input feed point total
| COUNTER   | `datakit_pipeline_point_total`                | Pipeline processed total points
| COUNTER   | `datakit_pipeline_drop_point_total`           | Pipeline total dropped points
| COUNTER   | `datakit_pipeline_error_point_total`          | Pipeline processed total error points
| SUMMARY   | `datakit_pipeline_cost`                       | Pipeline total running time(ms)
| GAUGE     | `datakit_pipeline_update_time`                | Pipeline last update time(unix timestamp)
| GAUGE     | `datakit_dialtesting_task_number`             | the number of tasks
| SUMMARY   | `datakit_dialtesting_pull_cost`               | time cost to pull tasks(in nanosecond)
| COUNTER   | `datakit_dialtesting_task_synchronized_total` | task synchronized number
| COUNTER   | `datakit_dialtesting_task_invalid_total`      | invalid task number
| SUMMARY   | `datakit_dialtesting_task_check_cost`         | task check time(in nanosecond)
| SUMMARY   | `datakit_dialtesting_task_run_cost`           | task run time(in nanosecond)
| COUNTER   | `datakit_kafkamq_consumer_message_total`      | Kafka consumer message numbers from Datakit start
| COUNTER   | `datakit_kafkamq_group_election_total`        | Kafka group election count
| GAUGE     | `datakit_inputs_instance`                     | Input instance count
| COUNTER   | `datakit_inputs_crash_total`                  | Input crash count
