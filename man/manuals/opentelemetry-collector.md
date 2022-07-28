# Opentelemetry Collector
---

## 视图预览
Opentelemetry Collector 性能指标展示：collector 在线时长、内存使用情况、exporter 相关指标、receiver 相关指标 等。

![](../imgs/opentelmetry-collector-1.png)

## 版本支持

操作系统：Linux / Windows<br />Opentelemetry Collector 版本：>=0.46.0

## 前置条件

- Opentelemetry Collector 服务器 <[安装 Datakit](../datakit/datakit-install.md)>

## 安装配置

说明：示例 Opentelemetry Collector 版本为 Linux 环境 docker 方式安装，各个不同版本指标可能存在差异。

### 部署实施

(Linux / Windows 环境相同)

#### 指标采集 (必选)
DataKit  有两种方案支持 otel-collector 指标采集，两种方案采集结果一致。

> 方案一 ：通过 prom 采集 Opentelemetry Collector 指标
>
> 方案二：通过 Opentelemetry  采集器采集 Opentelemetry Collector 指标


##### 方案一 ：通过 prom 采集 Opentelemetry Collector 指标

1. 开启 Opentelemetry Collector 指标端口，默认端口为：8888
```yaml
version: '3.3'

services:
    # Collector
    otel-collector:
        image: otel/opentelemetry-collector-contrib:0.46.0
        command: ["--config=/etc/otel-collector-config.yaml"]
        volumes:
            - ./otel-collector-config.yaml:/etc/otel-collector-config.yaml
        ports:
            - "1888:1888"   # pprof extension
            - "8888:8888"   # Prometheus metrics exposed by the collector
            - "8889:8889"   # Prometheus exporter metrics
            - "13133:13133" # health_check extension
            - "4350:4317"        # OTLP gRPC receiver
            - "55670:55679" # zpages extension
            - "4318:4318"

```

2. 访问 Opentelemetry Collector 指标 ， curl http://otel-collector-host:8888/metrics。

![image.png](../imgs/opentelmetry-collector-2.png)

3. 开启 Datakit prom 插件，复制 sample 文件
```
cd /usr/local/datakit/conf.d/prom/
cp prom.conf.sample prom-otelcol.conf
```

4. 修改 prom-otelcol.conf 配置文件

主要参数说明

- url：otel-collector 指标地址
- interval：采集频率
- source : 指标器别名
- response_timeout：响应超时时间 (默认5秒)
```toml

[[inputs.prom]]
  ## Exporter URLs
  urls = ["http://127.0.0.1:8888/metrics"]

  ## 忽略对 url 的请求错误
  ignore_req_err = false

  ## 采集器别名
  source = "otel-prom"

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

  ## 指标类型过滤, 可选值为 counter, gauge, histogram, summary, untyped
  # 默认只采集 counter 和 gauge 类型的指标
  # 如果为空，则不进行过滤
  metric_types = []

  ## 指标名称筛选：符合条件的指标将被保留下来
  # 支持正则，可以配置多个，即满足其中之一即可
  # 如果为空，则不进行筛选，所有指标均保留
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

  ## 过滤 tags, 可配置多个tag
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
  # 可以将包含前缀 prefix 的指标归为一类指标集
  # 自定义指标集名称配置优先 measurement_name 配置项
  #[[inputs.prom.measurements]]
  #  prefix = "cpu_"
  #  name = "cpu"

  # [[inputs.prom.measurements]]
  # prefix = "mem_"
  # name = "mem"

  ## 重命名 prom 数据中的 tag key
	[inputs.prom.tags_rename]
		overwrite_exist_tags = false
		[inputs.prom.tags_rename.mapping]
			# tag1 = "new-name-1"
			# tag2 = "new-name-2"
			# tag3 = "new-name-3"

  ## 自定义Tags
  [inputs.prom.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"

```

5. 重启 Datakit (如果需要开启日志，请配置日志采集再重启)
```
systemctl restart datakit
```

6. Opentelemetry Collector  指标采集验证，使用命令 /usr/local/datakit/datakit -M |egrep "最近采集|otel"

![image.png](../imgs/opentelmetry-collector-3.png)

##### 方案二：通过 Opentelemetry  采集器采集 Opentelemetry Collector 指标

1. collector 新增 otlp exporter。
```toml
receivers:
  otlp:
    protocols:
      grpc:
      http:
        cors:
          allowed_origins:
            - http://*
            - https://*
exporters:
  otlp:
    endpoint: "http://192.168.91.11:4319"
    tls:
      insecure: true
    compression: none # 不开启gzip

processors:
  batch:

extensions:
  health_check:
  pprof:
    endpoint: :1888
  zpages:
    endpoint: :55679

service:
  extensions: [pprof, zpages, health_check]
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
```
参数说明<br />otlp.endpoint：配置datakit opentelemetry grpc地址

