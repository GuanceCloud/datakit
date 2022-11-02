// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package opentelemetry is GRPC : trace & metric
package opentelemetry

import (
	"context"
	"fmt"
	"net"
	"strings"

	istorage "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/opentelemetry/collector"
	collectormetricepb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type otlpGrpcCollector struct {
	TraceEnable     bool   `toml:"trace_enable"`
	MetricEnable    bool   `toml:"metric_enable"`
	Addr            string `toml:"addr"`
	ExpectedHeaders map[string]string
	stopFunc        func()
}

func (o *otlpGrpcCollector) run(storage *collector.SpansStorage) {
	ln, err := net.Listen("tcp", o.Addr)
	if err != nil {
		log.Errorf("Failed to get an endpoint: %v", err)

		return
	}

	localCache.RegisterConsumer(istorage.OTEL_GRPC_KEY, func(buf []byte) error {
		trace := &tracepb.TracesData{}
		if err := proto.Unmarshal(buf, trace); err != nil {
			return err
		}
		spanStorage.AddSpans(trace.ResourceSpans)

		return nil
	})

	srv := grpc.NewServer()
	if o.TraceEnable {
		et := &ExportTrace{storage: storage}
		collectortracepb.RegisterTraceServiceServer(srv, et)
	}
	if o.MetricEnable {
		em := &ExportMetric{storage: storage}
		collectormetricepb.RegisterMetricsServiceServer(srv, em)
	}
	o.stopFunc = srv.Stop

	if err = srv.Serve(ln); err != nil {
		log.Error(err.Error())
	}
}

func (o *otlpGrpcCollector) stop() {
	if o.stopFunc != nil {
		o.stopFunc()
	}
}

type ExportTrace struct {
	collectortracepb.UnimplementedTraceServiceServer
	ExpectedHeaders map[string]string
	storage         *collector.SpansStorage
}

func (et *ExportTrace) Export(ctx context.Context, ets *collectortracepb.ExportTraceServiceRequest) (*collectortracepb.ExportTraceServiceResponse, error) { //nolint: lll
	md, haveHeader := metadata.FromIncomingContext(ctx)
	if haveHeader {
		if !checkHandler(et.ExpectedHeaders, md) {
			return nil, fmt.Errorf("invalid request haeders or nil headers")
		}
	}

	if rss := ets.GetResourceSpans(); len(rss) > 0 {
		if localCache == nil {
			et.storage.AddSpans(rss)
		} else {
			buf, err := proto.Marshal(&tracepb.TracesData{ResourceSpans: rss})
			if err != nil {
				log.Error(err.Error())

				return nil, err
			}

			if err = localCache.Put(istorage.OTEL_GRPC_KEY, buf); err != nil {
				log.Error(err.Error())

				return nil, err
			}
		}
	}

	return &collectortracepb.ExportTraceServiceResponse{}, nil
}

type ExportMetric struct {
	collectormetricepb.UnimplementedMetricsServiceServer
	ExpectedHeaders map[string]string
	storage         *collector.SpansStorage
}

func (em *ExportMetric) Export(ctx context.Context, ets *collectormetricepb.ExportMetricsServiceRequest) (*collectormetricepb.ExportMetricsServiceResponse, error) { //nolint: lll
	md, haveHeader := metadata.FromIncomingContext(ctx)
	if haveHeader {
		if !checkHandler(em.ExpectedHeaders, md) {
			return nil, fmt.Errorf("invalid request haeders or nil headers")
		}
	}

	if rss := ets.ResourceMetrics; len(rss) > 0 {
		orms := em.storage.ToDatakitMetric(rss)
		em.storage.AddMetric(orms)
	}

	return &collectormetricepb.ExportMetricsServiceResponse{}, nil
}

func checkHandler(headers map[string]string, md metadata.MD) bool {
	if len(headers) == 0 {
		return true
	}
	for k, v := range headers {
		strs := md.Get(strings.ToLower(k))
		mdVal := strings.Join(strs, ",")
		if mdVal != v {
			return false
		}
	}

	return true
}
