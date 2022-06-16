// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll
package cmds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"golang.org/x/term"
)

var (
	categoryMap = map[string]string{
		datakit.MetricDeprecated: "M",
		datakit.Metric:           "M",
		datakit.Network:          "N",
		datakit.KeyEvent:         "E",
		datakit.Object:           "O",
		datakit.Logging:          "L",
		datakit.Tracing:          "T",
		datakit.RUM:              "R",
		datakit.Security:         "S",
	}
	MaxTableWidth = 128

	inputsStatsCols = strings.Split(
		`Input,Cat,Freq,AvgFeed,Feeds,TotalPts,Filtered,1stFeed,LastFeed,AvgCost,MaxCost,Error(date)`, ",")
	plStatsCols = strings.Split(
		"Script,Cat,Namespace,Enabled,TotalPts,DropPts,ErrPts,PLUpdate,"+
			"Cost,AvgCost,1StTime,Update,UpdateTime,Deleted,Errors", ",")
	enabledInputCols = strings.Split(`Input,Instaces,Crashed`, ",")
	goroutineCols    = strings.Split(`Name,Done,Running,Total Cost,Min Cost,Max Cost,Failed`, ",")
	httpAPIStatCols  = strings.Split(`API,Total,Limited(%),Max Latency,Avg Latency,2xx,3xx,4xx,5xx`, ",")
	senderStatCols   = strings.Split(`Sink,Uptime,Count,Failed,Pts,Raw Bytes,Bytes,2xx,4xx,5xx,Timeout`, ",")
	filterRuleCols   = strings.Split("Category,Total,Filtered(%),Cost,Cost/Pts,Rules", ",")
)

func number(i interface{}) string {
	switch x := i.(type) {
	case int:
		return humanize.SI(float64(x), "")
	case uint:
		return humanize.SI(float64(x), "")
	case int64:
		return humanize.SI(float64(x), "")
	case uint64:
		return humanize.SI(float64(x), "")
	default:
		return ""
	}
}

