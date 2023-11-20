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
|*internal/config*|GAUGE|`datakit_config_datakit_ulimit`|`status`|Datakit ulimit|
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
|*internal/io/dataway*|SUMMARY|`datakit_io_build_body_batch_bytes`|`category,encoding,gzip`|Batch HTTP body size|
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
|*internal/metrics*|GAUGE|`datakit_goroutines`|`N/A`|Goroutine count within Datakit|
|*internal/metrics*|GAUGE|`datakit_heap_alloc_bytes`|`N/A`|Datakit memory heap bytes|
|*internal/metrics*|GAUGE|`datakit_sys_alloc_bytes`|`N/A`|Datakit memory system bytes|
|*internal/metrics*|GAUGE|`datakit_cpu_usage`|`N/A`|Datakit CPU usage(%)|
|*internal/metrics*|GAUGE|`datakit_open_files`|`N/A`|Datakit open files(only available on Linux)|
|*internal/metrics*|GAUGE|`datakit_cpu_cores`|`N/A`|Datakit CPU cores|
|*internal/metrics*|GAUGE|`datakit_uptime_seconds`|`hostname,resource_limit,lite,branch=?,os_arch=?,docker=?,auto_update=?,version=?,build_at=?`|Datakit uptime|
|*internal/metrics*|GAUGE|`datakit_data_overuse`|`N/A`|Does current workspace's data(metric/logging) usage(if 0 not beyond, or with a unix timestamp when overuse occurred)|
|*internal/metrics*|COUNTER|`datakit_process_ctx_switch_total`|`type`|Datakit process context switch count(Linux only)|
|*internal/metrics*|COUNTER|`datakit_process_io_count_total`|`type`|Datakit process IO count|
|*internal/metrics*|COUNTER|`datakit_process_io_bytes_total`|`type`|Datakit process IO bytes count|
|*internal/plugins/inputs/container/kubernetes*|GAUGE|`datakit_kubernetes_fetch_error`|`namespace,resource,error`|Kubernetes resource fetch error|
|*internal/plugins/inputs/container/kubernetes*|SUMMARY|`datakit_kubernetes_collect_cost_seconds`|`category`|Kubernetes collect cost|
|*internal/plugins/inputs/container/kubernetes*|SUMMARY|`datakit_kubernetes_collect_resource_cost_seconds`|`category,kind,fieldselector`|Kubernetes collect resource cost|
|*internal/plugins/inputs/container/kubernetes*|COUNTER|`datakit_kubernetes_collect_pts_total`|`category`|Kubernetes collect point total|
|*internal/plugins/inputs/container/kubernetes*|SUMMARY|`datakit_kubernetes_collect_resource_pts_num`|`category,kind,fieldselector`|Kubernetes resource collect point count|
|*internal/plugins/inputs/container/kubernetes*|COUNTER|`datakit_kubernetes_pod_metrics_query_total`|`target`|Kubernetes query pod metrics count|
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
|*internal/plugins/inputs/rum*|GAUGE|`datakit_rum_loaded_zips`|`platform`|RUM source map currently loaded zip archive count|
|*internal/plugins/inputs/rum*|SUMMARY|`datakit_rum_source_map_duration_seconds`|`sdk_name,app_id,env,version`|statistics elapsed time in RUM source map(unit: second)|
|*internal/plugins/inputs/rum*|SUMMARY|`datakit_rum_session_replay_upload_latency_seconds`|`app_id,env,version,service,status_code`|statistics elapsed time in session replay uploading|
|*internal/plugins/inputs/rum*|COUNTER|`datakit_rum_session_replay_dropped_total`|`app_id,env,version,service,status_code`|statistics count of dropped session replay points since uploading fail|
|*internal/prom*|SUMMARY|`datakit_prom_collect_points`|`mode,source`|Total number of prom collection points|
|*internal/prom*|SUMMARY|`datakit_prom_http_get_bytes`|`mode,source`|HTTP get bytes|
|*internal/prom*|SUMMARY|`datakit_prom_http_latency_in_second`|`mode,source`|HTTP latency(in second)|
|*internal/prom*|GAUGE|`datakit_prom_stream_size`|`mode,source`|Stream size|
|*internal/tailer*|COUNTER|`datakit_tailer_collect_multiline_state_total`|`source,filepath,multilinestate`|Tailer multiline state total|
|*internal/tailer*|COUNTER|`datakit_tailer_file_rotate_total`|`source,filepath`|Tailer rotate total|
|*internal/tailer*|COUNTER|`datakit_tailer_buffer_force_flush_total`|`source,filepath`|Tailer force flush total|
|*internal/tailer*|COUNTER|`datakit_tailer_parse_fail_total`|`source,filepath,mode`|Tailer parse fail total|
|*internal/tailer*|GAUGE|`datakit_tailer_open_file_num`|`mode`|Tailer open file total|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|COUNTER|`diskcache_put_bytes_total`|`path`|Cache Put() bytes count|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|COUNTER|`diskcache_get_total`|`path`|Cache Get() count|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|COUNTER|`diskcache_wakeup_total`|`path`|Wakeup count on sleeping write file|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|COUNTER|`diskcache_seek_back_total`|`path`|Seek back when Get() got any error|
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
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|COUNTER|`diskcache_dropped_total`|`path,reason`|Dropped files during Put() when capacity reached.|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|COUNTER|`diskcache_rotate_total`|`path`|Cache rotate count, mean file rotate from data to data.0000xxx|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|COUNTER|`diskcache_remove_total`|`path`|Removed file count, if some file read EOF, remove it from un-read list|
|*vendor/github.com/GuanceCloud/cliutils/diskcache*|COUNTER|`diskcache_put_total`|`path`|Cache Put() count|
|*vendor/github.com/GuanceCloud/cliutils/pipeline/stats*|COUNTER|`datakit_pipeline_point_total`|`category,name,namespace`|Pipeline processed total points|
|*vendor/github.com/GuanceCloud/cliutils/pipeline/stats*|COUNTER|`datakit_pipeline_drop_point_total`|`category,name,namespace`|Pipeline total dropped points|
|*vendor/github.com/GuanceCloud/cliutils/pipeline/stats*|COUNTER|`datakit_pipeline_error_point_total`|`category,name,namespace`|Pipeline processed total error points|
|*vendor/github.com/GuanceCloud/cliutils/pipeline/stats*|SUMMARY|`datakit_pipeline_cost_seconds`|`category,name,namespace`|Pipeline total running time|
|*vendor/github.com/GuanceCloud/cliutils/pipeline/stats*|GAUGE|`datakit_pipeline_last_update_timestamp_seconds`|`category,name,namespace`|Pipeline last update time|

### Golang Runtime Metrics {#go-runtime-metrics}

On API `/metrics`, Datakit exported Golang runtime metrics, it seems like this:

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
