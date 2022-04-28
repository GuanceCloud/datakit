{{.CSS}}

- DataKit 版本: {{.Version}}
- 文档发布日期: {{.ReleaseDate}}
- 操作系统支持: 全平台

# {{.InputName}}

本文档主要介绍 [Elastic Beats](https://www.elastic.co/products/beats/) 接收器。目前支持:
- [Filebeat](https://www.elastic.co/products/beats/filebeat/)

## 接收 Filebeat 采集的数据

### 配置 DataKit 接收

进入 DataKit 安装目录下的 `conf.d/beats_output/` 目录, 复制 `beats_output.conf.sample` 并命名为 `beats_output.conf`。示例如下: 

```toml
[[inputs.beats_output]]
  # listen address, with protocol scheme and port
  listen = "tcp://0.0.0.0:5044"

  ## source, if it's empty, use 'default'
  source = ""

  ## add service tag, if it's empty, use $source.
  service = ""

  ## grok pipeline script name
  pipeline = ""

  ## datakit read text from Files or Socket , default max_textline is 256k
  ## If your log text line exceeds 256Kb, please configure the length of your text,
  ## but the maximum length cannot exceed 256Mb
  maximum_length = 262144

  [inputs.beats_output.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
```

### 配置 Filebeat

将 Filebeat 目录下的 `filebeat.yml` 配置如下。

- `filebeat.inputs`:

```yml
filebeat.inputs:

# Each - is an input. Most options can be set at the input level, so
# you can use different inputs for various configurations.
# Below are the input specific configurations.

# filestream is an input for collecting log messages from files.
- type: filestream

  # Change to true to enable this input configuration.
  enabled: true

  # Paths that should be crawled and fetched. Glob based paths.
  paths:
    - /Users/mac/Downloads/tmp/1.log
```

- `output.logstash`:

```yml
output.logstash:
  # The Logstash hosts
  hosts: ["<Datakit-IP>:5044"]
```

这里的 `5044` 端口要与 `<Datakit 安装目录>/conf.d/beats_output/beats_output.conf` 中配置的 `listen` 端口一致。

这样就实现 Filebeat 采集日志文件 `/Users/mac/Downloads/tmp/1.log` 到 Datakit 了。

## 指标集

以下所有数据采集, 默认会追加名为 `host.name`(值为 Filebeat 所在主机名) 和 `log.file.path`(值为 Filebate 采集文件的全路径) 的全局 tag, 也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签: 

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }} 

## 其它

本接收器与日志采集器很相似，Pipeline 语法方面可参考[日志采集器](logging)。
