// +build linux

// linux only inputs
package telegraf_inputs

import (
	"github.com/influxdata/telegraf/plugins/inputs/iptables"
	"github.com/influxdata/telegraf/plugins/inputs/kernel"
	"github.com/influxdata/telegraf/plugins/inputs/systemd_units"
	"github.com/influxdata/telegraf/plugins/inputs/varnish"
)

var (
	telegrafInputsLinux = map[string]*TelegrafInput{ // Name: Catalog
		"iptables":      {name: "iptables", Catalog: "network", Input: &iptables.Iptables{}},
		"kernel":        {name: "kernel", Catalog: "host", Input: &kernel.Kernel{}},
		`systemd_units`: {name: "systemd_units", Catalog: "host", Input: &systemd_units.SystemdUnits{}},
		"varnish":       {name: "varnish", Catalog: "varnish", Input: &varnish.Varnish{}},
	}
)

func init() {
	for k, v := range telegrafInputsLinux {
		TelegrafInputs[k] = v
	}
}
