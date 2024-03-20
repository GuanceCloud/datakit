---
title     : 'Doris'
summary   : 'Collect metrics of Doris'
__int_icon      : 'icon/doris'
dashboard :
  - desc  : 'Doris'
    path  : 'dashboard/en/doris'
monitor   :
  - desc  : 'Doris'
    path  : 'monitor/en/doris'
---

<!-- markdownlint-disable MD025 -->
# Doris
<!-- markdownlint-enable -->
---

{{.AvailableArchs}}

---

Doris collector is used to collect metric data related to Doris, and currently it only supports data in Prometheus format.

## Configuration {#config}

Already tested version:

- [x] 2.0.0

### Preconditions {#requirements}

Doris defaults to enabling the Prometheus port

Check front-end: curl ip: 8030/metrics

Check backend: curl ip: 8040/metrics

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "host installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .

<!-- markdownlint-enable -->

## Metric {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}

