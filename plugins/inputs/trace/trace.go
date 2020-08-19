package trace

import (
	"fmt"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = "trace"

	traceConfigSample = `
#[inputs.trace.jaeger]
#	[inputs.trace.jaeger.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"
#
#[inputs.trace.zipkin]
#	[inputs.trace.zipkin.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"
#
#[inputs.trace.skywalkingV2]
#	grpcPort = 11800
#	[inputs.trace.skywalkingV2.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"
#
#[inputs.trace.skywalkingV3]
#	grpcPort = 13800
#	[inputs.trace.skywalkingV3.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"
`
	log *logger.Logger
)

var JaegerTags map[string]string
var ZipkinTags map[string]string
var SkywalkingV2Tags map[string]string
var SkywalkingV3Tags map[string]string

type Jaeger struct {
	Tags map[string]string
}

type Zipkin struct {
	Tags map[string]string
}

type Skywalking struct {
	GrpcPort int32
	Tags     map[string]string
}
type Trace struct {
	Jaeger       *Jaeger
	Zipkin       *Zipkin
	SkywalkingV2 *Skywalking
	SkywalkingV3 *Skywalking
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

	if t.Jaeger != nil {
		JaegerTags = t.Jaeger.Tags
	}

	if t.Zipkin != nil {
		ZipkinTags = t.Zipkin.Tags
	}

	if t.SkywalkingV2 != nil {
		SkywalkingV2Tags = t.SkywalkingV2.Tags
		go SkyWalkingServerRunV2(fmt.Sprintf(":%d", t.SkywalkingV2.GrpcPort))
	}

	if t.SkywalkingV3 != nil {
		SkywalkingV3Tags = t.SkywalkingV3.Tags
		if t.SkywalkingV3.GrpcPort != 0 {
			go SkyWalkingServerRunV3(t.SkywalkingV3)
		}
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		t := &Trace{}
		return t
	})
}
