package opentelemetry

import (
	"context"
	"encoding/json"
	"net"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric/metrictest"
	"go.opentelemetry.io/otel/metric/number"
	"go.opentelemetry.io/otel/metric/sdkapi"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/sum"
	"go.opentelemetry.io/otel/sdk/metric/export"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	"go.opentelemetry.io/otel/sdk/resource"
	collectormetricepb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func contextWithTimeout(parent context.Context, t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	d, ok := t.Deadline()
	if !ok {
		d = time.Now().Add(timeout)
	} else {
		d = d.Add(-1 * time.Millisecond)
		now := time.Now()
		if d.Sub(now) > timeout {
			d = now.Add(timeout)
		}
	}
	return context.WithDeadline(parent, d)
}

type MockOtlpGrpcCollector struct {
	trace    collectortracepb.TraceServiceServer
	metric   *ExportMetric
	Addr     string
	stopFunc func()
}

func (m *MockOtlpGrpcCollector) startServer(t *testing.T, endpoint string) {
	ln, err := net.Listen("tcp", endpoint)
	if err != nil {
		t.Fatalf("Failed to get an endpoint: %v", err)
		return
	}

	srv := grpc.NewServer()
	if m.trace != nil {
		collectortracepb.RegisterTraceServiceServer(srv, m.trace)
	}
	if m.metric != nil {
		collectormetricepb.RegisterMetricsServiceServer(srv, m.metric)
	}

	m.stopFunc = srv.Stop
	_ = srv.Serve(ln)
}

func newGRPCExporter(t *testing.T, ctx context.Context, endpoint string, additionalOpts ...otlptracegrpc.Option) *otlptrace.Exporter {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithReconnectionPeriod(50 * time.Millisecond),
	}

	opts = append(opts, additionalOpts...)
	client := otlptracegrpc.NewClient(opts...)
	exp, err := otlptrace.New(ctx, client)
	if err != nil {
		t.Fatalf("failed to create a new collector exporter: %v", err)
	}
	return exp
}

func newMetricGRPCExporter(t *testing.T, ctx context.Context, endpoint string, additionalOpts ...otlpmetricgrpc.Option) *otlpmetric.Exporter {
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithReconnectionPeriod(50 * time.Millisecond),
	}

	opts = append(opts, additionalOpts...)
	client := otlpmetricgrpc.NewClient(opts...)
	exp, err := otlpmetric.New(ctx, client)
	if err != nil {
		t.Fatalf("failed to create a new collector exporter: %v", err)
	}
	return exp
}

func MultiInstrumentationLibraryReader(records map[instrumentation.Library][]export.Record) export.InstrumentationLibraryReader {
	return instrumentationLibraryReader{records: records}
}

type instrumentationLibraryReader struct {
	records map[instrumentation.Library][]export.Record
}

var _ export.InstrumentationLibraryReader = instrumentationLibraryReader{}

func (m instrumentationLibraryReader) ForEach(fn func(instrumentation.Library, export.Reader) error) error {
	for library, records := range m.records {
		if err := fn(library, &metricReader{records: records}); err != nil {
			return err
		}
	}
	return nil
}

type metricReader struct {
	sync.RWMutex
	records []export.Record
}

var _ export.Reader = &metricReader{}

func (m *metricReader) ForEach(_ aggregation.TemporalitySelector, fn func(export.Record) error) error {
	for _, record := range m.records {
		if err := fn(record); err != nil && err != aggregation.ErrNoData {
			return err
		}
	}
	return nil
}

func getReader() export.InstrumentationLibraryReader {
	desc := metrictest.NewDescriptor(
		"foo",
		sdkapi.CounterInstrumentKind,
		number.Int64Kind,
	)
	agg := sum.New(1)
	if err := agg[0].Update(context.Background(), number.NewInt64Number(42), &desc); err != nil {
		panic(err)
	}
	start := time.Date(2020, time.December, 8, 19, 15, 0, 0, time.UTC)

	end := time.Date(2020, time.December, 8, 19, 16, 0, 0, time.UTC)

	labels := attribute.NewSet(attribute.String("abc", "def"), attribute.Int64("one", 1))

	rec := export.NewRecord(&desc, &labels, agg[0].Aggregation(), start, end)

	return MultiInstrumentationLibraryReader(
		map[instrumentation.Library][]export.Record{
			{
				Name: "onelib",
			}: {rec},
		})
}

var (
	oneRecord = getReader()

	testResource = resource.Empty()
)

func (o *otelResourceMetric) toString() string {
	bts, err := json.MarshalIndent(o, "    ", "    ")
	if err != nil {
		l.Errorf("marshal err= %v", err)
		return ""
	}
	return string(bts)
}

type MockTrace struct {
	collectortracepb.UnimplementedTraceServiceServer
	Rss     []*tracepb.ResourceSpans
	Headers metadata.MD
}

func (et *MockTrace) getResourceSpans() []*tracepb.ResourceSpans {
	return et.Rss
}

func (et *MockTrace) GetHeader() metadata.MD {
	return et.Headers
}

func (et *MockTrace) Export(ctx context.Context,
	ets *collectortracepb.ExportTraceServiceRequest) (*collectortracepb.ExportTraceServiceResponse, error) {
	l.Infof(ets.String())
	// ets.ProtoMessage()
	if rss := ets.GetResourceSpans(); len(rss) > 0 {
		et.Rss = rss
	}
	et.Headers, _ = metadata.FromOutgoingContext(ctx)

	res := &collectortracepb.ExportTraceServiceResponse{}
	return res, nil
}
