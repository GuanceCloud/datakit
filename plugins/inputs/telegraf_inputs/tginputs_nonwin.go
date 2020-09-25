// +build !windows

package telegraf_inputs

import (
	"github.com/influxdata/telegraf/plugins/inputs/processes"
	"github.com/influxdata/telegraf/plugins/inputs/varnish"
)

var (
	telegrafInputsNonWin = map[string]*TelegrafInput{ // Name: Catalog
		"varnish":   {name: "varnish", Catalog: "varnish", input: &varnish.Varnish{}},
		"processes": {name: "processes", Catalog: "processes", input: &processes.Processes{}},
	}
)

func init() {
	for k, v := range telegrafInputsNonWin {
		TelegrafInputs[k] = v
	}
}
