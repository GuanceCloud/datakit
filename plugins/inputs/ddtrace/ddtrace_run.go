package ddtrace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/tinylib/msgp/msgp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
)

const (
	// KeySamplingPriority is the key of the sampling priority value in the metrics map of the root span.
	keyPriority = "_sampling_priority_v1"
	// keySamplingRateGlobal is a metric key holding the global sampling rate.
	keySamplingRateGlobal = "_sample_rate"
)

func handleDDTraces(resp http.ResponseWriter, req *http.Request) {
	log.Debugf("%s: received tracing data from path: %s", inputName, req.URL.Path)

	traces, err := decodeDDTraces(req.URL.Path, req)
	if err != nil {
		log.Errorf(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
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

func decodeDDTraces(endpoint string, req *http.Request) (DDTraces, error) {
	var (
		traces DDTraces
		err    error
	)
	switch endpoint {
	case v1:
		var spans DDTrace
		if err := json.NewDecoder(req.Body).Decode(&spans); err != nil {
			return nil, err
		}
		traces = tracesFromSpans(spans)
	case v5:
		buf := bufpool.GetBuffer()
		defer bufpool.PutBuffer(buf)

		if _, err := io.Copy(buf, req.Body); err != nil {
			return nil, err
		}
		err = unmarshalTraceMsgDictionary(buf.Bytes(), &traces)
	default:
		err = decodeRequest(req, &traces)
	}

	return traces, err
}

func decodeRequest(req *http.Request, out *DDTraces) error {
	mediaType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		log.Debug(err.Error())
	}
	switch mediaType {
	case "application/msgpack":
		buf := bufpool.GetBuffer()
		defer bufpool.PutBuffer(buf)

		if _, err = io.Copy(buf, req.Body); err != nil {
			return err
		}
		_, err = out.UnmarshalMsg(buf.Bytes())
	case "application/json", "text/json", "":
		return json.NewDecoder(req.Body).Decode(out)
	default:
		// do our best
		if err1 := json.NewDecoder(req.Body).Decode(out); err1 != nil {
			buf := bufpool.GetBuffer()
			defer bufpool.PutBuffer(buf)

			_, err2 := io.Copy(buf, req.Body)
			if err2 != nil {
				err = fmt.Errorf("could not decode JSON (%q), nor Msgpack (%q)", err1, err2)
			}
			_, err2 = out.UnmarshalMsg(buf.Bytes())
			if err2 != nil {
				err = fmt.Errorf("could not decode JSON (%q), nor Msgpack (%q)", err1, err2)
			}
		}
	}

	return err
}

func tracesFromSpans(trace DDTrace) DDTraces {
	traces := DDTraces{}
	byID := make(map[uint64]DDTrace)
	for _, span := range trace {
		byID[span.TraceID] = append(byID[span.TraceID], span)
	}
	for _, trace := range byID {
		traces = append(traces, trace)
	}

	return traces
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
		str, bts, err = ParseStringBytes(bts)
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
	out.TraceID, bts, err = ParseUint64Bytes(bts)
	if err != nil {
		log.Debug(err.Error())

		return bts, err
	}
	// SpanID (4)
	out.SpanID, bts, err = ParseUint64Bytes(bts)
	if err != nil {
		log.Debug(err.Error())

		return bts, err
	}
	// ParentID (5)
	out.ParentID, bts, err = ParseUint64Bytes(bts)
	if err != nil {
		log.Debug(err.Error())

		return bts, err
	}
	// Start (6)
	out.Start, bts, err = ParseInt64Bytes(bts)
	if err != nil {
		log.Debug(err.Error())

		return bts, err
	}
	// Duration (7)
	out.Duration, bts, err = ParseInt64Bytes(bts)
	if err != nil {
		log.Debug(err.Error())

		return bts, err
	}
	// Error (8)
	out.Error, bts, err = ParseInt32Bytes(bts)
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
		val, bts, err = ParseFloat64Bytes(bts)
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
