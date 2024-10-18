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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/strarr"
)

func (app *monitorAPP) renderDWPointsTable(mfs map[string]*dto.MetricFamily, colArr []string) {
	table := app.dwptsTable

	if app.anyError != nil {
		return
	}

	if mfs == nil {
		table.SetTitle("Point [red]U[white]pload Info(no data collected)")
		return
	}

	table.SetTitle("Point [red]U[white]pload Info")

	// set table header
	for idx := range colArr {
		table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).
			SetMaxWidth(app.maxTableWidth).
			SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignRight))
	}

	dwPtsTotal := mfs["datakit_io_dataway_point_total"]
	dwBytesTotal := mfs["datakit_io_dataway_point_bytes_total"]

	if dwPtsTotal == nil {
		return
	}

	var (
		ptsTotal,
		ptsOK float64
		row  = 1
		cats = []string{}
	)

	for _, m := range dwPtsTotal.Metric {
		for _, lp := range m.GetLabel() {
			val := lp.GetValue()
			if lp.GetName() == labelCategory && !strarr.Contains(cats, val) {
				cats = append(cats, val)
				break
			}
		}
	}

	for _, cat := range cats {
		table.SetCell(row,
			0,
			tview.NewTableCell(point.CatString(cat).Alias()).
				SetMaxWidth(app.maxTableWidth).
				SetAlign(tview.AlignRight))

		// ok points
		if x := metricWithLabel(dwPtsTotal,
			point.CatString(cat).String(), http.StatusText(http.StatusOK)); x != nil {
			ptsOK = x.GetCounter().GetValue()
		}

		// total points
		if x := metricWithLabel(dwPtsTotal,
			point.CatString(cat).String(), "total"); x != nil {
			ptsTotal = x.GetCounter().GetValue()
		}

		// only show ok points and total points.
		table.SetCell(row, 1,
			tview.NewTableCell(fmt.Sprintf("%s/%s",
				number(ptsOK), number(ptsTotal))).
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))

		if dwBytesTotal != nil {
			var bytesOk, bytesTotal, bytesGzipTotal float64
			if x := metricWithLabel(dwBytesTotal,
				point.CatString(cat).String(), "raw", http.StatusText(http.StatusOK)); x != nil {
				bytesOk = x.GetCounter().GetValue()
			}

			// total raw bytes
			if x := metricWithLabel(dwBytesTotal,
				point.CatString(cat).String(), "raw", "total"); x != nil {
				bytesTotal = x.GetCounter().GetValue()
			}

			// total gzip bytes
			if x := metricWithLabel(dwBytesTotal,
				point.CatString(cat).String(), "gzip", "total"); x != nil {
				bytesGzipTotal = x.GetCounter().GetValue()
			}

			// only show ok points and total points.
			table.SetCell(row, 2,
				tview.NewTableCell(fmt.Sprintf("%s/%s(%s)",
					number(bytesOk), number(bytesTotal), number(bytesGzipTotal))).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
		}

		row++
	}
}
