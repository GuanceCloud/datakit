// +build linux

// linux only inputs
package inputs

import (
	"github.com/influxdata/telegraf/plugins/inputs/iptables"
	"github.com/influxdata/telegraf/plugins/inputs/kernel"
	"github.com/influxdata/telegraf/plugins/inputs/processes"
	"github.com/influxdata/telegraf/plugins/inputs/procstat"
	"github.com/influxdata/telegraf/plugins/inputs/systemd_units"
	"github.com/influxdata/telegraf/plugins/inputs/varnish"
)

var (
	telegrafInputsLinux = map[string]*TelegrafInput{ // Name: Catalog
		"iptables":      {name: "iptables", Catalog: "network", input: &iptables.Iptables{}},
		"kernel":        {name: "kernel", Catalog: "host", input: &kernel.Kernel{}},
		`systemd_units`: {name: "systemd_units", Catalog: "host", input: &systemd_units.SystemdUnits{}},
		"varnish":       {name: "varnish", Catalog: "varnish", input: &varnish.Varnish{}},
		"procstat":      {name: "procstat", Catalog: "processes", input: &procstat.Procstat{}},
		"processes":     {name: "processes", Catalog: "processes", input: &processes.Processes{}},
	}
)

func init() {
	for k, v := range telegrafInputsLinux {
		l.Debug("add telegraf plugin %s", k)
		TelegrafInputs[k] = v
	}
}
