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
)

type gRPC struct {
	Address string `toml:"addr" json:"addr"`

	otelSvr        *grpc.Server
	afterGatherRun itrace.AfterGatherHandler
	feeder         dkio.Feeder
}

func (gc *gRPC) runGRPCV1(addr string, ipt *Input) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorf("### opentelemetry grpc server v1 listening on %s failed: %v", addr, err.Error())

		return
	}
	log.Debugf("### opentelemetry grpc v1 listening on: %s", addr)

	gc.otelSvr = grpc.NewServer(itrace.DefaultGRPCServerOpts...)
	trace.RegisterTraceServiceServer(gc.otelSvr, &TraceServiceServer{Gather: gc.afterGatherRun, input: ipt})
	metrics.RegisterMetricsServiceServer(gc.otelSvr, &MetricsServiceServer{input: ipt})
	logs.RegisterLogsServiceServer(gc.otelSvr, &LogsServiceServer{input: ipt})

	if err = gc.otelSvr.Serve(listener); err != nil {
		log.Errorf("grpc server err=%v", err)
	}

	log.Info("OPenTelemetry grpc v1 exits")
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
	if tss.Gather != nil {
		if dktraces := tss.input.parseResourceSpans(tsreq.ResourceSpans); len(dktraces) != 0 {
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
	mss.input.parseResourceMetricsV2(msreq.ResourceMetrics)

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
	pts := l.input.parseLogRequest(logsReq.GetResourceLogs())
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
