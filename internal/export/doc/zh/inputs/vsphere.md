---
title     : 'vSphere'
summary   : '采集 vSphere 的指标数据'
tags:
  - 'VMWARE'
__int_icon      : 'icon/vsphere'
dashboard :
  - desc  : 'vSphere'
    path  : 'dashboard/zh/vsphere'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# vSphere
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

本采集器采集 vSphere 集群的资源使用指标，包括 CPU、内存和网络等资源，并把这些数据上报到观测云。

## 配置 {#config}

### 前置条件 {#requirements}

- 创建用户

在 vCenter 的管理界面中创建一个用户 `datakit`，并赋予 `read-only` 权限，并应用到需要监控的资源上。如果需要监控所有子对象，可以勾选 `Propagate to children` 选项。

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，重启 DataKit 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

## 指标 {#metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

<!-- markdownlint-disable MD046 -->
???+ attention

    下面的指标并非全部被采集到，具体可参阅[数据集合级别](https://docs.vmware.com/cn/VMware-vSphere/7.0/com.vmware.vsphere.monitoring.doc/GUID-25800DE4-68E5-41CC-82D9-8811E27924BC.html){:target="_blank"}中的说明。

<!-- markdownlint-enable -->
{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

<!-- markdownlint-disable MD024 -->
## 对象 {#object}

{{ range $i, $o := .Measurements }}

{{if eq $o.Type "object"}}

### `{{$o.Name}}`

{{$o.Desc}}

- 标签

{{$o.TagsMarkdownTable}}

- 指标列表

{{$o.FieldsMarkdownTable}}
{{end}}

{{ end }}

<!-- markdownlint-enable -->
## 日志 {#logging}

{{ range $i, $l := .Measurements }}

{{if eq $l.Type "logging"}}

### `{{$l.Name}}`

{{$l.Desc}}

- 标签

{{$l.TagsMarkdownTable}}

- 字段列表

{{$l.FieldsMarkdownTable}}
{{end}}

{{ end }}