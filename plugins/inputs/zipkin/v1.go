package zipkin

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	zpkcorev1 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/zipkin/corev1"
)

func unmarshalZipkinThriftV1(body io.ReadCloser) ([]*zpkcorev1.Span, error) {
	buffer := thrift.NewTMemoryBuffer()
	_, err := buffer.ReadFrom(body)
	if err != nil {
		return nil, err
	}

	var (
		transport = thrift.NewTBinaryProtocolTransport(buffer)
		size      int
	)
	if _, size, err = transport.ReadListBegin(); err != nil {
		return nil, err
	}

	var spans []*zpkcorev1.Span
	for i := 0; i < size; i++ {
		zs := &zpkcorev1.Span{}
		if err = zs.Read(transport); err != nil {
			log.Error(err.Error())
			continue
		}
		spans = append(spans, zs)
	}

	return spans, transport.ReadListEnd()
}

func thriftV1SpansToDkTrace(zpktrace []*zpkcorev1.Span) itrace.DatakitTrace {
	var (
		dktrace            itrace.DatakitTrace
		parentIDs, spanIDs = gatherZpkCoreV1SpansInfo(zpktrace)
	)
	for _, span := range zpktrace {
		if span == nil {
			continue
		}

		if span.ParentID == nil {
			span.ParentID = new(int64)
		}
		service := getServiceFromZpkCoreV1Span(span)
		dkspan := &itrace.DatakitSpan{
			TraceID:   fmt.Sprintf("%x", uint64(span.TraceID)),
			ParentID:  "0",
			SpanID:    fmt.Sprintf("%x", uint64(span.ID)),
			Service:   service,
			Resource:  span.Name,
			Operation: span.Name,
			Source:    inputName,
			SpanType:  itrace.FindSpanTypeInMultiServersIntSpanID(span.ID, *span.ParentID, service, spanIDs, parentIDs),
		}

		if span.ParentID != nil {
			dkspan.ParentID = fmt.Sprintf("%x", uint64(*span.ParentID))
		}

		if span.Timestamp != nil {
			dkspan.Start = (*span.Timestamp) * int64(time.Microsecond)
		} else {
			dkspan.Start = getStartTimestamp(span)
		}

		if span.Duration != nil {
			dkspan.Duration = (*span.Duration) * int64(time.Microsecond)
		} else {
			dkspan.Duration = getDurationThriftAno(span.Annotations)
		}

		dkspan.Status = itrace.STATUS_OK
		if _, ok := findZpkCoreV1BinaryAnnotation(span.BinaryAnnotations, "error"); ok {
			dkspan.Status = itrace.STATUS_ERR
		}

		if resource, ok := findZpkCoreV1BinaryAnnotation(span.BinaryAnnotations, "path.http"); ok {
			dkspan.Resource = resource
		}

		if project, ok := findZpkCoreV1BinaryAnnotation(span.BinaryAnnotations, "project"); ok {
			dkspan.Project = project
		}

		if version, ok := findZpkCoreV1BinaryAnnotation(span.BinaryAnnotations, "version"); ok {
			dkspan.Version = version
		}

		sourceTags := make(map[string]string)
		for _, tag := range span.BinaryAnnotations {
			sourceTags[tag.Key] = string(tag.Value)
		}
		dkspan.Tags = itrace.MergeInToCustomerTags(customerKeys, tags, sourceTags)

		if dkspan.ParentID == "0" && sampler != nil {
			dkspan.Priority = sampler.Priority
			dkspan.SamplingRateGlobal = sampler.SamplingRateGlobal
		}

		if buf, err := json.Marshal(zipkinConvThriftToJSON(span)); err != nil {
			log.Warn(err.Error())
		} else {
			dkspan.Content = string(buf)
		}

		dktrace = append(dktrace, dkspan)
	}

	return dktrace
}

func gatherZpkCoreV1SpansInfo(trace []*zpkcorev1.Span) (parentIDs map[int64]bool, spanIDs map[int64]string) {
	parentIDs = make(map[int64]bool)
	spanIDs = make(map[int64]string)
	for _, span := range trace {
		if span == nil {
			continue
		}
		spanIDs[span.ID] = getServiceFromZpkCoreV1Span(span)
		if span.ParentID != nil && *span.ParentID != 0 {
			parentIDs[*span.ParentID] = true
		}
	}

	return
}

func getServiceFromZpkCoreV1Span(span *zpkcorev1.Span) string {
	for _, anno := range span.Annotations {
		if anno.Host != nil && anno.Host.ServiceName != "" {
			return anno.Host.ServiceName
		}
	}
	for _, banno := range span.BinaryAnnotations {
		if banno.Host != nil && banno.Host.ServiceName != "" {
			return banno.Host.ServiceName
		}
	}

	return "zipkin_core_v1_unknown_service"
}

