package trace

import (
	"log"

	"github.com/influxdata/telegraf"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const traceConfigSample = `### active: whether to collect trace data.
### host: http server host.
### path: url path to recieve data.

#active = true
#host   = "127.0.0.1:54133"
#path   = "/trace"
#dataway_path="/v1/write/logging"
`

var (
	acc telegraf.Accumulator
)

type Trace struct {
	Active bool
	Host   string
	Path   string
	Ftdataway string   `toml:"dataway_path"`
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

func (_ *Trace) Gather(telegraf.Accumulator) error {
	return nil
}

func (t *Trace) Start(accumulator telegraf.Accumulator) error {
	if !t.Active {
		return nil
	}

	log.Printf("I! [trace] start")
	acc = accumulator

	go t.Serve()

	return nil
}

func (_ *Trace) Stop() {
}

func init() {
	inputs.Add("trace", func() inputs.Input {
		t := &Trace{}
		return t
	})
}
