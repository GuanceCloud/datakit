// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package skywalking

import (
	"context"
	"errors"
	"io"
	"net"

	commonv3old "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/v8.3.0/common/v3"
	agentv3old "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/v8.3.0/language/agent/v3"
	profilev3old "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/v8.3.0/language/profile/v3"
	mgmtv3old "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/v8.3.0/management/v3"

	configv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/v9.3.0/agent/configuration/v3"
	commonv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/v9.3.0/common/v3"
	eventv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/v9.3.0/event/v3"
	agentv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/v9.3.0/language/agent/v3"
	profilev3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/v9.3.0/language/profile/v3"
	loggingv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/v9.3.0/logging/v3"
	mgmtv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/v9.3.0/management/v3"
	"google.golang.org/grpc"
)

func runGRPCV3(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorf("### skywalking grpc server v3 listening on %s failed: %v", addr, err)

		return
	}
	log.Debugf("### skywalking grpc v3 listening on: %s", addr)

	skySvr = grpc.NewServer()
	// register API version 8.3.0
	agentv3old.RegisterTraceSegmentReportServiceServer(skySvr, &TraceReportServerV3Old{})
	agentv3old.RegisterJVMMetricReportServiceServer(skySvr, &JVMMetricReportServerV3Old{})
	profilev3old.RegisterProfileTaskServer(skySvr, &ProfileTaskServerV3Old{})
	mgmtv3old.RegisterManagementServiceServer(skySvr, &ManagementServerV3Old{})
	// register API version 9.3.0
	agentv3.RegisterTraceSegmentReportServiceServer(skySvr, &TraceReportServerV3{})
	eventv3.RegisterEventServiceServer(skySvr, &EventServerV3{})
	agentv3.RegisterJVMMetricReportServiceServer(skySvr, &JVMMetricReportServerV3{})
	loggingv3.RegisterLogReportServiceServer(skySvr, &LoggingServerV3{})
	profilev3.RegisterProfileTaskServer(skySvr, &ProfileTaskServerV3{})
	mgmtv3.RegisterManagementServiceServer(skySvr, &ManagementServerV3{})
	configv3.RegisterConfigurationDiscoveryServiceServer(skySvr, &DiscoveryServerV3{})

	if err = skySvr.Serve(listener); err != nil {
		log.Error(err.Error())
	}

	log.Debug("### skywalking grpc v3 exits")
}

type TraceReportServerV3Old struct {
	agentv3old.UnimplementedTraceSegmentReportServiceServer
}

