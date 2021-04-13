{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：{{.AvailableArchs}}

# 简介

oracle监控指标采集，参考datadog提供的指标，具有以下数据收集功能

- process相关
- tablespace相关数据
- system数据采集
- 自定义查询数据采集

## 前置条件

- 创建监控账号

```
-- Enable Oracle Script.
ALTER SESSION SET "_ORACLE_SCRIPT"=true;

-- Create the datakit user. Replace the password placeholder with a secure password.
CREATE USER datakit IDENTIFIED BY <PASSWORD>;

-- Grant access to the datakit user.
GRANT CONNECT TO datakit;
GRANT SELECT ON GV_$PROCESS TO datakit;
GRANT SELECT ON gv_$sysmetric TO datakit;
GRANT SELECT ON sys.dba_data_files TO datakit;
GRANT SELECT ON sys.dba_tablespaces TO datakit;
GRANT SELECT ON sys.dba_tablespace_usage_metrics TO datakit;
```

备注：oracle 11g, 需要以下设定

```
ALTER SESSION SET "_ORACLE_SCRIPT"=true;
```

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

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
