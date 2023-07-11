# jmx metric to DK
-----

*Author： 宋龙奇*

---
Send JMX metrics to DK. And how to customize indicators.

## requrements {#requrements}
Before collecting metrics from the JMX service, you must start the JMX service and open the port. 

Open parameters can refer to:

```shell
# default host is 127.0.0.1
-Dcom.sun.management.jmxremote \
-Dcom.sun.management.jmxremote.port=9000 \
-Dcom.sun.management.jmxremote.ssl=false \
-Dcom.sun.management.jmxremote.authenticate=false
```

check port:
```shell
[root@localhost ~]# netstat -anlp |grep 9000
tcp6       0     0 :::9000                 :::*                   LISTEN     9372/java
[root@localhost ~]#
```

check StatsD config [open statsd plugin](./statsd.md){:target="_blank"}

## JMXFetch {#ddtrace_jmxfetch}
By default, JVM information will be collected: JVM CPU, JVM MEM, JVM Thread, class, etc. Specific [index set list list](./jvm/#dd-jvm-measurement){:target="_blank"}

You can use `dd.jmxfetch.<INTEGRATION_NAME>.enabled=true` to enable the specified fetcher.

`INTEGRATION_NAME` list : [Three-party list supported](https://docs.datadoghq.com/integrations/){:target="_blank"}


## custom metrics {#custom_metric}
How to collect metrics through custom configuration.

Custom JVM thread metrics sent to DK : `jvm.total_thread_count`, `jvm.peak_thread_count`, `jvm.daemon_thread_count`

custom metric need add config file:

1. mkdir `/usr/local/ddtrace/conf.d`, Other directories can be used.
2. Create a configuration file under the folder `guance.d/conf.yaml`.
3. `conf.yaml` at end of doc.

start command is:
```shell
java -javaagent:/usr/local/dd-java-agent.jar \
  -Dcom.sun.management.jmxremote.host=127.0.0.1 \
  -Dcom.sun.manaagement.jmxremote.port=9012 \
  -Dcom.sun.management.jmxremote.ssl=false \
  -Dcom.sun.management.jmxremote.authenticate=false \
  -Ddd.jmxfetch.config.dir="/usr/local/ddtrace/conf.d/" \
  -Ddd.jmxfetch.config="guance.d/conf.yaml" \
  -jar tmall.jar
```

conf.yaml is:
```yaml
init_config:
  is_jmx: true
  collect_default_metrics: true

instances:
  - jvm_direct: true
    host: localhost
    port: 9012
    conf: 
      - include:
          domain: java.lang
          type: Threading
          attribute:
            TotalStartedThreadCount:
              alias: jvm.total_thread_count
              metric_type: gauge
            PeakThreadCount:
              alias: jvm.peak_thread_count
              metric_type: gauge
            DaemonThreadCount:
              alias: jvm.daemon_thread_count
              metric_type: gauge
```