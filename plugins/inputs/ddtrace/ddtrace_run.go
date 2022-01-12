package ddtrace

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"time"

	"github.com/tinylib/msgp/msgp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/msgpack"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

var ddtraceSpanType = map[string]string{
	"cache":         itrace.SPAN_SERVICE_CACHE,
	"cassandra":     itrace.SPAN_SERVICE_DB,
	"elasticsearch": itrace.SPAN_SERVICE_DB,
	"grpc":          itrace.SPAN_SERVICE_CUSTOM,
	"http":          itrace.SPAN_SERVICE_WEB,
	"mongodb":       itrace.SPAN_SERVICE_DB,
	"redis":         itrace.SPAN_SERVICE_CACHE,
	"sql":           itrace.SPAN_SERVICE_DB,
	"template":      itrace.SPAN_SERVICE_CUSTOM,
	"test":          itrace.SPAN_SERVICE_CUSTOM,
	"web":           itrace.SPAN_SERVICE_WEB,
	"worker":        itrace.SPAN_SERVICE_CUSTOM,
	"memcached":     itrace.SPAN_SERVICE_CACHE,
	"leveldb":       itrace.SPAN_SERVICE_DB,
	"dns":           itrace.SPAN_SERVICE_CUSTOM,
	"queue":         itrace.SPAN_SERVICE_CUSTOM,
	"consul":        itrace.SPAN_SERVICE_APP,
	"rpc":           itrace.SPAN_SERVICE_CUSTOM,
	"soap":          itrace.SPAN_SERVICE_CUSTOM,
	"db":            itrace.SPAN_SERVICE_DB,
	"hibernate":     itrace.SPAN_SERVICE_CUSTOM,
	"aerospike":     itrace.SPAN_SERVICE_DB,
	"datanucleus":   itrace.SPAN_SERVICE_CUSTOM,
	"graphql":       itrace.SPAN_SERVICE_CUSTOM,
	"custom":        itrace.SPAN_SERVICE_CUSTOM,
	"benchmark":     itrace.SPAN_SERVICE_CUSTOM,
	"build":         itrace.SPAN_SERVICE_CUSTOM,
	"":              itrace.SPAN_SERVICE_CUSTOM,
}

//nolint:lll
type Span struct {
	Service  string             `codec:"service" protobuf:"bytes,1,opt,name=service,proto3" json:"service" msg:"service"`                                                                                     // client code defined service name of span
	Name     string             `codec:"name" protobuf:"bytes,2,opt,name=name,proto3" json:"name" msg:"name"`                                                                                                 // client code defined operation name of span
	Resource string             `codec:"resource" protobuf:"bytes,3,opt,name=resource,proto3" json:"resource" msg:"resource"`                                                                                 // client code defined resource name of span
	TraceID  uint64             `codec:"trace_id" protobuf:"varint,4,opt,name=traceID,proto3" json:"trace_id" msg:"trace_id"`                                                                                 // id of root span
	SpanID   uint64             `codec:"span_id" protobuf:"varint,5,opt,name=spanID,proto3" json:"span_id" msg:"span_id"`                                                                                     // id of this span
	ParentID uint64             `codec:"parent_id" protobuf:"varint,6,opt,name=parentID,proto3" json:"parent_id" msg:"parent_id"`                                                                             // id of the span's direct parent
	Start    int64              `codec:"start" protobuf:"varint,7,opt,name=start,proto3" json:"start" msg:"start"`                                                                                            // span start time expressed in nanoseconds since epoch
	Duration int64              `codec:"duration" protobuf:"varint,8,opt,name=duration,proto3" json:"duration" msg:"duration"`                                                                                // duration of the span expressed in nanoseconds
	Error    int32              `codec:"error" protobuf:"varint,9,opt,name=error,proto3" json:"error" msg:"error"`                                                                                            // error status of the span; 0 means no errors
	Meta     map[string]string  `codec:"meta" protobuf:"bytes,10,rep,name=meta" json:"meta" msg:"meta" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`               // arbitrary map of meta data
	Metrics  map[string]float64 `codec:"metrics" protobuf:"bytes,11,rep,name=metrics" json:"metrics" msg:"metrics" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"fixed64,2,opt,name=value,proto3"` // arbitrary map of numeric metrics
	Type     string             `codec:"type" protobuf:"bytes,12,opt,name=type,proto3" json:"type" msg:"type"`                                                                                                // protocol associated with the span
}

