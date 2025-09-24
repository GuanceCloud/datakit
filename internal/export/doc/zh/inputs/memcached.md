---
title     : 'Memcached'
summary   : '采集 Memcached 的指标数据'
tags:
  - '缓存'
  - '中间件'
__int_icon      : 'icon/memcached'
dashboard :
  - desc  : 'Memcached'
    path  : 'dashboard/zh/memcached'
monitor   :
  - desc  : 'Memcached'
    path  : 'monitor/zh/memcached' 
---

{{.AvailableArchs}}

---

Memcached 采集器可以从 Memcached 实例中采集实例运行状态指标，并将指标采集到<<<custom_key.brand_name>>>，帮助监控分析 Memcached 各种异常情况。

## 配置 {#config}

### 前置条件 {#requirements}

- Memcached 版本 >= `1.5.0`。已测试的版本：
    - [x] 1.5.x
    - [x] 1.6.x

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

## 指标 {#metric}

以下所有数据采集，默认会追加全局选举 tag，也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}
