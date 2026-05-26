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

type Input struct {
	Pipelines           map[string]string `toml:"pipelines"`             // deprecated
	IgnoreAttributeKeys []string          `toml:"ignore_attribute_keys"` // deprecated
	CustomerTags        []string          `toml:"customer_tags"`
	CustomerTagsAll     bool              `toml:"customer_tags_all"`

	// Deprecated: 错误拼写字段。
	CustomerTagsAllDeprecated bool `toml:"costomer_tags_all"`

	TracingMetricEnable                bool `toml:"tracing_metric_enable"`
	TracingMetricDisableGlobalHostTags bool `toml:"tracing_metric_disable_global_host_tags"`

	TracingMetricTagBlacklist []string `toml:"tracing_metric_tag_blacklist"` // 指标黑名单。
	TracingMetricTagWhitelist []string `toml:"tracing_metric_tag_whitelist"`

	LogMaxLen  int         `toml:"log_max"` // KiB
	HTTPConfig *httpConfig `toml:"http"`
	GRPCConfig *gRPC       `toml:"grpc"`

	CompatibleDDTrace   bool `toml:"compatible_ddtrace"`
	CompatibleZhaoShang bool `toml:"compatible_zhaoshang"`
	CleanMessage        bool `toml:"clean_message"`

	SplitServiceName bool                         `toml:"split_service_name"`
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

	feeder          dkio.Feeder
	semStop         *cliutils.Sem // start stop signal
	Tagger          datakit.GlobalTagger
	workerPool      *workerpool.WorkerPool
	localCache      *storage.Storage
	commonAttrs     map[string]string
	customTagsX     *itrace.CustomTags
	commonAttrsRegs []*regexp.Regexp
	ptsOpts         []point.Option
	jmarshaler      jsonMarshaler
	labels          []string
}

func (*Input) Catalog() string { return inputName }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) Singleton() {}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&JVMMeasurement{},
		&itrace.TraceMeasurement{Name: inputName},
		&itrace.TracingMetricMeasurement{Source: "opentelemetry", Name: "OpenTelemetry"},
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
		log.Infof("unknown marshaler, use default protojsonMarshaler")
		ipt.jmarshaler = &protojsonMarshaler{}
	}
	ipt.customTagsX = itrace.NewCustomTags(ipt.CustomerTags, otelPubAttrs)

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
		labels := itrace.AddLabels(itrace.DefaultLabelNames, ipt.TracingMetricTagWhitelist)
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
					httpapi.HTTPStorageWrapper(storage.HTTP_KEY,
						httpStatusRespFunc,
						localCache,
						ipt.HTTPConfig.handleOTELTrace)), log, expectedHeaders))

		httpapi.RegHTTPHandler("POST",
			ipt.HTTPConfig.MetricAPI,
			httpapi.CheckExpectedHeaders(ipt.HTTPConfig.handleOTElMetrics, log, expectedHeaders))

		httpapi.RegHTTPHandler("POST",
			ipt.HTTPConfig.LogsAPI,
			httpapi.CheckExpectedHeaders(ipt.HTTPConfig.handleOTELLogging, log, expectedHeaders))

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
	if ipt.GRPCConfig != nil {
		ipt.GRPCConfig.stop()
		log.Info("grpc server stop")
	}
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
		commonAttrsRegs:  make([]*regexp.Regexp, 0),
		CleanMessage:     true,
		LogMaxLen:        500,
		// TracingMetricEnable: true,
		TracingMetricTagBlacklist: []string{"resource"},
	}
}

func init() { //nolint:gochecknoinits
	httpapi.RegInputHTTPRouteMatcher(func(method, path string) (string, bool) {
		if method != http.MethodPost {
			return "", false
		}

		switch path {
		case defaultTraceAPI, defaultMetricAPI, defaultLogAPI:
			return inputName, true
		default:
			return "", false
		}
	})

	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
