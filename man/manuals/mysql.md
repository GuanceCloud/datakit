{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：{{.AvailableArchs}}

# 简介

MySQL 指标采集，收集以下数据：

- mysql global status 基础数据采集
- scheam 相关数据
- innodb 相关指标(TODO)
- 主从模式(TODO)
- 支持自定义查询数据采集

## 前置条件

- 创建监控账号

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

```python
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

如需采集 `mysql` 的日志，将配置中`log`相关的配置打开，如需要开启msql慢查询日志，需要开启慢查询日志
```
SET GLOBAL slow_query_log = 'ON';

-- 未使用索引的查询也认为是一个可能的慢查询
set global log_queries_not_using_indexes = 'ON';
```

**注意**
- 日志路径需要填入绝对路径
- 在使用日志采集时，需要将datakit安装在redis服务同一台主机中，或使用其它方式将日志挂载到外部系统中

