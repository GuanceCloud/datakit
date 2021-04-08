{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}

# 简介

mem 采集器用于收集系统内存信息，一些通用的指标：

|字段|描述|
|:---|:---|
|total|主机中 RAM 的总量|
|available|可供程序分配的RAM|  
|available_percent|可供程序分配的RAM百分比|
|used|被程序使用的RAM|  
|used_percent|被程序使用的RAM百分比|

## 前置条件

暂无

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```python
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
