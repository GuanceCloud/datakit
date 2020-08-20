package traceZipkin

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = "traceZipkin"

	traceZipkinConfigSample = `
#[inputs.zipkin]
#	[inputs.zipkin.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"
`
	log *logger.Logger
)

var ZipkinTags       map[string]string

type Zipkin struct {
	Tags  map[string]string
}

type TraceZipkin struct {
	Zipkin       *Zipkin
}

func (_ *TraceZipkin) Catalog() string {
	return inputName
}

func (_ *TraceZipkin) SampleConfig() string {
	return traceZipkinConfigSample
}

func (t *TraceZipkin) Run() {
	log = logger.SLogger(inputName)
	log.Infof("%s input started...", inputName)

	if t.Zipkin != nil {
		ZipkinTags       = t.Zipkin.Tags
	}

	<-datakit.Exit.Wait()
	log.Infof("%s input exit", inputName)
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		t := &TraceZipkin{}
		return t
	})
}
