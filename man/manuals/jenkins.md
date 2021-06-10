{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`


# {{.InputName}}

Jenkins 采集器是通过插件 `Metrics` 采集数据监控 Jenkins，包括但不限于任务数，系统 cpu 使用，`jvm cpu`使用等

## 前置条件

- JenKins 版本 >= 2.277.4

- 安装 JenKins [参见](https://www.jenkins.io/doc/book/installing/)
      
- 下载 `Metric` 插件，[管理插件页面](https://www.jenkins.io/doc/book/managing/plugins/),[Metric 插件页面](https://plugins.jenkins.io/metrics/)

- 在 JenKins 管理页面 `your_manage_host/configure` 生成 `Metric Access keys`

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

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}


## 日志采集

如需采集 JenKins 的日志，可在 {{.InputName}}.conf 中 将 `files` 打开，并写入 JenKins 日志文件的绝对路径。比如：

```toml
    [[inputs.JenKins]]
      ...
      [inputs.JenKins.log]
        files = ["/var/log/jenkins/jenkins.log"]
```

  
开启日志采集以后，默认会产生日志来源（`source`）为 `jenkins` 的日志。

>注意：必须将 DataKit 安装在 JenKins 所在主机才能采集 JenKins 日志

## 日志 pipeline 功能切割字段说明

- JenKins 通用日志切割

通用日志文本示例:
```
2021-05-18 03:08:58.053+0000 [id=32]	INFO	jenkins.InitReactorRunner$1#onAttained: Started all plugins
```

切割后的字段列表如下：

| 字段名  |  字段值  | 说明 |
| ---    | ---     | --- |
|  status   | info     | 日志等级 |
|  id   | 32     | id |
|  time   | 1621278538000000000     | 纳秒时间戳（作为行协议时间）|