func (*TraceReportServerV3Old) Collect(tsr agentv3old.TraceSegmentReportService_CollectServer) error {
	for {
		segobj, err := tsr.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return tsr.SendAndClose(&commonv3old.Commands{})
			}
			log.Error(err.Error())

			return err
		}
		log.Debugf("### TraceReportServerV3Old:Collect SegmentObject:%#v", segobj)

		newSegObj := agentv3.SegmentObject{
			TraceId:         segobj.TraceId,
			TraceSegmentId:  segobj.TraceSegmentId,
			Spans:           make([]*agentv3.SpanObject, len(segobj.Spans)),
			Service:         segobj.Service,
			ServiceInstance: segobj.ServiceInstance,
			IsSizeLimited:   segobj.IsSizeLimited,
		}
		for i := range segobj.Spans {
			newSegObj.Spans[i] = &agentv3.SpanObject{
				SpanId:        segobj.Spans[i].SpanId,
				ParentSpanId:  segobj.Spans[i].ParentSpanId,
				StartTime:     segobj.Spans[i].StartTime,
				EndTime:       segobj.Spans[i].EndTime,
				Refs:          make([]*agentv3.SegmentReference, len(segobj.Spans[i].Refs)),
				OperationName: segobj.Spans[i].OperationName,
				Peer:          segobj.Spans[i].Peer,
				SpanType:      agentv3.SpanType(segobj.Spans[i].SpanType),
				SpanLayer:     agentv3.SpanLayer(segobj.Spans[i].SpanLayer),
				ComponentId:   segobj.Spans[i].ComponentId,
				IsError:       segobj.Spans[i].IsError,
				Tags:          make([]*commonv3.KeyStringValuePair, len(segobj.Spans[i].Tags)),
				Logs:          make([]*agentv3.Log, len(segobj.Spans[i].Logs)),
				SkipAnalysis:  segobj.Spans[i].SkipAnalysis,
			}
			for j := range segobj.Spans[i].Refs {
				newSegObj.Spans[i].Refs[j] = &agentv3.SegmentReference{
					RefType:                  agentv3.RefType(segobj.Spans[i].Refs[j].RefType),
					TraceId:                  segobj.Spans[i].Refs[j].TraceId,
					ParentTraceSegmentId:     segobj.Spans[i].Refs[j].ParentTraceSegmentId,
					ParentSpanId:             segobj.Spans[i].Refs[j].ParentSpanId,
					ParentService:            segobj.Spans[i].Refs[j].ParentService,
					ParentServiceInstance:    segobj.Spans[i].Refs[j].ParentServiceInstance,
					ParentEndpoint:           segobj.Spans[i].Refs[j].ParentEndpoint,
					NetworkAddressUsedAtPeer: segobj.Spans[i].Refs[j].NetworkAddressUsedAtPeer,
				}
			}
			for j := range segobj.Spans[i].Tags {
				newSegObj.Spans[i].Tags[j] = &commonv3.KeyStringValuePair{Key: segobj.Spans[i].Tags[j].Key, Value: segobj.Spans[i].Tags[j].Value}
			}
			for j := range segobj.Spans[i].Logs {
				newSegObj.Spans[i].Logs[j] = &agentv3.Log{
					Time: segobj.Spans[i].Logs[j].Time,
					Data: make([]*commonv3.KeyStringValuePair, len(segobj.Spans[i].Logs[j].Data)),
				}
				for k := range segobj.Spans[i].Logs[j].Data {
					newSegObj.Spans[i].Logs[j].Data[k] = &commonv3.KeyStringValuePair{
						Key:   segobj.Spans[i].Logs[j].Data[k].Key,
						Value: segobj.Spans[i].Logs[j].Data[k].Value,
					}
				}
			}
		}
		api.ProcessSegment(&newSegObj)
	}
}

func (*TraceReportServerV3Old) CollectInSync(ctx context.Context, seg *agentv3old.SegmentCollection) (*commonv3old.Commands, error) {
	log.Debugf("### TraceReportServerV3Old:CollectInSync SegmentCollection: %#v", seg)

	return &commonv3old.Commands{}, nil
}

type JVMMetricReportServerV3Old struct {
	agentv3old.UnimplementedJVMMetricReportServiceServer
}

func (*JVMMetricReportServerV3Old) Collect(ctx context.Context, jvm *agentv3old.JVMMetricCollection) (*commonv3old.Commands, error) {
	log.Debugf("### JVMMetricReportServerV3Old:Collect %#v", jvm)

	newJVM := agentv3.JVMMetricCollection{
		Metrics:         make([]*agentv3.JVMMetric, len(jvm.Metrics)),
		Service:         jvm.Service,
		ServiceInstance: jvm.ServiceInstance,
	}
	for i := range jvm.Metrics {
		newJVM.Metrics[i] = &agentv3.JVMMetric{
			Time:       jvm.Metrics[i].Time,
			Memory:     make([]*agentv3.Memory, len(jvm.Metrics[i].Memory)),
			MemoryPool: make([]*agentv3.MemoryPool, len(jvm.Metrics[i].MemoryPool)),
			Gc:         make([]*agentv3.GC, len(jvm.Metrics[i].Gc)),
		}
		if jvm.Metrics[i].Cpu != nil {
			newJVM.Metrics[i].Cpu = &commonv3.CPU{UsagePercent: jvm.Metrics[i].Cpu.UsagePercent}
		}
		if jvm.Metrics[i].Thread != nil {
			newJVM.Metrics[i].Thread = &agentv3.Thread{
				LiveCount:   jvm.Metrics[i].Thread.LiveCount,
				DaemonCount: jvm.Metrics[i].Thread.DaemonCount,
				PeakCount:   jvm.Metrics[i].Thread.PeakCount,
			}
		}
	}
	api.ProcessMetrics(&newJVM)

	return &commonv3old.Commands{}, nil
}