type Trace []*Span

type Traces []Trace

// TODO:.
func handleInfo(resp http.ResponseWriter, req *http.Request) { //nolint: unused,deadcode
	log.Errorf("%s not support now", req.URL.Path)
	resp.WriteHeader(http.StatusNotFound)
}

func handleTraces(pattern string) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		since := time.Now()
		traces, err := decodeRequest(pattern, req)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Warn(err.Error())
				resp.WriteHeader(http.StatusOK)
			} else {
				log.Error(err.Error())
				resp.WriteHeader(http.StatusBadRequest)
			}

			return
		}
		if len(traces) == 0 {
			log.Debug("empty traces")
			resp.WriteHeader(http.StatusOK)

			return
		}

		log.Debugf("show up all traces: %v", traces)

		pts, err := tracesToPoints(req, traces, filters...)
		if err != nil {
			log.Error(err.Error())
			resp.WriteHeader(http.StatusBadRequest)

			return
		}
		if len(pts) != 0 {
			if err = dkio.Feed(inputName, datakit.Tracing, pts,
				&dkio.Option{
					CollectCost: time.Since(since),
					HighFreq:    true,
				}); err != nil {
				log.Errorf("Feed: %s", err.Error())
			}
		} else {
			log.Debugf("empty points")
		}

		resp.WriteHeader(http.StatusOK)
	}
}

// TODO:.
func handleStats(resp http.ResponseWriter, req *http.Request) {
	log.Errorf("%s not support now", req.URL.Path)
	resp.WriteHeader(http.StatusNotFound)
}

func extractCustomerTags(customerKeys []string, meta map[string]string) map[string]string {
	customerTags := map[string]string{}
	for _, key := range customerKeys {
		if value, ok := meta[key]; ok {
			customerTags[key] = value
		}
	}

	return customerTags
}

func decodeRequest(pattern string, req *http.Request) (Traces, error) {
	mediaType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		log.Debugf("detect media-type failed fallback to application/json")
		mediaType = "application/json"
	}

	buf, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	traces := Traces{}
	if pattern == v5 {
		err = unmarshalTraceDictionary(buf, &traces)
	} else {
		switch mediaType {
		case "application/msgpack":
			err = msgpack.Unmarshal(bytes.NewBuffer(buf), &traces)
		case "application/json", "text/json", "":
			err = json.NewDecoder(bytes.NewBuffer(buf)).Decode(&traces)
		default:
			err = errors.New("unrecognized Content-Type to decode")
		}
	}

	return traces, err
}

