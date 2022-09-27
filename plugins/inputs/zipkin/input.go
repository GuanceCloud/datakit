// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package zipkin handle Zipkin APM traces.
package zipkin

import (
	"errors"
	"fmt"
	"time"

	cache "gitlab.jiagouyun.com/cloudcare-tools/cliutils/diskcache"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	_ inputs.InputV2   = &Input{}
	_ inputs.HTTPInput = &Input{}
)

const (
	inputName    = "zipkin"
	sampleConfig = `
[[inputs.zipkin]]
  pathV1 = "/api/v1/spans"
  pathV2 = "/api/v2/spans"

  ## customer_tags is a list of keys contains keys set by client code like span.SetTag(key, value)
  ## that want to send to data center. Those keys set by client code will take precedence over
  ## keys in [inputs.zipkin.tags]. DOT(.) IN KEY WILL BE REPLACED BY DASH(_) WHEN SENDING.
  # customer_tags = ["key1", "key2", ...]

  ## Keep rare tracing resources list switch.
  ## If some resources are rare enough(not presend in 1 hour), those resource will always send
  ## to data center and do not consider samplers and filters.
  # keep_rare_resource = false

  ## Ignore tracing resources map like service:[resources...].
  ## The service name is the full service name in current application.
  ## The resource list is regular expressions uses to block resource names.
  ## If you want to block some resources universally under all services, you can set the
  ## service name as "*". Note: double quotes "" cannot be omitted.
  # [inputs.zipkin.close_resource]
    # service1 = ["resource1", "resource2", ...]
    # service2 = ["resource1", "resource2", ...]
    # "*" = ["close_resource_under_all_services"]
    # ...

  ## Sampler config uses to set global sampling strategy.
  ## sampling_rate used to set global sampling rate.
  # [inputs.zipkin.sampler]
    # sampling_rate = 1.0

  # [inputs.zipkin.tags]
    # key1 = "value1"
    # key2 = "value2"
    # ...

  ## Threads config controls how many goroutines an agent cloud start.
  ## buffer is the size of jobs' buffering of worker channel.
  ## threads is the total number fo goroutines at running time.
  # [inputs.zipkin.threads]
    # buffer = 100
    # threads = 8

  ## Storage config a local storage space in hard dirver to cache trace data.
  ## path is the local file path used to cache data.
  ## capacity is total space size(MB) used to store data.
  # [inputs.zipkin.storage]
    # path = "./zipkin_storage"
    # capacity = 5120
`
)

var (
	log            = logger.DefaultSLogger(inputName)
	apiv1Path      = "/api/v1/spans"
	apiv2Path      = "/api/v2/spans"
	afterGatherRun itrace.AfterGatherHandler
	customerKeys   []string
	tags           map[string]string
	wpool          workerpool.WorkerPool
	storage        *itrace.Storage
)

type Input struct {
	Pipelines        map[string]string            `toml:"pipelines"` // deprecated
	PathV1           string                       `toml:"pathV1"`
	PathV2           string                       `toml:"pathV2"`
	CustomerTags     []string                     `toml:"customer_tags"`
	KeepRareResource bool                         `toml:"keep_rare_resource"`
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
			ipt.Storage.RunStorageConsumer(log, parseZipkinTraceFanout)
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

	if ipt.WPConfig != nil {
		wpool = workerpool.NewWorkerPool(ipt.WPConfig.Buffer)
		if err := wpool.Start(ipt.WPConfig.Threads); err != nil {
			log.Errorf("### start workerpool failed: %s", err.Error())
			wpool = nil
		}
	}

	if ipt.PathV1 == "" {
		ipt.PathV1 = apiv1Path
	}
	log.Debugf("### register handler for %s of agent %s", ipt.PathV1, inputName)
	http.RegHTTPHandler("POST", ipt.PathV1, handleZipkinTraceV1)

	if ipt.PathV2 == "" {
		ipt.PathV2 = apiv2Path
	}
	log.Debugf("### register handler for %s of agent %s", ipt.PathV2, inputName)
	http.RegHTTPHandler("POST", ipt.PathV2, handleZipkinTraceV2)
}

func (ipt *Input) Run() {
	customerKeys = ipt.CustomerTags
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
		log.Debug("### storage closed")
	}
}

func parseZipkinTraceFanout(param *itrace.TraceParameters) error {
	if param == nil || param.Meta == nil {
		return errors.New("invalid parameters")
	}

	switch param.Meta.UrlPath {
	case apiv1Path:
		return parseZipkinTraceV1(param)
	case apiv2Path:
		return parseZipkinTraceV2(param)
	default:
		return fmt.Errorf("invalid url path: %s", param.Meta.UrlPath)
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
