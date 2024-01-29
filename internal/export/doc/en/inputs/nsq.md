---
title     : 'NSQ'
summary   : 'Collect NSQ metrics'
__int_icon      : 'icon/nsq'
dashboard :
  - desc  : 'NSQ'
    path  : 'dashboard/en/nsq'
monitor   :
  - desc  : 'NSQ'
    path  : 'monitor/en/nsq'
---

<!-- markdownlint-disable MD025 -->
# NSQ
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Collect NSQ operation data and report it to Guance Cloud in the form of indicators.

## Configuration {#config}

### Preconditions {#requirements}

- NSQ installed（[NSQ official website](https://nsq.io/){:target="_blank"}）

- Recommend NSQ version >= 1.0.0, already tested version:

- [x] 1.2.1
- [x] 1.1.0
- [x] 0.3.8

### Collector Configuration {#input-config}

=== "Host Installation"
<!-- markdownlint-disable MD046 -->
    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    The NSQ collector is available in two configurations, `lookupd` and `nsqd`, as follows:
    
    - `lookupd`: Configure the `lookupd` address of the NSQ cluster, and the collector will automatically discover the NSQ Server and collect data, which is more scalable.
    - `nsqd`: Configure a fixed list of NSQ Daemon (`nsqd`) addresses for which the collector collects only NSQ Server data
    
    The above two configuration methods are mutually exclusive, and `lookupd` has higher priority, so it is recommended to use `lookupd` configuration method.
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
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

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}
