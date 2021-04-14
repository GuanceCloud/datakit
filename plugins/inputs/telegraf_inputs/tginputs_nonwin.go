// +build !windows

package telegraf_inputs

import (
	"github.com/influxdata/telegraf/plugins/inputs/varnish"
)

var (
	telegrafInputsNonWin = map[string]*TelegrafInput{ // Name: Catalog
		"varnish": {name: "varnish", Catalog: "varnish", Input: &varnish.Varnish{}},
	}
)

func init() {
	for k, v := range telegrafInputsNonWin {
		TelegrafInputs[k] = v
	}
}
