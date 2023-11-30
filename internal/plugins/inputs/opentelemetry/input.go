// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package opentelemetry handle OTEL APM trace
package opentelemetry

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

var (
	_ inputs.InputV2   = &Input{}
	_ inputs.HTTPInput = &Input{}
)

const (
	inputName    = "opentelemetry"
	sampleConfig = `
[[inputs.opentelemetry]]
  ## ignore_tags will work as a blacklist to prevent tags send to data center.
  ## Every value in this list is a valid string of regular expression.
  # ignore_tags = ["block1", "block2"]

  ## Keep rare tracing resources list switch.
  ## If some resources are rare enough(not presend in 1 hour), those resource will always send
  ## to data center and do not consider samplers and filters.
  # keep_rare_resource = false

  ## By default every error presents in span will be send to data center and omit any filters or
  ## sampler. If you want to get rid of some error status, you can set the error status list here.
  # omit_err_status = ["404"]

  ## compatible ddtrace: It is possible to compatible OTEL Trace with DDTrace trace
  # compatible_ddtrace=false

  ## delete trace message
  # del_message = true

  ## Ignore tracing resources map like service:[resources...].
  ## The service name is the full service name in current application.
  ## The resource list is regular expressions uses to block resource names.
  ## If you want to block some resources universally under all services, you can set the
  ## service name as "*". Note: double quotes "" cannot be omitted.
  # [inputs.opentelemetry.close_resource]
    # service1 = ["resource1", "resource2", ...]
    # service2 = ["resource1", "resource2", ...]
    # "*" = ["close_resource_under_all_services"]
    # ...

  ## Sampler config uses to set global sampling strategy.
  ## sampling_rate used to set global sampling rate.
  # [inputs.opentelemetry.sampler]
    # sampling_rate = 1.0

  # [inputs.opentelemetry.tags]
    # key1 = "value1"
    # key2 = "value2"
    # ...

  ## Threads config controls how many goroutines an agent cloud start to handle HTTP request.
  ## buffer is the size of jobs' buffering of worker channel.
  ## threads is the total number fo goroutines at running time.
  # [inputs.opentelemetry.threads]
    # buffer = 100
    # threads = 8

  ## Storage config a local storage space in hard dirver to cache trace data.
  ## path is the local file path used to cache data.
  ## capacity is total space size(MB) used to store data.
  # [inputs.opentelemetry.storage]
    # path = "./otel_storage"
    # capacity = 5120

  ## OTEL agent HTTP config for trace and metrics
  ## If enable set to be true, trace and metrics will be received on path respectively, by default is:
  ## trace : /otel/v1/trace
  ## metric: /otel/v1/metric
  ## and the client side should be configured properly with Datakit listening port(default: 9529)
  ## or custom HTTP request path.
  ## for example http://127.0.0.1:9529/otel/v1/trace
  ## The acceptable http_status_ok values will be 200 or 202.
  [inputs.opentelemetry.http]
   enable = true
   http_status_ok = 200
   trace_api = "/otel/v1/trace"
   metric_api = "/otel/v1/metric"

  ## OTEL agent GRPC config for trace and metrics.
  ## GRPC services for trace and metrics can be enabled respectively as setting either to be true.
  ## add is the listening on address for GRPC server.
  [inputs.opentelemetry.grpc]
   trace_enable = true
   metric_enable = true
   addr = "127.0.0.1:4317"

  ## If 'expected_headers' is well configed, then the obligation of sending certain wanted HTTP headers is on the client side,
  ## otherwise HTTP status code 400(bad request) will be provoked.
  ## Note: expected_headers will be effected on both trace and metrics if setted up.
  # [inputs.opentelemetry.expected_headers]
  # ex_version = "1.2.3"
  # ex_name = "env_resource_name"
  # ...
`
)

var (
	log               = logger.DefaultSLogger(inputName)
	statusOK          = 200
	defaultTraceAPI   = "/otel/v1/trace"
	defaultMetricAPI  = "/otel/v1/metric"
	afterGatherRun    itrace.AfterGatherHandler
	ignoreTags        []*regexp.Regexp
	getAttribute      getAttributeFunc
	extractAtrributes extractAttributesFunc
	tags              map[string]string
	wkpool            *workerpool.WorkerPool
	localCache        *storage.Storage
	otelSvr           *grpc.Server
	iptGlobal         *Input
	delMessage        bool
)

type httpConfig struct {
	Enabled      bool   `toml:"enable" json:"enable"`
	StatusCodeOK int    `toml:"http_status_ok" json:"http_status_ok"`
	TraceAPI     string `toml:"trace_api" json:"trace_api"`
	MetricAPI    string `toml:"metric_api" json:"metric_api"`
}

type grpcConfig struct {
	TraceEnabled  bool   `toml:"trace_enable" json:"trace_enable"`
	MetricEnabled bool   `toml:"metric_enable" json:"metric_enable"`
	Address       string `toml:"addr" json:"addr"`
}

type Input struct {
	Pipelines           map[string]string            `toml:"pipelines"`             // deprecated
	IgnoreAttributeKeys []string                     `toml:"ignore_attribute_keys"` // deprecated
	IgnoreTags          []string                     `toml:"ignore_tags"`
	HTTPConfig          *httpConfig                  `toml:"http"`
	GRPCConfig          *grpcConfig                  `toml:"grpc"`
	CompatibleDDTrace   bool                         `toml:"compatible_ddtrace"`
	CompatibleZhaoShang bool                         `toml:"compatible_zhaoshang"`
	DelMessage          bool                         `toml:"del_message"`
	ExpectedHeaders     map[string]string            `toml:"expected_headers"`
	KeepRareResource    bool                         `toml:"keep_rare_resource"`
	CloseResource       map[string][]string          `toml:"close_resource"`
	OmitErrStatus       []string                     `toml:"omit_err_status"`
	Sampler             *itrace.Sampler              `toml:"sampler"`
	Tags                map[string]string            `toml:"tags"`
	WPConfig            *workerpool.WorkerPoolConfig `toml:"threads"`
	LocalCacheConfig    *storage.StorageConfig       `toml:"storage"`

	feeder  dkio.Feeder
	semStop *cliutils.Sem // start stop signal
	Tagger  datakit.GlobalTagger
}

