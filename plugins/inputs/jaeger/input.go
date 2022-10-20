// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package jaeger handle Jaeger tracing metrics.
package jaeger

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
	inputName    = "jaeger"
	sampleConfig = `
[[inputs.jaeger]]
  # Jaeger endpoint for receiving tracing span over HTTP.
  # Default value set as below. DO NOT MODIFY THE ENDPOINT if not necessary.
  endpoint = "/apis/traces"

  # Jaeger agent host:port address for UDP transport.
  # address = "127.0.0.1:6831"

  ## customer_tags is a list of keys contains keys set by client code like span.SetTag(key, value)
  ## that want to send to data center. Those keys set by client code will take precedence over
  ## keys in [inputs.jaeger.tags]. DOT(.) IN KEY WILL BE REPLACED BY DASH(_) WHEN SENDING.
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
  # [inputs.jaeger.close_resource]
    # service1 = ["resource1", "resource2", ...]
    # service2 = ["resource1", "resource2", ...]
    # "*" = ["close_resource_under_all_services"]
    # ...

  ## Sampler config uses to set global sampling strategy.
  ## sampling_rate used to set global sampling rate.
  # [inputs.jaeger.sampler]
    # sampling_rate = 1.0

  # [inputs.jaeger.tags]
    # key1 = "value1"
    # key2 = "value2"
    # ...

  ## Threads config controls how many goroutines an agent cloud start to handle HTTP request.
  ## buffer is the size of jobs' buffering of worker channel.
  ## threads is the total number fo goroutines at running time.
  ## timeout is the duration(ms) before a job can return a result.
  # [inputs.jaeger.threads]
    # buffer = 100
    # threads = 8

  ## Storage config a local storage space in hard dirver to cache trace data.
  ## path is the local file path used to cache data.
  ## capacity is total space size(MB) used to store data.
  # [inputs.jaeger.storage]
    # path = "./jaeger_storage"
    # capacity = 5120
`
)

var (
	log            = logger.DefaultSLogger(inputName)
	afterGatherRun itrace.AfterGatherHandler
	customerKeys   []string
	tags           map[string]string
	wkpool         *workerpool.WorkerPool
	storage        *itrace.Storage
)

type Input struct {
	Path             string                       `toml:"path"`      // deprecated
	UDPAgent         string                       `toml:"udp_agent"` // deprecated
	Pipelines        map[string]string            `toml:"pipelines"` // deprecated
	Endpoint         string                       `toml:"endpoint"`
	Address          string                       `toml:"address"`
	CustomerTags     []string                     `toml:"customer_tags"`
	KeepRareResource bool                         `toml:"keep_rare_resource"`
	CloseResource    map[string][]string          `toml:"close_resource"`
	Sampler          *itrace.Sampler              `toml:"sampler"`
	Tags             map[string]string            `toml:"tags"`
	WPConfig         *workerpool.WorkerPoolConfig `toml:"threads"`
	Storage          *itrace.Storage              `toml:"storage"`
}

func (*Input) Catalog() string { return inputName }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleConfig() string { return sampleConfig }

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
			ipt.Storage.RunStorageConsumer(log, parseJaegerTraceFanout)
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
		var err error
		if wkpool, err = workerpool.NewWorkerPool(ipt.WPConfig, log); err != nil {
			log.Errorf("### new worker-pool failed: %s", err.Error())
		} else if err = wkpool.Start(); err != nil {
			log.Errorf("### start worker-pool failed: %s", err.Error())
			wkpool = nil
		}
	}

	log.Debugf("### register handler for %s of agent %s", ipt.Endpoint, inputName)
	if ipt.Endpoint != "" {
		// itrace.StartTracingStatistic()
		http.RegHTTPHandler("POST", ipt.Endpoint, handleJaegerTrace)
	}
}

func (ipt *Input) Run() {
	if ipt.Address != "" {
		log.Debugf("### %s UDP agent is starting...", inputName)
		// itrace.StartTracingStatistic()
		if err := StartUDPAgent(ipt.Address); err != nil {
			log.Errorf("### start %s UDP agent failed: %s", inputName, err.Error())
		}
	}

	customerKeys = ipt.CustomerTags
	tags = ipt.Tags

	log.Debugf("### %s agent is running...", inputName)
}

func (ipt *Input) Terminate() {
	if wkpool != nil {
		wkpool.Shutdown()
		log.Debug("### workerpool closed")
	}
	if storage != nil {
		if err := storage.Close(); err != nil {
			log.Error(err.Error())
		}
		log.Debug("### storage closed")
	}
}

func parseJaegerTraceFanout(param *itrace.TraceParameters) error {
	if param == nil || param.Meta == nil {
		return errors.New("invalid parameters")
	}

	switch param.Meta.Protocol {
	case "http":
		return parseJaegerTrace(param)
	case "udp":
		return parseJaegerTraceUDP(param)
	default:
		return fmt.Errorf("invalid protocol: %s", param.Meta.Protocol)
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
