### 采集器 elasticsearch

#### 指标相关
   - `telegraf` 采集的指标集有cluster, transport, breakers, indices, jvm, os, process, thread_pool等
   - `datadog` 大部分指标都能在telegraf上找到
    
 
#### conf

```
[[inputs.elasticsearch]]
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


#### 采集的指标集

- `elasticsearch_cluster_health`
- `elasticsearch_cluster_health_indices` 
- `elasticsearch_clusterstats_indices`
- `elasticsearch_clusterstats_nodes`
- `elasticsearch_transport`
- `elasticsearch_breakers`
- `elasticsearch_fs`
- `elasticsearch_http`
- `elasticsearch_indices`
- `elasticsearch_jvm`
- `elasticsearch_os`
- `elasticsearch_process`
- `elasticsearch_thread_pool`
- `elasticsearch_indices_stats_(primaries|total)`
- `elasticsearch_indices_stats_shards_total`
- `elasticsearch_indices_stats_shards`