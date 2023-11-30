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

	for _, lp := range metrics[0].GetLabel() {
		val := lp.GetValue()
		switch lp.GetName() {
		case "version":
			table.SetCell(row, 0, tview.NewTableCell("Version").SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			table.SetCell(row, 1, tview.NewTableCell(val).SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))
			row++
		case "build_at":
			table.SetCell(row, 0, tview.NewTableCell("Build").SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			table.SetCell(row, 1, tview.NewTableCell(val).SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))
			row++
		case "branch":
			table.SetCell(row, 0, tview.NewTableCell("Branch").SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			table.SetCell(row, 1, tview.NewTableCell(val).SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))
			row++
		case "lite":
			buildTag := "full"
			if val == "true" {
				buildTag = "lite"
			}
			table.SetCell(row, 0, tview.NewTableCell("Build Tag").SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			table.SetCell(row, 1, tview.NewTableCell(buildTag).SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))
			row++
		case "os_arch":
			table.SetCell(row, 0, tview.NewTableCell("OS/Arch").SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			table.SetCell(row, 1, tview.NewTableCell(val).SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))
			row++
		case "hostname":
			table.SetCell(row, 0, tview.NewTableCell("Hostname").SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			table.SetCell(row, 1, tview.NewTableCell(val).SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))
			row++
		case "resource_limit":
			table.SetCell(row, 0, tview.NewTableCell("Resource Limit").SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
			table.SetCell(row, 1, tview.NewTableCell(val).SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))
			row++
		default:
			// ignored
		}
	}

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
	table.SetCell(row, 1, tview.NewTableCell(app.url).SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))

	// show proxy info
	proxyBy := "no-proxy"
	table.SetCell(row, 0, tview.NewTableCell("Proxy").SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignRight))
	if app.proxy != "" {
		proxyBy = app.proxy
	}

	table.SetCell(row, 1, tview.NewTableCell(proxyBy).SetMaxWidth(app.maxTableWidth).SetAlign(tview.AlignLeft))
}
