- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}

# 简介

mysql监控指标采集，参考datadog提供的指标，具有以下数据收集功能
- mysql global status 基础数据采集
- scheam 相关数据
- 自定义查询数据采集
- innodb相关指标(待实现)
- 主从模式(待实现)

备注：
redis开发测试版本为v5.04, v6.0+待支持

## 前置条件
- 创建监控账号
```
CREATE USER 'datakitMonitor'@'localhost' IDENTIFIED BY '<UNIQUEPASSWORD>';

#mySQL 8.0+ create the datakitMonitor user with the native password hashing method
CREATE USER 'datakitMonitor'@'localhost' IDENTIFIED WITH mysql_native_password by '<UNIQUEPASSWORD>';
```
备注：@'localhost' 是本地连接，具体参考[](https://dev.mysql.com/doc/refman/8.0/en/creating-accounts.html)

- 授权
```
GRANT PROCESS ON *.* TO 'datakitMonitor'@'localhost';
show databases like 'performance_schema';
GRANT SELECT ON performance_schema.* TO 'datakitMonitor'@'localhost';
```

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.InputName}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```
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
