// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ddtrace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"

	"github.com/tinylib/msgp/msgp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

const (
	// headerRatesPayloadVersion contains the version of sampling rates.
	// If both agent and client have the same version, the agent won't return rates in API response.
	headerRatesPayloadVersion = "Datadog-Rates-Payload-Version"
)

const (
	// KeySamplingPriority is the key of the sampling priority value in the metrics map of the root span.
	keyPriority = "_sampling_priority_v1"
	// keySamplingRateGlobal is a metric key holding the global sampling rate.
	keySamplingRate = "_sample_rate"
)

func httpStatusRespFunc(resp http.ResponseWriter, req *http.Request, err error) {
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	switch req.URL.Path {
	case v1, v2, v3:
		io.WriteString(resp, "OK\n") // nolint: errcheck,gosec
	default:
		resp.Header().Set("Content-Type", "application/json")
		resp.Header().Set(headerRatesPayloadVersion, req.Header.Get(headerRatesPayloadVersion))
		resp.Write([]byte("{}")) // nolint: errcheck,gosec
	}
}

func handleDDTraces(resp http.ResponseWriter, req *http.Request) {
	log.Debugf("### received tracing data from path: %s", req.URL.Path)

	if req.Header.Get("Content-Length") == "0" || req.Header.Get("X-Datadog-Trace-Count") == "0" {
		log.Debug("empty request body")
		httpStatusRespFunc(resp, req, nil)

		return
	}

	pbuf := bufpool.GetBuffer()
	defer bufpool.PutBuffer(pbuf)

	_, err := io.Copy(pbuf, req.Body)
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	param := &itrace.TraceParameters{
		URLPath: req.URL.Path,
		Media:   req.Header.Get("Content-Type"),
		Body:    pbuf,
	}
	if err = parseDDTraces(param); err != nil {
		if errors.Is(err, msgp.ErrShortBytes) {
			log.Warn(err.Error())
		} else {
			log.Errorf("### parse ddtrace failed: %s", err.Error())
		}
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	httpStatusRespFunc(resp, req, nil)
}

// TODO:.
func handleDDInfo(resp http.ResponseWriter, req *http.Request) { // nolint: unused,deadcode
	log.Errorf("### %s unsupported yet", req.URL.Path)
	resp.WriteHeader(http.StatusNotFound)
}

// TODO:.
func handleDDStats(resp http.ResponseWriter, req *http.Request) {
	log.Errorf("### %s unsupported yet", req.URL.Path)
	resp.WriteHeader(http.StatusNotFound)
}

func parseDDTraces(param *itrace.TraceParameters) error {
	traces, err := decodeDDTraces(param)
	if err != nil {
		return err
	}
	if len(traces) == 0 {
		log.Debug("### get empty traces after decoding")

		return nil
	}

	var dktraces itrace.DatakitTraces
	for _, trace := range traces {
		if len(trace) == 0 {
			log.Debug("### empty trace in traces")
			continue
		}
		if dktrace := ddtraceToDkTrace(trace); len(dktrace) != 0 {
			dktraces = append(dktraces, dktrace)
		}
	}

	if len(dktraces) != 0 && afterGatherRun != nil {
		afterGatherRun.Run(inputName, dktraces, false)
	}

	return nil
}

func decodeDDTraces(param *itrace.TraceParameters) (DDTraces, error) {
	var (
		traces DDTraces
		err    error
	)
	switch param.URLPath {
	case v1:
		var spans DDTrace
		if err := json.NewDecoder(param.Body).Decode(&spans); err != nil {
			return nil, err
		}
		traces = mergeSpans(spans)
	case v5:
		if err = traces.UnmarshalMsgDictionary(param.Body.Bytes()); err == nil {
			traces = mergeTraces(traces)
		}
	default:
		if err = decodeRequest(param, &traces); err == nil {
			traces = mergeTraces(traces)
		}
	}

	return traces, err
}

func decodeRequest(param *itrace.TraceParameters, out *DDTraces) error {
	mediaType, _, err := mime.ParseMediaType(param.Media)
	if err != nil {
		log.Debug(err.Error())
	}
	switch mediaType {
	case "application/msgpack":
		_, err = out.UnmarshalMsg(param.Body.Bytes())
	case "application/json", "text/json", "":
		return json.NewDecoder(param.Body).Decode(out)
	default:
		// do our best
		if err1 := json.NewDecoder(param.Body).Decode(out); err1 != nil {
			if _, err2 := out.UnmarshalMsg(param.Body.Bytes()); err2 != nil {
				err = fmt.Errorf("### could not decode JSON (err:%s), nor Msgpack (err:%s)", err1.Error(), err2.Error()) // nolint:errorlint
			}
		}
	}

	return err
}

func mergeSpans(trace DDTrace) DDTraces {
	var (
		traces = DDTraces{}
		byID   = make(map[uint64]DDTrace)
	)
	for _, span := range trace {
		byID[span.TraceID] = append(byID[span.TraceID], span)
	}
	for _, trace := range byID {
		traces = append(traces, trace)
	}

	return traces
}

func mergeTraces(traces DDTraces) DDTraces {
	var (
		merged DDTraces
		byID   = make(map[uint64]DDTrace)
	)
	for i := range traces {
		if len(traces[i]) != 0 {
			byID[traces[i][0].TraceID] = append(byID[traces[i][0].TraceID], traces[i]...)
		}
	}
	for _, trace := range byID {
		merged = append(merged, trace)
	}

	return merged
}

func pickupMeta(dkspan *itrace.DatakitSpan, ddspan *DDSpan, keys ...string) {
	if dkspan.Tags == nil {
		dkspan.Tags = make(map[string]string)
	}

	for i := range keys {
		if v, ok := ddspan.Meta[keys[i]]; ok {
			dkspan.Tags[keys[i]] = v
		}
	}

	if pid, ok := ddspan.Metrics["system.pid"]; ok {
		dkspan.Tags[itrace.TAG_PID] = strconv.FormatInt(int64(pid), 10)
	}
	if runtimeid, ok := ddspan.Meta["runtime-id"]; ok {
		dkspan.Tags["runtime_id"] = runtimeid
	}
	if origin, ok := ddspan.Meta["_dd.origin"]; ok {
		dkspan.Tags["_dd_origin"] = origin
	}

	if dkspan.SourceType == itrace.SPAN_SOURCE_WEB {
		if host, ok := ddspan.Meta["http.host"]; ok {
			dkspan.Tags[itrace.TAG_HTTP_HOST] = host
		}
		if url, ok := ddspan.Meta["http.url"]; ok {
			dkspan.Tags[itrace.TAG_HTTP_URL] = url
		}
		if route, ok := ddspan.Meta["http.route"]; ok {
			dkspan.Tags[itrace.TAG_HTTP_ROUTE] = route
		}
		if method, ok := ddspan.Meta["http.method"]; ok {
			dkspan.Tags[itrace.TAG_HTTP_METHOD] = method
		}
		if statusCode, ok := ddspan.Meta["http.status_code"]; ok {
			dkspan.Tags[itrace.TAG_HTTP_STATUS_CODE] = statusCode
		}
	}

	if dkspan.Status == itrace.STATUS_ERR || dkspan.Status == itrace.STATUS_CRITICAL {
		if errType, ok := ddspan.Meta["error.type"]; ok {
			dkspan.Metrics[itrace.TAG_ERR_TYPE] = errType
		}
		if errStack, ok := ddspan.Meta["error.stack"]; ok {
			dkspan.Metrics[itrace.TAG_ERR_STACK] = errStack
		}
		if errMsg, ok := ddspan.Meta["error.msg"]; ok {
			dkspan.Metrics[itrace.TAG_ERR_MESSAGE] = errMsg
		}
	}
}

func ddtraceToDkTrace(trace DDTrace) itrace.DatakitTrace {
	var (
		dktrace            itrace.DatakitTrace
		parentIDs, spanIDs = gatherSpansInfo(trace)
	)
	for _, span := range trace {
		if span == nil {
			continue
		}

		dkspan := &itrace.DatakitSpan{
			TraceID:    strconv.FormatUint(span.TraceID, 10),
			ParentID:   strconv.FormatUint(span.ParentID, 10),
			SpanID:     strconv.FormatUint(span.SpanID, 10),
			Service:    span.Service,
			Resource:   span.Resource,
			Operation:  span.Name,
			Source:     inputName,
			SpanType:   itrace.FindSpanTypeInMultiServersIntSpanID(span.SpanID, span.ParentID, span.Service, spanIDs, parentIDs),
			SourceType: itrace.GetSpanSourceType(span.Type),
			Tags:       itrace.MergeInToCustomerTags(customerKeys, tags, span.Meta),
			Metrics:    make(map[string]interface{}),
			Start:      span.Start,
			Duration:   span.Duration,
		}

		dkspan.Status = itrace.STATUS_OK
		if span.Error != 0 {
			dkspan.Status = itrace.STATUS_ERR
		}

		if priority, ok := span.Metrics[keyPriority]; ok {
			dkspan.Metrics[itrace.FIELD_PRIORITY] = int(priority)
		}
		if rate, ok := span.Metrics[keySamplingRate]; ok {
			dkspan.Metrics[itrace.FIELD_SAMPLE_RATE] = rate
		}

		pickupMeta(dkspan, span, itrace.PROJECT, itrace.VERSION, itrace.ENV, itrace.CONTAINER_HOST)

		if buf, err := json.Marshal(span); err != nil {
			log.Warn(err.Error())
		} else {
			dkspan.Content = string(buf)
		}

		dktrace = append(dktrace, dkspan)
	}

	return dktrace
}

func gatherSpansInfo(trace DDTrace) (parentIDs map[uint64]bool, spanIDs map[uint64]string) {
	parentIDs = make(map[uint64]bool)
	spanIDs = make(map[uint64]string)
	for _, span := range trace {
		if span == nil {
			continue
		}
		spanIDs[span.SpanID] = span.Service
		if span.ParentID != 0 {
			parentIDs[span.ParentID] = true
		}
	}

	return
}
