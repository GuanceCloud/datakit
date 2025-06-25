---
title     : 'MySQL'
summary   : 'Collect MySQL metrics and logs'
tags:
  - 'DATA STORES'
__int_icon      : 'icon/mysql'
dashboard :
  - desc  : 'MySQL'
    path  : 'dashboard/en/mysql'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---


{{.AvailableArchs}}

---

MySQL metrics collection, which collects the following data:

- MySQL global status basic data collection
- Schema related data
- InnoDB related metrics
- Support custom query data collection

## Configuration {#config}

### Preconditions {#requirements}

- MySQL version 5.7+
- Create a monitoring account (in general, you need to log in with MySQL `root` account to create MySQL users)

```sql
CREATE USER 'datakit'@'localhost' IDENTIFIED BY '<UNIQUEPASSWORD>';

-- MySQL 8.0+ create the datakit user with the caching_sha2_password method
CREATE USER 'datakit'@'localhost' IDENTIFIED WITH caching_sha2_password by '<UNIQUEPASSWORD>';
```

- Authorization

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

    - All the above creation and authorization operations limit that the user `datakit` can only access MySQL on MySQL host (`localhost`). If MySQL is collected remotely, it is recommended to replace `localhost` with `%` (indicating that DataKit can access MySQL on any machine), or use a specific DataKit installation machine address.
    - Note that if you find the collector has the following error when using `localhost`, you need to replace the above `localhost` with `::1`:

        ```text
        Error 1045: Access denied for user 'datakit'@'localhost' (using password: YES)`
        ```
<!-- markdownlint-enable -->

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/db` directory under the DataKit installation directory, copy `mysql.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->

---

### Binlog Start {#binlog}

MySQL Binlog is not turned on. If you want to count the Binlog size, you need to turn on the Binlog function corresponding to MySQL:

```sql
-- ON: turn on, OFF: turn off
SHOW VARIABLES LIKE 'log_bin';
```

Binlog starts, see [this](https://stackoverflow.com/questions/40682381/how-do-i-enable-mysql-binary-logging){:target="_blank"} or [this answer](https://serverfault.com/questions/706699/enable-binlog-in-mysql-on-ubuntu){:target="_blank"}.

### Database Performance Metrics Collection {#performance-schema}

The database performance metrics come from MySQL's built-in database `performance_schema`, which provides a way to get the internal performance of the server at runtime. Through this database, DataKit can collect statistics of various metrics of historical query statements, execution plans of query statements and other related performance metrics. The collected performance metric data is saved as a log, and the sources are `mysql_dbm_metric`, `mysql_dbm_sample` and `mysql_dbm_activity`.

To turn it on, you need to perform the following steps.

- Modify the configuration file and start monitoring and collection

```toml
[[inputs.mysql]]

# Turn on database performance metric collection
dbm = true

...

# Monitor metric configuration
[inputs.mysql.dbm_metric]
  enabled = true

# Monitor sampling configuration
[inputs.mysql.dbm_sample]
  enabled = true

# Waiting for event collection
[inputs.mysql.dbm_activity]
  enabled = true   
...

```

- MySQL Configuration

Modify the configuration file (such as `mysql.conf`), open the `MySQL Performance Schema`, and configure the parameters:

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

- Account configuration

Account authorization

```sql
-- MySQL 5.6 & 5.7
GRANT REPLICATION CLIENT ON *.* TO datakit@'%' WITH MAX_USER_CONNECTIONS 5;
GRANT PROCESS ON *.* TO datakit@'%';

