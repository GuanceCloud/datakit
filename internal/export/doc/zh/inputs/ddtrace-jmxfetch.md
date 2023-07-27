
# 指标暴露和自定义指标

---

## JMXFetch {#ddtrace-jmxfetch}

DDTrace 以 agent 形式运行时，不需要用户特意的开通 jmx 端口，如果没有开通端口的话， agent 会随机打开一个本地端口。

JMXFetch 是从 JMX 服务器收集指标以 statsD 数据结构形式向外发送。本身集成在 *dd-java-agent* 中。

默认会采集 JVM 信息：JVM CPU、 Mem、Thread、Class 等。 具体 [指标集列表](jvm.md#metric)

默认情况下采集到的指标信息发送到 `localhost:8125` 确定已经 [开启 statsd 采集器](statsd.md) 是开启的。

如果是 k8s 环境下，需要配置 StatsD host 和 port：

```shell
DD_JMXFETCH_STATSD_HOST=datakit_url
DD_JMXFETCH_STATSD_PORT=8125
```

可以使用 `dd.jmxfetch.<INTEGRATION_NAME>.enabled=true` 开启指定的采集器。

填写 `INTEGRATION_NAME` 之前可以先查看 [默认支持的三方软件](https://docs.datadoghq.com/integrations/){:target="_blank"}

比如 tomcat：

```shell
-Ddd.jmxfetch.tomcat.enabled=true
```

## 如何通过自定义配置采集指标 {#custom-metric}

自定义 JVM 线程状态指标

- `jvm.total_thread_count`
- `jvm.peak_thread_count`
- `jvm.daemon_thread_count`

> `dd-java-agent` 从 v1.17.3-guance 开始已经将这三个指标内置，不需要再额外配置。但这种自定义的方式依然可以配置其他的 MBean 指标。

自定义指标需要增加配置文件：

1. 创建文件夹 */usr/local/ddtrace/conf.d* 目录随意（注意权限），下面用的着。
1. 在文件夹下创建配置文件 *guance.d/conf.yaml*, 文件必须是 yaml 格式。
1. *conf.yaml* 文件配置看最后

我的服务名为 `tmall.jar` 合并启动参数为：

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

conf.yaml 配置文件如下：

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