func getStartTimestamp(zs *zpkcorev1.Span) int64 {
	var (
		ts      int64 = 0x7FFFFFFFFFFFFFFF
		isFound bool
	)
	for _, ano := range zs.Annotations {
		if ano.Timestamp == 0 {
			continue
		}
		if ano.Timestamp < ts {
			isFound = true
			ts = ano.Timestamp
		}
	}

	if isFound {
		return ts * 1000
	}

	return time.Now().UnixNano()
}

func getDurationThriftAno(anos []*zpkcorev1.Annotation) int64 {
	if len(anos) < 2 {
		return 0
	}

	var (
		start int64 = 0x7FFFFFFFFFFFFFFF
		stop  int64
	)
	for _, ano := range anos {
		if ano.Timestamp == 0 {
			continue
		}

		if ano.Timestamp > stop {
			stop = ano.Timestamp
		}
		if ano.Timestamp < start {
			start = ano.Timestamp
		}
	}

	if stop > start {
		return (stop - start) * int64(time.Microsecond)
	}

	return 0
}

func findZpkCoreV1BinaryAnnotation(bannos []*zpkcorev1.BinaryAnnotation, key string) (string, bool) {
	for _, banno := range bannos {
		if banno != nil && banno.AnnotationType == zpkcorev1.AnnotationType_STRING && banno.Key == key {
			return string(banno.Value), true
		}
	}

	return "", false
}

func zipkinConvThriftToJSON(span *zpkcorev1.Span) *zpkcorev1.SpanJsonApater {
	zc := &zpkcorev1.SpanJsonApater{
		TraceID: uint64(span.TraceID),
		Name:    span.Name,
		ID:      uint64(span.ID),
		Debug:   span.Debug,
	}

	if span.ParentID != nil {
		zc.ParentID = uint64(*span.ParentID)
	}

	for _, anno := range span.Annotations {
		janno := zpkcorev1.AnnotationJsonApater{
			Timestamp: uint64(anno.Timestamp),
			Value:     anno.Value,
		}
		if anno.Host != nil {
			ep := &zpkcorev1.EndpointJsonApater{
				ServiceName: anno.Host.ServiceName,
				Ipv4:        net.IP(int2ip(uint32(anno.Host.Ipv4))),
				Ipv6:        anno.Host.Ipv6,
				Port:        anno.Host.Port,
			}
			janno.Host = ep
		}
		zc.Annotations = append(zc.Annotations, janno)
	}

	for _, banno := range span.BinaryAnnotations {
		jbanno := zpkcorev1.BinaryAnnotationJsonApater{
			Key:            banno.Key,
			Value:          banno.Value,
			AnnotationType: banno.AnnotationType,
		}
		if banno.Host != nil {
			ep := &zpkcorev1.EndpointJsonApater{
				ServiceName: banno.Host.ServiceName,
				Ipv4:        net.IP(int2ip(uint32(banno.Host.Ipv4))),
				Ipv6:        banno.Host.Ipv6,
				Port:        banno.Host.Port,
			}
			jbanno.Host = ep
		}
		zc.BinaryAnnotations = append(zc.BinaryAnnotations, jbanno)
	}

	if span.Timestamp != nil {
		zc.Timestamp = uint64(*span.Timestamp)
	}

	if span.Duration != nil {
		zc.Duration = uint64(*span.Duration)
	}

	if span.TraceIDHigh != nil {
		zc.TraceIDHigh = uint64(*span.TraceIDHigh)
	}

	return zc
}

func int2ip(i uint32) []byte {
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, i)

	return bs
}

type Endpoint struct {
	ServiceName string `json:"serviceName"`
	Ipv4        string `json:"ipv4"`
	Ipv6        string `json:"ipv6,omitempty"`
	Port        int16  `json:"port"`
}

type Annotation struct {
	Timestamp int64     `json:"timestamp"`
	Value     string    `json:"value"`
	Host      *Endpoint `json:"endpoint,omitempty"`
}

type BinaryAnnotation struct {
	Key   string    `json:"key"`
	Value string    `json:"value"`
	Host  *Endpoint `json:"endpoint,omitempty"`
}

type ZipkinSpanV1 struct {
	TraceID           string              `thrift:"traceId,1" db:"traceId" json:"traceId"`
	ParentID          string              `thrift:"parentId,5" db:"parentId" json:"parentId,omitempty"`
	ID                string              `thrift:"id,4" db:"id" json:"id"`
	Name              string              `thrift:"name,3" db:"name" json:"name"`
	Annotations       []*Annotation       `thrift:"annotations,6" db:"annotations" json:"annotations"`
	BinaryAnnotations []*BinaryAnnotation `thrift:"binary_annotations,8" db:"binary_annotations" json:"binaryAnnotations"`
	Timestamp         int64               `thrift:"timestamp,10" db:"timestamp" json:"timestamp,omitempty"`
	Duration          int64               `thrift:"duration,11" db:"duration" json:"duration,omitempty"`
	Debug             bool                `thrift:"debug,9" db:"debug" json:"debug,omitempty"`
}