-- MySQL >= 8.0
ALTER USER datakit@'%' WITH MAX_USER_CONNECTIONS 5;
GRANT REPLICATION CLIENT ON *.* TO datakit@'%';
GRANT PROCESS ON *.* TO datakit@'%';
```

Create a database

```sql
CREATE SCHEMA IF NOT EXISTS datakit;
GRANT EXECUTE ON datakit.* to datakit@'%';
GRANT CREATE TEMPORARY TABLES ON datakit.* TO datakit@'%';
```

Create the stored procedure `explain_statement` to get the SQL execution plan

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

Create a separate stored procedure for the database that needs to collect execution plans (optional)

```sql
DELIMITER $$
CREATE PROCEDURE <db_name>.explain_statement(IN query TEXT)
    SQL SECURITY DEFINER
BEGIN
    SET @explain := CONCAT('EXPLAIN FORMAT=json ', query);
    PREPARE stmt FROM @explain;
    EXECUTE stmt;
    DEALLOCATE PREPARE stmt;
END $$
DELIMITER ;
GRANT EXECUTE ON PROCEDURE <db_name>.explain_statement TO datakit@'%';
```

- `consumers` configuration

Method one (recommended): Dynamic configuration of `performance_schema.events_*` with `DataKit` requires the creation of the following stored procedure:

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

Method 2: Manually configure `consumers`

```sql
UPDATE performance_schema.setup_consumers SET enabled='YES' WHERE name LIKE 'events_statements_%';
UPDATE performance_schema.setup_consumers SET enabled='YES' WHERE name = 'events_waits_current';
```

### Replication Metrics Collection {#replication_metrics}

To collect replication metrics `mysql_replication`, you need to start MySQL replication. `mysql_replication` metrics are collected from the replication database, so you can confirm that the MySQL replication environment is working properly by entering them in the slave database:

```sql
SHOW SLAVE STATUS;
```

If the `Slave_IO_Running` and `Slave_SQL_Running` fields are `Yes`, the replication environment is working properly.

To capture group replication metrics such as `count_transactions_in_queue`, you need to add the group_replication plugin to the list of plugins loaded by the server at startup (group_replication has been supported since MySQL version 5.7.17). In the configuration file `/etc/my.cnf` for the replication database, add the line:

```toml
plugin_load_add ='group_replication.so'
```

You can confirm that the group replication plugin is installed by `showing plugins;`.

To turn it on, you need to perform the following steps.

- Modify the configuration file and start monitoring and collection

```toml
  [[inputs.mysql]]

  ## Set replication to true to collect replication metrics
  replication = true
  ## Set group_replication to true to collect group replication metrics
  group_replication = true
  ...
  
