---
title     : 'InfluxDB'
summary   : 'Collect InfluxDB metrics'
tags:
  - 'DATABASE'
__int_icon      : 'icon/influxdb'
dashboard :
  - desc  : 'InfluxDB'
    path  : 'dashboard/en/influxdb'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---


{{.AvailableArchs}}

---

The InfluxDB collector is used to collect the data of the InfluxDB.

## Configuration {#config}

### Preconditions {#requirements}

The influxdb collector is only applicable to influxdb v1.x, and the prom collector is required for influxdb v2.x.

Already tested version:

- [x] 1.8.10

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap injection collector configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->


### InfluxDB v2.x {#prom-config}

```toml
[[inputs.prom]]
  ## Exporter address
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

## Metric {#metric}

For all of the following data collections, the global election tags will added automatically, we can add extra tags in `[inputs.{{.InputName}}.tags]` if needed:

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

## Log Collection {#logging}

To collect the InfluxDB log, open `files` in {{.InputName}}.conf and write to the absolute path of the InfluxDB log file. For example:

```toml
[inputs.{{.InputName}}.log]
    # Fill in the absolute path
    files = ["/path/to/demo.log"] 
    ## grok pipeline script path
    pipeline = "influxdb.p"
```
