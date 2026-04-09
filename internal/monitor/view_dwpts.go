// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

import (
	"fmt"
	"sort"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/gdamore/tcell/v2"
	dto "github.com/prometheus/client_model/go"
	"github.com/rivo/tview"
)

func metricLabelValue(m *dto.Metric, name string) string {
	for _, lp := range m.GetLabel() {
		if lp.GetName() == name {
			return lp.GetValue()
		}
	}

	return ""
}

func sumCounterByLabels(mf *dto.MetricFamily, labels map[string]string) float64 {
	var sum float64

	if mf == nil {
		return sum
	}

	for _, m := range mf.Metric {
		matched := true
		for k, v := range labels {
			if metricLabelValue(m, k) != v {
				matched = false
				break
			}
		}

		if matched {
			sum += m.GetCounter().GetValue()
		}
	}

	return sum
}

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

	dwPtsTotal := mfs["datakit_io_endpoint_point_total"]
	dwBytesTotal := mfs["datakit_io_endpoint_point_bytes_total"]

	if dwPtsTotal == nil {
		return
	}

	var (
		row  = 1
		cats = map[string]struct{}{}
	)

	for _, m := range dwPtsTotal.Metric {
		if cat := metricLabelValue(m, labelCategory); cat != "" {
			cats[cat] = struct{}{}
		}
	}

	keys := make([]string, 0, len(cats))
	for k := range cats {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, cat := range keys {
		var (
			ptsTotal       = sumCounterByLabels(dwPtsTotal, map[string]string{labelCategory: cat, "status": "total"})
			ptsOK          = sumCounterByLabels(dwPtsTotal, map[string]string{labelCategory: cat, "status": "OK"})
			bytesTotal     = sumCounterByLabels(dwBytesTotal, map[string]string{labelCategory: cat, "enc": "raw", "status": "total"})
			bytesOK        = sumCounterByLabels(dwBytesTotal, map[string]string{labelCategory: cat, "enc": "raw", "status": "OK"})
			bytesGzipTotal = sumCounterByLabels(dwBytesTotal, map[string]string{labelCategory: cat, "enc": "gzip", "status": "total"})
		)

		table.SetCell(row,
			0,
			tview.NewTableCell(point.CatString(cat).Alias()).
				SetMaxWidth(app.maxTableWidth).
				SetAlign(tview.AlignRight))

		// only show ok points and total points.
		table.SetCell(row, 1,
			tview.NewTableCell(fmt.Sprintf("%s/%s",
				number(ptsOK), number(ptsTotal))).
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))

		if dwBytesTotal != nil {
			// only show ok points and total points.
			table.SetCell(row, 2,
				tview.NewTableCell(fmt.Sprintf("%s/%s(%s)",
					number(bytesOK), number(bytesTotal), number(bytesGzipTotal))).
					SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
		}

		row++
	}
}
