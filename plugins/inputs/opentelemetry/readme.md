# opentelemetry input

## 主要功能
- 接收从 opentelemetry 发送的 L/T/M 三种数据，目前仅支持 trace和metric
- 传输协议支持两种：grpc 和 http
- 数据编码格式仅支持 protobuf

### 参考：
- go开源地址 [opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go)
- 其他语言: [opentelemetry-java](https://github.com/open-telemetry/opentelemetry-go) 
- 官方使用手册 ：[opentelemetry-io-docs](https://opentelemetry.io/docs/)



## GRPC 协议 (目前的 otel 主流协议， 比 http 开销较小)

接收到的数据交给trace处理。

	本模块只做数据接收和组装 不做业务处理，并都是在(接收完成、返回客户端statusOK) 之后 再进行组装。
 
metric 在dk上的映射结构体为
``` go
type otelResourceMetric struct {
	Operation   string            `json:"operation"`   // metric.name
	Source      string            `json:"source"`      // inputName ： opentelemetry
	Attributes  map[string]string `json:"attributes"`  // tags
	Resource    string            `json:"resource"`    // global.Meter name
	Description string            `json:"description"` // metric.Description
	StartTime   uint64            `json:"start_time"`  // start time
	UnitTime    uint64            `json:"unit_time"`   // end time

	ValueType string      `json:"value_type"` // double | int | histogram | ExponentialHistogram | summary
	Value     interface{} `json:"value"`      // 5种类型 对应的值：int | float

	Content string `json:"content"` //

	// Exemplar 可获取 spanid 等
}
```



### HTTP 协议
