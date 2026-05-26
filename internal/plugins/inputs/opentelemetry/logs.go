// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"github.com/GuanceCloud/cliutils/otlp"
	"github.com/GuanceCloud/cliutils/point"
	common "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
	logs "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/logs/v1"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

func (ipt *Input) parseLogRequest(resourceLogss []*logs.ResourceLogs, remoteIP string) []*point.Point {
	globalFields := make(map[string]any, len(ipt.Tags))
	for key, value := range ipt.Tags {
		globalFields[key] = value
	}

	stringOpts := defaultOTLPStringMapOptions()
	fingerprint := datakit.DKHost + "_" + datakit.Version
	pts := otlp.ParseLogRequest(resourceLogss, otlp.LogsParserOptions{
		CollectorSourceIP:     remoteIP,
		DKFingerprint:         fingerprint,
		ResourceStringOptions: stringOpts,
		ScopeStringOptions:    stringOpts,
		RecordStringOptions:   stringOpts,
		MaxMessageLen:         ipt.LogMaxLen * 1024,
		GlobalFields:          globalFields,
		IDConverter:           ipt.convertBinID,
		ServiceAndSource:      getServiceAndSource,
		HostName:              getHostName,
		SeverityMapper:        getStatus,
	})

	for _, pt := range pts {
		pt.AddTag(itrace.TagCollectorSourceIP, remoteIP)
		pt.AddTag(itrace.TagDKFingerprintKey, fingerprint)
	}

	return pts
}

func defaultOTLPStringMapOptions() otlp.StringMapOptions {
	dropKeys := make(map[string]struct{}, len(delMetricKey))
	for _, key := range delMetricKey {
		dropKeys[key] = struct{}{}
	}

	return otlp.StringMapOptions{
		MaxValueLen: maxLogMetricFiledLen,
		DropKeys:    dropKeys,
	}
}

func splitByByteLength(s string, length int) []string {
	runes := []rune(s)
	var chunks []string
	for i := 0; i < len(runes); i += length {
		end := i + length
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}
	return chunks
}

func getStatus(severityNum logs.SeverityNumber, level string) string {
	switch severityNum {
	case logs.SeverityNumber_SEVERITY_NUMBER_TRACE,
		logs.SeverityNumber_SEVERITY_NUMBER_TRACE2,
		logs.SeverityNumber_SEVERITY_NUMBER_TRACE3,
		logs.SeverityNumber_SEVERITY_NUMBER_TRACE4:
		return "trace"
	case logs.SeverityNumber_SEVERITY_NUMBER_DEBUG,
		logs.SeverityNumber_SEVERITY_NUMBER_DEBUG2,
		logs.SeverityNumber_SEVERITY_NUMBER_DEBUG3,
		logs.SeverityNumber_SEVERITY_NUMBER_DEBUG4:
		return "debug"
	case logs.SeverityNumber_SEVERITY_NUMBER_INFO,
		logs.SeverityNumber_SEVERITY_NUMBER_INFO2,
		logs.SeverityNumber_SEVERITY_NUMBER_INFO3,
		logs.SeverityNumber_SEVERITY_NUMBER_INFO4:
		return "info"
	case logs.SeverityNumber_SEVERITY_NUMBER_WARN,
		logs.SeverityNumber_SEVERITY_NUMBER_WARN2,
		logs.SeverityNumber_SEVERITY_NUMBER_WARN3,
		logs.SeverityNumber_SEVERITY_NUMBER_WARN4:
		return "warn"
	case logs.SeverityNumber_SEVERITY_NUMBER_ERROR,
		logs.SeverityNumber_SEVERITY_NUMBER_ERROR2,
		logs.SeverityNumber_SEVERITY_NUMBER_ERROR3,
		logs.SeverityNumber_SEVERITY_NUMBER_ERROR4:
		return "error"
	case logs.SeverityNumber_SEVERITY_NUMBER_FATAL,
		logs.SeverityNumber_SEVERITY_NUMBER_FATAL2,
		logs.SeverityNumber_SEVERITY_NUMBER_FATAL3,
		logs.SeverityNumber_SEVERITY_NUMBER_FATAL4:
		return "fatal"
	case logs.SeverityNumber_SEVERITY_NUMBER_UNSPECIFIED:
		return "unknown"
	}

	return level
}

func getServiceAndSource(attr []*common.KeyValue) (service string, source string) {
	for _, keyValue := range attr {
		if keyValue.GetKey() == otelResourceServiceKey {
			service = keyValue.GetValue().GetStringValue()
		}
		if keyValue.GetKey() == "log.source" {
			source = keyValue.GetValue().GetStringValue()
		}
	}

	if source == "" {
		if service == "" {
			source = "otel_logs"
		} else {
			source = service
		}
	}
	if service == "" {
		service = "unSetServiceName"
	}
	return
}

func getHostName(attr []*common.KeyValue) (hostName string) {
	for _, keyValue := range attr {
		if keyValue.GetKey() == "host.name" {
			hostName = keyValue.GetValue().GetStringValue()
			break
		}
	}

	return
}
