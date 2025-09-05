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
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"google.golang.org/protobuf/proto"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
)

var (
	_ inputs.InputV2   = &Input{}
	_ inputs.HTTPInput = &Input{}

	log = logger.DefaultSLogger(inputName)
)

const (
	inputName    = "opentelemetry"
	sampleConfig = `
[[inputs.opentelemetry]]
  ## customer_tags will work as a whitelist to prevent tags send to data center.
  ## All . will replace to _ ,like this :
  ## "project.name" to send to center is "project_name"
  # customer_tags = ["sink_project", "custom.otel.tag"]

  ## If set to true, all Attributes will be extracted and message.Attributes will be empty.
  # customer_tags_all = false

  ## Keep rare tracing resources list switch.
  ## If some resources are rare enough(not presend in 1 hour), those resource will always send
  ## to data center and do not consider samplers and filters.
  # keep_rare_resource = false

  ## By default every error presents in span will be send to data center and omit any filters or
  ## sampler. If you want to get rid of some error status, you can set the error status list here.
  # omit_err_status = ["404"]

  ## compatible ddtrace: It is possible to compatible OTEL Trace with DDTrace trace
  # compatible_ddtrace=false

  ## split service.name form xx.system.
  ## see: https://github.com/open-telemetry/semantic-conventions/blob/main/docs/database/database-spans.md
  split_service_name = true

  ## delete trace message
  # del_message = true

  ## logging message data max length,default is 500kb
  log_max = 500

  ## JSON marshaler: set JSON marshaler. available marshaler are:
  ##   gojson/jsoniter/protojson
  ##
  ## For better performance, gojson and jsoniter is better than protojson,
  ## for compatible reason we still use protojson as default.
  jmarshaler = "protojson"

  ## cleaned the top-level fields in message. Default true
  clean_message = true

  ## tracing_metric_enable: trace_hits trace_hits_by_http_status trace_latency trace_errors trace_errors_by_http_status trace_apdex.
  ## Extract the above metrics from the collection traces.
  # tracing_metric_enable = true

  ## Blacklist of metric tags: There are many labels in the metric: "tracing_metrics".
  ## If you want to remove certain tag, you can use the blacklist to remove them.
  ## By default, it includes: source,span_name,env,service,status,version,resource,http_status_code,http_status_class
  ## and "customer_tags", k8s related tags, and others service.
  # tracing_metric_tag_blacklist = ["resource", "operation", "tag_a", "tag_b"]

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
  ## trace : /otel/v1/traces
  ## metric: /otel/v1/metrics
  ## and the client side should be configured properly with Datakit listening port(default: 9529)
  ## or custom HTTP request path.
  ## for example http://127.0.0.1:9529/otel/v1/traces
  ## The acceptable http_status_ok values will be 200 or 202.
  [inputs.opentelemetry.http]
   http_status_ok = 200
   trace_api = "/otel/v1/traces"
   metric_api = "/otel/v1/metrics"
   logs_api = "/otel/v1/logs"

  ## OTEL agent GRPC config for trace and metrics.
  ## GRPC services for trace and metrics can be enabled respectively as setting either to be true.
  ## add is the listening on address for GRPC server.
  [inputs.opentelemetry.grpc]
   addr = "127.0.0.1:4317"
   max_payload = 16777216 # default 16MiB

  ## If 'expected_headers' is well configed, then the obligation of sending certain wanted HTTP headers is on the client side,
  ## otherwise HTTP status code 400(bad request) will be provoked.
  ## Note: expected_headers will be effected on both trace and metrics if setted up.
  # [inputs.opentelemetry.expected_headers]
  # ex_version = "1.2.3"
  # ex_name = "env_resource_name"
  # ...
`
)

