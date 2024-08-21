// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

import (
	"fmt"
	"math"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/gdamore/tcell/v2"
	dto "github.com/prometheus/client_model/go"
	"github.com/rivo/tview"
)

func (app *monitorAPP) selected(x string) bool {
	if len(app.onlyInputs) == 0 {
		return true
	}

	for _, o := range app.onlyInputs {
		if o == x {
			return true
		}
	}

	return false
}

func (app *monitorAPP) renderInputsFeedTable(mfs map[string]*dto.MetricFamily, colArr []string) {
	table := app.inputsStatTable

	if app.anyError != nil {
		return
	}

	feedTotal := mfs["datakit_io_feed_total"]
	ptsSum := mfs["datakit_io_feed_point"]
	lastFeed := mfs["datakit_io_last_feed_timestamp_seconds"]
	cost := mfs["datakit_input_collect_latency_seconds"]
	ptsFilter := mfs["datakit_io_input_filter_point_total"]
	errCount := mfs["datakit_error_total"]
	feedCost := mfs["datakit_io_feed_cost_seconds"]

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
		var (
			inputName,
			cat string

			lps = m.GetLabel()
		)

		for _, lp := range lps {
			val := lp.GetValue()

			switch lp.GetName() {
			case "name":
				inputName = val

			case "category": //nolint:goconst
				cat = val

			default:
				// pass
			}
		}

		if !app.selected(inputName) {
			continue
		}

		// Input
		table.SetCell(row, 0,
			tview.NewTableCell(inputName).
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))

		// Cat
		table.SetCell(row, 1,
			tview.NewTableCell(point.CatString(cat).Alias() /* metric -> M */).
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))

		col := 2

		// Feeds
		table.SetCell(row, col, tview.NewTableCell(number(m.GetCounter().GetValue())).
			SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
		col++

		// P90Lat
		feedSum := metricWithLabel(feedCost, cat, inputName).GetSummary()
		feedLat := "-"
		if feedSum != nil {
			q := feedSum.GetQuantile()[1] // p90
			if v := q.GetValue(); math.IsNaN(v) {
				feedLat = "NaN"
			} else {
				feedLat = time.Duration(v * float64(time.Second)).String()
			}
		}
		table.SetCell(row, col, tview.NewTableCell(feedLat).SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
		col++

		// P90Pts
		if ptsSum != nil {
			x := metricWithLabel(ptsSum, cat, inputName)
			p90pts := "-"
			if x != nil {
				p90pts = number(x.GetSummary().GetQuantile()[1].GetValue())
			}
			table.SetCell(row, col, tview.NewTableCell(p90pts).
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
		}
		col++

		// Filtered
		table.SetCell(row, col, tview.NewTableCell("-").
			SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
		if ptsFilter != nil {
			x := metricWithLabel(ptsFilter, cat, inputName)
			if x != nil {
				table.SetCell(row, col, tview.NewTableCell(number(x.GetCounter().GetValue())).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			}
		}
		col++

		// LastFeed
		if lastFeed != nil {
			x := metricWithLabel(lastFeed, cat, inputName)
			if x == nil {
				table.SetCell(row, col, tview.NewTableCell("-").
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
			} else {
				since := fmt.Sprintf("%s ago", app.now.Sub(time.Unix(int64(x.GetGauge().GetValue()), 0)))

				table.SetCell(row, col, tview.NewTableCell(since).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			}
		}
		col++

		// AvgCost
		if cost != nil {
			x := metricWithLabel(cost, cat, inputName)
			if x == nil {
				table.SetCell(row, col, tview.NewTableCell("-").
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignCenter))
			} else {
				cost := time.Duration(
					float64(time.Second) * x.GetSummary().GetSampleSum() /
						float64(x.GetSummary().GetSampleCount()))
				table.SetCell(row, col, tview.NewTableCell(cost.String()).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			}
		}
		col++

		// Errors
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
