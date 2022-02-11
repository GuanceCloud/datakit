package opentelemetry

/*
	开通端口 接收 grpc 数据
*/

import (
	"context"
	"net"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	collectormetricepb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	trace "go.opentelemetry.io/proto/otlp/trace/v1"
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

type ExportTrace struct {
	collectortracepb.UnimplementedTraceServiceServer
	errors      []error
	requests    int
	mu          sync.RWMutex
	storage     []*trace.Span
	headers     metadata.MD
	exportBlock chan struct{}
}

func (et *ExportTrace) Export(ctx context.Context,
	ets *collectortracepb.ExportTraceServiceRequest) (*collectortracepb.ExportTraceServiceResponse, error) {
	l.Infof(ets.String())
	// ets.ProtoMessage()
	if rss := ets.GetResourceSpans(); rss != nil && len(rss) > 0 {
		storage.AddSpans(rss)
	}
	res := &collectortracepb.ExportTraceServiceResponse{}
	return res, nil
}

type ExportMetric struct {
	collectormetricepb.UnimplementedMetricsServiceServer
	errors      []error
	requests    int
	mu          sync.RWMutex
	storage     []*trace.Span
	headers     metadata.MD
	exportBlock chan struct{}
}

func (et *ExportMetric) Export(ctx context.Context,
	ets *collectormetricepb.ExportMetricsServiceRequest) (*collectormetricepb.ExportMetricsServiceResponse, error) {
	// header
	header, b := metadata.FromOutgoingContext(ctx)
	if b {
		l.Infof("len =%d", header.Len())
	}
	l.Infof(ets.String())
	// ets.ProtoMessage()
	if rss := ets.ResourceMetrics; rss != nil && len(rss) > 0 {
		for _, resourceMetrics := range rss {
			LibraryMetrics := resourceMetrics.GetInstrumentationLibraryMetrics()
			for _, libraryMetric := range LibraryMetrics {
				metrices := libraryMetric.GetMetrics()
				for _, metrice := range metrices {
					l.Debugf(metrice.Name)
					l.Infof("metric string=%s", metrice.String())
				}
			}
		}
	}
	res := &collectormetricepb.ExportMetricsServiceResponse{}
	return res, nil
}

type DKMetric struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *DKMetric) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

//nolint:lll
func (m *DKMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   inputName,
		Type:   "metric",
		Desc:   "opentelemetry 指标",
		Fields: map[string]interface{}{},
		Tags:   map[string]interface{}{},
	}
}
