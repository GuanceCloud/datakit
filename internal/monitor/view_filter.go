// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell/v2"
	dto "github.com/prometheus/client_model/go"
	"github.com/rivo/tview"
)

func (app *monitorAPP) renderFilterRulesStatsTable(mfs map[string]*dto.MetricFamily, colArr []string) {
	table := app.filterRulesStatsTable

	if app.anyError != nil {
		return
	}

	// set table header
	for idx := range colArr {
		table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).
			SetMaxWidth(app.maxTableWidth).
			SetTextColor(tcell.ColorGreen).
			SetAlign(tview.AlignRight))
	}

	ptsTotal := mfs["datakit_filter_point_total"]
	if ptsTotal == nil {
		return
	}

	ptsDroppedTotal := mfs["datakit_filter_point_dropped_total"]
	filterCost := mfs["datakit_filter_latency_seconds"]

	row := 1
	for _, m := range ptsTotal.Metric {
		lps := m.GetLabel()
		var (
			filter,
			source,
			cat string
			ptsTotal,
			ptsDropped float64
		)

		// Cat,Total,Filtered(%),Cost
		for _, lp := range lps {
			val := lp.GetValue()
			switch lp.GetName() {
			case "filters":
				filter = val
			case "source":
				source = val
			case "category":
				cat = val
				table.SetCell(row, 0, tview.NewTableCell(point.CatString(val).Alias()).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))

				ptsTotal = m.GetCounter().GetValue()
				table.SetCell(row, 1, tview.NewTableCell(number(ptsTotal)).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
			}
		}

		col := 2
		if ptsDroppedTotal != nil {
			x := metricWithLabel(ptsDroppedTotal, cat, filter, source)
			if x == nil {
				ptsDropped = 0.0
			} else {
				ptsDropped = x.GetCounter().GetValue()
			}

			table.SetCell(row, col, tview.NewTableCell(number(ptsDropped/ptsTotal*100.0)).
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
		}
		col++

		if filterCost != nil {
			x := metricWithLabel(filterCost, cat, filter, source)
			if x == nil {
				table.SetCell(row, col, tview.NewTableCell("-").
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
			} else {
				cost := time.Duration(float64(time.Second) * x.GetSummary().GetSampleSum())
				table.SetCell(row, col, tview.NewTableCell(cost.String()).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
			}
		}

		row++
	}
}

func (app *monitorAPP) renderFilterStatsTable(mfs map[string]*dto.MetricFamily) {
	table := app.filterStatsTable
	row := 0
	if app.anyError != nil { // some error occurred, we just gone
		return
	}

	pulled := mfs["datakit_filter_pull_latency_seconds"]
	if pulled == nil {
		return
	}

	pullCnt := 0.0
	if x := metricWithLabel(pulled, "ok"); x != nil {
		pullCnt += float64(x.GetSummary().GetSampleCount())
	}
	if x := metricWithLabel(pulled, "failed"); x != nil {
		pullCnt += float64(x.GetSummary().GetSampleCount())
	}

	table.SetCell(row, 0, tview.NewTableCell("Pulled").SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(number(pullCnt)).SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))
	row++

	lastPull := mfs["datakit_filter_last_update_timestamp_seconds"]
	if lastPull != nil {
		table.SetCell(row, 0, tview.NewTableCell("Last Updated").
			SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))

		t := time.Unix(int64(lastPull.Metric[0].GetGauge().GetValue()), 0)
		table.SetCell(row, 1, tview.NewTableCell(humanize.RelTime(t, time.Now(), "ago", "")).
			SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))
	}
}
