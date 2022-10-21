# kafka mq

目前仅支持 `SkyWalking`

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

## 分配策略
参考：
- [csnd-中文详细说明](https://blog.csdn.net/u010022158/article/details/106271208)
- [kafka-说明-英文](https://kafka.apache.org/10/javadoc/org/apache/kafka/clients/consumer/StickyAssignor.html)


## todo
- 剩下的消息 `meters` `managements` `profilings`
- 优化 wpool
- 对接其他 agent
- 测试用例
- ...