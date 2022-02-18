package ddtrace

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tinylib/msgp/msgp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/msgpack"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
)

const (
	// KeySamplingPriority is the key of the sampling priority value in the metrics map of the root span.
	keyPriority = "_sampling_priority_v1"
	// keySamplingRateGlobal is a metric key holding the global sampling rate.
	keySamplingRateGlobal = "_sample_rate"
)

var validContentTypes = map[string]bool{"application/json": true, "text/json": true, "application/msgpack": true}

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
	"":              itrace.SPAN_SERVICE_CUSTOM,
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

//nolint:lll
type DDSpan struct {
	Service  string             `protobuf:"bytes,1,opt,name=service,proto3" json:"service" msg:"service" codec:"service"`
	Name     string             `protobuf:"bytes,2,opt,name=name,proto3" json:"name" msg:"name" codec:"name"`
	Resource string             `protobuf:"bytes,3,opt,name=resource,proto3" json:"resource" msg:"resource" codec:"resource"`
	TraceID  uint64             `protobuf:"varint,4,opt,name=traceID,proto3" json:"trace_id" msg:"trace_id" codec:"trace_id"`
	SpanID   uint64             `protobuf:"varint,5,opt,name=spanID,proto3" json:"span_id" msg:"span_id" codec:"span_id"`
	ParentID uint64             `protobuf:"varint,6,opt,name=parentID,proto3" json:"parent_id" msg:"parent_id" codec:"parent_id"`
	Start    int64              `protobuf:"varint,7,opt,name=start,proto3" json:"start" msg:"start" codec:"start"`
	Duration int64              `protobuf:"varint,8,opt,name=duration,proto3" json:"duration" msg:"duration" codec:"duration"`
	Error    int32              `protobuf:"varint,9,opt,name=error,proto3" json:"error" msg:"error" codec:"error"`
	Meta     map[string]string  `protobuf:"bytes,10,rep,name=meta" json:"meta" msg:"meta" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3" codec:"meta"`
	Metrics  map[string]float64 `protobuf:"bytes,11,rep,name=metrics" json:"metrics" msg:"metrics" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"fixed64,2,opt,name=value,proto3" codec:"metrics"`
	Type     string             `protobuf:"bytes,12,opt,name=type,proto3" json:"type" msg:"type" codec:"type"`
}

type DDTrace []*DDSpan

type DDTraces []DDTrace

func handleDDTraces(pattern string) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		log.Debugf("%s: received tracing data from path: %s", inputName, req.URL.Path)

		switch req.URL.Path {
		case v3, v4, v5:
		default:
			log.Errorf("unrecognized ddtrace endpoint: %s", req.URL.Path)
			resp.WriteHeader(http.StatusBadRequest)

			return
		}

		buf, err := io.ReadAll(req.Body)
		if err != nil {
			log.Error(err.Error())
			resp.WriteHeader(http.StatusBadRequest)

			return
		}

		contentType := strings.ToLower(strings.TrimSpace(req.Header.Get("Content-Type")))
		if !validContentTypes[contentType] {
			if json.Valid(buf) {
				contentType = "application/json"
			} else {
				log.Errorf("unrecognized Content-Type: %s", contentType)
				resp.WriteHeader(http.StatusBadRequest)

				return
			}
		}

		var traces DDTraces
		switch req.URL.Path {
		case v3, v4:
			if traces, err = decodeV3V4Request(contentType, buf); err != nil {
				if errors.Is(err, io.EOF) {
					log.Warn(err.Error())
					resp.WriteHeader(http.StatusOK)
				} else {
					log.Error(err.Error())
					resp.WriteHeader(http.StatusBadRequest)
				}

				return
			}
		case v5:
			if contentType != "application/msgpack" {
				log.Errorf("invalid Content-Type: %s for MsgPack", contentType)
				resp.WriteHeader(http.StatusBadRequest)

				return
			}
			if err = unmarshalTraceMsgDictionary(buf, &traces); err != nil {
				log.Error(err.Error())
				resp.WriteHeader(http.StatusBadRequest)

				return
			}
		}
		if len(traces) == 0 {
			log.Debug("empty ddtraces")
			resp.WriteHeader(http.StatusOK)

			return
		}

		for _, trace := range traces {
			if len(trace) == 0 {
				continue
			}

			if dktrace := ddtraceToDkTrace(trace); len(dktrace) == 0 {
				log.Warn("empty datakit trace")
			} else {
				afterGatherRun.Run(inputName, dktrace, false)
			}
		}

		resp.WriteHeader(http.StatusOK)
	}
}

// TODO:.
func handleDDInfo(resp http.ResponseWriter, req *http.Request) { //nolint: unused,deadcode
	log.Errorf("%s unsupport yet", req.URL.Path)
	resp.WriteHeader(http.StatusNotFound)
}

