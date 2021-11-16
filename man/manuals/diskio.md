{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

diskio 采集器用于磁盘流量和时间的指标的采集

## 前置条件

对于部分旧版本 Windows 操作系统，如若遇到 Datakit 报错： **"The system cannot find the file specified."**

请以管理员身份运行 PowerShell，并执行：

```powershell
diskperf -Y
```

在执行成功后需要重启 Datakit 服务。

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

支持以环境变量的方式修改配置参数（只在 DataKit 以 K8s daemonset 方式运行时生效，主机部署的 DataKit 不支持此功能）：

| 环境变量名                            | 对应的配置参数项     | 参数示例                                                     |
| :---                                  | ---                  | ---                                                          |
| `ENV_INPUT_DISKIO_SKIP_SERIAL_NUMBER` | `skip_serial_number` | `true`/`false`                                               |
| `ENV_INPUT_DISKIO_TAGS`               | `tags`               | `tag1=value1,tag2=value2` 如果配置文件中有同名 tag，会覆盖它 |

## 指标集

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[[inputs.{{.InputName}}.tags]]` 另择 host 来命名。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
