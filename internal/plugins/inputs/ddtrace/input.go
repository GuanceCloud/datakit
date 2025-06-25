// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package ddtrace handle DDTrace APM traces.
package ddtrace

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
	"google.golang.org/protobuf/proto"
)

var (
	_ inputs.InputV2   = &Input{}
	_ inputs.HTTPInput = &Input{}

	customObjectFeedName = dkio.FeedSource(inputName, "CO")
)

const (
	inputName = "ddtrace"

	// TraceIDUpper Tag used to propagate the higher-order 64 bits of a 128-bit trace id encoded as a
	// lower-case hexadecimal string with no zero-padding or `0x` prefix.
	TraceIDUpper = "_dd.p.tid"

	sampleConfig = `
[[inputs.ddtrace]]
  ## DDTrace Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## NOTE: DO NOT EDIT.
  endpoints = ["/v0.3/traces", "/v0.4/traces", "/v0.5/traces"]

  ## customer_tags will work as a whitelist to prevent tags send to data center.
  ## All . will replace to _ ,like this :
  ## "project.name" to send to GuanCe center is "project_name"
  # customer_tags = ["sink_project", "custom_dd_tag"]

  ## Keep rare tracing resources list switch.
  ## If some resources are rare enough(not presend in 1 hour), those resource will always send
  ## to data center and do not consider samplers and filters.
  # keep_rare_resource = false

  ## By default every error presents in span will be send to data center and omit any filters or
  ## sampler. If you want to get rid of some error status, you can set the error status list here.
  # omit_err_status = ["404"]

  ## compatible otel: It is possible to compatible OTEL Trace with DDTrace trace.
  ## make span_id and parent_id to hex encoding.
  # compatible_otel=true

  ##  It is possible to compatible B3/B3Multi TraceID with DDTrace.
  # trace_id_64_bit_hex=true

  ## When true, the tracer generates 128 bit Trace IDs, 
  ## and encodes Trace IDs as 32 lowercase hexadecimal characters with zero padding.
  ## default is true.
  # trace_128_bit_id = true

  ## delete trace message
  # del_message = true

  ## max spans limit on each trace. default 100000 or set to -1 to remove this limit.
  # trace_max_spans = 100000

  ## max trace body(Content-Length) limit. default 32MiB or set to -1 to remove this limit.
  # max_trace_body_mb = 32

  ## Ignore tracing resources map like service:[resources...].
  ## The service name is the full service name in current application.
  ## The resource list is regular expressions uses to block resource names.
  ## If you want to block some resources universally under all services, you can set the
  ## service name as "*". Note: double quotes "" cannot be omitted.
  # [inputs.ddtrace.close_resource]
  #   service1 = ["resource1", "resource2", ...]
  #   service2 = ["resource1", "resource2", ...]
  #   "*" = ["close_resource_under_all_services"]
  #   ...

  ## Sampler config uses to set global sampling strategy.
  ## sampling_rate used to set global sampling rate.
  # [inputs.ddtrace.sampler]
  #   sampling_rate = 1.0

  # [inputs.ddtrace.tags]
  #   key1 = "value1"
  #   key2 = "value2"
  #   ...

  ## Threads config controls how many goroutines an agent cloud start to handle HTTP request.
  ## buffer is the size of jobs' buffering of worker channel.
  ## threads is the total number fo goroutines at running time.
  # [inputs.ddtrace.threads]
  #   buffer = 100
  #   threads = 8

  ## Storage config a local storage space in hard dirver to cache trace data.
  ## path is the local file path used to cache data.
  ## capacity is total space size(MB) used to store data.
  # [inputs.ddtrace.storage]
  #   path = "./ddtrace_storage"
  #   capacity = 5120
`
)

