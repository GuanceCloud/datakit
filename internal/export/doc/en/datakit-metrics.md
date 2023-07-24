# Datakit Exported Metrics

---

Foe better self-observability, Datakit has exported many Prometheus metrics among various submodules. We can use these metrics to trouble shooting errors or bugs during Datakit running.

## Exported Metrics {#metrics}

Since Datakit [Version 1.5.9](changelog.md#cl-1.5.9), we can visit `http://localhost:9529/metrics` to see these metrics(Different release may got some new or renamed metrics).

We have used these metrics in [Datakit monitor](datakit-monitor.md), and we can *playing* these metrics with some tricks like this(refresh CPU usage of Datakit): 

```shell
# play CPU usage of Datakit on every 3 seconds
$ watch -n 3 'curl -s http://localhost:9529/metrics | grep datakit_cpu_usage'

# HELP datakit_cpu_usage Datakit CPU usage(%)
# TYPE datakit_cpu_usage gauge
datakit_cpu_usage 4.9920266849857144
```

We can also playing other metrics too(change the `grep` string), all available metrics list below(current Datakit version is {{ .Version }}):

<!-- we can run `make show_metrics` go export all these metrics -->

``` not-set
TYPE                NAME                                               HELP
COUNTER             datakit_dns_domain_total                           DNS watched domain counter
COUNTER             datakit_dns_ip_updated_total                       Domain IP updated counter
COUNTER             datakit_dns_watch_run_total                        Watch run counter
SUMMARY             datakit_dns_cost_seconds                           DNS IP lookup cost
COUNTER             datakit_election_pause_total                       Input paused count when election failed
COUNTER             datakit_election_resume_total                      Input resume count when election OK
GAUGE               datakit_election_status                            Datakit election status, if metric = 0, meas not elected, or the elected time(unix timestamp second)
GAUGE               datakit_election_inputs                            Datakit election input count
SUMMARY             datakit_election_seconds                           Election latency
GAUGE               datakit_goroutine_alive                            Alive Goroutines
COUNTER             datakit_goroutine_stopped_total                    Stopped Goroutines
GAUGE               datakit_goroutine_groups                           Goroutine group count
SUMMARY             datakit_goroutine_cost_seconds                     Goroutine running duration
SUMMARY             datakit_http_api_elapsed_seconds                   API request cost
SUMMARY             datakit_http_api_req_size_bytes                    API request body size
COUNTER             datakit_http_api_total                             API request counter
COUNTER             datakit_httpcli_tcp_conn_total                     HTTP TCP connection count
COUNTER             datakit_httpcli_conn_reused_from_idle_total        HTTP connection reused from idle count
SUMMARY             datakit_httpcli_conn_idle_time_seconds             HTTP connection idle time
SUMMARY             datakit_httpcli_dns_cost_seconds                   HTTP DNS cost
SUMMARY             datakit_httpcli_tls_handshake_seconds              HTTP TLS handshake cost
SUMMARY             datakit_httpcli_http_connect_cost_seconds          HTTP connect cost
SUMMARY             datakit_httpcli_got_first_resp_byte_cost_seconds   Got first response byte cost
COUNTER             datakit_io_http_retry_total                        Dataway HTTP retried count
COUNTER             datakit_io_dataway_sink_total                      Dataway Sinked count, partitioned by category.
COUNTER             datakit_io_dataway_not_sink_point_total            Dataway not-Sinked points(condition or category not match)
COUNTER             datakit_io_dataway_sink_point_total                Dataway Sinked points, partitioned by category and point send status(ok/failed/dropped)
SUMMARY             datakit_io_flush_failcache_bytes                   IO flush fail-cache bytes(in gzip) summary
COUNTER             datakit_io_dataway_point_total                     Dataway uploaded points, partitioned by category and send status(HTTP status)
COUNTER             datakit_io_dataway_point_bytes_total               Dataway uploaded points bytes, partitioned by category and pint send status(HTTP status)
SUMMARY             datakit_io_dataway_api_latency_seconds             Dataway HTTP request latency partitioned by HTTP API(method@url) and HTTP status
GAUGE               datakit_filter_last_update_timestamp_seconds       Filter last update time
COUNTER             datakit_filter_point_total                         Filter points of filters
COUNTER             datakit_filter_point_dropped_total                 Dropped points of filters
SUMMARY             datakit_filter_pull_latency_seconds                Filter pull(remote) latency
SUMMARY             datakit_filter_latency_seconds                     Filter latency of these filters
COUNTER             datakit_filter_update_total                        Filters(remote) updated count
COUNTER             datakit_error_total                                Total errors, only count on error source, not include error message
COUNTER             datakit_io_feed_point_total                        Input feed point total
COUNTER             datakit_io_input_filter_point_total                Input filtered point total
COUNTER             datakit_io_feed_total                              Input feed total
GAUGE               datakit_io_last_feed_timestamp_seconds             Input last feed time(according to Datakit local time)
SUMMARY             datakit_input_collect_latency_seconds              Input collect latency
GAUGE               datakit_io_chan_usage                              IO channel usage(length of the channel)
GAUGE               datakit_io_chan_capacity                           IO channel capacity
SUMMARY             datakit_io_feed_cost_seconds                       IO feed waiting(on block mode) seconds
COUNTER             datakit_io_feed_drop_point_total                   IO feed drop(on non-block mode) points
GAUGE               datakit_io_flush_workers                           IO flush workers
COUNTER             datakit_io_flush_total                             IO flush total
GAUGE               datakit_io_queue_points                            IO module queued(cached) points
GAUGE               datakit_last_err                                   Datakit errors(when error occurred), these errors come from inputs or any sub modules
GAUGE               datakit_goroutines                                 Goroutine count within Datakit
GAUGE               datakit_heap_alloc_bytes                           Datakit memory heap bytes
GAUGE               datakit_sys_alloc_bytes                            Datakit memory system bytes
GAUGE               datakit_cpu_usage                                  Datakit CPU usage(%)
GAUGE               datakit_open_files                                 Datakit open files(only available on Linux)
GAUGE               datakit_cpu_cores                                  Datakit CPU cores
GAUGE               datakit_uptime_seconds                             Datakit uptime
GAUGE               datakit_data_overuse                               Does current workspace's data(metric/logging) usage(if 0 not beyond, or with a unix timestamp when overuse occurred)
COUNTER             datakit_process_ctx_switch_total                   Datakit process context switch count(Linux only)
COUNTER             datakit_process_ctx_switch_total                   Datakit process context switch count(Linux only)
COUNTER             datakit_process_io_count_total                     Datakit process IO count
COUNTER             datakit_process_io_count_total                     Datakit process IO count
COUNTER             datakit_process_io_bytes_total                     Datakit process IO bytes count
COUNTER             datakit_process_io_bytes_total                     Datakit process IO bytes count
COUNTER             datakit_pipeline_point_total                       Pipeline processed total points
COUNTER             datakit_pipeline_drop_point_total                  Pipeline total dropped points
COUNTER             datakit_pipeline_error_point_total                 Pipeline processed total error points
SUMMARY             datakit_pipeline_cost_seconds                      Pipeline total running time
GAUGE               datakit_pipeline_last_update_timestamp_seconds     Pipeline last update time
GAUGE               datakit_dialtesting_task_number                    The number of tasks
SUMMARY             datakit_dialtesting_pull_cost_seconds              Time cost to pull tasks
COUNTER             datakit_dialtesting_task_synchronized_total        Task synchronized number
COUNTER             datakit_dialtesting_task_invalid_total             Invalid task number
SUMMARY             datakit_dialtesting_task_check_cost_seconds        Task check time
SUMMARY             datakit_dialtesting_task_run_cost_seconds          Task run time
COUNTER             datakit_kafkamq_consumer_message_total             Kafka consumer message numbers from Datakit start
COUNTER             datakit_kafkamq_group_election_total               Kafka group election count
GAUGE               datakit_inputs_instance                            Input instance count
COUNTER             datakit_inputs_crash_total                         Input crash count
SUMMARY             datakit_prom_collect_points                        Total number of prom collection points
SUMMARY             datakit_prom_http_get_bytes                        HTTP get bytes
SUMMARY             datakit_prom_http_latency_in_second                HTTP latency(in second)
```
