{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

Redis 指标采集器，采集以下数据：

- 开启 AOF 数据持久化，会收集相关指标
- RDB 数据持久化指标
- Slowlog 监控指标
- bigkey scan 监控
- 主从replication


## 前置条件

- Redis 版本 v5.0+

在采集主从架构下数据时，请配置从节点的主机信息进行数据采集，可以得到主从相关的指标信息。

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
需要采集redis日志，需要开启Redis `redis.config`中日志文件输出配置

```toml
[inputs.redis.log]
    # 日志路径需要填入绝对路径
    files = ["/var/log/redis/*.log"] # 在使用日志采集时，需要将datakit安装在redis服务同一台主机中，或使用其它方式将日志挂载到外部系统中
```

## 日志 pipeline 功能切割字段说明

原始日志为 `122:M 14 May 2019 19:11:40.164 * Background saving terminated with success`

切割后的字段列表如下：

| 字段名  |  字段值  | 说明 |
| ---    | ---     | --- |
|  pid   | 122     | 进程id |
|  role  | "M"    | 角色 |
|  serverity   | "*"     | 服务  |
|  status   | notice     | 日志级别 |
|  msg   | Background saving terminated with success   | 日志内容 |
|  time   | 1557861100164000000     | 纳秒时间戳（作为行协议时间）|
