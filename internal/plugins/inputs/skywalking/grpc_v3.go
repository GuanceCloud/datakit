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
	"time"

	"github.com/GuanceCloud/cliutils/point"
	configv3 "github.com/GuanceCloud/tracing-protos/skywalking-gen-go/agent/configuration/v3"
	commonv3 "github.com/GuanceCloud/tracing-protos/skywalking-gen-go/common/v3"
	eventv3 "github.com/GuanceCloud/tracing-protos/skywalking-gen-go/event/v3"
	agentv3 "github.com/GuanceCloud/tracing-protos/skywalking-gen-go/language/agent/v3"
	agentv3old "github.com/GuanceCloud/tracing-protos/skywalking-gen-go/language/agent/v3/compat"
	profilev3 "github.com/GuanceCloud/tracing-protos/skywalking-gen-go/language/profile/v3"
	profilev3old "github.com/GuanceCloud/tracing-protos/skywalking-gen-go/language/profile/v3/compat"
	loggingv3 "github.com/GuanceCloud/tracing-protos/skywalking-gen-go/logging/v3"
	mgmtv3 "github.com/GuanceCloud/tracing-protos/skywalking-gen-go/management/v3"
	mgmtv3old "github.com/GuanceCloud/tracing-protos/skywalking-gen-go/management/v3/compat"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func runGRPCV3(ipt *Input) {
	listener, err := net.Listen("tcp", ipt.Address)
	if err != nil {
		log.Errorf("### skywalking grpc server v3 listening on %s failed: %v", ipt.Address, err)

		return
	}
	log.Debugf("### skywalking grpc v3 listening on: %s", ipt.Address)

	skySvr = grpc.NewServer(itrace.DefaultGRPCServerOpts...)
	// register API version 8.3.0
	agentv3old.RegisterTraceSegmentReportServiceServer(skySvr, &TraceReportServerV3Old{})
	agentv3old.RegisterJVMMetricReportServiceServer(skySvr, &JVMMetricReportServerV3Old{ipt: ipt})
	profilev3old.RegisterProfileTaskServer(skySvr, &ProfileTaskServerV3Old{})
	mgmtv3old.RegisterManagementServiceServer(skySvr, &ManagementServerV3Old{})
	// register API version 9.4.0
	agentv3.RegisterTraceSegmentReportServiceServer(skySvr, &TraceReportServerV3{})
	eventv3.RegisterEventServiceServer(skySvr, &EventServerV3{})
	agentv3.RegisterMeterReportServiceServer(skySvr, &MeterReportServiceServerImpl{feeder: ipt.feeder})
	agentv3.RegisterJVMMetricReportServiceServer(skySvr, &JVMMetricReportServerV3{ipt: ipt})
	loggingv3.RegisterLogReportServiceServer(skySvr, &LoggingServerV3{Ipt: ipt})
	profilev3.RegisterProfileTaskServer(skySvr, &ProfileTaskServerV3{})
	mgmtv3.RegisterManagementServiceServer(skySvr, &ManagementServerV3{})
	configv3.RegisterConfigurationDiscoveryServiceServer(skySvr, &DiscoveryServerV3{})

	if err = skySvr.Serve(listener); err != nil {
		log.Error(err.Error())
	}

	log.Debug("### skywalking v3 exits")
}

type MeterReportServiceServerImpl struct {
	feeder dkio.Feeder
	agentv3.UnimplementedMeterReportServiceServer
}

func (m *MeterReportServiceServerImpl) Collect(collect agentv3.MeterReportService_CollectServer) error {
	meters, err := collect.Recv()
	if err != nil {
		return err
	}

	pt := meterDataToPoint(meters)
	if pt == nil {
		return nil
	}

	err = m.feeder.Feed("skywalking_meter", point.Metric, []*point.Point{pt})
	if err != nil {
		log.Warnf("feeder error=%v", err)
	}
	return nil
}

func (m *MeterReportServiceServerImpl) CollectBatch(collects agentv3.MeterReportService_CollectBatchServer) error {
	meters, err := collects.Recv()
	if err != nil {
		return nil
	}
	datas := meters.GetMeterData()
	pts := make([]*point.Point, 0)
	for _, data := range datas {
		pt := meterDataToPoint(data)
		if pt != nil {
			pts = append(pts, pt)
		}
	}
	err = m.feeder.Feed("skywalking_meter", point.Metric, pts)
	if err != nil {
		log.Warnf("feeder error=%v", err)
	}

	return nil
}

