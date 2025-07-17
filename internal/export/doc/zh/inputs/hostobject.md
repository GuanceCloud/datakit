---
title     : '主机对象'
summary   : '采集主机基本信息'
tags:
  - '主机'
__int_icon      : 'icon/hostobject'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

{{.AvailableArchs}}

---

主机对象采集器用于收集主机基本信息，如硬件型号、基础资源消耗等。

## 配置 {#config}

成功安装 DataKit 并启动后，会默认开启主机对象采集器，无需手动开启。

<!-- markdownlint-disable MD046 -->

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

    也支持以环境变量的方式修改配置参数（需要在 ENV_DEFAULT_ENABLED_INPUTS 中加为默认采集器）：

{{ CodeBlock .InputENVSampleZh 4 }}

<!-- markdownlint-enable -->

### 开启云同步 {#cloudinfo}

DataKit 默认开启云同步，目前支持阿里云/腾讯云/AWS/华为云/微软云/火山引擎。可以通过设置 cloud_provider tag 显式指定云厂商，也可以由 DataKit 自动进行探测：

```toml
[inputs.hostobject.tags]
  # 此处目前支持 aliyun/tencent/aws/hwcloud/azure 几种，若不设置，则由 DataKit 自动探测并设置此 tag
  cloud_provider = "aliyun"
```

可以通过在配置文件中配置 `disable_cloud_provider_sync = true` 关闭云同步功能。

## 对象 {#object}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

```toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

???+ quote
    这里添加自定义 tag 时，尽量不要跟已有的 tag key/field key 同名。如果同名，DataKit 将选择配置里面的 tag 来覆盖采集的数据，可能导致一些数据问题。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}

如果开启了云同步，会多出如下一些字段（以同步到的字段为准）：

| 字段名                  | 描述           |  类型  |
| -----------------------:| -------------- | :----: |
| `cloud_provider`        | 云服务商       | string |
| `description`           | 描述           | string |
| `instance_id`           | 实例 ID        | string |
| `instance_name`         | 实例名         | string |
| `instance_type`         | 实例类型       | string |
| `instance_charge_type`  | 实例计费类型   | string |
| `instance_network_type` | 实例网络类型   | string |
| `instance_status`       | 实例状态       | string |
| `security_group_id`     | 实例分组       | string |
| `private_ip`            | 实例私网 IP    | string |
| `zone_id`               | 实例 Zone ID   | string |
| `region`                | 实例 Region ID | string |

### `message` 指标字段结构 {#message-struct}

`message` 字段基本结构如下：

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
    "config_file": ...,
  },

  "collectors": [ # 各个采集器的运行情况
    ...
  ]
}
```

#### `host.meta` {#host-meta}

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `host_name`        | 主机名                                         | string |
| `boot_time`        | 开机时间                                       |  int   |
| `os`               | 操作系统类型，如 `linux/windows/darwin`        | string |
| `platform`         | 平台名称，如 `ubuntu`                          | string |
| `platform_family`  | 平台分类，如 `ubuntu` 属于 `debian` 分类       | string |
| `platform_version` | 平台版本，如 `18.04`，即 Ubuntu 的某个分发版本 | string |
| `kernel_release`   | 内核版本，如 `4.15.0-139-generic`              | string |
| `arch`             | CPU 硬件架构，如 `x86_64/arm64` 等             | string |
| `extra_cloud_meta` | 开启云同步时，会带上一串云属性的 JSON 数据     | string |

#### `host.cpu` {#host-cpu}

| 字段名        | 描述                                                    |  类型  |
| -------------:| ------------------------------------------------------- | :----: |
| `vendor_id`   | 供应商 ID，如 `GenuineIntel`                            | string |
| `module_name` | CPU 型号，如 `Intel(R) Core(TM) i5-8210Y CPU @ 1.60GHz` | string |
| `cores`       | 核数                                                    |  int   |
| `mhz`         | 频率                                                    |  int   |
| `cache_size`  | L2 缓存大小（KB）                                       |  int   |

#### `host.mem` {#host-mem}

| 字段名         | 描述       | 类型 |
| --------------:| ---------- | :--: |
| `memory_total` | 总内存大小 | int  |
| `swap_total`:  | swap 大小  | int  |

#### `host.net` {#host-net}