2. 开启 prom 插件，复制 Sample 文件
```shell
cd /usr/local/datakit/conf.d/opentelemetry
cp opentelemetry.conf.sample opentelemetry.conf
```

3. 修改 opentelemetry.conf
```toml
[[inputs.opentelemetry]]
  ## 在创建'trace',Span','resource'时，会加入很多标签，这些标签最终都会出现在'Span'中
  ## 当您不希望这些标签太多造成网络上不必要的流量损失时，可选择忽略掉这些标签
  ## 支持正则表达，注意:将所有的'.'替换成'_'
  ## When creating 'trace', 'span' and 'resource', many labels will be added, and these labels will eventually appear in all 'spans'
  ## When you don't want too many labels to cause unnecessary traffic loss on the network, you can choose to ignore these labels
  ## Support regular expression. Note!!!: all '.' Replace with '_'
  # ignore_attribute_keys = ["os_*","process_*"]

  ## Keep rare tracing resources list switch.
  ## If some resources are rare enough(not presend in 1 hour), those resource will always send
  ## to data center and do not consider samplers and filters.
  # keep_rare_resource = false

  ## Ignore tracing resources map like service:[resources...].
  ## The service name is the full service name in current application.
  ## The resource list is regular expressions uses to block resource names.
  # [inputs.opentelemetry.close_resource]
    # service1 = ["resource1", "resource2", ...]
    # service2 = ["resource1", "resource2", ...]
    # ...

  ## Sampler config uses to set global sampling strategy.
  ## priority uses to set tracing data propagation level, the valid values are -1, 0, 1
  ##   -1: always reject any tracing data send to datakit
  ##    0: accept tracing data and calculate with sampling_rate
  ##    1: always send to data center and do not consider sampling_rate
  ## sampling_rate used to set global sampling rate
  # [inputs.opentelemetry.sampler]
    # priority = 0
    # sampling_rate = 1.0

  # [inputs.opentelemetry.tags]
    # key1 = "value1"
    # key2 = "value2"
    # ...

  [inputs.opentelemetry.expectedHeaders]
    ## 如有header配置 则请求中必须要携带 否则返回状态码500
	## 可作为安全检测使用,必须全部小写
	# ex_version = xxx
	# ex_name = xxx
	# ...

  ## grpc
  [inputs.opentelemetry.grpc]
  ## trace for grpc
  trace_enable = true

  ## metric for grpc
  metric_enable = true

  ## grpc listen addr
  # addr = "127.0.0.1:4317"
  addr = "0.0.0.0:4319"

  ## http
  [inputs.opentelemetry.http]
  ## if enable=true
  ## http path (do not edit):
  ##	trace : /otel/v1/trace
  ##	metric: /otel/v1/metric
  ## use as : http://127.0.0.1:9529/otel/v11/trace . Method = POST
  enable = true
  ## return to client status_ok_code :200/202
  http_status_ok = 200

```
参数说明

- trace_enable：true 		#开启grpc trace

- metric_enable： true 	    #开启grpc metric
- addr: 0.0.0.0:4319 		    #开启端口

4. 重启 DataKit
```shell
datakit --restart
```


#### 插件标签 (非必选）
参数说明

- 该配置为自定义标签，可以填写任意 key-value 值

- 以下示例配置完成后，所有 Opentelemetry Collector 指标都会带有 env= dev 的标签，可以进行快速查询
- 相关文档 <[DataFlux Tag 应用最佳实践](/best-practices/guance-skill/tag/)>

```
# 示例
[inputs.prom.tags]
   env= dev 
```
重启datakit
```
systemctl restart datakit
```
## 场景视图
<场景 - 新建仪表板 - 内置模板库 - Opentelemetry Collector 监控视图>

## 指标详解

| 指标 | 描述 |
| --- | --- |
| process_uptime | 在线时长 |
| process_memory_rss | 内存使用 |
| exporter_sent_log_records | exporter 发送 log 记录数 |
| exporter_sent_metric_points | exporter 发送 metric 记录数 |
| exporter_sent_spans | exporter 发送 span 记录数 |
| receiver_accepted_log_records | reveiver  接收 log 记录数 |
| receiver_accepted_metric_points | reveiver  接收 metric 记录数 |
| receiver_accepted_spans | reveiver  接收 span 记录数 |

## 常见问题排查
- [无数据上报排查](../datakit/why-no-data.md)
## 进一步阅读
- [**OpenTelemetry 链路数据接入最佳实践**](/best-practices/integrations/opentelemetry)

- [Opentelemetry to 观测云](/best-practices/monitor/opentelemetry-guance)
