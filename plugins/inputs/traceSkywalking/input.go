package traceSkywalking

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

var (
	inputName                   = "traceSkywalking"
	traceSkywalkingConfigSample = `
[[inputs.traceSkywalking]]
  ## Tracing data sample config, [rate] and [scope] together determine how many trace sample data
  ## will be send to DataFlux workspace.
  ## Sub item in sample_configs list with first priority.
  # [[inputs.traceSkywalking.sample_configs]]
    ## Sample rate, how many tracing data will be sampled
    # rate = 10
    ## Sample scope, the range to be covered in once sample action.
    # scope = 100
    ## Ignore tags list, keys appear in this list is transparent to sample function which means every trace carrying this tag will bypass sample function.
    # ignore_tags_list = []
    ## Sample target, program will search this [key, value] tag pairs to match a assgined sample config set in root span.
    # [inputs.traceSkywalking.sample_configs.target]
    # env = "prod"

  ## Sub item in sample_configs list with second priority.
  # [[inputs.traceSkywalking.sample_configs]]
    ## Sample rate, how many tracing data will be sampled.
    # rate = 100
    ## Sample scope, the range to be covered in once sample action.
    # scope = 1000
    ## Ignore tags list, keys appear in this list is transparent to sample function which means every trace carrying this tag will bypass sample function.
    # ignore_tags_list = []
    ## Sample target, program will search this [key, value] tag pairs to match a assgined sample config set in root span.
    # [inputs.traceSkywalking.sample_configs.target]
    # env = "dev"

    ## ...

  ## Sub item in sample_configs list with last priority.
  # [[inputs.traceSkywalking.sample_configs]]
    ## Sample rate, how many tracing data will be sampled.
    # rate = 10
    ## Sample scope, the range to be covered in once sample action.
    # scope = 100
    ## Ignore tags list, keys appear in this list is transparent to sample function which means every trace carrying this tag will bypass sample function.
    # ignore_tags_list = []
    ## Sample target, program will search this [key, value] tag pairs to match a assgined sample config set in root span.
    ## As general, the last item in sample_configs list without [tag, value] pair will be used as default sample rule
    ## only if all above rules mismatched.
    # [inputs.traceSkywalking.sample_configs.target]

  # [inputs.traceSkywalking.V2]
    #	grpcPort = 11800

    # [inputs.traceSkywalking.V2.tags]
      # tag1 = "tag1"
      # tag2 = "tag2"
      # ...

  # [inputs.traceSkywalking.V3]
    # grpcPort = 13800

    # [inputs.traceSkywalking.V3.tags]
      # tag1 = "tag1"
      # tag2 = "tag2"
      # ...
`
	skywalkingTagsV2 map[string]string
	skywalkingTagsV3 map[string]string
	log              = logger.DefaultSLogger(inputName)
)

var (
	sampleConfs  []*trace.TraceSampleConfig
	upstmFilters []upstmSegmentFilter
	swFilters    []swSegmentFilter
)

type Skywalking struct {
	GrpcPort int32             `toml:"grpcPort"`
	Tags     map[string]string `toml:"tags"`
}

type Input struct {
	TraceSampleConfs []*trace.TraceSampleConfig `toml:"sample_configs"`
	V2               *Skywalking                `toml:"V2"`
	V3               *Skywalking                `toml:"V3"`
}

func (_ *Input) Catalog() string {
	return inputName
}

func (_ *Input) SampleConfig() string {
	return traceSkywalkingConfigSample
}

func (t *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("%s input started...", inputName)

	sampleConfs = t.TraceSampleConfs
	for k, v := range sampleConfs {
		if v.Rate <= 0 || v.Scope < v.Rate {
			v.Rate = 100
			v.Scope = 100
			log.Warnf("%s input tracing sample config [%d] invalid, reset to default.", inputName, k)
		}
	}
	if len(sampleConfs) != 0 {
		upstmFilters = append(upstmFilters, upstmSegSample)
		swFilters = append(swFilters, swSegSample)
	}

	if t.V2 != nil {
		if t.V2.Tags != nil {
			skywalkingTagsV2 = t.V2.Tags
		}
		if t.V2.GrpcPort != 0 {
			//BoltDbInit()
			go SkyWalkingServerRunV2(fmt.Sprintf(":%d", t.V2.GrpcPort))
		}
	}

	if t.V3 != nil {
		if t.V3.Tags != nil {
			skywalkingTagsV3 = t.V3.Tags
		}
		if t.V3.GrpcPort != 0 {
			go SkyWalkingServerRunV3(t.V3)
		}
	}
}

func (t *Input) RegHttpHandler() {
	if t.V3 != nil {
		http.RegHttpHandler("POST", "/v3/segment", SkywalkingTraceHandle)
		http.RegHttpHandler("POST", "/v3/segments", SkywalkingTraceHandle)
		http.RegHttpHandler("POST", "/v3/management/reportProperties", SkywalkingTraceHandle)
		http.RegHttpHandler("POST", "/v3/management/keepAlive", SkywalkingTraceHandle)
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
