{{.CSS}}
# RabbitMQ
---

- DataKit 版本：{{.Version}}
- 操作系统支持：`{{.AvailableArchs}}`

RabbitMQ 采集器是通过插件 `rabbitmq-management` 采集数据监控 RabbitMQ ,它能够：

- RabbitMQ overview 总览，比如连接数、队列数、消息总数等
- 跟踪 RabbitMQ queue 信息，比如队列大小，消费者计数等
- 跟踪 RabbitMQ node 信息，比如使用的 `socket` `mem` 等
- 跟踪 RabbitMQ exchange 信息 ，比如 `message_publish_count` 等

## 视图预览
RabbitMQ 性能指标展示：包括连接数量、通道数量、队列量、消费者数、队列消息速率、队列消息数、交换机、节点、队列等。

![1631256451(1).png](../imgs/rabbitmq-1.png)

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
    sudo rabbitmqctl add_user guance <SECRET>
    sudo rabbitmqctl set_permissions  -p / guance "^aliveness-test$" "^amq\.default$" ".*"
    sudo rabbitmqctl set_user_tags guance monitoring
    ```

### 指标采集

#### 配置



1. 开启 rabbitmq_management 插件

RabbitMQ 采集器是通过插件 `rabbitmq_management` 采集数据监控 RabbitMQ ,它能够：

- RabbitMQ overview 总览，比如连接数、队列数、消息总数等
- 跟踪 RabbitMQ queue 信息，比如队列大小，消费者计数等
- 跟踪 RabbitMQ node 信息，比如使用的 `socket` `mem` 等
- 跟踪 RabbitMQ exchange 信息 ，比如 `message_publish_count` 等

      登录安装rabbitmq的服务器，执行如下命令：
```
rabbitmq-plugins enable rabbitmq_management
```
    

2. rabbitmq新增dataflux账号，并赋予monitoring角色

      登录安装rabbitmq的服务器，执行如下命令：
```
rabbitmqctl add_user dataflux Datakit1234
rabbitmqctl set_user_tags dataflux monitoring
rabbitmqctl set_permissions -p / dataflux "^aliveness-test$" "^amq\.default$" ".*"
```

3. 开启RabbitMQ插件，复制sample文件

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

4. 修改rabbitmq.conf 配置文件
```
vi rabbitmq.conf
```
参数说明

- url：rabbitmq的url
- username：步骤2中设置的用户名
- password：步骤2中设置的密码
- interval：数据采集频率
- insecure_skip_verify：是否忽略安全验证 (如果是 https，请设置为 true)
```
[[inputs.rabbitmq]]
  # rabbitmq url ,required
  url = "http://localhost:15672"

  # rabbitmq user, required
  username = "dataflux"

  # rabbitmq password, required
  password = "Datakit1234"

  # ##(optional) collection interval, default is 30s
  # interval = "30s"

  ## Optional TLS Config
  # tls_ca = "/xxx/ca.pem"
  # tls_cert = "/xxx/cert.cer"
  # tls_key = "/xxx/key.key"
  ## Use TLS but skip chain & host verification
  insecure_skip_verify = false

```

5. 重启 Datakit (如果需要开启日志，请配置日志采集再重启)
```
systemctl restart datakit
```

指标预览

![1631259475(1).png](../imgs/rabbitmq-2.png)


配置好后，重启 DataKit 即可。


### 日志采集

如需采集 RabbitMQ 的日志，可在 {{.InputName}}.conf 中 将 `files` 打开，并写入 RabbitMQ 日志文件的绝对路径。比如：

```toml
    [[inputs.rabbitmq]]
      ...
      [inputs.rabbitmq.log]
        files = ["/var/log/rabbitmq/rabbit@your-hostname.log"]
