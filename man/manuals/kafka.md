{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

采集 Kafka 指标和日志上报到观测云，帮助你监控分析 Kafka 各种异常情况

## 前置条件

安装或下载 [Jolokia](https://search.maven.org/remotecontent?filepath=org/jolokia/jolokia-jvm/1.6.2/jolokia-jvm-1.6.2-agent.jar)。DataKit 安装目录下的 `data` 目录中已经有下载好的 Jolokia jar 包。 

Jolokia 是作为 Kafka 的 java agent，基于 HTTP 协议提供了一个使用 json 作为数据格式的外部接口，提供给 DataKit 使用。 Kafka 启动时，先配置 `KAFKA_OPTS` 环境变量：(port 可根据实际情况修改成可用端口）

```shell
export KAFKA_OPTS="$KAFKA_OPTS -javaagent:/usr/local/datakit/data/jolokia-jvm-agent.jar=host=*,port=8080"
```

另外，也可以单独启动 Jolokia，将其指向 Kafka 进程 PID：

```shell
java -jar </path/to/jolokia-jvm-agent.jar> --host 127.0.0.1 --port=8080 start <Kafka-PID>
```

在开启 Kafka 服务后，如需采集 Producer/Consumer/Connector 指标，则需分别为其配置 Jolokia。

参考 [KAFKA QUICKSTART](https://kafka.apache.org/quickstart) ，以 Producer 为例，先配置 `KAFKA_OPTS` 环境变量，示例如下：

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

## 指标集

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
