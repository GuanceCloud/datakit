{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`


# {{.InputName}}

RabbitMQ 采集器是通过插件 `rabbitmq-management` 采集数据监控 RabbitMQ ,它能够：

- RabbitMQ overview 总览，比如连接数、队列数、消息总数等
- 跟踪 RabbitMQ queue 信息，比如队列大小，消费者计数等
- 跟踪 RabbitMQ node 信息，比如使用的 `socket` `mem` 等
- 跟踪 RabbitMQ exchange 信息 ，比如 `message_publish_count` 等


## 前置条件

- RabbitMQ 版本 >= 3.8.14

- 安装 `rabbitmq` 以 `Ubuntu` 为例

    ```shell
    sudo apt-get update
    sudo apt-get install rabbitmq-server
    sudo service rabbitmq-server start
    ```
      
- 开启 `REST API plug-ins` 
    
    ```shell
    sudo rabbitmq-plugins enable rabbitmq-management
    ```
      
- 创建 user，比如：
    
    ```shell
    sudo rabbitmqctl add_user dataflux <SECRET>
    sudo rabbitmqctl set_permissions  -p / dataflux "^aliveness-test$" "^amq\.default$" ".*"
    sudo rabbitmqctl set_user_tags dataflux monitoring
    ```

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 指标集

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

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

如需采集 RabbitMQ 的日志，可在 {{.InputName}}.conf 中 将 `files` 打开，并写入 RabbitMQ 日志文件的绝对路径。比如：

```toml
    [[inputs.rabbitmq]]
      ...
      [inputs.rabbitmq.log]
        files = ["/var/log/rabbitmq/rabbit@your-hostname.log"]
```

  
开启日志采集以后，默认会产生日志来源（`source`）为 `rabbitmq` 的日志。

>注意：必须将 DataKit 安装在 RabbitMQ 所在主机才能采集 RabbitMQ 日志

## 日志 pipeline 功能切割字段说明

原始日志为 `2021-05-26 14:20:06.105 [warning] <0.12897.46> rabbitmqctl node_health_check and its HTTP API counterpart are DEPRECATED. See https://www.rabbitmq.com/monitoring.html#health-checks for replacement options.`

切割后的字段列表如下：

| 字段名  |  字段值  | 说明 |
| ---    | ---     | --- |
| status    | warning     | 日志等级 |
| msg    | <0.12897.46>...replacement options     | 日志等级 |
|  time   | 1622010006000000000     | 纳秒时间戳（作为行协议时间）|
