{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

采集 Kubernetes 集群指标数据和对象数据，上报到观测云。

## 配置

DataKit 访问 Kubernetes API 采集各项数据，DataKit 以[daemonset 方式安装和运行](datakit-daemonset-deploy)会自动开启 Kubernetes 采集。

支持以环境变量的方式定制参数

| 环境变量名           | 对应的配置参数项 | 参数示例                                                     |
| :---                 | ---              | ---                                                          |
| `ENV_INPUT_K8S_TAGS` | `tags`           | `tag1=value1,tag2=value2` 如果配置文件中有同名 tag，会覆盖它 |

## 指标

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`
{{$m.Desc}}

-  标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## 对象

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "object"}}

### `{{$m.Name}}`
{{$m.Desc}}

-  标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## 日志

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

### `{{$m.Name}}`
{{$m.Desc}}

-  标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}
