package opentelemetry

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestExportTrace_Export(t *testing.T) {
	trace := &ExportTrace{}
	endpoint := "localhost:20010"
	m := MockOtlpGrpcCollector{trace: trace}
	go m.startServer(t, endpoint)
	<-time.After(5 * time.Millisecond)
	t.Log("start server")
	ctx := context.Background()
	exp := newGRPCExporter(t, ctx, endpoint)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(
			exp,
			// add following two options to ensure flush
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(10),
		),
	)
	t.Cleanup(func() { require.NoError(t, tp.Shutdown(ctx)) })

	tr := tp.Tracer("test-tracer")
	testKvs := []attribute.KeyValue{
		attribute.Int("Int", 1),
		attribute.Int64("Int64", int64(3)),
		attribute.Float64("Float64", 2.22),
		attribute.Bool("Bool", true),
		attribute.String("String", "test"),
	}
	_, span := tr.Start(ctx, "AlwaysSample")
	span.SetAttributes(testKvs...)
	time.Sleep(5 * time.Millisecond) // span.Duration
	span.End()
	t.Log("span end")
	// Flush and close.
	func() {
		ctx, cancel := contextWithTimeout(ctx, t, 10*time.Second)
		defer cancel()
		require.NoError(t, tp.Shutdown(ctx))
	}()

	// Wait >2 cycles.
	<-time.After(40 * time.Millisecond)

	// Now shutdown the exporter
	require.NoError(t, exp.Shutdown(ctx))

	// Shutdown the collector too so that we can begin
	// verification checks of expected data back.
	m.stopFunc()
	t.Log("stop server")
	expected := map[string]string{
		"Int":     "1",
		"Int64":   "3",
		"Float64": "2.22",
		"Bool":    "true",
		"String":  "test",
	}
	dktraces := storage.getDKTrace()
	storage.reset()

	if len(dktraces) != 1 {
		t.Errorf("dktraces.len != 1")
		return
	}

	for _, dktrace := range dktraces {
		if len(dktrace) != 1 {
			t.Errorf("dktrace.len != 1")
			return
		}
		for _, datakitSpan := range dktrace {
			if len(datakitSpan.Tags) < 5 {
				t.Errorf("tags count less 5")
				return
			}
			for key, val := range datakitSpan.Tags {
				for rkey := range expected {
					if key == rkey {
						if rval, ok := expected[rkey]; !ok || rval != val {
							t.Errorf("key=%s dk_span_val=%s  expetrd_val=%s", key, val, rval)
						}
					}
				}
			}
			if datakitSpan.Resource != "test-tracer" {
				t.Errorf("span.resource is %s  and real name is test-tracer", datakitSpan.Resource)
			}

			if datakitSpan.Operation != "AlwaysSample" {
				t.Errorf("span.Operation is %s  and real name is AlwaysSample", datakitSpan.Resource)
			}
			bts, _ := json.MarshalIndent(datakitSpan, "    ", "  ")
			t.Logf("json span = \n %s", string(bts))
		}
	}
}

func TestExportMetric_Export(t *testing.T) {
	metric := &ExportMetric{}
	endpoint := "localhost:20010"
	m := MockOtlpGrpcCollector{metric: metric}
	go m.startServer(t, endpoint)
	<-time.After(5 * time.Millisecond)
	t.Log("start server")

	ctx := context.Background()
	exp := newMetricGRPCExporter(t, ctx, endpoint)

	err := exp.Export(ctx, testResource, oneRecord)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := exp.Shutdown(ctx); err != nil {
			panic(err)
		}
	}()
	m.stopFunc()
	ms := storage.getDKMetric()
	if len(ms) != 1 {
		t.Errorf("metric len != 1")
	}
	want := &otelResourceMetric{
		Operation: "foo",
		Attributes: map[string]string{
			"abc": "def",
			"one": "1",
		},
		Source:    inputName,
		Resource:  "onelib",
		ValueType: "int",
		Value:     42,
		StartTime: uint64(time.Date(2020, time.December, 8, 19, 15, 0, 0, time.UTC).UnixNano()),
		UnitTime:  uint64(time.Date(2020, time.December, 8, 19, 16, 0, 0, time.UTC).UnixNano()),
	}

	got := ms[0]
	if !reflect.DeepEqual(got.Attributes, want.Attributes) {
		t.Errorf("tags got.tag=%+v  want.tag=%+v", got.Attributes, want.Attributes)
	}
	if got.Operation != want.Operation {
		t.Errorf("operation got=%s want=%s", got.Operation, want.Operation)
	}
}
