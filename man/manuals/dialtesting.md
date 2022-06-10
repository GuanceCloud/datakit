{{.CSS}}
# 网络拨测

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

该采集器是网络拨测结果数据采集，所有拨测产生的数据，上报观测云。

## 私有拨测节点部署

私有拨测节点部署，需在 [观测云页面创建私有拨测节点](https://www.yuque.com/dataflux/doc/phmtep)。创建完成后，将页面上相关信息填入 `conf.d/{{.Catalog}}/{{.InputName}}.conf` 即可：

```toml
#  中心任务存储的服务地址
server = "https://dflux-dial.guance.com"

# require，节点惟一标识ID
region_id = "reg_c2jlokxxxxxxxxxxx"

# 若server配为中心任务服务地址时，需要配置相应的ak或者sk
ak = "ZYxxxxxxxxxxxx"
sk = "BNFxxxxxxxxxxxxxxxxxxxxxxxxxxx"

[inputs.dialtesting.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

> 注意：目前只有 linux 的拨测节点才支持「路由跟踪」，跟踪数据会保存在相关指标的 [traceroute](#traceroute-字段描述) 字段中。

## 拨测部署图

![](imgs/dialtesting-net-arch.png)

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 指标集

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[[inputs.{{.InputName}}.tags]]` 另择 host 来命名。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}


## `traceroute` 字段描述

traceroute 是「路由跟踪」数据的 JSON 文本，整个数据是一个数组对象，对象中的每个数组元素记录了一次路由探测的相关情况，示例如下：

```json
[
    {
        "total": 2,
        "failed": 0,
        "loss": 0,
        "avg_cost": 12700395,
        "min_cost": 11902041,
        "max_cost": 13498750,
        "std_cost": 1129043,
        "items": [
            {
                "ip": "10.8.9.1",
                "response_time": 13498750
            },
            {
                "ip": "10.8.9.1",
                "response_time": 11902041
            }
        ]
    },
    {
        "total": 2,
        "failed": 0,
        "loss": 0,
        "avg_cost": 13775021,
        "min_cost": 13740084,
        "max_cost": 13809959,
        "std_cost": 49409,
        "items": [
            {
                "ip": "10.12.168.218",
                "response_time": 13740084
            },
            {
                "ip": "10.12.168.218",
                "response_time": 13809959
            }
        ]
    }
]
```

**字段描述：**

| 字段              | 类型   |  说明                                    |
| :---              | ---    |  ---                                     |
| `total` | number | 总探测次数 |
| `failed` | number | 失败次数 |
| `loss` | number | 失败百分比 |
| `avg_cost` | number | 平均耗时(ns) |
| `min_cost` | number | 最小耗时(ns) |
| `max_cost` | number | 最大耗时(ns) |
| `std_cost` | number | 耗时标准差(ns) |
| `items` | Item 的 Array | 每次探测信息([详见](#Item)) |

### Item

| 字段              | 类型   |  说明                                    |
| :---              | ---    |  ---                                     |
| `ip` | string | IP 地址，如果失败，值为 * |
| `response_time` | number | 响应时间(ns) |

