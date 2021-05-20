# solr采集器开发文档

配置示例

```toml
[[inputs.solr]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'
 
  ## specify a list of one or more Solr servers
  servers = ["http://localhost:8983"]
  
  ## Optional HTTP Basic Auth Credentials
  # username = "username"
  # password = "pa$$word"

  [inputs.solr.log]
    # files = []
    ## grok pipeline script path
    # pipeline = "solr.p"

  [inputs.solr.tags]
    # tag1 = "a"
```  

## Solr 指标采集

通过 Solr Metric API 获取数据

指标集：

* solr_cache:

  |指标|描述|数据类型|单位|
  |:-- |- |-|-|
  |cumulative_evictions|Number of cache evictions across all caches since this node has been running.|int|count|
  |cumulative_hitratio|Ratio of cache hits to lookups across all the caches since this node has been running.|float|%|
  |cumulative_hits|Number of cache hits across all the caches since this node has been running.|int|count|
  |cumulative_inserts|Number of cache insertions across all the caches since this node has been running.|int|count|
  |cumulative_lookups|Number of cache lookups across all the caches since this node has been running.|int|count|
  |evictions|Number of cache evictions for the current index searcher.|int|count|
  |hitratio|Ratio of cache hits to lookups for the current index searcher.|float|%|
  |hits|Number of hits for the current index searcher.|int|count|
  |inserts|Number of inserts into the cache.|int|count|
  |lookups|Number of lookups against the cache.|int|count|
  |max_ram|Maximum heap that should be used by the cache beyond which keys will be evicted.|int|MB|
  |ram_bytes_used|Actual heap usage of the cache at that particular instance.|int|Byte|
  |size|Number of entries in the cache at that particular instance.|int|count|
  |warmup|Warm-up time for the registered index searcher. This time is taken in account for the “auto-warming” of caches.|int|msec|

* solr_request_times:

  |指标|描述|数据类型|单位|
  |:-- |- |-|-|
  |count|Total number of requests made since the Solr process was started.|int|count|
  |max|Max of all the request processing time.|float|msec|
  |mean|Mean of all the request processing time.|float|msec|
  |median|Median of all the request processing time.|float|msec|
  |min|Min of all the request processing time.|float|msec|
  |p75|Request processing time for the request which belongs to the 75th Percentile.|float|msec|
  |p95|Request processing time in milliseconds for the request which belongs to the 95th Percentile.|float|msec|
  |p99|Request processing time in milliseconds for the request which belongs to the 99th Percentile.|float|msec|
  |p999|Request processing time in milliseconds for the request which belongs to the 99.9th Percentile.|float|msec|
  |rate_15min|Requests per second received over the past 15 minutes.|float|reqs/s|
  |rate_1min|Requests per second received over the past 1 minutes.|float|reqs/s|
  |rate_5min|Requests per second received over the past 5 minutes.|float|reqs/s|
  |rate_mean|Average number of requests per second received|float|reqs/s|
  |stddev|Stddev of all the request processing time.|float|msec|

* solr_searcher:

  |指标|描述|数据类型|单位|
  |:-- |- |-|-|
  |deleted_docs|The number of deleted documents.|int|count|
  |max_docs|The largest possible document number.|int|count|
  |num_docs|The total number of indexed documents.|int|count|
  |warmup|The time spent warming up.|int|msec|

## 日志采集

许配在配置文件中修改以下配置的 files 的值 以指向 solr 的日志

```toml
  [inputs.solr.log]
    # files = []
    ## grok pipeline script path
    # pipeline = "solr.p"
```
