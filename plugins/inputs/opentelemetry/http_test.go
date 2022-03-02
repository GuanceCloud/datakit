package opentelemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/opentelemetry/collector"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	v1 "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"
)

func Test_otlpHTTPCollector_apiOtlpTrace(t *testing.T) {
	cfg := collector.MockCollectorConfig{
		URLPath:         "/otel/v1/trace",
		Port:            20010,
		ExpectedHeaders: map[string]string{"header1": "1"}}

	o := otlpHTTPCollector{
		storage:         collector.NewSpansStorage(),
		Enable:          true,
		HTTPStatusOK:    200,
		ExpectedHeaders: map[string]string{"header1": "1"},
	}
	mockserver := collector.RunMockCollector(t, cfg, o.apiOtlpTrace)
	time.Sleep(time.Millisecond * 5) // 等待 server 端口开启

	// mock client
	ctx := context.Background()
	exp := collector.NewHTTPExporter(t, ctx, cfg.URLPath, "localhost:20010")

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
		ctx, cancel := collector.ContextWithTimeout(ctx, t, 10*time.Second)
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
	dktraces := o.storage.GetDKTrace()

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

// todo metric api test

func Test_otlpHTTPCollector_checkHeaders(t *testing.T) {
	type fields struct {
		Enable          bool
		HTTPStatusOK    int
		ExpectedHeaders map[string]string
	}
	type args struct {
		r *http.Request
	}
	req, err := http.NewRequest("post", "", nil)
	if err != nil {
		t.Errorf("err=%v", err)
		return
	}
	req.Header.Add("header", "1")
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "case1",
			fields: fields{Enable: true, HTTPStatusOK: 200, ExpectedHeaders: map[string]string{"header": "1"}},
			args:   args{r: req},
			want:   true,
		},
		{
			name:   "case2",
			fields: fields{Enable: true, HTTPStatusOK: 200, ExpectedHeaders: map[string]string{"header": "2"}},
			args:   args{r: req},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &otlpHTTPCollector{
				Enable:          tt.fields.Enable,
				HTTPStatusOK:    tt.fields.HTTPStatusOK,
				ExpectedHeaders: tt.fields.ExpectedHeaders,
			}
			if got := o.checkHeaders(tt.args.r); got != tt.want {
				t.Errorf("checkHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_unmarshalTraceRequest(t *testing.T) {
	type args struct {
		rawRequest  []byte
		contentType string
	}
	rss := &collectortracepb.ExportTraceServiceRequest{ResourceSpans: []*v1.ResourceSpans{&v1.ResourceSpans{SchemaUrl: "aaaaaaa"}}}
	bts, err := proto.Marshal(rss)
	if err != nil {
		t.Errorf("err=%v", err)
		return
	}

	tests := []struct {
		name    string
		args    args
		want    *collectortracepb.ExportTraceServiceRequest
		wantErr bool
	}{
		{
			name:    "case1",
			args:    args{rawRequest: bts, contentType: "application/x-protobuf"},
			want:    rss,
			wantErr: false,
		},
		{
			name:    "case2",
			args:    args{rawRequest: bts, contentType: ""},
			want:    rss,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshalTraceRequest(tt.args.rawRequest, tt.args.contentType)
			if err != nil {
				if tt.wantErr {
					return
				} else {
					t.Errorf("unmarshalTraceRequest() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}

			/*
				踩坑： 对象做 proto.marshal 之前和之后的数据是不一样的。所以 reflect.Equal 会返回 false
					got 通过 proto.marshal 之后，再进行对比就会返回 true。将下面四行删掉 就会返回false
			*/
			gbts, _ := proto.Marshal(got)
			if !bytes.Equal(gbts, tt.args.rawRequest) {
				t.Errorf("byte not equal")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unmarshalTraceRequest() got = %+v,\n want %+v", got, tt.want)
			}
		})
	}
}
