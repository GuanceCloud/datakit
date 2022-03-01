package collector

import (
	"context"
	"reflect"
	"testing"
	"time"

	DKtrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

var testKV = []*commonpb.KeyValue{
	{
		Key: "service.name",
		Value: &commonpb.AnyValue{
			Value: &commonpb.AnyValue_StringValue{
				StringValue: "service"},
		},
	},
	{
		Key: "http.method",
		Value: &commonpb.AnyValue{
			Value: &commonpb.AnyValue_StringValue{
				StringValue: "POST"},
		},
	},
	{
		Key: "http.status_code",
		Value: &commonpb.AnyValue{
			Value: &commonpb.AnyValue_IntValue{
				IntValue: 200,
			},
		},
	},
	{
		Key: "container.name",
		Value: &commonpb.AnyValue{
			Value: &commonpb.AnyValue_StringValue{
				StringValue: "hostName",
			},
		},
	},
	{
		Key: "process.pid",
		Value: &commonpb.AnyValue{
			Value: &commonpb.AnyValue_IntValue{
				IntValue: 2222,
			},
		},
	},
}

var allTag = map[string]string{
	"service.name":     "service",
	"http.method":      "POST",
	"http.status_code": "200",
	"container.name":   "hostName",
	"process.pid":      "2222",
}

func Test_mkDKTrace(t *testing.T) {
	/*
		mock server

		mock client 发送 readOnlySpans

		从export中获取 ResourceSpans

	*/
	trace := &MockTrace{}
	endpoint := "localhost:20010"
	m := MockOtlpGrpcCollector{Trace: trace}
	go m.StartServer(t, endpoint)
	<-time.After(5 * time.Millisecond)
	t.Log("start server")

	ctx := context.Background()
	exp := NewGRPCExporter(t, ctx, endpoint)

	roSpans, want := mockRoSpans(t)
	if err := exp.ExportSpans(ctx, roSpans); err != nil {
		t.Errorf("err=%v", err)
		return
	}
	time.Sleep(time.Millisecond * 40) // wait MockTrace
	rss := trace.getResourceSpans()
	type args struct {
		rss []*tracepb.ResourceSpans
	}
	tests := []struct {
		name string
		args args
		want []DKtrace.DatakitTrace
	}{
		{name: "case1", args: args{rss: rss}, want: want},
	}
	storage := &SpansStorage{RegexpString: "", GlobalTags: map[string]string{}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := storage.mkDKTrace(tt.args.rss)
			if !reflect.DeepEqual(got[0][0], tt.want[0][0]) {
				t.Errorf("mkDKTrace() = %+v,\n want %+v", got[0][0], tt.want[0][0])
			}
		})
	}
}

