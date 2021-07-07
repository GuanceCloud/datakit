{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# 简介

Microsoft IIS 采集器

## 前置条件

操作系统要求:

* Windows Vista 以上版本 (不包含 Windows Vista)
* Windows Server 2008 R2 及以上版本

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}} 
```

配置好后，重启 DataKit 即可。

## 指标集

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
  [inputs.{{.InputName}}.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
    # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

### 日志采集

如需采集 IIS 的日志，将配置中 log 相关的配置打开，如：

```toml
[inputs.iis.log]
    # 填入绝对路径
    files = ["C:/inetpub/logs/LogFiles/W3SVC1/*"] 
```
