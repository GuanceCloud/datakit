# Golang Example
---

In this paper, the common Web-side three-tier architecture mode is used to realize the link tracing and observability of OTEL.

Before using OTEL to send Trace to Datakit, make sure you have [configured the collector](opentelemetry.md).

## Next, implement in pseudo code {#code}

Simulation scenario: A user's login request flows through various modules on the server and is returned to the client. Add link tracking and marking in each process, and finally log in to the Guance Cloud platform to check the processing time and service status of each module in this process.

Process introduction: The user requests to the web layer, after analysis, send to the service layer, need to query the database of the dao layer, and finally return the results to the user.

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

## View Effect {#view}

Log in to [Guance Cloud](https://console.guance.com/tracing/service/table?time=15m){:target="_blank"} and then view `application performance monitoring` -> `links` -> Click on a single `link`

![](imgs/otel-go-example.png)

In the flame diagram, you can see the execution time, call flow and so on in each module.

--- 

## Reference {#more-readings}

- Source sample [github-opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go/tree/main/example/otel-collector){:target="_blank"}
- [Doc](https://opentelemetry.io/docs/instrumentation/go/getting-started/){:target="_blank"}
