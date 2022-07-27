{{.CSS}}
# CoreDNS
---

- 操作系统支持：{{.AvailableArchs}}

CoreDNS 采集器用于采集 CoreDNS 相关的指标数据。

## 前置条件 {#requirements}

- CoreDNS [配置](https://coredns.io/plugins/metrics/){:target="_blank"}启用 `prometheus` 插件

## 配置 {#input-config}

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 到 `conf.d/{{.Catalog}}` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}} 
```

配置好后，重启 DataKit 即可。

## 指标集 {#metrics}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
