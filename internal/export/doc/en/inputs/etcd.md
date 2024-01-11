---
title     : 'etcd'
summary   : 'Collect etcd metrics'
__int_icon      : 'icon/etcd'
dashboard :
  - desc  : 'etcd'
    path  : 'dashboard/zh/etcd'
  - desc  : 'etcd-k8s'
    path  : 'dashboard/zh/etcd-k8s'    
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# etcd
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

The tcd collector can take many metrics from the etcd instance, such as the status of the etcd server and network, and collect the metrics to DataFlux to help you monitor and analyze various abnormal situations of etcd.

## Configuration {#config}

### Preconditions {#requirements}

etcd version >= 3, Already tested version:

- [x] 3.5.7
- [x] 3.4.24
- [x] 3.3.27

Open etcd, the default metrics interface is `http://localhost:2379/metrics`, or you can modify it in your configuration file.

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

## Metric {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}
