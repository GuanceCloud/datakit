// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pinpoint handle Pinpoint APM traces.
package pinpoint

import (
	"context"
	"errors"
	"io"
	"net"

	istorage "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	v1 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/pinpoint/compiled/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func runGRPCV1(addr string) {
	if localCache != nil {
		localCache.RegisterConsumer(istorage.PINPOINT_GRPC_KEY, func(buf []byte) error {
			span := &v1.PSpanMessage{}
			if err := proto.Unmarshal(buf, span); err != nil {
				return err
			}

			dktrace, err := parsePPSpanMessage(span)
			if err != nil {
				return err
			}
			if spanSender != nil {
				log.Debugf("### span: %#v", dktrace)
				spanSender.Append(dktrace...)
			}

			return nil
		})
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorf("### grpc server v1 listening on %s failed: %v", addr, err)

		return
	}
	log.Debugf("### grpc server v1 listening on: %s", addr)

	ppsvr = grpc.NewServer()
	v1.RegisterAgentServer(ppsvr, &AgentServer{})
	v1.RegisterMetadataServer(ppsvr, &MetadataServer{})
	v1.RegisterProfilerCommandServiceServer(ppsvr, &ProfilerCommandServiceServer{})
	v1.RegisterStatServer(ppsvr, &StatServer{})
	v1.RegisterSpanServer(ppsvr, &SpanServer{})

	if err = ppsvr.Serve(listener); err != nil {
		log.Error(err.Error())
	}
	log.Debug("### grpc server v1 exits")
}

type ServiceInfo struct {
	Name string
	Libs []string
}

type AgentMetaData struct {
	Hostname     string
	IP           string
	Ports        string
	ServiceType  int32
	Pid          int32
	AgentVersion string
	VMVersion    string
	EndTimestamp int64
	EndStatus    int32
	Container    bool
	ServerInfo   string
	VMArg        []string
	ServiceInfo  []*ServiceInfo
	Version      int32
	GcType       v1.PJvmGcType
}

type AgentServer struct {
	v1.UnimplementedAgentServer
}

func (agtsvr *AgentServer) RequestAgentInfo(ctx context.Context, agInfo *v1.PAgentInfo) (*v1.PResult, error) {
	agentMetaData.Hostname = agInfo.Hostname
	agentMetaData.IP = agInfo.Ip
	agentMetaData.Ports = agInfo.Ports
	agentMetaData.ServiceType = agInfo.ServiceType
	agentMetaData.Pid = agInfo.Pid
	agentMetaData.AgentVersion = agInfo.AgentVersion
	agentMetaData.VMVersion = agInfo.VmVersion
	agentMetaData.EndTimestamp = agInfo.EndTimestamp
	agentMetaData.EndStatus = agInfo.EndStatus
	agentMetaData.Container = agInfo.Container
	agentMetaData.ServerInfo = agInfo.ServerMetaData.ServerInfo
	agentMetaData.VMArg = agInfo.ServerMetaData.VmArg
	agentMetaData.Version = agInfo.JvmInfo.Version
	agentMetaData.GcType = agInfo.JvmInfo.GcType
	for _, v := range agInfo.ServerMetaData.ServiceInfo {
		agentMetaData.ServiceInfo = append(agentMetaData.ServiceInfo, &ServiceInfo{
			Name: v.ServiceName,
			Libs: v.ServiceLib,
		})
	}

	log.Debugf("### agent meta: %#v", agentMetaData)
	for _, v := range agentMetaData.ServiceInfo {
		log.Debugf("### service info: %#v", v)
	}

	return &v1.PResult{Success: true, Message: "ok"}, nil
}

func (agtsvr *AgentServer) PingSession(ping v1.Agent_PingSessionServer) error {
	msg, err := ping.Recv()
	if err != nil {
		log.Error(err.Error())

		return err
	}

	return ping.SendMsg(msg)
}

type MetadataServer struct {
	v1.UnimplementedMetadataServer
}

func (mdsvr *MetadataServer) RequestSqlMetaData(ctx context.Context, meta *v1.PSqlMetaData) (*v1.PResult, error) { // nolint: stylecheck
	if reqMetaTab != nil && meta != nil {
		reqMetaTab.Store(meta.SqlId, v1.NewMetaData(meta.SqlId, meta))
	}

	return &v1.PResult{Success: true, Message: "ok"}, nil
}

func (mdsvr *MetadataServer) RequestApiMetaData(ctx context.Context, meta *v1.PApiMetaData) (*v1.PResult, error) { // nolint: stylecheck
	if reqMetaTab != nil && meta != nil {
		reqMetaTab.Store(meta.ApiId, v1.NewMetaData(meta.ApiId, meta))
	}

	return &v1.PResult{Success: true, Message: "ok"}, nil
}

