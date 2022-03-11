// Package opentelemetry is GRPC : trace & metric
package opentelemetry

import (
	"context"
	"fmt"
	"net"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/opentelemetry/collector"
	collectormetricepb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type otlpGrpcCollector struct {
	TraceEnable     bool   `toml:"trace_enable"`
	MetricEnable    bool   `toml:"metric_enable"`
	Addr            string `toml:"addr"`
	ExpectedHeaders map[string]string
	stopFunc        func()
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
	ExpectedHeaders map[string]string
	storage         *collector.SpansStorage
}

func (et *ExportTrace) Export(ctx context.Context, //nolint:structcheck,stylecheck
	ets *collectortracepb.ExportTraceServiceRequest) (*collectortracepb.ExportTraceServiceResponse, error) {
	md, haveHeader := metadata.FromIncomingContext(ctx)
	if haveHeader {
		if !checkHandler(et.ExpectedHeaders, md) {
			return nil, fmt.Errorf("invalid request haeders or nil headers")
		}
	}
	if rss := ets.GetResourceSpans(); len(rss) > 0 {
		et.storage.AddSpans(rss)
	}
	res := &collectortracepb.ExportTraceServiceResponse{}
	return res, nil
}

type ExportMetric struct { //nolint:structcheck,stylecheck
	collectormetricepb.UnimplementedMetricsServiceServer
	ExpectedHeaders map[string]string
	storage         *collector.SpansStorage
}

func (em *ExportMetric) Export(ctx context.Context, //nolint:structcheck,stylecheck
	ets *collectormetricepb.ExportMetricsServiceRequest) (*collectormetricepb.ExportMetricsServiceResponse, error) {
	md, haveHeader := metadata.FromIncomingContext(ctx)
	if haveHeader {
		if !checkHandler(em.ExpectedHeaders, md) {
			return nil, fmt.Errorf("invalid request haeders or nil headers")
		}
	}
	if rss := ets.ResourceMetrics; len(rss) > 0 {
		orms := em.storage.ToDatakitMetric(rss)
		em.storage.AddMetric(orms)
	}
	res := &collectormetricepb.ExportMetricsServiceResponse{}
	return res, nil
}

func checkHandler(headers map[string]string, md metadata.MD) bool {
	if len(headers) == 0 {
		return true
	}
	for k, v := range headers {
		strs := md.Get(strings.ToLower(k))
		mdVal := strings.Join(strs, ",")
		if mdVal != v {
			return false
		}
	}
	return true
}
