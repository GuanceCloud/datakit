# Datakit 自身指标

---

为便于 Datakit 自身可观测性，我们在开发 Datakit 过程中，给相关的业务模块增加了很多 Prometheus 指标暴露，通过暴露这些指标，我们能方便的排查 Datakit 运行过程中的一些问题。

## 指标列表 {#metrics}

自 Datakit [1.5.9 版本](changelog.md#cl-1.5.9)以来，通过访问 `http://localhost:9529/metrics` 即可获取当前的指标列表，不同 Datakit 版本可能会对一些相关指标做调整，或者增删一些指标。

这些指标，在 [Datakit monitor](datakit-monitor.md) 展示中也会用到，只是 monitor 中为了展示上的友好型，做了一些优化处理。如果要查看原始的指标（或者 monitor 上没有展示出来的指标），我们可以通过 `curl` 和 `watch` 命令的组合来查看，比如获取 Datakit 进程 CPU 的使用情况：

```shell
# 每隔 3s 获取一次 CPU 使用率指标
$ watch -n 3 'curl -s http://localhost:9529/metrics | grep -a datakit_cpu_usage'

# HELP datakit_cpu_usage Datakit CPU usage(%)
# TYPE datakit_cpu_usage gauge
datakit_cpu_usage 4.9920266849857144
```

其它指标也能通过类似方式来观察，目前已有的指标如下（当前版本 {{ .Version }}）：

<!-- 以下这些指标，通过执行 make show_metrics 方式能获取 -->

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
|*internal/io/dataway*|SUMMARY|`datakit_io_build_body_batch_bytes`|`category,encoding`|Batch HTTP body size|
|*internal/io/dataway*|COUNTER|`datakit_io_dataway_point_total`|`category,status`|Dataway uploaded points, partitioned by category and send status(HTTP status)|
|*internal/io/dataway*|COUNTER|`datakit_io_dataway_point_bytes_total`|`category,enc,status`|Dataway uploaded points bytes, partitioned by category and pint send status(HTTP status)|
|*internal/io/dataway*|SUMMARY|`datakit_io_dataway_api_latency_seconds`|`api,status`|Dataway HTTP request latency partitioned by HTTP API(method@url) and HTTP status|
|*internal/io/dataway*|COUNTER|`datakit_io_http_retry_total`|`api,status`|Dataway HTTP retried count|
|*internal/io/dataway*|COUNTER|`datakit_io_grouped_request_total`|`category`|Grouped requests under sinker|
|*internal/io/dataway*|SUMMARY|`datakit_io_flush_failcache_bytes`|`category`|IO flush fail-cache bytes(in gzip) summary|
|*internal/io/dataway*|SUMMARY|`datakit_io_build_body_cost_seconds`|`category,encoding`|Build point HTTP body cost|
|*internal/io/dataway*|SUMMARY|`datakit_io_build_body_batches`|`category,encoding`|Batch HTTP body batches|
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
|*internal/io*|SUMMARY|`datakit_io_feed_cost_seconds`|`category,from`|IO feed waiting(on block mode) seconds|
|*internal/io*|COUNTER|`datakit_io_feed_drop_point_total`|`category,from`|IO feed drop(on non-block mode) points|
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
|*internal/metrics*|GAUGE|`datakit_uptime_seconds`|`hostname,resource_limit,lite,version=?,build_at=?,branch=?,os_arch=?,docker=?,auto_update=?`|Datakit uptime|
|*internal/metrics*|GAUGE|`datakit_data_overuse`|`N/A`|Does current workspace's data(metric/logging) usage(if 0 not beyond, or with a unix timestamp when overuse occurred)|
|*internal/metrics*|COUNTER|`datakit_process_ctx_switch_total`|`type`|Datakit process context switch count(Linux only)|
|*internal/metrics*|COUNTER|`datakit_process_io_count_total`|`type`|Datakit process IO count|
|*internal/metrics*|COUNTER|`datakit_process_io_bytes_total`|`type`|Datakit process IO bytes count|
|*internal/plugins/inputs/container/kubernetes*|GAUGE|`datakit_kubernetes_fetch_error`|`namespace,resource,error`|Kubernetes resource fetch error|
|*internal/plugins/inputs/container/kubernetes*|SUMMARY|`datakit_kubernetes_collect_cost_seconds`|`category`|Kubernetes resource collect cost|
|*internal/plugins/inputs/container/kubernetes*|COUNTER|`datakit_kubernetes_collect_pts_total`|`category`|Kubernetes collect point total|
|*internal/plugins/inputs/container/kubernetes*|SUMMARY|`datakit_kubernetes_collect_resource_pts_num`|`category,kind`|Kubernetes resource collect point count|
|*internal/plugins/inputs/container*|SUMMARY|`datakit_container_collect_cost_seconds`|`category`|Container collect cost|
|*internal/plugins/inputs/container*|COUNTER|`datakit_container_collect_pts_total`|`category`|Container collect point total|
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
|*internal/plugins/inputs/rum*|SUMMARY|`datakit_rum_session_replay_upload_latency_seconds`|`app_id,env,version,service,status_code`|statistics elapsed time in session replay uploading|
|*internal/plugins/inputs/rum*|COUNTER|`datakit_session_replay_session_replay_dropped_total`|`app_id,env,version,service,status_code`|statistics count of dropped session replay points since uploading fail|
|*internal/prom*|SUMMARY|`datakit_prom_collect_points`|`source`|Total number of prom collection points|
|*internal/prom*|SUMMARY|`datakit_prom_http_get_bytes`|`source`|HTTP get bytes|
|*internal/prom*|SUMMARY|`datakit_prom_http_latency_in_second`|`source`|HTTP latency(in second)|
|*internal/tailer*|COUNTER|`datakit_tailer_collect_multiline_state_total`|`source,filepath,multilinestate`|Tailer multiline state total|
|*internal/tailer*|COUNTER|`datakit_tailer_file_rotate_total`|`source,filepath`|Tailer rotate total|
|*internal/tailer*|COUNTER|`datakit_tailer_parse_fail_total`|`source,filepath,mode`|Tailer parse fail total|
|*internal/tailer*|GAUGE|`datakit_tailer_open_file_num`|`mode`|Tailer open file total|
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
|*vendor/github.com/GuanceCloud/cliutils/pipeline/stats*|COUNTER|`datakit_pipeline_point_total`|`category,name,namespace`|Pipeline processed total points|
|*vendor/github.com/GuanceCloud/cliutils/pipeline/stats*|COUNTER|`datakit_pipeline_drop_point_total`|`category,name,namespace`|Pipeline total dropped points|
|*vendor/github.com/GuanceCloud/cliutils/pipeline/stats*|COUNTER|`datakit_pipeline_error_point_total`|`category,name,namespace`|Pipeline processed total error points|
|*vendor/github.com/GuanceCloud/cliutils/pipeline/stats*|SUMMARY|`datakit_pipeline_cost_seconds`|`category,name,namespace`|Pipeline total running time|
|*vendor/github.com/GuanceCloud/cliutils/pipeline/stats*|GAUGE|`datakit_pipeline_last_update_timestamp_seconds`|`category,name,namespace`|Pipeline last update time|