var (
	log                = logger.DefaultSLogger(inputName)
	v1, v2, v3, v4, v5 = "/v0.1/spans", "/v0.2/traces", "/v0.3/traces", "/v0.4/traces", "/v0.5/traces"
	stats              = "/v0.6/stats"
	apmTelemetry       = "/telemetry/proxy/api/v2/apmtelemetry"
	info               = "/info"
	afterGatherRun     itrace.AfterGatherHandler
	inputTags          map[string]string
	wkpool             *workerpool.WorkerPool
	localCache         *storage.Storage
	traceBase          = 10
	spanBase           = 10
	delMessage         bool
	traceMaxSpans      = 100000
	maxTraceBody       = int64(32 * (1 << 20))
	noStreaming        = false
	trace128BitID      bool
)

type Input struct {
	Path             string                       `toml:"path,omitempty"`           // deprecated
	TraceSampleConfs interface{}                  `toml:"sample_configs,omitempty"` // deprecated []*itrace.TraceSampleConfig
	TraceSampleConf  interface{}                  `toml:"sample_config"`            // deprecated *itrace.TraceSampleConfig
	IgnoreResources  []string                     `toml:"ignore_resources"`         // deprecated []string
	Pipelines        map[string]string            `toml:"pipelines"`                // deprecated
	CustomerTags     []string                     `toml:"customer_tags"`
	Endpoints        []string                     `toml:"endpoints"`
	CompatibleOTEL   bool                         `toml:"compatible_otel"`
	TraceID64BitHex  bool                         `toml:"trace_id_64_bit_hex"`
	Trace128BitID    bool                         `toml:"trace_128_bit_id"`
	DelMessage       bool                         `toml:"del_message"`
	KeepRareResource bool                         `toml:"keep_rare_resource"`
	OmitErrStatus    []string                     `toml:"omit_err_status"`
	CloseResource    map[string][]string          `toml:"close_resource"`
	Sampler          *itrace.Sampler              `toml:"sampler"`
	Tags             map[string]string            `toml:"tags"`
	WPConfig         *workerpool.WorkerPoolConfig `toml:"threads"`
	LocalCacheConfig *storage.StorageConfig       `toml:"storage"`

	TraceMaxSpans  int   `toml:"trace_max_spans"`
	MaxTraceBodyMB int64 `toml:"max_trace_body_mb"`

	NoStreaming bool `toml:"no_streaming,omitempty"`

	feeder  dkio.Feeder
	semStop *cliutils.Sem // start stop signal
	Tagger  datakit.GlobalTagger
	om      *Manager
}

func (*Input) Catalog() string { return inputName }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&itrace.TraceMeasurement{Name: inputName}, &jvmTelemetry{}}
}

