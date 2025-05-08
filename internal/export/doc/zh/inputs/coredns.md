---
title     : 'CoreDNS'
summary   : '采集 CoreDNS 的指标数据'
tags:
  - '中间件'
__int_icon      : 'icon/coredns'
dashboard :
  - desc  : 'CoreDNS'
    path  : 'dashboard/zh/coredns'
monitor   :
  - desc  : '暂无'
    path  : 'monitor/zh/coredns'
---


{{.AvailableArchs}}

---

CoreDNS 采集器用于采集 CoreDNS 相关的指标数据。

## 配置 {#config}

### 前置条件 {#requirements}

- CoreDNS [配置](https://coredns.io/plugins/metrics/){:target="_blank"}启用 `prometheus` 插件

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    通过 DataKit 来开启[ `kubernetesprometheus` 采集器(https://docs.<<<custom_key.brand_main_domain>>>/integrations/kubernetesprometheus/)。

    ```yaml
    [inputs.kubernetesprometheus]
      [[inputs.kubernetesprometheus.instances]]
          role       = "pod"
          namespaces = ["kube-system"]
          selector   = "k8s-app=kube-dns"
          port     = "__kubernetes_pod_container_coredns_port_metrics_number"
        [inputs.kubernetesprometheus.instances.custom]
          [inputs.kubernetesprometheus.instances.custom.tags]
            cluster = "demo"
    ```
<!-- markdownlint-enable -->

## 指标 {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
