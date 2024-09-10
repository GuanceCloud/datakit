---
title     : 'OceanBase'
summary   : 'Collect OceanBase metrics'
__int_icon      : 'icon/oceanbase'
dashboard :
  - desc  : 'OceanBase'
    path  : 'dashboard/en/oceanbase'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# OceanBase
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Collecting OceanBase performance metrics through the sys tenant.

Already tested version:

- [x] OceanBase Enterprise 3.2.4

## Configuration {#config}

### Precondition {#reqirement}

- Create a monitoring account

Create a monitoring account using a sys tenant account and grant the following privileges:

```sql
CREATE USER 'datakit'@'localhost' IDENTIFIED BY '<UNIQUEPASSWORD>';

-- MySQL 8.0+ create the datakit user with the caching_sha2_password method
CREATE USER 'datakit'@'localhost' IDENTIFIED WITH caching_sha2_password by '<UNIQUEPASSWORD>';

-- Grant the required permissions 
GRANT SELECT ON *.* TO 'datakit'@'localhost';
```

<!-- markdownlint-disable MD046 -->
???+ attention

    - Note that if you find the collector has the following error when using `localhost` , you need to replace the above `localhost` with `::1` <br/>
    `Error 1045: Access denied for user 'datakit'@'localhost' (using password: YES)`

    - All the above creation and authorization operations limit that the user `datakit` can only access OceanBase on local host (`localhost`). If OceanBase is collected remotely, it is recommended to replace `localhost` with `%` (indicating that DataKit can access OceanBase on any machine), or use a specific DataKit installation machine address.


### Collector Configuration {#input-config}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).

## Long Running Queries {#slow}

Datakit could reports the SQLs, those executed time exceeded the threshold time defined by user, to Guance Cloud, displays in the `Logs` side bar, the source name is `oceanbase_log`.

This function is disabled by default, user could enabling it by modify Datakit's OceanBase configuration like followings:

Change the string value after `slow_query_time` from `0s` to the threshold time, minimal value is 1 millsecond. Generally, recommand it to `10s`.

```conf

slow_query_time = "0s"

```

???+ info "Fields description"
    - `failed_obfuscate`：SQL obfuscated failed reason. Only exist when SQL obfuscated failed. Original SQL will be reported when SQL obfuscated failed.
    [More fields](https://www.oceanbase.com/docs/enterprise-oceanbase-database-cn-10000000000376688){:target="_blank"}.

???+ attention "Attention"
    - If the string value after `--slow-query-time` is `0s` or empty or less than 1 millisecond, this function is disabled, which is also the default state.
    - The SQL would not display here when NOT executed completed.

<!-- markdownlint-enable -->
## Metric {#metric}

For all of the following data collections, the global election tags will added automatically, we can add extra tags in `[inputs.{{.InputName}}.tags]` if needed:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric" }}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- Metric list

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## Log {#logging}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- Metric list

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}
