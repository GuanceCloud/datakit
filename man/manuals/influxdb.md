{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

InfluxDB 采集器，用于采集 InfluxDB 的数据。

## 前置条件

适用于 InfluxDB v1.x.

注：如需采集 InfluxDB v2.x 版本的数据可通过配置 **prom 采集器** 实现。示例如下：
```toml
[[inputs.prom]]
  ## Exporter 地址
  url = "http://127.0.0.1:8086/metrics"

  metric_types = ["counter", "gauge"]

  interval = "10s"

  ## TLS 配置
  tls_open = false
  # tls_ca = "/tmp/ca.crt"
  # tls_cert = "/tmp/peer.crt"
  # tls_key = "/tmp/peer.key"

  [[inputs.prom.measurements]]
    prefix = "boltdb_"
    name = "influxdb_v2_boltdb"

  [[inputs.prom.measurements]]
    prefix = "go_"
    name = "influxdb_v2_go"
  
  ## histogram 类型
  # [[inputs.prom.measurements]]
  #   prefix = "http_api_request_"
  #   name = "influxdb_v2_http_request"

  [[inputs.prom.measurements]]
    prefix = "influxdb_"
    name = "influxdb_v2"
  
  [[inputs.prom.measurements]]
    prefix = "service_"
    name = "influxdb_v2_service"

  [[inputs.prom.measurements]]
    prefix = "task_"
    name = "influxdb_v2_task" 

  ## 自定义Tags
  [inputs.prom.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"

```

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 指标集

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## 日志采集

如需采集 InfluxDB 的日志，可在 {{.InputName}}.conf 中 将 `files` 打开，并写入 InfluxDB 日志文件的绝对路径。比如：

```toml
[inputs.influxdb.log]
    # 填入绝对路径
    files = ["/path/to/demo.log"] 
    ## grok pipeline script path
    pipeline = "influxdb.p"
```
