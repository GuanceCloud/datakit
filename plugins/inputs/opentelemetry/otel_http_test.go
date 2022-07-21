// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/opentelemetry/collector"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/opentelemetry/mock"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	collectormetricpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	v12 "go.opentelemetry.io/proto/otlp/metrics/v1"
	v1 "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"
)

func Test_otlpHTTPCollector_apiOtlpTrace(t *testing.T) {
	cfg := mock.MockCollectorConfig{
		URLPath:         "/otel/v1/trace",
		Port:            20010,
		ExpectedHeaders: map[string]string{"header1": "1"},
	}

	o := otlpHTTPCollector{
		storage:         collector.NewSpansStorage(),
		Enable:          true,
		HTTPStatusOK:    200,
		ExpectedHeaders: map[string]string{"header1": "1"},
	}
	mockserver := mock.RunMockCollector(t, cfg, o.apiOtlpTrace)
	time.Sleep(time.Millisecond * 5) // 等待 server 端口开启

	// mock client
	ctx := context.Background()
	exp := mock.NewHTTPExporter(t, ctx, cfg.URLPath, "localhost:20010")

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
		ctx, cancel := mock.ContextWithTimeout(ctx, t, 10*time.Second)
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
			if datakitSpan.Resource != "AlwaysSample" {
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
	rss := &collectortracepb.ExportTraceServiceRequest{ResourceSpans: []*v1.ResourceSpans{{SchemaUrl: "aaaaaaa"}}}
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

func GZipBytes(data []byte) []byte {
	var input bytes.Buffer
	g := gzip.NewWriter(&input)
	_, _ = g.Write(data)
	_ = g.Close()
	return input.Bytes()
}

// func Test_readGzipBody(t *testing.T) {
// 	type args struct {
// 		body io.Reader
// 	}
// 	var in bytes.Buffer
// 	bts := GZipBytes([]byte{'a', 'b', 'c', 'd'})
// 	in.Write(bts)

// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    []byte
// 		wantErr bool
// 	}{
// 		{
// 			name:    "case",
// 			args:    args{body: &in},
// 			want:    []byte{'a', 'b', 'c', 'd'},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := readGzipBody(tt.args.body)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("readGzipBody() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("readGzipBody() got = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func Test_unmarshalMetricsRequest(t *testing.T) {
	type args struct {
		rawRequest  []byte
		contentType string
	}

	metrics := &collectormetricpb.ExportMetricsServiceRequest{ResourceMetrics: []*v12.ResourceMetrics{{SchemaUrl: "aaaaaaa"}}}
	bts, err := proto.Marshal(metrics)
	if err != nil {
		t.Errorf("err=%v", err)
		return
	}

	tests := []struct {
		name    string
		args    args
		want    *collectormetricpb.ExportMetricsServiceRequest
		wantErr bool
	}{
		{
			name: "case marshal",
			args: args{
				rawRequest:  bts,
				contentType: "application/x-protobuf",
			},
			want:    metrics,
			wantErr: false,
		},
		{
			name: "case marshal err",
			args: args{
				rawRequest:  bts,
				contentType: "application/txt",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshalMetricsRequest(tt.args.rawRequest, tt.args.contentType)
			if (err != nil) != tt.wantErr {
				t.Errorf("unmarshalMetricsRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				// 有错误时 返回的对象为空
				return
			}
			gbts, _ := proto.Marshal(got)
			if !bytes.Equal(gbts, tt.args.rawRequest) {
				t.Errorf("byte not equal")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unmarshalMetricsRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_otlpHTTPCollector_apiOtlpTrace1 :
// 与上一个测试不同的是，这个测试用例是用来测试覆盖率，覆盖到每一个 if/else 中。
func Test_otlpHTTPCollector_apiOtlpTrace1(t *testing.T) {
	type fields struct {
		storage         *collector.SpansStorage
		Enable          bool
		HTTPStatusOK    int
		ExpectedHeaders map[string]string

		addr        string
		pattern     string
		contentType string
		byteBuf     io.Reader
	}
	type args struct {
		wantCode int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "storage_is_nil",
			fields: fields{
				storage:         nil,
				Enable:          true,
				HTTPStatusOK:    200,
				ExpectedHeaders: nil,
				addr:            "127.0.0.1:8888",
				pattern:         "/testStorageIsNil",
				contentType:     "application/x-protobuf",
				byteBuf:         strings.NewReader("name=cjb"),
			},
			args: args{wantCode: 500},
		},
		{
			name: "check_header",
			fields: fields{
				storage:         collector.NewSpansStorage(),
				Enable:          true,
				HTTPStatusOK:    200,
				ExpectedHeaders: map[string]string{"header1": "1"},
				addr:            "127.0.0.1:8889",
				pattern:         "/checkHeader",
				contentType:     "application/x-protobuf",
				byteBuf:         strings.NewReader("name=cjb"),
			},
			args: args{wantCode: 400},
		},
		{
			name: "request_body",
			fields: fields{
				storage:         collector.NewSpansStorage(),
				Enable:          true,
				HTTPStatusOK:    200,
				ExpectedHeaders: map[string]string{},
				addr:            "127.0.0.1:8890",
				pattern:         "/request_body",
				contentType:     "application/x-protobuf",
				byteBuf:         nil,
			},
			args: args{wantCode: 200},
		},
		{
			name: "bad_body",
			fields: fields{
				storage:         collector.NewSpansStorage(),
				Enable:          true,
				HTTPStatusOK:    200,
				ExpectedHeaders: map[string]string{},
				addr:            "127.0.0.1:8890",
				pattern:         "/bad_body",
				contentType:     "application/x-protobuf",
				byteBuf:         strings.NewReader("name=cjb"),
			},
			args: args{wantCode: 400},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// server 监听端口
			o := &otlpHTTPCollector{
				storage:         tt.fields.storage,
				Enable:          tt.fields.Enable,
				HTTPStatusOK:    tt.fields.HTTPStatusOK,
				ExpectedHeaders: tt.fields.ExpectedHeaders,
			}
			http.HandleFunc(tt.fields.pattern, o.apiOtlpTrace)

			server := http.Server{
				Addr: tt.fields.addr,
			}

			go func() {
				err := server.ListenAndServe()
				if err != nil {
					return
				}
			}()
			time.Sleep(time.Millisecond * 80) // wait server
			defer func() {
				server.Close()
				time.Sleep(time.Millisecond * 50) // wait ListenAndServe close
			}()
			// 创建 request POST
			resp, err := http.Post("http://"+tt.fields.addr+tt.fields.pattern,
				tt.fields.contentType,
				tt.fields.byteBuf)
			if err != nil {
				t.Errorf("post err=%v", err)
				return
			}

			if resp.StatusCode != tt.args.wantCode {
				t.Errorf("want code =%d response code =%d", tt.args.wantCode, resp.StatusCode)
				return
			}
			if tt.args.wantCode != 200 {
				return
			}
			defer resp.Body.Close()
			_, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("error =%v", err)
				return
			}
		})
	}
}

// Test_otlpHTTPCollector_apiOtlpTrace1 :
// 与上一个测试不同的是，这个测试用例是用来测试覆盖率，覆盖到每一个 if/else 中。
func Test_otlpHTTPCollector_apiOtlpMetric(t *testing.T) {
	type fields struct {
		storage         *collector.SpansStorage
		Enable          bool
		HTTPStatusOK    int
		ExpectedHeaders map[string]string

		addr        string
		pattern     string
		contentType string
		byteBuf     io.Reader
	}
	type args struct {
		wantCode int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "storage_is_nil",
			fields: fields{
				storage:         nil,
				Enable:          true,
				HTTPStatusOK:    200,
				ExpectedHeaders: nil,
				addr:            "127.0.0.1:8888",
				pattern:         "/metricStorageIsNil",
				contentType:     "application/x-protobuf",
				byteBuf:         strings.NewReader("name=cjb"),
			},
			args: args{wantCode: 500},
		},
		{
			name: "check_header",
			fields: fields{
				storage:         collector.NewSpansStorage(),
				Enable:          true,
				HTTPStatusOK:    200,
				ExpectedHeaders: map[string]string{"header1": "1"},
				addr:            "127.0.0.1:8889",
				pattern:         "/metric_checkHeader",
				contentType:     "application/x-protobuf",
				byteBuf:         strings.NewReader("name=cjb"),
			},
			args: args{wantCode: 400},
		},
		{
			name: "request_body",
			fields: fields{
				storage:         collector.NewSpansStorage(),
				Enable:          true,
				HTTPStatusOK:    200,
				ExpectedHeaders: map[string]string{},
				addr:            "127.0.0.1:8890",
				pattern:         "/metric_request_body",
				contentType:     "application/x-protobuf",
				byteBuf:         nil,
			},
			args: args{wantCode: 200},
		},
		{
			name: "bad_body",
			fields: fields{
				storage:         collector.NewSpansStorage(),
				Enable:          true,
				HTTPStatusOK:    200,
				ExpectedHeaders: map[string]string{},
				addr:            "127.0.0.1:8890",
				pattern:         "/metricBad_body",
				contentType:     "application/x-protobuf",
				byteBuf:         strings.NewReader("name=cjb"),
			},
			args: args{wantCode: 400},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// server 监听端口
			o := &otlpHTTPCollector{
				storage:         tt.fields.storage,
				Enable:          tt.fields.Enable,
				HTTPStatusOK:    tt.fields.HTTPStatusOK,
				ExpectedHeaders: tt.fields.ExpectedHeaders,
			}
			http.HandleFunc(tt.fields.pattern, o.apiOtlpMetric)

			server := http.Server{
				Addr: tt.fields.addr,
			}

			go func() {
				err := server.ListenAndServe()
				if err != nil {
					return
				}
			}()
			time.Sleep(time.Millisecond * 80) // wait server
			defer func() {
				server.Close()
				time.Sleep(time.Millisecond * 50) // wait ListenAndServe close
			}()
			// 创建 request POST
			resp, err := http.Post("http://"+tt.fields.addr+tt.fields.pattern,
				tt.fields.contentType,
				tt.fields.byteBuf)
			if err != nil {
				t.Errorf("post err=%v", err)
				return
			}

			if resp.StatusCode != tt.args.wantCode {
				t.Errorf("want code =%d response code =%d", tt.args.wantCode, resp.StatusCode)
				return
			}
			if tt.args.wantCode != 200 {
				return
			}
			defer resp.Body.Close()
			_, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("error =%v", err)
				return
			}
		})
	}
}

func Test_unmarshalTraceRequest1(t *testing.T) {
	jsonBody := `{
	"resourceSpans": [{
		"resource": {
			"attributes": [{
				"key": "service.name",
				"value": {
					"stringValue": "front-app"
				}
			}, {
				"key": "telemetry.sdk.language",
				"value": {
					"stringValue": "webjs"
				}
			}, {
				"key": "telemetry.sdk.name",
				"value": {
					"stringValue": "opentelemetry"
				}
			}, {
				"key": "telemetry.sdk.version",
				"value": {
					"stringValue": "1.2.0"
				}
			}],
			"droppedAttributesCount": 0
		},
		"instrumentationLibrarySpans": [{
			"spans": [{
				"traceId": "b974aa3f8e95387f959024e0472c62d5",
				"spanId": "bd1b8a16de09d8fe",
				"name": "files-series-info-0",
				"kind": 1,
				"startTimeUnixNano": 1653030257075199700,
				"endTimeUnixNano": 1653030257141699800,
				"attributes": [],
				"droppedAttributesCount": 0,
				"events": [{
					"timeUnixNano": 1653030257141599700,
					"name": "fetching-span1-completed",
					"attributes": [],
					"droppedAttributesCount": 0
				}],
				"droppedEventsCount": 0,
				"status": {
					"code": 0
				},
				"links": [],
				"droppedLinksCount": 0
			}],
			"instrumentationLibrary": {
				"name": "example-tracer-web"
			}
		}]
	}]
}`
	type args struct {
		rawRequest  []byte
		contentType string
	}
	tests := []struct {
		name    string
		args    args
		want    *collectortracepb.ExportTraceServiceRequest
		wantErr bool
	}{
		{
			name: "json",
			args: args{
				rawRequest:  []byte(jsonBody),
				contentType: "application/json",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshalTraceRequest(tt.args.rawRequest, tt.args.contentType)
			if (err != nil) != tt.wantErr {
				t.Errorf("unmarshalTraceRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				t.Logf("%+v", got)
				t.Logf("%+v", got.ResourceSpans[0])
			} else {
				return
			}

			if len(got.ResourceSpans) != 1 && len(got.ResourceSpans[0].Resource.Attributes) != 4 {
				t.Errorf("json marshal error")
			}
		})
	}
}
