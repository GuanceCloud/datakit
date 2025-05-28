---
title     : 'MySQL'
summary   : '采集 MySQL 的指标数据'
tags:
  - '数据库'
__int_icon      : 'icon/mysql'
dashboard :
  - desc  : 'MySQL'
    path  : 'dashboard/zh/mysql'
monitor   :
  - desc  : 'MySQL'
    path  : 'monitor/zh/mysql'
---

{{.AvailableArchs}}

---

MySQL 指标采集，收集以下数据：

- MySQL Global Status 基础数据采集
- Schema 相关数据
- InnoDB 相关指标
- 支持自定义查询数据采集

## 配置 {#config}

### 前置条件 {#requirements}

- MySQL 版本 5.7+
- 创建监控账号（一般情况，需用 MySQL `root` 账号登陆才能创建 MySQL 用户）

```sql
CREATE USER 'datakit'@'localhost' IDENTIFIED BY '<UNIQUEPASSWORD>';

-- MySQL 8.0+ create the datakit user with the caching_sha2_password method
CREATE USER 'datakit'@'localhost' IDENTIFIED WITH caching_sha2_password by '<UNIQUEPASSWORD>';
```

- 授权

```sql
GRANT PROCESS ON *.* TO 'datakit'@'localhost';
GRANT SELECT ON *.* TO 'datakit'@'localhost';
show databases like 'performance_schema';
GRANT SELECT ON performance_schema.* TO 'datakit'@'localhost';
GRANT SELECT ON mysql.user TO 'datakit'@'localhost';
GRANT replication client on *.*  to 'datakit'@'localhost';
```

<!-- markdownlint-disable MD046 -->
???+ note

    - 如用 `localhost` 时发现采集器有如下报错，需要将上述步骤的 `localhost` 换成 `::1`：

        ```text
        `Error 1045: Access denied for user 'datakit'@'localhost' (using password: YES)`
        ```

    - 以上创建、授权操作，均限定了 `datakit` 这个用户，只能在 MySQL 主机上（`localhost`）访问 MySQL。如果需要对 MySQL 进行远程采集，建议将 `localhost` 替换成 `%`（表示 DataKit 可以在任意机器上访问 MySQL），也可用特定的 DataKit 安装机器地址。
<!-- markdownlint-enable -->

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

---

### Binlog 开启 {#binlog}

默认情况下，MySQL Binlog 是不开启的。如果要统计 Binlog 大小，需要开启 MySQL 对应 Binlog 功能：

```sql
-- ON: 开启/OFF: 关闭
SHOW VARIABLES LIKE 'log_bin';
```

