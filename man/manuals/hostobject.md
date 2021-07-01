{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

hostobject 用于收集主机基本信息，如硬件型号、基础资源消耗等。

## 前置条件

暂无

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 开启云同步

如果 DataKit 所在的主机是云主机（目前支持阿里云、腾讯云以及 AWS），那么可通过 `cloud_provider` 标签开启云同步：

```toml
[inputs.hostobject.tags]
	# 此处目前支持 aliyun/tencent/aws 三种
	cloud_provider = "aliyun"
```

## 指标集

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

### `message` 指标字段结构

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
		},

	"collectors": [ # 各个采集器的运行情况
		...
	]
}
```

#### `host.meta`

| 字段名             | 描述                                           | 类型   |
| ---                | ----                                           | :---:  |
| `host_name`        | 主机名                                         | string |
| `boot_time`        | 开机时间                                       | int    |
| `os`               | 操作系统类型，如 `linux/windows/darwin`        | string |
| `platform`         | 平台名称，如 `ubuntu`                          | string |
| `platform_family`  | 平台分类，如 `ubuntu` 属于 `debian` 分类       | string |
| `platform_version` | 平台版本，如 `18.04`，即 Ubuntu 的某个分发版本 | string |
| `kernel_release`   | 内核版本，如 `4.15.0-139-generic`              | string |
| `arch`             | CPU 硬件架构，如 `x86_64/arm64` 等             | string |

#### `host.cpu`

| 字段名        | 描述                                                    | 类型   |
| ---           | ----                                                    |:---:   |
| `vendor_id`   | 供应商 ID，如 `GenuineIntel`                            | string |
| `module_name` | CPU 型号，如 `Intel(R) Core(TM) i5-8210Y CPU @ 1.60GHz` | string |
| `cores`       | 核数                                                    | int    |
| `mhz`         | 频率                                                    | int    |
| `cache_size`  | L2 缓存大小（KB）                                       | int    |

#### `host.mem`

| 字段名         | 描述       | 类型 |
| ---            | ----       |:---: |
| `memory_total` | 总内存大小 | int  |
| `swap_total`:  | swap 大小  | int  |

#### `host.net`

| 字段名  | 描述               | 类型     |
| ---     | ----               |:---:     |
| `mtu`   | 最大传输单元       | int      |
| `name`  | 网卡名称           | string   |
| `mac`   | MAC 地址           | string   |
| `flags` | 状态位（可能多个） | []string |
| `ip4`   | IPv4 地址          | string   |
| `ip6`   | IPv6 地址          | string   |

#### `host.disk`

| 字段名       | 描述         | 类型   |
| ---          | ----         |:---:   |
| `device`     | 磁盘设备名   | string |
| `total`      | 磁盘总大小   | int    |
| `mountpoint` | 挂载点       | string |
| `fstype`     | 文件系统类型 | string |

#### `host.conntrack`

| 字段名		| 描述		| 类型 |
| --- | --- |:---: |
| `entries` | 当前连接数量| int |
| `entries_limit` | 连接跟踪表的大小 | int |
| `stat_found` | 成功的搜索条目数目  | int |
| `stat_invalid` | 不能被跟踪的包数目 | int |
| `stat_ignore` | 已经被跟踪的报数目 | int |
| `stat_insert` | 插入的包数目 | int |
| `stat_insert_failed` | 插入失败的包数目 | int |
| `stat_drop` | 跟踪失败被丢弃的包数目| int |
| `stat_early_drop` | 由于跟踪表满而导致部分已跟踪包条目被丢弃的数目| int |
| `stat_search_restart` | 由于hash表大小修改而导致跟踪表查询重启的数目 | int |

#### `host.filefd`

| 字段名		| 描述		| 类型 |
| --- | --- |:---: |
| `allocated` | 已分配文件句柄的数目| int |
| `maximum` | 文件句柄的最大数目 | int |

#### 单个采集器运行情况字段列表

| 字段名          | 描述                                           | 类型   |
| ---             | ----                                           | :---:  |
| `name`          | 采集器名称                                     | string |
| `count`         | 采集次数                                       | int    |
| `last_time`     | 最近一次采集时间                               | int    |
| `last_err`      | 最后一次报错信息(默认只报告 30 分钟以内的错误) | string |
| `last_err_time` | 最后一次报错时间                               | int    |
