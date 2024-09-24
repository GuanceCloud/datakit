---
title   : 'Cassandra'
summary : 'Collect Cassandra metrics'
tags    :
  - 'DATA STORES'
__int_icon : 'icon/cassandra'
dashboard  :
  - desc   : 'Cassandra'
    path   : 'dashboard/en/cassandra'
monitor    :
  - desc   : 'Cassandra'
    path   : 'monitor/en/cassandra'
---


{{.AvailableArchs}}

---

Cassandra metrics can be collected by using [DDTrace](ddtrace.md).
The flow of the collected data is as follows: Cassandra -> DDTrace -> DataKit(StatsD).

You can see that Datakit has integrated the [StatsD](https://github.com/statsd/statsd){:target="_blank"} server, DDTrace collects Cassandra metric data and reports it to Datakit using StatsD protocol.

## Configuration {#config}

### Preconditions {#requrements}

- Already tested Cassandra version:
    - [x] 5.0
    - [x] 4.1.3
    - [x] 3.11.15
    - [x] 3.0.24
    - [x] 2.1.22

### DDtrace Configuration {#config-ddtrace}

- Download `dd-java-agent.jar`, see [here](ddtrace.md){:target="_blank"};

- Datakit configuration:

See the configuration of [StatsD](statsd.md){:target="_blank"}.

Restart Datakit to make configuration take effect.

- Cassandra configuration:

Create the file `setenv.sh` under `/usr/local/cassandra/bin` and give it execute permission, then write the following:

```sh
export CATALINA_OPTS="-javaagent:dd-java-agent.jar \
                      -Ddd.jmxfetch.enabled=true \
                      -Ddd.jmxfetch.statsd.host=${DATAKIT_HOST} \
                      -Ddd.jmxfetch.statsd.port=${DATAKIT_STATSD_HOST} \
                      -Ddd.jmxfetch.cassandra.enabled=true"
```

The parameters are described below:

- `javaagent`: Fill in the full path to `dd-java-agent.jar`;
- `Ddd.jmxfetch.enabled`: Fill in `true`, which means the DDTrace collection function is enabled;
- `Ddd.jmxfetch.statsd.host`: Fill in the network address that Datakit listens to. No port number is included;
- `Ddd.jmxfetch.statsd.port`: Fill in the port number that Datakit listens to. Usually `11002`, as determined by the Datakit side configuration;
- `Ddd.jmxfetch.cassandra.enabled`: Fill in `true`, which means the Cassandra collect function of DDTrace is enabled. When enabled, the metrics set named `cassandra` will showing up;

Restart Datakit to make configuration take effect.

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host deployment"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .

<!-- markdownlint-enable -->
---

## Metric {#metric}

<!-- markdownlint-disable MD024 -->
{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- Tags

{{$m.TagsMarkdownTable}}

- Fields

{{$m.FieldsMarkdownTable}}

{{ end }}
<!-- markdownlint-enable -->
