---
title     : 'OpenTelemetry Golang'
summary   : 'Tracing Golang applications with OpenTelemetry'
tags      :
  - 'GOLANG'
  - 'OTEL'
  - 'APM'
  - 'TRACING'
__int_icon: 'icon/opentelemetry'
---

This example shows a simple three-layer web architecture in Go and how to add OTEL spans.

Before you export traces to DataKit, make sure [the OpenTelemetry collector input is configured](opentelemetry.md).

## Example Scenario {#code}

Workflow:

1. A client sends a login request.
2. Web layer receives request and creates a top-level span.
3. Web layer calls service layer.
4. Service layer executes database query.
5. Each stage creates child spans with attributes and timing.

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

// initProvider initializes OTLP gRPC exporter and trace provider.
func initProvider() func() {
    ctx := context.Background()

    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceNameKey.String("ServerName"),
        ),
    )
    handleErr(err, "failed to create resource")

    // For demo purposes, use local collector at 127.0.0.1:4317.
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

// user receives request and creates one span per business step.
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

## Reference {#more-readings}

- [OpenTelemetry Go sample](https://github.com/open-telemetry/opentelemetry-go/tree/main/example/otel-collector){:target="_blank"}
- [OpenTelemetry Go getting started](https://opentelemetry.io/docs/instrumentation/go/getting-started/){:target="_blank"}