func jsonV1SpansToDkTrace(zpktrace []*ZipkinSpanV1) itrace.DatakitTrace {
	var (
		dktrace            itrace.DatakitTrace
		parentIDs, spanIDs = gatherZpkV1SpansInfo(zpktrace)
	)
	for _, span := range zpktrace {
		if span == nil {
			continue
		}

		service := getServiceFromZpkV1Span(span)
		dkspan := &itrace.DatakitSpan{
			TraceID:   span.TraceID,
			ParentID:  span.ParentID,
			SpanID:    span.ID,
			Service:   service,
			Resource:  span.Name,
			Operation: span.Name,
			Source:    inputName,
			SpanType:  itrace.FindSpanTypeInMultiServersStrSpanID(span.ID, span.ParentID, service, spanIDs, parentIDs),
			Start:     getFirstTimestamp(span),
			Duration:  span.Duration * int64(time.Microsecond),
		}

		if isRootSpan(dkspan.ParentID) {
			dkspan.ParentID = "0"
		}

		if dkspan.Duration == 0 {
			dkspan.Duration = getDurationByAno(span.Annotations)
		}

		dkspan.Status = itrace.STATUS_OK
		if _, ok := findZpkV1BinaryAnnotation(span.BinaryAnnotations, "error"); ok {
			dkspan.Status = itrace.STATUS_ERR
		}

		if resource, ok := findZpkV1BinaryAnnotation(span.BinaryAnnotations, "path.http"); ok {
			dkspan.Resource = resource
		}

		if project, ok := findZpkV1BinaryAnnotation(span.BinaryAnnotations, "project"); ok {
			dkspan.Project = project
		}

		if version, ok := findZpkV1BinaryAnnotation(span.BinaryAnnotations, "version"); ok {
			dkspan.Version = version
		}

		sourceTags := make(map[string]string)
		for _, tag := range span.BinaryAnnotations {
			sourceTags[tag.Key] = tag.Value
		}
		dkspan.Tags = itrace.MergeInToCustomerTags(customerKeys, tags, sourceTags)

		if dkspan.ParentID == "0" && sampler != nil {
			dkspan.Priority = sampler.Priority
			dkspan.SamplingRateGlobal = sampler.SamplingRateGlobal
		}

		if buf, err := json.Marshal(span); err != nil {
			continue
		} else {
			dkspan.Content = string(buf)
		}

		dktrace = append(dktrace, dkspan)
	}

	return dktrace
}

func gatherZpkV1SpansInfo(trace []*ZipkinSpanV1) (parentIDs map[string]bool, spanIDs map[string]string) {
	parentIDs = make(map[string]bool)
	spanIDs = make(map[string]string)
	for _, span := range trace {
		if span == nil {
			continue
		}
		spanIDs[span.ID] = getServiceFromZpkV1Span(span)
		if span.ParentID != "" && span.ParentID != "0" {
			parentIDs[span.ParentID] = true
		}
	}

	return
}

func getServiceFromZpkV1Span(span *ZipkinSpanV1) string {
	for _, anno := range span.Annotations {
		if anno.Host != nil && anno.Host.ServiceName != "" {
			return anno.Host.ServiceName
		}
	}
	for _, bno := range span.BinaryAnnotations {
		if bno.Host != nil && bno.Host.ServiceName != "" {
			return bno.Host.ServiceName
		}
	}

	return "zipkin_v1_unknown_service"
}

func getFirstTimestamp(zs *ZipkinSpanV1) int64 {
	var (
		ts      int64 = 0x7FFFFFFFFFFFFFFF
		isFound bool
	)
	for _, ano := range zs.Annotations {
		if ano.Timestamp == 0 {
			continue
		}
		if ano.Timestamp < ts {
			isFound = true
			ts = ano.Timestamp
		}
	}

	if isFound {
		return ts * 1000
	}

	return time.Now().UnixNano()
}

func getDurationByAno(anos []*Annotation) int64 {
	if len(anos) < 2 {
		return 0
	}

	var (
		startTS int64 = 0x7FFFFFFFFFFFFFFF
		stopTS  int64
	)
	for _, ano := range anos {
		if ano.Timestamp == 0 {
			continue
		}
		if ano.Timestamp > stopTS {
			stopTS = ano.Timestamp
		}

		if ano.Timestamp < startTS {
			startTS = ano.Timestamp
		}
	}
	if stopTS > startTS {
		return (stopTS - startTS) * 1000
	}

	return 0
}

func findZpkV1BinaryAnnotation(bannos []*BinaryAnnotation, key string) (string, bool) {
	for _, banno := range bannos {
		if banno != nil && banno.Key == key {
			return banno.Value, true
		}
	}

	return "", false
}

func isRootSpan(parentID string) bool {
	if len(parentID) == 0 || parentID == "0" {
		return true
	} else {
		if i, err := strconv.ParseInt(parentID, 10, 64); err != nil {
			return false
		} else {
			return i == 0
		}
	}
}
