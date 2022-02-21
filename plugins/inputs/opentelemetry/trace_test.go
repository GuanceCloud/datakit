package opentelemetry

import (
	"context"
	"reflect"
	"testing"
	"time"

	DKtrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

func Test_mkDKTrace(t *testing.T) {
	/*
		mock server

		mock client 发送 readOnlySpans

		从export中获取 ResourceSpans

	*/
	trace := &MockTrace{}
	endpoint := "localhost:20010"
	m := MockOtlpGrpcCollector{trace: trace}
	go m.startServer(t, endpoint)
	<-time.After(5 * time.Millisecond)
	t.Log("start server")

	ctx := context.Background()
	exp := newGRPCExporter(t, ctx, endpoint)

	roSpans, want := mockRoSpans(t)
	if err := exp.ExportSpans(ctx, roSpans); err != nil {
		t.Fatalf("err=%v", err)
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
		{name: "case1", args: args{
			rss: rss,
		}, want: want},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mkDKTrace(tt.args.rss)
			if !reflect.DeepEqual(got[0][0], tt.want[0][0]) {
				t.Errorf("mkDKTrace() = %+v,\n want %+v", got[0][0], tt.want[0][0])
			}

		})
	}
}