// TODO:.
func handleDDStats(resp http.ResponseWriter, req *http.Request) {
	log.Errorf("%s  unsupport yet", req.URL.Path)
	resp.WriteHeader(http.StatusNotFound)
}

func decodeV3V4Request(contentType string, buf []byte) (DDTraces, error) {
	var (
		traces = DDTraces{}
		err    error
	)
	switch contentType {
	case "application/msgpack":
		err = msgpack.Unmarshal(bytes.NewBuffer(buf), &traces)
	case "application/json", "text/json", "":
		err = json.NewDecoder(bytes.NewBuffer(buf)).Decode(&traces)
	default:
		err = errors.New("unrecognized Content-Type: " + contentType)
	}

	return traces, err
}

func ddtraceToDkTrace(trace DDTrace) itrace.DatakitTrace {
	var (
		dktrace            itrace.DatakitTrace
		spanIDs, parentIDs = getSpanIDsAndParentIDs(trace)
	)
	for _, span := range trace {
		if span == nil {
			continue
		}

		dkspan := &itrace.DatakitSpan{
			TraceID:            fmt.Sprintf("%d", span.TraceID),
			ParentID:           fmt.Sprintf("%d", span.ParentID),
			SpanID:             fmt.Sprintf("%d", span.SpanID),
			Service:            span.Service,
			Resource:           span.Resource,
			Operation:          span.Name,
			Source:             inputName,
			SpanType:           itrace.FindSpanTypeInt(int64(span.SpanID), int64(span.ParentID), spanIDs, parentIDs),
			SourceType:         ddtraceSpanType[span.Type],
			Tags:               itrace.MergeInToCustomerTags(customerKeys, tags, span.Meta),
			ContainerHost:      span.Meta[itrace.CONTAINER_HOST],
			PID:                fmt.Sprintf("%d", int64(span.Metrics["system.pid"])),
			HTTPMethod:         span.Meta["http.method"],
			HTTPStatusCode:     span.Meta["http.status_code"],
			Start:              span.Start,
			Duration:           span.Duration,
			SamplingRateGlobal: span.Metrics[keySamplingRateGlobal],
		}

		if span.Meta[itrace.PROJECT] != "" {
			dkspan.Project = span.Meta[itrace.PROJECT]
		} else {
			dkspan.Project = tags[itrace.PROJECT]
		}

		if span.Meta[itrace.ENV] != "" {
			dkspan.Env = span.Meta[itrace.ENV]
		} else {
			dkspan.Env = tags[itrace.ENV]
		}

		if span.Meta[itrace.VERSION] != "" {
			dkspan.Version = span.Meta[itrace.VERSION]
		} else {
			dkspan.Version = tags[itrace.VERSION]
		}

		dkspan.Status = itrace.STATUS_OK
		if span.Error != 0 {
			dkspan.Status = itrace.STATUS_ERR
		}

		if priority := int(span.Metrics[keyPriority]); priority <= 0 {
			dkspan.Priority = itrace.PriorityReject
		}

		if dkspan.ParentID == "0" && defSampler != nil {
			dkspan.Priority = defSampler.Priority
			dkspan.SamplingRateGlobal = defSampler.SamplingRateGlobal
		}

		if buf, err := json.Marshal(span); err != nil {
			log.Warn(err.Error())
		} else {
			dkspan.Content = string(buf)
		}

		dktrace = append(dktrace, dkspan)
	}

	return dktrace
}

func getSpanIDsAndParentIDs(trace DDTrace) (map[int64]bool, map[int64]bool) {
	var (
		spanIDs   = make(map[int64]bool)
		parentIDs = make(map[int64]bool)
	)
	for _, span := range trace {
		if span == nil {
			continue
		}
		spanIDs[int64(span.SpanID)] = true
		if span.ParentID != 0 {
			parentIDs[int64(span.ParentID)] = true
		}
	}

	return spanIDs, parentIDs
}

// unmarshalTraceMsgDictionary decodes a trace using the specification from the v0.5 endpoint.
func unmarshalTraceMsgDictionary(bts []byte, out *DDTraces) error {
	var err error
	if _, bts, err = msgp.ReadArrayHeaderBytes(bts); err != nil {
		log.Debug(err.Error())

		return err
	}
	// read dictionary
	var sz uint32
	if sz, bts, err = msgp.ReadArrayHeaderBytes(bts); err != nil {
		log.Debug(err.Error())

		return err
	}
	dict := make([]string, sz)
	for i := range dict {
		var str string
		str, bts, err = msgpack.ParseStringBytes(bts)
		if err != nil {
			log.Debug(err.Error())

			return err
		}
		dict[i] = str
	}
	// read traces
	sz, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		log.Debug(err.Error())

		return err
	}
	if cap(*out) >= int(sz) {
		*out = (*out)[:sz]
	} else {
		*out = make(DDTraces, sz)
	}
	for i := range *out {
		sz, bts, err = msgp.ReadArrayHeaderBytes(bts)
		if err != nil {
			log.Debug(err.Error())

			return err
		}
		if cap((*out)[i]) >= int(sz) {
			(*out)[i] = (*out)[i][:sz]
		} else {
			(*out)[i] = make(DDTrace, sz)
		}
		for j := range (*out)[i] {
			if (*out)[i][j] == nil {
				(*out)[i][j] = new(DDSpan)
			}
			if bts, err = unmarshalSpanMsgDictionary(bts, dict, (*out)[i][j]); err != nil {
				log.Debug(err.Error())

				return err
			}
		}
	}

	return nil
}

