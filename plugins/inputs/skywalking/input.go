package skywalking

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName              = "skywalking"
	skywalkingConfigSample = `
[[inputs.skywalking]]
  [inputs.skywalking.V2]
    port = 11800
    # [inputs.skywalking.V2.tags]
      # tag1 = "tag1"
      # tag2 = "tag2"
      # ...

  [inputs.skywalking.V3]
    port = 13800
    # [inputs.skywalking.V3.tags]
      # tag1 = "tag1"
      # tag2 = "tag2"
      # ...
`
	log = logger.DefaultSLogger(inputName)
)

var (
	defSkyWalkingV2Port = 11800
	skywalkingV2Tags    map[string]string
	defSkyWalkingV3Port = 13800
	skywalkingV3Tags    map[string]string
)

type Skywalking struct {
	Port int32             `toml:"port"`
	Tags map[string]string `toml:"tags"`
}

type Input struct {
	V2 *Skywalking `toml:"V2"`
	V3 *Skywalking `toml:"V3"`
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) SampleConfig() string {
	return skywalkingConfigSample
}

func (i *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("%s input started...", inputName)

	if i.V2 != nil {
		if i.V2.Tags != nil {
			skywalkingV2Tags = i.V2.Tags
		}
		if i.V2.Port <= 0 {
			i.V2.Port = int32(defSkyWalkingV2Port)
		}
		log.Debug("start skywalking grpc v2 server")
		go skyWalkingV2ServerRun(fmt.Sprintf(":%d", i.V2.Port))
	}

	if i.V3 != nil {
		if i.V3.Tags != nil {
			skywalkingV3Tags = i.V3.Tags
		}
		if i.V3.Port <= 0 {
			i.V3.Port = int32(defSkyWalkingV3Port)
		}
		log.Debug("start skywalking grpc v3 server")
		go skyWalkingV3ServervRun(fmt.Sprintf(":%d", i.V3.Port))
	}
}

func (i *Input) RegHttpHandler() {
	if i.V3 != nil {
		http.RegHttpHandler("POST", "/v3/segment", ihttp.ProtectedHandlerFunc(handleSkyWalkingSegment, log))
		http.RegHttpHandler("POST", "/v3/segments", ihttp.ProtectedHandlerFunc(handleSkyWalkingSegments, log))
		http.RegHttpHandler("POST", "/v3/management/reportProperties", handleSkyWalkingProperties)
		http.RegHttpHandler("POST", "/v3/management/keepAlive", handleSkyWalkingKeepAlive)
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
