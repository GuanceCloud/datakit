---
title     : 'StatsD'
summary   : 'Collect metrics reported by StatsD'
__int_icon      : 'icon/statsd'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# StatsD data access
<!-- markdownlint-enable -->
---

{{.AvailableArchs}}

---

The indicator data collected by the DDTrace agent will be sent to port 8125 of the DK through the StatsD data type.

This includes the JVM CPU, memory, threads, and class loading information of the JVM runtime, as well as various collected JMX indicators such as Kafka, Tomcat, RabbitMQ, etc.

## Config {#config}

### Preconditions {#requrements}

When DDTrace runs as an agent, there is no need for the user to specifically open the jmx port. If no port is opened, the agent will randomly open a local port.

DDTrace will collect JVM information by default. By default, it will be sent to 'localhost: 8125'

if k8s:

```shell
DD_JMXFETCH_STATSD_HOST=datakit_url
DD_JMXFETCH_STATSD_PORT=8125
```

You can use ` dd.jmxfetch.<INTEGRATION_NAME>.enabled=true ` Enable the specified collector.

for `INTEGRATION_NAME`, You can check the [default supported third-party software](https://docs.datadoghq.com/integrations/){:target="_blank"} before.

For example, Tomcat or Kafka:

```shell
-Ddd.jmxfetch.tomcat.enabled=true
# or
-Ddd.jmxfetch.kafka.enabled=true 
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

    The collector can now be turned on by [configMap injection collector configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD046 -->
???+ info

    If find lot of Feed: io busy in the log, can configure interval='1s', minimum is 1s.
<!-- markdownlint-enable -->

### Tag data sources {#config-tag}

If you want to tag the hosts captured by DDTrace, you can do so by injecting tags:

- You can use environment variables, i.e. `DD_TAGS`, for example: `DD_TAGS=source_key:tomcat,host_key:cn-shanghai-sq5ei`
- You can use the command line, i.e. `dd.tags`, for example：`-Ddd.tags=source_key:tomcat,host_key:cn-shanghai-sq5ei`

In the above example, you need to specify that the key for source is `source_key` and the key for host is `host_key` in the Datakit configuration. You can change it to something else, but you have to make sure that the field names in Datakit are the same as the field names in DDTrace.

The end result is that you can see `statsd/tomcat/cn-shanghai-sq5ei` in the `datakit monitor`, which distinguishes it from other data sources reported to the statsd collector. If the above configuration is not done, then the default display is seen on the `datakit monitor`: `statsd/-/-`.

In addition, there is a configuration switch `save_above_key` that determine whether the tags corresponding to `statsd_source_key` and `statsd_host_key` are reported to the center. The default is not to report(`false`).

## Metric {#metric}

Statsd has no measurement definition at present, and all metrics are subject to the metrics sent by the network.

For example, if Tomcat or Kafka uses the default indicator set, [GitHub can view all indicator sets](https://docs.datadoghq.com/integrations/){:target="_blank"}