type TraceReportServerV3Old struct {
	agentv3old.UnimplementedTraceSegmentReportServiceServer
}

func (*TraceReportServerV3Old) Collect(tsr agentv3old.TraceSegmentReportService_CollectServer) error {
	for {
		segobj, err := tsr.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return tsr.SendAndClose(&commonv3.Commands{})
			}
			log.Error(err.Error())

			return err
		}
		log.Debugf("### TraceReportServerV3Old:Collect SegmentObject:%#v", segobj)

		bts, err := proto.Marshal(segobj)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		if localCache == nil || !localCache.Enabled() {
			newSegObj := &agentv3.SegmentObject{}
			if err = proto.Unmarshal(bts, newSegObj); err != nil {
				log.Error(err.Error())
				continue
			}
			dktrace := parseSegmentObjectV3(newSegObj)
			if len(dktrace) != 0 && afterGatherRun != nil {
				afterGatherRun.Run(inputName, itrace.DatakitTraces{dktrace})
			}
		} else {
			if err = localCache.Put(storage.SKY_WALKING_GRPC_KEY, bts); err != nil {
				log.Error(err.Error())
			}
		}
	}
}

func (*TraceReportServerV3Old) CollectInSync(ctx context.Context, col *agentv3.SegmentCollection) (*commonv3.Commands, error) {
	log.Debugf("### TraceReportServerV3Old:CollectInSync SegmentCollection: %#v", col)

	for _, segobj := range col.Segments {
		bts, err := proto.Marshal(segobj)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		if localCache == nil || !localCache.Enabled() {
			newSegObj := &agentv3.SegmentObject{}
			if err = proto.Unmarshal(bts, newSegObj); err != nil {
				log.Error(err.Error())
				continue
			}
			dktrace := parseSegmentObjectV3(newSegObj)
			if len(dktrace) != 0 && afterGatherRun != nil {
				afterGatherRun.Run(inputName, itrace.DatakitTraces{dktrace})
			}
		} else {
			if err = localCache.Put(storage.SKY_WALKING_GRPC_KEY, bts); err != nil {
				log.Error(err.Error())
			}
		}
	}

	return &commonv3.Commands{}, nil
}

type JVMMetricReportServerV3Old struct {
	agentv3old.UnimplementedJVMMetricReportServiceServer
	ipt *Input
}

func (r *JVMMetricReportServerV3Old) Collect(ctx context.Context, jvm *agentv3.JVMMetricCollection) (*commonv3.Commands, error) {
	log.Debugf("### JVMMetricReportServerV3Old:Collect %#v", jvm)

	start := time.Now()
	bts, err := proto.Marshal(jvm)
	if err != nil {
		log.Error(err.Error())

		return &commonv3.Commands{}, err
	}
	newjvm := &agentv3.JVMMetricCollection{}
	if err = proto.Unmarshal(bts, newjvm); err != nil {
		log.Error(err.Error())

		return &commonv3.Commands{}, err
	}

	metrics := processMetricsV3(newjvm, start, r.ipt)
	if len(metrics) != 0 {
		if err := r.ipt.feeder.Feed(jvmMetricName, point.Metric, metrics, &dkio.Option{CollectCost: time.Since(start)}); err != nil {
			r.ipt.feeder.FeedLastError(err.Error(),
				dkio.WithLastErrorInput(inputName),
				dkio.WithLastErrorSource(jvmMetricName),
			)
		}
	}

	return &commonv3.Commands{}, nil
}

type ProfileTaskServerV3Old struct {
	profilev3old.UnimplementedProfileTaskServer
}

func (*ProfileTaskServerV3Old) GetProfileTaskCommands(ctx context.Context,
	task *profilev3.ProfileTaskCommandQuery,
) (*commonv3.Commands, error) {
	log.Debugf("### ProfileTaskServerV3Old:GetProfileTaskCommands ProfileTaskCommandQuery: %#v", task)

	return &commonv3.Commands{}, nil
}

