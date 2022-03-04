// Package opentelemetry is GRPC : trace & metric
package opentelemetry

import (
	"context"
	"encoding/json"
	"net"

	"google.golang.org/grpc/metadata"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/opentelemetry/collector"
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

func (o *otlpGrpcCollector) run(storage *collector.SpansStorage) {
	ln, err := net.Listen("tcp", o.Addr)
	if err != nil {
		l.Errorf("Failed to get an endpoint: %v", err)
		return
	}
	srv := grpc.NewServer()
	if o.TraceEnable {
		et := &ExportTrace{storage: storage}
		collectortracepb.RegisterTraceServiceServer(srv, et)
	}
	if o.MetricEnable {
		em := &ExportMetric{storage: storage}
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
	storage *collector.SpansStorage
}

func (et *ExportTrace) Export(ctx context.Context, //nolint:structcheck,stylecheck
	ets *collectortracepb.ExportTraceServiceRequest) (*collectortracepb.ExportTraceServiceResponse, error) {
	headers, b := metadata.FromIncomingContext(ctx)
	if b {
		l.Infof("headers=%+v", headers)
	}
	l.Infof(ets.String())
	// ets.ProtoMessage()
	if rss := ets.GetResourceSpans(); len(rss) > 0 {
		et.storage.AddSpans(rss)
	}
	res := &collectortracepb.ExportTraceServiceResponse{}
	return res, nil
}

type ExportMetric struct { //nolint:structcheck,stylecheck
	collectormetricepb.UnimplementedMetricsServiceServer
	storage *collector.SpansStorage
}

func (em *ExportMetric) Export(ctx context.Context, //nolint:structcheck,stylecheck
	ets *collectormetricepb.ExportMetricsServiceRequest) (*collectormetricepb.ExportMetricsServiceResponse, error) {
	bts, err := json.MarshalIndent(ets.GetResourceMetrics(), "    ", "    ")
	if err == nil {
		l.Info(string(bts))
	}
	if rss := ets.ResourceMetrics; len(rss) > 0 {
		orms := em.storage.ToDatakitMetric(rss)
		em.storage.AddMetric(orms)
	}
	res := &collectormetricepb.ExportMetricsServiceResponse{}
	return res, nil
}
