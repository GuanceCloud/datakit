package traceJaeger

import (
	"fmt"

	"github.com/uber/jaeger-client-go/thrift-gen/jaeger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

type batchFilter func(*jaeger.Batch) *jaeger.Batch

func sample(batch *jaeger.Batch) *jaeger.Batch {
	var rootSpan *jaeger.Span
	for _, span := range batch.Spans {
		if span.ParentSpanId == 0 {
			rootSpan = span
			break
		}
	}
	if rootSpan == nil {
		log.Debugf("can not find root span in batch %v", batch)

		return batch
	}

	sconf := trace.SampleConfMatcher(sampleConfs, convertTags(rootSpan.Tags))
	if sconf == nil {
		log.Debug("can not find sample config")

		return batch
	}
	log.Debug(*sconf)

	for _, span := range batch.Spans {
		// bypass error
		for _, jtag := range span.Tags {
			if jtag.Key == "error" {
				return batch
			}
		}
		// bypass ignore keys
		if trace.SampleIgnoreKeys(convertTags(span.Tags), sconf.IgnoreTagsList) {
			return batch
		}
	}
	// do sample
	if trace.DefSampleFunc(uint64(trace.GetTraceId(rootSpan.TraceIdHigh, rootSpan.TraceIdLow)), sconf.Rate, sconf.Scope) {
		return batch
	} else {
		return nil
	}
}

func convertTags(jtags []*jaeger.Tag) map[string]string {
	tags := map[string]string{}
	for _, jtag := range jtags {
		tags[jtag.Key] = fmt.Sprintf("%v", getValueString(jtag))
	}

	return tags
}

func getValueString(tag *jaeger.Tag) interface{} {
	switch tag.VType {
	case jaeger.TagType_STRING:
		return *(tag.VStr)
	case jaeger.TagType_DOUBLE:
		return *(tag.VDouble)
	case jaeger.TagType_BOOL:
		return *(tag.VBool)
	case jaeger.TagType_LONG:
		return *(tag.VLong)
	case jaeger.TagType_BINARY:
		return tag.VBinary
	default:
		return nil
	}
}
