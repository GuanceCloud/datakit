// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"context"
	"net"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	logs "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/collector/logs/v1"
	metrics "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/collector/metrics/v1"
	trace "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/collector/trace/v1"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/peer"
)

type gRPC struct {
	Address    string `toml:"addr" json:"addr"`
	MaxPayload int    `toml:"max_payload" json:"max_payload"`

	trueAddr string

	otelSvr        *grpc.Server
	afterGatherRun itrace.AfterGatherHandler
	feeder         dkio.Feeder
}

func (gc *gRPC) runGRPCV1(ipt *Input) {
	listener, err := net.Listen("tcp", gc.Address)
	if err != nil {
		log.Errorf("opentelemetry grpc server v1 listening on %s failed: %v", gc.Address, err.Error())

		return
	}

	gc.trueAddr = listener.Addr().String()
	log.Debugf("opentelemetry grpc v1 listening on: %s", gc.trueAddr)

	if gc.MaxPayload <= 0 {
		gc.MaxPayload = 16777216
	}

	log.Infof("set gRPC max payload to %d", gc.MaxPayload)

	gc.otelSvr = grpc.NewServer(
		grpc.MaxRecvMsgSize(gc.MaxPayload),
		itrace.OTLPInterceptors(),
	)

	// register T/M/L gRPC server
	trace.RegisterTraceServiceServer(gc.otelSvr, &TraceServiceServer{Gather: gc.afterGatherRun, input: ipt})
	metrics.RegisterMetricsServiceServer(gc.otelSvr, &MetricsServiceServer{input: ipt})
	logs.RegisterLogsServiceServer(gc.otelSvr, &LogsServiceServer{input: ipt})

	if err = gc.otelSvr.Serve(listener); err != nil {
		log.Errorf("grpc server err=%v", err)
	}

	log.Info("otel grpc v1 exits")
}

func (gc *gRPC) stop() {
	if gc.otelSvr != nil {
		gc.otelSvr.GracefulStop()
	}
}

type TraceServiceServer struct {
	trace.UnimplementedTraceServiceServer
	Gather itrace.AfterGatherHandler
	input  *Input
}

func (tss *TraceServiceServer) Export(ctx context.Context, tsreq *trace.ExportTraceServiceRequest) (
	*trace.ExportTraceServiceResponse, error,
) {
	if dktraces := tss.input.parseResourceSpans(tsreq.ResourceSpans, getRemoteIP(ctx)); len(dktraces) != 0 {
		if tss.Gather != nil {
			tss.Gather.Run(inputName, dktraces)
		}
	}

	return &trace.ExportTraceServiceResponse{}, nil
}

type MetricsServiceServer struct {
	metrics.UnimplementedMetricsServiceServer
	input *Input
}

func (mss *MetricsServiceServer) Export(ctx context.Context, msreq *metrics.ExportMetricsServiceRequest) (
	*metrics.ExportMetricsServiceResponse, error,
) {
	mss.input.parseResourceMetricsV2(msreq.ResourceMetrics, getRemoteIP(ctx))

	return &metrics.ExportMetricsServiceResponse{}, nil
}

type LogsServiceServer struct {
	logs.UnimplementedLogsServiceServer
	input *Input
}

func (l *LogsServiceServer) Export(ctx context.Context, logsReq *logs.ExportLogsServiceRequest) (out *logs.ExportLogsServiceResponse, err error) {
	if logsReq == nil || len(logsReq.GetResourceLogs()) == 0 {
		return
	}
	start := time.Now()
	pts := l.input.parseLogRequest(logsReq.GetResourceLogs(), getRemoteIP(ctx))
	if len(pts) != 0 {
		if err := l.input.feeder.Feed(point.Logging, pts,
			dkio.WithSource(inputName),
			dkio.WithCollectCost(time.Since(start)),
		); err != nil {
			log.Error(err.Error())
		}
	}
	out = &logs.ExportLogsServiceResponse{PartialSuccess: &logs.ExportLogsPartialSuccess{}}

	return out, nil
}

func getRemoteIP(ctx context.Context) string {
	remoteIP := ""
	p, ok := peer.FromContext(ctx)
	if ok && p.Addr != nil {
		remoteIP, _, _ = net.SplitHostPort(p.Addr.String()) //nolint
	}

	return remoteIP
}
