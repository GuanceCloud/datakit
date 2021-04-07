- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}

# 简介

elasticsearch采集器主要采集节点运行情况、集群健康、JVM性能状况、索引性能、检索性能等。

## 前置条件

暂无

## 配置

进入 DataKit 安装目录下的 `conf.d/elasticsearch` 目录，复制 `elasticsearch.conf.sample` 并命名为 `elasticsearch.conf`。示例如下：

```
[[inputs.elasticsearch]]
  ## log file path
	#logFiles = ["/path/to/your/file.log"]

  ## specify a list of one or more Elasticsearch servers
  # you can add username and password to your url to use basic authentication:
  # servers = ["http://user:pass@localhost:9200"]
  servers = ["http://localhost:9200"]

  ## valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
  ## required
  interval = "10s"

  ## Timeout for HTTP requests to the elastic search server(s)
  http_timeout = "5s"

  ## When local is true (the default), the node will read only its own stats.
  ## Set local to false when you want to read the node stats from all nodes
  ## of the cluster.
  local = true

  ## Set cluster_health to true when you want to also obtain cluster health stats
  cluster_health = false

  ## Adjust cluster_health_level when you want to also obtain detailed health stats
  ## The options are
  ##  - indices (default)
  ##  - cluster
  # cluster_health_level = "indices"

  ## Set cluster_stats to true when you want to also obtain cluster stats.
  cluster_stats = false

  ## Only gather cluster_stats from the master node. To work this require local = true
  cluster_stats_only_from_master = true

  ## Indices to collect; can be one or more indices names or _all
  indices_include = ["_all"]

  ## One of "shards", "cluster", "indices"
  indices_level = "shards"

  ## node_stats is a list of sub-stats that you want to have gathered. Valid options
  ## are "indices", "os", "process", "jvm", "thread_pool", "fs", "transport", "http",
  ## "breaker". Per default, all stats are gathered.
  # node_stats = ["jvm", "http"]

  ## HTTP Basic Authentication username and password.
  # username = ""
  # password = ""

  ## Optional TLS Config
	tls_open = false
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

配置好后，重启 DataKit 即可。

## 指标集

### `elasticsearch_cluster_health`

-  标签
   
   标签名 | 描述 
   ---   | --- 
   name  |    

- 指标列表

  指标 | 描述 | 类型 | 单位
  --- | --- | --- | ---
  active_primary_shards | | integer  | 
  active_shards | | integer  | 
  active_shards_percent_as_number | | float  | 
  delayed_unassigned_shards | | integer  | 
  initializing_shards | | integer  | 
  number_of_data_nodes | | integer  | 
  number_of_in_flight_fetch | | integer  | 
  number_of_nodes | | integer  | 
  number_of_pending_tasks | | integer  | 
  relocating_shards | | integer  | 
  status | | string  | | one of green, yellow or red
  status_code | green = 1, yellow = 2, red = 3 | integer  |
  task_max_waiting_in_queue_millis | | integer  | 
  timed_out | | boolean  | 
  unassigned_shards | | integer  | 

### `elasticsearch_cluster_health_indices`
- 标签
   
   标签名 | 描述 
   ---   | --- 
   index |
   name  |

- 指标列表
  
  指标 | 描述 | 类型 | 单位
  --- | --- | --- | ---
  active_primary_shards ||integer|
  active_shards ||integer| 
  initializing_shards ||integer| 
  number_of_replicas ||integer| 
  number_of_shards ||integer| 
  relocating_shards ||integer| 
  status |one of green, yellow or red|string| 
  status_code |green = 1, yellow = 2, red = 3|integer| 
  unassigned_shards ||integer|

### `elasticsearch_cluster_stats`
- 标签
  
  标签名 | 描述 
   ---   | --- 
  cluster_name |
  node_name |
  status |

- 指标列表

  指标 | 描述 | 类型 | 单位
   --- | --- | --- | ---
  indices_completion_size_in_bytes ||float | 
  indices_count ||float | 
  indices_docs_count ||float | 
  indices_docs_deleted ||float | 
  indices_fielddata_evictions ||float | 
  indices_fielddata_memory_size_in_bytes ||float | 
  indices_query_cache_cache_count ||float | 
  indices_query_cache_cache_size ||float | 
  indices_query_cache_evictions ||float | 
  indices_query_cache_hit_count ||float | 
  indices_query_cache_memory_size_in_bytes ||float | 
  indices_query_cache_miss_count ||float | 
  indices_query_cache_total_count ||float | 
  indices_segments_count ||float | 
  indices_segments_doc_values_memory_in_bytes ||float | 
  indices_segments_fixed_bit_set_memory_in_bytes ||float | 
  indices_segments_index_writer_memory_in_bytes ||float | 
  indices_segments_max_unsafe_auto_id_timestamp ||float | 
  indices_segments_memory_in_bytes ||float | 
  indices_segments_norms_memory_in_bytes ||float | 
  indices_segments_points_memory_in_bytes ||float | 
  indices_segments_stored_fields_memory_in_bytes ||float | 
  indices_segments_term_vectors_memory_in_bytes ||float | 
  indices_segments_terms_memory_in_bytes ||float | 
  indices_segments_version_map_memory_in_bytes ||float | 
  indices_shards_index_primaries_avg ||float | 
  indices_shards_index_primaries_max ||float | 
  indices_shards_index_primaries_min ||float | 
  indices_shards_index_replication_avg ||float | 
  indices_shards_index_replication_max ||float | 
  indices_shards_index_replication_min ||float | 
  indices_shards_index_shards_avg ||float | 
  indices_shards_index_shards_max ||float | 
  indices_shards_index_shards_min ||float | 
  indices_shards_primaries ||float | 
  indices_shards_replication ||float | 
  indices_shards_total ||float | 
  indices_store_size_in_bytes ||float | 
  nodes_count_coordinating_only ||float | 
  nodes_count_data ||float | 
  nodes_count_ingest ||float | 
  nodes_count_master ||float | 
  nodes_count_total ||float | 
  nodes_fs_available_in_bytes ||float | 
  nodes_fs_free_in_bytes ||float | 
  nodes_fs_total_in_bytes ||float | 
  nodes_jvm_max_uptime_in_millis ||float | 
  nodes_jvm_mem_heap_max_in_bytes ||float | 
  nodes_jvm_mem_heap_used_in_bytes ||float | 
  nodes_jvm_threads ||float | 
  nodes_jvm_versions_0_count ||float | 
  nodes_jvm_versions_0_version ||string | 
  nodes_jvm_versions_0_vm_name ||string | 
  nodes_jvm_versions_0_vm_vendor ||string | 
  nodes_jvm_versions_0_vm_version ||string | 
  nodes_network_types_http_types_security4 ||float | 
  nodes_network_types_transport_types_security4 ||float | 
  nodes_os_allocated_processors ||float | 
  nodes_os_available_processors ||float | 
  nodes_os_mem_free_in_bytes ||float | 
  nodes_os_mem_free_percent ||float | 
  nodes_os_mem_total_in_bytes ||float | 
  nodes_os_mem_used_in_bytes ||float | 
  nodes_os_mem_used_percent ||float | 
  nodes_os_names_0_count ||float | 
  nodes_os_names_0_name ||string | 
  nodes_os_pretty_names_0_count ||float | 
  nodes_os_pretty_names_0_pretty_name ||string | 
  nodes_process_cpu_percent ||float | 
  nodes_process_open_file_descriptors_avg ||float | 
  nodes_process_open_file_descriptors_max ||float | 
  nodes_process_open_file_descriptors_min ||float | 
  nodes_versions_0 ||string | 

### `elasticsearch_node_stats`
- 标签
  
  标签名 | 描述 
   ---   | --- 
  cluster_name |
  node_attribute_ml.enabled |
  node_attribute_ml.machine_memory |
  node_attribute_ml.max_open_jobs |
  node_attribute_xpack.installed |
  node_host |
  node_id |
  node_name |

- 指标列表
  
  指标 | 描述 | 类型 | 单位
   --- | --- | --- | ---
  transport_rx_count ||float|
  transport_rx_size_in_bytes ||float|
  transport_server_open ||float|
  transport_tx_count ||float|
  transport_tx_size_in_bytes ||float|
  breakers_accounting_estimated_size_in_bytes ||float|
  breakers_accounting_limit_size_in_bytes ||float|
  breakers_accounting_overhead ||float|
  breakers_accounting_tripped ||float|
  breakers_fielddata_estimated_size_in_bytes ||float|
  breakers_fielddata_limit_size_in_bytes ||float|
  breakers_fielddata_overhead ||float|
  breakers_fielddata_tripped ||float|
  breakers_in_flight_requests_estimated_size_in_bytes ||float|
  breakers_in_flight_requests_limit_size_in_bytes ||float|
  breakers_in_flight_requests_overhead ||float|
  breakers_in_flight_requests_tripped ||float|
  breakers_parent_estimated_size_in_bytes ||float|
  breakers_parent_limit_size_in_bytes ||float|
  breakers_parent_overhead ||float|
  breakers_parent_tripped ||float|
  breakers_request_estimated_size_in_bytes ||float|
  breakers_request_limit_size_in_bytes ||float|
  breakers_request_overhead ||float|
  breakers_request_tripped ||float|
  fs_data_0_available_in_bytes ||float|
  fs_data_0_free_in_bytes ||float|
  fs_data_0_total_in_bytes ||float|
  fs_io_stats_devices_0_operations ||float|
  fs_io_stats_devices_0_read_kilobytes ||float|
  fs_io_stats_devices_0_read_operations ||float|
  fs_io_stats_devices_0_write_kilobytes ||float|
  fs_io_stats_devices_0_write_operations ||float|
  fs_io_stats_total_operations ||float|
  fs_io_stats_total_read_kilobytes ||float|
  fs_io_stats_total_read_operations ||float|
  fs_io_stats_total_write_kilobytes ||float|
  fs_io_stats_total_write_operations ||float|
  fs_timestamp ||float|
  fs_total_available_in_bytes ||float|
  fs_total_free_in_bytes ||float|
  fs_total_total_in_bytes ||float|
  http_current_open ||float|
  http_total_opened ||float|
  indices_completion_size_in_bytes ||float|
  indices_docs_count ||float|
  indices_docs_deleted ||float|
  indices_fielddata_evictions ||float|
  indices_fielddata_memory_size_in_bytes ||float|
  indices_flush_periodic ||float|
  indices_flush_total ||float|
  indices_flush_total_time_in_millis ||float|
  indices_get_current ||float|
  indices_get_exists_time_in_millis ||float|
  indices_get_exists_total ||float|
  indices_get_missing_time_in_millis ||float|
  indices_get_missing_total ||float|
  indices_get_time_in_millis ||float|
  indices_get_total ||float|
  indices_indexing_delete_current ||float|
  indices_indexing_delete_time_in_millis ||float|
  indices_indexing_delete_total ||float|
  indices_indexing_index_current ||float|
  indices_indexing_index_failed ||float|
  indices_indexing_index_time_in_millis ||float|
  indices_indexing_index_total ||float|
  indices_indexing_noop_update_total ||float|
  indices_indexing_throttle_time_in_millis ||float|
  indices_merges_current ||float|
  indices_merges_current_docs ||float|
  indices_merges_current_size_in_bytes ||float|
  indices_merges_total ||float|
  indices_merges_total_auto_throttle_in_bytes ||float|
  indices_merges_total_docs ||float|
  indices_merges_total_size_in_bytes ||float|
  indices_merges_total_stopped_time_in_millis ||float|
  indices_merges_total_throttled_time_in_millis ||float|
  indices_merges_total_time_in_millis ||float|
  indices_query_cache_cache_count ||float|
  indices_query_cache_cache_size ||float|
  indices_query_cache_evictions ||float|
  indices_query_cache_hit_count ||float|
  indices_query_cache_memory_size_in_bytes ||float|
  indices_query_cache_miss_count ||float|
  indices_query_cache_total_count ||float|
  indices_recovery_current_as_source ||float|
  indices_recovery_current_as_target ||float|
  indices_recovery_throttle_time_in_millis ||float|
  indices_refresh_listeners ||float|
  indices_refresh_total ||float|
  indices_refresh_total_time_in_millis ||float|
  indices_request_cache_evictions ||float|
  indices_request_cache_hit_count ||float|
  indices_request_cache_memory_size_in_bytes ||float|
  indices_request_cache_miss_count ||float|
  indices_search_fetch_current ||float|
  indices_search_fetch_time_in_millis ||float|
  indices_search_fetch_total ||float|
  indices_search_open_contexts ||float|
  indices_search_query_current ||float|
  indices_search_query_time_in_millis ||float|
  indices_search_query_total ||float|
  indices_search_scroll_current ||float|
  indices_search_scroll_time_in_millis ||float|
  indices_search_scroll_total ||float|
  indices_search_suggest_current ||float|
  indices_search_suggest_time_in_millis ||float|
  indices_search_suggest_total ||float|
  indices_segments_count ||float|
  indices_segments_doc_values_memory_in_bytes ||float|
  indices_segments_fixed_bit_set_memory_in_bytes ||float|
  indices_segments_index_writer_memory_in_bytes ||float|
  indices_segments_max_unsafe_auto_id_timestamp ||float|
  indices_segments_memory_in_bytes ||float|
  indices_segments_norms_memory_in_bytes ||float|
  indices_segments_points_memory_in_bytes ||float|
  indices_segments_stored_fields_memory_in_bytes ||float|
  indices_segments_term_vectors_memory_in_bytes ||float|
  indices_segments_terms_memory_in_bytes ||float|
  indices_segments_version_map_memory_in_bytes ||float|
  indices_store_size_in_bytes ||float|
  indices_translog_earliest_last_modified_age ||float|
  indices_translog_operations ||float|
  indices_translog_size_in_bytes ||float|
  indices_translog_uncommitted_operations ||float|
  indices_translog_uncommitted_size_in_bytes ||float|
  indices_warmer_current ||float|
  indices_warmer_total ||float|
  indices_warmer_total_time_in_millis ||float|
  jvm_buffer_pools_direct_count ||float|
  jvm_buffer_pools_direct_total_capacity_in_bytes ||float|
  jvm_buffer_pools_direct_used_in_bytes ||float|
  jvm_buffer_pools_mapped_count ||float|
  jvm_buffer_pools_mapped_total_capacity_in_bytes ||float|
  jvm_buffer_pools_mapped_used_in_bytes ||float|
  jvm_classes_current_loaded_count ||float|
  jvm_classes_total_loaded_count ||float|
  jvm_classes_total_unloaded_count ||float|
  jvm_gc_collectors_old_collection_count ||float|
  jvm_gc_collectors_old_collection_time_in_millis ||float|
  jvm_gc_collectors_young_collection_count ||float|
  jvm_gc_collectors_young_collection_time_in_millis ||float|
  jvm_mem_heap_committed_in_bytes ||float|
  jvm_mem_heap_max_in_bytes ||float|
  jvm_mem_heap_used_in_bytes ||float|
  jvm_mem_heap_used_percent ||float|
  jvm_mem_non_heap_committed_in_bytes ||float|
  jvm_mem_non_heap_used_in_bytes ||float|
  jvm_mem_pools_old_max_in_bytes ||float|
  jvm_mem_pools_old_peak_max_in_bytes ||float|
  jvm_mem_pools_old_peak_used_in_bytes ||float|
  jvm_mem_pools_old_used_in_bytes ||float|
  jvm_mem_pools_survivor_max_in_bytes ||float|
  jvm_mem_pools_survivor_peak_max_in_bytes ||float|
  jvm_mem_pools_survivor_peak_used_in_bytes ||float|
  jvm_mem_pools_survivor_used_in_bytes ||float|
  jvm_mem_pools_young_max_in_bytes ||float|
  jvm_mem_pools_young_peak_max_in_bytes ||float|
  jvm_mem_pools_young_peak_used_in_bytes ||float|
  jvm_mem_pools_young_used_in_bytes ||float|
  jvm_threads_count ||float|
  jvm_threads_peak_count ||float|
  jvm_timestamp ||float|
  jvm_uptime_in_millis ||float|
  os_cgroup_cpu_cfs_period_micros ||float|
  os_cgroup_cpu_cfs_quota_micros ||float|
  os_cgroup_cpu_stat_number_of_elapsed_periods ||float|
  os_cgroup_cpu_stat_number_of_times_throttled ||float|
  os_cgroup_cpu_stat_time_throttled_nanos ||float|
  os_cgroup_cpuacct_usage_nanos ||float|
  os_cpu_load_average_15m ||float|
  os_cpu_load_average_1m ||float|
  os_cpu_load_average_5m ||float|
  os_cpu_percent ||float|
  os_mem_free_in_bytes ||float|
  os_mem_free_percent ||float|
  os_mem_total_in_bytes ||float|
  os_mem_used_in_bytes ||float|
  os_mem_used_percent ||float|
  os_swap_free_in_bytes ||float|
  os_swap_total_in_bytes ||float|
  os_swap_used_in_bytes ||float|
  os_timestamp ||float|
  process_cpu_percent ||float|
  process_cpu_total_in_millis ||float|
  process_max_file_descriptors ||float|
  process_mem_total_virtual_in_bytes ||float|
  process_open_file_descriptors ||float|
  process_timestamp ||float|
  thread_pool_analyze_active ||float|
  thread_pool_analyze_completed ||float|
  thread_pool_analyze_largest ||float|
  thread_pool_analyze_queue ||float|
  thread_pool_analyze_rejected ||float|
  thread_pool_analyze_threads ||float|
  thread_pool_ccr_active ||float|
  thread_pool_ccr_completed ||float|
  thread_pool_ccr_largest ||float|
  thread_pool_ccr_queue ||float|
  thread_pool_ccr_rejected ||float|
  thread_pool_ccr_threads ||float|
  thread_pool_fetch_shard_started_active ||float|
  thread_pool_fetch_shard_started_completed ||float|
  thread_pool_fetch_shard_started_largest ||float|
  thread_pool_fetch_shard_started_queue ||float|
  thread_pool_fetch_shard_started_rejected ||float|
  thread_pool_fetch_shard_started_threads ||float|
  thread_pool_fetch_shard_store_active ||float|
  thread_pool_fetch_shard_store_completed ||float|
  thread_pool_fetch_shard_store_largest ||float|
  thread_pool_fetch_shard_store_queue ||float|
  thread_pool_fetch_shard_store_rejected ||float|
  thread_pool_fetch_shard_store_threads ||float|
  thread_pool_flush_active ||float|
  thread_pool_flush_completed ||float|
  thread_pool_flush_largest ||float|
  thread_pool_flush_queue ||float|
  thread_pool_flush_rejected ||float|
  thread_pool_flush_threads ||float|
  thread_pool_force_merge_active ||float|
  thread_pool_force_merge_completed ||float|
  thread_pool_force_merge_largest ||float|
  thread_pool_force_merge_queue ||float|
  thread_pool_force_merge_rejected ||float|
  thread_pool_force_merge_threads ||float|
  thread_pool_generic_active ||float|
  thread_pool_generic_completed ||float|
  thread_pool_generic_largest ||float|
  thread_pool_generic_queue ||float|
  thread_pool_generic_rejected ||float|
  thread_pool_generic_threads ||float|
  thread_pool_get_active ||float|
  thread_pool_get_completed ||float|
  thread_pool_get_largest ||float|
  thread_pool_get_queue ||float|
  thread_pool_get_rejected ||float|
  thread_pool_get_threads ||float|
  thread_pool_index_active ||float|
  thread_pool_index_completed ||float|
  thread_pool_index_largest ||float|
  thread_pool_index_queue ||float|
  thread_pool_index_rejected ||float|
  thread_pool_index_threads ||float|
  thread_pool_listener_active ||float|
  thread_pool_listener_completed ||float|
  thread_pool_listener_largest ||float|
  thread_pool_listener_queue ||float|
  thread_pool_listener_rejected ||float|
  thread_pool_listener_threads ||float|
  thread_pool_management_active ||float|
  thread_pool_management_completed ||float|
  thread_pool_management_largest ||float|
  thread_pool_management_queue ||float|
  thread_pool_management_rejected ||float|
  thread_pool_management_threads ||float|
  thread_pool_ml_autodetect_active ||float|
  thread_pool_ml_autodetect_completed ||float|
  thread_pool_ml_autodetect_largest ||float|
  thread_pool_ml_autodetect_queue ||float|
  thread_pool_ml_autodetect_rejected ||float|
  thread_pool_ml_autodetect_threads ||float|
  thread_pool_ml_datafeed_active ||float|
  thread_pool_ml_datafeed_completed ||float|
  thread_pool_ml_datafeed_largest ||float|
  thread_pool_ml_datafeed_queue ||float|
  thread_pool_ml_datafeed_rejected ||float|
  thread_pool_ml_datafeed_threads ||float|
  thread_pool_ml_utility_active ||float|
  thread_pool_ml_utility_completed ||float|
  thread_pool_ml_utility_largest ||float|
  thread_pool_ml_utility_queue ||float|
  thread_pool_ml_utility_rejected ||float|
  thread_pool_ml_utility_threads ||float|
  thread_pool_refresh_active ||float|
  thread_pool_refresh_completed ||float|
  thread_pool_refresh_largest ||float|
  thread_pool_refresh_queue ||float|
  thread_pool_refresh_rejected ||float|
  thread_pool_refresh_threads ||float|
  thread_pool_rollup_indexing_active ||float|
  thread_pool_rollup_indexing_completed ||float|
  thread_pool_rollup_indexing_largest ||float|
  thread_pool_rollup_indexing_queue ||float|
  thread_pool_rollup_indexing_rejected ||float|
  thread_pool_rollup_indexing_threads ||float|
  thread_pool_search_active ||float|
  thread_pool_search_completed ||float|
  thread_pool_search_largest ||float|
  thread_pool_search_queue ||float|
  thread_pool_search_rejected ||float|
  thread_pool_search_threads ||float|
  thread_pool_search_throttled_active ||float|
  thread_pool_search_throttled_completed ||float|
  thread_pool_search_throttled_largest ||float|
  thread_pool_search_throttled_queue ||float|
  thread_pool_search_throttled_rejected ||float|
  thread_pool_search_throttled_threads ||float|
  thread_pool_security-token-key_active ||float|
  thread_pool_security-token-key_completed ||float|
  thread_pool_security-token-key_largest ||float|
  thread_pool_security-token-key_queue ||float|
  thread_pool_security-token-key_rejected ||float|
  thread_pool_security-token-key_threads ||float|
  thread_pool_snapshot_active ||float|
  thread_pool_snapshot_completed ||float|
  thread_pool_snapshot_largest ||float|
  thread_pool_snapshot_queue ||float|
  thread_pool_snapshot_rejected ||float|
  thread_pool_snapshot_threads ||float|
  thread_pool_warmer_active ||float|
  thread_pool_warmer_completed ||float|
  thread_pool_warmer_largest ||float|
  thread_pool_warmer_queue ||float|
  thread_pool_warmer_rejected ||float|
  thread_pool_warmer_threads ||float|
  thread_pool_watcher_active ||float|
  thread_pool_watcher_completed ||float|
  thread_pool_watcher_largest ||float|
  thread_pool_watcher_queue ||float|
  thread_pool_watcher_rejected ||float|
  thread_pool_watcher_threads ||float|
  thread_pool_write_active ||float|
  thread_pool_write_completed ||float|
  thread_pool_write_largest ||float|
  thread_pool_write_queue ||float|
  thread_pool_write_rejected ||float|
  thread_pool_write_threads ||float|



### `elasticsearch_indices_stats`
- 标签
  
  标签名 | 描述 
   ---   | --- 
  index_name | 
  
- 指标列表
  
  指标 | 描述 | 类型 | 单位
   --- | --- | --- | ---
  (primaries/total)_completion_size_in_bytes ||float|
  (primaries/total)_docs_count ||float|
  (primaries/total)_docs_deleted ||float|
  (primaries/total)_fielddata_evictions ||float|
  (primaries/total)_fielddata_memory_size_in_bytes ||float|
  (primaries/total)_flush_periodic ||float|
  (primaries/total)_flush_total ||float|
  (primaries/total)_flush_total_time_in_millis ||float|
  (primaries/total)_get_current ||float|
  (primaries/total)_get_exists_time_in_millis ||float|
  (primaries/total)_get_exists_total ||float|
  (primaries/total)_get_missing_time_in_millis ||float|
  (primaries/total)_get_missing_total ||float|
  (primaries/total)_get_time_in_millis ||float|
  (primaries/total)_get_total ||float|
  (primaries/total)_indexing_delete_current ||float|
  (primaries/total)_indexing_delete_time_in_millis ||float|
  (primaries/total)_indexing_delete_total ||float|
  (primaries/total)_indexing_index_current ||float|
  (primaries/total)_indexing_index_failed ||float|
  (primaries/total)_indexing_index_time_in_millis ||float|
  (primaries/total)_indexing_index_total ||float|
  (primaries/total)_indexing_is_throttled ||float|
  (primaries/total)_indexing_noop_update_total ||float|
  (primaries/total)_indexing_throttle_time_in_millis ||float|
  (primaries/total)_merges_current ||float|
  (primaries/total)_merges_current_docs ||float|
  (primaries/total)_merges_current_size_in_bytes ||float|
  (primaries/total)_merges_total ||float|
  (primaries/total)_merges_total_auto_throttle_in_bytes ||float|
  (primaries/total)_merges_total_docs ||float|
  (primaries/total)_merges_total_size_in_bytes ||float|
  (primaries/total)_merges_total_stopped_time_in_millis ||float|
  (primaries/total)_merges_total_throttled_time_in_millis ||float|
  (primaries/total)_merges_total_time_in_millis ||float|
  (primaries/total)_query_cache_cache_count ||float|
  (primaries/total)_query_cache_cache_size ||float|
  (primaries/total)_query_cache_evictions ||float|
  (primaries/total)_query_cache_hit_count ||float|
  (primaries/total)_query_cache_memory_size_in_bytes ||float|
  (primaries/total)_query_cache_miss_count ||float|
  (primaries/total)_query_cache_total_count ||float|
  (primaries/total)_recovery_current_as_source ||float|
  (primaries/total)_recovery_current_as_target ||float|
  (primaries/total)_recovery_throttle_time_in_millis ||float|
  (primaries/total)_refresh_external_total ||float|
  (primaries/total)_refresh_external_total_time_in_millis ||float|
  (primaries/total)_refresh_listeners ||float|
  (primaries/total)_refresh_total ||float|
  (primaries/total)_refresh_total_time_in_millis ||float|
  (primaries/total)_request_cache_evictions ||float|
  (primaries/total)_request_cache_hit_count ||float|
  (primaries/total)_request_cache_memory_size_in_bytes ||float|
  (primaries/total)_request_cache_miss_count ||float|
  (primaries/total)_search_fetch_current ||float|
  (primaries/total)_search_fetch_time_in_millis ||float|
  (primaries/total)_search_fetch_total ||float|
  (primaries/total)_search_open_contexts ||float|
  (primaries/total)_search_query_current ||float|
  (primaries/total)_search_query_time_in_millis ||float|
  (primaries/total)_search_query_total ||float|
  (primaries/total)_search_scroll_current ||float|
  (primaries/total)_search_scroll_time_in_millis ||float|
  (primaries/total)_search_scroll_total ||float|
  (primaries/total)_search_suggest_current ||float|
  (primaries/total)_search_suggest_time_in_millis ||float|
  (primaries/total)_search_suggest_total ||float|
  (primaries/total)_segments_count ||float|
  (primaries/total)_segments_doc_values_memory_in_bytes ||float|
  (primaries/total)_segments_fixed_bit_set_memory_in_bytes ||float|
  (primaries/total)_segments_index_writer_memory_in_bytes ||float|
  (primaries/total)_segments_max_unsafe_auto_id_timestamp ||float|
  (primaries/total)_segments_memory_in_bytes ||float|
  (primaries/total)_segments_norms_memory_in_bytes ||float|
  (primaries/total)_segments_points_memory_in_bytes ||float|
  (primaries/total)_segments_stored_fields_memory_in_bytes ||float|
  (primaries/total)_segments_term_vectors_memory_in_bytes ||float|
  (primaries/total)_segments_terms_memory_in_bytes ||float|
  (primaries/total)_segments_version_map_memory_in_bytes ||float|
  (primaries/total)_store_size_in_bytes ||float|
  (primaries/total)_translog_earliest_last_modified_age ||float|
  (primaries/total)_translog_operations ||float|
  (primaries/total)_translog_size_in_bytes ||float|
  (primaries/total)_translog_uncommitted_operations ||float|
  (primaries/total)_translog_uncommitted_size_in_bytes ||float|
  (primaries/total)_warmer_current ||float|
  (primaries/total)_warmer_total ||float|
  (primaries/total)_warmer_total_time_in_millis ||float|

### `elasticsearch_indices_stats_shards_total`
- 标签
- 指标列表
  
  指标 | 描述 | 类型 | 单位
   --- | --- | --- | ---
  failed || float |
  successful || float |
  total || float |

### `elasticsearch_indices_stats_shards`
- 标签

  标签名 | 描述
  --- | ----
  index_name |
  node_name |
  shard_name |
  type |

- 指标列表
  
  指标 | 描述 | 类型 | 单位
   --- | --- | --- | ---
  commit_generation || float |
  commit_num_docs || float |
  completion_size_in_bytes || float |
  docs_count || float |
  docs_deleted || float |
  fielddata_evictions || float |
  fielddata_memory_size_in_bytes || float |
  flush_periodic || float |
  flush_total || float |
  flush_total_time_in_millis || float |
  get_current || float |
  get_exists_time_in_millis || float |
  get_exists_total || float |
  get_missing_time_in_millis || float |
  get_missing_total || float |
  get_time_in_millis || float |
  get_total || float |
  indexing_delete_current || float |
  indexing_delete_time_in_millis || float |
  indexing_delete_total || float |
  indexing_index_current || float |
  indexing_index_failed || float |
  indexing_index_time_in_millis || float |
  indexing_index_total || float |
  indexing_is_throttled || bool |
  indexing_noop_update_total || float |
  indexing_throttle_time_in_millis || float |
  merges_current || float |
  merges_current_docs || float |
  merges_current_size_in_bytes || float |
  merges_total || float |
  merges_total_auto_throttle_in_bytes || float |
  merges_total_docs || float |
  merges_total_size_in_bytes || float |
  merges_total_stopped_time_in_millis || float |
  merges_total_throttled_time_in_millis || float |
  merges_total_time_in_millis || float |
  query_cache_cache_count || float |
  query_cache_cache_size || float |
  query_cache_evictions || float |
  query_cache_hit_count || float |
  query_cache_memory_size_in_bytes || float |
  query_cache_miss_count || float |
  query_cache_total_count || float |
  recovery_current_as_source || float |
  recovery_current_as_target || float |
  recovery_throttle_time_in_millis || float |
  refresh_external_total || float |
  refresh_external_total_time_in_millis || float |
  refresh_listeners || float |
  refresh_total || float |
  refresh_total_time_in_millis || float |
  request_cache_evictions || float |
  request_cache_hit_count || float |
  request_cache_memory_size_in_bytes || float |
  request_cache_miss_count || float |
  retention_leases_primary_term || float |
  retention_leases_version || float |
  routing_state || int |
  search_fetch_current || float |
  search_fetch_time_in_millis || float |
  search_fetch_total || float |
  search_open_contexts || float |
  search_query_current || float |
  search_query_time_in_millis || float |
  search_query_total || float |
  search_scroll_current || float |
  search_scroll_time_in_millis || float |
  search_scroll_total || float |
  search_suggest_current || float |
  search_suggest_time_in_millis || float |
  search_suggest_total || float |
  segments_count || float |
  segments_doc_values_memory_in_bytes || float |
  segments_fixed_bit_set_memory_in_bytes || float |
  segments_index_writer_memory_in_bytes || float |
  segments_max_unsafe_auto_id_timestamp || float |
  segments_memory_in_bytes || float |
  segments_norms_memory_in_bytes || float |
  segments_points_memory_in_bytes || float |
  segments_stored_fields_memory_in_bytes || float |
  segments_term_vectors_memory_in_bytes || float |
  segments_terms_memory_in_bytes || float |
  segments_version_map_memory_in_bytes || float |
  seq_no_global_checkpoint || float |
  seq_no_local_checkpoint || float |
  seq_no_max_seq_no || float |
  shard_path_is_custom_data_path || bool |
  store_size_in_bytes || float |
  translog_earliest_last_modified_age || float |
  translog_operations || float |
  translog_size_in_bytes || float |
  translog_uncommitted_operations || float |
  translog_uncommitted_size_in_bytes || float |
  warmer_current || float |
  warmer_total || float |
  warmer_total_time_in_millis || float |
  