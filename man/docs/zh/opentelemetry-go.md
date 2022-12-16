# Golang 示例
---

本文以常见 Web 端三层架构模式实现 OTEL 的链路追踪及可观测性。

在使用 OTEL 发送 Trace 到 Datakit 之前，请先确定您已经[配置好了采集器](opentelemetry.md)。

## 接下来使用伪代码实现 {#code}

模拟场景：一条用户的登录请求在服务端的各个模块流转并返回到客户端的过程。在每一个过程中都加上链路追踪并标记，最后登录观测云平台查看在这个过程中每个模块的处理时间和服务状态。

流程介绍：用户请求到 web 层，解析后发送到 service 层，需要查询数据库的 dao 层，最终将结果返回到用户。

``` go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel"
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

// Initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func initProvider() func() {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String("ServerName"),
			// semconv.FaaSIDKey.String(""),
		),
		//resource.WithOS(), // and so on ...
	)

	handleErr(err, "failed to create resource")
	var bsp sdktrace.SpanProcessor

	// If the OpenTelemetry Collector is running on a local cluster (minikube or
	// microk8s), it should be accessible through the NodePort service at the
	// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
	// endpoint of your cluster. If you run the app inside k8s, then you can
	// probably connect directly to the service through dns
	conn, err := grpc.DialContext(ctx, "10.200.14.226:4317", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	handleErr(err, "failed to create gRPC connection to collector")
	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	handleErr(err, "failed to create trace exporter")

	bsp = sdktrace.NewBatchSpanProcessor(traceExporter)

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return func() {
		// Shutdown will flush any remaining spans and shut down the exporter.
		handleErr(tracerProvider.Shutdown(ctx), "failed to shutdown TracerProvider")
		time.Sleep(time.Second)
	}
}

var tracer = otel.Tracer("tracer_user_login")

// web handler 处理请求数据
func web(w http.ResponseWriter, r *http.Request) {
	// ... 接收客户端请求
	log.Println("doing web")
	// labels represent additional key-value descriptors that can be bound to a
	// metric observer or recorder.
	commonLabels := []attribute.KeyValue{attribute.String("key1", "val1")}
	// work begins
	ctx, span := tracer.Start(
		context.Background(),
		"span-Example",
		trace.WithAttributes(commonLabels...))
	defer span.End()
	<-time.After(time.Millisecond * 50)
	service(ctx)

	log.Printf("Doing really hard work")
	<-time.After(time.Millisecond * 40)

	log.Printf("Done!")
	w.Write([]byte("ok"))
}

// service 调用service层 处理业务
func service(ctx context.Context) {
	log.Println("service")
	ctx1, iSpan := tracer.Start(ctx, "Sample-service")
	<-time.After(time.Second / 2) // do something...
	dao(ctx1)
	iSpan.End()
}

// dao 数据访问层
func dao(ctx context.Context) {
	log.Println("dao")
	ctxD, iSpan := tracer.Start(ctx, "Sample-dao")
	<-time.After(time.Second / 2)

	// 创建子 span 查询数据库等操作
	_, sqlSpan := tracer.Start(ctxD, "do_sql")
	sqlSpan.SetStatus(codes.Ok, "is ok") //
	<-time.After(time.Second)
	sqlSpan.End()

	iSpan.End()
}

func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}

func main() {
	shutdown := initProvider()
	defer shutdown()
	log.Println("connect ...")
	http.HandleFunc("/user", web)
	go handleErr(http.ListenAndServe(":4317", nil), "open server")
	time.Sleep(time.Minute * 2)
	os.Exit(0)
}
```

## 效果查看 {#view}

登录 [观测云](https://console.guance.com/tracing/service/table?time=15m){:target="_blank"} 后查看 `应用性能监测` -> `链路` -> 点击单条 `链路`

![](imgs/otel-go-example.png)

在火焰图中可看到每一个模块中执行的时间、调用流程等。

--- 

## 参考 {#more-readings}

- 源码示例 [github-opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go/tree/main/example/otel-collector){:target="_blank"}
- 文档 [官方文档](https://opentelemetry.io/docs/instrumentation/go/getting-started/){:target="_blank"}
