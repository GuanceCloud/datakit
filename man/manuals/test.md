
# Elasticsearch 性能指标采集

- catalog: SuperCloud IT 监控
- thumbnail: attachments/elasticsearch.png
- tags: IT运维,指标采集

<!-- s/\%u00a0/  /g 1000000000 -->

## 简介

Elasticsearch提供了许多指标，可以帮助您检测故障迹象并在遇到不可靠的节点，内存不足的错误以及较长的垃圾收集时间等问题时采取措施。需要监视的几个关键领域是：

* 集群运行状况和节点可用性
* 主机的网络和系统
* 搜索性能指标
* 索引性能指标
* 内存使用和GC指标
* 资源饱和度和错误

## 场景视图

![场景视图](./attachments/ElasticSearch 场景视图1.png)
![场景视图](./attachments/ElasticSearch 场景视图2.png)
![场景视图](./attachments/ElasticSearch 场景视图3.png)

场景视图Json文件![场景视图](./attachments/ElasticSearch 场景视图.json)

## 前置条件

-  已安装 DataKit（[DataKit 安装文档](../01-datakit安装)）

## 配置

### 监控指标采集

进入 DataKit 安装目录下的 `conf.d/db` 目录，复制 `elasticsearch.conf.sample` 并命名为 `elasticsearch.conf`。示例如下：

```python
 [[inputs.elasticsearch]]

  ## specify a list of one or more Elasticsearch servers
  # you can add username and password to your url to use basic authentication:
  # servers = ["http://user:pass@localhost:9200"]
  servers = ["http://localhost:9200"]

  ## Timeout for HTTP requests to the elastic search server(s)
  http_timeout = "5s"

  ## When local is true (the default), the node will read only its own stats.
  ## Set local to false when you want to read the node stats from all nodes
  ## of the cluster.
  local = true

  ## Set cluster_health to true when you want to also obtain cluster health stats
  cluster_health = true

  ## Adjust cluster_health_level when you want to also obtain detailed health stats
  ## The options are
  ##  - indices (default)
  ##  - cluster
  # cluster_health_level = "cluster"

  ## Set cluster_stats to true when you want to also obtain cluster stats.
  cluster_stats = true

  ## Only gather cluster_stats from the master node. To work this require local = true
  cluster_stats_only_from_master = true

  ## Indices to collect; can be one or more indices names or _all
  indices_include = ["_all"]

  ## One of "shards", "cluster", "indices"
  indices_level = "shards"

  ## node_stats is a list of sub-stats that you want to have gathered. Valid options
  ## are "indices", "os", "process", "jvm", "thread_pool", "fs", "transport", "http",
  ## "breaker". Per default, all stats are gathered.
  node_stats = ["jvm", "http","indices","os","process","thread_pool","fs","transport"]

  ## HTTP Basic Authentication username and password.
  # username = ""
  # password = ""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
  
```

   
重新启动datakit生效

`systemctl restart datakit`

### 日志采集

进入 DataKit 安装目录下的 `conf.d/log` 目录，复制 `tailf.conf.sample` 并命名为 `tailf.conf`。示例如下：

```python
[[inputs.tailf]]
    # glob logfiles
    # required
    logfiles = ["/var/log/elasticsearch/solution.log"]

    # glob filteer
    ignore = [""]

    # required
    source = "es_clusterlog"

    # grok pipeline script path
    pipeline = "elasticsearch_cluster_log.p"

    # read file from beginning
    # if from_begin was false, off auto discovery file
    from_beginning = true

    ## characters are replaced using the unicode replacement character
    ## When set to the empty string the data is not decoded to text.
    ## ex: character_encoding = "utf-8"
    ##     character_encoding = "utf-16le"
    ##     character_encoding = "utf-16le"
    ##     character_encoding = "gbk"
    ##     character_encoding = "gb18030"
    ##     character_encoding = ""
    # character_encoding = ""

    ## The pattern should be a regexp
    ## Note the use of '''XXX'''
    # match = '''^\d{4}-\d{2}-\d{2}'''

	#Add Tag for elasticsearch cluster
    [inputs.tailf.tags]
      cluster_name = "solution"

```
Elasticsearch集群信息日志切割grok脚本
![grok脚本](./attachments/elasticsearch_cluster_log.p)

重新启动datakit生效

`systemctl restart datakit`

## 采集指标

配置采集器后默认指标以行协议的方式收集

### 指标集 `elasticsearch_cluster_health`

前置条件：
- `cluster_health = true`

#### 标签

| 标签名  | 描述
|---:     |---------
| `name`  |  

#### 指标

| 指标                               | 描述                                         | 类型                                    | 单位
| ---:                               | :----:                                       | ---                                     | ----
| `active_primary_shards`            | 活动主分区的数量                             | integer                                 | -
| `active_shards`                    | 活动主分区和副本分区的总数                   | integer                                 | -
| `active_shards_percent_as_number`  | 群集中活动碎片的比率，以百分比表示           | float                                   | -
| `delay_unassigned_shards`          | 其分配因超时设置而延迟的分片数               | integer                                 | -
| `initializing_shards`              | 正在初始化的分片数                           | integer                                 | -
| `number_of_data_nodes`             | 作为专用数据节点的节点数                     | integer                                 | -
| `number_of_in_flight_fetch`        | 未完成的访存数量                             | integer                                 | -
| `number_of_nodes`                  | 集群中的节点数                               | integer                                 | -
| `number_of_pending_tasks`          | 尚未执行的集群级别更改的数量                 | integer                                 | -
| `relocating_shards`                | 正在重定位的分片的数量                       | integer                                 | -
| `status`                           | 集群的运行状况，基于其主要和副本分片的状态   | enum("green","yellow","red")            | -
| `status_code`                      | 集群的运行状况，基于其主要和副本分片的状态   | integer, green = 1, yellow = 2, red = 3 | -
| `task_max_waiting_in_queue_millis` | 自最早的初始化任务等待执行以来的时间         | integer                                 | ms
| `timed_out`                        | 如果false响应在timeout参数指定的时间段内返回 | boolean                                 | -
| `unassigned_shards`                | 未分配的分片数                               | boolean                                 | -

status状态说明

| 颜色   | 描述
| ---:   | :----
| green  | 所有分片均已分配
| yellow | 所有主分片均已分配，但一个或多个副本分片未分配。如果群集中的某个节点发生故障，则在修复该节点之前，某些数据可能不可用
| red    | 未分配一个或多个主分片，因此某些数据不可用。在集群启动期间，这可能会短暂发生，因为已分配了主要分片


```golang
package main

import "fmt"

func main() {
	fmt.Println("hello world!")
}
```
