// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package sinkotel is for opentemetry
package sinkotel

import (
	"context"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/sinkfuncs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

const (
	creatorOtelID = "otel"
	defaultScheme = "grpc"
	defaultHost   = "localhost"
	defaultPath   = "/v1/traces"
	defaultPort   = "4317"
)

var l = logger.DefaultSLogger("sink-otel")

type otelSink struct {
	id     string
	IDStr  string // MD5 origin string.
	scheme string
	host   string // 包含 host
	port   string
	path   string // http url path

	// export
	exp tracesdk.SpanExporter
}

func (s *otelSink) LoadConfig(mConf map[string]interface{}) error {
	l = logger.SLogger("sink-otel")

	if id, str, err := sinkfuncs.GetSinkCreatorID(mConf); err != nil {
		return err
	} else {
		s.id = id
		s.IDStr = str
	}
	// DK_SINK_T="otel://localhost:7201?scheme=http"
	if scheme, err := dkstring.GetMapAssertString("scheme", mConf); err != nil {
		return err
	} else {
		addrNew, err := dkstring.CheckNotEmpty(scheme, "scheme")
		if err != nil {
			s.scheme = defaultScheme
		} else {
			s.scheme = addrNew
		}
	}

	if addr, err := dkstring.GetMapAssertString("host", mConf); err != nil {
		return err
	} else {
		addrNew, err := dkstring.CheckNotEmpty(addr, "host")
		if err != nil {
			s.host = defaultHost
		} else {
			s.host = addrNew
		}
	}

	if path, err := dkstring.GetMapAssertString("path", mConf); err != nil {
		return err
	} else {
		addrNew, err := dkstring.CheckNotEmpty(path, "path")
		if err != nil {
			s.path = defaultPath
		} else {
			s.path = addrNew
		}
	}
	if port, err := dkstring.GetMapAssertString("port", mConf); err != nil {
		return err
	} else {
		addrNew, err := dkstring.CheckNotEmpty(port, "port")
		if err != nil {
			s.port = defaultPort
		} else {
			s.port = addrNew
		}
	}
	switch s.scheme {
	case "http":
		s.exp = newHTTPExport(context.Background(), s.path, s.host)
	case "grpc":
		s.exp = newGrpcExport(context.Background(), s.host)
	default:
		return fmt.Errorf("the protocol must be http or grpc")
	}
	l.Infof("init {otel = %+v } ok", s)
	sinkcommon.AddImpl(s)
	return nil
}

func (s *otelSink) Write(category string, pts []*point.Point) error {
	spans := pointToTrace(pts)
	err := s.exp.ExportSpans(context.Background(), spans)
	if err != nil {
		l.Errorf("export span to remote error :%v", err)
	}
	return err
}

func (s *otelSink) GetInfo() *sinkcommon.SinkInfo {
	return &sinkcommon.SinkInfo{
		ID:         s.id,
		IDStr:      s.IDStr,
		CreateID:   creatorOtelID,
		Categories: []string{datakit.SinkCategoryTracing},
	}
}

// newGrpcExport is use http, send msg to remote.
func newHTTPExport(cxt context.Context, path string, endpoint string) *otlptrace.Exporter {
	httpClient, err := otlptracehttp.New(
		cxt,
		otlptracehttp.WithURLPath(path),
		otlptracehttp.WithEndpoint(endpoint), // host:port
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		l.Errorf("error to create a new collector exporter: %v", err)
		return nil
	}
	return httpClient
}

// newGrpcExport use tcp.
func newGrpcExport(cxt context.Context, endpoint string) *otlptrace.Exporter {
	grpcClient, err := otlptracegrpc.New(
		cxt,
		otlptracegrpc.WithTimeout(time.Second*30),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithReconnectionPeriod(50*time.Millisecond),
	)
	if err != nil {
		l.Errorf("error to create a new collector exporter: %v", err)
		return nil
	}
	return grpcClient
}

func init() { //nolint:gochecknoinits
	sinkcommon.AddCreator(creatorOtelID, func() sinkcommon.ISink {
		return &otelSink{}
	})
}