```

## Metric {#metric}

For all of the following data collections, the global election tags will added automatically, we can add extra tags in `[inputs.{{.InputName}}.tags]` if needed:

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

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}{{end}}

{{ end }}

## Object {#object}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "object"}}

### `{{$m.Name}}`

{{$m.Desc}}

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

### `message` Metric Field Structure {#message-struct}

The basic structure of the `message` field is as follows:

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

#### `setting` {#setting}

  The `setting` field contains data sourced from the `performance_schema.global_variables` table, which holds global variable information for the MySQL server. Detailed fields can be referenced in the [MySQL Documentation](https://dev.mysql.com/doc/refman/5.7/en/server-system-variables.html){:target="_blank"}

#### `databases` {#databases}

The `databases` field stores information about all databases on the MySQL server. Each database entry includes the following:

| Field Name              | Description                 | Type   |
| ------------------:| ---------------------------------------------- | :----: |
| `name`                       | Database name                | string |
| `default_character_set_name` | Default character set for the database (e.g., utf8mb4)  | string |
| `default_collation_name`     | Default collation for the database (e.g., utf8mb4_general_ci)      | string |
| `tables`                     | List containing table information      | list |

##### `tables` {#databases-tables}

The `tables` field contains information about all tables in the database. Each table entry includes:

| Field Name              | Description                 | Type   |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | Table name                                         | string |
| `columns`        | List containing column information              | list |
| `indexes`        | List containing index information               | list |
| `foreign_keys`   | List containing foreign key information         | list |
| `partitions`     | List containing partition information           | list |

Details of the list fields:

- `tables.columns` field

The `tables.columns` field contains information about all columns in the table. Each column entry includes:

| Field Name              | Description                 | Type   |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | Column name                                         | string |
| `data_type`        |  Data type (e.g., int)                                        | string |
| `default`        |  Default value (NULL is converted to empty string)           | string |
| `nullable`        |  Whether NULL values are allowed (True indicates NULL is allowed, corresponding to NULL in SQL) | bool |
| `ordinal_position`        |   The sequential position of the column in the table (starting from 1)           | string |

- `tables.indexes`

The `tables.indexes` field contains information about all indexes in the table. Each index entry includes:

| Field Name              | Description                 | Type   |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | Index name                                         | string |
| `cardinality` | Estimated number of unique values in the index | string |
| `index_type` | Index type | string |
| `columns` | Columns included in the index | list |
| `non_unique` | Whether the index allows duplicate values (True indicates non-unique) | bool |
| `expression` | Index expression (only present if the index is based on an expression) | string |

The index column information field `indexes.columns` contains details about the columns included in the index. Each column entry includes:

| Field Name              | Description                 | Type   |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | Column name                                         | string |
| `sub_part` | Number of characters indexed for partial indexes (e.g., 10 for indexing the first 10 characters of a `varchar` column) | int |
| `collation` | Column collation                                             | string |
| `packed` | Index storage format                                               | string |
| `nullable` | Whether the column allows NULL values                                         | string |

- `tables.foreign_keys`

The `tables.foreign_keys` field contains information about all foreign keys in the table. Each foreign key entry includes:

| Field Name              | Description                 | Type   |
| ------------------:| ---------------------------------------------- | :----: |
| `constraint_schema`| Database to which the foreign key belongs (typically the same as the table's database) |string |
| `name`| Foreign key constraint name |string |
| `column_names`| Names of foreign key columns (comma-separated for multiple columns, e.g., user_id, order_id) |string |
| `referenced_table_schema`| Database of the referenced table |string |
| `referenced_table_name`| Name of the referenced table |string |
| `referenced_column_names`| Names of referenced columns (comma-separated) |string |
| `update_action`| Cascade update rule (e.g., CASCADE, RESTRICT) |string |
| `delete_action`| Cascade delete rule (e.g., CASCADE, SET NULL) |string |

- `tables.partitions`

The `tables.partitions` field contains information about all partitions in the table. Each partition entry includes:

| Field Name              | Description                 | Type   |
| ------------------:| ---------------------------------------------- | :----: |
| `name` |   Partition name | string |
| `subpartitions` |   List of dictionaries containing `subpartition` information (only valid if `subpartitions` exist) | list |
| `partition_ordinal_position` |   Sequential position of the partition in the table | int |
| `partition_method` |   Partitioning method (e.g., RANGE, LIST) | string |
| `partition_expression` |   Partition expression (e.g., COLUMN(id)) | string |
| `partition_description` |   Partition description (e.g., VALUES LESS THAN (100)) | string |
| `table_rows` |   Number of rows in the partition (total rows including all `subpartitions`) | int |
| `data_length` |   Size of partition data (in bytes, total size including all `subpartitions`) | int |

The `subpartition` information field `partitions.subpartitions` contains details about `subpartitions`. Each `subpartition` entry includes:

| Field Name              | Description                 | Type   |
| --- | --- | --- |
| `subpartition_ordinal_position` | Sequential position of the `subpartition` within the partition| string |
| `subpartition_method` | `Subpartitioning` method (e.g., HASH, KEY)| string |
| `subpartition_expression` | `Subpartition` expression| string |
| `table_rows` | Number of rows in the `subpartition`| int |
| `data_length` | Size of `subpartition` data (in bytes)| int |

## Log {#logging}

[:octicons-tag-24: Version-1.4.6](../datakit/changelog.md#cl-1.4.6)

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

### `{{$m.Name}}`

{{$m.Desc}}

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}{{end}}

{{ end }}
<!-- markdownlint-enable -->

### MySQL Run Log {#mysql-logging}

If you need to collect MySQL log, open the log-related configuration in the configuration. If you need to open MySQL slow query log, you need to open the slow query log. Execute the following statements in MySQL.

```sql
SET GLOBAL slow_query_log = 'ON';

