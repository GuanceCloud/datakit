{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`


# {{.InputName}}

SQL Server 采集器采集 SQL Server `waitstats` `database_io` 等相关指标

## 前置条件

- SQL Server 版本 >= 2019

- 创建用户：

Linux、Windows:

```
USE master;
GO
CREATE LOGIN [DataFlux] WITH PASSWORD = N'yourpassword';
GO
GRANT VIEW SERVER STATE TO [DataFlux];
GO
GRANT VIEW ANY DEFINITION TO [DataFlux];
GO
```

aliyun RDS SQL Server:

```
USE master;
GO
CREATE LOGIN [DataFlux] WITH PASSWORD = N'yourpassword';
GO

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


## 日志采集

如需采集 SQL Server 的日志，可在 {{.InputName}}.conf 中 将 `files` 打开，并写入 SQL Server 日志文件的绝对路径。比如：

```toml
    [[inputs.sqlserver]]
      ...
      [inputs.sqlserver.log]
        files = ["/var/opt/mssql/log/error.log"]
```

  
开启日志采集以后，默认会产生日志来源（`source`）为 `sqlserver` 的日志。

>注意：必须将 DataKit 安装在 SQL Server 所在主机才能采集 SQL Server 日志

## 日志 pipeline 功能切割字段说明

- SQL Server 通用日志切割 

通用日志文本示例：
```
2021-05-28 10:46:07.78 spid10s     0 transactions rolled back in database 'msdb' (4:0). This is an informational message only. No user action is required
```

切割后的字段列表如下：

| 字段名  |  字段值  | 说明 |
| ---    | ---     | --- |
|  msg   | spid...   | 日志内容 |
|  time   | 1622169967780000000     | 纳秒时间戳（作为行协议时间）|
|  origin   | spid10s     | 源 |

