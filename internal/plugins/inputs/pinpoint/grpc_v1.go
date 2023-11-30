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
	"time"

	ppv1 "github.com/GuanceCloud/tracing-protos/pinpoint-gen-go/v1"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/pinpoint/cache"
	istorage "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/proto"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

var agentCache *cache.AgentCache

func runGRPCV1(addr string) {
	if localCache != nil {
		localCache.RegisterConsumer(istorage.PINPOINT_GRPC_KEY, func(buf []byte) error {
			spanw := &ppv1.PSpanMessageWithMeta{}
			if err := proto.Unmarshal(buf, spanw); err != nil {
				return err
			}
			parsePPSpanMessage(makeMeta(spanw.Meta), spanw.SpanMessage)
			return nil
		})
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorf("### grpc server v1 listening on %s failed: %s", addr, err.Error())

		return
	}
	log.Debugf("### grpc server v1 listening on: %s", addr)

	gsvr = grpc.NewServer(itrace.DefaultGRPCServerOpts...)
	ppv1.RegisterAgentServer(gsvr, &AgentServer{})
	ppv1.RegisterMetadataServer(gsvr, &MetadataServer{})
	ppv1.RegisterProfilerCommandServiceServer(gsvr, &ProfilerCommandServiceServer{})
	ppv1.RegisterStatServer(gsvr, &StatServer{})
	ppv1.RegisterSpanServer(gsvr, &SpanServer{})

	if err = gsvr.Serve(listener); err != nil {
		log.Error(err.Error())
	}
	log.Debug("### grpc server v1 exits")
}

type ServiceInfo struct {
	Name string
	Libs []string
}

type AgentServer struct {
	ppv1.UnimplementedAgentServer
}

func (agtsvr *AgentServer) RequestAgentInfo(ctx context.Context, agInfo *ppv1.PAgentInfo) (*ppv1.PResult, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if vals := md.Get("agentid"); len(vals) > 0 {
			log.Debugf("agent info agentid=%s", vals[0])
			agentCache.SetAgentInfo(vals[0], agInfo)
		}
	}
	log.Debugf("agentinfo =%v", agInfo)
	return &ppv1.PResult{Success: true, Message: "ok"}, nil
}

func (agtsvr *AgentServer) PingSession(ping ppv1.Agent_PingSessionServer) error {
	msg, err := ping.Recv()
	if err != nil {
		log.Errorf("pingSession err=%v", err)
		return err
	}

	return ping.SendMsg(msg)
}

type MetadataServer struct {
	ppv1.UnimplementedMetadataServer
}

func (mdsvr *MetadataServer) RequestSqlMetaData(ctx context.Context, meta *ppv1.PSqlMetaData) (*ppv1.PResult, error) { // nolint: stylecheck
	if agentCache != nil {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if vals := md.Get("agentid"); len(vals) > 0 {
				agentCache.StoreMeta(vals[0], meta)
			}
		}
	}

	return &ppv1.PResult{Success: true, Message: "ok"}, nil
}

func (mdsvr *MetadataServer) RequestApiMetaData(ctx context.Context, meta *ppv1.PApiMetaData) (*ppv1.PResult, error) { // nolint: stylecheck
	if agentCache != nil {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if vals := md.Get("agentid"); len(vals) > 0 {
				log.Debugf("store agentid=%s MD=%+v  meta=%v", vals[0], md, meta)
				agentCache.StoreMeta(vals[0], meta)
			}
		}
	}

	return &ppv1.PResult{Success: true, Message: "ok"}, nil
}

func (mdsvr *MetadataServer) RequestStringMetaData(ctx context.Context, meta *ppv1.PStringMetaData) (*ppv1.PResult, error) {
	if agentCache != nil {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if vals := md.Get("agentid"); len(vals) > 0 {
				log.Debugf("store agentid=%s MD=%+v  meta=%v", vals[0], md, meta)
				agentCache.StoreMeta(vals[0], meta)
			}
		}
	}

	return &ppv1.PResult{Success: true, Message: "ok"}, nil
}

type ProfilerCommandServiceServer struct {
	ppv1.UnimplementedProfilerCommandServiceServer
}

func (*ProfilerCommandServiceServer) HandleCommand(handler ppv1.ProfilerCommandService_HandleCommandServer) error {
	if _, err := handler.Recv(); err != nil {
		log.Error(err.Error())

		return err
	}
	time.Sleep(time.Second)
	return nil
}

func (*ProfilerCommandServiceServer) HandleCommandV2(handler ppv1.ProfilerCommandService_HandleCommandV2Server) error {
	msg, err := handler.Recv()
	if err != nil {
		log.Errorf("close from pinpoint client err=%v", err)
		return err
	}
	log.Debugf("### profiler handle command v2 %#v", msg)
	time.Sleep(time.Second)

	return nil
}

