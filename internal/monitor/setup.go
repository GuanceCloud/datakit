// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

import (
	"os"

	"github.com/rivo/tview"
)

func (app *monitorAPP) setup() {
	// basic info
	app.basicInfoTable = tview.NewTable().SetFixed(1, 1).SetSelectable(true, false).SetBorders(false)
	app.basicInfoTable.SetBorder(true).SetTitle("[red]B[white]asic Info").SetTitleAlign(tview.AlignLeft)

	app.golangRuntime = tview.NewTable().SetFixed(1, 1).SetSelectable(true, false).SetBorders(false)
	app.golangRuntime.SetBorder(true).SetTitle("[red]R[white]untime Info").SetTitleAlign(tview.AlignLeft)

	// inputs running stats
	app.inputsStatTable = tview.NewTable().SetFixed(1, 1).SetSelectable(true, false).SetBorders(false).SetSeparator(tview.Borders.Vertical)
	app.inputsStatTable.SetBorder(true).SetTitle("[red]In[white]puts Info").SetTitleAlign(tview.AlignLeft)

	// pipeline running stats
	app.plStatTable = tview.NewTable().SetFixed(1, 1).SetSelectable(true, false).SetBorders(false).SetSeparator(tview.Borders.Vertical)
	app.plStatTable.SetBorder(true).SetTitle("[red]P[white]ipeline Info").SetTitleAlign(tview.AlignLeft)

	// enabled inputs
	app.enabledInputTable = tview.NewTable().SetFixed(1, 1).SetSelectable(true, false).SetBorders(false)
	app.enabledInputTable.SetBorder(true).SetTitle("Enabled [red]In[white]puts").SetTitleAlign(tview.AlignLeft)

	// goroutine stats
	app.goroutineStatTable = tview.NewTable().SetFixed(1, 1).SetSelectable(true, false).SetBorders(false).SetSeparator(tview.Borders.Vertical)
	app.goroutineStatTable.SetBorder(true).SetTitle("[red]G[white]oroutine Groups").SetTitleAlign(tview.AlignLeft)

	// 9592 http stats
	app.httpServerStatTable = tview.NewTable().SetFixed(1, 1).SetSelectable(true, false).SetBorders(false).SetSeparator(tview.Borders.Vertical)
	app.httpServerStatTable.SetBorder(true).SetTitle("[red]H[white]TTP APIs").SetTitleAlign(tview.AlignLeft)

	// sender stats
	app.ioStatTable = tview.NewTable().SetFixed(1, 1).SetSelectable(true, false).SetBorders(false).SetSeparator(tview.Borders.Vertical)
	app.ioStatTable.SetBorder(true).SetTitle("[red]IO[white] Info").SetTitleAlign(tview.AlignLeft)

	// filter stats
	app.filterStatsTable = tview.NewTable().SetFixed(1, 1).SetSelectable(true, false).SetBorders(false)
	app.filterStatsTable.SetBorder(true).SetTitle("[red]F[white]ilter").SetTitleAlign(tview.AlignLeft)

	app.filterRulesStatsTable = tview.NewTable().SetFixed(1, 1).SetSelectable(true, false).SetBorders(false).SetSeparator(tview.Borders.Vertical)
	app.filterRulesStatsTable.SetBorder(true).SetTitle("[red]F[white]ilter Rules").SetTitleAlign(tview.AlignLeft)

	// bottom prompt
	app.exitPrompt = tview.NewTextView().SetDynamicColors(true)
	// error prompt
	app.anyErrorPrompt = tview.NewTextView().SetDynamicColors(true)

	app.flex = tview.NewFlex()
	app.setupFlex()

	if err := app.app.SetRoot(app.flex, true).EnableMouse(true).Run(); err != nil {
		os.Exit(1)
	}
}
