package traceSkywalking

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

var (
	defRate         = 15
	defScope        = 100
	traceSampleConf *trace.TraceSampleConfig
)

var (
	inputName                   = "traceSkywalking"
	traceSkywalkingConfigSample = `
[[inputs.traceSkywalking]]
## trace sample config, sample_rate and sample_scope together determine how many trace sample data will send to io
[inputs.traceSkywalking.sample_config]
  ## sample rate, how many will be sampled
  # rate = ` + fmt.Sprintf("%d", defRate) + `
  ## sample scope, the range to sample
  # scope = ` + fmt.Sprintf("%d", defScope) + `
  ## ignore tags list for samplingx
  # ignore_tags_list = []

[inputs.traceSkywalking.V2]
  #	grpcPort = 11800

  [inputs.traceSkywalking.V2.tags]
    # tag1 = "tag1"
    # tag2 = "tag2"
    # tag3 = "tag3"

[inputs.traceSkywalking.V3]
  # grpcPort = 13800

  [inputs.traceSkywalking.V3.tags]
    # tag1 = "tag1"
    # tag2 = "tag2"
    # tag3 = "tag3"
`
	log = logger.DefaultSLogger(inputName)
)

type Skywalking struct {
	GrpcPort int32             `toml:"grpcPort"`
	Tags     map[string]string `toml:"tags"`
}

type SkywalkingTrace struct {
	TraceSampleConf *trace.TraceSampleConfig `toml:"sample_config"`
	V2              *Skywalking              `toml:"V2"`
	V3              *Skywalking              `toml:"V3"`
}

var SkywalkingTagsV2 map[string]string
var SkywalkingTagsV3 map[string]string

func (_ *SkywalkingTrace) Catalog() string {
	return inputName
}

func (_ *SkywalkingTrace) SampleConfig() string {
	return traceSkywalkingConfigSample
}

func (t *SkywalkingTrace) Run() {
	log = logger.SLogger(inputName)
	log.Infof("%s input started...", inputName)

	if t.TraceSampleConf != nil {
		if t.TraceSampleConf.Rate <= 0 {
			t.TraceSampleConf.Rate = defRate
		}
		if t.TraceSampleConf.Scope <= 0 {
			t.TraceSampleConf.Scope = defScope
		}
		traceSampleConf = t.TraceSampleConf
	}

	if t.V2 != nil {
		SkywalkingTagsV2 = t.V2.Tags
		if t.V2.GrpcPort != 0 {
			//BoltDbInit()
			go SkyWalkingServerRunV2(fmt.Sprintf(":%d", t.V2.GrpcPort))
		}
	}

	if t.V3 != nil {
		SkywalkingTagsV3 = t.V3.Tags
		if t.V3.GrpcPort != 0 {
			go SkyWalkingServerRunV3(t.V3)
		}
	}

	<-datakit.Exit.Wait()
	log.Infof("%s input exit", inputName)
}

func (t *SkywalkingTrace) RegHttpHandler() {
	if t.V3 != nil {
		http.RegHttpHandler("POST", "/v3/segment", SkywalkingTraceHandle)
		http.RegHttpHandler("POST", "/v3/segments", SkywalkingTraceHandle)
		http.RegHttpHandler("POST", "/v3/management/reportProperties", SkywalkingTraceHandle)
		http.RegHttpHandler("POST", "/v3/management/keepAlive", SkywalkingTraceHandle)
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		t := &SkywalkingTrace{}

		return t
	})
}
