package ddtrace

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"

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
	keySamplingRateGlobal = "_sample_rate"
)

func handleDDTraceWithVersion(v string) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
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
				log.Debug("empty ddtrace")
				continue
			}

			if dktrace := ddtraceToDkTrace(trace); len(dktrace) == 0 {
				log.Warn("empty datakit trace")
			} else {
				afterGatherRun.Run(inputName, dktrace, false)
			}
		}

		switch v {
		case v1, v2, v3:
			io.WriteString(resp, "OK\n") // nolint: errcheck,gosec
		default:
			resp.Header().Set("Content-Type", "application/json")
			resp.Header().Set(headerRatesPayloadVersion, req.Header.Get(headerRatesPayloadVersion))
			resp.Write([]byte("{}")) // nolint: errcheck,gosec
		}
	}
}

// TODO:.
func handleDDInfo(resp http.ResponseWriter, req *http.Request) { // nolint: unused,deadcode
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
		err = traces.UnmarshalMsgDictionary(buf.Bytes())
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
				err = fmt.Errorf("could not decode JSON (err:%s), nor Msgpack (err:%s)", err1.Error(), err2.Error()) // nolint:errorlint
			}
			_, err2 = out.UnmarshalMsg(buf.Bytes())
			if err2 != nil {
				err = fmt.Errorf("could not decode JSON (err:%s), nor Msgpack (err:%s)", err1.Error(), err2.Error()) // nolint:errorlint
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
		parentIDs, spanIDs = gatherSpansInfo(trace)
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
			SpanType:           itrace.FindSpanTypeInMultiServersIntSpanID(int64(span.SpanID), int64(span.ParentID), span.Service, spanIDs, parentIDs),
			SourceType:         getDDTraceSourceType(span.Type),
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

		if dkspan.ParentID == "0" && sampler != nil {
			dkspan.Priority = sampler.Priority
			dkspan.SamplingRateGlobal = sampler.SamplingRateGlobal
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

func gatherSpansInfo(trace DDTrace) (parentIDs map[int64]bool, spanIDs map[int64]string) {
	parentIDs = make(map[int64]bool)
	spanIDs = make(map[int64]string)
	for _, span := range trace {
		if span == nil {
			continue
		}
		spanIDs[int64(span.SpanID)] = span.Service
		if span.ParentID != 0 {
			parentIDs[int64(span.ParentID)] = true
		}
	}

	return
}
