---
title     : 'Apache IoTDB'
summary   : '采集 Apache IoTDB 服务端指标'
tags:
  - '数据库'
__int_icon      : 'icon/iotdb'
---

{{.AvailableArchs}}

---

IoTDB 采集器通过 IoTDB 的 Prometheus Reporter 读取服务端运行指标，只采集监控数据，不向 IoTDB 写入业务数据。

## 配置 {#config}

在 IoTDB 中启用 Prometheus 指标 Reporter，并确保 DataKit 可以访问指标导出端口，然后配置采集地址：

```toml
[[inputs.iotdb]]
  endpoint = "http://127.0.0.1:9091/metrics"
  interval = "30s"
  timeout = "10s"
  election = true

  [inputs.iotdb.tags]
  # cluster_name = "testing"
```

`endpoint` 支持 HTTP 或 HTTPS URL；只填写主机地址时默认访问 `/metrics`。如在指标端点前部署了需要认证的代理，可配置 `username` 和 `password` 发送 HTTP Basic Auth。

## 指标 {#metric}

所有样本以 `iotdb` measurement 输出，Prometheus 指标名、指标类型和原始标签分别作为 `metric_name`、`metric_type` 和标签字段，数值存放在 `value` 字段。Histogram 的桶、总和和计数按标准 `_bucket`、`_sum`、`_count` 名称展开；Summary 分位数使用 `quantile` 标签。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.DescZh}}

{{$m.MarkdownTable}}

{{ end }}
