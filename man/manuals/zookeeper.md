# Zookeeper
---

> 操作系统支持：windows/amd64,windows/386,linux/arm,linux/arm64,linux/386,linux/amd64,darwin/amd64

## 视图预览

![image.png](../imgs/zookeeper-1.png)

## 安装部署

说明：示例 Zookeeper 版本为： zookeeper 3.6.3 (CentOS)，zookeeper 3.6 +的版本会比之前版本多出许多指标，如果您使用的是 3.6 之前的版本可能会存在部分指标采集不到的情况。

### 前置条件

- [安装 Datakit](../datakit/datakit-install.md)

- 开启 Zookeeper中的 Metrics Providers 配置（默认配置文件位置：Zookeeper安装目录/conf/zoo.cfg）并添加 `4lw.commands.whitelist=*`

```
## Metrics Providers

# [https://prometheus.io](https://prometheus.io) Metrics Exporter
metricsProvider.className=org.apache.zookeeper.metrics.prometheus.PrometheusMetricsProvider
metricsProvider.httpPort=7000
metricsProvider.exportJvmInfo=true
4lw.commands.whitelist=*
```

- 重启 Zookeeper 集群应用配置

- 在 Zookeeper 集群中下载安装 [zookeeper_exporter](https://github.com/carlpett/zookeeper_exporter/releases/download/v1.1.0/zookeeper_exporter)， chmod +x 赋予执行权限后启动即可。默认端口为 9141 ,可用通过命令进行验证数据

```bash
[root@d ~]# curl 0.0.0.0:9141/metrics
zk_sync_process_time{quantile="0_5",zk_host="172.16.0.196:2181"} NaN
zk_proposal_latency_sum{zk_host="172.16.0.196:2181"} 0.0
zk_read_final_proc_time_ms{quantile="0_5",zk_host="172.16.0.196:2181"} NaN
zk_process_start_time_seconds{zk_host="172.16.0.23:2181"} 1.645516624379E9
zk_jvm_buffer_pool_capacity_bytes{pool="direct",zk_host="172.16.0.194:2181"} 287560.0
zk_time_waiting_empty_pool_in_commit_processor_read_ms_sum{zk_host="172.16.0.194:2181"} 0.0
zk_stale_requests_dropped{zk_host="172.16.0.23:2181"} 0.0
zk_open_file_descriptor_count{zk_host="172.16.0.196:2181"} 83.0
zk_commit_commit_proc_req_queued_sum{zk_host="172.16.0.194:2181"} 0.0
zk_write_final_proc_time_ms_count{zk_host="172.16.0.194:2181"} 0.0
zk_election_time{quantile="0_5",zk_host="172.16.0.23:2181"} NaN
zk_connection_token_deficit_count{zk_host="172.16.0.23:2181"} 2.0
......
```

### 配置实施


#### 指标采集 (必选)

开启 Datakit Prom 插件，复制 sample 文件

```bash
/usr/local/datakit/conf.d/prom
cp prom.conf.sample prom.conf
```

修改 `prom.conf` 配置文件

```bash
vi prom.conf
```

配置如下：

```toml
[[inputs.prom]]
 ## Exporter URLs
 urls = ["http://127.0.0.1:9141/metrics"]

 ## 忽略对 url 的请求错误
 ignore_req_err = false

 ## 采集器别名
 source = "zookeeper"

 ## 采集数据输出源
 # 配置此项，可以将采集到的数据写到本地文件而不将数据打到中心
 # 之后可以直接用 datakit --prom-conf /path/to/this/conf 命令对本地保存的指标集进行调试
 # 如果已经将 url 配置为本地文件路径，则 --prom-conf 优先调试 output 路径的数据
 # output = "/abs/path/to/file"

 ## 采集数据大小上限，单位为字节
 # 将数据输出到本地文件时，可以设置采集数据大小上限
 # 如果采集数据的大小超过了此上限，则采集的数据将被丢弃
 # 采集数据大小上限默认设置为32MB
 # max_file_size = 0

 ## 指标类型过滤, 可选值为 counter, gauge, histogram, summary
 # 默认只采集 counter 和 gauge 类型的指标
 # 如果为空，则不进行过滤
 metric_types = ["counter", "gauge"]

 ## 指标名称过滤
 # 支持正则，可以配置多个，即满足其中之一即可
 # 如果为空，则不进行过滤
 # metric_name_filter = ["cpu"]

 ## 指标集名称前缀
 # 配置此项，可以给指标集名称添加前缀
 measurement_prefix = ""

 ## 指标集名称
 # 默认会将指标名称以下划线"_"进行切割，切割后的第一个字段作为指标集名称，剩下字段作为当前指标名称
 # 如果配置measurement_name, 则不进行指标名称的切割
 # 最终的指标集名称会添加上measurement_prefix前缀
 # measurement_name = "prom"

 ## 采集间隔 "ns", "us" (or "µs"), "ms", "s", "m", "h"
 interval = "10s"

 ## 过滤tags, 可配置多个tag
 # 匹配的tag将被忽略
 # tags_ignore = ["xxxx"]

 ## TLS 配置
 tls_open = false
 # tls_ca = "/tmp/ca.crt"
 # tls_cert = "/tmp/peer.crt"
 # tls_key = "/tmp/peer.key"

 ## 自定义认证方式，目前仅支持 Bearer Token
 # token 和 token_file: 仅需配置其中一项即可
 # [inputs.prom.auth]
 # type = "bearer_token"
 # token = "xxxxxxxx"
 # token_file = "/tmp/token"

 ## 自定义指标集名称
 # 可以将包含前缀prefix的指标归为一类指标集
 # 自定义指标集名称配置优先measurement_name配置项
 #[[inputs.prom.measurements]]
 #  prefix = "cpu_"
 #  name = "cpu"

 # [[inputs.prom.measurements]]
 # prefix = "mem_"
 # name = "mem"

 ## 自定义Tags
 [inputs.prom.tags]
 #some_tag = "some_value"
 # more_tag = "some_other_value"
 ```

重启 Datakit (如果需要开启日志，请配置日志采集再重启)

```bash
systemctl restart datakit
```

Zookeeper 指标采集验证 `/usr/local/datakit/datakit -M |egrep "最近采集|Zookeeper"`

![image.png](../imgs/zookeeper-2.png)

DQL 验证

```bash
[root@df-solution-ecs-018 prom]# datakit dql
dql > M::zookeeper LIMIT 1
-----------------[ r1.zookeeper.s1 ]-----------------
      add_dead_watcher_stall_time 0
            approximate_data_size 44
                auth_failed_count 0
      avg_close_session_prep_time '1.0'
avg_commit_commit_proc_req_queued '0.0'
          avg_commit_process_time '0.0'
          .....
    sync_processor_request_queued 2
                    throttled_ops 0
                             time 2022-02-22 16:00:10 +0800 CST
           tls_handshake_exceeded 0
        unrecoverable_error_count 0
           unsuccessful_handshake 0
                           uptime 2858680025
                          version '3.7.0-e3704b390a6697bfdf4b0bef79e3da7a4f6bac4b'
                        version_1 <nil>
                      watch_bytes 0
                      watch_count 0
                      znode_count 5
---------
1 rows, 1 series, cost 40.297037ms

```

指标预览

![image.png](../imgs/zookeeper-3.png)

#### 插件标签 (非必选)

参数说明

- 该配置为自定义标签，可以填写任意 key-value 值
- 以下示例配置完成后，所有 Flink 指标都会带有 service = "zookeeper" 的标签，可以进行快速查询

```toml
# 示例
[inputs.prom.tags]
  service = "zookeeper"
```

重启 Datakit

```shell
systemctl restart datakit
```

## 场景视图

场景 - 新建场景 - Zookeeper 监控视图

## 异常检测

异常检测库 - 新建检测库 - Zookeeper 检测库 
 

| 序号 | 规则名称                   | 触发条件                    | 级别 | 检测频率 |
| ---  | ---                        | ---                         | ---  | ---      |
| 1    | Zookeeper 堆积请求数过大   | Zookeeper 堆积请求数 > 10   | 紧急 | 1m       |
| 2    | Zookeeper 平均响应延迟过高 | Zookeeper 平均响应延迟 > 20 | 紧急 | 1m       |
| 3    | Zookeeper 服务器宕机       | Zookeeper  运行时间 = 0     | 紧急 | 1m       |

相关文档 <[DataFlux 内置检测库](https://www.yuque.com/dataflux/doc/br0rm2)>

## 指标详解

| **名称**                                                              | **描述**                                                                     | **指标类型** | **Availability**      |
| ---                                                                   | ---                                                                          | ---          | ---                   |
| `zookeeper_approximate_data_size`                                     |                                                                              | 度量         |                       |
| `zookeeper_avg_latency`                                               | 服务器响应客户端请求所需的时间。(ms)                                         | 度量         |                       |
| `zookeeper_bytes_received`                                            | 接收的字节数                                                                 | 度量         |                       |
| `zookeeper_bytes_sent`                                                | 发送的字节数                                                                 | 度量         |                       |
| `zookeeper_connections`                                               | 客户端连接的总数。(连接数)                                                   | 度量         |                       |
| `zookeeper_ephemerals_count`                                          |                                                                              | 度量         |                       |
| `zookeeper_instances`                                                 |                                                                              | 度量         |                       |
| `zookeeper_latency.avg`                                               | 服务器响应客户端请求所需的时间。（ms）                                       | 度量         |                       |
| `zookeeper_latency.max`                                               | 服务器响应客户端请求所需的时间。（ms）                                       | 度量         |                       |
| `zookeeper_latency.min`                                               | 服务器响应客户端请求所需的时间。（ms）                                       | 度量         |                       |
| `zookeeper_max_file_descriptor_count`                                 |                                                                              | 度量         |                       |
| `zookeeper_max_latency`                                               | 服务器响应客户端请求所需的时间。（ms）                                       | 度量         |                       |
| `zookeeper_min_latency`                                               | 服务器响应客户端请求所需的时间。（ms）                                       | 度量         |                       |
| `zookeeper_nodes`                                                     | ZooKeeper 命名空间（数据）中的 znode 数量。（节点数）                        | 度量         |                       |
| `zookeeper_num_alive_connections`                                     | 客户端连接的总数。（连接数）                                                 | 度量         |                       |
| `zookeeper_open_file_descriptor_count`                                |                                                                              | 度量         |                       |
| `zookeeper_outstanding_requests`                                      | 当服务器负载不足并且接收的持续请求数超过其处理能力时排队的请求数。（请求数） | 度量         |                       |
| `zookeeper_packets.received`                                          | 收到的数据包数。                                                             | 度量         |                       |
| `zookeeper_packets.sent`                                              | 发送的数据包数。                                                             | 度量         |                       |
| `zookeeper_packets_received`                                          | 收到的数据包数。                                                             | 度量         |                       |
| `zookeeper_packets_sent`                                              | 发送的数据包数。                                                             | 度量         |                       |
| `zookeeper_server_state`                                              |                                                                              | 度量         |                       |
| `zookeeper_watch_count`                                               |                                                                              | 度量         |                       |
| `zookeeper_znode_count`                                               | ZooKeeper 命名空间（数据）中的 znode 数量。                                  | 度量         |                       |
| `zookeeper_zxid.count`                                                |                                                                              | 度量         |                       |
| `zookeeper_zxid.epoch`                                                |                                                                              | 度量         |                       |
| `zookeeper_add_dead_watcher_stall_time`                               |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_bytes_received_count`                                      | 接收到的字节数。（byte）                                                     | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_close_session_prep_time`                                   | close_session_prep_time 的直方图                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_close_session_prep_time_count`                             | close_session_prep_time 的总计数                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_close_session_prep_time_sum`                               | close_session_prep_time 的总和                                               | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_commit_commit_proc_req_queued`                             | commit_commit_proc_req_queued 的直方图                                       | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_commit_commit_proc_req_queued_count`                       | commit_commit_proc_req_queued 的总数                                         | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_commit_commit_proc_req_queued_sum`                         | commit_commit_proc_req_queued 的总和                                         | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_commit_count`                                              | 在leader上执行的提交次数                                                     | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_commit_process_time`                                       | commit_process_time 的直方图                                                 | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_commit_process_time_count`                                 | commit_process_time 的总计数                                                 | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_commit_process_time_sum`                                   | commit_process_time 的总和                                                   | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_commit_propagation_latency`                                | commit_propagation_latency 的直方图                                          | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_commit_propagation_latency_count`                          | commit_propagation_latency 的总数                                            | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_commit_propagation_latency_sum`                            | commit_propagation_latency 的总和                                            | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_concurrent_request_processing_in_commit_processor`         | concurrent_request_processing_in_commit_processor 的直方图                   | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_concurrent_request_processing_in_commit_processor_count`   | concurrent_request_processing_in_commit_processor 的总数                     | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_concurrent_request_processing_in_commit_processor_sum`     | concurrent_request_processing_in_commit_processor 的总和                     | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_connection_drop_count`                                     | 连接断开计数                                                                 | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_connection_drop_probability`                               | 连接掉线概率                                                                 | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_connection_rejected`                                       | 连接被拒绝计数                                                               | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_connection_request_count`                                  | 传入客户端连接请求数                                                         | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_connection_revalidate_count`                               | 连接重新验证计数                                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_connection_token_deficit`                                  | connection_token_deficit 的直方图                                            | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_connection_token_deficit_count`                            | connection_token_deficit 总数                                                | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_connection_token_deficit_sum`                              | connection_token_deficit 的总和                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_dbinittime`                                                | 重新加载数据库的时间                                                         | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_dbinittime_count`                                          | Time to reload database                                                      | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_dbinittime_sum`                                            | 重新加载数据库的时间                                                         | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_dead_watchers_cleaner_latency`                             | dead_watchers_cleaner_latency 的直方图                                       | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_dead_watchers_cleaner_latency_count`                       | dead_watchers_cleaner_latency 的总数                                         | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_dead_watchers_cleaner_latency_sum`                         | dead_watchers_cleaner_latency 的总和                                         | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_dead_watchers_cleared`                                     |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_dead_watchers_queued`                                      |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_diff_count`                                                | 执行的差异同步数                                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_digest_mismatches_count`                                   |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_election_time`                                             | 进入和离开选举之间的时间                                                     | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_election_time_count`                                       | 进入和离开选举之间的时间                                                     | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_election_time_sum`                                         | 进入和离开选举之间的时间                                                     | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_ensemble_auth_fail`                                        |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_ensemble_auth_skip`                                        |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_ensemble_auth_success`                                     |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_follower_sync_time`                                        | follower与leader同步的时间                                                   | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_follower_sync_time_count`                                  | follower与leader同步的时间count                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_follower_sync_time_sum`                                    | follower与leader同步的时间sum                                                | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_fsynctime`                                                 | fsync事务日志的时间                                                          | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_fsynctime_count`                                           | fsync事务日志的时间                                                          | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_fsynctime_sum`                                             | fsync事务日志的时间                                                          | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_global_sessions`                                           | 全局会话计数                                                                 | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_buffer_pool_capacity_bytes`                            |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_buffer_pool_used_buffers`                              |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_buffer_pool_used_bytes`                                |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_classes_loaded`                                        |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_classes_loaded_total`                                  |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_classes_unloaded_total`                                |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_gc_collection_seconds_count`                           |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_gc_collection_seconds_sum`                             |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_info`                                                  |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_memory_bytes_committed`                                |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_memory_bytes_init`                                     |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_memory_bytes_max`                                      |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_memory_bytes_used`                                     |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_memory_pool_allocated_bytes_total`                     |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_memory_pool_bytes_committed`                           |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_memory_pool_bytes_init`                                | （byte）                                                                     | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_memory_pool_bytes_max`                                 | （byte）                                                                     | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_memory_pool_bytes_used`                                | （byte）                                                                     | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_threads_current`                                       |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_threads_daemon`                                        |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_threads_deadlocked`                                    |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_threads_deadlocked_monitor`                            |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_threads_peak`                                          |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_threads_started_total`                                 |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_jvm_threads_state`                                         |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_large_requests_rejected`                                   |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_last_client_response_size`                                 |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_learner_commit_received_count`                             |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_learner_proposal_received_count`                           |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_local_sessions`                                            | 本地会话计数                                                                 | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_local_write_committed_time_ms`                             | local_write_committed_time_ms 的直方图                                       | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_local_write_committed_time_ms_count`                       | local_write_committed_time_ms 的总计数                                       | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_local_write_committed_time_ms_sum`                         | local_write_committed_time_ms 的总和                                         | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_looking_count`                                             | 转换为查看状态的次数                                                         | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_max_client_response_size**<br `                            |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_min_client_response_size`                                  |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_netty_queued_buffer_capacity`                              | netty_queued_buffer_capacity 的直方图                                        | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_netty_queued_buffer_capacity_count`                        | netty_queued_buffer_capacity 的总数                                          | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_netty_queued_buffer_capacity_sum`                          | netty_queued_buffer_capacity 的总和                                          | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_node_changed_watch_count`                                  | node_changed_watch_count 的直方图                                            | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_node_changed_watch_count_count`                            | node_changed_watch_count 的总数                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_node_changed_watch_count_sum`                              | node_changed_watch_count 的总和                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_node_children_watch_count`                                 | node_children_watch_count 的直方图                                           | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_node_children_watch_count_count`                           | node_children_watch_count 的总数                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_node_children_watch_count_sum`                             | node_children_watch_count 的总和                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_node_created_watch_count`                                  | node_created_watch_count 的直方图                                            | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_node_created_watch_count_count`                            | node_created_watch_count 的总数                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_node_created_watch_count_sum`                              | node_created_watch_count 的总和                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_node_deleted_watch_count`                                  | node_deleted_watch_count 的直方图                                            | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_node_deleted_watch_count_count`                            | node_deleted_watch_count 的总数                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_node_deleted_watch_count_sum`                              | node_deleted_watch_count 的总和                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_om_commit_process_time_ms`                                 | om_commit_process_time_ms 的直方图                                           | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_om_commit_process_time_ms_count`                           | om_commit_process_time_ms 的总数                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_om_commit_process_time_ms_sum`                             | om_commit_process_time_ms 的总和                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_om_proposal_process_time_ms`                               | om_proposal_process_time_ms 的直方图                                         | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_om_proposal_process_time_ms_count`                         | om_proposal_process_time_ms 的总数                                           | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_om_proposal_process_time_ms_sum`                           | om_proposal_process_time_ms 的总和                                           | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_outstanding_changes_queued`                                |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_outstanding_changes_removed`                               |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_outstanding_tls_handshake`                                 |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_pending_session_queue_size`                                | pending_session_queue_size 的直方图                                          | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_pending_session_queue_size_count`                          | pending_session_queue_size 的总数                                            | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_pending_session_queue_size_sum`                            | pending_session_queue_size 的总和                                            | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_prep_process_time`                                         | prep_process_time 的直方图                                                   | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_prep_process_time_count`                                   | prep_process_time 的总计数                                                   | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_prep_process_time_sum`                                     | prep_process_time 的总和                                                     | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_prep_processor_queue_size`                                 | prep_processor_queue_size 的直方图                                           | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_prep_processor_queue_size_count`                           | prep_processor_queue_size 的总数                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_prep_processor_queue_size_sum`                             | prep_processor_queue_size 的总和                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_prep_processor_queue_time_ms`                              | prep_processor_queue_time_ms 的直方图                                        | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_prep_processor_queue_time_ms_count`                        | prep_processor_queue_time_ms 的总数                                          | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_prep_processor_queue_time_ms_sum`                          | prep_processor_queue_time_ms 的总和                                          | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_prep_processor_request_queued`                             |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_process_cpu_seconds_total`                                 |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_process_max_fds`                                           |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_process_open_fds`                                          |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_process_resident_memory_bytes`                             | (byte)                                                                       | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_process_start_time_seconds`                                |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_process_virtual_memory_bytes`                              | (byte)                                                                       | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_propagation_latency`                                       | 更新的端到端延迟，从leader提议到给定主机上的提交数据树                       | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_propagation_latency_count`                                 | 更新的端到端延迟，从leader提议到给定主机上的提交数据树                       | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_propagation_latency_sum`                                   | 更新的端到端延迟，从leader提议到给定主机上的提交数据树                       | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_proposal_ack_creation_latency`                             | proposal_ack_creation_latency 的直方图                                       | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_proposal_ack_creation_latency_count`                       | proposal_ack_creation_latency 的总数                                         | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_proposal_ack_creation_latency_sum`                         | proposal_ack_creation_latency 的总和                                         | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_proposal_count`                                            |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_proposal_latency`                                          | proposal_latency 的直方图                                                    | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_proposal_latency_count`                                    | proposal_latency 的总数                                                      | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_proposal_latency_sum`                                      | proposal_latency 的总和                                                      | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_quit_leading_due_to_disloyal_voter`                        |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_quorum_ack_latency`                                        | quorum_ack_latency 的直方图                                                  | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_quorum_ack_latency_count`                                  | quorum_ack_latency 的总数                                                    | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_quorum_ack_latency_sum`                                    | quorum_ack_latency 的总和                                                    | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_read_commit_proc_issued`                                   | read_commit_proc_issued 的直方图                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_read_commit_proc_issued_count`                             | read_commit_proc_issued 的总数                                               | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_read_commit_proc_issued_sum`                               | read_commit_proc_issued 的总和                                               | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_read_commit_proc_req_queued`                               | read_commit_proc_req_queued 的直方图                                         | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_read_commit_proc_req_queued_count`                         | read_commit_proc_req_queued 的总数                                           | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_read_commit_proc_req_queued_sum`                           | read_commit_proc_req_queued 的总和                                           | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_read_commitproc_time_ms`                                   | read_commitproc_time_ms 的直方图                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_read_commitproc_time_ms_count`                             | read_commitproc_time_ms 的总计数                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_read_commitproc_time_ms_sum`                               | read_commitproc_time_ms 的总和                                               | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_read_final_proc_time_ms`                                   | read_final_proc_time_ms 的直方图                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_read_final_proc_time_ms_count`                             | read_final_proc_time_ms 的总计数                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_read_final_proc_time_ms_sum`                               | read_final_proc_time_ms 的总和                                               | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_readlatency`                                               | readlatency 直方图 读取请求延迟                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_readlatency_count`                                         | readlatency总计数读取请求延迟                                                | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_readlatency_sum`                                           | readlatency 总和 读取请求延迟                                                | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_reads_after_write_in_session_queue`                        | reads_after_write_in_session_queue 的直方图                                  | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_reads_after_write_in_session_queue_count`                  | reads_after_write_in_session_queue 的总数                                    | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_reads_after_write_in_session_queue_sum`                    | reads_after_write_in_session_queue 的总和                                    | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_reads_issued_from_session_queue`                           | reads_issued_from_session_queue 的直方图                                     | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_reads_issued_from_session_queue_count`                     | reads_issued_from_session_queue 的总数                                       | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_reads_issued_from_session_queue_sum`                       | reads_issued_from_session_queue 的总和                                       | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_request_commit_queued`                                     | 排队的请求提交计数                                                           | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_request_throttle_wait_count`                               |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_requests_in_session_queue`                                 | requests_in_session_queue 的直方图                                           | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_requests_in_session_queue_count`                           | requests_in_session_queue 的总数                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_requests_in_session_queue_sum`                             | requests_in_session_queue 的总和                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_response_packet_cache_hits`                                |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_response_packet_cache_misses`                              |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_response_packet_get_children_cache_hits`                   |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_response_packet_get_children_cache_misses`                 |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_revalidate_count`                                          |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_server_write_committed_time_ms`                            | server_write_committed_time_ms 的直方图                                      | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_server_write_committed_time_ms_count`                      | server_write_committed_time_ms 的总计数                                      | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_server_write_committed_time_ms_sum`                        | server_write_committed_time_ms 的总和                                        | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_session_queues_drained`                                    | session_queues_drained 的直方图                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_session_queues_drained_count`                              | session_queues_drained 的总数                                                | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_session_queues_drained_sum`                                | session_queues_drained 的总和                                                | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_sessionless_connections_expired`                           |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_snap_count`                                                | 执行的快照同步次数                                                           | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_snapshottime`                                              | snapshottime直方图写快照的时间                                               | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_snapshottime_count`                                        | 快照时间总数 写入快照的时间                                                  | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_snapshottime_sum`                                          | 快照时间总和 写入快照的时间                                                  | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_stale_replies`                                             |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_stale_requests`                                            |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_stale_requests_dropped`                                    |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_stale_sessions_expired`                                    |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_startup_snap_load_time`                                    | startup_snap_load_time 的直方图                                              |              |                       |
| `zookeeper_startup_snap_load_time_count`                              | startup_snap_load_time 的总计数                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_startup_snap_load_time_sum`                                | startup_snap_load_time的总和                                                 | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_startup_txns_load_time`                                    | startup_txns_load_time 的直方图                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_startup_txns_load_time_count`                              | startup_txns_load_time 的总计数                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_startup_txns_load_time_sum`                                | startup_txns_load_time的总和                                                 | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_startup_txns_loaded`                                       | startup_txns_loaded 的直方图                                                 | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_startup_txns_loaded_count`                                 | startup_txns_loaded 的总数                                                   | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_startup_txns_loaded_sum`                                   | startup_txns_loaded 的总和                                                   | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_sync_process_time`                                         | sync_process_time 的直方图                                                   | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_sync_process_time_count`                                   | sync_process_time 的总数                                                     | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_sync_process_time_sum`                                     | sync_process_time 的总和                                                     | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_sync_processor_batch_size`                                 | sync_processor_batch_size 的直方图                                           | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_sync_processor_batch_size_count`                           | sync_processor_batch_size 的总数                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_sync_processor_batch_size_sum`                             | sync_processor_batch_size 的总和                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_sync_processor_queue_and_flush_time_ms`                    | sync_processor_queue_and_flush_time_ms 的直方图                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_sync_processor_queue_and_flush_time_ms_count`              | sync_processor_queue_and_flush_time_ms 的总数                                | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_sync_processor_queue_and_flush_time_ms_sum`                | sync_processor_queue_and_flush_time_ms 的总和                                | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_sync_processor_queue_flush_time_ms`                        | sync_processor_queue_flush_time_ms 的直方图                                  | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_sync_processor_queue_flush_time_ms_count`                  | sync_processor_queue_flush_time_ms 的总数                                    | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_sync_processor_queue_flush_time_ms_sum`                    | sync_processor_queue_flush_time_ms 的总和                                    | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_sync_processor_queue_size`                                 | sync_processor_queue_size 的直方图                                           | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_sync_processor_queue_size_count`                           | sync_processor_queue_size 的总数                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_sync_processor_queue_size_sum`                             | sync_processor_queue_size 的总和                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_sync_processor_queue_time_ms`                              | sync_processor_queue_time_ms 的直方图                                        | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_sync_processor_queue_time_ms_count`                        | sync_processor_queue_time_ms 的总数                                          | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_sync_processor_queue_time_ms_sum`                          | sync_processor_queue_time_ms 的总和                                          | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_sync_processor_request_queued`                             |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_time_waiting_empty_pool_in_commit_processor_read_ms`       | time_waiting_empty_pool_in_commit_processor_read_ms 的直方图                 | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_time_waiting_empty_pool_in_commit_processor_read_ms_count` | time_waiting_empty_pool_in_commit_processor_read_ms 的总计数                 | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_time_waiting_empty_pool_in_commit_processor_read_ms_sum`   | time_waiting_empty_pool_in_commit_processor_read_ms 的总和                   | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_tls_handshake_exceeded`                                    |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_unrecoverable_error_count`                                 |                                                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_updatelatency`                                             | updatelatency 直方图 更新请求延迟                                            | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_updatelatency_count`                                       | updatelatency 更新请求延迟的总数                                             | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_updatelatency_sum`                                         | updatelatency总和 更新请求延迟                                               | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_uptime`                                                    | 对等点处于表领先/跟随/观察状态的正常运行时间                                 | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_write_batch_time_in_commit_processor`                      | write_batch_time_in_commit_processor 的直方图                                | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_write_batch_time_in_commit_processor_count`                | write_batch_time_in_commit_processor 的总数                                  | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_write_batch_time_in_commit_processor_sum`                  | write_batch_time_in_commit_processor 的总和                                  | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_write_commit_proc_issued`                                  | write_commit_proc_issued 的直方图                                            | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_write_commit_proc_issued_count`                            | write_commit_proc_issued 的总数                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_write_commit_proc_issued_sum`                              | write_commit_proc_issued 的总和                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_write_commit_proc_req_queued`                              | write_commit_proc_req_queued 的直方图                                        | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_write_commit_proc_req_queued_count`                        | write_commit_proc_req_queued 的总数                                          | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_write_commit_proc_req_queued_sum`                          | write_commit_proc_req_queued 的总和                                          | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_write_commitproc_time_ms`                                  | write_commitproc_time_ms 的直方图                                            | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_write_commitproc_time_ms_count`                            | write_commitproc_time_ms 的总计数                                            | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_write_commitproc_time_ms_sum`                              | write_commitproc_time_ms 的总和                                              | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_write_final_proc_time_ms`                                  | write_final_proc_time_ms 的直方图                                            | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_write_final_proc_time_ms_count`                            | write_final_proc_time_ms 的总计数                                            | 度量         | 适用于 Zookeeper 3.6+ |
| `zookeeper_write_final_proc_time_ms_sum`                              | write_final_proc_time_ms 的总和                                              | 度量         | 适用于 Zookeeper 3.6+ |

以下指标仍会发送，但已经弃用：

- `zookeeper.bytes_received`
- `zookeeper.bytes_sent`

## 最佳实践

<暂无>

## 故障排查

- [无数据上报排查](../datakit/why-no-data.md)

## 进一步阅读

- [DataFlux Tag 应用最佳实践](/best-practices/guance-skill/tag.md)
