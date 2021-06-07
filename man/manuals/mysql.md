{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

MySQL 指标采集，收集以下数据：

- mysql global status 基础数据采集
- scheam 相关数据
- innodb 相关指标 
- 主从模式 
- 支持自定义查询数据采集

>  主从模式相关的 MySQL 指标采集尚未支持（Comming Soon...）

## 前置条件

- MySQL 版本 5.7+

- 创建监控账号（一般情况，需用 MySQL `root` 账号登陆才能创建 MySQL 用户）

```sql
    CREATE USER 'datakitMonitor'@'localhost' IDENTIFIED BY '<UNIQUEPASSWORD>';
    
    -- MySQL 8.0+ create the datakitMonitor user with the native password hashing method
    CREATE USER 'datakitMonitor'@'localhost' IDENTIFIED WITH mysql_native_password by '<UNIQUEPASSWORD>';
```

备注：`localhost` 是本地连接，具体参考[这里](https://dev.mysql.com/doc/refman/8.0/en/creating-accounts.html)

- 授权

```sql
    GRANT PROCESS ON *.* TO 'datakitMonitor'@'localhost';
    show databases like 'performance_schema';
    GRANT SELECT ON performance_schema.* TO 'datakitMonitor'@'localhost';
```

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

## Binlog 开启

默认情况下，MySQL binlog 是不开启的。如果要统计 binlog 大小，需要开启 MySQL 对应 binlog 功能：

```sql
-- ON:开启, OFF:关闭
SHOW VARIABLES LIKE 'log_bin';
```

binlog 开启，参见[这个问答](https://stackoverflow.com/questions/40682381/how-do-i-enable-mysql-binary-logging)，或者[这个问答](https://serverfault.com/questions/706699/enable-binlog-in-mysql-on-ubuntu)

## 日志采集

如需采集 MySQL 的日志，将配置中 log 相关的配置打开，如需要开启 MySQL 慢查询日志，需要开启慢查询日志，在 MySQL 中执行以下语句

```sql
SET GLOBAL slow_query_log = 'ON';

-- 未使用索引的查询也认为是一个可能的慢查询
set global log_queries_not_using_indexes = 'ON';
```

```python
[inputs.mysql.log]
    # 填入绝对路径
    files = ["/var/log/mysql/*.log"] 
```

> 注意：在使用日志采集时，需要将 DataKit 安装在 MySQL 服务同一台主机中，或使用其它方式将日志挂载到 DataKit 所在机器

## 日志 pipeline 功能切割字段说明

- Mysql 日志

**日志原文**

```
2017-12-29T12:33:33.095243Z         2 Query     SELECT TABLE_SCHEMA, TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE CREATE_OPTIONS LIKE '%partitioned%';
```

切割后的字段列表如下：

| 字段名  | 字段值                                                   | 说明                         |
| ---     | ---                                                      | ---                          |
| `status | `Warning`                                                | 日志级别                     |
| `msg`   | `System table 'plugin' is expected to be transactional.` | 日志内容                     |
| `time`  | `1514520249954078000`                                    | 纳秒时间戳（作为行协议时间） |

- 慢查询日志

**日志原文**

```
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
| `db_ip`             | `1.2.3.4`                                                                                   | ip                             |
| `db_slow_statement` | `SET timestamp=1574851393;\nSELECT * FROM fruit f1, fruit f2, fruit f3, fruit f4, fruit f5` | 慢查询sql                      |
| `db_user`           | `root[root]`                                                                                | 用户                           |
| `lock_time`         | `0.000184`                                                                                  | 锁时间                         |
| `query_id`          | `35`                                                                                        | 查询id                         |
| `query_time`        | `0.2l4922`                                                                                  | SQL执行所消耗的时间            |
| `rows_examined`     | `72`                                                                                        | 为了返回查询的数据所读取的行数 |
| `rows_sent`         | `248832`                                                                                    | 查询返回的行数                 |
| `thread_id`         | `55`                                                                                        | 线程id                         |
| `time`              | `1514520249954078000`                                                                       | 纳秒时间戳（作为行协议时间）   |
