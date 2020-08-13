package trace

import (
	"io"
	"fmt"
	"encoding/json"

	swV3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace/skywalking/v3"
)

type SkywalkingServerV3 struct {}

func (s *SkywalkingServerV3) Collect(tsc swV3.TraceSegmentReportService_CollectServer) error {
	for {
		sgo, err := tsc.Recv()
		if err == io.EOF {
			return tsc.SendAndClose(&swV3.Commands{})
		}
		if err != nil {
			return err
		}
		err = skywalkGrpcToLineProto(sgo)
		if err != nil {
			return err
		}
	}
	return nil
}

func skywalkGrpcToLineProto(sg *swV3.SegmentObject) error {
	for _, span := range sg.Spans {
		t := TraceAdapter{}

		t.source = "skywalking"

		t.duration = (span.EndTime -span.StartTime)*1000
		t.timestampUs = span.StartTime * 1000
		js ,err := json.Marshal(span)
		if err != nil {
			return err
		}
		t.content = string(js)
		t.class         = "tracing"
		t.serviceName   = sg.Service
		t.operationName = span.OperationName
		if span.SpanType == swV3.SpanType_Entry {
			if len(span.Refs) > 0 {
				t.parentID      = fmt.Sprintf("%s%d", span.Refs[0].ParentTraceSegmentId,
					span.Refs[0].ParentSpanId)
			}
		} else {
			t.parentID      = fmt.Sprintf("%s%d", sg.TraceSegmentId, span.ParentSpanId)
		}

		t.traceID       = sg.TraceId
		t.spanID        = fmt.Sprintf("%s%d", sg.TraceSegmentId, span.SpanId)
		if span.IsError {
			t.isError   = "true"
		}
		if span.SpanType == swV3.SpanType_Entry {
			t.spanType  = SPAN_TYPE_ENTRY
		} else if span.SpanType == swV3.SpanType_Local {
			t.spanType  = SPAN_TYPE_LOCAL
		} else {
			t.spanType  = SPAN_TYPE_EXIT
		}
		t.endPoint      = span.Peer

		t.mkLineProto()
	}
	return nil
}
