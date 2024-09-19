---
title     : 'Mem'
summary   : 'Collect metrics of host memory'
__int_icon      : 'icon/mem'
dashboard :
  - desc  : 'memory'
    path  : 'dashboard/en/mem'
monitor   :
  - desc  : 'host detection library'
    path  : 'monitor/en/host'  
---

<!-- markdownlint-disable MD025 -->
# Memory
<!-- markdownlint-enable -->

<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Mem collector is used to collect system memory information, some general metrics such as total memory, used memory and so on.


## Configuration {#config}

After successfully installing and starting DataKit, the Mem collector will be enabled by default without the need for manual activation.

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->

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

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.mem.tags]`:

``` toml
 [inputs.mem.tags]
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
