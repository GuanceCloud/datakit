# Java Example

---

## Install Dependency {#dependence}

Download the latest ddtrace agent dd-java-agent.jar, see [Download Instructions](ddtrace.md#doc-example).

## Run {#run}

You can run your java Code in a variety of ways, such as IDE, Maven, Gradle, or directly through the java-jar command. Start the application with the java command below:

```shell
java -javaagent:/path/to/dd-java-agent.jar \
-Ddd.logs.injection=true \
-Ddd.service.name=my-app \
-Ddd.env=staging \
-Ddd.version=1.0.0 \
-Ddd.agent.host=localhost \
-Ddd.trace.agent.port=9529 \
-jar path/to/your/app.jar
```

## Startup Parameters {#start-options}

- dd.env: Set the environment variable for the service, corresponding to the environment variable DD_ENV.
- dd.version: APP version number, corresponding to the environment variable DD_VERSION.
- dd.service.name: Set the service name corresponding to the environment variable DD_SERVICE.
- dd.trace.agent.timeout: The client network send timeout defaults to 10s, corresponding to the environment variable DD_TRACE_AGENT_TIMEOUT.
- dd.logs.injection: Whether to start Java application log injection, so that the log can be associated with link data, the default is true, corresponding to the environment variable DD_LOGS_INJECTION.
- dd.tags: Add the default Tags for each Span, corresponding to the environment variable DD_TAGS.
- dd.agent.host: The name of the address where Datakit listens, default localhost, corresponding to the environment variable DD_AGENT_HOST.
- dd.trace.agent.port: The port number on which Datakit listens, default 9529, corresponding to the environment variable DD_TRACE_AGENT_PORT.
- dd.trace.sample.rate: Set the sampling rate from 0.0 (0%) to 1.0 (100%).
- dd.jmxfetch.enabled: Start JMX metrics collection, default value is true, corresponding to environment variable DD_JMXFETCH_ENABLED.
- dd.jmxfetch.config.dir: Additional JMX metrics collection configuration directory. The Java Agent will look for jvm_direct: true in the instance section of the yaml configuration file to modify the configuration, corresponding to the environment variable DD_JMXFETCH_CONFIG_DIR.
- dd.jmxfetch.config: Additional JMX metrics collection configuration file. The JAVA agent will look for jvm_direct: true in the instance section of the yaml configuration file to modify the configuration corresponding environment variable,  DD_JMXFETCH_CONFIG
- dd.jmxfetch.check-period: JMX metrics send frequency (ms), default 1500, corresponding to the environment variable DD_JMXFETCH_CHECK_PERIOD.
- dd.jmxfetch.refresh-beans-period: Refresh the JMX beans frequency (s), default value 600, corresponding to the environment variable  DD_JMXFETCH_REFRESH_BEANS_PERIOD.
- dd.jmxfetch.statsd.host: The Statsd host address is used to receive JMX metrics, and if you use unix Domain Socket, use a host address like `unix://PATH_TO_UDS_SOCKET`. The default value is the same as agent. host, corresponding to the environment variable DD_JMXFETCH_STATSD_HOST.
- dd.jmxfetch.statsd.port: The StatsD port number is used to receive JMX metrics. If using Unix Domain Socket, make 0. The default value is the same as the environment variable DD_JMXFETCH_STATSD_PORT corresponding to agent. port.
