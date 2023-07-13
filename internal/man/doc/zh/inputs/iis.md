---
title     : 'IIS'
summary   : '采集 IIS 指标数据'
__int_icon      : 'icon/iis'
dashboard :
  - desc  : 'IIS'
    path  : 'dashboard/zh/iis'
monitor   :
  - desc  : 'IIS'
    path  : 'monitor/zh/iis'
---

<!-- markdownlint-disable MD025 -->
# IIS
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

采集 IIS 指标数据。

## 配置 {#config}

### 前置条件 {#requirements}

操作系统要求：

- Windows 7 以上版本（含 Windows 7）
- Windows Server 2008 R2 及以上版本

### 采集器配置 {#input-config}

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}} 
```

配置好后，重启 DataKit 即可。

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
[inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

## 指标 {#metric}

{{ range $i, $m := .Measurements }}

{{if or (eq $m.Type "metric") (eq $m.Type "")}}

### `{{$m.Name}}`
{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## 日志 {#logging}

如需采集 IIS 的日志，将配置中 log 相关的配置打开，如：

```toml
[inputs.iis.log]
    # 填入绝对路径
    files = ["C:/inetpub/logs/LogFiles/W3SVC1/*"] 
```