type Input struct {
	Pipelines           map[string]string `toml:"pipelines"`             // deprecated
	IgnoreAttributeKeys []string          `toml:"ignore_attribute_keys"` // deprecated
	CustomerTags        []string          `toml:"customer_tags"`
	CustomerTagsAll     bool              `toml:"customer_tags_all"`
	// Deprecated: 错误拼写字段。
	CustomerTagsAllDeprecated bool `toml:"costomer_tags_all"`

	TracingMetricEnable       bool     `toml:"tracing_metric_enable"`        // 开关，默认打开。
	TracingMetricTagBlacklist []string `toml:"tracing_metric_tag_blacklist"` // 指标黑名单。

	LogMaxLen  int         `toml:"log_max"` // KiB
	HTTPConfig *httpConfig `toml:"http"`
	GRPCConfig *gRPC       `toml:"grpc"`

	CompatibleDDTrace   bool `toml:"compatible_ddtrace"`
	CompatibleZhaoShang bool `toml:"compatible_zhaoshang"`
	CleanMessage        bool `toml:"clean_message"`

	SplitServiceName bool                         `toml:"spilt_service_name"`
	DelMessage       bool                         `toml:"del_message"`
	ExpectedHeaders  map[string]string            `toml:"expected_headers"`
	KeepRareResource bool                         `toml:"keep_rare_resource"`
	CloseResource    map[string][]string          `toml:"close_resource"`
	OmitErrStatus    []string                     `toml:"omit_err_status"`
	Sampler          *itrace.Sampler              `toml:"sampler"`
	Tags             map[string]string            `toml:"tags"`
	WPConfig         *workerpool.WorkerPoolConfig `toml:"threads"`
	LocalCacheConfig *storage.StorageConfig       `toml:"storage"`

	JSONMarshaler string `toml:"jmarshaler"`

	feeder      dkio.Feeder
	semStop     *cliutils.Sem // start stop signal
	Tagger      datakit.GlobalTagger
	workerPool  *workerpool.WorkerPool
	localCache  *storage.Storage
	commonAttrs map[string]string

	ptsOpts    []point.Option
	jmarshaler jsonMarshaler
	labels     []string
}

func (*Input) Catalog() string { return inputName }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) Singleton() {}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&JVMMeasurement{},
		&itrace.TraceMeasurement{Name: inputName},
		&itrace.TracingMetricMeasurement{},
	}
}

func (ipt *Input) setup() *Input {
	log = logger.SLogger(inputName)

	switch ipt.JSONMarshaler {
	case "gojson":
		ipt.jmarshaler = &gojsonMarshaler{}
	case "jsoniter":
		ipt.jmarshaler = &jsoniterMarshaler{}
	default:
		ipt.jmarshaler = &protojsonMarshaler{}
	}

	// setup common attributes.
	for k, v := range otelPubAttrs { // deep copy
		ipt.commonAttrs[k] = v
	}

	// NOTE: CustomerTags may overwrite public common attribytes
	for _, key := range ipt.CustomerTags {
		ipt.commonAttrs[key] = strings.ReplaceAll(key, ".", "_")
	}

	ipt.ptsOpts = append(point.CommonLoggingOptions(), point.WithExtraTags(ipt.Tagger.HostTags()))
	return ipt
}

func (ipt *Input) RegHTTPHandler() {
	ipt = ipt.setup()

	if ipt.HTTPConfig == nil && ipt.GRPCConfig == nil {
		log.Infof("all otel web protocol are not enabled")

		return
	}
	if ipt.TracingMetricEnable {
		// 默认的标签 + custom tags
		labels := itrace.AddLabels(itrace.DefaultLabelNames, ipt.CustomerTags)
		labels = itrace.DelLabels(labels, ipt.TracingMetricTagBlacklist)
		ipt.labels = labels
		initP8SMetrics(labels)
	}

	var err error
	var wkpool *workerpool.WorkerPool
	if ipt.WPConfig != nil {
		if wkpool, err = workerpool.NewWorkerPool(ipt.WPConfig, log); err != nil {
			log.Errorf("new worker-pool failed: %s", err.Error())
		}

		if err = wkpool.Start(); err != nil {
			log.Errorf("start worker-pool failed: %s", err.Error())
		} else {
			ipt.workerPool = wkpool
		}
	}
	if ipt.CustomerTagsAllDeprecated {
		ipt.CustomerTagsAll = true
	}
	var localCache *storage.Storage
	if ipt.LocalCacheConfig != nil && ipt.HTTPConfig != nil {
		log.Debug("start register")
		if localCache, err = storage.NewStorage(ipt.LocalCacheConfig, log); err != nil {
			log.Errorf("new local-cache failed: %s", err.Error())
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
						log.Errorf("parse raw URL: %s failed: %s", reqpb.Url, err.Error())
					}
					ipt.HTTPConfig.handleOTELTrace(&httpapi.NopResponseWriter{}, req)

					log.Debugf("process status: buffer-size: %dkb, cost: %dms, err: %v", len(reqpb.Body)>>10, time.Since(start)/time.Millisecond, err)

					return nil
				}
			})
			if err = localCache.RunConsumeWorker(); err != nil {
				log.Errorf("run local-cache consumer failed: %s", err.Error())
			}
		}
	}

	var afterGather *itrace.AfterGather
	if localCache != nil && localCache.Enabled() {
		afterGather = itrace.NewAfterGather(
			itrace.WithLogger(log),
			itrace.WithRetry(100*time.Millisecond),
			itrace.WithPointOptions(point.WithExtraTags(ipt.Tagger.HostTags())),
			itrace.WithFeeder(ipt.feeder))
		ipt.localCache = localCache
	} else {
		afterGather = itrace.NewAfterGather(itrace.WithLogger(log),
			itrace.WithPointOptions(point.WithExtraTags(ipt.Tagger.HostTags())), itrace.WithFeeder(ipt.feeder))
	}

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

	if ipt.GRPCConfig != nil {
		ipt.GRPCConfig.afterGatherRun = afterGather
		ipt.GRPCConfig.feeder = ipt.feeder
	}

	if ipt.HTTPConfig != nil {
		ipt.HTTPConfig.input = ipt
		ipt.HTTPConfig.initConfig(afterGather)

		httpapi.RegHTTPHandler("POST", ipt.HTTPConfig.TraceAPI,
			httpapi.CheckExpectedHeaders(
				workerpool.HTTPWrapper(httpStatusRespFunc, wkpool,
					httpapi.HTTPStorageWrapper(storage.HTTP_KEY, httpStatusRespFunc, localCache, ipt.HTTPConfig.handleOTELTrace)), log, expectedHeaders))
		httpapi.RegHTTPHandler("POST", ipt.HTTPConfig.MetricAPI, httpapi.CheckExpectedHeaders(ipt.HTTPConfig.handleOTElMetrics, log, expectedHeaders))
		httpapi.RegHTTPHandler("POST", ipt.HTTPConfig.LogsAPI, httpapi.CheckExpectedHeaders(ipt.HTTPConfig.handleOTELLogging, log, expectedHeaders))

		log.Infof("register handler:trace:%s metric: %s logs:%s  of agent %s",
			ipt.HTTPConfig.TraceAPI, ipt.HTTPConfig.MetricAPI, ipt.HTTPConfig.LogsAPI, inputName)
	}
}

