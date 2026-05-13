// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"encoding/binary"
	"encoding/hex"
	"strconv"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"

	"github.com/GuanceCloud/cliutils/otlp"
	"github.com/GuanceCloud/cliutils/point"
	common "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
	trace "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/trace/v1"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gopkg.in/CodapeWild/dd-trace-go.v1/ddtrace/ext"
)

func (ipt *Input) parseResourceSpans(resspans []*trace.ResourceSpans, remoteIP string) itrace.DatakitTraces {
	ipt.ensureTraceParser()

	opts := otlp.TracesParserOptions{
		PointName:         inputName,
		Source:            inputName,
		CollectorSourceIP: remoteIP,
		DKFingerprint:     datakit.DKHost + "_" + datakit.Version,
		GlobalTags:        ipt.Tags,
		IDConverter:       ipt.convertBinID,
		SelectAttrs:       ipt.selectTraceAttrs,
		SpanType:          itrace.FindSpanTypeStrSpanID,
		SpanStatus:        getDKSpanStatus,
		SpanKind: func(kind int32) string {
			return spanKinds[kind]
		},
		SourceType:     getSourceType,
		BaseService:    ipt.traceBaseService,
		DecorateKVs:    ipt.decorateTraceKVs,
		MessageEncoder: ipt.traceMessageEncoder,
	}
	if ipt.CleanMessage {
		opts.CleanSpan = ipt.cleanTraceSpanForMessage
	}

	batches := otlp.ParseResourceSpans(resspans, opts)
	dktraces := make(itrace.DatakitTraces, 0, len(batches))
	values := make([]string, len(ipt.labels))

	for _, batch := range batches {
		dktrace := make(itrace.DatakitTrace, 0, len(batch))
		for _, pt := range batch {
			values = values[:0]
			if ipt.TracingMetricEnable {
				spanMetrics(pt, ipt.labels, values)
			}
			dktrace = append(dktrace, &itrace.DkSpan{Point: pt})
		}
		if len(dktrace) != 0 {
			dktraces = append(dktraces, dktrace)
		}
	}

	return dktraces
}

func (ipt *Input) ensureTraceParser() {
	if ipt.customTagsX == nil {
		ipt.customTagsX = itrace.NewCustomTags(ipt.CustomerTags, otelPubAttrs)
	}
	if ipt.jmarshaler == nil {
		ipt.jmarshaler = &protojsonMarshaler{}
	}
}

func (ipt *Input) selectTraceAttrs(attrs []*common.KeyValue) (point.KVs, []*common.KeyValue) {
	if !ipt.CleanMessage {
		return ipt.selectAttrs(attrs)
	}

	filtered := make([]*common.KeyValue, 0, len(attrs))
	for _, attr := range attrs {
		if attr == nil {
			continue
		}
		switch attr.GetKey() {
		case otelResourceServiceKey, itrace.FieldRuntimeID:
			continue
		default:
			filtered = append(filtered, attr)
		}
	}

	return ipt.selectAttrs(filtered)
}

func (ipt *Input) traceBaseService(attrs map[string]string) string {
	if !ipt.SplitServiceName {
		return ""
	}
	for _, key := range []string{"db.system", "rpc.system", "messaging.system"} {
		if system := attrs[key]; system != "" {
			return system
		}
	}

	return ""
}

func (ipt *Input) decorateTraceKVs(kvs point.KVs, _ *trace.ResourceSpans, _ *trace.ScopeSpans, _ *trace.Span) point.KVs {
	if kv := kvs.Get(itrace.FieldRuntimeID); kv != nil {
		if runtimeID, ok := kv.Raw().(string); ok && runtimeID != "" {
			kvs = kvs.Set(ext.RuntimeID, runtimeID) // NOTE: ext.RuntimeID deprecated.
		}
	}

	if ipt.Tagger != nil {
		for key, value := range ipt.Tagger.HostTags() {
			kvs = kvs.AddTag(key, value)
		}
	}

	return kvs
}

func (ipt *Input) traceMessageEncoder(span *trace.Span) (string, error) {
	if ipt.DelMessage || span == nil {
		return "", nil
	}

	buf, err := ipt.jmarshaler.Marshal(span)
	if err != nil {
		log.Warn(err.Error())
		return "", err
	}

	return string(buf), nil
}

func (ipt *Input) cleanTraceSpanForMessage(span *trace.Span) *trace.Span {
	if span == nil {
		return nil
	}
	for idx, event := range span.GetEvents() {
		if event != nil && event.GetName() == ExceptionEventName {
			span.Events[idx] = nil
			break
		}
	}

	return ipt.cleanSpan(span)
}

// cleanSpan try to remove fields of the span and make the marshaled json smaller.
func (ipt *Input) cleanSpan(span *trace.Span) *trace.Span {
	span.TraceId = nil
	span.SpanId = nil
	span.ParentSpanId = nil
	span.Name = ""
	span.Kind = trace.Span_SPAN_KIND_UNSPECIFIED
	span.StartTimeUnixNano = 0
	span.EndTimeUnixNano = 0
	span.Status = nil

	return span
}

func byteToString(buf []byte) string {
	if len(buf) == 0 || string(buf) == "0" || isZeroID(buf) {
		return "0"
	}

	return hex.EncodeToString(buf)
}

func isZeroID(buf []byte) bool {
	for _, b := range buf {
		if b != 0 {
			return false
		}
	}

	return true
}

func (ipt *Input) convertBinID(id []byte) string {
	switch {
	case ipt.CompatibleDDTrace:
		if len(id) >= 8 {
			bts := id[len(id)-8:]
			num := binary.BigEndian.Uint64(bts[:8])
			return strconv.FormatUint(num, 10)
		} else {
			log.Debugf("traceid or spanid is %s ,can not convert to [8]byte", string(id))
			return "0"
		}

	case ipt.CompatibleZhaoShang:
		if len(id) > 8 {
			return string(id)
		} else {
			return byteToString(id)
		}

	default:
		return byteToString(id)
	}
}

// getDKSpanStatus 从otel的status转成dk的span_status.
func getDKSpanStatus(statuspb *trace.Status) string {
	status := itrace.StatusInfo
	if statuspb == nil {
		return status
	}
	switch statuspb.Code {
	case trace.Status_STATUS_CODE_OK, trace.Status_STATUS_CODE_UNSET:
		status = itrace.StatusOk
	case trace.Status_STATUS_CODE_ERROR:
		status = itrace.StatusErr
	default:
	}

	return status
}

func getSourceType(tags point.KVs) string {
	var hasDB, hasMessaging, hasWeb bool

	for _, v := range tags {
		switch v.Key {
		case otelHTTPSchemeKey, otelHTTPMethodKey, otelRPCSystemKey:
			hasWeb = true
		case otelDBSystemKey:
			hasDB = true
		case otelMessagingSystemKey:
			hasMessaging = true
		}
	}

	switch {
	case hasDB:
		return itrace.SpanSourceDb
	case hasMessaging:
		return itrace.SpanSourceMsgque
	case hasWeb:
		return itrace.SpanSourceWeb
	}

	return itrace.SpanSourceCustomer
}
