// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package skywalking handle SkyWalking tracing metrics.
package skywalking

import (
	"context"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/skywalkingapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"google.golang.org/grpc"
)

var _ inputs.InputV2 = &Input{}

const (
	inputName    = "skywalking"
	sampleConfig = `
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
	log     = logger.DefaultSLogger(inputName)
	address = "localhost:11800"
	api     *skywalkingapi.SkyAPI
	skySvr  *grpc.Server
)

type Input struct {
	V2               interface{}            `toml:"V2"`        // deprecated *skywalkingConfig
	V3               interface{}            `toml:"V3"`        // deprecated *skywalkingConfig
	Pipelines        map[string]string      `toml:"pipelines"` // deprecated
	Address          string                 `toml:"address"`
	Plugins          []string               `toml:"plugins"`
	CustomerTags     []string               `toml:"customer_tags"`
	KeepRareResource bool                   `toml:"keep_rare_resource"`
	CloseResource    map[string][]string    `toml:"close_resource"`
	Sampler          *itrace.Sampler        `toml:"sampler"`
	Tags             map[string]string      `toml:"tags"`
	LocalCacheConfig *storage.StorageConfig `toml:"storage"`
}

func (*Input) Catalog() string { return inputName }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleConfig() string { return sampleConfig }

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&skywalkingapi.MetricMeasurement{}}
}

func (ipt *Input) Run() {
	log = logger.SLogger(inputName)

	api = skywalkingapi.InitApiPluginAges(ipt.Plugins, ipt.LocalCacheConfig, ipt.CloseResource,
		ipt.KeepRareResource, ipt.Sampler, ipt.CustomerTags, ipt.Tags, inputName)
	log.Debug("start skywalking grpc v3 server")

	// start up grpc v3 routine
	if len(ipt.Address) == 0 {
		ipt.Address = address
	}
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_skywalking"})
	g.Go(func(ctx context.Context) error {
		runGRPCV3(ipt.Address)

		return nil
	})

	<-datakit.Exit.Wait()
	ipt.Terminate()
}

func (ipt *Input) Terminate() {
	api.StopStorage()
	if skySvr != nil {
		skySvr.Stop()
	}
	api.CloseLocalCache()
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
