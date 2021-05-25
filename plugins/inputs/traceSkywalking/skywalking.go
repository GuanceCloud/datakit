package traceSkywalking

import (
	"fmt"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = "traceSkywalking"

	traceJaegerConfigSample = `
#[inputs.traceSkywalking.V2]
#	grpcPort = 11800
#	[inputs.traceSkywalking.V2.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"
#
#[inputs.traceSkywalking.V3]
#	grpcPort = 13800
#	[inputs.traceSkywalking.V3.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"
`
	log *logger.Logger
)

type Skywalking struct {
	GrpcPort int32
	Tags     map[string]string
}

type SkywalkingTrace struct {
	V2 *Skywalking
	V3 *Skywalking
}

var SkywalkingTagsV2 map[string]string
var SkywalkingTagsV3 map[string]string

func (_ *SkywalkingTrace) Catalog() string {
	return inputName
}

func (_ *SkywalkingTrace) SampleConfig() string {
	return traceJaegerConfigSample
}

func (t *SkywalkingTrace) Run() {
	log = logger.SLogger(inputName)
	log.Infof("%s input started...", inputName)

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
