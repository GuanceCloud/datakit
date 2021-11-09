{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

net 采集器用于采集主机网络信息，如各网络接口的流量信息等。对于 Linux 将采集系统范围 TCP 和 UDP 统计信息。

## 前置条件

暂无

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

支持以环境变量的方式修改配置参数（只在 DataKit 以 K8s daemonset 方式运行时生效，主机部署的 DataKit 不支持此功能）：

| 环境变量名                                | 对应的配置参数项            | 参数示例                                                     |
| :---                                      | ---                         | ---                                                          |
| `ENV_INPUT_NET_IGNORE_PROTOCOL_STATS`     | `ignore_protocol_stats`     | `true`/`false`                                               |
| `ENV_INPUT_NET_ENABLE_VIRTUAL_INTERFACES` | `enable_virtual_interfaces` | `true`/`false`                                               |
| `ENV_INPUT_NET_TAGS`                      | `tags`                      | `tag1=value1,tag2=value2` 如果配置文件中有同名 tag，会覆盖它 |

## 指标集

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
