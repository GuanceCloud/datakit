
# 订阅 Kafka 中的数据
---

{{.AvailableArchs}}

***作者：宋龙奇***

---

Datakit 支持从 kafka 中订阅消息采集链路、指标和日志信息。目前仅支持 `SkyWalking` 、`Jaeger` 以及自定义 topic.

## 如何配置 {#config}

配置文件示例：

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

---

> 注意：自 v1.6.0 开始全部支持采样和速率限制，之前的版本只有自定义消息支持。

配置文件注意的地方：

1. `kafka_version`: 长度为 3，例如：`1.0.0`，`1.2.1` 等等
1. `offsets`: 注意是 `Newest` 还是 `Oldest`
1. `SASL` : 如果开启了安全认证，请正确配置用户和密码，如果 Kafka 监听地址是域名形式，请在 `/etc/hosts` 添加映射 IP

## SkyWalking {#kafkamq-skywalking}

kafka 插件默认会将 `traces/JVM metrics/logging/Instance Properties/profiled snapshots` 发送到 Kafka 集群中。

该功能默认是关闭的。需要将 *kafka-reporter-plugin-x.y.z.jar* 从 *agent/optional-reporter-plugins* 放到 *agent/plugins* 目录才会生效。

配置文件及说明：

```toml
# skywalking custom
[inputs.kafkamq.skywalking]

  # !!!Required: send to datakit skywalking input.
  dk_endpoint="http://localhost:9529"
  
  topics = [
    "skywalking-metrics",
    "skywalking-profilings",
    "skywalking-segments",
    "skywalking-managements",
    "skywalking-meters",
    "skywalking-logging",
  ]
  namespace = ""
```

将注释打开即可开启订阅，订阅的主题在 SkyWalking agent 配置文件 *config/agent.config* 中。

> 注意：该采集器只是将订阅的数据转发到 Datakit SkyWalking 采集器中，请打开 [SkyWalking](skywalking.md) 采集器，并将 `dk_endpoint` 注释打开。

## Jaeger {#jaeger}

配置文件：

```toml
# Jaeger from kafka. !!!Note: Make sure Datakit Jaeger collector is open.
[inputs.kafkamq.jaeger]
    ## !!!Required: ipv6 is "[::1]:9529"
    dk_endpoint="http://localhost:9529"
    
    ## !!!Required: topics 
    topics=["jaeger-spans","jaeger-my-spans"]
```

> 注意： 该采集器只是将订阅的数据转发到 Datakit Jaeger 采集器中，请打开 [jaeger](jaeger.md) 采集器，并将 `dk_endpoint` 注释打开。

## 自定义 Topic {#kafka-custom}

有些时候用户使用的并不是市面上常用的工具，有些的三方库并不是开源的，数据结构也不是公开的。这样就需要根据采集到的数据结构手动进行处理，这时候就体现到 Pipeline 的强大之处，用户可通过自定义配置进行订阅并消费消息。

更多的情况往往是现有的系统已经将数据发送到 kafka，而随着开发运维人员迭代，进行修改输出变的复杂难以实现，这时候使用自定义模式便是很好的方式。

配置文件：

```toml
# user custom message with PL script.
[inputs.kafkamq.custom]
# spilt_json_body = true
[inputs.kafkamq.custom.log_topic_map]
  "log_topic"="log.p"
  "log"="rum_apm.p"

[inputs.kafkamq.custom.metric_topic_map]
  "metric_topic"="rum_apm.p"
  
[inputs.kafkamq.custom.rum_topic_map]
  "rum"="rum.p"

```

