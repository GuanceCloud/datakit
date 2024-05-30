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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/election"
)

type baseInfRender struct {
	inf [7][2]string
}

func (r *baseInfRender) set(k, v string) {
	switch k {
	case "version":
		r.inf[0] = [2]string{"Version", v}
	case "build_at":
		r.inf[1] = [2]string{"Build", v}
	case "branch":
		r.inf[2] = [2]string{"Branch", v}
	case "lite", "elinker":
		r.inf[3][0] = "Build Tag"
		if v == "true" {
			r.inf[3][1] = k
		}
		if r.inf[3][1] == "" {
			r.inf[3][1] = "full"
		}
	case "os_arch":
		r.inf[4] = [2]string{"OS/Arch", v}
	case "hostname":
		r.inf[5] = [2]string{"Hostname", v}
	case "resource_limit":
		r.inf[6] = [2]string{"Resource Limit", v}
	default:
		// ignored
	}
}

func (r *baseInfRender) render(table *tview.Table, tableWidth, row int) int {
	for _, v := range r.inf {
		if v[0] != "" {
			table.SetCell(row, 0, tview.NewTableCell(v[0]).SetMaxWidth(tableWidth).SetAlign(tview.AlignRight))
			table.SetCell(row, 1, tview.NewTableCell(v[1]).SetMaxWidth(tableWidth).SetAlign(tview.AlignLeft))
			row++
		}
	}
	return row
}

func (app *monitorAPP) renderBasicInfoTable(mfs map[string]*dto.MetricFamily) {
	table := app.basicInfoTable
	row := 0

	if app.anyError != nil { // some error occurred or empty, we just gone
		table.SetCell(row, 0, tview.NewTableCell("Error").SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 1, tview.NewTableCell(app.anyError.Error()).SetMaxWidth(app.maxTableWidth).
			SetAlign(tview.AlignLeft).SetTextColor(tcell.ColorRed))
		return
	}

	mf := mfs["datakit_uptime_seconds"]
	if mf == nil {
		return
	}

	metrics := mf.GetMetric()
	if len(metrics) == 0 {
		return
	}

	table.SetCell(row, 0, tview.NewTableCell("Uptime").SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
	uptime := time.Duration(metrics[0].GetGauge().GetValue()) * time.Second
	table.SetCell(row, 1, tview.NewTableCell(uptime.String()).
		SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))
	row++

	baseinfoRender := baseInfRender{}
	for _, lp := range metrics[0].GetLabel() {
		baseinfoRender.set(lp.GetName(), lp.GetValue())
	}

	row = baseinfoRender.render(table, app.maxTableWidth, row)

	// show election info
	mf = mfs["datakit_election_status"]
	if mf != nil {
		table.SetCell(row, 0, tview.NewTableCell("Elected").SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
		str := "not-ready"
		if ei := election.MetricElectionInfo(mf); ei != nil {
			if ei.ElectedTime > 0 {
				str = fmt.Sprintf("%s::%s|%s(elected: %s)",
					ei.Namespace, ei.Status, ei.WhoElected, ei.ElectedTime.String())
			} else {
				str = fmt.Sprintf("%s::%s|%s", ei.Namespace, ei.Status, ei.WhoElected)
			}
		}

		table.SetCell(row, 1, tview.NewTableCell(str).SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))
		row++
	}

	table.SetCell(row, 0, tview.NewTableCell("From").SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(app.src.URL()).SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))
	row++

	// show proxy info
	proxyBy := "no-proxy"
	table.SetCell(row, 0, tview.NewTableCell("Proxy").SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
	if app.proxy != "" {
		proxyBy = app.proxy
	}

	table.SetCell(row, 1, tview.NewTableCell(proxyBy).SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))
}
