---
title     : 'etcd'
summary   : '采集 etcd 的指标数据'
__int_icon      : 'icon/etcd'
dashboard :
  - desc  : 'etcd'
    path  : 'dashboard/zh/etcd'
  - desc  : 'etcd-k8s'
    path  : 'dashboard/zh/etcd-k8s'    
monitor   :
  - desc  : '暂无'            # 缺少监控视图示例
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# etcd
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

etcd 采集器可以从 etcd 实例中采取很多指标，比如 etcd 服务器状态和网络的状态等多种指标，并将指标采集到 DataFlux，帮助你监控分析 etcd 各种异常情况。

## 配置 {#config}

### 前置条件 {#requirements}

etcd 版本 >= 3, 已测试的版本：

- [x] 3.5.7
- [x] 3.4.24
- [x] 3.3.27

### 采集器配置 {#input-config}

开启 etcd，默认的 metrics 接口是 `http://localhost:2379/metrics`，也可以自行在配置文件中修改。

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

## 指标 {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
