// +build windows

package telegraf_inputs

import (
	"github.com/influxdata/telegraf/plugins/inputs/win_perf_counters"
	"github.com/influxdata/telegraf/plugins/inputs/win_services"
)

var (
	telegrafInputsWin = map[string]*TelegrafInput{
		"win_services":      {name: "win_services", Catalog: "windows", input: &win_services.WinServices{}},
		"win_perf_counters": {name: "win_perf_counters", Catalog: "windows", input: &win_perf_counters.Win_PerfCounters{}},
		`dotnetclr`:         {name: "dotnetclr", Catalog: "windows", Sample: samples["dotnetclr"], input: &win_perf_counters.Win_PerfCounters{}},
		`aspdotnet`:         {name: "aspdotnet", Catalog: "windows", Sample: samples["aspdotnet"], input: &win_perf_counters.Win_PerfCounters{}},
		`msexchange`:        {name: "msexchange", Catalog: "windows", Sample: samples["msexchange"], input: &win_perf_counters.Win_PerfCounters{}},
		`iis`:               {name: "iis", Catalog: "iis", Sample: samples["iis"], input: &win_perf_counters.Win_PerfCounters{}},
		`active_directory`:  {name: "active_directory", Catalog: "active_directory", Sample: samples["active_directory"], input: &win_perf_counters.Win_PerfCounters{}},
	}
)

func init() {
	for k, v := range telegrafInputsWin {
		TelegrafInputs[k] = v
	}
}
