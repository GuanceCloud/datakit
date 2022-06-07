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
 
trace 的具体数据结构 查看 [json后数据结构](mate)
 
接收的数据结构为 ：*collectortracepb.ExportTraceServiceRequest ,其中 GetResourceSpans() 可获取 ResourceSpans 其结构类型如下

- ResourceSpans
    - Resource                      // 进程初始化 otel 时，该进程的 otel 单例对象，可注入一下标签
    - InstrumentationLibrarySpans   // library 数组：进程中每一个服务模块对应一个 library 对象。
        - InstrumentationLibrary    // 单个 library 对象。同时也是一条 trace 对象
        - Spans                     // span 数组，单个trace 可产生多个 span 对象
            - span
                - TraceId           // span 中的各种信息
                - SpanId
                - TraceState
                - adn others ...
            - span
        - SchemaUrl
    - SchemaUrl
- ResourceSpans
    - Resource
    - InstrumentationLibrarySpans
    - SchemaUrl    

由以上数据结构可知：
1. client 端发送的数据是单进程也是单 export 中的所有服务在单位时间产生的 span 对象，并按照 trace 进行分层发送出来。
1. 接收数据端应该按照 library 对应一个 dktrace。

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


```
20220530 变更如下
``` text
一级 - 固定 service_name:"otel_service"
二级  - 子类：指标A（metricName）
            tag：{tagA:A,tagB:b} // 观测云上筛选使用
            field : {metricName : metricVal} // 指标的时序图展示使用
        子类：指标B
            ...
```
        
---
 
### 如何测试
1. 单元测试中有 mock 数据可以进行测试。
1. 源码测试可使用 [github.example](https://github.com/open-telemetry/opentelemetry-go/blob/main/example/otel-collector/main.go)
1. 最佳实践 [语雀文档](todo)