---
title     : 'SQLServer'
summary   : '采集 SQLServer 的指标数据'
tags:
  - '数据库'
__int_icon      : 'icon/sqlserver'
dashboard :
  - desc  : 'SQLServer'
    path  : 'dashboard/zh/sqlserver'
  - desc  : 'SQLServer-v2'
    path  : 'dashboard/zh/sqlserver-v2'
monitor   :
  - desc  : 'SQLServer 监控器'
    path  : 'monitor/zh/sqlserver'
  - desc  : 'SQLServer 监控器-v2'
    path  : 'monitor/zh/sqlserver-v2'
---

{{.AvailableArchs}}

---

SQL Server 采集器采集 SQL Server `waitstats`、`database_io` 等相关指标

## 配置 {#config}

SQL Server 版本 >= 2008, 已测试的版本：

- [x] 2017
- [x] 2019
- [x] 2022

### 前置条件 {#requrements}

- 创建用户：

Linux、Windows:

```sql
USE master;
GO
CREATE LOGIN [datakit] WITH PASSWORD = N'yourpassword';
GO
GRANT VIEW SERVER STATE TO [datakit];
GO
GRANT VIEW ANY DEFINITION TO [datakit];
GO
```

Aliyun RDS SQL Server:

```sql
USE master;
GO
CREATE LOGIN [datakit] WITH PASSWORD = N'yourpassword';
GO
```

<!-- markdownlint-disable MD046 -->
???+ note "注意事项"

    注意，执行上述操作需要相应权限的帐号，否则可能会导致用户创建失败或者授权失败。

    - 自建的 SQL Server 需要具备 WITH GRANT OPTION、CREATE ANY LOGIN、CREATE ANY USER、ALTER ANY LOGIN 权限的用户，也可以直接使用具有 sysadmin 角色的用户或者 local 用户授权。
    - RDS for SQL Server 则需要使用高权限账号进行授权。

---

### 采集器配置 {#input-config}

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

### 日志采集配置 {#logging-config}

<!-- markdownlint-disable MD046 -->
???+ note

    必须将 DataKit 安装在 SQLServer 所在主机才能采集日志。
<!-- markdownlint-enable -->

如需采集 SQL Server 的日志，可在 *{{.InputName}}.conf* 中 将 `files` 打开，并写入 SQL Server 日志文件的绝对路径。比如：

```toml hl_lines="4"
[[inputs.sqlserver]]
    ...
    [inputs.sqlserver.log]
        files = ["/var/opt/mssql/log/error.log"]
```

开启日志采集以后，默认会产生日志来源（*source*）为 `sqlserver` 的日志。

## 数据库监控 (DBM) {#dbm}

数据库监控（Database Monitoring，DBM）功能提供对 SQL Server 数据库性能的深度可见性，通过收集查询指标、活动会话和执行计划来帮助分析和优化数据库性能。

### 启用 DBM {#dbm-enabled}

```toml
[inputs.sqlserver.dbm]
  # 启用 DBM 功能（默认：false）
  enabled = true
  # 存储过程字符数限制（默认：500）
  stored_procedure_characters_limit = 500
```

### 查询指标 (Metric) {#dbm-metric}

收集 SQL 查询的累积执行统计信息，按查询签名、执行计划哈希和数据库进行聚合。包含执行次数、CPU 时间、逻辑读取、物理读取、等待时间等，通过计算导数指标（两次采集之间的差值）来反映查询的实际执行情况。

> 注意：只有连续两个采集窗口都出现的查询才会上报指标。首次采集到的查询仅作为基线，不会上报指标。

```toml
[inputs.sqlserver.dbm.metric]
  # 启用查询指标采集（默认： true）
  enabled = true
  # 采集间隔（默认： 60s）
  collection_interval = "60s"
  # sys.dm_exec_query_stats 查询的最大行数限制（默认： 10000）
  # 这限制了聚合前的初始查询结果大小
  dm_exec_query_stats_row_limit = 10000
  # 每个采集间隔报告的最大查询数（默认： 500）
  # 按导数执行时间排序后，只报告前 N 个查询
  max_queries = 500
  # 查询过滤的回看窗口（秒，默认： 300）
  # 只收集在此时间窗口内执行的查询（基于 last_execution_time + last_elapsed_time）
  lookback_window = 300
  # 启用执行计划采集（默认： true）
  plan_enabled = true
  # 执行计划对象缓存 TTL（默认： 1h）
  plan_cache_ttl = "1h"
  # 执行计划采集的最大运行时间（秒，默认： 30）
  # 如果采集时间超过此值，将跳过执行计划采集
  max_run_time = 30
```

**执行计划 (Plan)**：当 `plan_enabled = true` 时，会同时采集执行计划。

### 活动查询 (Activity) {#dbm-activity}

收集当前正在执行的查询和活动会话信息。记录会话状态、等待事件、阻塞信息、SQL 文本（已脱敏）等，用于实时监控数据库的当前活动状态。

**会话指标 (Session)**：基于活动查询数据，会自动生成会话指标。

**连接指标 (Connection)**：独立查询数据库连接信息，按用户名、状态、数据库等维度统计连接数。

```toml
[inputs.sqlserver.dbm.activity]
  # 启用活动查询信息采集（默认： true）
  enabled = true
  # 活动指标采集间隔（默认： 10s）
  collection_interval = "10s"
  # sys.dm_exec_sessions 查询的最大行数限制（默认： 1000）
  dm_exec_sessions_row_limit = 1000
```