func Test_byteToString(t *testing.T) {
	type args struct {
		bts []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "nil", args: args{bts: []byte{}}, want: "0"},
		{name: "100", args: args{bts: []byte{1, 0, 0}}, want: "010000"},
		{name: "a1", args: args{bts: []byte{0xa1}}, want: "a1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := byteToString(tt.args.bts); got != tt.want {
				t.Errorf("byteToInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dkTags_addGlobalTags(t *testing.T) {
	type fields struct {
		tags       map[string]string
		globalTags map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		want   *dkTags
	}{
		{
			name:   "add a:b",
			fields: fields{tags: map[string]string{}, globalTags: map[string]string{"globalTag_a": "b"}},
			want:   &dkTags{replaceTags: map[string]string{"globalTag_a": "b"}, tags: map[string]string{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := &dkTags{
				tags:        tt.fields.tags,
				replaceTags: map[string]string{},
			}
			if got := dt.addGlobalTags(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("addGlobalTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dkTags_addOtherTags(t *testing.T) {
	type fields struct {
		tags map[string]string
	}
	type args struct {
		span *tracepb.Span
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *dkTags
	}{
		{
			name: "case",
			fields: fields{
				tags: map[string]string{},
			},
			args: args{
				span: &tracepb.Span{
					DroppedEventsCount: 1,                                               // drop event count = 1
					Events:             []*tracepb.Span_Event{{Name: "1"}, {Name: "1"}}, // events = 2
					Links:              []*tracepb.Span_Link{{TraceState: "1"}},         // links = 1
				}},
			want: &dkTags{
				tags: map[string]string{"links_count": "1", "events_count": "2", "dropped_events_count": "1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := &dkTags{
				tags: tt.fields.tags,
			}
			if got := dt.addOtherTags(tt.args.span); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("addOtherTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dkTags_checkAllTagsKey(t *testing.T) {
	type fields struct {
		tags map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		want   *dkTags
	}{
		{
			name:   "case",
			fields: fields{tags: map[string]string{"a.b": "c"}},
			want: &dkTags{
				tags:        map[string]string{"a.b": "c"},
				replaceTags: map[string]string{"a_b": "c"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := &dkTags{
				tags:        tt.fields.tags,
				replaceTags: map[string]string{},
			}
			if got := dt.checkAllTagsKey(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("checkAllTagsKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dkTags_resource(t *testing.T) {
	type fields struct {
		replaceTags map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]string
	}{
		{
			name:   "get resource",
			fields: fields{replaceTags: map[string]string{"a": "b"}},
			want:   map[string]string{"a": "b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := &dkTags{
				replaceTags: tt.fields.replaceTags,
			}
			if got := dt.resource(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("resource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dkTags_setAttributesToTags(t *testing.T) {
	type fields struct {
		tags map[string]string
	}
	type args struct {
		attr []*commonpb.KeyValue
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *dkTags
	}{
		{
			name:   "case1",
			fields: fields{tags: map[string]string{}},
			args:   args{attr: testKV},
			want:   &dkTags{tags: allTag},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := &dkTags{
				tags: tt.fields.tags,
			}
			if got := dt.setAttributesToTags(tt.args.attr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("setAttributesToTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDKSpanStatus(t *testing.T) {
	type args struct {
		code *tracepb.Status
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case1",
			args: args{code: &tracepb.Status{Code: tracepb.Status_STATUS_CODE_UNSET}},
			want: "info"},

		{
			name: "case2",
			args: args{code: &tracepb.Status{Code: tracepb.Status_STATUS_CODE_OK}},
			want: "ok",
		},

		{
			name: "case3",
			args: args{code: &tracepb.Status{Code: tracepb.Status_STATUS_CODE_ERROR}},
			want: "error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDKSpanStatus(tt.args.code); got != tt.want {
				t.Errorf("getDKSpanStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newEmptyTags(t *testing.T) {
	type args struct {
		regexp     string
		globalTags map[string]string
	}
	tests := []struct {
		name string
		args args
		want *dkTags
	}{
		{
			name: "case",
			args: args{regexp: "", globalTags: map[string]string{}},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newEmptyTags(tt.args.regexp, tt.args.globalTags); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newEmptyTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_replace(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "case", args: args{key: "mysql.select"}, want: "mysql_select"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := replace(tt.args.key); got != tt.want {
				t.Errorf("replace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dkTags_getAttributeVal(t *testing.T) {
	type fields struct {
		tags        map[string]string
		replaceTags map[string]string
	}
	type args struct {
		keyName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "case1",
			fields: fields{tags: allTag},
			args:   args{keyName: otelResourceServiceKey},
			want:   "service",
		},
		{
			name:   "case2",
			fields: fields{tags: allTag},
			args:   args{keyName: otelResourceHTTPMethodKey},
			want:   "POST",
		},
		{
			name:   "case3",
			fields: fields{tags: allTag},
			args:   args{keyName: otelResourceContainerNameKey},
			want:   "hostName",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := &dkTags{
				tags:        tt.fields.tags,
				replaceTags: tt.fields.replaceTags,
			}
			if got := dt.getAttributeVal(tt.args.keyName); got != tt.want {
				t.Errorf("getAttributeVal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dkTags_checkCustomTags1(t *testing.T) {
	type fields struct {
		tags        map[string]string
		replaceTags map[string]string
	}
	tests := []struct {
		name      string
		regexpStr string
		fields    fields
		want      *dkTags
	}{
		{
			name:      "regexp-1",
			regexpStr: "os_*|process_*",
			fields: fields{
				tags:        map[string]string{},
				replaceTags: map[string]string{"os_name": "linux", "other_key": "other_value"},
			},
			want: &dkTags{
				tags:        map[string]string{},
				replaceTags: map[string]string{"other_key": "other_value"},
			},
		},

		{
			name:      "regexp-2",
			regexpStr: "os_*|process_*",
			fields: fields{
				tags:        map[string]string{},
				replaceTags: map[string]string{"os_name": "linux", "process_id": "123", "other_key": "other_value"},
			},
			want: &dkTags{
				tags:        map[string]string{},
				replaceTags: map[string]string{"other_key": "other_value"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := &dkTags{
				tags:        tt.fields.tags,
				replaceTags: tt.fields.replaceTags,
			}
			dt.regexpString = tt.regexpStr
			if got := dt.checkCustomTags(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("checkCustomTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
