// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"context"
	"net"

	"github.com/GuanceCloud/cliutils/point"
	metrics "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/collector/metrics/v1"
	trace "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/collector/trace/v1"
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
	metrics.RegisterMetricsServiceServer(otelSvr, &MetricsServiceServer{Ipt: ipt})

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
	Ipt *Input
}

func (mss *MetricsServiceServer) Export(ctx context.Context, msreq *metrics.ExportMetricsServiceRequest) (
	*metrics.ExportMetricsServiceResponse, error,
) {
	omcs := parseResourceMetrics(msreq.ResourceMetrics)
	var points []*point.Point
	for i := range omcs {
		if pts := omcs[i].getPoints(); len(pts) != 0 {
			points = append(points, pts...)
		}
	}
	if len(points) != 0 {
		if err := mss.Ipt.feeder.Feed(inputName, point.Metric, points, nil); err != nil {
			log.Error(err.Error())
		}
	}

	return &metrics.ExportMetricsServiceResponse{}, nil
}
