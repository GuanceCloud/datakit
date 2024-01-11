---
title     : 'Host Object'
summary   : 'Collect Basic Host Information'
__int_icon      : 'icon/hostobject'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Host Object
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Hostobject is used to collect basic host information, such as hardware model, basic resource consumption and so on.

## Configuration {#config}

In general, the host object is turned on by default and does not need to be configured.

<!-- markdownlint-disable MD046 -->

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, restart DataKit.

=== "Kubernetes"

    In general, the host object is turned on by default and does not need to be configured. In Kubernetes, it is supported to modify default parameters in the form of environment variables:
    
    | Environment Variable Name                                           | Corresponding Configuration Parameter Item                | Parameter Description                                                           | Parameter Example                                                                                                   |
    | :---                                                 | ---                             | ---                                                                | ---                                                                                                        |
    | `ENV_INPUT_HOSTOBJECT_ENABLE_NET_VIRTUAL_INTERFACES` | `enable_net_virtual_interfaces` | Allow collection of virtual network card                                                   | `true`/`false`                                                                                             |
    | `ENV_INPUT_HOSTOBJECT_ENABLE_ZERO_BYTES_DISK`        | `ignore_zero_bytes_disk`        | Ignore disks with size 0                                                | `true`/`false`                                                                                             |
    | `ENV_INPUT_HOSTOBJECT_TAGS`                          | `tags`                          | Add additional labels                                                       | `tag1=value1,tag2=value2`; If there is a tag with the same name in the configuration file, it will be overwritten.                                               |
    | `ENV_INPUT_HOSTOBJECT_ONLY_PHYSICAL_DEVICE`          | `only_physical_device`          | Ignore non-physical disks (such as network disk, NFS, etc., only collect local hard disk/CD ROM/USB disk, etc.) | Just give an arbitrary string value                                                                                     |
    | `ENV_INPUT_HOSTOBJECT_EXCLUDE_DEVICE`                      | `exclude_device`                | ignored device                                | `"/dev/loop0","/dev/loop1"` separated by English commas                      |
    | `ENV_INPUT_HOSTOBJECT_EXTRA_DEVICE`                        | `extra_device`                  | Additional device                            | `"/nfsdata"` separated by English commas                      |
    | `ENV_CLOUD_PROVIDER`                                 | `tags`                          | Designate cloud service provider                                                       | `aliyun/aws/tencent/hwcloud/azure`                                                                         |

<!-- markdownlint-enable -->

### Turn on Cloud Synchronization {#cloudinfo}

Datakit turns on cloud synchronization by default, and currently supports Alibaba Cloud/Tencent Cloud/AWS/Huawei Cloud/Microsoft Cloud. You can specify the cloud vendor explicitly by setting the cloud_provider tag, or you can detect it automatically by Datakit:

```toml
[inputs.hostobject.tags]
  # There are several kinds of aliyun/tencent/aws/hwcloud/azure supported at present. If not set, Datakit will detect and set this tag automatically
  cloud_provider = "aliyun"
```

You can turn off cloud synchronization by configuring `disable_cloud_provider_sync = true` in the hostobject configuration file.

## Object {#object}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.{{.InputName}}.tags]`:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

> Note: When adding custom tags here, try not to have the same name as the existing tag key/field key. If it has the same name, DataKit will choose to configure the tag inside to overwrite the collected data, which may cause some data problems.

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}

If cloud synchronization is turned on, the following additional fields will be added (whichever field is synchronized to):

| Field Name                  | Description           | Type   |
| ---                     | ----           | :---:  |
| `cloud_provider`        | Cloud service provider       | string |
| `description`           | Description           | string |
| `instance_id`           | Instance ID        | string |
| `instance_name`         | Instance name         | string |
| `instance_type`         | Instance type       | string |
| `instance_charge_type`  | Instance billing type   | string |
| `instance_network_type` | Instance network type   | string |
| `instance_status`       | Instance state       | string |
| `security_group_id`     | Instance grouping       | string |
| `private_ip`            | Instance private network IP    | string |
| `zone_id`               | Instance Zone ID   | string |
| `region`                | Instance Region ID | string |

### `message` Metric Field Structure {#message-struct}

The basic structure of the `message` field is as follows:

```json
{
  "host": {
    "meta": ...,
    "cpu": ...,
    "mem": ...,
    "net": ...,
    "disk": ...,
    "conntrack": ...,
    "filefd": ...,
    "election": ...,
  },

  "collectors": [ # Operation of each collector
    ...
  ]
}
```

#### `host.meta` {#host-meta}