// tracesToPoints work as a adapter to convert traces to points.
// parameter traces is raw traces, filters contains all the functional filters
// like resource filter, sample.
//nolint: cyclop
func tracesToPoints(req *http.Request, traces Traces, filters ...traceFilter) ([]*dkio.Point, error) {
	var pts []*dkio.Point
	for _, trace := range traces {
		if trace == nil || len(trace) == 0 {
			log.Warnf("got empty trace, request headers: %v", req.Header)
			continue
		}
		// run all filters
		if runFiltersWithBreak(trace, filters...) == nil {
			continue
		}

		spanIDs, parentIDs := getSpanIDsAndParentIDs(trace)
		for _, span := range trace {
			if span == nil {
				log.Warnf("got nil span, request headers: %v", req.Header)
				continue
			}

			var (
				spanInfo = &dkio.SpanInfo{
					Toolkit:  inputName,
					Service:  span.Service,
					Resource: span.Resource,
					Duration: time.Duration(span.Duration),
				}
				spanType = itrace.FindSpanType(int64(span.SpanID), int64(span.ParentID), spanIDs, parentIDs)
				tags     = make(map[string]string)
				field    = make(map[string]interface{})
				tm       = &itrace.TraceMeasurement{Name: "ddtrace"}
			)

			tags[itrace.TAG_SPAN_TYPE] = spanType
			if spanType == itrace.SPAN_TYPE_ENTRY {
				spanInfo.IsEntry = true
				spanInfo.IsErr = span.Error != 0
			}

			tags[itrace.TAG_SERVICE] = span.Service
			tags[itrace.TAG_OPERATION] = span.Name
			tags[itrace.TAG_TYPE] = ddtraceSpanType[span.Type]
			if span.Error == 0 {
				tags[itrace.TAG_SPAN_STATUS] = itrace.STATUS_OK
			} else {
				tags[itrace.TAG_SPAN_STATUS] = itrace.STATUS_ERR
			}
			tags[itrace.TAG_PROJECT] = span.Meta[itrace.PROJECT]
			if tags[itrace.TAG_PROJECT] == "" {
				tags[itrace.TAG_PROJECT] = ddTags[itrace.PROJECT]
			}
			spanInfo.Project = tags[itrace.TAG_PROJECT]

			tags[itrace.TAG_ENV] = span.Meta[itrace.ENV]
			if tags[itrace.TAG_ENV] == "" {
				tags[itrace.TAG_ENV] = ddTags[itrace.ENV]
			}

			tags[itrace.TAG_VERSION] = span.Meta[itrace.VERSION]
			if tags[itrace.TAG_VERSION] == "" {
				tags[itrace.TAG_VERSION] = ddTags[itrace.VERSION]
			}
			spanInfo.Version = tags[itrace.TAG_VERSION]

			// send span info
			dkio.SendSpanInfo(spanInfo)

			tags[itrace.TAG_CONTAINER_HOST] = span.Meta[itrace.CONTAINER_HOST]
			tags[itrace.TAG_HTTP_METHOD] = span.Meta["http.method"]
			tags[itrace.TAG_HTTP_CODE] = span.Meta["http.status_code"]

			customerTags := extractCustomerTags(customerKeys, span.Meta)
			for k, v := range customerTags {
				tags[k] = v
			}
			for k, v := range ddTags {
				tags[k] = v
			}

			field[itrace.FIELD_RESOURCE] = span.Resource
			field[itrace.FIELD_PARENTID] = fmt.Sprintf("%d", span.ParentID)
			field[itrace.FIELD_TRACEID] = fmt.Sprintf("%d", span.TraceID)
			field[itrace.FIELD_SPANID] = fmt.Sprintf("%d", span.SpanID)
			field[itrace.FIELD_DURATION] = span.Duration / int64(time.Microsecond)
			field[itrace.FIELD_START] = span.Start / int64(time.Microsecond)
			if v, ok := span.Metrics["system.pid"]; ok {
				field[itrace.FIELD_PID] = fmt.Sprintf("%v", v)
			}
			if buf, err := json.Marshal(span); err != nil {
				return nil, err
			} else {
				field[itrace.FIELD_MSG] = string(buf)
			}

			tm.Tags = tags
			tm.Fields = field
			tm.TS = time.Unix(span.Start/int64(time.Second), span.Start%int64(time.Second))
			if pt, err := tm.LineProto(); err != nil {
				return nil, err
			} else {
				pts = append(pts, pt)
			}
		}
	}

	return pts, nil
}

