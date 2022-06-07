// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package sinkotel is for opentemetry export to Jaeger
package sinkotel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/jaeger"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

const (
	jaegerID = "jaeger"
)

var (
	defScheme = "http"
	defHost   = "localhost"
	// 14268  collector: accept "jaeger thrift" directly from clients
	// 6831  agent: port
	defPort = "14268"
	defPath = "/api/traces"
)

// sinkJaeger: use otel jaeger.exporter
type sinkJaeger struct {
	id     string
	scheme string
	host   string
	port   string
	path   string

	// exporter
	exp tracesdk.SpanExporter
}

func (s *sinkJaeger) LoadConfig(mConf map[string]interface{}) error {
	l = logger.SLogger("jarger")

	if id, err := dkstring.GetMapMD5String(mConf); err != nil {
		return err
	} else {
		s.id = id
	}

	if scheme, err := dkstring.GetMapAssertString("scheme", mConf); err != nil {
		return err
	} else {
		addrNew, err := dkstring.CheckNotEmpty(scheme, "scheme")
		if err != nil {
			s.scheme = defScheme
		} else {
			s.scheme = addrNew
		}
	}

	if addr, err := dkstring.GetMapAssertString("host", mConf); err != nil {
		return err
	} else {
		addrNew, err := dkstring.CheckNotEmpty(addr, "host")
		if err != nil {
			s.host = defHost
		} else {
			s.host = addrNew
		}
	}

	if addr, err := dkstring.GetMapAssertString("port", mConf); err != nil {
		return err
	} else {
		addrNew, err := dkstring.CheckNotEmpty(addr, "port")
		if err != nil {
			s.port = defPort
		} else {
			s.port = addrNew
		}
	}

	if path, err := dkstring.GetMapAssertString("path", mConf); err != nil {
		return err
	} else {
		addrNew, err := dkstring.CheckNotEmpty(path, "path")
		if err != nil {
			s.path = defPath
		} else {
			s.path = addrNew
		}
	}
	switch s.scheme {
	case "http":
		u := "http://" + s.host + ":" + s.port + s.path
		exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(u)))
		if err != nil {
			return fmt.Errorf("jaeger new export err = %w", err)
		}
		s.exp = exp
	case "grpc":
		exp, err := jaeger.New(jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(s.host),
			jaeger.WithAgentPort(s.port),
		))
		if err != nil {
			l.Errorf("jaeger new export err =%v", err)
			return err
		}
		s.exp = exp
	default:
		return fmt.Errorf("the protocol must be HTTP or UDP")
	}

	l.Infof("init {jaeger = %+v } ok", s)
	sinkcommon.AddImpl(s)
	return nil
}

func (s *sinkJaeger) Write(pts []sinkcommon.ISinkPoint) error {
	spans := pointToTrace(pts)
	err := s.exp.ExportSpans(context.Background(), spans)
	if err != nil {
		l.Errorf("export span to remote error :%v", err)
	}
	return err
}

func (s *sinkJaeger) GetInfo() *sinkcommon.SinkInfo {
	return &sinkcommon.SinkInfo{
		ID:         s.id,
		CreateID:   jaegerID,
		Categories: []string{datakit.SinkCategoryTracing},
	}
}

func init() { //nolint:gochecknoinits
	sinkcommon.AddCreator(jaegerID, func() sinkcommon.ISink {
		return &sinkJaeger{}
	})
}
