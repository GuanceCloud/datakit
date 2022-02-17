// Package opentelemetry is GRPC : trace & metric
package opentelemetry

import (
	"context"
	"encoding/json"
	"net"

	collectormetricepb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
)

type otlpGrpcCollector struct {
	TraceEnable  bool   `toml:"trace_enable"`
	MetricEnable bool   `toml:"metric_enable"`
	Addr         string `toml:"addr"`
	stopFunc     func()
}

func (o *otlpGrpcCollector) run() {
	ln, err := net.Listen("tcp", o.Addr)
	if err != nil {
		l.Fatalf("Failed to get an endpoint: %v", err)
		return
	}
	srv := grpc.NewServer()
	if o.TraceEnable {
		et := &ExportTrace{}
		collectortracepb.RegisterTraceServiceServer(srv, et)
	}
	if o.MetricEnable {
		em := &ExportMetric{}
		collectormetricepb.RegisterMetricsServiceServer(srv, em)
	}
	o.stopFunc = srv.Stop
	_ = srv.Serve(ln)
}

func (o *otlpGrpcCollector) stop() {
	if o.stopFunc != nil {
		o.stopFunc()
	}
}

type ExportTrace struct { //nolint:structcheck,stylecheck
	collectortracepb.UnimplementedTraceServiceServer
}

func (et *ExportTrace) Export(ctx context.Context, //nolint:structcheck,stylecheck
	ets *collectortracepb.ExportTraceServiceRequest) (*collectortracepb.ExportTraceServiceResponse, error) {
	l.Infof(ets.String())
	// ets.ProtoMessage()
	if rss := ets.GetResourceSpans(); len(rss) > 0 {
		storage.AddSpans(rss)
	}
	res := &collectortracepb.ExportTraceServiceResponse{}
	return res, nil
}

type ExportMetric struct { //nolint:structcheck,stylecheck
	collectormetricepb.UnimplementedMetricsServiceServer
}

func (et *ExportMetric) Export(ctx context.Context, //nolint:structcheck,stylecheck
	ets *collectormetricepb.ExportMetricsServiceRequest) (*collectormetricepb.ExportMetricsServiceResponse, error) {
	bts, err := json.MarshalIndent(ets.GetResourceMetrics(), "    ", "    ")
	if err == nil {
		l.Info(string(bts))
	}
	if rss := ets.ResourceMetrics; len(rss) > 0 {
		orms := toDatakitMetric(rss)
		storage.AddMetric(orms)
	}
	res := &collectormetricepb.ExportMetricsServiceResponse{}
	return res, nil
}
