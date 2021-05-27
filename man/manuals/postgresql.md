{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# 简介

Postgresql 采集器可以从 Postgresql 实例中采集实例运行状态指标，并将指标采集到 DataFlux ，帮助你监控分析 Postgresql 各种异常情况

## 前置条件

暂无

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 指标集

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[[inputs.{{.InputName}}.tags]]` 另择 host 来命名。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## 日志采集
如需采集 Postgresql 的日志，可在{{.InputName}}.conf 中 将 `files` 打开，并写入 Postgresql 日志文件的绝对路径。比如:
```
[inputs.postgresql.log]
files = ["/tmp/pgsql/postgresql.log"]
```
开启日志采集后，默认会产生日志来源(`source`)为`postgresql`的日志。
### 注意
- 日志采集仅支持已安装 DataKit 主机上的日志。
- Postgresql 日志默认是输出至`stderr`，如需开启文件日志，可在配置文件进行配置:

```
logging_collector = on    # 开启日志写入文件功能
                          
log_directory = 'pg_log'  # 设置文件存放目录，绝对路径或相对路径(相对PGDATA)

log_filename = 'pg.log'   # 日志文件名称
log_statement = 'all'     # 记录所有查询

#log_duration = on
log_line_prefix= '%m [%p] %d [%a] %u [%h] %c ' # 日志行前缀
log_file_mode = 0644
## For Windows
#log_destination = 'eventlog'
```
更多配置，请参考[官方文档](https://www.postgresql.org/docs/11/runtime-config-logging.html)。