| Field Name             | Description                                           | Type   |
| ---                | ----                                           | :---:  |
| `host_name`        | hostname                                         | string |
| `boot_time`        | Startup time                                       | int    |
| `os`               | Operating system type, such as `linux/windows/darwin`        | string |
| `platform`         | Platform name, such as `ubuntu`                          | string |
| `platform_family`  | Platform classification, such as `ubuntu` belongs to `debian` classification       | string |
| `platform_version` | Platform version, such as `18.04`, that is, a distribution version of Ubuntu | string |
| `kernel_release`   | Kernel version, such as `4.15.0-139-generic`              | string |
| `arch`             | Switch hardware architecture, such as `x86_64/arm64`            | string |
| `extra_cloud_meta` | When cloud synchronization is turned on, it will bring a string of JSON data with cloud attributes.     | string |

#### `host.cpu` {#host-cpu}

| Field Name        | Description                                                    | Type   |
| ---           | ----                                                    |:---:   |
| `vendor_id`   | Vendor ID, such as `GenuineIntel`                            | string |
| `module_name` | CPU model, such as `Intel(R) Core(TM) i5-8210Y CPU @ 1.60GHz` | string |
| `cores`       | Audit                                                    | int    |
| `mhz`         | Frequency                                                    | int    |
| `cache_size`  | L2 Cache size (KB)                                       | int    |

#### `host.mem` {#host-mem}

| Field Name         | Description       | Type |
| ---            | ----       |:---: |
| `memory_total` | Total memory size | int  |
| `swap_total`:  | swap size  | int  |

#### `host.net` {#host-net}

| Field Name  | Description               | Type     |
| ---     | ----               |:---:     |
| `mtu`   | Maximum transmission unit       | int      |
| `name`  | NIC Name           | string   |
| `mac`   | MAC address           | string   |
| `flags` | Status bits (may be multiple) | []string |
| `ip4`   | IPv4 address          | string   |
| `ip6`   | IPv6 address          | string   |
| `ip4_all`| all IPv4 address     | []string |
| `ip6_all`| all IPv6 address     | []string |

#### `host.disk` {#host-disk}

| Field Name       | Description         | Type   |
| ---          | ----         |:---:   |
| `device`     | Disk device name   | string |
| `total`      | Total disk size   | int    |
| `mountpoint` | Mount point       | string |
| `fstype`     | File system type | string |

#### `host.election` {#host-election}

> Note: This field is null when the `enable_election` option is turned off in the configuration file

| Field Name      | Description     | Type   |
| ---         | ----     | :---:  |
| `elected`   | Election status | string |
| `namespace` | Election space | string |

#### `host.conntrack` {#host-conntrack}

<!-- markdownlint-disable MD046 -->

???+ attention

    `conntrack` 仅 Linux 平台支持

<!-- markdownlint-enable -->

| Field Name                | Description                                           | Type  |
| ---                   | ---                                            | :---: |
| `entries`             | Current number of connections                                   | int   |
| `entries_limit`       | Size of Connection Trace Table                               | int   |
| `stat_found`          | Number of successful search terms                             | int   |
| `stat_invalid`        | Number of packets that cannot be tracked                             | int   |
| `stat_ignore`         | Number of reports that have been tracked                             | int   |
| `stat_insert`         | Number of packets inserted                                   | int   |
| `stat_insert_failed`  | Number of packets that failed to insert                               | int   |
| `stat_drop`           | Trace failed the number of discarded packets                         | int   |
| `stat_early_drop`     | Number of partially tracked packet entries discarded due to full trace table | int   |
| `stat_search_restart` | Number of trace table queries restarted due to hash table size modification   | int   |

#### `host.filefd` {#host-filefd}

<!-- markdownlint-disable MD046 -->

???+ attention

    `filefd` Linux platform only

<!-- markdownlint-enable -->

| Field Name         | Description                                                 | Type  |
| ---            | ---                                                  | :---: |
| `allocated`    | Number of allocated file handles                                 | int   |
| `maximum`      | Maximum number of file handles (deprecated, replaced by `maximum_mega`) | int   |
| `maximum_mega` | Maximum number of file handles in M(10^6)                     | float |

#### Collector Performance Field List {#inputs-stats}

The `collectors` field is a list of objects with the following fields for each object:

| Field Name          | Description                                             | Type   |
| ---             | ----                                             | :---:  |
| `name`          | Collector name                                       | string |
| `count`         | Collection times                                         | int    |
| `last_err`      | For the last error message, only the errors within the last 30 seconds (inclusive) are reported. | string |
| `last_err_time` | The last time an error was reported (Unix timestamp in seconds).        | int    |
| `last_time`     | Last collection time (Unix timestamp in seconds)       | int    |

## FAQ {#faq}

### :material-chat-question: Why no `entries` and `entries_limit`, the value shows -1？ {#no-entries}

Need to load `nf_conntrack` module, run `modprobe nf_conntrack` in a terminal.
