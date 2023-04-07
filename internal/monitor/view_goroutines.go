// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	dto "github.com/prometheus/client_model/go"
	"github.com/rivo/tview"
)

func (app *monitorAPP) renderGoroutineTable(mfs map[string]*dto.MetricFamily, colArr []string) {
	if !exitsStr(app.onlyModules, moduleGoroutine) && !app.verbose {
		return
	}

	table := app.goroutineStatTable

	if app.anyError != nil {
		return
	}

	grAlive := mfs["datakit_goroutine_alive"]
	grStopped := mfs["datakit_goroutine_stopped_total"]
	grCost := mfs["datakit_goroutine_cost"]

	if grAlive == nil {
		return
	} else {
		table.SetTitle(fmt.Sprintf("[red]G[white]oroutine Groups(%d groups)", len(grAlive.Metric)))
	}

	// set table header
	for idx := range colArr {
		table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).
			SetMaxWidth(app.maxTableWidth).
			SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignRight))
	}

	// Fill columns row by row:
	//   Name,Running,Done,Total Cost`

	row := 1
	for _, m := range grAlive.Metric {
		lps := m.GetLabel()
		var group string

		for _, lp := range lps {
			val := lp.GetValue()
			switch lp.GetName() {
			case "name":
				group = val
				table.SetCell(row, 0, tview.NewTableCell(val).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			default:
				// pass
			}
		}

		col := 1

		// Running
		table.SetCell(row, col, tview.NewTableCell(number(m.GetGauge().GetValue())).
			SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
		col++

		// Done
		if grStopped != nil {
			x := metricWithLabel(grStopped, group)
			if x == nil {
				table.SetCell(row, col, tview.NewTableCell("-").
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
			} else {
				table.SetCell(row, col, tview.NewTableCell(number(x.GetCounter().GetValue())).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
			}
		}
		col++

		// TotalCost
		if grCost != nil {
			x := metricWithLabel(grCost, group)
			if x == nil {
				table.SetCell(row, col, tview.NewTableCell("-").
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
			} else {
				avg := time.Duration(uint64(x.GetSummary().GetSampleSum()) / x.GetSummary().GetSampleCount())
				table.SetCell(row, col, tview.NewTableCell(avg.String()).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			}
		}

		row++
	}
}
