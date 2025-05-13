---
title     : 'DDTrace Java'
summary   : 'Tracing Java application with DDTrace'
tags      :
  - 'DDTRACE'
  - 'JAVA'
  - 'APM'
  - 'TRACING'
__int_icon: 'icon/ddtrace'
---

Integrating APM into Java is quite convenient, as it does not require any modifications to the business code; you only need to inject the corresponding agent.

## Preconditions {#requirements}

Install DataKit and enable the [DDTrace Collector](ddtrace.md){:target="_block"}. If you want to collect some JVM runtime metrics, you need to enable the [StatsD Collector](statsd.md){:target="_block"}.

Require JDK version 1.8 or above.

## Install Dependencies {#dependence}

<!-- markdownlint-disable MD046 -->
=== "Forked Version"

    To add support for more middleware, we have enhanced the [DDTrace-Java implementation](ddtrace-ext-java.md).

    ```shell
    wget -O dd-java-agent.jar 'https://static.<<<custom_key.brand_main_domain>>>/dd-image/dd-java-agent.jar'
    ```

=== "Datadog Version"

    ```shell
    wget -O dd-java-agent.jar 'https://dtdg.co/latest-java-tracer'
    ```
<!-- markdownlint-enable -->

## Running the Application {#instrument}

<!-- markdownlint-disable MD046 -->
=== "Host Application"

    You can run your Java code through various means, such as IDE, Maven, Gradle, or directly via the `java -jar` command. The following example starts the application with the `java` command:

    ```shell hl_lines="2-7" linenums="1"
    java \
      -javaagent:/path/to/dd-java-agent.jar \
      -Ddd.logs.injection=true \
      -Ddd.service.name=<YOUR-SERVICE-NAME> \
      -Ddd.env=<YOUR-ENV-NAME> \
      -Ddd.agent.host=<YOUR-DATAKIT-HOST> \
      -Ddd.trace.agent.port=9529 \
      -jar path/to/your/app.jar
    ```

    Fill in your basic parameter configurations for `<YOUR-...>` here. In addition to these, there are some optional parameters as follows:

    ### Enable Profiling {#instrument-profiling}

    > The [Profiling Collector](profile.md) needs to be enabled here.

    After enabling Profiling, we can see more information about Java runtime:

    ```shell linenums="1" hl_lines="3-4"
    java \
      -javaagent:/path/to/dd-java-agent.jar \
      -Ddd.profiling.enabled=true \
      -XX:FlightRecorderOptions=stackdepth=256 \
      ...
    ```

    ### Enable Sampling Rate {#instrument-sampling}

    We can enable a sampling rate to reduce the actual amount of data generated:

    ```shell hl_lines="3" linenums="1"
    java \
      -javaagent:/path/to/dd-java-agent.jar \
      -Ddd.trace.sample.rate=0.8 \
      ...
    ```

    ### Enable JVM Metrics Collection {#instrument-jvm-metrics}

    > The [statsd Collector](statsd.md) needs to be enabled here.

    ```shell hl_lines="3-6" linenums="1"
    java \
      -javaagent:/path/to/dd-java-agent.jar \
      -Ddd.jmxfetch.enabled=true \
      -Ddd.jmxfetch.check-period=1000 \
      -Ddd.jmxfetch.statsd.host=<YOUR-DATAKIT-HOST>  \
      -Ddd.jmxfetch.statsd.port=8125 \
      ...
    ```

