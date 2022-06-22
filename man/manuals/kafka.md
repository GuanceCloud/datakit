{{.CSS}}
# Kafka
---

- DataKit 版本：{{.Version}}
- 操作系统支持：`{{.AvailableArchs}}`

采集 Kafka 指标和日志上报到观测云，帮助你监控分析 Kafka 各种异常情况

## 前置条件

安装或下载 [Jolokia](https://search.maven.org/remotecontent?filepath=org/jolokia/jolokia-jvm/1.6.2/jolokia-jvm-1.6.2-agent.jar){:target="_blank"}。DataKit 安装目录下的 `data` 目录中已经有下载好的 Jolokia jar 包。 

Jolokia 是作为 Kafka 的 java agent，基于 HTTP 协议提供了一个使用 json 作为数据格式的外部接口，提供给 DataKit 使用。 Kafka 启动时，先配置 `KAFKA_OPTS` 环境变量：(port 可根据实际情况修改成可用端口）

```shell
export KAFKA_OPTS="$KAFKA_OPTS -javaagent:/usr/local/datakit/data/jolokia-jvm-agent.jar=host=*,port=8080"
```

另外，也可以单独启动 Jolokia，将其指向 Kafka 进程 PID：

```shell
java -jar </path/to/jolokia-jvm-agent.jar> --host 127.0.0.1 --port=8080 start <Kafka-PID>
```

在开启 Kafka 服务后，如需采集 Producer/Consumer/Connector 指标，则需分别为其配置 Jolokia。

参考 [KAFKA QUICKSTART](https://kafka.apache.org/quickstart){:target="_blank"} ，以 Producer 为例，先配置 `KAFKA_OPTS` 环境变量，示例如下：

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
  [[inputs.kafka.metric]]
    name       = "kafka_producer"
    mbean      = "kafka.producer:type=*,client-id=*"
    tag_keys   = ["client-id", "type"]
```

重启 Datakit，这时 Datakit 便可采集到 Producer 实例的指标。

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

## 日志采集

如需采集 Kafka 的日志，可在 {{.InputName}}.conf 中 将 `files` 打开，并写入 kafka 日志文件的绝对路径。比如：

```toml
    [[inputs.kafka]]
      ...
      [inputs.kafka.log]
		files = ["/usr/local/var/log/kafka/error.log","/usr/local/var/log/kafka/kafka.log"]
```


开启日志采集以后，默认会产生日志来源（`source`）为 `kafka` 的日志。

>注意：必须将 DataKit 安装在 Kafka 所在主机才能采集 Kafka 日志

切割日志示例：

```
[2020-07-07 15:04:29,333] DEBUG Progress event: HTTP_REQUEST_COMPLETED_EVENT, bytes: 0 (io.confluent.connect.s3.storage.S3OutputStream:286)
```

切割后的字段列表如下：



| 字段名 | 字段值                                                 |
| ------ | ------------------------------------------------------ |
| msg    | Progress event: HTTP_REQUEST_COMPLETED_EVENT, bytes: 0 |
| name   | io.confluent.connect.s3.storage.S3OutputStream:286     |
| status | DEBUG                                                  |
| time   | 1594105469333000000                                    |

## 视图预览

### 场景视图

Kafka 观测场景主要展示了 Kafka的 基础信息，topic 信息和性能信息。

![image](imgs/input-kafka-1.png)

## 安装部署

说明：示例 Kafka 版本为：Kafka 2.11 (CentOS)，各个不同版本指标可能存在差异

### 前置条件

安装或下载 [Jolokia](https://search.maven.org/remotecontent?filepath=org/jolokia/jolokia-jvm/1.6.2/jolokia-jvm-1.6.2-agent.jar)。DataKit 安装目录下的 `data` 目录中已经有下载好的 Jolokia jar 包。

### 配置实施

#### 指标采集 (必选)


1、开启 Datakit Kafka 插件，复制 sample 文件
```bash
/usr/local/datakit/conf.d/kafka
cp kafka.conf.sample kafka.conf
```

2、修改 `kafka.conf` 配置文件
```bash
vi kafka.conf
```
参数说明

- default_tag_prefix：设置默认tag前缀(默认为空)
- default_field_prefix：设置默认field前缀(默认为空)
- default_field_separator：设置默认field分割(默认为".")
- username：要采集的 kafka 的用户名
- password：要采集的 kafka 的密码
- response_timeout：超时时间
- interval：采集指标频率
- urls：jolokia的地址

```yaml
[[inputs.kafka]]
# default_tag_prefix      = ""
# default_field_prefix    = ""
# default_field_separator = "."

# username = ""
# password = ""
# response_timeout = "5s"

## Optional TLS config
# tls_ca   = "/var/private/ca.pem"
# tls_cert = "/var/private/client.pem"
# tls_key  = "/var/private/client-key.pem"
# insecure_skip_verify = false

## Monitor Intreval
# interval   = "60s"

# Add agents URLs to query
urls = ["http://localhost:12346/jolokia/"]

## Add metrics to read
[[inputs.kafka.metric]]
  name         = "kafka_controller"
  mbean        = "kafka.controller:name=*,type=*"
  field_prefix = "#1."

[[inputs.kafka.metric]]
  name         = "kafka_replica_manager"
  mbean        = "kafka.server:name=*,type=ReplicaManager"
  field_prefix = "#1."

[[inputs.kafka.metric]]
  name         = "kafka_purgatory"
  mbean        = "kafka.server:delayedOperation=*,name=*,type=DelayedOperationPurgatory"
  field_prefix = "#1."
  field_name   = "#2"

[[inputs.kafka.metric]]
  name     = "kafka_client"
  mbean    = "kafka.server:client-id=*,type=*"
  tag_keys = ["client-id", "type"]

[[inputs.kafka.metric]]
  name         = "kafka_request"
  mbean        = "kafka.network:name=*,request=*,type=RequestMetrics"
  field_prefix = "#1."
  tag_keys     = ["request"]

[[inputs.kafka.metric]]
  name         = "kafka_topics"
  mbean        = "kafka.server:name=*,type=BrokerTopicMetrics"
  field_prefix = "#1."

[[inputs.kafka.metric]]
  name         = "kafka_topic"
  mbean        = "kafka.server:name=*,topic=*,type=BrokerTopicMetrics"
  field_prefix = "#1."
  tag_keys     = ["topic"]

[[inputs.kafka.metric]]
  name       = "kafka_partition"
  mbean      = "kafka.log:name=*,partition=*,topic=*,type=Log"
  field_name = "#1"
  tag_keys   = ["topic", "partition"]

[[inputs.kafka.metric]]
  name       = "kafka_partition"
  mbean      = "kafka.cluster:name=UnderReplicated,partition=*,topic=*,type=Partition"
  field_name = "UnderReplicatedPartitions"
  tag_keys   = ["topic", "partition"]
```

3、重启 Datakit (如果需要开启日志，请配置日志采集再重启)

```bash
systemctl restart datakit
```

4、Kafka 指标采集验证 `/usr/local/datakit/datakit -M |egrep "最近采集|kafka"`

![image](imgs/input-kafka-2.png)

5、指标预览

![image](imgs/input-kafka-3.png)

#### 日志采集 (非必选)

1、修改 `kafka.conf` 配置文件

参数说明

- files：日志文件路径 (通常填写访问日志和错误日志)
- pipeline：日志切割文件(内置)，实际文件路径 /usr/local/datakit/pipeline/kafka.p
- 相关文档 <[DataFlux pipeline 文本数据处理](/datakit/pipeline.md)>

```
[inputs.kafka.log]
  files = ["/usr/local/kafka/logs/server.log",
    "/usr/local/kafka/logs/controller.log"
  ]
```

3、重启 Datakit (如果需要开启自定义标签，请配置插件标签再重启)

```
systemctl restart datakit
```

4、Kafka 日志采集验证  /usr/local/datakit/datakit -M |egrep "最近采集|kafka_log"

![image](imgs/input-kafka-4.png)

5、日志预览

![image](imgs/input-kafka-5.png)

#### 插件标签 (非必选)
参数说明

- 该配置为自定义标签，可以填写任意 key-value 值
- 以下示例配置完成后，所有 Kafka 指标都会带有 service = "kafka" 的标签，可以进行快速查询
- 相关文档 <[DataFlux Tag 应用最佳实践](../best-practices/guance-skill/tag.md)>

```
# 示例
[inputs.kafka.tags]
		service = "kafka"
    # some_tag = "some_value"
    # more_tag = "some_other_value"
```

重启 Datakit

```
systemctl restart datakit
```

## 场景视图

场景 - 新建场景 - kafka 监控场景 

## 异常检测

异常检测库 - 新建检测库 - kafka 检测库 

| 序号 | 规则名称 | 触发条件 | 级别 | 检测频率 |
| --- | --- | --- | --- | --- |

## 指标详解

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## 最佳实践

[<Kafka可观测最佳实践>](../best-practices/integrations/kafka.md)

## 故障排查

<[无数据上报排查](why-no-data.md)>

