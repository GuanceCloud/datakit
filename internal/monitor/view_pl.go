// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

/*
func (app *monitorAPP) renderPLStatTable(colArr []string) {
		table := m.plStatTable

		if m.anyError != nil {
			return
		}

		if len(ds.PLStats) == 0 {
			table.SetTitle("[red]P[white]ipeline Info(no data collected)")
			return
		} else {
			table.SetTitle(fmt.Sprintf("[red]P[white]ipeline Info(%d scripts)", len(ds.PLStats)))
		}

		// set table header
		for idx := range colArr {
			table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).
				SetMaxWidth(m.maxTableWidth).
				SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignRight))
		}

		row := 1
		now := time.Now()

		//
		// render all pl script, row by row
		//
		for _, plStats := range ds.PLStats {
			table.SetCell(row, 0,
				tview.NewTableCell(fmt.Sprintf("%12s", plStats.Name)).
					SetMaxWidth(m.maxTableWidth).SetAlign(tview.AlignRight))

			table.SetCell(row, 1, tview.NewTableCell(func() string {
				if v, ok := datakit.CategoryMap[plStats.Category]; ok {
					return fmt.Sprintf("%8s", v)
				} else {
					return "-"
				}
			}()).SetMaxWidth(m.maxTableWidth).SetAlign(tview.AlignCenter))

			table.SetCell(row, 2, tview.NewTableCell(func() string {
				if plStats.NS == "" {
					return "-"
				}
				return fmt.Sprintf("%8s", plStats.NS)
			}()).SetMaxWidth(m.maxTableWidth).SetAlign(tview.AlignRight))

			table.SetCell(row, 3,
				tview.NewTableCell(fmt.Sprintf("%8v", plStats.Enable)).
					SetMaxWidth(m.maxTableWidth).SetAlign(tview.AlignRight))

			table.SetCell(row, 4,
				tview.NewTableCell(fmt.Sprintf("%12s", number(plStats.Pt))).
					SetMaxWidth(m.maxTableWidth).SetAlign(tview.AlignRight))

			table.SetCell(row, 5,
				tview.NewTableCell(fmt.Sprintf("%12s", number(plStats.PtDrop))).
					SetMaxWidth(m.maxTableWidth).SetAlign(tview.AlignRight))

			table.SetCell(row, 6,
				tview.NewTableCell(fmt.Sprintf("%12s", number(plStats.PtError))).
					SetMaxWidth(m.maxTableWidth).SetAlign(tview.AlignRight))

			table.SetCell(row, 7,
				tview.NewTableCell(fmt.Sprintf("%12s", number(plStats.ScriptUpdateTimes))).
					SetMaxWidth(m.maxTableWidth).SetAlign(tview.AlignRight))

			table.SetCell(row, 8,
				tview.NewTableCell(fmt.Sprintf("%12s", time.Duration(plStats.TotalCost))).
					SetMaxWidth(m.maxTableWidth).SetAlign(tview.AlignRight))

			var avgCost time.Duration
			if plStats.Pt > 0 {
				avgCost = time.Duration(plStats.TotalCost / int64(plStats.Pt))
			}
			table.SetCell(row, 9,
				tview.NewTableCell(fmt.Sprintf("%12s", avgCost)).
					SetMaxWidth(m.maxTableWidth).SetAlign(tview.AlignRight))

			table.SetCell(row, 10,
				tview.NewTableCell(humanize.RelTime(plStats.FirstTS, now, "ago", "")).
					SetMaxWidth(m.maxTableWidth).SetAlign(tview.AlignRight))

			table.SetCell(row, 11, tview.NewTableCell(humanize.RelTime(plStats.MetaTS, now, "ago", "")).
				SetMaxWidth(m.maxTableWidth).SetAlign(tview.AlignRight))

			table.SetCell(row, 12, tview.NewTableCell(humanize.RelTime(plStats.ScriptTS, now, "ago", "")).
				SetMaxWidth(m.maxTableWidth).SetAlign(tview.AlignRight))

			table.SetCell(row, 13,
				tview.NewTableCell(fmt.Sprintf("%8v", plStats.Deleted)).
					SetMaxWidth(m.maxTableWidth).SetAlign(tview.AlignRight))

			var errInfo string
			if plStats.CompileError != "" {
				errInfo = "Compile Error: " + plStats.CompileError + "\n\n"
			}
			if len(plStats.RunLastErrs) > 0 {
				errInfo += "Run Error:\n"
				for _, e := range plStats.RunLastErrs {
					errInfo += "  " + e + "\n"
				}
			}
			click := "__________script__________\n" +
				plStats.Script +
				"__________________________\n"
			if errInfo == "" {
				errInfo = "-"
			} else {
				click = errInfo + click
			}

			lastErrCell := tview.NewTableCell(errInfo).
				SetMaxWidth(m.maxTableWidth).SetAlign(tview.AlignCenter)

			lastErrCell.SetClickedFunc(func() bool {
				m.setupLastErr(click)
				return true
			})

			table.SetCell(row, 14, lastErrCell)

			row++
		}
}

*/
