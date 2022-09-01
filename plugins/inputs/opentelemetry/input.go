// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package opentelemetry is input for opentelemetry

package opentelemetry

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	cache "gitlab.jiagouyun.com/cloudcare-tools/cliutils/diskcache"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkHTTP "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/opentelemetry/collector"
)

var (
	_ inputs.InputV2   = &Input{}
	_ inputs.HTTPInput = &Input{}
)

const (
	inputName    = "opentelemetry"
	sampleConfig = `
[[inputs.opentelemetry]]
  ## 在创建'trace',Span','resource'时，会加入很多标签，这些标签最终都会出现在'Span'中
  ## 当您不希望这些标签太多造成网络上不必要的流量损失时，可选择忽略掉这些标签
  ## 支持正则表达，注意:将所有的'.'替换成'_'
  ## When creating 'trace', 'span' and 'resource', many labels will be added, and these labels will eventually appear in all 'spans'
  ## When you don't want too many labels to cause unnecessary traffic loss on the network, you can choose to ignore these labels
  ## Support regular expression. Note!!!: all '.' Replace with '_'
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

  ## Threads config controls how many goroutines an agent cloud start.
  ## buffer is the size of jobs' buffering of worker channel.
  ## threads is the total number fo goroutines at running time.
  # [inputs.opentelemetry.threads]
    # buffer = 100
    # threads = 8

  ## Storage config a local storage space in hard dirver to cache trace data.
  ## path is the local file path used to cache data.
  ## capacity is total space size(MB) used to store data.
  # [inputs.opentelemetry.storage]
    # path = "./opentelemetry_storage"
    # capacity = 5120

  [inputs.opentelemetry.expectedHeaders]
  # 如有header配置 则请求中必须要携带 否则返回状态码500
  ## 可作为安全检测使用,必须全部小写
  # ex_version = xxx
  # ex_name = xxx
  # ...

  ## grpc
  [inputs.opentelemetry.grpc]
  ## trace for grpc
  trace_enable = true

  ## metric for grpc
  metric_enable = true

  ## grpc listen addr
  addr = "127.0.0.1:4317"

  ## http
  [inputs.opentelemetry.http]
  ## if enable=true
  ## http path (do not edit):
  ##	trace : /otel/v1/trace
  ##	metric: /otel/v1/metric
  ## use as : http://127.0.0.1:9529/otel/v1/trace . Method = POST
  enable = true
  ## return to client status_ok_code :200/202
  http_status_ok = 200
`
)

var (
	log         = logger.DefaultSLogger(inputName)
	spanStorage *collector.SpansStorage
	afterGather *itrace.AfterGather
	wpool       workerpool.WorkerPool
	storage     *itrace.Storage
)

type Input struct {
	Pipelines           map[string]string            `toml:"pipelines"` // deprecated
	GRPCCol             *otlpGrpcCollector           `toml:"grpc"`
	HTTPCol             *otlpHTTPCollector           `toml:"http"`
	KeepRareResource    bool                         `toml:"keep_rare_resource"`
	CloseResource       map[string][]string          `toml:"close_resource"`
	OmitErrStatus       []string                     `toml:"omit_err_status"`
	Sampler             *itrace.Sampler              `toml:"sampler"`
	IgnoreAttributeKeys []string                     `toml:"ignore_attribute_keys"`
	Tags                map[string]string            `toml:"tags"`
	WPConfig            *workerpool.WorkerPoolConfig `toml:"threads"`
	Storage             *itrace.Storage              `toml:"storage"`
	ExpectedHeaders     map[string]string            `toml:"expectedHeaders"`
	inputName           string
	semStop             *cliutils.Sem // start stop signal
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&itrace.TraceMeasurement{Name: inputName}}
}

func (ipt *Input) RegHTTPHandler() {
	log = logger.SLogger(ipt.inputName)

	if ipt.Storage != nil {
		if cache, err := cache.Open(ipt.Storage.Path, &cache.Option{Capacity: int64(ipt.Storage.Capacity) << 20}); err != nil {
			log.Errorf("### open cache %s with cap %dMB failed, cache.Open: %s", ipt.Storage.Path, ipt.Storage.Capacity, err)
		} else {
			ipt.Storage.SetCache(cache)
			ipt.Storage.RunStorageConsumer(log, ipt.parseOtelTraceFanout)
			storage = ipt.Storage
			log.Infof("### open cache %s with cap %dMB OK", ipt.Storage.Path, ipt.Storage.Capacity)
		}
	}

	if storage == nil {
		afterGather = itrace.NewAfterGather()
	} else {
		afterGather = itrace.NewAfterGather(itrace.WithRetry(100 * time.Millisecond))
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

	spanStorage = collector.NewSpansStorage(afterGather)
	spanStorage.GlobalTags = ipt.Tags
	if len(ipt.IgnoreAttributeKeys) > 0 {
		spanStorage.RegexpString = strings.Join(ipt.IgnoreAttributeKeys, "|")
	}

	if ipt.WPConfig != nil {
		wpool = workerpool.NewWorkerPool(ipt.WPConfig.Buffer)
		if err := wpool.Start(ipt.WPConfig.Threads); err != nil {
			log.Errorf("### start workerpool failed: %s", err.Error())
			wpool = nil
		}
	}

	log.Debugf("### register handler for /otel/v1/trace of agent %s", inputName)
	log.Debugf("### register handler for /otel/v1/metric of agent %s", inputName)
	dkHTTP.RegHTTPHandler("POST", "/otel/v1/trace", ipt.HTTPCol.apiOtlpTrace)
	dkHTTP.RegHTTPHandler("POST", "/otel/v1/metric", ipt.HTTPCol.apiOtlpMetric)
}

func (ipt *Input) Run() {
	open := false
	// 从配置文件 开启
	if ipt.HTTPCol.Enable {
		// add option
		ipt.HTTPCol.spanStorage = spanStorage
		ipt.HTTPCol.ExpectedHeaders = ipt.ExpectedHeaders
		open = true
	}
	if ipt.GRPCCol.TraceEnable || ipt.GRPCCol.MetricEnable {
		open = true
		ipt.GRPCCol.ExpectedHeaders = ipt.ExpectedHeaders
		go ipt.GRPCCol.run(spanStorage)
	}
	if open {
		go spanStorage.Run()
		for {
			select {
			case <-datakit.Exit.Wait():
				ipt.exit()
				log.Infof("### %s exit", ipt.inputName)

				return
			case <-ipt.semStop.Wait():
				ipt.exit()
				log.Infof("### %s return", ipt.inputName)

				return
			}
		}
	}

	log.Debugf("### %s agent is running...", inputName)
}

func (ipt *Input) exit() {
	ipt.GRPCCol.stop()
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
	if wpool != nil {
		wpool.Shutdown()
		log.Debug("### workerpool closed")
	}
	if storage != nil {
		if err := storage.Close(); err != nil {
			log.Error(err.Error())
		}
		log.Debugf("### storage closed")
	}
}

func (ipt *Input) parseOtelTraceFanout(param *itrace.TraceParameters) error {
	if param == nil || param.Meta == nil {
		return errors.New("invalid parameters")
	}

	switch param.Meta.Protocol {
	case "http":
		return ipt.HTTPCol.parseOtelTrace(param)
	case "grpc":
		return parseOtelTraceGRPC(spanStorage, param)
	default:
		return fmt.Errorf("invalid protocol: %s", param.Meta.Protocol)
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
