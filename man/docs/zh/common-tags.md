# 常见 Tag 整理
---

在 DataKit 采集的数据中，Tag 是所有数据的关键字段，它影响数据的过滤和分组，一旦 Tag 数据有误，将导致 Web 页面数据展示错误。另外，Tag 的标定还会影响时序数据的用量统计。故在设计和变更 Tag 的过程中，应该深思熟虑，全盘考虑对应的变动是否会造成相关的问题。本文档主要列举一下当前 DataKit 中常见的 Tag，一来用以明确每个 Tag 的具体意义，二来，在未来新加 Tag 的时候，应该沿用、遵循以下这些 Tag 的命名和标定，避免出现不一致的情况。

下面将从全局 Tag 和特定数据类型专属 Tag 两个维度来进行罗列。

## 全局类 Tag {#global-tags}

这些 Tag 跟具体数据类型无关，它可以追加到任意数据类型上。

| Tag                | 描述                                                                                                |
| ---                | ---                                                                                                 |
| host               | 主机名，daemonset 安装和主机安装都能带上这个 tag，在某些特定的情况下，用户可以 rename 这个 tag 的值 |
| project            | 项目名，一般都是由用户设置                                                                          |
| cluster            | 集群名，一般在 daemonset 安装中，由用户设置                                                         |
| election_namespace | 选举所在的命名空间，默认不追加，详见[文档](datakit-daemonset-deploy.md#env-elect)                   |
| version            | 版本号，所有涉及版本信息的 tag 字段，都应该以该 tag 来表示                                          |

### Kubernates/容器常见 Tag {#k8s-tags}

这些 tag 在采集到的数据中，一般都会有追加，但涉及时序采集的时候，默认会忽略一些多变的 tag（比如 `pod_name`），以节约时间线。

| Tag            | 描述                    |
| ---            | ---                     |
| pod_name       | pod 名称                |
| deployment     | k8s 中 Deployment 名称  |
| service        | k8s 中 Service 名称     |
| namespace      | k8s 中 Namespace 名称   |
| job            | k8s 中 Job 名称         |
| image          | k8s 中 镜像全称         |
| image_name     | k8s 中镜像名简称        |
| container_name | ks8/容器中的容器名      |
| cronjob        | k8s 中 CronJob 名称     |
| daemonset      | k8s 中 Daemonset 名称   |
| replica_set    | k8s 中 ReplicaSet 名称 |
| node_name      | k8s 中 Node 名称        |
| node_ip        | k8s 中 Node IP          |

## 按特定数据类型的 Tag 分类 {#tag-classes}

### 日志 {#L}

| Tag                | 描述                                                                                                |
| ---                | ---                                                                                                 |
| source | 日志来源，在行协议上，它并不是以 tag 形式存在，而是作为指标集名称，但中心将其作为 tag 存为日志的 source 字段 |
| service | 日志的 service 名称，如果不填写，其值等同于 source 字段 |
| status | 日志等级，如果不填写，采集器会默认将其值置为 `unknown`，常见的 status 列表在[这里](logging.md#status) |

### 对象 {#O}

| Tag                | 描述                                                                                                |
| ---                | ---                                                                                                 |
| class | 对象分类，在行协议上，它并不是以 tag 形式存在，而是作为指标集名称，但中心将其作为 tag 存为对象的 class 字段 |
| name | 对象名称，中心会结合 hash(class + name) 来唯一确定某个工作空间中的对象 |

### 指标 {#M}

指标由于数据来源纷杂，除了全局类 tag 外，没有固定的 tag。

### APM {#T}

Tracing 类数据的 tag 统一在[这里](ddtrace.md#measurements)

### RUM {#R}

详见 RUM 文档：

- [Web](../real-user-monitoring/web/app-data-collection.md)
- [Android](../real-user-monitoring/android/app-data-collection.md)
- [iOS](../real-user-monitoring/ios/app-data-collection.md)
- [小程序](../real-user-monitoring/miniapp/app-data-collection.md)
- [Flutter](../real-user-monitoring/flutter/app-data-collection.md)
- [React Native](../real-user-monitoring/react-native/app-data-collection.md)

### Scheck {#S}

参见 [Scheck 对应文档](../scheck/scheck-how-to.md)

### Profile {#P}

参见[采集器文档](profile.md#measurements)

### Network {#N}

参见[采集器文档](ebpf.md#measurements)

### Event {#E}

参见[设计文档](../events/generating.md)
