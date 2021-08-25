{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

进程采集器可以对系统中各种运行的进程进行实施监控， 获取、分析进程运行时各项指标，包括内存使用率、占用CPU时间、进程当前状态等，并根据进程运行时的各项指标信息，用户可以在 DataFlux 中配置相关告警，使用户了解进程的状态，在进程发生故障时，可以及时对发生故障的进程进行维护

## 前置条件

- 进程采集器默认不采集进程指标数据，如需采集指标相关数据，可在 `{{.InputName}}.conf` 中 将 `open_metric` 设置为 `true`。比如：
                              
  ```
      [[inputs.host_processes]]
        ...
         open_metric = true
  ```

## 配置

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
