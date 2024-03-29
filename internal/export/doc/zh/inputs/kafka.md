---
title     : 'Kafka'
summary   : '采集 Kafka 的指标数据'
__int_icon      : 'icon/kafka'
dashboard :
  - desc  : 'Kafka'
    path  : 'dashboard/zh/kafka'
monitor   :
  - desc  : 'Kafka'
    path  : 'monitor/zh/kafka'
---

<!-- markdownlint-disable MD025 -->
# Kafka
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

采集 Kafka 指标和日志上报到观测云，帮助你监控分析 Kafka 各种异常情况。

## 配置 {#config}

### 前置条件 {#requirements}

安装或下载 [Jolokia](https://search.maven.org/remotecontent?filepath=org/jolokia/jolokia-jvm/1.6.2/jolokia-jvm-1.6.2-agent.jar){:target="_blank"}。DataKit 安装目录下的 `data` 目录中已经有下载好的 Jolokia jar 包。

Jolokia 是作为 Kafka 的 Java agent，基于 HTTP 协议提供了一个使用 JSON 作为数据格式的外部接口，提供给 DataKit 使用。 Kafka 启动时，先配置 `KAFKA_OPTS` 环境变量：(port 可根据实际情况修改成可用端口）

```shell
export KAFKA_OPTS="$KAFKA_OPTS -javaagent:/usr/local/datakit/data/jolokia-jvm-agent.jar=host=*,port=8080"
```

另外，也可以单独启动 Jolokia，将其指向 Kafka 进程 PID：

```shell
java -jar </path/to/jolokia-jvm-agent.jar> --host 127.0.0.1 --port=8080 start <Kafka-PID>
```

<!-- markdownlint-disable MD046 -->

???+ attention

    Jolokia 不允许运行过程中修改端口号。如果发现通过 `--port` 命令无法修改端口号，就是这个原因。

    若想修改 Jolokia 端口号必须先退出 Jolokia 再启动才能成功。

???+ tip

    退出 Jolokia 命令是： `java -jar </path/to/jolokia-jvm-agent.jar> --quiet stop <Kafka-PID>`

    更多 Jolokia 命令信息可参考[这里](https://jolokia.org/reference/html/agents.html#jvm-agent){:target="_blank"}。

<!-- markdownlint-enable -->

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

## 指标 {#metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## 日志 {#logging}

如需采集 Kafka 的日志，可在 {{.InputName}}.conf 中 将 `files` 打开，并写入 kafka 日志文件的绝对路径。比如：

```toml
[[inputs.{{.InputName}}]]
  ...
  [inputs.{{.InputName}}.log]
    files = ["/usr/local/var/log/kafka/error.log","/usr/local/var/log/kafka/kafka.log"]
```

开启日志采集以后，默认会产生日志来源（`source`）为 `kafka` 的日志。

> 注意：必须将 DataKit 安装在 Kafka 所在主机才能采集 Kafka 日志

切割日志示例：

``` log
[2020-07-07 15:04:29,333] DEBUG Progress event: HTTP_REQUEST_COMPLETED_EVENT, bytes: 0 (io.confluent.connect.s3.storage.S3OutputStream:286)
```

切割后的字段列表如下：

| 字段名 | 字段值                                                 |
| ------ | ------------------------------------------------------ |
| msg    | Progress event: HTTP_REQUEST_COMPLETED_EVENT, bytes: 0 |
| name   | io.confluent.connect.s3.storage.S3OutputStream:286     |
| status | DEBUG                                                  |
| time   | 1594105469333000000                                    |

## FAQ {#faq}

<!-- markdownlint-disable MD013 -->
### :material-chat-question: 为什么看不到 `kafka_producer` / `kafka_consumer` / `kafka_connect` 指标集？ {#faq-no-data}

在开启 Kafka 服务后，如需采集 Producer/Consumer/Connector 指标，则需分别为其配置 Jolokia。

参考 [Kafka Quick Start](https://kafka.apache.org/quickstart){:target="_blank"} ，以 Producer 为例，先配置 `KAFKA_OPTS` 环境变量，示例如下：

```shell
export KAFKA_OPTS="-javaagent:/usr/local/datakit/data/jolokia-jvm-agent.jar=host=127.0.0.1,port=8090"
```

进入 Kafka 目录下启动一个 Producer：

```shell
bin/kafka-console-producer.sh --topic quickstart-events --bootstrap-server localhost:9092
```

复制出一个 kafka.conf 以开启多个 Kafka 采集器，并配置该 url：

```toml
  urls = ["http://localhost:8090/jolokia"]
```

并将采集 Producer 指标部分的字段去掉注释：

```toml
  # The following metrics are available on producer instances.  
  [[inputs.{{.InputName}}.metric]]
    name       = "kafka_producer"
    mbean      = "kafka.producer:type=*,client-id=*"
    tag_keys   = ["client-id", "type"]
```

重启 Datakit，这时 Datakit 便可采集到 Producer 实例的指标。

<!-- markdownlint-enable -->