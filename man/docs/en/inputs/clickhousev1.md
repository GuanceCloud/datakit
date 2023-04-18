
# ClickHouse
---

{{.AvailableArchs}}

---

ClickHouse collector can collect various metrics actively exposed by ClickHouse server instances, such as the number of statements executed, memory storage, IO interaction and other metrics, and collect the metrics into Guance Cloud to help you monitor and analyze various abnormal situations of ClickHouse.

## Preconditions {#requirements}

ClickHouse version >=v20.1.2.4

Find the following code snippet in the config.xml configuration file of clickhouse-server, uncomment it, and set the port number exposed by metrics (which is unique if you choose it yourself). Restart after modification (if it is a cluster, every machine needs to operate).

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

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, [restart DataKit](datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    At present, you can [inject collector configuration in ConfigMap mode](datakit-daemonset-deploy.md#configmap-setting)。

## Measurements {#measurements}

For all the following data collections, a global tag named `host` is appended by default (the tag value is the host name where the DataKit is located), or other tags can be customized in the configuration through `[inputs.prom.tags]`(Hostname can be added to the cluster).

``` toml
    [inputs.prom.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
```

## Metrics {#metrics}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- field list

{{$m.FieldsMarkdownTable}} {{end}}

{{ end }}