Binlog 开启，参见[这个问答](https://stackoverflow.com/questions/40682381/how-do-i-enable-mysql-binary-logging){:target="_blank"}，或者[这个问答](https://serverfault.com/questions/706699/enable-binlog-in-mysql-on-ubuntu){:target="_blank"}

### 数据库性能指标采集 {#performance-schema}

数据库性能指标主要来源于 MySQL 的内置数据库 `performance_schema`, 该数据库提供了一个能够在运行时获取服务器内部执行情况的方法。通过该数据库，DataKit 能够采集历史查询语句的各种指标统计和查询语句的执行计划，以及其他相关性能指标。采集的性能指标数据保存为日志，source 分别为 `mysql_dbm_metric`, `mysql_dbm_sample` 和 `mysql_dbm_activity`。

如需开启，需要执行以下步骤。

- 修改配置文件，开启监控采集

```toml
[[inputs.mysql]]

# 开启数据库性能指标采集
dbm = true

...

# 监控指标配置
[inputs.mysql.dbm_metric]
  enabled = true

# 监控采样配置
[inputs.mysql.dbm_sample]
  enabled = true

# 等待事件采集
[inputs.mysql.dbm_activity]
  enabled = true   
...

```

- MySQL 配置

修改配置文件（如 *mysql.conf*），开启 `MySQL Performance Schema`， 并配置相关参数：

```toml
[mysqld]
performance_schema = on
max_digest_length = 4096
performance_schema_max_digest_length = 4096
performance_schema_max_sql_text_length = 4096
performance-schema-consumer-events-statements-current = on
performance-schema-consumer-events-waits-current = on
performance-schema-consumer-events-statements-history-long = on
performance-schema-consumer-events-statements-history = on

```

- 账号配置

账号授权

```sql
-- MySQL 5.6 & 5.7
GRANT REPLICATION CLIENT ON *.* TO datakit@'%' WITH MAX_USER_CONNECTIONS 5;
GRANT PROCESS ON *.* TO datakit@'%';

-- MySQL >= 8.0
ALTER USER datakit@'%' WITH MAX_USER_CONNECTIONS 5;
GRANT REPLICATION CLIENT ON *.* TO datakit@'%';
GRANT PROCESS ON *.* TO datakit@'%';
```

创建数据库

```sql
CREATE SCHEMA IF NOT EXISTS datakit;
GRANT EXECUTE ON datakit.* to datakit@'%';
GRANT CREATE TEMPORARY TABLES ON datakit.* TO datakit@'%';
```

创建存储过程 `explain_statement`，用于获取 SQL 执行计划

```sql
DELIMITER $$
CREATE PROCEDURE datakit.explain_statement(IN query TEXT)
    SQL SECURITY DEFINER
BEGIN
    SET @explain := CONCAT('EXPLAIN FORMAT=json ', query);
    PREPARE stmt FROM @explain;
    EXECUTE stmt;
    DEALLOCATE PREPARE stmt;
END $$
DELIMITER ;
```

为需要采集执行计划的数据库单独创建存储过程（可选）

```sql
DELIMITER $$
CREATE PROCEDURE <数据库名称>.explain_statement(IN query TEXT)
    SQL SECURITY DEFINER
BEGIN
    SET @explain := CONCAT('EXPLAIN FORMAT=json ', query);
    PREPARE stmt FROM @explain;
    EXECUTE stmt;
    DEALLOCATE PREPARE stmt;
END $$
DELIMITER ;
GRANT EXECUTE ON PROCEDURE <数据库名称>.explain_statement TO datakit@'%';
```

- `consumers` 配置

方法一（推荐）：通过 `DataKit` 动态配置 `performance_schema.events_*`，需要创建以下存储过程：

```sql
DELIMITER $$
CREATE PROCEDURE datakit.enable_events_statements_consumers()
    SQL SECURITY DEFINER
BEGIN
    UPDATE performance_schema.setup_consumers SET enabled='YES' WHERE name LIKE 'events_statements_%';
    UPDATE performance_schema.setup_consumers SET enabled='YES' WHERE name = 'events_waits_current';
END $$
DELIMITER ;

GRANT EXECUTE ON PROCEDURE datakit.enable_events_statements_consumers TO datakit@'%';
```

方法二：手动配置 `consumers`

```sql
UPDATE performance_schema.setup_consumers SET enabled='YES' WHERE name LIKE 'events_statements_%';
UPDATE performance_schema.setup_consumers SET enabled='YES' WHERE name = 'events_waits_current';
```

### 主从复制指标采集 {#replication_metrics}

采集主从复制 `mysql_replication` 指标的前提是开启主从复制，`mysql_replication` 指标都是由从数据库采集的，确认主从复制环境是否正常可以在从数据库输入：

```sql
SHOW SLAVE STATUS;
```

可以看到 `Replica_IO_Running`、`Replica_SQL_Running` 的值均为 Yes，说明主从复制环境状态正常。

若要采集组复制指标如 `count_transactions_in_queue`，需要将组复制插件添加到服务器在启动时加载的插件列表（group_replication 从 MySQL 版本 5.7.17 开始支持）。在从数据库的配置文件 `/etc/my.cnf` 中，添加一行

```toml
plugin_load_add ='group_replication.so'
```

可以通过 `show plugins;` 确认组复制插件已安装。

如需开启，需要执行以下步骤。

- 修改配置文件，开启监控采集

```toml
[[inputs.mysql]]

  ## Set replication to true to collect replication metrics
  replication = true
  ## Set group_replication to true to collect group replication metrics
  group_replication = true  
  ...

```

## 指标 {#metric}

以下所有数据采集，默认会追加全局选举 tag，也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

```toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

<!-- markdownlint-disable MD024 -->
{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## 对象 {#object}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "object"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

### `message` 指标字段结构 {#message-struct}

`message` 字段基本结构如下：

```json
{
  "setting": {
    "auto_generate_certs": "ON",
    ...
  },

  "databases": [ # databases information
    {
      "name": "db1",
      "default_character_set_name": "utf8mb4",
      "default_collation_name": "utf8mb4_general_ci",
      "tables": [ # tables information
        {
          "name": "table1",
          "columns": [], # columns information
          "indexes": [], # indexes information
          "foreign_keys": [], # foreign keys information
          "partitions": [] # partitions information
        }
        ...
      ]
    }
    ...
  ]
}
```

#### `setting` {#host-meta}

  `setting` 字段中的数据来源于 `performance_schema.global_variables` 表，该表包含了 MySQL 服务器的全局变量信息，详细字段可以参考 [MySQL 文档](https://dev.mysql.com/doc/refman/5.7/en/server-system-variables.html){:target="_blank"}。

#### `databases` {#databases}

`databases` 字段保存了 MySQL 服务器上所有数据库的信息，每个数据库的信息如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | 数据库名称                                         | string |
| `default_character_set_name`        | 数据库默认字符集（如 utf8mb4）  | string |
| `default_collation_name`        | 数据库默认排序规则（如 utf8mb4_general_ci）      | string |
| `tables`        | 包含表信息的列表      | list |

##### `tables` {#databases-tables}

`tables` 字段包含了数据库中所有表的信息，每个表的信息如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | 表名称                                         | string |
| `columns`        | 包含列信息的列表                             | list |
| `indexes`        | 包含索引信息的列表                             | list |
| `foreign_keys`        | 包含外键信息的列表                             | list |
| `partitions`        | 包含分区信息的列表                             | list |

`tables` 中类型为 list 的字段的详细结构如下：

- `tables.columns` 字段

`columns` 字段包含了表中所有列的信息，每个列的信息如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | 列名称                                         | string |
| `data_type`        | 数据类型（如 int）                                         | string |
| `default`        |  默认值（NULL 会转换为空字符串）                                        | string |
| `nullable`        |  是否允许为空（True 表示允许，对应 SQL 中的 NULL）            | bool |
| `ordinal_position`        |  列在表中的顺序位置（从 1 开始）            | string |

- `tables.indexes`

`tables.indexes` 字段包含了表中所有索引的信息，每个索引的信息如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | 索引名称                                         | string |
| `cardinality` | 索引中唯一值的估计数量 | string |
| `index_type` | 索引类型 | string |
| `columns` | 索引包含的列 | list |
| `non_unique` | 是否为非唯一索引（True 表示允许重复值） | bool |
| `expression` | 索引表达式（仅当索引基于表达式创建时存在） | string |

索引列信息字段 `indexes.columns` 包含了索引中包含的列的信息，每个列的信息如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | 列名称                                         | string |
| `sub_part` | 部分索引的字符数（如 `varchar` 列只索引前 10 个字符时为 10） | int |
| `collation` | 列的排序规则                                             | string |
| `packed` | 索引存储格式                                               | string |
| `nullable` | 列是否允许为 NULL                                         | string |

- `tables.foreign_keys`

`foreign_keys` 字段包含了表中所有外键的信息，每个外键的信息如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `constraint_schema` |外键所属的数据库（通常与表所在数据库一致） | string |
| `name` |外键约束名称 | string |
| `column_names` |外键列名称（多个列用逗号分隔，如 user_id, order_id） | string |
| `referenced_table_schema` |引用表所在的数据库 | string |
| `referenced_table_name` |引用表名称 | string |
| `referenced_column_names` |引用列名称（多个列用逗号分隔） | string |
| `update_action` |级联更新规则（如 CASCADE, RESTRICT） | string |
| `delete_action` |级联删除规则（如 CASCADE, SET NULL） | string |

- `tables.partitions`

`partitions` 字段包含了表中所有分区的信息，每个分区的信息如下：

| 字段名             | 描述                                           |  类型  |
| ------------------:| ---------------------------------------------- | :----: |
| `name` |分区名称 | string |
| `subpartitions` |包含子分区信息的字典列表（仅当存在子分区时有效） | list |
| `partition_ordinal_position` |分区在表中的顺序位置 | int |
| `partition_method` |分区方法（如 RANGE, LIST） | string |
| `partition_expression` |分区表达式（如 COLUMN(id)） | string |
| `partition_description` |分区描述（如 VALUES LESS THAN (100)） | string |
| `table_rows` |分区中的行数（包含所有子分区的总行数） | int |
| `data_length` |分区数据大小（字节，包含所有子分区的总大小） | int |

子分区信息字段 `partitions.subpartitions` 包含了子分区的信息，每个子分区的信息如下：

| 字段名 | 描述 | 数据类型 |
| --- | --- | --- |
| `name` |分区名称 | string |
| `subpartition_ordinal_position` |子分区在分区中的顺序位置 | string |
| `subpartition_method` |子分区方法（如 HASH, KEY） | string |
| `subpartition_expression` |子分区表达式 | string |
| `table_rows` |子分区中的行数 | int |
| `data_length` |子分区数据大小（字节） | int |

## 日志 {#logging}

[:octicons-tag-24: Version-1.4.6](../datakit/changelog.md#cl-1.4.6)

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

<!-- markdownlint-enable -->

### MySQL 运行日志 {#mysql-logging}

如需采集 MySQL 的日志，将配置中 log 相关的配置打开，如需要开启 MySQL 慢查询日志，需要开启慢查询日志，在 MySQL 中执行以下语句

```sql
SET GLOBAL slow_query_log = 'ON';

-- 未使用索引的查询也认为是一个可能的慢查询
set global log_queries_not_using_indexes = 'ON';
```

```toml
[inputs.mysql.log]
    # 填入绝对路径
    files = ["/var/log/mysql/*.log"]
```

> 注意：在使用日志采集时，需要将 DataKit 安装在 MySQL 服务同一台主机中，或使用其它方式将日志挂载到 DataKit 所在机器

MySQL 日志分为普通日志和慢日志两种。

### MySQL 普通日志 {#mysql-app-logging}

日志原文：

``` log
2017-12-29T12:33:33.095243Z         2 Query     SELECT TABLE_SCHEMA, TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE CREATE_OPTIONS LIKE '%partitioned%';
```

切割后的字段列表如下：

| 字段名   | 字段值                                                   | 说明                         |
| -------- | -------------------------------------------------------- | ---------------------------- |
| `status` | `Warning`                                                | 日志级别                     |
| `msg`    | `System table 'plugin' is expected to be transactional.` | 日志内容                     |
| `time`   | `1514520249954078000`                                    | 纳秒时间戳（作为行协议时间） |

### MySQL 慢查询日志 {#mysql-slow-logging}

日志原文：

``` log
# Time: 2019-11-27T10:43:13.460744Z
# User@Host: root[root] @ localhost [1.2.3.4]  Id:    35
# Query_time: 0.214922  Lock_time: 0.000184 Rows_sent: 248832  Rows_examined: 72
# Thread_id: 55   Killed: 0  Errno: 0
# Bytes_sent: 123456   Bytes_received: 0
SET timestamp=1574851393;
SELECT * FROM fruit f1, fruit f2, fruit f3, fruit f4, fruit f5
```

切割后的字段列表如下：

| 字段名              | 字段值                                                                                      | 说明                           |
| ---                 | ---                                                                                         | ---                            |
| `bytes_sent`        | `123456`                                                                                    | 发送字节数                     |
| `db_host`           | `localhost`                                                                                 | hostname                       |
| `db_ip`             | `1.2.3.4`                                                                                   | IP                             |
| `db_slow_statement` | `SET timestamp=1574851393;\nSELECT * FROM fruit f1, fruit f2, fruit f3, fruit f4, fruit f5` | 慢查询 SQL                     |
| `db_user`           | `root[root]`                                                                                | 用户                           |
| `lock_time`         | `0.000184`                                                                                  | 锁时间                         |
| `query_id`          | `35`                                                                                        | 查询 ID                        |
| `query_time`        | `0.2l4922`                                                                                  | SQL 执行所消耗的时间           |
| `rows_examined`     | `72`                                                                                        | 为了返回查询的数据所读取的行数 |
| `rows_sent`         | `248832`                                                                                    | 查询返回的行数                 |
| `thread_id`         | `55`                                                                                        | 线程 ID                        |
| `time`              | `1514520249954078000`                                                                       | 纳秒时间戳（作为行协议时间）   |

## FAQ {#faq}

<!-- markdownlint-disable MD013 -->
### 阿里云 RDS 采集时，指标 `mysql_user_status` 没有上报数据？ {#faq-user-no-data}

该指标需要开启 `performance_schema`，可以通过以下 SQL 查询：

```sql
show variables like "performance_schema";

+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| performance_schema | ON    |
+--------------------+-------+

```

如果值是 `OFF`，请参考阿里云相关 [文档](https://help.aliyun.com/document_detail/41726.html?spm=a2c4g.276975.0.i9){:target="_blank"} 进行开启。

<!-- markdownlint-enable -->
