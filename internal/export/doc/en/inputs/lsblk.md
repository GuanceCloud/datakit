---
title     : 'Lsblk'
summary   : 'Collect metrics of block device'
tags:
  - 'HOST'
__int_icon      : 'icon/disk'
dashboard :
  - desc  : 'Lsblk'
    path  : 'dashboard/lsblk'
---

{{.AvailableArchs}}

---

Lsblk collector is used for Linux host block device information collection, such as device name, primary and secondary device number, filesystem available size, filesystem type, filesystem used size, filesystem usage percentage, device mount location, etc.

## Configuration {#config}

<!-- markdownlint-disable MD046 -->

### Collector Configuration {#input-config}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .

    Can also be turned on by environment variables, (needs to be added as the default collector in ENV_DEFAULT_ENABLED_INPUTS):
    
{{ CodeBlock .InputENVSample 4 }}

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

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}

{{ end }}
