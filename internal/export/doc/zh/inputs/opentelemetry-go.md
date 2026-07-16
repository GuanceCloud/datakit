---
title     : 'OpenTelemetry Golang'
summary   : 'OpenTelemetry Golang 集成'
tags      :
  - 'GOLANG'
  - 'OTEL'
  - 'APM'
  - '链路追踪'
__int_icon: 'icon/opentelemetry'
---

本示例以常见三层 Web 架构展示如何使用 OpenTelemetry 在 Go 服务中接入链路追踪。

在向 DataKit 上报 Trace 前，请先完成 [OpenTelemetry 采集器配置](opentelemetry.md)。

## 示例流程 {#code}

示例流程：

1. 客户端发起登录请求；
2. Web 层接收请求并创建一个顶层 Span；
3. Web 层调用服务层；
4. 服务层执行数据库查询；
5. 各层均创建子 Span，并记录时延与属性。

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "time"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
    "go.opentelemetry.io/otel/trace"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

// 初始化 OTLP gRPC exporter 与 trace provider。
func initProvider() func() {
    ctx := context.Background()

    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceNameKey.String("ServerName"),
        ),
    )
    handleErr(err, "failed to create resource")

    // 示例环境中，Collector 运行在本机 127.0.0.1:4317
    conn, err := grpc.DialContext(ctx, "127.0.0.1:4317",
        grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
    handleErr(err, "failed to create gRPC connection to collector")

    traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
    handleErr(err, "failed to create trace exporter")

    bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
    tracerProvider := sdktrace.NewTracerProvider(
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
        sdktrace.WithResource(res),
        sdktrace.WithSpanProcessor(bsp),
    )
    otel.SetTracerProvider(tracerProvider)
    otel.SetTextMapPropagator(propagation.TraceContext{})

    return func() {
        handleErr(tracerProvider.Shutdown(ctx), "failed to shutdown TracerProvider")
        time.Sleep(time.Second)
    }
}

var tracer = otel.Tracer("tracer_user_login")

// user 接收请求并按业务步骤创建 span。
func user(w http.ResponseWriter, r *http.Request) {
    log.Println("receiving user request")
    commonLabels := []attribute.KeyValue{attribute.String("key1", "val1")}

    ctx, span := tracer.Start(
        context.Background(),
        "span-Example",
        trace.WithAttributes(commonLabels...),
    )
    defer span.End()

    <-time.After(time.Millisecond * 50)
    service(ctx)

    log.Printf("Done!")
    w.Write([]byte("ok"))
}

func service(ctx context.Context) {
    ctx1, iSpan := tracer.Start(ctx, "Sample-service")
    defer iSpan.End()

    <-time.After(time.Second / 2)
    dao(ctx1)
}

func dao(ctx context.Context) {
    ctxD, iSpan := tracer.Start(ctx, "Sample-dao")
    defer iSpan.End()

    _, sqlSpan := tracer.Start(ctxD, "do_sql")
    sqlSpan.SetStatus(codes.Ok, "query done")
    <-time.After(time.Second)
    sqlSpan.End()
}

func handleErr(err error, message string) {
    if err != nil {
        log.Fatalf("%s: %v", message, err)
    }
}

func main() {
    shutdown := initProvider()
    defer shutdown()

    log.Println("listening on :8080")
    http.HandleFunc("/user", user)
    go handleErr(http.ListenAndServe(":8080", nil), "open server")
    time.Sleep(time.Minute * 2)
    os.Exit(0)
}
```

## 参考 {#more-readings}

- [OpenTelemetry Go 示例](https://github.com/open-telemetry/opentelemetry-go/tree/main/example/otel-collector){:target="_blank"}
- [OpenTelemetry Go 官方文档](https://opentelemetry.io/docs/instrumentation/go/getting-started/){:target="_blank"}
