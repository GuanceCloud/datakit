{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：{{.AvailableArchs}}

# 简介

Redis 指标采集器，采集以下数据：

- 开启 AOF 数据持久化，会收集相关指标
- RDB 数据持久化指标
- Slowlog 监控指标
- bigkey scan 监控
- 主从replication

> Redis 开发测试版本为 v5.04, v6.0+ 待支持

## 前置条件

暂无

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

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

## 日志采集
需要采集redis日志，需要开启Redis `redis.config`中日志文件输出配置
```
logfile "/usr/redis/log/redis.log"
```

如需采集 redis 的日志，可在 {{.InputName}}.conf 中log相关的配置打开

**注意**
- 日志路径需要填入绝对路径
- 在使用日志采集时，需要将datakit安装在redis服务同一台主机中，或使用其它方式将日志挂载到外部系统中
