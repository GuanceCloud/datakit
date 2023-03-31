// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	dto "github.com/prometheus/client_model/go"
	"github.com/rivo/tview"
)

func (app *monitorAPP) renderEnabledInputTable(mfs map[string]*dto.MetricFamily, colArr []string) {
	table := app.enabledInputTable

	if app.anyError != nil {
		return
	}

	instance, panics := mfs["datakit_inputs_instance"], mfs["datakit_inputs_crash_total"]
	if instance == nil {
		app.enabledInputTable.SetTitle("Enabled [red]In[white]puts(no inputs enabled)")
		return
	}

	app.enabledInputTable.SetTitle(fmt.Sprintf("Enabled [red]In[white]puts(%d inputs)", len(instance.Metric)))

	// set table header
	for idx := range colArr {
		table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).
			SetMaxWidth(app.maxTableWidth).
			SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignRight))
	}

	row := 1
	for _, metric := range instance.Metric {
		lps := metric.GetLabel()
		metric.GetGauge().GetValue()

		name := lps[0].GetValue()
		clicked := app.inputClicked(name)

		table.SetCell(row, 0, tview.NewTableCell(name).SetClickedFunc(clicked).
			SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))

		table.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("%d", int64(metric.GetGauge().GetValue()))).
			SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))

		if panics != nil {
			for _, pm := range panics.Metric {
				if pm.GetLabel()[0].GetName() == name {
					table.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("%d", int64(pm.GetCounter().GetValue()))).
						SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
				}
			}
		} else {
			table.SetCell(row, 2, tview.NewTableCell("0").
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
		}

		row++
	}
}
