package skywalking

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"

	common "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/v3/skywalking/network/common/v3"
	lang "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/v3/skywalking/network/language/agent/v3"
	mgment "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/v3/skywalking/network/management/v3"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
	"google.golang.org/grpc"
)

func skyWalkingV3ServervRun(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorf("start skywalking V3 grpc server %s failed: %v", addr, err)

		return
	}
	log.Infof("skywalking v3 listening on: %s", addr)

	srv := grpc.NewServer()
	lang.RegisterTraceSegmentReportServiceServer(srv, &SkyWalkingServerV3{})
	lang.RegisterJVMMetricReportServiceServer(srv, &SkyWalkingJVMMetricReportServerV3{})
	mgment.RegisterManagementServiceServer(srv, &SkyWalkingManagementServerV3{})
	if err := srv.Serve(listener); err != nil {
		log.Error(err)
	}
	log.Info("skywalking v3 exits")
}

type SkyWalkingServerV3 struct {
	lang.UnimplementedTraceSegmentReportServiceServer
}

func (s *SkyWalkingServerV3) Collect(tsc lang.TraceSegmentReportService_CollectServer) (err error) {
	defer func() {
		if err != nil {
			log.Error(err)
		}
	}()

	for {
		segobj, err := tsc.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return tsc.SendAndClose(&common.Commands{})
			}
			log.Error(err.Error())

			return err
		}

		log.Debug("segment received")

		group, err := segobjToAdapters(segobj)
		if err != nil {
			log.Error(err.Error())

			return err
		}

		if len(group) != 0 {
			trace.MkLineProto(group, inputName)
		} else {
			log.Debug("empty v3 segment")
		}
	}
}

func segobjToAdapters(segment *lang.SegmentObject) ([]*trace.TraceAdapter, error) {
	var group []*trace.TraceAdapter
	for _, span := range segment.Spans {
		adapter := &trace.TraceAdapter{Source: inputName}
		adapter.Duration = (span.EndTime - span.StartTime) * 1000000
		adapter.Start = span.StartTime * 1000000
		js, err := json.Marshal(span)
		if err != nil {
			return nil, err
		}
		adapter.Content = string(js)
		adapter.ServiceName = segment.Service
		adapter.OperationName = span.OperationName
		if span.SpanType == lang.SpanType_Entry {
			if len(span.Refs) > 0 {
				adapter.ParentID = fmt.Sprintf("%s%d", span.Refs[0].ParentTraceSegmentId,
					span.Refs[0].ParentSpanId)
			}
		} else {
			adapter.ParentID = fmt.Sprintf("%s%d", segment.TraceSegmentId, span.ParentSpanId)
		}

		adapter.TraceID = segment.TraceId
		adapter.SpanID = fmt.Sprintf("%s%d", segment.TraceSegmentId, span.SpanId)
		adapter.Status = trace.STATUS_OK
		if span.IsError {
			adapter.Status = trace.STATUS_ERR
		}

		switch span.SpanType {
		case lang.SpanType_Entry:
			adapter.SpanType = trace.SPAN_TYPE_ENTRY
		case lang.SpanType_Local:
			adapter.SpanType = trace.SPAN_TYPE_LOCAL
		case lang.SpanType_Exit:
			adapter.SpanType = trace.SPAN_TYPE_EXIT
		default:
			log.Warnf("unknown span type %d, use SPAN_TYPE_EXIT", span.SpanType)
			adapter.SpanType = trace.SPAN_TYPE_EXIT
		}

		adapter.EndPoint = span.Peer
		adapter.Tags = skywalkingV3Tags

		group = append(group, adapter)
	}

	return group, nil
}

type SkyWalkingManagementServerV3 struct {
	mgment.UnimplementedManagementServiceServer
}

func (*SkyWalkingManagementServerV3) ReportInstanceProperties(ctx context.Context,
	mng *mgment.InstanceProperties) (*common.Commands, error) {
	var kvpStr string
	cmd := &common.Commands{}

	for _, kvp := range mng.Properties {
		kvpStr += fmt.Sprintf("[%v:%v]", kvp.Key, kvp.Value)
	}
	log.Debugf("ReportInstanceProperties service:%v instance:%v properties:%v", mng.Service, mng.ServiceInstance, kvpStr)

	return cmd, nil
}

func (*SkyWalkingManagementServerV3) KeepAlive(ctx context.Context,
	ping *mgment.InstancePingPkg) (*common.Commands, error) {
	cmd := &common.Commands{}
	log.Debugf("KeepAlive service:%v instance:%v", ping.Service, ping.ServiceInstance)

	return cmd, nil
}

type SkyWalkingJVMMetricReportServerV3 struct {
	lang.UnimplementedJVMMetricReportServiceServer
}

func (*SkyWalkingJVMMetricReportServerV3) Collect(ctx context.Context,
	jvm *lang.JVMMetricCollection) (*common.Commands, error) {
	cmd := &common.Commands{}
	log.Debugf("JVMMetricReportService service:%v instance:%v", jvm.Service, jvm.ServiceInstance)

	return cmd, nil
}
