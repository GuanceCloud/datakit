// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package opentelemetry handle OTEL APM traces
package opentelemetry

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
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
  ## OTEL agent HTTP config for traces and metrics
  [inputs.opentelemetry.http]
  ## if enable=true
  ## http path (do not edit):
  ##  trace : /otel/v1/trace
  ##  metric: /otel/v1/metric
  ##  url: http://127.0.0.1:9529/otel/v1/trace
	##  method = POST
   enable = true
  ## return to client status_ok_code :200/202
   http_status_ok = 200

  ## OTEL agent GRPC config for traces and metrics
  [inputs.opentelemetry.grpc]
  ## enable trace
   trace_enable = true
  ## enable metrics
   metric_enable = true
  ## grpc listening on address.
   addr = "127.0.0.1:4317"

  ## During creating 'trace', 'span' and 'resource', many labels will be added, and these labels will eventually appear in all 'spans'
  ## When you don't want too many labels to cause unnecessary traffic loss on the network, you can choose to ignore these labels
  ## Support regular expression. Note!!!: '.' WILL BE REPLACED BY '_'.
  # ignore_attribute_keys = ["os_*","process_*"]

  ## Keep rare tracing resources list switch.
  ## If some resources are rare enough(not presend in 1 hour), those resource will always send
  ## to data center and do not consider samplers and filters.
  # keep_rare_resource = false

  ## By default every error presents in span will be send to data center and omit any filters or
  ## sampler. If you want to get rid of some error status, you can set the error status list here.
  # omit_err_status = ["404"]

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

  ## If 'expectedHeaders' is well configed, then the obligation of sending certain wanted HTTP headers is on the client side,
  ## otherwise HTTP status code 400(bad request) will be provoked.
  # [inputs.opentelemetry.expectedHeaders]
  # ex_version = "12.3"
  # ex_name = "my_service_name"
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
`
)

var (
	log        = logger.DefaultSLogger(inputName)
	greg       string
	tags       map[string]string
	wkpool     *workerpool.WorkerPool
	localCache *storage.Storage
)

type otlpHTTPCollector struct {
	Enabled bool   `toml:"enable"`
	Address string `toml:"addr"`
}

type otlpGrpcCollector struct {
	TraceEnabled  bool   `toml:"trace_enable"`
	MetricEnabled bool   `toml:"metric_enable"`
	Address       string `toml:"addr"`
}

type Input struct {
	Pipelines           map[string]string            `toml:"pipelines"` // deprecated
	HTTPConfig          *otlpHTTPCollector           `toml:"http"`
	GRPCConfig          *otlpGrpcCollector           `toml:"grpc"`
	KeepRareResource    bool                         `toml:"keep_rare_resource"`
	CloseResource       map[string][]string          `toml:"close_resource"`
	OmitErrStatus       []string                     `toml:"omit_err_status"`
	Sampler             *itrace.Sampler              `toml:"sampler"`
	IgnoreAttributeKeys []string                     `toml:"ignore_attribute_keys"`
	Tags                map[string]string            `toml:"tags"`
	WPConfig            *workerpool.WorkerPoolConfig `toml:"threads"`
	LocalCacheConfig    *storage.StorageConfig       `toml:"storage"`
	ExpectedHeaders     map[string]string            `toml:"expectedHeaders"`
}

func (*Input) Catalog() string { return inputName }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&itrace.TraceMeasurement{Name: inputName}}
}

func (ipt *Input) RegHTTPHandler() {
	log = logger.SLogger(ipt.inputName)

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
					handleOTELTraces(&ihttp.NopResponseWriter{}, req)

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
		afterGather = itrace.NewAfterGather(itrace.WithLogger(log), itrace.WithRetry(100*time.Millisecond), itrace.WithBlockIOModel(true))
	} else {
		afterGather = itrace.NewAfterGather(itrace.WithLogger(log))
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
	if ipt.KeepRareResource {
		keepRareResource := &itrace.KeepRareResource{}
		keepRareResource.UpdateStatus(ipt.KeepRareResource, time.Hour)
		afterGather.AppendFilter(keepRareResource.Keep)
	}
	// add sampler
	var sampler *itrace.Sampler
	if ipt.Sampler != nil && (ipt.Sampler.SamplingRateGlobal >= 0 && ipt.Sampler.SamplingRateGlobal <= 1) {
		sampler = ipt.Sampler
	} else {
		sampler = &itrace.Sampler{SamplingRateGlobal: 1}
	}
	afterGather.AppendFilter(sampler.Sample)

	tags = ipt.Tags
	if len(ipt.IgnoreAttributeKeys) > 0 {
		greg = strings.Join(ipt.IgnoreAttributeKeys, "|")
	}

	var expectedHeaders = map[string][]string{"Content-Type": {"application/x-protobuf", "application/json"}}
	for k, v := range ipt.ExpectedHeaders {
		expectedHeaders[k] = append(expectedHeaders[k], v)
	}

	log.Infof("### register handler for /otel/v1/trace of agent %s", inputName)
	dkhttp.RegHTTPHandler("POST", "/otel/v1/trace",
		ihttp.ExpectedHeaders(
			workerpool.HTTPWrapper(httpStatusRespFunc, wkpool,
				storage.HTTPWrapper(storage.HTTP_KEY, httpStatusRespFunc, localCache, handleOTELTraces)), log, expectedHeaders))

	// log.Infof("### register handler for /otel/v1/metric of agent %s", inputName)
	// dkhttp.RegHTTPHandler("POST", "/otel/v1/metric", ipt.HTTPCol.apiOtlpMetric)
}

func (ipt *Input) Run() {
	log.Infof("### %s agent is running...", inputName)

	open := false
	if ipt.HTTPConfig.Enabled {
		open = true
	}

	// g := goroutine.NewGroup(goroutine.Option{Name: "inputs_opentelemetry"})
	// if ipt.GRPCCol.TraceEnable || ipt.GRPCCol.MetricEnable {
	// 	open = true
	// 	ipt.GRPCCol.ExpectedHeaders = ipt.ExpectedHeaders
	// 	func(spanStorage *collector.SpansStorage) {
	// 		g.Go(func(ctx context.Context) error {
	// 			ipt.GRPCCol.run(spanStorage)

	// 			return nil
	// 		})
	// 	}(spanStorage)
	// }
	if open {
		// g.Go(func(ctx context.Context) error {
		// 	spanStorage.Run()

		// 	return nil
		// })
		select {
		case <-datakit.Exit.Wait():
			log.Infof("### %s exit", ipt.inputName)
		case <-ipt.semStop.Wait():
			log.Infof("### %s return", ipt.inputName)
		}
		ipt.exit()
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
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			inputName: inputName,
			semStop:   cliutils.NewSem(),
		}
	})
}