> 注意：metric 的 Pipeline 脚本应该放在 *pipeline/metric/* 目录下，RUM 的 Pipeline 脚本应该放到 *pipeline/rum/* 目录下。

理论上每一个消息体应该是一条日志或者一个指标，如果您的消息是多条日志，可以使用 `spilt_json_body` 开启 JSON 切割数组功能，当数据是 JSON 数组，配合 PL 可以将数组切割成单个日志或者指标数据。

### 示例 {#example}

以一个简单的 metric 为例，介绍如何使用自定义配置订阅消息。

当不知道发送到 Kafka 上的数据结构时什么格式时。可以先将 Datakit 的日志级别改为 Debug。将订阅打开，在 Datakit 日志中会有输出。假设拿到的如下数据：

```shell
# 打开 debug 日志级别之后，查看日志，Datakit 会将消息信息打印出来。
tailf /var/log/datakit/log | grep "kafka_message"
```

假设拿到的这是一个 metric 的 JSON 格式纯文本字符串：

```json
{"time": 1666492218, "dimensions": {"bk_biz_id": 225,"ip": "10.200.64.45" },  "metrics": { "cpu_usage_pct": 0.01}, "exemplar": null}
```

有了数据格式，就可以手写 Pipeline 脚本。登录「观测云 -> 管理 -> 文本处理（Pipeline）编写脚本」。 如：

```python
data = load_json(message)
drop_origin_data()

hostip = data["dimensions"]["ip"]
bkzid = data["bk_biz_id"]
cast(bkzid,"sttr")

set_tag(hostip,hostip)
set_tag(bk_biz_id,bkzid)

add_key(cpu_usage_pct,data["metrics"]["cpu_usage_pct"])

# 注意 此处为行协议缺省值，Pipeline 脚本通过之后 这个 message_len 就可以删掉了。
drop_key(message_len)
```

将文件放到 */usr/local/datakit/pipeline/metric/* 目录下。

> 注意：指标数据的 Pipeline 脚本放到 *metric/* 下，logging 数据的 Pipeline 脚本放到 *pipeline/* 目录下。

配置好 Pipeline 脚本，重启 Datakit 即可。

### 基准测试 {#benchmark}

消息的消费能力受限于网络和带宽的限制，所以基准测试只是测试 Datakit 的消费能力而不是 IO 能力。本次测试的机器配置是 4 核 8 线程、16G 内存。测试过程中 CPU 峰值 60%~70%，内存增加 10%。

| 消息数量 | 用时    | 每秒消费能力（条） |
| -------  | ------- | -----------        |
| 100k     | 5s~7s   | 16k                |
| 1000k    | 1m30s   | 11k                |

另外减少日志输出、关闭 cgroup 限制、增加内网和公网带宽等，可以增加消费能力。

### 多台 Datakit 负载均衡 {#datakit-assignor}

当消息量很大，一台 Datakit 消费能力不足时可以增加多台 Datakit 进行消费，这里有三点需要注意：

1. 确保 Topic 分区不是一个（至少 2 个），这个可以通过工具 [`kafka-map`](https://github.com/dushixiang/kafka-map/releases){:target="_blank"}查看
1. 确保 KafkaMQ 采集器的配置是 `assignor = "roundrobin"`（负载均衡策略的一种），`group_id="datakit"`（组名称必须一致，否则会重复消费）
1. 确保消息的生产者将消息发送多分区，语言不同方法不同 这里不列出代码了，自行查找相关实现

### 问题排查 {#some-problems}

当写好 Pipeline 脚本之后不确定是否能切割正确，可以使用测试命令：

```shell
datakit pipeline -P metric.p -T '{"time": 1666492218,"dimensions":{"bk_biz_id": 225,"ip": "172.253.64.45"},"metrics": {"cpu_usage_pct": 0.01}, "exemplar": null}'
```

切割正确之后，可以查看行协议数据是否正确，暂时将 output_file 设置为本地文件：

```shell
vim conf/datakit.conf

# 设置为本地文件，就不会输出到 io，测试结束之后赋值为空即可。
output_file = "/usr/local/datakit/out.pts"
# 查看文件 out.pts 是否正确
```

连接失败可能是版本问题，请在配置文件中正确填写 kafka 版本。目前支持的版本列表：[0.8.2] - [3.3.1]

其他问题： 通过 `datakit monitor` 命令查看，或者 `datakit monitor -V` 查看。
