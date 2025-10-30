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

- 开启 `pg_stat_statements` 扩展（可选）

    [PostgreSQL 对象](postgresql.md#object) 采集时，部分指标如 `qps/tps/avg_query_time` 等，需要开启 `pg_stat_statements` 扩展。具体步骤如下：

    - **修改配置文件**

        找到并编辑 PostgreSQL 配置文件（通常位于 `/var/lib/pgsql/data/postgresql.conf` 或 `/etc/postgresql/<版本>/main/postgresql.conf`）：

        ```ini
        # 启用 pg_stat_statements 扩展
        shared_preload_libraries = 'pg_stat_statements'
        
        # 可选配置
        pg_stat_statements.track = 'all'  # 收集所有 SQL 语句
        pg_stat_statements.max = 10000  # 最多收集的 SQL 语句数量
        pg_stat_statements.track_utility = off # 忽略 utility 语句，只跟踪常规 SQL 查询如 SELECT、INSERT、UPDATE、DELETE
        
        ```

    - **重启 PostgreSQL 服务**

        修改好配置文件后，需要重启 PostgreSQL 服务。

    - **数据库中创建扩展**

        连接到目标数据库， 执行以下 SQL：

        ```sql
        CREATE EXTENSION pg_stat_statements;
        ```

    - **验证扩展是否成功开启**

        ```sql
        SELECT * FROM pg_extension WHERE extname = 'pg_stat_statements';
        SELECT * FROM pg_stat_statements LIMIT 10;
        ```

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机部署"

    进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

### 数据库性能指标采集 {#performance-schema}

[:octicons-tag-24: Version-1.84.0](../datakit/changelog-2025.md#cl-1.84.0)

数据库性能指标主要来源于 PostgreSQL 的内置系统视图和扩展插件，其中最核心的包括 `pg_stat_activity` 和 `pg_stat_statements`。这些工具提供了在运行时获取数据库内部执行情况的方法：`pg_stat_activity` 实时展示当前会话的活动状态、执行的查询及等待事件等信息；`pg_stat_statements` 则记录历史 SQL 语句的执行统计数据，包括执行次数、耗时、IO 情况等。

通过这些视图和插件，DataKit 能够采集实时会话活动、历史查询的性能指标统计以及相关执行信息。采集的性能指标数据保存为日志，source 分别为 `postgresql_dbm_metric`、`postgresql_dbm_sample` 和 `postgresql_dbm_activity`。

如需开启，需要执行以下步骤。

- 修改配置文件，开启监控采集

```toml
[[inputs.postgresql]]

  ## Set dbm to true to collect database activity 
  dbm = false

  ## Config dbm metric 
  [inputs.postgresql.dbm_metric]
    enabled = true
  
  ## Config dbm sample 
  [inputs.postgresql.dbm_sample]
    enabled = true  

  ## Config dbm activity
  [inputs.postgresql.dbm_activity]
    enabled = true  

...

```

- PostgreSQL 配置

修改配置文件（如 *postgresql.conf*），配置相关参数：

```toml

shared_preload_libraries = 'pg_stat_statements'
track_activity_query_size = 4096 a# Required for collection of larger queries.

```

- 权限配置

<!-- markdownlint-disable MD046 -->
=== "Postgres >=15"

    账号授权

    ```sql
    ALTER ROLE datakit INHERIT;
    ```

    在每个数据库中执行以下 SQL：

    ```sql
    CREATE SCHEMA datakit;
    GRANT USAGE ON SCHEMA datakit TO datakit;
    GRANT USAGE ON SCHEMA public TO datakit;
    GRANT pg_monitor TO datakit;
    CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

    ```

=== "Postgres >=10"

    在每个数据库中执行以下 SQL：

    ```sql
    CREATE SCHEMA datakit;
    GRANT USAGE ON SCHEMA datakit TO datakit;
    GRANT USAGE ON SCHEMA public TO datakit;
    GRANT pg_monitor TO datakit;
    CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

    ```

=== "Postgres 9.6"

    在每个数据库中执行以下 SQL：

    ```sql
    CREATE SCHEMA datakit;
    GRANT USAGE ON SCHEMA datakit TO datakit;
    GRANT USAGE ON SCHEMA public TO datakit;
    GRANT SELECT ON pg_stat_database TO datakit;
    CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

    CREATE OR REPLACE FUNCTION datakit.pg_stat_activity() RETURNS SETOF pg_stat_activity AS
      $$ SELECT * FROM pg_catalog.pg_stat_activity; $$
    LANGUAGE sql
    SECURITY DEFINER;
    CREATE OR REPLACE FUNCTION datakit.pg_stat_statements() RETURNS SETOF pg_stat_statements AS
        $$ SELECT * FROM pg_stat_statements; $$
    LANGUAGE sql
    SECURITY DEFINER;

    ```

<!-- markdownlint-enable -->

## 指标 {#metric}

以下所有数据采集，默认会追加全局选举 tag，也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

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
  "setting": {
    "DateStyle":"ISO, MDY",
    ...
  },

  "databases": [ # databases information
    {
      "name": "db1",
      "encoding": "utf8",
      "owner": "datakit",
      "schemas": [ # schemas information
        {
          "name": "schema1",
          "owner": "datakit",
          "tables": [ # tables information
            {
              "name": "table1",
              "columns": [], # columns information
              "indexes": [], # indexes information
              "foreign_keys": [], # foreign keys information
            }
            ...
          ]
        },
        ...
      ]
    }
    ...
  ]
}
```

#### `setting` {#setting}

  `setting` 字段中的数据来源于 `pg_settings` 系统视图，用于展示当前数据库系统的配置参数信息，详细信息可以参考 [PostgreSQL 文档](https://www.postgresql.org/docs/current/view-pg-settings.html){:target="_blank"}。

#### `databases` {#databases}

`databases` 字段保存了 `PostgreSQL` 服务器上所有数据库的信息，每个数据库的信息如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | 数据库名称                                           | string |
| `encoding`        | 数据库编码                                        | string |
| `owner`        | 角色名称                                            | string |
| `description`        | 描述文本                                      | string |
| `schemas`        | 包含 `schema` 信息的列表                           | list |

`schemas` 字段包含了数据库中的所有 `schema` 信息，每个 `schema` 的信息如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | `schema` 名称                                           | string |
| `owner`        | 角色名称                                            | string |
| `tables`        | 包含 `table` 信息的列表                           | list |

`tables` 字段包含了数据库中所有表的信息，每个表的信息如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | 表名称                                         | string |
| `owner`        | 角色名称                             | string |
| `has_indexes`        | 是否有索引                             | bool |
| `has_partitions`        | 是否有分区                             | bool |
| `toast_table`        | `toast` 表名称                             | string |
| `partition_key`        | 分区键                             | string |
| `num_partitions`        | 分区数量                             | int64 |
| `foreign_keys`        | 包含外键信息的列表                             | list |
| `columns`        | 包含列信息的列表                             | list |
| `indexes`        | 包含索引信息的列表                             | list |

- `tables.columns`

`columns` 字段包含了数据库中表的所有列的信息，每个列的信息如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | `column` 名称                                           | string |
| `data_type`        | 数据类型                                            | string |
| `nullable`        | 是否可以为空                                            | bool |
| `default`        | 默认值                                            | string |

- `tables.indexes`

`indexes` 字段包含了数据库中表的所有索引的信息，每个索引的信息如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | `index` 名称                                           | string |
| `columns`     | 索引包含的列                                              | list |
| `index_type`  | 索引类型                                                  | string |
| `definition`        | 索引定义                                            | string |
| `is_unique`        | 是否唯一                                            | bool |
| `is_primary`        | 是否主键                                            | bool |
| `is_exclusion`        | 是否为排除约束索引                                            | bool |
| `is_immediate`        | 每条语句执行结束后是否立即检查约束                      | bool |
| `is_valid`        | 是否有效                                            | bool |
| `is_clustered`        | 是否为聚簇索引                                            | bool |
| `is_checkxmin`        | 是否检查 `xmin`                                            | bool |
| `is_ready`        | 是否就绪                                            | bool |
| `is_live`        | 是否活跃                                            | bool |
| `is_replident`        | 是否是行标识索引                                            | bool |
| `is_partial`        | 是否是部分索引                                            | bool |

索引列信息字段 `indexes.columns` 包含了索引中包含的列的信息，每个列的信息如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | 列名称                                         | string |

- `tables.foreign_keys`

`foreign_keys` 字段包含了数据库中表的所有外键的信息，每个外键的信息如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | `foreign_key` 名称                                           | string |
| `definition`        | 外键定义                                            | string |
| `constraint_schema` |外键所属的数据库（通常与表所在数据库一致） | string |
| `column_names` |外键列名称（多个列用逗号分隔，如 user_id, order_id） | string |
| `referenced_table_schema` |引用表所在的数据库 | string |
| `referenced_table_name` |引用表名称 | string |
| `referenced_column_names` |引用列名称（多个列用逗号分隔） | string |
| `update_action` |级联更新规则（如 CASCADE, RESTRICT） | string |
| `delete_action` |级联删除规则（如 CASCADE, SET NULL） | string |

## 日志 {#logging}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}
{{end}}

{{ end }}

## 文件日志 {#file-log}

### 日志采集 {#log}

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

- PostgreSQL 采集器默认是未开启日志采集功能，可在 *conf.d/samples/{{.InputName}}.conf* 中 将 `files` 打开，并写入 PostgreSQL 日志文件的绝对路径。比如：

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
