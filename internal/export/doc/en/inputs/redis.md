---
title     : 'Redis'
summary   : 'Collect Redis metrics and logs'
__int_icon      : 'icon/redis'
dashboard :
  - desc  : 'Redis'
    path  : 'dashboard/en/redis'
monitor:
  - desc: 'Redis'
    path: 'monitor/en/redis'
---

<!-- markdownlint-disable MD025 -->
# Redis
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Redis indicator collector, which collects the following data:

- Turn on AOF data persistence and collect relevant metrics
- RDB data persistence metrics
- Slow Log monitoring metrics
- Big Key scan monitoring
- Master-slave replication

## Configuration {#config}

Already tested version:

- [x] 7.0.11
- [x] 6.2.12
- [x] 6.0.8
- [x] 5.0.14
- [x] 4.0.14

### Precondition {#reqirement}

- Redis version v5.0+

When collecting data under the master-slave architecture, please configure the host information of the slave node for data collection, and you can get the metric information related to the master-slave.

Create Monitor User (**optional**)

redis6.0+ goes to the `redis-cli` command line, create the user and authorize

```sql
ACL SETUSER username >password
ACL SETUSER username on +@dangerous +ping
```

- goes to the `redis-cli` command line, authorization statistics `hotkey/bigkey` information

```sql
CONFIG SET maxmemory-policy allkeys-lfu
ACL SETUSER username on +get +@read +@connection +@keyspace ~*
```

- collect hotkey & `bigkey` remote, need install redis-cli (collect local need not install it)

```shell
# ubuntu 
apt-get install redis-tools

# centos
yum install -y  redis
```

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .

---

???+ attention

    If it is Alibaba Cloud Redis and the corresponding username and PASSWORD are set, the `<PASSWORD>` should be set to `your-user:your-password`, such as `datakit:Pa55W0rd`.
<!-- markdownlint-enable -->

### Log Collection Configuration {#logging-config}

To collect Redis logs, you need to open the log file `redis.config` output configuration in Redis:

```toml
[inputs.redis.log]
    # Log path needs to be filled with absolute path
    files = ["/var/log/redis/*.log"]
```

<!-- markdownlint-disable MD046 -->
???+ attention

    When configuring log collection, you need to install the DataKit on the same host as the Redis service, or otherwise mount the log on the DataKit machine.
    
    In K8s, Redis logs can be exposed to stdout, and DataKit can automatically find its corresponding log.
<!-- markdownlint-enable -->

## Metrics {#metric}
<!-- markdownlint-disable MD009 -->
For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.{{.InputName}}.tags]`:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- field list

{{$m.FieldsMarkdownTable}} {{end}}

{{ end }}

## Logging {#logging}

<!-- markdownlint-disable MD024 -->
{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- field list

{{$m.FieldsMarkdownTable}} {{end}}

{{ end }}
<!-- markdownlint-enable -->
### Logging Pipeline {#pipeline}

The original log is:

```log
122:M 14 May 2019 19:11:40.164 * Background saving terminated with success
```

The list of cut fields is as follows:

| Field Name  | Field Value                                 | Description                                  |
| ---         | ---                                         | ---                                          |
| `pid`       | `122`                                       | process id                                   |
| `role`      | `M`                                         | role                                         |
| `serverity` | `*`                                         | service                                      |
| `statu`     | `notice`                                    | log level                                    |
| `msg`       | `Background saving terminated with success` | log content                                  |
| `time`      | `1557861100164000000`                       | Nanosecond timestamp (as line protocol time) |
