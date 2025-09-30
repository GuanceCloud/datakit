// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ddtrace

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	jsoniter "github.com/json-iterator/go"
	"github.com/tinylib/msgp/msgp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
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
)

var jsonIterator = jsoniter.ConfigFastest

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
		resp.WriteHeader(http.StatusOK)
		resp.Write([]byte(`{}`)) // nolint: errcheck,gosec
	}
}

func handleDDTraces(resp http.ResponseWriter, req *http.Request) {
	clStr := req.Header.Get("Content-Length")
	ntraceStr := req.Header.Get("X-Datadog-Trace-Count")
	if clStr == "0" || ntraceStr == "0" {
		log.Debug("empty request body")
		httpStatusRespFunc(resp, req, nil)

		return
	}
	remoteIP, _ := net.RemoteAddr(req)
	ntrace, err := strconv.ParseInt(ntraceStr, 10, 64)
	if err != nil {
		log.Debugf("invalid X-Datadog-Trace-Count: %q, ignored", ntraceStr)
	}

	cl, err := strconv.ParseInt(clStr, 10, 64)
	if err != nil {
		log.Warnf("invalid Content-Length: %q, ignored", clStr)
	} else if maxTraceBody > 0 && cl > maxTraceBody {
		if ntrace > 0 {
			droppedTraces.WithLabelValues(req.URL.Path).Add(float64(ntrace))
		}

		log.Warnf("dropped %d trace: too large request body(%q bytes > %d bytes)", ntrace, clStr, maxTraceBody)
		return
	}

	pbuf := bufpool.GetBuffer()
	defer bufpool.PutBuffer(pbuf)

	_, err = io.Copy(pbuf, req.Body)
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	param := &itrace.TraceParameters{
		URLPath:  req.URL.Path,
		Media:    itrace.GetContentType(req),
		Body:     pbuf,
		RemoteIP: remoteIP,
	}

	log.Debugf("param body len=%d", param.Body.Len())

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
func handleDDStats(resp http.ResponseWriter, req *http.Request) {
	log.Infof("### %s unsupported yet", req.URL.Path)
	resp.WriteHeader(http.StatusNotFound)
}

func handleDDInfo(resp http.ResponseWriter, req *http.Request) {
	log.Debugf("### %s unsupported yet", req.URL.Path)
	resp.WriteHeader(http.StatusNotFound)
}

func (ipt *Input) handleDDProxy(resp http.ResponseWriter, req *http.Request) {
	bts, err := io.ReadAll(req.Body)
	defer req.Body.Close() //nolint
	if err != nil {
		log.Warnf("read body err=%v", err)
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	if ipt.om != nil {
		ipt.om.parseTelemetryRequest(req.Header, bts)
	}

	resp.WriteHeader(http.StatusOK)
}

func parseDDTraces(param *itrace.TraceParameters) error {
	dktraces, err := decodeDDTraces(param)
	if err != nil {
		return err
	}

	if len(dktraces) != 0 && afterGatherRun != nil {
		log.Debugf("feed %d traces", len(dktraces))
		afterGatherRun.Run(inputName, dktraces)
	}

	return nil
}

func decodeDDTraces(param *itrace.TraceParameters) (itrace.DatakitTraces, error) {
	var (
		err      error
		dktraces itrace.DatakitTraces
	)
	traces := ddtracePool.Get().(DDTraces)
	defer func() {
		traces.reset()
		ddtracePool.Put(traces) //nolint
	}()

	switch param.URLPath {
	case v1:
		var spans DDTrace
		if err := jsonIterator.Unmarshal(param.Body.Bytes(), &spans); err != nil {
			return nil, err
		}
		traces = mergeSpans(spans)
	case v5:
		if err = traces.UnmarshalMsgDictionary(param.Body.Bytes()); err == nil {
			traces = mergeTraces(traces)
		}
	default:
		if err = decodeRequest(param, &traces); err != nil {
			// traces = mergeTraces(traces)
			return nil, err
		}
	}

	curSpans := 0
	maxBatch := 100

	log.Debugf("transform ddtrace to dkspan, noStreaming=%v", noStreaming)
	values := make([]string, 0, len(labels))
	if len(traces) != 0 {
		for _, trace := range traces {
			if len(trace) == 0 {
				log.Debug("### empty trace in traces")
				continue
			}

			// decode single ddtrace into dktrace
			dktrace := ddtraceToDkTrace(trace, values, param.RemoteIP)
			if nspan := len(dktrace); nspan > 0 {
				if nspan > maxBatch && !noStreaming { // flush large trace ASAP.
					log.Debugf("streaming feed %d spans", nspan)
					afterGatherRun.Run(inputName, itrace.DatakitTraces{dktrace})
				} else {
					dktraces = append(dktraces, dktrace)
					curSpans += nspan

					if curSpans > maxBatch && !noStreaming { // multiple traces got too many spans, flush ASAP.
						log.Debugf("streaming feed %d spans within %d traces", curSpans, len(dktraces))
						afterGatherRun.Run(inputName, dktraces)

						// clear and reset
						dktraces = dktraces[:0]
						curSpans = 0
					}
				}
			}
		}
	}

	log.Debugf("curSpans: %d", curSpans)
	return dktraces, err
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
		return jsonIterator.Unmarshal(param.Body.Bytes(), out)
	default:
		// do our best
		if err1 := jsonIterator.Unmarshal(param.Body.Bytes(), out); err1 != nil {
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

var (
	ignoreTraceIDFromTag = false
	traceOpts            = []point.Option{}
)

func ddtraceToDkTrace(trace DDTrace, values []string, remoteIP string) itrace.DatakitTrace {
	var (
		parentIDs, spanIDs = gatherSpansInfo(trace) // NOTE: we should gather before truncate
		dktrace            = make(itrace.DatakitTrace, 0, len(trace))
		strTraceID         = ""
	)

	traceSpans.WithLabelValues(inputName).Observe(float64(len(trace)))

	// truncate too large spans
	if traceMaxSpans > 0 && len(trace) > traceMaxSpans {
		// append info to last span's meta
		lastSpan := trace[traceMaxSpans-1]
		if lastSpan.Meta == nil {
			lastSpan.Meta = map[string]string{}
		}

		lastSpan.Meta["__datakit_span_truncated"] = fmt.Sprintf("large trace that got spans %d, max span limit is %d(%d spans truncated)",
			len(trace), traceMaxSpans, len(trace)-traceMaxSpans)

		log.Warnf("truncate %d spans from service %q", len(trace)-traceMaxSpans, trace[0].Service)
		truncatedTraceSpans.WithLabelValues(inputName).Add(float64(len(trace) - traceMaxSpans))
		trace = trace[:traceMaxSpans] // truncated too large spans
	}

	for _, span := range trace {
		values = values[:0]
		if span == nil {
			continue
		}

		if strTraceID == "" {
			strTraceID = strconv.FormatUint(span.TraceID, traceBase)

			if v, ok := span.Meta[TraceIDUpper]; trace128BitID && ok {
				strTraceID = v + Int64ToPaddedString(span.TraceID)
			}

			if ignoreTraceIDFromTag {
				strTraceID = Int64ToPaddedString(span.TraceID)
			}
		}

		var spanKV point.KVs
		spanKV = spanKV.AddTag(itrace.TagRemoteIP, remoteIP)
		priority, ok := span.Metrics[keyPriority]
		if ok {
			if priority == -1 || priority == -3 || priority == 0 {
				log.Debugf("drop this traceID=%s service=%s", span.TraceID, span.Service)
				return []*itrace.DkSpan{} // 此处应该返回空的数组。
			}

			if p, ok := itrace.DDPriorityMap[int(priority)]; ok {
				// 在采样的结果放到行协议中，如果 DK 有配置采样，则需要该值进行过滤。
				spanKV = spanKV.SetTag(itrace.SampleRateKey, p)
			}
		}

		resource := span.Resource
		if strings.Contains(span.Resource, "\n") {
			resource = strings.ReplaceAll(span.Resource, "\n", " ")
		}

		spanKV = spanKV.Add(itrace.FieldTraceID, strTraceID).
			Add(itrace.FieldParentID, itrace.FormatSpanIDByBase(span.ParentID, spanBase)).
			Add(itrace.FieldSpanid, itrace.FormatSpanIDByBase(span.SpanID, spanBase)).
			AddTag(itrace.TagService, span.Service).
			Add(itrace.FieldResource, resource).
			AddTag(itrace.TagOperation, span.Name).
			AddTag(itrace.TagSource, inputName).
			AddTag(itrace.TagSpanType,
				itrace.FindSpanTypeInMultiServersIntSpanID(span.SpanID,
					span.ParentID,
					span.Service,
					spanIDs,
					parentIDs)).
			AddTag(itrace.TagSourceType, itrace.GetSpanSourceType(span.Type)).
			Add(itrace.FieldStart, span.Start/int64(time.Microsecond)).
			Add(itrace.FieldDuration, span.Duration/int64(time.Microsecond))

		// runtime_id 作为链路和 profiling 关联的字段，由于历史问题，需要增加一个冗余字段。
		runTimeIDKey := "runtime-id"
		if v, ok := span.Meta[runTimeIDKey]; ok {
			spanKV = spanKV.AddTag("runtime_id", v).AddTag(runTimeIDKey, v)
			delete(span.Meta, runTimeIDKey)
		}

		for k, v := range inputTags {
			spanKV = spanKV.AddTag(k, v)
		}

		for k, v := range span.Meta {
			ddTagsLock.RLock()
			if replace, ok := ddTags[k]; ok {
				if len(v) > 1024 {
					spanKV = spanKV.Set(replace, v)
				} else {
					spanKV = spanKV.SetTag(replace, v)
				}
				// 从 message 中删除 key.
				delete(span.Meta, k)
			}
			if k == "db.type" { // db 类型。
				spanKV = spanKV.AddTag("db_host", span.Meta["peer.hostname"])
			}
			ddTagsLock.RUnlock()
		}
		if code := spanKV.GetTag(itrace.TagHttpStatusCode); code != "" {
			spanKV = spanKV.AddTag(itrace.TagHttpStatusClass, itrace.GetClass(code))
		}

		if span.Error != 0 {
			spanKV = spanKV.AddTag(itrace.TagSpanStatus, itrace.StatusErr)
		} else {
			spanKV = spanKV.AddTag(itrace.TagSpanStatus, itrace.StatusOk)
		}

		if !delMessage {
			span.ParentID = 0
			span.SpanID = 0
			span.TraceID = 0
			if buf, err := jsonIterator.Marshal(span); err != nil {
				log.Warn(err.Error())
			} else {
				spanKV = spanKV.Add(itrace.FieldMessage, string(buf))
			}
		}

		t := time.Unix(0, span.Start)
		pt := point.NewPoint(inputName, spanKV, append(traceOpts, point.WithTime(t))...)
		if isTracingMetricsEnable {
			spanMetrics(pt, labels, values) // span 指标化。
		}

		dktrace = append(dktrace, &itrace.DkSpan{Point: pt})
	}

	log.Debugf("build %d trace point from %d trace spans",
		len(dktrace), cap(dktrace)) // cap(dktrace) is the origin trace span count

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
		parentIDs[span.ParentID] = true
	}

	return
}

func Int64ToPaddedString(num uint64) string {
	str := strconv.FormatUint(num, 16)
	if len(str) < 16 {
		str = strings.Repeat("0", 16-len(str)) + str
	}

	return str
}