func (ipt *Input) Run() {
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_opentelemetry"})
	g.Go(func(ctx context.Context) error {
		if ipt.GRPCConfig != nil {
			ipt.GRPCConfig.runGRPCV1(ipt)
		}

		return nil
	})

	log.Infof("%s agent is running...", inputName)

	ticker := time.NewTicker(time.Second * 60)
	for {
		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			log.Info("opentelemetry exit")
			return
		case <-ipt.semStop.Wait():
			ipt.exit()
			log.Info("opentelemetry return")
			return
		case <-ticker.C:
			if ipt.TracingMetricEnable {
				ipt.gatherMetrics()
			}
		}
	}
}

func (ipt *Input) exit() {
	if ipt.workerPool != nil {
		ipt.workerPool.Shutdown()
		log.Info("workerpool closed")
	}
	if ipt.localCache != nil {
		if err := ipt.localCache.Close(); err != nil {
			log.Errorf("close localCache err=%v", err)
		}
		log.Info("storage closed")
	}
	if ipt.GRPCConfig != nil {
		ipt.GRPCConfig.stop()
		log.Info("grpc server stop")
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}

	httpapi.RemoveHTTPRoute("POST", ipt.HTTPConfig.TraceAPI)
	httpapi.RemoveHTTPRoute("POST", ipt.HTTPConfig.MetricAPI)
	httpapi.RemoveHTTPRoute("POST", ipt.HTTPConfig.LogsAPI)
}

func (ipt *Input) gatherMetrics() {
	startTime := time.Now()
	// 发送指标
	pts := itrace.GatherPoints(reg, ipt.Tags)
	if len(pts) > 0 {
		err := ipt.feeder.Feed(point.Metric, pts,
			dkio.WithSource(dkio.FeedSource(inputName, itrace.TracingMetricName)),
			dkio.WithCollectCost(time.Since(startTime)))
		if err != nil {
			log.Errorf("opentelemetry send metrics points error: %v", err)
		}
	}
	// reset
	reset()
}

func defaultInput() *Input {
	return &Input{
		feeder:           dkio.DefaultFeeder(),
		semStop:          cliutils.NewSem(),
		Tagger:           datakit.DefaultGlobalTagger(),
		SplitServiceName: true,
		commonAttrs:      map[string]string{},
		CleanMessage:     true,
		LogMaxLen:        500,
		// TracingMetricEnable: true,
		TracingMetricTagBlacklist: []string{"resource", "operation"},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
