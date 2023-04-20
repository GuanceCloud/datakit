# kafka mq

目前仅支持 `SkyWalking` `jaeger` `自定义数据`

kafkamq 理论上不做数据的二次处理，交给指定的采集器处理，如果是自定义的数据结构也交给 PL 处理。定制化的东西不应该出现在这里，这样会增加维护成本。

## SkyWalking
订阅 kafka 的消息并消费发送到观测云。 其中 topic 是固定的。

- skywalking-metrics
- skywalking-profilings 
- skywalking-segments
- skywalking-managements
- skywalking-meters
- skywalking-logging

如果用户配置了 namespace， 那么 topic 就是 `$namespace-topic`

其中 `segments` `metrics` `logging` 会发送到观测云。剩下的目前仅接收 不做处理。

数据结构也是固定的，数据结构 go 代码 [github](https://github.com/apache/skywalking-data-collect-protocol)  本地位置 inputs/skywalking/compiled/

## Jaeger 
订阅 Jaeger Span 的消息，消费并发送到 Jaeger 采集器处理，其中主题是可变的但数据格式不是可变的，是有固定格式 这点很重要！ 否则无法接入。


## 自定义数据
也叫自定义消息，比如从 kafka 中订阅日志信息通过 PL 处理发送到观测上。


## Kafka 策略

### 分配策略参考：

- [csnd-中文详细说明](https://blog.csdn.net/u010022158/article/details/106271208)
- [kafka-说明-英文](https://kafka.apache.org/10/javadoc/org/apache/kafka/clients/consumer/StickyAssignor.html)

### 安全策略
kafka 开启 SASL 之后，注意 user 和 pw 还有协议。

### 限速和采样
从 v1.6.0 开始，全部支持限速和采样，并将自定义中的移除。


## todo
- 集成测试，文档补全（kafka容易配置错的地方） en
- 剩下的消息 `meters` `managements` `profilings`
- 对接其他 agent
- 测试用例
- ...