type ProfileTaskServerV3Old struct {
	profilev3old.UnimplementedProfileTaskServer
}

func (*ProfileTaskServerV3Old) GetProfileTaskCommands(ctx context.Context,
	task *profilev3old.ProfileTaskCommandQuery,
) (*commonv3old.Commands, error) {
	log.Debugf("### ProfileTaskServerV3Old:GetProfileTaskCommands ProfileTaskCommandQuery: %#v", task)

	return &commonv3old.Commands{}, nil
}

func (*ProfileTaskServerV3Old) CollectSnapshot(psrv profilev3old.ProfileTask_CollectSnapshotServer) error {
	for {
		profile, err := psrv.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return psrv.SendAndClose(&commonv3old.Commands{})
			}
			log.Debug(err.Error())

			return err
		}
		log.Debugf("### ProfileTaskServerV3Old:CollectSnapshot ThreadSnapshot: %#v", profile)

		newProfile := profilev3.ThreadSnapshot{
			TaskId:         profile.TaskId,
			TraceSegmentId: profile.TraceSegmentId,
			Time:           profile.Time,
			Sequence:       profile.Sequence,
		}
		if profile.Stack != nil {
			newProfile.Stack = &profilev3.ThreadStack{CodeSignatures: profile.Stack.CodeSignatures}
		}
		api.ProcessProfile(&newProfile)
	}
}

func (*ProfileTaskServerV3Old) ReportTaskFinish(ctx context.Context, reporter *profilev3old.ProfileTaskFinishReport) (*commonv3old.Commands, error) {
	log.Debugf("### ProfileTaskServerV3Old:ReportTaskFinish ProfileTaskFinishReport: %#v", reporter)

	return &commonv3old.Commands{}, nil
}

type ManagementServerV3Old struct {
	mgmtv3old.UnimplementedManagementServiceServer
}

func (*ManagementServerV3Old) ReportInstanceProperties(ctx context.Context, mgmt *mgmtv3old.InstanceProperties) (*commonv3old.Commands, error) {
	log.Debugf("### ManagementServerV3Old:ReportInstanceProperties InstanceProperties: %#v", mgmt)

	return &commonv3old.Commands{}, nil
}

func (*ManagementServerV3Old) KeepAlive(ctx context.Context, ping *mgmtv3old.InstancePingPkg) (*commonv3old.Commands, error) {
	log.Debugf("### ManagementServerV3Old:KeepAlive InstancePingPkg: %#v", ping)

	return &commonv3old.Commands{}, nil
}

type TraceReportServerV3 struct {
	agentv3.UnimplementedTraceSegmentReportServiceServer
}

func (*TraceReportServerV3) Collect(tsr agentv3.TraceSegmentReportService_CollectServer) error {
	for {
		segobj, err := tsr.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return tsr.SendAndClose(&commonv3.Commands{})
			}
			log.Error(err.Error())

			return err
		}
		log.Debugf("### TraceReportServerV3:Collect SegmentObject: %#v", segobj)

		api.ProcessSegment(segobj)
	}
}

func (*TraceReportServerV3) CollectInSync(ctx context.Context, seg *agentv3.SegmentCollection) (*commonv3.Commands, error) {
	log.Debugf("### TraceReportServerV3:CollectInSync SegmentCollection: %#v", seg)

	return &commonv3.Commands{}, nil
}

