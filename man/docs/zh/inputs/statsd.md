
# StatsD 数据接入

---

{{.AvailableArchs}}

---

DDTrace agent 采集的指标数据会通过 StatsD 数据类型发送到 DK 的 8125 端口上。
其中包括 JVM 运行时的 CPU 、内存、线程、类加载信息，也包括开启的各种采集上来的 JMX 指标， 如： Kafka、Tomcat、RabbitMQ 等。

## 前置条件 {#requrements}

DDTrace 以 agent 形式运行时，不需要用户特意的开通 jmx 端口，如果没有开通端口的话， agent 会随机打开一个本地端口。

DDTrace 默认会采集 JVM 信息。默认情况下会发送到 `localhost:8125`.

如果是 k8s 环境下，需要配置 StatsD host 和 port：

```shell
DD_JMXFETCH_STATSD_HOST=datakit_url
DD_JMXFETCH_STATSD_PORT=8125
```

可以使用 `dd.jmxfetch.<INTEGRATION_NAME>.enabled=true` 开启指定的采集器。

填写 `INTEGRATION_NAME` 之前可以先查看 [默认支持的三方软件](https://docs.datadoghq.com/integrations/){:target="_blank"}

比如 Tomcat 或者 Kafka：

```shell
-Ddd.jmxfetch.tomcat.enabled=true
# or
-Ddd.jmxfetch.kafka.enabled=true 
```

## 配置 {#config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，重启 DataKit 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

## 指标集 {#measurement}

statsD 暂无指标集定义，所有指标以网络发送过来的指标为准。

使用 Agent 默认的指标集的情况下，[GitHub 上可以查看所有的指标集](https://docs.datadoghq.com/integrations/){:target="_blank"}