-- Queries that do not use indexes are also considered a possible slow query
set global log_queries_not_using_indexes = 'ON';
```

```toml
[inputs.mysql.log]
    # Fill in the absolute path
    files = ["/var/log/mysql/*.log"]
```

> Note: When using log collection, you need to install the DataKit on the same host as the MySQL service, or use other methods to mount the log on the machine where the DataKit is located.

MySQL logs are divided into normal logs and slow logs.

### MySQL Normal Logs {#mysql-app-logging}

Original log:

``` log
2017-12-29T12:33:33.095243Z         2 Query     SELECT TABLE_SCHEMA, TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE CREATE_OPTIONS LIKE '%partitioned%';
```

The list of cut fields is as follows:

| Field Name   | Field Value                                                   | Description                         |
| -------- | -------------------------------------------------------- | ---------------------------- |
| `status` | `Warning`                                                | log level                     |
| `msg`    | `System table 'plugin' is expected to be transactional.` | log content                     |
| `time`   | `1514520249954078000`                                    | Nanosecond timestamp (as row protocol time) |

### MySQL Slow Query Log {#mysql-slow-logging}

Original log:

``` log
# Time: 2019-11-27T10:43:13.460744Z
# User@Host: root[root] @ localhost [1.2.3.4]  Id:    35
# Query_time: 0.214922  Lock_time: 0.000184 Rows_sent: 248832  Rows_examined: 72
# Thread_id: 55   Killed: 0  Errno: 0
# Bytes_sent: 123456   Bytes_received: 0
SET timestamp=1574851393;
SELECT * FROM fruit f1, fruit f2, fruit f3, fruit f4, fruit f5
```

The list of cut fields is as follows:

| Field Name              | Field Value                                                                                      | Description                           |
| ---                 | ---                                                                                         | ---                            |
| `bytes_sent`        | `123456`                                                                                    | Number of bytes sent                     |
| `db_host`           | `localhost`                                                                                 | hostname                       |
| `db_ip`             | `1.2.3.4`                                                                                   | ip                             |
| `db_slow_statement` | `SET timestamp=1574851393;\nSELECT * FROM fruit f1, fruit f2, fruit f3, fruit f4, fruit f5` | Slow query SQL                     |
| `db_user`           | `root[root]`                                                                                | User                           |
| `lock_time`         | `0.000184`                                                                                  | Lock time                         |
| `query_id`          | `35`                                                                                        | query id                        |
| `query_time`        | `0.2l4922`                                                                                  | Time spent on SQL execution           |
| `rows_examined`     | `72`                                                                                        | Number of rows read to return queried data |
| `rows_sent`         | `248832`                                                                                    | Number of rows returned by query                 |
| `thread_id`         | `55`                                                                                        | Thread id                        |
| `time`              | `1514520249954078000`                                                                       | Nanosecond timestamp (as line protocol time)   |

## FAQ {#faq}

<!-- markdownlint-disable MD013 -->
### Why the measurement `mysql_user_status` is not collected for Aliyun RDS? {#faq-user-no-data}

The measurement is collected from MySQL `performance_schema`. You should check if it is enabled by the SQL below：

```sql
show variables like "performance_schema";

+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| performance_schema | ON    |
+--------------------+-------+

```

If the value is `OFF`, please refer to the [document](https://help.aliyun.com/document_detail/41726.html?spm=a2c4g.276975.0.i9){:target="_blank"} to enable it.

<!-- markdownlint-enable -->