func (mdsvr *MetadataServer) RequestStringMetaData(ctx context.Context, meta *v1.PStringMetaData) (*v1.PResult, error) {
	if reqMetaTab != nil && meta != nil {
		reqMetaTab.Store(meta.StringId, v1.NewMetaData(meta.StringId, meta))
	}

	return &v1.PResult{Success: true, Message: "ok"}, nil
}

type ProfilerCommandServiceServer struct {
	v1.UnimplementedProfilerCommandServiceServer
}

func (*ProfilerCommandServiceServer) HandleCommand(handler v1.ProfilerCommandService_HandleCommandServer) error {
	if _, err := handler.Recv(); err != nil {
		log.Error(err.Error())

		return err
	}

	return nil
}

func (*ProfilerCommandServiceServer) HandleCommandV2(handler v1.ProfilerCommandService_HandleCommandV2Server) error {
	msg, err := handler.Recv()
	if err != nil {
		log.Error(err.Error())

		return err
	}
	log.Debugf("### profiler handle command v2 %#v", msg)

	return nil
}

func (*ProfilerCommandServiceServer) CommandEcho(ctx context.Context, resp *v1.PCmdEchoResponse) (*emptypb.Empty, error) {
	log.Debugf("### profiler echo command %#v", resp)

	return &emptypb.Empty{}, nil
}

func (*ProfilerCommandServiceServer) CommandStreamActiveThreadCount(stream v1.ProfilerCommandService_CommandStreamActiveThreadCountServer) error {
	for {
		resp, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return stream.SendAndClose(&emptypb.Empty{})
			}
			log.Error(err.Error())

			return err
		}

		log.Debugf("### profiler stream active thread count command %#v", resp)
	}
}

func (*ProfilerCommandServiceServer) CommandActiveThreadDump(ctx context.Context, resp *v1.PCmdActiveThreadDumpRes) (
	*emptypb.Empty, error,
) {
	log.Debugf("### profiler active thread dump command %#v", resp)

	return &emptypb.Empty{}, nil
}

func (pcss *ProfilerCommandServiceServer) CommandActiveThreadLightDump(ctx context.Context,
	resp *v1.PCmdActiveThreadLightDumpRes,
) (*emptypb.Empty, error) {
	log.Debugf("### profiler active thread light dump command %#v", resp)

	return &emptypb.Empty{}, nil
}

type StatServer struct {
	v1.UnimplementedStatServer
}

func (*StatServer) SendAgentStat(statSvr v1.Stat_SendAgentStatServer) error {
	for {
		select {
		case <-statSvr.Context().Done():
			if err := statSvr.Context().Err(); err != nil {
				if errors.Is(err, io.EOF) {
					return statSvr.SendAndClose(&emptypb.Empty{})
				}
				log.Debug(err.Error())

				return err
			}
		default:
		}

		msg, err := statSvr.Recv()
		if err != nil {
			log.Debug(err.Error())

			return err
		}

		log.Debugf("### stat: %#v", msg.GetAgentStat())
	}
}

type SpanServer struct {
	v1.UnimplementedSpanServer
}

func (*SpanServer) SendSpan(spanSvr v1.Span_SendSpanServer) error {
	for {
		msg, err := spanSvr.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return spanSvr.SendAndClose(&emptypb.Empty{})
			}
			log.Debug(err.Error())

			return err
		}

		if localCache == nil || !localCache.Enabled() {
			dktrace, err := parsePPSpanMessage(msg)
			if err != nil {
				log.Debug(err.Error())
				continue
			}
			if spanSender != nil {
				log.Debugf("### span: %#v", dktrace)
				spanSender.Append(dktrace...)
			}
		} else {
			buf, err := proto.Marshal(msg)
			if err != nil {
				log.Error(err.Error())
				continue
			}

			if err = localCache.Put(istorage.PINPOINT_GRPC_KEY, buf); err != nil {
				log.Error(err.Error())
			}
		}
	}
}

func parsePPSpanMessage(msg *v1.PSpanMessage) (itrace.DatakitTrace, error) {
	var trace itrace.DatakitTrace
	if ppspan := msg.GetSpan(); ppspan != nil {
		trace = ppspan.ConvertToDKTrace(inputName, reqMetaTab)
	} else if ppchunk := msg.GetSpanChunk(); ppchunk != nil {
		trace = ppchunk.ConvertToDKTrace(inputName, reqMetaTab)
	} else {
		return nil, errors.New("### empty span message")
	}

	// add on global tags
	if len(tags) != 0 {
		for _, span := range trace {
			if span.Tags == nil {
				span.Tags = make(map[string]string)
			}
			for k, v := range tags {
				span.Tags[k] = v
			}
		}
	}

	return trace, nil
}
