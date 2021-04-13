{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：{{.AvailableArchs}}


# 简介

RabbitMQ 采集器是通过插件 `rabbitmq-management` 采集数据监控 RabbitMQ ,它能够：

- RabbitMQ overview 总览，比如连接数、队列数、消息总数等
- 跟踪 RabbitMQ queue 信息，比如队列大小，消费者计数等
- 跟踪 RabbitMQ node 信息，比如使用的 `socket` `mem` 等
- 跟踪 RabbitMQ exchange 信息 ，比如 `message_publish_count` 等


## 前置条件

- 安装 `rabbitmq` 以 `Ubuntu` 为例

    ```shell script
    sudo apt-get update
    sudo apt-get install rabbitmq-server
    sudo service rabbitmq-server start
    ```
      
- 开启 `REST API plug-ins` 

    ```shell script
    sudo rabbitmq-plugins enable rabbitmq-management
    ```
      
- 创建 user，比如：

    ```shell script
    rabbitmqctl add_user dataflux <SECRET>
    rabbitmqctl set_permissions  -p / dataflux "^aliveness-test$" "^amq\.default$" ".*"
    rabbitmqctl set_user_tags dataflux monitoring
    ```
  
- 如需采集 RabbitMQ 的日志，可在 rabbitmq.conf 中 将 `files` 打开并写入 RabbitMQ 日志文件的绝对路径。目前仅支持在 DataKit 安装主机上面的日志采集 

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 指标集

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
