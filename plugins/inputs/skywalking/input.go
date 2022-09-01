// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package skywalking handle SkyWalking tracing metrics.
package skywalking

import (
	"time"

	cache "gitlab.jiagouyun.com/cloudcare-tools/cliutils/diskcache"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"google.golang.org/grpc"
)

var _ inputs.InputV2 = &Input{}

const (
	inputName     = "skywalking"
	jvmMetricName = "skywalking_jvm"
	sampleConfig  = `
[[inputs.skywalking]]
  ## Skywalking grpc server listening on address.
  address = "localhost:11800"

  ## plugins is a list contains all the widgets used in program that want to be regarded as service.
  ## every key words list in plugins represents a plugin defined as special tag by skywalking.
  ## the value of the key word will be used to set the service name.
  # plugins = ["db.type"]

  ## customer_tags is a list of keys contains keys set by client code like span.SetTag(key, value)
  ## that want to send to data center. Those keys set by client code will take precedence over
  ## keys in [inputs.skywalking.tags]. DOT(.) IN KEY WILL BE REPLACED BY DASH(_) WHEN SENDING.
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
  # [inputs.skywalking.close_resource]
    # service1 = ["resource1", "resource2", ...]
    # service2 = ["resource1", "resource2", ...]
    # "*" = ["close_resource_under_all_services"]
    # ...

  ## Sampler config uses to set global sampling strategy.
  ## sampling_rate used to set global sampling rate.
  # [inputs.skywalking.sampler]
    # sampling_rate = 1.0

  # [inputs.skywalking.tags]
    # key1 = "value1"
    # key2 = "value2"
    # ...

  ## Storage config a local storage space in hard dirver to cache trace data.
  ## path is the local file path used to cache data.
  ## capacity is total space size(MB) used to store data.
  # [inputs.skywalking.storage]
    # path = "./skywalking_storage"
    # capacity = 5120
`
)

var (
	log            = logger.DefaultSLogger(inputName)
	address        = "localhost:11800"
	plugins        []string
	afterGatherRun itrace.AfterGatherHandler
	customerKeys   []string
	tags           map[string]string
	skySvr         *grpc.Server
	storage        *itrace.Storage
)

type Input struct {
	V2               interface{}         `toml:"V2"`        // deprecated *skywalkingConfig
	V3               interface{}         `toml:"V3"`        // deprecated *skywalkingConfig
	Pipelines        map[string]string   `toml:"pipelines"` // deprecated
	Address          string              `toml:"address"`
	Plugins          []string            `toml:"plugins"`
	CustomerTags     []string            `toml:"customer_tags"`
	KeepRareResource bool                `toml:"keep_rare_resource"`
	CloseResource    map[string][]string `toml:"close_resource"`
	Sampler          *itrace.Sampler     `toml:"sampler"`
	Tags             map[string]string   `toml:"tags"`
	Storage          *itrace.Storage     `toml:"storage"`
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

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&skywalkingMetricMeasurement{}}
}

func (ipt *Input) Run() {
	log = logger.SLogger(inputName)

	plugins = ipt.Plugins

	if ipt.Storage != nil {
		if cache, err := cache.Open(ipt.Storage.Path, &cache.Option{Capacity: int64(ipt.Storage.Capacity) << 20}); err != nil {
			log.Errorf("### open cache %s with cap %dMB failed, cache.Open: %s", ipt.Storage.Path, ipt.Storage.Capacity, err)
		} else {
			ipt.Storage.SetCache(cache)
			ipt.Storage.RunStorageConsumer(log, parseSegmentObjectWrapper)
			storage = ipt.Storage
			log.Infof("### open cache %s with cap %dMB OK", ipt.Storage.Path, ipt.Storage.Capacity)
		}
	}

	var afterGather *itrace.AfterGather
	if storage == nil {
		afterGather = itrace.NewAfterGather()
	} else {
		afterGather = itrace.NewAfterGather(itrace.WithRetry(100 * time.Millisecond))
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

	customerKeys = ipt.CustomerTags
	tags = ipt.Tags

	log.Debug("start skywalking grpc v3 server")

	// start up grpc v3 routine
	if len(ipt.Address) == 0 {
		ipt.Address = address
	}
	go registerServerV3(ipt.Address)
}

func (ipt *Input) Terminate() {
	if storage != nil {
		if err := storage.Close(); err != nil {
			log.Error(err.Error())
		}
		log.Debug("### storage closed")
	}
	if skySvr != nil {
		skySvr.Stop()
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
