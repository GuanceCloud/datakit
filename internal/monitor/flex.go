// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

import (
	"github.com/rivo/tview"
)

func (app *monitorAPP) setupFlex() {
	app.flex.Clear()

	if app.verbose { // with -V, we show more stats info
		app.flex.SetDirection(tview.FlexRow).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
										AddItem(app.basicInfoTable, 0, 10, false).               // basic info
										AddItem(app.golangRuntime, 0, 10, false), 0, 10, false). // golang runtime stats
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn). // all inputs running stats
										AddItem(app.enabledInputTable, 0, 2, false). // inputs config stats
										AddItem(app.inputsStatTable, 0, 8, false), 0, 15, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn). // input config/goroutine/9529 http stats
										AddItem(app.goroutineStatTable, 0, 10, false).  // goroutine group stats
										AddItem(app.httpServerStatTable, 0, 10, false), // 9529 HTTP server stats
										0, 10, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn). // filter related stats
										AddItem(app.filterStatsTable, 0, 2, false).      // filter stats
										AddItem(app.filterRulesStatsTable, 0, 8, false), // filter rules stats
				0, 10, false).
			AddItem(app.plStatTable, 0, 15, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(app.ioStatTable, 0, 10, false).
				AddItem(app.dwTable, 0, 10, false), 0, 10, false).
			AddItem(app.anyErrorPrompt, 0, 1, false).
			AddItem(app.exitPrompt, 0, 1, false)
		return
	}

	if len(app.onlyModules) > 0 {
		flex := app.flex.SetDirection(tview.FlexRow)
		if exitsStr(app.onlyModules, moduleBasic) {
			flex.AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).AddItem(app.basicInfoTable, 0, 10, false), 0, 10, false)
		}

		if exitsStr(app.onlyModules, moduleRuntime) {
			flex.AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).AddItem(app.golangRuntime, 0, 10, false), 0, 10, false)
		}

		if exitsStr(app.onlyModules, moduleFilter) {
			flex.AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).AddItem(app.filterStatsTable, 0, 10, false), 0, 10, false)
			flex.AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).AddItem(app.filterRulesStatsTable, 0, 10, false), 0, 10, false)
		}

		if exitsStr(app.onlyModules, moduleGoroutine) {
			flex.AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).AddItem(app.goroutineStatTable, 0, 10, false), 0, 10, false)
		}

		if exitsStr(app.onlyModules, moduleHTTP) {
			flex.AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).AddItem(app.httpServerStatTable, 0, 10, false), 0, 10, false)
		}

		if exitsStr(app.onlyModules, moduleInputs) {
			flex.AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).AddItem(app.enabledInputTable, 0, 10, false), 0, 10, false)
			flex.AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).AddItem(app.inputsStatTable, 0, 10, false), 0, 10, false)
		}

		if exitsStr(app.onlyModules, modulePipeline) {
			flex.AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).AddItem(app.plStatTable, 0, 10, false), 0, 10, false)
		}

		if exitsStr(app.onlyModules, moduleIO) {
			flex.AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).AddItem(app.ioStatTable, 0, 10, false), 0, 10, false)
		}

		if exitsStr(app.onlyModules, moduleDataway) {
			flex.AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).AddItem(app.dwTable, 0, 10, false), 0, 10, false)
		}

		flex.AddItem(app.anyErrorPrompt, 0, 1, false).AddItem(app.exitPrompt, 0, 1, false)

		return
	}

	app.flex.SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
									AddItem(app.basicInfoTable, 0, 10, false).
									AddItem(app.golangRuntime, 0, 10, false), 0, 10, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn). // all inputs running stats
									AddItem(app.enabledInputTable, 0, 3, false). // inputs config stats
									AddItem(app.inputsStatTable, 0, 7, false), 0, 15, false).
		AddItem(app.anyErrorPrompt, 0, 1, false).
		AddItem(app.exitPrompt, 0, 1, false)
}

func exitsStr(sli []string, str []string) bool {
	var exits bool
	for _, m := range sli {
		for _, s := range str {
			if m == s {
				exits = true
				break
			}
		}
	}
	return exits
}
