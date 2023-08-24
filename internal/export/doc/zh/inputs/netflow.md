---
title     : 'NetFlow'
summary   : 'NetFlow 采集器可以用来可视化和监控已开启 NetFlow 的设备'
__int_icon      : 'icon/netflow'
dashboard :
  - desc  : 'NetFlow'
    path  : 'dashboard/zh/netflow'
monitor   :
  - desc  : 'NetFlow'
    path  : 'monitor/zh/netflow'
---

<!-- markdownlint-disable MD025 -->
# NetFlow
<!-- markdownlint-enable -->
---

{{.AvailableArchs}}

---

NetFlow 采集器可以用来可视化和监控已开启 NetFlow 的设备，并将日志采集到观测云，帮助监控分析 NetFlow 各种异常情况。

## 什么是 NetFlow {#what}

NetFlow 是最广泛使用的流量数据统计标准，由 Cisco 开发，用于监控和记录进出接口的所有流量。NetFlow 分析它收集的流量数据，以提供关于流量和流量的可见性，并跟踪流量从何处来、流向何处以及在任何时候生成的流量。记录的信息可用于使用情况监视、异常检测和其他各种网络管理任务。

目前 Datakit 支持以下协议：

- netflow5
- netflow9
- sflow5
- ipfix

## 配置 {#config}

### 前置条件 {#requirements}

- 支持 NetFlow 功能的设备并且已开启 NetFlow 功能。每个设备的开启方法不尽相同，建议参考官方文档。例如：[在 Cisco ASA 上开启 NetFlow](https://www.petenetlive.com/KB/Article/0000055){:target="_blank"}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

## 数据结构说明 {#structure}

以下是一个日志示例：

```json
{
    "flush_timestamp":1692077978547,
    "type":"netflow5",
    "sampling_rate":0,
    "direction":"ingress",
    "start":1692077976,
    "end":1692077976,
    "bytes":668,
    "packets":1588,
    "ether_type":"IPv4",
    "ip_protocol":"TCP",
    "device":{
        "namespace":"namespace"
    },
    "exporter":{
        "ip":"10.200.14.142"
    },
    "source":{
        "ip":"130.240.103.204",
        "port":"4627",
        "mac":"00:00:00:00:00:00",
        "mask":"130.240.96.0/20"
    },
    "destination":{
        "ip":"152.222.36.168",
        "port":"424",
        "mac":"00:00:00:00:00:00",
        "mask":"152.0.0.0/8"
    },
    "ingress":{
        "interface":{
            "index":0
        }
    },
    "egress":{
        "interface":{
            "index":0
        }
    },
    "host":"MacBook-Air-2.local",
    "next_hop":{
        "ip":"20.104.52.139"
    }
}
```

解释如下：

- Root/NetFlow 节点

|  字段   | 说明  |
|  ----:  | :----  |
| flush_timestamp | 上报时间              |
| type            | 协议                 |
| sampling_rate   | 采样频率              |
| direction       | 方向                 |
| start           | 开始时间              |
| end             | 结束时间              |
| bytes           | 传输字节数             |
| packets         | 传输包数量             |
| ether_type      | 以太网类型（IPv4/IPv6） |
| ip_protocol     | IP 协议（TCP/UDP）     |
| device          | 设备信息节点            |
| exporter        | Exporter 信息节点      |
| source          | Flow 来源端信息节点     |
| destination     | Flow 去向端信息节点     |
| ingress         | 入口网关信息节点         |
| egress          | 出口网关信息节点         |
| host            | Collector 的 Hostname |
| tcp_flags       | TCP 标记               |
| next_hop        | Next_Hop 属性信息节点   |

- `device` 节点

|  字段   | 说明  |
|  ----:  | :----  |
| namespace | 命名空间 |

- `exporter` 节点

|  字段   | 说明  |
|  ----:  | :----  |
| ip | Exporter 的 IP |

- `source` 节点

|  字段   | 说明  |
|  ----:  | :----  |
| ip   | 来源端的 IP 地址  |
| port | 来源端的端口      |
| mac  | 来源端的 MAC 地址 |
| mask | 来源端的网络掩码   |

- `destination` 节点

|  字段   | 说明  |
|  ----:  | :----  |
| ip   | 去向端的 IP 地址    |
| port | 去向端的端口        |
| mac  | 去向端的 MAC 地址   |
| mask | 去向端的 IP 网络掩码 |

- `ingress` 节点

|  字段   | 说明  |
|  ----:  | :----  |
| interface | 网口编号 |

- `egress` 节点

|  字段   | 说明  |
|  ----:  | :----  |
| interface | 网口编号 |

- `next_hop` 节点

|  字段   | 说明  |
|  ----:  | :----  |
| ip | Next_Hop 属性中去往目的地的下一跳 IP 地址 |

## 指标集 {#metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}` {#{{$m.Name}}}

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
