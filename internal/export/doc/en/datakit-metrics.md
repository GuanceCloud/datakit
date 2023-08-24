# Datakit Exported Metrics

---

Foe better self-observability, Datakit has exported many Prometheus metrics among various submodules. We can use these metrics to trouble shooting errors or bugs during Datakit running.

## Exported Metrics {#metrics}

Since Datakit [Version 1.5.9](changelog.md#cl-1.5.9), we can visit `http://localhost:9529/metrics` to see these metrics(Different release may got some new or renamed metrics).

We have used these metrics in [Datakit monitor](datakit-monitor.md), and we can *playing* these metrics with some tricks like this(refresh CPU usage of Datakit): 

```shell
# play CPU usage of Datakit on every 3 seconds
$ watch -n 3 'curl -s http://localhost:9529/metrics | grep -a datakit_cpu_usage'

# HELP datakit_cpu_usage Datakit CPU usage(%)
# TYPE datakit_cpu_usage gauge
datakit_cpu_usage 4.9920266849857144
```

We can also playing other metrics too(change the `grep` string), all available metrics list below(current Datakit version is {{ .Version }}):

<!-- we can run `make show_metrics` go export all these metrics -->

|POSITION|TYPE|NAME|LABELS|HELP|
|---|---|---|---|---|
|*internal/dnswatcher*|COUNTER|`datakit_dns_domain_total`|`N/A`|DNS watched domain counter|
|*internal/dnswatcher*|COUNTER|`datakit_dns_ip_updated_total`|`domain`|Domain IP updated counter|
|*internal/dnswatcher*|COUNTER|`datakit_dns_watch_run_total`|`interval`|Watch run counter|
|*internal/dnswatcher*|SUMMARY|`datakit_dns_cost_seconds`|`domain,status`|DNS IP lookup cost|
|*internal/election*|COUNTER|`datakit_election_pause_total`|`id,namespace`|Input paused count when election failed|
|*internal/election*|COUNTER|`datakit_election_resume_total`|`id,namespace`|Input resume count when election OK|
|*internal/election*|GAUGE|`datakit_election_status`|`elected_id,id,namespace,status`|Datakit election status, if metric = 0, meas not elected, or the elected time(unix timestamp second)|
|*internal/election*|GAUGE|`datakit_election_inputs`|`namespace`|Datakit election input count|
|*internal/election*|SUMMARY|`datakit_election_seconds`|`namespace,status`|Election latency|
|*internal/goroutine*|GAUGE|`datakit_goroutine_alive`|`name`|Alive Goroutine count|
|*internal/goroutine*|COUNTER|`datakit_goroutine_stopped_total`|`name`|Stopped Goroutine count|
|*internal/goroutine*|GAUGE|`datakit_goroutine_groups`|`N/A`|Goroutine group count|
|*internal/goroutine*|SUMMARY|`datakit_goroutine_cost_seconds`|`name`|Goroutine running duration|
|*internal/httpapi*|SUMMARY|`datakit_http_api_elapsed_seconds`|`api,method,status`|API request cost|
|*internal/httpapi*|SUMMARY|`datakit_http_api_req_size_bytes`|`api,method,status`|API request body size|
|*internal/httpapi*|COUNTER|`datakit_http_api_total`|`api,method,status`|API request counter|
|*internal/httpcli*|COUNTER|`datakit_httpcli_tcp_conn_total`|`from,remote,type`|HTTP TCP connection count|
|*internal/httpcli*|COUNTER|`datakit_httpcli_conn_reused_from_idle_total`|`from`|HTTP connection reused from idle count|
|*internal/httpcli*|SUMMARY|`datakit_httpcli_conn_idle_time_seconds`|`from`|HTTP connection idle time|
|*internal/httpcli*|SUMMARY|`datakit_httpcli_dns_cost_seconds`|`from`|HTTP DNS cost|
|*internal/httpcli*|SUMMARY|`datakit_httpcli_tls_handshake_seconds`|`from`|HTTP TLS handshake cost|
|*internal/httpcli*|SUMMARY|`datakit_httpcli_http_connect_cost_seconds`|`from`|HTTP connect cost|
|*internal/httpcli*|SUMMARY|`datakit_httpcli_got_first_resp_byte_cost_seconds`|`from`|Got first response byte cost|
|*internal/io/dataway*|COUNTER|`datakit_io_http_retry_total`|`api,status`|Dataway HTTP retried count|
|*internal/io/dataway*|COUNTER|`datakit_io_dataway_sink_total`|`category`|Dataway Sinked count, partitioned by category.|
|*internal/io/dataway*|COUNTER|`datakit_io_dataway_not_sink_point_total`|`category`|Dataway not-Sinked points(condition or category not match)|
|*internal/io/dataway*|COUNTER|`datakit_io_dataway_sink_point_total`|`category,status`|Dataway Sinked points, partitioned by category and point send status(ok/failed/dropped)|
|*internal/io/dataway*|SUMMARY|`datakit_io_flush_failcache_bytes`|`category`|IO flush fail-cache bytes(in gzip) summary|
|*internal/io/dataway*|COUNTER|`datakit_io_dataway_point_total`|`category,status`|Dataway uploaded points, partitioned by category and send status(HTTP status)|
|*internal/io/dataway*|COUNTER|`datakit_io_dataway_point_bytes_total`|`category,enc,status`|Dataway uploaded points bytes, partitioned by category and pint send status(HTTP status)|
|*internal/io/dataway*|SUMMARY|`datakit_io_dataway_api_latency_seconds`|`api,status`|Dataway HTTP request latency partitioned by HTTP API(method@url) and HTTP status|
|*internal/io/filter*|COUNTER|`datakit_filter_update_total`|`N/A`|Filters(remote) updated count|
|*internal/io/filter*|GAUGE|`datakit_filter_last_update_timestamp_seconds`|`N/A`|Filter last update time|
|*internal/io/filter*|COUNTER|`datakit_filter_point_total`|`category,filters,source`|Filter points of filters|
|*internal/io/filter*|GAUGE|`datakit_filter_parse_error`|`error,filters`|Filter parse error|
|*internal/io/filter*|COUNTER|`datakit_filter_point_dropped_total`|`category,filters,source`|Dropped points of filters|
|*internal/io/filter*|SUMMARY|`datakit_filter_pull_latency_seconds`|`status`|Filter pull(remote) latency|
|*internal/io/filter*|SUMMARY|`datakit_filter_latency_seconds`|`category,filters,source`|Filter latency of these filters|
|*internal/io*|COUNTER|`datakit_error_total`|`source,category`|Total errors, only count on error source, not include error message|
|*internal/io*|COUNTER|`datakit_io_feed_point_total`|`name,category`|Input feed point total|
|*internal/io*|COUNTER|`datakit_io_input_filter_point_total`|`name,category`|Input filtered point total|
|*internal/io*|COUNTER|`datakit_io_feed_total`|`name,category`|Input feed total|
|*internal/io*|GAUGE|`datakit_io_last_feed_timestamp_seconds`|`name,category`|Input last feed time(according to Datakit local time)|
|*internal/io*|SUMMARY|`datakit_input_collect_latency_seconds`|`name,category`|Input collect latency|
|*internal/io*|GAUGE|`datakit_io_chan_usage`|`category`|IO channel usage(length of the channel)|
|*internal/io*|GAUGE|`datakit_io_chan_capacity`|`category`|IO channel capacity|
|*internal/io*|SUMMARY|`datakit_io_feed_cost_seconds`|`N/A`|IO feed waiting(on block mode) seconds|
|*internal/io*|COUNTER|`datakit_io_feed_drop_point_total`|`N/A`|IO feed drop(on non-block mode) points|
|*internal/io*|GAUGE|`datakit_io_flush_workers`|`category`|IO flush workers|
|*internal/io*|COUNTER|`datakit_io_flush_total`|`category`|IO flush total|
|*internal/io*|GAUGE|`datakit_io_queue_points`|`category`|IO module queued(cached) points|
|*internal/io*|GAUGE|`datakit_last_err`|`input,source,category,error`|Datakit errors(when error occurred), these errors come from inputs or any sub modules|
|*internal/metrics*|GAUGE|`datakit_goroutines`|`N/A`|Goroutine count within Datakit|
|*internal/metrics*|GAUGE|`datakit_heap_alloc_bytes`|`N/A`|Datakit memory heap bytes|
|*internal/metrics*|GAUGE|`datakit_sys_alloc_bytes`|`N/A`|Datakit memory system bytes|
|*internal/metrics*|GAUGE|`datakit_cpu_usage`|`N/A`|Datakit CPU usage(%)|
|*internal/metrics*|GAUGE|`datakit_open_files`|`N/A`|Datakit open files(only available on Linux)|
|*internal/metrics*|GAUGE|`datakit_cpu_cores`|`N/A`|Datakit CPU cores|
|*internal/metrics*|GAUGE|`datakit_uptime_seconds`|`hostname,cgroup,branch=?,os_arch=?,docker=?,auto_update=?,version=?,build_at=?`|Datakit uptime|
|*internal/metrics*|GAUGE|`datakit_data_overuse`|`N/A`|Does current workspace's data(metric/logging) usage(if 0 not beyond, or with a unix timestamp when overuse occurred)|
|*internal/metrics*|COUNTER|`datakit_process_ctx_switch_total`|`type`|Datakit process context switch count(Linux only)|
|*internal/metrics*|COUNTER|`datakit_process_io_count_total`|`type`|Datakit process IO count|
|*internal/metrics*|COUNTER|`datakit_process_io_bytes_total`|`type`|Datakit process IO bytes count|
|*internal/pipeline/stats*|COUNTER|`datakit_pipeline_point_total`|`category,name,namespace`|Pipeline processed total points|
|*internal/pipeline/stats*|COUNTER|`datakit_pipeline_drop_point_total`|`category,name,namespace`|Pipeline total dropped points|
|*internal/pipeline/stats*|COUNTER|`datakit_pipeline_error_point_total`|`category,name,namespace`|Pipeline processed total error points|
|*internal/pipeline/stats*|SUMMARY|`datakit_pipeline_cost_seconds`|`category,name,namespace`|Pipeline total running time|
|*internal/pipeline/stats*|GAUGE|`datakit_pipeline_last_update_timestamp_seconds`|`category,name,namespace`|Pipeline last update time|
|*internal/plugins/inputs/dialtesting*|GAUGE|`datakit_dialtesting_worker_job_chan_number`|`type`|The number of the channel for the jobs|
|*internal/plugins/inputs/dialtesting*|GAUGE|`datakit_dialtesting_worker_job_number`|`N/A`|The number of the jobs to send data in parallel|
|*internal/plugins/inputs/dialtesting*|GAUGE|`datakit_dialtesting_worker_cached_points_number`|`region,protocol`|The number of cached points|
|*internal/plugins/inputs/dialtesting*|GAUGE|`datakit_dialtesting_worker_send_points_number`|`region,protocol,status`|The number of the points which have been sent|
|*internal/plugins/inputs/dialtesting*|SUMMARY|`datakit_dialtesting_worker_send_cost_seconds`|`region,protocol`|Time cost to send points|
|*internal/plugins/inputs/dialtesting*|GAUGE|`datakit_dialtesting_task_number`|`region,protocol`|The number of tasks|
|*internal/plugins/inputs/dialtesting*|GAUGE|`datakit_dialtesting_dataway_send_failed_number`|`region,protocol,dataway`|The number of failed sending for each Dataway|
|*internal/plugins/inputs/dialtesting*|SUMMARY|`datakit_dialtesting_pull_cost_seconds`|`region,is_first`|Time cost to pull tasks|
|*internal/plugins/inputs/dialtesting*|COUNTER|`datakit_dialtesting_task_synchronized_total`|`region,protocol`|Task synchronized number|
|*internal/plugins/inputs/dialtesting*|COUNTER|`datakit_dialtesting_task_invalid_total`|`region,protocol,fail_reason`|Invalid task number|
|*internal/plugins/inputs/dialtesting*|SUMMARY|`datakit_dialtesting_task_check_cost_seconds`|`region,protocol,status`|Task check time|
|*internal/plugins/inputs/dialtesting*|SUMMARY|`datakit_dialtesting_task_run_cost_seconds`|`region,protocol`|Task run time|
|*internal/plugins/inputs/kafkamq*|COUNTER|`datakit_kafkamq_consumer_message_total`|`topic,partition,status`|Kafka consumer message numbers from Datakit start|
|*internal/plugins/inputs/kafkamq*|COUNTER|`datakit_kafkamq_group_election_total`|`N/A`|Kafka group election count|
|*internal/plugins/inputs*|GAUGE|`datakit_inputs_instance`|`input`|Input instance count|
|*internal/plugins/inputs*|COUNTER|`datakit_inputs_crash_total`|`input`|Input crash count|
|*internal/plugins/inputs/rum*|COUNTER|`datakit_rum_locate_statistics_total`|`app_id,ip_status,locate_status`|locate by ip addr statistics|
|*internal/plugins/inputs/rum*|COUNTER|`datakit_rum_source_map_total`|`app_id,sdk_name,status,remark`|source map result statistics|
|*internal/plugins/inputs/rum*|GAUGE|`datakit_rum_loaded_zip_cnt`|`platform`|RUM source map currently loaded zip archive count|
|*internal/plugins/inputs/rum*|SUMMARY|`datakit_rum_source_map_duration_seconds`|`sdk_name,app_id,env,version`|statistics elapsed time in RUM source map(unit: second)|
|*internal/prom*|SUMMARY|`datakit_prom_collect_points`|`source`|Total number of prom collection points|
|*internal/prom*|SUMMARY|`datakit_prom_http_get_bytes`|`source`|HTTP get bytes|
|*internal/prom*|SUMMARY|`datakit_prom_http_latency_in_second`|`source`|HTTP latency(in second)|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|COUNTER|`diskcache_put_bytes_total`|`path`|Cache Put() bytes count|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|COUNTER|`diskcache_get_total`|`path`|Cache Get() count|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|COUNTER|`diskcache_wakeup_total`|`path`|Wakeup count on sleeping write file|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|COUNTER|`diskcache_get_bytes_total`|`path`|Cache Get() bytes count|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|GAUGE|`diskcache_capacity`|`path`|Current capacity(in bytes)|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|GAUGE|`diskcache_max_data`|`path`|Max data to Put(in bytes), default 0|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|GAUGE|`diskcache_batch_size`|`path`|Data file size(in bytes)|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|GAUGE|`diskcache_size`|`path`|Current cache size(in bytes)|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|GAUGE|`diskcache_open_time`|`no_fallback_on_error,no_lock,no_pos,no_sync,path`|Current cache Open time in unix timestamp(second)|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|GAUGE|`diskcache_last_close_time`|`path`|Current cache last Close time in unix timestamp(second)|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|GAUGE|`diskcache_datafiles`|`path`|Current un-read data files|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|SUMMARY|`diskcache_get_latency`|`path`|Get() time cost(micro-second)|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|SUMMARY|`diskcache_put_latency`|`path`|Put() time cost(micro-second)|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|COUNTER|`diskcache_dropped_bytes_total`|`path`|Dropped bytes during Put() when capacity reached.|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|COUNTER|`diskcache_dropped_total`|`path`|Dropped files during Put() when capacity reached.|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|COUNTER|`diskcache_rotate_total`|`path`|Cache rotate count, mean file rotate from data to data.0000xxx|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|COUNTER|`diskcache_remove_total`|`path`|Removed file count, if some file read EOF, remove it from un-read list|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|COUNTER|`diskcache_put_total`|`path`|Cache Put() count|
