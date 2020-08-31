// +build windows

package inputs

import (
	"github.com/influxdata/telegraf/plugins/inputs/win_perf_counters"
	"github.com/influxdata/telegraf/plugins/inputs/win_services"
)

var (
	telegrafInputsWin = map[string]*TelegrafInput{
		"win_services":      {name: "win_services", Catalog: "windows", input: &win_services.WinServices{}},
		"win_perf_counters": {name: "win_perf_counters", Catalog: "windows", input: &win_perf_counters.Win_PerfCounters{}},
		`dotnetclr`:         {name: "dotnetclr", Catalog: "windows", input: &win_perf_counters.Win_PerfCounters{}},
		`aspdotnet`:         {name: "aspdotnet", Catalog: "windows", input: &win_perf_counters.Win_PerfCounters{}},
		`msexchange`:        {name: "msexchange", Catalog: "windows", input: &win_perf_counters.Win_PerfCounters{}},
	}
)

func init() {
	for k, v := range telegrafInputsWin {
		TelegrafInputs[k] = v
	}
}
