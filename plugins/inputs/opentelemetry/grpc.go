package opentelemetry

/*
	开通端口 接收 grpc 数据
*/

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	collectormetricpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
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
		fmt.Printf("Failed to get an endpoint: %v", err)
	}

	srv := grpc.NewServer()
	if o.TraceEnable {
		et := &ExportTrace{}
		collectortracepb.RegisterTraceServiceServer(srv, et)
	}
	if o.MetricEnable {
		em := &MetricService{}
		collectormetricpb.RegisterMetricsServiceServer(srv, em)
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
	rss := ets.GetResourceSpans()
	if rss != nil && len(rss) > 0 {
		log.Printf("rss len =%d", len(rss))
		for _, rs := range rss {
			ls := rs.GetInstrumentationLibrarySpans()
			log.Printf("ls len =%d", len(ls))
			for _, l := range ls {
				spans := l.GetSpans()
				log.Printf("span len =%d", len(spans))
				for _, span := range spans {
					// todo span
					fmt.Println(span.Name)
				}
			}
		}
	}
	return nil, nil
}

type MetricService struct {
	collectormetricpb.UnimplementedMetricsServiceServer

	requests int
	errors   []error

	headers metadata.MD
	mu      sync.RWMutex
	// storage otlpmetrictest.MetricsStorage
	delay time.Duration
}

func (et *MetricService) Export(ctx context.Context,
	ets *collectormetricpb.ExportMetricsServiceRequest) (*collectormetricpb.ExportMetricsServiceResponse, error) {
	rss := ets.GetResourceMetrics()
	if rss != nil && len(rss) > 0 {
		log.Printf("rss len =%d", len(rss))
		for _, rs := range rss {
			ls := rs.GetInstrumentationLibraryMetrics()
			log.Printf("ls len =%d", len(ls))
			for _, l := range ls {
				spans := l.GetMetrics()
				log.Printf("span len =%d", len(spans))
				for _, span := range spans {
					// todo metric
					fmt.Println(span.Name)
				}
			}
		}
	}
	return nil, nil
}
