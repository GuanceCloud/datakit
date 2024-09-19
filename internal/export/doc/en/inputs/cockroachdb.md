---
title     : 'CockroachDB'
summary   : 'Collect CockroachDB metrics'
__int_icon      : 'icon/cockroachdb'
dashboard :
  - desc  : 'CockroachDB'
    path  : 'dashboard/en/cockroachdb'
monitor   :
  - desc  : 'CockroachDB'
    path  : 'monitor/en/cockroachdb'
---

<!-- markdownlint-disable MD025 -->
# CockroachDB
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

The CockroachDB collector is used to collect CockroachDB-related indicator data.
Currently, it only supports data in Prometheus format.

Already tested CockroachDB version:

- [x] CockroachDB 19.2
- [x] CockroachDB 20.2
- [x] CockroachDB 21.2
- [x] CockroachDB 22.2
- [x] CockroachDB 23.2.4

## Configuration {#config}

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host deployment"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .

<!-- markdownlint-enable -->
---

## Metric {#metric}

<!-- markdownlint-disable MD024 -->
{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- Tags

{{$m.TagsMarkdownTable}}

- Fields

{{$m.FieldsMarkdownTable}}

{{ end }}
<!-- markdownlint-enable -->
