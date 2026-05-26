---
title     : 'Doris'
summary   : 'Collect metrics of Doris'
tags:
  - 'DATABASE'
__int_icon      : 'icon/doris'
dashboard :
  - desc  : 'Doris'
    path  : 'dashboard/en/doris'
monitor   :
  - desc  : 'Doris'
    path  : 'monitor/en/doris'
---


{{.AvailableArchs}}

---

Doris collector uses the MySQL-compatible protocol on FE Query Port to collect Doris database objects and custom SQL metrics.
To collect full FE/BE Prometheus metrics, keep using the Prometheus collector configuration in the sample.

## Configuration {#config}

Already tested version:

- [x] 2.0.0

### Preconditions {#requirements}

Connect to FE Query Port with a Doris admin account and create the `datakit` collection account:

```sql
CREATE USER 'datakit'@'%' IDENTIFIED BY '123456';
GRANT NODE_PRIV ON *.*.* TO 'datakit'@'%';
```

Use the `datakit` account to check the FE SQL connection:

```shell
mysql -h <FE_HOST> -P 9030 -u datakit -p
```

After connecting, run the following SQL statements to verify the data required by object collection:

```sql
SHOW FRONTENDS;
SHOW BACKENDS;
```

If custom queries are configured, make sure the `datakit` account has query privileges on the related tables or views.

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "host installation"

    Go to the `conf.d/samples` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .

<!-- markdownlint-enable -->

### Custom Query {#custom-query}

Use `[[inputs.doris.custom_queries]]` to run custom SQL and report the query result as metrics.

```toml
[[inputs.doris.custom_queries]]
  sql = '''
    SELECT
      TABLE_SCHEMA AS table_schema,
      COUNT(*) AS table_count
    FROM information_schema.tables
    GROUP BY TABLE_SCHEMA
  '''
  metric = "doris_custom"
  tags = ["table_schema"]
  fields = ["table_count"]
  interval = "10s"
```

- `metric`: measurement name to report.
- `tags`: result columns used as tags.
- `fields`: result columns used as fields. Field values must be numeric.
- `interval`: collection interval for this query. If not configured, `inputs.doris.interval` is used.

## Metric {#metric}

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

{{$m.MarkdownTable}}

{{ end }}
{{ end }}