func (*Input) Catalog() string { return inputName }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&Measurement{}}
}

func (ipt *Input) RegHTTPHandler() {
	log = logger.SLogger(inputName)

	var err error
	if ipt.WPConfig != nil {
		if wkpool, err = workerpool.NewWorkerPool(ipt.WPConfig, log); err != nil {
			log.Errorf("### new worker-pool failed: %s", err.Error())
		} else if err = wkpool.Start(); err != nil {
			log.Errorf("### start worker-pool failed: %s", err.Error())
		}
	}
	if ipt.LocalCacheConfig != nil {
		log.Debug("### start register")
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
					handleOTELTrace(&httpapi.NopResponseWriter{}, req)

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
			itrace.WithIOBlockingMode(true),
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
	// add rare resource keeper
	if ipt.KeepRareResource && ipt.Sampler != nil {
		keepRareResource := &itrace.KeepRareResource{}
		keepRareResource.UpdateStatus(ipt.KeepRareResource, time.Hour)
		afterGather.AppendFilter(keepRareResource.Keep)
	}
	// add sampler
	var sampler *itrace.Sampler
	if ipt.Sampler != nil && (ipt.Sampler.SamplingRateGlobal >= 0 && ipt.Sampler.SamplingRateGlobal <= 1) {
		sampler = ipt.Sampler.Init()
		afterGather.AppendFilter(sampler.Sample)
	}

	expectedHeaders := map[string][]string{"Content-Type": {"application/x-protobuf", "application/json"}}
	for k, v := range ipt.ExpectedHeaders {
		expectedHeaders[k] = append(expectedHeaders[k], v)
	}

	if ipt.HTTPConfig == nil || !ipt.HTTPConfig.Enabled {
		log.Debugf("### HTTP server in OpenTelemetry are not enabled")

		return
	}
	// 路由可能为空，为版本兼容设置默认值。
	if ipt.HTTPConfig.TraceAPI == "" {
		ipt.HTTPConfig.TraceAPI = defaultTraceAPI
	}

	if ipt.HTTPConfig.MetricAPI == "" {
		ipt.HTTPConfig.MetricAPI = defaultMetricAPI
	}

	log.Debugf("### register handler for %s of agent %s", ipt.HTTPConfig.TraceAPI, inputName)
	statusOK = ipt.HTTPConfig.StatusCodeOK
	httpapi.RegHTTPHandler("POST", ipt.HTTPConfig.TraceAPI,
		httpapi.CheckExpectedHeaders(
			workerpool.HTTPWrapper(httpStatusRespFunc, wkpool,
				httpapi.HTTPStorageWrapper(storage.HTTP_KEY, httpStatusRespFunc, localCache, handleOTELTrace)), log, expectedHeaders))

	log.Debugf("### register handler for %s of agent %s", ipt.HTTPConfig.MetricAPI, inputName)

	iptGlobal = ipt
	httpapi.RegHTTPHandler("POST", ipt.HTTPConfig.MetricAPI, httpapi.CheckExpectedHeaders(handleOTElMetrics, log, expectedHeaders))
}

func (ipt *Input) Run() {
	if (ipt.HTTPConfig == nil || !ipt.HTTPConfig.Enabled) &&
		(ipt.GRPCConfig == nil || (!ipt.GRPCConfig.MetricEnabled && !ipt.GRPCConfig.TraceEnabled)) {
		log.Debugf("### All OpenTelemetry web protocol are not enabled")

		return
	}
	for _, v := range ipt.IgnoreTags {
		if rexp, err := regexp.Compile(v); err != nil {
			log.Debug(err.Error())
		} else {
			ignoreTags = append(ignoreTags, rexp)
		}
	}
	traceOpts = append(point.DefaultLoggingOptions(), point.WithExtraTags(datakit.DefaultGlobalTagger().HostTags()))
	delMessage = ipt.DelMessage
	tags = ipt.Tags
	convertToDD = ipt.CompatibleDDTrace
	convertToZhaoShang = ipt.CompatibleZhaoShang
	getAttribute = getAttrWrapper(ignoreTags)
	extractAtrributes = extractAttrsWrapper(ignoreTags)

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_opentelemetry"})
	g.Go(func(ctx context.Context) error {
		runGRPCV1(ipt.GRPCConfig.Address, ipt)

		return nil
	})

	log.Debugf("### %s agent is running...", inputName)

	select {
	case <-datakit.Exit.Wait():
		ipt.exit()
		log.Info("opentelemetry exit")
		return
	case <-ipt.semStop.Wait():
		ipt.exit()
		log.Info("opentelemetry return")
		return
	}
}

func (ipt *Input) exit() {
	if wkpool != nil {
		wkpool.Shutdown()
		log.Info("### workerpool closed")
	}
	if localCache != nil {
		if err := localCache.Close(); err != nil {
			log.Error(err.Error())
		}
		log.Info("### storage closed")
	}
	if otelSvr != nil {
		otelSvr.GracefulStop()
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func defaultInput() *Input {
	return &Input{
		feeder:  dkio.DefaultFeeder(),
		semStop: cliutils.NewSem(),
		Tagger:  datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