func (*ProfileTaskServerV3Old) CollectSnapshot(psrv profilev3old.ProfileTask_CollectSnapshotServer) error {
	for {
		profile, err := psrv.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return psrv.SendAndClose(&commonv3.Commands{})
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
		processProfileV3(&newProfile)
	}
}

func (*ProfileTaskServerV3Old) ReportTaskFinish(ctx context.Context, reporter *profilev3.ProfileTaskFinishReport) (*commonv3.Commands, error) {
	log.Debugf("### ProfileTaskServerV3Old:ReportTaskFinish ProfileTaskFinishReport: %#v", reporter)

	return &commonv3.Commands{}, nil
}

type ManagementServerV3Old struct {
	mgmtv3old.UnimplementedManagementServiceServer
}

func (*ManagementServerV3Old) ReportInstanceProperties(ctx context.Context, mgmt *mgmtv3.InstanceProperties) (*commonv3.Commands, error) {
	log.Debugf("### ManagementServerV3Old:ReportInstanceProperties InstanceProperties: %#v", mgmt)

	return &commonv3.Commands{}, nil
}

func (*ManagementServerV3Old) KeepAlive(ctx context.Context, ping *mgmtv3.InstancePingPkg) (*commonv3.Commands, error) {
	log.Debugf("### ManagementServerV3Old:KeepAlive InstancePingPkg: %#v", ping)

	return &commonv3.Commands{}, nil
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

		if localCache == nil || !localCache.Enabled() {
			dktrace := parseSegmentObjectV3(segobj)
			if len(dktrace) != 0 && afterGatherRun != nil {
				afterGatherRun.Run(inputName, itrace.DatakitTraces{dktrace})
			}
		} else {
			if bts, err := proto.Marshal(segobj); err != nil {
				log.Error(err.Error())
			} else {
				if err = localCache.Put(storage.SKY_WALKING_GRPC_KEY, bts); err != nil {
					log.Error(err.Error())
				}
			}
		}
	}
}

func (*TraceReportServerV3) CollectInSync(ctx context.Context, col *agentv3.SegmentCollection) (*commonv3.Commands, error) {
	log.Debugf("### TraceReportServerV3:CollectInSync SegmentCollection: %#v", col)

	for _, segobj := range col.Segments {
		if localCache == nil || !localCache.Enabled() {
			dktrace := parseSegmentObjectV3(segobj)
			if len(dktrace) != 0 && afterGatherRun != nil {
				afterGatherRun.Run(inputName, itrace.DatakitTraces{dktrace})
			}
		} else {
			if bts, err := proto.Marshal(segobj); err != nil {
				log.Error(err.Error())
				continue
			} else {
				if err = localCache.Put(storage.SKY_WALKING_GRPC_KEY, bts); err != nil {
					log.Error(err.Error())
				}
			}
		}
	}

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
	ipt *Input
}

func (r *JVMMetricReportServerV3) Collect(ctx context.Context, jvm *agentv3.JVMMetricCollection) (*commonv3.Commands, error) {
	log.Debugf("### JVMMetricReportServerV3:Collect JVMMetricCollection: %#v", jvm)

	start := time.Now()
	metrics := processMetricsV3(jvm, start, r.ipt)
	if len(metrics) != 0 {
		if err := r.ipt.feeder.Feed(jvmMetricName, point.Metric, metrics, &dkio.Option{CollectCost: time.Since(start)}); err != nil {
			r.ipt.feeder.FeedLastError(err.Error(),
				dkio.WithLastErrorInput(inputName),
				dkio.WithLastErrorSource(jvmMetricName),
			)
		}
	}

	return &commonv3.Commands{}, nil
}

type LoggingServerV3 struct {
	loggingv3.UnsafeLogReportServiceServer
	Ipt *Input
}

func (ls *LoggingServerV3) Collect(server loggingv3.LogReportService_CollectServer) error {
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

		if pt, err := processLogV3(logData); err != nil {
			log.Error(err.Error())
		} else {
			if err = ls.Ipt.feeder.Feed(logData.Service, point.Logging, []*point.Point{pt}, nil); err != nil {
				log.Error(err.Error())
			}
		}
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

		processProfileV3(profile)
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
