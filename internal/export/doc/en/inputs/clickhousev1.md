---
title      : 'ClickHouse'
summary    : 'Collect metrics of ClickHouse'
__int_icon : 'icon/clickhouse'
tags:
  - 'DATA STORES'
dashboard :
  - desc  : 'ClickHouse'
    path  : 'dashboard/en/clickhouse'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---


{{.AvailableArchs}}

---

ClickHouse collector can collect various metrics actively exposed by ClickHouse server instances, such as the number of statements executed, memory storage, IO interaction and other metrics, and collect the metrics into <<<custom_key.brand_name>>> to help you monitor and analyze various abnormal situations of ClickHouse.

## Configuration {#config}

### Preconditions {#requirements}

ClickHouse version >=v20.1.2.4

Find the following code snippet in the config.xml configuration file of ClickHouse-server, uncomment it, and set the port number exposed by metrics (which is unique if you choose it yourself). Restart after modification (if it is a cluster, every machine needs to operate).

```shell
vim /etc/clickhouse-server/config.xml
```

```xml
<prometheus>
    <endpoint>/metrics</endpoint>
    <port>9363</port>
    <metrics>true</metrics>
    <events>true</events>
    <asynchronous_metrics>true</asynchronous_metrics>
</prometheus>
```

Field description:

- HTTP Routing of `endpoint` Prometheus Server Fetch Metrics
- `port` number of the port endpoint
- `metrics` grabs exposed metrics flags from ClickHouse's `system.metrics` table
- `events` grabs exposed event flags from ClickHouse's `table.events`.
- `asynchronous_metrics` grabs exposed asynchronous_metrics flags from ClickHouse's `system.asynchronous_metrics` table

See [ClickHouse official documents](https://ClickHouse.com/docs/en/operations/server-configuration-parameters/settings/#server_configuration_parameters-prometheus){:target="_blank"}

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    At present, you can [inject collector configuration in ConfigMap mode](../datakit/datakit-daemonset-deploy.md#configmap-setting)。
<!-- markdownlint-enable -->

## Metric {#metric}

For all of the following data collections, the global election tags will added automatically, we can add extra tags in `[inputs.{{.InputName}}.tags]` if needed:

``` toml
[inputs.prom.tags]
# some_tag = "some_value"
# more_tag = "some_other_value"
```

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.Desc}}

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}{{end}}

{{ end }}
