// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

/*

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (app *monitorAPP) setupLastErr(lastErr string) {
	if app.lastErrText != nil { // change to another `last error`
		app.lastErrText.Clear()
		app.flex.RemoveItem(app.lastErrText)
	}

	app.lastErrText = tview.NewTextView().SetWordWrap(true).SetDynamicColors(true)

	app.lastErrText.SetBorder(true)
	fmt.Fprintf(app.lastErrText, "[red]%s \n\n[green]ESC/Enter to close the message", lastErr)

	app.flex.AddItem(app.lastErrText, 0, 5, false)

	app.lastErrText.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyESC || key == tcell.KeyEnter {
			app.lastErrText.Clear()
			app.flex.RemoveItem(app.lastErrText)
		}
	})
} */
