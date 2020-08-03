package trace

import (
	"fmt"
	"io"
	"net"
	"encoding/json"

	"google.golang.org/grpc"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	swV3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace/skywalking/v3"
)

type SkywalkingServer struct {}

func (s *SkywalkingServer) Collect(tsc swV3.TraceSegmentReportService_CollectServer) error {
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

func SkyWalkingServer(addr string) {
	log.Infof("trace gRPC starting...")

	rpcListener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorf("start gRPC server %s failed: %v", addr, err)
		return
	}

	log.Infof("start gRPC server on %s ok", addr)

	rpcServer := grpc.NewServer()
	swV3.RegisterTraceSegmentReportServiceServer(rpcServer, &SkywalkingServer{})

	go func() {
		if err := rpcServer.Serve(rpcListener); err != nil {
			log.Error(err)
		}

		log.Info("gRPC server exit")
	}()

	<-datakit.Exit.Wait()
	log.Info("stopping gRPC server...")
	rpcServer.Stop()

	log.Info("gRPC server exit")
	return
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
		} else {
			t.spanType  = SPAN_TYPE_LOCAL
		}
		t.endPoint      = span.Peer

		t.mkLineProto()
	}
	return nil
}