func (m *monitorAPP) renderGolangRuntimeTable(ds *dkhttp.DatakitStats) {
	table := m.golangRuntime
	row := 0

	if m.anyError != nil { // some error occurred, we just gone
		return
	}

	if ds.GolangRuntime == nil { // on older version datakit, no golang runtime responded
		m.golangRuntime.SetTitle("Runtime Info(unavailable)").SetTitleAlign(tview.AlignLeft)
		return
	}

	table.SetCell(row, 0,
		tview.NewTableCell("Goroutines").SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1,
		tview.NewTableCell(fmt.Sprintf("%d", ds.GolangRuntime.Goroutines)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0,
		tview.NewTableCell("CPU(%)").SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1,
		tview.NewTableCell(fmt.Sprintf("%f", ds.GolangRuntime.CPUUsage*100)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0,
		tview.NewTableCell("Mem").SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1,
		tview.NewTableCell(number(ds.GolangRuntime.HeapAlloc)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0,
		tview.NewTableCell("SysMem").SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1,
		tview.NewTableCell(number(ds.GolangRuntime.Sys)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0,
		tview.NewTableCell("GC Pasued").SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1,
		tview.NewTableCell(time.Duration(ds.GolangRuntime.GCPauseTotal).String()).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0,
		tview.NewTableCell("GC Count").SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1,
		tview.NewTableCell(fmt.Sprintf("%d", ds.GolangRuntime.GCNum)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0,
		tview.NewTableCell("OpenFiles").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1,
		tview.NewTableCell(fmt.Sprintf("%d", ds.OpenFiles)).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))
}

func (m *monitorAPP) renderFilterStatsTable(ds *dkhttp.DatakitStats) {
	table := m.filterStatsTable
	row := 0
	if m.anyError != nil { // some error occurred, we just gone
		return
	}

	fs := ds.FilterStats
	now := time.Now()

	table.SetCell(row, 0, tview.NewTableCell("Pulled").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(number(fs.PullCount)).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("Rules From").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(fs.RuleSource).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("Pull Interval").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(fs.PullInterval.String()).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("Pull Failed").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(number(fs.PullFailed)).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("Pull Cost.Avg").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(fs.PullCostAvg.String()).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("Pull Cost.Max").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(fs.PullCostMax.String()).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("Last Updated").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(humanize.RelTime(fs.LastUpdate, now, "ago", "")).
		SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("Last Pull.Err").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))

	lastErrCell := tview.NewTableCell(
		func() string {
			if fs.LastErr == "" {
				return "-"
			}
			return fmt.Sprintf("%s(%s)", fs.LastErr, humanize.RelTime(fs.LastErrTime, now, "ago", ""))
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignLeft)

	if fs.LastErr != "" {
		lastErrCell.SetClickedFunc(func() bool {
			m.setupLastErr(fmt.Sprintf("%s(%s)", fs.LastErr, humanize.RelTime(fs.LastErrTime, now, "ago", "")))
			return true
		})
	}

	table.SetCell(row, 1, lastErrCell)
}

func (m *monitorAPP) renderBasicInfoTable(ds *dkhttp.DatakitStats) {
	table := m.basicInfoTable
	row := 0

	if m.anyError != nil { // some error occurred, we just gone
		table.SetCell(row, 0, tview.NewTableCell("Error").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 1, tview.NewTableCell(m.anyError.Error()).SetMaxWidth(MaxTableWidth).
			SetAlign(tview.AlignLeft).SetTextColor(tcell.ColorRed))
		return
	}

	table.SetCell(row, 0, tview.NewTableCell("Hostname").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(ds.HostName).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("Version").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(ds.Version).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("Build").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(ds.BuildAt).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("Branch").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(ds.Branch).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("Uptime").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(ds.Uptime).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("CGroup").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(ds.Cgroup).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("OS/Arch").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(ds.OSArch).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("IO").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(ds.IOChanStat).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("Elected").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(ds.Elected).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("From").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(m.url).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))
}

func (m *monitorAPP) renderEnabledInputTable(ds *dkhttp.DatakitStats, colArr []string) {
	table := m.enabledInputTable

	if m.anyError != nil {
		return
	}

	if len(ds.EnabledInputs) == 0 {
		m.enabledInputTable.SetTitle("Enabled Inputs(no inputs enabled)")
		return
	} else {
		m.enabledInputTable.SetTitle(fmt.Sprintf("Enabled Inputs(%d inputs)", len(ds.EnabledInputs)))
	}

	// set table header
	for idx := range colArr {
		table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).
			SetMaxWidth(*flagMonitorMaxTableWidth).
			SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignRight))
	}

	// sort enabled inputs(by name)
	names := []string{}
	for k := range ds.EnabledInputs {
		names = append(names, k)
	}
	sort.Strings(names)

	row := 1

	// sort inputs(by name)
	for _, k := range names {
		ei := ds.EnabledInputs[k]
		table.SetCell(row, 0, tview.NewTableCell(ei.Input).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("%d", ei.Instances)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("%d", ei.Panics)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		row++
	}
}

func (m *monitorAPP) renderFilterRulesStatsTable(ds *dkhttp.DatakitStats, colArr []string) {
	table := m.filterRulesStatsTable
	if m.anyError != nil {
		return
	}

	// set table header
	for idx := range colArr {
		table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).
			SetMaxWidth(*flagMonitorMaxTableWidth).
			SetTextColor(tcell.ColorGreen).
			SetAlign(tview.AlignRight))
	}
	row := 1

	rs := ds.FilterStats.RuleStats

	keys := []string{}
	for k := range rs {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, x := range keys {
		v := rs[x]
		col := 0
		table.SetCell(row, col, tview.NewTableCell(categoryMap[x]).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignCenter))
		col++
		table.SetCell(row, col, tview.NewTableCell(number(v.Total)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		col++

		if v.Total == 0 {
			table.SetCell(row, col, tview.NewTableCell("-").
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		} else {
			table.SetCell(row, col, tview.NewTableCell(
				fmt.Sprintf("%s(%.2f)", number(v.Filtered), 100.0*float64(v.Filtered)/float64(v.Total))).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		}

		col++
		table.SetCell(row, col, tview.NewTableCell(v.Cost.String()).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		col++
		table.SetCell(row, col, tview.NewTableCell(v.CostPerPoint.String()).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		col++
		table.SetCell(row, col, tview.NewTableCell(number(v.Conditions)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))

		row++
	}
}

func (m *monitorAPP) renderHTTPStatTable(ds *dkhttp.DatakitStats, colArr []string) {
	table := m.httpServerStatTable
	if m.anyError != nil {
		return
	}

	// set table header
	for idx := range colArr {
		table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).
			SetMaxWidth(*flagMonitorMaxTableWidth).
			SetTextColor(tcell.ColorGreen).
			SetAlign(tview.AlignRight))
	}

	row := 1

	// sort goroutines(by name)
	apis := []string{}
	for k := range ds.HTTPMetrics {
		apis = append(apis, k)
	}

	sort.Strings(apis)

	for _, x := range apis {
		v := ds.HTTPMetrics[x]

		table.SetCell(row, 0, tview.NewTableCell(x).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 1, tview.NewTableCell(number(v.TotalCount)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("%s(%.2f)", number(v.Limited), v.LimitedPercent)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 3, tview.NewTableCell(v.MaxLatency.String()).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 4, tview.NewTableCell(v.AvgLatency.String()).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 5, tview.NewTableCell(number(v.Status2xx)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 6, tview.NewTableCell(number(v.Status3xx)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 7, tview.NewTableCell(number(v.Status4xx)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 8, tview.NewTableCell(number(v.Status5xx)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))

		row++
	}
}

func (m *monitorAPP) renderGoroutineTable(ds *dkhttp.DatakitStats, colArr []string) {
	table := m.goroutineStatTable

	if m.anyError != nil {
		return
	}

	// set table header
	for idx := range colArr {
		table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).
			SetMaxWidth(*flagMonitorMaxTableWidth).
			SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignRight))
	}

	row := 1

	// sort goroutines(by name)
	names := []string{}
	for k := range ds.GoroutineStats.Items {
		names = append(names, k)
	}
	sort.Strings(names)

	for _, name := range names {
		v := ds.GoroutineStats.Items[name]

		table.SetCell(row, 0, tview.NewTableCell(name).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("%d", v.Total)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("%d", v.RunningTotal)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 3, tview.NewTableCell(v.CostTime).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 4, tview.NewTableCell(v.MinCostTime).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 5, tview.NewTableCell(v.MaxCostTime).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 6, tview.NewTableCell(fmt.Sprintf("%d", v.ErrCount)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))

		row++
	}
}

func (m *monitorAPP) renderExitPrompt() {
	fmt.Fprintf(m.exitPrompt, "[green]Refresh: %s. monitor: %s | Double ctrl+c to exit.",
		*flagMonitorRefreshInterval, time.Since(m.start).String())
}

func (m *monitorAPP) renderAnyError() {
	fmt.Fprintf(m.anyErrorPrompt, "[red]%s", m.anyError)
}

func (m *monitorAPP) renderInputsStatTable(ds *dkhttp.DatakitStats, colArr []string) {
	table := m.inputsStatTable

	if m.anyError != nil {
		return
	}

	if len(ds.InputsStats) == 0 {
		m.inputsStatTable.SetTitle("Inputs Info(no data collected)")
		return
	} else {
		m.inputsStatTable.SetTitle(fmt.Sprintf("Inputs Info(%d inputs)", len(ds.InputsStats)))
	}

	// set table header
	for idx := range colArr {
		table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).
			SetMaxWidth(*flagMonitorMaxTableWidth).
			SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignRight))
	}

	row := 1
	now := time.Now()

	isSpecifiedInputs := func(n string) bool {
		for _, x := range *flagMonitorOnlyInputs {
			if x == n {
				return true
			}
		}
		return false
	}

	// sort inputs(by name)
	inputsNames := []string{}
	for k := range ds.InputsStats {
		if len(*flagMonitorOnlyInputs) == 0 || isSpecifiedInputs(k) {
			inputsNames = append(inputsNames, k)
		}
	}
	sort.Strings(inputsNames)

	if len(*flagMonitorOnlyInputs) > 0 {
		m.inputsStatTable.SetTitle(fmt.Sprintf("Inputs Info(total %d, %d selected)",
			len(ds.InputsStats), len(*flagMonitorOnlyInputs)))
	}

	//
	// render all inputs, row by row
	//
	for _, name := range inputsNames {
		v := ds.InputsStats[name]
		table.SetCell(row, 0,
			tview.NewTableCell(name).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 1, tview.NewTableCell(func() string {
			if v, ok := categoryMap[v.Category]; ok {
				return v
			} else {
				return "-"
			}
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignCenter))
		table.SetCell(row, 2, tview.NewTableCell(func() string {
			if v.Frequency == "" {
				return "-"
			}
			return v.Frequency
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 3,
			tview.NewTableCell(fmt.Sprintf("%d", v.AvgSize)).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 4,
			tview.NewTableCell(number(v.Count)).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 5,
			tview.NewTableCell(number(v.Total)).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 6,
			tview.NewTableCell(number(v.Filtered)).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))

		table.SetCell(row, 7, tview.NewTableCell(func() string {
			return humanize.RelTime(v.First, now, "ago", "")
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 8, tview.NewTableCell(func() string {
			return humanize.RelTime(v.Last, now, "ago", "")
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 9,
			tview.NewTableCell(v.AvgCollectCost.String()).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 10,
			tview.NewTableCell(v.MaxCollectCost.String()).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))

		// carefully treat the error message column

		lastErrCell := tview.NewTableCell(
			func() string {
				if v.LastErr == "" {
					return "-"
				}
				return fmt.Sprintf("%s(%s)", v.LastErr, humanize.RelTime(v.LastErrTS, now, "ago", ""))
			}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignCenter)

		if v.LastErr != "" {
			lastErrCell.SetClickedFunc(func() bool {
				m.setupLastErr(fmt.Sprintf("%s(%s)", v.LastErr, humanize.RelTime(v.LastErrTS, now, "ago", "")))
				return true
			})
		}

		table.SetCell(row, 11, lastErrCell)

		row++
	}
}

func (m *monitorAPP) renderPLStatTable(ds *dkhttp.DatakitStats, colArr []string) {
	table := m.plStatTable

	if m.anyError != nil {
		return
	}

	if len(ds.PLStats) == 0 {
		table.SetTitle("Pipeline Info(no data collected)")
		return
	} else {
		table.SetTitle(fmt.Sprintf("Pipeline Info(%d scripts)", len(ds.PLStats)))
	}

	// set table header
	for idx := range colArr {
		table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).
			SetMaxWidth(*flagMonitorMaxTableWidth).
			SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignRight))
	}

	row := 1
	now := time.Now()

	//
	// render all pl script, row by row
	//
	for _, plStats := range ds.PLStats {
		table.SetCell(row, 0,
			tview.NewTableCell(fmt.Sprintf("%12s", plStats.Name)).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))

		table.SetCell(row, 1, tview.NewTableCell(func() string {
			if v, ok := categoryMap[plStats.Category]; ok {
				return fmt.Sprintf("%8s", v)
			} else {
				return "-"
			}
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignCenter))

		table.SetCell(row, 2, tview.NewTableCell(func() string {
			if plStats.NS == "" {
				return "-"
			}
			return fmt.Sprintf("%8s", plStats.NS)
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))

		table.SetCell(row, 3,
			tview.NewTableCell(fmt.Sprintf("%8v", plStats.Enable)).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))

		table.SetCell(row, 4,
			tview.NewTableCell(fmt.Sprintf("%12s", number(plStats.Pt))).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))

		table.SetCell(row, 5,
			tview.NewTableCell(fmt.Sprintf("%12s", number(plStats.PtDrop))).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))

		table.SetCell(row, 6,
			tview.NewTableCell(fmt.Sprintf("%12s", number(plStats.PtError))).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))

		table.SetCell(row, 7,
			tview.NewTableCell(fmt.Sprintf("%12s", number(plStats.ScriptUpdateTimes))).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))

		table.SetCell(row, 8,
			tview.NewTableCell(fmt.Sprintf("%12s", time.Duration(plStats.TotalCost))).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))

		var avgCost time.Duration
		if plStats.Pt > 0 {
			avgCost = time.Duration(plStats.TotalCost / int64(plStats.Pt))
		}
		table.SetCell(row, 9,
			tview.NewTableCell(fmt.Sprintf("%12s", avgCost)).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))

		table.SetCell(row, 10,
			tview.NewTableCell(humanize.RelTime(plStats.FirstTS, now, "ago", "")).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))

		table.SetCell(row, 11, tview.NewTableCell(humanize.RelTime(plStats.MetaTS, now, "ago", "")).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))

		table.SetCell(row, 12, tview.NewTableCell(humanize.RelTime(plStats.ScriptTS, now, "ago", "")).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))

		table.SetCell(row, 13,
			tview.NewTableCell(fmt.Sprintf("%8v", plStats.Deleted)).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))

		var errInfo string
		if plStats.CompileError != "" {
			errInfo = "Compile Error: " + plStats.CompileError + "\n\n"
		}
		if len(plStats.RunLastErrs) > 0 {
			errInfo += "Run Error:\n"
			for _, e := range plStats.RunLastErrs {
				errInfo += "  " + e + "\n"
			}
		}
		click := "__________script__________\n" +
			plStats.Script +
			"__________________________\n"
		if errInfo == "" {
			errInfo = "-"
		} else {
			click = errInfo + click
		}

		lastErrCell := tview.NewTableCell(errInfo).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignCenter)

		lastErrCell.SetClickedFunc(func() bool {
			m.setupLastErr(click)
			return true
		})

		table.SetCell(row, 14, lastErrCell)

		row++
	}
}

