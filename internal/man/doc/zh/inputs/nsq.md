---
title     : 'NSQ'
summary   : '采集 NSQ 的指标数据'
__int_icon      : 'icon/nsq'
dashboard :
  - desc  : 'NSQ'
    path  : 'dashboard/zh/nsq'
monitor   :
  - desc  : 'NSQ'
    path  : 'monitor/zh/nsq'
---

<!-- markdownlint-disable MD025 -->
# NSQ
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

采集 NSQ 运行数据并以指标的方式上报到观测云。


## 配置 {#config}

### 前置条件 {#requirements}

NSQ 版本 >= 1.0.0

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，[重启 DataKit](datakit-service-how-to.md#manage-service) 即可。

    ???+ tip "NSQ 采集器提供两种配置方式，分别为 `lookupd` 和 `nsqd`"
    
        - `lookupd`：配置 NSQ 集群的 `lookupd` 地址，采集器会自动发现 NSQ Server 并采集数据，扩展性更佳
        - `nsqd`：配置固定的 NSQ Daemon（`nsqd`）地址列表，采集器只会采集该列表的 NSQ Server 数据
        
        以上两种配置方式是互斥的，**`lookupd` 优先级更高，推荐使用 `lookupd` 配置方式**。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

## 指标 {#metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
