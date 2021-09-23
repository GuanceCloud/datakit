{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# etcd

etcd 采集器可以从 etcd 实例中采取很多指标，比如etcd服务器状态和网络的状态等多种指标，并将指标采集到 DataFlux ，帮助你监控分析 etcd 各种异常情况

## 前置条件

- etcd 版本  >=3

- 开启etcd，默认的metrics接口是http://localhost:2379/metrics，也可以自己去配置文件中修改。

## 配置

进入 DataKit 安装目录下的 `conf.d/prom` 目录，复制如下示例 并命名为 `prom.conf`。示例如下：

```toml
[[inputs.prom]]
  ## Exporter 地址
  url = "http://127.0.0.1:2379/metrics"

	## 采集器别名
	source = "prom"

  ## 指标类型过滤, 可选值为 counter, gauge, histogram, summary
  # 默认只采集 counter 和 gauge 类型的指标
  # 如果为空，则不进行过滤
  metric_types = ["counter", "gauge"]

  ## 指标名称过滤
  # 支持正则，可以配置多个，即满足其中之一即可
  # 如果为空，则不进行过滤
  metric_name_filter = ["^etcd_server","^etcd_network"]

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

  ## 自定义指标集名称
  # 可以将包含前缀prefix的指标归为一类指标集
  # 自定义指标集名称配置优先measurement_name配置项
  [[inputs.prom.measurements]]
    prefix = "etcd_"
    name = "etcd"

  ## 自定义认证方式，目前仅支持 Bearer Token
  # [inputs.prom.auth]
  # type = "bearer_token"
  # token = "xxxxxxxx"
  # token_file = "/tmp/token"

  ## 自定义Tags
```

配置好后，重启 DataKit 即可。

## 指标集

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}