func (ipt *Input) RegHTTPHandler() {
	log = logger.SLogger(inputName)
	log.Infof("DdTrace start init and register HTTP. Input=%s", ipt.string())
	inputTags = ipt.Tags
	if ipt.CompatibleOTEL {
		spanBase = 16
	}
	if ipt.TraceID64BitHex {
		ignoreTraceIDFromTag = true
		traceBase = 16
		spanBase = 16
	}
	if len(ipt.CustomerTags) != 0 {
		setCustomTags(ipt.CustomerTags)
	}

	if ipt.TraceMaxSpans != 0 {
		traceMaxSpans = ipt.TraceMaxSpans
	}

	if ipt.MaxTraceBodyMB != 0 {
		maxTraceBody = ipt.MaxTraceBodyMB * (1 << 20)
	}

	trace128BitID = ipt.Trace128BitID
	noStreaming = ipt.NoStreaming
	delMessage = ipt.DelMessage
	traceOpts = append(point.CommonLoggingOptions(), point.WithExtraTags(ipt.Tagger.HostTags()))

	var err error
	if ipt.WPConfig != nil {
		if wkpool, err = workerpool.NewWorkerPool(ipt.WPConfig, log); err != nil {
			log.Errorf("### new worker-pool failed: %s", err.Error())
		} else if err = wkpool.Start(); err != nil {
			log.Errorf("### start worker-pool failed: %s", err.Error())
		}
	}
	if ipt.LocalCacheConfig != nil {
		if localCache, err = storage.NewStorage(ipt.LocalCacheConfig, log); err != nil {
			log.Errorf("### new local-cache failed: %s", err.Error())
		} else {
			localCache.RegisterConsumer(storage.HTTP_KEY, func(buf []byte) error {
				start := time.Now()
				reqpb := &storage.Request{}
				if err := proto.Unmarshal(buf, reqpb); err != nil {
					return err
				} else {
					req := &http.Request{
						Method:           reqpb.Method,
						Proto:            reqpb.Proto,
						ProtoMajor:       int(reqpb.ProtoMajor),
						ProtoMinor:       int(reqpb.ProtoMinor),
						Header:           storage.ConvertMapEntriesToMap(reqpb.Header),
						Body:             io.NopCloser(bytes.NewBuffer(reqpb.Body)),
						ContentLength:    reqpb.ContentLength,
						TransferEncoding: reqpb.TransferEncoding,
						Close:            reqpb.Close,
						Host:             reqpb.Host,
						Form:             storage.ConvertMapEntriesToMap(reqpb.Form),
						PostForm:         storage.ConvertMapEntriesToMap(reqpb.PostForm),
						RemoteAddr:       reqpb.RemoteAddr,
						RequestURI:       reqpb.RequestUri,
					}
					if req.URL, err = url.Parse(reqpb.Url); err != nil {
						log.Errorf("### parse raw URL: %s failed: %s", reqpb.Url, err.Error())
					}
					handleDDTraces(&httpapi.NopResponseWriter{}, req)

					log.Debugf("### process status: buffer-size: %dkb, cost: %dms, err: %v", len(reqpb.Body)>>10, time.Since(start)/time.Millisecond, err)

					return nil
				}
			})
			if err = localCache.RunConsumeWorker(); err != nil {
				log.Errorf("### run local-cache consumer failed: %s", err.Error())
			}
		}
	}

	var afterGather *itrace.AfterGather
	if localCache != nil && localCache.Enabled() {
		afterGather = itrace.NewAfterGather(
			itrace.WithLogger(log),
			itrace.WithRetry(100*time.Millisecond),
			itrace.WithPointOptions(point.WithExtraTags(ipt.Tagger.HostTags())),
			itrace.WithFeeder(ipt.feeder),
		)
	} else {
		afterGather = itrace.NewAfterGather(itrace.WithLogger(log),
			itrace.WithPointOptions(point.WithExtraTags(ipt.Tagger.HostTags())), itrace.WithFeeder(ipt.feeder))
	}
	afterGatherRun = afterGather

	// add filters: the order of appending filters into AfterGather is important!!!
	// the order of appending represents the order of that filter executes.
	// add close resource filter
	if len(ipt.CloseResource) != 0 {
		closeResource := &itrace.CloseResource{}
		closeResource.UpdateIgnResList(ipt.CloseResource)
		afterGather.AppendFilter(closeResource.Close)
	}

	// add error status penetration
	afterGather.AppendFilter(itrace.PenetrateErrorTracing)
	// add omit certain error status list
	if len(ipt.OmitErrStatus) != 0 {
		afterGather.AppendFilter(itrace.OmitHTTPStatusCodeFilterWrapper(ipt.OmitErrStatus))
	}
	// add rare resource keeper
	if ipt.KeepRareResource && ipt.Sampler != nil && ipt.Sampler.SamplingRateGlobal < 1 {
		keepRareResource := &itrace.KeepRareResource{}
		keepRareResource.UpdateStatus(ipt.KeepRareResource, time.Hour)
		afterGather.AppendFilter(keepRareResource.Keep)
	}
	// add penetration filter for rum
	afterGather.AppendFilter(func(log *logger.Logger, dktrace itrace.DatakitTrace) (itrace.DatakitTrace, bool) {
		for i := range dktrace {
			if dktrace[i].GetTag("_dd.origin") == "rum" {
				log.Debugf("penetrate rum trace, tid: %s service: %s resource: %s.",
					dktrace[i].Get(itrace.FieldTraceID), dktrace[i].GetTag(itrace.TagService), dktrace[i].GetTag(itrace.FieldResource))

				return dktrace, true
			}
		}

		return dktrace, false
	})

	if ipt.Sampler != nil && (ipt.Sampler.SamplingRateGlobal >= 0 && ipt.Sampler.SamplingRateGlobal <= 1) {
		sampler := ipt.Sampler.Init()
		afterGather.AppendFilter(sampler.Sample)
	}

	log.Debugf("### register handlers %v for %s agent", ipt.Endpoints, inputName)
	var isReg bool
	for _, endpoint := range ipt.Endpoints {
		switch endpoint {
		case v1, v2, v3, v4, v5:
			httpapi.RegHTTPHandler(http.MethodPost, endpoint,
				workerpool.HTTPWrapper(httpStatusRespFunc, wkpool,
					httpapi.HTTPStorageWrapper(storage.HTTP_KEY, httpStatusRespFunc, localCache, handleDDTraces)))
			httpapi.RegHTTPHandler(http.MethodPut, endpoint,
				workerpool.HTTPWrapper(httpStatusRespFunc, wkpool,
					httpapi.HTTPStorageWrapper(storage.HTTP_KEY, httpStatusRespFunc, localCache, handleDDTraces)))
			isReg = true
			log.Debugf("### pattern %s registered for %s agent", endpoint, inputName)
		default:
			log.Debugf("### unrecognized pattern %s for %s agent", endpoint, inputName)
		}
	}
	if isReg {
		httpapi.RegHTTPHandler(http.MethodGet, info, handleDDInfo)
		httpapi.RegHTTPHandler(http.MethodGet, stats, handleDDStats)
		ipt.OMInitAndRunning()
		httpapi.RegHTTPHandler(http.MethodPost, apmTelemetry, ipt.handleDDProxy)
	}
	log.Infof("### %s agent is running...", inputName)
}

