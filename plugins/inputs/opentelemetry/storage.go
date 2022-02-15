// Package opentelemetry storage

package opentelemetry

import (
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	DKtrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

// SpansStorage stores the spans
type SpansStorage struct {
	rsm       []DKtrace.DatakitTrace
	spanCount int
	max       chan int
	stop      chan struct{}
}

// NewSpansStorage creates a new spans storage.
func NewSpansStorage() SpansStorage {
	return SpansStorage{
		rsm: make([]DKtrace.DatakitTrace, 0),
		max: make(chan int, 1),
	}
}

// AddSpans adds spans to the spans storage.
func (s *SpansStorage) AddSpans(rss []*tracepb.ResourceSpans) {
	traces := mkDKTrace(rss)
	s.rsm = append(s.rsm, traces...)
	s.spanCount += len(traces)
	if s.spanCount >= maxSend {
		s.max <- 0
	}
}

func setTag(tags map[string]string, attr []*commonpb.KeyValue) map[string]string {
	for _, kv := range attr {
		key := replace(kv.Key)
		tags[key] = kv.GetValue().String()
	}
	return tags
}

// GetResourceSpans returns the stored resource spans.
func (s *SpansStorage) getDKTrace() []DKtrace.DatakitTrace {
	rss := make([]DKtrace.DatakitTrace, 0, len(s.rsm))
	rss = append(rss, s.rsm...)
	return rss
}

func (s *SpansStorage) getCount() int {
	return s.spanCount
}

func (s *SpansStorage) run() {
	// 定时发送 或者长度超过100
	for {
		select {
		case <-s.max:
			traces := s.getDKTrace()
			for _, trace := range traces {
				afterGather.Run(inputName, trace, false)
			}
			s.reset()
		case <-time.After(time.Duration(interval) * time.Second):
			if s.getCount() > 0 {
				traces := s.getDKTrace()
				for _, trace := range traces {
					afterGather.Run(inputName, trace, false)
				}
				s.reset()
			}
		case <-s.stop:
			return
		}
	}
}

func (s *SpansStorage) reset() {
	// 归零
	s.spanCount = 0
	s.rsm = s.rsm[:0]
}

func mkDKTrace(rss []*tracepb.ResourceSpans) []DKtrace.DatakitTrace {
	dkTraces := make([]DKtrace.DatakitTrace, 0)
	for _, spans := range rss {
		ls := spans.GetInstrumentationLibrarySpans()

		// opentelemetry/filter.go:15      resource = attributes:{key:"service.name" value:{string_value:"test-service"}}
		l.Infof("resource = %s", spans.Resource.String())

		service := getServiceName(spans.Resource.Attributes)
		for _, librarySpans := range ls {
			l.Infof("librarySpans.InstrumentationLibrary.Name = %s", librarySpans.InstrumentationLibrary.Name)
			l.Infof("librarySpans.InstrumentationLibrary.Version = %s", librarySpans.InstrumentationLibrary.Version)
			l.Infof("schemaurl = %s", librarySpans.SchemaUrl)
			dktrace := make([]*DKtrace.DatakitSpan, 0)
			for _, span := range librarySpans.Spans {
				tags := toDatakitTags(span.Attributes)
				tags = setTag(tags, spans.Resource.Attributes)
				dkSpan := &DKtrace.DatakitSpan{
					TraceID:            hex.EncodeToString(span.GetTraceId()),
					ParentID:           byteToInt64(span.GetParentSpanId()),
					SpanID:             byteToInt64(span.GetSpanId()),
					Service:            service,
					Resource:           librarySpans.InstrumentationLibrary.Name,
					Operation:          span.Name,
					Source:             inputName,
					SpanType:           span.Kind.String(),
					SourceType:         "",
					Env:                "",
					Project:            "",
					Version:            librarySpans.InstrumentationLibrary.Version,
					Tags:               tags,
					EndPoint:           "",
					HTTPMethod:         "",
					HTTPStatusCode:     "",
					ContainerHost:      "",
					PID:                "",
					Start:              int64(span.StartTimeUnixNano),                        //  注意单位 nano
					Duration:           int64(span.EndTimeUnixNano - span.StartTimeUnixNano), // 单位 nano
					Status:             getDKSpanStatus(span.GetStatus().Code),               // 使用 dk status
					Content:            "",
					Priority:           0,
					SamplingRateGlobal: 0,
				}
				bts, err := json.Marshal(dkSpan)
				if err == nil {
					dkSpan.Content = string(bts)
				}
				l.Infof("dkspan = %+v", dkSpan)
				l.Infof("span = %+v", span)
				dktrace = append(dktrace, dkSpan)
			}
			dkTraces = append(dkTraces, dktrace)
		}
	}
	return dkTraces
}

// toDatakitTags : make attributes to tags
func toDatakitTags(attr []*commonpb.KeyValue) map[string]string {
	m := make(map[string]string, len(attr))
	for _, kv := range attr {
		key := replace(kv.Key)
		m[key] = kv.GetValue().GetStringValue()
		/*
			switch kv.GetValue().Value.(type) {
			// For slice attributes, serialize as JSON list string.
			case *v1.AnyValue_StringValue:
				m[kv.Key] = kv.GetValue().GetStringValue()
			case *v1.AnyValue_BoolValue:
			case *v1.AnyValue_IntValue:
			case *v1.AnyValue_DoubleValue:
			case *v1.AnyValue_ArrayValue:
			case *v1.AnyValue_KvlistValue:
			case *v1.AnyValue_BytesValue:

			default:
				m[(string)(kv.Key)] = kv.Value.Emit()
			}
		*/
	}

	return m
}

func byteToInt64(bts []byte) string {
	string16 := hex.EncodeToString(bts)
	n, err := strconv.ParseUint(string16, 16, 64)
	if err != nil {
		return string16
	}
	return strconv.FormatUint(n, 10)
}

// getDKSpanStatus 从otel的status转成dk的span_status
func getDKSpanStatus(code tracepb.Status_StatusCode) string {
	status := DKtrace.STATUS_INFO
	switch code {
	case tracepb.Status_STATUS_CODE_UNSET:
		status = DKtrace.STATUS_INFO
	case tracepb.Status_STATUS_CODE_OK:
		status = DKtrace.STATUS_OK
	case tracepb.Status_STATUS_CODE_ERROR:
		status = DKtrace.STATUS_ERR
	default:
	}
	return status
}

func getServiceName(attr []*commonpb.KeyValue) string {
	for _, kv := range attr {
		if kv.Key == "service.name" {
			return kv.Value.GetStringValue()
			/*if stringVal, ok := kv.Value.Value.(*commonpb.AnyValue_StringValue); ok {
				return stringVal.StringValue
			}*/
		}
	}
	return ""
}

// replace 行协议中的tag的key禁止点 全部替换掉
func replace(key string) string {
	return strings.ReplaceAll(key, ".", "")
}
