{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

solr 采集器，用于采集 solr cache 和 request times 等的统计信息。

## 前置条件

DataKit 使用 Solr Metrics API 采集指标数据，支持 Solr 7.0 及以上版本。可用于 Solr 6.6，但指标数据不完整。

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

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

## 日志采集

如需采集 Solr 的日志，可在 {{.InputName}}.conf 中 将 `files` 打开，并写入 Solr 日志文件的绝对路径。比如：

```toml
[inputs.solr.log]
    # 填入绝对路径
    files = ["/path/to/demo.log"] 
```
