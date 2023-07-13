---
title: 'DataKit 自身指标采集'
summary: '采集 Datakit 自身运行指标'
__int_icon: 'icon/dk'
dashboard:
  - desc: 'Datakit 内置视图'
    path: 'dashboard/zh/dk'

monitor:
  - desc: '暂无'
    path: '-'
---

<!-- markdownlint-disable MD025 -->
# DataKit 自身指标
<!-- markdownlint-enable -->

---

{{.AvailableArchs}} · [:octicons-tag-24: Version-1.9.2](changelog.md#cl-1.9.2)

---

Datakit 采集器用于自身基本信息的采集，包括运行环境信息、CPU、内存占用、各个核心模块指标等。

## 配置 {#config}

Datakit 启动后。默认会暴露一些 [Prometheus 指标](datakit-metrics.md)，没有额外的操作需要执行，本采集器也是默认启动的，替代了之前的 `self` 采集器。

<!-- markdownlint-disable MD046 -->
=== "主机部署"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，重启 DataKit 即可。

=== "Kubernetes"

    Kubernetes 中支持以环境变量的方式修改配置参数：

    | 环境变量名                        | 说明                            | 参数示例                                                                               |
    | :---                              | ---                             | ---                                                                                    |
    | `ENV_INPUT_DK_ENABLE_ALL_METRICS` | 开启所有指标采集                | 任意非空字符串，如 `on/yes/`                                                           |
    | `ENV_INPUT_DK_ADD_METRICS`        | 追加指标列表（JSON 数组）       | `["datakit_io_.*", "datakit_pipeline_.*"]`，可用的指标名参见[这里](datakit-metrics.md) |
    | `ENV_INPUT_DK_ONLY_METRICS`       | **只开启**指定指标（JSON 数组） | `["datakit_io_.*", "datakit_pipeline_.*"]`                                             |
<!-- markdownlint-enable -->

## 指标 {#metric}

Datakit 自身指标主要是一些 Prometheus 指标，其文档参见[这里](datakit-metrics.md)

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
