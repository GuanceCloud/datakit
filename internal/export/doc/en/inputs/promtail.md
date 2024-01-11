---
title     : 'Promtail'
summary   : 'Collect log data reported by Promtail'
__int_icon      : 'icon/promtail'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Promtail Data Access
<!-- markdownlint-enable -->

---

:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker:

---

Start an HTTP endpoint to listen and receive promtail log data and report it to Guance Cloud.

## Configuration {#config}

Already tested version:

- [x] 2.8.2
- [x] 2.0.0
- [x] 1.5.0
- [x] 1.0.0
- [x] 0.1.0

### Collector Configuration {#input-config}

Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    After configuration, [Restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

### API Version {#API version}

For Promtail versions `v0.3.0` and before, Datakit's configuration should set `legacy = true`, by using legacy API [`POST /api/prom/push`](https://grafana.com/docs/loki/latest/api/#post-apiprompush){:target="_blank"} to receiving logging data from Promtail.

Using the default Datakit's configuration, namely `legacy = false` for the rest of Promtail versions, by using new API [`POST /loki/api/v1/push`](https://grafana.com/docs/loki/latest/api/#post-lokiapiv1push){:target="_blank"}.

### Custom Tags {#custom tags}

You can add custom tags to log data by configuring `[inputs.{{.InputName}}.tags]`, as shown below:

```toml
  [inputs.{{.InputName}}.tags]
    some_tag = "some_value"
    more_tag = "some_other_value"
```

After configuration, restart DataKit.

### Supported parameter {#args}

The promtail collector supports adding parameters to the HTTP URL. The list of parameters is as follows:

- `source`: Identifies the data source. Such as `nginx` or `redis`（`/v1/write/promtail?source=nginx`), With `source` set to `default`by default;
- `pipeline`: Specify the pipeline name required for the data, Such as `nginx.p`（`/v1/write/promtail?pipeline=nginx.p`）；
- `tags`: Add custom tags, separated by English commas `,`, such as `key1=value1` and `key2=value2`（`/v1/write/promtail?tags=key1=value1,key2=value2`）。

## Best Practice {#best practice}

Promtail's data was originally sent to Loki, which is, `/loki/api/v1/push`. Change the `url` in Promtail's configuration to Datakit, after enabled Datakit's promtail collector, Promtail would send its data to Datakit's promtail collector.

Promtail's configuration is like below:

```yaml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://localhost:9529/v1/write/promtail    # Send to the endpoint that the promtail collector listens on

scrape_configs:
  - job_name: system
    static_configs:
      - targets:
          - localhost
        labels:
          job: varlogs
          __path__: /var/log/*log
```

## Logging {#logging}

The logs delivered by Promtail shall prevail.
