// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

import (
	"fmt"
	"net/http"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/gdamore/tcell/v2"
	dto "github.com/prometheus/client_model/go"
	"github.com/rivo/tview"
)

func (app *monitorAPP) renderIOTable(mfs map[string]*dto.MetricFamily, colArr []string) {
	table := app.ioStatTable

	if app.anyError != nil {
		return
	}

	if mfs == nil {
		table.SetTitle("[red]IO[white] Info(no data collected)")
		return
	}

	table.SetTitle("[red]IO[white] Info")

	// set table header
	for idx := range colArr {
		table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).
			SetMaxWidth(app.maxTableWidth).
			SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignRight))
	}

	chanCap := mfs["datakit_io_chan_capacity"]
	chanUsage := mfs["datakit_io_chan_usage"]
	dwPtsTotal := mfs["datakit_io_dataway_point_total"]
	dwBytesTotal := mfs["datakit_io_dataway_point_bytes_total"]

	if chanUsage == nil {
		return
	}

	row := 1
	for _, m := range chanUsage.Metric {
		lps := m.GetLabel()

		var (
			cat string

			used,
			capacity int64

			ptsTotal, ptsOK float64
		)

		for _, lp := range lps {
			val := lp.GetValue()
			if lp.GetName() == "category" {
				cat = val

				table.SetCell(row, 0, tview.NewTableCell(point.CatString(cat).Alias()).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
				used = int64(m.GetGauge().GetValue())
			}
		}

		if chanCap != nil {
			x := metricWithLabel(chanCap, "all-the-same")
			if x == nil {
				capacity = 0 // should not been here
			} else {
				capacity = int64(x.GetGauge().GetValue())
			}

			table.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("%d/%d", used, capacity)).
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
		}

		if dwPtsTotal != nil {
			if x := metricWithLabel(dwPtsTotal,
				point.CatString(cat).String(), http.StatusText(http.StatusOK)); x != nil {
				ptsOK = x.GetCounter().GetValue()
			}

			if x := metricWithLabel(dwPtsTotal,
				point.CatString(cat).String(), "total"); x != nil {
				ptsTotal = x.GetCounter().GetValue()
			}

			// only show ok points and total points.
			table.SetCell(row, 2,
				tview.NewTableCell(fmt.Sprintf("%s/%s",
					number(ptsOK), number(ptsTotal))).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))

			// For failed points, there maybe more reasons(more tags), so do not
			// show here, we can see them via /metrics API.
		}

		if dwBytesTotal != nil {
			var bytesOk, bytesTotal, bytesGzipTotal float64
			if x := metricWithLabel(dwBytesTotal,
				point.CatString(cat).String(), "raw", http.StatusText(http.StatusOK)); x != nil {
				bytesOk = x.GetCounter().GetValue()
			}

			if x := metricWithLabel(dwBytesTotal,
				point.CatString(cat).String(), "raw", "total"); x != nil {
				bytesTotal = x.GetCounter().GetValue()
			}

			if x := metricWithLabel(dwBytesTotal,
				point.CatString(cat).String(), "gzip", "total"); x != nil {
				bytesGzipTotal = x.GetCounter().GetValue()
			}

			// only show ok points and total points.
			table.SetCell(row, 3,
				tview.NewTableCell(fmt.Sprintf("%s/%s(%s)",
					number(bytesOk), number(bytesTotal), number(bytesGzipTotal))).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))

			// For failed points, there maybe more reasons(more tags), so do not
			// show here, we can see them via /metrics API.
		}

		row++
	}
}
