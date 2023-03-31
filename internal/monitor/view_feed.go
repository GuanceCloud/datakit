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

func (app *monitorAPP) renderInputsFeedTable(mfs map[string]*dto.MetricFamily, colArr []string) {
	table := app.inputsStatTable

	if app.anyError != nil {
		return
	}

	feedTotal := mfs["datakit_io_feed_total"]
	ptsTotal := mfs["datakit_io_feed_point_total"]
	lastFeed := mfs["datakit_io_last_feed"]
	cost := mfs["datakit_input_collect_latency"]
	ptsFilter := mfs["datakit_io_input_filter_point_total"]
	errCount := mfs["datakit_error_total"]

	if feedTotal == nil {
		app.inputsStatTable.SetTitle("[red]In[white]puts Info(no data collected)")
		return
	} else {
		app.inputsStatTable.SetTitle(fmt.Sprintf("[red]In[white]puts Info(%d inputs)", len(feedTotal.Metric)))
	}

	// set table header
	for idx := range colArr {
		table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).
			SetMaxWidth(app.maxTableWidth).
			SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignRight))
	}

	if len(app.onlyInputs) > 0 {
		table.SetTitle(fmt.Sprintf("[red]In[white]puts Info(total %d, %d selected)",
			len(feedTotal.Metric), len(app.onlyInputs)))
	}

	row := 1

	//
	// render all inputs feed info, row by row
	//
	for _, m := range feedTotal.Metric {
		lps := m.GetLabel()
		var inputName,
			cat string

		for _, lp := range lps {
			val := lp.GetValue()

			switch lp.GetName() {
			case "name":
				inputName = val
				table.SetCell(row, 0,
					tview.NewTableCell(val).
						SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))

			case "category":

				cat = val

				table.SetCell(row, 1,
					tview.NewTableCell(point.CatString(val).Alias() /* metric -> M */).
						SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
			default:
				// pass
			}
		}

		col := 2

		table.SetCell(row, col, tview.NewTableCell(number(m.GetCounter().GetValue())).
			SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
		col++

		if ptsTotal != nil {
			x := metricWithLabel(ptsTotal, cat, inputName)
			if x == nil {
				table.SetCell(row, col, tview.NewTableCell("-").
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
			} else {
				table.SetCell(row, col, tview.NewTableCell(number(x.GetCounter().GetValue())).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			}
		}
		col++

		if ptsFilter != nil {
			x := metricWithLabel(ptsFilter, cat, inputName)
			if x == nil {
				table.SetCell(row, col, tview.NewTableCell("-").
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
			} else {
				table.SetCell(row, col, tview.NewTableCell(number(x.GetCounter().GetValue())).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			}
		}
		col++

		now := time.Now()
		if lastFeed != nil {
			x := metricWithLabel(lastFeed, cat, inputName)
			if x == nil {
				table.SetCell(row, col, tview.NewTableCell("-").
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
			} else {
				since := humanize.RelTime(time.Unix(int64(x.GetGauge().GetValue()), 0), now, "ago", "")
				table.SetCell(row, col, tview.NewTableCell(since).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			}
		}
		col++

		if cost != nil {
			x := metricWithLabel(cost, cat, inputName)
			if x == nil {
				table.SetCell(row, col, tview.NewTableCell("-").
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
			} else {
				avgMicrosec := time.Duration(uint64(x.GetSummary().GetSampleSum()) / x.GetSummary().GetSampleCount())
				table.SetCell(row, col, tview.NewTableCell(avgMicrosec.String()).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			}
		}
		col++

		if errCount != nil {
			errcnt := 0.0
			for _, em := range errCount.Metric {
				lpCat, lpSource := em.GetLabel()[0], em.GetLabel()[1]

				if lpCat.GetValue() == cat && lpSource.GetValue() == inputName {
					errcnt = em.GetCounter().GetValue()
					break
				}
			}

			table.SetCell(row, col, tview.NewTableCell(number(errcnt)).
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
		}
		row++
	}
}
