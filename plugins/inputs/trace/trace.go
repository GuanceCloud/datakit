package trace

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = "trace"

	traceConfigSample = `
#[inputs.trace]
#	skywalkingGrpcV3 = ":11800"
#	skywalkingGrpcV2 = ":13800"
#	[inputs.trace.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"
`
	log *logger.Logger
)

var gTags map[string]string

type Trace struct {
	SkywalkingGrpcV3 string
	SkywalkingGrpcV2 string
	Tags             map[string]string
}

func (_ *Trace) Catalog() string {
	return "trace"
}

func (_ *Trace) SampleConfig() string {
	return traceConfigSample
}

func (t *Trace) Run() {
	log = logger.SLogger("trace")
	log.Infof("trace input started...")
	gTags = t.Tags

	if t.SkywalkingGrpcV3 != "" {
		go SkyWalkingServerV3(t.SkywalkingGrpcV3)
	}

	if t.SkywalkingGrpcV2 != "" {
		go SkyWalkingServerV2(t.SkywalkingGrpcV2)
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		t := &Trace{}
		return t
	})
}
