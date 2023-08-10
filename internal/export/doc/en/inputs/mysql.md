
# MySQL

---

{{.AvailableArchs}}

---

MySQL metrics collection, which collects the following data:

- MySQL global status basic data collection
- Scheam related data
- InnoDB related metrics
- Support custom query data collection

## Preconditions {#requirements}

- MySQL version 5.7+
- Create a monitoring account (in general, you need to log in with MySQL `root` account to create MySQL users)

```sql
CREATE USER 'datakit'@'localhost' IDENTIFIED BY '<UNIQUEPASSWORD>';

-- MySQL 8.0+ create the datakit user with the native password hashing method
CREATE USER 'datakit'@'localhost' IDENTIFIED WITH mysql_native_password by '<UNIQUEPASSWORD>';
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

All the above creation and authorization operations limit that the user `datakit` can only access MySQL on MySQL host (`localhost`). If MySQL is collected remotely, it is recommended to replace `localhost` with `%` (indicating that DataKit can access MySQL on any machine), or use a specific DataKit installation machine address.

> Note that if you find the collector has the following error when using `localhost` , you need to replace the above `localhost` with `::1`

```
Error 1045: Access denied for user 'datakit'@'::1' (using password: YES)
```

## Configuration {#config}

=== "Host Installation"

    Go to the `conf.d/db` directory under the DataKit installation directory, copy `mysql.conf.sample` and name it `mysql.conf`. Examples are as follows:
    
    ```toml
        
    [[inputs.mysql]]
      host = "localhost"
      user = "datakit"
      pass = "<PASS>"
      port = 3306
      # sock = "<SOCK>"
      # charset = "utf8"
    
      ## @param connect_timeout - number - optional - default: 10s
      # connect_timeout = "10s"
    
      ## Deprecated
      # service = "<SERVICE>"
    
      interval = "10s"
    
      ## @param inno_db
      innodb = true
    
      ## table_schema
      tables = []
    
      ## user
      users = []
    
      ## Start database performance index collection
      # dbm = false
    
      # [inputs.mysql.log]
      # #required, glob logfiles
      # files = ["/var/log/mysql/*.log"]
    
      ## glob filteer
      #ignore = [""]
    
      ## optional encodings:
      ##    "utf-8", "utf-16le", "utf-16le", "gbk", "gb18030" or ""
      #character_encoding = ""
    
      ## The pattern should be a regexp. Note the use of '''this regexp'''
      ## regexp link: https://golang.org/pkg/regexp/syntax/#hdr-Syntax
      #multiline_match = '''^(# Time|\d{4}-\d{2}-\d{2}|\d{6}\s+\d{2}:\d{2}:\d{2}).*'''
    
      ## grok pipeline script path
      #pipeline = "mysql.p"
    
      ## Set true to enable election
      election = true
    
      # [[inputs.mysql.custom_queries]]
      #   sql = "SELECT foo, COUNT(*) FROM table.events GROUP BY foo"
      #   metric = "xxxx"
      #   tagKeys = ["column1", "column1"]
      #   fieldKeys = ["column3", "column1"]
      
      ## Monitoring metric configuration
      [inputs.mysql.dbm_metric]
        enabled = true
      
      ## Monitoring sampling configuration
      [inputs.mysql.dbm_sample]
        enabled = true  
    
      ## Waiting for event collection
      [inputs.mysql.dbm_activity]
        enabled = true  
    
      [inputs.mysql.tags]
        # some_tag = "some_value"
        # more_tag = "some_other_value"
    
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).


---

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.mysql.tags]`:

```toml
 [inputs.mysql.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

### Binlog Start {#binlog}

MySQL binlog is not turned on. If you want to count the binlog size, you need to turn on the binlog function corresponding to MySQL:

```sql
-- ON:开启, OFF:关闭
SHOW VARIABLES LIKE 'log_bin';
```

binlog starts, see [this](https://stackoverflow.com/questions/40682381/how-do-i-enable-mysql-binary-logging){:target="_blank"} or [this answer](https://serverfault.com/questions/706699/enable-binlog-in-mysql-on-ubuntu){:target="_blank"}.

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
-- 	MySQL 5.6 & 5.7
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

Create the stored procedure `explain_statement` to get the sql execution plan

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

### Measurements {#measurement}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

#### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}} {{end}}

{{ end }}




### Log {#logging}

[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

#### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- Metric list

{{$m.FieldsMarkdownTable}} {{end}}

{{ end }}


## MySQL Run Log {#mysql-logging}

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

```
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

```
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
| `db_slow_statement` | `SET timestamp=1574851393;\nSELECT * FROM fruit f1, fruit f2, fruit f3, fruit f4, fruit f5` | Slow query sql                     |
| `db_user`           | `root[root]`                                                                                | User                           |
| `lock_time`         | `0.000184`                                                                                  | Lock time                         |
| `query_id`          | `35`                                                                                        | query id                        |
| `query_time`        | `0.2l4922`                                                                                  | Time spent on SQL execution           |
| `rows_examined`     | `72`                                                                                        | Number of rows read to return queried data |
| `rows_sent`         | `248832`                                                                                    | Number of rows returned by query                 |
| `thread_id`         | `55`                                                                                        | Thread id                        |
| `time`              | `1514520249954078000`                                                                       | Nanosecond timestamp (as line protocol time)   |

## FAQ {#faq}

### :material-chat-question: Why the measurement `mysql_user_status` is not collected for Aliyun RDS? {#faq-user-no-data}

The measurment is collected from MySQL `performance_schema`. You should check if it is enabled by the SQL below：

```sql
show variables like "performance_schema";

+--------------------+-------+
| Variable_name      | Value |
+--------------------+-------+
| performance_schema | ON    |
+--------------------+-------+

```

If the value is `OFF`, please refer to the [document](https://help.aliyun.com/document_detail/41726.html?spm=a2c4g.276975.0.i9) to enable it.
