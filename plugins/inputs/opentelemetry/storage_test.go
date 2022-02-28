package opentelemetry

import (
	"reflect"
	"testing"
	"time"

	DKtrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	resouecepb "go.opentelemetry.io/proto/otlp/resource/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

type mockStorage struct {
	SpansStorage
}

func newMockStorage() mockStorage {
	return mockStorage{
		SpansStorage{
			rsm:         make([]DKtrace.DatakitTrace, 0),
			otelMetrics: make([]*otelResourceMetric, 0),
			max:         make(chan int, 1),
		},
	}
}

var mockResourceSpan = &tracepb.ResourceSpans{
	Resource: &resouecepb.Resource{
		Attributes:             testKV,
		DroppedAttributesCount: 0,
	},
	SchemaUrl: "",
	InstrumentationLibrarySpans: []*tracepb.InstrumentationLibrarySpans{
		{
			InstrumentationLibrary: &commonpb.InstrumentationLibrary{
				Name:    "test-tracer",
				Version: "",
			},
			Spans: []*tracepb.Span{
				{
					Name: "span_name_a",
				},
				{
					Name: "span_name_b",
				},
				{
					Name: "span_name_c",
				},
			},
			SchemaUrl: "",
		},
	},
}

func TestSpansStorage_AddSpans(t *testing.T) {
	type args struct {
		rss []*tracepb.ResourceSpans
	}
	tests := []struct {
		name   string
		fields mockStorage
		args   args
	}{
		{name: "case1", fields: newMockStorage(), args: args{rss: []*tracepb.ResourceSpans{mockResourceSpan}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SpansStorage{
				rsm:         tt.fields.rsm,
				otelMetrics: tt.fields.otelMetrics,
				Count:       tt.fields.Count,
				max:         tt.fields.max,
				stop:        tt.fields.stop,
			}
			s.AddSpans(tt.args.rss)
			if s.Count != 3 {
				t.Errorf("span count not 3")
			}
		})
	}
}

func TestNewSpansStorage(t *testing.T) {
	tests := []struct {
		name string
		want SpansStorage
	}{
		{name: "new", want: SpansStorage{rsm: make([]DKtrace.DatakitTrace, 0),
			otelMetrics: make([]*otelResourceMetric, 0),
			max:         make(chan int, 1)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSpansStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSpansStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpansStorage_getCount(t *testing.T) {
	tests := []struct {
		name   string
		fields mockStorage
		want   int
	}{
		{
			name:   "case",
			fields: mockStorage{SpansStorage{Count: 1}},
			want:   1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SpansStorage{
				rsm:         tt.fields.rsm,
				otelMetrics: tt.fields.otelMetrics,
				Count:       tt.fields.Count,
				max:         tt.fields.max,
				stop:        tt.fields.stop,
			}
			if got := s.getCount(); got != tt.want {
				t.Errorf("getCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpansStorage_getDKTrace(t *testing.T) {
	type fields struct {
		rsm         []DKtrace.DatakitTrace
		otelMetrics []*otelResourceMetric
		Count       int
		max         chan int
		stop        chan struct{}
	}
	tests := []struct {
		name   string
		fields fields
		want   []DKtrace.DatakitTrace
	}{
		{
			name: "case",
			fields: fields{
				rsm: []DKtrace.DatakitTrace{
					{&DKtrace.DatakitSpan{TraceID: "000001"}},
					{&DKtrace.DatakitSpan{TraceID: "000002"}},
					{&DKtrace.DatakitSpan{TraceID: "000003"}},
				},
				otelMetrics: nil,
				Count:       0,
				max:         nil,
				stop:        nil,
			},
			want: []DKtrace.DatakitTrace{
				{&DKtrace.DatakitSpan{TraceID: "000001"}},
				{&DKtrace.DatakitSpan{TraceID: "000002"}},
				{&DKtrace.DatakitSpan{TraceID: "000003"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SpansStorage{
				rsm:         tt.fields.rsm,
				otelMetrics: tt.fields.otelMetrics,
				Count:       tt.fields.Count,
				max:         tt.fields.max,
				stop:        tt.fields.stop,
			}
			if got := s.getDKTrace(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDKTrace() = %v, want %v", got, tt.want)
			}
			t.Logf("storage.rsm len=%d", len(s.rsm))
		})
	}
}

func TestSpansStorage_run(t *testing.T) {
	type fields struct {
		rsm         []DKtrace.DatakitTrace
		otelMetrics []*otelResourceMetric
		Count       int
		max         chan int
		stop        chan struct{}
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "case1",
			fields: fields{
				rsm:         make([]DKtrace.DatakitTrace, 0),
				otelMetrics: make([]*otelResourceMetric, 0),
				max:         make(chan int, 1),
				stop:        make(chan struct{}, 1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SpansStorage{
				rsm:         tt.fields.rsm,
				otelMetrics: tt.fields.otelMetrics,
				Count:       tt.fields.Count,
				max:         tt.fields.max,
				stop:        tt.fields.stop,
			}
			go s.run()
			go func() {
				s.stop <- struct{}{}
				t.Log("set stop")
			}()
			time.Sleep(time.Millisecond * 200) // wait s.stop()
			t.Log("wait stop")
			if res, ok := <-s.stop; ok {
				t.Errorf("not close res=%v", res)
			}
		})
	}
}
