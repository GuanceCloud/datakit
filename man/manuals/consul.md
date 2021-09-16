{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# consul 相关指标采集

consul采集器用于采集consul相关的指标数据，目前只支持Prometheus格式的数据

## 前置条件
- 安装consul-exporter
- 安装Promethues采集器，并配置好consul-exporter的相关信息

## 配置

进入 DataKit 安装目录下的 `conf.d/prom` 目录，复制 `prom.conf.sample` 并命名为 `prom.conf`。
配置如下：
```toml
[[inputs.prom]]
  ## Exporter 地址
  url = "http://127.0.0.1:9107/metrics"

  ## 采集器别名
  source = "consul"

  ## 指标类型过滤, 可选值为 counter, gauge, histogram, summary
  # 默认只采集 counter 和 gauge 类型的指标
  # 如果为空，则不进行过滤
  metric_types = ["counter", "gauge"]

  ## 指标名称过滤
  # 支持正则，可以配置多个，即满足其中之一即可
  # 如果为空，则不进行过滤
  metric_name_filter = ["^consul"]

  ## 指标集名称前缀
  # 配置此项，可以给指标集名称添加前缀
  measurement_prefix = ""

  ## 采集间隔 "ns", "us" (or "µs"), "ms", "s", "m", "h"
  interval = "10s"

  ## 自定义指标集名称
  # 可以将包含前缀prefix的指标归为一类指标集
  # 自定义指标集名称配置优先measurement_name配置项
  [[inputs.prom.measurements]]
  	prefix = "consul_"
	name = "consul"
```

配置好后，重启 DataKit 即可。


## 指标集
|    <div align = center>指标  </div>   | <div align = center>描述 </div> | <div align = center>单位</div>  | 数据类型 |
| :--------------------------: | :----------------------------------------------------------: |:---: |:-----:|
| catalog_service_node_healthy |                   某个服务在结点上是否健康                   |gauge |int|
|       catalog_services       |                      该集群中有多少个服务                     |gauge |int|
|      health_node_status      |  结点的健康检查状态，status有critical, maintenance, passing,warning四种 |gauge |int|
|    health_service_status     |  服务的健康检查状态，status有critical, maintenance, passing,warning四种 |gauge |int|
|         raft_leader          |                    raft集群中有多少个leader                   |gauge |int|
|          raft_peers          |                     raft集群中有多少个peer                    |gauge |int|
|    serf_lan_member_status    |  集群里成员的状态，其中1表示Alive，2表示Leaving，3表示Left，4表示Failed |gauge |int|
|       serf_lan_members       |                       集群中有多少个成员                      |gauge |int|
|        service_checks        |                    服务id和服务名能否对应上                   |gauge |int|


## 日志采集
如需采集consul的日志，需要在启动consul时需要
