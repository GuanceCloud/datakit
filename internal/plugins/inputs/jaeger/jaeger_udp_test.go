// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jaeger

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"github.com/uber/jaeger-client-go/thrift"
	"github.com/uber/jaeger-client-go/thrift-gen/agent"
	"github.com/uber/jaeger-client-go/thrift-gen/jaeger"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

type tAfterGather struct {
	t   *testing.T
	pts chan *point.Point
}

func newtAfterGather(t *testing.T) *tAfterGather {
	t.Helper()
	return &tAfterGather{
		t:   t,
		pts: make(chan *point.Point, 10),
	}
}

func (t *tAfterGather) Run(inputName string, dktraces itrace.DatakitTraces) {
	for _, dktrace := range dktraces {
		for _, span := range dktrace {
			t.pts <- span.Point
		}
	}
}

var protocolMarshalers = map[string]func(t *testing.T, batch *jaeger.Batch) []byte{
	CompactProtocol: compactBatch,
	BinaryProtocol:  binaryBatch,
}

func binaryBatch(t *testing.T, batch *jaeger.Batch) []byte {
	t.Helper()
	transport := thrift.NewTMemoryBufferLen(1000)
	protocolFactory := thrift.NewTBinaryProtocolFactoryConf(&thrift.TConfiguration{})
	client := agent.NewAgentClientFactory(transport, protocolFactory)

	err := client.EmitBatch(context.Background(), batch)
	if err != nil {
		t.Errorf("emitBatch err=%v", err)
		return transport.Bytes()
	}

	return transport.Bytes()
}

func compactBatch(t *testing.T, batch *jaeger.Batch) []byte {
	t.Helper()
	transport := thrift.NewTMemoryBufferLen(1000)
	protocolFactory := thrift.NewTCompactProtocolFactoryConf(&thrift.TConfiguration{})
	client := agent.NewAgentClientFactory(transport, protocolFactory)

	err := client.EmitBatch(context.Background(), batch)
	if err != nil {
		t.Errorf("emitBatch err=%v", err)
		return transport.Bytes()
	}
	t.Logf("transport=%s", transport.String())
	return transport.Bytes()
}

func mockBatch(spanCount int, tags map[string]string) *jaeger.Batch {
	p := jaeger.NewProcess()
	spanTags := make([]*jaeger.Tag, 0)
	for k, v := range tags {
		vp := v
		spanTags = append(spanTags, &jaeger.Tag{Key: k, VType: 0, VStr: &vp})
	}
	p.ServiceName = "test_service"
	project := "project"
	p.Tags = []*jaeger.Tag{{Key: "project", VType: 0, VStr: &project}}
	spans := make([]*jaeger.Span, 0)
	for i := 0; i < spanCount; i++ {
		span := &jaeger.Span{
			TraceIdLow:    12345678,
			TraceIdHigh:   987654,
			SpanId:        987654321,
			ParentSpanId:  0,
			OperationName: "test_span_" + strconv.Itoa(i),
			References:    nil,
			Flags:         0,
			StartTime:     2345,
			Duration:      1000,
			Tags:          spanTags,
			Logs:          nil,
		}
		spans = append(spans, span)
	}
	nn := int64(16)
	batch := &jaeger.Batch{
		Process: p,
		Spans:   spans,
		SeqNo:   &nn,
		Stats: &jaeger.ClientStats{
			FullQueueDroppedSpans: nn,
			TooLargeDroppedSpans:  nn,
			FailedToEmitSpans:     nn,
		},
	}

	return batch
}

func send(t *testing.T, addr string, bts []byte) {
	t.Helper()
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return
	}

	n, err := conn.Write(bts)
	if err != nil {
		t.Error(err)
		return
	}
	conn.Close()
	t.Logf("send to udp conn bts len=%d", n)
}

func TestStartUDPAgent(t *testing.T) {
	type args struct {
		protocol  string
		addr      string
		spanCount int
		semStop   *cliutils.Sem
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test_" + BinaryProtocol,
			args: args{
				protocol:  BinaryProtocol,
				addr:      "127.0.0.1:16832",
				spanCount: 10,
				semStop:   cliutils.NewSem(),
			},
		},
		{
			name: "test_" + CompactProtocol,
			args: args{
				protocol:  CompactProtocol,
				addr:      "127.0.0.1:16831",
				spanCount: 5,
				semStop:   cliutils.NewSem(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("test for %s", tt.name)
			af := newtAfterGather(t)
			afterGatherRun = af
			go StartUDPAgent(tt.args.protocol, tt.args.addr, tt.args.semStop)
			time.Sleep(time.Second * 2)
			batch := mockBatch(tt.args.spanCount, map[string]string{"tag_a": "a"})
			bts := protocolMarshalers[tt.args.protocol](t, batch)
			send(t, tt.args.addr, bts)

			ticker := time.NewTicker(time.Second * 20)
			for i := 0; i < tt.args.spanCount; i++ {
				select {
				case p := <-af.pts:
					t.Logf("point trace_id =%s", p.Get("trace_id"))
					assert.EqualValues(t, p.GetTag(itrace.TagService), "test_service")
					assert.EqualValues(t, p.GetTag("tag_a"), "a")
					assert.Len(t, p.Get(itrace.FieldTraceID), 32)
					assert.Equal(t, p.Get(itrace.FieldParentID), "0")
					assert.Equal(t, p.Get(itrace.Project), "project")
					assert.Equal(t, p.Get(itrace.FieldSpanid), strconv.FormatUint(uint64(987654321), 16))
					assert.Equal(t, p.Get(itrace.FieldDuration), int64(1000))
				case <-ticker.C:
					t.Errorf("timeout! count:%d != spanCount:%d", i, tt.args.spanCount)
					tt.args.semStop.Close()
					return
				}
			}
		})
	}
}

func Test_EmitBatch(t *testing.T) {
	transport := thrift.NewTMemoryBufferLen(1000)
	protocolFactory := thrift.NewTBinaryProtocolFactoryConf(&thrift.TConfiguration{})
	client := agent.NewAgentClientFactory(transport, protocolFactory)

	batch := &jaeger.Batch{
		Process: &jaeger.Process{
			ServiceName: "service_name",
		},
		Spans: []*jaeger.Span{
			{
				TraceIdLow:    1111111,
				TraceIdHigh:   222222222,
				SpanId:        3333333,
				ParentSpanId:  0,
				OperationName: "service_name",
				References:    nil,
				Flags:         0,
				StartTime:     1234567,
				Duration:      0,
				Logs:          nil,
			},
		},
	}

	err := client.EmitBatch(context.Background(), batch)
	if err != nil {
		t.Errorf("emitBatch err=%v", err)
		return
	}

	tmbuf := thrift.NewTMemoryBufferLen(len(transport.Bytes()))
	_, err = tmbuf.Write(transport.Bytes())
	if err != nil {
		t.Errorf("write err=%v", err)
		return
	}

	tprot := thrift.NewTBinaryProtocolFactoryConf(&thrift.TConfiguration{}).GetProtocol(tmbuf)

	ctx := context.Background()

	if _, _, _, err = tprot.ReadMessageBegin(ctx); err != nil { //nolint:dogsled
		t.Errorf("read message err=%v", err)
	}
	defer func() {
		if err := tprot.ReadMessageEnd(ctx); err != nil {
			log.Error("read message end failed :%s,", err.Error())
		}
	}()

	batch1 := &agent.AgentEmitBatchArgs{}
	if err = batch1.Read(ctx, tprot); err != nil {
		t.Errorf("read message err=%v", err)
		return
	}
	t.Logf("batch = %s", batch1.Batch.String())
}
