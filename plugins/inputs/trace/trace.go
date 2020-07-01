package trace

import (
	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

var (
	traceConfigSample = `### active: whether to collect trace data.
### path: url path to recieve data.

#active = true
#path   = "/trace"
`
	log *zap.SugaredLogger
)

type Trace struct {
	Active bool
	Path   string
}

func (_ *Trace) Catalog() string {
	return "trace"
}

func (_ *Trace) SampleConfig() string {
	return traceConfigSample
}

func (_ *Trace) Description() string {
	return "Collect Trace Data"
}

func (t *Trace) Run() {
	if !t.Active {
		return
	}

	log = logger.SLogger("mock")
	log.Info("mock trace started...")

	io.RegisterRoute(t.Path, writeTracing)
}

func init() {
	inputs.Add("trace", func() inputs.Input {
		t := &Trace{}
		return t
	})
}
