### 简介
`Postgresql`采集，参考datadog指标，是基于`sql`查询的方式来获取相关指标。

### 配置
```
[[inputs.postgresql]]
## 服务器地址
# url格式 
#	postgres://[pqgotest[:password]]@localhost[/dbname]?sslmode=[disable|verify-ca|verify-full]
# 简单字符串格式
# 	host=localhost user=pqgotest password=... sslmode=... dbname=app_production

address = "postgres://postgres@localhost/test?sslmode=disable"

## 配置采集的数据库，默认会采集所有的数据库
# ignored_databases = ["db1"]
# databases = ["db1"]

## 设置服务器Tag，默认是基于服务器地址生成
# outputaddress = "db01"

## 采集间隔
# 单位 "ns", "us" (or "µs"), "ms", "s", "m", "h"
interval = "10s"

## 自定义Tag
[inputs.postgresql.tags]
# a = "b"

```

### 指标集
指标集名称为`postgresql`
指标来源的表有`pg_stat_database`, `pg_stat_bgwriter`