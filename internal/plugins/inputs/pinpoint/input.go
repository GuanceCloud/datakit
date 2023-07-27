// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pinpoint handle Pinpoint APM traces.
package pinpoint

import (
	"context"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"google.golang.org/grpc"
)

var _ inputs.InputV2 = &Input{}

const (
	inputName    = "pinpoint"
	sampleConfig = `
[[inputs.pinpoint]]
  ## Pinpoint service endpoint for
  ## - Span Server
  ## - Agent Server(unimplemented, for service intactness and compatibility)
  ## - Metadata Server(unimplemented, for service intactness and compatibility)
  ## - Profiler Server(unimplemented, for service intactness and compatibility)
  address = "127.0.0.1:9991"

  ## Keep rare tracing resources list switch.
  ## If some resources are rare enough(not presend in 1 hour), those resource will always send
  ## to data center and do not consider samplers and filters.
  # keep_rare_resource = false

  ## Ignore tracing resources map like service:[resources...].
  ## The service name is the full service name in current application.
  ## The resource list is regular expressions uses to block resource names.
  ## If you want to block some resources universally under all services, you can set the
  ## service name as "*". Note: double quotes "" cannot be omitted.
  # [inputs.pinpoint.close_resource]
    # service1 = ["resource1", "resource2", ...]
    # service2 = ["resource1", "resource2", ...]
    # "*" = ["close_resource_under_all_services"]
    # ...

  ## Sampler config uses to set global sampling strategy.
  ## sampling_rate used to set global sampling rate.
  # [inputs.pinpoint.sampler]
    # sampling_rate = 1.0

  # [inputs.pinpoint.tags]
    # key1 = "value1"
    # key2 = "value2"
    # ...

  ## Storage config a local storage space in hard dirver to cache trace data.
  ## path is the local file path used to cache data.
  ## capacity is total space size(MB) used to store data.
  # [inputs.pinpoint.storage]
    # path = "./pinpoint_storage"
    # capacity = 5120
`
)

var (
	log            = logger.DefaultSLogger(inputName)
	afterGatherRun itrace.AfterGatherFunc
	tags           map[string]string
	reqMetaTab     = &sync.Map{}
	agentMetaData  = &AgentMetaData{}
	gsvr           *grpc.Server
	localCache     *storage.Storage
	spanSender     *itrace.SpanSender
)

type Input struct {
	Address          string                 `toml:"address"`
	KeepRareResource bool                   `toml:"keep_rare_resource"`
	CloseResource    map[string][]string    `toml:"close_resource"`
	Sampler          *itrace.Sampler        `toml:"sampler"`
	Tags             map[string]string      `toml:"tags"`
	LocalCacheConfig *storage.StorageConfig `toml:"storage"`

	feeder  dkio.Feeder
	semStop *cliutils.Sem // start stop signal
}

func (*Input) Catalog() string { return inputName }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&itrace.TraceMeasurement{Name: inputName}}
}

func (ipt *Input) Run() {
	log = logger.SLogger(inputName)

	var err error
	if ipt.LocalCacheConfig != nil {
		if localCache, err = storage.NewStorage(ipt.LocalCacheConfig, log); err != nil {
			log.Errorf("### new local-cache failed: %s", err.Error())
		}
	}

	var afterGather *itrace.AfterGather
	if localCache != nil && localCache.Enabled() {
		afterGather = itrace.NewAfterGather(
			itrace.WithLogger(log),
			itrace.WithRetry(100*time.Millisecond),
			itrace.WithIOBlockingMode(true),
			itrace.WithPointOptions(point.WithExtraTags(dkpt.GlobalHostTags())),
			itrace.WithFeeder(ipt.feeder),
		)
	} else {
		afterGather = itrace.NewAfterGather(itrace.WithLogger(log),
			itrace.WithPointOptions(point.WithExtraTags(dkpt.GlobalHostTags())), itrace.WithFeeder(ipt.feeder))
	}
	afterGatherRun = afterGather.Run

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

	if spanSender, err = itrace.NewSpanSender(inputName, 256, time.Second, afterGatherRun, log); err != nil {
		log.Errorf("### SpanSender is essential for pinpoint agent and failed to initialize: %s", err.Error())

		return
	}
	spanSender.Start()

	tags = ipt.Tags

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_pinpoint"})
	g.Go(func(ctx context.Context) error {
		runGRPCV1(ipt.Address)

		return nil
	})

	select {
	case <-datakit.Exit.Wait():
		ipt.exit()
		log.Info("pinpoint exit")
		return
	case <-ipt.semStop.Wait():
		ipt.exit()
		log.Info("pinpoint return")
		return
	}
}

func (*Input) exit() {
	if localCache != nil {
		if err := localCache.Close(); err != nil {
			log.Error(err.Error())
		}
		log.Debug("### local storage closed")
	}
	if gsvr != nil {
		gsvr.Stop()
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
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
