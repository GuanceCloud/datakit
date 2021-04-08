- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}

# 简介

system 采集器

## 前置条件

无

## 配置

进入 DataKit 安装目录下的 `conf.d/host` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 指标集

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
