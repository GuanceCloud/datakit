---
title     : 'Disk'
summary   : 'Collect metrics of disk'
tags:
  - 'HOST'
__int_icon      : 'icon/disk'
dashboard :
  - desc  : 'disk'
    path  : 'dashboard/en/disk'
monitor   :
  - desc  : 'host detection library'
    path  : 'monitor/en/host'
---


{{.AvailableArchs}}

---

Disk collector is used to collect disk information, such as disk storage space, inode usage, etc.

## Configuration {#config}

After successfully installing and starting DataKit, the disk collector will be enabled by default without the need for manual activation.

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

<!-- markdownlint-disable MD046 -->
???+ info "Source of disk metrics"
    Under Linux, we first get device and mount info from */proc/self/mountinfo*, then get disk usage metrics via `statfs()` syscall. For Windows, we get device and mount info via Windows APIs like `GetLogicalDriveStringsW()`, and get disk usage by another API `GetDiskFreeSpaceExW()`

    In the [:octicons-tag-24: Version-1.66.0](../datakit/changelog-2025.md#cl-1.66.0) release, the disk collector has been optimized. However, mount points for the same device will still be merged into one, with only the first mount point being taken. If you need to collect all mount points, a specific flag(`merge_on_device/ENV_INPUT_DISK_MERGE_ON_DEVICE`) must be disable. While this flag disabled, this may result in a significant increase in the number of time series in the disk measurement.
<!-- markdownlint-enable -->

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}

{{ end }}
