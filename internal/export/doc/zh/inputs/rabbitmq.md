---
title     : 'RabbitMQ'
summary   : '采集 RabbitMQ 的指标数据'
tags:
  - '消息队列'
  - '中间件'
__int_icon      : 'icon/rabbitmq'
dashboard :
  - desc  : 'RabbitMQ'
    path  : 'dashboard/zh/rabbitmq'
monitor   :
  - desc  : 'RabbitMQ'
    path  : 'monitor/zh/rabbitmq'
---

{{.AvailableArchs}}

---

RabbitMQ 采集器是通过插件 `rabbitmq-management` 采集数据监控 RabbitMQ，它能够：

- RabbitMQ overview 总览，比如连接数、队列数、消息总数等
- 跟踪 RabbitMQ queue 信息，比如队列大小，消费者计数等
- 跟踪 RabbitMQ node 信息，比如使用的 `socket` `mem` 等
- 跟踪 RabbitMQ exchange 信息 ，比如 `message_publish_count` 等

## 配置 {#config}

### 前置条件 {#reqirement}

- RabbitMQ 版本 >= `3.6.0`; 已测试的版本：
    - [x] 3.11.x
    - [x] 3.10.x
    - [x] 3.9.x
    - [x] 3.8.x
    - [x] 3.7.x
    - [x] 3.6.x

- 安装 `rabbitmq`，以 `Ubuntu` 为例

    ```shell
    sudo apt-get update
    sudo apt-get install rabbitmq-server
    sudo service rabbitmq-server start
    ```

- 开启 `REST API plug-ins`

    ```shell
    sudo rabbitmq-plugins enable rabbitmq_management
    ```

- 创建 user，比如：

    ```shell
    sudo rabbitmqctl add_user guance <SECRET>
    sudo rabbitmqctl set_permissions  -p / guance "^aliveness-test$" "^amq\.default$" ".*"
    sudo rabbitmqctl set_user_tags guance monitoring
    ```

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 *conf.d/{{.Catalog}}* 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

## 指标 {#metric}

以下所有数据采集，默认会追加全局选举 tag，也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
{{ end }}

## 自定义对象 {#object}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "custom_object"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## 日志 {#logging}

<!-- markdownlint-disable MD046 -->
???+ note

    必须将 DataKit 安装在 RabbitMQ 所在主机才能采集 RabbitMQ 日志
<!-- markdownlint-enable -->

如需采集 RabbitMQ 的日志，可在 *{{.InputName}}.conf* 中 将 `files` 打开，并写入 RabbitMQ 日志文件的绝对路径。比如：

```toml
[[inputs.rabbitmq]]
  ...
  [inputs.rabbitmq.log]
    files = ["/var/log/rabbitmq/rabbit@your-hostname.log"]
```

开启日志采集以后，默认会产生日志来源（`source`）为 `rabbitmq` 的日志。

### 日志 Pipeline 功能切割字段说明 {#pipeline}

- RabbitMQ 通用日志切割

通用日志文本示例：

``` log
2021-05-26 14:20:06.105 [warning] <0.12897.46> rabbitmqctl node_health_check and its HTTP API counterpart are DEPRECATED. See https://www.rabbitmq.com/monitoring.html#health-checks for replacement options.
```

切割后的字段列表如下：

| 字段名 | 字段值                             | 说明                         |
| ---    | ---                                | ---                          |
| status | warning                            | 日志等级                     |
| msg    | <0.12897.46>...replacement options | 日志等级                     |
| time   | 1622010006000000000                | 纳秒时间戳（作为行协议时间） |
