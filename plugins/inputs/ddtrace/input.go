// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package ddtrace handle DDTrace APM traces.
package ddtrace

import (
	"net/http"
	"time"

	cache "gitlab.jiagouyun.com/cloudcare-tools/cliutils/diskcache"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	_ inputs.InputV2   = &Input{}
	_ inputs.HTTPInput = &Input{}
)

const (
	inputName    = "ddtrace"
	sampleConfig = `
[[inputs.ddtrace]]
  ## DDTrace Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/v0.3/traces", "/v0.4/traces", "/v0.5/traces"]

  ## customer_tags is a list of keys contains keys set by client code like span.SetTag(key, value)
  ## that want to send to data center. Those keys set by client code will take precedence over
  ## keys in [inputs.ddtrace.tags]. DOT(.) IN KEY WILL BE REPLACED BY DASH(_) WHEN SENDING.
  # customer_tags = ["key1", "key2", ...]

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
  # [inputs.ddtrace.close_resource]
    # service1 = ["resource1", "resource2", ...]
    # service2 = ["resource1", "resource2", ...]
    # "*" = ["close_resource_under_all_services"]
    # ...

  ## Sampler config uses to set global sampling strategy.
  ## sampling_rate used to set global sampling rate.
  # [inputs.ddtrace.sampler]
    # sampling_rate = 1.0

  # [inputs.ddtrace.tags]
    # key1 = "value1"
    # key2 = "value2"
    # ...

  ## Threads config controls how many goroutines an agent cloud start.
  ## buffer is the size of jobs' buffering of worker channel.
  ## threads is the total number fo goroutines at running time.
  # [inputs.ddtrace.threads]
    # buffer = 100
    # threads = 8

  ## Storage config a local storage space in hard dirver to cache trace data.
  ## path is the local file path used to cache data.
  ## capacity is total space size(MB) used to store data.
  # [inputs.ddtrace.storage]
    # path = "./ddtrace_storage"
    # capacity = 5120
`
)

var (
	log                = logger.DefaultSLogger(inputName)
	v1, v2, v3, v4, v5 = "/v0.1/spans", "/v0.2/traces", "/v0.3/traces", "/v0.4/traces", "/v0.5/traces"
	info, stats        = "/info", "/v0.6/stats"
	afterGatherRun     itrace.AfterGatherHandler
	customerKeys       []string
	tags               map[string]string
	wpool              workerpool.WorkerPool
	storage            *itrace.Storage
)

