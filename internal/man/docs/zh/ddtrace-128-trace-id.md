# 支持 128 位的链路 ID

[:octicons-tag-24: Datakit-1.8.0](changelog.md#cl-1.8.0)
[:octicons-tag-24: DDTrace-1.4.0-guance](ddtrace-ext-changelog.md#cl-1.14.0-guance)

DDTrace agent 默认的 trace-id 是 64 位，Datakit 在接收到的链路数据中 trace-id 也是 64 位，从 v1.11.0 开始支持 W3C 协议并支持接收 128 位的 trace-id。但是发送到链路中的 trace-id 依旧是 64 位。

为此，观测云做了二次开发，将 trace_128_bit_id 放到链路数据中一并发往 Datakit ，这样就能实现 DDTrace 和 OTEL 的链路串联。

可以查看参考：[GitHub issue](https://github.com/GuanceCloud/dd-trace-java/issues/37){:target="_blank"}


## 实现方式 {#how}
从 dd v1.11 开始已经支持 128 位的 traceID，目前 观测云的版本是 1.12.1。启动命令参数：

```shell
-Ddd.trace.128.bit.traceid.generation.enabled=true
# 设置透传协议
-Ddd.trace.propagation.style=tracecontext
```

但是，仅仅在 dd 内部能拿到这个 128 长度的 traceID。最后序列化再发送出去的时候其实还是还是一个 uint64，要想将这个 128ID 传递出去必须修改传输协议中的结构体。
这样的后果就是完全版本不兼容，代码也是很大改动，引起的问题会很多。

我们想到的做法是：将 `"trace_128_bit_id":xxxxxx` 放到 span 的 tags 中。在 DK 收到数据包之后发现有这个 key 则替换掉原始的 `trace_id`

放入的地方 span.build：

```java
    private DDSpan buildSpan() {
      DDSpan span = DDSpan.create(timestampMicro, buildSpanContext());
      if (span.isLocalRootSpan()) {
        EndpointTracker tracker = tracer.onRootSpanStarted(span);
        span.setEndpointTracker(tracker);
      }
      span.setTag("trace_128_bit_id",span.getTraceId().toString()); 
      return span;
    }
```

之后，所有的 span 都会在初始化的时候，将这个 key 放进去。

这样还是不够的，还需要在 Datakit 中进行筛选，如果有 `trace_128_bit_id` 则替换掉旧的 `trace-id` 。

在 观测云 链路中，所有的链路 id 都会成为 128 位的。

## OTEL 与 DDTrace 实现串联 {#otel-to-ddtrace}
OTEL 的客户端发送 http 请求到 dd 的服务端：OTEL 默认会在请求头上带上 `traceparent:00-815cf7a2d315279413e6ceb43971225f-14f64a9c3fb05612-01` （W3C 规范） 依次为 version - trace-id - parent-id - trace-flags


这样 dd 在收到请求并初始化 span 时就会将 trace-id 作为链路 id，parent-id 作为父 spanID。

效果：

<!-- markdownlint-disable MD046 MD033 -->
<figure >
  <img src="https://github.com/GuanceCloud/dd-trace-java/assets/31207055/9b599678-1ebc-4f1f-9993-f863fb25280b" style="height: 600px" alt="链路详情">
  <figcaption> 链路详情 </figcaption>
</figure>



## 更多 {#more}
目前仅实现 DDTrace 与 OTEL 串联，与其他 APM 厂商暂时没有测试。

