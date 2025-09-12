---
title     : 'vSphere'
summary   : 'Collect vSphere metrics'
tags:
  - 'VMWARE'
__int_icon      : 'icon/vsphere'
dashboard :
  - desc  : 'vSphere'
    path  : 'dashboard/en/vsphere'
monitor   :
  - desc  : '-'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# vSphere
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

This collector gathers resource usage metrics from vSphere clusters, including resources such as CPU, memory, and network, and reports this data to the <<<custom_key.brand_name>>>.

## Configuration {#config}

### Preconditions {#requrements}

- Create a vSphere account:

In the vCenter management interface, create a user `datakit` and assign `read-only` permissions, applying these to the resources that need to be monitored. If monitoring of all child objects is required, you can select the `Propagate to children` option.

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, restart DataKit.

=== "Kubernetes"

    The collector can now be turned on by [configMap injection collector configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->

## Metric {#metric}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.{{.InputName}}.tags]`:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

<!-- markdownlint-disable MD046 -->
???+ note

    - Not all of the metrics listed below are collected; for specifics, refer to the explanations in the [Data Collection Levels](https://techdocs.broadcom.com/us/en/vmware-cis/vsphere/vsphere/9-0/vsphere-monitoring-and-performance/monitoring-inventory-objects/data-collection-levels.html){:target="_blank"}

<!-- markdownlint-enable -->
{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}{{end}}

{{ end }}

<!-- markdownlint-disable MD024 -->
## Object {#object}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "object"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}
{{end}}

{{ end }}

<!-- markdownlint-enable -->
## Logs {#logging}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}{{end}}

{{ end }}
