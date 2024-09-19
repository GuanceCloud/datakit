---
title     : 'CoreDNS'
summary   : 'Collect CoreDNS metrics and logs'
__int_icon      : 'icon/coredns'
dashboard :
  - desc  : 'CoreDNS'
    path  : 'dashboard/en/coredns'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# CoreDNS
<!-- markdownlint-enable -->
---

{{.AvailableArchs}}

---

CoreDNS collector is used to collect metric data related to CoreDNS.

## Configuration {#config}

### Preconditions {#requirements}

- CoreDNS [configuration](https://coredns.io/plugins/metrics/){:target="_blank"}; Enable the `prometheus` plug-in

### Collector Configuration {#input-conifg}

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
