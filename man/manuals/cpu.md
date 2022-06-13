{{.CSS}}
# CPU
---

- DataKit 版本：{{.Version}}
- 操作系统支持：`{{.AvailableArchs}}`

CPU 采集器用于系统 CPU 使用率的采集。

![](imgs/input-cpu-1.png) 

## 前置条件

暂无

## 配置  {#input-config}

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 指标查看

数据采集上来后，即可在页面上看到如下 CPU 指标数据：

![](imgs/input-cpu-2.png) 

### 通过环境变量修改配置参数 {#envs}

支持以环境变量的方式修改配置参数（只在 Daemonset 方式运行时生效）：

| 环境变量名                                  | 对应的配置参数项              | 参数示例                                                                              |
| :---                                        | ---                           | ---                                                                                   |
| `ENV_INPUT_CPU_PERCPU`                      | `percpu`                      | `true/false`                                                                          |
| `ENV_INPUT_CPU_ENABLE_TEMPERATURE`          | `enable_temperature`          | `true/false`                                                                          |
| `ENV_INPUT_CPU_TAGS`                        | `tags`                        | `tag1=value1,tag2=value2` 如果配置文件中有同名 tag，会覆盖它                          |
| `ENV_INPUT_CPU_INTERVAL`                    | `interval`                    | `10s`                                                                                 |
| `ENV_INPUT_CPU_DISABLE_TEMPERATURE_COLLECT` | `disable_temperature_collect` | `false/true`。给任意字符串就认为是 `true`，没定义就是 `false`。                       |
| `ENV_INPUT_CPU_ENABLE_LOAD5S`               | `enable_load5s`               | `false/true`。给任意字符串就认为是。给任意字符串就认为是 `true`，没定义就是 `false`。 |

## 指标集 {#measurements}

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

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## 场景试图

<场景 - 新建仪表板 - 内置模板库 - CPU>

## 异常检测

<监控 - 模板新建 - 主机检测库>
