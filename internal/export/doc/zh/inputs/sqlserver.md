---
title     : 'SQLServer'
summary   : '采集 SQLServer 的指标数据'
__int_icon      : 'icon/sqlserver'
dashboard :
  - desc  : 'SQLServer'
    path  : 'dashboard/zh/sqlserver'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# SQLServer
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

SQL Server 采集器采集 SQL Server `waitstats`、`database_io` 等相关指标


## 配置 {#config}

### 前置条件 {#requrements}

- SQL Server 版本 >= 2019

- 创建用户：

Linux、Windows:

```sql
USE master;
GO
CREATE LOGIN [guance] WITH PASSWORD = N'yourpassword';
GO
GRANT VIEW SERVER STATE TO [guance];
GO
GRANT VIEW ANY DEFINITION TO [guance];
GO
```

Aliyun RDS SQL Server:

```sql
USE master;
GO
CREATE LOGIN [guance] WITH PASSWORD = N'yourpassword';
GO
```

<!-- markdownlint-disable MD046 -->
???+ attention "注意事项"

    注意，执行上述操作需要相应权限的帐号，否则可能会导致用户创建失败或者授权失败。

    - 自建的 SQL Server 需要具备 WITH GRANT OPTION、CREATE ANY LOGIN、CREATE ANY USER、ALTER ANY LOGIN 权限的用户，也可以直接使用具有 sysadmin 角色的用户或者 local 用户授权。
    - RDS for SQL Server 则需要使用高权限账号进行授权。

### 采集器配置 {#input-config}

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

#### 日志采集配置 {#logging-config}

<!-- markdownlint-disable MD046 -->
???+ attention

    必须将 DataKit 安装在 SQLServer 所在主机才能采集日志。
<!-- markdownlint-enable -->

如需采集 SQL Server 的日志，可在 *{{.InputName}}.conf* 中 将 `files` 打开，并写入 SQL Server 日志文件的绝对路径。比如：

```toml hl_lines="4"
[[inputs.sqlserver]]
    ...
    [inputs.sqlserver.log]
        files = ["/var/opt/mssql/log/error.log"]
```

开启日志采集以后，默认会产生日志来源（*source*）为 `sqlserver` 的日志。

## 指标 {#metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

<!-- markdownlint-disable MD024 -->
{{ range $i, $m := .Measurements }}
{{if eq $m.Type "metric"}}
### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}

{{ end }}
{{ end }}

## 日志 {#logging}

以下指标集均以日志形式收集，所有日志等级均为 `info`。

{{ range $i, $m := .Measurements }}
{{if eq $m.Type "logging"}}
### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}

{{ end }}
{{ end }}
<!-- markdownlint-enable -->

### 日志 Pipeline 功能切割字段说明 {#pipeline}

SQL Server 通用日志文本示例：

```log
2021-05-28 10:46:07.78 spid10s     0 transactions rolled back in database 'msdb' (4:0). This is an informational message only. No user action is required
```

切割后的字段列表如下：

| 字段名   | 字段值                | 说明                                          |
| ---      | ---                   | ---                                           |
| `msg`    | `spid...`             | 日志内容                                      |
| `time`   | `1622169967780000000` | 纳秒时间戳（作为行协议时间）                  |
| `origin` | `spid10s`             | 源                                            |
| `status` | `info`                | 由于日志没有明确字段说明日志等级，默认为 info |
