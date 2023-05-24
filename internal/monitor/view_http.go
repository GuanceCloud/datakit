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

func (app *monitorAPP) renderHTTPStatTable(mfs map[string]*dto.MetricFamily, colArr []string) {
	table := app.httpServerStatTable
	if app.anyError != nil {
		return
	}

	apiTotal := mfs["datakit_http_api_total"]
	apiElapsed := mfs["datakit_http_api_elapsed_seconds"]
	reqSize := mfs["datakit_http_api_req_size_bytes"]

	if apiTotal == nil {
		return
	} else {
		table.SetTitle(fmt.Sprintf("[red]H[white]TTP APIs(%d inputs)", len(apiTotal.Metric)))
	}

	// API,Status,Total,Limited(%),Avg Latency
	// set table header
	for idx := range colArr {
		table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).
			SetMaxWidth(app.maxTableWidth).
			SetTextColor(tcell.ColorGreen).
			SetAlign(tview.AlignRight))
	}

	row := 1

	for _, m := range apiTotal.Metric {
		lps := m.GetLabel()
		var api,
			status,
			method string

		for _, lp := range lps {
			val := lp.GetValue()

			switch lp.GetName() {
			case "api":
				api = val
			case "method":
				method = val
			case "status":
				status = val
			}
		}

		table.SetCell(row, 0,
			tview.NewTableCell(fmt.Sprintf("%s@%s", api, method)).
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))

		table.SetCell(row, 1,
			tview.NewTableCell(status).
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))

		table.SetCell(row, 2, tview.NewTableCell(number(m.GetCounter().GetValue())).
			SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))

		col := 3
		if apiElapsed != nil {
			x := metricWithLabel(apiElapsed, api, method, status)
			if x == nil {
				table.SetCell(row, col, tview.NewTableCell("-").
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			} else {
				sum := x.GetSummary()
				avg := time.Duration(float64(time.Second) * sum.GetSampleSum() / float64(sum.GetSampleCount()))

				table.SetCell(row, col, tview.NewTableCell(avg.String()).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			}
		}
		col++

		if reqSize != nil {
			x := metricWithLabel(reqSize, api, method, status)
			if x == nil {
				table.SetCell(row, col, tview.NewTableCell("-").
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			} else {
				table.SetCell(row, col, tview.NewTableCell(number(x.GetSummary().GetSampleSum())).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			}
		}

		row++
	}
}