func (ipt *Input) Run() {
	select {
	case <-datakit.Exit.Wait():
		ipt.exit()
		log.Info("ddtrace exit")

		return
	case <-ipt.semStop.Wait():
		ipt.exit()
		log.Info("ddtrace return")

		return
	}
}

func (ipt *Input) exit() {
	traceOpts = []point.Option{}
	if wkpool != nil {
		wkpool.Shutdown()
		log.Debug("### workerpool closed")
	}
	if localCache != nil {
		if err := localCache.Close(); err != nil {
			log.Error(err.Error())
		}
		log.Debug("### local storage closed")
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}

	// remove route
	isReg := false
	for _, endpoint := range ipt.Endpoints {
		switch endpoint {
		case v1, v2, v3, v4, v5:
			httpapi.RemoveHTTPRoute(http.MethodPost, endpoint)
			httpapi.RemoveHTTPRoute(http.MethodPut, endpoint)
			isReg = true
			log.Debugf("### pattern %s removed for %s agent", endpoint, inputName)
		default:
			log.Debugf("### unrecognized pattern %s for %s agent", endpoint, inputName)
		}
	}
	if isReg {
		httpapi.RemoveHTTPRoute(http.MethodGet, stats)
		httpapi.RemoveHTTPRoute(http.MethodGet, info)
		httpapi.RemoveHTTPRoute(http.MethodPost, apmTelemetry)
	}
}

func (ipt *Input) string() string {
	bts, err := json.Marshal(ipt)
	if err != nil {
		return ""
	} else {
		return string(bts)
	}
}

func defaultInput() *Input {
	return &Input{
		feeder:        dkio.DefaultFeeder(),
		semStop:       cliutils.NewSem(),
		Tagger:        datakit.DefaultGlobalTagger(),
		TraceMaxSpans: traceMaxSpans,
		Trace128BitID: true,
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
