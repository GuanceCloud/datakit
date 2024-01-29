---
title     : 'Memcached'
summary   : 'Collect memcached metrics data'
__int_icon      : 'icon/memcached'
dashboard :
  - desc  : 'Memcached'
    path  : 'dashboard/en/memcached'
monitor   :
  - desc  : 'N/A'
    path  : '-' 
---

<!-- markdownlint-disable MD025 -->
# Memcached
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Memcached collector can collect the running status metrics from Memcached instances, and collect the metrics to the observation cloud to help monitor and analyze various abnormal situations of Memcached.

## Config {#config}

### Preconditions {#requirements}

- Memcached version >= `1.5.0`. Already tested version:
    - [x] 1.5.x
    - [x] 1.6.x

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->

## Metric {#metric}

For all the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.{{.InputName}}.tags]`:

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
