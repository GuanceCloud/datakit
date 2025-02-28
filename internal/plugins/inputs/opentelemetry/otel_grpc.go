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

func runGRPCV1(addr string, ipt *Input) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorf("### opentelemetry grpc server v1 listening on %s failed: %v", addr, err.Error())

		return
	}
	log.Debugf("### opentelemetry grpc v1 listening on: %s", addr)

	otelSvr = grpc.NewServer(itrace.DefaultGRPCServerOpts...)
	trace.RegisterTraceServiceServer(otelSvr, &TraceServiceServer{})
	metrics.RegisterMetricsServiceServer(otelSvr, &MetricsServiceServer{})
	logs.RegisterLogsServiceServer(otelSvr, &LogsServiceServer{Ipt: ipt})

	if err = otelSvr.Serve(listener); err != nil {
		log.Error(err.Error())
	}

	log.Debug("### opentelemetry grpc v1 exits")
}

type TraceServiceServer struct {
	trace.UnimplementedTraceServiceServer
}

func (tss *TraceServiceServer) Export(ctx context.Context, tsreq *trace.ExportTraceServiceRequest) (
	*trace.ExportTraceServiceResponse, error,
) {
	if afterGatherRun != nil {
		if dktraces := parseResourceSpans(tsreq.ResourceSpans); len(dktraces) != 0 {
			afterGatherRun.Run(inputName, dktraces)
		}
	}

	return &trace.ExportTraceServiceResponse{}, nil
}

type MetricsServiceServer struct {
	metrics.UnimplementedMetricsServiceServer
}

func (mss *MetricsServiceServer) Export(ctx context.Context, msreq *metrics.ExportMetricsServiceRequest) (
	*metrics.ExportMetricsServiceResponse, error,
) {
	parseResourceMetricsV2(msreq.ResourceMetrics)

	return &metrics.ExportMetricsServiceResponse{}, nil
}

type LogsServiceServer struct {
	logs.UnimplementedLogsServiceServer
	Ipt *Input
}

func (l *LogsServiceServer) Export(ctx context.Context, logsReq *logs.ExportLogsServiceRequest) (out *logs.ExportLogsServiceResponse, err error) {
	if logsReq == nil || len(logsReq.GetResourceLogs()) == 0 {
		return
	}
	start := time.Now()
	pts := ParseLogsRequest(logsReq.GetResourceLogs())
	if len(pts) != 0 {
		if err := l.Ipt.feeder.FeedV2(point.Logging, pts,
			dkio.WithInputName(inputName),
			dkio.WithCollectCost(time.Since(start)),
		); err != nil {
			log.Error(err.Error())
		}
	}
	out = &logs.ExportLogsServiceResponse{PartialSuccess: &logs.ExportLogsPartialSuccess{}}

	return out, nil
}