func (m *monitorAPP) renderSenderTable(ds *dkhttp.DatakitStats, colArr []string) {
	table := m.senderStatTable

	if m.anyError != nil {
		return
	}

	if ds.SenderStat == nil {
		m.senderStatTable.SetTitle("Sender Info(no data collected)")
		return
	} else {
		m.senderStatTable.SetTitle("Sender Info")
	}

	// set table header
	for idx := range colArr {
		table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).
			SetMaxWidth(*flagMonitorMaxTableWidth).
			SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignRight))
	}

	sinkNames := []string{}
	for category := range ds.SenderStat {
		sinkNames = append(sinkNames, category)
	}
	sort.Strings(sinkNames)

	row := 1

	for _, name := range sinkNames {
		stat := ds.SenderStat[name]
		table.SetCell(row, 0, tview.NewTableCell(func() string {
			return name
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignCenter))
		table.SetCell(row, 1, tview.NewTableCell(func() string {
			return stat.Uptime.String()
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignCenter))
		table.SetCell(row, 2, tview.NewTableCell(func() string {
			return humanize.SI(float64(stat.Count), "")
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignCenter))
		table.SetCell(row, 3, tview.NewTableCell(func() string {
			return humanize.SI(float64(stat.Failed), "")
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignCenter))
		table.SetCell(row, 4, tview.NewTableCell(func() string {
			return humanize.SI(float64(stat.Pts), "")
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignCenter))
		table.SetCell(row, 5, tview.NewTableCell(func() string {
			return humanize.SI(float64(stat.RawBytes), "")
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignCenter))
		table.SetCell(row, 6, tview.NewTableCell(func() string {
			return humanize.SI(float64(stat.Bytes), "")
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignCenter))
		table.SetCell(row, 7, tview.NewTableCell(func() string {
			return humanize.SI(float64(stat.Status2XX), "")
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignCenter))
		table.SetCell(row, 8, tview.NewTableCell(func() string {
			return humanize.SI(float64(stat.Status4XX), "")
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignCenter))
		table.SetCell(row, 9, tview.NewTableCell(func() string {
			return humanize.SI(float64(stat.Status5XX), "")
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignCenter))
		table.SetCell(row, 10, tview.NewTableCell(func() string {
			return humanize.SI(float64(stat.TimeoutCount), "")
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignCenter))

		row++
	}
}

type monitorAPP struct {
	app *tview.Application

	// UI elements
	basicInfoTable      *tview.Table
	golangRuntime       *tview.Table
	inputsStatTable     *tview.Table
	plStatTable         *tview.Table
	enabledInputTable   *tview.Table
	goroutineStatTable  *tview.Table
	httpServerStatTable *tview.Table
	senderStatTable     *tview.Table

	filterStatsTable      *tview.Table
	filterRulesStatsTable *tview.Table

	exitPrompt     *tview.TextView
	anyErrorPrompt *tview.TextView
	lastErrText    *tview.TextView

	flex *tview.Flex

	ds *dkhttp.DatakitStats

	anyError error

	refresh time.Duration
	start   time.Time
	url     string
}

func (m *monitorAPP) setupLastErr(lastErr string) {
	if m.lastErrText != nil { // change to another `last error`
		m.lastErrText.Clear()
		m.flex.RemoveItem(m.lastErrText)
	}

	m.lastErrText = tview.NewTextView().SetWordWrap(true).SetDynamicColors(true)

	m.lastErrText.SetBorder(true)
	fmt.Fprintf(m.lastErrText, "[red]%s \n\n[green]ESC/Enter to close the message", lastErr)

	m.flex.AddItem(m.lastErrText, 0, 5, false)

	m.lastErrText.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyESC || key == tcell.KeyEnter {
			m.lastErrText.Clear()
			m.flex.RemoveItem(m.lastErrText)
		}
	})
}

func (m *monitorAPP) setupFlex() {
	m.flex.Clear()

	if *flagMonitorVerbose { // with -V, we show more stats info
		m.flex.SetDirection(tview.FlexRow).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
										AddItem(m.basicInfoTable, 0, 10, false).               // basic info
										AddItem(m.golangRuntime, 0, 10, false), 0, 10, false). // golang runtime stats
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn). // all inputs running stats
										AddItem(m.enabledInputTable, 0, 2, false). // inputs config stats
										AddItem(m.inputsStatTable, 0, 8, false), 0, 15, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn). // input config/goroutine/9529 http stats
										AddItem(m.goroutineStatTable, 0, 10, false).  // goroutine group stats
										AddItem(m.httpServerStatTable, 0, 10, false), // 9529 HTTP server stats
										0, 10, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn). // filter related stats
										AddItem(m.filterStatsTable, 0, 2, false).      // filter stats
										AddItem(m.filterRulesStatsTable, 0, 8, false), // filter rules stats
				0, 10, false).
			AddItem(m.plStatTable, 0, 15, false).
			AddItem(m.senderStatTable, 0, 14, false).
			AddItem(m.anyErrorPrompt, 0, 1, false).
			AddItem(m.exitPrompt, 0, 1, false)
	} else {
		m.flex.SetDirection(tview.FlexRow).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
										AddItem(m.basicInfoTable, 0, 10, false).
										AddItem(m.golangRuntime, 0, 10, false), 0, 10, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn). // all inputs running stats
										AddItem(m.enabledInputTable, 0, 3, false). // inputs config stats
										AddItem(m.inputsStatTable, 0, 7, false), 0, 15, false).
			AddItem(m.anyErrorPrompt, 0, 1, false).
			AddItem(m.exitPrompt, 0, 1, false)
	}
}

func (m *monitorAPP) setup() {
	// basic info
	m.basicInfoTable = tview.NewTable().SetSelectable(true, false).SetBorders(false)
	m.basicInfoTable.SetBorder(true).SetTitle("Basic Info").SetTitleAlign(tview.AlignLeft)

	m.golangRuntime = tview.NewTable().SetSelectable(true, false).SetBorders(false)
	m.golangRuntime.SetBorder(true).SetTitle("Runtime Info").SetTitleAlign(tview.AlignLeft)

	// inputs running stats
	m.inputsStatTable = tview.NewTable().SetSelectable(true, false).SetBorders(false).SetSeparator(tview.Borders.Vertical)
	m.inputsStatTable.SetBorder(true).SetTitle("Inputs Info").SetTitleAlign(tview.AlignLeft)

	// pipeline running stats
	m.plStatTable = tview.NewTable().SetSelectable(true, false).SetBorders(false).SetSeparator(tview.Borders.Vertical)
	m.plStatTable.SetBorder(true).SetTitle("Pipeline Info").SetTitleAlign(tview.AlignLeft)

	// enabled inputs
	m.enabledInputTable = tview.NewTable().SetSelectable(true, false).SetBorders(false)
	m.enabledInputTable.SetBorder(true).SetTitle("Enabled Inputs").SetTitleAlign(tview.AlignLeft)

	// goroutine stats
	m.goroutineStatTable = tview.NewTable().SetSelectable(true, false).SetBorders(false).SetSeparator(tview.Borders.Vertical)
	m.goroutineStatTable.SetBorder(true).SetTitle("Goroutine Groups").SetTitleAlign(tview.AlignLeft)

	// 9592 http stats
	m.httpServerStatTable = tview.NewTable().SetSelectable(true, false).SetBorders(false).SetSeparator(tview.Borders.Vertical)
	m.httpServerStatTable.SetBorder(true).SetTitle("HTTP APIs").SetTitleAlign(tview.AlignLeft)

	// sender stats
	m.senderStatTable = tview.NewTable().SetSelectable(true, false).SetBorders(false).SetSeparator(tview.Borders.Vertical)
	m.senderStatTable.SetBorder(true).SetTitle("Sender Info").SetTitleAlign(tview.AlignLeft)
	// filter stats
	m.filterStatsTable = tview.NewTable().SetSelectable(true, false).SetBorders(false)
	m.filterStatsTable.SetBorder(true).SetTitle("Filter").SetTitleAlign(tview.AlignLeft)
	m.filterRulesStatsTable = tview.NewTable().SetSelectable(true, false).SetBorders(false).SetSeparator(tview.Borders.Vertical)
	m.filterRulesStatsTable.SetBorder(true).SetTitle("Filter Rules").SetTitleAlign(tview.AlignLeft)

	// bottom prompt
	m.exitPrompt = tview.NewTextView().SetDynamicColors(true)
	// error prompt
	m.anyErrorPrompt = tview.NewTextView().SetDynamicColors(true)

	m.flex = tview.NewFlex()
	m.setupFlex()

	go func() {
		tick := time.NewTicker(m.refresh)
		defer tick.Stop()
		var err error

		for {
			m.anyError = nil

			l.Debugf("try get stats...")

			m.ds, err = requestStats(m.url)
			if err != nil {
				m.anyError = fmt.Errorf("request stats failed: %w", err)
			}

			m.render()
			m.app = m.app.Draw()
			<-tick.C // wait
		}
	}()

	if err := m.app.SetRoot(m.flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func (m *monitorAPP) run() error {
	return m.app.Run()
}

func (m *monitorAPP) render() {
	m.anyErrorPrompt.Clear()
	if m.anyError != nil {
		m.renderAnyError()
		goto end
	}

	m.basicInfoTable.Clear()
	m.golangRuntime.Clear()

	m.inputsStatTable.Clear()
	m.enabledInputTable.Clear()

	m.plStatTable.Clear()
	if *flagMonitorVerbose {
		m.goroutineStatTable.Clear()

		// HTTPMetrics maybe nil(request timeout), so keep original table
		if m.ds.HTTPMetrics != nil {
			m.httpServerStatTable.Clear()
		}

		m.senderStatTable.Clear()
		m.filterStatsTable.Clear()
		m.filterRulesStatsTable.Clear()
	}

	m.renderBasicInfoTable(m.ds)
	m.renderGolangRuntimeTable(m.ds)
	m.renderEnabledInputTable(m.ds, enabledInputCols)
	m.renderInputsStatTable(m.ds, inputsStatsCols)
	m.renderPLStatTable(m.ds, plStatsCols)
	if *flagMonitorVerbose {
		m.renderGoroutineTable(m.ds, goroutineCols)

		if m.ds.HTTPMetrics != nil {
			m.renderHTTPStatTable(m.ds, httpAPIStatCols)
		}

		m.renderSenderTable(m.ds, senderStatCols)
		if m.ds.FilterStats != nil {
			m.renderFilterStatsTable(m.ds)
			m.renderFilterRulesStatsTable(m.ds, filterRuleCols)
		} else {
			l.Debugf("ds: %+#v")
		}
	}

end:
	m.exitPrompt.Clear()
	m.renderExitPrompt()
}

func runMonitorFlags() error {
	if *flagMonitorRefreshInterval < time.Second {
		*flagMonitorRefreshInterval = time.Second
	}

	to := config.Cfg.HTTPAPI.Listen
	if *flagMonitorTo != "" {
		to = *flagMonitorTo
	}

	m := monitorAPP{
		app:     tview.NewApplication(),
		refresh: *flagMonitorRefreshInterval,
		url:     fmt.Sprintf("http://%s/stats", to),
		start:   time.Now(),
	}

	m.setup()

	return m.run()
}

func requestStats(url string) (*dkhttp.DatakitStats, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s", string(body))
	}

	ds := dkhttp.DatakitStats{
		DisableMonofont: true,
	}
	if err = json.Unmarshal(body, &ds); err != nil {
		return nil, err
	}

	return &ds, nil
}

// cmdMonitor deprecated.
func cmdMonitor(interval time.Duration, verbose bool) {
	addr := fmt.Sprintf("http://%s/stats", config.Cfg.HTTPAPI.Listen)

	if interval < time.Second {
		interval = time.Second
	}

	run := func() {
		fmt.Print("\033[H\033[2J") // clean screen

		x, err := doCMDMonitor(addr, verbose)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println(string(x))
			fmt.Printf("(Refresh at %s)Press ctrl+c to exit.\n", interval)
		}
	}

	run() // run before sleep

	tick := time.NewTicker(interval)
	defer tick.Stop()
	for range tick.C {
		run()
	}
}

func doCMDMonitor(url string, verbose bool) ([]byte, error) {
	ds, err := requestStats(url)
	if err != nil {
		return nil, err
	}

	mdtxt, err := ds.Markdown("", verbose)
	if err != nil {
		return nil, err
	}

	width := 100
	if term.IsTerminal(0) {
		if width, _, err = term.GetSize(0); err != nil {
			width = 100
		}
	}

	leftPad := 2
	if err != nil {
		return nil, err
	} else {
		if len(mdtxt) == 0 {
			return nil, fmt.Errorf("no monitor info available")
		} else {
			result := markdown.Render(string(mdtxt), width, leftPad)
			return result, nil
		}
	}
}