func (*ProfilerCommandServiceServer) CommandEcho(ctx context.Context, resp *ppv1.PCmdEchoResponse) (*emptypb.Empty, error) {
	log.Debugf("### profiler echo command %#v", resp)

	return &emptypb.Empty{}, nil
}

// CommandStreamActiveThreadCount 此方法在测试中并没有调用，可能与agent配置有关.
func (*ProfilerCommandServiceServer) CommandStreamActiveThreadCount(stream ppv1.ProfilerCommandService_CommandStreamActiveThreadCountServer) error {
	for {
		resp, err := stream.Recv()
		if err != nil {
			log.Errorf("CommandStreamActiveThreadCount err=%v", err)
			if errors.Is(err, io.EOF) {
				return stream.SendAndClose(&emptypb.Empty{})
			}

			return err
		}

		log.Debugf("### profiler stream active thread count command %#v", resp)
	}
}

func (*ProfilerCommandServiceServer) CommandActiveThreadDump(ctx context.Context, resp *ppv1.PCmdActiveThreadDumpRes) (
	*emptypb.Empty, error,
) {
	log.Debugf("### profiler active thread dump command %#v", resp)

	return &emptypb.Empty{}, nil
}

func (pcss *ProfilerCommandServiceServer) CommandActiveThreadLightDump(ctx context.Context,
	resp *ppv1.PCmdActiveThreadLightDumpRes,
) (*emptypb.Empty, error) {
	log.Debugf("### profiler active thread light dump command %#v", resp)

	return &emptypb.Empty{}, nil
}

// StatServer agent metric.
type StatServer struct {
	ppv1.UnimplementedStatServer
}

func (*StatServer) SendAgentStat(statSvr ppv1.Stat_SendAgentStatServer) error {
	for {
		md, ok := metadata.FromIncomingContext(statSvr.Context())
		if ok {
			log.Debugf("agent stat md=%+v", md)
		}
		msg, err := statSvr.Recv()
		if err != nil {
			log.Errorf("agentStat err=%v", err)
			agentCache.SyncMetaConn(md, false)
			if errors.Is(err, io.EOF) {
				return statSvr.SendAndClose(&emptypb.Empty{})
			}

			return err
		}

		if agentCache != nil {
			ParsePPAgentStatMessage(md, msg)
			agentCache.SyncMetaConn(md, true)
		}

		log.Debugf("### stat: %s", msg.String())
	}
}

type SpanServer struct {
	ppv1.UnimplementedSpanServer
}

func (*SpanServer) SendSpan(spanSvr ppv1.Span_SendSpanServer) error {
	for {
		md, _ := metadata.FromIncomingContext(spanSvr.Context())
		p, ok := peer.FromContext(spanSvr.Context())
		if ok {
			md.Set("addr", p.Addr.String())
		}
		msg, err := spanSvr.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return spanSvr.SendAndClose(&emptypb.Empty{})
			}
			log.Debug(err.Error())

			return err
		}

		if localCache == nil || !localCache.Enabled() {
			parsePPSpanMessage(md, msg)
		} else {
			spanw := &ppv1.PSpanMessageWithMeta{
				SpanMessage: msg,
				Meta:        makePSpanMeta(md),
			}
			buf, err := proto.Marshal(spanw)
			if err != nil {
				log.Errorf("proto marshal err=%v", err)
				continue
			}

			if err = localCache.Put(istorage.PINPOINT_GRPC_KEY, buf); err != nil {
				log.Errorf("local cache put err=%v", err)
			}
		}
	}
}

func parsePPSpanMessage(meta metadata.MD, msg *ppv1.PSpanMessage) {
	if ppspan := msg.GetSpan(); ppspan != nil {
		if ppspan.ParentSpanId != -1 {
			log.Debugf("store span %+v", ppspan)
			agentCache.SetSpan(ppspan.SpanId, ppspan, meta, 0)
		} else {
			ConvertPSpanToDKTrace(ppspan, meta) // span
		}
	}

	if ppchunk := msg.GetSpanChunk(); ppchunk != nil {
		ConvertPSpanChunkToDKTrace(ppchunk, meta) // spanChunk
	}
}

func makePSpanMeta(md metadata.MD) map[string]*ppv1.StringSlice {
	if len(md) == 0 {
		return nil
	}

	spmd := make(map[string]*ppv1.StringSlice)
	for k, v := range md {
		spmd[k] = &ppv1.StringSlice{Values: v}
	}

	return spmd
}

func makeMeta(spmd map[string]*ppv1.StringSlice) metadata.MD {
	if len(spmd) == 0 {
		return nil
	}

	md := make(metadata.MD)
	for k, v := range spmd {
		md[k] = v.Values
	}

	return md
}
