---
title     : 'Kingbase'
summary   : '采集 Kingbase 的指标数据'
tags:
  - '数据库'
__int_icon      : 'icon/kingbase'
dashboard :
  - desc  : 'Kingbase'
    path  : 'dashboard/zh/kingbase'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

{{.AvailableArchs}}

---

Kingbase 采集器可以从 Kingbase 实例中采集实例运行状态指标。

## 配置 {#config}

### 前置条件 {#reqirement}

- 创建监控帐号

```sql
-- 创建监控用户
CREATE USER datakit with password 'datakit';
-- 授权
GRANT sys_monitor TO datakit;
```

- 开启 sys_stat_statements 扩展日志记录

编辑 `data` 目录下的 `kingbase.conf` 配置文件，修改 `sys_stat_statements.track` 的值为 `top`。

```bash
# 跟踪统计 SQL 语句访问，推荐 top，默认 none
sys_stat_statements.track = 'top'
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

- 如需开启 Kingbase 的运行日志，可在 data 目录下的 Kingbase 配置文件 `kingbase.conf` ， 进行如下配置：

```toml
log_destination = 'stderr'

logging_collector = on
log_directory = 'sys_log'
```

- Kingbase 采集器默认是未开启日志采集功能，可在 *conf.d/{{.Catalog}}/{{.InputName}}.conf* 中 将 `files` 打开，并写入 Kingbase 日志文件的绝对路径。比如：

```toml
[[inputs.kingbase]]

  ...

  [inputs.kingbases.log]
    files = ["/tmp/kingbase.log"]
```

开启日志采集后，默认会产生日志来源（`source`）为 `kingbase` 的日志。

> 注意：日志采集仅支持已安装 DataKit 主机上的日志。

### 日志 Pipeline 切割 {#pipeline}

原始日志为

``` log
2025-06-17 13:07:10.952 UTC [999] ERROR:  relation "sys_stat_activity" does not exist at character 240
```

切割后的字段说明：

| 字段名              | 字段值                                                           | 说明                                                       |
| ---                | ---                                                              | ---                                                        |
| `msg`              | `relation "sys_stat_activity" does not exist at character 240`   | 日志内容                                                    |
| `db_name`          | `test`                                                           | 访问的数据库                                                |
| `process_id`       | `999`                                                            | 当前连接的客户端进程 ID                                      |
| `status`           | `ERROR`                                                          | 当前日志的级别（LOG,ERROR,FATAL,PANIC,WARNING,NOTICE,INFO）  |
| `time`             | `1750136961776000000`                                            | 日志产生时间                                                |
