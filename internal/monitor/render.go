// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

func (app *monitorAPP) render() {
	app.anyErrorPrompt.Clear()
	if app.anyError != nil {
		app.renderAnyError()
		goto end
	}

	app.basicInfoTable.Clear()
	app.golangRuntime.Clear()
	app.inputsStatTable.Clear()
	app.enabledInputTable.Clear()
	app.plStatTable.Clear()
	app.goroutineStatTable.Clear()
	app.ioStatTable.Clear()
	app.dwTable.Clear()
	app.filterStatsTable.Clear()
	app.filterRulesStatsTable.Clear()

	app.renderBasicInfoTable(app.mfs)
	app.renderGolangRuntimeTable(app.mfs)
	app.renderEnabledInputTable(app.mfs, enabledInputCols)

	app.renderInputsFeedTable(app.mfs, inputsFeedCols)
	app.renderGoroutineTable(app.mfs, goroutineCols)
	app.renderHTTPStatTable(app.mfs, httpAPIStatCols)

	app.renderFilterStatsTable(app.mfs)
	app.renderFilterRulesStatsTable(app.mfs, filterRuleCols)
	app.renderPLStatTable(app.mfs, plStatsCols)
	app.renderIOTable(app.mfs, ioStatCols)
	app.renderDatawayTable(app.mfs, dwCols)

end:
	app.exitPrompt.Clear()
	app.renderExitPrompt(app.mfs)
}
