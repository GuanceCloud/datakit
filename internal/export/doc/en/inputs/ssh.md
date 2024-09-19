---
title     : 'SSH'
summary   : 'Collect SSH metrics'
__int_icon      : 'icon/ssh'
dashboard :
  - desc  : 'SSH'
    path  : 'dashboard/en/ssh'
monitor   :
  - desc  : 'SSH'
    path  : 'monitor/en/ssh'
---

<!-- markdownlint-disable MD025 -->
# SSH
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Monitor SSH/SFTP services and report data to Guance Cloud.

## Configuration {#config}

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

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}
