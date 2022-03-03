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
 
### 如何测试
1. 单元测试中有 mock 数据可以进行测试。
1. 源码测试可使用 [github.example](https://github.com/open-telemetry/opentelemetry-go/blob/main/example/otel-collector/main.go)
1. 最佳实践 [语雀文档](todo)