---
title     : 'TDengine'
summary   : '采集 TDengine 的指标数据'
__int_icon      : 'icon/tdengine'
dashboard :
  - desc  : 'TDengine'
    path  : 'dashboard/zh/tdengine'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# TDengine
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

TDengine 是一款高性能、分布式、支持 SQL 的时序数据库 (Database)。在开通采集器之前请先熟悉 [TDengine 基本概念](https://docs.taosdata.com/concept/){:target="_blank"}

TDengine 采集器需要的连接 `taos_adapter` 才可以正常工作，taosAdapter 从 TDengine v2.4.0.0 版本开始成为 TDengine 服务端软件 的一部分，本文主要是指标集的详细介绍。

## 配置  {#config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。

???+ tip

    - 连接 taoAdapter 之前请先确定端口是开放的。并且连接用户需要有 read 权限。
    - 若仍连接失败，[请参考此处](https://docs.taosdata.com/2.6/train-faq/faq/){:target="_blank"}。
<!-- markdownlint-enable -->

## 指标 {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

> - 数据库中有些表中没有 `ts` 字段，Datakit 会使用当前采集的时间。
