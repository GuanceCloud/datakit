---
title     : 'DDTrace JMX'
summary   : '通过 DDTrace Java Agent 采集 JVM JMX 指标'
tags      :
  - 'DDTRACE'
  - 'JAVA'
  - 'JMX'
__int_icon: 'icon/ddtrace'
---

## JMXFetch {#ddtrace-jmxfetch}

JMXFetch 内置于 `dd-java-agent`，用于读取 JVM MBean 并以 DogStatsD 指标发送。它与 trace 接收完全独立：

`Java Agent JMXFetch` → `DogStatsD` → `DataKit StatsD` 采集器 → `<<<custom_key.brand_name>>>`

默认 JVM 配置使用 `jvm_direct: true` 在**当前 JVM 进程内**读取 MBean，因此通常不需要为了 JMXFetch 再开放 `com.sun.management.jmxremote.port`。远程 JMX 是另一种运维能力；如果业务确需开放它，应单独配置认证、TLS 和网络访问控制，绝不能按旧示例在生产环境关闭认证并暴露端口。

### 前置条件 {#prerequisites}

1. Java 应用已通过 [DDTrace Java](ddtrace-java.md) 加载 Agent。
1. DataKit 已启用 [StatsD 采集器](statsd.md)。其默认监听为 UDP `:8125`；跨主机或跨 Pod 时需确认 Service、NetworkPolicy 和 UDP 网络可达。
1. 应用侧 `DD_JMXFETCH_STATSD_HOST` 和 `DD_JMXFETCH_STATSD_PORT` 指向 DataKit 的 StatsD 接收端，而不是 trace 端口 `9529`。

### 默认指标与内置集成 {#default-metrics}

JMXFetch 通常会采集 JVM 的内存、线程、类加载和 GC 等运行时指标；实际指标会随 Java Agent 版本、JVM 和配置变化。请以 [JVM 指标集](jvm.md#metric)中的实际数据为准。

部分 Java Agent 版本还带有第三方产品的 JMX 配置。可通过如下形式按需启用：

```shell
-Ddd.jmxfetch.tomcat.enabled=true
```

其中 `tomcat` 是集成名。启用前先确认当前 Agent 版本内置了该集成，并只开启确实需要的检查，避免无效查询与额外指标基数。

## 配置 DataKit 接收 DogStatsD {#statsd-target}

本机部署可直接使用默认地址；跨容器或 Kubernetes 部署时，应显式配置：

```shell
DD_JMXFETCH_ENABLED=true \
DD_JMXFETCH_STATSD_HOST=datakit-service.datakit.svc.cluster.local \
DD_JMXFETCH_STATSD_PORT=8125 \
java -javaagent:/opt/dd-java-agent.jar \
  -Ddd.agent.host=datakit-service.datakit.svc.cluster.local \
  -Ddd.trace.agent.port=9529 \
  -jar /opt/my-app.jar
```

`DD_AGENT_HOST` / `DD_TRACE_AGENT_PORT` 配置 trace 目标；`DD_JMXFETCH_STATSD_HOST` / `DD_JMXFETCH_STATSD_PORT` 配置指标目标。两组配置不能互换。

## 通过自定义配置采集指标 {#custom-metric}

当内置指标不能满足需求时，可给 Java Agent 提供额外的 JMXFetch YAML。建议把配置以 ConfigMap、受版本控制的镜像文件或受管卷的形式提供，并保证 JVM 运行用户可读。

### 1. 创建配置目录 {#create-config-dir}

```shell
mkdir -p /opt/ddtrace/conf.d/ext.d
```

将下方内容保存为 `/opt/ddtrace/conf.d/ext.d/conf.yaml`。

### 2. 编写 YAML {#write-yaml}

以下示例直接读取当前 JVM 的 `java.lang:type=Threading` MBean；`jvm_direct: true` 时无需配置远程 JMX host、port 或关闭认证：

```yaml title="conf.yaml"
init_config:
  is_jmx: true
  collect_default_metrics: true

instances:
  - jvm_direct: true
    name: application-jvm
    conf:
      - include:
          domain: java.lang
          type: Threading
          attribute:
            ThreadCount:
              alias: jvm.thread.count
              metric_type: gauge
            PeakThreadCount:
              alias: jvm.thread.peak_count
              metric_type: gauge
            DaemonThreadCount:
              alias: jvm.thread.daemon_count
              metric_type: gauge
```

`alias` 是上报到 DataKit 的指标名；请使用稳定、可读且不会因实例 ID 等动态值无限增长的名称。新增大量 MBean 或标签前先评估指标基数和采集开销。

### 3. 让 Agent 加载配置 {#load-config}

通过目录与相对文件名指定配置：

```shell
java -javaagent:/opt/dd-java-agent.jar \
  -Ddd.jmxfetch.config.dir=/opt/ddtrace/conf.d \
  -Ddd.jmxfetch.config=ext.d/conf.yaml \
  -Ddd.jmxfetch.statsd.host=datakit-service.datakit.svc.cluster.local \
  -Ddd.jmxfetch.statsd.port=8125 \
  -jar /opt/my-app.jar
```

也可由环境变量配置为 `DD_JMXFETCH_CONFIG_DIR`、`DD_JMXFETCH_CONFIG`、`DD_JMXFETCH_STATSD_HOST` 和 `DD_JMXFETCH_STATSD_PORT`。同一配置不要同时设置相互冲突的 JVM 属性与环境变量。

### 4. 验证 {#verify}

1. 临时开启 Agent 启动日志或 JMXFetch 日志，确认 YAML 文件被加载且 DogStatsD 目标正确。
1. 检查 DataKit 的 StatsD 采集器已启动并监听预期协议和端口。
1. 等待一个采集周期后，在指标浏览中查找 `jvm.thread.count`、`jvm.thread.peak_count` 和 `jvm.thread.daemon_count`。
1. 若无数据，依次检查配置文件挂载路径/权限、UDP 连通性、JMXFetch 是否启用，以及 YAML 的缩进和 MBean 名称。

## 相关配置 {#configuration}

| JVM 属性 | 环境变量 | 说明 |
| --- | --- | --- |
| `dd.jmxfetch.enabled` | `DD_JMXFETCH_ENABLED` | 启用或关闭 JMXFetch。 |
| `dd.jmxfetch.check-period` | `DD_JMXFETCH_CHECK_PERIOD` | 发送指标的周期（毫秒）。 |
| `dd.jmxfetch.refresh-beans-period` | `DD_JMXFETCH_REFRESH_BEANS_PERIOD` | 刷新 MBean 列表的周期（秒）。 |
| `dd.jmxfetch.config.dir` | `DD_JMXFETCH_CONFIG_DIR` | 额外配置目录。 |
| `dd.jmxfetch.config` | `DD_JMXFETCH_CONFIG` | 额外 YAML 配置文件。 |
| `dd.jmxfetch.statsd.host` | `DD_JMXFETCH_STATSD_HOST` | DataKit StatsD 地址。 |
| `dd.jmxfetch.statsd.port` | `DD_JMXFETCH_STATSD_PORT` | DataKit StatsD 端口，通常为 `8125`。 |

完整参数与内置集成以 [Datadog Java 配置文档](https://docs.datadoghq.com/tracing/trace_collection/library_config/java/){:target="_blank"} 和所用 Agent 版本为准。
