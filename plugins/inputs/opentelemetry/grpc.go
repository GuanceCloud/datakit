// Package opentelemetry is http
package opentelemetry

/*
	开通端口 接收 grpc 数据
*/

import (
	"context"
	"encoding/json"
	"net"

	collectormetricepb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
	// errors      []error
	// requests    int
	// mu          sync.RWMutex
	// storage     []*trace.Span
	// headers     metadata.MD
	// exportBlock chan struct{}
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
	// errors      []error
	// requests    int
	// mu          sync.RWMutex
	// storage     []*trace.Span
	// headers     metadata.MD
	// exportBlock chan struct{}
}

func (et *ExportMetric) Export(ctx context.Context, //nolint:structcheck,stylecheck
	ets *collectormetricepb.ExportMetricsServiceRequest) (*collectormetricepb.ExportMetricsServiceResponse, error) {
	// header
	header, b := metadata.FromOutgoingContext(ctx)
	if b {
		l.Infof("len =%d", header.Len())
	}
	l.Infof(ets.String())
	bts, err := json.MarshalIndent(ets.GetResourceMetrics(), "    ", "")
	if err == nil {
		l.Info(string(bts))
	}
	// ets.ProtoMessage()
	orms := make([]*otelResourceMetric, 0)
	if rss := ets.ResourceMetrics; len(rss) > 0 {
		for _, resourceMetrics := range rss {
			tags := toDatakitTags(resourceMetrics.Resource.Attributes)
			LibraryMetrics := resourceMetrics.GetInstrumentationLibraryMetrics()
			for _, libraryMetric := range LibraryMetrics {
				metrices := libraryMetric.GetMetrics()
				for _, metrice := range metrices {
					l.Debugf(metrice.Name)
					bts, err := json.MarshalIndent(metrice, "\t", "")
					if err == nil {
						l.Info(string(bts))
					}
					l.Infof("metric string=%s", metrice.String())
					ps := getData(metrice)
					for _, p := range ps {
						orm := &otelResourceMetric{
							name: metrice.Name, attributes: tags,
							description: metrice.Description, dataType: p.typeName, startTime: p.startTime,
							unitTime: p.unitTime, data: p.val,
						}
						orms = append(orms, orm)
						// todo 将 orms 转换成 行协议格式 并发送到IO
					}
				}
			}
		}
	}
	l.Infof("orms len=%d", len(orms))
	res := &collectormetricepb.ExportMetricsServiceResponse{}
	return res, nil
}