| 字段名    | 描述               |   类型   |
| ---------:| ------------------ | :------: |
| `mtu`     | 最大传输单元       |   int    |
| `name`    | 网卡名称           |  string  |
| `mac`     | MAC 地址           |  string  |
| `flags`   | 状态位（可能多个） | []string |
| `ip4`     | IPv4 地址          |  string  |
| `ip6`     | IPv6 地址          |  string  |
| `ip4_all` | 所有 IPv4 地址     | []string |
| `ip6_all` | 所有 IPv6 地址     | []string |

#### `host.disk` {#host-disk}

???+ quote
    之前的版本中，同一个设备只会采集一个挂载点（具体采集哪一个，以具体挂载点在 */proc/self/mountpoint* 出现的顺序为准）。在 [:octicons-tag-24: Version-1.66.0](../datakit/changelog-2025.md#cl-1.66.0) 版本中，主机对象中的磁盘部份会将符合条件（比如设备名以 `/dev` 开头）挂载点都采集上来，其目的是为了展示 DataKit 能看到的所有设备，避免遗漏。

| 字段名       | 描述         |  类型  |
| ------------:| ------------ | :----: |
| `device`     | 磁盘设备名   | string |
| `total`      | 磁盘总大小   |  int   |
| `mountpoint` | 挂载点       | string |
| `fstype`     | 文件系统类型 | string |

#### `host.election` {#host-election}

???+ quote
    当配置文件中 `enable_election` 选项关闭时，该字段为 null

| 字段名      | 描述     |  类型  |
| -----------:| -------- | :----: |
| `elected`   | 选举状态 | string |
| `namespace` | 选举空间 | string |

#### `host.conntrack` {#host-conntrack}

<!-- markdownlint-disable MD046 -->
???+ quote
    - `conntrack` 仅 Linux 平台支持
    - Linux 下有时候这俩个指标采集不到，显示为 -1。此时我们需要加载 `nf_conntrack` 模块，终端执行如下命令即可：

        ```shell
        modprobe nf_conntrack
        ```
<!-- markdownlint-enable -->

| 字段名                | 描述                                           | 类型 |
| ---------------------:| ---------------------------------------------- | :--: |
| `entries`             | 当前连接数量                                   | int  |
| `entries_limit`       | 连接跟踪表的大小                               | int  |
| `stat_found`          | 成功的搜索条目数目                             | int  |
| `stat_invalid`        | 不能被跟踪的包数目                             | int  |
| `stat_ignore`         | 已经被跟踪的报数目                             | int  |
| `stat_insert`         | 插入的包数目                                   | int  |
| `stat_insert_failed`  | 插入失败的包数目                               | int  |
| `stat_drop`           | 跟踪失败被丢弃的包数目                         | int  |
| `stat_early_drop`     | 由于跟踪表满而导致部分已跟踪包条目被丢弃的数目 | int  |
| `stat_search_restart` | 由于 hash 表大小修改而导致跟踪表查询重启的数目 | int  |

#### `host.filefd` {#host-filefd}

<!-- markdownlint-disable MD046 -->
???+ quote
    `filefd` 仅 Linux 平台支持
<!-- markdownlint-enable -->

| 字段名         | 描述                                                 | 类型  |
| --------------:| ---------------------------------------------------- | :---: |
| `allocated`    | 已分配文件句柄的数目                                 |  int  |
| `maximum`      | 文件句柄的最大数目（已弃用，用 `maximum_mega` 替代） |  int  |
| `maximum_mega` | 文件句柄的最大数目，单位 M(10^6)                     | float |

#### `host.config_file` {#host-config-file}

`config_file` 是一个 `{"file-path": "file-content"}` 的 map，每个字段的含义如下：

| 字段名         | 描述                                                 | 类型   |
| --------------:| ---------------------------------------------------- | :---:  |
| `file-path`    | 配置文件的绝对路径                                   | string |
| `file-content` | 配置文件的内容                                       | string |

#### 采集器运行情况字段列表 {#inputs-stats}

`collectors` 字段是一个对象列表，每个对象的字段如下：

| 字段名          | 描述                                               |  类型  |
| ---------------:| -------------------------------------------------- | :----: |
| `name`          | 采集器名称                                         | string |
| `count`         | 采集次数                                           |  int   |
| `last_err`      | 最后一次报错信息，只报告最近 30 秒（含）以内的错误 | string |
| `last_err_time` | 最后一次报错时间（Unix 时间戳，单位为秒）          |  int   |
| `last_time`     | 最近一次采集时间（Unix 时间戳，单位为秒）          |  int   |
