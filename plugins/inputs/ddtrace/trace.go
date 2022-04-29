// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ddtrace

//go:generate msgp -file=span.pb.go -o span_gen.go -io=false
//go:generate msgp -o trace_gen.go -io=false

import itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"

var ddtraceSpanType = map[string]string{
	"consul":        itrace.SPAN_SERVICE_APP,
	"cache":         itrace.SPAN_SERVICE_CACHE,
	"memcached":     itrace.SPAN_SERVICE_CACHE,
	"redis":         itrace.SPAN_SERVICE_CACHE,
	"aerospike":     itrace.SPAN_SERVICE_DB,
	"cassandra":     itrace.SPAN_SERVICE_DB,
	"db":            itrace.SPAN_SERVICE_DB,
	"elasticsearch": itrace.SPAN_SERVICE_DB,
	"leveldb":       itrace.SPAN_SERVICE_DB,
	"mongodb":       itrace.SPAN_SERVICE_DB,
	"sql":           itrace.SPAN_SERVICE_DB,
	"http":          itrace.SPAN_SERVICE_WEB,
	"web":           itrace.SPAN_SERVICE_WEB,
	"benchmark":     itrace.SPAN_SERVICE_CUSTOM,
	"build":         itrace.SPAN_SERVICE_CUSTOM,
	"custom":        itrace.SPAN_SERVICE_CUSTOM,
	"datanucleus":   itrace.SPAN_SERVICE_CUSTOM,
	"dns":           itrace.SPAN_SERVICE_CUSTOM,
	"graphql":       itrace.SPAN_SERVICE_CUSTOM,
	"grpc":          itrace.SPAN_SERVICE_CUSTOM,
	"hibernate":     itrace.SPAN_SERVICE_CUSTOM,
	"queue":         itrace.SPAN_SERVICE_CUSTOM,
	"rpc":           itrace.SPAN_SERVICE_CUSTOM,
	"soap":          itrace.SPAN_SERVICE_CUSTOM,
	"template":      itrace.SPAN_SERVICE_CUSTOM,
	"test":          itrace.SPAN_SERVICE_CUSTOM,
	"worker":        itrace.SPAN_SERVICE_CUSTOM,
}

func getDDTraceSourceType(spanType string) string {
	if t, ok := ddtraceSpanType[spanType]; ok {
		return t
	}

	return itrace.SPAN_SERVICE_UNKNOW
}

type DDTrace []*DDSpan

type DDTraces []DDTrace
