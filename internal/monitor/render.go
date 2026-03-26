// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

import (
	"fmt"
	"time"
)

// setupCols reset column names when quantile set.
func (app *monitorAPP) setupCols() {
	if app.quantile != "" {
		inputsFeedCols[3] = fmt.Sprintf("Lat(p%s)", app.quantile)
		inputsFeedCols[4] = fmt.Sprintf("Pts(p%s)", app.quantile)
		inputsFeedCols[7] = fmt.Sprintf("Cost(p%s)", app.quantile)

		httpAPIStatCols[3] = fmt.Sprintf("Lat(p%s)", app.quantile)
		httpAPIStatCols[4] = fmt.Sprintf("Body(p%s)", app.quantile)

		filterRuleCols[3] = fmt.Sprintf("Cost(p%s)", app.quantile)

		plStatsCols[7] = fmt.Sprintf("Cost(p%s)", app.quantile)

		dwCols[3] = fmt.Sprintf("Lat(p%s)", app.quantile)
	}
}

func (app *monitorAPP) render() {
	app.anyErrorPrompt.Clear()
	if app.anyError != nil {
		app.renderAnyError()
		goto end
	}

	if !app.specifiedNow {
		app.now = time.Now()
	}

	app.basicInfoTable.Clear()
	app.golangRuntime.Clear()
	app.inputsStatTable.Clear()
	app.enabledInputTable.Clear()
	app.plStatTable.Clear()
	app.walStatTable.Clear()
	app.goroutineStatTable.Clear()
	app.dwptsTable.Clear()
	app.dwTable.Clear()
	app.httpServerStatTable.Clear()
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
	app.renderWALStatTable(app.mfs, walStatsCols)
	app.renderDWPointsTable(app.mfs, dwptsStatCols)
	app.renderDatawayTable(app.mfs, dwCols)

end:
	app.exitPrompt.Clear()
	app.renderExitPrompt(app.mfs)
}
