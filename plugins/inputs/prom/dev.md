### 简介
prom 采集器可以获取各种Prometheus Exporters的监控数据，用户只要配置相应的Endpoint，就可以将监控数据接入。支持指标过滤、指标集重命名等。

### 配置
```
[[inputs.prom]]
  # Exporter URLs.
  url = "http://127.0.0.1:9100/metrics"

  ## Metrics type whitelist. Optional: counter, gauge, histogram, summary
  # Default only collect 'counter' and 'gauge'.
  # Collect all if empty.
  metric_types = ["counter", "gauge"]

  ## Metrics name whitelist.
  # Regex supported. Multi supported, conditions met when one matched.
  # Collect all if empty.
  # metric_name_filter = ["cpu"]

  ## Measurement prefix.
  # Add prefix to measurement set name.
  measurement_prefix = "prom_"

  ## Measurement name.
  # If measurement_name is empty, split metric name by '_', the first field after split as measurement set name, the rest as current metric name.
  # If measurement_name is not empty, using this as measurement set name.
  # Always add 'measurement_prefix' prefix at last.
  # measurement_name = "prom"

  ## Customize measurement set name.
  # Treat those metrics with prefix as one set.
  # Prioritier over 'measurement_name' configuration.
  #[[inputs.prom.measurements]]
  #  Name match, supports Regex.
  #  pattern = "cpu"
  #  Measurement set name
  #  name = "prom_cpu"

  # [[inputs.prom.measurements]]
  # disable_prefix = 0
  # pattern = "mem"
  # name = "mem"

  ## Collect interval, support "ns", "us" (or "µs"), "ms", "s", "m", "h".
  interval = "10s"
  
  ## TLS configuration.
  tls_open = false
  # tls_ca = "/tmp/ca.crt"
  # tls_cert = "/tmp/peer.crt"
  # tls_key = "/tmp/peer.key"

  ## Customize tags.
  [inputs.prom.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"

```
