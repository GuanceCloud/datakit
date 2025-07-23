---
title     : 'SQLServer'
summary   : 'Collect SQLServer Metrics'
tags:
  - 'DATA STORES'
__int_icon      : 'icon/sqlserver'
dashboard :
  - desc  : 'SQLServer'
    path  : 'dashboard/en/sqlserver'
monitor   :
  - desc  : 'SQLServer'
    path  : 'monitor/en/sqlserver'
---


{{.AvailableArchs}}

---

SQL Server Collector collects SQL Server `waitstats`, `database_io` and other related metrics.


## Configuration {#config}

SQL Server  version >= 2008, tested version:

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
CREATE LOGIN [datakit] WITH PASSWORD = N'yourpassword';
GO
GRANT VIEW SERVER STATE TO [datakit];
GO
GRANT VIEW ANY DEFINITION TO [datakit];
GO
```

Aliyun RDS SQL Server:

```sql
USE master;
GO
CREATE LOGIN [datakit] WITH PASSWORD = N'yourpassword';
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
???+ note

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

For all of the following data collections, the global election tags will be added automatically, we can add extra tags in `[inputs.{{.InputName}}.tags]` if needed:

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

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}

{{ end }}
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
  "setting": [ # settings information
    {
      "name": "recovery interval (min)",
      "value": "123",
      "value_in_use": "0",
      "maximum": "10",
      "minimum": "0",
      "is_dynamic": true,
      "is_advanced": true
    }
    ...
  ],

  "databases": [ # databases information
    {
      "name": "db1",
      "owner_name": "dbo",
      "collation": "SQL_Latin1_General_CP1_CI_AS",
      "schemas": [
        {
          "name": "schema1",
          "owner_name": "dbo",
          "tables": [
            {
              "name": "table1",
              "columns": [], # columns information
              "indexes": [], # indexes information
              "foreign_keys": [], # foreign keys information
              "partitions": { # partitions information
                "partition_count": 1
              } 
            }
            ...
          ]
        }
      ]
    }
    ...
  ]
}
```

#### `setting` {#setting}

  The data in the `setting` field is derived from the `sys.configurations` system view, which contains information about the global variables of the SQL Server instance. For detailed fields, you can refer to the [SQL Server documentation](https://docs.microsoft.com/en-us/sql/relational-databases/system-catalog-views/sys-configurations-transact-sql?view=sql-server-ver16){:target="_blank"}.

#### `databases` {#databases}

The `databases` field stores information about all databases on the SQL Server instance. The information for each database is as follows:

| Field Name       | Description                                 |  Type   |
| :--------------- | :------------------------------------------ | :-----: |
| `name`           | The name of the database                    | string  |
| `owner_name`     | The name of the database owner (e.g., dbo)  | string  |
| `collation`      | The default collation of the database (e.g., SQL_Latin1_General_CP1_CI_AS) | string  |
| `schemas`        | A list containing table information         |  list   |

#### `schemas` {#schemas}

The `schemas` field stores information about all schemas in the database. The information for each schema is as follows:

| Field Name       | Description                                         |  Type  |
| :--------------- | :-------------------------------------------------- | :----: |
| `name`           | The name of the schema                              | string |
| `owner_name`     | The name of the schema owner (e.g., dbo)            | string |
| `tables`         | A list containing table information                 |  list  |

The `tables` field stores information about all tables included in the schema. The information for each table is as follows:

| Field Name       | Description                                         |  Type  |
| :--------------- | :-------------------------------------------------- | :----: |
| `name`           | The name of the table                               | string |
| `columns`        | A list containing column information                |  list  |
| `indexes`        | A list containing index information                 |  list  |
| `foreign_keys`   | A list containing foreign key information           |  list  |
| `partitions`     | A dictionary containing partition information       |  dict  |

The `tables.columns` field stores information about all columns included in the table. The information for each column is as follows:

| Field Name       | Description                                         |  Type  |
| :--------------- | :-------------------------------------------------- | :----: |
| `name`           | The name of the column                              | string |
| `data_type`      | The data type of the column | string |
| `nullable`       | Whether the column allows null values | string |
| `default`        | The default value of the column                     | string |

The `tables.indexes` field stores information about all indexes included in the table. The information for each index is as follows:

| Field Name             | Description                                         |  Type  |
| :--------------------- | :-------------------------------------------------- | :----: |
| `name`                 | The name of the index                               | string |
| `type`                 | The type of the index                               | string |
| `is_unique`            | Whether the index is unique                         | string |
| `is_primary_key`       | Whether the index is a primary key                  | string |
| `column_names`         | The names of the columns included in the index      | string |
| `is_disabled`          | Whether the index is disabled                       | string |
| `is_unique_constraint` | Whether the index is a unique constraint            | string |

The `tables.foreign_keys` field stores information about all foreign keys included in the table. The information for each foreign key is as follows:

| Field Name           | Description                                         |  Type  |
| :------------------- | :-------------------------------------------------- | :----: |
| `foreign_key_name`   | The name of the foreign key                         | string |
| `referencing_table`  | The name of the referencing table                    | string |
| `referenced_table`   | The name of the referenced table                    | string |
| `referencing_columns`| The names of the referencing columns                | string |
| `referenced_columns` | The names of the referenced columns                 | string |
| `update_action`      | The cascading action for updates                    | string |
| `delete_action`      | The cascading action for deletions                  | string |

The `tables.partitions` field stores the number of all partitions included in the table. The specific field description is as follows:

| Field Name         | Description                                         |  Type  |
| :----------------- | :-------------------------------------------------- | :----: |
| `partition_count`  | The number of partitions                            | number |

## Logging {#logging}

Following measurements are collected as logs with the level of `info`.

{{ range $i, $m := .Measurements }}
{{if eq $m.Type "logging"}}

### `{{$m.Name}}`

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

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
