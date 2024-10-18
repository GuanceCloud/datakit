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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/strarr"
)

func (app *monitorAPP) renderWALStatTable(mfs map[string]*dto.MetricFamily, colArr []string) {
	table := app.walStatTable

	if app.anyError != nil {
		return
	}

	if mfs == nil {
		table.SetTitle("[red]WAL[white] Info(no data collected)")
		return
	}

	table.SetTitle("[red]WAL[white] Info")

	// set table header
	for idx := range colArr {
		table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).
			SetMaxWidth(app.maxTableWidth).
			SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignRight))
	}

	ptsTotal := mfs["datakit_io_wal_point_total"]

	if ptsTotal == nil {
		return
	}

	var (
		row  = 1
		cats []string
	)

	for _, m := range ptsTotal.Metric {
		for _, lp := range m.GetLabel() {
			val := lp.GetValue()
			if lp.GetName() == labelCategory && !strarr.Contains(cats, val) {
				cats = append(cats, val)
				break
			}
		}
	}

	for _, cat := range cats {
		var total,
			mem,
			disk,
			drop float64

		table.SetCell(row,
			0,
			tview.NewTableCell(cat).
				SetMaxWidth(app.maxTableWidth).
				SetAlign(tview.AlignRight))
		if x := metricWithLabel(ptsTotal, cat, "M"); x != nil {
			mem = x.GetCounter().GetValue()
		}

		if x := metricWithLabel(ptsTotal, cat, "D"); x != nil {
			disk = x.GetCounter().GetValue()
		}

		if x := metricWithLabel(ptsTotal, cat, "drop"); x != nil {
			drop = x.GetCounter().GetValue()
		}

		total = mem + disk + drop

		// only show ok points and total points.
		table.SetCell(row, 1,
			tview.NewTableCell(fmt.Sprintf("%s/%s/%s/%s",
				number(mem), number(disk), number(drop), number(total))).
				SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))

		row++
	}
}
