// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

import (
	"time"

	"github.com/gdamore/tcell/v2"
	dto "github.com/prometheus/client_model/go"
	"github.com/rivo/tview"
)

func (app *monitorAPP) renderDatawayTable(mfs map[string]*dto.MetricFamily, colArr []string) {
	table := app.dwTable

	if app.anyError != nil {
		return
	}

	apiLatency := mfs["datakit_io_dataway_api_latency_seconds"]
	apiRetry := mfs["datakit_io_http_retry_total"]

	if apiLatency == nil {
		app.dwTable.SetTitle("Data[red]W[white]ay Info(no data collected)")
		return
	} else {
		app.dwTable.SetTitle("Data[red]W[white]ay APIs")
	}

	// set table header
	for idx := range colArr {
		table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).
			SetMaxWidth(app.maxTableWidth).
			SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignRight))
	}

	row := 1

	for _, m := range apiLatency.Metric {
		var (
			lps         = m.GetLabel()
			api, status string
			cnt         uint64
		)

		for _, lp := range lps {
			val := lp.GetValue()

			switch lp.GetName() {
			case "api":
				api = val
				table.SetCell(row, 0,
					tview.NewTableCell(api).
						SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))

			case "status": //nolint:goconst
				status = val
				cnt = m.GetSummary().GetSampleCount()

			default:
				// pass
			}
		}

		table.SetCell(row, 1,
			tview.NewTableCell(status).
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))

		table.SetCell(row, 2,
			tview.NewTableCell(number(cnt)).
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))

		col := 3

		if x := metricWithLabel(apiLatency, api, status); x != nil {
			lat := x.GetSummary().GetSampleSum() / float64(x.GetSummary().GetSampleCount())
			table.SetCell(row, col, tview.NewTableCell(time.Duration(lat*float64(time.Second)).String()).
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
		}

		col++

		retried := 0
		if apiRetry != nil {
			x := metricWithLabel(apiRetry, api, status)
			if x == nil {
				retried = 0
			} else {
				retried = int(x.GetCounter().GetValue())
			}
		}

		table.SetCell(row, col, tview.NewTableCell(number(retried)).
			SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))

		row++
	}
}
