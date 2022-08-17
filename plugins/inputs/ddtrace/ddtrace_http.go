// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ddtrace

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"time"

	cache "gitlab.jiagouyun.com/cloudcare-tools/cliutils/diskcache"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
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

type parameters struct {
	URLPath string
	Media   string
	Buf     []byte
	body    *bytes.Buffer
}

func handleDDTraceWithVersion(v string) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		log.Debugf("### received tracing data from path: %s", req.URL.Path)

		var (
			readbodycost = time.Now()
			enterWPoolOK = false
		)
		pbuf := bufpool.GetBuffer()
		defer func() {
			if !enterWPoolOK {
				bufpool.PutBuffer(pbuf)
			}
		}()

		_, err := io.Copy(pbuf, req.Body)
		if err != nil {
			log.Error(err.Error())
			resp.WriteHeader(http.StatusBadRequest)

			return
		}

		param := &parameters{
			URLPath: req.URL.Path,
			Media:   req.Header.Get("Content-Type"),
			body:    pbuf,
		}

		log.Debugf("### path: %s, Content-Type: %s, body-size: %dkb, read-body-cost: %dms",
			param.URLPath, param.Media, pbuf.Len()>>10, time.Since(readbodycost)/time.Millisecond)

		if wpool == nil {
			if storage == nil {
				if err = parseDDTraces(param); err != nil {
					log.Error(err.Error())
					resp.WriteHeader(http.StatusBadRequest)

					return
				}
			} else {
				if err = sendToStorage(param); err != nil {
					log.Error(err.Error())
					resp.WriteHeader(http.StatusBadRequest)

					return
				}
			}
		} else {
			job, err := workerpool.NewJob(workerpool.WithInput(param),
				workerpool.WithProcess(parseDDTracesAdapter),
				workerpool.WithProcessCallback(func(input, output interface{}, cost time.Duration) {
					if param, ok := input.(*parameters); ok {
						bufpool.PutBuffer(param.body)
					}
					log.Debugf("### job status: input: %v, output: %v, cost: %dms", input, output, cost/time.Millisecond)
				}),
			)
			if err != nil {
				log.Error(err.Error())
				resp.WriteHeader(http.StatusBadRequest)

				return
			}

			if err = wpool.MoreJob(job); err != nil {
				log.Error(err)
				resp.WriteHeader(http.StatusTooManyRequests)

				return
			}
			enterWPoolOK = true
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
	log.Errorf("%s unsupported yet", req.URL.Path)
	resp.WriteHeader(http.StatusNotFound)
}

// TODO:.
func handleDDStats(resp http.ResponseWriter, req *http.Request) {
	log.Errorf("%s unsupported yet", req.URL.Path)
	resp.WriteHeader(http.StatusNotFound)
}

func sendToStorage(param *parameters) error {
	var err error
	if param.Buf, err = io.ReadAll(param.body); err != nil {
		return err
	}

	buffer := bytes.NewBuffer(nil)
	encoder := gob.NewEncoder(buffer)
	if err = encoder.Encode(param); err != nil {
		return err
	}

	return storage.Put(buffer.Bytes())
}

func parseDDTracesAdapter(input interface{}) (output interface{}) {
	param, ok := input.(*parameters)
	if !ok {
		return errors.New("type assertion failed")
	}

	if storage != nil {
		return sendToStorage(param)
	} else {
		return parseDDTraces(param)
	}
}

func parseDDTraces(param *parameters) error {
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

func decodeDDTraces(param *parameters) (DDTraces, error) {
	var (
		traces DDTraces
		err    error
	)
	switch param.URLPath {
	case v1:
		var spans DDTrace
		if err := json.NewDecoder(param.body).Decode(&spans); err != nil {
			return nil, err
		}
		traces = mergeSpans(spans)
	case v5:
		if err = traces.UnmarshalMsgDictionary(param.body.Bytes()); err == nil {
			traces = mergeTraces(traces)
		}
	default:
		if err = decodeRequest(param, &traces); err == nil {
			traces = mergeTraces(traces)
		}
	}

	return traces, err
}

func decodeRequest(param *parameters, out *DDTraces) error {
	mediaType, _, err := mime.ParseMediaType(param.Media)
	if err != nil {
		log.Debug(err.Error())
	}
	switch mediaType {
	case "application/msgpack":
		_, err = out.UnmarshalMsg(param.body.Bytes())
	case "application/json", "text/json", "":
		return json.NewDecoder(param.body).Decode(out)
	default:
		// do our best
		if err1 := json.NewDecoder(param.body).Decode(out); err1 != nil {
			if _, err2 := out.UnmarshalMsg(param.body.Bytes()); err2 != nil {
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
			TraceID:        fmt.Sprintf("%d", span.TraceID),
			ParentID:       fmt.Sprintf("%d", span.ParentID),
			SpanID:         fmt.Sprintf("%d", span.SpanID),
			Service:        span.Service,
			Resource:       span.Resource,
			Operation:      span.Name,
			Source:         inputName,
			SpanType:       itrace.FindSpanTypeInMultiServersIntSpanID(int64(span.SpanID), int64(span.ParentID), span.Service, spanIDs, parentIDs),
			SourceType:     itrace.GetSpanSourceType(span.Type),
			Tags:           itrace.MergeInToCustomerTags(customerKeys, tags, span.Meta),
			ContainerHost:  span.Meta[itrace.CONTAINER_HOST],
			PID:            fmt.Sprintf("%d", int64(span.Metrics["system.pid"])),
			HTTPMethod:     span.Meta["http.method"],
			HTTPStatusCode: span.Meta["http.status_code"],
			Start:          span.Start,
			Duration:       span.Duration,
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

		if id, ok := span.Meta["runtime-id"]; ok {
			dkspan.Tags["runtime_id"] = id
		}

		dkspan.Status = itrace.STATUS_OK
		if span.Error != 0 {
			dkspan.Status = itrace.STATUS_ERR
		}

		if priority, ok := span.Metrics[keyPriority]; ok {
			dkspan.Metrics = make(map[string]interface{})
			dkspan.Metrics[itrace.FIELD_PRIORITY] = int(priority)
		}
		if rate, ok := span.Metrics[keySamplingRate]; ok {
			dkspan.Metrics[itrace.FIELD_SAMPLE_RATE] = rate
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

func consumeStorageWorker() {
	for {
		start := time.Now()
		if err := storage.Get(func(buf []byte) error {
			param := &parameters{}
			err := gob.NewDecoder(bytes.NewBuffer(buf)).Decode(param)
			if err == nil {
				param.body = bytes.NewBuffer(param.Buf)
				err := parseDDTraces(param)

				log.Debugf("### process status: buffer-size: %dkb, cost: %dms, err: %v", len(param.Buf)>>10, time.Since(start)/time.Millisecond, err)
			}

			return err
		}); err != nil {
			if errors.Is(err, cache.ErrEOF) {
				log.Debug(err.Error())
				time.Sleep(time.Second)
			} else {
				log.Error(err.Error())
			}
		}
	}
}
