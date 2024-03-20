---
title     : 'SQLServer'
summary   : 'Collect SQLServer Metrics'
__int_icon      : 'icon/sqlserver'
dashboard :
  - desc  : 'SQLServer'
    path  : 'dashboard/en/sqlserver'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# SQLServer
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

SQL Server Collector collects SQL Server `waitstats`, `database_io` and other related metrics.


## Configuration {#config}

SQL Server  version >= 2012, tested version:

- [x] 2017
- [x] 2019
- [x] 2022

### Prerequisites {#requrements}

- SQL Server version >= 2019

- Create a user:

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
### Collector Configuration {#input-config}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, restart DataKit.

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->

#### Log Collector Configuration {#logging-config}

<!-- markdownlint-disable MD046 -->
???+ attention

     DataKit must be installed on the host where SQLServer is running.
<!-- markdownlint-enable -->

To collect SQL Server logs, enable `files` in *{{.InputName}}.conf* and write to the absolute path of the SQL Server log file. For example:

```toml hl_lines="4"
[[inputs.sqlserver]]
    ...
    [inputs.sqlserver.log]
        files = ["/var/opt/mssql/log/error.log"]
```

When log collection is turned on, a log with a log (aka *source*) of`sqlserver` is collected.

## Metrics {#measurements}

For all of the following data collections, a global tag name `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.{{.InputName}}.tags]`:

``` toml
 [inputs.sqlserver.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

<!-- markdownlint-disable MD024 -->
{{ range $i, $m := .Measurements }}
{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

- tag

{{$m.TagsMarkdownTable}}

- field list

{{$m.FieldsMarkdownTable}}

{{ end }}
{{ end }}

## Logging {#logging}

Following measurements are collected as logs with the level of `info`.

{{ range $i, $m := .Measurements }}
{{if eq $m.Type "logging"}}

### `{{$m.Name}}`

- tag

{{$m.TagsMarkdownTable}}

- field list

{{$m.FieldsMarkdownTable}}

{{ end }}
{{ end }}
<!-- markdownlint-enable -->

### Pipeline for  SQLServer logging {#pipeline}

- SQL Server Common Log Pipeline

Example of common log text:

```log
2021-05-28 10:46:07.78 spid10s     0 transactions rolled back in database 'msdb' (4:0). This is an informational message only. No user action is required
```

The list of extracted fields are as follows:

| Field Name | Field Value         | Description                                                                                |
| ---------- | ------------------- | ------------------------------------------------------------------------------------------ |
| `msg`      | spid...             | log content                                                                                |
| `time`     | 1622169967780000000 | nanosecond timestamp (as row protocol time)                                                |
| `origin`   | spid10s             | source                                                                                     |
| `status`   | info                | As the log does not have an explicit field to describe the log level, the default is info. |
