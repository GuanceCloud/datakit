---
title     : 'TDengine'
summary   : 'Collect TDengine metrics'
__int_icon      : 'icon/tdengine'
dashboard :
  - desc  : 'TDengine'
    path  : 'dashboard/en/tdengine'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# TDengine
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

TDEngine is a high-performance, distributed, SQL-enabled time series Database (Database). Familiarize yourself with the [basic concepts of TDEngine](https://docs.taosdata.com/concept/){:target="_blank"} before opening the collector.

TDengine collector needs to connect `taos_adapter` can work normally, taosAdapter from TDengine v2.4. 0.0 version comes to becoming a part of TDengine server software, this paper is mainly a detailed introduction of measurement.

## Configuration  {#config}

### Collector Config {#input-config}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    At present, the collector can be turned on by [injecting the collector configuration in ConfigMap mode](../datakit/datakit-daemonset-deploy.md#configmap-setting).

???+ tip

    Please make sure the port is open before connecting to the taoAdapter. And the connecting user needs to have read permission.
    If the connection still fails, [please refer to](https://docs.taosdata.com/2.6/train-faq/faq/){:target="_blank"}

## Metric {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}

> - Some tables in the database do not have the `ts` field, and Datakit uses the current collection time.
