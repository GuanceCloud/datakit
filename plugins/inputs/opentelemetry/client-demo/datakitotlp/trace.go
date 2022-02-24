package main

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
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
			semconv.ServiceNameKey.String(globalConf.ServerName),
			// semconv.FaaSIDKey.String(""),
		),
		resource.WithOS(), // and so on ...
	)

	handleErr(err, "failed to create resource")
	var bsp sdktrace.SpanProcessor
	if globalConf.GrpcOpen {
		// If the OpenTelemetry Collector is running on a local cluster (minikube or
		// microk8s), it should be accessible through the NodePort service at the
		// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
		// endpoint of your cluster. If you run the app inside k8s, then you can
		// probably connect directly to the service through dns
		conn, err := grpc.DialContext(ctx, globalConf.GrpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		handleErr(err, "failed to create gRPC connection to collector")
		// Set up a trace exporter
		traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
		handleErr(err, "failed to create trace exporter")

		bsp = sdktrace.NewBatchSpanProcessor(traceExporter)
	}
	if !globalConf.GrpcOpen && globalConf.HTTPOpen {
		// todo http 最佳实践
		httpClent, err := otlptracehttp.New(
			ctx,
			otlptracehttp.WithURLPath(globalConf.HttpAddr),
			otlptracehttp.WithEndpoint(globalConf.HttpEndpoint),
			otlptracehttp.WithInsecure(),
			otlptracehttp.WithHeaders(map[string]string{"header1": "1"}))
		handleErr(err, "init http conn")
		bsp = sdktrace.NewBatchSpanProcessor(httpClent)
	}

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
