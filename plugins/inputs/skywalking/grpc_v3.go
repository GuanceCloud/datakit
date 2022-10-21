// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package skywalking

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"

	loggingv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/logging/v3"

	configv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/agent/configuration/v3"
	commonv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/common/v3"
	eventv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/event/v3"
	agentv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/language/agent/v3"
	profilev3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/language/profile/v3"
	mgmtv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/management/v3"
	"google.golang.org/grpc"
)

func registerServerV3(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorf("### skywalking grpc server v3 listening on %s failed: %v", addr, err)

		return
	}
	log.Debugf("### skywalking grpc v3 listening on: %s", addr)

	skySvr = grpc.NewServer()
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
	log.Debug("### skywalking v3 exits")
}

type TraceReportServerV3 struct {
	agentv3.UnimplementedTraceSegmentReportServiceServer
}

func (*TraceReportServerV3) Collect(tsc agentv3.TraceSegmentReportService_CollectServer) error {
	for {
		segobj, err := tsc.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return tsc.SendAndClose(&commonv3.Commands{})
			}
			log.Error(err.Error())

			return err
		}
		log.Debug("### v3 segment received")
		api.ProcessSegment(segobj)
	}
}

func (*TraceReportServerV3) CollectInSync(ctx context.Context, seg *agentv3.SegmentCollection) (*commonv3.Commands, error) {
	log.Debugf("### reveived collect sync: %s", seg.String())

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

		log.Debugf("### reveived service event: %s", event.String())
	}
}

type JVMMetricReportServerV3 struct {
	agentv3.UnimplementedJVMMetricReportServiceServer
}

func (*JVMMetricReportServerV3) Collect(ctx context.Context, jvm *agentv3.JVMMetricCollection) (*commonv3.Commands, error) {
	log.Debugf("### JVMMetricReportService service:%v instance:%v", jvm.Service, jvm.ServiceInstance)
	api.ProcessMetrics(jvm)
	return &commonv3.Commands{}, nil
}

type LoggingServerV3 struct {
	loggingv3.UnsafeLogReportServiceServer
}

func (l LoggingServerV3) Collect(server loggingv3.LogReportService_CollectServer) error {
	logData, err := server.Recv()
	if err != nil {
		return err
	}
	api.ProcessLog(logData)
	return nil
}

type ProfileTaskServerV3 struct {
	profilev3.UnimplementedProfileTaskServer
}

func (*ProfileTaskServerV3) GetProfileTaskCommands(ctx context.Context, task *profilev3.ProfileTaskCommandQuery) (*commonv3.Commands, error) {
	return nil, nil
}

func (*ProfileTaskServerV3) CollectSnapshot(psrv profilev3.ProfileTask_CollectSnapshotServer) error {
	profile, err := psrv.Recv()
	if err != nil {
		log.Warnf("recover profile err=%v", err)
		return err
	}
	api.ProcessProfile(profile)
	return nil
}

func (*ProfileTaskServerV3) ReportTaskFinish(ctx context.Context, reporter *profilev3.ProfileTaskFinishReport) (*commonv3.Commands, error) {
	return nil, nil
}

type ManagementServerV3 struct {
	mgmtv3.UnimplementedManagementServiceServer
}

func (*ManagementServerV3) ReportInstanceProperties(ctx context.Context, mgmt *mgmtv3.InstanceProperties) (*commonv3.Commands, error) {
	var kvpStr string
	for _, kvp := range mgmt.Properties {
		kvpStr += fmt.Sprintf("[%v:%v]", kvp.Key, kvp.Value)
	}
	log.Debugf("### ReportInstanceProperties service:%v instance:%v properties:%v", mgmt.Service, mgmt.ServiceInstance, kvpStr)

	return &commonv3.Commands{}, nil
}

func (*ManagementServerV3) KeepAlive(ctx context.Context, ping *mgmtv3.InstancePingPkg) (*commonv3.Commands, error) {
	log.Debugf("### KeepAlive service:%v instance:%v", ping.Service, ping.ServiceInstance)

	return &commonv3.Commands{}, nil
}

type DiscoveryServerV3 struct {
	configv3.UnimplementedConfigurationDiscoveryServiceServer
}

func (*DiscoveryServerV3) FetchConfigurations(ctx context.Context, cfgReq *configv3.ConfigurationSyncRequest) (*commonv3.Commands, error) {
	log.Debugf("### DiscoveryServerV3 service: %s", cfgReq.String())

	return &commonv3.Commands{}, nil
}
