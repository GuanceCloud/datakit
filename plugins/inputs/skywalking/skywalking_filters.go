package skywalking

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/v2/common"
	swV2 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/v2/language-agent-v2"
)

type upstmSegmentFilter func(segment *swV2.UpstreamSegment) *swV2.UpstreamSegment

// func upstmSegSample(segment *swV2.UpstreamSegment) *swV2.UpstreamSegment {
// 	var rootSpan *swV2.SpanObjectV2
// 	for _, span := range segment.Segment.Spans {
// 		if span.ParentSpanId == 0 {
// 			rootSpan = span
// 			break
// 		}
// 	}
// 	if rootSpan == nil {
// 		log.Debugf("can not find root span in upstream segment %v", segment)

// 		return segment
// 	}

// 	sconf := trace.SampleConfMatcher(sampleConfs, convertKVTags(rootSpan.Tags))
// 	if sconf == nil {
// 		log.Debug("can not find sample config")

// 		return segment
// 	}
// 	log.Debug(*sconf)

// 	for _, span := range segment.Segment.Spans {
// 		// bypass error
// 		if span.IsError {
// 			return segment
// 		}
// 		// bypass ignore keys
// 		if trace.SampleIgnoreKeys(convertKVTags(span.Tags), sconf.IgnoreTagsList) {
// 			return segment
// 		}
// 	}
// 	// do sample
// 	if trace.DefSampleFunc(uint64(segment.GlobalTraceIds[0].IdParts[2]), sconf.Rate, sconf.Scope) {
// 		return segment
// 	} else {
// 		return nil
// 	}
// }

func convertKVTags(sktags []*common.KeyStringValuePair) map[string]string {
	var tags = map[string]string{}
	for _, sktag := range sktags {
		tags[sktag.Key] = sktag.Value
	}

	return tags
}

type swSegmentFilter func(segment *SkyWalkingSegment) *SkyWalkingSegment

// func swSegSample(segment *SkyWalkingSegment) *SkyWalkingSegment {
// 	var rootSpan *SkyWalkingSpan
// 	for _, span := range segment.Spans {
// 		if span.ParentSpanId == 0 {
// 			rootSpan = span
// 		}
// 	}
// 	if rootSpan == nil {
// 		log.Debugf("can not find root span in skywalk segment %v", *segment)

// 		return segment
// 	}

// 	sconf := trace.SampleConfMatcher(sampleConfs, convertSkywalkTags(rootSpan.Tags))
// 	if sconf == nil {
// 		log.Debug("can not find sample config")

// 		return segment
// 	}
// 	log.Debug(*sconf)

// 	for _, span := range segment.Spans {
// 		// bypass error
// 		if span.IsError {
// 			return segment
// 		}
// 		// bypass ignore keys
// 		if trace.SampleIgnoreKeys(convertSkywalkTags(span.Tags), sconf.IgnoreTagsList) {
// 			return segment
// 		}
// 	}
// 	// do sample
// 	traceid, err := strconv.ParseInt(segment.TraceId, 10, 64)
// 	if err != nil {
// 		log.Debug(err)

// 		return segment
// 	}
// 	if trace.DefSampleFunc(uint64(traceid), sconf.Rate, sconf.Scope) {
// 		return segment
// 	} else {
// 		return nil
// 	}
// }

func convertSkywalkTags(swtags []*SkyWalkingTag) map[string]string {
	var tags map[string]string
	for _, swtag := range swtags {
		tags[swtag.Key] = fmt.Sprintf("%v", swtag.Value)
	}

	return tags
}
