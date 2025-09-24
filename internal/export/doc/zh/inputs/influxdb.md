---
title     : 'InfluxDB'
summary   : '采集 InfluxDB 指标数据'
tags:
  - '数据库'
__int_icon      : 'icon/influxdb'
dashboard :
  - desc  : 'InfluxDB'
    path  : 'dashboard/zh/influxdb'
  - desc  : 'InfluxDB v2'
    path  : 'dashboard/zh/influxdb_v2'
monitor   :
  - desc  : 'InfluxDB v2'
    path  : 'monitor/zh/influxdb_v2'
---


{{.AvailableArchs}}

---

InfluxDB 采集器，用于采集 InfluxDB 的数据。

## InfluxDB 采集器配置 {#config}

### 前置条件 {#requirements}

- InfluxDB 采集器，仅适用于 InfluxDB v1.x
- InfluxDB v2.x ，需要使用 prom 采集器进行采集

已测试的版本：

- [x] 1.8.10

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

### InfluxDB v2.x {#prom-config}

```toml
[[inputs.prom]]
  ## Exporter HTTP URL.
  url = "http://127.0.0.1:8086/metrics"

  metric_types = ["counter", "gauge"]

  interval = "10s"

  ## TLS configuration.
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
  
  ## Histogram type.
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

  ## Customize tags.
  [inputs.prom.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"

```

## 指标 {#metric}

以下所有数据采集，默认会追加全局选举 tag，也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}

## 日志 {#logging}

如需采集 InfluxDB 的日志，可在 {{.InputName}}.conf 中 将 `files` 打开，并写入 InfluxDB 日志文件的绝对路径。比如：

```toml
[inputs.{{.InputName}}.log]
    # 填入绝对路径
    files = ["/path/to/demo.log"] 
    ## grok pipeline script path
    pipeline = "influxdb.p"
```