// spanPropertyCount specifies the number of top-level properties that a span
// has.
const spanPropertyCount = 12

// unmarshalSpanMsgDictionary decodes a span from the given decoder dc, looking up strings
// in the given dictionary dict.
func unmarshalSpanMsgDictionary(bts []byte, dict []string, out *DDSpan) ([]byte, error) {
	var (
		sz  uint32
		err error
	)
	sz, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		log.Debug(err.Error())

		return bts, err
	}
	if sz != spanPropertyCount {
		return bts, errors.New("encoded span needs exactly 12 elements in array")
	}
	// Service (0)
	out.Service, bts, err = dictionaryString(bts, dict)
	if err != nil {
		log.Debug(err.Error())

		return bts, err
	}
	// Name (1)
	out.Name, bts, err = dictionaryString(bts, dict)
	if err != nil {
		log.Debug(err.Error())

		return bts, err
	}
	// Resource (2)
	out.Resource, bts, err = dictionaryString(bts, dict)
	if err != nil {
		log.Debug(err.Error())

		return bts, err
	}
	// TraceID (3)
	out.TraceID, bts, err = msgpack.ParseUint64Bytes(bts)
	if err != nil {
		log.Debug(err.Error())

		return bts, err
	}
	// SpanID (4)
	out.SpanID, bts, err = msgpack.ParseUint64Bytes(bts)
	if err != nil {
		log.Debug(err.Error())

		return bts, err
	}
	// ParentID (5)
	out.ParentID, bts, err = msgpack.ParseUint64Bytes(bts)
	if err != nil {
		log.Debug(err.Error())

		return bts, err
	}
	// Start (6)
	out.Start, bts, err = msgpack.ParseInt64Bytes(bts)
	if err != nil {
		log.Debug(err.Error())

		return bts, err
	}
	// Duration (7)
	out.Duration, bts, err = msgpack.ParseInt64Bytes(bts)
	if err != nil {
		log.Debug(err.Error())

		return bts, err
	}
	// Error (8)
	out.Error, bts, err = msgpack.ParseInt32Bytes(bts)
	if err != nil {
		log.Debug(err.Error())

		return bts, err
	}
	// Meta (9)
	sz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		log.Debug(err.Error())

		return bts, err
	}
	if out.Meta == nil && sz > 0 {
		out.Meta = make(map[string]string, sz)
	} else if len(out.Meta) > 0 {
		for key := range out.Meta {
			delete(out.Meta, key)
		}
	}
	for sz > 0 {
		sz--
		var key, val string
		key, bts, err = dictionaryString(bts, dict)
		if err != nil {
			log.Debug(err.Error())

			return bts, err
		}
		val, bts, err = dictionaryString(bts, dict)
		if err != nil {
			log.Debug(err.Error())

			return bts, err
		}
		out.Meta[key] = val
	}
	// Metrics (10)
	sz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		log.Debug(err.Error())

		return bts, err
	}
	if out.Metrics == nil && sz > 0 {
		out.Metrics = make(map[string]float64, sz)
	} else if len(out.Metrics) > 0 {
		for key := range out.Metrics {
			delete(out.Metrics, key)
		}
	}
	for sz > 0 {
		sz--
		var (
			key string
			val float64
		)
		key, bts, err = dictionaryString(bts, dict)
		if err != nil {
			return bts, err
		}
		val, bts, err = msgpack.ParseFloat64Bytes(bts)
		if err != nil {
			log.Debug(err.Error())

			return bts, err
		}
		out.Metrics[key] = val
	}
	// Type (11)
	out.Type, bts, err = dictionaryString(bts, dict)
	if err != nil {
		log.Debug(err.Error())

		return bts, err
	}

	return bts, nil
}

// dictionaryString reads an int from decoder dc and returns the string
// at that index from dict.
func dictionaryString(bts []byte, dict []string) (string, []byte, error) {
	var (
		ui  uint32
		err error
	)
	ui, bts, err = msgp.ReadUint32Bytes(bts)
	if err != nil {
		return "", bts, err
	}
	idx := int(ui)
	if idx >= len(dict) {
		return "", bts, fmt.Errorf("dictionary index %d out of range", idx)
	}

	return dict[idx], bts, nil
}
