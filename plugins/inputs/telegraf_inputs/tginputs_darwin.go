// +build darwin

package telegraf_inputs

import (
	"github.com/influxdata/telegraf/plugins/inputs/processes"
	"github.com/influxdata/telegraf/plugins/inputs/procstat"
	"github.com/influxdata/telegraf/plugins/inputs/varnish"
)

var (
	telegrafInputsDarwin = map[string]*TelegrafInput{ // Name: Catalog
		"varnish":   {name: "varnish", Catalog: "varnish", input: &varnish.Varnish{}},
		"procstat":  {name: "procstat", Catalog: "processes", input: &procstat.Procstat{}},
		"processes": {name: "processes", Catalog: "processes", input: &processes.Processes{}},
	}
)

func init() {
	for k, v := range telegrafInputsDarwin {
		l.Debug("add telegraf plugin %s", k)
		TelegrafInputs[k] = v
	}
}
