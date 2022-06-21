{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

Datakit 内嵌的 Zipkin Agent 用于接收，运算，分析 Zipkin Tracing 协议数据。

## Zipkin 文档

- [Quickstart](https://zipkin.io/pages/quickstart.html)
- [Docs](https://zipkin.io/pages/instrumenting.html)
- [Souce Code](https://github.com/openzipkin/zipkin)

## 配置 Zipkin Agent

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

## Tracing 数据

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "tracing"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}
