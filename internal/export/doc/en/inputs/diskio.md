---
title     : 'Disk IO'
summary   : 'Collect metrics of disk io'
tags:
  - 'HOST'
__int_icon      : 'icon/diskio'
dashboard :
  - desc  : 'Disk IO'
    path  : 'dashboard/en/diskio'
monitor   :
  - desc  : 'Host detection library'
    path  : 'monitor/en/host'
---


{{.AvailableArchs}}

---

Diskio collector is used to collect the index of disk flow and time.

## Configuration {#config}

After successfully installing and starting DataKit, the DiskIO collector will be enabled by default without the need for manual activation.

### Precondition {#requirement}

For some older versions of Windows operating systems, if you encounter an error with DataKit: **"The system cannot find the file specified."**

Run PowerShell as an administrator and execute:

```powershell
diskperf -Y
```

The DataKit service needs to be restarted after successful execution.

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->

=== "Host Installation"

    Go to the `conf.d/samples` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
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

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or it can be named by `[[inputs.diskio.tags]]` alternative host in the configuration.

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}

## FAQ {#faq}

### What is the data source on Linux hosts {#linux-diskio}

On Linux hosts, the metrics are parsed and calculated from the */proc/diskstats* file; an explanation of each column can be found in [docs](https://www.kernel.org/doc/Documentation/ABI/testing/procfs-diskstats){:target="_blank"};

The corresponding relationship between some data source columns and indicators is as follows:

| Fields                                     | `diskio` metrics                                                                                               |
| ---                                        | ---                                                                                                            |
| col04: reads completed successfully        | `reads`                                                                                                        |
| col05: reads merged                        | `merged_reads`                                                                                                 |
| col06: sectors read                        | `read_bytes = col06 * sector_size`; `read_bytes/sec = (read_bytes - last(read_bytes))/(time - last(time))`     |
| col07: time spent reading (ms)             | `read_time`                                                                                                    |
| col08: writes completed                    | `writes`                                                                                                       |
| col09: writes merged                       | `merged_writes`                                                                                                |
| col10: sectors written                     | `write_bytes = col10 * sector_size`; `write_bytes/sec = (write_bytes - last(write_bytes))/(time - last(time))` |
| col11: time spent writing (ms)             | `write_time`                                                                                                   |
| col12: I/Os currently in progress          | `iops_in_progress`                                                                                             |
| col13: time spent doing I/Os (ms)          | `io_time`                                                                                                      |
| col14: weighted time spent doing I/Os (ms) | `weighted_io_time`                                                                                             |

Attention:

1. `sector_size` is 512 bytes.
1. All are count but `read_bytes/sec` and `write_bytes/sec`.