type Input struct {
	Path             string                       `toml:"path,omitempty"`           // deprecated
	TraceSampleConfs interface{}                  `toml:"sample_configs,omitempty"` // deprecated []*itrace.TraceSampleConfig
	TraceSampleConf  interface{}                  `toml:"sample_config"`            // deprecated *itrace.TraceSampleConfig
	IgnoreResources  []string                     `toml:"ignore_resources"`         // deprecated []string
	Pipelines        map[string]string            `toml:"pipelines"`                // deprecated
	Endpoints        []string                     `toml:"endpoints"`
	CustomerTags     []string                     `toml:"customer_tags"`
	KeepRareResource bool                         `toml:"keep_rare_resource"`
	OmitErrStatus    []string                     `toml:"omit_err_status"`
	CloseResource    map[string][]string          `toml:"close_resource"`
	Sampler          *itrace.Sampler              `toml:"sampler"`
	Tags             map[string]string            `toml:"tags"`
	WPConfig         *workerpool.WorkerPoolConfig `toml:"threads"`
	Storage          *itrace.Storage              `toml:"storage"`
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
	log = logger.SLogger(inputName)

	if ipt.Storage != nil {
		if cache, err := cache.Open(ipt.Storage.Path, &cache.Option{Capacity: int64(ipt.Storage.Capacity) << 20}); err != nil {
			log.Errorf("### open cache %s with cap %dMB failed, cache.Open: %s", ipt.Storage.Path, ipt.Storage.Capacity, err)
		} else {
			ipt.Storage.SetCache(cache)
			ipt.Storage.RunStorageConsumer(log, parseDDTraces)
			storage = ipt.Storage
			log.Infof("### open cache %s with cap %dMB OK", ipt.Storage.Path, ipt.Storage.Capacity)
		}
	}

	var afterGather *itrace.AfterGather
	if storage == nil {
		afterGather = itrace.NewAfterGather()
	} else {
		afterGather = itrace.NewAfterGather(itrace.WithRetry(100*time.Millisecond), itrace.WithBlockIOModel(true))
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
	// add RespectUserRule filter to obey client priority rules.
	afterGather.AppendFilter(itrace.RespectUserRule)
	// add error status penetration
	afterGather.AppendFilter(itrace.PenetrateErrorTracing)
	// add omit certain error status list
	if len(ipt.OmitErrStatus) != 0 {
		afterGather.AppendFilter(itrace.OmitHTTPStatusCodeFilterWrapper(ipt.OmitErrStatus))
	}
	// add rare resource keeper
	if ipt.KeepRareResource {
		keepRareResource := &itrace.KeepRareResource{}
		keepRareResource.UpdateStatus(ipt.KeepRareResource, time.Hour)
		afterGather.AppendFilter(keepRareResource.Keep)
	}
	// add penetration filter for rum
	afterGather.AppendFilter(func(dktrace itrace.DatakitTrace) (itrace.DatakitTrace, bool) {
		for i := range dktrace {
			if dktrace[i].Tags["_dd.origin"] == "rum" {
				log.Debugf("penetrate rum trace, tid: %s service: %s resource: %s.", dktrace[i].TraceID, dktrace[i].Service, dktrace[i].Resource)

				return dktrace, true
			}
		}

		return dktrace, false
	})
	// add sampler
	var sampler *itrace.Sampler
	if ipt.Sampler != nil && (ipt.Sampler.SamplingRateGlobal >= 0 && ipt.Sampler.SamplingRateGlobal <= 1) {
		sampler = ipt.Sampler
	} else {
		sampler = &itrace.Sampler{SamplingRateGlobal: 1}
	}
	afterGather.AppendFilter(sampler.Sample)

	if ipt.WPConfig != nil {
		wpool = workerpool.NewWorkerPool(ipt.WPConfig.Buffer)
		if err := wpool.Start(ipt.WPConfig.Threads); err != nil {
			log.Errorf("### start workerpool failed: %s", err.Error())
			wpool = nil
		}
	}

	log.Debugf("### register handlers for %s agent", inputName)
	var isReg bool
	for _, endpoint := range ipt.Endpoints {
		switch endpoint {
		case v1, v2, v3, v4, v5:
			dkhttp.RegHTTPHandler(http.MethodPost, endpoint, handleDDTraceWithVersion(endpoint))
			dkhttp.RegHTTPHandler(http.MethodPut, endpoint, handleDDTraceWithVersion(endpoint))
			isReg = true
			log.Debugf("### pattern %s registered for %s agent", endpoint, inputName)
		default:
			log.Debugf("### unrecognized pattern %s for %s agent", endpoint, inputName)
		}
	}
	if isReg {
		// itrace.StartTracingStatistic()
		// unsupported api yet
		dkhttp.RegHTTPHandler(http.MethodPost, info, handleDDInfo)
		dkhttp.RegHTTPHandler(http.MethodPost, stats, handleDDStats)
	}
}

func (ipt *Input) Run() {
	for i := range ipt.CustomerTags {
		customerKeys = append(customerKeys, ipt.CustomerTags[i])
	}
	tags = ipt.Tags

	log.Debugf("### %s agent is running...", inputName)
}

func (ipt *Input) Terminate() {
	if wpool != nil {
		wpool.Shutdown()
		log.Debug("### workerpool closed")
	}
	if storage != nil {
		if err := storage.Close(); err != nil {
			log.Error(err.Error())
		}
		log.Debug("### local storage closed")
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
