// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package monitor implements datakit tool monitor
package monitor

import (
	"fmt"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/dustin/go-humanize"
	dto "github.com/prometheus/client_model/go"
	"github.com/rivo/tview"
)

var (
	l = logger.DefaultSLogger("monitor")

	inputsFeedCols = strings.Split(
		`Input,Cat,Feeds,TotalPts,Filtered,LastFeed,AvgCost,Errors`,
		",")
	plStatsCols      = strings.Split("Script,Cat,Namespace,TotalPts,DropPts,ErrPts,PLUpdate,AvgCost", ",")
	enabledInputCols = strings.Split(`Input,Instaces,Crashed`, ",")
	goroutineCols    = strings.Split(`Name,Running,Done,TotalCost`, ",")
	httpAPIStatCols  = strings.Split(`API,Status,Total,Latency,BodySize`, ",")
	ioStatCols       = strings.Split(`Cat,ChanUsage,UploadPts`, ",")
	filterRuleCols   = strings.Split("Cat,Total,Filtered(%),Cost", ",")

	moduleGoroutine = []string{"G", "goroutine"}
	moduleBasic     = []string{"B", "basic"}
	moduleRuntime   = []string{"R", "runtime"}
	moduleFilter    = []string{"F", "filter"}
	moduleHTTP      = []string{"H", "http"}
	moduleInputs    = []string{"In", "inputs"}
	modulePipeline  = []string{"P", "pipeline"}
	moduleIO        = []string{"IO", "io_stats"}
)

type monitorAPP struct {
	app *tview.Application

	// UI elements
	basicInfoTable      *tview.Table
	golangRuntime       *tview.Table
	inputsStatTable     *tview.Table
	plStatTable         *tview.Table
	enabledInputTable   *tview.Table
	goroutineStatTable  *tview.Table
	httpServerStatTable *tview.Table
	ioStatTable         *tview.Table

	filterStatsTable      *tview.Table
	filterRulesStatsTable *tview.Table

	exitPrompt     *tview.TextView
	anyErrorPrompt *tview.TextView

	flex *tview.Flex

	mfs map[string]*dto.MetricFamily

	inputsStats map[string]string

	anyError error

	start time.Time

	// options
	verbose                 bool
	maxTableWidth           int
	maxRun                  int
	refresh                 time.Duration
	isURL                   string
	url                     string
	onlyInputs, onlyModules []string
}

func defaultApp() *monitorAPP {
	return &monitorAPP{
		app:     tview.NewApplication(),
		start:   time.Now(),
		refresh: time.Second * 5,
		url:     "localhost:9529",
	}
}

func Start(opts ...APPOption) {
	l = logger.SLogger("monitor")

	app := defaultApp()

	for _, opt := range opts {
		if opt != nil {
			opt(app)
		}
	}

	go app.refreshData()
	app.setup()
	if err := app.run(); err != nil {
		l.Errorf("app.run: %s", err.Error())
	}
}

func (app *monitorAPP) refreshData() {
	go func() {
		tick := time.NewTicker(app.refresh)
		defer tick.Stop()
		var err error

		n := 0

		for {
			app.anyError = nil

			app.mfs, err = requestMetrics(app.url)
			if err != nil {
				app.anyError = fmt.Errorf("request stats failed: %w", err)
			}

			// app.inputsStats, err = requestInputInfo(app.isURL)
			// if err != nil {
			//	app.anyError = fmt.Errorf("request input stats failed: %w", err)
			//}

			app.render()
			app.app = app.app.Draw() // NOTE: cause DATA RACE

			<-tick.C // wait
			n++

			if app.maxRun > 0 && n >= app.maxRun {
				app.app.Stop()
				return
			}
		}
	}()
}

func (app *monitorAPP) run() error {
	return app.app.Run() // NOTE: cause DATA RACE
}

func (app *monitorAPP) inputClicked(input string) func() bool {
	if app.anyError != nil { // some error occurred, we just gone
		return func() bool {
			return true
		}
	}

	return func() bool {
		conf := "没有此采集器的配置"
		if app.inputsStats != nil && len(app.inputsStats[input]) > 0 {
			conf = app.inputsStats[input]
		}
		app.renderInputConfigView(conf)

		return true
	}
}

func number(i interface{}) string {
	switch x := i.(type) {
	case int:
		return humanize.SI(float64(x), "")
	case uint:
		return humanize.SI(float64(x), "")
	case int64:
		return humanize.SI(float64(x), "")
	case uint64:
		return humanize.SI(float64(x), "")
	case float32:
		return humanize.SI(float64(x), "")
	case float64:
		return humanize.SI(x, "")
	default:
		return ""
	}
}

func metricWithLabel(mf *dto.MetricFamily, vals ...string) *dto.Metric {
	labelMatch := func(lps []*dto.LabelPair) bool {
		if len(lps) < len(vals) {
			return false
		}

		for i, v := range vals {
			if lps[i].GetValue() != v {
				return false
			}
		}
		return true
	}

	for _, m := range mf.Metric {
		if labelMatch(m.GetLabel()) {
			return m
		}
	}
	return nil
}
