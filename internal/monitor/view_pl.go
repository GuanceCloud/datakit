// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell/v2"
	dto "github.com/prometheus/client_model/go"
	"github.com/rivo/tview"
)

func (app *monitorAPP) renderPLStatTable(mfs map[string]*dto.MetricFamily, colArr []string) {
	table := app.plStatTable

	if app.anyError != nil {
		return
	}

	totalPts := mfs["datakit_pipeline_point_total"]
	totalErrorPts := mfs["datakit_pipeline_error_point_total"]
	totalDropPts := mfs["datakit_pipeline_drop_point_total"]
	lastUpdate := mfs["datakit_pipeline_last_update_timestamp_seconds"]
	cost := mfs["datakit_pipeline_cost_seconds"]

	if totalPts == nil {
		table.SetTitle("[red]P[white]ipeline Info(no data collected)")
		return
	} else {
		table.SetTitle(fmt.Sprintf("[red]P[white]ipeline Info(%d scripts)", len(totalPts.Metric)))
	}

	// set table header
	for idx := range colArr {
		table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).
			SetMaxWidth(app.maxTableWidth).
			SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignRight))
	}

	row := 1
	now := time.Now()
	for _, m := range totalPts.Metric {
		lps := m.GetLabel()
		var cat, name, ns string

		for _, lp := range lps {
			val := lp.GetValue()
			switch lp.GetName() {
			case "name":
				name = val
				table.SetCell(row, 0,
					tview.NewTableCell(val).
						SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			case "category":
				cat = val
				table.SetCell(row, 1,
					tview.NewTableCell(point.CatString(val).Alias()).
						SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			case "namespace":
				ns = val
				table.SetCell(row, 2,
					tview.NewTableCell(val).
						SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			}
		}

		col := 3
		table.SetCell(row, col, tview.NewTableCell(number(m.GetCounter().GetValue())).
			SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
		col++

		if totalErrorPts != nil {
			x := metricWithLabel(totalErrorPts, cat, name, ns)
			if x == nil {
				table.SetCell(row, col, tview.NewTableCell("-").
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
			} else {
				table.SetCell(row, col, tview.NewTableCell(number(x.GetCounter().GetValue())).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			}
		} else {
			table.SetCell(row, col, tview.NewTableCell("-").
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
		}
		col++

		if totalDropPts != nil {
			x := metricWithLabel(totalDropPts, cat, name, ns)
			if x == nil {
				table.SetCell(row, col, tview.NewTableCell("-").
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
			} else {
				table.SetCell(row, col, tview.NewTableCell(number(x.GetCounter().GetValue())).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			}
		} else {
			table.SetCell(row, col, tview.NewTableCell("-").
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
		}
		col++

		if lastUpdate != nil {
			x := metricWithLabel(lastUpdate, cat, name, ns)
			if x == nil {
				table.SetCell(row, col, tview.NewTableCell("-").
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
			} else {
				since := humanize.RelTime(time.Unix(int64(x.GetGauge().GetValue()), 0), now, "ago", "")
				table.SetCell(row, col, tview.NewTableCell(since).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			}
		} else {
			table.SetCell(row, col, tview.NewTableCell("-").
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
		}
		col++

		if cost != nil {
			x := metricWithLabel(cost, cat, name, ns)
			if x == nil {
				table.SetCell(row, col, tview.NewTableCell("-").
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
			} else {
				avg := x.GetSummary().GetSampleSum() / float64(x.GetSummary().GetSampleCount())
				table.SetCell(row, col, tview.NewTableCell(time.Duration(avg*float64(time.Second)).String()).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			}
		} else {
			table.SetCell(row, col, tview.NewTableCell("-").
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
		}

		row++
	}
}
