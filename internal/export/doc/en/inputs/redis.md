<!-- markdownlint-disable MD025 -->
# Redis
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Redis indicator collector, which collects the following data:

- Turn on AOF data persistence and collect relevant metrics
- RDB data persistence metrics
- Slowlog monitoring metrics
- bigkey scan monitoring
- Master-slave replication

## Configuration {#config}

Already tested version:

- [x] 7.0.11
- [x] 6.2.12
- [x] 5.0.14
- [x] 4.0.14

### Precondition {#reqirement}

- Redis version v5.0+

When collecting data under the master-slave architecture, please configure the host information of the slave node for data collection, and you can get the metric information related to the master-slave.

Create Monitor User

redis6.0+ goes to the rediss-cli command line, create the user and authorize

```sql
ACL SETUSER username >password
ACL SETUSER username on +@dangerous
ACL SETUSER username on +ping
```

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap injection collector configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).

---

???+ attention

    If it is Alibaba Cloud Redis and the corresponding username and PASSWORD are set, the `<PASSWORD>` should be set to `your-user:your-password`, such as `datakit:Pa55W0rd`.
<!-- markdownlint-enable -->

### Log Collection {#redis-logging}

To collect Redis logs, you need to open the log file `redis.config` output configuration in Redis:

```toml
[inputs.redis.log]
    # Log path needs to be filled with absolute path
    files = ["/var/log/redis/*.log"]
```

???+ attention

    When configuring log collection, you need to install the DataKit on the same host as the Redis service, or otherwise mount the log on the DataKit machine.
    
    In K8s, Redis logs can be exposed to stdout, and DataKit can automatically find its corresponding log.

## Metrics {#metric}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.redis.tags]`:

``` toml
 [inputs.redis.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

#### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- feld list

{{$m.FieldsMarkdownTable}} {{end}}

{{ end }}

## Logging {#logging}

[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

#### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- field list

{{$m.FieldsMarkdownTable}} {{end}}

{{ end }}

### Logging Pipeline {#pipeline}

The original log is:

```
122:M 14 May 2019 19:11:40.164 * Background saving terminated with success
```

The list of cut fields is as follows:

| Field Name  | Field Value                                 | Description                                  |
| ----------- | ------------------------------------------- | -------------------------------------------- |
| `pid`       | `122`                                       | process id                                   |
| `role`      | `M`                                         | role                                         |
| `serverity` | `*`                                         | service                                      |
| `statu`     | `notice`                                    | log level                                    |
| `msg`       | `Background saving terminated with success` | log content                                  |
| `time`      | `1557861100164000000`                       | Nanosecond timestamp (as line protocol time) |