```

参数说明

- files：日志文件路径 (通常填写访问日志和错误日志)
- pipeline：日志切割文件(内置)，实际文件路径 /usr/local/datakit/pipeline/rabbitmq.p
- 相关文档 <[DataFlux pipeline 文本数据处理](/datakit/pipeline/)>

  
开启日志采集以后，默认会产生日志来源（`source`）为 `rabbitmq` 的日志。

>注意：必须将 DataKit 安装在 RabbitMQ 所在主机才能采集 RabbitMQ 日志

#### 日志 pipeline 功能切割字段说明

- RabbitMQ 通用日志切割

通用日志文本示例:

```
2021-05-26 14:20:06.105 [warning] <0.12897.46> rabbitmqctl node_health_check and its HTTP API counterpart are DEPRECATED. See https://www.rabbitmq.com/monitoring.html#health-checks for replacement options.
```

切割后的字段列表如下：

| 字段名  |  字段值  | 说明 |
| ---    | ---     | --- |
| status    | warning     | 日志等级 |
| msg    | <0.12897.46>...replacement options     | 日志等级 |
|  time   | 1622010006000000000     | 纳秒时间戳（作为行协议时间）|


## 指标详解

### `rabbitmq_overview`

- 标签
| 标签名 | 描述 |
| --- | --- |
| `cluster_name` | rabbitmq cluster name |
| `rabbitmq_version` | rabbitmq version |
| `url` | rabbitmq url |


- 指标列表
| 指标 | 描述 | 数据类型 | 单位 |
| --- | --- | --- | --- |
| `message_ack_count` | Number of messages delivered to clients and acknowledged | int | count |
| `message_ack_rate` | Rate of messages delivered to clients and acknowledged per second | float | percent |
| `message_confirm_count` | Count of messages confirmed | int | count |
| `message_confirm_rate` | Rate of messages confirmed per second | float | percent |
| `message_deliver_get_count` | Sum of messages delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get | int | count |
| `message_deliver_get_rate` | Rate per second of the sum of messages delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get | float | percent |
| `message_publish_count` | Count of messages published | int | count |
| `message_publish_in_count` | Count of messages published from channels into this overview | int | count |
| `message_publish_in_rate` | Rate of messages published from channels into this overview per sec | float | percent |
| `message_publish_out_count` | Count of messages published from this overview into queues | int | count |
| `message_publish_out_rate` | Rate of messages published from this overview into queues per second | float | percent |
| `message_publish_rate` | Rate of messages published per second | float | percent |
| `message_redeliver_count` | Count of subset of messages in deliver_get which had the redelivered flag set | int | count |
| `message_redeliver_rate` | Rate of subset of messages in deliver_get which had the redelivered flag set per second | float | percent |
| `message_return_unroutable_count` | Count of messages returned to publisher as unroutable | int | count |
| `message_return_unroutable_count_rate` | Rate of messages returned to publisher as unroutable per second | float | percent |
| `object_totals_channels` | Total number of channels | int | count |
| `object_totals_connections` | Total number of connections | int | count |
| `object_totals_consumers` | Total number of consumers | int | count |
| `object_totals_queues` | Total number of queues | int | count |
| `queue_totals_messages_count` | Total number of messages (ready plus unacknowledged) | int | count |
| `queue_totals_messages_rate` | Total rate of messages (ready plus unacknowledged) | float | percent |
| `queue_totals_messages_ready_count` | Number of messages ready for delivery | int | count |
| `queue_totals_messages_ready_rate` | Rate of number of messages ready for delivery | float | percent |
| `queue_totals_messages_unacknowledged_count` | Number of unacknowledged messages | int | count |
| `queue_totals_messages_unacknowledged_rate` | Rate of number of unacknowledged messages | float | percent |


### `rabbitmq_queue`

- 标签
| 标签名 | 描述 |
| --- | --- |
| `node_name` | rabbitmq node name |
| `queue_name` | rabbitmq queue name |
| `url` | rabbitmq url |


- 指标列表
| 指标 | 描述 | 数据类型 | 单位 |
| --- | --- | --- | --- |
| `bindings_count` | Number of bindings for a specific queue | int | count |
| `consumer_utilization` | Number of consumers | float | percent |
| `consumers` | The ratio of time that a queue's consumers can take new messages | int | count |
| `head_message_timestamp` | Timestamp of the head message of the queue. Shown as millisecond | int | msec |
| `memory` | Bytes of memory consumed by the Erlang process associated with the queue, including stack, heap and internal structures | int | B |
| `message` | Count of the total messages in the queue | int | count |
| `message_ack_count` | Number of messages in queues delivered to clients and acknowledged | int | count |
| `message_ack_rate` | Number per second of messages delivered to clients and acknowledged | float | percent |
| `message_deliver_count` | Count of messages delivered in acknowledgement mode to consumers | int | count |
| `message_deliver_get_count` | Sum of messages in queues delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get. | int | count |
| `message_deliver_get_rate` | Rate per second of the sum of messages in queues delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get. | float | percent |
| `message_deliver_rate` | Rate of messages delivered in acknowledgement mode to consumers | float | percent |
| `message_publish_count` | Count of messages in queues published | int | count |
| `message_publish_rate` | Rate per second of messages published | float | percent |
| `message_redeliver_count` | Count of subset of messages in queues in deliver_get which had the redelivered flag set | int | count |
| `message_redeliver_rate` | Rate per second of subset of messages in deliver_get which had the redelivered flag set | float | percent |
| `messages_rate` | Count per second of the total messages in the queue | float | percent |
| `messages_ready` | Number of messages ready to be delivered to clients | int | count |
| `messages_ready_rate` | Number per second of messages ready to be delivered to clients | float | percent |
| `messages_unacknowledged` | Number of messages delivered to clients but not yet acknowledged | int | count |
| `messages_unacknowledged_rate` | Number per second of messages delivered to clients but not yet acknowledged | float | percent |


### `rabbitmq_exchange`

- 标签
| 标签名 | 描述 |
| --- | --- |
| `auto_delete` | If set, the exchange is deleted when all queues have finished using it |
| `durable` | If set when creating a new exchange, the exchange will be marked as durable. Durable exchanges remain active when a server restarts. Non-durable exchanges (transient exchanges) are purged if/when a server restarts. |
| `exchange_name` | rabbitmq exchange name |
| `internal` | If set, the exchange may not be used directly by publishers, but only when bound to other exchanges. Internal exchanges are used to construct wiring that is not visible to applications |
| `type` | rabbitmq exchange type |
| `url` | rabbitmq url |
| `vhost` | rabbitmq exchange virtual hosts |


- 指标列表
| 指标 | 描述 | 数据类型 | 单位 |
| --- | --- | --- | --- |
| `message_ack_count` | Number of messages in exchanges delivered to clients and acknowledged | int | count |
| `message_ack_rate` | Rate of messages in exchanges delivered to clients and acknowledged per second | float | percent |
| `message_confirm_count` | Count of messages in exchanges confirmed | int | count |
| `message_confirm_rate` | Rate of messages in exchanges confirmed per second | float | percent |
| `message_deliver_get_count` | Sum of messages in exchanges delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get | int | count |
| `message_deliver_get_rate` | Rate per second of the sum of exchange messages delivered in acknowledgement mode to consumers, in no-acknowledgement mode to consumers, in acknowledgement mode in response to basic.get, and in no-acknowledgement mode in response to basic.get | float | percent |
| `message_publish_count` | Count of messages in exchanges published | int | count |
| `message_publish_in_count` | Count of messages published from channels into this exchange | int | count |
| `message_publish_in_rate` | Rate of messages published from channels into this exchange per sec | float | percent |
| `message_publish_out_count` | Count of messages published from this exchange into queues | int | count |
| `message_publish_out_rate` | Rate of messages published from this exchange into queues per second | float | percent |
| `message_publish_rate` | Rate of messages in exchanges published per second | float | percent |
| `message_redeliver_count` | Count of subset of messages in exchanges in deliver_get which had the redelivered flag set | int | count |
| `message_redeliver_rate` | Rate of subset of messages in exchanges in deliver_get which had the redelivered flag set per second | float | percent |
| `message_return_unroutable_count` | Count of messages in exchanges returned to publisher as unroutable | int | count |
| `message_return_unroutable_count_rate` | Rate of messages in exchanges returned to publisher as unroutable per second | float | percent |


### `rabbitmq_node`

- 标签
| 标签名 | 描述 |
| --- | --- |
| `node_name` | rabbitmq node name |
| `url` | rabbitmq url |


- 指标列表
| 指标 | 描述 | 数据类型 | 单位 |
| --- | --- | --- | --- |
| `disk_free` | Current free disk space | int | B |
| `disk_free_alarm` | Does the node have disk alarm | bool | - |
| `fd_used` | Used file descriptors | int | - |
| `io_read_avg_time` | avg wall time (milliseconds) for each disk read operation in the last statistics interval | float | ms |
| `io_seek_avg_time` | average wall time (milliseconds) for each seek operation in the last statistics interval | float | ms |
| `io_sync_avg_time` | average wall time (milliseconds) for each fsync() operation in the last statistics interval | float | ms |
| `io_write_avg_time` | avg wall time (milliseconds) for each disk write operation in the last statistics interval | float | ms |
| `mem_alarm` | Does the node have mem alarm | bool | - |
| `mem_limit` | Memory usage high watermark in bytes | int | B |
| `mem_used` | Memory used in bytes | int | B |
| `run_queue` | Average number of Erlang processes waiting to run | int | count |
| `running` | Is the node running or not | bool | - |
| `sockets_used` | Number of file descriptors used as sockets | int | count |


## 场景视图
场景 - 新建空白场景 - 系统视图 - Rabbitmq监控视图

## 故障排查
- [无数据上报排查](why-no-data.md)
