---
title     : 'PostgreSQL'
summary   : 'Collect PostgreSQL metrics'
tags:
  - 'DATABASE'
__int_icon      : 'icon/postgresql'
dashboard :
  - desc  : 'PostgrepSQL'
    path  : 'dashboard/en/postgresql'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---


{{.AvailableArchs}}

---

PostgreSQL collector can collect the running status index from PostgreSQL instance, and collect the index to <<<custom_key.brand_name>>> to help monitor and analyze various abnormal situations of PostgreSQL.

## Configuration {#config}

### Preconditions {#reqirement}

- PostgreSQL version >= 9.0
- Create user

    ```sql
    -- PostgreSQL >= 10
    create user datakit with password '<PASSWORD>';
    grant pg_monitor to datakit;
    grant SELECT ON pg_stat_database to datakit;

    -- PostgreSQL < 10
    create user datakit with password '<PASSWORD>';
    grant SELECT ON pg_stat_database to datakit;
    ```

- Enable the `pg_stat_statements` Extension (Optional)

    When collecting [PostgreSQL Object](postgresql.md#object), some key metrics such as `qps/tps/avg_query_time` rely on the `pg_stat_statements` extension. The specific steps to enable this extension are as follows:

    - **Modify the Configuration File**

        Locate and edit the PostgreSQL configuration file (common paths are `/var/lib/pgsql/data/postgresql.conf` or `/etc/postgresql/<version>/main/postgresql.conf`), and add or modify the following configuration items:

        ```ini
        # Enable the pg_stat_statements
        shared_preload_libraries = 'pg_stat_statements'

        # Optional configuration
        pg_stat_statements.track = 'all'  # collect all SQL statements
        pg_stat_statements.max = 10000  # max number of SQL statements to track
        pg_stat_statements.track_utility = off # ignore utility statements, only track regular SQL statements like `SELECT`, `INSERT`, `UPDATE`, `DELETE`

        ```

    - **Restart the PostgreSQL Service**

        After modifying the configuration file, it is necessary to restart the PostgreSQL service for the configurations to take effect.​

    - **Create the Extension**

        Connect to the target database and execute the following SQL statement to create the extension:

        ```sql
        CREATE EXTENSION pg_stat_statements;
        ```

        - **Verify the Successful Enablement of the Extension:**

        ```sql
        SELECT * FROM pg_extension WHERE extname = 'pg_stat_statements';
        SELECT * FROM pg_stat_statements LIMIT 10;
        ```

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/samples` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->

### Database Performance Metrics Collection {#performance-schema}

[:octicons-tag-24: Version-1.84.0](../datakit/changelog-2025.md#cl-1.84.0)

Database performance metrics are mainly derived from the built-in system views and extension plugins of PostgreSQL, with the most core ones including pg_stat_activity and pg_stat_statements. These tools provide methods to obtain the internal execution status of the database at runtime: pg_stat_activity displays real-time information such as the activity status of current sessions, executed queries, and waiting events; pg_stat_statements records the execution statistics of historical SQL statements, including execution times, time consumption, IO status, etc.
Through these views and plugins, DataKit can collect real-time session activities, performance metric statistics of historical queries, and relevant execution information. The collected performance metric data is saved as logs, with the sources (source) being `postgresql_dbm_metric`, `postgresql_dbm_sample`, and `postgresql_dbm_activity` respectively.

To enable this feature, the following steps need to be performed.

- Modify the configuration file to enable monitoring and collection

```toml
[[inputs.postgresql]]

  ## Set dbm to true to collect database activity 
  dbm = false

  ## Config dbm metric 
  [inputs.postgresql.dbm_metric]
    enabled = true
  
  ## Config dbm sample 
  [inputs.postgresql.dbm_sample]
    enabled = true  

  ## Config dbm activity
  [inputs.postgresql.dbm_activity]
    enabled = true  

...

```

- PostgreSQL Configuration

Modify the configuration file (e.g., `postgresql.conf`) and configure the relevant parameters:

```toml

shared_preload_libraries = 'pg_stat_statements'
track_activity_query_size = 4096 a# Required for collection of larger queries.

```

- Permission Configuration

<!-- markdownlint-disable MD046 -->
=== "Postgres >=15"

    Account Authorization

    ```sql
    ALTER ROLE datakit INHERIT;
    ```

    Execute the following SQL in each database:

    ```sql
    CREATE SCHEMA datakit;
    GRANT USAGE ON SCHEMA datakit TO datakit;
    GRANT USAGE ON SCHEMA public TO datakit;
    GRANT pg_monitor TO datakit;
    CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

    ```

=== "Postgres >=10"

    Execute the following SQL in each database:

    ```sql
    CREATE SCHEMA datakit;
    GRANT USAGE ON SCHEMA datakit TO datakit;
    GRANT USAGE ON SCHEMA public TO datakit;
    GRANT pg_monitor TO datakit;
    CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

    ```

=== "Postgres 9.6"

    Execute the following SQL in each database:

    ```sql
    CREATE SCHEMA datakit;
    GRANT USAGE ON SCHEMA datakit TO datakit;
    GRANT USAGE ON SCHEMA public TO datakit;
    GRANT SELECT ON pg_stat_database TO datakit;
    CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

    CREATE OR REPLACE FUNCTION datakit.pg_stat_activity() RETURNS SETOF pg_stat_activity AS
      $$ SELECT * FROM pg_catalog.pg_stat_activity; $$
    LANGUAGE sql
    SECURITY DEFINER;
    CREATE OR REPLACE FUNCTION datakit.pg_stat_statements() RETURNS SETOF pg_stat_statements AS
        $$ SELECT * FROM pg_stat_statements; $$
    LANGUAGE sql
    SECURITY DEFINER;

    ```

<!-- markdownlint-enable -->

## Metric {#metric}

For all of the following data collections, the global election tags will added automatically, we can add extra tags in `[inputs.{{.InputName}}.tags]` if needed:

{{ range $i, $m := .Measurements }}
{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}
{{ end }}

## Object {#object}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "object"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}
{{end}}

{{ end }}


### Structure of the `message` field {#message-struct}

The basic structure of the `message` field is as follows:​

```json
{
  "setting": {
    "DateStyle":"ISO, MDY",
    ...
  },

  "databases": [ # databases information
    {
      "name": "db1",
      "encoding": "utf8",
      "owner": "datakit",
      "schemas": [ # schemas information
        {
          "name": "schema1",
          "owner": "datakit",
          "tables": [ # tables information
            {
              "name": "table1",
              "columns": [], # columns information
              "indexes": [], # indexes information
              "foreign_keys": [], # foreign keys information
            }
            ...
          ]
        },
        ...
      ]
    }
    ...
  ]
}
```

#### `setting` {#setting}

  The data in the `setting` field is sourced from the `pg_settings` system view and is used to display the configuration parameter information of the current database. For more details, please refer to the [PostgreSQL documentation](https://www.postgresql.org/docs/current/view-pg-settings.html){:target="_blank"}.

#### `databases` {#databases}

The `databases` field stores information about all databases on the `PostgreSQL` server. The detailed information of each database is shown in the following table:

| Field Name              | Description                 | Type   |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | Database name                                           | string |
| `encoding`        | Database encoding                                        | string |
| `owner`        | Role name                                            | string |
| `description`        | Description text                                      | string |
| `schemas`        | List containing `schema` information                           | list |

The `schemas` field contains information about all the `schemas` in the database. The information for each `schema` is as follows:

| Field Name              | Description                 | Type   |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | `schema` name                                           | string |
| `owner`        | Role name                                            | string |
| `tables`        | List containing `table` information                           | list |

The `tables` field contains information about all the `tables` in the database. The information for each `table` is as follows:

| Field Name              | Description                 | Type   |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | `table` name                                         | string |
| `owner`        | `table` owner                                            | string |
| `has_indexes`        | Whether there are indexes                             | bool |
| `has_partitions`        | Whether there are partitions                             | bool |
| `toast_table`        | `toast` table name                             | string |
| `partition_key`        | Partition key                             | string |
| `num_partitions`        | Number of partitions                             | int64 |
| `foreign_keys`        | List containing foreign key information                             | list |
| `columns`        | List containing `column` information                           | list |
| `indexes`        | List containing `index` information                           | list |

- `tables.columns`

The `columns` field contains information about all the `columns` in the `tables`. The information for each `column` is as follows:

| Field Name              | Description                 | Type   |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | Name                                           | string |
| `data_type`        | Data type                                            | string |
| `nullable`        | Whether `column` can be null                                            | bool |
| `default`        | Default value                                            | string |

- `tables.indexes`

The `indexes` field contains information about all the `indexes` in the database tables. The information for each `index` is as follows:

| Field Name              | Description                 | Type   |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | Name                                           | string |
| `columns`     | Columns included in the index                                              | list |
| `index_type`  | Index type                                                  | string |
| `definition`        | Index definition                                            | string |
| `is_unique`        | Whether unique                                            | bool |
| `is_primary`        | Whether primary                                            | bool |
| `is_exclusion`        | Whether exclusion constraint index                                            | bool |
| `is_immediate`        | Whether to check constraints immediately after each statement execution                      | bool |
| `is_valid`        | Whether it is valid                                            | bool |
| `is_clustered`        | Whether it is a clustered index                                            | bool |
| `is_checkxmin`        | Whether to check `xmin`                                            | bool |
| `is_ready`        | Whether it is ready                                            | bool |
| `is_live`        | Whether it is live                                            | bool |
| `is_replident`        | Whether it is a row identifier index                                            | bool |
| `is_partial`        | Whether it is a partial index                                            | bool |

The `indexes.columns` field, which contains information about the columns included in the `index`, has the following details for each column:

| Field Name              | Description                 | Type   |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | Name                                         | string |

- `tables.foreign_keys`

The `foreign_keys` field contains information about all the `foreign_keys` in the database tables. The information for each `foreign_key` is as follows:

| Field Name              | Description                 | Type   |
| ------------------:| ---------------------------------------------- | :----: |
| `name`        | Name                                           | string |
| `definition`        | Foreign key definition                                            | string |
| `constraint_schema` |Foreign key schema                                            | string |
| `column_names` |Foreign key column names                                            | string |
| `referenced_table_schema` |Referenced table schema                                            | string |
| `referenced_table_name` |Referenced table name                                            | string |
| `referenced_column_names` |Referenced column names                                            | string |
| `update_action` |Cascade update rule (such as CASCADE, RESTRICT)                                            | string |
| `delete_action` |Cascade delete rule (such as CASCADE, SET NULL)                                            | string |

## Logging {#logging}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}
{{end}}

{{ end }}

## File log {#file-log}

### Log Collection {#log}

- PostgreSQL logs are output to `stderr` by default. To open file logs, configure them in postgresql's configuration file `/etc/postgresql/<VERSION>/main/postgresql.conf` as follows:

```toml
logging_collector = on    # Enable log writing to files

log_directory = 'pg_log'  # Set the file storage directory, absolute path or relative path (relative PGDATA)

log_filename = 'pg.log'   # Log file name
log_statement = 'all'     # Record all queries

#log_duration = on
log_line_prefix= '%m [%p] %d [%a] %u [%h] %c ' # 日志行前缀
log_file_mode = 0644

# For Windows
#log_destination = 'eventlog'
```

For more configuration, please refer to the [doc](https://www.postgresql.org/docs/11/runtime-config-logging.html){:target="_blank"}。

- The PostgreSQL collector does not have log collection enabled by default. You can open `files` in `conf.d/db/postgresql.conf`  and write to the absolute path of the PostgreSQL log file. For example:

```toml
[[inputs.postgresql]]

  ...

  [inputs.postgresql.log]
  files = ["/tmp/pgsql/postgresql.log"]
```

When log collection is turned on, a log with a log `source` of `postgresql` is generated by default.

**Notices:**

- Log collection only supports logs on hosts where DataKit is installed.

### Log Pipeline Cut {#pipeline}

The original log is

``` log
2021-05-31 15:23:45.110 CST [74305] test [pgAdmin 4 - DB:postgres] postgres [127.0.0.1] 60b48f01.12241 LOG:  statement:
        SELECT psd.*, 2^31 - age(datfrozenxid) as wraparound, pg_database_size(psd.datname) as pg_database_size
        FROM pg_stat_database psd
        JOIN pg_database pd ON psd.datname = pd.datname
        WHERE psd.datname not ilike 'template%'   AND psd.datname not ilike 'rdsadmin'
        AND psd.datname not ilike 'azure_maintenance'   AND psd.datname not ilike 'postgres'
```

Description of the cut field:

| Field name         | Field Value               | Description                                                    |
| ------------------ | ------------------------- | -------------------------------------------------------------- |
| `application_name` | `pgAdmin 4 - DB:postgres` | The name of the application connecting to the current database |
| `db_name`          | `test`                    | Database accessed                                              |
| `process_id`       | `74305`                   | The client process ID of the current connection                |
| `remote_host`      | `127.0.0.1`               | Address of the client                                          |
| `session_id`       | `60b48f01.12241`          | ID of the current session                                      |
| `user`             | `postgres`                | Current Access User Name                                       |
| `status`           | `LOG`                     | Current log level (LOG,ERROR,FATAL,PANIC,WARNING,NOTICE,INFO)  |
| `time`             | `1622445825110000000`     | Log generation time                                            |

## FAQ {#faq}

<!-- markdownlint-disable MD013 -->
### Missing metrics {#faq-missing-relation-metrics}

For metrics `postgresql_lock/postgresql_stat/postgresql_index/postgresql_size/postgresql_statio`, the `relations` field in the configuration file needs to be enabled. If some of these metrics are partially missing, it may be because there is no data for the relevant metrics.

<!-- markdownlint-enable -->
