{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# 简介

这只是个采集器开发示例。

注意

- 这里进行文档描述的时候，一些英文描述，比如 `Oracle`，不要写成 `oracle`，`NGINX` 不应该写成 `nginx`，至少也应该写成 `Nginx`。这写都是一些专用名词
- 中英文之间用空格，比如：`这是一个 Oracle 采集器`。不要写成 `这是一个Oracle采集器`
- 不要滥用代码字体，比如 DataFlux 不要写成 `DataFlux`
- 这里统一用中文标点符号

## 前置条件

注意：

- 这里尽量说明下必要前置条件，比如 Redis 版本要求，需要额外安装的软件等等
- 这里不用加 `安装 DataKit` 这个条件，实属废话

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}} 
```

Sample 注意事项：

1. 这里不要写太多英文描述
2. 请做好 sample 格式化（对齐好格式），不要用 tab 缩进，统一用空格。因为用户终端 tab 宽度显示可能有差异
3. 一些分段点，如 `[[inputs.oracle.options]]`，不要加注释，因为部分用户改完 `options` 下的配置后，可能会忘记打开这一行的配置，导致解析失败（用户甚至不知道需要打开这一行）
4. 一些默认打开的选项，不要注释掉了，不然用户使用的时候，还要手动去打开
5. 当某些采集器无需额外配置时，在 sample 中加一行 `# 这里无需额外配置`，让用户知道不用其它配置了
6. 总体原则是，配置项能不注释就不注释

配置好后，重启 DataKit 即可。

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

## 指标

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`
{{$m.Desc}}

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

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

- 指标列表

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

- 指标列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}
