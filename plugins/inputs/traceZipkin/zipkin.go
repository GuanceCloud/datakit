package traceZipkin

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = "traceZipkin"

	traceZipkinConfigSample = `
#[inputs.traceZipkin]
#	pathV1 = "/api/v1/spans"
#	pathV2 = "/api/v2/spans"
#	[inputs.traceZipkin.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"
`
	log *logger.Logger
)

const (
	defaultZipkinPathV1 = "/api/v1/spans"
	defaultZipkinPathV2 = "/api/v2/spans"
)

var ZipkinTags map[string]string

type Zipkin struct {
	Tags map[string]string
}

type TraceZipkin struct {
	PathV1 string
	PathV2 string
	Tags   map[string]string
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

	if t != nil {
		ZipkinTags = t.Tags
	}

	<-datakit.Exit.Wait()
	log.Infof("%s input exit", inputName)
}

func (t *TraceZipkin) RegHttpHandler() {
	if t.PathV1 == "" {
		t.PathV1 = defaultZipkinPathV1
	}
	http.RegHttpHandler("POST", t.PathV1, ZipkinTraceHandleV1)

	if t.PathV2 == "" {
		t.PathV2 = defaultZipkinPathV2
	}
	http.RegHttpHandler("POST", t.PathV2, ZipkinTraceHandleV2)
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		t := &TraceZipkin{}
		return t
	})
}
