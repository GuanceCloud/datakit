
# Statsd Data Access
---

{{.AvailableArchs}}

---
The indicator data collected by the DDTrace agent will be sent to port 8125 of the DK through the StatsD data type.

This includes the JVM CPU, memory, threads, and class loading information of the JVM runtime, as well as various collected JMX indicators such as Kafka, Tomcat, RabbitMQ, etc.


## Preconditions {#requrements}

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


## Configuration {#config}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, restart DataKit.

=== "Kubernetes"

    The collector can now be turned on by [configMap injection collector configuration](datakit-daemonset-deploy.md#configmap-setting).

## Measurement {#measurement}

Statsd has no measurement definition at present, and all metrics are subject to the metrics sent by the network.

For example, if Tomcat or Kafka uses the default indicator set, [GitHub can view all indicator sets](https://docs.datadoghq.com/integrations/){:target="_blank"}
