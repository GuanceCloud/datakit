---
title     : 'System'
summary   : '采集主机系统相关的指标数据'
__int_icon      : 'icon/system'
dashboard :
  - desc  : 'System'
    path  : 'dashboard/zh/system'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# System
<!-- markdownlint-enable -->

<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

System 采集器收集系统负载、正常运行时间、CPU 核心数量以及登录的用户数。

## 配置 {#config}

成功安装 DataKit 并启动后，会默认开启 System 采集器，无需手动开启。

<!-- markdownlint-disable MD046 -->

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    支持以环境变量的方式修改配置参数：

    | 环境变量名              | 对应的配置参数项 | 参数示例                                                     |
    | :---                    | ---              | ---                                                          |
    | `ENV_INPUT_SYSTEM_TAGS` | `tags`           | `tag1=value1,tag2=value2` 如果配置文件中有同名 tag，会覆盖它 |
    | `ENV_INPUT_SYSTEM_INTERVAL` | `interval` | `10s` |

<!-- markdownlint-enable -->

---

## 指标 {#metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

```toml
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
