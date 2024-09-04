// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

import (
	"fmt"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/rivo/tview"
)

func (app *monitorAPP) renderGolangRuntimeTable(mfs map[string]*dto.MetricFamily) {
	table := app.golangRuntime
	row := 0

	if app.anyError != nil { // some error occurred, we just gone
		return
	}

	goroutines := app.mfs["datakit_goroutines"]
	heapAlloc := app.mfs["datakit_heap_alloc_bytes"]
	sysAlloc := app.mfs["datakit_sys_alloc_bytes"]
	memStat := app.mfs["datakit_mem_stat"]
	cpuUsage := app.mfs["datakit_cpu_usage"]
	gcSummary := app.mfs["datakit_gc_summary_seconds"]
	openFiles := app.mfs["datakit_open_files"]

	if goroutines != nil && len(goroutines.Metric) == 1 {
		m := goroutines.Metric[0]
		table.SetCell(row, 0,
			tview.NewTableCell("Goroutines").SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 1,
			tview.NewTableCell(fmt.Sprintf("%d", int64(m.GetGauge().GetValue()))).
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))
		row++
	}

	var heap, sysMem float64
	if heapAlloc != nil && len(heapAlloc.Metric) == 1 {
		m := heapAlloc.Metric[0]
		heap = m.GetGauge().GetValue()
	}

	if sysAlloc != nil && len(sysAlloc.Metric) == 1 {
		m := sysAlloc.Metric[0]
		sysMem = m.GetGauge().GetValue()
	}

	table.SetCell(row, 0,
		tview.NewTableCell("Sys/Heap").SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1,
		tview.NewTableCell(fmt.Sprintf("%s/%s", number(sysMem), number(heap))).
			SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))
	row++

	var rss, vms float64
	for _, m := range memStat.Metric {
		lps := m.GetLabel()
		if len(lps) != 1 {
			continue
		}

		switch lps[0].GetValue() {
		case "rss":
			rss = m.GetGauge().GetValue()
		case "vms":
			vms = m.GetGauge().GetValue()
		default: // pass
		}
	}

	// only show RSS memory usage
	table.SetCell(row, 0,
		tview.NewTableCell("RSS/VMS").SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1,
		tview.NewTableCell(fmt.Sprintf("%s/%s", number(rss), number(vms))).
			SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))
	row++

	if cpuUsage != nil && len(cpuUsage.Metric) == 1 {
		m := cpuUsage.Metric[0]
		table.SetCell(row, 0,
			tview.NewTableCell("CPU(%)").SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 1,
			tview.NewTableCell(fmt.Sprintf("%.2f", m.GetGauge().GetValue())).
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))
		row++
	}

	if gcSummary != nil && len(gcSummary.Metric) == 1 {
		m := gcSummary.Metric[0]
		table.SetCell(row, 0,
			tview.NewTableCell("GC Paused").SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 1,
			tview.NewTableCell(
				time.Duration(float64(time.Second)*m.GetSummary().GetSampleSum()).String()+
					"/"+
					number(m.GetSummary().GetSampleCount())).
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))
		row++
	}

	if openFiles != nil && len(openFiles.Metric) == 1 {
		m := openFiles.Metric[0]
		table.SetCell(row, 0,
			tview.NewTableCell("OpenFiles").SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 1,
			tview.NewTableCell(number(m.GetGauge().GetValue())).
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))
	}
}
