package opentelemetry

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func Test_otlpHTTPCollector_apiOtlpTrace(t *testing.T) {
	cfg := mockCollectorConfig{
		URLPath:         "/otel/v1/trace",
		Port:            20010,
		ExpectedHeaders: map[string]string{"header1": "1"}}

	o := otlpHTTPCollector{
		Enable:          true,
		HTTPStatusOK:    200,
		ExpectedHeaders: map[string]string{"header1": "1"},
	}
	mockserver := runMockCollector(t, cfg, o.apiOtlpTrace)
	time.Sleep(time.Millisecond * 5) // 等待 server 端口开启

	// mock client
	ctx := context.Background()
	exp := newHTTPExporter(t, ctx, cfg.URLPath, "localhost:20010")

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
	_, span := tr.Start(ctx, "AlwaysSample", trace.WithSpanKind(1))
	span.SetAttributes(testKvs...)
	time.Sleep(5 * time.Millisecond) // span.Duration
	// span.End()
	span.End(trace.WithStackTrace(true))
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
	_ = mockserver.Stop()
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