## 指标 {#metric}

以下所有数据采集，默认会追加全局选举 tag，也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

<!-- markdownlint-disable MD024 -->
{{ range $i, $m := .Measurements }}
{{if eq $m.Type "metric"}}
### `{{$m.Name}}`

{{$m.DescZh}}

{{$m.MarkdownTable}}

{{ end }}
{{ end }}

## 对象 {#object}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "object"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}
{{end}}

{{ end }}


### `message` 指标字段结构 {#message-struct}

`message` 字段基本结构如下：

```json
{
  "setting": [ # settings information
    {
      "name": "recovery interval (min)",
      "value": "123",
      "value_in_use": "0",
      "maximum": "10",
      "minimum": "0",
      "is_dynamic": true,
      "is_advanced": true
    }
    ...
  ],

  "databases": [ # databases information
    {
      "name": "db1",
      "owner_name": "dbo",
      "collation": "SQL_Latin1_General_CP1_CI_AS",
      "schemas": [
        {
          "name": "schema1",
          "owner_name": "dbo",
          "tables": [
            {
              "name": "table1",
              "columns": [], # columns information
              "indexes": [], # indexes information
              "foreign_keys": [], # foreign keys information
              "partitions": { # partitions information
                "partition_count": 1
              } 
            }
            ...
          ]
        }
      ]
    }
    ...
  ]
}
```

#### `setting` {#setting}

  `setting` 字段中的数据来源于 `sys.configurations` 系统视图，该表包含了 SQL Server 服务器的全局变量信息，详细字段可以参考 [SQL Server 文档](https://docs.microsoft.com/en-us/sql/relational-databases/system-catalog-views/sys-configurations-transact-sql?view=sql-server-ver16){:target="_blank"}。

#### `databases` {#databases}

`databases` 字段保存了 SQL Server 服务器上所有数据库的信息，每个数据库的信息如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | 数据库名称                                         | string |
| `owner_name`        | 数据库所有者名称（如 dbo）      | string |
| `collation`        | 数据库默认排序规则（如 SQL_Latin1_General_CP1_CI_AS）      | string |
| `schemas`        | 包含表信息的列表      | list |

#### `schemas` {#schemas}

`schemas` 字段保存了数据库中所有架构的信息，每个架构的信息如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | 架构名称                                         | string |
| `owner_name`        | 架构所有者名称（如 dbo）      | string |
| `tables`        | 包含表信息的列表      | list |

`tables` 字段保存了架构中包含的所有表的信息，每个表的信息如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | 表名称                                         | string |
| `columns`        | 包含列信息的列表      | list |
| `indexes`        | 包含索引信息的列表      | list |
| `foreign_keys`        | 包含外键信息的列表      | list |
| `partitions`        | 包含分区信息的列表      | dict |

`tables.columns` 字段保存了表包含的所有列的信息，每个列的信息如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | 列名称                                         | string |
| `data_type`        | 列数据类型      | string |
| `nullable`        | 列是否为空      | string |
| `default`        | 列默认值      | string |

`tables.indexes` 字段保存了表包含的所有索引的信息，每个索引的信息如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | 索引名称                                         | string |
| `type`        | 索引类型      | string |
| `is_unique`        | 索引是否唯一      | string |
| `is_primary_key`        | 索引是否为主键      | string |
| `column_names`        | 索引包含的列名称      | string |
| `is_disabled`        | 索引是否被禁用      | string |
| `is_unique_constraint`        | 索引是否为唯一约束      | string |

`tables.foreign_keys` 字段保存了表包含的所有外键的信息，每个外键的信息如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `foreign_key_name`        | 外键名称                                         | string |
| `referencing_table`        | 引用表名称                                         | string |
| `referenced_table`        | 被引用表名称                                         | string |
| `referencing_columns`        | 引用列名称                                         | string |
| `referenced_columns`        | 被引用列名称                                         | string |
| `update_action`        | 更新时的级联操作      | string |
| `delete_action`        | 删除时的级联操作      | string |

`tables.partitions` 字段保存了表包含的所有分区的数量，具体字段描述如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `partition_count`        | 分区数量                                         | number |

## 日志 {#logging}

以下指标集均以日志形式收集，所有日志等级均为 `info`。

{{ range $i, $m := .Measurements }}
{{if eq $m.Type "logging"}}
### `{{$m.Name}}`

{{$m.DescZh}}

{{$m.MarkdownTable}}

{{ end }}
{{ end }}
<!-- markdownlint-enable -->

### 日志 Pipeline 功能切割字段说明 {#pipeline}

SQL Server 通用日志文本示例：

```log
2021-05-28 10:46:07.78 spid10s     0 transactions rolled back in database 'msdb' (4:0). This is an informational message only. No user action is required
```

切割后的字段列表如下：

| 字段名      | 字段值                   | 说明                        |
|----------|-----------------------|---------------------------|
| `msg`    | `spid...`             | 日志内容                      |
| `time`   | `1622169967780000000` | 纳秒时间戳（作为行协议时间）            |
| `origin` | `spid10s`             | 源                         |
| `status` | `info`                | 由于日志没有明确字段说明日志等级，默认为 info |
