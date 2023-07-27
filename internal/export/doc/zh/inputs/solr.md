---
title     : 'Solr'
summary   : '采集 Solr 的指标数据'
__int_icon      : 'icon/solr'
dashboard :
  - desc  : 'Solr'
    path  : 'dashboard/zh/solr'
monitor   :
  - desc  : 'Solr'
    path  : 'monitor/zh/solr'
---

<!-- markdownlint-disable MD025 -->
# Solr
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Solr 采集器，用于采集 Solr Cache 和 Request Times 等的统计信息。

## 配置 {#config}

### 前置条件 {#requrements}

- DataKit 使用 Solr Metrics API 采集指标数据，支持 Solr 7.0 及以上版本。
- 也可用于 Solr 6.6，但指标数据不完整。

已测试的版本：

- [x] 8.11.2
- [x] 7.0.0

### 采集器配置 {#input-config}

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

---

如需采集 Solr 的日志，可在 *{{.InputName}}.conf* 中 将 `files` 打开，并写入 Solr 日志文件的绝对路径。比如：

```toml
[inputs.solr.log]
    # 填入绝对路径
    files = ["/path/to/demo.log"]
```

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

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## 日志 {#logging}

切割日志示例：

```log
2013-10-01 12:33:08.319 INFO (org.apache.solr.core.SolrCore) [collection1] webapp.reporter
```

切割后字段：

| 字段名     | 字段值                          |
| --------   | -----------------------------   |
| `Reporter` | `webapp.reporter`               |
| `status`   | `INFO`                          |
| `thread`   | `org.apache.solr.core.SolrCore` |
| `time`     | `1380630788319000000`           |
