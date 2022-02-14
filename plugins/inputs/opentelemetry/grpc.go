// Package opentelemetry is http
package opentelemetry

/*
	开通端口 接收 grpc 数据
*/

import (
	"context"
	"encoding/json"
	"net"
	"time"

	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

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
						orm := &otelResourceMetric{name: metrice.Name, attributes: tags,
							description: metrice.Description, dataType: p.typeName, startTime: p.startTime,
							unitTime: p.unitTime, data: p.val}
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

type point struct {
	typeName  string
	startTime uint64
	unitTime  uint64
	val       interface{}
}

func getData(metric *metricpb.Metric) []*point {
	ps := make([]*point, 0)
	switch t := metric.GetData().(type) {
	case *metricpb.Metric_IntGauge: // 弃用
	case *metricpb.Metric_Gauge:
		for _, p := range t.Gauge.DataPoints {
			point := &point{}
			if double, ok := p.Value.(*metricpb.NumberDataPoint_AsDouble); ok {
				point.val = double.AsDouble
				point.typeName = "double"
			} else if i, ok := p.Value.(*metricpb.NumberDataPoint_AsInt); ok {
				point.val = i.AsInt
				point.typeName = "int"
			}
			point.startTime = p.StartTimeUnixNano
			point.unitTime = p.TimeUnixNano
			ps = append(ps, point)
		}
	case *metricpb.Metric_IntSum: // 弃用
	case *metricpb.Metric_Sum:
		for _, p := range t.Sum.DataPoints {
			point := &point{}
			if double, ok := p.Value.(*metricpb.NumberDataPoint_AsDouble); ok {
				point.val = double.AsDouble
				point.typeName = "double"
			} else if i, ok := p.Value.(*metricpb.NumberDataPoint_AsInt); ok {
				point.val = i.AsInt
				point.typeName = "int"
			}
			//	t.Sum.AggregationTemporality
			point.startTime = p.StartTimeUnixNano
			point.unitTime = p.TimeUnixNano
			ps = append(ps, point)
		}
	case *metricpb.Metric_IntHistogram: // 弃用
	case *metricpb.Metric_Histogram:
		for _, p := range t.Histogram.DataPoints {
			point := &point{}
			point.val = p.Sum
			point.typeName = "histogram"
			point.startTime = p.StartTimeUnixNano
			point.unitTime = p.TimeUnixNano
			ps = append(ps, point)
		}
	case *metricpb.Metric_ExponentialHistogram:
		for _, p := range t.ExponentialHistogram.DataPoints {
			point := &point{
				typeName:  "ExponentialHistogram",
				startTime: p.StartTimeUnixNano,
				unitTime:  p.TimeUnixNano,
				val:       p.Sum,
			}
			ps = append(ps, point)
		}
	case *metricpb.Metric_Summary:
		for _, p := range t.Summary.DataPoints {
			point := &point{
				typeName:  "summary",
				startTime: p.StartTimeUnixNano,
				unitTime:  p.TimeUnixNano,
				val:       p.Sum,
			}
			ps = append(ps, point)
		}
	default:

	}
	return ps
}

type otelResourceMetric struct {
	name        string
	attributes  map[string]string
	description string
	dataType    string // int bool int64 float
	startTime   uint64
	unitTime    uint64
	data        interface{}
	// Exemplar 可获取 spanid 等
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
