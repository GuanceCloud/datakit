# opentelemetry input 开发文档

## 主要功能
- 接收从 opentelemetry 发送的 L/T/M 三种数据，目前仅支持 trace和metric
- 传输协议支持两种：grpc 和 http  (目前otel的主流协议)
- 数据编码格式仅支持 protobuf 
- 测试

### 参考：
- go开源地址 [opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go)
- 官方使用手册 ：[opentelemetry-io-docs](https://opentelemetry.io/docs/)


### trace 流程
接收到的数据交给trace处理。

本模块只做数据接收和组装 不做业务处理，并都是在(接收完成、返回客户端statusOK) 之后 再进行组装。
 
trace 的具体数据结构 查看 [json后数据结构](mate.md)
 
---
 
### Metric
代码及注释中会用到很多 OTEL 专业的名称，这里做统一解释
- instrumentation_library.name: 指标集名称
- Provider : Metric 处理器,一个进城一般只有一个处理器
- attributes:标签
- Exporter : 将产生的 metric 发送出去

重新梳理 Metric 数据结构和观测云上的展示后。决定重构 OtelResourceMetric 对象。 20220303

分层结构如下：
``` text
resource (server.name or inputName)
    指标集名称（libraryMetric.InstrumentationLibrary.Name）
        子类：指标A（metricName）
            tag：{tagA:A,tagB:b} // 观测云上筛选使用
            field : {metricName : metricVal} // 指标的时序图展示使用
        子类：指标B
            ...

观测云展示示例
指标集
resource_指标集名称  ------>  指标A
                             指标B
```

        
---
 
### 如何测试
1. 单元测试中有 mock 数据可以进行测试。
1. 源码测试可使用 [github.example](https://github.com/open-telemetry/opentelemetry-go/blob/main/example/otel-collector/main.go)
1. 最佳实践 [语雀文档](todo)