{{.CSS}}
# DataKit 自身指标
---

{{.AvailableArchs}}

---

self 采集器用于 DataKit 自身基本信息的采集，包括运行环境信息、CPU、内存占用情况等。

## 前置条件 {#reqirement}

暂无

## 配置 {#config}

self 采集器会自动运行，无需配置，且无法关闭。

## 指标 {#measurements}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## 延申阅读 {#more-reading}

- [主机采集器](hostobject.md)
