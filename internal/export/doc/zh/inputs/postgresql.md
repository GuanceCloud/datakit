---
title     : 'PostgreSQL'
summary   : '采集 PostgreSQL 的指标数据'
tags:
  - '数据库'
__int_icon      : 'icon/postgresql'
dashboard :
  - desc  : 'PostgrepSQL'
    path  : 'dashboard/zh/postgresql'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

{{.AvailableArchs}}

---

PostgreSQL 采集器可以从 PostgreSQL 实例中采集实例运行状态指标，并将指标采集到<<<custom_key.brand_name>>>，帮助监控分析 PostgreSQL 各种异常情况。

## 配置 {#config}

### 前置条件 {#reqirement}

- PostgreSQL 版本 >= 9.0
- 创建监控帐号

```sql
-- PostgreSQL >= 10
create user datakit with password '<PASSWORD>';
grant pg_monitor to datakit;
grant SELECT ON pg_stat_database to datakit;

-- PostgreSQL < 10
create user datakit with password '<PASSWORD>';
grant SELECT ON pg_stat_database to datakit;
```

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机部署"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

## 指标 {#metric}

以下所有数据采集，默认会追加全局选举 tag，也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
{{ end }}

## 自定义对象 {#object}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "custom_object"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## 日志 {#logging}

- PostgreSQL 日志默认是输出至 `stderr`，如需开启文件日志，可在 PostgreSQL 的配置文件 `/etc/postgresql/<VERSION>/main/postgresql.conf` ， 进行如下配置：

```toml
logging_collector = on    # 开启日志写入文件功能

log_directory = 'pg_log'  # 设置文件存放目录，绝对路径或相对路径（相对 PGDATA）

log_filename = 'pg.log'   # 日志文件名称
log_statement = 'all'     # 记录所有查询

#log_duration = on
log_line_prefix= '%m [%p] %d [%a] %u [%h] %c ' # 日志行前缀
log_file_mode = 0644

# For Windows
#log_destination = 'eventlog'
```

更多配置，请参考[官方文档](https://www.postgresql.org/docs/11/runtime-config-logging.html){:target="_blank"}。

- PostgreSQL 采集器默认是未开启日志采集功能，可在 *conf.d/{{.Catalog}}/{{.InputName}}.conf* 中 将 `files` 打开，并写入 PostgreSQL 日志文件的绝对路径。比如：

```toml
[[inputs.postgresql]]

  ...

  [inputs.postgresql.log]
  files = ["/tmp/pgsql/postgresql.log"]
```

开启日志采集后，默认会产生日志来源（`source`）为 PostgreSQL 的日志。

> 注意：日志采集仅支持已安装 DataKit 主机上的日志。

### 日志 Pipeline 切割 {#pipeline}

原始日志为

``` log
2021-05-31 15:23:45.110 CST [74305] test [pgAdmin 4 - DB:postgres] postgres [127.0.0.1] 60b48f01.12241 LOG:  statement:
        SELECT psd.*, 2^31 - age(datfrozenxid) as wraparound, pg_database_size(psd.datname) as pg_database_size
        FROM pg_stat_database psd
        JOIN pg_database pd ON psd.datname = pd.datname
        WHERE psd.datname not ilike 'template%'   AND psd.datname not ilike 'rdsadmin'
        AND psd.datname not ilike 'azure_maintenance'   AND psd.datname not ilike 'postgres'
```

切割后的字段说明：

| 字段名             | 字段值                    | 说明                                                        |
| ---                | ---                       | ---                                                         |
| `application_name` | `pgAdmin 4 - DB:postgres` | 连接当前数据库的应用的名称                                  |
| `db_name`          | `test`                    | 访问的数据库                                                |
| `process_id`       | `74305`                   | 当前连接的客户端进程 ID                                     |
| `remote_host`      | `127.0.0.1`               | 客户端的地址                                                |
| `session_id`       | `60b48f01.12241`          | 当前会话的 ID                                               |
| `user`             | `postgres`                | 当前访问用户名                                              |
| `status`           | `LOG`                     | 当前日志的级别（LOG,ERROR,FATAL,PANIC,WARNING,NOTICE,INFO） |
| `time`             | `1622445825110000000`     | 日志产生时间                                                |

## FAQ {#faq}

<!-- markdownlint-disable MD013 -->
### 缺失指标 {#faq-missing-relation-metrics}

对于 `postgresql_lock/postgresql_stat/postgresql_index/postgresql_size/postgresql_statio` 这些指标，需要开启配置文件中的 `relations` 字段。如果这些指标存在部分缺失，可能是因为相关指标不存在数据导致的。

<!-- markdownlint-enable -->