=== "Kubernetes"

    In Kubernetes, you can inject the trace agent through the [DataKit Operator](../datakit/datakit-operator.md#datakit-operator-inject-lib), or manually mount the trace agent into the application container.

    ```yaml hl_lines="10-19" linenums="1"
    apiVersion: apps/v1
    kind: Deployment
    spec:
      template:
        spec:
          containers:
            - name: <CONTAINER_NAME>
              image: <CONTAINER_IMAGE>/<TAG>
              env:
                - name: DD_AGENT_HOST
                  value: "datakit-service.datakit.svc"
                - name: DD_TRACE_AGENT_PORT
                  value: "9529"
                - name: DD_ENV
                  value: <YOUR-ENV-NAME>
                - name: DD_SERVICE
                  value: <YOUR-SERVICE-NAME>
                - name: DD_LOGS_INJECTION
                  value: "true"
    ```

    For more other parameter settings, refer to the corresponding ENV fields in the [Parameter Explanation](ddtrace-java.md#start-options) below.
<!-- markdownlint-enable -->

## Parameter Explanation {#start-options}

Below are the explanations for each command-line parameter and their corresponding environment variable configurations. For full parameter support, refer to the [DataDog Official Documentation](https://docs.datadoghq.com/tracing/trace_collection/library_config/java){:target="_blank"}.

- **`dd.env`**

    **ENV**: `DD_ENV`

    Set the environment information for the service, such as `testing/prod`.

- **`dd.version`**

    **ENV**: `DD_VERSION`

    The version number of the APP.

- **`dd.service.name`**

    Set the service name.
    **ENV**: `DD_SERVICE`

- **`dd.trace.agent.timeout`**

    **ENV**: `DD_TRACE_AGENT_TIMEOUT`

    The client network send timeout defaults to 10s.

- **`dd.logs.injection`**

    **ENV**: `DD_LOGS_INJECTION`

    Whether to enable Java application log injection to associate logs with trace data, defaults to true.

- **`dd.tags`**

    **ENV**: `DD_TAGS`

    Add default Tags to each Span.

- **`dd.agent.host`**

    **ENV**: `DD_AGENT_HOST`

    The hostname where DataKit is listening, default is localhost.

- **`dd.trace.agent.port`**

    **ENV**: `DD_TRACE_AGENT_PORT`

    The port number where DataKit is listening, default is 9529.

- **`dd.trace.sample.rate`**

    **ENV**: `DD_TRACE_SAMPLE_RATE`

    Set the sampling rate from 0.0 (0%) to 1.0 (100%).

- **`dd.jmxfetch.enabled`**

    **ENV**: `DD_JMXFETCH_ENABLED`

    Enable JMX metrics collection, default value is true.

- **`dd.jmxfetch.config.dir`**

    **ENV**: `DD_JMXFETCH_CONFIG_DIR`

    Extra JMX metrics collection configuration directory. The Java Agent will look for `jvm_direct: true` in the instance section of the yaml configuration file to modify the configuration.

- **`dd.jmxfetch.config`**

    **ENV**: `DD_JMXFETCH_CONFIG`

    Extra JMX metrics collection configuration file. The JAVA agent will look for `jvm_direct: true` in the instance section of the yaml configuration file to modify the configuration.

- **`dd.jmxfetch.check-period`**

    **ENV**: `DD_JMXFETCH_CHECK_PERIOD`

    The frequency of sending JMX metrics (ms), default value is 1500.

- **`dd.jmxfetch.refresh-beans-period`**

    **ENV**: `DD_JMXFETCH_REFRESH_BEANS_PERIOD`

    The frequency of refreshing JMX beans (s), default value is 600.

- **`dd.jmxfetch.statsd.host`**

    **ENV**: `DD_JMXFETCH_STATSD_HOST`

    The Statsd host address for receiving JMX metrics, if using Unix Domain Socket please use a host address like `unix://PATH_TO_UDS_SOCKET`. The default value is the same as `agent.host`.

- **`dd.jmxfetch.statsd.port`**

    **ENV**: `DD_JMXFETCH_STATSD_PORT`

    The StatsD port number for receiving JMX metrics, if using Unix Domain Socket please fill in 0. The default value is the same as agent.port.

- **`dd.profiling.enabled`**

    **ENV**: `DD_PROFILING_ENABLED`

    Enable Profiling control, after enabling, the Profiling information during the Java application runtime will also be collected and reported to DataKit.


## More {#more-reading}

- Secondary development version [DDTrace JAVA extend](ddtrace-ext-java.md){:target="_blank"}
- By default, JVM metrics will be collected, specific metrics: [metrics](jvm.md#metric){:target="_blank"}
