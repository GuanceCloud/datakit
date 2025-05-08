---
title     : 'Memcached'
summary   : 'Collect memcached metrics data'
tags:
  - 'CACHING'
  - 'MIDDLEWARE'
__int_icon      : 'icon/memcached'
dashboard :
  - desc  : 'Memcached'
    path  : 'dashboard/en/memcached'
monitor   :
  - desc  : 'N/A'
    path  : '-' 
---


{{.AvailableArchs}}

---

Memcached collector can collect the running status metrics from Memcached instances, and collect the metrics to the <<<custom_key.brand_name>>> to help monitor and analyze various abnormal situations of Memcached.

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

For all of the following data collections, the global election tags will added automatically, we can add extra tags in `[inputs.{{.InputName}}.tags]` if needed:

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
