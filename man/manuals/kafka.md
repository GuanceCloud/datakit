{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：{{.AvailableArchs}}

# 简介

采集 kafka 指标和日志上报到 DataFlux 中，帮助你监控分析 Kafka 各种异常情况

## 前置条件

安装或下载 [Jolokia](https://search.maven.org/remotecontent?filepath=org/jolokia/jolokia-jvm/1.6.2/jolokia-jvm-1.6.2-agent.jar)。DataKit 安装目录下的 `data` 目录中已经有下载好的 Jolokia jar 包。 


## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```python
{{.InputSample}}
```

## 指标集

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}


## 日志采集

如需采集 kafka 的日志，可在 {{.InputName}}.conf 中 将 `files` 打开，并写入 Kafka 日志文件的绝对路径。比如：

```python
    [[inputs.kafka]]
      ...
      [inputs.kafka.log]
		files = ["/usr/local/var/log/kafka/error.log","/usr/local/var/log/kafka/kafka.log"]
```


开启日志采集以后，默认会产生日志来源（`source`）为 `kafka` 的日志。

**注意**

- 日志采集仅支持采集已安装 DataKit 主机上的日志
