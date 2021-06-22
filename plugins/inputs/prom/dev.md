### 简介
prom 采集器可以获取各种Prometheus Exporters的监控数据，用户只要配置相应的Endpoint，就可以将监控数据接入。支持指标过滤、指标集重命名等。

### 配置
```
[[inputs.prom]]
  ## Exporter 地址
  url = "http://127.0.0.1:9100/metrics"

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
  measurement_prefix = "prom_"

  ## 指标集名称
  # 默认会将指标名称以下划线"_"进行切割，切割后的第一个字段作为指标集名称，剩下字段作为当前指标名称
  # 如果配置`measurement_name`, 则不进行指标名称的切割
  # 最终的指标集名称会添加上`measurement_prefix`前缀
  # measurement_name = "prom"

  ## 自定义指标集名称
  # 可以将名称满足指定`pattern`的指标归为一类指标集
  # 自定义指标集名称配置优先`measurement_name`配置项
  #[[inputs.prom.measurements]]
  #  名称匹配, 支持正则
  #  pattern = "cpu"
  #  指标集名称
  #  name = "prom_cpu"

  # [[inputs.prom.measurements]]
  # disable_prefix = 0
  # pattern = "mem"
  # name = "mem"

  ## 采集间隔 "ns", "us" (or "µs"), "ms", "s", "m", "h"
  interval = "10s"
  
  ## TLS 配置
  tls_open = false
  # tls_ca = "/tmp/ca.crt"
  # tls_cert = "/tmp/peer.crt"
  # tls_key = "/tmp/peer.key"

  ## 自定义Tags
  [inputs.prom.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"

```