type EventServerV3 struct {
	eventv3.UnimplementedEventServiceServer
}

func (*EventServerV3) Collect(esrv eventv3.EventService_CollectServer) error {
	for {
		event, err := esrv.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return esrv.SendAndClose(&commonv3.Commands{})
			}
			log.Debug(err.Error())

			return err
		}
		log.Debugf("### EventServerV3:Collect Event: %#v", event)
	}
}

type JVMMetricReportServerV3 struct {
	agentv3.UnimplementedJVMMetricReportServiceServer
}

func (*JVMMetricReportServerV3) Collect(ctx context.Context, jvm *agentv3.JVMMetricCollection) (*commonv3.Commands, error) {
	log.Debugf("### JVMMetricReportServerV3:Collect JVMMetricCollection: %#v", jvm)

	api.ProcessMetrics(jvm)

	return &commonv3.Commands{}, nil
}

type LoggingServerV3 struct {
	loggingv3.UnsafeLogReportServiceServer
}

func (*LoggingServerV3) Collect(server loggingv3.LogReportService_CollectServer) error {
	for {
		logData, err := server.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return server.SendAndClose(&commonv3.Commands{})
			}
			log.Debug(err.Error())

			return err
		}
		log.Debugf("### LoggingServerV3:Collect LogData: %#v", logData)

		api.ProcessLog(logData)
	}
}

type ProfileTaskServerV3 struct {
	profilev3.UnimplementedProfileTaskServer
}

func (*ProfileTaskServerV3) GetProfileTaskCommands(ctx context.Context, task *profilev3.ProfileTaskCommandQuery) (*commonv3.Commands, error) {
	log.Debugf("### ProfileTaskServerV3:GetProfileTaskCommands ProfileTaskCommandQuery: %#v", task)

	return &commonv3.Commands{}, nil
}

func (*ProfileTaskServerV3) CollectSnapshot(psrv profilev3.ProfileTask_CollectSnapshotServer) error {
	for {
		profile, err := psrv.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return psrv.SendAndClose(&commonv3.Commands{})
			}
			log.Debug(err.Error())

			return err
		}
		log.Debugf("### ProfileTaskServerV3:CollectSnapshot ThreadSnapshot: %#v", profile)

		api.ProcessProfile(profile)
	}
}

func (*ProfileTaskServerV3) ReportTaskFinish(ctx context.Context, reporter *profilev3.ProfileTaskFinishReport) (*commonv3.Commands, error) {
	log.Debugf("### ProfileTaskServerV3:ReportTaskFinish ProfileTaskFinishReport: %#v", reporter)

	return &commonv3.Commands{}, nil
}

type ManagementServerV3 struct {
	mgmtv3.UnimplementedManagementServiceServer
}

func (*ManagementServerV3) ReportInstanceProperties(ctx context.Context, mgmt *mgmtv3.InstanceProperties) (*commonv3.Commands, error) {
	log.Debugf("### ManagementServerV3:ReportInstanceProperties InstanceProperties: %#v", mgmt)

	return &commonv3.Commands{}, nil
}

func (*ManagementServerV3) KeepAlive(ctx context.Context, ping *mgmtv3.InstancePingPkg) (*commonv3.Commands, error) {
	log.Debugf("### ManagementServerV3:KeepAlive InstancePingPkg: %#v", ping)

	return &commonv3.Commands{}, nil
}

type DiscoveryServerV3 struct {
	configv3.UnimplementedConfigurationDiscoveryServiceServer
}

func (*DiscoveryServerV3) FetchConfigurations(ctx context.Context, cfgReq *configv3.ConfigurationSyncRequest) (*commonv3.Commands, error) {
	log.Debugf("### DiscoveryServerV3:FetchConfigurations ConfigurationSyncRequest: %#v", cfgReq)

	return &commonv3.Commands{}, nil
}
