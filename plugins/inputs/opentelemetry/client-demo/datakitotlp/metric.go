package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric/global"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

func initMetric() func() {
	ctxM := context.Background()
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint("10.200.14.226:9550"),
		otlpmetricgrpc.WithReconnectionPeriod(50 * time.Millisecond),
		otlpmetricgrpc.WithHeaders(map[string]string{"header": "1"}), // 开启校验 header
	}

	// opts = append(opts, additionalOpts...)
	client := otlpmetricgrpc.NewClient(opts...)
	exp, err := otlpmetric.New(ctxM, client)
	handleErr(err, "otlpmetric.New")
	res, _ := resource.New(ctxM, resource.WithHost(), resource.WithOS(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("serviceNameForMetric"),
			semconv.ProcessPIDKey.Int(os.Getpid()), // set pid
			// and so on ...
		),
	)

	pusher := controller.New(
		processor.NewFactory(
			simple.NewWithHistogramDistribution(),
			exp,
		),
		controller.WithResource(res),
		controller.WithExporter(exp),
		controller.WithCollectPeriod(2*time.Second),
	)

	global.SetMeterProvider(pusher)

	if err := pusher.Start(ctxM); err != nil {
		log.Fatalf("could not start metric controoler: %v", err)
	}
	/*defer func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		// pushes any last exports to the receiver
		if err := pusher.Stop(ctx); err != nil {
			otel.Handle(err)
		}
	}()*/
	return func() {
		handleErr(pusher.Stop(ctxM), "failed to shutdown pusher")
	}
}
