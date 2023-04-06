// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

import (
	"fmt"
	"time"

	dto "github.com/prometheus/client_model/go"
)

func (app *monitorAPP) renderExitPrompt(mfs map[string]*dto.MetricFamily) {
	beyondUsage := false
	mf := mfs["datakit_data_overuse"]
	if mf != nil {
		m := mf.Metric[0]
		beyondUsage = (m.GetGauge().GetValue() > 0.0)
	}

	if !beyondUsage {
		fmt.Fprintf(app.exitPrompt, "[green]Refresh: %s. monitor: %s | Double ctrl+c to exit",
			app.refresh, time.Since(app.start).String())
	} else {
		fmt.Fprintf(app.exitPrompt, "[green]Refresh: %s. monitor: %s | Double ctrl+c to exit | [red] Beyond Usage",
			app.refresh, time.Since(app.start).String())
	}
}