func getSpanIDsAndParentIDs(trace Trace) (map[int64]bool, map[int64]bool) {
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

// unmarshalTraceDictionary decodes a trace using the specification from the v0.5 endpoint.
// For details, see the documentation for endpoint v0.5 in pkg/trace/api/version.go
//nolint:cyclop
func unmarshalTraceDictionary(bts []byte, out *Traces) error {
	if out == nil {
		return errors.New("nil pointer")
	}

	var err error
	if _, bts, err = msgp.ReadArrayHeaderBytes(bts); err != nil {
		return err
	}
	// read dictionary
	var sz uint32
	if sz, bts, err = msgp.ReadArrayHeaderBytes(bts); err != nil {
		return err
	}
	dict := make([]string, sz)
	for i := range dict {
		var str string
		str, bts, err = msgpack.ParseStringBytes(bts)
		if err != nil {
			return err
		}
		dict[i] = str
	}
	// read traces
	sz, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		return err
	}
	if cap(*out) >= int(sz) {
		*out = (*out)[:sz]
	} else {
		*out = make(Traces, sz)
	}
	for i := range *out {
		sz, bts, err = msgp.ReadArrayHeaderBytes(bts)
		if err != nil {
			return err
		}
		if cap((*out)[i]) >= int(sz) {
			(*out)[i] = (*out)[i][:sz]
		} else {
			(*out)[i] = make(Trace, sz)
		}
		for j := range (*out)[i] {
			if (*out)[i][j] == nil {
				(*out)[i][j] = new(Span)
			}
			if bts, err = unmarshalSpanDictionary(bts, dict, (*out)[i][j]); err != nil {
				return err
			}
		}
	}

	return nil
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

// spanPropertyCount specifies the number of top-level properties that a span
// has.
const spanPropertyCount = 12

// unmarshalSpanDictionary decodes a span from the given decoder dc, looking up strings
// in the given dictionary dict. For details, see the documentation for endpoint v0.5
// in pkg/trace/api/version.go
//nolint:cyclop
func unmarshalSpanDictionary(bts []byte, dict []string, out *Span) ([]byte, error) {
	if out == nil {
		return nil, errors.New("nil pointer")
	}

	var (
		sz  uint32
		err error
	)
	sz, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		return bts, err
	}
	if sz != spanPropertyCount {
		return bts, errors.New("encoded span needs exactly 12 elements in array")
	}
	// Service (0)
	out.Service, bts, err = dictionaryString(bts, dict)
	if err != nil {
		return bts, err
	}
	// Name (1)
	out.Name, bts, err = dictionaryString(bts, dict)
	if err != nil {
		return bts, err
	}
	// Resource (2)
	out.Resource, bts, err = dictionaryString(bts, dict)
	if err != nil {
		return bts, err
	}
	// TraceID (3)
	out.TraceID, bts, err = msgpack.ParseUint64Bytes(bts)
	if err != nil {
		return bts, err
	}
	// SpanID (4)
	out.SpanID, bts, err = msgpack.ParseUint64Bytes(bts)
	if err != nil {
		return bts, err
	}
	// ParentID (5)
	out.ParentID, bts, err = msgpack.ParseUint64Bytes(bts)
	if err != nil {
		return bts, err
	}
	// Start (6)
	out.Start, bts, err = msgpack.ParseInt64Bytes(bts)
	if err != nil {
		return bts, err
	}
	// Duration (7)
	out.Duration, bts, err = msgpack.ParseInt64Bytes(bts)
	if err != nil {
		return bts, err
	}
	// Error (8)
	out.Error, bts, err = msgpack.ParseInt32Bytes(bts)
	if err != nil {
		return bts, err
	}
	// Meta (9)
	sz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
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
			return bts, err
		}
		val, bts, err = dictionaryString(bts, dict)
		if err != nil {
			return bts, err
		}
		out.Meta[key] = val
	}
	// Metrics (10)
	sz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
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
			return bts, err
		}
		out.Metrics[key] = val
	}
	// Type (11)
	out.Type, bts, err = dictionaryString(bts, dict)
	if err != nil {
		return bts, err
	}

	return bts, nil
}
